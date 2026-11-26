package lower

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// In x && y the left operand reaches the result only when it's falsy, and a represented object is true
// (rulings.md:184, conditions are JavaScript's ToBoolean). So an arm of x's type that can only be an object
// never reaches the result, and the relation into the result's slot is judged with those arms dropped
// (#y1mw6wj, Adamic's ruling on #w6qgtzg). tsc's builder.ts:1601, `oldProgram && oldProgram.getProgramOrUndefined()`,
// relates only undefined then, which is what the checker's own type for the && says, where judging the
// whole BuilderProgram | undefined refused a view no value can take.

// andOfLeft is the && expression node is the left operand of, through parentheses, or nil.
func andOfLeft(node *ast.Node) *ast.Node {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent != nil && parent.Kind == ast.KindBinaryExpression &&
		parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.AsBinaryExpression().Left == child {
		return parent
	}
	return nil
}

// andFalsyDropped reports whether the && expression and, through parentheses, is consumed where a falsy result is
// dropped before any slot sees it (Adamic's ruling of Nov 19 on #snypn8c, rule 3, extending rulings.md:184): the left
// operand of an ||, which then evaluates its right instead, the operand of !, or the condition of an if, a while, a
// do, a for or a conditional, each of which reads only the result's truth. A falsy result there is the left's, so
// the left's value reaches no slot through the &&, and only the right's value is judged. tsc's checker.ts:21987,
// `!strictVariance && compareTypes(...) || compareTypes(...)`, relates only the Ternary its right returns then, where
// judging the left's false against the Ternary slot refused a value no slot is handed. The right of an || hands the
// && result on whole, falsy or not, and ?? keeps every falsy value, so neither drops it, and a result stored or passed
// first is judged where it goes, as it is. Every walk that judges an && left asks this one helper
// (TestAndFalsyDroppedWalks), so no walk drops the left by a copy of its own.
func andFalsyDropped(and *ast.Node) bool {
	child := and
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken && parent.AsBinaryExpression().Left == child
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == child
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == child
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == child
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == child
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == child
	}
	return false
}

// falsyArms is own without the arms that can never be falsy, the arms that can reach the result of an && it's the
// left of; nil when no arm is left, so the left never reaches the result. A left's falsy values are its type's actual
// falsy values (Adamic's ruling of Nov 19 on #snypn8c, rule 3): false, 0 and NaN for number, 0n, "", null and
// undefined. So an arm that can only be an object is dropped (onlyObject), and so is one that can only be truthy
// (truthyLiteral): true, a number or numeric enum literal other than 0, a bigint literal other than 0n, and a string
// or string enum literal other than "". boolean is true | false to the checker, so its arms split already, and false
// is kept. number, string and bigint are kept whole, since NaN, 0, "" and 0n are among their values. A numeric
// enum's whole type is open (rulings.md:254): it holds any number, NaN and 0 among them, so when every member of one
// is among the arms, as tsc flattens the enum's own type, those members are kept, and only a member literal or a
// strict subset of an enum's members is taken at its values.
func (l *lowering) falsyArms(own *checker.Type) *checker.Type {
	arms := []*checker.Type{own}
	if own.Flags()&checker.TypeFlagsUnion != 0 {
		arms = own.Types()
	}
	whole, impure := l.wholeNumericEnums(arms)
	var kept []*checker.Type
	for _, arm := range arms {
		if l.onlyObject(arm, map[*checker.Type]bool{}) {
			continue
		}
		if l.truthyLiteral(arm) && !l.openEnumMember(arm, whole, impure) {
			continue
		}
		kept = append(kept, arm)
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

// openEnumMember reports whether arm, one of a union's arms, is a member literal of a numeric enum every member of
// which is among the arms (whole, as wholeNumericEnums reads them), so it's part of that enum's open type, or a
// numeric member literal beside an enum that isn't purely numeric (impure), which fails closed.
func (l *lowering) openEnumMember(arm *checker.Type, whole []wholeNumericEnum, impure bool) bool {
	if arm.Flags()&checker.TypeFlagsEnumLiteral == 0 || arm.Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return false
	}
	return impure || slices.ContainsFunc(whole, func(enum wholeNumericEnum) bool { return enum.symbol == l.enumIdentity(arm) })
}

// truthyLiteral reports whether t is a literal type whose one value is truthy under JavaScript's ToBoolean
// (rulings.md:184): true, a number literal other than 0 (and -0, which compares equal to it), a bigint literal other
// than 0n, or a string literal other than "", an enum's member literals among them. NaN is no literal type.
func (l *lowering) truthyLiteral(t *checker.Type) bool {
	switch flags := t.Flags(); {
	case flags&checker.TypeFlagsBooleanLiteral != 0:
		return l.checker.TypeToString(t) == "true"
	case flags&checker.TypeFlagsNumberLiteral != 0:
		number := reflect.ValueOf(t.AsLiteralType().Value())
		return number.Kind() == reflect.Float64 && number.Float() != 0
	case flags&checker.TypeFlagsBigIntLiteral != 0:
		return fmt.Sprint(t.AsLiteralType().Value()) != "0"
	case flags&checker.TypeFlagsStringLiteral != 0:
		text, isText := t.AsLiteralType().Value().(string)
		return isText && text != ""
	}
	return false
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
