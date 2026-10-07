package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Optional syntax slots, including absent for-statement header expressions.
func (p *Program) loopHeader(out *fields, _ *checker.Checker, _ *ast.SourceFile, node *ast.Node, question string) error {
	if question != "loop-header" || node.Kind != ast.KindForStatement {
		return fmt.Errorf("loop-header requires a for statement")
	}
	statement := node.AsForStatement()
	for _, slot := range []*ast.Node{statement.Initializer, statement.Condition, statement.Incrementor} {
		out.yes(slot != nil)
		if slot != nil {
			out.text(strings.TrimPrefix(slot.Kind.String(), "Kind"))
			out.number(uint64(slot.Pos()))
			out.number(uint64(slot.End()))
		}
	}
	return nil
}
func init() { additionalQuestions["loop-header"] = (*Program).loopHeader }
