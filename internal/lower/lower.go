// Package lower turns a checked program into Adamic's IR.
//
// Stage 0 lowers a little and refuses the rest. Every construct it can't lower yet is an error
// naming the construct and where it is, never a skip: a compiler that drops a statement it doesn't
// understand produces a program that looks right and isn't, which is the one thing Adamic must
// never do.
package lower

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// NotYet is a construct stage 0 can't lower yet, with where it is.
type NotYet struct {
	Where string
	What  string
}

func (n *NotYet) Error() string {
	return fmt.Sprintf("%s: stage 0 can't lower %s yet", n.Where, n.What)
}

// Refused is something Adamic 0.1 doesn't allow at all (docs/0.1.md), as opposed to something stage 0
// hasn't learned yet. The difference is a promise, so the message keeps them apart.
type Refused struct {
	Where string
	What  string
	Fix   string
}

func (r *Refused) Error() string {
	return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %s", r.Where, r.What, r.Fix)
}

// Lower lowers a checked program, entry file first.
func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {
	files := program.Files()
	if len(files) != 1 {
		return nil, fmt.Errorf("lower: stage 0 compiles a program from one entry file, got %d", len(files))
	}
	entry := files[0]
	// Stage 0 checks single-threaded, so one checker answers for every file.
	typeChecker, release := program.Checker(ctx, entry)
	defer release()

	lowering := &lowering{program: program, checker: typeChecker, result: &ir.Program{}, this: -1, functionIndex: -1}
	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(entry))
	modules, err := lowering.moduleOrder(entry)
	if err != nil {
		return nil, err
	}
	lowering.noteInheritance(modules)
	for _, module := range modules {
		if err := lowering.refuse(module); err != nil {
			return nil, err
		}
	}
	for _, module := range modules {
		if err := lowering.declareModule(module.Statements.Nodes); err != nil {
			return nil, err
		}
	}
	for _, module := range modules {
		body, err := lowering.statements(module.Statements.Nodes)
		if err != nil {
			return nil, err
		}
		lowering.result.Main = append(lowering.result.Main, body...)
	}
	if lowering.unlowerable != nil {
		return nil, lowering.unlowerable
	}
	lowering.result.Main = append(lowering.forwarderValues, lowering.result.Main...)
	lowering.finishClassCalls()
	if err := lowering.exceptions(); err != nil {
		return nil, err
	}
	if err := lowering.findCycles(modules); err != nil {
		return nil, err
	}
	borrow(lowering.result)
	counters(lowering.result)
	return lowering.result, nil
}

type lowering struct {
	program *load.Program
	checker *checker.Checker
	result  *ir.Program

	// strings indexes result.Strings, so a constant used twice is stored once.
	strings map[string]int

	// locals maps each variable's symbol to its index in result.Locals.
	locals map[*ast.Symbol]int

	// functions maps each function declaration's symbol to its index in result.Functions.
	functions map[*ast.Symbol]int

	// function is the function being lowered, or nil for the module's top level.
	function *ir.Function

	// this is the local this is in a method or constructor, or -1.
	this int

	// functionIndex is the function being lowered, -1 for the module's top level, and closures the
	// closures being lowered, outermost first, each inside the one before.
	functionIndex int
	closures      []int

	// classes maps each module class's symbol to its declaration, and instances each instantiation
	// already lowered (class.go).
	classes          map[*ast.Symbol]*ast.Node
	instances        map[string]*instance
	derivedAncestors map[*ast.Symbol]bool

	// instance is the class instantiation being lowered, if any.
	instance *instance

	// substitution is what each type parameter stands for in the instantiation being lowered.
	substitution map[*checker.Type]ir.Type

	// unlowerable is the first construct found not lowerable somewhere that can't return an error.
	unlowerable error

	// alwaysUndefined are parameters that only ever receive undefined, whatever the checker calls
	// their type: the first of an Array.from callback's (from.go). Each is a reference that's always
	// missing.
	alwaysUndefined map[*ast.Symbol]bool

	// signed are the functions whose signatures are written and whose bodies aren't lowered yet, by
	// index, each with what its body still needs (signature).
	signed map[int]signed

	// forwarders are the globals holding the function values made for module functions read as values,
	// by the function each forwards to, and forwarderValues the declarations that make them, which run
	// before anything else (functionValue).
	forwarders      map[int]int
	forwarderValues []ir.Statement

	// unsetUntil is where, in the constructor being lowered, its last assignment of a field without an
	// initializer ends: before it, this is only for reading and writing fields (useOfThis).
	unsetUntil int

	// generics maps each generic module function's symbol to its declaration, and genericInstances
	// each instantiation already lowered to its function (generic.go). genericDepth counts the
	// instantiations being lowered inside one another.
	generics         map[*ast.Symbol]*ast.Node
	genericInstances map[string]int
	genericDepth     int

	// caught are the variables a catch binds, each always an Error; tries are the try statements
	// lowered (exceptions.go).
	caught map[*ast.Symbol]bool
	tries  []tryRecord

	// initializing is the variables whose initializers are being lowered, by where each is declared.
	initializing map[int]*ast.Node

	// For the cycle finder (cycles.go): the checker's type of each local, and where it's declared;
	// every function value made, with its type; and the class type being instantiated, for this.
	localTypes     map[int]*checker.Type
	localAlso      map[int][]*checker.Type
	localNodes     map[int]*ast.Node
	closureRecords []closureRecord
	classType      *checker.Type
	classNode      *ast.Node

	// typeMapper is what the type parameters of the instantiation being lowered, and of those it's
	// inside, stand for; instantiated is every class type an instantiation was made for (instantiate.go).
	typeMapper   *typeMapper
	instantiated []*checker.Type

	// writeSites is every write into a slot lowering made, by its IR node's Site less one (fresh.go).
	writeSites []writeSite
}

// moduleOrder is the order the program's modules run in, ECMAScript's: each module's imports first,
// depth-first in the order they're written, then the module itself. The prelude's 'adamic' module
// has no body to run. An import cycle is refused, as 0.1 says (docs/0.1.md).
func (l *lowering) moduleOrder(entry *ast.SourceFile) ([]*ast.SourceFile, error) {
	order := []*ast.SourceFile{}
	state := map[*ast.SourceFile]int{} // 1 while its imports are being visited, 2 once placed
	var visit func(module *ast.SourceFile, from *ast.Node) error
	visit = func(module *ast.SourceFile, from *ast.Node) error {
		switch state[module] {
		case 1:
			return &Refused{Where: l.program.Where(from), What: "an import cycle", Fix: "move what both modules need into a third that neither imports"}
		case 2:
			return nil
		}
		state[module] = 1
		for _, statement := range module.Statements.Nodes {
			if statement.Kind != ast.KindImportDeclaration && statement.Kind != ast.KindExportDeclaration {
				continue
			}
			specifier := statement.ModuleSpecifier()
			if specifier == nil {
				continue
			}
			target := l.checker.GetSymbolAtLocation(specifier)
			if target == nil || len(target.Declarations) == 0 || target.Declarations[0].Kind != ast.KindSourceFile {
				// 'adamic' is an ambient module in the prelude: nothing to run.
				continue
			}
			imported := target.Declarations[0].AsSourceFile()
			if load.IsPrelude(imported) || load.IsLibrary(imported) {
				continue
			}
			if err := visit(imported, statement); err != nil {
				return err
			}
		}
		state[module] = 2
		order = append(order, module)
		return nil
	}
	if err := visit(entry, entry.AsNode()); err != nil {
		return nil, err
	}
	return order, nil
}

// declareModule registers the module's globals and functions before anything is lowered, so a
// function can call one declared below it and read a global declared below it, as JavaScript
// allows. Every function's signature is written, from the checker, before any body is lowered, so a
// call knows what the function it calls returns wherever that function is declared, and two
// functions can call each other. Then each body is lowered.
func (l *lowering) declareModule(statements []*ast.Node) error {
	declarations := []*ast.Node{}
	for _, statement := range statements {
		switch statement.Kind {
		case ast.KindVariableStatement:
			list := statement.AsVariableStatement().DeclarationList
			if list.Flags&ast.NodeFlagsBlockScoped == 0 {
				continue // statements() refuses var where it stands
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				// A module's const [a, b] = tuple, or { x, y } = object, declares globals too, each name
				// its own.
				if declaration.Name().Kind == ast.KindArrayBindingPattern || declaration.Name().Kind == ast.KindObjectBindingPattern {
					for _, binding := range declaration.Name().AsBindingPattern().Elements.Nodes {
						if skipped(binding) || !ast.IsIdentifier(binding.Name()) {
							continue
						}
						local, err := l.declareLocal(binding.Name())
						if err != nil {
							return err
						}
						l.result.Locals[local].Global = true
					}
					continue
				}
				if ast.IsIdentifier(declaration.Name()) {
					local, err := l.declareLocal(declaration.Name())
					if err != nil {
						return err
					}
					l.result.Locals[local].Global = true
				}
			}
		case ast.KindClassDeclaration:
			if l.classes == nil {
				l.classes = map[*ast.Symbol]*ast.Node{}
			}
			l.classes[l.symbol(statement.Name())] = statement
		case ast.KindFunctionDeclaration:
			symbol := l.symbol(statement.Name())
			if len(statement.TypeParameters()) > 0 {
				// A generic function is lowered once per instantiation, where it's called (generic.go).
				if l.generics == nil {
					l.generics = map[*ast.Symbol]*ast.Node{}
				}
				l.generics[symbol] = statement
				continue
			}
			if l.functions == nil {
				l.functions = map[*ast.Symbol]int{}
			}
			l.functions[symbol] = len(l.result.Functions)
			l.result.Functions = append(l.result.Functions, ir.Function{Name: statement.Name().Text()})
			declarations = append(declarations, statement)
		}
	}
	for _, declaration := range declarations {
		if err := l.signature(l.functions[l.symbol(declaration.Name())], declaration, -1); err != nil {
			return err
		}
	}
	for _, declaration := range declarations {
		if err := l.functionBody(declaration); err != nil {
			return err
		}
	}
	return nil
}

// functionBody lowers a module function declaration's parameters and body.
func (l *lowering) functionBody(declaration *ast.Node) error {
	return l.lowerFunction(l.functions[l.symbol(declaration.Name())], declaration, -1)
}

// defaulted is a parameter with a default: the local the body declares, the one the argument arrives
// in, and the default.
type defaulted struct {
	local, incoming int
	initializer     *ast.Node
}

// signed is what a function whose signature is written still needs to lower its body: the local its
// this is, or -1, its parameters' defaults, and its destructured parameters.
type signed struct {
	this     int
	defaults []defaulted
	patterns []patterned
}

// patterned is a destructured parameter, ([key, value]) or ({ x, y }): the pattern, the parameter,
// and the local it arrives in whole.
type patterned struct {
	pattern   *ast.Node
	parameter *ast.Node
	incoming  int
}

// lowerFunction lowers a function-like declaration into the function at index: its signature, unless
// signature has written it already, then its body. this, when not -1, is the local a method's or
// constructor's this is, passed first.
//
// It works on a copy and writes it back by index: lowering a body can instantiate a class, which
// appends functions, and a pointer into the slice would be left pointing at the old one.
func (l *lowering) lowerFunction(index int, declaration *ast.Node, this int) error {
	if _, isSigned := l.signed[index]; !isSigned {
		if err := l.signature(index, declaration, this); err != nil {
			return err
		}
	}
	pending := l.signed[index]
	delete(l.signed, index)
	return l.lowerBody(index, declaration, pending.this, pending.defaults, pending.patterns)
}

// signature writes the function at index's parameters and result, from the checker, without lowering
// its body, so a call to it lowers whether or not its body has been. this is as for lowerFunction.
func (l *lowering) signature(index int, declaration *ast.Node, this int) error {
	function := l.result.Functions[index]
	if this >= 0 && declaration.Kind != ast.KindConstructor {
		// A method receives this; a constructor makes it.
		function.Parameters = append(function.Parameters, this)
	}
	if declaration.Kind != ast.KindConstructor {
		signature := l.checker.GetSignatureFromDeclaration(declaration)
		returns := l.checker.GetReturnTypeOfSignature(signature)
		// A function that never returns (it panics on every path, as (why) => panic(why) does) has no
		// result to hold, as one returning void hasn't. An arrow whose expression is never for another
		// reason, a variable the checker narrowed to nothing, isn't one.
		neverArrow := returns.Flags()&checker.TypeFlagsNever != 0 && declaration.Body() != nil && declaration.Body().Kind != ast.KindBlock && !l.isPanicCall(declaration.Body())
		if returns.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 || neverArrow {
			valueType, isKnown := l.representation(returns)
			if !isKnown {
				// An arrow function has no name to point at, so it's pointed at whole.
				where := declaration.Name()
				if where == nil {
					where = declaration
				}
				return l.notYet(where, "a function returning "+l.checker.TypeToString(returns))
			}
			function.Returns = valueType
		}
	}
	outerIndexForParameters := l.functionIndex
	l.functionIndex = index
	defer func() { l.functionIndex = outerIndexForParameters }()
	// A parameter with a default arrives as what may be missing, and the body begins by declaring the
	// parameter itself: the argument, or the default when it's undefined, as JavaScript decides.
	defaults := []defaulted{}
	// A destructured parameter, ([key, value]) or ({ x, y }), arrives whole in a parameter of its own,
	// and its names are declared from it before the body runs.
	patterns := []patterned{}
	for _, parameter := range declaration.Parameters() {
		declared := parameter.AsParameterDeclaration()
		if name := parameter.Name(); (name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern) && declared.DotDotDotToken == nil && declared.Initializer == nil && declared.QuestionToken == nil {
			incoming := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "destructured", Type: ir.Object, Function: index})
			function.Parameters = append(function.Parameters, incoming)
			patterns = append(patterns, patterned{pattern: name, parameter: parameter, incoming: incoming})
			continue
		}
		if !ast.IsIdentifier(parameter.Name()) || declared.DotDotDotToken != nil {
			return l.notYet(parameter, "a parameter that isn't a plain name")
		}
		local, err := l.declareLocal(parameter.Name())
		if err != nil {
			return err
		}
		if function.Closure && (declared.Initializer != nil || declared.QuestionToken != nil) {
			// A function value is called with the arguments its caller has, and no more.
			return l.notYet(parameter, "a function value with an optional parameter")
		}
		if function.Closure && slotless(l.result.Locals[local].Type) {
			// Its arguments are each one adamic_value.
			return l.notYet(parameter, "a function value taking "+l.checker.TypeToString(l.checker.GetTypeAtLocation(parameter.Name())))
		}
		if declared.Initializer == nil {
			function.Parameters = append(function.Parameters, local)
			continue
		}
		missing := ir.Maybe(l.result.Locals[local].Type)
		if !missing.IsMaybe() && !missing.IsReference() {
			return l.notYet(parameter, "a default for a "+typeName(missing)+" parameter")
		}
		incoming := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: parameter.Name().Text(), Type: missing, Function: index})
		function.Parameters = append(function.Parameters, incoming)
		defaults = append(defaults, defaulted{local: local, incoming: incoming, initializer: declared.Initializer})
	}
	if function.Closure && slotless(function.Returns) {
		// A function value's arguments and result are each one adamic_value, and number | undefined
		// needs two words.
		return l.notYet(declaration, "a function value returning "+typeName(function.Returns))
	}
	if declaration.Body() == nil && !ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAbstract) {
		return l.notYet(declaration, "a function without a body")
	}
	l.result.Functions[index] = function
	if l.signed == nil {
		l.signed = map[int]signed{}
	}
	l.signed[index] = signed{this: this, defaults: defaults, patterns: patterns}
	return nil
}

// lowerBody lowers the body of the function at index, whose signature is written.
func (l *lowering) lowerBody(index int, declaration *ast.Node, this int, defaults []defaulted, patterns []patterned) error {
	function := l.result.Functions[index]
	if !function.Closure {
		// A function declaration, a method or a constructor has its body lowered where a use of it is first met,
		// a closure's body included, but it's declared at the top level and captures nothing from
		// the closures being lowered there: a local of its own read in a closure of its own must not
		// land in their environments.
		outerClosures := l.closures
		l.closures = nil
		defer func() { l.closures = outerClosures }()
	}
	body := declaration.Body()
	outer, outerThis, outerIndex := l.function, l.this, l.functionIndex
	l.function, l.functionIndex = &function, index
	if this >= 0 {
		l.this = this
	}
	// Defaults run before the body, in order, after a constructor's fields, as JavaScript runs them.
	var prologue, lowered []ir.Statement
	var err error
	if len(patterns) > 0 && len(defaults) > 0 {
		// JavaScript binds parameters in order, a default perhaps reading a name destructured
		// before it; the two together aren't lowered yet.
		err = l.notYet(declaration, "a destructured parameter beside a parameter with a default")
	}
	for _, parameter := range patterns {
		if err != nil {
			break
		}
		var destructured []ir.Statement
		patternType := l.checker.GetTypeAtLocation(parameter.parameter)
		heldAs, _ := l.representation(patternType)
		destructured, err = l.destructureFrom(parameter.pattern, patternType, heldAs, parameter.incoming)
		prologue = append(prologue, destructured...)
	}
	for _, parameter := range defaults {
		if err != nil {
			break
		}
		var fallback ir.Expression
		if fallback, err = l.expression(parameter.initializer); err != nil {
			break
		}
		of := l.result.Locals[parameter.local].Type
		if fallback = fit(fallback, of); fallback.Type() != of {
			err = l.notYet(parameter.initializer, "a default of another type than its parameter")
			break
		}
		incoming := ir.Read{Local: parameter.incoming, Of: l.result.Locals[parameter.incoming].Type}
		prologue = append(prologue, ir.Declare{Local: parameter.local, Value: ir.Coalesce{Value: incoming, Fallback: fallback, Of: of}})
	}
	if err != nil {
		l.function, l.this, l.functionIndex = outer, outerThis, outerIndex
		return err
	}
	if body.Kind == ast.KindBlock {
		lowered, err = l.statements(body.AsBlock().Statements.Nodes)
	} else if function.Returns == 0 || l.isPanicCall(body) {
		// An arrow function's expression body, when it returns nothing or never returns (() =>
		// panic('why')), is a statement.
		lowered, err = l.expressionStatement(body)
	} else {
		// Otherwise its value is what it returns.
		var value ir.Expression
		if value, err = l.expression(body); err == nil {
			lowered = []ir.Statement{ir.Return{Value: fit(value, function.Returns)}}
		}
	}
	l.function, l.this, l.functionIndex = outer, outerThis, outerIndex
	if err != nil {
		return err
	}
	// The environment may have grown while the body was lowered (captures are found as they're
	// read), so it's taken from what's recorded, not from this copy.
	function.Environment = l.result.Functions[index].Environment
	function.Body = append(function.Body, prologue...)
	function.Body = append(function.Body, lowered...)
	l.result.Functions[index] = function
	return nil
}
func (l *lowering) notYet(node *ast.Node, what string) error {
	return &NotYet{Where: l.program.Where(node), What: what}
}

// statements lowers a list of statements.
func (l *lowering) statements(nodes []*ast.Node) ([]ir.Statement, error) {
	lowered := []ir.Statement{}
	for _, node := range nodes {
		statements, err := l.statement(node)
		if err != nil {
			return nil, err
		}
		lowered = append(lowered, statements...)
	}
	return lowered, nil
}

// statement lowers one statement to none, one or several.
func (l *lowering) statement(node *ast.Node) ([]ir.Statement, error) {
	switch node.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		// Types erase to nothing, and so does an empty statement.
		return nil, nil
	case ast.KindImportDeclaration:
		// What an import brings in is resolved through the checker at each use, and the module it
		// names runs first (moduleOrder).
		return nil, nil
	case ast.KindExportDeclaration, ast.KindExportAssignment:
		return nil, &Refused{Where: l.program.Where(node), What: describe(node), Fix: "export where you declare: export function, export const (one name for one thing)"}
	case ast.KindFunctionDeclaration:
		if l.function != nil {
			return nil, l.notYet(node, "a function inside a function (a closure)")
		}
		// Lowered already, by declareModule.
		return nil, nil
	case ast.KindClassDeclaration:
		if l.function != nil {
			return nil, l.notYet(node, "a class inside a function")
		}
		// Lowered at each instantiation, by instantiate.
		return nil, nil
	case ast.KindReturnStatement:
		return l.returnStatement(node)
	case ast.KindExpressionStatement:
		return l.expressionStatement(node.AsExpressionStatement().Expression)
	case ast.KindVariableStatement:
		return l.variables(node.AsVariableStatement().DeclarationList)
	case ast.KindBlock:
		body, err := l.statements(node.AsBlock().Statements.Nodes)
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Block{Body: body}}, nil
	case ast.KindIfStatement:
		return l.ifStatement(node)
	case ast.KindForStatement:
		return l.forStatement(node)
	case ast.KindWhileStatement, ast.KindDoStatement:
		return l.whileStatement(node)
	case ast.KindForOfStatement:
		return l.forOf(node)
	case ast.KindSwitchStatement:
		return l.switchStatement(node)
	case ast.KindBreakStatement, ast.KindContinueStatement:
		if node.Label() != nil {
			return nil, l.notYet(node, "a labeled "+strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(node.Kind.String(), "Kind"), "Statement")))
		}
		if node.Kind == ast.KindBreakStatement {
			return []ir.Statement{ir.Break{}}, nil
		}
		return []ir.Statement{ir.Continue{}}, nil
	case ast.KindThrowStatement:
		return l.throwStatement(node)
	case ast.KindTryStatement:
		return l.tryStatement(node)
	}
	return nil, l.notYet(node, describe(node))
}

// expressionStatement lowers an expression used as a statement: a console call, an assignment, or
// ++ and --. Any other expression's value would be thrown away, and stage 0 doesn't lower that yet.
func (l *lowering) expressionStatement(expression *ast.Node) ([]ir.Statement, error) {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindCallExpression:
		if ast.SkipParentheses(expression.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
			return l.superStatement(expression)
		}
		if err := l.optionalCall(expression); err != nil {
			return nil, err
		}
		if l.isConsole(expression.AsCallExpression().Expression) {
			statement, err := l.console(expression)
			if err != nil {
				return nil, err
			}
			return []ir.Statement{statement}, nil
		}
		if l.isPreludeFunction(expression.AsCallExpression().Expression, "panic") {
			arguments := expression.AsCallExpression().Arguments.Nodes
			if len(arguments) != 1 {
				return nil, errors.New("lower: " + l.program.Where(expression) + ": panic takes one argument, and the checker let another count through")
			}
			message, err := l.expression(arguments[0])
			if err != nil {
				return nil, err
			}
			return []ir.Statement{ir.Panic{Message: message}}, nil
		}
		// A builtin's result thrown away, like map.set(key, value) or array.push(value).
		if lowered, isBuiltin, err := l.builtin(expression); isBuiltin {
			if err != nil {
				return nil, err
			}
			return []ir.Statement{ir.Evaluate{Value: lowered}}, nil
		}
		// A call for its effects, void or not.
		call, err := l.callOrMethod(expression)
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Evaluate{Value: call}}, nil
	case ast.KindBinaryExpression:
		return l.assignment(expression)
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		return l.increment(expression)
	}
	return nil, l.notYet(expression, describe(expression)+" as a statement")
}

// console lowers console.log and console.error with one string.
func (l *lowering) console(call *ast.Node) (ir.Statement, error) {
	stream, err := l.consoleStream(call.AsCallExpression().Expression)
	if err != nil {
		return nil, err
	}
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		// The prelude declares one parameter, so the checker has already refused any other count.
		return nil, errors.New("lower: " + l.program.Where(call) + ": console takes one argument, and the checker let another count through")
	}
	value, err := l.expression(arguments[0])
	if err != nil {
		return nil, err
	}
	return ir.WriteLine{Stream: stream, Value: value}, nil
}

// variables lowers const and let declarations, each to a local of its own.
func (l *lowering) variables(list *ast.Node) ([]ir.Statement, error) {
	if list.Flags&ast.NodeFlagsBlockScoped == 0 {
		return nil, &Refused{Where: l.program.Where(list), What: "var", Fix: "use const or let"}
	}
	statements := []ir.Statement{}
	for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
		name := declaration.Name()
		if name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern {
			// const [a, b] = tuple, and const { x, y } = object (collections.go).
			destructured, err := l.destructure(name, declaration.AsVariableDeclaration().Initializer)
			if err != nil {
				return nil, err
			}
			statements = append(statements, destructured...)
			continue
		}
		if !ast.IsIdentifier(name) {
			return nil, l.notYet(name, "a destructuring declaration")
		}
		local, err := l.declareLocal(name)
		if err != nil {
			return nil, err
		}
		var value ir.Expression
		if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
			if l.initializing == nil {
				l.initializing = map[int]*ast.Node{}
			}
			l.initializing[local] = name
			value, err = l.expression(initializer)
			delete(l.initializing, local)
			if err != nil {
				return nil, err
			}
		}
		statements = append(statements, ir.Declare{Local: local, Value: fit(value, l.result.Locals[local].Type)})
	}
	return statements, nil
}

// skipped reports whether an element of an array binding pattern is a hole, as in const [, b]: the
// checker's tree has a binding element with no name there.
func skipped(binding *ast.Node) bool {
	return binding.Kind == ast.KindOmittedExpression || binding.Name() == nil
}

// declareLocal makes a new local for a declared name, typed by what the checker proved for it.
func (l *lowering) declareLocal(name *ast.Node) (int, error) {
	symbol := l.symbol(name)
	if symbol == nil {
		return 0, errors.New("lower: " + l.program.Where(name) + ": the checker gave a declaration no symbol")
	}
	valueType := ir.Object
	if !l.alwaysUndefined[symbol] && !l.caught[symbol] {
		var err error
		if valueType, err = l.typeOf(name); err != nil {
			return 0, err
		}
	}
	if l.locals == nil {
		l.locals = map[*ast.Symbol]int{}
	}
	if local, isDeclared := l.locals[symbol]; isDeclared {
		// A global, registered before anything was lowered.
		return local, nil
	}
	l.locals[symbol] = len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name.Text(), Type: valueType, Function: l.functionIndex})
	l.noteLocal(l.locals[symbol], l.checker.GetTypeAtLocation(name), name)
	return l.locals[symbol], nil
}

// noteAlso keeps another checker type a local has: a this shared by more than one instantiation.
func (l *lowering) noteAlso(local int, proven *checker.Type) {
	if l.localAlso == nil {
		l.localAlso = map[int][]*checker.Type{}
	}
	l.localAlso[local] = append(l.localAlso[local], proven)
}

// noteLocal keeps a local's checker type and declaration for the cycle finder.
func (l *lowering) noteLocal(local int, proven *checker.Type, node *ast.Node) {
	if l.localTypes == nil {
		l.localTypes, l.localNodes = map[int]*checker.Type{}, map[int]*ast.Node{}
	}
	l.localTypes[local], l.localNodes[local] = l.concrete(proven), node
	if l.instance != nil {
		l.instance.templates = append(l.instance.templates, template{local: local, proven: proven})
	}
}

// symbol is what a name refers to, the same whichever file names it: an import resolves to what it
// imports, and an exported declaration to its export symbol, so one declaration is one symbol.
func (l *lowering) symbol(node *ast.Node) *ast.Symbol {
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	return l.checker.GetExportSymbolOfSymbol(symbol)
}

// local finds the local an identifier refers to, and notes a capture if it belongs to another
// function.
func (l *lowering) local(identifier *ast.Node) (int, bool) {
	symbol := l.symbol(identifier)
	local, isLocal := l.locals[symbol]
	if isLocal {
		l.touch(local)
		if declared := l.result.Locals[local]; declared.Captured && slotless(declared.Type) && l.unlowerable == nil {
			// A cell holds one adamic_value, and number | undefined needs two words. Lower says so once
			// it's done, since the capture is found here, where nothing can return an error.
			l.unlowerable = l.notYet(identifier, "a "+typeName(declared.Type)+" variable a function value captures")
		}
	}
	return local, isLocal
}

// touch notes that the function being lowered reads or writes a local. One declared in another
// function, and not a global, is captured: it moves into a cell, and every closure being lowered
// between its function and this one carries that cell in its environment.
func (l *lowering) touch(local int) {
	declared := l.result.Locals[local]
	if declared.Global || declared.Function == l.functionIndex {
		return
	}
	l.result.Locals[local].Captured = true
	if name, isInitializing := l.initializing[local]; isInitializing && l.unlowerable == nil {
		// const f = () => f(): the function value is made before the variable's cell is, so it has
		// nothing to capture yet.
		l.unlowerable = l.notYet(name, "a function value that captures the variable its own initializer declares")
	}
	start := 0
	for position, closure := range l.closures {
		if closure == declared.Function {
			start = position + 1
		}
	}
	for _, closure := range l.closures[start:] {
		function := &l.result.Functions[closure]
		if !slices.Contains(function.Environment, local) {
			function.Environment = append(function.Environment, local)
		}
	}
}

// assignment lowers =, and the compound assignments, to a local.
func (l *lowering) assignment(node *ast.Node) ([]ir.Statement, error) {
	binary := node.AsBinaryExpression()
	operator, isCompound := compoundAssignments[binary.OperatorToken.Kind]
	if binary.OperatorToken.Kind != ast.KindEqualsToken && !isCompound {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	target := ast.SkipParentheses(binary.Left)
	if target.Kind == ast.KindPropertyAccessExpression {
		if isCompound {
			return l.updateProperty(node, target, operator, binary.Right)
		}
		return l.setProperty(target, binary.Right)
	}
	// array[index] += value never reaches here: the element may be missing, so the checker refuses it
	// (noUncheckedIndexedAccess).
	if target.Kind == ast.KindElementAccessExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		return l.setIndex(target, binary.Right)
	}
	if target.Kind == ast.KindArrayLiteralExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		return l.destructuringAssignment(target, binary.Right)
	}
	if target.Kind == ast.KindElementAccessExpression && isCompound {
		return l.updateIndex(node, target, operator, binary.Right)
	}
	local, isLocal := l.local(target)
	if !ast.IsIdentifier(target) || !isLocal {
		return nil, l.notYet(target, "assigning to "+describe(target))
	}
	if l.caught[l.symbol(target)] {
		return nil, l.notYet(target, "assigning to what a catch caught")
	}
	if l.alwaysUndefined[l.symbol(target)] {
		// Its type is unknown, so anything could be written to it, and it holds only undefined.
		return nil, l.notYet(target, "assigning to a parameter that only ever receives undefined")
	}
	value, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	if isCompound {
		current := ir.Expression(ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)})
		if operator == ast.KindPlusToken {
			current, value = l.spelled(target, current), l.spelled(binary.Right, value)
		}
		if value, err = l.combine(node, operator, current, value); err != nil {
			return nil, err
		}
	}
	return []ir.Statement{ir.Assign{Local: local, Value: fit(value, l.result.Locals[local].Type), Checked: l.checked(local)}}, nil
}

// tupleField reads element index of the tuple held in the local held, as the tuple's element type
// holds it, and makes it what the name it goes to holds: a string element going to a string | number
// name is boxed on the way.
func (l *lowering) tupleField(where *ast.Node, held int, elements []*checker.Type, index int, to ir.Type) (ir.Expression, error) {
	if index >= len(elements) {
		return nil, l.notYet(where, "a name past its tuple's elements")
	}
	of, isKnown := l.representation(elements[index])
	if !isKnown || slotless(of) {
		return nil, l.notYet(where, "a tuple element of type "+l.checker.TypeToString(elements[index]))
	}
	value := fit(ir.Expression(ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: strconv.Itoa(index), Of: of}), to)
	if value.Type() != to {
		return nil, l.notYet(where, "a tuple element of type "+l.checker.TypeToString(elements[index])+" given to a "+typeName(to))
	}
	return value, nil
}

// destructuringAssignment lowers [a, b] = tuple: the tuple is evaluated whole and held, then each
// name is assigned its field, in order, as JavaScript assigns them. A name left out ([, b]) is
// skipped. Anything but plain names, or a value that isn't a tuple, is not lowered yet.
func (l *lowering) destructuringAssignment(pattern *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	if !checker.IsTupleType(l.checker.GetTypeAtLocation(valueNode)) {
		return nil, l.notYet(pattern, "destructuring other than a tuple")
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(pattern, "destructuring a "+typeName(value.Type()))
	}
	elements := l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(valueNode))
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "tuple", Type: ir.Object, Function: l.functionIndex})
	statements := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for index, element := range pattern.AsArrayLiteralExpression().Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			continue
		}
		local, isLocal := l.local(element)
		if !ast.IsIdentifier(element) || !isLocal || l.alwaysUndefined[l.symbol(element)] {
			return nil, l.notYet(element, "assigning to "+describe(element)+" in a destructuring assignment")
		}
		field, err := l.tupleField(element, held, elements, index, l.result.Locals[local].Type)
		if err != nil {
			return nil, err
		}
		statements = append(statements, ir.Assign{Local: local, Value: field, Checked: l.checked(local)})
	}
	// The held tuple is the block's, released when it ends.
	return []ir.Statement{ir.Block{Body: statements}}, nil
}

// checked reports whether touching a local must be checked against the temporal dead zone: a
// global, from inside a function, which may run before the global's declaration has.
func (l *lowering) checked(local int) bool {
	return l.function != nil && l.result.Locals[local].Global
}

// returnStatement lowers return, with or without a value.
func (l *lowering) returnStatement(node *ast.Node) ([]ir.Statement, error) {
	if l.function == nil {
		return nil, errors.New("lower: " + l.program.Where(node) + ": return outside a function, and the checker let it through")
	}
	expression := node.AsReturnStatement().Expression
	if expression == nil {
		return []ir.Statement{ir.Return{}}, nil
	}
	if l.isPanicCall(expression) {
		// return panic('why'): panic never returns, so there is nothing to return, and it is the panic.
		return l.expressionStatement(expression)
	}
	value, err := l.expression(expression)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Return{Value: fit(value, l.function.Returns)}}, nil
}

// isPanicCall reports whether an expression is a call to the prelude's panic, which never returns.
func (l *lowering) isPanicCall(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	return expression.Kind == ast.KindCallExpression && l.isPreludeFunction(expression.AsCallExpression().Expression, "panic")
}

var compoundAssignments = map[ast.Kind]ast.Kind{
	ast.KindPlusEqualsToken:             ast.KindPlusToken,
	ast.KindMinusEqualsToken:            ast.KindMinusToken,
	ast.KindAsteriskEqualsToken:         ast.KindAsteriskToken,
	ast.KindSlashEqualsToken:            ast.KindSlashToken,
	ast.KindPercentEqualsToken:          ast.KindPercentToken,
	ast.KindAsteriskAsteriskEqualsToken: ast.KindAsteriskAsteriskToken,

	ast.KindAmpersandEqualsToken:                         ast.KindAmpersandToken,
	ast.KindBarEqualsToken:                               ast.KindBarToken,
	ast.KindCaretEqualsToken:                             ast.KindCaretToken,
	ast.KindLessThanLessThanEqualsToken:                  ast.KindLessThanLessThanToken,
	ast.KindGreaterThanGreaterThanEqualsToken:            ast.KindGreaterThanGreaterThanToken,
	ast.KindGreaterThanGreaterThanGreaterThanEqualsToken: ast.KindGreaterThanGreaterThanGreaterThanToken,
}

// increment lowers ++ and -- on a number local or field, as a statement, where prefix and postfix
// agree.
func (l *lowering) increment(node *ast.Node) ([]ir.Statement, error) {
	var operator ast.Kind
	var operand *ast.Node
	if node.Kind == ast.KindPrefixUnaryExpression {
		operator, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
	} else {
		operator, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
	}
	if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	operand = ast.SkipParentheses(operand)
	if operand.Kind == ast.KindPropertyAccessExpression {
		step := ast.KindPlusToken
		if operator == ast.KindMinusMinusToken {
			step = ast.KindMinusToken
		}
		return l.updateProperty(node, operand, step, nil)
	}
	local, isLocal := l.local(operand)
	if !ast.IsIdentifier(operand) || !isLocal {
		return nil, l.notYet(operand, "incrementing "+describe(operand))
	}
	step := ir.Add
	if operator == ast.KindMinusMinusToken {
		step = ir.Subtract
	}
	current := ir.Read{Local: local, Of: ir.Number, Checked: l.checked(local)}
	return []ir.Statement{ir.Assign{Local: local, Value: ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}}, Checked: l.checked(local)}}, nil
}

func (l *lowering) ifStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsIfStatement()
	condition, err := l.condition(statement.Expression)
	if err != nil {
		return nil, err
	}
	then, err := l.statement(statement.ThenStatement)
	if err != nil {
		return nil, err
	}
	lowered := ir.If{Condition: condition, Then: then}
	if statement.ElseStatement != nil {
		if lowered.Else, err = l.statement(statement.ElseStatement); err != nil {
			return nil, err
		}
	}
	return []ir.Statement{lowered}, nil
}

// forStatement lowers for (initializer; condition; incrementor) to a loop inside a block, so the
// loop's own declarations are scoped to it.
func (l *lowering) forStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsForStatement()
	block := ir.Block{}
	if initializer := statement.Initializer; initializer != nil {
		var err error
		var initial []ir.Statement
		if initializer.Kind == ast.KindVariableDeclarationList {
			initial, err = l.variables(initializer)
		} else {
			initial, err = l.expressionStatement(initializer)
		}
		if err != nil {
			return nil, err
		}
		block.Body = append(block.Body, initial...)
	}
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}}
	for _, statement := range block.Body {
		if declared, isDeclare := statement.(ir.Declare); isDeclare && initializerIsLet(statement, node) {
			loop.PerIteration = append(loop.PerIteration, declared.Local)
		}
	}
	if statement.Condition != nil {
		condition, err := l.condition(statement.Condition)
		if err != nil {
			return nil, err
		}
		loop.Condition = condition
	}
	if statement.Incrementor != nil {
		update, err := l.expressionStatement(statement.Incrementor)
		if err != nil {
			return nil, err
		}
		loop.Update = update
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, err
	}
	loop.Body = body
	block.Body = append(block.Body, loop)
	return []ir.Statement{block}, nil
}

// whileStatement lowers while and do...while.
func (l *lowering) whileStatement(node *ast.Node) ([]ir.Statement, error) {
	var expression, body *ast.Node
	checkAfter := node.Kind == ast.KindDoStatement
	if checkAfter {
		expression, body = node.AsDoStatement().Expression, node.AsDoStatement().Statement
	} else {
		expression, body = node.AsWhileStatement().Expression, node.AsWhileStatement().Statement
	}
	condition, err := l.condition(expression)
	if err != nil {
		return nil, err
	}
	loweredBody, err := l.statement(body)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Loop{Condition: condition, Body: loweredBody, CheckAfter: checkAfter}}, nil
}

// condition lowers an expression that decides a branch or a loop. 0.1 requires a boolean there
// (docs/0.1.md), so `if (name)` on a string is refused rather than given JavaScript's truthiness.
func (l *lowering) condition(node *ast.Node) (ir.Expression, error) {
	condition, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	if condition.Type() != ir.Boolean {
		return nil, &Refused{Where: l.program.Where(node), What: "a " + typeName(condition.Type()) + " as a condition", Fix: "compare it explicitly, like name.length > 0 or count !== 0"}
	}
	return condition, nil
}

// isPreludeFunction reports whether a callee is the prelude's function of that name, as imported
// from 'adamic', and not one of the program's that shares the name.
func (l *lowering) isPreludeFunction(callee *ast.Node, name string) bool {
	callee = ast.SkipParentheses(callee)
	if !ast.IsIdentifier(callee) {
		return false
	}
	symbol := l.symbol(callee)
	return symbol != nil && symbol.Name == name && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// isConsole reports whether a callee is a method of the prelude's console.
func (l *lowering) isConsole(callee *ast.Node) bool {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(callee.AsPropertyAccessExpression().Expression)
	return symbol != nil && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// consoleStream is the stream a callee writes to, when it is the prelude's console.log or
// console.error. A local named console is someone else's, and isn't lowered as this one.
func (l *lowering) consoleStream(callee *ast.Node) (ir.Stream, error) {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return 0, l.notYet(callee, "a call to "+describe(callee))
	}
	object := callee.AsPropertyAccessExpression().Expression
	symbol := l.checker.GetSymbolAtLocation(object)
	if symbol == nil || len(symbol.Declarations) == 0 || !load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return 0, l.notYet(callee, "a method call")
	}
	switch callee.Name().Text() {
	case "log":
		return ir.Stdout, nil
	case "error":
		return ir.Stderr, nil
	}
	return 0, l.notYet(callee, "console."+callee.Name().Text())
}

func (l *lowering) constant(value string) int {
	if l.strings == nil {
		l.strings = map[string]int{}
	}
	if index, isKnown := l.strings[value]; isKnown {
		return index
	}
	l.strings[value] = len(l.result.Strings)
	l.result.Strings = append(l.result.Strings, value)
	return l.strings[value]
}

// describe names a node's kind for a person: "a VariableStatement", "an ImportDeclaration".
func describe(node *ast.Node) string {
	name := strings.TrimPrefix(node.Kind.String(), "Kind")
	if strings.ContainsAny(name[:1], "AEIOU") {
		return "an " + name
	}
	return "a " + name
}

// initializerIsLet reports whether a for loop declares its variables with let, which JavaScript gives
// a fresh binding each iteration (const can't change, so a copy would be the same).
func initializerIsLet(_ ir.Statement, node *ast.Node) bool {
	initializer := node.AsForStatement().Initializer
	return initializer != nil && initializer.Kind == ast.KindVariableDeclarationList && initializer.Flags&ast.NodeFlagsLet != 0
}
