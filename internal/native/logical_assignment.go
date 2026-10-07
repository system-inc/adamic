package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) logicalAssignment(expression ir.LogicalAssignment) string {
	var left, receiver, key, index string
	var storage ir.Type
	var right ir.Expression
	switch write := expression.Write.(type) {
	case ir.Assign:
		left = e.value(expression.Read)
		storage, right = e.program.Locals[write.Local].Type, write.Value
	case ir.SetProperty:
		receiver = e.snapshot(ir.Object, e.value(write.Object))
		key = cString(write.Name)
		if expression.Key != nil {
			key = e.snapshot(ir.String, e.value(expression.Key)) + "->bytes"
		}
		storage, right = write.Value.Type(), write.Value
		if storage == ir.MaybeNumber {
			left = e.snapshot(storage, fmt.Sprintf("adamic_object_maybe_number(%s, %s, &%s)", receiver, key, e.cache()))
		} else {
			slot := e.temporary()
			e.line("adamic_value *%s = adamic_object_optional_field(%s, %s, &%s);", slot, receiver, key, e.cache())
			field := unslotted(storage, slot+"->"+member(storage))
			if storage.IsReference() {
				left = e.own(storage, fmt.Sprintf("%s == NULL ? NULL : (%s)adamic_retain(%s)", slot, cType(storage), field))
			} else if storage.IsMaybe() {
				left = e.snapshot(storage, fmt.Sprintf("%s == NULL ? %s : %s", slot, zero(storage), field))
			} else {
				left = e.snapshot(storage, field)
			}
		}
	case ir.SetIndex:
		receiver = e.snapshot(ir.Array, e.value(write.Array))
		index = e.snapshot(ir.Number, e.value(write.Index))
		storage, right = write.Element, write.Value
		slot := e.temporary()
		e.line("adamic_value *%s = adamic_array_at(%s, %s);", slot, receiver, index)
		readType := expression.Read.Type()
		if readType.IsMaybe() {
			left = e.snapshot(readType, maybeSlot(storage, slot))
		} else {
			left = e.own(readType, fmt.Sprintf("%s == NULL ? NULL : (%s)adamic_retain(%s->reference)", slot, cType(readType), slot))
		}
	default:
		panic(fmt.Sprintf("native: no logical assignment for %T", expression.Write))
	}
	leftType := expression.Read.Type()
	left = e.snapshot(leftType, left)
	condition := ""
	if expression.Operator == "??=" {
		condition = left + " == NULL"
		if leftType.IsMaybe() {
			condition = "!" + left + ".present"
		} else if !leftType.IsReference() {
			condition = "false"
		}
	} else {
		condition = e.toBoolean(leftType, left)
		if expression.Operator == "||=" {
			condition = "!(" + condition + ")"
		}
	}
	result := e.temporary()
	e.line("%s %s;", cType(expression.Of), result)
	e.line("if (%s) {", unwrap(condition))
	// The aside holds the right side's temporaries only on the branch taking the write.
	text, value, owned := e.aside(right)
	e.out.WriteString(text)
	e.indent++
	value = e.snapshot(storage, value)
	switch write := expression.Write.(type) {
	case ir.Assign:
		if write.Checked {
			e.checkReady(write.Local)
		}
		e.store(write.Local, value, false)
	case ir.SetProperty:
		e.line("adamic_object_check_write(%s, %s);", receiver, key)
		slot := e.temporary()
		e.line("adamic_value *%s = adamic_object_write_field(%s, %s, &%s);", slot, receiver, key, e.cache())
		if storage.IsReference() {
			old := e.temporary()
			e.line("void *%s = %s->reference;", old, slot)
			e.line("%s->reference = %s;", slot, retained(value))
			e.line("adamic_release(%s);", old)
		} else {
			e.line("%s->%s = %s;", slot, member(storage), slotted(storage, value))
		}
	case ir.SetIndex:
		stored := value
		if storage.IsReference() {
			stored = retained(stored)
		}
		e.line("adamic_array_set(%s, %s, (adamic_value){.%s = %s});", receiver, index, member(storage), slotted(storage, stored))
	}
	e.assignLogicalResult(result, storage, expression.Of, value)
	for i := len(owned) - 1; i >= 0; i-- {
		e.line("adamic_release(%s);", owned[i])
	}
	e.indent--
	e.line("} else {")
	e.indent++
	e.assignLogicalResult(result, leftType, expression.Of, left)
	e.indent--
	e.line("}")
	if expression.Of.IsReference() {
		e.owned = append(e.owned, result)
	}
	return result
}

func (e *emitter) assignLogicalResult(result string, from, to ir.Type, value string) {
	if from.IsMaybe() && to == from.Present() {
		value, from = value+"."+member(from.Present()), from.Present()
	}
	value, fresh := logicalConverted(from, to, value)
	if to.IsReference() && !fresh {
		value = retained(value)
	}
	e.line("%s = %s;", result, value)
}
