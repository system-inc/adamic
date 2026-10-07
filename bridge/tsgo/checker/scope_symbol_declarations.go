package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) scopeSymbolDeclarations(out *fields, c *checker.Checker, _ *ast.SourceFile, node *ast.Node, question string) error {
	if question != "scope-symbol-declarations" {
		return fmt.Errorf("unexpected scope symbol suffix")
	}
	symbols := c.GetSymbolsInScope(node, ^ast.SymbolFlags(0))
	out.number(uint64(len(symbols)))
	for _, symbol := range symbols {
		out.text(symbol.Name)
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			source := ast.GetSourceFileOfNode(declaration)
			out.yes(source != nil)
			if source != nil {
				out.text(source.FileName())
				out.yes(source.IsDeclarationFile)
				out.number(uint64(declaration.Pos()))
				out.number(uint64(declaration.End()))
			}
		}
	}
	return nil
}

func init() { additionalQuestions["scope-symbol-declarations"] = (*Program).scopeSymbolDeclarations }
