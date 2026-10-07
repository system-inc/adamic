package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) nodeRequireGlobal(node *ast.Node, name string) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) || node.Text() != name {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	file := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return load.IsPrelude(file) || load.IsNodeLibrary(file)
}

func (l *lowering) nodeRequireCall(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return "", false
	}
	call := node.AsCallExpression()
	if !l.nodeRequireGlobal(call.Expression, "require") {
		return "", false
	}
	if len(call.Arguments.Nodes) != 1 || call.Arguments.Nodes[0].Kind != ast.KindStringLiteral {
		return "", true
	}
	module, builtin := load.NodeBuiltin(call.Arguments.Nodes[0].Text())
	if !builtin {
		return "", true
	}
	return module, true
}

// Ambient host modules have no module body. Their immutable namespace bindings
// are resolved by the checker at each use, just as namespace imports are.
func (l *lowering) nodeRequireBinding(declaration *ast.Node) bool {
	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) {
		return false
	}
	module, call := l.nodeRequireCall(declaration.AsVariableDeclaration().Initializer)
	return call && (module == "node:fs" || module == "node:path" || module == "node:perf_hooks") && declaration.Parent.Flags&ast.NodeFlagsConst != 0
}

func (l *lowering) refuseNodeRequire(node *ast.Node) error {
	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression || callee.Kind == ast.KindElementAccessExpression {
			return l.refuseNodeRequire(callee)
		}
	}
	if module, call := l.nodeRequireCall(node); call {
		if module == "" {
			return &Refused{Where: l.program.Where(node), What: "require() without one string literal naming a Node builtin", Fix: "use an import"}
		}
		if module != "node:fs" && module != "node:path" && module != "node:perf_hooks" {
			return l.notYet(node, "require("+module+"): the builtin host module")
		}
		outer := node
		for outer.Parent != nil && (outer.Parent.Kind == ast.KindParenthesizedExpression || (module == "node:perf_hooks" && outer.Parent.Kind == ast.KindAsExpression)) {
			outer = outer.Parent
		}
		if module == "node:perf_hooks" && outer.Parent != nil && outer.Parent.Kind == ast.KindVariableDeclaration && outer.Parent.Name().Kind == ast.KindObjectBindingPattern {
			return nil
		}
		if outer.Parent == nil || !l.nodeRequireBinding(outer.Parent) {
			return l.notYet(node, "require("+module+") outside a plain const namespace binding; use an import")
		}
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if l.nodeRequireGlobal(access.Expression, "require") || l.nodeRequireGlobal(access.Expression, "module") {
			return &Refused{Where: l.program.Where(node), What: "CommonJS indexed access", Fix: "use an import (and named exports for module.exports)"}
		}
	}
	if node.Kind == ast.KindIdentifier && l.nodeRequireGlobal(node, "require") {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		if outer.Parent == nil || outer.Parent.Kind != ast.KindCallExpression || outer.Parent.AsCallExpression().Expression != outer {
			return &Refused{Where: l.program.Where(node), What: "require read as a value", Fix: "use an import"}
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if l.nodeRequireGlobal(access.Expression, "require") || (l.nodeRequireGlobal(access.Expression, "module") && access.Name().Text() == "exports") {
			return &Refused{Where: l.program.Where(node), What: "CommonJS " + access.Expression.Text() + "." + access.Name().Text(), Fix: "use an import (and named exports for module.exports)"}
		}
	}
	return nil
}

// Immediate performance destructuring cannot retain or write the module view.
// Only the unchanged declared member type, optionally absent, may be projected.
func (l *lowering) nodeRequirePerformanceProjection(node *ast.Node) bool {
	if node.Kind != ast.KindAsExpression {
		return false
	}
	module, call := l.nodeRequireCall(node.AsAsExpression().Expression)
	if !call || module != "node:perf_hooks" {
		return false
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	declaration := outer.Parent
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || declaration.Name().Kind != ast.KindObjectBindingPattern {
		return false
	}
	for _, binding := range declaration.Name().AsBindingPattern().Elements.Nodes {
		element := binding.AsBindingElement()
		name := binding.Name().Text()
		if element.PropertyName != nil {
			name = element.PropertyName.Text()
		}
		if name != "performance" || !ast.IsIdentifier(binding.Name()) || element.Initializer != nil || element.DotDotDotToken != nil {
			return false
		}
	}
	source := l.checker.GetTypeAtLocation(node.AsAsExpression().Expression)
	target := l.checker.GetTypeAtLocation(node)
	original := l.checker.GetPropertyOfType(source, "performance")
	viewed := l.checker.GetPropertyOfType(target, "performance")
	return original != nil && viewed != nil && l.withoutUndefined(l.checker.GetTypeOfSymbol(viewed)) == l.checker.GetTypeOfSymbol(original)
}
