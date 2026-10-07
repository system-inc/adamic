package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func (p *Program) symbolOrigins(out *fields, c *checker.Checker, node *ast.Node) error {
	symbol := c.GetSymbolAtLocation(node)
	out.number(p.symbolID(symbol))
	if symbol == nil {
		out.number(0)
		out.text("")
		out.number(0)
		return nil
	}
	out.number(uint64(symbol.Flags))
	out.text(strings.ToValidUTF8(symbol.Name, "�"))
	out.number(uint64(len(symbol.Declarations)))
	for _, decl := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(decl)
		if file == nil {
			return fmt.Errorf("symbol declaration without source")
		}
		out.text(strings.TrimPrefix(decl.Kind.String(), "Kind"))
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.Path()))
	}
	return nil
}

func init() { additionalQuestions["symbol-origins"] = (*Program).symbolOrigins }
