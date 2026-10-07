package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// Raw program texts and module resolutions. Native code selects entry files,
// import kinds, seeds and transitive closure and decides blocking.
func (p *Program) wave27ModuleSources(out *fields, node *ast.Node, question string) (string, error) {
	if question != "wave-27-module-sources" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("module-sources requires a SourceFile")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		out.yes(ast.IsExternalModule(file))
		out.text(file.Text())
		type edge struct {
			kind     string
			typeOnly bool
			literal  bool
			target   string
		}
		var edges []edge
		add := func(kind string, typeOnly bool, specifier *ast.Node) {
			e := edge{kind: kind, typeOnly: typeOnly}
			if specifier != nil && ast.IsStringLiteralLike(specifier) {
				e.literal = true
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, specifier)
				if resolved != nil && resolved.IsResolved() {
					if target := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName); target != nil {
						e.target = target.FileName()
					}
				}
			}
			edges = append(edges, e)
		}
		if !file.IsDeclarationFile {
			for _, statement := range file.Statements.Nodes {
				switch statement.Kind {
				case ast.KindImportDeclaration:
					d := statement.AsImportDeclaration()
					add("import", d.ImportClause != nil && d.ImportClause.IsTypeOnly(), d.ModuleSpecifier)
				case ast.KindExportDeclaration:
					d := statement.AsExportDeclaration()
					add("export", d.IsTypeOnly, d.ModuleSpecifier)
				case ast.KindImportEqualsDeclaration:
					d := statement.AsImportEqualsDeclaration()
					if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
						add("import-equals", d.IsTypeOnly, d.ModuleReference.AsExternalModuleReference().Expression)
					}
				}
			}
			if strings.Contains(file.Text(), "import(") || strings.Contains(file.Text(), "require(") {
				var visit func(*ast.Node) bool
				visit = func(n *ast.Node) bool {
					if n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false)) {
						var specifier *ast.Node
						if args := n.AsCallExpression().Arguments; args != nil && len(args.Nodes) > 0 {
							specifier = args.Nodes[0]
						}
						add("call", false, specifier)
					}
					n.ForEachChild(visit)
					return false
				}
				file.AsNode().ForEachChild(visit)
			}
		}
		out.number(uint64(len(edges)))
		for _, edge := range edges {
			out.text(edge.kind)
			out.yes(edge.typeOnly)
			out.yes(edge.literal)
			out.text(edge.target)
		}
	}
	return out.String(), nil
}
