package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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
	castClassAncestry                   // a class cast with no nominal ancestry between the two
	castClassTarget                     // a cast to a class from a value that isn't a class instance
	castRefinement                      // a cast from a union to something narrower than a member, not a member
	castUnionWrite                      // a union whose writable field a member can't hold every value of
	castAmbiguousTag                    // a union whose every member holds a literal in a field, but no two apart
	castCallable                        // a cast whose target or source has a call signature
	// Refused until Adamic rules them (Q1 to Q3 on #b5w3ycg, Nov 22).
	castIndexSignature // a target with an index signature
	castTypeParameter  // a source that is a type parameter
	castIntersection   // an intersection target
	// NotYet in .ts: the untagged downcast row rules an unprovable cast to an object target a checked view,
	// and the row beside it rules an unbuilt view contract NotYet, never a refusal. Refused in .a (item 8).
	castUntaggedUnion // a union source whose members no literal tag tells apart
	castUnionTarget   // an object cast to a union target that no literal tag tells apart
	castKeepsNullish  // a target that keeps undefined or null
	castSideways      // an object target the source isn't a supertype of
	castMixedSource   // a union source holding primitives, cast to an object target
	castFromObject    // a source typed object, which names no fields to read
	// No branch above fits: the sentence main spoke before the split, so nothing is lost.
	castUnnamed
)

// castStop words a cast castProof found no road for. In .a every cause keeps main's refusal word for word,
// since item 8 refuses every checked downcast there and keptCheckDowncastRefusal names the narrowing. In
// .ts each cause says what stopped it: a Refused names its own fix, and a NotYet names the check it needs.
// source and target are the cast's as written; field is the field castUnionWrite names.
func (l *lowering) castStop(node *ast.Node, cause castCause, source, target *checker.Type, field string) error {
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
	case castClassAncestry:
		return refused("a cast from "+from+" to "+to+", which isn't a class descending from it", "downcast along nominal class ancestry (a class that extends "+from+"), or build the value as a "+to+rule)
	case castClassTarget:
		return refused("a cast from "+from+" to the class "+to+", which "+from+" isn't by ancestry", "build the value with new, or downcast from a class along its ancestry"+rule)
	case castRefinement:
		return refused("a cast from "+from+" to "+to+", which is narrower than a member of "+from+" without being one, and a tag test can't prove what makes it narrower", "cast to the member itself and test the rest, or declare "+to+" as a member with a tag of its own"+rule)
	case castUnionWrite:
		return refused("a cast over "+from+", whose writable field "+field+" some member can't hold every value of, so a write through the cast could break that member", "declare "+field+" readonly in "+from+"'s members, or cast from one member at a time"+rule)
	case castAmbiguousTag:
		return refused("a cast from "+from+" to "+to+", whose tag field "+field+" holds a literal in every member but doesn't tell each apart (a value two members share, two enums' members equal at run time, or a member that holds any number)", "give each member of "+from+" a "+field+" of its own, one literal per member"+rule)
	case castCallable:
		return refused("a cast from "+from+" to "+to+", which asserts a call signature", "prove the callable body rather than asserting its signature"+rule)
	case castIndexSignature:
		return refused("a cast from "+from+" to "+to+", whose index signature a checked view doesn't hold", l.castRepair(node))
	case castTypeParameter:
		return refused("a cast from the type parameter "+from+" to "+to+", which a check would have to read through its erased type", l.castRepair(node))
	case castIntersection:
		return refused("a cast from "+from+" to the intersection "+to, l.castRepair(node))
	case castUntaggedUnion:
		return l.notYet(node, "a checked view of "+to+" over the union "+from+", whose members no literal tag tells apart")
	case castUnionTarget:
		return l.notYet(node, "a checked view of the union "+to+" over "+from+", whose members no literal tag tells apart")
	case castKeepsNullish:
		return l.notYet(node, "a checked cast from "+from+" to "+to+", which keeps "+l.castKeptNullish(target))
	case castSideways:
		return l.notYet(node, "a checked view of "+to+" over "+from+", which "+to+" doesn't extend")
	case castFromObject:
		return l.notYet(node, "a checked view of "+to+" over a value typed "+from+", which names no fields")
	case castMixedSource:
		return l.notYet(node, "a checked cast from "+from+", which holds primitives, to "+to+", which needs a typeof test before its view")
	}
	return &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: l.castRepair(node)}
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
		return castUnionTarget
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
	return castUntaggedUnion
}
