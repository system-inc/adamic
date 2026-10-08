package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// censusRestParameter admits rest only where every call passes a freshly packed array.
// Function values and methods have a different calling convention and remain NotYet.
func (l *lowering) censusRestParameter(declaration, parameter *ast.Node) error {
	if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.TypeParameters()) != 0 {
		return l.notYet(parameter, "a rest parameter outside a nongeneric named function")
	}
	if l.result.Functions[l.functionIndex].Closure {
		return l.notYet(parameter, "a rest parameter in a named closure")
	}
	parameters := declaration.Parameters()
	if parameters[len(parameters)-1] != parameter {
		return l.notYet(parameter, "a rest parameter before another parameter")
	}
	proven := l.checker.GetTypeAtLocation(parameter.Name())
	if !l.checker.IsArrayType(proven) {
		return l.notYet(parameter, "a rest parameter other than an array")
	}
	element, known := l.kept(l.checker.GetElementTypeOfArrayType(proven))
	if !known || slotless(element) {
		return l.notYet(parameter, "a rest array of "+l.checker.TypeToString(proven))
	}
	return nil
}

// censusRestDeclaration recovers the declaration without adding state to lower.go.
func (l *lowering) censusRestDeclaration(function int) *ast.Node {
	for symbol, index := range l.functions {
		if index != function {
			continue
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil {
				continue
			}
			parameters := declaration.Parameters()
			if len(parameters) > 0 && parameters[len(parameters)-1].AsParameterDeclaration().DotDotDotToken != nil {
				return declaration
			}
		}
	}
	return nil
}

// censusRestCall evaluates fixed arguments before rest items, and copies each spread
// when encountered. The array is new even when the only item is a spread.
func (l *lowering) censusRestCall(call *ast.CallExpression, function int, declaration *ast.Node) (ir.Expression, error) {
	parameters := declaration.Parameters()
	fixed := len(parameters) - 1
	proven := l.checker.GetTypeAtLocation(parameters[fixed].Name())
	element, _ := l.kept(l.checker.GetElementTypeOfArrayType(proven))
	rest := ir.ArrayLiteral{Element: element}
	arguments := []ir.Expression{}
	for index, argument := range call.Arguments.Nodes {
		spread := argument.Kind == ast.KindSpreadElement
		valueNode := argument
		if spread {
			if index < fixed {
				return nil, l.notYet(argument, "a spread filling fixed parameters before a rest parameter")
			}
			valueNode = argument.AsSpreadElement().Expression
		}
		value, err := l.expression(valueNode)
		if err != nil {
			return nil, err
		}
		if index < fixed {
			arguments = append(arguments, value)
			continue
		}
		if spread {
			source := l.checker.GetTypeAtLocation(valueNode)
			if !l.checker.IsArrayType(source) {
				return nil, l.notYet(argument, "a rest spread other than an array")
			}
			held, known := l.kept(l.checker.GetElementTypeOfArrayType(source))
			if !known || held != element {
				return nil, l.notYet(argument, "a rest spread with differently held elements")
			}
		} else {
			value = fit(value, element)
			if value.Type() != element {
				return nil, l.notYet(argument, "a rest argument with another representation")
			}
		}
		rest.Elements = append(rest.Elements, value)
		rest.Spread = append(rest.Spread, spread)
	}
	// Missing fixed arguments must keep their positions before the rest array.
	for len(arguments) < fixed {
		of := l.result.Locals[l.result.Functions[function].Parameters[len(arguments)]].Type
		if !of.IsMaybe() && !of.IsReference() {
			return nil, l.notYet(parameters[len(arguments)], "a missing fixed argument before rest")
		}
		arguments = append(arguments, fit(ir.Undefined{}, of))
	}
	arguments = append(arguments, rest)
	return l.censusOverloadResult(call, ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns})
}

// Boolean fields have a tagged byte; differently held union fields keep a boxed reference.
func censusFieldSlotless(of ir.Type) bool {
	return slotless(of) && of != ir.MaybeBoolean && of != ir.Union
}

// Only the implementation of a named overload set has executable code.
func (l *lowering) censusImplementation(declaration *ast.Node) *ast.Node {
	if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil {
		return nil
	}
	symbol := l.symbol(declaration.Name())
	if symbol == nil {
		return nil
	}
	for _, candidate := range symbol.Declarations {
		if candidate.Kind == ast.KindFunctionDeclaration && candidate.Body() != nil {
			return candidate
		}
	}
	return nil
}

// Use the same nominal, invariant and strictly contravariant relation as class overrides.
func (l *lowering) censusRelated(from, to *checker.Type) bool {
	from, to = l.concrete(from), l.concrete(to)
	return l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil
}

func (l *lowering) censusOverloads(implementation *ast.Node) error {
	if implementation.Kind != ast.KindFunctionDeclaration || implementation.Body() == nil {
		return nil
	}
	symbol := l.symbol(implementation.Name())
	if symbol == nil {
		return nil
	}
	ordinal := 0
	for _, overload := range symbol.Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		ordinal++
		if err := l.censusOverload(implementation, overload, ordinal); err != nil {
			return err
		}
	}
	return nil
}

// Compare generic signatures under the same rigid type parameters, never under one
// lucky runtime instantiation. Mapping the implementation's binders is alpha-renaming.
func (l *lowering) censusOverload(implementation, overload *ast.Node, ordinal int) error {
	label := fmt.Sprintf("overload %d of %s", ordinal, implementation.Name().Text())
	declaredTypes, servedTypes := overload.TypeParameters(), implementation.TypeParameters()
	if len(servedTypes) > len(declaredTypes) {
		return l.notYet(overload, label+" with additional implementation type parameters")
	}
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	if len(servedTypes) > 0 {
		sources, targets := []*checker.Type{}, []*checker.Type{}
		for index, parameter := range servedTypes {
			sources = append(sources, l.checker.GetTypeAtLocation(parameter.Name()))
			targets = append(targets, l.checker.GetTypeAtLocation(declaredTypes[index].Name()))
		}
		l.typeMapper = newTypeMapper(sources, targets)
		for index, parameter := range servedTypes {
			if constraint := l.checker.GetBaseConstraintOfType(sources[index]); constraint != nil && !l.censusRelated(targets[index], constraint) {
				return &Refused{Where: l.program.Where(declaredTypes[index]), What: label + " type parameter " + declaredTypes[index].Name().Text() + " cannot satisfy implementation parameter " + parameter.Name().Text(), Fix: "make implementation constraints accept every type admitted by the overload"}
			}
		}
	}
	declared, served := overload.Parameters(), implementation.Parameters()
	for index := 0; index <= max(len(declared), len(served)); index++ {
		given, parameter, givenRest, err := l.censusOverloadParameter(declared, index)
		if err != nil {
			return l.notYet(overload, label+" with a non-array rest parameter")
		}
		takes, actual, takesRest, err := l.censusOverloadParameter(served, index)
		if err != nil {
			return l.notYet(overload, label+" with a non-array implementation rest parameter")
		}
		if actual == nil {
			continue
		} // JavaScript evaluates and ignores extra arguments.
		if parameter == nil && takesRest {
			continue
		} // Missing rest items make an empty array.
		if parameter == nil {
			given = l.checker.GetUndefinedType()
			parameter = actual
		}
		if givenRest && !takesRest {
			given = l.checker.GetUnionType([]*checker.Type{given, l.checker.GetUndefinedType()})
		}
		if !l.censusRelated(given, takes) {
			name := func(node *ast.Node) string {
				if ast.IsIdentifier(node.Name()) {
					return node.Name().Text()
				}
				return fmt.Sprintf("%d", index+1)
			}
			return &Refused{Where: l.program.Where(parameter), What: label + " parameter " + name(parameter) + " cannot be served by implementation parameter " + name(actual), Fix: "make the implementation accept every value admitted by this overload, without mutable widening or bivariance"}
		}
	}
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	produced := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation))
	if !l.censusRelated(produced, promised) {
		if l.censusNullableOverloadResult(produced, promised) {
			return nil // Each resolved call proves the result or checks its presence.
		}
		return &Refused{Where: l.program.Where(overload), What: label + " result " + l.checker.TypeToString(promised) + " cannot be served by implementation result " + l.checker.TypeToString(produced), Fix: "make the implementation result covariant with every overload result"}
	}
	return nil
}

func (l *lowering) censusOverloadParameter(parameters []*ast.Node, index int) (*checker.Type, *ast.Node, bool, error) {
	if len(parameters) == 0 {
		return nil, nil, false, nil
	}
	parameter := parameters[min(index, len(parameters)-1)]
	rest := parameter.AsParameterDeclaration().DotDotDotToken != nil
	if index >= len(parameters) && !rest {
		return nil, nil, false, nil
	}
	proven := l.checker.GetTypeAtLocation(parameter)
	if ast.IsIdentifier(parameter.Name()) {
		proven = l.censusCallableParameterType(l.symbol(parameter.Name()))
	}
	proven = l.concrete(proven)
	if rest {
		if !l.checker.IsArrayType(proven) {
			return nil, parameter, true, l.notYet(parameter, "an overload rest other than an array")
		}
		proven = l.checker.GetElementTypeOfArrayType(proven)
	}
	return proven, parameter, rest, nil
}

// Calls use the resolved overload's result representation, even when the
// implementation proves a narrower result. A plain false must become a present
// boolean | undefined, and a scalar result must be boxed when the caller sees a union.
func (l *lowering) censusOverloadResult(call *ast.CallExpression, value ir.Expression) (ir.Expression, error) {
	symbol := l.symbol(ast.SkipParentheses(call.Expression))
	overloaded := false
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil && l.censusImplementation(declaration) != nil {
				overloaded = true
				break
			}
		}
	}
	if !overloaded || value.Type() == 0 {
		return value, nil
	}
	resolved := l.checker.GetResolvedSignature(call.AsNode())
	if resolved != nil && resolved.Declaration() != nil {
		overload := resolved.Declaration()
		implementation := l.censusImplementation(overload)
		if implementation != nil && overload.Body() == nil {
			ordinal := 0
			for _, declaration := range symbol.Declarations {
				if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
					ordinal++
					if declaration == overload {
						break
					}
				}
			}
			if !l.censusProveOverloadResult(implementation, overload) {
				promised := l.checker.GetReturnTypeOfSignature(resolved)
				produced := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation))
				if l.censusHasUndefined(produced) && !l.censusHasUndefined(promised) {
					of, err := l.typeOf(call.AsNode())
					if err != nil {
						return nil, err
					}
					message := fmt.Sprintf("overload %d of %s result: expected %s, got undefined", ordinal, implementation.Name().Text(), l.checker.TypeToString(promised))
					return ir.Coalesce{Value: value, Of: of, Panic: ir.StringConstant{Index: l.constant(message)}}, nil
				}
			}
		}
	}
	of, err := l.typeOf(call.AsNode())
	if err != nil {
		return nil, err
	}
	value = fit(value, of)
	if value.Type() != of {
		return nil, l.notYet(call.AsNode(), "an overload result requiring another representation")
	}
	return value, nil
}

// Callable slots use the same tagged boolean representation as fields.
func censusCallableSlotless(of ir.Type) bool { return slotless(of) && of != ir.MaybeBoolean }

// A default's body-local type excludes undefined, but its incoming argument can be undefined.
func (l *lowering) censusCallableParameter(parameter *ast.Symbol) (ir.Type, bool) {
	of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
	for _, declaration := range parameter.Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer != nil {
			of = ir.Maybe(of)
		}
	}
	return of, known
}

// Signature symbols describe the body-local type of a defaulted parameter. The
// callable contract also accepts undefined, which selects that parameter's default.
func (l *lowering) censusCallableParameterType(parameter *ast.Symbol) *checker.Type {
	of := l.checker.GetTypeOfSymbol(parameter)
	for _, declaration := range parameter.Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer != nil {
			return l.checker.GetUnionType([]*checker.Type{of, l.checker.GetUndefinedType()})
		}
	}
	return of
}

// Only exactly true is truthy in boolean | undefined. Equality evaluates the value once.
func censusBooleanCondition(value ir.Expression) ir.Expression {
	if value.Type() == ir.MaybeBoolean {
		return ir.Binary{Operator: ir.Equal, Left: value, Right: fit(ir.BooleanConstant{Value: true}, ir.MaybeBoolean)}
	}
	return value
}

func (l *lowering) censusBooleanLogical(node *ast.Node, operator ast.Kind, left, right ir.Expression) (ir.Expression, bool) {
	if operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken {
		return nil, false
	}
	boolean := func(of ir.Type) bool { return of == ir.Boolean || of == ir.MaybeBoolean }
	if !boolean(left.Type()) || !boolean(right.Type()) {
		return nil, false
	}
	of, err := l.typeOf(node)
	if err != nil || !boolean(of) {
		// Two booleans give a boolean whatever the checker narrowed the whole to: a branch it
		// knows is dead types the expression never, and that branch must still lower.
		if left.Type() != ir.Boolean || right.Type() != ir.Boolean {
			return nil, false
		}
		of = ir.Boolean
	}
	lowered := ir.And
	if operator == ast.KindBarBarToken {
		lowered = ir.Or
	}
	if lowered == ir.And && left.Type() == ir.MaybeBoolean {
		of = ir.MaybeBoolean
	}
	if of == ir.MaybeBoolean {
		left, right = fit(left, of), fit(right, of)
	} else {
		left, right = censusBooleanCondition(left), censusBooleanCondition(right)
	}
	return ir.Binary{Operator: lowered, Left: left, Right: right}, true
}

// ToBoolean never calls user conversion methods. Every represented value can be tested.
// NumberCall already carries a single evaluated argument through both backends and analyses.
func censusCondition(value ir.Expression) ir.Expression {
	if value.Type() == ir.Boolean {
		return value
	}
	return ir.NumberCall{Function: "toBoolean", Arguments: []ir.Expression{value}}
}

// A never-rest signature is an erased callable marker. No ordinary argument can
// inhabit its rest element; storage compares results without pairing parameters.
func (l *lowering) censusNeverRestSignature(signature *checker.Signature) bool {
	parameters := signature.Parameters()
	if len(parameters) != 1 {
		return false
	}
	rest := false
	for _, declaration := range parameters[0].Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
			rest = true
		}
	}
	if !rest {
		return false
	}
	parameter := l.concrete(l.checker.GetTypeOfSymbol(parameters[0]))
	if l.checker.IsArrayType(parameter) {
		parameter = l.checker.GetElementTypeOfArrayType(parameter)
	}
	return parameter.Flags()&checker.TypeFlagsNever != 0
}
