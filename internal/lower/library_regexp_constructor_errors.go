package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// An invalid constructor is executable ECMAScript, unlike an invalid literal.
// Keep every original argument expression in order and let the runtime parser
// throw, preserving its Node message and ordinary exception cleanup paths.
func (l *lowering) regexpConstructorError(node *ast.Node, args []*ast.Node, values []ir.Expression) (ir.Expression, error) {
	arguments := make([]ir.Expression, 0, 2)
	for i, value := range values {
		if _, missing := value.(ir.Undefined); missing {
			value = ir.StringConstant{Index: l.constant("")}
		}
		if i == 0 && value.Type() == ir.Object && l.isLibraryType(l.checker.GetTypeAtLocation(args[i]), "RegExp") {
			value = ir.Property{Object: value, Name: "source", Of: ir.String}
		}
		if value.Type() != ir.String && value.Type() != ir.Maybe(ir.String) {
			return nil, l.notYet(node, "RegExp SyntaxError argument other than a string, undefined or intrinsic RegExp")
		}
		arguments = append(arguments, value)
	}
	for len(arguments) < 2 {
		arguments = append(arguments, ir.StringConstant{Index: l.constant("")})
	}
	return l.runtimeRegExpNew(arguments), nil
}
