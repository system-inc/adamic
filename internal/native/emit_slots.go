package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A lookup's slot distinguishes a missing element or entry from a present null reference. Keep
// that presence until typeof has classified it, while sharing the ordinary value read's lookup.
func (e *emitter) arrayIndexSlot(expression ir.ArrayIndex) string {
	array := e.value(expression.Array)
	if expression.View != "" {
		array = e.own(ir.Array, "adamic_retain("+array+")")
		if expression.Optional {
			snapshot, slot := e.temporary(), e.temporary()
			e.line("adamic_value %s;", snapshot)
			e.line("adamic_value *%s=NULL;", slot)
			text, index, owned := e.aside(expression.Index)
			e.line("if(%s!=NULL){", array)
			e.out.WriteString(text)
			expression.Optional = false
			selected := e.emitViewArrayRead(expression, array, index)
			e.line("if(%s!=NULL){%s=*%s;%s=&%s;}", selected, snapshot, selected, slot, snapshot)
			for i := len(owned) - 1; i >= 0; i-- {
				e.line("adamic_release(%s);", owned[i])
			}
			e.line("}")
			return slot
		}
		index := e.value(expression.Index)
		return e.emitViewArrayRead(expression, array, index)
	}
	if expression.Optional {
		// Snapshot the receiver before the key can replace its binding. The key's
		// work and temporary cleanup belong only to the present branch.
		array = e.snapshot(ir.Array, array)
		slot := e.temporary()
		e.line("adamic_value *%s = NULL;", slot)
		text, index, owned := e.aside(expression.Index)
		e.line("if (%s != NULL) {", array)
		e.out.WriteString(text)
		e.indent++
		e.line("%s = adamic_array_at(%s, %s);", slot, array, index)
		for index := len(owned) - 1; index >= 0; index-- {
			e.line("adamic_release(%s);", owned[index])
		}
		e.indent--
		e.line("}")
		return slot
	}
	index := e.value(expression.Index)
	slot := e.temporary()
	lookup := "adamic_array_at"
	if expression.Relative {
		lookup = "adamic_array_at_relative"
	} else if read, isRead := expression.Index.(ir.Read); isRead && e.program.Locals[read.Local].Counter {
		lookup, index = "adamic_array_at_integer", e.localName(read.Local)
	}
	e.line("adamic_value *%s = %s(%s, %s);", slot, lookup, array, index)
	return slot
}

func (e *emitter) mapGetSlot(expression ir.MapGet) string {
	object := e.value(expression.Map)
	key := e.value(expression.Key)
	slot := e.temporary()
	e.line("adamic_value *%s = adamic_map_get(%s, %s);", slot, object, borrowed(expression.KeyType, key))
	return slot
}

// typeOfReference evaluates a reference once and retains the lookup's presence where null and
// undefined could otherwise collapse into the same pointer. The classifier remains union_typeof.
func (e *emitter) typeOfReference(observation ir.TypeOf) (string, string) {
	if observation.Null {
		var slot string
		var of ir.Type
		retain := true
		switch value := observation.Value.(type) {
		case ir.ArrayIndex:
			slot, of = e.arrayIndexSlot(value), value.Element
		case ir.MapGet:
			slot, of = e.mapGetSlot(value), value.ValueType
		case ir.ArrayPop:
			array := e.snapshot(ir.Array, e.value(value.Array))
			slot, of, retain = e.temporary(), value.Element, false
			e.line("adamic_value *%s = %s->length == 0 ? NULL : &%s->elements[--%s->length];", slot, array, array, array)
		}
		if slot != "" {
			reference := slot + "->reference"
			if retain {
				reference = "adamic_retain(" + reference + ")"
			}
			operand := e.own(of, fmt.Sprintf("%s == NULL ? NULL : (%s)%s", slot, cType(of), reference))
			return operand, slot + " != NULL"
		}
	}
	null := observation.Null
	if _, literal := observation.Value.(ir.Null); literal {
		null = true
	}
	return e.value(observation.Value), fmt.Sprintf("%t", null)
}
