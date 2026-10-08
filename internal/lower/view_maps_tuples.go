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

func (l *lowering) readTupleViewElement(node *ast.Node, object ir.Expression, name string, of ir.Type) (ir.Expression, error) {
	receiver := node.AsElementAccessExpression().Expression
	receiverType := l.checker.GetTypeAtLocation(receiver)
	position, err := strconv.Atoi(name)
	if err != nil || position < 0 {
		return ir.Property{Object: object, Name: name, Of: of}, nil
	}
	var declared *checker.Type
	absent := false
	alternatives := tupleAlternativesType(receiverType)
	tail := false
	if alternatives {
		declared = l.concrete(l.checker.GetTypeAtLocation(node))
		for _, member := range receiverType.Types() {
			flags := member.TargetTupleType().ElementFlags()
			absent = absent || position >= len(flags)
			if position < len(flags) {
				absent = absent || flags[position] == checker.ElementFlagsOptional
			}
		}
	} else {
		elements := l.checker.GetTypeArguments(receiverType)
		flags := receiverType.TargetTupleType().ElementFlags()
		rest := len(elements) > 0 && flags[len(elements)-1] == checker.ElementFlagsRest
		declaredPosition := position
		if rest && position >= len(elements)-1 {
			declaredPosition = len(elements) - 1
		}
		if declaredPosition >= len(elements) {
			return ir.Property{Object: object, Name: name, Of: of}, nil
		}
		declared = l.concrete(elements[declaredPosition])
		tail = rest && position >= len(elements)-1
		absent = flags[declaredPosition] == checker.ElementFlagsOptional || tail
	}
	if alternatives || tail {
		if l.result.CheckedFields == nil {
			l.result.CheckedFields = map[string]bool{}
		}
		l.result.CheckedFields[name] = true
	}
	property := ir.Property{Object: object, Name: name, Of: of, Absent: absent, Readiness: sourceExpression(node), View: sourceExpression(node), ViewWhere: l.program.Where(node), ViewType: l.checker.TypeToString(declared), ViewTypeID: int(declared.Id()), ViewReceiverTypeID: int(receiverType.Id()), ViewAllowed: l.viewLiterals(declared)}
	property.ViewContract = l.result.ViewContractTypes[property.ViewTypeID]
	if alternatives {
		property.ViewContract, err = l.viewContract(node, declared)
		if err != nil {
			return nil, err
		}
	}
	// A narrowed demand cannot reinterpret absence as numeric bits or a pointer.
	if absent && of == ir.Number {
		property.Of = ir.MaybeNumber
		if l.acceptsUndefined(node) {
			return property, nil
		}
		return ir.Unwrap{Value: property}, nil
	}
	if absent && of.IsReference() && !l.includesUndefined(l.checker.GetTypeAtLocation(node)) && !l.acceptsUndefined(node) {
		return ir.Defined{Value: property, Message: "cast failed: field read failed: " + property.View + " is not initialized; expected " + property.ViewType + ", found missing"}, nil
	}
	return property, nil
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
