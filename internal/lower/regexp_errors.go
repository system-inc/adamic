package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

var builtinErrorNames = []string{"Error", "TypeError", "SyntaxError", "RangeError", "ReferenceError", "EvalError", "URIError"}

func (l *lowering) errorKind(node *ast.Node) (int, bool) {
	if node == nil {
		return 0, false
	}
	for kind, name := range builtinErrorNames {
		if l.isLibraryGlobal(node, name) {
			return kind, true
		}
	}
	return 0, false
}

func (l *lowering) errorBuiltin(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind == ast.KindNewExpression {
		kind, known := l.errorKind(node.AsNewExpression().Expression)
		if !known || kind == 0 {
			return nil, false, nil
		}
		value, err := l.newError(node)
		if err != nil {
			return nil, true, err
		}
		return ir.BuiltinError{Message: value.(ir.MakeError).Message, Kind: kind}, true, nil
	}
	if node.Kind != ast.KindBinaryExpression {
		return nil, false, nil
	}
	binary := node.AsBinaryExpression()
	left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
	exact := binary.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken
	if exact && right.Kind == ast.KindPropertyAccessExpression {
		left, right = right, left
	}
	kind, known := l.errorKind(right)
	if !known {
		return nil, false, nil
	}
	operand := left
	if exact {
		if left.Kind != ast.KindPropertyAccessExpression || left.Name().Text() != "constructor" {
			return nil, false, nil
		}
		operand = ast.SkipParentheses(left.AsPropertyAccessExpression().Expression)
	} else if binary.OperatorToken.Kind != ast.KindInstanceOfKeyword {
		return nil, false, nil
	}
	// Only catches are proven to hold builtin errors; user classes and arbitrary
	// structural Error objects must use ordinary lowering or be refused.
	if !ast.IsIdentifier(operand) || !l.caught[l.symbol(operand)] {
		return nil, false, nil
	}
	local, _ := l.local(operand)
	value := ir.Expression(ir.ErrorIs{Value: ir.Read{Local: local, Of: ir.Object, Checked: l.checked(local)}, Kind: kind, Exact: exact})
	if binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
		value = ir.Unary{Operator: ir.Not, Operand: value}
	}
	return value, true, nil
}

// V8's diagnostic wording is not part of Pattern semantics. Until those
// messages are implemented, do not compile a constructor whose thrown message
// could be observed anywhere in its module graph.
func (l *lowering) regexErrorMessageObserved() bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	caught := map[*ast.Symbol]bool{}
	var collect ast.Visitor
	collect = func(node *ast.Node) bool {
		if node.Kind == ast.KindCatchClause {
			if declaration := node.AsCatchClause().VariableDeclaration; declaration != nil && ast.IsIdentifier(declaration.Name()) {
				caught[l.symbol(declaration.Name())] = true
			}
		}
		return node.ForEachChild(collect)
	}
	for _, module := range modules {
		collect(module.AsNode())
	}
	observed := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if observed {
			return true
		}
		if ast.IsIdentifier(node) && caught[l.symbol(node)] {
			parent := node.Parent
			safe := parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.Name() == node
			if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node {
				safe = parent.Name().Text() == "constructor" || parent.Name().Text() == "name"
			}
			if parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindInstanceOfKeyword && parent.AsBinaryExpression().Left == node {
				safe = true
			}
			if parent != nil && parent.Kind == ast.KindTypeOfExpression {
				safe = true
			}
			if !safe {
				observed = true
				return true
			}
		}
		if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "message" {
			receiver := node.AsPropertyAccessExpression().Expression
			if l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Error", "TypeError", "SyntaxError", "RangeError", "ReferenceError", "EvalError", "URIError") {
				observed = true
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		visit(module.AsNode())
	}
	return observed
}
