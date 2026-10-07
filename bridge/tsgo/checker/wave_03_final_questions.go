package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) inspectWave03Final(out *fields, c *checker.Checker, node *ast.Node, mode, question string) (string, error) {
	switch mode {
	case "generic-signature-shape":
		return p.genericSignatureShape(out, c, node, question)
	case "indexed-type-shape":
		return p.indexedTypeShape(out, c, node, question)
	case "callable-return-shape":
		return p.callableReturnShape(out, c, node, question)
	default:
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
}
