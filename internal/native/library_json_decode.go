package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) jsonDecode(expression ir.JSONDecode) string {
	e.declarations = append(e.declarations, `#include "json_decode.h"`)
	name := e.temporary() + "_decode"
	nodes := []string{}
	for i, n := range expression.Schema.Nodes {
		fields := "NULL"
		if len(n.Fields) > 0 {
			values := []string{}
			for _, f := range n.Fields {
				values = append(values, fmt.Sprintf("{%s, %d, %t}", cString(f.Name), f.Node, f.Optional))
			}
			fields = fmt.Sprintf("%s_fields_%d", name, i)
			e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_decode_field %s[] = {%s};", fields, strings.Join(values, ", ")))
		}
		children := "NULL"
		if len(n.Children) > 0 {
			values := []string{}
			for _, c := range n.Children {
				values = append(values, fmt.Sprint(c))
			}
			children = fmt.Sprintf("%s_children_%d", name, i)
			e.declarations = append(e.declarations, fmt.Sprintf("static const size_t %s[] = {%s};", children, strings.Join(values, ", ")))
		}
		literal := "NULL"
		if n.Kind == "literal" && n.Of == ir.String {
			literal = fmt.Sprintf("%s_literal_%d", name, i)
			e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", literal, cString(n.Literal)))
			literal = "&" + literal
		}
		nodes = append(nodes, fmt.Sprintf("{%s, %d, %s, %s, %s, %t, %d, %s, %d, %s, %s}", cString(n.Kind), n.Of, cString(n.Expected), literal, cNumber(n.Number), n.Boolean, len(n.Children), children, len(n.Fields), fields, cString(n.Discriminant)))
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_decode_node %s_nodes[] = {%s};", name, strings.Join(nodes, ", ")))
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_decode_schema %s = {%s_nodes, %d};", name, name, expression.Schema.Root))
	return e.own(ir.Object, fmt.Sprintf("adamic_json_decode(%s, &%s)", e.value(expression.Text), name))
}

// Optional fields compact decoded layouts, so the whole-program fixed-offset proof must defer
// their reads to the runtime shape cache.
func hasJsonDecode(program *ir.Program) bool {
	found := false
	walkExpressions(program, func(expression ir.Expression) {
		if _, ok := expression.(ir.JSONDecode); ok {
			found = true
		}
	})
	return found
}
