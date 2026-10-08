package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Port of V8 src/builtins/array-from.tq's iterable branch, with no mapper.
// String for...of yields Unicode code points, not UTF-16 index units.
func (l *lowering) libraryArrayFromIterable(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "from" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array") {
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 || written[0].Kind == ast.KindSpreadElement {
		return nil, false, nil
	}
	of, known := l.representation(l.checker.GetTypeAtLocation(written[0]))
	if !known || of != ir.String || l.includesUndefined(l.checker.GetTypeAtLocation(written[0])) || l.includesNull(l.checker.GetTypeAtLocation(written[0])) {
		return nil, false, nil
	}
	source, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{source})
	result := b.declare("result", ir.ArrayLiteral{Element: ir.String})
	item := b.local("element", ir.String)
	b.body = append(b.body, ir.ForOf{Iterable: b.read(b.parameters[0]), Element: ir.String, Local: item, Body: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: b.read(result), Value: b.read(item), Element: ir.String, Site: l.writeSite(node)}}}})
	return b.finish("array_from_string", b.read(result)), true, nil
}
