package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) inspectWave03React(out *fields, c *checker.Checker, node *ast.Node, mode, question string) (string, error) {
	switch mode {
	case "foreign-source-syntax-tree":
		return p.foreignSourceSyntaxTree(out, node, question)
	case "foreign-node-symbol-context":
		return p.foreignNodeSymbolContext(out, c, node, question)
	case "source-token-start":
		return p.sourceTokenStart(out, node, question)
	case "source-syntax-tree":
		return p.sourceSyntaxTree(out, node, question)
	default:
		return p.inspectWave03Final(out, c, node, mode, question)
	}
}
