package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func (p *Program) processQuestions(c *checker.Checker, node *ast.Node, question string) (string, error) {
	mode := strings.Split(question, "\n")[0]
	switch mode {
	case "syntax-reference-facts":
		return p.syntaxReferenceFacts(c, node, question)
	case "syntax-projection":
		return p.syntaxProjection(node, question)
	case "syntax-control-flow":
		return p.syntaxControlFlow(node, question)
	case "process-symbol-details":
		return p.processSymbolDetails(c, node, question)
	case "resolved-call-declaration":
		return p.resolvedCallDeclaration(c, node, question)
	case "program-imports":
		return p.programImports(node, question)
	}
	return "", fmt.Errorf("unsupported checker question: %s", question)
}
