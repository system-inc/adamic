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
	source := l.concrete(l.checker.GetTypeAtLocation(as.Expression))
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.sameKeeping(source, target, map[[2]*checker.Type]bool{}) {
		return nil, l.notYet(node, "a cast that changes the runtime representation or ownership of a reference")
	}
	value, err := l.expression(as.Expression)
	if err != nil {
		return nil, err
	}
	if proof.structuralView {
		return l.structuralViewCast(node, value, source, target)
	}
	if proof.interfaceView {
		return l.interfaceCast(node, value, l.concrete(l.checker.GetTypeAtLocation(as.Expression)), l.concrete(l.checker.GetTypeAtLocation(node)))
	}
	if len(proof.allowed) == 0 && len(proof.classes) == 0 {
		return value, nil
	}
	refused := &Refused{Where: l.program.Where(node), What: "a cast without an object tag representation", Fix: castRepair}
	if value.Type() != ir.Object {
		return nil, refused
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
	if _, err := l.view(node, value, target); err != nil {
		return nil, err
	}
	cast.CheckedFields = true
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
