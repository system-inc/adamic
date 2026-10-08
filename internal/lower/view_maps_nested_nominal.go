package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) mapNestedNominalType(target *checker.Type, seen map[*checker.Type]bool) bool {
	if seen[target] {
		return false
	}
	seen[target] = true
	if isClassInstance(l.checker.GetNonNullableType(target)) {
		return true
	}
	if target.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range target.Types() {
			if l.mapNestedNominalType(member, seen) {
				return true
			}
		}
	}
	if target.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if l.viewArrayBase(target) != nil || checker.IsTupleType(target) {
		for _, child := range l.checker.GetTypeArguments(target) {
			if l.mapNestedNominalType(child, seen) {
				return true
			}
		}
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if l.mapNestedNominalType(l.checker.GetTypeOfSymbol(field), seen) {
			return true
		}
	}
	return false
}

// Finite structural paths are certified at the producer. Every class field
// read separately checks identity, including after mutation through an alias.
// Fixed tuples use their existing adapter; readonly object cycles use a visited witness.
func (l *lowering) mapNestedNominalEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	if !l.mapNominalPathAcyclic(target, map[*checker.Type]bool{}) && !l.recursiveNominalObjectPath(target, map[*checker.Type]bool{}) {
		return 0
	}
	if l.result.NominalEntryContracts == nil {
		l.result.NominalEntryContracts = map[int]ir.ViewContractID{}
	}
	if id := l.result.NominalEntryContracts[int(target.Id())]; id != 0 {
		return id
	}
	present := l.checker.GetNonNullableType(target)
	if present != target {
		of, known := l.representation(target)
		if !known {
			return 0
		}
		child := l.mapNestedNominalEntrySlot(node, present)
		if child == 0 {
			return 0
		}
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewNullable, Of: of, Element: child, Null: l.includesNull(target), Undefined: l.includesUndefined(target), Name: l.checker.TypeToString(target)})
		return ir.ViewContractID(len(l.result.ViewContracts))
	}
	if l.viewArrayBase(target) != nil {
		return l.mapNominalArrayEntrySlot(node, target)
	}
	if target.Flags()&checker.TypeFlagsObject == 0 || checker.IsTupleType(target) || len(l.checker.GetIndexInfosOfType(target)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 {
		return 0
	}
	contract := ir.ViewContract{Kind: ir.ViewObject, Of: ir.Object, Name: l.checker.TypeToString(target), Unsupported: "pending nominal producer graph"}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.NominalEntryContracts[int(target.Id())] = id
	for _, field := range l.checker.GetPropertiesOfType(target) {
		child := l.mapEntrySlot(node, l.concrete(l.checker.GetTypeOfSymbol(field)))
		if child == 0 {
			return 0
		}
		if field.Flags&ast.SymbolFlagsOptional != 0 && !l.checker.IsReadonlySymbol(field) {
			// Present class/undefined slots share the checked source-write adapter.
			// Other optional aggregate storage still needs its own conversion proof.
			descriptor := l.result.ViewContracts[child-1]
			if descriptor.Kind == ir.ViewNullable && descriptor.Element != 0 {
				descriptor = l.result.ViewContracts[descriptor.Element-1]
			}
			if (l.result.ViewContracts[child-1].Of != ir.Object && l.result.ViewContracts[child-1].Of != ir.Union) || descriptor.NominalClass == 0 {
				return 0
			}
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: field.Name, Contract: child, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(field)})
		// A readonly schema can describe a mutable allocation held by another alias.
		// Retain source certificates for all writes to these class field names.
		if ir.HasMapNominalWitness(l.result, child) && (l.result.ViewContracts[child-1].Of == ir.Object || l.result.ViewContracts[child-1].Of == ir.Union) {
			l.optionalViewWriteField(field.Name)
		}
		if l.result.CheckedFields == nil {
			l.result.CheckedFields = map[string]bool{}
		}
		l.result.CheckedFields[field.Name] = true
	}
	contract.Unsupported = ""
	l.result.ViewContracts[id-1] = contract
	return id
}

// Finite private producer paths keep their existing adapters. Readonly object
// cycles separately require the runtime visited witness.
func (l *lowering) mapNominalPathAcyclic(target *checker.Type, active map[*checker.Type]bool) bool {
	if active[target] {
		return false
	}
	active[target] = true
	defer delete(active, target)
	if target.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range target.Types() {
			if !l.mapNominalPathAcyclic(member, active) {
				return false
			}
		}
		return true
	}
	if target.Flags()&checker.TypeFlagsObject == 0 {
		return true
	}
	if l.viewArrayBase(target) != nil || checker.IsTupleType(target) {
		for _, child := range l.checker.GetTypeArguments(target) {
			if !l.mapNominalPathAcyclic(child, active) {
				return false
			}
		}
		return true
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		child := l.checker.GetTypeOfSymbol(field)
		if len(l.checker.GetSignaturesOfType(child, checker.SignatureKindCall)) != 0 {
			return false
		}
		if !l.mapNominalPathAcyclic(child, active) {
			return false
		}
	}
	return true
}

func (l *lowering) mapNominalArrayEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	element := l.viewArrayElementType(target)
	if element == nil {
		return 0
	}
	child := l.mapEntrySlot(node, l.concrete(element))
	if child == 0 {
		return 0
	}
	of := l.result.ViewContracts[child-1].Of
	if of != ir.Object && of != ir.Array && !(of == ir.Union && isClassInstance(l.checker.GetNonNullableType(element))) {
		return 0
	}
	contract := ir.ViewContract{Kind: ir.ViewArray, Of: ir.Array, Name: l.checker.TypeToString(target), Element: child, ArrayReadonly: l.isLibraryType(l.viewArrayBase(target), "ReadonlyArray")}
	for _, field := range l.viewArrayOwnProperties(target, l.viewArrayBase(target)) {
		declared := l.concrete(l.checker.GetTypeOfSymbol(field))
		if l.mapNestedNominalType(declared, map[*checker.Type]bool{}) {
			return 0
		}
		slot := l.mapEntrySlot(node, declared)
		if slot == 0 {
			return 0
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: field.Name, Contract: slot, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(field)})
	}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	return ir.ViewContractID(len(l.result.ViewContracts))
}

// The visited-witness adapter currently covers recursive readonly object paths.
// Recursive arrays, mutable aggregate edges and recursive class declarations
// retain their existing refusals until their storage producers have this proof.
func (l *lowering) recursiveNominalObjectPath(target *checker.Type, seen map[*checker.Type]bool) bool {
	target = l.concrete(target)
	if seen[target] {
		return true
	}
	seen[target] = true
	present := l.checker.GetNonNullableType(target)
	if present != target {
		return l.recursiveNominalObjectPath(present, seen)
	}
	if target.Flags()&checker.TypeFlagsObject == 0 {
		return target.Flags()&checker.TypeFlagsUnion == 0
	}
	if l.viewArrayBase(target) != nil || checker.IsTupleType(target) {
		return !l.mapNestedNominalType(target, map[*checker.Type]bool{})
	}
	if len(l.checker.GetIndexInfosOfType(target)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 {
		return false
	}
	nominal := isClassInstance(target)
	if nominal && !l.mapNominalEntryTypeProven(target, map[*checker.Type]bool{}) {
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if !nominal && !l.checker.IsReadonlySymbol(field) {
			return false
		}
		child := l.concrete(l.checker.GetTypeOfSymbol(field))
		if !l.recursiveNominalObjectPath(child, seen) {
			return false
		}
	}
	return true
}
