package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The integrator wires this hook into lazy dictionary classification. Descriptors
// remain unsupported until producer and indexed-read dispatch exist together.
var viewDictionaryContractHook viewContractHook = internDictionaryViewContract

func internDictionaryViewContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (ir.ViewContractID, error) {
	if l.checker.IsArrayType(target) || checker.IsTupleType(target) || build == nil {
		return 0, l.notYet(node, "a string-key dictionary contract")
	}
	var element *checker.Type
	for _, index := range l.checker.GetIndexInfosOfType(target) {
		if index.KeyType().Flags()&checker.TypeFlagsString != 0 {
			element = index.ValueType()
		} else {
			// Numeric/template/symbol index constraints cannot silently disappear.
			return 0, l.notYet(node, "a dictionary with a non-string index contract")
		}
	}
	if element == nil {
		return 0, l.notYet(node, "a dictionary without a string index contract")
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	before := len(l.result.ViewContracts)
	id := ir.ViewContractID(before + 1)
	contract := ir.ViewContract{Kind: ir.ViewUnknown, Of: ir.Object, Name: l.checker.TypeToString(target), Unsupported: "dictionary source dispatch"}
	// Reserve before building children: recursive record/object graphs share ids.
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	rollback := func(err error) (ir.ViewContractID, error) {
		l.result.ViewContracts = l.result.ViewContracts[:before]
		for key, child := range l.result.ViewContractTypes {
			if int(child) > before {
				delete(l.result.ViewContractTypes, key)
			}
		}
		return 0, err
	}
	var err error
	contract.Element, err = build(element)
	if err != nil {
		return rollback(err)
	}
	if contract.Element <= 0 || int(contract.Element) > len(l.result.ViewContracts) {
		return rollback(l.notYet(node, "an unavailable dictionary element contract"))
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		child, err := build(l.checker.GetTypeOfSymbol(property))
		if err != nil {
			return rollback(err)
		}
		if child <= 0 || int(child) > len(l.result.ViewContracts) {
			return rollback(l.notYet(node, "an unavailable dictionary named field contract"))
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: property.Name, Contract: child, Optional: property.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(property)})
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}
