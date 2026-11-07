package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) recordCall(call ir.RecordCall) string {
	args := make([]string, len(call.Arguments))
	for index, argument := range call.Arguments {
		args[index] = nativeEmitter.value(argument)
	}
	switch call.Method {
	case "get":
		slot := nativeEmitter.temporary()
		lookup := "adamic_record_get"
		if call.OwnOnly {
			lookup = "adamic_record_get_own"
		}
		nativeEmitter.line("adamic_value *%s = %s(%s, %s);", slot, lookup, args[0], args[1])
		if call.Returns.IsMaybe() {
			return nativeEmitter.snapshot(call.Returns, maybeSlot(call.Element, slot))
		}
		return nativeEmitter.own(
			call.Returns,
			fmt.Sprintf("%s == NULL ? NULL : adamic_retain(%s->reference)", slot, slot),
		)
	case "set":
		value := nativeEmitter.snapshot(call.Element, args[2])
		nativeEmitter.recordSet(
			args[0],
			retained(args[1]),
			held(call.Element, value),
			args[1],
			nativeEmitter.program.RecordWriteMayThrow(call),
		)
		return value
	case "delete":
		if !nativeEmitter.program.RecordWriteMayThrow(call) {
			return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_delete(%s, %s)", args[0], args[1]))
		}
		deleted := nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_record_delete_checked(%s, %s)", args[0], args[1]),
		)
		// A fixed key's delete throws: strict code never sees false.
		nativeEmitter.line("if (!%s) adamic_record_refused(%s, true);", deleted, args[1])
		nativeEmitter.checkThrown()
		return deleted
	case "defineHidden":
		// The record takes key and value, the ones it refuses too; the caller throws on false.
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf(
			"adamic_record_define_hidden(%s, %s, %s, %s)",
			args[0],
			retained(args[1]),
			held(call.Element, args[2]),
			recordScalar(call.Element),
		))
	case "has":
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_has(%s, %s)", args[1], args[0]))
	case "hasOwn":
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_has_own(%s, %s)", args[0], args[1]))
	case "keys":
		return nativeEmitter.own(ir.Array, fmt.Sprintf("adamic_record_keys(%s)", args[0]))
	case "values", "entries":
		element := call.Element
		if call.Method == "entries" {
			// Entries are [key, value] tuples, references whatever the values are.
			element = ir.Union
		}
		result := nativeEmitter.own(ir.Array, fmt.Sprintf("adamic_array_make(0, %s)", arrayElement(element)))
		nativeEmitter.recordWalk(args[0], func(key, slot string) {
			value := slot
			if call.Element.IsReference() {
				value = fmt.Sprintf("(adamic_value){.reference = adamic_retain(%s.reference)}", slot)
			}
			if call.Method == "entries" {
				shape := nativeEmitter.shapeOf([]string{"0", "1"}, []ir.Type{ir.String, call.Element})
				pair := nativeEmitter.temporary()
				nativeEmitter.line("adamic_object *%s = adamic_object_new(&%s);", pair, shape)
				// [key, value] is a tuple to a view, and its slots carry their representations as a literal's do.
				nativeEmitter.line("%s->tuple = true;", pair)
				nativeEmitter.line("adamic_object_field_types(%s)[0] = %d;", pair, ir.String)
				nativeEmitter.line("adamic_object_field_types(%s)[1] = %d;", pair, call.Element)
				nativeEmitter.line("%s->slots[0].reference = adamic_retain(%s);", pair, key)
				nativeEmitter.line("%s->slots[1] = %s;", pair, value)
				value = fmt.Sprintf("(adamic_value){.reference = %s}", pair)
			}
			nativeEmitter.line("adamic_array_push(%s, %s);", result, value)
		})
		return result
	}
	panic("native: unknown record operation")
}

// Nothing called by walk can mutate its source or throw; outputs are retained before advancing.
func (nativeEmitter *emitter) recordWalk(record string, visit func(string, string)) {
	iterator, key, slot := nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary()
	nativeEmitter.line("adamic_record_iterator *%s = adamic_record_iterate(%s);", iterator, record)
	nativeEmitter.line("adamic_string *%s;", key)
	nativeEmitter.line("adamic_value %s;", slot)
	nativeEmitter.line("while (adamic_record_iterator_next(%s, &%s, &%s)) {", iterator, key, slot)
	nativeEmitter.indent++
	visit(key, slot)
	nativeEmitter.indent--
	nativeEmitter.line("}")
	nativeEmitter.line("adamic_release(%s);", iterator)
}
func (nativeEmitter *emitter) dictionaryLiteral(literal ir.RecordLiteral) string {
	result := nativeEmitter.own(ir.Record, fmt.Sprintf("adamic_record_new(%s)", recordValueKind(literal.Element)))
	if literal.Spread != nil {
		source := nativeEmitter.value(literal.Spread)
		nativeEmitter.recordWalk(source, func(key, slot string) {
			value := slot
			if literal.Element.IsReference() {
				value = fmt.Sprintf("(adamic_value){.reference = adamic_retain(%s.reference)}", slot)
			}
			nativeEmitter.line("adamic_record_define(%s, adamic_retain(%s), %s);", result, key, value)
		})
	}
	for _, entry := range literal.Entries {
		key := nativeEmitter.value(entry.Key)
		value := nativeEmitter.value(entry.Value)
		nativeEmitter.line("adamic_record_define(%s, %s, %s);", result, retained(key), held(literal.Element, value))
	}
	return result
}

func (nativeEmitter *emitter) recordCoalesce(coalesce ir.RecordCoalesce) string {
	record, key := nativeEmitter.value(coalesce.Record), nativeEmitter.value(coalesce.Key)
	slot := nativeEmitter.temporary()
	nativeEmitter.line("adamic_value *%s = adamic_record_get(%s, %s);", slot, record, key)
	result := nativeEmitter.temporary()
	nativeEmitter.line("%s %s;", cType(coalesce.Element), result)
	present := slot + " != NULL"
	if coalesce.Element.IsReference() {
		present += " && " + slot + "->reference != NULL"
	}
	nativeEmitter.line("if (%s) {", present)
	value := slot + "->" + member(coalesce.Element)
	if coalesce.Element.IsReference() {
		value = retained(value)
	}
	nativeEmitter.line("\t%s = %s;", result, value)
	text, fallback, owned := nativeEmitter.aside(coalesce.Value)
	nativeEmitter.line("} else {")
	nativeEmitter.out.WriteString(text)
	nativeEmitter.line("\t%s = %s;", result, fallback)
	if coalesce.Element.IsReference() {
		nativeEmitter.line("\t%s = adamic_retain(%s);", result, result)
	}
	checked := nativeEmitter.program.RecordWriteMayThrow(coalesce)
	if checked {
		nativeEmitter.line(
			"\tif (!adamic_record_set_checked(%s, %s, %s)) adamic_record_refused(%s, false);",
			record,
			retained(key),
			held(coalesce.Element, result),
			key,
		)
	} else {
		nativeEmitter.line("\tadamic_record_set(%s, %s, %s);", record, retained(key), held(coalesce.Element, result))
	}
	for index := len(owned) - 1; index >= 0; index-- {
		nativeEmitter.line("\tadamic_release(%s);", owned[index])
	}
	nativeEmitter.line("}")
	if coalesce.Element.IsReference() {
		nativeEmitter.owned = append(nativeEmitter.owned, result)
	}
	if checked {
		nativeEmitter.checkThrown()
	}
	return result
}

// recordSet emits a record write of value under key into record, each the C the call takes, and named,
// the key's string, borrowed. A checked one, whose key may be one the program hides (ir.Program's
// RecordWriteMayThrow), is adamic_record_set_checked, refused on a read-only key, and throws Node's
// TypeError naming it; any other is adamic_record_set, as without a define.
func (nativeEmitter *emitter) recordSet(record, key, value, named string, checked bool) {
	if !checked {
		nativeEmitter.line("adamic_record_set(%s, %s, %s);", record, key, value)
		return
	}
	nativeEmitter.line(
		"if (!adamic_record_set_checked(%s, %s, %s)) adamic_record_refused(%s, false);",
		record,
		key,
		value,
		named,
	)
	nativeEmitter.checkThrown()
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
