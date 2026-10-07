package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw call syntax and declaration-file provenance. No lint judgment is returned.
func (p *Program) callSymbolShape(out *fields, c *checker.Checker, _ *ast.SourceFile, node *ast.Node, question string) error {
	if question != "call-symbol-shape" || node.Kind != ast.KindCallExpression {
		return fmt.Errorf("call-symbol-shape requires a call")
	}
	out.number(uint64(len(node.Arguments())))
	expression := node.Expression()
	if expression != nil {
		expression = ast.SkipParentheses(expression)
	}
	// Keep the nullable callee kind separate from its presence.
	out.yes(expression != nil)
	if expression == nil {
		return nil
	}
	out.number(uint64(expression.Kind))
	text := ""
	if expression.Kind == ast.KindIdentifier {
		text = expression.Text()
	}
	out.text(text)
	symbol := c.GetSymbolAtLocation(expression)
	out.yes(symbol != nil)
	if symbol == nil {
		return nil
	}
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		out.yes(file != nil && file.IsDeclarationFile)
	}
	return nil
}
func init() { additionalQuestions["call-symbol-shape"] = (*Program).callSymbolShape }
