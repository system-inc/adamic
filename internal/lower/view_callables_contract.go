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
	if len(signatures) == 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return l.notYet(node, "viewed callable read with no call signature or construct signatures")
	}
	for _, signature := range signatures {
		if len(signature.TypeParameters()) != 0 || signature.HasRestParameter() {
			return l.notYet(node, "viewed callable read with generic or rest parameters")
		}
	}
	if len(signatures) > 1 {
		members := make([]ir.ViewContractID, 0, len(signatures))
		for _, signature := range signatures {
			member := ir.ViewContractID(len(l.result.ViewContracts) + 1)
			l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewCallable, Of: ir.Closure, Name: l.checker.TypeToString(target)})
			if err := l.completeViewCallableSignature(node, signature, member, build); err != nil {
				return err
			}
			members = append(members, member)
		}
		contract := l.result.ViewContracts[id-1]
		contract.Members = members
		contract.Parameters = l.result.ViewContracts[members[0]-1].Parameters
		contract.Result = l.result.ViewContracts[members[0]-1].Result
		l.result.ViewContracts[id-1] = contract
		return nil
	}
	return l.completeViewCallableSignature(node, signatures[0], id, build)
}

func (l *lowering) completeViewCallableSignature(node *ast.Node, signature *checker.Signature, id ir.ViewContractID, build viewContractBuilder) error {
	if len(signature.TypeParameters()) != 0 || signature.HasRestParameter() {
		return l.notYet(node, "viewed callable read with generic or rest parameters")
	}
	contract := l.result.ViewContracts[id-1]
	parameters := make([]ir.ViewContractID, len(signature.Parameters()))
	for index, parameter := range signature.Parameters() {
		child, err := build(l.censusCallableParameterType(parameter))
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
