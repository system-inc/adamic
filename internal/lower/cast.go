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
	proof, err := l.castProof(node)
	if err != nil {
		return nil, err
	}
	value, err := l.expression(as.Expression)
	if err != nil {
		return nil, err
	}
	sourceType, targetType := l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node)
	if value.Type() == ir.Object && targetType.Flags()&checker.TypeFlagsObject != 0 {
		for _, field := range l.checker.GetPropertiesOfType(targetType) {
			member := l.enumAnnotatedMember(field)
			if member == nil || !l.openNumericEnumType(l.checker.GetTypeOfSymbol(field)) {
				continue
			}
			source := l.checker.GetPropertyOfType(sourceType, field.Name)
			if source != nil && l.enumAnnotatedMember(source) == member {
				continue
			}
			constant, err := l.enumConstant(member)
			if err != nil {
				return nil, err
			}
			value = ir.CheckedCast{Value: value, Field: field.Name, FieldType: ir.Number, Allowed: []ir.Expression{constant}, Message: "numeric enum member view failed: field " + field.Name}
		}
	}

	if len(proof.allowed) == 0 && len(proof.classes) == 0 {
		return value, nil
	}
	source := l.concrete(l.checker.GetTypeAtLocation(as.Expression))
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	refused := &Refused{Where: l.program.Where(node), What: "a cast without an object tag representation", Fix: castRepair}
	if value.Type() != ir.Object {
		return nil, refused
	}
	message := "cast failed: this " + l.checker.TypeToString(source) + " is not a " + l.checker.TypeToString(target)
	if len(proof.classes) > 0 {
		return l.checkedClassCast(node, value, proof.classes, message)
	}
	value, err = l.enumTagViewAs(node, value, source, target)
	if err != nil {
		return nil, err
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

func (l *lowering) enumAnnotatedMember(field *ast.Symbol) *ast.Node {
	if len(field.Declarations) == 0 {
		return nil
	}
	annotation := field.Declarations[0].Type()
	if annotation == nil || annotation.Kind != ast.KindTypeReference {
		return nil
	}
	symbol := l.symbol(annotation.AsTypeReferenceNode().TypeName)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnumMember != 0 && l.numericEnum(l.symbol(symbol.ValueDeclaration.Parent.Name())) {
		return symbol.ValueDeclaration
	}
	return nil
}
