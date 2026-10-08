package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) jsonContract(c *ir.JSONContract) string {
	if c == nil {
		return "NULL"
	}
	e.declarations = append(e.declarations, `#include "checked_json.h"`)
	element := e.jsonContract(c.Element)
	var fields, alternatives []string
	for _, field := range c.Fields {
		fields = append(fields, fmt.Sprintf("{%s, %t, %s}", cString(field.Name), field.Optional, e.jsonContract(field.Contract)))
	}
	for _, part := range c.Alternatives {
		alternatives = append(alternatives, e.jsonContract(part))
	}
	name := e.temporary() + "_json_contract"
	list, choices := "NULL", "NULL"
	if len(fields) != 0 {
		list = name + "_fields"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_contract_field %s[] = {%s};", list, strings.Join(fields, ", ")))
	}
	if len(alternatives) != 0 {
		choices = name + "_alternatives"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_contract *const %s[] = {%s};", choices, strings.Join(alternatives, ", ")))
	}
	count := len(fields)
	if len(alternatives) != 0 {
		count = len(alternatives)
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_contract %s = {%s, %s, %s, %d, %s, %s, %s, %d, %s, %t};", name, cString(c.Kind), cString(c.Name), element, count, list, choices, cString(c.LiteralText), len(c.LiteralText), cNumber(c.LiteralNumber), c.LiteralBoolean))
	return "&" + name
}
func (e *emitter) checkedJSON(v ir.CheckedJSON) string {
	value := e.value(v.Value)
	schema := e.jsonContract(v.Contract)
	e.line("adamic_check_json(%s, %s, %s);", value, schema, cString(v.Path))
	return e.jsonCheckedValue(value, v.Type())
}
func (e *emitter) jsonCheckedValue(value string, to ir.Type) string {
	switch to {
	case ir.Number:
		return e.snapshot(to, fmt.Sprintf("((adamic_number_box *)%s)->number", value))
	case ir.Boolean:
		return e.snapshot(to, fmt.Sprintf("((adamic_boolean_box *)%s)->boolean", value))
	case ir.MaybeNumber:
		return e.snapshot(to, fmt.Sprintf("(%s == NULL ? (adamic_maybe_number){false, 0} : (adamic_maybe_number){true, ((adamic_number_box *)%s)->number})", value, value))
	case ir.MaybeBoolean:
		return e.snapshot(to, fmt.Sprintf("(%s == NULL ? (adamic_maybe_boolean){false, false} : (adamic_maybe_boolean){true, ((adamic_boolean_box *)%s)->boolean})", value, value))
	default:
		return e.own(to, fmt.Sprintf("(%s)adamic_retain(%s)", cType(to), value))
	}
}
func runtimeJSONType(to ir.Type) *ir.JSONContract {
	kind := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Array: "array", ir.Object: "object"}[to]
	if to == ir.Union {
		return &ir.JSONContract{Kind: "union", Name: "JSON value", Alternatives: []*ir.JSONContract{{Kind: "undefined", Name: "undefined"}, {Kind: "null", Name: "null"}, {Kind: "number", Name: "number"}, {Kind: "boolean", Name: "boolean"}, {Kind: "string", Name: "string"}, {Kind: "object", Name: "object"}, {Kind: "array", Name: "array"}}}
	}
	if to.IsMaybe() {
		base := runtimeJSONType(to.Present())
		return &ir.JSONContract{Kind: "union", Name: base.Name + " | undefined", Alternatives: []*ir.JSONContract{base, {Kind: "undefined", Name: "undefined"}}}
	}
	if kind == "" {
		panic("unsupported checked JSON read type")
	}
	base := &ir.JSONContract{Kind: kind, Name: kind}
	if to.IsReference() {
		return &ir.JSONContract{Kind: "union", Name: kind + " | undefined", Alternatives: []*ir.JSONContract{base, {Kind: "undefined", Name: "undefined"}}}
	}
	return base
}
func (e *emitter) checkedJSONProperty(v ir.Property) string {
	object := e.value(v.Object)
	read := fmt.Sprintf("adamic_dynamic_property((adamic_heap *)%s, %s)", object, cString(v.Name))
	if v.Optional {
		read = fmt.Sprintf("(%s==NULL ? NULL : %s)", object, read)
	}
	value := e.own(ir.Union, read)
	schema := e.jsonContract(runtimeJSONType(v.Type()))
	e.line("adamic_check_json(%s, %s, %s);", value, schema, cString(v.Name))
	return e.jsonCheckedValue(value, v.Type())
}
func (e *emitter) checkedJSONArrayIndex(v ir.ArrayIndex) string {
	array, index := e.value(v.Array), e.value(v.Index)
	e.declarations = append(e.declarations, `#include "checked_json.h"`)
	value := e.own(ir.Union, fmt.Sprintf("adamic_json_array_get(%s, %s)", array, index))
	schema := e.jsonContract(runtimeJSONType(v.Type()))
	e.line("adamic_check_json(%s, %s, \"array element\");", value, schema)
	return e.jsonCheckedValue(value, v.Type())
}
