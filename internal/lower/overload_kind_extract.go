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
// "kind">> over produced, the implementation's result, one Extract per member of it as the checker distributes
// it, whose every member a check can tell by its kind: each arm names one type parameter T, and each member of
// produced is an object whose kind is a plain field typed exactly one enum member's literal, a value no other
// member's kind carries. A getter-backed kind (any accessor named kind the program writes, since an
// interface's field can be served by one, accessorName), an intersection member, a kind that is a union of
// literals or a whole enum, and two members whose kinds share a value, one enum's aliases or two enums'
// members, each leave it refused.
func (l *lowering) overloadKindExtract(promised, produced *checker.Type) bool {
	if l.accessorName("kind") {
		return false
	}
	members := []*checker.Type{produced}
	if produced.Flags()&checker.TypeFlagsUnion != 0 {
		members = produced.Types()
	}
	// Every arm picks kind alone, and the arms' subjects are produced's members, each once.
	subjects, keys, read := l.overloadKindExtractArms(promised)
	if !read || keys.Flags()&checker.TypeFlagsStringLiteral == 0 || keys.AsLiteralType().Value() != "kind" {
		return false
	}
	covered := map[*checker.Type]bool{}
	for _, subject := range subjects {
		if checker.Checker_isTypeIdenticalTo(l.checker, subject, produced) {
			for _, member := range members {
				covered[member] = true
			}
			continue
		}
		found := false
		for _, member := range members {
			if !covered[member] && checker.Checker_isTypeIdenticalTo(l.checker, subject, member) {
				covered[member], found = true, true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(covered) != len(members) {
		return false
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

// overloadKindExtractArms reads promised, an overload's declared result, as a distributed Extract<U, Pick<T,
// K>>: the checker distributes Extract over a union U where it's declared, so the promise arrives as one
// Extract<member, Pick<T, K>> per member of U, and as one Extract only when U isn't a union. Each arm is the
// library's Extract over the library's Pick of one type parameter T, the same T and the same K in every arm.
// It returns the arms' subjects, U's members as the checker distributed them, and K, which its callers judge:
// overloadKindExtract serves only K exactly "kind", and overloadKindExtractIntersection names the gap for a
// K that holds "kind" among other keys too.
func (l *lowering) overloadKindExtractArms(promised *checker.Type) ([]*checker.Type, *checker.Type, bool) {
	arms := []*checker.Type{promised}
	if promised.Alias() == nil && promised.Flags()&checker.TypeFlagsUnion != 0 {
		arms = promised.Types()
	}
	var binder, keys *checker.Type
	subjects := []*checker.Type{}
	for _, arm := range arms {
		extract := arm.Alias()
		if extract == nil || extract.Symbol().Name != "Extract" || !l.librarySymbol(extract.Symbol()) || len(extract.TypeArguments()) != 2 {
			return nil, nil, false
		}
		pick := extract.TypeArguments()[1].Alias()
		if pick == nil || pick.Symbol().Name != "Pick" || !l.librarySymbol(pick.Symbol()) || len(pick.TypeArguments()) != 2 {
			return nil, nil, false
		}
		if pick.TypeArguments()[0].Flags()&checker.TypeFlagsTypeParameter == 0 {
			return nil, nil, false
		}
		if binder != nil && (pick.TypeArguments()[0] != binder || pick.TypeArguments()[1] != keys) {
			return nil, nil, false
		}
		binder, keys = pick.TypeArguments()[0], pick.TypeArguments()[1]
		subjects = append(subjects, extract.TypeArguments()[0])
	}
	return subjects, keys, true
}

// tsc's transformNamedEvaluation (transformers/namedEvaluation.ts:469 at the stage 3 pin) promises
// Extract<NamedEvaluation, Pick<T, "kind" | keyof T & "operatorToken" | keyof T & "name">>, whose members
// are intersections, a node type with its name or operator narrowed, and one of tsc's callers instantiates
// it as never. Its implementation returns the plain node types, so no member serves an arm by identity,
// and a kind check at the call can't tell a narrowed member from its plain node. Serving it is the
// adaptation road, annotating tsc's helpers, so until then it says so as its own gap rather than as the
// covariance refusal every other unserved result gets (#0kkpw2e set 2, Adamic's ruling (c)).

// overloadKindExtractIntersection is whether promised reads as a distributed Extract by kind
// (overloadKindExtractArms, with K "kind" alone or a union holding it) one of whose subjects is an
// intersection or isn't produced or one of its members, so the implementation's result can't serve it.
func (l *lowering) overloadKindExtractIntersection(promised, produced *checker.Type) bool {
	subjects, keys, read := l.overloadKindExtractArms(promised)
	if !read {
		return false
	}
	kinds := []*checker.Type{keys}
	if keys.Flags()&checker.TypeFlagsUnion != 0 {
		kinds = keys.Types()
	}
	byKind := false
	for _, key := range kinds {
		if key.Flags()&checker.TypeFlagsStringLiteral != 0 && key.AsLiteralType().Value() == "kind" {
			byKind = true
		}
	}
	if !byKind {
		return false
	}
	members := []*checker.Type{produced}
	if produced.Flags()&checker.TypeFlagsUnion != 0 {
		members = append(members, produced.Types()...)
	}
	for _, subject := range subjects {
		if subject.Flags()&checker.TypeFlagsIntersection != 0 {
			return true
		}
		served := false
		for _, member := range members {
			if checker.Checker_isTypeIdenticalTo(l.checker, subject, member) {
				served = true
			}
		}
		if !served {
			return true
		}
	}
	return false
}
