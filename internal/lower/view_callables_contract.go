package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Called only at a reached viewed member read, after its placeholder has been
// interned. Building argument contracts during cast admission would break laziness.
// A zero Result remains the explicit unknown signature until this hook succeeds.
func (l *lowering) completeViewCallableShapeContract(node *ast.Node, target *checker.Type, id ir.ViewContractID, build viewContractBuilder) error {
	signatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(signatures) != 1 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return l.notYet(node, "viewed callable read with overload or construct signatures")
	}
	signature := signatures[0]
	if len(signature.TypeParameters()) != 0 || signature.HasRestParameter() || signature.MinArgumentCount() != len(signature.Parameters()) {
		return l.notYet(node, "viewed callable read with generic, rest or optional parameters")
	}
	contract := l.result.ViewContracts[id-1]
	parameters := make([]ir.ViewContractID, len(signature.Parameters()))
	for index, parameter := range signature.Parameters() {
		child, err := build(l.checker.GetTypeOfSymbol(parameter))
		if err != nil {
			return err
		}
		if child == 0 {
			return l.notYet(node, "viewed callable read with an unknown parameter contract")
		}
		parameters[index] = child
	}
	result, err := build(l.checker.GetReturnTypeOfSignature(signature))
	if err != nil {
		return err
	}
	if result == 0 {
		return l.notYet(node, "viewed callable read with an unknown result contract")
	}
	contract.Parameters, contract.Result = parameters, result
	l.result.ViewContracts[id-1] = contract
	return nil
}
