package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (l *lowering) intrinsicUndefined(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return ast.IsIdentifier(node) && l.checker.IsUndefinedSymbol(l.checker.GetSymbolAtLocation(node))
}

// A flow-narrowed undefined type is not a compile-time value: a called closure
// can change a mutable binding. Resolve only the intrinsic or stable aliases.
func (l *lowering) constantUndefined(node *ast.Node, depth int) bool {
	if depth > 32 {
		return false
	}
	node = ast.SkipParentheses(node)
	if l.intrinsicUndefined(node) && l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	d := symbol.Declarations[0]
	if d.Kind != ast.KindVariableDeclaration || d.Parent == nil || d.AsVariableDeclaration().Initializer == nil || !l.regexStableBinding(symbol, d) {
		return false
	}
	return l.constantUndefined(d.AsVariableDeclaration().Initializer, depth+1)
}

// Resolve a RegExp's immutable source and flags, while construction still
// evaluates its input object at runtime and starts the new lastIndex at zero.
func (l *lowering) constantRegExp(node *ast.Node, depth int) (string, string, bool) {
	if depth > 32 {
		return "", "", false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNewExpression || node.Kind == ast.KindCallExpression {
		var callee *ast.Node
		var args []*ast.Node
		if node.Kind == ast.KindNewExpression {
			n := node.AsNewExpression()
			callee = n.Expression
			if n.Arguments != nil {
				args = n.Arguments.Nodes
			}
		} else {
			n := node.AsCallExpression()
			callee = n.Expression
			if n.Arguments != nil {
				args = n.Arguments.Nodes
			}
		}
		if !l.isLibraryGlobal(callee, "RegExp") || len(args) > 2 {
			return "", "", false
		}
		pattern, flags := "", ""
		if len(args) > 0 {
			var ok bool
			if l.constantUndefined(args[0], depth+1) {
				ok = true
			} else if l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(args[0])), "RegExp") {
				pattern, flags, ok = l.constantRegExp(args[0], depth+1)
			} else {
				pattern, ok = l.constantPattern(args[0], depth+1)
			}
			if !ok {
				return "", "", false
			}
		}
		if len(args) > 1 && !l.constantUndefined(args[1], depth+1) {
			var ok bool
			flags, ok = l.constantPattern(args[1], depth+1)
			if !ok {
				return "", "", false
			}
		}
		return pattern, flags, true
	}
	if node.Kind == ast.KindRegularExpressionLiteral {
		text := node.Text()
		for i := len(text) - 1; i > 0; i-- {
			if text[i] == '/' {
				return text[1:i], text[i+1:], true
			}
		}
	}
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			d := symbol.Declarations[0]
			if d.Kind == ast.KindVariableDeclaration && d.Parent != nil && d.AsVariableDeclaration().Initializer != nil && l.regexStableBinding(symbol, d) {
				return l.constantRegExp(d.AsVariableDeclaration().Initializer, depth+1)
			}
		}
	}
	return "", "", false
}

// A let-bound primitive can be compiled just like const when no write in any
// reachable module can change it. Scan all syntax, including closures and dead
// branches: an uncertain write refuses compilation instead of freezing a value.
// Arguments are still evaluated at runtime, retaining TDZ and evaluation order.
func (l *lowering) regexStableBinding(symbol *ast.Symbol, declaration *ast.Node) bool {
	if declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
		return true
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	written := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if written {
			return true
		}
		var target *ast.Node
		switch node.Kind {
		case ast.KindBinaryExpression:
			if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
				target = node.AsBinaryExpression().Left
			}
		case ast.KindPrefixUnaryExpression:
			u := node.AsPrefixUnaryExpression()
			if u.Operator == ast.KindPlusPlusToken || u.Operator == ast.KindMinusMinusToken {
				target = u.Operand
			}
		case ast.KindPostfixUnaryExpression:
			target = node.AsPostfixUnaryExpression().Operand
		case ast.KindForInStatement, ast.KindForOfStatement:
			target = node.AsForInOrOfStatement().Initializer
		}
		if target != nil {
			var check ast.Visitor
			check = func(part *ast.Node) bool {
				if ast.IsIdentifier(part) && l.symbol(part) == symbol {
					written = true
					return true
				}
				return part.ForEachChild(check)
			}
			check(target)
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		visit(module.AsNode())
	}
	return !written
}
