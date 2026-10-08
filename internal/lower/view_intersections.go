package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The shared dispatcher must store these ids as a conjunction, never a union.
// This helper does not enable cast admission or certify any runtime payload.
func (l *lowering) viewIntersectionContracts(node *ast.Node, target *checker.Type, build viewContractBuilder) ([]ir.ViewContractID, error) {
	if target.Flags()&checker.TypeFlagsIntersection == 0 || l.phantomBase(target) != nil {
		return nil, l.notYet(node, "a nonprimitive intersection checked-view contract")
	}
	if build == nil {
		return nil, l.notYet(node, "an intersection checked view without a member builder")
	}
	var members []ir.ViewContractID
	for _, part := range target.Types() {
		if l.viewIntersectionPhantom(part) {
			continue
		}
		child, err := build(part)
		if err != nil {
			return nil, err
		}
		if child <= 0 || int(child) > len(l.result.ViewContracts) || l.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown {
			return nil, l.notYet(node, "an unavailable intersection member contract for "+l.checker.TypeToString(part))
		}
		members = append(members, child)
	}
	if len(members) == 0 {
		return nil, l.notYet(node, "an intersection without a runtime constituent")
	}
	return members, nil
}

// Reuse the brand lane's field rule. An empty object is not evidence of a brand,
// and callable/indexed objects must retain their runtime obligations.
func (l *lowering) viewIntersectionPhantom(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsObject == 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return false
	}
	fields := l.checker.GetPropertiesOfType(target)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) {
			return false
		}
	}
	return true
}

// Structural intersections share object storage, but retain every constituent's
// read obligations. Checker properties resolve duplicate fields by intersection,
// so their child descriptor retains the conjunction rather than picking an arm.
func (l *lowering) structuralViewIntersection(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsIntersection == 0 || l.phantomBase(target) != nil {
		return false
	}
	runtimeParts := 0
	for _, part := range target.Types() {
		if l.viewIntersectionPhantom(part) {
			continue
		}
		runtimeParts++
		if part.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(part) || l.checker.IsArrayType(part) || checker.IsTupleType(part) || l.callableViewContract(part) || len(l.checker.GetIndexInfosOfType(part)) != 0 {
			return false
		}
	}
	return runtimeParts > 0
}

// The lazy dispatcher reserves this descriptor before walking descendants.
// Unsupported children remain obligations at their own reads, never cast gates.
func (l *lowering) internStructuralViewIntersection(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if !l.structuralViewIntersection(target) {
		return 0, l.notYet(node, "a structural object intersection view")
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Intersection: true, Kind: ir.ViewObject, Of: ir.Object, Name: l.checker.TypeToString(target)}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	runtimeFields := map[string]bool{}
	for _, part := range target.Types() {
		if l.viewIntersectionPhantom(part) {
			continue
		}
		child, err := l.viewContract(node, part)
		if err != nil {
			return 0, err
		}
		contract.Members = append(contract.Members, child)
		for _, field := range l.checker.GetPropertiesOfType(part) {
			runtimeFields[field.Name] = true
		}
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if !runtimeFields[field.Name] {
			continue
		}
		child, err := l.viewContract(node, l.checker.GetTypeOfSymbol(field))
		if err != nil {
			return 0, err
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: field.Name, Contract: child, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(field)})
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}
