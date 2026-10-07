package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) inspectWave03Next(out *fields, c *checker.Checker, node *ast.Node, mode, question string) (string, error) {
	switch mode {
	case "source-module-context":
		return p.sourceModuleContext(out, node, question)
	case "symbol-context":
		return p.symbolContext(out, c, node, question)
	case "resolved-call-declaration":
		return p.resolvedCallDeclaration(out, c, node, question)
	case "code-path-graph":
		return p.codePathGraph(out, node, question)
	case "program-module-edges":
		return p.programModuleEdges(out, node, question)
	default:
		return p.inspectWave03More(out, c, node, mode, question)
	}
}
