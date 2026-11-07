package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// censusRestParameter admits rest only where every call passes a freshly packed array.
// Function values and methods have a different calling convention and remain NotYet.
func (lowering *lowering) censusRestParameter(declaration, parameter *ast.Node) error {
	if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.TypeParameters()) != 0 {
		return lowering.notYet(parameter, "a rest parameter outside a nongeneric named function")
	}
	if lowering.result.Functions[lowering.functionIndex].Closure {
		return lowering.notYet(parameter, "a rest parameter in a named closure")
	}
	parameters := declaration.Parameters()
	if parameters[len(parameters)-1] != parameter {
		return lowering.notYet(parameter, "a rest parameter before another parameter")
	}
	proven := lowering.checker.GetTypeAtLocation(parameter.Name())
	if !lowering.isArrayType(proven) {
		return lowering.notYet(parameter, "a rest parameter other than an array")
	}
	element, known := lowering.kept(lowering.elementTypeOfArray(proven))
	if !known || slotless(element) {
		return lowering.notYet(parameter, "a rest array of "+lowering.checker.TypeToString(proven))
	}
	return nil
}

// censusRestDeclaration recovers the declaration without adding state to lower.go.
func (lowering *lowering) censusRestDeclaration(function int) *ast.Node {
	for symbol, index := range lowering.functions {
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
func (lowering *lowering) censusRestCall(
	call *ast.CallExpression,
	function int,
	declaration *ast.Node,
) (ir.Expression, error) {
	parameters := declaration.Parameters()
	fixed := len(parameters) - 1
	proven := lowering.checker.GetTypeAtLocation(parameters[fixed].Name())
	element, _ := lowering.kept(lowering.elementTypeOfArray(proven))
	rest := ir.ArrayLiteral{Element: element}
	arguments := []ir.Expression{}
	for index, argument := range call.Arguments.Nodes {
		spread := argument.Kind == ast.KindSpreadElement
		valueNode := argument
		if spread {
			if index < fixed {
				return nil, lowering.notYet(argument, "a spread filling fixed parameters before a rest parameter")
			}
			valueNode = argument.AsSpreadElement().Expression
		}
		value, expressionError := lowering.expression(valueNode)
		if expressionError != nil {
			return nil, expressionError
		}
		if index < fixed {
			arguments = append(arguments, value)
			continue
		}
		if spread {
			source := lowering.checker.GetTypeAtLocation(valueNode)
			if !lowering.isArrayType(source) {
				return nil, lowering.notYet(argument, "a rest spread other than an array")
			}
			held, known := lowering.kept(lowering.elementTypeOfArray(source))
			if !known || held != element {
				return nil, lowering.notYet(argument, "a rest spread with differently held elements")
			}
		} else {
			value = fit(value, element)
			if value.Type() != element {
				return nil, lowering.notYet(argument, "a rest argument with another representation")
			}
		}
		rest.Elements = append(rest.Elements, value)
		rest.Spread = append(rest.Spread, spread)
	}
	// Missing fixed arguments must keep their positions before the rest array.
	for len(arguments) < fixed {
		of := lowering.result.Locals[lowering.result.Functions[function].Parameters[len(arguments)]].Type
		if !of.IsMaybe() && !of.IsReference() {
			return nil, lowering.notYet(parameters[len(arguments)], "a missing fixed argument before rest")
		}
		arguments = append(arguments, fit(ir.Undefined{}, of))
	}
	arguments = append(arguments, rest)
	return lowering.censusOverloadResult(
		call,
		ir.Call{Function: function, Arguments: arguments, Returns: lowering.result.Functions[function].Returns},
	)
}

// Boolean fields have a tagged byte; other slotless representations remain refused.
func censusFieldSlotless(of ir.Type) bool {
	return slotless(of) && of != ir.MaybeBoolean && of != ir.Union
}

// Only the implementation of a named overload set has executable code.
func (lowering *lowering) censusImplementation(declaration *ast.Node) *ast.Node {
	if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil {
		return nil
	}
	symbol := lowering.symbol(declaration.Name())
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
func (lowering *lowering) censusRelated(from, to *checker.Type) bool {
	from, to = lowering.concrete(from), lowering.concrete(to)
	return lowering.classAssignable(from, to) && lowering.widened(from, to, map[[2]*checker.Type]bool{}) == nil
}

func (lowering *lowering) censusOverloads(implementation *ast.Node) error {
	if implementation.Kind != ast.KindFunctionDeclaration || implementation.Body() == nil {
		return nil
	}
	symbol := lowering.symbol(implementation.Name())
	if symbol == nil {
		return nil
	}
	ordinal := 0
	for _, overload := range symbol.Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		ordinal++
		if censusOverloadError := lowering.censusOverload(
			implementation,
			overload,
			ordinal,
		); censusOverloadError != nil {
			return censusOverloadError
		}
	}
	return nil
}

// Compare generic signatures under the same rigid type parameters, never under one
// lucky runtime instantiation. Mapping the implementation's binders is alpha-renaming.
func (lowering *lowering) censusOverload(implementation, overload *ast.Node, ordinal int) error {
	label := fmt.Sprintf("overload %d of %s", ordinal, implementation.Name().Text())
	declaredTypes, servedTypes := overload.TypeParameters(), implementation.TypeParameters()
	outerMapper := lowering.typeMapper
	defer func() { lowering.typeMapper = outerMapper }()
	if len(servedTypes) > 0 {
		sources, targets := []*checker.Type{}, []*checker.Type{}
		inferred := map[*checker.Type]*checker.Type{}
		if len(servedTypes) > len(declaredTypes) {
			// Extra binders need a witness from the entire signature, not positional
			// alpha-renaming: implementations may insert a binder before an old one.
			for index, parameter := range implementation.Parameters() {
				if index < len(overload.Parameters()) {
					lowering.inferTypes(
						lowering.checker.GetTypeAtLocation(parameter),
						lowering.checker.GetTypeAtLocation(overload.Parameters()[index]),
						inferred,
					)
				}
			}
			lowering.inferTypes(
				lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(implementation)),
				lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(overload)),
				inferred,
			)
		}
		for index, parameter := range servedTypes {
			source := lowering.checker.GetTypeAtLocation(parameter.Name())
			target := inferred[source]
			if len(servedTypes) <= len(declaredTypes) {
				target = lowering.checker.GetTypeAtLocation(declaredTypes[index].Name())
			} else if target == nil {
				// A binder with no evidence can use bottom as a candidate witness.
				// This is not permission to erase it: every constraint, incoming
				// parameter and outgoing result below must still hold for that witness.
				target = lowering.checker.GetNonNullableType(lowering.checker.GetUndefinedType())
			}
			sources = append(sources, source)
			targets = append(targets, target)
		}
		lowering.typeMapper = newTypeMapper(sources, targets)
		for index, parameter := range servedTypes {
			if constraint := lowering.checker.GetBaseConstraintOfType(
				sources[index],
			); constraint != nil && !lowering.censusRelated(targets[index], constraint) {
				where, name := overload, "type "+lowering.checker.TypeToString(targets[index])
				if len(servedTypes) <= len(declaredTypes) {
					where, name = declaredTypes[index], "type parameter "+declaredTypes[index].Name().Text()
				}
				return &Refused{
					Where: lowering.program.Where(where),
					What:  label + " " + name + " cannot satisfy implementation parameter " + parameter.Name().Text(),
					Fix:   "make implementation constraints accept every type admitted by the overload",
				}
			}
		}
	}
	declared, served := overload.Parameters(), implementation.Parameters()
	for index := 0; index <= max(len(declared), len(served)); index++ {
		given, parameter, givenRest, censusOverloadParameterError := lowering.censusOverloadParameter(declared, index)
		if censusOverloadParameterError != nil {
			return lowering.notYet(overload, label+" with a non-array rest parameter")
		}
		takes, actual, takesRest, censusOverloadParameterError := lowering.censusOverloadParameter(served, index)
		if censusOverloadParameterError != nil {
			return lowering.notYet(overload, label+" with a non-array implementation rest parameter")
		}
		if actual == nil {
			continue
		} // JavaScript evaluates and ignores extra arguments.
		if parameter == nil && takesRest {
			continue
		} // Missing rest items make an empty array.
		if parameter == nil {
			given = lowering.checker.GetUndefinedType()
			parameter = actual
		}
		if givenRest && !takesRest {
			given = lowering.checker.GetUnionType([]*checker.Type{given, lowering.checker.GetUndefinedType()})
		}
		if !lowering.censusRelated(given, takes) {
			name := func(node *ast.Node) string {
				if ast.IsIdentifier(node.Name()) {
					return node.Name().Text()
				}
				return fmt.Sprintf("%d", index+1)
			}
			return &Refused{
				Where: lowering.program.Where(parameter),
				What: label +
					" parameter " +
					name(parameter) +
					" cannot be served by implementation parameter " +
					name(actual),
				Fix: "make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
			}
		}
	}
	promised := lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(overload))
	produced := lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(implementation))
	if !lowering.censusRelated(produced, promised) {
		from, to := lowering.concrete(produced), lowering.concrete(promised)
		if (lowering.hasPhantom(from) || lowering.hasPhantom(to)) &&
			lowering.phantomCast(from, to) &&
			lowering.phantomAssignable(from, to) &&
			lowering.phantomAssignable(to, from) {
			return nil
		}
		if lowering.censusNullableOverloadResult(produced, promised) {
			return nil // Each resolved call proves the result or checks its presence.
		}
		proven, failed := lowering.predicateOverloadResultProven(implementation, overload)
		if proven {
			// Every call's callback is proven where it's passed: an overload's predicate parameter whose
			// implementation takes a plain boolean callback isn't a checked claim (checkedFunctionTypeClaim), so
			// predicateArguments demands a named function or arrow whose body proves both branches.
			return nil
		}
		what := label +
			" result " +
			lowering.checker.TypeToString(promised) +
			" cannot be served by implementation result " +
			lowering.checker.TypeToString(produced)
		if failed != nil {
			what += " (return at " + lowering.program.Where(failed) + " isn't the value its predicate checked)"
		}
		return &Refused{
			Where: lowering.program.Where(overload),
			What:  what,
			Fix:   "make the implementation result covariant with every overload result",
		}
	}
	return nil
}

func (lowering *lowering) censusOverloadParameter(
	parameters []*ast.Node,
	index int,
) (*checker.Type, *ast.Node, bool, error) {
	if len(parameters) == 0 {
		return nil, nil, false, nil
	}
	parameter := parameters[min(index, len(parameters)-1)]
	rest := parameter.AsParameterDeclaration().DotDotDotToken != nil
	if index >= len(parameters) && !rest {
		return nil, nil, false, nil
	}
	proven := lowering.checker.GetTypeAtLocation(parameter)
	if ast.IsIdentifier(parameter.Name()) {
		proven = lowering.censusCallableParameterType(lowering.symbol(parameter.Name()))
	}
	proven = lowering.concrete(proven)
	if rest {
		if !lowering.isArrayType(proven) {
			return nil, parameter, true, lowering.notYet(parameter, "an overload rest other than an array")
		}
		proven = lowering.elementTypeOfArray(proven)
	}
	return proven, parameter, rest, nil
}

// Calls use the resolved overload's result representation, even when the
// implementation proves a narrower result. A plain false must become a present
// boolean | undefined, and a scalar result must be boxed when the caller sees a union.
func (lowering *lowering) censusOverloadResult(call *ast.CallExpression, value ir.Expression) (ir.Expression, error) {
	symbol := lowering.symbol(ast.SkipParentheses(call.Expression))
	overloaded := false
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration &&
				declaration.Body() == nil &&
				lowering.censusImplementation(declaration) != nil {
				overloaded = true
				break
			}
		}
	}
	if !overloaded {
		return lowering.checkedPredicateResult(call, value, nil, nil)
	}
	resolved := lowering.checker.GetResolvedSignature(call.AsNode())
	if resolved != nil && resolved.Declaration() != nil {
		overload := resolved.Declaration()
		implementation := lowering.censusImplementation(overload)
		if implementation != nil && overload.Body() == nil {
			if overload.Type() != nil && overload.Type().Kind == ast.KindTypePredicate {
				return lowering.checkedPredicateResult(call, value, implementation, overload)
			}
			ordinal := 0
			for _, declaration := range symbol.Declarations {
				if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
					ordinal++
					if declaration == overload {
						break
					}
				}
			}
			if !lowering.censusProveOverloadResult(implementation, overload) {
				promised := lowering.checker.GetReturnTypeOfSignature(resolved)
				produced := lowering.checker.GetReturnTypeOfSignature(
					lowering.checker.GetSignatureFromDeclaration(implementation),
				)
				if lowering.censusHasUndefined(produced) && !lowering.censusHasUndefined(promised) {
					of, typeOfError := lowering.typeOf(call.AsNode())
					if typeOfError != nil {
						return nil, typeOfError
					}
					message := fmt.Sprintf(
						"overload %d of %s result: expected %s, got undefined",
						ordinal,
						implementation.Name().Text(),
						lowering.checker.TypeToString(promised),
					)
					return ir.Coalesce{
						Value: value,
						Of:    of,
						Panic: ir.StringConstant{Index: lowering.constant(message)},
					}, nil
				}
			}
		}
	}
	if value.Type() == 0 {
		return value, nil
	}
	of, typeOfError := lowering.typeOf(call.AsNode())
	if typeOfError != nil {
		return nil, typeOfError
	}
	value = fit(value, of)
	if value.Type() != of {
		return nil, lowering.notYet(call.AsNode(), "an overload result requiring another representation")
	}
	return value, nil
}

// Callable slots use the same tagged boolean representation as fields.
func censusCallableSlotless(of ir.Type) bool {
	return slotless(of) && of != ir.MaybeBoolean && of != ir.Union
}

// A default's body-local type excludes undefined, but its incoming argument can be undefined.
func (lowering *lowering) censusCallableParameter(parameter *ast.Symbol) (ir.Type, bool) {
	of, known := lowering.representation(lowering.checker.GetTypeOfSymbol(parameter))
	for _, declaration := range parameter.Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer != nil {
			of = ir.Maybe(of)
		}
	}
	return of, known
}

// Signature symbols describe the body-local type of a defaulted parameter. The
// callable contract also accepts undefined, which selects that parameter's default.
func (lowering *lowering) censusCallableParameterType(parameter *ast.Symbol) *checker.Type {
	of := lowering.checker.GetTypeOfSymbol(parameter)
	for _, declaration := range parameter.Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer != nil {
			return lowering.checker.GetUnionType([]*checker.Type{of, lowering.checker.GetUndefinedType()})
		}
	}
	return of
}

// callableParameterTypeAt is a signature's parameter type at position, as a call there sees it: its own
// parameter's (censusCallableParameterType, which adds undefined for one with an initializer), or, at or
// past a rest parameter, the rest's element at that position (GetTypeAtPosition), never the rest's whole
// tuple or array (#6xzs6gj).
func (lowering *lowering) callableParameterTypeAt(signature *checker.Signature, position int) *checker.Type {
	parameters := signature.Parameters()
	if position < len(parameters) && !(signature.HasRestParameter() && position == len(parameters)-1) {
		return lowering.censusCallableParameterType(parameters[position])
	}
	return lowering.checker.GetTypeAtPosition(signature, position)
}

// Only exactly true is truthy in boolean | undefined. Equality evaluates the value once.
func censusBooleanCondition(value ir.Expression) ir.Expression {
	if value.Type() == ir.MaybeBoolean {
		return ir.Binary{Operator: ir.Equal, Left: value, Right: fit(ir.BooleanConstant{Value: true}, ir.MaybeBoolean)}
	}
	return value
}

func (lowering *lowering) censusBooleanLogical(
	node *ast.Node,
	operator ast.Kind,
	left, right ir.Expression,
) (ir.Expression, bool) {
	if operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken {
		return nil, false
	}
	boolean := func(of ir.Type) bool { return of == ir.Boolean || of == ir.MaybeBoolean }
	if !boolean(left.Type()) || !boolean(right.Type()) {
		return nil, false
	}
	of, typeOfError := lowering.typeOf(node)
	if typeOfError != nil || !boolean(of) {
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
func (lowering *lowering) censusNeverRestSignature(signature *checker.Signature) bool {
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
	parameter := lowering.concrete(lowering.checker.GetTypeOfSymbol(parameters[0]))
	if lowering.isArrayType(parameter) {
		parameter = lowering.elementTypeOfArray(parameter)
	}
	return parameter.Flags()&checker.TypeFlagsNever != 0
}
