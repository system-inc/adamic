package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Prove only direct reads of enum bindings whose declaration has already run.
// Calls and instance initializers are deferred code: their actual binding reads
// keep runtime readiness checks, regardless of unrelated pending declarations.
func (l *lowering) enumInitialization(modules []*ast.SourceFile) error {
	if l.provenModuleReads == nil {
		l.provenModuleReads = map[*ast.Node]bool{}
	}
	initialized := map[*ast.Node]bool{}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil || ast.IsFunctionLike(node) || ast.IsTypeNode(node) {
			return false
		}
		if node.Kind == ast.KindPropertyDeclaration && !ast.HasStaticModifier(node) {
			return false
		}
		if ast.IsIdentifier(node) {
			symbol := l.symbol(node)
			if symbol != nil && symbol.ValueDeclaration != nil && initialized[symbol.ValueDeclaration] {
				l.provenModuleReads[node] = true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		for _, statement := range module.Statements.Nodes {
			if statement.Kind == ast.KindEnumDeclaration {
				initialized[statement] = true
				continue
			}
			visit(statement)
		}
	}
	return nil
}
