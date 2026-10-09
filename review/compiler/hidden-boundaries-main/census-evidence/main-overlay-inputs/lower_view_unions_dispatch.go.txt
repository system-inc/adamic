package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// All members are interned before a union can supply a read certificate.
func buildUnionReadContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (ir.ViewContractID, error) {
	if of, known := l.representation(target); known && of == ir.Closure {
		return l.prepareUntaggedCallableUnionRead(node, target)
	}
	l.prepareUntaggedStructuralRead(node, target, map[*checker.Type]bool{})
	id, err := internMixedUnionViewContract(l, node, target, build)
	if err != nil {
		return 0, err
	}
	contract := l.result.ViewContracts[id-1]
	contract.Undefined = l.includesUndefined(target)
	contract.Null = l.includesNull(target)
	for _, member := range contract.Members {
		child := l.result.ViewContracts[member-1]
		if child.Kind == ir.ViewArray {
			contract.Unsupported = "views-v3: array element kind: " + child.Name
		}
	}
	if contract.Of == ir.Object {
		for _, field := range l.checker.GetPropertiesOfType(target) {
			child, err := build(l.checker.GetTypeOfSymbol(field))
			if err != nil {
				return 0, err
			}
			contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: field.Name, Contract: child, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(field)})
		}
		l.result.ViewContracts[id-1] = contract
		if contract.Unsupported == "" && !l.supportsUntaggedRead(contract) {
			contract.Unsupported = "untagged object union"
		}
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}

// Union members need an element/position graph even before array consumers
// land. No standalone array consumer or dictionary adapter is registered here.
func (l *lowering) unionAggregateContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Name: l.checker.TypeToString(target)}
	l.result.ViewContractTypes[int(target.Id())] = id
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	if checker.IsTupleType(target) {
		contract.Kind = ir.ViewObject
		contract.Of = ir.Object
		contract.FixedTuple = true
		for position, element := range l.checker.GetTypeArguments(target) {
			child, err := l.viewContract(node, element)
			if err != nil {
				return 0, err
			}
			flags := target.TargetTupleType().ElementFlags()
			if flags[position] != checker.ElementFlagsRequired {
				return 0, l.notYet(node, "a variable tuple checked-view contract")
			}
			contract.Tuple = append(contract.Tuple, child)
			contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: strconv.Itoa(position), Contract: child, Readonly: target.TargetTupleType().IsReadonly()})
		}
	} else {
		contract.Kind = ir.ViewArray
		contract.Unsupported = "views-v3: array element kind"
		contract.Of = ir.Array
		element := l.viewArrayElementType(target)
		if element == nil {
			return 0, l.notYet(node, "an unavailable array union element")
		}
		child, err := l.viewContract(node, element)
		if err != nil {
			return 0, err
		}
		contract.Element = child
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}

// Only registered, complete union reads bypass the pre-existing slot refusal.
// The backend must run that certificate before returning the read value.
func (l *lowering) unionReadCertificate(node *ast.Node, of ir.Type) bool {
	if of != ir.Union {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node.Name())
	if symbol == nil {
		return false
	}
	id := l.result.ViewContractTypes[int(l.checker.GetTypeOfSymbol(symbol).Id())]
	if id <= 0 || int(id) > len(l.result.ViewContracts) {
		return false
	}
	contract := l.result.ViewContracts[id-1]
	return contract.Kind == ir.ViewUnion && contract.Unsupported == "" && len(contract.Members) != 0
}
