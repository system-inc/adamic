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
func (l *lowering) castValue(node *ast.Node) (ir.Expression, error) {
	as := node.AsAsExpression()
	proof, err := l.castProof(node)
	if err != nil {
		return nil, err
	}
	if proof.speculation {
		return l.checkedSpeculativeResult(node)
	}
	if proof.nodeArray {
		return l.constructNodeArray(node)
	}
	if proof.construction {
		return l.constructLiteral(node)
	}
	if proof.brandMint != nil {
		return l.brandMintValue(node, as, proof.brandMint)
	}
	if proof.overloadDefault {
		// The pair's hop to unknown makes no value of its own: x is lowered as itself, and it must already
		// be stored as the target reads it.
		operand := ast.SkipParentheses(as.Expression).AsAsExpression().Expression
		from, target := l.concrete(l.checker.GetTypeAtLocation(operand)), l.concrete(l.checker.GetTypeAtLocation(node))
		if !l.sameKeeping(from, target, map[[2]*checker.Type]bool{}) {
			return nil, l.notYet(node, "a proven default cast that changes the runtime representation of "+l.checker.TypeToString(from))
		}
		// x goes where the pair goes, never into the hop's unknown, which would read it as a value viewed
		// as unknown (valueSlotTags).
		l.keepSlot(operand, l.checker.GetTypeAtLocation(node))
		return l.expression(operand)
	}
	if proof.present {
		// The claim is presence alone, so the cast is the check `!` makes, with the value and type `!` gives:
		// one presence check for both spellings, not a second one here.
		value, err := l.expression(as.Expression)
		if err != nil {
			return nil, err
		}
		return l.nonNullValue(node, value)
	}
	if proof.dictionaryView {
		// The view changes the representation on purpose, a union to the record it tests for, so it's lowered
		// before sameKeeping, which would stop it.
		return l.dictionaryViewValue(node)
	}
	source := l.concrete(l.checker.GetTypeAtLocation(as.Expression))
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	// The message names the source as written, undefined included; everything after the absent check
	// reads the present members (absentArm).
	message := l.castFailure(source, target)
	if proof.arrayArm != nil {
		return l.arrayArmValue(node, proof.arrayArm, source, target, message)
	}
	if proof.keepsNullish {
		return l.keepsNullishValue(node, proof, source, target, message)
	}
	if proof.absent {
		source = l.absentArm(source, target)
	}
	if !l.sameKeeping(source, target, map[[2]*checker.Type]bool{}) {
		return nil, l.notYet(node, "a cast that changes the runtime representation or ownership of a reference")
	}
	value, err := l.expression(as.Expression)
	if err != nil {
		return nil, err
	}
	if proof.absent {
		// The absent check comes first, with the cast's own message: an undefined value is a member the
		// target doesn't hold, so it fails the cast as a wrong tag would, and the tag is never read.
		// An object that may be undefined is held as an Object whose undefined is a null reference.
		if value.Type() != ir.Object {
			return nil, l.notYet(node, "a member cast whose source holds undefined, held as other than an object")
		}
		value = ir.Coalesce{Value: value, Panic: ir.StringConstant{Index: l.constant(message)}, Of: ir.Object}
	}
	return l.castRoadValue(node, proof, value, source, target, message)
}

// castFailure is row 9's message (docs/rulings.md:9), the words every tag check panics with where a value isn't the
// type it's claimed as: `cast failed: this <source> is not a <target>`, each type as the checker prints it. A cast's
// check, a non-null assertion written as a cast, and ruling 17's dispatch check (dispatch_lower.go) all say it.
func (l *lowering) castFailure(source, target *checker.Type) string {
	return "cast failed: this " + l.checker.TypeToString(source) + " is not a " + l.checker.TypeToString(target)
}

// keepsNullishValue lowers a cast that keeps its source's undefined or null (keepsNullishCast): the operand
// once, into a local, and then a nullish value is itself, and a present one takes the road castRoad found
// between the present types, with the cast's own message. The road's test reads the tag, so it sits only in
// the present arm. A present pair with no test at all (an upcast) is the value itself. The test is the one
// optionalIndex makes, IsNull beside IsUndefined where the source holds null, since the JavaScript backend
// holds an object's null as null and IsUndefined there is === undefined.
func (l *lowering) keepsNullishValue(node *ast.Node, proof castProof, source, target *checker.Type, message string) (ir.Expression, error) {
	present, wanted := l.checker.GetNonNullableType(source), l.checker.GetNonNullableType(target)
	// castValue's check, made on the pair the road runs on: castRoadValue doesn't make it.
	if !l.sameKeeping(present, wanted, map[[2]*checker.Type]bool{}) {
		return nil, l.notYet(node, "a cast that changes the runtime representation or ownership of a reference")
	}
	value, err := l.expression(node.AsAsExpression().Expression)
	if err != nil {
		return nil, err
	}
	// Only an upcast, a proof that asks no check, is the value itself. checksRoad is castRoadValue's own list, so a
	// road added there is a road here: a list of this function's own once missed the untagged arm and the class
	// road, and a cast that kept undefined over an untagged union lowered to the bare operand (#ejthypb, J-zewecvk).
	if !proof.checksRoad() {
		return value, nil
	}
	// An object that may be undefined or null is held as an Object whose nullish value is a null reference,
	// and the result is one too. A value held otherwise, a Union, would need its own tests: it keeps the stop.
	held, known := l.representation(target)
	if value.Type() != ir.Object || !known || held != ir.Object {
		return nil, l.castStop(node, castKeepsNullish, source, target, "")
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "cast_operand", Type: ir.Object, Function: function, ExpressionAssigned: true})
	read := ir.Read{Local: local, Of: ir.Object}
	cast, err := l.castRoadValue(node, proof, read, present, wanted, message)
	if err != nil {
		return nil, err
	}
	nullish := ir.Expression(ir.IsUndefined{Value: read})
	if l.includesNull(source) {
		nullish = ir.Binary{Operator: ir.Or, Left: nullish, Right: ir.IsNull{Value: read}}
	}
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: value}}, Result: ir.Conditional{
		Condition: nullish, WhenTrue: read, WhenNot: cast, Of: ir.Object,
	}}, nil
}

// checksRoad is whether castRoadValue lowers the proof to a check: a view (overlap, untagged, structural or
// interface), a kind union's test, a tag test, or a class test. A proof that answers false is an upcast, which
// castRoadValue gives back as the value itself. It's the one list of castRoadValue's roads, read by castRoadValue's
// own fallback, by keepsNullishValue and by keptCheckDowncastRefusal, so a road added to castRoadValue is added here
// and every reader sees it (Compiler's one-list ask on #eg3krrp, after #ejthypb).
func (proof castProof) checksRoad() bool {
	return proof.overlap || proof.untagged || proof.kindUnion != nil || proof.structuralView || proof.interfaceView || len(proof.allowed) > 0 || len(proof.classes) > 0
}

// castRoadValue lowers the road a proof found over a value already lowered: the view, the tag test or the
// class test, with message as its failure. It takes the value, source and target rather than reading the
// cast's operand, so a composed cast can lower its inner road over a value it made itself (the present
// value of a cast that keeps nullish, once tested, #b5w3ycg). node is the cast, for its place and its
// view's site.
func (l *lowering) castRoadValue(node *ast.Node, proof castProof, value ir.Expression, source, target *checker.Type, message string) (ir.Expression, error) {
	if proof.overlap {
		return l.overlapViewCast(node, value, source, target)
	}
	if proof.untagged {
		return l.untaggedUnionCast(node, value, source, target)
	}
	if proof.kindUnion != nil {
		return l.kindUnionValue(node, value, source, proof.kindUnion, message)
	}
	if proof.structuralView {
		return l.structuralViewCast(node, value, source, target)
	}
	if proof.interfaceView {
		checked, err := l.interfaceCast(node, value, source, target)
		// castProof found the tag on an object placeholder. A value held otherwise, an array cast to a
		// tuple whose length is the tag (never[] as []), gets no checked cast back, and nil is no value.
		if err == nil && checked == nil {
			return nil, l.notYet(node, "a checked cast from "+l.checker.TypeToString(source)+" to "+l.checker.TypeToString(target)+" whose value isn't a plain object")
		}
		return checked, err
	}
	if !proof.checksRoad() {
		return value, nil
	}
	refused := &Refused{Where: l.program.Where(node), What: "a cast without an object tag representation", Fix: l.castRepair(node)}
	if value.Type() != ir.Object {
		return nil, refused
	}
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
	if _, err := l.view(node, value, source, target); err != nil {
		return nil, err
	}
	l.viewSources[len(l.viewSources)-1].cast = true
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
