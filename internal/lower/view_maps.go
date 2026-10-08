package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Certificates come from the constructor's declared source schema, never from a
// cast. Complete scalar and structural schemas are authorized as map storage.
func (l *lowering) mapViewContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Kind: ir.ViewMap, Of: ir.Map, Name: l.checker.TypeToString(target), MapReadonly: l.isLibraryType(target, "ReadonlyMap")}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	args := l.typeArguments(target)
	if len(args) != 2 {
		contract.Unsupported = "Map key/value certificate"
	} else {
		if args[0].Flags()&checker.TypeFlagsIntersection != 0 || args[1].Flags()&checker.TypeFlagsIntersection != 0 {
			contract.Unsupported = "Map phantom/intersection certificate"
		}
		contract.Key = l.slotContract(node, l.concrete(args[0]))
		contract.Element = l.slotContract(node, l.concrete(args[1]))
		for _, child := range []ir.ViewContractID{contract.Key, contract.Element} {
			if child == 0 || !mapEntryContract(l.result.ViewContracts[child-1]) {
				contract.Unsupported = "Map key/value certificate"
			}
		}
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}
func (l *lowering) mapProducer(node *ast.Node, value ir.MapNew) ir.MapNew {
	args := l.typeArguments(l.concrete(l.checker.GetTypeAtLocation(node)))
	if len(args) != 2 || args[0].Flags()&checker.TypeFlagsIntersection != 0 || args[1].Flags()&checker.TypeFlagsIntersection != 0 {
		return value
	}
	key := l.slotContract(node, l.concrete(args[0]))
	element := l.slotContract(node, l.concrete(args[1]))
	if key != 0 && element != 0 && mapEntryContract(l.result.ViewContracts[key-1]) && mapEntryContract(l.result.ViewContracts[element-1]) {
		l.result.MapCertificates = append(l.result.MapCertificates, [2]ir.ViewContractID{key, element})
		value.KeyContract = key
		value.ValueContract = element
		value.ContractName = l.checker.TypeToString(l.checker.GetTypeAtLocation(node))
	}
	return value
}

// slotContract has already checked every member recursively. Unknown or lazy
// descriptors cannot certify entries merely because their storage is a pointer.
func mapEntryContract(contract ir.ViewContract) bool {
	return contract.Unsupported == "" && (contract.Kind == ir.ViewScalar || contract.Kind == ir.ViewObject)
}
