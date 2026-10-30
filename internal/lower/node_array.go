package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Recognize declaration ancestry, never a coincidental type name or structural
// array-like object. Element and extra storage are then resolved by the checker.
func (l *lowering) nodeArrayBase(proven *checker.Type, seen map[*checker.Type]bool) *checker.Type {
	if proven == nil || seen[proven] || proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	seen[proven] = true
	if l.isLibraryType(proven, "Array", "ReadonlyArray") {
		return proven
	}
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	if target.ObjectFlags()&checker.ObjectFlagsInterface == 0 {
		return nil
	}
	for _, base := range l.checker.GetBaseTypes(target) {
		if found := l.nodeArrayBase(base, seen); found != nil {
			return found
		}
	}
	return nil
}

func (l *lowering) nodeArrayLayoutOf(proven *checker.Type) (ArrayLayout, *checker.Type, bool) {
	if proven == nil || l.checker.IsArrayType(proven) || checker.IsTupleType(proven) {
		return ArrayLayout{}, nil, false
	}
	base := l.nodeArrayBase(proven, map[*checker.Type]bool{})
	if base == nil {
		return ArrayLayout{}, nil, false
	}
	element := l.checker.GetIndexTypeOfType(proven, l.checker.GetNumberType())
	if element == nil {
		return ArrayLayout{}, nil, false
	}
	if l.nodeArrayBase(element, map[*checker.Type]bool{}) != nil && !l.checker.IsArrayType(element) {
		return ArrayLayout{}, nil, false
	}
	held, known := l.representation(element)
	if !known || slotless(held) {
		return ArrayLayout{}, nil, false
	}
	extras := []FieldContract{}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if l.checker.GetPropertyOfType(base, property.Name) != nil {
			continue
		}
		declared := l.checker.GetTypeOfSymbol(property)
		of, known := l.representation(declared)
		if !known {
			return ArrayLayout{}, nil, false
		}
		extras = append(extras, FieldContract{Name: property.Name, DeclaredType: l.checker.TypeToString(declared), Of: of, Optional: property.Flags&ast.SymbolFlagsOptional != 0})
	}
	layout, err := NodeArrayLayout(held, extras)
	return layout, element, err == nil
}

func (l *lowering) nodeArrayCast(node *ast.Node) bool {
	as := node.AsAsExpression()
	if ast.SkipParentheses(as.Expression).Kind != ast.KindArrayLiteralExpression {
		return false
	}
	layout, element, known := l.nodeArrayLayoutOf(l.concrete(l.checker.GetTypeAtLocation(node)))
	if !known || len(layout.Fields) == 0 {
		return false
	}
	actual := l.checker.GetElementTypeOfArrayType(l.checker.GetTypeAtLocation(as.Expression))
	// Mutable array elements keep their representation and their full element contract.
	return actual != nil && l.castAssignable(actual, element) && l.castAssignable(element, l.checker.GetWidenedType(actual)) && l.sameKeeping(actual, element, map[[2]*checker.Type]bool{})
}

func (l *lowering) constructNodeArray(node *ast.Node) (ir.Expression, error) {
	layout, element, _ := l.nodeArrayLayoutOf(l.concrete(l.checker.GetTypeAtLocation(node)))
	value, err := l.arrayLiteral(ast.SkipParentheses(node.AsAsExpression().Expression))
	if err != nil {
		return nil, err
	}
	literal, known := value.(ir.ArrayLiteral)
	if !known {
		return nil, l.notYet(node, "NodeArray construction from a nonliteral array operation")
	}
	held, _ := l.representation(element)
	if literal.Element != held {
		return nil, l.notYet(node, "NodeArray element storage conversion")
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

func (l *lowering) nodeArrayProperty(node *ast.Node, object ir.Expression, name string) (ir.Expression, bool, error) {
	layout, _, known := l.nodeArrayLayoutOf(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression)))
	if !known {
		return nil, false, nil
	}
	field, err := layout.Read(name)
	if err != nil {
		return nil, true, l.notYet(node, "a NodeArray field outside its fixed metadata layout")
	}
	declared := l.checker.GetTypeOfSymbol(l.checker.GetSymbolAtLocation(node.Name()))
	value := ir.DynamicProperty{Object: fit(object, ir.Union), Name: name}
	checked, err := l.unsetSpecialization(node, value, declared, "field '"+name+"', factory 'NodeArray allocation'")
	_ = field // Layout membership is proved independently of observed presence.
	return checked, true, err
}
