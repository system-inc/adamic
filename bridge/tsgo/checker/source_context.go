package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) sourceContext(_ *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "source-context"
	if question != mode || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source-context requires a SourceFile without suffix")
	}
	source := node.AsSourceFile()
	out := &fields{}
	out.number(1)
	out.text(mode)
	out.yes(ast.IsExternalModule(source))
	out.yes(source.IsDeclarationFile)
	return out.String(), nil
}
