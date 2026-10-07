package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// typeDeclarationAncestors returns raw syntax ancestry of a type's own symbol
// declarations. Counts and annotation links are syntax facts, never rule verdicts.
func (p *Program) typeDeclarationAncestors(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("type-declaration-ancestors requires an identity")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	var declarations []*ast.Node
	if symbol := p.typesByID[id-1].Symbol(); symbol != nil {
		declarations = symbol.Declarations
	}
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		var chain []*ast.Node
		for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
			chain = append(chain, ancestor)
		}
		out.number(uint64(len(chain)))
		source := ast.GetSourceFileOfNode(declaration)
		for _, ancestor := range chain {
			out.text(strings.TrimPrefix(ancestor.Kind.String(), "Kind"))
			out.text(source.FileName())
			out.number(uint64(ancestor.Pos()))
			out.number(uint64(ancestor.End()))
			name := ""
			if n := ancestor.Name(); n != nil && (n.Kind == ast.KindIdentifier || n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNumericLiteral) {
				name = n.Text()
			}
			out.text(name)
			children := uint64(0)
			ancestor.ForEachChild(func(child *ast.Node) bool { children++; return false })
			out.number(children)
			var annotation *ast.Node
			if ancestor.Kind == ast.KindTypeAliasDeclaration {
				annotation = ancestor.AsTypeAliasDeclaration().Type
				for annotation != nil && annotation.Kind == ast.KindParenthesizedType {
					annotation = annotation.AsParenthesizedTypeNode().Type
				}
			}
			out.yes(annotation != nil)
			if annotation != nil {
				out.text(strings.TrimPrefix(annotation.Kind.String(), "Kind"))
				out.number(uint64(annotation.Pos()))
				out.number(uint64(annotation.End()))
			}
		}
	}
	return out.String(), nil
}
