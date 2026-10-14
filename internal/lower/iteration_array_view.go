package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An inherited ReadonlyArray view uses the array's existing storage and identity.
// Object constructions with this structural type still need protocol dispatch.
func (l *lowering) readonlyArrayView(proven *checker.Type) *checker.Type {
	proven = l.concrete(proven)
	base := l.readonlyArrayBase(proven, map[*checker.Type]bool{})
	if base == nil {
		return nil
	}
	index := l.checker.GetIndexInfoOfType(proven, l.checker.GetNumberType())
	if index == nil || !index.IsReadonly() {
		return nil
	}
	for _, name := range []string{"copyWithin", "fill", "pop", "push", "reverse", "shift", "sort", "splice", "unshift"} {
		if l.checker.GetPropertyOfType(proven, name) != nil {
			return nil
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(base) {
		if !l.librarySymbol(l.checker.GetPropertyOfType(proven, property.Name)) {
			return nil
		}
	}
	return base
}

func (l *lowering) readonlyArrayBase(proven *checker.Type, visited map[*checker.Type]bool) *checker.Type {
	if proven == nil || visited[proven] || proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	if l.checker.IsArrayType(proven) {
		return nil
	}
	visited[proven] = true
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	symbol := target.Symbol()
	if symbol == nil {
		return nil
	}
	var declaration *ast.Node
	for _, node := range symbol.Declarations {
		if node.Kind == ast.KindInterfaceDeclaration {
			declaration = node
			break
		}
	}
	if declaration == nil {
		return nil
	}
	mapper := l.typeMapper
	if len(declaration.TypeParameters()) > 0 && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		mapper = l.typeMapperOf(declaration, proven)
	}
	for _, base := range l.checker.GetBaseTypes(target) {
		if mapper != nil {
			base = instantiateType(l.checker, base, mapper)
		}
		if l.isLibraryType(base, "ReadonlyArray") {
			return base
		}
		if found := l.readonlyArrayBase(base, visited); found != nil {
			return found
		}
	}
	return nil
}

func (l *lowering) readonlyArrayValue(node *ast.Node, value ir.Expression) error {
	own := l.checker.GetTypeAtLocation(node)
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if l.readonlyArrayView(l.checker.GetNonNullableType(own)) == nil && (contextual == nil || l.readonlyArrayView(l.checker.GetNonNullableType(contextual)) == nil) {
		return nil
	}
	if _, missing := value.(ir.Undefined); missing {
		return nil
	}
	if _, missing := value.(ir.Null); missing {
		return nil
	}
	if value.Type() != ir.Array {
		return l.notYet(node, "a structural ReadonlyArray object needing iterator protocol dispatch")
	}
	if contextual != nil {
		to, known := l.representation(contextual)
		if !known || (to != ir.Array && to != ir.Union) {
			return l.notYet(node, "an inherited ReadonlyArray viewed through another native storage representation")
		}
	}
	return nil
}
