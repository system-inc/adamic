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

// Only immutable structural paths are certified here. Fixed tuples use their
// existing adapter. Array, nullable aggregate and mutable paths retain a named
// read refusal until their stores carry the same recursive witness.
func (l *lowering) mapNestedNominalEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	if !l.mapNominalPathAcyclic(target, map[*checker.Type]bool{}) {
		return 0
	}
	if target.Flags()&checker.TypeFlagsObject == 0 || l.viewArrayBase(target) != nil || checker.IsTupleType(target) || len(l.checker.GetIndexInfosOfType(target)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 {
		return 0
	}
	contract := ir.ViewContract{Kind: ir.ViewObject, Of: ir.Object, Name: l.checker.TypeToString(target)}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if !l.checker.IsReadonlySymbol(field) || field.Flags&ast.SymbolFlagsOptional != 0 {
			return 0
		}
		child := l.mapEntrySlot(node, l.concrete(l.checker.GetTypeOfSymbol(field)))
		if child == 0 {
			return 0
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: field.Name, Contract: child, Readonly: true})
		if l.result.CheckedFields == nil {
			l.result.CheckedFields = map[string]bool{}
		}
		l.result.CheckedFields[field.Name] = true
	}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	return ir.ViewContractID(len(l.result.ViewContracts))
}

// Private producer descriptors are finite trees. Recursive aggregate paths must
// acquire a runtime visited witness before they can certify a nested class.
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
