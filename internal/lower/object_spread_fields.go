package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The added-field path keeps actual own presence, including an empty undefined
// source. The legacy same-shape path still synthesizes slots and keeps its guard.
func (l *lowering) objectSpreadExtends(node *ast.Node) bool {
	properties := node.AsObjectLiteralExpression().Properties.Nodes
	if len(properties) == 0 || properties[0].Kind != ast.KindSpreadAssignment {
		return false
	}
	source := properties[0].AsSpreadAssignment().Expression
	for _, property := range properties[1:] {
		name, known := l.methodName(property)
		if !known {
			continue
		}
		if !l.hasProperty(source, name) || l.declaredField(node, name) == ir.Union {
			return true
		}
		if property.Kind == ast.KindPropertyAssignment {
			if held, known := l.representation(l.checker.GetTypeAtLocation(property.AsPropertyAssignment().Initializer)); known && held == ir.Union {
				return true
			}
		}
	}
	return false
}

// Runtime shape names are C strings. A hidden NUL-bearing field cannot be merged
// by name without losing its identity; keep this conservative until names carry lengths.
func (l *lowering) objectSpreadNamesKnown() bool {
	known := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !known {
			return true
		}
		if name := node.Name(); name != nil && name.Kind == ast.KindStringLiteral && strings.ContainsRune(name.Text(), 0) {
			known = false
			return true
		}
		node.ForEachChild(visit)
		return false
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return known
}
