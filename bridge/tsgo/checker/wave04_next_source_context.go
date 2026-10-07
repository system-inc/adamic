package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Raw top-level parse context flags, including TypeScript's await reparse.
func (p *Program) wave04NextSourceContext(out *fields, node *ast.Node, question string) (string, error) {
	if question != "wave04-next-source-context" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source context requires a source file")
	}
	statements := node.AsSourceFile().Statements.Nodes
	out.number(uint64(len(statements)))
	for _, statement := range statements {
		out.number(uint64(statement.Pos()))
		out.number(uint64(statement.End()))
		out.number(uint64(statement.Flags))
		awaits := false
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			if ast.IsFunctionLike(n) {
				return false
			}
			if n.Kind == ast.KindAwaitExpression {
				awaits = true
			}
			n.ForEachChild(visit)
			return false
		}
		visit(statement)
		out.yes(awaits)
	}
	return out.String(), nil
}
