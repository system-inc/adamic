package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// symbolAncestry returns binder facts only. Native rules decide platform identity.
func (p *Program) symbolAncestry(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbol-ancestry" && question != "symbol-ancestry\nown" {
		return "", fmt.Errorf("unexpected symbol-ancestry suffix")
	}
	s := c.GetSymbolAtLocation(node)
	if s != nil && s.Flags&ast.SymbolFlagsAlias != 0 && question == "symbol-ancestry" {
		s = c.GetAliasedSymbol(s)
	}
	out := &fields{}
	out.number(1)
	out.text("symbol-ancestry")
	out.yes(s != nil)
	if s != nil {
		out.number(p.symbolID(s))
		out.number(uint64(s.Flags))
		out.number(uint64(len(s.Declarations)))
		for _, d := range s.Declarations {
			f := ast.GetSourceFileOfNode(d)
			if f == nil {
				return "", fmt.Errorf("symbol declaration has no source")
			}
			out.text(f.FileName().AsString())
			out.text(strings.TrimPrefix(d.Kind.String(), "Kind"))
			out.number(uint64(d.Pos()))
			out.number(uint64(d.End()))
			out.yes(f.IsDeclarationFile)
			out.yes(p.Compiler.IsSourceFileDefaultLibrary(f.PathKey()))
			out.yes(ast.IsExternalModule(f))
			var parents []*ast.Node
			for parent := d.Parent; parent != nil; parent = parent.Parent {
				parents = append(parents, parent)
			}
			out.number(uint64(len(parents)))
			for _, parent := range parents {
				out.text(strings.TrimPrefix(parent.Kind.String(), "Kind"))
				name := ""
				if n := parent.Name(); n != nil {
					name = n.Text()
				}
				out.text(name)
				out.number(uint64(parent.Flags))
				keyword := ""
				if parent.Kind == ast.KindModuleDeclaration {
					keyword = strings.TrimPrefix(parent.AsModuleDeclaration().Keyword.String(), "Kind")
				}
				out.text(keyword)
			}
		}
	}
	return out.String(), nil
}
