package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) recordCall(call ir.RecordCall) string {
	args := make([]string, len(call.Arguments))
	for i, a := range call.Arguments {
		args[i] = e.value(a)
	}
	switch call.Method {
	case "get":
		slot := e.temporary()
		lookup := "adamic_record_get"
		if call.OwnOnly {
			lookup = "adamic_record_get_own"
		}
		e.line("adamic_value *%s = %s(%s, %s);", slot, lookup, args[0], args[1])
		if call.Returns.IsMaybe() {
			return e.snapshot(call.Returns, maybeSlot(call.Element, slot))
		}
		return e.own(call.Returns, fmt.Sprintf("%s == NULL ? NULL : adamic_retain(%s->reference)", slot, slot))
	case "set":
		value := e.snapshot(call.Element, args[2])
		e.recordSet(args[0], retained(args[1]), held(call.Element, value), args[1], e.program.RecordWriteMayThrow(call))
		return value
	case "delete":
		if !e.program.RecordWriteMayThrow(call) {
			return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_delete(%s, %s)", args[0], args[1]))
		}
		deleted := e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_delete_checked(%s, %s)", args[0], args[1]))
		// A fixed key's delete throws: strict code never sees false.
		e.line("if (!%s) adamic_record_refused(%s, true);", deleted, args[1])
		e.checkThrown()
		return deleted
	case "defineHidden":
		// The record takes key and value, the ones it refuses too; the caller throws on false.
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_define_hidden(%s, %s, %s, %s)", args[0], retained(args[1]), held(call.Element, args[2]), recordScalar(call.Element)))
	case "has":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_has(%s, %s)", args[1], args[0]))
	case "hasOwn":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_has_own(%s, %s)", args[0], args[1]))
	case "keys":
		return e.own(ir.Array, fmt.Sprintf("adamic_record_keys(%s)", args[0]))
	case "values", "entries":
		element := call.Element
		if call.Method == "entries" {
			// Entries are [key, value] tuples, references whatever the values are.
			element = ir.Union
		}
		result := e.own(ir.Array, fmt.Sprintf("adamic_array_make(0, %s)", arrayElement(element)))
		e.recordWalk(args[0], func(key, slot string) {
			value := slot
			if call.Element.IsReference() {
				value = fmt.Sprintf("(adamic_value){.reference = adamic_retain(%s.reference)}", slot)
			}
			if call.Method == "entries" {
				shape := e.shapeOf([]string{"0", "1"}, []ir.Type{ir.String, call.Element})
				pair := e.temporary()
				e.line("adamic_object *%s = adamic_object_new(&%s);", pair, shape)
				// [key, value] is a tuple to a view, and its slots carry their representations as a literal's do.
				e.line("%s->tuple = true;", pair)
				e.line("adamic_object_field_types(%s)[0] = %d;", pair, ir.String)
				e.line("adamic_object_field_types(%s)[1] = %d;", pair, call.Element)
				e.line("%s->slots[0].reference = adamic_retain(%s);", pair, key)
				e.line("%s->slots[1] = %s;", pair, value)
				value = fmt.Sprintf("(adamic_value){.reference = %s}", pair)
			}
			e.line("adamic_array_push(%s, %s);", result, value)
		})
		return result
	}
	panic("native: unknown record operation")
}

// Nothing called by walk can mutate its source or throw; outputs are retained before advancing.
func (e *emitter) recordWalk(record string, visit func(string, string)) {
	iterator, key, slot := e.temporary(), e.temporary(), e.temporary()
	e.line("adamic_record_iterator *%s = adamic_record_iterate(%s);", iterator, record)
	e.line("adamic_string *%s;", key)
	e.line("adamic_value %s;", slot)
	e.line("while (adamic_record_iterator_next(%s, &%s, &%s)) {", iterator, key, slot)
	e.indent++
	visit(key, slot)
	e.indent--
	e.line("}")
	e.line("adamic_release(%s);", iterator)
}
func (e *emitter) dictionaryLiteral(literal ir.RecordLiteral) string {
	result := e.own(ir.Record, fmt.Sprintf("adamic_record_new(%s)", recordValueKind(literal.Element)))
	if literal.Spread != nil {
		source := e.value(literal.Spread)
		e.recordWalk(source, func(key, slot string) {
			value := slot
			if literal.Element.IsReference() {
				value = fmt.Sprintf("(adamic_value){.reference = adamic_retain(%s.reference)}", slot)
			}
			e.line("adamic_record_define(%s, adamic_retain(%s), %s);", result, key, value)
		})
	}
	for _, entry := range literal.Entries {
		key := e.value(entry.Key)
		value := e.value(entry.Value)
		e.line("adamic_record_define(%s, %s, %s);", result, retained(key), held(literal.Element, value))
	}
	return result
}

func (e *emitter) recordCoalesce(c ir.RecordCoalesce) string {
	record, key := e.value(c.Record), e.value(c.Key)
	slot := e.temporary()
	e.line("adamic_value *%s = adamic_record_get(%s, %s);", slot, record, key)
	result := e.temporary()
	e.line("%s %s;", cType(c.Element), result)
	present := slot + " != NULL"
	if c.Element.IsReference() {
		present += " && " + slot + "->reference != NULL"
	}
	e.line("if (%s) {", present)
	value := slot + "->" + member(c.Element)
	if c.Element.IsReference() {
		value = retained(value)
	}
	e.line("\t%s = %s;", result, value)
	text, fallback, owned := e.aside(c.Value)
	e.line("} else {")
	e.out.WriteString(text)
	e.line("\t%s = %s;", result, fallback)
	if c.Element.IsReference() {
		e.line("\t%s = adamic_retain(%s);", result, result)
	}
	checked := e.program.RecordWriteMayThrow(c)
	if checked {
		e.line("\tif (!adamic_record_set_checked(%s, %s, %s)) adamic_record_refused(%s, false);", record, retained(key), held(c.Element, result), key)
	} else {
		e.line("\tadamic_record_set(%s, %s, %s);", record, retained(key), held(c.Element, result))
	}
	for i := len(owned) - 1; i >= 0; i-- {
		e.line("\tadamic_release(%s);", owned[i])
	}
	e.line("}")
	if c.Element.IsReference() {
		e.owned = append(e.owned, result)
	}
	if checked {
		e.checkThrown()
	}
	return result
}

// recordSet emits a record write of value under key into record, each the C the call takes, and named,
// the key's string, borrowed. A checked one, whose key may be one the program hides (ir.Program's
// RecordWriteMayThrow), is adamic_record_set_checked, refused on a read-only key, and throws Node's
// TypeError naming it; any other is adamic_record_set, as without a define.
func (e *emitter) recordSet(record, key, value, named string, checked bool) {
	if !checked {
		e.line("adamic_record_set(%s, %s, %s);", record, key, value)
		return
	}
	e.line("if (!adamic_record_set_checked(%s, %s, %s)) adamic_record_refused(%s, false);", record, key, value, named)
	e.checkThrown()
}

// recordScalar is how adamic_record_define_hidden compares what a record of element's slots holds.
func recordScalar(element ir.Type) string {
	switch element {
	case ir.Number:
		return "adamic_record_numbers"
	case ir.MaybeNumber:
		return "adamic_record_maybe_numbers"
	case ir.Boolean:
		return "adamic_record_booleans"
	}
	return "adamic_record_references"
}

// recordValueKind is the adamic_field_kind a record of the element type is made with (record.c), which
// its shape keeps, so a run-time key read boxes its values by it. Unlike a field's storage kind
// (fieldStorageKind), boolean | undefined is a boolean here: the packed byte reads as one, 2 for
// undefined, as number | undefined's packed double does as a number.
func recordValueKind(element ir.Type) string {
	if element.IsReference() {
		return "adamic_field_reference"
	}
	if element == ir.Boolean || element == ir.MaybeBoolean {
		return "adamic_field_boolean"
	}
	return "adamic_field_number"
}
