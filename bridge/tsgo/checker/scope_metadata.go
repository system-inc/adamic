package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Raw binder container membership; native callers choose their own scope partition.
func (p *Program) scopeMetadata(out *fields, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "scope-metadata" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("scope-metadata requires a SourceFile")
	}
	out.yes(ast.IsExternalModule(source))
	var containers []*ast.Node
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if ast.IsLocalsContainer(n) {
			containers = append(containers, n)
		}
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(node)
	out.number(uint64(len(containers)))
	for _, n := range containers {
		out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
		out.number(uint64(n.Pos()))
		out.number(uint64(n.End()))
	}
	return out.String(), nil
}

func init() {
	questionExtensions["scope-metadata"] = scopeMetadataQuestion
}
func scopeMetadataQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	return p.scopeMetadata(out, source, node, question)
}
