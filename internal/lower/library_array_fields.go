package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Numeric array-like keys need the checker's canonical spelling at construction.
// Keep this out of method-name proofs, whose consumers include Object observations.
func (l *lowering) libraryArrayNumericField(node, property *ast.Node, literal *ir.ObjectLiteral) (bool, error) {
	if property.Kind != ast.KindPropertyAssignment || property.Name().Kind != ast.KindNumericLiteral {
		return false, nil
	}
	fieldName, known := l.libraryArrayLikeFieldName(property.Name())
	if !known {
		return true, l.notYet(property.Name(), "a numeric field without a canonical name")
	}
	if literal.Spread != nil && !l.hasProperty(node.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression, fieldName) {
		return true, l.notYet(property, "a spread that adds a field the source doesn't have")
	}
	initializer := property.AsPropertyAssignment().Initializer
	if l.uninitializedInitializer(initializer) {
		declared := l.declaredField(node, fieldName)
		if declared == 0 || declared == ir.MaybeBoolean {
			return true, l.notYet(property, "an uninitialized object field without a supported declared slot type")
		}
		value := ir.Expression(zeroValue(declared))
		if declared.IsReference() {
			value = ir.Undefined{Of: declared}
		}
		literal.Fields = append(literal.Fields, ir.Field{Name: fieldName, Value: value, Uninitialized: true})
		return true, nil
	}
	value, err := l.expression(initializer)
	if err != nil {
		return true, err
	}
	if declared := l.declaredField(node, fieldName); declared != 0 && !censusFieldSlotless(declared) {
		value = fit(value, declared)
	}
	if censusFieldSlotless(value.Type()) {
		return true, l.notYet(property, "a field holding "+typeName(value.Type()))
	}
	literal.Fields = append(literal.Fields, ir.Field{Name: fieldName, Value: value})
	return true, nil
}
