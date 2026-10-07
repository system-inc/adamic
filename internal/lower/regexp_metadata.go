package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Only direct intrinsic observations are admitted. No receiver evaluation is
// discarded, no method value escapes, and shadowed constructors never qualify.
func (l *lowering) regexIntrinsicMethod(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return "", false
	}
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	if receiver.Kind != ast.KindPropertyAccessExpression || receiver.Name().Text() != "prototype" || !l.isLibraryGlobal(receiver.AsPropertyAccessExpression().Expression, "RegExp") {
		return "", false
	}
	switch name := node.Name().Text(); name {
	case "exec", "test", "toString", "compile":
		return name, true
	}
	return "", false
}
func (l *lowering) regexMethodObservation(node *ast.Node) bool {
	if _, ok := l.regexIntrinsicMethod(node); !ok {
		return false
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	switch parent.Name().Text() {
	case "length", "name", "prototype":
		p := parent.Parent
		if p != nil && p.Kind == ast.KindBinaryExpression && p.AsBinaryExpression().Left == parent && ast.IsAssignmentOperator(p.AsBinaryExpression().OperatorToken.Kind) {
			return false
		}
		if p != nil && (p.Kind == ast.KindPrefixUnaryExpression || p.Kind == ast.KindPostfixUnaryExpression) {
			return false
		}
		return true
	case "hasOwnProperty", "propertyIsEnumerable":
		return called(parent)
	}
	return false
}
func (l *lowering) regexMethodMetadata(node *ast.Node) (ir.Expression, bool) {
	access := node.AsPropertyAccessExpression()
	method, ok := l.regexIntrinsicMethod(access.Expression)
	if !ok || access.QuestionDotToken != nil {
		return nil, false
	}
	switch access.Name().Text() {
	case "length":
		n := 1.0
		if method == "compile" {
			n = 2
		}
		if method == "toString" {
			n = 0
		}
		return ir.NumberConstant{Value: n}, true
	case "name":
		return ir.StringConstant{Index: l.constant(method)}, true
	case "prototype":
		return ir.Undefined{}, true
	}
	return nil, false
}
func (l *lowering) regexMethodMetadataCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	_, ok := l.regexIntrinsicMethod(callee.AsPropertyAccessExpression().Expression)
	if !ok {
		return nil, false, nil
	}
	name := callee.Name().Text()
	if name != "hasOwnProperty" && name != "propertyIsEnumerable" {
		return nil, false, nil
	}
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) != 1 || ast.SkipParentheses(args[0]).Kind != ast.KindStringLiteral {
		return nil, true, l.notYet(node, "intrinsic RegExp method metadata with an evaluated key")
	}
	key := ast.SkipParentheses(args[0]).Text()
	return ir.BooleanConstant{Value: name == "hasOwnProperty" && (key == "length" || key == "name")}, true, nil
}
