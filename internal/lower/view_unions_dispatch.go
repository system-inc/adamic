package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// All members are interned before a union can supply a read certificate.
func buildUnionReadContract(
	lowering *lowering,
	node *ast.Node,
	target *checker.Type,
	build viewContractBuilder,
) (ir.ViewContractID, error) {
	if of, known := lowering.representation(target); known && of == ir.Closure {
		return lowering.prepareUntaggedCallableUnionRead(node, target)
	}
	lowering.prepareUntaggedStructuralRead(node, target, map[*checker.Type]bool{})
	id, internMixedUnionViewContractError := internMixedUnionViewContract(lowering, node, target, build)
	if internMixedUnionViewContractError != nil {
		return 0, internMixedUnionViewContractError
	}
	contract := lowering.result.ViewContracts[id-1]
	contract.Undefined = lowering.includesUndefined(target)
	contract.Null = lowering.includesNull(target)
	for _, member := range contract.Members {
		child := lowering.result.ViewContracts[member-1]
		if child.Kind == ir.ViewArray {
			contract.Unsupported = "views-v3: array element kind: " + child.Name
		}
	}
	if contract.Of == ir.Object {
		for _, field := range lowering.checker.GetPropertiesOfType(target) {
			child, buildError := build(lowering.checker.GetTypeOfSymbol(field))
			if buildError != nil {
				return 0, buildError
			}
			contract.Fields = append(contract.Fields, ir.ViewFieldContract{
				Name:     field.Name,
				Contract: child,
				Optional: field.Flags&ast.SymbolFlagsOptional != 0,
				Readonly: lowering.checker.IsReadonlySymbol(field),
			})
		}
		lowering.result.ViewContracts[id-1] = contract
		if contract.Unsupported == "" && !lowering.supportsUntaggedRead(contract) {
			contract.Unsupported = "untagged object union"
		}
	}
	lowering.result.ViewContracts[id-1] = contract
	return id, nil
}

// Union members need an element/position graph even before array consumers
// land. No standalone array consumer or dictionary adapter is registered here.
func (lowering *lowering) unionAggregateContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if id := lowering.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	id := ir.ViewContractID(len(lowering.result.ViewContracts) + 1)
	contract := ir.ViewContract{Name: lowering.checker.TypeToString(target)}
	lowering.result.ViewContractTypes[int(target.Id())] = id
	lowering.result.ViewContracts = append(lowering.result.ViewContracts, contract)
	if checker.IsTupleType(target) {
		contract.Kind = ir.ViewObject
		contract.Of = ir.Object
		contract.FixedTuple = true
		for position, element := range lowering.checker.GetTypeArguments(target) {
			child, viewContractError := lowering.viewContract(node, element)
			if viewContractError != nil {
				return 0, viewContractError
			}
			flags := target.TargetTupleType().ElementFlags()
			if flags[position] != checker.ElementFlagsRequired {
				return 0, lowering.notYet(node, "a variable tuple checked-view contract")
			}
			contract.Tuple = append(contract.Tuple, child)
			contract.Fields = append(contract.Fields, ir.ViewFieldContract{
				Name:     strconv.Itoa(position),
				Contract: child,
				Readonly: target.TargetTupleType().IsReadonly(),
			})
		}
	} else {
		contract.Kind = ir.ViewArray
		contract.Unsupported = "views-v3: array element kind"
		contract.Of = ir.Array
		element := lowering.viewArrayElementType(target)
		if element == nil {
			return 0, lowering.notYet(node, "an unavailable array union element")
		}
		child, viewContractError := lowering.viewContract(node, element)
		if viewContractError != nil {
			return 0, viewContractError
		}
		contract.Element = child
	}
	lowering.result.ViewContracts[id-1] = contract
	return id, nil
}

// Only registered, complete union reads bypass the pre-existing slot refusal.
// The backend must run that certificate before returning the read value.
func (lowering *lowering) unionReadCertificate(node *ast.Node, of ir.Type) bool {
	if of != ir.Union {
		return false
	}
	symbol := lowering.checker.GetSymbolAtLocation(node.Name())
	if symbol == nil {
		return false
	}
	id := lowering.result.ViewContractTypes[int(lowering.checker.GetTypeOfSymbol(symbol).Id())]
	if id <= 0 || int(id) > len(lowering.result.ViewContracts) {
		return false
	}
	contract := lowering.result.ViewContracts[id-1]
	return contract.Kind == ir.ViewUnion && contract.Unsupported == "" && len(contract.Members) != 0
}
