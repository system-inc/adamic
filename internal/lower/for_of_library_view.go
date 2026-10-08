package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A private consumer can retain the library iterator convention when every
// reference is a direct call with a fresh Map/Set iterator. Other structural
// views still need a protocol representation, including optional return.
func (l *lowering) forOfLibraryIteratorView(source, name *ast.Node) (ir.Type, bool) {
	proven := l.checker.GetTypeAtLocation(source)
	if !l.isLibraryType(proven, "Iterable", "IterableIterator") || !ast.IsIdentifier(source) {
		return 0, false
	}
	symbol := l.symbol(source)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindParameter {
		return 0, false
	}
	parameter := symbol.Declarations[0]
	declared := parameter.AsParameterDeclaration()
	owner := parameter.Parent
	if owner == nil || owner.Kind != ast.KindFunctionDeclaration || owner.Name() == nil || len(owner.TypeParameters()) != 0 || ast.HasSyntacticModifier(owner, ast.ModifierFlagsExport) || declared.Initializer != nil || declared.QuestionToken != nil || declared.DotDotDotToken != nil {
		return 0, false
	}
	function := l.symbol(owner.Name())
	if function == nil || len(function.Declarations) != 1 {
		return 0, false
	}
	position := -1
	for index, argument := range owner.Parameters() {
		if argument == parameter {
			position = index
		}
	}
	element, err := l.typeOf(name)
	if position < 0 || err != nil || (element != ir.String && element != ir.Number && element != ir.Boolean && element != ir.Object) {
		return 0, false
	}
	file := ast.GetSourceFileOfNode(owner)
	// A module's unexported function cannot have callers in another file.
	// Inspect its source even during selected replay. Global scripts need a
	// wider caller proof; ordinary Adamic loading forces module detection.
	if !ast.IsExternalModule(file) {
		return 0, false
	}
	called, sound := false, true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !sound {
			return true
		}
		if ast.IsIdentifier(node) {
			referenced := l.symbol(node)
			if referenced == symbol && node != parameter.Name() {
				// The parameter cannot escape, be reassigned, or expose next/return.
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindForOfStatement || parent.AsForInOrOfStatement().Expression != node {
					sound = false
					return true
				}
			}
			if referenced == function && node != owner.Name() {
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindCallExpression || parent.AsCallExpression().Expression != node {
					sound = false
					return true
				}
				call := parent.AsCallExpression()
				if position >= len(call.Arguments.Nodes) {
					sound = false
					return true
				}
				for _, argument := range call.Arguments.Nodes {
					if argument.Kind == ast.KindSpreadElement {
						sound = false
						return true
					}
				}
				argument := ast.SkipParentheses(call.Arguments.Nodes[position])
				if !l.forOfLibraryIteratorFactory(argument, element) {
					sound = false
					return true
				}
				called = true
			}
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	return element, sound && called
}

func (l *lowering) forOfLibraryIteratorFactory(node *ast.Node, element ir.Type) bool {
	if node.Kind != ast.KindCallExpression || len(node.AsCallExpression().Arguments.Nodes) != 0 {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || !l.libraryMember(callee) {
		return false
	}
	part := callee.Name().Text()
	if part != "keys" && part != "values" && part != "entries" {
		return false
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if !l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		return false
	}
	result := l.checker.GetTypeAtLocation(node)
	if !l.isLibraryType(result, "MapIterator", "SetIterator") {
		return false
	}
	arguments := l.typeArguments(result)
	if len(arguments) != 1 {
		return false
	}
	// A Weak view reads an object, but its collection stores handles. Keep the
	// physical slot proof rather than confusing the read type with its storage.
	stored, known := l.kept(arguments[0])
	return known && stored == element
}
