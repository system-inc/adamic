package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// An overload promising Extract<U, Pick<T, "kind">>, as tsc's injectClassNamedEvaluationHelperBlockIfMissing
// does (transformers/namedEvaluation.ts:173), is a narrowing the declaration can't judge: T is rigid there, so
// the promise is a union of deferred conditional types, not a member overloadResultCheckable can read. At a
// call it is concrete, the members of U whose kind the argument's kind carries. So it is served by
// rulings.md:128's call-site check, in .ts only, as every narrow overload is (narrowOverloadJudged), where the
// tag the check reads is sound by rulings.md:247's premise: a kind's value says which member a value is only
// when no other member could carry that value (#0kkpw2e set 1, Adamic's ruling).

// overloadKindExtract is whether promised, an overload's declared result, is a distributed Extract<U, Pick<T,
// "kind">> over produced, the implementation's result, whose every member a check can tell by its kind: U is
// produced itself, T a type parameter, and each member of produced is an object whose kind is a plain field
// typed exactly one enum member's literal, a value no other member's kind carries. A getter-backed kind (any
// accessor named kind the program writes, since an interface's field can be served by one, accessorName), an
// intersection member, a kind that is a union of literals or a whole enum, and two members whose kinds share a
// value, one enum's aliases or two enums' members, each leave it refused.
func (l *lowering) overloadKindExtract(promised, produced *checker.Type) bool {
	extract := promised.Alias()
	if extract == nil || extract.Symbol().Name != "Extract" || !l.librarySymbol(extract.Symbol()) || len(extract.TypeArguments()) != 2 {
		return false
	}
	pick := extract.TypeArguments()[1].Alias()
	if pick == nil || pick.Symbol().Name != "Pick" || !l.librarySymbol(pick.Symbol()) || len(pick.TypeArguments()) != 2 {
		return false
	}
	keys := pick.TypeArguments()[1]
	if pick.TypeArguments()[0].Flags()&checker.TypeFlagsTypeParameter == 0 || keys.Flags()&checker.TypeFlagsStringLiteral == 0 || keys.AsLiteralType().Value() != "kind" {
		return false
	}
	if !checker.Checker_isTypeIdenticalTo(l.checker, extract.TypeArguments()[0], produced) || l.accessorName("kind") {
		return false
	}
	members := []*checker.Type{produced}
	if produced.Flags()&checker.TypeFlagsUnion != 0 {
		members = produced.Types()
	}
	values := map[any]bool{}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject == 0 {
			return false
		}
		property := l.checker.GetPropertyOfType(member, "kind")
		if property == nil || property.Flags&(ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor|ast.SymbolFlagsMethod|ast.SymbolFlagsOptional) != 0 {
			return false
		}
		kind := l.checker.GetTypeOfSymbol(property)
		if kind.Flags()&checker.TypeFlagsEnumLiteral == 0 || kind.Flags()&(checker.TypeFlagsNumberLiteral|checker.TypeFlagsStringLiteral) == 0 || l.openNumericEnumType(kind) {
			return false
		}
		value := kind.AsLiteralType().Value()
		if values[value] {
			return false // Another member carries this value, so the kind can't say which member a value is.
		}
		values[value] = true
	}
	return true
}
