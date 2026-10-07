package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Only conditional expressions move to cohere in this unit. Other invariant
// views and all nominal views retain Adamic's proofs and diagnostics.
func (l *lowering) refuseConditionalCopies(module *ast.SourceFile) error {
	sites := map[[2]int]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.SkipParentheses(node).Kind == ast.KindConditionalExpression && !l.stringReadOnlyArgument(node) {
			line, column := scanner.GetLineAndCharacterOfPosition(module, scanner.GetTokenPosOfNode(node, module, false))
			sites[[2]int{line + 1, column + 1}] = true
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	if len(sites) == 0 {
		return nil
	}
	findings, err := (cohereRuleRunner{}).RunRule(l.program.CompilerProgram(), l.checker, []*ast.SourceFile{module}, "adamic/invariant-mutable")
	if err != nil {
		return err
	}
	for _, finding := range findings {
		if !sites[[2]int{finding.Line, finding.Column}] {
			continue
		}
		return &Refused{
			Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), finding.Line, finding.Column),
			What:  strings.Join(strings.Fields(finding.Message), " "),
			Fix:   "use a readonly view or an unconditional copy (" + finding.Rule + ")",
		}
	}
	return nil
}
