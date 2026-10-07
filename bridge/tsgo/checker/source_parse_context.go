package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The source file's parser module context, independent of any lint predicate.
func (p *Program) sourceParseContext(out *fields, node *ast.Node, question string) (string, error) {
	if question != "source-parse-context" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source-parse-context requires a SourceFile")
	}
	out.yes(ast.IsExternalModule(node.AsSourceFile()))
	return out.String(), nil
}
