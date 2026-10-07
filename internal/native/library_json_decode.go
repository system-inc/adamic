package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) jsonDecode(expression ir.JSONDecode) string {
	e.declarations = append(e.declarations, `#include "json_decode.h"`)
	name := e.jsonDecodeSchema(expression.Schema)
	return e.own(ir.Object, fmt.Sprintf("adamic_json_decode(%s, &%s)", e.value(expression.Text), name))
}

// Preserve the decoder's existing C descriptor layout.
func (e *emitter) jsonDecodeSchema(schema ir.JSONDecodeSchema) string {
	return e.jsonDataSchema(schema, false)
}

// Both C descriptor views are emitted from the same checker-built schema graph.
func (e *emitter) jsonDataSchema(schema ir.JSONDecodeSchema, encode bool) string {
	prefix, fieldConst := "adamic_decode", "const "
	if encode {
		prefix, fieldConst = "adamic_encode", ""
	}
	name := e.temporary() + "_decode"
	nodes := []string{}
	for i, n := range schema.Nodes {
		fields := "NULL"
		if len(n.Fields) > 0 {
			values := []string{}
			for _, f := range n.Fields {
				initializer := fmt.Sprintf("{%s, %d, %t}", cString(f.Name), f.Node, f.Optional)
				if encode {
					initializer = fmt.Sprintf("{.name=%s, .node=%d, .optional=%t}", cString(f.Name), f.Node, f.Optional)
				}
				values = append(values, initializer)
			}
			fields = fmt.Sprintf("%s_fields_%d", name, i)
			e.declarations = append(e.declarations, fmt.Sprintf("static %s%s_field %s[] = {%s};", fieldConst, prefix, fields, strings.Join(values, ", ")))
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
		kind := cString(n.Kind)
		if encode {
			kind = "adamic_encode_" + n.Kind
		}
		nodes = append(nodes, fmt.Sprintf("{%s, %d, %s, %s, %s, %t, %d, %s, %d, %s, %s}", kind, n.Of, cString(n.Expected), literal, cNumber(n.Number), n.Boolean, len(n.Children), children, len(n.Fields), fields, cString(n.Discriminant)))
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const %s_node %s_nodes[] = {%s};", prefix, name, strings.Join(nodes, ", ")))
	e.declarations = append(e.declarations, fmt.Sprintf("static const %s_schema %s = {%s_nodes, %d};", prefix, name, name, schema.Root))
	return name
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
