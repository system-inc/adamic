package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Shared union read dispatch interns a complete graph with its member builder.
// Reserve before recursion, but discard every newly reserved id on failure:
// a later lookup must not mistake a partial graph for a complete contract.
func internMixedUnionViewContract(
	lowering *lowering,
	node *ast.Node,
	target *checker.Type,
	build viewContractBuilder,
) (id ir.ViewContractID, internMixedUnionViewContractError error) {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return 0, lowering.notYet(node, "a mixed union contract for a non-union type")
	}
	of, known := lowering.viewRepresentation(target)
	if !known {
		return 0, lowering.notYet(node, "checked union representation for "+lowering.checker.TypeToString(target))
	}
	if build == nil {
		return 0, lowering.notYet(node, "a checked union without a recursive member builder")
	}
	if lowering.result.ViewContractTypes == nil {
		lowering.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if cached := lowering.result.ViewContractTypes[int(target.Id())]; cached != 0 {
		return cached, nil
	}
	start := len(lowering.result.ViewContracts)
	id = ir.ViewContractID(start + 1)
	contract := ir.ViewContract{Kind: ir.ViewUnion, Name: lowering.checker.TypeToString(target), Of: of}
	lowering.result.ViewContracts = append(lowering.result.ViewContracts, contract)
	lowering.result.ViewContractTypes[int(target.Id())] = id
	defer func() {
		if internMixedUnionViewContractError == nil {
			return
		}
		lowering.result.ViewContracts = lowering.result.ViewContracts[:start]
		for key, value := range lowering.result.ViewContractTypes {
			if int(value) > start {
				delete(lowering.result.ViewContractTypes, key)
			}
		}
		id = 0
	}()
	for _, member := range target.Types() {
		var child ir.ViewContractID
		child, internMixedUnionViewContractError = build(member)
		if internMixedUnionViewContractError != nil {
			return 0, internMixedUnionViewContractError
		}
		if child <= 0 ||
			int(child) > len(lowering.result.ViewContracts) ||
			lowering.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown {
			return 0, lowering.notYet(
				node,
				"an unavailable checked union member contract for "+lowering.checker.TypeToString(member),
			)
		}
		contract.Members = append(contract.Members, child)
	}
	lowering.result.ViewContracts[int(id)-1] = contract
	return id, nil
}
