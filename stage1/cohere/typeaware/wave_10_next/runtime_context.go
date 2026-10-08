// Package wave10next supplies raw compiler observations for the continuation.
// It has no lint messages, control-flow decisions or fixes.
package wave10next

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
)

type fields struct{ strings.Builder }

func (f *fields) text(s string)   { fmt.Fprintf(&f.Builder, "%d\n%s", len(utf16.Encode([]rune(s))), s) }
func (f *fields) number(n uint64) { f.text(strconv.FormatUint(n, 10)) }
func (f *fields) yes(b bool) {
	if b {
		f.text("1")
	} else {
		f.text("0")
	}
}
func (f *fields) declaration(p *compiler.Program, n *ast.Node) {
	file := ast.GetSourceFileOfNode(n)
	f.text(file.FileName().AsString())
	f.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
	f.number(uint64(n.Pos()))
	f.number(uint64(n.End()))
	f.yes(file.IsDeclarationFile)
	f.yes(p.IsSourceFileDefaultLibrary(file.PathKey()))
	f.yes(ast.IsExternalModule(file))
	f.text(file.Text())
	f.number(uint64(n.Flags))
	f.number(uint64(ast.GetFunctionFlags(n)))
	ancestors := []*ast.Node{}
	for current := n.Parent; current != nil; current = current.Parent {
		ancestors = append(ancestors, current)
	}
	f.number(uint64(len(ancestors)))
	for _, a := range ancestors {
		f.text(strings.TrimPrefix(a.Kind.String(), "Kind"))
		name := ""
		if a.Name() != nil && (ast.IsIdentifier(a.Name()) || ast.IsStringLiteralLike(a.Name()) || a.Name().Kind == ast.KindNumericLiteral) {
			name = a.Name().Text()
		}
		f.text(name)
		f.number(uint64(a.Flags))
		f.number(uint64(a.Pos()))
		f.number(uint64(a.End()))
		f.yes(ast.IsGlobalScopeAugmentation(a))
		nameKind := ""
		if a.Name() != nil {
			nameKind = strings.TrimPrefix(a.Name().Kind.String(), "Kind")
		}
		f.text(nameKind)
	}
}

// Inspect accepts an AST anchor already validated by the bridge. Its modes
// return only compiler facts; all lint classification belongs in Adamic.
func Inspect(p *compiler.Program, file *ast.SourceFile, selected *ast.Node, c *checker.Checker, question string) (string, error) {
	pieces := strings.Split(question, "\n")
	if len(pieces) != 2 || pieces[0] != "runtime-context" {
		return "", fmt.Errorf("invalid runtime context question")
	}
	mode := pieces[1]
	if mode != "origin" && mode != "alias-origin" && mode != "signature" && mode != "program" && mode != "syntax" {
		return "", fmt.Errorf("unknown runtime context mode")
	}
	if p == nil || file == nil || selected == nil || ast.GetSourceFileOfNode(selected) != file {
		return "", fmt.Errorf("runtime anchor does not belong to source")
	}

	if mode != "program" && c == nil {
		return "", fmt.Errorf("missing leased checker")
	}

	out := &fields{}
	out.number(1)
	out.text("runtime-context")
	out.text(mode)
	switch mode {
	case "syntax":
		syntaxFields(out, selected)
	case "origin", "alias-origin":
		symbol := c.GetSymbolAtLocation(selected)
		if mode == "alias-origin" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = c.GetAliasedSymbol(symbol)
		}
		out.yes(symbol != nil)
		if symbol != nil {
			out.number(uint64(symbol.Flags))
			out.text(symbol.Name)
			out.number(uint64(len(symbol.Declarations)))
			for _, d := range symbol.Declarations {
				out.declaration(p, d)
			}
		}
	case "signature":
		if selected.Kind != ast.KindCallExpression {
			return "", fmt.Errorf("runtime signature requires CallExpression")
		}
		sig := c.GetResolvedSignature(selected)
		out.yes(sig != nil && sig.Declaration() != nil)
		if sig != nil && sig.Declaration() != nil {
			out.declaration(p, sig.Declaration())
			t := c.GetReturnTypeOfSignature(sig)
			flags := uint64(0)
			if t != nil {
				flags = uint64(t.Flags())
			}
			out.number(flags)
		}
	case "program":
		files := p.GetSourceFiles()
		out.number(uint64(len(files)))
		for _, source := range files {
			out.text(source.FileName().AsString())
			out.yes(source.IsDeclarationFile)
			out.yes(ast.IsExternalModule(source))
			out.text(source.Text())
			type edge struct {
				kind      string
				typeOnly  bool
				computed  bool
				specifier string
				target    string
			}
			var edges []edge
			add := func(n *ast.Node, k string, typeOnly bool) {
				e := edge{kind: k, typeOnly: typeOnly, computed: n == nil || !ast.IsStringLiteralLike(n)}
				if !e.computed {
					e.specifier = n.Text()
					resolved := p.GetResolvedModuleFromModuleSpecifier(source, n)
					if resolved != nil && resolved.IsResolved() {
						if target := p.GetSourceFile(resolved.ResolvedFileName); target != nil {
							e.target = target.FileName().AsString()
						}
					}
				}
				edges = append(edges, e)
			}
			if !source.IsDeclarationFile {
				for _, s := range source.Statements.Nodes {
					switch s.Kind {
					case ast.KindImportDeclaration:
						d := s.AsImportDeclaration()
						add(d.ModuleSpecifier, "import", d.ImportClause != nil && d.ImportClause.IsTypeOnly())
					case ast.KindExportDeclaration:
						d := s.AsExportDeclaration()
						if d.ModuleSpecifier != nil {
							add(d.ModuleSpecifier, "export", d.IsTypeOnly)
						}
					case ast.KindImportEqualsDeclaration:
						d := s.AsImportEqualsDeclaration()
						if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
							add(d.ModuleReference.AsExternalModuleReference().Expression, "import-equals", d.IsTypeOnly)
						}
					}
				}
				var calls func(*ast.Node)
				calls = func(n *ast.Node) {
					if n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false)) {
						args := n.AsCallExpression().Arguments
						var first *ast.Node
						if args != nil && len(args.Nodes) > 0 {
							first = args.Nodes[0]
						}
						k := "require"
						if ast.IsImportCall(n) {
							k = "dynamic-import"
						}
						add(first, k, false)
					}
					n.ForEachChild(func(child *ast.Node) bool { calls(child); return false })
				}
				calls(source.AsNode())
			}
			out.number(uint64(len(edges)))
			for _, e := range edges {
				out.text(e.kind)
				out.yes(e.typeOnly)
				out.yes(e.computed)
				out.text(e.specifier)
				out.text(e.target)
			}
		}
	}
	return out.String(), nil
}
