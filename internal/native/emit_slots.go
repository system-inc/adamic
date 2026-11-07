package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A lookup's slot distinguishes a missing element or entry from a present null reference. Keep
// that presence until typeof has classified it, while sharing the ordinary value read's lookup.
func (nativeEmitter *emitter) arrayIndexSlot(expression ir.ArrayIndex) string {
	array := nativeEmitter.value(expression.Array)
	if expression.Optional {
		// Snapshot the receiver before the key can replace its binding. The key's
		// work and temporary cleanup belong only to the present branch.
		array = nativeEmitter.snapshot(ir.Array, array)
		slot := nativeEmitter.temporary()
		nativeEmitter.line("adamic_value *%s = NULL;", slot)
		text, index, owned := nativeEmitter.aside(expression.Index)
		nativeEmitter.line("if (%s != NULL) {", array)
		nativeEmitter.out.WriteString(text)
		nativeEmitter.indent++
		nativeEmitter.line("%s = adamic_array_at(%s, %s);", slot, array, index)
		for index := len(owned) - 1; index >= 0; index-- {
			nativeEmitter.line("adamic_release(%s);", owned[index])
		}
		nativeEmitter.indent--
		nativeEmitter.line("}")
		return slot
	}
	index := nativeEmitter.value(expression.Index)
	slot := nativeEmitter.temporary()
	lookup := "adamic_array_at"
	if expression.Relative {
		lookup = "adamic_array_at_relative"
	} else if read, isRead := expression.Index.(ir.Read); isRead && nativeEmitter.integerCounter(read.Local) {
		lookup, index = "adamic_array_at_integer", nativeEmitter.localName(read.Local)
	}
	nativeEmitter.line("adamic_value *%s = %s(%s, %s);", slot, lookup, array, index)
	return slot
}

func (nativeEmitter *emitter) mapGetSlot(expression ir.MapGet) string {
	object := nativeEmitter.value(expression.Map)
	key := nativeEmitter.value(expression.Key)
	slot := nativeEmitter.temporary()
	nativeEmitter.line("adamic_value *%s = adamic_map_get(%s, %s);", slot, object, borrowed(expression.KeyType, key))
	return slot
}

// typeOfReference evaluates a reference once and retains the lookup's presence where null and
// undefined could otherwise collapse into the same pointer. The classifier remains union_typeof.
func (nativeEmitter *emitter) typeOfReference(observation ir.TypeOf) (string, string) {
	if observation.Null {
		var slot string
		var of ir.Type
		retain := true
		switch value := observation.Value.(type) {
		case ir.ArrayIndex:
			slot, of = nativeEmitter.arrayIndexSlot(value), value.Element
		case ir.MapGet:
			slot, of = nativeEmitter.mapGetSlot(value), value.ValueType
		case ir.ArrayPop:
			array := nativeEmitter.snapshot(ir.Array, nativeEmitter.value(value.Array))
			slot, of, retain = nativeEmitter.temporary(), value.Element, false
			nativeEmitter.line(
				"adamic_value *%s = %s->length == 0 ? NULL : &%s->elements[--%s->length];",
				slot,
				array,
				array,
				array,
			)
		}
		if slot != "" {
			reference := slot + "->reference"
			if retain {
				reference = "adamic_retain(" + reference + ")"
			}
			operand := nativeEmitter.own(of, fmt.Sprintf("%s == NULL ? NULL : (%s)%s", slot, cType(of), reference))
			return operand, slot + " != NULL"
		}
	}
	null := observation.Null
	if _, literal := observation.Value.(ir.Null); literal {
		null = true
	}
	return nativeEmitter.value(observation.Value), fmt.Sprintf("%t", null)
}
