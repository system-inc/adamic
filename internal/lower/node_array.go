package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Recognize declaration ancestry, never a coincidental type name or structural
// array-like object. Element and extra storage are then resolved by the checker.
func (lowering *lowering) nodeArrayBase(proven *checker.Type, seen map[*checker.Type]bool) *checker.Type {
	if proven == nil || seen[proven] || proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	seen[proven] = true
	if lowering.isLibraryType(proven, "Array", "ReadonlyArray") {
		return proven
	}
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	if target.ObjectFlags()&checker.ObjectFlagsInterface == 0 {
		return nil
	}
	for _, base := range lowering.checker.GetBaseTypes(target) {
		if found := lowering.nodeArrayBase(base, seen); found != nil {
			return found
		}
	}
	return nil
}

func (lowering *lowering) nodeArrayLayoutOf(proven *checker.Type) (ArrayLayout, *checker.Type, bool) {
	if proven == nil || lowering.isArrayType(proven) || checker.IsTupleType(proven) {
		return ArrayLayout{}, nil, false
	}
	base := lowering.nodeArrayBase(proven, map[*checker.Type]bool{})
	if base == nil {
		return ArrayLayout{}, nil, false
	}
	element := lowering.checker.GetIndexTypeOfType(proven, lowering.checker.GetNumberType())
	if element == nil {
		return ArrayLayout{}, nil, false
	}
	if lowering.nodeArrayBase(element, map[*checker.Type]bool{}) != nil && !lowering.isArrayType(element) {
		return ArrayLayout{}, nil, false
	}
	held, known := lowering.representation(element)
	if !known || slotless(held) {
		return ArrayLayout{}, nil, false
	}
	extras := []FieldContract{}
	for _, property := range lowering.checker.GetPropertiesOfType(proven) {
		if lowering.checker.GetPropertyOfType(base, property.Name) != nil {
			continue
		}
		declared := lowering.checker.GetTypeOfSymbol(property)
		of, known := lowering.representation(declared)
		if !known {
			return ArrayLayout{}, nil, false
		}
		extras = append(extras, FieldContract{
			Name:         property.Name,
			DeclaredType: lowering.checker.TypeToString(declared),
			Of:           of,
			Optional:     property.Flags&ast.SymbolFlagsOptional != 0,
		})
	}
	layout, nodeArrayLayoutError := NodeArrayLayout(held, extras)
	return layout, element, nodeArrayLayoutError == nil
}

func (lowering *lowering) nodeArrayCast(node *ast.Node) bool {
	as := node.AsAsExpression()
	if ast.SkipParentheses(as.Expression).Kind != ast.KindArrayLiteralExpression {
		return false
	}
	layout, element, known := lowering.nodeArrayLayoutOf(lowering.concrete(lowering.checker.GetTypeAtLocation(node)))
	if !known || len(layout.Fields) == 0 {
		return false
	}
	actual := lowering.elementTypeOfArray(lowering.checker.GetTypeAtLocation(as.Expression))
	// Mutable array elements keep their representation and their full element contract.
	return actual != nil &&
		lowering.castAssignable(actual, element) &&
		lowering.castAssignable(element, lowering.checker.GetWidenedType(actual)) &&
		lowering.sameKeeping(actual, element, map[[2]*checker.Type]bool{})
}

func (lowering *lowering) constructNodeArray(node *ast.Node) (ir.Expression, error) {
	layout, element, _ := lowering.nodeArrayLayoutOf(lowering.concrete(lowering.checker.GetTypeAtLocation(node)))
	value, arrayLiteralError := lowering.arrayLiteral(ast.SkipParentheses(node.AsAsExpression().Expression))
	if arrayLiteralError != nil {
		return nil, arrayLiteralError
	}
	literal, known := value.(ir.ArrayLiteral)
	if !known {
		return nil, lowering.notYet(node, "NodeArray construction from a nonliteral array operation")
	}
	held, _ := lowering.representation(element)
	if literal.Element != held {
		return nil, lowering.notYet(node, "NodeArray element storage conversion")
	}
	for _, field := range layout.Fields {
		zero := ir.Expression(zeroValue(field.Of))
		if field.Of.IsReference() {
			zero = ir.Undefined{Of: field.Of}
		}
		literal.Metadata = append(literal.Metadata, ir.Field{Name: field.Name, Value: zero, Absent: true})
	}
	return literal, nil
}

func (lowering *lowering) nodeArrayProperty(
	node *ast.Node,
	object ir.Expression,
	name string,
) (ir.Expression, bool, error) {
	layout, _, known := lowering.nodeArrayLayoutOf(
		lowering.checker.GetNonNullableType(
			lowering.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression),
		),
	)
	if !known {
		return nil, false, nil
	}
	field, readError := layout.Read(name)
	if readError != nil {
		return nil, true, lowering.notYet(node, "a NodeArray field outside its fixed metadata layout")
	}
	declared := lowering.checker.GetTypeOfSymbol(lowering.checker.GetSymbolAtLocation(node.Name()))
	value := ir.DynamicProperty{Object: fit(object, ir.Union), Name: name}
	checked, unsetSpecializationError := lowering.unsetSpecialization(
		node,
		value,
		declared,
		"field '"+name+"', factory 'NodeArray allocation'",
	)
	_ = field // Layout membership is proved independently of observed presence.
	return checked, true, unsetSpecializationError
}
