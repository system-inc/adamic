package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// hostBindingAbsentMemberRefusal keeps the constant process.browser read exact: no program may
// install that member, and run-time key operations never read an ordinary object slot. A structural
// view naming browser is still unregistered (nodeHostViewReach), before its reads or writes lower.
func (l *lowering) hostBindingAbsentMemberRefusal(node *ast.Node) error {
	stop := func() error { return l.notYet(node, "process.browser outside a plain property read") }
	if node.Kind == ast.KindDeleteExpression && l.processPath(node.AsDeleteExpression().Expression) == "process.browser" {
		return stop()
	}
	if node.Kind == ast.KindPropertyAccessExpression && l.processPath(node) == "process.browser" {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		if ast.IsAssignmentTarget(outer) || outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression {
			return stop()
		}
	}
	if node.Kind == ast.KindElementAccessExpression && l.processPath(node.AsElementAccessExpression().Expression) == "process" {
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if key.Kind != ast.KindStringLiteral || key.Text() == "browser" {
			return stop()
		}
	}
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword && l.processPath(node.AsBinaryExpression().Right) == "process" {
		key := ast.SkipParentheses(node.AsBinaryExpression().Left)
		if key.Kind != ast.KindStringLiteral || key.Text() == "browser" {
			return stop()
		}
	}
	if node.Kind == ast.KindSpreadAssignment && l.processPath(node.AsSpreadAssignment().Expression) == "process" {
		return stop()
	}
	if node.Kind == ast.KindVariableDeclaration && node.Initializer() != nil && l.processPath(node.Initializer()) == "process" && node.Name().Kind == ast.KindIdentifier {
		// Judge a later view before the generic process-as-a-value stop can hide its member name.
		own := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.Initializer()))
		if row, _ := l.hostBindingOf(node.Initializer()); row != nil && hostBindingRouted(*row) {
			if model := l.hostBindingModel(*row); model != nil {
				own = model
			}
		}
		view := l.checker.GetTypeAtLocation(node.Name())
		if l.checker.GetPropertyOfType(l.checker.GetNonNullableType(view), "browser") != nil {
			if reached := l.nodeHostViewAt(own, view); reached != "" {
				return l.nodeHostViewRefusal(node.Initializer(), own, view, reached)
			}
		}
	}
	if node.Kind == ast.KindVariableDeclaration && node.Initializer() != nil && l.processPath(node.Initializer()) == "process" && node.Name().Kind == ast.KindObjectBindingPattern {
		for _, binding := range node.Name().AsBindingPattern().Elements.Nodes {
			name := binding.AsBindingElement().PropertyName
			if name == nil {
				name = binding.Name()
			}
			// A computed or nested binding isn't a plain read of the absent member either.
			if name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral && name.Kind != ast.KindNumericLiteral || name.Text() == "browser" {
				return stop()
			}
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		if len(call.Arguments.Nodes) == 0 || l.processPath(call.Arguments.Nodes[0]) != "process" {
			return nil
		}
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return nil
		}
		access := callee.AsPropertyAccessExpression()
		if l.isLibraryGlobal(access.Expression, "Object") {
			switch access.Name().Text() {
			case "assign", "defineProperty", "defineProperties", "setPrototypeOf":
				return stop()
			}
		}
		if l.isLibraryGlobal(access.Expression, "Reflect") {
			switch access.Name().Text() {
			case "get", "set", "defineProperty", "deleteProperty", "setPrototypeOf":
				return stop()
			}
		}
	}
	return nil
}

func init() {
	refusalRules.register(70, "hostBindingAbsentMemberRefusal", (*lowering).hostBindingAbsentMemberRefusal)
}
