package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Extensions register from their own files; Inspect's default arm is the only
// shared edit needed to admit independently owned questions.
var additionalQuestions = map[string]func(*Program, *fields, *checker.Checker, *ast.SourceFile, *ast.Node, string) error{}

func inspectAdditionalQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, mode, question string) (string, error) {
	answer := additionalQuestions[mode]
	if answer == nil {
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	if err := answer(p, out, c, source, node, question); err != nil {
		return "", err
	}
	return out.String(), nil
}
