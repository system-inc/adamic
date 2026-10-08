package checker

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Symbols are opaque, program-owned identities, never addresses. In shorthand
// assignments this selects the value symbol, rather than the property's symbol.
func (p *Program) writeSymbol(out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node) {
	var symbol *ast.Symbol
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
		symbol = c.GetShorthandAssignmentValueSymbol(parent)
	}
	if symbol == nil {
		symbol = c.GetSymbolAtLocation(node)
	}
	if symbol == nil {
		symbol = node.Symbol()
	}
	if symbol == nil {
		out.number(0)
		out.text("")
		out.number(0)
		return
	}
	if p.symbolIDs == nil {
		p.symbolIDs = make(map[*ast.Symbol]uint64)
	}
	id := p.symbolIDs[symbol]
	if id == 0 {
		id = uint64(len(p.symbolIDs)) + 1
		if id > (1<<53)-1 {
			panic("symbol identity space exhausted")
		}
		p.symbolIDs[symbol] = id
	}
	out.number(id)
	out.text(symbol.Name)
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.yes(ast.GetSourceFileOfNode(declaration) == source)
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		name, property, module := "", "", ""
		if n := declaration.Name(); n != nil {
			name = n.Text()
		}
		if n := declaration.PropertyName(); n != nil {
			property = n.Text()
		}
		switch declaration.Kind {
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
			for parent := declaration.Parent; parent != nil; parent = parent.Parent {
				if parent.Kind == ast.KindImportDeclaration {
					if n := parent.AsImportDeclaration().ModuleSpecifier; n != nil && ast.IsStringLiteralLike(n) {
						module = n.Text()
					}
					break
				}
			}
		}
		out.text(name)
		out.text(property)
		out.text(module)
	}
}
