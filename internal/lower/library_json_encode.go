package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) encodeJson(node *ast.Node) (ir.Expression, bool, error) {
	if !l.isPreludeFunction(node.AsCallExpression().Expression, "encodeJson") {
		return nil, false, nil
	}
	// Use exactly the decoder's schema and refusal policy. There is one type walker.
	schema, err := l.jsonSchemaType(node, jsonSchemaEncode)
	if err != nil {
		return nil, true, err
	}
	// A structural view can widen an unboxed scalar slot to a boxed union.
	// Existing layouts record reference ownership, not complete scalar types.
	// This unit deliberately uses compile-time NotYet for mixed scalar unions
	// inside containers; runtime shape metadata changes belong to another unit.
	for _, parent := range schema.Nodes {
		if parent.Kind == "object" || parent.Kind == "tuple" {
			for _, field := range parent.Fields {
				if schema.Nodes[field.Node].Of == ir.Union {
					return nil, true, l.notYet(node, "encodeJson mixed scalar union fields require complete runtime slot metadata")
				}
			}
		}
		if parent.Kind == "array" && schema.Nodes[parent.Children[0]].Of == ir.Union {
			return nil, true, l.notYet(node, "encodeJson mixed scalar union elements require complete runtime slot metadata")
		}
	}
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "encodeJson requires one value argument")
	}
	value, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	return ir.JSONEncode{Value: fit(value, schema.Nodes[schema.Root].Of), Schema: schema}, true, nil
}
