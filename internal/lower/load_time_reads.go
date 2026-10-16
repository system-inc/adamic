package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// proveModuleReads marks the global reads in a cyclic program that the entry's actual ESM order proves run after their
// provider finished: a module function declaration, which is hoisted, or a value whose module runs before the reader's.
// Every other global read keeps its runtime dead-zone check, so the program is accepted whatever a cycle reads at load.
// Cohere's nexus/correctness-no-import-cycle-load-time-read inventories those reads lint-side; lowering never runs it,
// so the compiler links none of cohere's lint (#321063n). The ledger in cycleledger holds the rule's findings beside these
// decisions.
func (l *lowering) proveModuleReads(files []*ast.SourceFile) {
	l.provenModuleReads = map[*ast.Node]bool{}
	positions := map[*ast.SourceFile]int{}
	for index, file := range files {
		positions[file] = index
	}
	for _, file := range files {
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node == nil || ast.IsFunctionLike(node) {
				return false
			}
			switch node.Kind {
			case ast.KindImportDeclaration, ast.KindExportDeclaration:
				return false
			case ast.KindHeritageClause:
				if node.AsHeritageClause().Token == ast.KindExtendsKeyword && ast.IsClassLike(node.Parent) {
					for _, element := range node.AsHeritageClause().Types.Nodes {
						visit(element.AsExpressionWithTypeArguments().Expression)
					}
				}
				return false
			case ast.KindExpressionWithTypeArguments:
				visit(node.AsExpressionWithTypeArguments().Expression)
				return false
			case ast.KindPropertyDeclaration:
				if !ast.HasStaticModifier(node) {
					return false
				}
			}
			if ast.IsTypeNode(node) {
				return false
			}
			if ast.IsIdentifier(node) || ast.IsPropertyAccessExpression(node) {
				symbol := l.checker.GetSymbolAtLocation(node)
				if ast.IsIdentifier(node) && node.Parent != nil && ast.IsShorthandPropertyAssignment(node.Parent) {
					symbol = l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
				}
				if symbol != nil {
					symbol = l.checker.SkipAlias(symbol)
					if symbol != nil && symbol.ValueDeclaration != nil {
						declaration := symbol.ValueDeclaration
						provider := ast.GetSourceFileOfNode(declaration)
						index, present := positions[provider]
						if (declaration.Kind == ast.KindFunctionDeclaration && declaration.Parent != nil && declaration.Parent.Kind == ast.KindSourceFile) || (present && provider != file && index < positions[file]) {
							l.provenModuleReads[node] = true
						}
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
}

// ProvenModuleReads is proveModuleReads for the ledger in cycleledger, which links cohere's rule so lowering never has to.
func ProvenModuleReads(typeChecker *checker.Checker, files []*ast.SourceFile) map[*ast.Node]bool {
	analysis := &lowering{checker: typeChecker}
	analysis.proveModuleReads(files)
	return analysis.provenModuleReads
}

func (l *lowering) checkedModuleRead(node *ast.Node, local int) bool {
	return !l.provenModuleReads[node] && l.checked(local)
}
