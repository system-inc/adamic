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
func (lowering *lowering) cast(node *ast.Node) (ir.Expression, error) {
	as := node.AsAsExpression()
	proof, castProofError := lowering.castProof(node)
	if castProofError != nil {
		return nil, castProofError
	}
	if proof.speculation {
		return lowering.checkedSpeculativeResult(node)
	}
	if proof.nodeArray {
		return lowering.constructNodeArray(node)
	}
	if proof.construction {
		return lowering.constructLiteral(node)
	}
	if proof.brandMint != nil {
		return lowering.brandMintValue(node, as, proof.brandMint)
	}
	if proof.overloadDefault {
		// The pair's hop to unknown makes no value of its own: x is lowered as itself, and it must already
		// be stored as the target reads it.
		operand := ast.SkipParentheses(as.Expression).AsAsExpression().Expression
		from, target := lowering.concrete(
			lowering.checker.GetTypeAtLocation(operand),
		), lowering.concrete(lowering.checker.GetTypeAtLocation(node))
		if !lowering.sameKeeping(from, target, map[[2]*checker.Type]bool{}) {
			return nil, lowering.notYet(
				node,
				"a proven default cast that changes the runtime representation of "+lowering.checker.TypeToString(from),
			)
		}
		// x goes where the pair goes, never into the hop's unknown, which would read it as a value viewed
		// as unknown (valueSlotTags).
		lowering.keepSlot(operand, lowering.checker.GetTypeAtLocation(node))
		return lowering.expression(operand)
	}
	source := lowering.concrete(lowering.checker.GetTypeAtLocation(as.Expression))
	target := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	if !lowering.sameKeeping(source, target, map[[2]*checker.Type]bool{}) {
		return nil, lowering.notYet(node, "a cast that changes the runtime representation or ownership of a reference")
	}
	value, expressionError := lowering.expression(as.Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	if proof.structuralView {
		return lowering.structuralViewCast(node, value, source, target)
	}
	if proof.interfaceView {
		checked, interfaceCastError := lowering.interfaceCast(node, value, source, target)
		// castProof found the tag on an object placeholder. A value held otherwise, an array cast to a
		// tuple whose length is the tag (never[] as []), gets no checked cast back, and nil is no value.
		if interfaceCastError == nil && checked == nil {
			return nil, lowering.notYet(node, "a checked cast from "+
				lowering.checker.TypeToString(source)+
				" to "+
				lowering.checker.TypeToString(target)+
				" whose value isn't a plain object")
		}
		return checked, interfaceCastError
	}
	if len(proof.allowed) == 0 && len(proof.classes) == 0 {
		return value, nil
	}
	refused := &Refused{
		Where: lowering.program.Where(node),
		What:  "a cast without an object tag representation",
		Fix:   lowering.castRepair(node),
	}
	if value.Type() != ir.Object {
		return nil, refused
	}
	message := "cast failed: this " +
		lowering.checker.TypeToString(source) +
		" is not a " +
		lowering.checker.TypeToString(target)
	if len(proof.classes) > 0 {
		return lowering.checkedClassCast(node, value, proof.classes, message)
	}
	cast := ir.CheckedCast{Value: value, Field: proof.field, Message: message}
	for _, literal := range proof.allowed {
		allowed, fieldType, constant := lowering.literalConstant(literal)
		if !constant || (cast.FieldType != 0 && cast.FieldType != fieldType) {
			return nil, refused
		}
		cast.Allowed = append(cast.Allowed, allowed)
		cast.FieldType = fieldType
	}
	if _, viewError := lowering.view(node, value, source, target); viewError != nil {
		return nil, viewError
	}
	lowering.viewSources[len(lowering.viewSources)-1].cast = true
	cast.CheckedFields = true
	return cast, nil
}

// fieldLiteral is a member's type for a field when it's a single literal ('Circle', 1, true), the
// kind of field a runtime check can read the member from.
func (lowering *lowering) fieldLiteral(member *checker.Type, field string) *checker.Type {
	for _, property := range lowering.checker.GetPropertiesOfType(member) {
		if property.Name != field {
			continue
		}
		fieldType := lowering.checker.GetTypeOfSymbol(property)
		if fieldType.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral) != 0 {
			return fieldType
		}
	}
	return nil
}

// literalConstant is a literal type as a constant to compare a field with.
func (lowering *lowering) literalConstant(literal *checker.Type) (ir.Expression, ir.Type, bool) {
	flags := literal.Flags()
	switch {
	case flags&checker.TypeFlagsStringLiteral != 0:
		text, isText := literal.AsLiteralType().Value().(string)
		return ir.StringConstant{Index: lowering.constant(text)}, ir.String, isText
	case flags&checker.TypeFlagsNumberLiteral != 0:
		value := reflect.ValueOf(literal.AsLiteralType().Value())
		if value.Kind() != reflect.Float64 {
			return nil, 0, false
		}
		return ir.NumberConstant{Value: value.Float()}, ir.Number, true
	case flags&checker.TypeFlagsBooleanLiteral != 0:
		return ir.BooleanConstant{Value: lowering.checker.TypeToString(literal) == "true"}, ir.Boolean, true
	}
	return nil, 0, false
}
