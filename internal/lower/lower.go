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
	classes   map[*ast.Symbol]*ast.Node
	instances map[string]*instance

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
// allows, and then lowers each function's body.
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
			if l.functions == nil {
				l.functions = map[*ast.Symbol]int{}
			}
			l.functions[symbol] = len(l.result.Functions)
			l.result.Functions = append(l.result.Functions, ir.Function{Name: statement.Name().Text()})
			declarations = append(declarations, statement)
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

// lowerFunction lowers a function-like declaration into the function at index. this, when not -1,
// is the local a method's or constructor's this is, passed first.
//
// It works on a copy and writes it back by index: lowering a body can instantiate a class, which
// appends functions, and a pointer into the slice would be left pointing at the old one.
func (l *lowering) lowerFunction(index int, declaration *ast.Node, this int) error {
	function := l.result.Functions[index]
	if this >= 0 && declaration.Kind != ast.KindConstructor {
		// A method receives this; a constructor makes it.
		function.Parameters = append(function.Parameters, this)
	}
	if declaration.Kind != ast.KindConstructor {
		signature := l.checker.GetSignatureFromDeclaration(declaration)
		returns := l.checker.GetReturnTypeOfSignature(signature)
		if returns.Flags()&checker.TypeFlagsVoid == 0 {
			valueType, isKnown := l.representation(returns)
			if !isKnown {
				return l.notYet(declaration.Name(), "a function returning "+l.checker.TypeToString(returns))
			}
			function.Returns = valueType
		}
	}
	outerIndexForParameters := l.functionIndex
	l.functionIndex = index
	defer func() { l.functionIndex = outerIndexForParameters }()
	// A parameter with a default arrives as what may be missing, and the body begins by declaring the
	// parameter itself: the argument, or the default when it's undefined, as JavaScript decides.
	type defaulted struct {
		local, incoming int
		initializer     *ast.Node
	}
	defaults := []defaulted{}
	for _, parameter := range declaration.Parameters() {
		declared := parameter.AsParameterDeclaration()
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
	body := declaration.Body()
	if body == nil {
		return l.notYet(declaration, "a function without a body")
	}
	// The signature is written back before the body is lowered, so a recursive call inside it knows
	// what the function returns.
	l.result.Functions[index] = function
	outer, outerThis, outerIndex := l.function, l.this, l.functionIndex
	l.function, l.functionIndex = &function, index
	if this >= 0 {
		l.this = this
	}
	// Defaults run before the body, in order, after a constructor's fields, as JavaScript runs them.
	var prologue, lowered []ir.Statement
	var err error
	for _, parameter := range defaults {
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
	} else if function.Returns == 0 {
		// An arrow function's expression body, when it returns nothing, is a statement.
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
	}
	return nil, l.notYet(node, describe(node))
}

// expressionStatement lowers an expression used as a statement: a console call, an assignment, or
// ++ and --. Any other expression's value would be thrown away, and stage 0 doesn't lower that yet.
func (l *lowering) expressionStatement(expression *ast.Node) ([]ir.Statement, error) {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindCallExpression:
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
		if !ast.IsIdentifier(name) {
			return nil, l.notYet(name, "a destructuring declaration")
		}
		local, err := l.declareLocal(name)
		if err != nil {
			return nil, err
		}
		var value ir.Expression
		if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
			if value, err = l.expression(initializer); err != nil {
				return nil, err
			}
		}
		statements = append(statements, ir.Declare{Local: local, Value: fit(value, l.result.Locals[local].Type)})
	}
	return statements, nil
}

// declareLocal makes a new local for a declared name, typed by what the checker proved for it.
func (l *lowering) declareLocal(name *ast.Node) (int, error) {
	symbol := l.symbol(name)
	if symbol == nil {
		return 0, errors.New("lower: " + l.program.Where(name) + ": the checker gave a declaration no symbol")
	}
	valueType := ir.Object
	if !l.alwaysUndefined[symbol] {
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
	return l.locals[symbol], nil
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
	local, isLocal := l.local(target)
	if !ast.IsIdentifier(target) || !isLocal {
		return nil, l.notYet(target, "assigning to "+describe(target))
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
	value, err := l.expression(expression)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Return{Value: fit(value, l.function.Returns)}}, nil
}

var compoundAssignments = map[ast.Kind]ast.Kind{
	ast.KindPlusEqualsToken:             ast.KindPlusToken,
	ast.KindMinusEqualsToken:            ast.KindMinusToken,
	ast.KindAsteriskEqualsToken:         ast.KindAsteriskToken,
	ast.KindSlashEqualsToken:            ast.KindSlashToken,
	ast.KindPercentEqualsToken:          ast.KindPercentToken,
	ast.KindAsteriskAsteriskEqualsToken: ast.KindAsteriskAsteriskToken,
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
