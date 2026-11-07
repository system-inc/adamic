package lower

import (
	"fmt"
	_ "unsafe"

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
	if !l.isArrayType(proven) {
		return l.notYet(parameter, "a rest parameter other than an array")
	}
	element, known := l.kept(l.elementTypeOfArray(proven))
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
	element, _ := l.kept(l.elementTypeOfArray(proven))
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
			if !l.isArrayType(source) {
				return nil, l.notYet(argument, "a rest spread other than an array")
			}
			held, known := l.kept(l.elementTypeOfArray(source))
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

// Boolean fields have a tagged byte; other slotless representations remain refused.
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
// lucky runtime instantiation. The witness for the implementation's binders is only a
// guess: the constraint, parameter and result checks below are the proof, so any witness
// that passes them all proves the overload, and one that fails proves nothing. The first
// guess renames binders by position when the implementation has no more of them than the
// overload, and otherwise pairs parameters by index. When that is refused, a second guess
// reads the whole signature (censusInferWitness). It may only find a better witness, never
// weaken a check: if it fails too, the first refusal stands, in its own words (#yq1xjde,
// tsc's or, whose T is the arguments' tuple and whose P is the predicate's subject).
func (l *lowering) censusOverload(implementation, overload *ast.Node, ordinal int) error {
	err := l.censusOverloadWitness(implementation, overload, ordinal, false)
	if _, refused := err.(*Refused); !refused || len(implementation.TypeParameters()) == 0 {
		return err
	}
	if l.censusOverloadWitness(implementation, overload, ordinal, true) == nil {
		return nil
	}
	return err
}

// censusOverloadWitness checks one overload against the implementation under one witness for
// its binders: the first guess, or with read the second, censusInferWitness's.
func (l *lowering) censusOverloadWitness(implementation, overload *ast.Node, ordinal int, read bool) error {
	label := fmt.Sprintf("overload %d of %s", ordinal, implementation.Name().Text())
	declaredTypes, servedTypes := overload.TypeParameters(), implementation.TypeParameters()
	positional := !read && len(servedTypes) <= len(declaredTypes)
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	if len(servedTypes) > 0 {
		sources, targets := []*checker.Type{}, []*checker.Type{}
		inferred := map[*checker.Type]*checker.Type{}
		if read {
			inferred = l.censusInferWitness(implementation, overload)
		} else if !positional {
			// Extra binders need a witness from the entire signature, not positional
			// alpha-renaming: implementations may insert a binder before an old one.
			for index, parameter := range implementation.Parameters() {
				if index < len(overload.Parameters()) {
					l.inferTypes(l.checker.GetTypeAtLocation(parameter), l.checker.GetTypeAtLocation(overload.Parameters()[index]), inferred)
				}
			}
			l.inferTypes(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation)), l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload)), inferred)
		}
		for index, parameter := range servedTypes {
			source := l.checker.GetTypeAtLocation(parameter.Name())
			target := inferred[source]
			if positional {
				target = l.checker.GetTypeAtLocation(declaredTypes[index].Name())
			} else if target == nil {
				// A binder with no evidence can use bottom as a candidate witness.
				// This is not permission to erase it: every constraint, incoming
				// parameter and outgoing result below must still hold for that witness.
				target = l.checker.GetNonNullableType(l.checker.GetUndefinedType())
			}
			sources = append(sources, source)
			targets = append(targets, target)
		}
		l.typeMapper = newTypeMapper(sources, targets)
		for index, parameter := range servedTypes {
			if constraint := l.checker.GetBaseConstraintOfType(sources[index]); constraint != nil && !l.censusRelated(targets[index], constraint) {
				where, name := overload, "type "+l.checker.TypeToString(targets[index])
				if positional {
					where, name = declaredTypes[index], "type parameter "+declaredTypes[index].Name().Text()
				}
				return &Refused{Where: l.program.Where(where), What: label + " " + name + " cannot satisfy implementation parameter " + parameter.Name().Text(), Fix: "make implementation constraints accept every type admitted by the overload"}
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
		from, to := l.concrete(produced), l.concrete(promised)
		if (l.hasPhantom(from) || l.hasPhantom(to)) && l.phantomCast(from, to) && l.phantomAssignable(from, to) && l.phantomAssignable(to, from) {
			return nil
		}
		if l.censusNullableOverloadResult(produced, promised) {
			return nil // Each resolved call proves the result or checks its presence.
		}
		if l.censusCheckedClaimResult(overload, produced) {
			return nil // Every call through the promised signature checks its claim (R-T80).
		}
		proven, failed := l.predicateOverloadResultProven(implementation, overload)
		if proven {
			// Every call's callback is proven where it's passed: an overload's predicate parameter whose
			// implementation takes a plain boolean callback isn't a checked claim (checkedFunctionTypeClaim), so
			// predicateArguments demands a named function or arrow whose body proves both branches.
			return nil
		}
		what := label + " result " + l.checker.TypeToString(promised) + " cannot be served by implementation result " + l.checker.TypeToString(produced)
		if failed != nil {
			what += " (return at " + l.program.Where(failed) + " isn't the value its predicate checked)"
		}
		return &Refused{Where: l.program.Where(overload), What: what, Fix: "make the implementation result covariant with every overload result"}
	}
	return nil
}

// censusInferWitness reads what each of the implementation's binders stands for under an
// overload from the whole signature: each overload parameter against the implementation
// parameter at its position, or against a rest's element at and past the rest (the reading
// censusOverloadParameter does), and the overload's result against the implementation's.
// A function parameter whose own rest is typed by a binder, (...args: T) => U, takes the
// overload's parameters from there as the tuple the checker itself reads them as, so (p1: P)
// gives T = [P]. A binder with no evidence is left out; censusOverloadWitness gives it bottom.
func (l *lowering) censusInferWitness(implementation, overload *ast.Node) map[*checker.Type]*checker.Type {
	inferred := map[*checker.Type]*checker.Type{}
	binders := map[*checker.Type]bool{}
	for _, parameter := range implementation.TypeParameters() {
		binders[l.checker.GetTypeAtLocation(parameter.Name())] = true
	}
	served := implementation.Parameters()
	for index, parameter := range overload.Parameters() {
		if len(served) == 0 {
			break
		}
		takes := served[min(index, len(served)-1)]
		rest := takes.AsParameterDeclaration().DotDotDotToken != nil
		if index >= len(served) && !rest {
			break
		}
		taken := l.checker.GetTypeAtLocation(takes)
		if rest {
			if !l.isArrayType(taken) {
				continue
			}
			taken = l.elementTypeOfArray(taken)
		}
		l.censusInferParameter(taken, l.checker.GetTypeAtLocation(parameter), binders, inferred)
	}
	l.inferTypes(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation)), l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload)), inferred)
	return inferred
}

// censusRestTypeAtPosition is a signature's parameters from pos on, as the tuple a rest there
// would take, read by the pinned checker as its own inference reads it.
//
//go:linkname censusRestTypeAtPosition github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getRestTypeAtPosition
func censusRestTypeAtPosition(receiver *checker.Checker, source *checker.Signature, pos int, readonly bool) *checker.Type

// censusInferParameter infers from one overload parameter against the implementation's, as
// inferTypes does, except that a single call signature whose rest is typed by one of the
// implementation's binders binds that binder to the overload signature's parameters from the
// rest's position on. inferTypes pairs parameters by index, which would bind T to P.
func (l *lowering) censusInferParameter(taken, given *checker.Type, binders map[*checker.Type]bool, into map[*checker.Type]*checker.Type) {
	takenSignatures, givenSignatures := l.checker.GetSignaturesOfType(taken, checker.SignatureKindCall), l.checker.GetSignaturesOfType(given, checker.SignatureKindCall)
	if len(takenSignatures) == 1 && len(givenSignatures) == 1 && takenSignatures[0].HasRestParameter() {
		parameters := takenSignatures[0].Parameters()
		last := len(parameters) - 1
		if rest := l.checker.GetTypeOfSymbol(parameters[last]); binders[rest] {
			if _, known := into[rest]; !known {
				into[rest] = l.concrete(censusRestTypeAtPosition(l.checker, givenSignatures[0], last, false))
			}
			for index := 0; index < last; index++ {
				l.inferTypes(l.checker.GetTypeOfSymbol(parameters[index]), l.checker.GetTypeAtPosition(givenSignatures[0], index), into)
			}
			l.inferTypes(l.checker.GetReturnTypeOfSignature(takenSignatures[0]), l.checker.GetReturnTypeOfSignature(givenSignatures[0]), into)
			return
		}
	}
	l.inferTypes(taken, given, into)
}

// censusCheckedClaimResult is whether an implementation's result serves an overload whose result is
// a function claiming `p is R`, because the claim is checked at every call through it: in .ts, a claim
// checkedFunctionTypeClaim makes checked is tested at each call that resolves to the signature, against
// the target the call resolved to, so the implementation never has to prove it (Adamic's ruling on
// #1w5bve2, R-T80's own reason). Then the result serves the promise when it relates to the promised
// signature with its claim erased to boolean. Three things keep it narrow (Lower's design, accepted by
// Compiler): the claim is read off the overload's own annotation, a function type, so .a, where claims
// aren't checked, never sees one; an asserts claim isn't erased, since nothing checks one on a function
// type; and the promised type is that one call signature and nothing else, so an intersection with a
// property, a construct signature or a second call signature stays refused. The false side needs
// nothing: a call through the value isn't a direct predicate call, so the false-direction walk
// (predicate_false_flow.go) makes a read narrowed on false NotYet, as for every checked claim.
func (l *lowering) censusCheckedClaimResult(overload *ast.Node, produced *checker.Type) bool {
	annotation := overload.Type()
	if annotation == nil || annotation.Kind != ast.KindFunctionType {
		return false
	}
	claim := annotation.Type()
	if claim == nil || claim.Kind != ast.KindTypePredicate || claim.AsTypePredicateNode().AssertsModifier != nil || !l.checkedFunctionTypeClaim(claim) {
		return false
	}
	promised := l.concrete(l.checker.GetTypeFromTypeNode(annotation))
	promisedSignatures := l.checker.GetSignaturesOfType(promised, checker.SignatureKindCall)
	if len(promisedSignatures) != 1 || len(promisedSignatures[0].TypeParameters()) != 0 || promisedSignatures[0].HasRestParameter() ||
		len(l.checker.GetSignaturesOfType(promised, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetPropertiesOfType(promised)) != 0 || len(l.checker.GetIndexInfosOfType(promised)) != 0 {
		return false
	}
	producedSignatures := l.checker.GetSignaturesOfType(l.concrete(produced), checker.SignatureKindCall)
	if len(producedSignatures) != 1 || len(producedSignatures[0].TypeParameters()) != 0 {
		return false
	}
	given, takes := promisedSignatures[0], producedSignatures[0]
	for index := 0; index < max(len(given.Parameters()), len(takes.Parameters())); index++ {
		if index >= len(takes.Parameters()) && !takes.HasRestParameter() {
			break // JavaScript evaluates and ignores extra arguments.
		}
		argument := l.checker.GetUndefinedType()
		if index < len(given.Parameters()) {
			argument = l.callableParameterTypeAt(given, index)
		}
		if !l.censusRelated(argument, l.callableParameterTypeAt(takes, index)) {
			return false
		}
	}
	return l.censusRelated(l.checker.GetReturnTypeOfSignature(takes), l.checker.GetBooleanType())
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
		if !l.isArrayType(proven) {
			return nil, parameter, true, l.notYet(parameter, "an overload rest other than an array")
		}
		proven = l.elementTypeOfArray(proven)
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
	if !overloaded {
		return l.checkedPredicateResult(call, value, nil, nil)
	}
	resolved := l.checker.GetResolvedSignature(call.AsNode())
	if resolved != nil && resolved.Declaration() != nil {
		overload := resolved.Declaration()
		implementation := l.censusImplementation(overload)
		if implementation != nil && overload.Body() == nil {
			if overload.Type() != nil && overload.Type().Kind == ast.KindTypePredicate {
				return l.checkedPredicateResult(call, value, implementation, overload)
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
	if value.Type() == 0 {
		return value, nil
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
func censusCallableSlotless(of ir.Type) bool {
	return slotless(of) && of != ir.MaybeBoolean && of != ir.Union
}

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

// callableParameterTypeAt is a signature's parameter type at position, as a call there sees it: its own
// parameter's (censusCallableParameterType, which adds undefined for one with an initializer), or, at or
// past a rest parameter, the rest's element at that position (GetTypeAtPosition), never the rest's whole
// tuple or array (#6xzs6gj).
func (l *lowering) callableParameterTypeAt(signature *checker.Signature, position int) *checker.Type {
	parameters := signature.Parameters()
	if position < len(parameters) && !(signature.HasRestParameter() && position == len(parameters)-1) {
		return l.censusCallableParameterType(parameters[position])
	}
	return l.checker.GetTypeAtPosition(signature, position)
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
	if l.isArrayType(parameter) {
		parameter = l.elementTypeOfArray(parameter)
	}
	return parameter.Flags()&checker.TypeFlagsNever != 0
}
