package lower

import (
	"os"
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
	// Refused, #53amvnn: no check can hold the claim once another alias writes the value, or a ruling refuses it.
	castWidenedTarget  // an object downcast to a target that widens a writable field the source holds
	castMutableArray   // a downcast to a mutable array, which a write through the source can break (invariant-mutable)
	castTupleFromArray // a mutable array cast to a tuple, whose length another alias of the array can still change
	// Adamic's three rulings (Q1 to Q3 on #b5w3ycg, Nov 22), under one principle: in .ts a cast's stop is
	// NotYet when a run-time check could decide the claim and nobody has built it, and Refused only when no
	// check ever could.
	// The view is built for object to a dictionary of unknown values (dictionaryView), and NotYet for the rest.
	castIndexSignature // a target with an index signature: NotYet, the dictionary view (rulings.md:26 narrowed)
	castTypeParameter  // a type parameter source: NotYet under an object constraint, Refused unconstrained
	castIntersection   // an intersection target: NotYet, unless a member is refused on its own
	// NotYet in .ts: the untagged downcast row rules an unprovable cast to an object target a checked view,
	// and the row beside it rules an unbuilt view contract NotYet, never a refusal. Refused in .a (item 8).
	castUntaggedUnion // a union source whose members no literal tag tells apart
	castUnionTarget   // an object cast to a union target that no literal tag tells apart
	castKeepsNullish  // a target that keeps undefined or null, past keepsNullishCast's road
	castSideways      // an object target the source isn't a supertype of
	castMixedSource   // a union source holding primitives, cast to an object target
	// NotYet under the same row by Adamic's ruling on #bbssdcy (item 19, Nov 24): a gap, not a lie, whose
	// checked view is the untagged view road's to build.
	castFieldRefinement // an object target the source is a supertype of, which reads a field of the source narrower
	// NotYet in .ts by Adamic's principle, as Compiler applied it (R4 and R4b on #b5w3ycg, Nov 22): instanceof
	// decides each at run time, and castProof's class road only wants every member a class.
	castClassAncestry      // a class cast with no nominal ancestry between the two
	castClassTarget        // a cast to a class from a value that isn't a class instance
	castFromObject         // a source typed object, which names no fields to read
	castDropsNullish       // a union source cast to its members without undefined or null, past presentCast's road
	castIntersectionMember // a union target one of whose members is an intersection
	// NotYet in .ts, #53amvnn, by the same principle: a check could decide each.
	castTargetTypeParameter // a type parameter target: NotYet under an object constraint, Refused unconstrained
	castArrayArm            // a union of arrays and other members, cast to one side: Array.isArray decides
	castTuple               // an array cast to a tuple: a length check, and a tuple's own representation
	castFromEmpty           // a source typed {}, which holds any value but null and undefined
	// NotYet by Compiler's ruling on #53amvnn (Nov 23): views v3's per-read element check decides a readonly
	// array's claim at each read, so a write through another alias is caught where it's read.
	castArrayElements // a downcast of a readonly array's elements
	// No branch above fits: the sentence main spoke before the split, so nothing is lost.
	castUnnamed
)

// castStop words a cast castProof found no road for. In .a every cause keeps main's refusal word for word,
// since item 8 refuses every checked downcast there and keptCheckDowncastRefusal names the narrowing. In
// .ts each cause says what stopped it: a Refused names its own fix, and a NotYet names the check it needs.
// source and target are the cast's as written; field is the field castUnionWrite or castFieldRefinement names.
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
	case castWidenedTarget:
		return refused("a cast from "+from+" to "+to+", "+field, "declare the field readonly in both, or build the value as a "+to+rule)
	case castMutableArray:
		return refused("a cast from "+from+" to the mutable array "+to+", which a write through "+from+" could break later", "cast to a readonly array, or build a new "+to+" of the checked elements (adamic/invariant-mutable)")
	case castTupleFromArray:
		return refused("a cast from "+from+" to the tuple "+to+", whose length a check at the cast can't hold, since another alias of the array can still change it", "build the tuple from the checked elements ([first, ...rest]), or keep the array and check its length where it's read"+rule)
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
		return l.notYet(node, "a checked view of "+to+" over the union "+from+", whose members no literal tag tells apart"+l.untaggedWhy(source, target))
	case castUnionTarget:
		if l.distinctTagField(castMembers(l.checker.GetNonNullableType(target))) {
			return l.notYet(node, "a checked view of the union "+to+" over "+from+", whose tag the kind-union road can't test from "+from)
		}
		return l.notYet(node, "a checked view of the union "+to+" over "+from+", whose members no literal tag tells apart"+l.untaggedTargetWhy(source, target))
	case castIntersectionMember:
		return l.notYet(node, "a checked view of the union "+to+" over "+from+", whose member "+field+" is an intersection")
	case castDropsNullish:
		return l.notYet(node, "a checked cast from "+from+" to "+to+", which removes "+l.castKeptNullish(source))
	case castKeepsNullish:
		return l.notYet(node, "a checked cast from "+from+" to "+to+", which keeps "+l.castKeptNullish(target))
	case castSideways:
		return l.notYet(node, "a checked view of "+to+" over "+from+", which "+to+" doesn't extend")
	case castFieldRefinement:
		return l.notYet(node, "a checked view of "+to+" over "+from+", which reads the field "+field+" of "+from+" narrower")
	case castFromObject:
		return l.notYet(node, "a checked view of "+to+" over a value typed "+from+", which names no fields")
	case castTargetTypeParameter:
		present := l.checker.GetNonNullableType(target)
		constraint := l.checker.GetBaseConstraintOfType(present)
		if constraint == nil || constraint == present || constraint.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsAny) != 0 || castPrimitiveMember(constraint) {
			return refused("a cast from "+from+" to the type parameter "+to+", which nothing at run time can test, since "+to+" may be any type", "constrain "+to+" to an object type, or return the value at its own type"+rule)
		}
		return l.notYet(node, "a checked view of the type parameter "+to+", constrained to "+l.checker.TypeToString(constraint)+", over "+from)
	case castArrayArm:
		return l.notYet(node, "a checked cast from "+from+" to "+to+", which needs Array.isArray to tell "+from+"'s arrays from its other members")
	case castTuple:
		return l.notYet(node, "a checked cast from "+from+" to the tuple "+to+", which needs its length checked")
	case castArrayElements:
		return l.notYet(node, "a checked view of the readonly array "+to+"'s elements over "+from)
	case castFromEmpty:
		return l.notYet(node, "a checked view of "+to+" over a value typed {}, which holds anything but null and undefined")
	case castClassAncestry, castClassTarget:
		return l.notYet(node, "a checked class cast from "+from+" to "+to+" by instanceof")
	case castMixedSource:
		return l.notYet(node, "a checked cast from "+from+", which holds primitives, to "+to+", which needs a typeof test before its view")
	}
	return &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: l.castRepair(node)}
}

// arrayCastCause is why a cast with an array on either side found no road (#53amvnn), source given by its
// members and target without its nullish members. A union of arrays and other members cast to one side is told
// by Array.isArray, when the side it picks holds no element downcast; a cast to a tuple needs its length; a
// downcast of elements is refused, readonly or mutable, since another alias can still write the array.
func (l *lowering) arrayCastCause(members []*checker.Type, present *checker.Type) castCause {
	arrays, others := []*checker.Type{}, 0
	for _, member := range members {
		switch {
		case member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) != 0:
		case l.castArrayLike(member):
			arrays = append(arrays, member)
		default:
			others++
		}
	}
	targetArray := l.castArrayLike(present)
	switch {
	case targetArray && checker.IsTupleType(present) && len(arrays) > 0:
		// A tuple or readonly array source is decided by a length and element check at each read; a mutable
		// array's length can still change through another alias, so it's refused.
		for _, member := range arrays {
			if !checker.IsTupleType(member) && !l.isLibraryType(member, "ReadonlyArray") {
				return castTupleFromArray
			}
		}
		return castTuple
	case !targetArray && len(arrays) > 0 && others > 0:
		return castArrayArm
	case targetArray && others > 0:
		for _, member := range arrays {
			if l.castAssignable(member, present) && l.castAssignable(present, member) {
				return castArrayArm
			}
		}
	}
	if !targetArray || len(arrays) == 0 {
		return castUnnamed
	}
	// The checker's array question answers for ReadonlyArray too, so readonly is asked by name.
	if l.isArrayType(present) && !l.isLibraryType(present, "ReadonlyArray") {
		return castMutableArray
	}
	return castArrayElements
}

// castSignatures is whether proven has a call or construct signature: a function type holds no properties, so
// the {} source cause must ask before calling one empty.
func (l *lowering) castSignatures(proven *checker.Type) bool {
	return len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0
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

// castExtendsArray is whether some member of proven, nullish aside, is an interface whose base-type chain, as
// the checker resolves it (nodeArrayBase), reaches the global Array or ReadonlyArray: tsc's NodeArray<T>,
// MutableNodeArray<T>, MissingList<T> and SortedArray<T>. IsArrayType doesn't take such an interface, so its
// number index reads as an index signature, and a cast to or from one would be judged a dictionary view where
// Node runs it over an array (#2v3xtgm, Compiler's ruling). It's asked only where a cause is chosen, never in
// arrayLikeType, whose other readers would move. A chain the checker can't read gives nodeArrayBase nil, so
// the cast keeps the cause it had: it's never read as an array, or as a dictionary, on a guess.
func (l *lowering) castExtendsArray(proven *checker.Type) bool {
	for _, member := range castMembers(l.checker.GetNonNullableType(proven)) {
		if l.nodeArrayBase(member, map[*checker.Type]bool{}) != nil {
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
	case castPrimitiveMember(present):
		return castPrimitive
	case present.Flags()&checker.TypeFlagsTypeParameter != 0:
		return castTargetTypeParameter
	// An interface extending an array is an array's cast too, before its number index can read as a dictionary
	// (#2v3xtgm): arrayCastCause, which counts only real arrays, answers castUnnamed for it. The mutant:
	// castExtendsArray answers false, and TestCastStopNamesItsCause/missing_list goes red, a NotYet dictionary view
	// where the array's refusal is wanted.
	case l.castArrayLike(source) || l.castArrayLike(present) || l.castExtendsArray(source) || l.castExtendsArray(present):
		return l.arrayCastCause(castMembers(source), present)
	case source.Flags()&checker.TypeFlagsObject != 0 && len(l.checker.GetPropertiesOfType(source)) == 0 && !l.castSignatures(source) && present.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(present) && !l.castSignatures(present):
		return castFromEmpty
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
	case present.Flags()&checker.TypeFlagsObject != 0 && l.checker.IsTypeAssignableTo(present, source) && l.widened(present, source, map[[2]*checker.Type]bool{}) != nil:
		return castWidenedTarget
	case present.Flags()&checker.TypeFlagsObject != 0 && !l.checker.IsTypeAssignableTo(present, source):
		return castSideways
	case l.refinedField(source, present) != "":
		return castFieldRefinement
	}
	return castUnnamed
}

// refinedField is the first field a field-refinement downcast reads narrower, or "" when the cast is none
// (Adamic's item 19 on #bbssdcy): present, the target without its nullish members, is a plain object with no
// index, call or construct signature; source is an object, or an intersection of objects, so it holds no
// primitive, type parameter or nullish member; present is assignable to source; every property of present is a
// data property, as is each one source holds of the same name, never a method or an accessor; and some field of
// present holds a type assignable to source's for that field, but not the reverse. Item 19 rules a field read
// narrower, so a method that returns something narrower (lib's Promise seen over PromiseLike, whose then does) is
// none, and keeps main's sentence. That it widens none of source's writable fields is objectCastCause's order,
// which asks castWidenedTarget first, and a target that does keeps that refusal: no view can hold its claim once
// source writes the wider value. tsc's toBuilderStateFileInfoForMultiEmit (builder.ts:2233 in TypeScript 6.0.3) is
// one: its source is an Omit<...> & { readonly signature } in the adapted tree, which castRoad's structural view
// doesn't take, since that view wants a source that is one object.
func (l *lowering) refinedField(source, present *checker.Type) string {
	if present.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(present) || len(l.checker.GetIndexInfosOfType(present)) != 0 || l.castSignatures(present) {
		return ""
	}
	if source.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range source.Types() {
			if member.Flags()&checker.TypeFlagsObject == 0 {
				return ""
			}
		}
	} else if source.Flags()&checker.TypeFlagsObject == 0 {
		return ""
	}
	if !l.checker.IsTypeAssignableTo(present, source) {
		return ""
	}
	const notField = ast.SymbolFlagsMethod | ast.SymbolFlagsGetAccessor | ast.SymbolFlagsSetAccessor
	field := ""
	for _, property := range l.checker.GetPropertiesOfType(present) {
		inside := l.checker.GetPropertyOfType(source, property.Name)
		if property.Flags&notField != 0 || inside != nil && inside.Flags&notField != 0 {
			return ""
		}
		if inside == nil || field != "" {
			continue
		}
		read, held := l.checker.GetTypeOfSymbol(property), l.checker.GetTypeOfSymbol(inside)
		if l.checker.IsTypeAssignableTo(read, held) && !l.checker.IsTypeAssignableTo(held, read) {
			field = property.Name
		}
	}
	return field
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
	if castPrimitiveMember(present) {
		return castPrimitive
	}
	if present.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return castTargetTypeParameter
	}
	// As in objectCastCause, an interface extending an array is kept out of castIndexSignature (#2v3xtgm). The
	// mutant: castExtendsArray answers false, and TestCastStopNamesItsCause/node_or_node_array goes red.
	for _, member := range append(append([]*checker.Type{}, members...), present) {
		if l.castArrayLike(member) || l.castExtendsArray(member) {
			return l.arrayCastCause(members, present)
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

// widenedWords says how a target seen as the source widens a writable field, for castWidenedTarget and
// castWidenedMember alike.
func (l *lowering) widenedWords(member, source *checker.Type) string {
	found := l.widened(member, source, map[[2]*checker.Type]bool{})
	if found == nil {
		return "which " + l.checker.TypeToString(source) + " can write wider"
	}
	if found.readonlyField != "" {
		return "whose field " + found.readonlyField + " " + l.checker.TypeToString(member) + " keeps readonly and " + l.checker.TypeToString(source) + " can write"
	}
	return "which " + l.checker.TypeToString(source) + " can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read"
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

// untaggedWhy is proof-only (#eg3krrp scout, never merged): with ADAMIC_UNTAGGED_WHY=1, the first of
// untaggedUnionArm's conditions that declined this cast, so the family splits by cause.
func (l *lowering) untaggedWhy(source, target *checker.Type) string {
	if os.Getenv("ADAMIC_UNTAGGED_WHY") != "1" {
		return ""
	}
	members, targets := castMembers(source), castMembers(target)
	plainWhy := func(proven *checker.Type) string {
		switch {
		case proven.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) != 0:
			return "nullish"
		case proven.Flags()&checker.TypeFlagsIntersection != 0:
			return "intersection"
		case proven.Flags()&checker.TypeFlagsObject == 0:
			return "not an object"
		case isClassInstance(proven):
			return "class instance"
		case l.arrayLikeType(proven) || checker.IsTupleType(proven):
			return "array or tuple"
		}
		if of, known := l.representation(proven); !known || of != ir.Object {
			return "not held as an object"
		}
		if base, _ := l.phantomExtension(proven); base != nil {
			return "phantom extension"
		}
		switch {
		case len(l.checker.GetIndexInfosOfType(proven)) != 0:
			return "index signature"
		case len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0:
			return "callable"
		}
		return ""
	}
	why := ""
	switch {
	case len(members) < 2:
		why = "one member"
	case len(targets) == 0:
		why = "no target members"
	case l.unionHoldsLiterals(members):
		why = "members hold a literal field"
	case len(targets) > 1 && l.unionHoldsLiterals(targets):
		why = "target union holds a literal field"
	}
	if why == "" {
		if w := l.widened(source, target, map[[2]*checker.Type]bool{}); w != nil {
			kind := "a field"
			switch {
			case w.readonlyField != "":
				kind = "readonly field " + w.readonlyField + " becomes writable"
			case w.enum:
				kind = "an enum literal slot"
			case w.methodField != "":
				kind = "method " + w.methodField
			case w.parameter:
				kind = "a parameter"
			case w.genericMethod != "":
				kind = "generic method " + w.genericMethod
			}
			slot := ""
			if w.source != nil && w.target != nil {
				slot = " (" + l.checker.TypeToString(w.source) + " as " + l.checker.TypeToString(w.target) + ")"
			}
			if len(slot) > 120 {
				slot = slot[:120] + "...)"
			}
			why = "widened: " + kind + slot
		}
	}
	if why == "" {
		for _, member := range members {
			if reason := plainWhy(member); reason != "" {
				why = "member " + l.checker.TypeToString(member) + ": " + reason
				break
			}
		}
	}
	if why == "" {
		for _, wanted := range targets {
			if reason := plainWhy(wanted); reason != "" {
				why = "target " + l.checker.TypeToString(wanted) + ": " + reason
				break
			}
			downcast := false
			for _, member := range members {
				downcast = downcast || l.castAssignable(wanted, member)
			}
			if !downcast {
				why = "target " + l.checker.TypeToString(wanted) + ": sideways"
				break
			}
		}
	}
	if why == "" {
		why = "the arm would take it: reached by another road"
	}
	return " [why: " + why + "]"
}

// untaggedTargetWhy is proof-only (#eg3krrp scout, never merged): with ADAMIC_UNTAGGED_WHY=1, why an object
// cast to an untagged union target isn't union target (a)'s arm, read as that arm would with the source as its
// one member.
func (l *lowering) untaggedTargetWhy(source, target *checker.Type) string {
	if os.Getenv("ADAMIC_UNTAGGED_WHY") != "1" {
		return ""
	}
	members, targets := castMembers(source), castMembers(target)
	why := ""
	if len(members) >= 2 {
		why = "a union source"
	}
	if why == "" && len(targets) > 1 && l.unionHoldsLiterals(targets) {
		why = "target union holds a literal field"
	}
	if why == "" {
		if w := l.widened(source, target, map[[2]*checker.Type]bool{}); w != nil {
			switch {
			case w.readonlyField != "":
				why = "widened: readonly field " + w.readonlyField + " becomes writable"
			case w.enum:
				why = "widened: an enum literal slot"
			default:
				why = "widened: another slot"
			}
		}
	}
	if why == "" && !l.plainViewObject(members[0]) {
		why = "source not a plain view object"
	}
	if why == "" {
		for _, wanted := range targets {
			if !l.plainViewObject(wanted) {
				why = "target member " + l.checker.TypeToString(wanted) + " not a plain view object"
				break
			}
			if !l.castAssignable(wanted, members[0]) {
				why = "target member " + l.checker.TypeToString(wanted) + " sideways"
				break
			}
		}
	}
	if why == "" {
		why = "one object source: union target (a)'s arm would take it but for wanting a union source"
	}
	return " [why: " + why + "]"
}
