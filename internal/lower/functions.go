// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

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
	for position, parameter := range declaration.Parameters() {
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
		if parameterProperty(parameter) {
			// The binder gives a parameter property both a field symbol and a lexical parameter
			// symbol. Reads and assignments in the body refer to the latter.
			parameters := l.checker.GetSignatureFromDeclaration(declaration).Parameters()
			l.locals[parameters[position]] = local
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
		if declaration.Kind == ast.KindConstructor {
			if err = l.parameterPropertyDefault(declaration, parameter.initializer); err != nil {
				break
			}
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
	if declaration.Kind == ast.KindConstructor && l.instance.base == nil {
		assigned, assignErr := l.parameterPropertyStores(declaration.Parent, this)
		if assignErr != nil {
			l.function, l.this, l.functionIndex = outer, outerThis, outerIndex
			return assignErr
		}
		prologue = append(prologue, assigned...)
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
	if body.Kind == ast.KindBlock && declaration.Kind != ast.KindConstructor && declaration.Flags&ast.NodeFlagsHasImplicitReturn != 0 && l.permitsImplicitReturn(declaration) {
		// The checker permits an end that is reachable only when the result can be undefined
		// (or void). Put that exit in the IR so every backend and ownership pass sees it.
		returned := ir.Return{}
		if function.Returns != 0 {
			returned.Value = fit(ir.Undefined{}, function.Returns)
		}
		lowered = append(lowered, returned)
	}
	// The environment may have grown while the body was lowered (captures are found as they're
	// read), so it's taken from what's recorded, not from this copy.
	function.Environment = l.result.Functions[index].Environment
	function.Body = append(function.Body, prologue...)
	function.Body = append(function.Body, lowered...)
	l.result.Functions[index] = function
	return nil
}

// The binder marks a syntactic end even for a switch the checker proved exhaustive. Only the
// checker's result type can authorize returning undefined at that end.
func (l *lowering) permitsImplicitReturn(declaration *ast.Node) bool {
	result := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(declaration))
	return result.Flags()&checker.TypeFlagsVoid != 0 || l.includesUndefined(result)
}
