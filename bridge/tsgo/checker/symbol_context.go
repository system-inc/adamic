package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Declaration identities and origins, without rule decisions.
func (p *Program) symbolContext(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbol-context" {
		return "", fmt.Errorf("unexpected symbol-context suffix")
	}
	s := c.GetSymbolAtLocation(node)
	out.number(p.symbolID(s))
	flags := uint64(0)
	count := 0
	if s != nil {
		flags = uint64(s.Flags)
		count = len(s.Declarations)
	}
	out.number(flags)
	out.number(uint64(count))
	if s != nil {
		for _, d := range s.Declarations {
			f := ast.GetSourceFileOfNode(d)
			if f == nil {
				return "", fmt.Errorf("declaration has no source")
			}
			out.text(f.FileName().AsString())
			out.text(strings.TrimPrefix(d.Kind.String(), "Kind"))
			out.number(uint64(d.Pos()))
			out.number(uint64(d.End()))
			out.yes(p.Compiler.IsSourceFileDefaultLibrary(f.PathKey()))
			out.yes(d.Kind == ast.KindNamespaceImport && ast.IsTypeOnlyImportOrExportDeclaration(d))
			parentFlags := uint64(0)
			if d.Parent != nil {
				parentFlags = uint64(d.Parent.Flags)
			}
			out.number(parentFlags)
		}
	}
	return out.String(), nil
}
