package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Question modules register independently, so concurrent ports need no switch edits.
type additionalQuestion func(*Program, *fields, *checker.Checker, *ast.Node) error

var additionalQuestions = map[string]additionalQuestion{}

func (p *Program) additionalAnswer(out *fields, c *checker.Checker, node *ast.Node, question string) (string, bool, error) {
	mode := strings.SplitN(question, "\n", 2)[0]
	handler, ok := additionalQuestions[mode]
	if !ok {
		return "", false, nil
	}
	if question != mode {
		return "", true, fmt.Errorf("%s accepts no suffix", mode)
	}
	if err := handler(p, out, c, node); err != nil {
		return "", true, err
	}
	return out.String(), true, nil
}
