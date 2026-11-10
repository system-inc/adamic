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
	element := e.jsonSchema(schema.Element)
	fields := []string{}
	for _, f := range schema.Fields {
		child := e.jsonSchema(f.Schema)
		fields = append(fields, fmt.Sprintf("{&%s, %d, %s}", stringName(e.program.Strings[f.Name]), f.Slot, child))
	}
	// An object reference is read as a union is, by its heap header, and then described from its
	// allocation by json_objects.c.
	kind := schema.Kind
	if kind == "reference" {
		kind = "union"
	}
	name, exists := e.sharedDeclaration("adamic_json_schema", kind+"/"+element+"/"+strings.Join(fields, ","))
	if exists {
		return "&" + name
	}
	list := "NULL"
	if len(fields) > 0 {
		list = name + "_fields"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_field %s[] = {%s};", list, strings.Join(fields, ", ")))
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_schema %s = {adamic_json_%s, %s, %d, %s};", name, kind, element, len(fields), list))
	return "&" + name
}

// jsonReferences reports whether a schema reaches an object reference, which only json_objects.c's
// provider can describe.
func jsonReferences(schema *ir.JSONSchema) bool {
	if schema == nil {
		return false
	}
	if schema.Kind == "reference" || jsonReferences(schema.Element) {
		return true
	}
	for _, field := range schema.Fields {
		if jsonReferences(field.Schema) {
			return true
		}
	}
	return false
}

// jsonStringifiesReferences reports whether a call may print an object reference, so that the call
// takes the provider and every object in the program records its slots' representation bytes.
func jsonStringifiesReferences(expression ir.JSONStringify) bool {
	return jsonReferences(expression.Schema) || jsonReferences(expression.ReplacerSchema) || jsonReferences(expression.SpaceSchema)
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
	call := "adamic_json_stringify"
	if jsonStringifiesReferences(expression) {
		call = "adamic_json_stringify_objects"
	}
	result := e.own(ir.String, fmt.Sprintf("%s(%s, %s, %s, %s, %s, %s, %s)", call, first, schema, replacer, replacerSchema, space, spaceSchema))
	// Past V8's longest, a RangeError is pending and the result empty, let go with the rest.
	e.checkThrown()
	return result
}
