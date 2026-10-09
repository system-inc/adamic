package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) jsonContract(root *ir.JSONContract) string {
	if root == nil {
		return "NULL"
	}
	e.declarations = append(e.declarations, `#include "checked_json.h"`)
	nodes := ir.JSONContractGraph(root)
	definitions := map[int]*ir.JSONContract{}
	for _, c := range nodes {
		if c.Kind != "ref" {
			definitions[c.ID] = c
		}
	}
	name := e.temporary() + "_json_contract"
	ids := map[*ir.JSONContract]int{}
	for i, c := range nodes {
		ids[c] = i
	}
	pointer := func(c *ir.JSONContract) string {
		if c == nil {
			return "NULL"
		}
		return fmt.Sprintf("&%s[%d]", name, ids[c])
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_contract %s[%d];", name, len(nodes)))
	var entries []string
	for i, c := range nodes {
		var fields, alternatives []string
		for _, f := range c.Fields {
			fields = append(fields, fmt.Sprintf("{%s, %t, %s}", cString(f.Name), f.Optional, pointer(f.Contract)))
		}
		for _, a := range c.Alternatives {
			alternatives = append(alternatives, pointer(a))
		}
		list, choices := "NULL", "NULL"
		if len(fields) != 0 {
			list = fmt.Sprintf("%s_%d_fields", name, i)
			e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_contract_field %s[] = {%s};", list, strings.Join(fields, ", ")))
		}
		if len(alternatives) != 0 {
			choices = fmt.Sprintf("%s_%d_alternatives", name, i)
			e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_contract *const %s[] = {%s};", choices, strings.Join(alternatives, ", ")))
		}
		count := len(fields)
		if len(alternatives) != 0 {
			count = len(alternatives)
		}
		element := c.Element
		if c.Kind == "ref" {
			element = definitions[c.Reference]
			if element == nil {
				panic("missing JSON contract reference")
			}
		}
		entries = append(entries, fmt.Sprintf("{%s, %s, %s, %d, %s, %s, %s, %d, %s, %t}", cString(c.Kind), cString(c.Name), pointer(element), count, list, choices, cString(c.LiteralText), len(c.LiteralText), cNumber(c.LiteralNumber), c.LiteralBoolean))
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_contract %s[%d] = {%s};", name, len(nodes), strings.Join(entries, ", ")))
	return pointer(root)
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

// Optional tagged receivers preserve both nullish tags and evaluate once.
func (e *emitter) dynamicProperty(v ir.DynamicProperty) string {
	value := e.value(v.Object)
	read := fmt.Sprintf("adamic_dynamic_property(%s, %s)", value, cString(v.Name))
	if v.Optional {
		read = fmt.Sprintf("(%s==NULL || %s==&adamic_null ? NULL : %s)", value, value, read)
	}
	return e.own(ir.Union, read)
}
