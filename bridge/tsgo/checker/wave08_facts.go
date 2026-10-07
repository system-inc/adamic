package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Each question owns its implementation; shared files contain only registration.
func (p *Program) wave08Facts(out *fields, c *checker.Checker, node *ast.Node, mode, question string) error {
	switch mode {
	case "wave08-symbol":
		return p.wave08Symbol(out, c, node, question)
	case "syntax-metadata":
		return p.syntaxMetadata(out, node, question)
	case "module-links":
		return p.moduleLinks(out, node, question)
	}
	return nil
}
