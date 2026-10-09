package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Union membership admits only fixed scalar signatures until the callable
// adapter supplies stronger proofs. A physical object tag is never that proof.
func (l *lowering) untaggedCallableShape(target *checker.Type) bool {
	signatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(signatures) != 1 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	signature := signatures[0]
	if len(signature.TypeParameters()) != 0 || signature.HasRestParameter() {
		return false
	}
	for _, parameter := range signature.Parameters() {
		of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || !(of == ir.Number || of == ir.Boolean || of == ir.String || of == ir.MaybeNumber || of == ir.MaybeBoolean) {
			return false
		}
	}
	result := l.checker.GetReturnTypeOfSignature(signature)
	return interfaceScalar(result) || result.Flags()&checker.TypeFlagsVoid != 0
}

func (l *lowering) prepareUntaggedCallableMember(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if !l.untaggedCallableShape(target) {
		return 0, l.notYet(node, "an untagged callable member without a fixed scalar signature")
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	id := l.result.ViewContractTypes[int(target.Id())]
	if id == 0 {
		id = ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{})
		l.result.ViewContractTypes[int(target.Id())] = id
	}
	signature := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)[0]
	contract := ir.ViewContract{Kind: ir.ViewCallable, Of: ir.Closure, Name: l.checker.TypeToString(target)}
	for _, parameter := range signature.Parameters() {
		child, err := l.viewContract(node, l.checker.GetTypeOfSymbol(parameter))
		if err != nil {
			return 0, err
		}
		contract.Parameters = append(contract.Parameters, child)
	}
	result := l.checker.GetReturnTypeOfSignature(signature)
	if result.Flags()&checker.TypeFlagsVoid != 0 {
		contract.DiscardResult = true
	} else {
		child, err := l.viewContract(node, result)
		if err != nil {
			return 0, err
		}
		contract.Result = child
	}
	l.result.ViewContracts[id-1] = contract
	if l.untaggedCallableTargets == nil {
		l.untaggedCallableTargets = map[ir.ViewContractID]*checker.Type{}
	}
	l.untaggedCallableTargets[id] = target
	return id, nil
}
