package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Intrinsic key parameters may be unknown. Preserve the distinct null tag when boxing a
// nullable reference there, while keeping unrelated host receiver protocols unchanged.
func (l *lowering) collectionKeyContext(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if !l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		return false
	}
	switch callee.Name().Text() {
	case "set", "get", "has", "delete", "add":
		return true
	}
	return false
}

// Box each selected union arm using its own nullability, after saving effects once.
func (l *lowering) unionMember(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	own := l.concrete(l.checker.GetTypeAtLocation(node))
	if _, literal := value.(ir.Null); !literal && value.Type().IsReference() && value.Type() != ir.Union && l.includesNull(own) {
		if l.includesUndefined(own) {
			return nil, l.notYet(node, "a nullable lookup boxed as a union without its presence slot")
		}
		b := l.libraryArrayBuilder([]ir.Expression{value})
		held := b.read(b.parameters[0])
		return b.finish("nullable_union_key", ir.Conditional{
			Condition: ir.IsNull{Value: held}, WhenTrue: ir.Box{Value: ir.Null{}}, WhenNot: ir.Box{Value: held}, Of: ir.Union,
		}), nil
	}
	return fit(value, ir.Union), nil
}
