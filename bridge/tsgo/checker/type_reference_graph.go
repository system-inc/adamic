package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// typeReferenceGraph exposes type syntax and direct symbol declarations, including
// declarations in other files. It returns no reachability or circularity verdict.
// Constructor nodes remain opaque records; the native caller chooses which kinds
// are transparent to its relation. IDs are local to this answer, not handles.
func (p *Program) typeReferenceGraph(c *checker.Checker, root *ast.Node, question string) (string, error) {
	if question != "type-reference-graph" {
		return "", fmt.Errorf("unexpected type-reference-graph suffix")
	}
	type record struct {
		kind                   string
		symbol                 uint64
		body                   uint64
		children, declarations []uint64
	}
	nodes := map[*ast.Node]uint64{}
	var records []record
	var add func(*ast.Node) uint64
	add = func(node *ast.Node) uint64 {
		if node == nil {
			return 0
		}
		if id := nodes[node]; id != 0 {
			return id
		}
		id := uint64(len(records) + 1)
		nodes[node] = id
		records = append(records, record{})
		r := record{kind: strings.TrimPrefix(node.Kind.String(), "Kind")}
		switch node.Kind {
		case ast.KindIdentifier:
			symbol := c.GetSymbolAtLocation(node)
			r.symbol = p.symbolID(symbol)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					r.declarations = append(r.declarations, add(declaration))
				}
			}
		case ast.KindTypeAliasDeclaration:
			r.body = add(node.AsTypeAliasDeclaration().Type)
		case ast.KindIndexSignature:
			r.body = add(node.AsIndexSignatureDeclaration().Type)
		case ast.KindMappedType:
			r.body = add(node.AsMappedTypeNode().Type)
		case ast.KindParenthesizedType:
			r.body = add(node.AsParenthesizedTypeNode().Type)
		case ast.KindInterfaceDeclaration:
			if list := node.AsInterfaceDeclaration().Members; list != nil {
				for _, member := range list.Nodes {
					r.children = append(r.children, add(member))
				}
			}
		case ast.KindTypeLiteral, ast.KindUnionType, ast.KindIntersectionType,
			ast.KindConditionalType, ast.KindIndexedAccessType, ast.KindTypeReference:
			node.ForEachChild(func(child *ast.Node) bool { r.children = append(r.children, add(child)); return false })
		}
		records[id-1] = r
		return id
	}
	rootID := add(root)
	out := &fields{}
	out.number(1)
	out.text("type-reference-graph")
	out.number(rootID)
	out.number(uint64(len(records)))
	for _, r := range records {
		out.text(r.kind)
		out.number(r.symbol)
		out.number(r.body)
		out.ids(r.children)
		out.ids(r.declarations)
	}
	return out.String(), nil
}
