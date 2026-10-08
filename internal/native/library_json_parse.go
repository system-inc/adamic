package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) jsonParseSchema(s *ir.JSONParseSchema) string {
	if s == nil {
		return "NULL"
	}
	element := e.jsonParseSchema(s.Element)
	fields := []string{}
	names := []string{}
	types := []ir.Type{}
	for _, f := range s.Fields {
		child := e.jsonParseSchema(f.Schema)
		fields = append(fields, fmt.Sprintf("{&adamic_string_%d, %s}", e.jsonParseStringIndex(f.Name), child))
		names = append(names, f.Name)
		types = append(types, f.Schema.Of)
	}
	members := []string{}
	for _, m := range s.Members {
		members = append(members, e.jsonParseSchema(m))
	}
	name := e.temporary() + "_parse_schema"
	fieldList, memberList, shape := "NULL", "NULL", "NULL"
	if len(fields) > 0 {
		fieldList = name + "_fields"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_parse_field %s[] = {%s};", fieldList, strings.Join(fields, ", ")))
	}
	if len(members) > 0 {
		memberList = name + "_members"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_parse_schema *const %s[] = {%s};", memberList, strings.Join(members, ", ")))
	}
	if s.Kind == "object" {
		n := e.shapeOf(names, types)
		shape = "&" + n
		if e.dynamicProperties() && len(names) > 0 {
			e.line("adamic_register_shape_types(&%s_metadata);", n)
		}
	}
	literal := "NULL"
	if s.HasLiteral {
		literal = fmt.Sprintf("&adamic_string_%d", e.jsonParseStringIndex(s.Literal))
	}
	count := len(fields)
	if s.Kind == "union" {
		count = len(members)
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_parse_schema %s = {json_%s, %d, &adamic_string_%d, %t, %s, %s, %d, %s, %s, %s};", name, s.Kind, s.Of, e.jsonParseStringIndex(s.Name), s.Optional, literal, element, count, fieldList, memberList, shape))
	return "&" + name
}
func (e *emitter) jsonParse(p ir.JSONParse) string {
	e.declarations = append(e.declarations, `#include "json_parse.h"`)
	check, layout := e.jsonParseSchema(p.Check), e.jsonParseSchema(p.Layout)
	input := e.value(p.Text)
	temp := e.temporary()
	e.line("adamic_value %s = adamic_json_parse(%s, %s, %s);", temp, input, check, layout)
	e.checkThrown()
	switch p.Of {
	case ir.Number:
		return temp + ".number"
	case ir.Boolean:
		return temp + ".boolean"
	case ir.MaybeNumber:
		return "adamic_maybe_number_unpack(" + temp + ".number)"
	}
	return e.own(p.Of, fmt.Sprintf("(%s)%s.reference", cType(p.Of), temp))
}

func (e *emitter) jsonParseStringIndex(text string) int {
	for i, s := range e.program.Strings {
		if s == text {
			return i
		}
	}
	panic("JSON.parse schema string missing from IR")
}
