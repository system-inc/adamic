package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// maximumGenericDepth is how many instantiations of generic functions may be lowered inside one
// another. Ordinary programs nest a few; a function that calls itself with an ever larger type
// (polymorphic recursion) would instantiate without end, and 0.1 refuses it (docs/0.1.md).
const maximumGenericDepth = 32

// instantiateFunction lowers a generic module function for the type arguments of one call, once per
// distinct set of checker type identities, as a generic class is (instantiate). Within it,
// each type parameter is what this call made it.
//
// The checker resolves the call's signature with its type arguments substituted, but doesn't export
// the mapping itself, so it's read back the way it was made: each declared parameter's type, and the
// result's, against the resolved signature's, through arrays and tuples. A type parameter that can't
// be read back that way is left unmapped, and the body says not yet wherever it needs to know it.
func (lowering *lowering) instantiateFunction(call *ast.Node, declaration *ast.Node) (int, error) {
	return lowering.instantiateFunctionSignature(call, declaration, lowering.checker.GetResolvedSignature(call))
}

func (lowering *lowering) instantiateFunctionSignature(
	call *ast.Node,
	declaration *ast.Node,
	resolved *checker.Signature,
) (int, error) {
	target := lowering.checker.GetSignatureFromDeclaration(declaration)
	if resolved == nil || target == nil {
		return 0, lowering.notYet(call, "a call to a generic function whose signature the checker didn't resolve")
	}
	concreteTypes := map[*checker.Type]*checker.Type{}
	// Class-generic calls supply this callee's resolved mapper. Ordinary recursive
	// calls must infer their own arguments: the outer mapper can still describe the
	// same declaration's previous instantiation, as nest<T>([item], depth - 1) does.
	if lowering.genericUsesClasses(declaration, map[*ast.Node]bool{}) {
		for _, parameter := range declaration.TypeParameters() {
			declaredType := lowering.checker.GetTypeAtLocation(parameter.Name())
			concrete := lowering.concrete(declaredType)
			if concrete != declaredType && concrete.Flags()&checker.TypeFlagsTypeParameter == 0 {
				concreteTypes[declaredType] = concrete
			}
		}
	}
	declared, given := target.Parameters(), resolved.Parameters()
	for index := range declared {
		if index < len(given) {
			lowering.inferTypes(
				lowering.checker.GetTypeOfSymbol(declared[index]),
				lowering.checker.GetTypeOfSymbol(given[index]),
				concreteTypes,
			)
		}
	}
	lowering.inferTypes(
		lowering.checker.GetReturnTypeOfSignature(target),
		lowering.checker.GetReturnTypeOfSignature(resolved),
		concreteTypes,
	)

	// An explicit type argument can make tsc view a mutable argument through a wider type without
	// giving the argument that contextual type. Judge the instantiated parameter directly: the
	// ordinary refusal walk cannot see this view at the argument node.
	if call.Kind == ast.KindCallExpression &&
		call.AsCallExpression().TypeArguments != nil &&
		len(call.AsCallExpression().TypeArguments.Nodes) > 0 {
		for index, argument := range call.AsCallExpression().Arguments.Nodes {
			if index >= len(given) ||
				argument.Kind == ast.KindObjectLiteralExpression ||
				argument.Kind == ast.KindArrayLiteralExpression {
				continue
			}
			from, to := lowering.checker.GetTypeAtLocation(argument), lowering.checker.GetTypeOfSymbol(given[index])
			if found := lowering.widened(from, to, map[[2]*checker.Type]bool{}); found != nil {
				return 0, &Refused{
					Where: lowering.program.Where(argument),
					What: "a type argument makes a value of type " +
						lowering.checker.TypeToString(from) +
						" seen as " +
						lowering.checker.TypeToString(to) +
						", which can write " +
						lowering.checker.TypeToString(found.target) +
						" where " +
						lowering.checker.TypeToString(found.source) +
						" is read",
					Fix: "use the value's invariant type argument, or make the parameter readonly (adamic/invariant-mutable)",
				}
			}
		}
	}

	if refuseGenericVoidReadsError := lowering.refuseGenericVoidReads(
		call,
		declaration,
		concreteTypes,
		target,
		resolved,
	); refuseGenericVoidReadsError != nil {
		return 0, refuseGenericVoidReadsError
	}

	substitution := map[*checker.Type]ir.Type{}

	owner := -1
	if local, ok := lowering.locals[lowering.symbol(
		declaration.Name(),
	)]; ok && lowering.result.Locals[local].NestedFunction < 0 {
		owner = lowering.result.Locals[local].Function
	}
	key := lowering.program.Where(declaration)
	name := declaration.Name().Text()
	if owner >= 0 {
		key += ",frame:" + strconv.Itoa(owner)
	}
	for _, parameter := range declaration.TypeParameters() {
		parameterType := lowering.checker.GetTypeAtLocation(parameter.Name())
		concrete, isKnown := concreteTypes[parameterType]
		if !isKnown {
			// Not read back (it stands only inside a union or an object, Result<Value>): left
			// unmapped, so whatever in the body needs to know how it's held says not yet, there.
			key += ",unread"
			name += "_unread"
			continue
		}
		held, hasRepresentation := lowering.representation(concrete)
		if !hasRepresentation {
			key += ",unread"
			name += "_unread"
			continue
		}
		substitution[parameterType] = held
		key += "," + lowering.genericTypeKey(concrete)
	}
	// Every request is judged, hit or miss, before the table is read: a body shared by requests that
	// differ, an unread parameter's, is judged with that parameter unmapped (predicates_branded.go).
	if brandedRequestError := lowering.brandedRequest(
		key,
		declaration,
		lowering.instanceMapper(declaration, concreteTypes),
	); brandedRequestError != nil {
		return 0, brandedRequestError
	}
	if existing, isLowered := lowering.genericInstances[key]; isLowered {
		return existing, nil
	}
	if lowering.genericDepth >= maximumGenericDepth {
		return 0, &Refused{
			Where: lowering.program.Where(call),
			What:  "a generic function instantiated without end (polymorphic recursion)",
			Fix:   "call it with the same type arguments it was called with, or write a function per type",
		}
	}

	source := lowering.sourceIdentity(declaration)
	for _, parameter := range declaration.TypeParameters() {
		concrete := concreteTypes[lowering.checker.GetTypeAtLocation(parameter.Name())]
		if concrete != nil {
			source.Specialization += ";" + lowering.sourceTypeKey(concrete, map[*checker.Type]bool{})
		} else {
			source.Specialization += ";unread"
		}
	}
	index := len(lowering.result.Functions)
	name += "_" + strconv.Itoa(index)
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: name, Source: source, Closure: owner >= 0, NestedParent: owner + 1},
	)
	if lowering.genericInstances == nil {
		lowering.genericInstances = map[string]int{}
	}
	lowering.genericInstances[key] = index

	// Lower with this instantiation's meaning of each type parameter, and locals of its own, outside
	// whatever function or closure called it.
	outerSubstitution, outerLocals, outerClosures, outerTypeMapper := lowering.substitution, lowering.locals, lowering.closures, lowering.typeMapper
	lowering.substitution, lowering.closures = substitution, nil
	// Validation may return before body lowering; restore the caller on every exit.
	defer func() {
		lowering.substitution, lowering.locals, lowering.closures, lowering.typeMapper = outerSubstitution, outerLocals, outerClosures, outerTypeMapper
	}()
	if owner >= 0 {
		for proven, held := range outerSubstitution {
			if _, ok := substitution[proven]; !ok {
				substitution[proven] = held
			}
		}
		lowering.closures = append(append([]int{}, outerClosures...), index)
	}
	lowering.typeMapper = lowering.instanceMapper(declaration, concreteTypes)
	if refuseInstantiatedMutationError := lowering.refuseInstantiatedMutation(
		declaration,
	); refuseInstantiatedMutationError != nil {
		return 0, refuseInstantiatedMutationError
	}
	lowering.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if lowering.result.Locals[local].Global || owner >= 0 {
			lowering.locals[symbol] = local
		}
	}
	lowering.genericDepth++
	defer func() {
		lowering.genericDepth--
	}()
	if owner >= 0 {
		lowering.closureRecords = append(lowering.closureRecords, closureRecord{
			proven:   lowering.concrete(lowering.checker.GetTypeAtLocation(declaration.Name())),
			function: index,
			node:     declaration,
		})
	}
	if lowerFunctionError := lowering.lowerFunction(index, declaration, -1); lowerFunctionError != nil {
		return 0, lowerFunctionError
	}
	return index, nil
}

// instanceMapper is the type mapper an instance's body is lowered under: each type parameter read back
// mapped to what it stands for, or, when none is, the mapper in force.
func (lowering *lowering) instanceMapper(
	declaration *ast.Node,
	concreteTypes map[*checker.Type]*checker.Type,
) *typeMapper {
	sources, targets := []*checker.Type{}, []*checker.Type{}
	for _, parameter := range declaration.TypeParameters() {
		parameterType := lowering.checker.GetTypeAtLocation(parameter.Name())
		if concrete, ok := concreteTypes[parameterType]; ok {
			sources, targets = append(sources, parameterType), append(targets, concrete)
		}
	}
	if len(sources) == 0 {
		return lowering.typeMapper
	}
	return newTypeMapper(sources, targets)
}

// genericTypeKey is the shared identity key for class and function type arguments.
// Keep the concrete representatives in instantiated, the type universe the cycle
// finder already visits. Checker identity includes union members, nested containers,
// signatures and structural fields; a native representation alone loses those facts.
// This is also where nullable representations can extend the key.
func (lowering *lowering) genericTypeKey(proven *checker.Type) string {
	held, known := lowering.representation(proven)
	if !known {
		return "unread"
	}
	for index, representative := range lowering.instantiated {
		if identicalTypes(lowering.checker, proven, representative) {
			return typeName(held) + ":" + strconv.Itoa(index)
		}
	}
	index := len(lowering.instantiated)
	lowering.instantiated = append(lowering.instantiated, proven)
	return typeName(held) + ":" + strconv.Itoa(index)
}

// refuseInstantiatedMutation checks writes whose safety depends on a type parameter's constraint.
// TypeScript checks the generic body against the constraint, so Pack extends Animal[] permits a Cat
// to be pushed. An instantiation with Pack = Dog[] would then put that Cat in a Dog[]; monomorphizing
// the body must recheck that write against Dog rather than trusting the wider constraint.
// JSON inputs also need their concrete schemas checked before lowering can erase their types.
func (lowering *lowering) refuseInstantiatedMutation(declaration *ast.Node) error {
	var refused error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if refused != nil {
			return true
		}
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression &&
				callee.Name().Text() == "stringify" &&
				lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") {
				for _, argument := range call.Arguments.Nodes {
					if checkInstantiatedJsonInputError := lowering.checkInstantiatedJSONInput(
						argument,
					); checkInstantiatedJsonInputError != nil {
						refused = checkInstantiatedJsonInputError
						return true
					}
				}
			}
			if callee.Kind == ast.KindPropertyAccessExpression {
				receiver := callee.AsPropertyAccessExpression().Expression
				argumentIndex, typeIndex := -1, -1
				switch callee.Name().Text() {
				case "push", "add":
					argumentIndex, typeIndex = 0, 0
				case "set":
					argumentIndex, typeIndex = 1, 1
				}
				concrete := lowering.concrete(lowering.checker.GetTypeAtLocation(receiver))
				if argumentIndex < 0 ||
					concrete.Flags()&checker.TypeFlagsObject == 0 ||
					concrete.ObjectFlags()&checker.ObjectFlagsReference == 0 {
					return node.ForEachChild(visit)
				}
				arguments := lowering.checker.GetTypeArguments(concrete)
				if argumentIndex >= 0 && argumentIndex < len(call.Arguments.Nodes) && typeIndex < len(arguments) {
					argument := call.Arguments.Nodes[argumentIndex]
					from, to := lowering.concrete(lowering.checker.GetTypeAtLocation(argument)), arguments[typeIndex]
					if !lowering.checker.IsTypeAssignableTo(from, to) {
						refused = &Refused{
							Where: lowering.program.Where(argument),
							What: "instantiating a generic function makes a value of type " +
								lowering.checker.TypeToString(from) +
								" written where " +
								lowering.checker.TypeToString(to) +
								" is read",
							Fix: "make the collection readonly, or use a type parameter for the value being written (adamic/invariant-mutable)",
						}
						return true
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	declaration.ForEachChild(visit)
	return refused
}

// inferTypes records what each type parameter in declared stands for, by where it stands in
// instantiated: the type itself, a referenced type's arguments, or a function's parameters and
// result. The caller's type mapper makes an outer function's parameter concrete before it is saved,
// so a generic function calling another with its own type parameter passes the concrete type on.
func (lowering *lowering) inferTypes(
	declared *checker.Type,
	instantiated *checker.Type,
	into map[*checker.Type]*checker.Type,
) {
	if declared == nil || instantiated == nil {
		return
	}
	// Optional implementation parameters and results can wrap the same generic
	// binder that the resolved overload exposes directly. Infer from the present
	// member; an absent argument supplies no evidence about that binder.
	if declared.Flags()&checker.TypeFlagsUnion != 0 && lowering.censusHasUndefined(declared) {
		present := []*checker.Type{}
		for _, member := range declared.Types() {
			if member.Flags()&checker.TypeFlagsUndefined == 0 {
				present = append(present, member)
			}
		}
		given := instantiated
		if lowering.censusHasUndefined(given) && given.Flags()&checker.TypeFlagsUnion != 0 {
			members := []*checker.Type{}
			for _, member := range given.Types() {
				if member.Flags()&checker.TypeFlagsUndefined == 0 {
					members = append(members, member)
				}
			}
			given = lowering.checker.GetUnionType(members)
		}
		// GetNonNullableType also turns a rigid T into T & {}. Removing only
		// the optional member preserves the binder we are trying to infer.
		if len(present) == 1 && given.Flags()&checker.TypeFlagsUndefined == 0 {
			lowering.inferTypes(present[0], given, into)
		}
		return
	}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		var missing *checker.Type
		for _, member := range declared.Types() {
			known, found := into[member]
			if !found && member.Flags()&checker.TypeFlagsTypeParameter != 0 {
				if missing != nil {
					return // More than one unknown binder has no unique witness.
				}
				missing = member
				continue
			}
			if !found {
				known = lowering.concrete(member)
			}
			if !lowering.checker.IsTypeAssignableTo(known, lowering.concrete(instantiated)) {
				return
			}
		}
		if missing != nil {
			into[missing] = lowering.concrete(instantiated)
		}
		return
	}
	if declared.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if _, isSet := into[declared]; !isSet {
			into[declared] = lowering.concrete(instantiated)
		}
		return
	}
	if declared.Flags()&checker.TypeFlagsObject != 0 && instantiated.Flags()&checker.TypeFlagsObject != 0 {
		if declared.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
			instantiated.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			declaredArguments, instantiatedArguments := lowering.checker.GetTypeArguments(
				declared,
			), lowering.checker.GetTypeArguments(instantiated)
			for index := range declaredArguments {
				if index < len(instantiatedArguments) {
					lowering.inferTypes(declaredArguments[index], instantiatedArguments[index], into)
				}
			}
		}
		declaredSignatures, instantiatedSignatures := lowering.checker.GetSignaturesOfType(
			declared,
			checker.SignatureKindCall,
		), lowering.checker.GetSignaturesOfType(instantiated, checker.SignatureKindCall)
		if len(declaredSignatures) > 0 && len(instantiatedSignatures) > 0 {
			declaredParameters, instantiatedParameters := declaredSignatures[0].Parameters(), instantiatedSignatures[0].Parameters()
			for index := range declaredParameters {
				if index < len(instantiatedParameters) {
					lowering.inferTypes(
						lowering.checker.GetTypeOfSymbol(declaredParameters[index]),
						lowering.checker.GetTypeOfSymbol(instantiatedParameters[index]),
						into,
					)
				}
			}
			lowering.inferTypes(
				lowering.checker.GetReturnTypeOfSignature(declaredSignatures[0]),
				lowering.checker.GetReturnTypeOfSignature(instantiatedSignatures[0]),
				into,
			)
		}
	}
}

// JSON descriptors must be checked against the actual type arguments too. Without
// this check a type parameter can hide a container inside the scalar union schema.
func (lowering *lowering) checkInstantiatedJSONInput(node *ast.Node) error {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNullKeyword {
		return nil
	}
	if node.Kind == ast.KindArrayLiteralExpression {
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
				continue
			}
			if checkInstantiatedJsonInputError := lowering.checkInstantiatedJSONInput(
				element,
			); checkInstantiatedJsonInputError != nil {
				return checkInstantiatedJsonInputError
			}
		}
		return nil
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
			if field.Kind == ast.KindPropertyAssignment {
				if checkInstantiatedJsonInputError := lowering.checkInstantiatedJSONInput(
					field.AsPropertyAssignment().Initializer,
				); checkInstantiatedJsonInputError != nil {
					return checkInstantiatedJsonInputError
				}
			}
		}
		return nil
	}
	concrete := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	if concrete.Flags()&checker.TypeFlagsUnion != 0 {
		held, known := lowering.representation(concrete)
		if known && held == ir.Union {
			for _, member := range concrete.Types() {
				if lowering.isArrayType(member) {
					return lowering.notYet(node, "JSON.stringify unions containing arrays (adamic/json-union-array)")
				}
			}
		}
	}
	_, jsonTypeError := lowering.jsonType(node, concrete, 0)
	return jsonTypeError
}
