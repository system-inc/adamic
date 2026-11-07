package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Union membership admits only fixed scalar signatures until the callable
// adapter supplies stronger proofs. A physical object tag is never that proof.
func (lowering *lowering) untaggedCallableShape(target *checker.Type) bool {
	signatures := lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(signatures) != 1 || len(lowering.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	signature := signatures[0]
	if len(signature.TypeParameters()) != 0 || signature.HasRestParameter() {
		return false
	}
	for _, parameter := range signature.Parameters() {
		of, known := lowering.representation(lowering.checker.GetTypeOfSymbol(parameter))
		if of == ir.Union && known {
			primitive := true
			for _, member := range lowering.checker.GetTypeOfSymbol(parameter).Types() {
				primitive = primitive && interfaceScalar(member)
			}
			if primitive {
				continue
			}
		}
		if !known ||
			!(of == ir.Number || of == ir.Boolean || of == ir.String || of == ir.MaybeNumber || of == ir.MaybeBoolean) {
			return false
		}
	}
	result := lowering.checker.GetReturnTypeOfSignature(signature)
	return interfaceScalar(result) || result.Flags()&checker.TypeFlagsVoid != 0
}

func (lowering *lowering) prepareUntaggedCallableMember(
	node *ast.Node,
	target *checker.Type,
) (ir.ViewContractID, error) {
	if !lowering.untaggedCallableShape(target) {
		return 0, lowering.notYet(node, "an untagged callable member without a fixed scalar signature")
	}
	if lowering.result.ViewContractTypes == nil {
		lowering.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	id := lowering.result.ViewContractTypes[int(target.Id())]
	if id == 0 {
		id = ir.ViewContractID(len(lowering.result.ViewContracts) + 1)
		lowering.result.ViewContracts = append(lowering.result.ViewContracts, ir.ViewContract{})
		lowering.result.ViewContractTypes[int(target.Id())] = id
	}
	signature := lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall)[0]
	contract := ir.ViewContract{Kind: ir.ViewCallable, Of: ir.Closure, Name: lowering.checker.TypeToString(target)}
	for _, parameter := range signature.Parameters() {
		child, viewContractError := lowering.viewContract(node, lowering.checker.GetTypeOfSymbol(parameter))
		if viewContractError != nil {
			return 0, viewContractError
		}
		if members, ok := ir.PrimitiveViewMembers(lowering.result, child); ok {
			for _, member := range members {
				lowering.result.ViewContracts[child-1].RepresentationMask |= 1 << member.Of
			}
		}
		contract.Parameters = append(contract.Parameters, child)
	}
	result := lowering.checker.GetReturnTypeOfSignature(signature)
	if result.Flags()&checker.TypeFlagsVoid != 0 {
		contract.DiscardResult = true
	} else {
		child, viewContractError := lowering.viewContract(node, result)
		if viewContractError != nil {
			return 0, viewContractError
		}
		contract.Result = child
	}
	lowering.result.ViewContracts[id-1] = contract
	if lowering.untaggedCallableTargets == nil {
		lowering.untaggedCallableTargets = map[ir.ViewContractID]*checker.Type{}
	}
	lowering.untaggedCallableTargets[id] = target
	return id, nil
}

// This is descriptor construction, not an array consumer admission.
func (lowering *lowering) viewArrayElementType(target *checker.Type) *checker.Type {
	if !lowering.isArrayType(target) {
		return nil
	}
	return lowering.elementTypeOfArray(target)
}
