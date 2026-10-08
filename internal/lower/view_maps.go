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
		contract.Key = l.mapEntrySlot(node, l.concrete(args[0]))
		contract.Element = l.mapEntrySlot(node, l.concrete(args[1]))
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
	key := l.mapEntrySlot(node, l.concrete(args[0]))
	element := l.mapEntrySlot(node, l.concrete(args[1]))
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
	return contract.Unsupported == "" && (contract.Kind == ir.ViewScalar || contract.Kind == ir.ViewObject || contract.Kind == ir.ViewArray || contract.Kind == ir.ViewUnion || contract.Kind == ir.ViewNullable || contract.Kind == ir.ViewNull || contract.Kind == ir.ViewUndefined)
}

func (l *lowering) mapEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	if !l.mapEntryTypeProven(target, map[*checker.Type]bool{}) {
		return 0
	}
	present := l.checker.GetNonNullableType(target)
	array := l.viewArrayBase(present) != nil
	optionalObject := l.includesUndefined(target) && !l.includesNull(target) && present.Flags()&checker.TypeFlagsObject != 0
	if !array && !optionalObject && target.Flags()&checker.TypeFlagsUnion == 0 && target.Flags()&checker.TypeFlagsNull == 0 {
		return l.slotContract(node, target)
	}
	id, err := l.viewContract(node, target)
	if err != nil || !mapEntryDescriptorProven(l.result, id, map[ir.ViewContractID]bool{}) {
		return 0
	}
	return id
}

// A primitive phantom intersection has no runtime nominal witness. Reject it
// throughout entries, including beneath structural fields or array elements.
func (l *lowering) mapEntryTypeProven(target *checker.Type, seen map[*checker.Type]bool) bool {
	target = l.concrete(target)
	if target.Flags()&checker.TypeFlagsIntersection != 0 || isClassInstance(target) {
		return false
	}
	if seen[target] {
		return true
	}
	seen[target] = true
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if !l.mapEntryTypeProven(member, seen) {
				return false
			}
		}
		return true
	}
	if base := l.viewArrayBase(target); base != nil {
		if !l.mapEntryTypeProven(l.checker.GetElementTypeOfArrayType(base), seen) {
			return false
		}
		for _, field := range l.viewArrayOwnProperties(target, base) {
			if !l.mapEntryTypeProven(l.checker.GetTypeOfSymbol(field), seen) {
				return false
			}
		}
	} else if target.Flags()&checker.TypeFlagsObject != 0 {
		for _, field := range l.checker.GetPropertiesOfType(target) {
			if !l.mapEntryTypeProven(l.checker.GetTypeOfSymbol(field), seen) {
				return false
			}
		}
	}
	return true
}

func mapEntryDescriptorProven(program *ir.Program, id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id == 0 {
		return false
	}
	if seen[id] {
		return true
	}
	seen[id] = true
	c := program.ViewContracts[id-1]
	if !mapEntryContract(c) {
		return false
	}
	for _, f := range c.Fields {
		if !mapEntryDescriptorProven(program, f.Contract, seen) {
			return false
		}
	}
	for _, member := range c.Members {
		if !mapEntryDescriptorProven(program, member, seen) {
			return false
		}
	}
	if c.Element != 0 && !mapEntryDescriptorProven(program, c.Element, seen) {
		return false
	}
	return true
}
