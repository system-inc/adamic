package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Context supplies a callable contract, never an erased generic entry. The checker
// derives the substitution; read-back evidence prevents inference defaults from
// choosing unknown for a binder absent from the contextual contract.
func (l *lowering) genericFunctionValue(node, declaration *ast.Node) (ir.Expression, error) {
	context := l.concrete(l.checker.GetContextualType(node, checker.ContextFlagsNone))
	if context == nil {
		return nil, l.unfixedGenericValue(node, declaration, "without a contextual callable slot")
	}
	signatures := l.checker.GetSignaturesOfType(context, checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) != 0 {
		return nil, l.unfixedGenericValue(node, declaration, "without one concrete contextual callable signature")
	}
	target := l.checker.GetSignatureFromDeclaration(declaration)
	if target == nil {
		return nil, l.notYet(node, "a generic function value without a checker signature")
	}
	evidence := map[*checker.Type]*checker.Type{}
	declared, given := target.Parameters(), signatures[0].Parameters()
	for i, parameter := range declared {
		if i < len(given) {
			l.inferTypes(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(given[i]), evidence)
		}
	}
	l.inferTypes(l.checker.GetReturnTypeOfSignature(target), l.checker.GetReturnTypeOfSignature(signatures[0]), evidence)
	for _, parameter := range declaration.TypeParameters() {
		binder := l.checker.GetTypeAtLocation(parameter.Name())
		concrete := evidence[binder]
		if concrete == nil || concrete.Flags()&checker.TypeFlagsTypeParameter != 0 {
			return nil, l.unfixedGenericValue(node, declaration, "with unresolved type parameter "+parameter.Name().Text())
		}
	}
	resolved := checker.Checker_instantiateSignatureInContextOf(l.checker, target, signatures[0], nil, nil)
	instance, err := l.instantiateFunctionSignature(node, declaration, resolved)
	if err != nil {
		return nil, err
	}
	// Module declarations have one source identity regardless of selected call ABI.
	source := l.program.Where(declaration)
	identity := instance + 1
	for _, function := range l.result.Functions {
		if function.GenericSource == source {
			identity = function.SourceIdentity
			break
		}
	}
	l.result.Functions[instance].GenericSource = source
	l.result.Functions[instance].SourceIdentity = identity
	return l.functionValue(node, instance)
}

func (l *lowering) unfixedGenericValue(node, declaration *ast.Node, reason string) error {
	names := []string{}
	for _, parameter := range declaration.TypeParameters() {
		names = append(names, parameter.Name().Text())
	}
	return &Refused{Where: l.program.Where(node), What: "generic function " + declaration.Name().Text() + " as a value " + reason + "; type parameters: " + strings.Join(names, ", "), Fix: "give this value a concrete callable slot fixing every type parameter (adamic/generic-function-values)"}
}

func (l *lowering) functionValueType(node *ast.Node, callee ir.Function) *checker.Type {
	if callee.SourceIdentity != 0 {
		return l.concrete(l.checker.GetContextualType(node, checker.ContextFlagsNone))
	}
	return l.concrete(l.checker.GetTypeAtLocation(node))
}
