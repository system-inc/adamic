package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A readonly unknown slot preserves the closure's heap identity and actual
// producer metadata. This is storage permission, never a signature certificate.
func (l *lowering) readonlyViewCallableStorage(node *ast.Node, producer *checker.Type) bool {
	if node.Parent == nil || node.Parent.Kind != ast.KindPropertyAssignment || !l.runtimeViewCallableShape(producer) {
		return false
	}
	property := node.Parent
	literal := property.Parent
	if literal == nil || literal.Kind != ast.KindObjectLiteralExpression || property.Name() == nil || !ast.IsIdentifier(property.Name()) {
		return false
	}
	contextual := l.checker.GetContextualType(literal, checker.ContextFlagsNone)
	if contextual == nil {
		return false
	}
	field := l.checker.GetPropertyOfType(contextual, property.Name().Text())
	return field != nil && l.checker.IsReadonlySymbol(field) && dynamicObjectType(l.checker.GetTypeOfSymbol(field))
}
