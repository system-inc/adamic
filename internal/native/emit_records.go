package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) recordCall(call ir.RecordCall) string {
	if call.DictionaryRead != nil && len(e.program.ViewOrigins) != 0 {
		return e.dictionaryRead(*call.DictionaryRead)
	}
	args := make([]string, len(call.Arguments))
	for i, a := range call.Arguments {
		args[i] = e.value(a)
	}
	receiver := args[0]
	if call.Method == "has" {
		receiver = args[1]
	}
	if call.Element != 0 {
		e.line("adamic_record_storage_check(%s, %d);", receiver, call.Element)
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
		e.line("adamic_record_set(%s, %s, %s);", args[0], retained(args[1]), held(call.Element, value))
		return value
	case "delete":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_delete(%s, %s)", args[0], args[1]))
	case "has":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_has(%s, %s)", args[1], args[0]))
	case "hasOwn":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_record_has_own(%s, %s)", args[0], args[1]))
	case "keys":
		return e.own(ir.Array, fmt.Sprintf("adamic_record_keys(%s)", args[0]))
	case "values", "entries":
		result := e.own(ir.Array, fmt.Sprintf("adamic_array_new(0, %t)", call.Method == "entries" || call.Element.IsReference()))
		e.recordWalk(args[0], func(key, slot string) {
			value := slot
			if call.Element.IsReference() {
				value = fmt.Sprintf("(adamic_value){.reference = adamic_retain(%s.reference)}", slot)
			}
			if call.Method == "entries" {
				shape := e.shapeOf([]string{"0", "1"}, []ir.Type{ir.String, call.Element})
				pair := e.temporary()
				e.line("adamic_object *%s = adamic_object_new(&%s);", pair, shape)
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
func (e *emitter) recordLiteral(literal ir.RecordLiteral) string {
	result := e.own(ir.Record, fmt.Sprintf("adamic_record_new_typed(%t, %d)", literal.Element.IsReference(), literal.Element))
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
	e.line("adamic_record_storage_check(%s, %d);", record, c.Element)
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
	e.line("\tadamic_record_set(%s, %s, %s);", record, retained(key), held(c.Element, result))
	for i := len(owned) - 1; i >= 0; i-- {
		e.line("\tadamic_release(%s);", owned[i])
	}
	e.line("}")
	if c.Element.IsReference() {
		e.owned = append(e.owned, result)
	}
	return result
}
