package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// typeDeclarationAncestry returns raw symbol declarations and their AST parent chains.
// It knows no lint rule names, accepted paths, union completeness or diagnostics.
func (p *Program) typeDeclarationAncestry(out *fields, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("type-declaration-ancestry requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	symbol := checker.Type_symbol(p.typesByID[id-1])
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return "", fmt.Errorf("declaration has no source")
		}
		out.text(file.FileName())
		var chain []*ast.Node
		for node := declaration; node != nil; node = node.Parent {
			chain = append(chain, node)
		}
		out.number(uint64(len(chain)))
		for _, node := range chain {
			out.text(strings.TrimPrefix(node.Kind.String(), "Kind"))
			out.number(uint64(node.Pos()))
			out.number(uint64(node.End()))
			name := ""
			if node.Kind == ast.KindTypeAliasDeclaration || node.Kind == ast.KindInterfaceDeclaration {
				if node.Name() != nil {
					name = node.Name().Text()
				}
			}
			out.text(name)
			var members []*ast.Node
			if node.Kind == ast.KindUnionType {
				members = node.AsUnionTypeNode().Types.Nodes
			}
			out.number(uint64(len(members)))
			var declared *ast.Node
			if node.Kind == ast.KindTypeAliasDeclaration {
				declared = node.AsTypeAliasDeclaration().Type
			}
			out.yes(declared != nil)
			if declared != nil {
				out.number(uint64(declared.Pos()))
				out.number(uint64(declared.End()))
				out.text(strings.TrimPrefix(declared.Kind.String(), "Kind"))
			}
		}
	}
	return out.String(), nil
}
