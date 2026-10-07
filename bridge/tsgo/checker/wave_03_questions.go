package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The fallback keeps the common dispatcher edit to one line. Each question's
// fact implementation is isolated in its own file.
func (p *Program) inspectWave03(out *fields, c *checker.Checker, node *ast.Node, mode, question string) (string, error) {
	switch mode {
	case "signature-kinds":
		return p.signatureKinds(out, c, question)
	case "node-binding":
		return p.nodeBinding(out, c, node, question)
	case "import-runtime-options":
		return p.importRuntimeOptions(out, node, question)
	default:
		return p.inspectWave03Next(out, c, node, mode, question)
	}
}
