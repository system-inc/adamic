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
// distinct set of what they're held as at runtime, as a generic class is (instantiate). Within it,
// each type parameter is what this call made it.
//
// The checker resolves the call's signature with its type arguments substituted, but doesn't export
// the mapping itself, so it's read back the way it was made: each declared parameter's type, and the
// result's, against the resolved signature's, through arrays and tuples. A type parameter that can't
// be read back that way is left unmapped, and the body says not yet wherever it needs to know it.
func (l *lowering) instantiateFunction(call *ast.Node, declaration *ast.Node) (int, error) {
	resolved := l.checker.GetResolvedSignature(call)
	target := l.checker.GetSignatureFromDeclaration(declaration)
	if resolved == nil || target == nil {
		return 0, l.notYet(call, "a call to a generic function whose signature the checker didn't resolve")
	}
	concreteTypes := map[*checker.Type]*checker.Type{}
	// Class-generic calls supply this callee's resolved mapper. Ordinary recursive
	// calls must infer their own arguments: the outer mapper can still describe the
	// same declaration's previous instantiation, as nest<T>([item], depth - 1) does.
	if l.genericUsesClasses(declaration, map[*ast.Node]bool{}) {
		for _, parameter := range declaration.TypeParameters() {
			declaredType := l.checker.GetTypeAtLocation(parameter.Name())
			concrete := l.concrete(declaredType)
			if concrete != declaredType && concrete.Flags()&checker.TypeFlagsTypeParameter == 0 {
				concreteTypes[declaredType] = concrete
			}
		}
	}
	declared, given := target.Parameters(), resolved.Parameters()
	for index := range declared {
		if index < len(given) {
			l.inferTypes(l.checker.GetTypeOfSymbol(declared[index]), l.checker.GetTypeOfSymbol(given[index]), concreteTypes)
		}
	}
	l.inferTypes(l.checker.GetReturnTypeOfSignature(target), l.checker.GetReturnTypeOfSignature(resolved), concreteTypes)

	// An explicit type argument can make tsc view a mutable argument through a wider type without
	// giving the argument that contextual type. Judge the instantiated parameter directly: the
	// ordinary refusal walk cannot see this view at the argument node.
	if call.AsCallExpression().TypeArguments != nil && len(call.AsCallExpression().TypeArguments.Nodes) > 0 {
		for index, argument := range call.AsCallExpression().Arguments.Nodes {
			if index >= len(given) || argument.Kind == ast.KindObjectLiteralExpression || argument.Kind == ast.KindArrayLiteralExpression {
				continue
			}
			from, to := l.checker.GetTypeAtLocation(argument), l.checker.GetTypeOfSymbol(given[index])
			if !l.nullableViewsMatch(from, to, map[[2]*checker.Type]bool{}) {
				return 0, l.notYet(argument, "a nullable reference view changes its empty case or adds a second empty case (nullable reference needs an empty-case tag)")
			}
			if found := l.widened(from, to, map[[2]*checker.Type]bool{}); found != nil {
				return 0, &Refused{Where: l.program.Where(argument), What: "a type argument makes a value of type " + l.checker.TypeToString(from) + " seen as " + l.checker.TypeToString(to) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read", Fix: "use the value's invariant type argument, or make the parameter readonly (adamic/invariant-mutable)"}
			}
		}
	}

	substitution := map[*checker.Type]ir.Type{}

	key := l.program.Where(declaration)
	name := declaration.Name().Text()
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
		if l.includesNull(concrete) && l.includesUndefined(concrete) {
			return 0, l.notYet(call, nullableTagReason)
		}
		held, hasRepresentation := l.representation(concrete)
		if concrete.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			held, hasRepresentation = ir.Object, true
		}
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
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name})
	if l.genericInstances == nil {
		l.genericInstances = map[string]int{}
	}
	l.genericInstances[key] = index

	// Lower with this instantiation's meaning of each type parameter, and locals of its own, outside
	// whatever function or closure called it.
	outerSubstitution, outerLocals, outerClosures, outerTypeMapper := l.substitution, l.locals, l.closures, l.typeMapper
	l.substitution, l.closures = substitution, nil
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
		if l.result.Locals[local].Global {
			l.locals[symbol] = local
		}
	}
	l.genericDepth++
	defer func() {
		l.substitution, l.locals, l.closures, l.typeMapper = outerSubstitution, outerLocals, outerClosures, outerTypeMapper
		l.genericDepth--
	}()
	if err := l.lowerFunction(index, declaration, -1); err != nil {
		return 0, err
	}
	return index, nil
}

// genericTypeKey distinguishes every instantiation whose lowered operations or object layout differ.
// Containers include how each argument is held, so number[] and string[] differ, while T[], T[][]
// and deeper arrays reach a fixed point once T itself is an array. Class layouts keep their full
// concrete type because Box<number> and Box<Box<number>> have different fields.
func (l *lowering) genericTypeKey(proven *checker.Type) string {
	proven = l.concrete(proven)
	// The empty case changes operations even though both reference types use NULL.
	if l.includesNull(proven) || l.includesUndefined(proven) {
		return l.checker.TypeToString(proven)
	}
	held, known := l.representation(proven)
	if !known {
		return "unread"
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		(l.checker.IsArrayType(proven) || l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
		key := typeName(held)
		for _, argument := range l.checker.GetTypeArguments(proven) {
			argumentHeld, argumentKnown := l.representation(argument)
			if !argumentKnown {
				key += ",unread"
			} else {
				if l.includesNull(argument) || l.includesUndefined(argument) {
					key += "," + l.checker.TypeToString(argument)
				} else {
					key += "," + typeName(argumentHeld)
				}
			}
		}
		return key
	}
	if isClassInstance(proven) {
		return l.checker.TypeToString(proven)
	}
	return typeName(held)
}

// refuseInstantiatedMutation checks writes whose safety depends on a type parameter's constraint.
// TypeScript checks the generic body against the constraint, so Pack extends Animal[] permits a Cat
// to be pushed. An instantiation with Pack = Dog[] would then put that Cat in a Dog[]; monomorphizing
// the body must recheck that write against Dog rather than trusting the wider constraint.
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
	l.inferTypesSeen(declared, instantiated, into, map[[2]*checker.Type]bool{})
}

func (l *lowering) inferTypesSeen(declared *checker.Type, instantiated *checker.Type, into map[*checker.Type]*checker.Type, visited map[[2]*checker.Type]bool) {
	if declared == nil || instantiated == nil {
		return
	}
	if visited[[2]*checker.Type{declared, instantiated}] {
		return
	}
	visited[[2]*checker.Type{declared, instantiated}] = true
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
					l.inferTypesSeen(declaredArguments[index], instantiatedArguments[index], into, visited)
				}
			}
		}
		if declared.ObjectFlags()&checker.ObjectFlagsReference == 0 && len(l.checker.GetSignaturesOfType(declared, checker.SignatureKindCall)) == 0 {
			for _, property := range l.checker.GetPropertiesOfType(declared) {
				if given := l.checker.GetTypeOfPropertyOfType(instantiated, property.Name); given != nil {
					l.inferTypesSeen(l.checker.GetTypeOfPropertyOfType(declared, property.Name), given, into, visited)
				}
			}
		}
		declaredSignatures, instantiatedSignatures := l.checker.GetSignaturesOfType(declared, checker.SignatureKindCall), l.checker.GetSignaturesOfType(instantiated, checker.SignatureKindCall)
		if len(declaredSignatures) > 0 && len(instantiatedSignatures) > 0 {
			declaredParameters, instantiatedParameters := declaredSignatures[0].Parameters(), instantiatedSignatures[0].Parameters()
			for index := range declaredParameters {
				if index < len(instantiatedParameters) {
					l.inferTypesSeen(l.checker.GetTypeOfSymbol(declaredParameters[index]), l.checker.GetTypeOfSymbol(instantiatedParameters[index]), into, visited)
				}
			}
			l.inferTypesSeen(l.checker.GetReturnTypeOfSignature(declaredSignatures[0]), l.checker.GetReturnTypeOfSignature(instantiatedSignatures[0]), into, visited)
		}
	}
}
