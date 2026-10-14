package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// evolvingObject finds a storage type for an unannotated, uninitialized local
// whose assignments all have proven object types. Explicit any and any-valued
// reads remain unsupported; the checker's flow types still decide every use.
func (l *lowering) evolvingObject(name *ast.Node) *checker.Type {
	if name.Parent == nil || name.Parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	declaration := name.Parent.AsVariableDeclaration()
	if declaration.Type != nil || declaration.Initializer != nil || l.checker.GetTypeAtLocation(name).Flags()&checker.TypeFlagsAny == 0 {
		return nil
	}
	root := name.Parent
	for root.Parent != nil && !ast.IsFunctionLike(root) {
		root = root.Parent
	}
	symbol := l.symbol(name)
	var inferred *checker.Type
	valid := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindBinaryExpression {
			binary := node.AsBinaryExpression()
			target := ast.SkipParentheses(binary.Left)
			if ast.IsIdentifier(target) && l.symbol(target) == symbol {
				_, compound := compoundAssignments[binary.OperatorToken.Kind]
				if binary.OperatorToken.Kind == ast.KindEqualsToken || compound || logicalAssignment(binary.OperatorToken.Kind) {
					proven := l.checker.GetTypeAtLocation(binary.Right)
					of, known := l.representation(proven)
					if binary.OperatorToken.Kind != ast.KindEqualsToken || !known || of != ir.Object {
						valid = false
					} else if inferred == nil {
						inferred = proven
					} else if inferred != proven {
						valid = false
					}
				}
			}
		}
		if node.Kind == ast.KindIdentifier && node != name && l.symbol(node) == symbol && l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsAny != 0 {
			// A write target can still have auto/any before its first assignment.
			if node.Parent == nil || node.Parent.Kind != ast.KindBinaryExpression || node.Parent.AsBinaryExpression().Left != node || node.Parent.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
				valid = false
			}
		}
		node.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	if !valid {
		return nil
	}
	return inferred
}
