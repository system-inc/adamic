package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// A factory proves the method ABI only while the closed program cannot replace
// it. Include wider object aliases: retaining the factory identity must not hide
// a write through an erased view of the same object.
func (l *lowering) generatorProtocolStable(where *ast.Node) error {
	proven := l.concrete(l.checker.GetTypeAtLocation(where))
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	related := func(node *ast.Node) bool {
		other := l.concrete(l.checker.GetTypeAtLocation(node))
		return l.checker.IsTypeAssignableTo(proven, other) || l.checker.IsTypeAssignableTo(other, proven)
	}
	var hazard *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if hazard != nil {
			return true
		}
		if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			target := ast.SkipParentheses(node.AsBinaryExpression().Left)
			switch target.Kind {
			case ast.KindPropertyAccessExpression:
				access := target.AsPropertyAccessExpression()
				name := access.Name().Text()
				if (name == "next" || name == "return" || name == "throw") && related(access.Expression) {
					hazard = target
					return true
				}
			case ast.KindElementAccessExpression:
				access := target.AsElementAccessExpression()
				name, known := constantStringKey(access.ArgumentExpression)
				if (!known || name == "next" || name == "return" || name == "throw") && related(access.Expression) {
					hazard = target
					return true
				}
			}
		}
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && len(call.Arguments.Nodes) > 0 {
				access := callee.AsPropertyAccessExpression()
				name := access.Name().Text()
				if l.isLibraryGlobal(access.Expression, "Object") && (name == "assign" || name == "defineProperty" || name == "defineProperties" || name == "setPrototypeOf") && related(call.Arguments.Nodes[0]) {
					hazard = node
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if hazard != nil {
		return l.notYet(hazard, "generator suspension protocol replacement needs a proved receiver convention")
	}
	return nil
}
