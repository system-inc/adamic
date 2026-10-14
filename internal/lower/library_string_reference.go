package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// These cases preserve the area's proven element and intrinsic-prototype representations.
func (l *lowering) stringReferenceValue(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if value.Type() == ir.Array && !l.includesUndefined(l.checker.GetTypeAtLocation(node)) && !l.includesNull(l.checker.GetTypeAtLocation(node)) {
		layer := l.checker.GetTypeAtLocation(node)
		for depth := 0; depth <= 32; depth++ {
			elementType := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(layer))
			if elementType == nil {
				break
			}
			element, known := l.kept(elementType)
			if !known {
				break
			}
			if element == ir.Array {
				layer = elementType
				continue
			}
			if element == ir.Number || element == ir.MaybeNumber || element == ir.String || element == ir.Boolean {
				return ir.ArrayJoin{Array: value, Separator: ir.StringConstant{Index: l.constant(",")}, Element: element, Depth: depth}, nil
			}
			break
		}
	}
	if value.Type() == ir.Map && !l.includesUndefined(l.checker.GetTypeAtLocation(node)) && !l.includesNull(l.checker.GetTypeAtLocation(node)) {
		// Maps and Sets have fixed intrinsic prototypes and refuse expando properties.
		text := "[object Map]"
		if !l.isLibraryType(l.checker.GetTypeAtLocation(node), "Map", "ReadonlyMap", "Set", "ReadonlySet") {
			return nil, l.notYet(node, "String conversion of a mixed Map and Set view")
		}
		if l.isLibraryType(l.checker.GetTypeAtLocation(node), "Set", "ReadonlySet") {
			text = "[object Set]"
		}
		function, _ := l.stringHelper("map_primitive", []ir.Expression{value})
		l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: ir.StringConstant{Index: l.constant(text)}}}
		return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
	}
	return nil, l.notYet(node, "String conversion of a nullable or unrepresented reference")
}

// The base converts call operands eagerly. Until the sibling's bound-value dispatch is
// present, refuse reference arguments whose ToPrimitive would run before later arguments
// or for an unused raw substitution. The sibling calls the value converter directly.
func (l *lowering) stringReferenceNeedsBoundArguments(node *ast.Node) bool {
	of, known := l.representation(l.checker.GetTypeAtLocation(node))
	if !known || (of != ir.Array && of != ir.Map) {
		return false
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.Name().Text() == "raw" && l.isLibraryGlobal(access.Expression, "String") {
		return true
	}
	if access.Name().Text() == "call" {
		_, intrinsic := l.stringPrototypeMethod(access.Expression)
		return intrinsic
	}
	if access.Name().Text() == "concat" {
		held, _ := l.representation(l.checker.GetTypeAtLocation(access.Expression))
		return held == ir.String
	}
	return false
}
