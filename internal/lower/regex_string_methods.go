package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// An immediate intrinsic .call supplies this; it does not detach a method. Reuse the same
// RegExp path as a direct String call, preserving receiver, pattern and replacement order.
func (l *lowering) regexStringPrototypeCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "call" {
		return nil, false, nil
	}
	name, intrinsic := l.stringPrototypeMethod(callee.AsPropertyAccessExpression().Expression)
	if !intrinsic {
		return nil, false, nil
	}
	switch name {
	case "split", "match", "matchAll", "search", "replace", "replaceAll":
	default:
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) < 2 || !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(written[1])), "RegExp") {
		return nil, false, nil
	}
	of, _ := l.representation(l.checker.GetTypeAtLocation(written[0]))
	if of != ir.String {
		return nil, false, nil
	}
	if l.mayBeUndefined(written[0]) || l.includesNull(l.checker.GetTypeAtLocation(written[0])) {
		return nil, true, l.notYet(node, "String RegExp prototype call on a possibly null or undefined receiver")
	}
	return l.regexMethod(node, written[0], name, written[1:])
}
