package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A tag tells members apart only by the value it holds at run time, never by the type the checker gives
// it. Two enum members with one value, `enum K { First = 1, Alias = 1 }`, are two literal types to the
// checker and one value to a test, so a test of K.First is true for a K.Alias. rulings.md:13 refuses a
// cast over such a union ("duplicate enum values"), and these two rules are that condition, in one place
// for every reader: the cast (cast_proof.go), a branded claim's tagged arm (brandedTaggedArm, Adamic's
// ruling on #fx9waqy, Nov 7), and the predicate verifier's discriminant proof (#fvmqee2).

// tagLiteralsDistinct is whether a run-time test of one tag literal can never be true of the other: both
// are the same representation, string, number or boolean, and their values differ. Two representations
// fail too, since a mixed tag is refused, not compared (rulings.md:13's "mixed tag representations").
func tagLiteralsDistinct(first, second *checker.Type) bool {
	mask := checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBooleanLiteral
	return first.Flags()&mask == second.Flags()&mask && !reflect.DeepEqual(first.AsLiteralType().Value(), second.AsLiteralType().Value())
}

// enumValueUnique is whether tag, a tag literal, holds a value no other member of its own enum holds. A
// literal that isn't an enum member has no aliases, so it's unique. Every declaration of the enum is read,
// since declarations merge, and a member whose value the checker can't compute fails closed, since it may
// be the alias.
func (l *lowering) enumValueUnique(tag *checker.Type) bool {
	symbol := tag.Symbol()
	if tag.Flags()&checker.TypeFlagsEnumLiteral == 0 || symbol == nil || symbol.Flags&ast.SymbolFlagsEnumMember == 0 {
		return true
	}
	if symbol.ValueDeclaration == nil {
		return false
	}
	value := l.checker.GetConstantValue(symbol.ValueDeclaration)
	enum := l.enumIdentity(tag)
	if value == nil || enum == nil {
		return false
	}
	for _, declaration := range enum.Declarations {
		if declaration.Kind != ast.KindEnumDeclaration {
			return false
		}
		for _, member := range declaration.AsEnumDeclaration().Members.Nodes {
			if member == symbol.ValueDeclaration {
				continue
			}
			other := l.checker.GetConstantValue(member)
			if other == nil || reflect.DeepEqual(other, value) {
				return false
			}
		}
	}
	return true
}
