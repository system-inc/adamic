package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A named-capture dictionary contains only strings (or indices arrays), never
// functions. Object.prototype shortcuts must not synthesize inherited methods.
func (l *lowering) regexGroupCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || !l.regexGroups(callee.AsPropertyAccessExpression().Expression) {
		return nil, false, nil
	}
	if callee.AsPropertyAccessExpression().QuestionDotToken != nil || node.AsCallExpression().QuestionDotToken != nil {
		return nil, true, l.notYet(node, "an optional named-group method call")
	}
	label, ok := regexCallLabel(callee)
	if !ok {
		return nil, true, l.notYet(node, "a named-group method call without a proven V8 diagnostic expression")
	}
	receiver, err := l.expression(callee.AsPropertyAccessExpression().Expression)
	if err != nil {
		return nil, true, err
	}
	values := []ir.Expression{receiver}
	for _, arg := range node.AsCallExpression().Arguments.Nodes {
		value, err := l.expression(arg)
		if err != nil {
			return nil, true, err
		}
		if _, absent := value.(ir.Undefined); !absent {
			values = append(values, value)
		}
	}
	returns, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	function, _ := l.stringHelper("regexp_group_noncallable", values)
	l.result.Functions[function].Returns = returns
	l.result.Functions[function].Body = []ir.Statement{ir.Throw{Value: ir.MakeError{
		Name:    ir.StringConstant{Index: l.constant("TypeError")},
		Message: ir.StringConstant{Index: l.constant(label + " is not a function")},
	}}}
	return ir.Call{Function: function, Arguments: values, Returns: returns}, true, nil
}

func regexCallLabel(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		return node.Text(), true
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		left, ok := regexCallLabel(node.AsPropertyAccessExpression().Expression)
		return left + "." + node.Name().Text(), ok
	}
	return "", false
}
