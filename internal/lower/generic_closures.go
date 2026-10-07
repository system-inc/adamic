package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A generic function value keeps one ordinary counted closure per concrete call
// signature. The containing object preserves identity and captures; choosing a
// slot at the call preserves runtime property lookup, including replacement.
func (l *lowering) genericSignature(proven *checker.Type) *checker.Signature {
	if proven == nil {
		return nil
	}
	proven = l.present(l.concrete(proven))
	if proven == nil {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) == 1 && signatures[0].Declaration() != nil && len(signatures[0].Declaration().TypeParameters()) > 0 {
		return signatures[0]
	}
	return nil
}

func (l *lowering) genericClosureArguments(call *ast.Node, declared *checker.Signature) ([]*checker.Type, bool) {
	declared = l.checker.GetSignatureFromDeclaration(declared.Declaration())
	resolved := l.checker.GetResolvedSignature(call)
	if resolved == nil {
		return nil, false
	}
	inferred := map[*checker.Type]*checker.Type{}
	parameters, given := declared.Parameters(), resolved.Parameters()
	for index, parameter := range parameters {
		if index < len(given) {
			l.inferTypes(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(given[index]), inferred)
		}
	}
	l.inferTypes(l.checker.GetReturnTypeOfSignature(declared), l.checker.GetReturnTypeOfSignature(resolved), inferred)
	var arguments []*checker.Type
	for _, parameter := range declared.Declaration().TypeParameters() {
		argument := inferred[l.checker.GetTypeAtLocation(parameter.Name())]
		if argument == nil {
			return nil, false
		}
		argument = l.concrete(argument)
		if _, known := l.representation(argument); !known {
			return nil, false
		}
		arguments = append(arguments, argument)
	}
	return arguments, true
}

// Slots use the resolved parameter/result identities, rather than type-parameter
// names or order. A structurally compatible signature may rename its parameters.
func (l *lowering) genericClosureKey(call *ast.Node) string {
	signature := l.checker.GetResolvedSignature(call)
	keys := []string{"generic"}
	for _, parameter := range signature.Parameters() {
		keys = append(keys, l.genericTypeKey(l.concrete(l.checker.GetTypeOfSymbol(parameter))))
	}
	keys = append(keys, l.genericTypeKey(l.concrete(l.checker.GetReturnTypeOfSignature(signature))))
	return strings.Join(keys, ";")
}

// Discover concrete signatures before making the runtime value. Compatible
// generic signatures share a family, so interface aliases and nested forwarding
// helpers get the same slots. Unresolved call families stay NotYet, never an
// absent runtime slot. Generic calls inside another generic body are not a new
// concrete seed until the enclosing mapper supplies their type arguments.
func (l *lowering) genericClosure(node *ast.Node) (ir.Expression, error) {
	if l.typeMapper != nil {
		return nil, l.notYet(node, "a generic function value created inside another instantiation")
	}
	if l.genericDepth >= maximumGenericDepth {
		return nil, l.notYet(node, "a generic closure instantiated without end")
	}
	declaredType := l.checker.GetTypeAtLocation(node)
	declared := l.checker.GetSignatureFromDeclaration(node)
	argumentsByKey, keys := l.genericClosureFamilies(declaredType)

	if len(keys) == 0 {
		return nil, l.notYet(node, "a generic function value with no concrete call family")
	}
	bundle := ir.ObjectLiteral{}
	for _, key := range keys {
		closure, err := l.genericClosureInstance(node, declared, argumentsByKey[key])
		if err != nil {
			return nil, err
		}
		bundle.Fields = append(bundle.Fields, ir.Field{Name: key, Value: closure})
	}
	return bundle, nil
}

// Only source-concrete calls seed a family. Calls in a generic body may use an
// already seeded slot, but cannot silently introduce a slot after a value was
// constructed. More general higher-order specialization remains NotYet.
func (l *lowering) genericClosureFamilies(declaredType *checker.Type) (map[string][]*checker.Type, []string) {
	outerMapper, outerSubstitution := l.typeMapper, l.substitution
	l.typeMapper, l.substitution = nil, nil
	defer func() { l.typeMapper, l.substitution = outerMapper, outerSubstitution }()
	declaration := l.genericSignature(declaredType).Declaration()
	argumentsByKey := map[string][]*checker.Type{}
	var keys []string
	for _, module := range l.program.Files() {
		var visit ast.Visitor
		visit = func(candidate *ast.Node) bool {
			if ast.IsPartOfTypeNode(candidate) {
				return false
			}
			if candidate.Kind == ast.KindCallExpression {
				calleeType := l.checker.GetTypeAtLocation(candidate.AsCallExpression().Expression)
				signature := l.genericSignature(calleeType)
				if signature != nil && len(signature.Declaration().TypeParameters()) == len(declaration.TypeParameters()) && l.checker.IsTypeAssignableTo(declaredType, calleeType) && l.checker.IsTypeAssignableTo(calleeType, declaredType) {
					if arguments, known := l.genericClosureArguments(candidate, l.checker.GetSignatureFromDeclaration(declaration)); known {
						key := l.genericClosureKey(candidate)
						if _, exists := argumentsByKey[key]; !exists {
							argumentsByKey[key] = arguments
							keys = append(keys, key)
						}
					}
				}
			}
			return candidate.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	return argumentsByKey, keys
}

func (l *lowering) genericClosureInstance(node *ast.Node, signature *checker.Signature, arguments []*checker.Type) (ir.Expression, error) {
	outerLocals, outerSubstitution, outerMapper := l.locals, l.substitution, l.typeMapper
	l.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		l.locals[symbol] = local
	}
	l.substitution = map[*checker.Type]ir.Type{}
	for source, held := range outerSubstitution {
		l.substitution[source] = held
	}
	var sources []*checker.Type
	for index, parameter := range signature.Declaration().TypeParameters() {
		source := l.checker.GetTypeAtLocation(parameter.Name())
		sources = append(sources, source)
		held, _ := l.representation(arguments[index])
		l.substitution[source] = held
	}
	l.typeMapper = newTypeMapper(sources, arguments)
	l.genericDepth++
	defer func() {
		l.locals, l.substitution, l.typeMapper = outerLocals, outerSubstitution, outerMapper
		l.genericDepth--
	}()
	if err := l.refuseInstantiatedMutation(node); err != nil {
		return nil, err
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "generic_closure_" + strconv.Itoa(index), Closure: true})
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: index, node: node})
	l.closures = append(l.closures, index)
	err := l.lowerFunction(index, node, -1)
	l.closures = l.closures[:len(l.closures)-1]
	if err != nil {
		return nil, err
	}
	return ir.MakeClosure{Function: index}, nil
}

func (l *lowering) callGenericClosure(node *ast.Node, signature *checker.Signature) (ir.Expression, error) {
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, l.notYet(node, "an optional call through a generic function value")
	}
	_, known := l.genericClosureArguments(node, signature)
	if !known {
		return nil, l.notYet(node, "a generic function value call without concrete type arguments")
	}
	families, _ := l.genericClosureFamilies(l.checker.GetTypeAtLocation(node.AsCallExpression().Expression))
	key := l.genericClosureKey(node)
	if _, seeded := families[key]; !seeded {
		return nil, l.notYet(node, "a generic function value call without a source-concrete family")
	}
	bundle, err := l.expression(node.AsCallExpression().Expression)
	if err != nil {
		return nil, err
	}
	closure := ir.Property{Object: bundle, Name: key, Of: ir.Closure}
	return l.callClosureValue(node, closure, l.checker.GetResolvedSignature(node))
}

// Function declarations are in scope before their statement, including declarations
// following the factory's return. They need no runtime function-valued binding:
// each direct call makes the concrete closure with that invocation's cells.
func (l *lowering) nestedGenericFunction(node *ast.Node) ([]ir.Statement, error) {
	if l.generics == nil {
		l.generics = map[*ast.Symbol]*ast.Node{}
	}
	l.generics[l.symbol(node.Name())] = node
	return nil, nil
}

func (l *lowering) registerNestedGenerics(body *ast.Node) {
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			if node.Kind == ast.KindFunctionDeclaration && len(node.TypeParameters()) > 0 {
				_, _ = l.nestedGenericFunction(node)
			}
			return false
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
}

func nestedGenericDeclaration(declaration *ast.Node) bool {
	for parent := declaration.Parent; parent != nil; parent = parent.Parent {
		if ast.IsFunctionLike(parent) {
			return true
		}
	}
	return false
}

func (l *lowering) callNestedGeneric(node *ast.Node, declaration *ast.Node) (ir.Expression, error) {
	signature := l.checker.GetSignatureFromDeclaration(declaration)
	arguments, known := l.genericClosureArguments(node, signature)
	if !known {
		return nil, l.notYet(node, "a nested generic call without concrete type arguments")
	}
	if l.genericDepth >= maximumGenericDepth {
		return nil, l.notYet(node, "a nested generic call instantiated without end")
	}
	closure, err := l.genericClosureInstance(declaration, signature, arguments)
	if err != nil {
		return nil, err
	}
	return l.callClosureValue(node, closure, l.checker.GetResolvedSignature(node))
}
