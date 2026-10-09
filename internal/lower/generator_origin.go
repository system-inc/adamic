package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The static Generator interface does not encode the native receiver ABI. Keep
// opaque implementations stopped until protocol views preserve that ABI.
func (l *lowering) generatorOrigin(node *ast.Node, depth int) bool {
	if node == nil || depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindCallExpression:
		signature := l.checker.GetResolvedSignature(node)
		if signature == nil || signature.Declaration() == nil {
			return false
		}
		declaration := signature.Declaration()
		if isGenerator(declaration) {
			return true
		}
		if declaration.Body() == nil {
			return false
		}
		seen, known := false, true
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if ast.IsFunctionLike(n) {
				return false
			}
			if n.Kind == ast.KindReturnStatement {
				seen = true
				known = known && l.generatorOrigin(n.AsReturnStatement().Expression, depth+1)
			}
			return n.ForEachChild(visit)
		}
		declaration.Body().ForEachChild(visit)
		return seen && known
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration {
			return false
		}
		if declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			for parent := declaration.Parent; parent != nil; parent = parent.Parent {
				if ast.IsFunctionLike(parent) {
					return false
				}
			}
			modules, err := l.moduleOrder(l.program.Files()[0])
			if err != nil {
				return false
			}
			known, seen := true, false
			if initial := declaration.AsVariableDeclaration().Initializer; initial != nil {
				known = l.checker.GetTypeAtLocation(initial).Flags()&checker.TypeFlagsUndefined != 0 || l.generatorOrigin(initial, depth+1)
				seen = true
			}
			var visit ast.Visitor
			visit = func(n *ast.Node) bool {
				if n.Kind == ast.KindBinaryExpression && n.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
					b := n.AsBinaryExpression()
					left := ast.SkipParentheses(b.Left)
					if ast.IsIdentifier(left) && l.symbol(left) == symbol {
						seen = true
						known = known && (l.checker.GetTypeAtLocation(b.Right).Flags()&checker.TypeFlagsUndefined != 0 || l.generatorOrigin(b.Right, depth+1))
					}
				}
				return n.ForEachChild(visit)
			}
			for _, module := range modules {
				module.AsNode().ForEachChild(visit)
			}
			return known && seen
		}
		return l.generatorOrigin(declaration.AsVariableDeclaration().Initializer, depth+1)
	case ast.KindConditionalExpression:
		c := node.AsConditionalExpression()
		return l.generatorOrigin(c.WhenTrue, depth+1) && l.generatorOrigin(c.WhenFalse, depth+1)
	}
	return false
}
