package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// globalSource exposes module classification, source language and the resolved
// compiler option. Directive and ancestor strictness remain native decisions.
func (p *Program) globalSource(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "global-source" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("global-source requires SourceFile without suffix")
	}
	source := node.AsSourceFile()
	out.yes(source.ExternalModuleIndicator != nil)
	out.yes(ast.IsSourceFileJS(source))
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().AlwaysStrict))
	return out.String(), nil
}
