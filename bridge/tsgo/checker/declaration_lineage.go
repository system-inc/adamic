package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Stable AST identities and complete parent chains, without classifying outcomes.
func lineageKey(node *ast.Node) string {
	if node == nil {
		return ""
	}
	source := ast.GetSourceFileOfNode(node)
	if source == nil {
		panic("lineage node has no source")
	}
	return fmt.Sprintf("%s:%s:%d:%d", source.FileName(), node.Kind.String(), node.Pos(), node.End())
}
func (p *Program) declarationLineage(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	var symbol *ast.Symbol
	if len(parts) == 1 {
		symbol = c.GetSymbolAtLocation(node)
	} else if len(parts) == 2 {
		id, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
			return "", fmt.Errorf("unknown checker type identity")
		}
		symbol = p.typesByID[id-1].Symbol()
	} else {
		return "", fmt.Errorf("declaration-lineage takes an optional type identity")
	}
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		var chain []*ast.Node
		for parent := declaration; parent != nil; parent = parent.Parent {
			chain = append(chain, parent)
		}
		out.number(uint64(len(chain)))
		for _, parent := range chain {
			source := ast.GetSourceFileOfNode(parent)
			out.text(lineageKey(parent))
			out.text(strings.TrimPrefix(parent.Kind.String(), "Kind"))
			name := ""
			if n := parent.Name(); n != nil && n.Kind == ast.KindIdentifier {
				name = n.Text()
			}
			out.text(name)
			out.text(source.FileName().AsString())
			out.yes(source.IsDeclarationFile)
			out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.PathKey()))
			var children []*ast.Node
			parent.ForEachChild(func(child *ast.Node) bool { children = append(children, child); return false })
			out.number(uint64(len(children)))
			for _, child := range children {
				out.text(lineageKey(child))
			}
			declaredType := ""
			if parent.Kind == ast.KindTypeAliasDeclaration {
				declaredType = lineageKey(parent.AsTypeAliasDeclaration().Type)
			}
			out.text(declaredType)
		}
	}
	return out.String(), nil
}
