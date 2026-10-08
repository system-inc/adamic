package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// IDs reuse the checker's structural identity representatives, including generic
// instantiations. Proofs are directional: a broad value never proves a narrow slot.
func (l *lowering) contractTypeID(declared *checker.Type) int {
	for index, representative := range l.instantiated {
		if identicalTypes(l.checker, declared, representative) {
			return index + 1
		}
	}
	l.instantiated = append(l.instantiated, declared)
	return len(l.instantiated)
}

func (l *lowering) referenceContract(contract *ir.FieldContract, declared *checker.Type) {
	contract.TypeID = l.contractTypeID(declared)
	contract.Reference = contract.Kind == ir.Object || contract.Kind == ir.Array || contract.Kind == ir.Map
	l.result.WriteContracts = append(l.result.WriteContracts, contract)
	// Both allocation-field proofs and present incoming-write proofs are needed.
	// Present writes have already checked nullability; allocation fields have not.
	l.refreshContractProofs(contract)
	for _, expected := range l.result.WriteContracts {
		if expected != contract {
			l.addContractProof(expected, contract.TypeID)
		}
	}
	if contract.Kind == ir.Object {
		contract.Fields, contract.Structural = l.referenceFields(l.checker.GetNonNullableType(declared), map[*checker.Type]bool{}, 0)
	}
}

func (l *lowering) addContractProof(contract *ir.FieldContract, source int) {
	from, to := l.instantiated[source-1], l.instantiated[contract.TypeID-1]
	if l.checker.IsTypeAssignableTo(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil && !slices.Contains(contract.ProvenFields, source) {
		contract.ProvenFields = append(contract.ProvenFields, source)
	}
	a, b := l.checker.GetNonNullableType(from), l.checker.GetNonNullableType(to)
	if l.checker.IsTypeAssignableTo(a, b) && l.widened(a, b, map[[2]*checker.Type]bool{}) == nil && !slices.Contains(contract.ProvenWrites, source) {
		contract.ProvenWrites = append(contract.ProvenWrites, source)
	}
	slices.Sort(contract.ProvenFields)
	slices.Sort(contract.ProvenWrites)
}

func (l *lowering) refreshContractProofs(contract *ir.FieldContract) {
	for source := range l.instantiated {
		l.addContractProof(contract, source+1)
	}
}

// Complex reference types use the directional proofs above. The fallback checks
// finite plain allocation shapes, never invokes getters, and does not reinterpret
// current broad payloads as lifetime guarantees for a narrow referenced object.
func (l *lowering) referenceFields(declared *checker.Type, seen map[*checker.Type]bool, depth int) ([]ir.ContractField, bool) {
	declared = l.contractType(declared)
	if depth >= 8 || seen[declared] || !l.structured(declared) || len(l.containers(declared)) != 0 || l.callableViewContract(declared) || isClassInstance(declared) {
		return nil, false
	}
	seen[declared] = true
	defer delete(seen, declared)
	fields := []ir.ContractField{}
	for _, field := range l.checker.GetPropertiesOfType(declared) {
		own := l.contractType(l.checker.GetTypeOfSymbol(field))
		kind, known := l.representation(own)
		if !known || accessorSymbol(field) || field.Flags&ast.SymbolFlagsMethod != 0 || (kind != ir.Number && kind != ir.Boolean && kind != ir.String && kind != ir.MaybeNumber && kind != ir.Object) {
			return nil, false
		}
		child := &ir.FieldContract{Kind: kind, Declared: l.checker.TypeToString(own), Nullable: l.includesUndefined(own) || l.includesNull(own), TypeID: l.contractTypeID(own)}
		// The fallback uses actual allocation field types, not snapshot value narrowing.
		l.result.WriteContracts = append(l.result.WriteContracts, child)
		l.refreshContractProofs(child)
		for _, expected := range l.result.WriteContracts {
			if expected != child {
				l.addContractProof(expected, child.TypeID)
			}
		}
		if kind == ir.Object {
			var supported bool
			child.Fields, supported = l.referenceFields(l.checker.GetNonNullableType(own), seen, depth+1)
			if !supported {
				return nil, false
			}
		}
		fields = append(fields, ir.ContractField{Name: field.Name, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Contract: child})
	}
	return fields, true
}

func (l *lowering) checkedWriteType(target, value *ast.Node) int {
	if len(l.result.CheckedWrites) == 0 && !l.result.CheckedElements {
		return 0
	}
	declared := l.contractType(l.checker.GetTypeAtLocation(value))
	source := l.contractTypeID(declared)
	for _, contract := range l.result.WriteContracts {
		l.addContractProof(contract, source)
	}
	return source
}

func (l *lowering) allocationContractType(node *ast.Node) int {
	declared := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if declared == nil {
		declared = l.checker.GetTypeAtLocation(node)
	}
	source := l.contractTypeID(l.checker.GetNonNullableType(l.contractType(declared)))
	for _, contract := range l.result.WriteContracts {
		l.addContractProof(contract, source)
	}
	return source
}
