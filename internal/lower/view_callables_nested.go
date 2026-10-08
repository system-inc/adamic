package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// One nested descriptor is supported here. Recursive higher-order signatures,
// generic callbacks and callback-valued results remain explicit frontiers.
func (l *lowering) simpleNestedCallableShape(proven *checker.Type) bool {
	if l.includesUndefined(proven) {
		return false
	}
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) != 0 || signatures[0].HasRestParameter() || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	for _, parameter := range signatures[0].Parameters() {
		child := l.censusCallableParameterType(parameter)
		of, known := l.representation(child)
		if !known || of == ir.Closure || !l.viewCallableBoxedRepresentation(child) {
			return false
		}
	}
	result := l.checker.GetReturnTypeOfSignature(signatures[0])
	if result.Flags()&checker.TypeFlagsVoid != 0 {
		return true
	}
	of, known := l.representation(result)
	return known && of != ir.Closure && l.viewCallableBoxedRepresentation(result)
}

func (l *lowering) nestedCallableArguments(call *ast.Node) ([]ir.ViewContractID, error) {
	resolved := l.checker.GetResolvedSignature(call)
	if resolved == nil {
		return nil, nil
	}
	contracts := make([]ir.ViewContractID, len(call.AsCallExpression().Arguments.Nodes))
	active := false
	for index, parameter := range resolved.Parameters() {
		if index >= len(contracts) {
			break
		}
		proven := l.censusCallableParameterType(parameter)
		of, known := l.representation(proven)
		if !known || of != ir.Closure {
			continue
		}
		if !l.simpleNestedCallableShape(proven) {
			return nil, l.notYet(call, "a nested callable argument without a complete descriptor")
		}
		id, err := l.prepareViewCallableRead(call, proven)
		if err != nil {
			return nil, err
		}
		contracts[index] = id
		active = true
	}
	if !active {
		return nil, nil
	}
	return contracts, nil
}
