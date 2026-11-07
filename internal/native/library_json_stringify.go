package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (nativeEmitter *emitter) jsonSchema(schema *ir.JSONSchema) string {
	if schema == nil {
		return "NULL"
	}
	element := nativeEmitter.jsonSchema(schema.Element)
	fields := []string{}
	for _, field := range schema.Fields {
		child := nativeEmitter.jsonSchema(field.Schema)
		fields = append(
			fields,
			fmt.Sprintf("{&%s, %d, %s}", stringName(nativeEmitter.program.Strings[field.Name]), field.Slot, child),
		)
	}
	name, exists := nativeEmitter.sharedDeclaration(
		"adamic_json_schema",
		schema.Kind+"/"+element+"/"+strings.Join(fields, ","),
	)
	if exists {
		return "&" + name
	}
	list := "NULL"
	if len(fields) > 0 {
		list = name + "_fields"
		nativeEmitter.declarations = append(
			nativeEmitter.declarations,
			fmt.Sprintf("static const adamic_json_field %s[] = {%s};", list, strings.Join(fields, ", ")),
		)
	}
	nativeEmitter.declarations = append(nativeEmitter.declarations, fmt.Sprintf(
		"static const adamic_json_schema %s = {adamic_json_%s, %s, %d, %s};",
		name,
		schema.Kind,
		element,
		len(fields),
		list,
	))
	return "&" + name
}
func (nativeEmitter *emitter) jsonSlot(expression ir.Expression) string {
	value := nativeEmitter.value(expression)
	// A maybe boolean has two words as a value, and is not a slot type. Make its scalar-or-
	// undefined union for the duration of this call without allocating a number box.
	if expression.Type() == ir.MaybeBoolean {
		temp := nativeEmitter.temporary()
		nativeEmitter.line("adamic_maybe_boolean %s = %s;", temp, value)
		return fmt.Sprintf(
			"(adamic_value){.reference = %s.present ? (%s.boolean ? (void *)&adamic_box_true : (void *)&adamic_box_false) : NULL}",
			temp,
			temp,
		)
	}
	return borrowed(expression.Type(), value)
}
func (nativeEmitter *emitter) jsonStringify(expression ir.JSONStringify) string {
	nativeEmitter.declarations = append(nativeEmitter.declarations, `#include "json_stringify.h"`)
	schema := nativeEmitter.jsonSchema(expression.Schema)
	value := nativeEmitter.jsonSlot(expression.Value)
	// Pin scalar slots as well: evaluating a later argument may write an earlier local.
	first := nativeEmitter.temporary()
	nativeEmitter.line("adamic_value %s = %s;", first, value)
	replacer, space := "(adamic_value){.reference = NULL}", "(adamic_value){.reference = NULL}"
	if expression.Replacer != nil {
		replacer = nativeEmitter.jsonSlot(expression.Replacer)
		temp := nativeEmitter.temporary()
		nativeEmitter.line("adamic_value %s = %s;", temp, replacer)
		replacer = temp
	}
	if expression.Space != nil {
		space = nativeEmitter.jsonSlot(expression.Space)
	}
	replacerSchema := nativeEmitter.jsonSchema(expression.ReplacerSchema)
	spaceSchema := nativeEmitter.jsonSchema(expression.SpaceSchema)
	return nativeEmitter.own(ir.String, fmt.Sprintf(
		"adamic_json_stringify(%s, %s, %s, %s, %s, %s)",
		first,
		schema,
		replacer,
		replacerSchema,
		space,
		spaceSchema,
	))
}
