package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Certify data classes at the entry root. Nested nominal stores and callable
// members need recursive producer checks and remain outside this certificate.
func (l *lowering) mapNominalEntryTypeProven(target *checker.Type, seen map[*checker.Type]bool) bool {
	present := l.checker.GetNonNullableType(target)
	if len(seen) != 0 || l.classNodeFor(present) == nil {
		return false
	}
	seen[present] = true
	for _, field := range l.checker.GetPropertiesOfType(present) {
		child := l.concrete(l.checker.GetTypeOfSymbol(field))
		if len(l.checker.GetSignaturesOfType(child, checker.SignatureKindCall)) != 0 || !l.mapEntryTypeProven(child, seen) {
			return false
		}
	}
	return true
}

func (l *lowering) mapNominalEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	if !l.mapNominalEntryTypeProven(target, map[*checker.Type]bool{}) {
		return 0
	}
	present := l.checker.GetNonNullableType(target)
	declaration := l.classNodeFor(present)
	name := l.checker.TypeToString(present)
	nominal := l.program.Where(declaration) + ":" + name
	id := ir.ViewContractID(0)
	for index, contract := range l.result.ViewContracts {
		if contract.Nominal == nominal && contract.NominalClass != 0 {
			id = ir.ViewContractID(index + 1)
			break
		}
	}
	if id == 0 {
		contract := ir.ViewContract{Kind: ir.ViewObject, Of: ir.Object, Name: name, Nominal: nominal, NominalBases: l.viewNominalBases(present, map[*checker.Type]bool{}), NominalClass: l.classIdentity(declaration)}
		if l.result.CheckedFields == nil {
			l.result.CheckedFields = map[string]bool{}
		}
		for _, field := range l.checker.GetPropertiesOfType(present) {
			child := l.mapEntrySlot(node, l.concrete(l.checker.GetTypeOfSymbol(field)))
			if child == 0 {
				return 0
			}
			contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: field.Name, Contract: child, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(field)})
			l.result.CheckedFields[field.Name] = true
		}
		l.result.ViewContracts = append(l.result.ViewContracts, contract)
		id = ir.ViewContractID(len(l.result.ViewContracts))
	}
	if target == present {
		return id
	}
	of, known := l.representation(target)
	if !known {
		return 0
	}
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewNullable, Of: of, Element: id, Null: l.includesNull(target), Undefined: l.includesUndefined(target), Name: l.checker.TypeToString(target)})
	return ir.ViewContractID(len(l.result.ViewContracts))
}
