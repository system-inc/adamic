package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func init() { viewArrayContractHook = internArrayViewContract }

// Adapt lane 2's element callback into the shared logical contract graph. Reserve
// the id first: arrays and interfaces can contain each other recursively.
func internArrayViewContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (ir.ViewContractID, error) {
	if checker.IsTupleType(target) {
		var childError error
		id := l.tupleViewSlot(node, target, func(element *checker.Type) ir.ViewContractID {
			id, err := build(element)
			if err != nil {
				childError = err
				return 0
			}
			return id
		})
		if childError != nil {
			return 0, childError
		}
		if id == 0 {
			return 0, l.notYet(node, "a checked tuple view with optional or rest positions")
		}
		return id, nil
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Kind: ir.ViewArray, Of: ir.Array, Name: l.checker.TypeToString(target), ArrayReadonly: l.isLibraryType(l.viewArrayBase(target), "ReadonlyArray")}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	handled, err := l.viewArrayContract(node, target, func(element *checker.Type) error {
		var childError error
		contract.Element, childError = build(element)
		if childError == nil && l.callableViewContract(element) {
			l.result.ViewContracts[contract.Element-1].Unsupported = "callable"
		}
		return childError
	})
	if err != nil {
		return 0, err
	}
	if !handled {
		return 0, l.notYet(node, "an array checked view without an element contract")
	}
	base := l.viewArrayBase(target)
	for _, property := range l.viewArrayOwnProperties(target, base) {
		child, err := build(l.checker.GetTypeOfSymbol(property))
		if err != nil {
			return 0, err
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: property.Name, Contract: child, Optional: property.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(property)})
	}
	l.result.ViewContracts[int(id)-1] = contract
	return id, nil
}

// Optional arrays keep the array descriptor and its element edge. Undefined is
// a field-presence alternative, not an object union requiring a discriminant.
func (l *lowering) viewOptionalArrayContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, bool, error) {
	if !l.includesUndefined(target) || l.viewArrayBase(l.checker.GetNonNullableType(target)) == nil {
		return 0, false, nil
	}
	present, err := l.viewContract(node, l.checker.GetNonNullableType(target))
	if err != nil {
		return 0, true, err
	}
	contract := l.result.ViewContracts[present-1]
	contract.Undefined = true
	contract.Name = l.checker.TypeToString(target)
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	return id, true, nil
}
