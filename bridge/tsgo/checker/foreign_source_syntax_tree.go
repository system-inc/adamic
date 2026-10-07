package checker

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Read an already-loaded virtual source through a real source anchor. No filesystem lookup.
func (p *Program) foreignSourceSyntaxTree(out *fields, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 || parts[0] != "foreign-source-syntax-tree" || node.Kind != ast.KindSourceFile || !utf8.ValidString(parts[1]) || strings.ContainsRune(parts[1], 0) {
		return "", fmt.Errorf("foreign-source-syntax-tree requires a SourceFile and loaded filename")
	}
	source := p.Compiler.GetSourceFile(parts[1])
	if source == nil {
		return "", fmt.Errorf("foreign source is not loaded: %s", parts[1])
	}
	return p.sourceSyntaxTree(out, source.AsNode(), "source-syntax-tree")
}
