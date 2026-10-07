package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw source metadata and option tristates. Native rules decide strictness.
func (p *Program) inspectGlobalSourceFacts(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "global-source-facts" {
		return p.inspectGlobalBindingFacts(c, node, question)
	}
	if node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("global-source-facts requires a SourceFile")
	}
	source := node.AsSourceFile()
	out := &fields{}
	out.number(1)
	out.text(question)
	out.yes(source.ExternalModuleIndicator != nil)
	out.yes(ast.IsSourceFileJS(source))
	out.number(uint64(p.Compiler.Options().Strict))
	out.number(uint64(p.Compiler.Options().AlwaysStrict))
	return out.String(), nil
}
