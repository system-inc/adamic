package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Declaration syntax, including external declarations, without a lint judgment.
func (p *Program) inspectJsxDeclarationSyntax(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if strings.Split(question, "\n")[0] != "jsx-declaration-syntax" {
		return p.inspectCallbackSymbolFacts(out, c, node, question)
	}
	if question != "jsx-declaration-syntax" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("jsx-declaration-syntax requires an Identifier and no suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if symbol == nil {
		out.number(0)
		return out.String(), nil
	}
	out.number(uint64(len(symbol.Declarations)))
	kind := func(n *ast.Node) string {
		if n == nil {
			return ""
		}
		return strings.TrimPrefix(n.Kind.String(), "Kind")
	}
	for _, decl := range symbol.Declarations {
		parent := decl.Parent
		var grand *ast.Node
		if parent != nil {
			grand = parent.Parent
		}
		out.text(kind(decl))
		out.text(kind(parent))
		out.text(kind(grand))
		imported, module := "", ""
		var initializer *ast.Node
		switch decl.Kind {
		case ast.KindImportSpecifier:
			spec := decl.AsImportSpecifier()
			name := spec.PropertyName
			if name == nil {
				name = spec.Name()
			}
			if name != nil && name.Kind == ast.KindIdentifier {
				imported = name.Text()
			}
			if parent != nil && parent.Kind == ast.KindNamedImports && grand != nil && grand.Kind == ast.KindImportClause && grand.Parent != nil && grand.Parent.Kind == ast.KindImportDeclaration {
				n := grand.Parent.AsImportDeclaration().ModuleSpecifier
				if n != nil && ast.IsStringLiteralLike(n) {
					module = n.Text()
				}
			}
		case ast.KindVariableDeclaration:
			initializer = decl.AsVariableDeclaration().Initializer
		case ast.KindBindingElement:
			if grand != nil && grand.Kind == ast.KindVariableDeclaration {
				initializer = grand.AsVariableDeclaration().Initializer
			}
		}
		out.text(imported)
		out.text(module)
		raw := ""
		if initializer != nil {
			source := ast.GetSourceFileOfNode(initializer)
			if source == nil {
				return "", fmt.Errorf("declaration initializer has no source")
			}
			raw = source.Text()[initializer.Pos():initializer.End()]
		}
		out.text(raw)
	}
	return out.String(), nil
}
