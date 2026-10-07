package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"sort"
	"strings"
)

// Binder tables are metadata. The native rule chooses which bindings collide.
func scopeTables(out *fields, source *ast.Node) {
	type table struct {
		node    *ast.Node
		kind    string
		symbols ast.SymbolTable
	}
	var tables []table
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if ast.IsLocalsContainer(node) {
			tables = append(tables, table{node, "locals", node.Locals()})
		}
		if node.Kind == ast.KindEnumDeclaration && node.Symbol() != nil {
			tables = append(tables, table{node, "exports", node.Symbol().Exports})
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source)
	out.number(uint64(len(tables)))
	for _, table := range tables {
		out.text(table.kind)
		out.text(strings.TrimPrefix(table.node.Kind.String(), "Kind"))
		out.number(uint64(table.node.Pos()))
		out.number(uint64(table.node.End()))
		var names []string
		for name, symbol := range table.symbols {
			if symbol != nil {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		out.number(uint64(len(names)))
		for _, name := range names {
			symbol := table.symbols[name]
			out.text(name)
			out.number(uint64(symbol.Flags))
			out.number(uint64(len(symbol.Declarations)))
			for _, declaration := range symbol.Declarations {
				out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
				out.number(uint64(declaration.Pos()))
				out.number(uint64(declaration.End()))
			}
		}
	}
}
