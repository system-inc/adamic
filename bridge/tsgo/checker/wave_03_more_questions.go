package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) inspectWave03More(out *fields, c *checker.Checker, node *ast.Node, mode, question string) (string, error) {
	switch mode {
	case "source-access-context":
		return p.sourceAccessContext(out, node, question)
	default:
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
}
