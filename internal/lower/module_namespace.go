package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// An ESM namespace qualification denotes the export's live binding, even through
// re-export barrels. A structural object with the same type is not that binding.
func (l *lowering) moduleNamespace(node *ast.Node) bool {
	symbol := l.symbol(ast.SkipParentheses(node))
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindSourceFile && !load.IsLibrary(declaration.AsSourceFile()) && !load.IsPrelude(declaration.AsSourceFile()) {
			return true
		}
	}
	return false
}

func (l *lowering) moduleNamespaceRefusal(node *ast.Node) error {
	if l.namespaceValueNode(node) && l.moduleNamespace(node) && l.namespaceEnumViewOf(node) == nil {
		parent := node.Parent
		if parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.Expression() != node {
			return l.notYet(node, "an ESM namespace object used as a value; use qualified live bindings")
		}
	}
	if node.Kind != ast.KindPropertyAccessExpression || !l.moduleNamespace(node.Expression()) {
		return nil
	}
	symbol := l.symbol(node)
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if ast.IsFunctionLike(declaration) && namespaceOwnThis(declaration) {
			return l.notYet(node, "an ESM namespace function observing its receiver; the module object is not represented")
		}
	}
	return nil
}

// namespaceEnumView is the enum declaration (ns as Readonly<Record<"K", V>>).K reads, the way tsc's
// debug.ts reads an enum object through its own ESM namespace, or nil. It's admitted only when all of
// these hold (Adamic's ruling on #52zja36, Nov 6):
//   - ns is an ESM namespace;
//   - the cast's type has one property, K, read only, and no index, so K is one string literal;
//   - the namespace's export K resolves, through export *, to an enum declaration whose object exists;
//   - V is a readonly string-index view whose value type holds every value the enum's table holds:
//     string | number for a numeric enum, whose reverse keys hold its names.
//
// Such a read is the enum's live binding, the same one ns.K reads. Every other use of the namespace
// object stays as it was.
func (l *lowering) namespaceEnumView(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindPropertyAccessExpression || node.AsPropertyAccessExpression().QuestionDotToken != nil || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil
	}
	cast := ast.SkipParentheses(node.Expression())
	if cast.Kind != ast.KindAsExpression {
		return nil
	}
	namespace := ast.SkipParentheses(cast.AsAsExpression().Expression)
	if !ast.IsIdentifier(namespace) || !l.moduleNamespace(namespace) {
		return nil
	}
	name := node.Name().Text()
	viewed := l.checker.GetTypeAtLocation(cast)
	properties := l.checker.GetPropertiesOfType(viewed)
	if len(properties) != 1 || properties[0].Name != name || !l.checker.IsReadonlySymbol(properties[0]) || len(l.checker.GetIndexInfosOfType(viewed)) != 0 {
		return nil
	}
	target := l.namespaceExport(namespace, name)
	if target == nil || target.Flags&ast.SymbolFlagsEnum == 0 || target.ValueDeclaration == nil || target.ValueDeclaration.Kind != ast.KindEnumDeclaration || len(target.Declarations) != 1 {
		return nil
	}
	declaration := target.ValueDeclaration
	if !l.enumObjectBuilt(declaration) {
		return nil
	}
	view := l.checker.GetTypeOfSymbol(properties[0])
	infos := l.checker.GetIndexInfosOfType(view)
	if len(infos) != 1 || infos[0].KeyType().Flags() != checker.TypeFlagsString || !infos[0].IsReadonly() || len(l.checker.GetPropertiesOfType(view)) != 0 {
		return nil
	}
	fields, err := l.enumFields(declaration)
	if err != nil {
		return nil
	}
	for _, field := range fields {
		held := l.checker.GetStringType()
		if field.Value.Type() == ir.Number {
			held = l.checker.GetNumberType()
		}
		if !l.checker.IsTypeAssignableTo(held, infos[0].ValueType()) {
			return nil
		}
	}
	return declaration
}

// namespaceEnumViewOf is the enum namespaceEnumView admits where node is the namespace it casts, or nil.
func (l *lowering) namespaceEnumViewOf(node *ast.Node) *ast.Node {
	access := l.namespaceCastAccess(node)
	if access == nil {
		return nil
	}
	return l.namespaceEnumView(access)
}

// namespaceCastAccess is (node as T).K, the property read of a cast whose operand is node, or nil. The
// operand may be parenthesized, ((ts) as T).K, and node either the name or its parentheses.
func (l *lowering) namespaceCastAccess(node *ast.Node) *ast.Node {
	cast := node.Parent
	for cast != nil && cast.Kind == ast.KindParenthesizedExpression {
		cast = cast.Parent
	}
	if cast == nil || cast.Kind != ast.KindAsExpression || ast.SkipParentheses(cast.AsAsExpression().Expression) != ast.SkipParentheses(node) {
		return nil
	}
	access := cast.Parent
	for access != nil && access.Kind == ast.KindParenthesizedExpression {
		access = access.Parent
	}
	if access == nil || access.Kind != ast.KindPropertyAccessExpression || ast.SkipParentheses(access.Expression()) != cast {
		return nil
	}
	return access
}

// namespaceExport is the value a namespace's export name resolves to, through export * and aliases,
// as recordExports resolves them, or nil. A type-only export has no value.
func (l *lowering) namespaceExport(namespace *ast.Node, name string) *ast.Symbol {
	for _, exported := range l.namespaceExportSymbols(namespace) {
		if exported.Name == name {
			return l.resolvedExport(exported)
		}
	}
	return nil
}

func (l *lowering) namespaceExportSymbols(namespace *ast.Node) []*ast.Symbol {
	module := l.symbol(namespace)
	if module == nil {
		return nil
	}
	return l.checker.GetExportsOfModule(module)
}

func (l *lowering) resolvedExport(exported *ast.Symbol) *ast.Symbol {
	if l.checker.GetTypeOnlyAliasDeclaration(exported) != nil {
		return nil
	}
	if exported.Flags&ast.SymbolFlagsAlias != 0 {
		exported = l.checker.GetAliasedSymbol(exported)
	}
	return l.checker.GetExportSymbolOfSymbol(exported)
}
