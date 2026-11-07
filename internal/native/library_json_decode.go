package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (nativeEmitter *emitter) jsonDecode(expression ir.JSONDecode) string {
	nativeEmitter.declarations = append(nativeEmitter.declarations, `#include "json_decode.h"`)
	name := nativeEmitter.jsonDecodeSchema(expression.Schema)
	return nativeEmitter.own(
		ir.Object,
		fmt.Sprintf("adamic_json_decode(%s, &%s)", nativeEmitter.value(expression.Text), name),
	)
}

// Preserve the decoder's existing C descriptor layout.
func (nativeEmitter *emitter) jsonDecodeSchema(schema ir.JSONDecodeSchema) string {
	return nativeEmitter.jsonDataSchema(schema, false)
}

// Both C descriptor views are emitted from the same checker-built schema graph.
func (nativeEmitter *emitter) jsonDataSchema(schema ir.JSONDecodeSchema, encode bool) string {
	prefix, fieldConst := "adamic_decode", "const "
	if encode {
		prefix, fieldConst = "adamic_encode", ""
	}
	name, exists := nativeEmitter.sharedDeclaration(prefix+"_schema", fmt.Sprintf("%#v", schema))
	if exists {
		return name
	}
	nodes := []string{}
	for index, node := range schema.Nodes {
		fields := "NULL"
		if len(node.Fields) > 0 {
			values := []string{}
			for _, field := range node.Fields {
				initializer := fmt.Sprintf("{%s, %d, %t}", cString(field.Name), field.Node, field.Optional)
				if encode {
					initializer = fmt.Sprintf(
						"{.name=%s, .node=%d, .optional=%t}",
						cString(field.Name),
						field.Node,
						field.Optional,
					)
				}
				values = append(values, initializer)
			}
			fields = fmt.Sprintf("%s_fields_%d", name, index)
			nativeEmitter.declarations = append(
				nativeEmitter.declarations,
				fmt.Sprintf("static %s%s_field %s[] = {%s};", fieldConst, prefix, fields, strings.Join(values, ", ")),
			)
		}
		children := "NULL"
		if len(node.Children) > 0 {
			values := []string{}
			for _, child := range node.Children {
				values = append(values, fmt.Sprint(child))
			}
			children = fmt.Sprintf("%s_children_%d", name, index)
			nativeEmitter.declarations = append(
				nativeEmitter.declarations,
				fmt.Sprintf("static const size_t %s[] = {%s};", children, strings.Join(values, ", ")),
			)
		}
		literal := "NULL"
		if node.Kind == "literal" && node.Of == ir.String {
			literal = fmt.Sprintf("%s_literal_%d", name, index)
			nativeEmitter.declarations = append(
				nativeEmitter.declarations,
				fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", literal, cString(node.Literal)),
			)
			literal = "&" + literal
		}
		kind := cString(node.Kind)
		if encode {
			kind = "adamic_encode_" + node.Kind
		}
		nodes = append(nodes, fmt.Sprintf(
			"{%s, %d, %s, %s, %s, %t, %d, %s, %d, %s, %s}",
			kind,
			node.Of,
			cString(node.Expected),
			literal,
			cNumber(node.Number),
			node.Boolean,
			len(node.Children),
			children,
			len(node.Fields),
			fields,
			cString(node.Discriminant),
		))
	}
	nativeEmitter.declarations = append(
		nativeEmitter.declarations,
		fmt.Sprintf("static const %s_node %s_nodes[] = {%s};", prefix, name, strings.Join(nodes, ", ")),
	)
	nativeEmitter.declarations = append(
		nativeEmitter.declarations,
		fmt.Sprintf("static const %s_schema %s = {%s_nodes, %d};", prefix, name, name, schema.Root),
	)
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
