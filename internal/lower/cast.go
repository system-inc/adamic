package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// cast lowers value as T. as const and a cast to a type the value already has (an upcast) are the
// value itself. A downcast to members of a discriminated union is checked at runtime: the tag is
// there to read. Any other cast is a type the compiler would have to trust, and 0.1 refuses it
// (adamic/no-unchecked-cast).
func (l *lowering) cast(node *ast.Node) (ir.Expression, error) {
	as := node.AsAsExpression()
	if l.nodeRequirePerformanceProjection(node) {
		return l.expression(as.Expression)
	}
	value, err := l.expression(as.Expression)
	if err != nil {
		return nil, err
	}
	if as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
		return value, nil
	}
	source := l.concrete(l.checker.GetTypeAtLocation(as.Expression))
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	if handled, err := l.phantomArrayCast(node, as.Expression, source, target); handled {
		return value, err
	}
	if l.phantomCast(source, target) {
		return value, nil
	}
	// An object literal's type is fresh, and a fresh type is held to excess properties, so { name, age }
	// as Named would read as not assignable. Its widened type isn't fresh; its own keeps the literal
	// fields a discriminated union needs. Either assignable makes an upcast candidate; the sound
	// relation below still has to prove its writable slots, function views and nominal ancestry.
	if l.checker.IsTypeAssignableTo(source, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target) {
		if err := l.provenRelation(node, as.Expression, target); err != nil {
			return nil, err
		}
		return value, nil
	}
	if checked, err := l.viewArrayCast(node, value, source, target); checked != nil || err != nil {
		return checked, err
	}
	if checked, err := l.viewProvenClassCast(node, value, source, target); checked != nil || err != nil {
		return checked, err
	}
	refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: "narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast)"}
	if source.Flags()&checker.TypeFlagsUnion == 0 {
		if checked, err := l.interfaceCast(node, value, source, target); checked != nil || err != nil {
			return checked, err
		}
	}
	if source.Flags()&checker.TypeFlagsUnion == 0 {
		if checked, err := l.structuralViewCast(node, value, source, target); checked != nil || err != nil {
			return checked, err
		}
	}
	if value.Type() != ir.Object || source.Flags()&checker.TypeFlagsUnion == 0 {
		return nil, refused
	}
	proof, err := l.castProof(node)
	if err != nil {
		return nil, err
	}
	message := "cast failed: this " + l.checker.TypeToString(source) + " is not a " + l.checker.TypeToString(target)
	if len(proof.classes) > 0 {
		return l.checkedClassCast(node, value, proof.classes, message)
	}
	cast := ir.CheckedCast{Value: value, Field: proof.field, Message: message}
	for _, literal := range proof.allowed {
		allowed, fieldType, constant := l.literalConstant(literal)
		if !constant || (cast.FieldType != 0 && cast.FieldType != fieldType) {
			return nil, refused
		}
		cast.Allowed = append(cast.Allowed, allowed)
		cast.FieldType = fieldType
	}
	return cast, nil
}

// fieldLiteral is a member's type for a field when it's a single literal ('Circle', 1, true), the
// kind of field a runtime check can read the member from.
func (l *lowering) fieldLiteral(member *checker.Type, field string) *checker.Type {
	for _, property := range l.checker.GetPropertiesOfType(member) {
		if property.Name != field {
			continue
		}
		fieldType := l.checker.GetTypeOfSymbol(property)
		if fieldType.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral) != 0 {
			return fieldType
		}
	}
	return nil
}

// literalConstant is a literal type as a constant to compare a field with.
func (l *lowering) literalConstant(literal *checker.Type) (ir.Expression, ir.Type, bool) {
	flags := literal.Flags()
	switch {
	case flags&checker.TypeFlagsStringLiteral != 0:
		text, isText := literal.AsLiteralType().Value().(string)
		return ir.StringConstant{Index: l.constant(text)}, ir.String, isText
	case flags&checker.TypeFlagsNumberLiteral != 0:
		value := reflect.ValueOf(literal.AsLiteralType().Value())
		if value.Kind() != reflect.Float64 {
			return nil, 0, false
		}
		return ir.NumberConstant{Value: value.Float()}, ir.Number, true
	case flags&checker.TypeFlagsBooleanLiteral != 0:
		return ir.BooleanConstant{Value: l.checker.TypeToString(literal) == "true"}, ir.Boolean, true
	}
	return nil, 0, false
}
