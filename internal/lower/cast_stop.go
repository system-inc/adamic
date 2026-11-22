package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// castCause is the branch of castProof that found no road for a cast, so its stop can say which one
// (#b5w3ycg). Until Nov 22 every branch spoke one sentence, "a cast the runtime can't check", at 591 sites in
// tsc, and a reader couldn't tell a cast the ledger refuses from one it rules a check nobody has built yet.
type castCause int

const (
	// Refused in .ts as in .a, each by a row of docs/rulings.md's casts and views.
	castDouble         castCause = iota // an `as unknown as` pair that is no mint and no overload default
	castAny                             // a cast from or to any
	castFromUnknown                     // a cast from unknown to an object type
	castPrimitive                       // a cast whose target holds a primitive, or whose source is one
	castClassArguments                  // a class downcast whose type arguments the source's ancestry doesn't fix
	castRefinement                      // a cast from a union to something narrower than a member, not a member
	castUnionWrite                      // a union whose writable field a member can't hold every value of
	castAmbiguousTag                    // a union whose every member holds a literal in a field, but no two apart
	castCallable                        // a cast whose target or source has a call signature
	castWidenedMember                   // a union target whose member widens a writable field the source holds
	// Adamic's three rulings (Q1 to Q3 on #b5w3ycg, Nov 22), under one principle: in .ts a cast's stop is
	// NotYet when a run-time check could decide the claim and nobody has built it, and Refused only when no
	// check ever could.
	castIndexSignature // a target with an index signature: NotYet, the dictionary view (rulings.md:26 narrowed)
	castTypeParameter  // a type parameter source: NotYet under an object constraint, Refused unconstrained
	castIntersection   // an intersection target: NotYet, unless a member is refused on its own
	// NotYet in .ts: the untagged downcast row rules an unprovable cast to an object target a checked view,
	// and the row beside it rules an unbuilt view contract NotYet, never a refusal. Refused in .a (item 8).
	castUntaggedUnion // a union source whose members no literal tag tells apart
	castUnionTarget   // an object cast to a union target that no literal tag tells apart
	castKeepsNullish  // a target that keeps undefined or null
	castSideways      // an object target the source isn't a supertype of
	castMixedSource   // a union source holding primitives, cast to an object target
	// NotYet in .ts by Adamic's principle, as Compiler applied it (R4 and R4b on #b5w3ycg, Nov 22): instanceof
	// decides each at run time, and castProof's class road only wants every member a class.
	castClassAncestry      // a class cast with no nominal ancestry between the two
	castClassTarget        // a cast to a class from a value that isn't a class instance
	castFromObject         // a source typed object, which names no fields to read
	castDropsNullish       // a union source cast to its own members without undefined or null: a presence test
	castIntersectionMember // a union target one of whose members is an intersection
	// No branch above fits: the sentence main spoke before the split, so nothing is lost.
	castUnnamed
)

// castStop words a cast castProof found no road for. In .a every cause keeps main's refusal word for word,
// since item 8 refuses every checked downcast there and keptCheckDowncastRefusal names the narrowing. In
// .ts each cause says what stopped it: a Refused names its own fix, and a NotYet names the check it needs.
// source and target are the cast's as written; field is the field castUnionWrite names.
func (l *lowering) castStop(node *ast.Node, cause castCause, source, target *checker.Type, field string) error {
	if cause == castUnnamed && !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		// Scratch only, never lands: why the structural road missed each unnamed cast (#53amvnn's scout).
		present := l.checker.GetNonNullableType(target)
		from, fromKnown := l.representation(source)
		to, toKnown := l.representation(present)
		widened := l.widened(present, source, map[[2]*checker.Type]bool{}) != nil
		diag := []string{}
		add := func(name string, on bool) {
			if on {
				diag = append(diag, name)
			}
		}
		add("sourceUnion", source.Flags()&checker.TypeFlagsUnion != 0)
		add("targetUnion", present.Flags()&checker.TypeFlagsUnion != 0)
		add("sourceArray", l.castArrayLike(source))
		add("targetArray", l.castArrayLike(present))
		add("sourceTuple", checker.IsTupleType(l.checker.GetNonNullableType(source)))
		add("targetTuple", checker.IsTupleType(present))
		add("targetTypeParameter", present.Flags()&checker.TypeFlagsTypeParameter != 0)
		add("sourceTypeParameter", source.Flags()&checker.TypeFlagsTypeParameter != 0)
		add("targetNullish", l.includesUndefined(target) || l.includesNull(target))
		add("sourceNullish", l.includesUndefined(source) || l.includesNull(source))
		add("sourcePrimitive", castPrimitiveMember(source))
		add("targetAssignable", l.checker.IsTypeAssignableTo(present, source))
		add("widened", widened)
		add("targetIndex", len(l.checker.GetIndexInfosOfType(present)) != 0)
		add("fromUnknownRep", !fromKnown)
		add("toUnknownRep", !toKnown)
		add("fromNotObject", fromKnown && from != ir.Object)
		add("toNotObject", toKnown && to != ir.Object)
		add("targetClass", isClassInstance(present))
		return &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check: UNNAMED [" + strings.Join(diag, " ") + "] " + l.checker.TypeToString(source) + " -> " + l.checker.TypeToString(target), Fix: l.castRepair(node)}
	}
	if strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") || cause == castUnnamed {
		return &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: l.castRepair(node)}
	}
	from, to := l.checker.TypeToString(source), l.checker.TypeToString(target)
	// fix ends in its rule, as castRepair's does.
	refused := func(what, fix string) error {
		return &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check: " + what, Fix: fix}
	}
	const rule = " (adamic/no-unchecked-cast)"
	switch cause {
	case castDouble:
		return refused("an `as unknown as` pair from "+from+" to "+to+" proves nothing about the value", "write a checked view, a cast to "+to+" from the value's own type, or a phantom brand's mint"+rule)
	case castAny:
		return refused("a cast from "+from+" to "+to+", where any has no type to check", "type it: give the value its type where it's made, and cast from that type (adamic/no-explicit-any)")
	case castFromUnknown:
		return refused("a cast from unknown to "+to, "narrow it first with a type predicate or a typeof and in test, then cast from what they prove"+rule)
	case castPrimitive:
		if !castPrimitiveMember(l.checker.GetNonNullableType(target)) {
			return refused("a cast from "+from+", which holds a primitive, to the object type "+to, l.castRepair(node))
		}
		return refused("a cast from "+from+" to "+to+", which holds a primitive no view can check", "compare the primitive's typeof first (`if (typeof value !== 'string') panic('why')`), or convert it (`String(value)`)"+rule)
	case castClassArguments:
		return refused("a downcast from "+from+" to "+to+" whose type arguments "+from+"'s class ancestry doesn't fix, so its identity test can't tell "+to+" from another instantiation", "downcast to a class whose type parameters "+from+"'s ancestry passes through, or narrow by a field only "+to+" holds"+rule)
	case castRefinement:
		return refused("a cast from "+from+" to "+to+", which is narrower than a member of "+from+" without being one, and a tag test can't prove what makes it narrower", "cast to the member itself and test the rest, or declare "+to+" as a member with a tag of its own"+rule)
	case castUnionWrite:
		return refused("a cast over "+from+", whose writable field "+field+" some member can't hold every value of, so a write through the cast could break that member", "declare "+field+" readonly in "+from+"'s members, or cast from one member at a time"+rule)
	case castAmbiguousTag:
		return refused("a cast from "+from+" to "+to+", whose tag field "+field+" holds a literal in every member but doesn't tell each apart (a value two members share, two enums' members equal at run time, or a member that holds any number)", "give each member of "+from+" a "+field+" of its own, one literal per member"+rule)
	case castCallable:
		return refused("a cast from "+from+" to "+to+", which asserts a call signature", "prove the callable body rather than asserting its signature"+rule)
	case castWidenedMember:
		return refused("a cast from "+from+" to "+to+", whose member "+field, "declare the field readonly in both, or cast to a member whose fields "+from+" holds as written"+rule)
	case castIndexSignature:
		return l.notYet(node, "a checked dictionary view of "+to+" over "+from)
	case castTypeParameter:
		constraint := l.checker.GetBaseConstraintOfType(source)
		if constraint == nil || constraint == source || constraint.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsAny) != 0 || castPrimitiveMember(constraint) {
			return refused("a cast from the type parameter "+from+" to "+to+", which nothing at run time can test, since "+from+" may be any type", "constrain "+from+" to an object type (`"+from+" extends object`), or cast from the value's concrete type"+rule)
		}
		return l.notYet(node, "a checked view of "+to+" over the type parameter "+from+", constrained to "+l.checker.TypeToString(constraint))
	case castIntersection:
		// A member castProof refuses alone keeps that refusal, in the member's own words and fix.
		for _, member := range l.checker.GetNonNullableType(target).Types() {
			if cause := l.objectCastCause(source, member); cause <= castCallable {
				return l.castStop(node, cause, source, member, "")
			}
		}
		return l.notYet(node, "a checked view of the intersection "+to+" over "+from)
	case castUntaggedUnion:
		return l.notYet(node, "a checked view of "+to+" over the union "+from+", whose members no literal tag tells apart")
	case castUnionTarget:
		if l.distinctTagField(castMembers(l.checker.GetNonNullableType(target))) {
			return l.notYet(node, "a checked view of the union "+to+" over "+from+", whose tag the kind-union road can't test from "+from)
		}
		return l.notYet(node, "a checked view of the union "+to+" over "+from+", whose members no literal tag tells apart")
	case castIntersectionMember:
		return l.notYet(node, "a checked view of the union "+to+" over "+from+", whose member "+field+" is an intersection")
	case castDropsNullish:
		return l.notYet(node, "a checked cast from "+from+" to "+to+", which removes "+l.castKeptNullish(source))
	case castKeepsNullish:
		return l.notYet(node, "a checked cast from "+from+" to "+to+", which keeps "+l.castKeptNullish(target))
	case castSideways:
		return l.notYet(node, "a checked view of "+to+" over "+from+", which "+to+" doesn't extend")
	case castFromObject:
		return l.notYet(node, "a checked view of "+to+" over a value typed "+from+", which names no fields")
	case castClassAncestry, castClassTarget:
		return l.notYet(node, "a checked class cast from "+from+" to "+to+" by instanceof")
	case castMixedSource:
		return l.notYet(node, "a checked cast from "+from+", which holds primitives, to "+to+", which needs a typeof test before its view")
	}
	return &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: l.castRepair(node)}
}

// castArrayLike is whether some member of proven, nullish aside, is an array or a tuple. A cast between
// arrays isn't named yet: its own cause, an element checked one by one, is a later commit, so it keeps main's
// sentence rather than borrow a field's words (an array's methods read as writable fields to castUnionWrites,
// and its number index as an index signature).
func (l *lowering) castArrayLike(proven *checker.Type) bool {
	for _, member := range castMembers(l.checker.GetNonNullableType(proven)) {
		if l.arrayLikeType(member) || checker.IsTupleType(member) {
			return true
		}
	}
	return false
}

// castKeptNullish names what a target keeps beside its objects: undefined, null, or both.
func (l *lowering) castKeptNullish(target *checker.Type) string {
	switch undefined, null := l.includesUndefined(target), l.includesNull(target); {
	case undefined && null:
		return "undefined and null"
	case null:
		return "null"
	}
	return "undefined"
}

// castPrimitiveMember is whether some member of proven, nullish members aside, is a primitive.
func castPrimitiveMember(proven *checker.Type) bool {
	for _, member := range castMembers(proven) {
		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) == 0 && member.Flags()&checker.TypeFlagsPrimitive != 0 {
			return true
		}
	}
	return false
}

// objectCastCause is why castProof's one-object branch found no road: no tag the target declares, no kind
// union, no structural view. The order puts a ledger refusal before a NotYet, so a cast that is both (a
// primitive target that keeps undefined) keeps its refusal.
func (l *lowering) objectCastCause(source, target *checker.Type) castCause {
	present := l.checker.GetNonNullableType(target)
	switch {
	case l.castArrayLike(source) || l.castArrayLike(present):
		return castUnnamed
	case source.Flags()&checker.TypeFlagsUnknown != 0:
		return castFromUnknown
	case castPrimitiveMember(present) || castPrimitiveMember(source):
		return castPrimitive
	case source.Flags()&checker.TypeFlagsTypeParameter != 0:
		return castTypeParameter
	case source.Flags()&checker.TypeFlagsNonPrimitive != 0 && present.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(present) && len(l.checker.GetIndexInfosOfType(present)) == 0 && len(l.checker.GetSignaturesOfType(present, checker.SignatureKindCall)) == 0 && !l.includesUndefined(target) && !l.includesNull(target):
		return castFromObject
	case present.Flags()&checker.TypeFlagsIntersection != 0:
		return castIntersection
	case len(l.checker.GetIndexInfosOfType(present)) != 0:
		return castIndexSignature
	case isClassInstance(present):
		return castClassTarget
	case len(l.checker.GetSignaturesOfType(present, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(source, checker.SignatureKindCall)) != 0:
		return castCallable
	case l.includesUndefined(target) || l.includesNull(target):
		return castKeepsNullish
	case present.Flags()&checker.TypeFlagsUnion != 0:
		cause, _ := l.unionTargetCause(source, present)
		return cause
	case present.Flags()&checker.TypeFlagsObject != 0 && !l.checker.IsTypeAssignableTo(present, source):
		return castSideways
	}
	return castUnnamed
}

// distinctTagField is whether some field holds a literal in every member, each apart from the others, as
// unionMemberProof's tag must.
func (l *lowering) distinctTagField(members []*checker.Type) bool {
	for _, property := range l.checker.GetPropertiesOfType(members[0]) {
		literals := []*checker.Type{}
		for _, member := range members {
			literal := l.fieldLiteral(member, property.Name)
			if literal == nil || l.tagHoldsAnyNumber(literal) {
				literals = nil
				break
			}
			for _, previous := range literals {
				if !tagLiteralsDistinct(previous, literal) {
					literals = nil
				}
			}
			if literals == nil {
				break
			}
			literals = append(literals, literal)
		}
		if literals != nil {
			return true
		}
	}
	return false
}

// unionCastCause is why unionMemberProof found no road for a union source's members cast to target, once a
// writable field and a refinement are ruled out.
func (l *lowering) unionCastCause(members []*checker.Type, target *checker.Type) castCause {
	present := l.checker.GetNonNullableType(target)
	for _, member := range append(append([]*checker.Type{}, members...), present) {
		if l.castArrayLike(member) {
			return castUnnamed
		}
	}
	primitive := false
	for _, member := range members {
		primitive = primitive || castPrimitiveMember(member)
	}
	switch {
	case castPrimitiveMember(present):
		return castPrimitive
	case present.Flags()&checker.TypeFlagsIntersection != 0:
		return castIntersection
	case len(l.checker.GetIndexInfosOfType(present)) != 0:
		return castIndexSignature
	case primitive:
		return castMixedSource
	case l.includesUndefined(target) || l.includesNull(target):
		return castKeepsNullish
	}
	nullish, others := false, []*checker.Type{}
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) != 0 {
			nullish = true
		} else {
			others = append(others, member)
		}
	}
	if nullish && len(others) > 0 {
		dropped := true
		for _, member := range others {
			dropped = dropped && l.castAssignable(member, present)
		}
		if dropped {
			return castDropsNullish
		}
	}
	return castUntaggedUnion
}

// unionTargetCause is why kindUnionCast took no cast from an object source to the union target present,
// in its own order: an intersection member (a later head), a member that widens a writable field the
// source holds (no check can stop a later write through the source), and otherwise the view nobody has
// built. detail words the member for castStop.
func (l *lowering) unionTargetCause(source, present *checker.Type) (castCause, string) {
	for _, member := range present.Types() {
		if member.Flags()&checker.TypeFlagsIntersection != 0 {
			return castIntersectionMember, l.checker.TypeToString(member)
		}
	}
	for _, member := range present.Types() {
		if member.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(member) || !l.checker.IsTypeAssignableTo(member, source) {
			continue
		}
		if found := l.widened(member, source, map[[2]*checker.Type]bool{}); found != nil {
			detail := l.checker.TypeToString(member) + " can't be seen as " + l.checker.TypeToString(source) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read"
			if found.readonlyField != "" {
				detail = l.checker.TypeToString(member) + " can't be seen as " + l.checker.TypeToString(source) + ", whose field " + found.readonlyField + " " + l.checker.TypeToString(member) + " keeps readonly and " + l.checker.TypeToString(source) + " can write"
			}
			return castWidenedMember, detail
		}
	}
	return castUnionTarget, ""
}
