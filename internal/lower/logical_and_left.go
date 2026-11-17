package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// In x && y the left operand reaches the result only when it's falsy, and a represented object is true
// (rulings.md:184, conditions are JavaScript's ToBoolean). So an arm of x's type that can only be an object
// never reaches the result, and the relation into the result's slot is judged with those arms dropped
// (#y1mw6wj, Adamic's ruling on #w6qgtzg). tsc's builder.ts:1601, `oldProgram && oldProgram.getProgramOrUndefined()`,
// relates only undefined then, which is what the checker's own type for the && says, where judging the
// whole BuilderProgram | undefined refused a view no value can take.

// andLeftOperand reports whether node is the left operand of &&, through parentheses.
func andLeftOperand(node *ast.Node) bool {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	return parent != nil && parent.Kind == ast.KindBinaryExpression &&
		parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.AsBinaryExpression().Left == child
}

// falsyArms is own without its arms that can only be objects, the arms that can reach the result of an && it's
// the left of; nil when no arm is left, so the left never reaches the result.
func (l *lowering) falsyArms(own *checker.Type) *checker.Type {
	arms := []*checker.Type{own}
	if own.Flags()&checker.TypeFlagsUnion != 0 {
		arms = own.Types()
	}
	var kept []*checker.Type
	for _, arm := range arms {
		if !l.onlyObject(arm, map[*checker.Type]bool{}) {
			kept = append(kept, arm)
		}
	}
	switch len(kept) {
	case 0:
		return nil
	case len(arms):
		return own
	case 1:
		return kept[0]
	}
	return l.checker.GetUnionType(kept)
}

// onlyObject reports whether no primitive can be a value of t, so a value of t is always truthy, proved by t's own
// declaration rather than by asking whether a primitive is assignable to it: tsc's assignability holds one relation
// at a time, and a primitive reaches a type it isn't assignable to through another it is ("" as Sized
// { length: number; foo?: number }, then Sized as { foo?: number }, Language's must on #y1mw6wj). So an object type is
// proved when it has a call or construct signature, which no primitive has, or a required property that none of
// string's, number's, boolean's, bigint's and symbol's apparent types has by name, since every view of the value on
// the way kept that property required. `{}`, `Object`, a weak type, a type of index signatures alone and a
// structural type whose required members a primitive has, such as `{ length: number }`, are kept. A union is proved
// when each member is; a type parameter when its base constraint is. An intersection is kept, a brand among them,
// and so is anything else: `unknown`, `any`, a primitive, an unconstrained type parameter.
func (l *lowering) onlyObject(t *checker.Type, visiting map[*checker.Type]bool) bool {
	if visiting[t] {
		return false
	}
	visiting[t] = true
	switch {
	case t.Flags()&checker.TypeFlagsTypeParameter != 0:
		constraint := l.checker.GetBaseConstraintOfType(t)
		return constraint != nil && constraint != t && l.onlyObject(constraint, visiting)
	case t.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range t.Types() {
			if !l.onlyObject(member, visiting) {
				return false
			}
		}
		return len(t.Types()) != 0
	case t.Flags()&checker.TypeFlagsObject == 0:
		return false
	}
	if len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) != 0 {
		return true
	}
	for _, property := range l.checker.GetPropertiesOfType(t) {
		if property.Flags&ast.SymbolFlagsOptional == 0 && !l.primitiveHas(property.Name) {
			return true
		}
	}
	return false
}

// primitiveHas reports whether any primitive's apparent type has a property of this name: String, Number,
// Boolean, BigInt or Symbol, each with Object's members, as the checker reads a property off a primitive.
func (l *lowering) primitiveHas(name string) bool {
	for _, primitive := range []*checker.Type{l.checker.GetStringType(), l.checker.GetNumberType(), l.checker.GetBooleanType(), l.checker.GetBigIntType(), l.checker.GetESSymbolType()} {
		if l.checker.GetPropertyOfType(primitive, name) != nil {
			return true
		}
	}
	return false
}
