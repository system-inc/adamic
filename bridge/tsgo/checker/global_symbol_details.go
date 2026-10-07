package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The global symbol's declarations are facts, not a builtin-global verdict.
func (p *Program) globalSymbolDetails(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "global-symbol-details" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("global-symbol-details requires an Identifier")
	}
	p.writeSymbolDetails(out, c.GetGlobalSymbol(node.Text(), ast.SymbolFlagsAll, nil))
	return out.String(), nil
}

func init() {
	questionExtensions["global-symbol-details"] = globalSymbolDetailsQuestion
}
func globalSymbolDetailsQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	return p.globalSymbolDetails(out, c, node, question)
}
