package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) logicalAssignment(expression ir.LogicalAssignment) string {
	var left, receiver, key, index string
	var storage ir.Type
	var right ir.Expression
	// Through a checked view, the field is read and written as a plain read and = through it are.
	viewed := false
	switch write := expression.Write.(type) {
	case ir.Assign:
		left = nativeEmitter.value(expression.Read)
		storage, right = nativeEmitter.program.Locals[write.Local].Type, write.Value
	case ir.SetProperty:
		receiver = nativeEmitter.snapshot(ir.Object, nativeEmitter.value(write.Object))
		key = cString(write.Name)
		if expression.Key != nil {
			key = nativeEmitter.snapshot(ir.String, nativeEmitter.value(expression.Key)) + "->bytes"
		}
		storage, right = write.Value.Type(), write.Value
		viewed = expression.Key == nil && nativeEmitter.program.CheckedField(write.ViewReceiverTypeID, write.Name)
		if viewed {
			read, isProperty := expression.Read.(ir.Property)
			if !isProperty || read.View == "" {
				panic("compiler bug: a logical assignment through a checked view without its checked read")
			}
			left = nativeEmitter.viewFieldAt(read, receiver)
		} else if storage == ir.MaybeNumber {
			left = nativeEmitter.snapshot(
				storage,
				fmt.Sprintf("adamic_object_maybe_number(%s, %s, &%s)", receiver, key, nativeEmitter.cache()),
			)
		} else {
			slot := nativeEmitter.temporary()
			nativeEmitter.line(
				"adamic_value *%s = adamic_object_optional_field(%s, %s, &%s);",
				slot,
				receiver,
				key,
				nativeEmitter.cache(),
			)
			field := unslotted(storage, slot+"->"+member(storage))
			if storage.IsReference() {
				left = nativeEmitter.own(
					storage,
					fmt.Sprintf("%s == NULL ? NULL : (%s)adamic_retain(%s)", slot, cType(storage), field),
				)
			} else if storage.IsMaybe() {
				left = nativeEmitter.snapshot(storage, fmt.Sprintf("%s == NULL ? %s : %s", slot, zero(storage), field))
			} else {
				left = nativeEmitter.snapshot(storage, field)
			}
		}
	case ir.SetIndex:
		receiver = nativeEmitter.snapshot(ir.Array, nativeEmitter.value(write.Array))
		index = nativeEmitter.snapshot(ir.Number, nativeEmitter.value(write.Index))
		storage, right = write.Element, write.Value
		slot := nativeEmitter.temporary()
		nativeEmitter.line("adamic_value *%s = adamic_array_at(%s, %s);", slot, receiver, index)
		readType := expression.Read.Type()
		if readType.IsMaybe() {
			left = nativeEmitter.snapshot(readType, maybeSlot(storage, slot))
		} else {
			left = nativeEmitter.own(
				readType,
				fmt.Sprintf("%s == NULL ? NULL : (%s)adamic_retain(%s->reference)", slot, cType(readType), slot),
			)
		}
	default:
		panic(fmt.Sprintf("native: no logical assignment for %T", expression.Write))
	}
	leftType := expression.Read.Type()
	left = nativeEmitter.snapshot(leftType, left)
	condition := ""
	if expression.Operator == "??=" {
		condition = left + " == NULL"
		if leftType == ir.Union {
			// A union holds null as its own box, as ?? tests it.
			condition += " || " + left + " == &adamic_null"
		}
		if leftType.IsMaybe() {
			condition = "!" + left + ".present"
		} else if !leftType.IsReference() {
			condition = "false"
		}
	} else {
		condition = nativeEmitter.toBoolean(leftType, left)
		if expression.Operator == "||=" {
			condition = "!(" + condition + ")"
		}
	}
	result := nativeEmitter.temporary()
	nativeEmitter.line("%s %s;", cType(expression.Of), result)
	nativeEmitter.line("if (%s) {", unwrap(condition))
	// The aside holds the right side's temporaries only on the branch taking the write.
	text, value, owned := nativeEmitter.aside(right)
	nativeEmitter.out.WriteString(text)
	nativeEmitter.indent++
	value = nativeEmitter.snapshot(storage, value)
	switch write := expression.Write.(type) {
	case ir.Assign:
		if write.Checked {
			nativeEmitter.checkReady(write, write.Local)
		}
		nativeEmitter.store(write.Local, value, false)
	case ir.SetProperty:
		if viewed {
			nativeEmitter.writeProperty(write, receiver, value, owned...)
			break
		}
		nativeEmitter.line("adamic_object_check_write(%s, %s);", receiver, key)
		// A frozen object's TypeError lets go of the right side's temporaries on the way out.
		if nativeEmitter.program.FrozenWriteMayThrow(write) {
			nativeEmitter.checkThrown(owned...)
		}
		slot := nativeEmitter.temporary()
		nativeEmitter.line(
			"adamic_value *%s = adamic_object_write_field(%s, %s, &%s);",
			slot,
			receiver,
			key,
			nativeEmitter.cache(),
		)
		if storage.IsReference() {
			old := nativeEmitter.temporary()
			nativeEmitter.line("void *%s = %s->reference;", old, slot)
			nativeEmitter.line("%s->reference = %s;", slot, retained(value))
			nativeEmitter.line("adamic_release(%s);", old)
		} else {
			nativeEmitter.line("%s->%s = %s;", slot, member(storage), slotted(storage, value))
		}
	case ir.SetIndex:
		stored := value
		if storage.IsReference() {
			stored = retained(stored)
		}
		nativeEmitter.line(
			"adamic_array_set(%s, %s, (adamic_value){.%s = %s});",
			receiver,
			index,
			member(storage),
			slotted(storage, stored),
		)
	}
	nativeEmitter.assignLogicalResult(result, storage, expression.Of, value)
	for index := len(owned) - 1; index >= 0; index-- {
		nativeEmitter.line("adamic_release(%s);", owned[index])
	}
	nativeEmitter.indent--
	nativeEmitter.line("} else {")
	nativeEmitter.indent++
	nativeEmitter.assignLogicalResult(result, leftType, expression.Of, left)
	nativeEmitter.indent--
	nativeEmitter.line("}")
	if expression.Of.IsReference() {
		nativeEmitter.owned = append(nativeEmitter.owned, result)
	}
	return result
}

func (nativeEmitter *emitter) assignLogicalResult(result string, from, to ir.Type, value string) {
	if from.IsMaybe() && to == from.Present() {
		value, from = value+"."+member(from.Present()), from.Present()
	}
	value, fresh := logicalConverted(from, to, value)
	if to.IsReference() && !fresh {
		value = retained(value)
	}
	nativeEmitter.line("%s = %s;", result, value)
}
