package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) jsonSchema(schema *ir.JSONSchema) string {
	if schema == nil {
		return "NULL"
	}
	e.declarations = append(e.declarations, `#include "json_stringify.h"`)
	element := e.jsonSchema(schema.Element)
	fields := []string{}
	for _, f := range schema.Fields {
		child := e.jsonSchema(f.Schema)
		fields = append(fields, fmt.Sprintf("{&adamic_string_%d, %d, %s}", f.Name, f.Slot, child))
	}
	name := e.temporary() + "_json_schema"
	list := "NULL"
	if len(fields) > 0 {
		list = name + "_fields"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_field %s[] = {%s};", list, strings.Join(fields, ", ")))
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_schema %s = {adamic_json_%s, %s, %d, %s, %t};", name, schema.Kind, element, len(fields), list, schema.Null))
	return "&" + name
}
func (e *emitter) jsonSlot(expression ir.Expression) string {
	value := e.value(expression)
	// A maybe boolean has two words as a value, and is not a slot type. Make its scalar-or-
	// undefined union for the duration of this call without allocating a number box.
	if expression.Type() == ir.MaybeBoolean {
		temp := e.temporary()
		e.line("adamic_maybe_boolean %s = %s;", temp, value)
		return fmt.Sprintf("(adamic_value){.reference = %s.present ? (%s.boolean ? (void *)&adamic_box_true : (void *)&adamic_box_false) : NULL}", temp, temp)
	}
	return borrowed(expression.Type(), value)
}
func (e *emitter) jsonStringify(expression ir.JSONStringify) string {
	e.declarations = append(e.declarations, `#include "json_stringify.h"`)
	schema := e.jsonSchema(expression.Schema)
	value := e.jsonSlot(expression.Value)
	// Pin scalar slots as well: evaluating a later argument may write an earlier local.
	first := e.temporary()
	e.line("adamic_value %s = %s;", first, value)
	replacer, space := "(adamic_value){.reference = NULL}", "(adamic_value){.reference = NULL}"
	if expression.Replacer != nil {
		replacer = e.jsonSlot(expression.Replacer)
		temp := e.temporary()
		e.line("adamic_value %s = %s;", temp, replacer)
		replacer = temp
	}
	if expression.Space != nil {
		space = e.jsonSlot(expression.Space)
	}
	replacerSchema := e.jsonSchema(expression.ReplacerSchema)
	spaceSchema := e.jsonSchema(expression.SpaceSchema)
	if expression.Schema.CallsUserCode() || expression.ReplacerSchema.CallsUserCode() {
		result := e.temporary()
		e.line("adamic_string *%s = adamic_json_stringify(%s, %s, %s, %s, %s, %s);", result, first, schema, replacer, replacerSchema, space, spaceSchema)
		e.checkThrown()
		return e.own(ir.String, result)
	}
	return e.own(ir.String, fmt.Sprintf("adamic_json_stringify(%s, %s, %s, %s, %s, %s)", first, schema, replacer, replacerSchema, space, spaceSchema))
}
