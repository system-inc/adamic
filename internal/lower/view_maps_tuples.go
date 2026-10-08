package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// Tuple certificates share object slots and retain their declared arity domain.
func (l *lowering) mapTupleEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	id := l.tupleViewSlot(node, target, func(element *checker.Type) ir.ViewContractID {
		child := l.mapEntrySlot(node, l.concrete(element))
		if tupleCallableEntry(l.result, child, map[ir.ViewContractID]bool{}) {
			return 0
		}
		return child
	})
	if !mapEntryDescriptorProven(l.result, id, map[ir.ViewContractID]bool{}) || tupleCallableEntry(l.result, id, map[ir.ViewContractID]bool{}) {
		return 0
	}
	return id
}

func (l *lowering) tupleViewSlot(node *ast.Node, target *checker.Type, build func(*checker.Type) ir.ViewContractID) ir.ViewContractID {
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	tuple := target.TargetTupleType()
	minimum, variable := 0, false
	for position, flag := range tuple.ElementFlags() {
		if flag == checker.ElementFlagsRequired {
			if variable {
				return 0
			}
			minimum++
		} else if flag == checker.ElementFlagsOptional {
			variable = true
		} else if flag == checker.ElementFlagsRest && position == len(tuple.ElementFlags())-1 {
			variable = true
		} else {
			return 0
		}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		if l.result.ViewContracts[id-1].Unsupported == "tuple building" {
			return 0
		}
		if l.result.ViewContracts[id-1].FixedTuple {
			return id
		}
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewUnknown, Unsupported: "tuple building"})
	l.result.ViewContractTypes[int(target.Id())] = id
	contract := ir.ViewContract{Kind: ir.ViewObject, Of: ir.Object, Name: l.checker.TypeToString(target), ArrayReadonly: tuple.IsReadonly(), FixedTuple: true, TupleVariable: variable, TupleMinimum: minimum}
	for position, element := range l.checker.GetTypeArguments(target) {
		child := build(element)
		if child == 0 {
			return 0
		}
		name := strconv.Itoa(position)
		if tuple.ElementFlags()[position] == checker.ElementFlagsRest {
			contract.TupleRest = child
			contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: name, Contract: child, Optional: true, Readonly: tuple.IsReadonly()})
			l.result.CheckedFields[name] = true
			continue
		}
		contract.Tuple = append(contract.Tuple, child)
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: name, Contract: child, Optional: tuple.ElementFlags()[position] == checker.ElementFlagsOptional, Readonly: tuple.IsReadonly()})
		l.result.CheckedFields[name] = true
	}
	l.result.ViewContracts[id-1] = contract
	l.result.ViewContractTypes[int(target.Id())] = id
	return id
}

func (l *lowering) readTupleViewElement(node *ast.Node, object ir.Expression, name string, of ir.Type) ir.Expression {
	receiver := node.AsElementAccessExpression().Expression
	elements := l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(receiver))
	position, err := strconv.Atoi(name)
	tuple := l.checker.GetTypeAtLocation(receiver).TargetTupleType()
	rest := len(elements) > 0 && tuple.ElementFlags()[len(elements)-1] == checker.ElementFlagsRest
	declaredPosition := position
	if rest && position >= len(elements)-1 {
		declaredPosition = len(elements) - 1
	}
	if err != nil || position < 0 || declaredPosition >= len(elements) {
		return ir.Property{Object: object, Name: name, Of: of}
	}
	if rest && position >= len(elements)-1 {
		if l.result.CheckedFields == nil {
			l.result.CheckedFields = map[string]bool{}
		}
		l.result.CheckedFields[name] = true
	}
	declared := l.concrete(elements[declaredPosition])
	property := ir.Property{Object: object, Name: name, Of: of, Readiness: sourceExpression(node), View: sourceExpression(node), ViewWhere: l.program.Where(node), ViewType: l.checker.TypeToString(declared), ViewTypeID: int(declared.Id()), ViewReceiverTypeID: int(l.checker.GetTypeAtLocation(receiver).Id())}
	property.Absent = tuple.ElementFlags()[declaredPosition] == checker.ElementFlagsOptional || rest && position >= len(elements)-1
	property.ViewAllowed = l.viewLiterals(declared)
	property.ViewContract = l.result.ViewContractTypes[property.ViewTypeID]
	return property
}

func (l *lowering) mapNullableTupleEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	present := l.checker.GetNonNullableType(target)
	child := l.mapTupleEntrySlot(node, present)
	if child == 0 || target == present {
		return child
	}
	of, known := l.representation(target)
	if !known {
		return 0
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewNullable, Of: of, Element: child, Null: l.includesNull(target), Undefined: l.includesUndefined(target), Name: l.checker.TypeToString(target)})
	l.result.ViewContractTypes[int(target.Id())] = id
	return id
}

// Callable producers beneath a tuple need the closure lane's recursive store proof.
func tupleCallableEntry(program *ir.Program, id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id == 0 || seen[id] {
		return false
	}
	seen[id] = true
	contract := program.ViewContracts[id-1]
	if contract.Kind == ir.ViewCallable {
		return true
	}
	if tupleCallableEntry(program, contract.Element, seen) {
		return true
	}
	for _, field := range contract.Fields {
		if tupleCallableEntry(program, field.Contract, seen) {
			return true
		}
	}
	for _, member := range contract.Members {
		if tupleCallableEntry(program, member, seen) {
			return true
		}
	}
	return false
}
