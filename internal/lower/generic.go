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
// The resolved signature supplies both the return contract and the checker's
// actual type argument mapper. Reconstructing that mapper from parameter shapes
// loses arguments inside unions and conditional types.
func (l *lowering) instantiateFunction(call *ast.Node, declaration *ast.Node) (int, error) {
	resolved := l.checker.GetResolvedSignature(call)
	target := l.checker.GetSignatureFromDeclaration(declaration)
	if resolved == nil || target == nil {
		return 0, l.notYet(call, "a call to a generic function whose signature the checker didn't resolve")
	}
	// A resolved overload describes its promise, not the implementation's
	// binders or wider result. Keep the existing implementation inference and
	// result boundary; ordinary calls use the checker's exact resolved mapper.
	overload := resolved.Declaration() != nil && resolved.Declaration().Body() == nil && l.censusImplementation(resolved.Declaration()) == declaration
	var resolvedReturn *checker.Type
	if !overload {
		var err error
		resolvedReturn, err = l.resolvedFunctionReturn(call, resolved)
		if err != nil {
			return 0, err
		}
	}
	concreteTypes := map[*checker.Type]*checker.Type{}
	classContext := l.genericUsesClasses(declaration, map[*ast.Node]bool{})
	given := resolved.Parameters()
	if overload {
		if classContext {
			for _, parameter := range declaration.TypeParameters() {
				declaredType := l.checker.GetTypeAtLocation(parameter.Name())
				concrete := l.concrete(declaredType)
				if concrete != declaredType && concrete.Flags()&checker.TypeFlagsTypeParameter == 0 {
					concreteTypes[declaredType] = concrete
				}
			}
		}
		for index, parameter := range target.Parameters() {
			if index < len(given) {
				l.inferTypes(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(given[index]), concreteTypes)
			}
		}
		l.inferTypes(l.checker.GetReturnTypeOfSignature(target), l.checker.GetReturnTypeOfSignature(resolved), concreteTypes)
	} else {
		mapper := resolvedTypeMapper(resolved)
		if mapper == nil {
			return 0, l.notYet(call, "a generic function whose type argument mapper the checker didn't resolve")
		}
		signatureParameters := target.TypeParameters()
		if resolved.Target() != nil {
			signatureParameters = resolved.Target().TypeParameters()
		}
		if len(signatureParameters) < len(declaration.TypeParameters()) {
			return 0, l.notYet(call, "a generic function whose resolved type parameter correspondence is unknown")
		}
		for index, parameter := range declaration.TypeParameters() {
			declaredType := l.checker.GetTypeAtLocation(parameter.Name())
			concrete := l.concrete(checker.Checker_instantiateType(l.checker, signatureParameters[index], mapper))
			if classContext {
				if active := l.concrete(declaredType); active != declaredType && active.Flags()&checker.TypeFlagsTypeParameter == 0 {
					concrete = active
				}
			}
			if concrete == nil || concrete == declaredType || concrete.Flags()&checker.TypeFlagsAny != 0 {
				return 0, l.notYet(call, "a generic function whose type argument "+parameter.Name().Text()+" is any or unresolved")
			}
			concreteTypes[declaredType] = concrete
		}
	}

	// An explicit type argument can make tsc view a mutable argument through a wider type without
	// giving the argument that contextual type. Judge the instantiated parameter directly: the
	// ordinary refusal walk cannot see this view at the argument node.
	if call.AsCallExpression().TypeArguments != nil && len(call.AsCallExpression().TypeArguments.Nodes) > 0 {
		for index, argument := range call.AsCallExpression().Arguments.Nodes {
			if index >= len(given) || argument.Kind == ast.KindObjectLiteralExpression || argument.Kind == ast.KindArrayLiteralExpression {
				continue
			}
			from, to := l.checker.GetTypeAtLocation(argument), l.checker.GetTypeOfSymbol(given[index])
			if found := l.widened(from, to, map[[2]*checker.Type]bool{}); found != nil {
				return 0, &Refused{Where: l.program.Where(argument), What: "a type argument makes a value of type " + l.checker.TypeToString(from) + " seen as " + l.checker.TypeToString(to) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read", Fix: "use the value's invariant type argument, or make the parameter readonly (adamic/invariant-mutable)"}
			}
		}
	}

	substitution := map[*checker.Type]ir.Type{}

	owner := -1
	if local, ok := l.locals[l.symbol(declaration.Name())]; ok && l.result.Locals[local].NestedFunction < 0 {
		owner = l.result.Locals[local].Function
	}
	key := l.program.Where(declaration)
	name := declaration.Name().Text()
	if owner >= 0 {
		key += ",frame:" + strconv.Itoa(owner)
	}
	for _, parameter := range declaration.TypeParameters() {
		parameterType := l.checker.GetTypeAtLocation(parameter.Name())
		concrete, isKnown := concreteTypes[parameterType]
		if !isKnown {
			// Not read back (it stands only inside a union or an object, Result<Value>): left
			// unmapped, so whatever in the body needs to know how it's held says not yet, there.
			key += ",unread"
			name += "_unread"
			continue
		}
		held, hasRepresentation := l.representation(concrete)
		if !hasRepresentation {
			key += ",unread"
			name += "_unread"
			continue
		}
		substitution[parameterType] = held
		key += "," + l.genericTypeKey(concrete)
	}
	if existing, isLowered := l.genericInstances[key]; isLowered {
		return existing, nil
	}
	if l.genericDepth >= maximumGenericDepth {
		return 0, &Refused{Where: l.program.Where(call), What: "a generic function instantiated without end (polymorphic recursion)", Fix: "call it with the same type arguments it was called with, or write a function per type"}
	}

	index := len(l.result.Functions)
	name += "_" + strconv.Itoa(index)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name, Closure: owner >= 0, NestedParent: owner + 1})
	if l.genericInstances == nil {
		l.genericInstances = map[string]int{}
	}
	l.genericInstances[key] = index

	// Lower with this instantiation's meaning of each type parameter, and locals of its own, outside
	// whatever function or closure called it.
	outerSubstitution, outerLocals, outerClosures, outerTypeMapper := l.substitution, l.locals, l.closures, l.typeMapper
	l.substitution, l.closures = substitution, nil
	if owner >= 0 {
		for proven, held := range outerSubstitution {
			if _, ok := substitution[proven]; !ok {
				substitution[proven] = held
			}
		}
		l.closures = append(append([]int{}, outerClosures...), index)
	}
	sources, targets := []*checker.Type{}, []*checker.Type{}
	for _, parameter := range declaration.TypeParameters() {
		parameterType := l.checker.GetTypeAtLocation(parameter.Name())
		if concrete, ok := concreteTypes[parameterType]; ok {
			sources, targets = append(sources, parameterType), append(targets, concrete)
		}
	}
	if len(sources) > 0 {
		l.typeMapper = newTypeMapper(sources, targets)
	}
	if err := l.refuseInstantiatedMutation(declaration); err != nil {
		return 0, err
	}
	l.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if l.result.Locals[local].Global || owner >= 0 {
			l.locals[symbol] = local
		}
	}
	l.genericDepth++
	defer func() {
		l.substitution, l.locals, l.closures, l.typeMapper = outerSubstitution, outerLocals, outerClosures, outerTypeMapper
		l.genericDepth--
	}()
	if owner >= 0 {
		l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(declaration.Name())), function: index, node: declaration})
	}
	if err := l.signatureReturn(index, declaration, -1, resolvedReturn); err != nil {
		return 0, err
	}
	if err := l.lowerFunction(index, declaration, -1); err != nil {
		return 0, err
	}
	return index, nil
}

// genericTypeKey is the shared identity key for class and function type arguments.
// Keep the concrete representatives in instantiated, the type universe the cycle
// finder already visits. Checker identity includes union members, nested containers,
// signatures and structural fields; a native representation alone loses those facts.
// This is also where nullable representations can extend the key.
func (l *lowering) genericTypeKey(proven *checker.Type) string {
	held, known := l.representation(proven)
	if !known {
		return "unread"
	}
	for index, representative := range l.instantiated {
		if identicalTypes(l.checker, proven, representative) {
			return typeName(held) + ":" + strconv.Itoa(index)
		}
	}
	index := len(l.instantiated)
	l.instantiated = append(l.instantiated, proven)
	return typeName(held) + ":" + strconv.Itoa(index)
}

// refuseInstantiatedMutation checks writes whose safety depends on a type parameter's constraint.
// TypeScript checks the generic body against the constraint, so Pack extends Animal[] permits a Cat
// to be pushed. An instantiation with Pack = Dog[] would then put that Cat in a Dog[]; monomorphizing
// the body must recheck that write against Dog rather than trusting the wider constraint.
// JSON inputs also need their concrete schemas checked before lowering can erase their types.
func (l *lowering) refuseInstantiatedMutation(declaration *ast.Node) error {
	var refused error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if refused != nil {
			return true
		}
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "stringify" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") {
				for _, argument := range call.Arguments.Nodes {
					if err := l.checkInstantiatedJSONInput(argument); err != nil {
						refused = err
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
				concrete := l.concrete(l.checker.GetTypeAtLocation(receiver))
				if argumentIndex < 0 || concrete.Flags()&checker.TypeFlagsObject == 0 || concrete.ObjectFlags()&checker.ObjectFlagsReference == 0 {
					return node.ForEachChild(visit)
				}
				arguments := l.checker.GetTypeArguments(concrete)
				if argumentIndex >= 0 && argumentIndex < len(call.Arguments.Nodes) && typeIndex < len(arguments) {
					argument := call.Arguments.Nodes[argumentIndex]
					from, to := l.concrete(l.checker.GetTypeAtLocation(argument)), arguments[typeIndex]
					if !l.checker.IsTypeAssignableTo(from, to) {
						refused = &Refused{Where: l.program.Where(argument), What: "instantiating a generic function makes a value of type " + l.checker.TypeToString(from) + " written where " + l.checker.TypeToString(to) + " is read", Fix: "make the collection readonly, or use a type parameter for the value being written (adamic/invariant-mutable)"}
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
func (l *lowering) inferTypes(declared *checker.Type, instantiated *checker.Type, into map[*checker.Type]*checker.Type) {
	if declared == nil || instantiated == nil {
		return
	}
	// Optional implementation parameters and results can wrap the same generic
	// binder that the resolved overload exposes directly. Infer from the present
	// member; an absent argument supplies no evidence about that binder.
	if declared.Flags()&checker.TypeFlagsUnion != 0 && l.censusHasUndefined(declared) {
		present := []*checker.Type{}
		for _, member := range declared.Types() {
			if member.Flags()&checker.TypeFlagsUndefined == 0 {
				present = append(present, member)
			}
		}
		given := instantiated
		if l.censusHasUndefined(given) && given.Flags()&checker.TypeFlagsUnion != 0 {
			members := []*checker.Type{}
			for _, member := range given.Types() {
				if member.Flags()&checker.TypeFlagsUndefined == 0 {
					members = append(members, member)
				}
			}
			given = l.checker.GetUnionType(members)
		}
		// GetNonNullableType also turns a rigid T into T & {}. Removing only
		// the optional member preserves the binder we are trying to infer.
		if len(present) == 1 && given.Flags()&checker.TypeFlagsUndefined == 0 {
			l.inferTypes(present[0], given, into)
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
				known = l.concrete(member)
			}
			if !l.checker.IsTypeAssignableTo(known, l.concrete(instantiated)) {
				return
			}
		}
		if missing != nil {
			into[missing] = l.concrete(instantiated)
		}
		return
	}
	if declared.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if _, isSet := into[declared]; !isSet {
			into[declared] = l.concrete(instantiated)
		}
		return
	}
	if declared.Flags()&checker.TypeFlagsObject != 0 && instantiated.Flags()&checker.TypeFlagsObject != 0 {
		if declared.ObjectFlags()&checker.ObjectFlagsReference != 0 && instantiated.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			declaredArguments, instantiatedArguments := l.checker.GetTypeArguments(declared), l.checker.GetTypeArguments(instantiated)
			for index := range declaredArguments {
				if index < len(instantiatedArguments) {
					l.inferTypes(declaredArguments[index], instantiatedArguments[index], into)
				}
			}
		}
		declaredSignatures, instantiatedSignatures := l.checker.GetSignaturesOfType(declared, checker.SignatureKindCall), l.checker.GetSignaturesOfType(instantiated, checker.SignatureKindCall)
		if len(declaredSignatures) > 0 && len(instantiatedSignatures) > 0 {
			declaredParameters, instantiatedParameters := declaredSignatures[0].Parameters(), instantiatedSignatures[0].Parameters()
			for index := range declaredParameters {
				if index < len(instantiatedParameters) {
					l.inferTypes(l.checker.GetTypeOfSymbol(declaredParameters[index]), l.checker.GetTypeOfSymbol(instantiatedParameters[index]), into)
				}
			}
			l.inferTypes(l.checker.GetReturnTypeOfSignature(declaredSignatures[0]), l.checker.GetReturnTypeOfSignature(instantiatedSignatures[0]), into)
		}
	}
}

// JSON descriptors must be checked against the actual type arguments too. Without
// this check a type parameter can hide a container inside the scalar union schema.
func (l *lowering) checkInstantiatedJSONInput(node *ast.Node) error {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNullKeyword {
		return nil
	}
	if node.Kind == ast.KindArrayLiteralExpression {
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
				continue
			}
			if err := l.checkInstantiatedJSONInput(element); err != nil {
				return err
			}
		}
		return nil
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
			if field.Kind == ast.KindPropertyAssignment {
				if err := l.checkInstantiatedJSONInput(field.AsPropertyAssignment().Initializer); err != nil {
					return err
				}
			}
		}
		return nil
	}
	concrete := l.concrete(l.checker.GetTypeAtLocation(node))
	if concrete.Flags()&checker.TypeFlagsUnion != 0 {
		held, known := l.representation(concrete)
		if known && held == ir.Union {
			for _, member := range concrete.Types() {
				if l.checker.IsArrayType(member) {
					return l.notYet(node, "JSON.stringify unions containing arrays (adamic/json-union-array)")
				}
			}
		}
	}
	_, err := l.jsonType(node, concrete, 0)
	return err
}

// resolvedFunctionReturn preserves the checker's substituted call contract.
// A failed resolution is a refusal boundary, never a representation of any.
func (l *lowering) resolvedFunctionReturn(call *ast.Node, signature *checker.Signature) (*checker.Type, error) {
	if signature == nil {
		return nil, l.notYet(call, "a generic function whose signature the checker didn't resolve")
	}
	returns := l.concrete(l.checker.GetReturnTypeOfSignature(signature))
	if returns == nil {
		return nil, l.notYet(call, "a generic function whose return type the checker didn't resolve")
	}
	if returns.Flags()&checker.TypeFlagsAny != 0 {
		return nil, l.notYet(call, "a generic function whose resolved return type is any")
	}
	return returns, nil
}
