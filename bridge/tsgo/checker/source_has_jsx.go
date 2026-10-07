package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func init() { questionExtensions["source-has-jsx"] = sourceHasJsxQuestion }
func sourceHasJsxQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "source-has-jsx" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source-has-jsx requires SourceFile")
	}
	found := false
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindJsxElement || n.Kind == ast.KindJsxSelfClosingElement || n.Kind == ast.KindJsxFragment {
			found = true
			return true
		}
		return n.ForEachChild(visit)
	}
	visit(node)
	out.yes(found)
	return out.String(), nil
}
