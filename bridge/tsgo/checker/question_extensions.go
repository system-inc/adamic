package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Question files register during package initialization. Requests only read the
// table, after the program has selected and leased its exact source and checker.
type questionExtension func(*Program, *fields, *checker.Checker, *ast.SourceFile, *ast.Node, string) (string, error)

var questionExtensions = map[string]questionExtension{}

func (p *Program) inspectExtension(out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if answer := questionExtensions[strings.Split(question, "\n")[0]]; answer != nil {
		return answer(p, out, c, source, node, question)
	}
	return "", fmt.Errorf("unsupported checker question: %s", question)
}
