package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// hostBindingProcessSyntax is the cheap gate for rule 70: only the process spelling, or
// globalThis.process / globalThis["process"], can need a host lookup. It proves no binding identity.
func hostBindingProcessSyntax(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for {
		node = ast.SkipParentheses(node)
		if node.Kind != ast.KindAsExpression {
			break
		}
		node = node.AsAsExpression().Expression
	}
	if ast.IsIdentifier(node) {
		if node.Text() == "process" {
			return node
		}
		return nil
	}
	var global *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if node.Name().Text() != "process" {
			return nil
		}
		global = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if (key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNoSubstitutionTemplateLiteral) || key.Text() != "process" {
			return nil
		}
		global = node.AsElementAccessExpression().Expression
	default:
		return nil
	}
	global = ast.SkipParentheses(global)
	if ast.IsIdentifier(global) && global.Text() == "globalThis" {
		return node
	}
	return nil
}

// hostBindingProcessReceiver proves the binding after the syntactic gate. globalThis's process is
// the host's too; keep these writes closed independently of the namespace-value refusal (#zqz01vh).
func (l *lowering) hostBindingProcessReceiver(node *ast.Node) bool {
	node = hostBindingProcessSyntax(node)
	if node == nil {
		return false
	}
	if ast.IsIdentifier(node) {
		return l.processPath(node) == "process"
	}
	return l.globalObject(ast.SkipParentheses(node.Expression()))
}

// hostBindingAbsentKey says whether a key names browser, or could at run time, and which it is.
func hostBindingAbsentKey(node *ast.Node) (browser, runtime bool) {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text() == "browser", false
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return false, false
	}
	return true, true
}

// hostBindingAbsentAccess extracts only an access that names browser or has a run-time key.
func hostBindingAbsentAccess(node *ast.Node) (receiver *ast.Node, runtime bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "browser" {
		return node.AsPropertyAccessExpression().Expression, false
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if browser, runtime := hostBindingAbsentKey(access.ArgumentExpression); browser {
			return access.Expression, runtime
		}
	}
	return nil, false
}

// hostBindingAbsentMemberRefusal keeps the constant process.browser read exact: no program may
// install that member, and run-time key operations never read an ordinary object slot. A structural
// view naming browser is still unregistered (nodeHostViewReach), before its reads or writes lower.
func (l *lowering) hostBindingAbsentMemberRefusal(node *ast.Node) error {
	stop := func(runtime bool) error {
		if runtime {
			return l.notYet(node, "a run-time key on process, which could name process.browser")
		}
		return l.notYet(node, "process.browser outside a plain property read")
	}
	if node.Kind == ast.KindDeleteExpression {
		if receiver, runtime := hostBindingAbsentAccess(node.AsDeleteExpression().Expression); l.hostBindingProcessReceiver(receiver) {
			return stop(runtime)
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "browser" {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		if ast.IsAssignmentTarget(outer) || outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression {
			if l.hostBindingProcessReceiver(node.AsPropertyAccessExpression().Expression) {
				return stop(false)
			}
		}
	}
	if node.Kind == ast.KindElementAccessExpression {
		if receiver, runtime := hostBindingAbsentAccess(node); l.hostBindingProcessReceiver(receiver) {
			return stop(runtime)
		}
	}
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword {
		binary := node.AsBinaryExpression()
		if browser, runtime := hostBindingAbsentKey(binary.Left); browser && l.hostBindingProcessReceiver(binary.Right) {
			return stop(runtime)
		}
	}
	if node.Kind == ast.KindSpreadAssignment && l.hostBindingProcessReceiver(node.AsSpreadAssignment().Expression) {
		return stop(false)
	}
	if node.Kind == ast.KindVariableDeclaration && node.Initializer() != nil && l.hostBindingProcessReceiver(node.Initializer()) {
		if node.Name().Kind == ast.KindIdentifier {
			// Judge a later view before the generic process-as-a-value stop hides its member name.
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
		if node.Name().Kind == ast.KindObjectBindingPattern {
			for _, binding := range node.Name().AsBindingPattern().Elements.Nodes {
				name := binding.AsBindingElement().PropertyName
				if name == nil {
					name = binding.Name()
				}
				// Computed and nested bindings aren't plain reads of the absent member either.
				if name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral && name.Kind != ast.KindNumericLiteral || name.Text() == "browser" {
					return stop(false)
				}
			}
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression || len(call.Arguments.Nodes) == 0 || !l.hostBindingProcessReceiver(call.Arguments.Nodes[0]) {
			return nil
		}
		access := callee.AsPropertyAccessExpression()
		if l.isLibraryGlobal(access.Expression, "Object") {
			switch access.Name().Text() {
			case "assign", "defineProperty", "defineProperties", "setPrototypeOf":
				return stop(false)
			}
		}
		if l.isLibraryGlobal(access.Expression, "Reflect") {
			switch access.Name().Text() {
			case "get", "set", "defineProperty", "deleteProperty", "setPrototypeOf":
				return stop(false)
			}
		}
	}
	return nil
}

func init() {
	refusalRules.register(70, "hostBindingAbsentMemberRefusal", (*lowering).hostBindingAbsentMemberRefusal)
}
