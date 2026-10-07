package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// declarationLineage exposes symbol declarations and their ancestors, not rule predicates.
func (p *Program) declarationLineage(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) < 2 {
		return "", fmt.Errorf("declaration-lineage requires a selector")
	}
	var symbol *ast.Symbol
	var declarations []*ast.Node
	switch split[1] {
	case "node":
		if len(split) != 2 {
			return "", fmt.Errorf("invalid selector")
		}
		symbol = c.GetSymbolAtLocation(node)
	case "signature":
		if len(split) != 2 || node.Kind != ast.KindCallExpression {
			return "", fmt.Errorf("signature requires a call")
		}
		if sig := c.GetResolvedSignature(node); sig != nil && sig.Declaration() != nil {
			declarations = []*ast.Node{sig.Declaration()}
		}
	case "type":
		if len(split) != 3 {
			return "", fmt.Errorf("type selector requires identity")
		}
		id, err := strconv.ParseUint(split[2], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[2] {
			return "", fmt.Errorf("unknown checker type identity")
		}
		symbol = p.typesByID[id-1].Symbol()
	default:
		return "", fmt.Errorf("unknown lineage selector")
	}
	flags := uint64(0)
	name := ""
	if symbol != nil {
		flags = uint64(symbol.Flags)
		name = symbol.Name
		declarations = symbol.Declarations
	}
	out.number(p.symbolID(symbol))
	out.number(flags)
	out.text(name)
	out.number(uint64(len(declarations)))
	for _, d := range declarations {
		f := ast.GetSourceFileOfNode(d)
		if f == nil {
			return "", fmt.Errorf("declaration has no source")
		}
		out.text(f.FileName())
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(f.Path()))
		out.yes(ast.IsTypeOnlyImportOrExportDeclaration(d))
		var chain []*ast.Node
		for cur := d; cur != nil; cur = cur.Parent {
			chain = append(chain, cur)
		}
		out.number(uint64(len(chain)))
		for _, cur := range chain {
			out.text(strings.TrimPrefix(cur.Kind.String(), "Kind"))
			name := ""
			if n := cur.Name(); n != nil {
				name = n.Text()
			}
			out.text(name)
			nameKind := ""
			if n := cur.Name(); n != nil {
				nameKind = strings.TrimPrefix(n.Kind.String(), "Kind")
			}
			out.text(nameKind)
			out.number(uint64(cur.Pos()))
			out.number(uint64(cur.End()))
			out.number(uint64(cur.Flags))
		}
	}
	return out.String(), nil
}
