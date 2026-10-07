package metamorphic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// aliasSites are the constants read twice or more after their statement: each site declares an alias
// right after the statement, const alias = name, and reads it from the second read on. The alias
// holds the very value the constant does, so every read sees what it saw; what changes is that a
// value has two names, which is what ownership has to get right.
func aliasSites(p *program) []site {
	found := p.references()
	var sites []site
	visit(p.file.AsNode(), func(node *ast.Node) bool {
		if node.Kind != ast.KindVariableStatement || !inStatementList(node) {
			return true
		}
		list := node.AsVariableStatement().DeclarationList
		if list.Flags&ast.NodeFlagsConst == 0 {
			return true
		}
		for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
			name := declaration.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				continue
			}
			symbol := found.symbolOf[name]
			if symbol == nil {
				continue
			}
			var after []reference
			for _, entry := range found.reads(symbol) {
				if entry.node.Pos() >= node.End() && !ast.IsPartOfTypeNode(entry.node) {
					after = append(after, entry)
				}
			}
			if len(after) < 2 {
				continue
			}
			alias := p.fresh(name.Text() + "Alias")
			change := site{edits: []edit{{node.End(), node.End(), "\n" + p.indentation(node) + "const " + alias + " = " + name.Text() + ";"}}}
			for _, entry := range after[1:] {
				start := p.start(entry.node)
				if entry.shorthand {
					change.edits = append(change.edits, edit{start, entry.node.End(), name.Text() + ": " + alias})
				} else {
					change.edits = append(change.edits, edit{start, entry.node.End(), alias})
				}
			}
			sites = append(sites, change)
		}
		return true
	})
	return sites
}

// inStatementList says whether a statement is one of a list of statements, where another can go
// beside it: the file's, a block's. A case clause's statements are left alone.
func inStatementList(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && (parent.Kind == ast.KindSourceFile || parent.Kind == ast.KindBlock)
}
