package lower

import (
	"slices"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Mutable locations are invariant in 0.1 (docs/0.1.md, adamic/invariant-mutable): tsc relates an
// array's elements, a property and a map's values covariantly, so a Dog[] is accepted as an Animal[],
// and the wider name can then write a cat where the dogs are read. Stage 0 doesn't run cohere's lint,
// so it refuses the same views itself, before lowering: every place a value goes into a typed slot
// (an initializer, an assignment, an argument, a return, a literal's property or element), the value's
// type is walked against the slot's, all the way down, as the Weak view check walks them.
//
// At each mutable slot the wider side can write, what it writes must be something the narrower side
// can read: the slot's type in the target must be assignable to its type in the source. tsc proved
// the other direction when it accepted the site, so together the slot is invariant. A readonly slot
// only reads, and stays covariant.
//
// In .a, an index signature is a slot too, under every name a declared property doesn't take: a
// declared property seen through a wider one can be written what it can't hold, and a dictionary
// seen through a declared property can be read as what it doesn't hold (records_declared_keys.go).
//
// Two shapes reach a mutable array where tsc's relation can't be walked part by part, and each is
// judged as cohere's rule judges it (cohere 838c6516):
//   - A type parameter target (const pack: Pack = narrow, Narrow extends Pack extends Animal[]): what
//     Pack holds is no type to compare against until it's instantiated, so a type parameter whose
//     constraint can be written is one mutable slot, and only the parameter itself (or it narrowed,
//     as T & {}) fills it.
//   - An intersection with an array (Dog[] & { tag } into Animal[] & { tag }): its array, tuple and
//     container members are paired member by member, as well as its properties.

// widening is a view refused: the parts of the source and the target at the slot where the target can
// write what the source can't read.
type widening struct {
	source, target *checker.Type

	// readonlyField names a readonly field seen as a writable one, when that's the slot.
	readonlyField string
	enum          bool

	// methodField names a method seen as a writable field, when that's the slot (#9dmh4em).
	methodField string

	// parameter is a function's parameter of type source seen as taking target, which it can't.
	parameter bool

	// genericMethod names a generic method, of type source, seen as the view's method of type target,
	// which stage 0 doesn't relate yet (widenedProperties).
	genericMethod string

	// unfixedGeneric is a generic function, of type source, seen as target, which fixes none of the
	// source's type parameters that its instantiation needs, so one falls back to a constraint holding
	// any (rule G, widened).
	unfixedGeneric bool

	// declaredKey names a declared property an index signature reaches in .a (records_declared_keys.go):
	// the target's index signature can write target under it where source is declared, or, when
	// dictionary is set, the source's index signature holds source where the target reads it as target.
	declaredKey string
	dictionary  bool

	// viewed is the value seen, when it's one branch of a conditional whose own type tsc reduced to
	// the wider side (flag ? options : dict is a Dict), so a refusal can name it.
	viewed *checker.Type
}

// widened finds the first mutable slot at which a value of type from, seen as type to, could be
// written something it can't hold, or returns nil.
func (l *lowering) widened(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	// #bv59a0r: undefined and null enter a union through their own member, not through a
	// writable type parameter left after withoutUndefined removes the undefined member.
	if (from.Flags() == checker.TypeFlagsUndefined || from.Flags() == checker.TypeFlagsNull) && to.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range to.Types() {
			if member.Flags()&from.Flags() != 0 {
				return nil
			}
		}
	}
	// A value held as this union admits only writes every arm accepts. Reads
	// through one arm go through a checked cast or a narrowing, which rule 3350
	// (narrowedArmReadGap in narrowed_nullish.go) guards against calls that undo it.
	// Keep this premise when lifting 3350 to read the declared type (#3twb8ab).
	if l.sameUnionView(from, to) {
		return nil
	}
	from, to = l.withoutUndefined(from), l.withoutUndefined(to)
	if from == to || visited[[2]*checker.Type{from, to}] || from.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	visited[[2]*checker.Type{from, to}] = true
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool { return member.Flags()&checker.TypeFlagsTypeParameter != 0 }) {
		// In a body a type parameter is known only by its constraint, and every value of it is a value of
		// that constraint (docs/rulings.md:123), so it's read at its base constraint, through chains (U extends
		// T, T extends number reads number), and judged by every arm below as a value of that type is. The
		// open enum arms never saw the constraint, so tsc's createBaseToken<Token<TKind>> relating TKind,
		// TKind extends SyntaxKind, to Node's kind was refused as an unproven value (#snypn8c c5, Adamic's
		// ruling). It's never a pass of its own: T extends number into an enum member slot is number there,
		// and refused as number is (docs/rulings.md:247). A slot that is a type parameter, or a union naming
		// one (T | null), is left to the arms that judge it, which relate the source by its own name: Narrow
		// seen as T is a view of T's slot, which reading Narrow at its constraint would no longer name.
		if constraint := l.checker.GetBaseConstraintOfType(from); constraint != nil {
			return l.widened(constraint, to, visited)
		}
	}
	if l.openNumericEnumType(to) {
		for _, member := range l.definedMembers(from) {
			if !l.enumAssignable(member, to) {
				return &widening{source: from, target: to, enum: true}
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsNumberLiteral != 0 && l.openNumericEnumType(from) {
		return &widening{source: from, target: to, enum: true}
	}
	if members := l.definedMembers(from); len(members) > 1 {
		// A union is any one of its members, each seen as the target.
		for _, member := range members {
			if found := l.widened(member, to, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if members := l.definedMembers(to); len(members) > 1 {
		// Seen as a union, a value is seen as each member it can be, and narrowing picks any of them
		// to write through. For a declared key, a value entering a union that names its own type stays
		// that member (#6c4swvt G2, records.go), and seeing the union as another member later is a view
		// of its own, judged there.
		if l.declaredKeys {
			for _, member := range members {
				if member == from {
					return nil
				}
			}
		}
		if from.Flags()&checker.TypeFlagsNumberLike != 0 && l.openEnumSlot(to) {
			// A number's arm is judged against the slot's enum part alone, and a whole numeric enum there
			// holds any number (rulings.md:243), as the open enum arm above says of the enum on its own. tsc
			// lets a number into each of its member literals, so the walk below would pick the first and
			// refuse what the slot takes (#snypn8c c2). An object's arm still walks the object members below.
			return nil
		}
		taken := false
		for _, member := range members {
			if !l.checker.IsTypeAssignableTo(from, member) {
				continue
			}
			taken = true
			if found := l.widened(from, member, visited); found != nil {
				return found
			}
		}
		if !taken && l.structured(from) {
			// A value no one member takes is one tsc let in by decomposing a discriminated union: its
			// whole open enum tag split into the members' literals, each slot then promising one literal
			// the value may not hold (#egmxf20, openEnumTagDecomposed).
			for _, member := range members {
				if member.Flags()&checker.TypeFlagsPrimitive != 0 {
					continue
				}
				for _, viewed := range l.checker.GetPropertiesOfType(member) {
					inside := l.checker.GetPropertyOfType(from, viewed.Name)
					if inside == nil {
						continue
					}
					source, slot := l.checker.GetTypeOfSymbol(inside), l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(viewed))
					if l.openEnumTagDecomposed(source, slot) {
						return &widening{source: source, target: slot, enum: true}
					}
				}
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if l.narrowedFrom(from, to) {
			return nil
		}
		if constraint := l.checker.GetBaseConstraintOfType(to); constraint != nil && l.canWrite(constraint, map[*checker.Type]bool{}) {
			return &widening{source: from, target: to}
		}
		return nil
	}
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		// What it can be is what its constraint allows: seen as anything wider, it's that seen so. A
		// constrained one reaches here only through a slot that is or names a type parameter, which the
		// read above leaves to the arms between.
		constraint := l.checker.GetBaseConstraintOfType(from)
		if constraint == nil {
			return nil
		}
		return l.widened(constraint, to, visited)
	}
	if !l.enumAssignable(from, to) {
		return &widening{source: from, target: to, enum: true}
	}
	if !l.structured(from) || !l.structured(to) {
		// A class instance target is walked as any object is: its fields are slots as a literal's
		// are, so Box<Dog> seen as Box<Animal>, or a plain object seen as a class, is judged by them.
		return nil
	}
	if l.declaredKeys {
		if found := l.declaredKeyWidened(from, to, nil, false); found != nil {
			return found
		}
	}
	if source, target := l.recordElement(from), l.recordElement(to); source != nil && target != nil {
		// Nothing writes through a readonly index, so a record seen as a readonly record may widen what its
		// values are read as (records_readonly.go), as a readonly array may widen its element. The way back
		// is a widening whatever the two index types say: a readonly record may hold a narrower record (tsc's
		// WatchOptions seen as Readonly<OptionsBase>), so seen as a writable record it could write what that
		// record's own reads don't promise, and tsc, unlike for arrays, lets it.
		fromReadonly, toReadonly := l.readonlyIndexType(l.concrete(from)), l.readonlyIndexType(l.concrete(to))
		if fromReadonly && !toReadonly {
			return &widening{source: from, target: to}
		}
		if !toReadonly && (!l.checker.IsTypeAssignableTo(target, source) || l.writeHidesField(target, source)) {
			return &widening{source: source, target: target}
		}
		return l.widened(source, target, visited)
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		// A function seen as another is handed the other's arguments, and its results are seen as
		// the other's: each a view of its own.
		if l.censusNeverRestSignature(toSignatures[0]) {
			if l.censusDiscardedMarkerPredicate(fromSignatures[0], toSignatures[0]) {
				return nil
			}
			source := l.checker.GetReturnTypeOfSignature(fromSignatures[0])
			target := l.checker.GetReturnTypeOfSignature(toSignatures[0])
			if !l.checker.IsTypeAssignableTo(source, target) {
				return &widening{source: source, target: target}
			}
			return l.widened(source, target, visited)
		}
		if len(fromSignatures) == 1 && len(toSignatures) == 1 && len(fromSignatures[0].TypeParameters()) > 0 && !slices.Equal(fromSignatures[0].TypeParameters(), toSignatures[0].TypeParameters()) {
			// A generic function seen as another signature is related as tsc relates it (relater.go,
			// compareSignaturesRelated): the target's own type parameters held opaque (its canonical
			// signature), and the source instantiated in that context by the checker's inference, so
			// @types/node's setTimeout<TArgs> seen as a slot's <A>(callback: (...args: NoInfer<A>) => void,
			// ms, ...args: A) is judged with TArgs as A (rule G, #7fd51gk). The walk below is unchanged.
			// A type parameter the target doesn't fix falls back to its constraint, and one holding any
			// would hand the walk an any it accepts anything into, so that instantiation stops instead.
			target := toSignatures[0]
			if len(target.TypeParameters()) > 0 {
				target = checker.Checker_getCanonicalSignature(l.checker, target)
			}
			instantiated := instantiateSignatureInContextOf(l.checker, fromSignatures[0], target, nil, nil)
			if l.signatureAny(instantiated, target) != nil && l.signatureAny(fromSignatures[0], target) == nil {
				return &widening{source: from, target: to, unfixedGeneric: true}
			}
			fromSignatures, toSignatures = []*checker.Signature{instantiated}, []*checker.Signature{target}
		}
		if found := l.widenedParameters(fromSignatures[0], toSignatures[0], visited); found != nil {
			return found
		}
		return l.widened(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]), visited)
	}
	fromContainers := l.containers(from)
	for _, viewed := range l.containers(to) {
		for _, inside := range fromContainers {
			if !l.sameContainer(inside, viewed) {
				continue
			}
			if found := l.widenedElements(inside, viewed, visited); found != nil {
				return found
			}
			break
		}
	}
	if to.Flags()&checker.TypeFlagsObject != 0 && len(l.containers(to)) > 0 {
		// An array, tuple, map or set: its elements are its slots, and its own properties (length,
		// size) are numbers on both sides.
		return nil
	}
	return l.widenedProperties(from, to, nil, false, visited)
}

// widenedParameters is the half of widened's signature relation that judges what a function is handed:
// each of from's parameters seen as taking to's at the same position. A literal's method judges its
// parameters here and what it returns at its own return statements (literalMethodRefusal).
func (l *lowering) widenedParameters(from *checker.Signature, to *checker.Signature, visited map[[2]*checker.Type]bool) *widening {
	// Each position the viewed signature hands an argument to (anyRelatedPositions), typed as a
	// call there sees it: a rest parameter gives its element at that position, never its whole
	// tuple or array, so a listener `(seen: Error) => void` seen as `(...args: [error: Error]) =>
	// void` is related Error to Error (#6xzs6gj).
	for position := range anyRelatedPositions(from, to) {
		takes, given := l.callableParameterTypeAt(from, position), l.callableParameterTypeAt(to, position)
		if !l.enumAssignable(given, takes) || !l.checker.IsTypeAssignableTo(given, takes) {
			// tsc relates a method's parameters both ways (method bivariance), so a method taking
			// a Dog can be seen as one taking any Animal, and handed a Cat.
			return &widening{source: takes, target: given, parameter: true}
		}
		if found := l.widened(given, takes, visited); found != nil {
			return found
		}
	}
	return nil
}

// widenedProperties walks to's properties against from's, but those skip names: each field a slot
// to can write unless it's readonly, and each method a function view of its own. A copy's fields
// (fresh) are slots of their own, which nothing else writes, so only what they hold is walked.
func (l *lowering) widenedProperties(from *checker.Type, to *checker.Type, skip map[string]bool, fresh bool, visited map[[2]*checker.Type]bool) *widening {
	for _, viewed := range l.checker.GetPropertiesOfType(to) {
		inside := l.checker.GetPropertyOfType(from, viewed.Name)
		if inside == nil || skip[viewed.Name] {
			continue
		}
		source, target := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)
		if viewed.Flags&ast.SymbolFlagsMethod == 0 && inside.Flags&ast.SymbolFlagsMethod != 0 && !l.hostDeclaredSymbol(inside) && !fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source) || l.writeHidesField(target, source)) {
			// A method seen as a writable field is a slot the view writes: tsc admits
			// `const view: { fail: () => void } = { fail(): never { ... } }`, and view.fail = () => {} then
			// puts a returning function where the method's type promises never. A write to the method
			// itself is refused as an unbound method, but the field can be written through the view, so the
			// field's rule applies: the view may write only what the method's type holds (#9dmh4em,
			// Optimizer's probe on #8n7eet3, Adamic's ruling). A readonly field, and a method seen as a
			// method, write nothing here. A host's own method (hostDeclaredSymbol: process's, a Hash's, a
			// RegExp's) is judged by the host rules, which stop a view of it or a write to it in their own
			// words (nodeHostObject, the unstored-member rule, the host binding's relation).
			return &widening{source: source, target: target, methodField: viewed.Name}
		}
		if viewed.Flags&ast.SymbolFlagsMethod != 0 || inside.Flags&ast.SymbolFlagsMethod != 0 {
			// A method isn't a slot anything writes, but what it returns and takes are views: the
			// program's own methods. The library's (an Iterable's, an array's own) are left to the
			// rules for the library's types.
			if len(viewed.Declarations) == 0 || load.IsLibrary(ast.GetSourceFileOfNode(viewed.Declarations[0])) {
				continue
			}
			if l.generic(target) || l.generic(source) {
				// A generic method, on either side, has no one type to compare until its type parameters
				// are related, and tsc relates a method's parameters both ways: a handle<T>(x: Dog) seen as
				// handle<T>(x: Animal) is handed a Cat (#q6gs98r). Only a method type identical to the
				// view's, under the checker's identity relation, is the same method. Any other view of a
				// generic method is related by rule G when each side is one signature, and an overload set
				// stops unjudged rather than passing (#7fd51gk). A host's own generic method,
				// declared only in TypeScript's library or Node's (hostDeclaredSymbol), keeps the path it had
				// before this stop: FSWatcher's on, seen as sys.ts's FsWatchWorkerWatcher's, is judged by its
				// run-time event check and its any overload's relation, and Performance's members by
				// nodeProcessStorageABI's signature scan. A generic the view alone brings (process.chdir
				// seen as chdir<T>(directory: string | number)) isn't the host's, so that view stops too.
				if source == target || checker.Checker_isTypeIdenticalTo(l.checker, source, target) || l.hostDeclaredSymbol(inside) && l.generic(source) {
					continue
				}
				if len(l.checker.GetSignaturesOfType(source, checker.SignatureKindCall)) != 1 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 1 {
					// An overload set has no one signature to relate, so it stays unjudged and stops.
					return &widening{source: source, target: target, genericMethod: viewed.Name}
				}
				// One signature on each side is related as a function value is (widened's rule G), so the
				// parameters a view narrows are refused there, with the method's own types.
				if found := l.widened(source, target, visited); found != nil {
					return found
				}
				continue
			}
			if found := l.widened(source, target, visited); found != nil {
				return found
			}
			continue
		}
		if !fresh && !l.checker.IsReadonlySymbol(viewed) && l.checker.IsReadonlySymbol(inside) {
			// tsc ignores readonly when it relates properties, so a readonly field, which may hold
			// something narrower than its type says (it's covariant), can be seen as a writable one,
			// and the wider type written into it. Only the identical type, which this walk never
			// enters, is the same field.
			return &widening{source: source, target: target, readonlyField: viewed.Name}
		}
		if !fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source) || l.writeHidesField(target, source)) {
			return &widening{source: source, target: target}
		}
		if found := l.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// hostDeclaredSymbol reports whether every declaration of a symbol, through its roots, is in TypeScript's
// bundled library or Node's types: a member the host implements, never a body the program wrote.
func (l *lowering) hostDeclaredSymbol(symbol *ast.Symbol) bool {
	roots := l.checker.GetRootSymbols(symbol)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if !load.IsLibrary(file) && !load.IsNodeLibrary(file) {
				return false
			}
		}
	}
	return true
}

// signatureAny is the first position where a member typed any in from meets one in to that isn't any or
// unknown, through the return (covariantly) and every parameter a caller of to passes (contravariantly),
// as anyRelatedSignatures walks them, or nil.
func (l *lowering) signatureAny(from, to *checker.Signature) *anyRelation {
	seen := map[[2]*checker.Type]bool{}
	if found := l.anyContained(l.checker.GetReturnTypeOfSignature(from), l.checker.GetReturnTypeOfSignature(to), 0, seen); found != nil {
		return found
	}
	for position := range anyRelatedPositions(from, to) {
		if found := l.anyContained(l.checker.GetTypeAtPosition(to, position), l.checker.GetTypeAtPosition(from, position), 0, seen); found != nil {
			return found
		}
	}
	return nil
}

// generic reports whether a function type has a type parameter of its own.
func (l *lowering) generic(proven *checker.Type) bool {
	for _, signature := range l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}

// widenedElements walks an array's, a tuple's, a map's or a set's type arguments in from against
// to's, each a slot to can write unless it's read only.
func (l *lowering) widenedElements(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	mutable := !l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(to) && to.TargetTupleType().IsReadonly())
	return l.widenedArguments(from, to, mutable, visited)
}

// widenedArguments is widenedElements with whether to's own slots can be written given: a fresh
// container's can't be written through anything else (freshValue).
func (l *lowering) widenedArguments(from *checker.Type, to *checker.Type, mutable bool, visited map[[2]*checker.Type]bool) *widening {
	fromArguments, toArguments := l.checker.GetTypeArguments(from), l.checker.GetTypeArguments(to)
	if checker.IsTupleType(from) && !checker.IsTupleType(to) && len(toArguments) == 1 {
		// A tuple seen as an array: every element of it is seen as the array's element.
		widenedTo := make([]*checker.Type, len(fromArguments))
		for index := range widenedTo {
			widenedTo[index] = toArguments[0]
		}
		toArguments = widenedTo
	}
	for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
		source, target := fromArguments[index], toArguments[index]
		if mutable && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source) || l.writeHidesField(target, source)) {
			return &widening{source: source, target: target}
		}
		if found := l.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// writeHidesField says whether a write of a value of type written, into a slot whose readers read it as
// read, can store a value hiding a field those readers expect (#4ysqm2m): tsc lets the write through, since
// written is assignable to read, but written lacks a field read declares optional, or holds it only as a
// library member the runtime's shape lacks. It's the hiding rule's walk (optionalWidened, rule 3300's) run in
// the write direction, which no read site judges: the slot's own readers read what was stored as read, so a
// Target[] seen as Source[] takes a Source whose hidden label a read through Target then takes as a string.
// Each road the invariant-mutable rule judges a write by assignability asks it too: a mutable container's
// element (widenedArguments), a writable field (widenedProperties), a method seen as a writable field, a
// writable record's value, and a literal setter's parameter. It judges the type, never the value that flows,
// so an honest Source is refused where a lying one would be, as the declared-keys rule judges (#npqr3sx).
func (l *lowering) writeHidesField(written, read *checker.Type) bool {
	return l.optionalWidened(written, read, nil, map[[2]*checker.Type]bool{}) != nil
}

// containers is a type's arrays, tuples, maps and sets: itself, or its members that are, when it's an
// intersection.
func (l *lowering) containers(proven *checker.Type) []*checker.Type {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsIntersection != 0 {
		members = proven.Types()
	}
	var found []*checker.Type
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject != 0 && member.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
			(l.isArrayType(member) || checker.IsTupleType(member) || l.isLibraryType(member, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
			found = append(found, member)
		}
	}
	return found
}

// sameContainer reports whether two containers are the same kind, whose type arguments pair up: two
// arrays or tuples (readonly or not), two maps, or two sets.
func (l *lowering) sameContainer(from *checker.Type, to *checker.Type) bool {
	switch {
	case l.isArrayType(to) || checker.IsTupleType(to):
		return l.isArrayType(from) || checker.IsTupleType(from)
	case l.isLibraryType(to, "Map", "ReadonlyMap"):
		return l.isLibraryType(from, "Map", "ReadonlyMap")
	default:
		return l.isLibraryType(from, "Set", "ReadonlySet")
	}
}

// canWrite reports whether something can be written through a value of a type, at any depth: an
// array, tuple, map or set that isn't read only, or a property that isn't readonly, or one inside a
// readonly one.
func (l *lowering) canWrite(proven *checker.Type, visited map[*checker.Type]bool) bool {
	proven = l.withoutUndefined(proven)
	if proven == nil || visited[proven] || !l.structured(proven) || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 {
		return false
	}
	visited[proven] = true
	if l.recordElement(proven) != nil {
		return true
	}
	containers := l.containers(proven)
	for _, container := range containers {
		if !l.isLibraryType(container, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(container) && container.TargetTupleType().IsReadonly()) {
			return true
		}
		for _, element := range l.checker.GetTypeArguments(container) {
			if l.canWrite(element, visited) {
				return true
			}
		}
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 && len(containers) > 0 {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if !l.checker.IsReadonlySymbol(property) || l.canWrite(l.checker.GetTypeOfSymbol(property), visited) {
			return true
		}
	}
	return false
}

// narrowedFrom reports whether from is the type parameter to itself, or it narrowed: an intersection
// with it among its members (T & {}, what a check against undefined leaves).
func (l *lowering) narrowedFrom(from *checker.Type, to *checker.Type) bool {
	if from == to {
		return true
	}
	if from.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range from.Types() {
			if member == to {
				return true
			}
		}
	}
	return false
}

// structured reports whether a type has parts to walk: an object or an intersection.
func (l *lowering) structured(proven *checker.Type) bool {
	return proven.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0
}

// sameUnionView compares checker identity after instantiation and removal of
// added nullish members. The checker retains named union origins when filtering.
func (l *lowering) sameUnionView(from, to *checker.Type) bool {
	from, to = l.concrete(from), l.concrete(to)
	if from.Flags()&checker.TypeFlagsUnion == 0 || to.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	sourceNullish, viewNullish := unionNullishFlags(from), unionNullishFlags(to)
	if sourceNullish&^viewNullish != 0 {
		return false
	}
	difference := sourceNullish ^ viewNullish
	if difference == 0 {
		return from == to
	}
	// The subset check makes this filter a no-op on the source.
	keep := func(member *checker.Type) bool { return member.Flags()&difference == 0 }
	return filterUnionView(l.checker, from, keep) == filterUnionView(l.checker, to, keep)
}

func unionNullishFlags(union *checker.Type) checker.TypeFlags {
	var flags checker.TypeFlags
	for _, member := range union.Types() {
		flags |= member.Flags() & (checker.TypeFlagsUndefined | checker.TypeFlagsNull)
	}
	return flags
}

//go:linkname filterUnionView github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).filterType
func filterUnionView(receiver *checker.Checker, union *checker.Type, keep func(*checker.Type) bool) *checker.Type

// withoutUndefined is a type with undefined taken out of it: the one type left, or the union itself
// when several are, whose members widened walks one by one (definedMembers).
func (l *lowering) withoutUndefined(proven *checker.Type) *checker.Type {
	if members := l.definedMembers(proven); len(members) == 1 {
		return members[0]
	}
	return proven
}

// definedMembers is a union's members but undefined, or the type itself when it isn't a union.
func (l *lowering) definedMembers(proven *checker.Type) []*checker.Type {
	if proven == nil {
		return nil
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return []*checker.Type{proven}
	}
	var members []*checker.Type
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined == 0 {
			members = append(members, member)
		}
	}
	return members
}

// isClassInstance is the instance side of a class, generic or not.
func isClassInstance(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if proven.ObjectFlags()&checker.ObjectFlagsClass != 0 {
		return true
	}
	return proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil && proven.Target().ObjectFlags()&checker.ObjectFlagsClass != 0
}

// viewSite reports whether an expression is a value going into a typed slot: an initializer (a
// variable's, a parameter's default, a class field's), the right of an assignment (=, ??=, ||= or &&=),
// an argument, a returned value, a literal's property (shorthand too) or element, or an arrow's
// expression body.
// Parentheses around it are the same site.
func viewSite(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		// A destructuring declaration without a type keeps nothing of the value: its names take the
		// types of its parts, and the pattern's own type (its fields any) is no view of it.
		declaration := parent.AsVariableDeclaration()
		if ast.IsBindingPattern(declaration.Name()) && declaration.Type == nil {
			return false
		}
		return declaration.Initializer == node
	case ast.KindBinaryExpression:
		// ??=, ||= and &&= store their right into the left's slot as = does, so the right of each is a
		// view site too. Taking only = let a Dog[] into an Animal[] by ??= and a readonly field become
		// writable by ??=, both lowered wrong (#6jq01hc).
		binary := parent.AsBinaryExpression()
		return (binary.OperatorToken.Kind == ast.KindEqualsToken || logicalAssignment(binary.OperatorToken.Kind)) && binary.Right == node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression != node
	case ast.KindReturnStatement, ast.KindArrayLiteralExpression:
		return true
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindArrowFunction:
		return parent.AsArrowFunction().Body == node
	}
	return false
}

// refuseWidening refuses a value seen through a type that can write what it can't hold.
func (l *lowering) refuseWidening(node *ast.Node) error {
	if l.nodeFSFileReadOnlyArgument(node) || l.regexCallbackIntrinsicView(node) {
		return nil
	}
	if literalMember(node) && !judgedLiteralMembers[node.Kind] {
		return l.unjudgedLiteralMember(node)
	}
	var own, contextual *checker.Type
	var found *widening
	switch {
	case node.Kind == ast.KindAsExpression:
		as := node.AsAsExpression()
		if as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
			return nil
		}
		source, target := l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node)
		if l.checker.IsTypeAssignableTo(source, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target) {
			if err := l.provenRelation(node, as.Expression, target); err != nil {
				return err
			}
		} else {
			// Preserve the existing checks on downcasts; their lowering belongs to cast.go.
			own, contextual = source, target
			if found = l.freshOrWidened(as.Expression, own, contextual); found != nil {
				break
			}
		}
		own, contextual, found = l.castSlotWidened(node)
	case node.Kind == ast.KindSatisfiesExpression:
		satisfies := node.AsSatisfiesExpression()
		if err := l.provenRelation(node, satisfies.Expression, l.checker.GetTypeAtLocation(satisfies.Type)); err != nil {
			return err
		}
		own, contextual, found = l.castSlotWidened(node)
	case node.Kind == ast.KindShorthandPropertyAssignment:
		// { pets } is { pets: pets }: the variable seen as the literal's property, judged exactly as the
		// named form's initializer is, against the slot its name has (slotType). Reading the property off
		// the literal's slot instead missed it whenever that slot was a union holding undefined, such as a
		// return type `{ p: W } | undefined`: undefined has no p, so the union had none, and a readonly R
		// went out as a writable W unjudged (#20nf0sm). The name's slot is the checker's contextual type
		// for the property, or carriedSlot's, both of which look past undefined. Where the checker names no
		// slot for the name but the literal has one, the property of each member the literal can sit in is
		// judged instead, so a slot the checker can't name never lets the value through unjudged
		// (literalMemberSlots). A shorthand's key is its name, which is always read.
		slots, key, _ := l.literalMemberSlots(node)
		own = l.checker.GetTypeAtLocation(node.Name())
		contextual, found = l.literalMemberWidened(node, node.Name(), own, slots, key)
	case node.Kind == ast.KindGetAccessor && literalMember(node):
		// { get p(): R { return r; } } is r seen as the literal's property p, judged exactly as { p: r } is:
		// each value the getter returns against each slot of p. Nothing judged it before #wjqtypf, since a
		// getter isn't an expression, so a readonly R went out as a writable W. Each return statement's
		// value is judged at its own site, not the getter's declared type, so a literal made right there
		// keeps its freshness (freshOrWidened): `get p(): { readonly t: number } { return { t: 1 }; }` is a
		// new object at every read, which no other name holds. A getter judges what the property reads as; a
		// setter beside it is judged in its own case, by what it takes.
		slots, key, readable := l.literalMemberSlots(node)
		if !readable {
			return l.unreadableLiteralMember(node, "a getter")
		}
		for _, returned := range returnedExpressions(node) {
			own = l.checker.GetTypeAtLocation(returned)
			if contextual, found = l.literalMemberWidened(node, returned, own, slots, key); found != nil {
				break
			}
		}
	case node.Kind == ast.KindSetAccessor && literalMember(node):
		return l.literalSetterRefusal(node)
	case node.Kind == ast.KindMethodDeclaration && literalMember(node):
		// { get(): R { return r; } } is a function seen as the literal's property, judged in
		// literalMethodRefusal (#wjqtypf).
		slots, _, readable := l.literalMemberSlots(node)
		if !readable {
			return l.unreadableLiteralMember(node, "a method")
		}
		return l.literalMethodRefusal(node, slots)
	case node.Kind == ast.KindSpreadAssignment:
		// { ...kennel } copies kennel's fields, not what they hold: each field not written again after
		// it is kennel's value seen as the literal's field. It's judged against each member of the
		// literal's slot the literal can sit in, since a union's own properties are only those every
		// member has, and `{ p: W } | undefined` has none (#20nf0sm, Lower's neighbor).
		literal := l.slotType(node.Parent)
		if literal == nil {
			return nil
		}
		skip := map[string]bool{}
		after := false
		for _, property := range node.Parent.AsObjectLiteralExpression().Properties.Nodes {
			if property == node {
				after = true
			} else if after && property.Name() != nil {
				skip[property.Name().Text()] = true
			}
		}
		own = l.checker.GetTypeAtLocation(node.AsSpreadAssignment().Expression)
		for _, member := range l.literalSlotMembers(node.Parent, literal) {
			contextual = member
			if l.declaredKeys {
				// A copy writes each of the spread's entries under its key, so a dictionary's can land
				// under the literal's declared properties; its own declared properties are only copied.
				found = l.declaredKeyWidened(own, member, skip, true)
			}
			if found == nil {
				found = l.widenedProperties(own, member, skip, true, map[[2]*checker.Type]bool{})
			}
			if found != nil {
				break
			}
		}
	default:
		if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
			return nil
		}
		if l.genericFunction(node) {
			// same<Item> written where a (value: number) => number goes is instantiated to exactly
			// that type, Item = number, so there's no wider view of it to judge.
			return nil
		}
		if pattern := destructuringTarget(node); pattern != nil {
			// [a, b] = tuple and { a, b } = value keep nothing of the pattern: each element or field
			// is read out and stored into its target, so each is the view, seen as its target's type,
			// and the pattern isn't one.
			if pattern.Kind == ast.KindObjectLiteralExpression {
				return l.refuseFieldWidening(node, pattern)
			}
			return l.refuseElementWidening(node, pattern)
		}
		if viewSite(node) {
			contextual = l.slotType(node)
		}
		read := false
		if contextual == nil && !l.conditionalBranchInSlot(node) {
			// An operand of ||, ?? or ?: is judged where its expression lands (operandLanding): against a
			// declared slot's type, or as a read where nothing holds it.
			contextual, read = l.operandLanding(node)
			if contextual == nil {
				// With no type written for it, a value can still be taken into a wider one tsc made: the
				// union of a conditional's branches reduced to the wider (flag ? dogs : animals is an
				// Animal[]), a literal's elements likewise, a function's returns likewise.
				contextual = l.impliedTarget(node)
			}
		}
		if contextual == nil {
			return nil
		}
		own = l.checker.GetTypeAtLocation(node)
		if andLeftOperand(node) {
			// Only the left's falsy arms reach an &&'s result: one that can only be an object never does.
			if own = l.falsyArms(own); own == nil {
				return nil
			}
		}
		if read {
			found = l.readWidened(node, own, contextual)
		} else {
			found = l.freshOrWidened(node, own, contextual)
		}
		if found == nil && node.Parent != nil && node.Parent.Kind == ast.KindPropertyAssignment && node.Parent.AsPropertyAssignment().Initializer == node {
			if name := node.Parent.Name(); name != nil && (ast.IsIdentifier(name) || ast.IsStringLiteral(name) || ast.IsNumericLiteral(name)) {
				if literal := l.slotType(node.Parent.Parent); literal != nil {
					contextual, found = l.memberLiteralSlotWidened(node, node.Parent.Parent, own, literal, name.Text(), contextual)
				}
			}
		}
	}
	if found == nil {
		return nil
	}
	return l.wideningRefusal(node, own, contextual, found)
}

// literalSlotMembers is each member of an object literal's slot the literal is judged against: the slot
// itself when it isn't a union, and otherwise every member that isn't a primitive (undefined and null
// among them) and that the literal's own type is assignable to, since a member it isn't assignable to
// never holds it. That filter only narrows which members are judged, and it's no relation of its own:
// tsc lets a literal into a discriminated union it fits as a whole but fits no one member of, by
// splitting its discriminant (`{ ...h, k }` with k: 'x' | 'y', into `{ k: 'x'; p: W } | { k: 'y'; p: W }`),
// so when no member passes the filter, every non-primitive member is judged, and the list is never empty
// while the slot has an object member (#20nf0sm, Compiler's road on d2a29ebf).
func (l *lowering) literalSlotMembers(literal *ast.Node, slot *checker.Type) []*checker.Type {
	if slot.Flags()&checker.TypeFlagsUnion == 0 {
		return []*checker.Type{slot}
	}
	own := l.checker.GetTypeAtLocation(literal)
	var members, objects []*checker.Type
	for _, member := range slot.Types() {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		objects = append(objects, member)
		if l.checker.IsTypeAssignableTo(own, member) {
			members = append(members, member)
		}
	}
	if len(members) == 0 {
		return objects
	}
	return members
}

// judgedLiteralMembers are the kinds of an object literal's member refuseWidening judges against the slot
// its property goes into: a property, by its initializer; a shorthand, a spread, a getter, a setter and a
// method, each in its own case. A member of any other kind, one the language adds later, is the value of a
// property nothing judges, so it falls to unjudgedLiteralMember and fails closed (#wjqtypf): a list of the
// kinds that are judged, never of the ones that aren't, so a kind nobody listed can't lower unjudged.
var judgedLiteralMembers = map[ast.Kind]bool{
	ast.KindPropertyAssignment:          true,
	ast.KindShorthandPropertyAssignment: true,
	ast.KindSpreadAssignment:            true,
	ast.KindGetAccessor:                 true,
	ast.KindSetAccessor:                 true,
	ast.KindMethodDeclaration:           true,
}

// literalMember is whether node is a member of an object literal.
func literalMember(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression
}

// unjudgedLiteralMember is not yet for a member of an object literal whose kind refuseWidening doesn't judge
// (judgedLiteralMembers), wherever a slot names its property: its value would go into that slot unjudged.
// Where no slot names it, nothing sees it as anything, and it's let through, as is one whose key can't be
// read into a slot with an index signature (literalSlotIndexed).
func (l *lowering) unjudgedLiteralMember(member *ast.Node) error {
	slots, _, readable := l.literalMemberSlots(member)
	if readable && len(slots) == 0 || !readable && l.literalSlotIndexed(member) {
		return nil
	}
	return l.notYet(member, describe(member)+" in an object literal written where a slot names its property, which stage 0 judges for a property, a shorthand, a spread, a getter, a setter and a method only")
}

// literalSetterRefusal judges a setter in an object literal against each slot of its property. A writer
// through the slot can set anything the slot's type holds, so the setter must take all of it, and what it
// takes must not be seen as a type that can write what it can't hold. tsc relates only what the property
// reads as, the getter's type, so `{ get p(): Dog, set p(v: Dog) }` was accepted as `{ p: Animal }` and
// handed a Cat; a getter beside the setter judges the reads, never this road (#wjqtypf, Compiler's must on
// 0ee9e50b). A setter with no getter beside it is not yet wherever a slot names its property, since the
// property then reads as undefined where the slot says it holds the slot's type.
func (l *lowering) literalSetterRefusal(member *ast.Node) error {
	slots, key, readable := l.literalMemberSlots(member)
	if !readable {
		return l.unreadableLiteralMember(member, "a setter")
	}
	if len(slots) == 0 {
		return nil
	}
	parameters := l.checker.GetSignatureFromDeclaration(member).Parameters()
	if len(parameters) != 1 {
		return l.notYet(member, "a setter in an object literal without exactly one parameter")
	}
	takes := l.checker.GetTypeOfSymbol(parameters[0])
	for _, slot := range slots {
		if !l.enumAssignable(slot, takes) || !l.checker.IsTypeAssignableTo(slot, takes) || l.writeHidesField(slot, takes) {
			return &Refused{Where: l.program.Where(member), What: "a setter of " + key + " taking " + l.checker.TypeToString(takes) + " seen as the property " + key + ": " + l.checker.TypeToString(slot) + ", so a write through it can hand the setter what it can't take", Fix: "take " + l.checker.TypeToString(slot) + " in the setter, or make " + key + " readonly in the type it's seen as, which can't write (adamic/invariant-mutable)"}
		}
		if found := l.widened(slot, takes, map[[2]*checker.Type]bool{}); found != nil {
			return l.wideningRefusal(member, slot, takes, found)
		}
	}
	for _, sibling := range member.Parent.AsObjectLiteralExpression().Properties.Nodes {
		if sibling.Kind != ast.KindGetAccessor {
			continue
		}
		if siblingKey, siblingReadable := l.literalMemberKey(sibling); siblingReadable && siblingKey == key {
			return nil
		}
	}
	return l.notYet(member, "a setter with no getter in an object literal written where a slot names its property, which reads as undefined where the slot says it holds "+l.checker.TypeToString(slots[0]))
}

// literalMethodRefusal judges a method in an object literal against each slot of its property, as a
// function seen as the slot's: its parameters both ways, as widened's signatures branch relates them
// (widenedParameters), and each value it returns against the slot's result at its own return statement,
// not the method's declared result, so a literal made right there keeps its freshness (freshOrWidened):
// `next() { return { value: 1, done: false }; }` seen as returning `{ value: number; done: boolean }` is
// a new object every call, which nothing else holds. A method with type parameters of its own gives up
// that freedom: its return statements are typed by them, which widened passes when unconstrained, so
// `{ get<T>(x: T): T { return x; } }` seen as `{ get(x: R): W }` would let R out as W. Its declared result,
// as the slot instantiates it (T := R), is judged instead, which fails closed (Lower's must on 29bc3e33).
// widened relates a function to one non-generic signature only, so a slot with type parameters
// (PromiseLike's then<TResult1, TResult2>) or overloads is not yet, never refused by a relation that
// doesn't hold, for any method. That not yet is l.unlowerable, returned only once every module has lowered
// (lower.go), so any other stop is the one a program hears: a literal thenable's own stops ("return of
// thenables") are raised in moduleStatements, after refuse, where refusal.deferred, refuse's lowest stop,
// would still have come first. It fails closed, since the program stops on it when nothing else does. An
// async or generator method's return statements don't give what it returns, a promise or an iterator does,
// so it's judged whole by widened. A slot with no call signature calls nothing through it, and judges
// nothing here (#wjqtypf).
func (l *lowering) literalMethodRefusal(method *ast.Node, slots []*checker.Type) error {
	own := l.checker.GetTypeAtLocation(method)
	whole := generatorDeclaration(method) || ast.HasSyntacticModifier(method, ast.ModifierFlagsAsync)
	signature := l.checker.GetSignatureFromDeclaration(method)
	for _, slot := range slots {
		for _, member := range l.methodSlotMembers(own, slot) {
			slotSignatures := l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)
			if len(slotSignatures) == 0 {
				continue
			}
			if len(slotSignatures) > 1 {
				if l.unlowerable == nil {
					l.unlowerable = l.notYet(method, "a method in an object literal seen as an overloaded method, which stage 0 doesn't judge")
				}
				return nil
			}
			if len(slotSignatures[0].TypeParameters()) > 0 {
				if l.unlowerable == nil {
					l.unlowerable = l.notYet(method, "a method in an object literal seen as a generic method, which stage 0 doesn't judge")
				}
				return nil
			}
			if whole {
				if found := l.widened(own, member, map[[2]*checker.Type]bool{}); found != nil {
					return l.wideningRefusal(method, own, member, found)
				}
				continue
			}
			if l.censusDiscardedMarkerPredicate(signature, slotSignatures[0]) {
				continue
			}
			judged := signature
			if len(judged.TypeParameters()) > 0 {
				judged = instantiateSignatureInContextOf(l.checker, judged, slotSignatures[0], nil, nil)
			}
			result := l.checker.GetReturnTypeOfSignature(slotSignatures[0])
			if l.censusNeverRestSignature(slotSignatures[0]) {
				// Handed nothing it can take, so only what it returns is a view (widened's census branch).
				if source := l.checker.GetReturnTypeOfSignature(judged); !l.checker.IsTypeAssignableTo(source, result) {
					return l.wideningRefusal(method, own, member, &widening{source: source, target: result})
				}
			} else if found := l.widenedParameters(judged, slotSignatures[0], map[[2]*checker.Type]bool{}); found != nil {
				return l.wideningRefusal(method, own, member, found)
			}
			if len(signature.TypeParameters()) > 0 {
				// Its own type parameters type its return statements, and widened passes an unconstrained
				// one, so its result as the slot instantiates it is judged instead.
				if found := l.widened(l.checker.GetReturnTypeOfSignature(judged), result, map[[2]*checker.Type]bool{}); found != nil {
					return l.wideningRefusal(method, own, member, found)
				}
				continue
			}
			for _, returned := range returnedExpressions(method) {
				if found := l.freshOrWidened(returned, l.checker.GetTypeAtLocation(returned), result); found != nil {
					return l.wideningRefusal(method, own, member, found)
				}
			}
		}
	}
	return nil
}

// methodSlotMembers is each member of a method's slot it's judged against: the slot itself when it isn't a
// union of several, and otherwise every member the method's type is assignable to, as widened sees a value
// seen as a union. That filter only narrows which members are judged, so when none passes, every member
// with a call signature is judged, and the list is never empty while one has one (literalSlotMembers'
// fix, #20nf0sm, Lower's must on b3e90772).
func (l *lowering) methodSlotMembers(own *checker.Type, slot *checker.Type) []*checker.Type {
	members := l.definedMembers(slot)
	if len(members) <= 1 {
		return members
	}
	var assignable, callable []*checker.Type
	for _, member := range members {
		if len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) > 0 {
			callable = append(callable, member)
		}
		if l.checker.IsTypeAssignableTo(own, member) {
			assignable = append(assignable, member)
		}
	}
	if len(assignable) == 0 {
		return callable
	}
	return assignable
}

// returnedExpressions is each value a function's own return statements give back, and none of a function
// or class nested in it, whose returns are its own.
func returnedExpressions(function *ast.Node) []*ast.Node {
	var returned []*ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) || ast.IsClassLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			if expression := node.AsReturnStatement().Expression; expression != nil {
				returned = append(returned, expression)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	if body := function.Body(); body != nil {
		body.ForEachChild(visit)
	}
	return returned
}

// siteResults is each value a function gives back: an arrow's expression body, or each value its own
// return statements give (returnedExpressions).
func siteResults(function *ast.Node) []*ast.Node {
	if body := function.Body(); function.Kind == ast.KindArrowFunction && body != nil && body.Kind != ast.KindBlock {
		return []*ast.Node{body}
	}
	return returnedExpressions(function)
}

// freshFunctionWidened judges a function value written at a view site, when it's fresh for what each call
// returns (freshFunction), against each member of its slot as literalMethodRefusal judges a literal's
// method: its parameters both ways (widenedParameters), and each value it returns as made at the site
// against the slot's result (madeAtSiteWidened), never by the function's inferred result, whose own slots
// no other name sees. A slot member that's overloaded, generic or a census never-rest signature isn't
// related here, so judged is false and the whole value takes widened's path, as a function that isn't fresh
// does. A member with no call signature calls nothing through it, but can still see the function's own
// properties (a Function's readonly name, seen as a writable one), so widened walks it as it did before
// (#k8ed1f2, Adamic's ruling).
func (l *lowering) freshFunctionWidened(function *ast.Node, own *checker.Type, contextual *checker.Type) (*widening, bool) {
	if !l.freshFunction(function) {
		return nil, false
	}
	signature := l.checker.GetSignatureFromDeclaration(function)
	visited := map[[2]*checker.Type]bool{}
	for _, member := range l.methodSlotMembers(own, contextual) {
		slotSignatures := l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)
		if len(slotSignatures) == 0 {
			if found := l.widened(own, member, visited); found != nil {
				return found, true
			}
			continue
		}
		if len(slotSignatures) > 1 || len(slotSignatures[0].TypeParameters()) > 0 || l.censusNeverRestSignature(slotSignatures[0]) {
			return nil, false
		}
		if found := l.widenedParameters(signature, slotSignatures[0], visited); found != nil {
			return found, true
		}
		result := l.checker.GetReturnTypeOfSignature(slotSignatures[0])
		for _, returned := range siteResults(function) {
			if found := l.madeAtSiteWidened(returned, result, visited); found != nil {
				return found, true
			}
		}
	}
	return nil, true
}

// freshFunction reports whether a function written at a view site, an arrow or a function expression, is
// fresh for what each call returns, #wjqtypf's per-call freshness of a literal's own method carried to a
// function value. Each of its three conditions is read in the function's own body, never across a call
// (#k8ed1f2, Adamic's ruling):
//   - Every value it returns is made right there (madeAtSite). A return with no value, or falling off the
//     end, gives undefined, which holds nothing.
//   - It stores that value nowhere. A value built in the return expression itself can't be named by any
//     statement before the return, and nothing after it, a finally, can name it either. The one way it can
//     still get out is through its own members: a method, an accessor or a function expression in the
//     returned literal is handed the literal as this when it's called, and can push it anywhere, under a
//     type of its own (lendsItself).
//   - The function has no handle but the slot: it's the expression judged at the site, so it's new and only
//     the slot holds it, and a function expression doesn't mention its own name, by which it could call or
//     store itself (namesItself). An async function or a generator gives a promise or an iterator, not what
//     its return statements give, so it's judged whole by its type, as is one with type parameters of its
//     own, whose returns its own parameters type.
func (l *lowering) freshFunction(function *ast.Node) bool {
	if function.Kind != ast.KindArrowFunction && function.Kind != ast.KindFunctionExpression {
		return false
	}
	if generatorDeclaration(function) || ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) || len(l.checker.GetSignatureFromDeclaration(function).TypeParameters()) > 0 {
		return false
	}
	if function.Kind == ast.KindFunctionExpression && namesItself(function) {
		return false
	}
	for _, returned := range siteResults(function) {
		if !l.madeAtSite(returned) {
			return false
		}
		if lendsItself(returned) {
			return false
		}
	}
	return true
}

// madeAtSite reports whether a value a function returns is made right where it's returned, so no statement
// before the return can name it: an object or an array literal, a new Map, Set or Array or a native copy
// (freshValue), or a conditional whose branches both are, through parentheses. A call, a name, a parameter
// and a property read are not, nor are ||, ??, && and =, since each can give back what something else holds
// (#k8ed1f2, Adamic's ruling).
func (l *lowering) madeAtSite(value *ast.Node) bool {
	value = ast.SkipParentheses(value)
	switch value.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		return true
	case ast.KindConditionalExpression:
		conditional := value.AsConditionalExpression()
		return l.madeAtSite(conditional.WhenTrue) && l.madeAtSite(conditional.WhenFalse)
	case ast.KindNewExpression, ast.KindCallExpression:
		return l.freshValue(value)
	}
	return false
}

// lendsItself reports whether an object literal a function returns, at any branch of a conditional, can
// hand itself to something else when it's called: it has a method, a getter or a setter, or a this inside
// a function in it that isn't an arrow, any of which is given the literal as this and can store it where
// the slot never sees (#k8ed1f2, Adamic's ruling).
func lendsItself(value *ast.Node) bool {
	value = ast.SkipParentheses(value)
	switch value.Kind {
	case ast.KindConditionalExpression:
		conditional := value.AsConditionalExpression()
		return lendsItself(conditional.WhenTrue) || lendsItself(conditional.WhenFalse)
	case ast.KindObjectLiteralExpression:
		for _, member := range value.AsObjectLiteralExpression().Properties.Nodes {
			switch member.Kind {
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				return true
			}
		}
		return receivesThis(value, false)
	}
	return false
}

// freshObjectLiteral reports whether the literal's own slots can be built solely as its contextual
// type. A member that isn't an arrow (a method, a getter or a setter, or a property holding a function
// expression) is given the object as this, and can store it or read it at that type while a wider view
// writes its fields. A member whose this is the literal's own type keeps it fresh: the literal's
// contextual type, or its own type with or without its freshness. A narrower this, declared as a
// parameter (#69vkn0j) or given by a contextual ThisType (#n7mgy3s), doesn't, and nor does a this the
// checker types any other way, since it can't be shown to be the literal's own.
func (l *lowering) freshObjectLiteral(literal *ast.Node) bool {
	for _, member := range literal.AsObjectLiteralExpression().Properties.Nodes {
		switch member.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		case ast.KindPropertyAssignment:
			member = ast.SkipParentheses(member.AsPropertyAssignment().Initializer)
			if member.Kind != ast.KindFunctionExpression {
				continue
			}
		default:
			continue
		}
		if ast.GetThisParameter(member) != nil {
			return false
		}
		if keyword := boundThis(member); keyword != nil && !l.literalOwnThis(literal, l.checker.GetTypeAtLocation(keyword)) {
			return false
		}
	}
	return true
}

// literalOwnThis reports whether given, the type a member of literal is given as this, is the literal's own:
// the literal's slot (slotType), or the literal's own type, fresh or regular. Only
// identity counts, so a ThisType's narrower view, a union, or a type tsc made anew is never the literal's.
func (l *lowering) literalOwnThis(literal *ast.Node, given *checker.Type) bool {
	if given == l.slotType(literal) {
		return true
	}
	own := l.checker.GetTypeAtLocation(literal)
	return given == own || given == regularTypeOfObjectLiteral(l.checker, own)
}

// boundThis is the first this keyword in a literal's member that the member itself binds, or nil: not one
// in its name, nor one inside a function under it that isn't an arrow, or a class, either of which binds
// its own (receivesThis).
func boundThis(member *ast.Node) *ast.Node {
	var found *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindThisKeyword {
			found = node
			return true
		}
		if (ast.IsFunctionLike(node) && node.Kind != ast.KindArrowFunction) || ast.IsClassLike(node) {
			return false
		}
		return node.ForEachChild(visit)
	}
	name := member.Name()
	member.ForEachChild(func(child *ast.Node) bool {
		return child != name && visit(child)
	})
	return found
}

// receivesThis reports whether a this keyword under node sits inside a function that isn't an arrow, or a
// class, either of which binds this to what it's called on. bound is whether node is already inside one.
func receivesThis(node *ast.Node, bound bool) bool {
	if node.Kind == ast.KindThisKeyword && bound {
		return true
	}
	if (ast.IsFunctionLike(node) && node.Kind != ast.KindArrowFunction) || ast.IsClassLike(node) {
		bound = true
	}
	return node.ForEachChild(func(child *ast.Node) bool { return receivesThis(child, bound) })
}

// namesItself reports whether a function expression with a name of its own mentions that name in its
// parameters or its body, by which it could call itself or store itself. A name that only shadows it
// counts too, which can only keep a function from being fresh (#k8ed1f2, Adamic's ruling).
func namesItself(function *ast.Node) bool {
	name := function.Name()
	if name == nil {
		return false
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node == name {
			return false
		}
		if ast.IsIdentifier(node) && node.Text() == name.Text() {
			return true
		}
		return node.ForEachChild(visit)
	}
	return function.ForEachChild(visit)
}

// unreadableLiteralMember is a getter's, a setter's or a method's stop when literalMemberKey can't read its
// key ([key]() {}). Into a slot with an index signature it's a keyed entry of the literal, the write its
// direct twin `handlers[key] = value` makes, which the registry's dictionary and declared-key rules judge,
// so it's let through to them rather than stopped first by a rule that can't read it (#wjqtypf, w10).
// Anywhere else it's not yet.
func (l *lowering) unreadableLiteralMember(member *ast.Node, what string) error {
	if l.literalSlotIndexed(member) {
		return nil
	}
	return l.notYet(member, what+" in an object literal whose key can't be read off the literal's slot, where it would be judged against its property")
}

// literalSlotIndexed is whether a member of an object literal goes into a slot with a string or a number
// index signature, in any member of it the literal can sit in (literalSlotMembers).
func (l *lowering) literalSlotIndexed(member *ast.Node) bool {
	literal := l.slotType(member.Parent)
	if literal == nil {
		return false
	}
	for _, slotMember := range l.literalSlotMembers(member.Parent, literal) {
		if l.checker.GetIndexTypeOfType(slotMember, l.checker.GetStringType()) != nil || l.checker.GetIndexTypeOfType(slotMember, l.checker.GetNumberType()) != nil {
			return true
		}
	}
	return false
}

// literalMemberSlots is each slot a member of an object literal goes into, with the key it's read under
// (literalMemberKey): its name's own slot (slotType), or, where the checker names none for the name but the
// literal has a slot, the key's property in each member of that slot the literal can sit in
// (literalSlotMembers), so a slot the checker can't name never lets the member through unjudged (#20nf0sm). A
// getter's or a method's name is a declaration's, never an expression the checker gives a slot, so theirs
// are always read off the literal's (#wjqtypf). readable is false when the literal has a slot but the key
// can't be read, and the caller fails closed.
func (l *lowering) literalMemberSlots(member *ast.Node) (slots []*checker.Type, key string, readable bool) {
	key, readable = l.literalMemberKey(member)
	if name := member.Name(); name != nil && ast.IsIdentifier(name) {
		if slot := l.slotType(name); slot != nil {
			return []*checker.Type{slot}, key, true
		}
	}
	literal := l.slotType(member.Parent)
	if literal == nil {
		return nil, key, true
	}
	if !readable {
		return nil, key, false
	}
	for _, slotMember := range l.literalSlotMembers(member.Parent, literal) {
		if property := l.checker.GetPropertyOfType(slotMember, key); property != nil {
			slots = append(slots, l.checker.GetTypeOfSymbol(property))
		}
	}
	return slots, key, true
}

// literalMemberKey is the key a member of an object literal declares its property under: its name's text, or
// for a computed name ([Symbol.iterator]), the name of the literal's own property it declares, the key tsc
// binds it under. ok is false when it's neither.
func (l *lowering) literalMemberKey(member *ast.Node) (key string, ok bool) {
	name := member.Name()
	if name == nil {
		return "", false
	}
	if ast.IsIdentifier(name) || ast.IsStringLiteral(name) || ast.IsNumericLiteral(name) {
		return name.Text(), true
	}
	for _, property := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(member.Parent)) {
		if slices.Contains(property.Declarations, member) {
			return property.Name, true
		}
	}
	return "", false
}

// literalMemberWidened judges own, the value of a literal's member, against each slot of its property
// (literalMemberSlots), as a property's initializer is judged against its own, and then, for a union slot,
// against each member's literal slot (memberLiteralSlotWidened). value is the node the value is judged at.
func (l *lowering) literalMemberWidened(member *ast.Node, value *ast.Node, own *checker.Type, slots []*checker.Type, key string) (*checker.Type, *widening) {
	var contextual *checker.Type
	for _, slot := range slots {
		contextual = slot
		if found := l.freshOrWidened(value, own, contextual); found != nil {
			return contextual, found
		}
	}
	if literal := l.slotType(member.Parent); literal != nil && len(slots) > 0 {
		return l.memberLiteralSlotWidened(value, member.Parent, own, literal, key, contextual)
	}
	return contextual, nil
}

// conditionalBranchInSlot is whether node is a branch of a conditional that goes straight into a typed
// slot. The conditional's own type, its branches' union reduced to the wider, is then held by nothing:
// the value goes into the slot, and refuseWidening judges each branch against that slot at the
// conditional (freshOrWidened's conditional arm). Judging a branch against the reduced union as well
// refused tsc's `this._alternatives = spec ? parseRange(spec) : emptyArray` (semver.ts:211, #2vjqseg),
// a never[] seen as Comparator[][] that no reference ever holds at that type, where the slot is
// readonly. A conditional with no slot keeps its implied target: in `const all = flag ? dogs :
// animals` the variable holds the union, and a push through it writes into either branch.
func (l *lowering) conditionalBranchInSlot(node *ast.Node) bool {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil || parent.Kind != ast.KindConditionalExpression {
		return false
	}
	conditional := parent.AsConditionalExpression()
	if conditional.WhenTrue != child && conditional.WhenFalse != child {
		return false
	}
	return viewSite(parent) && l.slotType(parent) != nil
}

// operandLanding is where an operand of ||, ?? or ?: lands, read up through parentheses and the nested
// operators of those three, which pass the landing down (Adamic's ruling on #r11wppy, applying #7n03qzg,
// rulings.md:71 and :93). The operator's own type holds nothing, has no storage, and no name writes
// through it, so the operand is judged once, where the value goes:
//   - A declared slot (an annotated binding, parameter default or field, an argument's parameter, a
//     return or an arrow's body under a declared return type): slot is that slot's type, as if the
//     operand were written there, so a readonly T[] lowers and a T[] still refuses the mutable widening.
//   - A consumer that holds nothing (a spread's operand, a for...of's iterable): read is set, and the
//     operand is judged against the operator's implied type as a read (readWidened).
//
// Anything else answers neither, and the operand keeps the operator's implied type: an unannotated
// binding, whose inferred type is the mutable one, a receiver ((a || empty).push(x)), an assignment or
// an element write, a cast or a satisfies, an unannotated return. So does a declared slot the checker
// gives no type, which fails closed. Every later handle is judged at its own site, as always.
func (l *lowering) operandLanding(node *ast.Node) (slot *checker.Type, read bool) {
	value := node
	for value.Parent != nil && value.Parent.Kind == ast.KindParenthesizedExpression {
		value = value.Parent
	}
	operand := false
	for parent := value.Parent; parent != nil && passesLanding(parent, value); parent = value.Parent {
		operand = true
		value = parent
		for value.Parent != nil && value.Parent.Kind == ast.KindParenthesizedExpression {
			value = value.Parent
		}
	}
	parent := value.Parent
	if !operand || parent == nil {
		return nil, false
	}
	declared := false
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindPropertyDeclaration:
		// An unannotated binding holds the operator's inferred type, which is mutable: no landing.
		declared = parent.Initializer() == value && parent.Type() != nil
	case ast.KindCallExpression:
		declared = parent.AsCallExpression().Expression != value
	case ast.KindNewExpression:
		declared = parent.AsNewExpression().Expression != value
	case ast.KindReturnStatement:
		function := parent
		for function != nil && !ast.IsFunctionLike(function) {
			function = function.Parent
		}
		declared = function != nil && function.Type() != nil
	case ast.KindArrowFunction:
		declared = parent.Body() == value && parent.Type() != nil
	case ast.KindSpreadElement:
		return nil, parent.AsSpreadElement().Expression == value
	case ast.KindForOfStatement:
		return nil, parent.AsForInOrOfStatement().Expression == value
	}
	if !declared {
		return nil, false
	}
	return l.slotType(value), false
}

// passesLanding is whether value, under its parentheses, is an operand of parent that the value passes
// through to parent's own landing: a conditional's branch, or either operand of || or ??.
func passesLanding(parent, value *ast.Node) bool {
	switch parent.Kind {
	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		return conditional.WhenTrue == value || conditional.WhenFalse == value
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return true
		}
	}
	return false
}

// readWidened judges an operand that lands where nothing holds it (operandLanding): a spread or a
// for...of reads its elements out once, so nothing writes through the operator's type, and the
// operand's own slots are covariant there, as a fresh container's are (freshOrWidened). What it holds
// is kept by the reader, an element bound to the loop's name or copied into the new array, so that is
// still walked as a view: `for (const list of dogLists || animalLists) list.push(cat)` writes a cat
// among the dogs. An operand that isn't a container pairing one in the implied type is judged as before.
func (l *lowering) readWidened(node *ast.Node, own *checker.Type, contextual *checker.Type) *widening {
	source, target := l.withoutUndefined(own), l.withoutUndefined(contextual)
	for _, viewed := range l.containers(target) {
		for _, inside := range l.containers(source) {
			if l.sameContainer(inside, viewed) {
				return l.widenedArguments(inside, viewed, false, map[[2]*checker.Type]bool{})
			}
		}
	}
	return l.freshOrWidened(node, own, contextual)
}

// castSlotWidened judges an as or satisfies at a view site as the value it is: the cast's own type
// against the slot it goes into, as the default branch judges any other value there. The cast's relation
// to its operand is judged apart from it, and passing that says nothing of the slot, so without this
// inner satisfies string[] went into a (string | number)[] and wrote a number past the strings (#b6m00y7).
// The operand under the casts decides freshness, so a literal made right there stays fresh.
func (l *lowering) castSlotWidened(node *ast.Node) (*checker.Type, *checker.Type, *widening) {
	if !viewSite(node) {
		return nil, nil, nil
	}
	contextual := l.slotType(node)
	if contextual == nil {
		return nil, nil, nil
	}
	operand := ast.SkipParentheses(node)
	for operand.Kind == ast.KindAsExpression || operand.Kind == ast.KindSatisfiesExpression {
		operand = ast.SkipParentheses(operand.Expression())
	}
	own := l.checker.GetTypeAtLocation(node)
	return own, contextual, l.freshOrWidened(operand, own, contextual)
}

// castSlotRefusal refuses a cast whose own type goes into a slot that can write what it can't hold, the
// judgment castSlotWidened makes, for a checked cast, whose relation to its operand stringWideningRefusal
// leaves to cast.go.
func (l *lowering) castSlotRefusal(node *ast.Node) error {
	if l.nodeFSFileReadOnlyArgument(node) || l.regexCallbackIntrinsicView(node) {
		return nil
	}
	own, contextual, found := l.castSlotWidened(node)
	if found == nil {
		return nil
	}
	return l.wideningRefusal(node, own, contextual, found)
}

// An object literal written where a union goes is built as one of its members, so a property is
// judged against each member's own slot, not the union's property. For { kind: Kind.Text } |
// { kind: Kind.Count } that property is Kind.Text | Kind.Count, the whole open enum, which takes any
// number; the member it's built as promises one literal. Each member-literal slot the value can be
// assigned to is judged as if it were the only one; the union's own property stays the answer when
// none of them refuses. A value no member's slot takes reaches one only by tsc decomposing the union,
// and an open enum there is refused (openEnumTagDecomposed, #egmxf20), unless some member takes object,
// the literal, whole: it's built as that member, and another member's literal slot promises it nothing.
func (l *lowering) memberLiteralSlotWidened(node *ast.Node, object *ast.Node, own, literal *checker.Type, name string, contextual *checker.Type) (*checker.Type, *widening) {
	if literal.Flags()&checker.TypeFlagsUnion == 0 {
		return contextual, nil
	}
	for _, member := range literal.Types() {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		property := l.checker.GetPropertyOfType(member, name)
		if property == nil {
			continue
		}
		slot := l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(property))
		if slot.Flags()&checker.TypeFlagsNumberLiteral == 0 && !l.numericEnum(l.enumIdentity(slot)) || l.openNumericEnumType(slot) {
			continue
		}
		if l.openEnumTagDecomposed(own, slot) && !l.takenWhole(l.checker.GetTypeAtLocation(object), literal) {
			return slot, &widening{source: own, target: slot, enum: true}
		}
		if !l.checker.IsTypeAssignableTo(own, slot) {
			continue
		}
		if found := l.freshOrWidened(node, own, slot); found != nil {
			return slot, found
		}
	}
	return contextual, nil
}

// takenWhole is whether some member of a union that isn't a primitive takes from as it is, so a value
// of type from goes in as that member, never by tsc decomposing the union.
func (l *lowering) takenWhole(from, union *checker.Type) bool {
	return slices.ContainsFunc(union.Types(), func(member *checker.Type) bool {
		return member.Flags()&checker.TypeFlagsPrimitive == 0 && l.checker.IsTypeAssignableTo(from, member)
	})
}

// openEnumTagDecomposed is whether own, a value's type, reaches slot, one member literal of a numeric enum,
// only by tsc decomposing a discriminated union: own is that enum's whole open type, or a union of two or
// more of its member literals, so the member the value is built as promises one literal it may not hold.
// A numeric enum holds any number (rulings.md, adamic/enum-literal), and narrowing that excludes members by
// !== leaves the rest open too, so `kind !== Kind.Triangle` is still 7. One literal is the slot's own
// promise, and a slot that is the whole enum (an enum of one member) takes any number.
func (l *lowering) openEnumTagDecomposed(own, slot *checker.Type) bool {
	slot = l.checker.GetNonNullableType(slot)
	if slot.Flags()&checker.TypeFlagsNumberLiteral == 0 || !l.numericEnum(l.enumIdentity(slot)) || l.openNumericEnumType(slot) {
		return false
	}
	own = l.withoutUndefined(own)
	if l.openNumericEnumType(own) {
		return l.enumIdentity(l.checker.GetNonNullableType(own)) == l.enumIdentity(slot)
	}
	members := l.definedMembers(own)
	if len(members) < 2 {
		return false
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsNumberLiteral == 0 || l.enumIdentity(member) != l.enumIdentity(slot) {
			return false
		}
	}
	return true
}

func (l *lowering) wideningRefusal(node *ast.Node, own, contextual *checker.Type, found *widening) error {
	if found.declaredKey != "" {
		return l.declaredKeyRefusal(node, own, contextual, found)
	}
	if found.genericMethod != "" {
		return l.notYet(node, "a generic method "+found.genericMethod+" of type "+l.checker.TypeToString(found.source)+" seen as "+l.checker.TypeToString(found.target))
	}
	if found.unfixedGeneric {
		return l.notYet(node, "a generic function of type "+l.checker.TypeToString(found.source)+" seen as "+l.checker.TypeToString(found.target)+", which doesn't fix every type parameter it's instantiated with, so one falls back to a constraint holding any")
	}
	if found.enum {
		if enumObjectSymbol(found.target) != nil {
			return &Refused{Where: l.program.Where(node), What: "a structural object seen as " + l.checker.TypeToString(found.target) + "; the complete enum shape is unproven", Fix: "use the enum's runtime object or a typeof alias, or give the ordinary object an explicit interface"}
		}
		if l.flagEnum(l.enumIdentity(found.target)) && !l.numericEnum(l.enumIdentity(found.target)) {
			return l.flagWriteRefusal(node, found.target)
		}
		if l.numericEnum(l.enumIdentity(found.target)) || l.numericEnum(l.enumIdentity(found.source)) {
			return &Refused{Where: l.program.Where(node), What: "an unproven value assigned to a numeric literal or enum member slot " + l.checker.TypeToString(found.target), Fix: "compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"}
		}
		return &Refused{Where: l.program.Where(node), What: "an arbitrary number or a value from another enum assigned to " + l.checker.TypeToString(found.target) + "; its members are a closed union", Fix: "use a declared member of this enum, or compare the number with its members and return the matching member (adamic/enum-members)"}
	}
	what := "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read"
	fix := "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
	if found.parameter {
		what = "a function taking " + l.checker.TypeToString(found.source) + " seen as one taking " + l.checker.TypeToString(found.target) + " (tsc relates a method's parameters both ways), so it can be handed what it can't take"
		fix = "write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)"
	}
	if found.methodField != "" {
		what = "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", whose method " + found.methodField + " becomes a writable field: a write of " + l.checker.TypeToString(found.target) + " would replace a method of type " + l.checker.TypeToString(found.source)
		fix = "make " + found.methodField + " readonly in the type it's seen as (readonly " + found.methodField + ": ...), or declare it there as a method, which isn't written (adamic/invariant-mutable)"
	}
	if found.readonlyField != "" {
		what = "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", whose readonly field " + found.readonlyField + " becomes writable: a readonly field may hold something narrower than " + l.checker.TypeToString(found.source) + ", which a write of " + l.checker.TypeToString(found.target) + " would replace"
		fix = "keep " + found.readonlyField + " readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
	}
	if found.target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		what = "a value of type " + l.checker.TypeToString(found.source) + " seen as " + l.checker.TypeToString(found.target) + ", a type parameter whose constraint " + l.checker.TypeToString(l.checker.GetBaseConstraintOfType(found.target)) + " can be written, so it can write what " + l.checker.TypeToString(found.source) + " can't hold"
		fix = "take it as " + l.checker.TypeToString(found.source) + ", or constrain " + l.checker.TypeToString(found.target) + " to something readonly, which can't write (adamic/invariant-mutable)"
	}
	return &Refused{Where: l.program.Where(node), What: what, Fix: fix}
}

// freshOrWidened is widened for a value at a site, but a fresh one: a container made right there and
// held by nothing else (freshValue) can't be written through another name, so its own slots are
// covariant, and only what's inside it, held elsewhere too, is walked as a view. A function written
// there is fresh the same way for what each call returns, when its own body makes every result
// (freshFunctionWidened), and so is the copy a native map makes with such a callback
// (freshContainerWidened).
func (l *lowering) freshOrWidened(node *ast.Node, own *checker.Type, contextual *checker.Type) *widening {
	node = ast.SkipParentheses(node)
	own = l.contextualGenericType(node, own)
	// A numeric enum read narrowed to never has a non-returning IR check at this exact site.
	// It cannot write a value into the contextual slot, including a closed string-enum slot.
	if own.Flags()&checker.TypeFlagsNever != 0 && l.enumNeverIdentity(node, map[*ast.Node]bool{}) != nil {
		return nil
	}
	// A slot is a literal one by its enum part alone (openEnumSlot), so a whole enum beside an object, as in
	// `CompilerOptions | BuilderFileEmit`, is open.
	targetLiteral := (contextual.Flags()&checker.TypeFlagsNumberLiteral != 0 || l.numericEnum(l.enumIdentity(contextual))) && !l.openEnumSlot(contextual)
	if targetLiteral {
		if declared := l.enumStoredType(node); declared != nil {
			if l.openNumericEnumType(declared) && !l.enumMemberOrigin(node, l.enumIdentity(declared), map[*ast.Node]bool{}) {
				return &widening{source: declared, target: contextual, enum: true}
			}
		}
	}
	if found := l.flagMemberWidened(node, contextual); found != nil && !l.numericEnum(l.enumIdentity(found.target)) {
		return found
	}
	if target := l.flagTarget(contextual); target != nil && l.flagDomain(node, target) {
		return nil
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		if l.freshObjectLiteral(node) {
			return nil
		}
	case ast.KindArrayLiteralExpression:
		// Made as the type it's written into, held by nothing else: its own parts are sites.
		return nil
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		// A function written right here has no handle but the slot, so it's only ever called as the slot
		// sees it, and when every value it returns is made in its own body and stored nowhere, each call's
		// result is new, as a literal's own method's is (literalMethodRefusal): its own slots are covariant,
		// and only what they hold is a view (freshFunctionWidened, #k8ed1f2, Adamic's ruling). Any other
		// function is judged by its type below.
		// Only where today's walk refuses: an arrow whose own type the slot already holds whole (one written
		// `(): ScriptTargetFeatures =>`) is let through as before, and its returns stop at their own sites.
		if l.widened(own, contextual, map[[2]*checker.Type]bool{}) != nil {
			if found, judged := l.freshFunctionWidened(node, own, contextual); judged {
				return found
			}
		}
	case ast.KindConditionalExpression:
		// Either branch is the value: each judged as it is.
		conditional := node.AsConditionalExpression()
		for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
			if found := l.freshOrWidened(branch, l.checker.GetTypeAtLocation(branch), contextual); found != nil {
				if found.viewed == nil {
					found.viewed = l.checker.GetTypeAtLocation(branch)
				}
				return found
			}
		}
		return nil
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken:
			// A fresh literal held only by the chain's targets, each declaring what the slot reads
			// (freshAssignmentHandles). Anything else is judged by its own type below.
			if l.freshAssignmentHandles(node, contextual) {
				return nil
			}
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			// Either operand is the value, as a conditional's branches are: each judged as it is, so
			// a fresh one keeps its freshness (#r11wppy). The left keeps its own type: no truthy
			// narrowing is invented for it, so a left that refuses on its own still refuses.
			for _, operand := range []*ast.Node{binary.Left, binary.Right} {
				if found := l.freshOrWidened(operand, l.checker.GetTypeAtLocation(operand), contextual); found != nil {
					if found.viewed == nil {
						found.viewed = l.checker.GetTypeAtLocation(operand)
					}
					return found
				}
			}
			return nil
		case ast.KindAmpersandAmpersandToken:
			// Only the right is the value, and the left reaches the result only by its falsy arms
			// (andLeftOperand, falsyArms), so the left is judged by those, and not at all when it has none,
			// and the right as itself, as || judges each operand: a fresh one keeps its freshness (#k8ed1f2,
			// applying #7n03qzg). moduleNameResolver.ts's `resolved && { ... }` is a literal made right
			// there, or undefined. It's consulted only where today's walk of the whole value refuses: judging each
			// operand on its own reaches operands that walk never judged (an enum literal, a field read), so an
			// && the walk lets through is let through as before.
			if l.widened(own, contextual, map[[2]*checker.Type]bool{}) == nil {
				return nil
			}
			if falsy := l.falsyArms(l.checker.GetTypeAtLocation(binary.Left)); falsy != nil {
				if found := l.freshOrWidened(binary.Left, falsy, contextual); found != nil {
					if found.viewed == nil {
						found.viewed = falsy
					}
					return found
				}
			}
			if found := l.freshOrWidened(binary.Right, l.checker.GetTypeAtLocation(binary.Right), contextual); found != nil {
				if found.viewed == nil {
					found.viewed = l.checker.GetTypeAtLocation(binary.Right)
				}
				return found
			}
			return nil
		}
	}
	visited := map[[2]*checker.Type]bool{}
	if !l.freshValue(node) {
		return l.widened(own, contextual, visited)
	}
	return l.freshContainerWidened(own, contextual, l.freshMapCallback(node), visited)
}

// freshContainerWidened judges a container made at the site (freshValue, or an array literal a fresh
// function returns) seen as contextual: no other name writes its own slots, so they're covariant, and an
// array, tuple, map or set pairing one in contextual is judged by its elements as views (widenedArguments),
// anything else by widened. mapped is a native map's callback that's fresh for what each call returns
// (freshMapCallback), or nil. The copy's elements are then exactly what the callback returns, each a value
// made in it, so each is judged as made at the site against the slot's element (madeAtSiteWidened), never
// by map's own element type, which tsc infers from the callback's return and not from the slot. Any other
// copy (slice, filter, concat, Array.from, a new Map) holds elements something else holds too, and is
// walked as a view of them (#k8ed1f2, Adamic's ruling).
func (l *lowering) freshContainerWidened(own *checker.Type, contextual *checker.Type, mapped *ast.Node, visited map[[2]*checker.Type]bool) *widening {
	own, contextual = l.withoutUndefined(own), l.withoutUndefined(contextual)
	for _, viewed := range l.containers(contextual) {
		for _, inside := range l.containers(own) {
			if !l.sameContainer(inside, viewed) {
				continue
			}
			if mapped != nil && !checker.IsTupleType(viewed) && len(l.checker.GetTypeArguments(viewed)) == 1 && l.widenedArguments(inside, viewed, false, map[[2]*checker.Type]bool{}) != nil {
				// Only where today's walk of the copy's elements refuses, so a copy it already lets through
				// is judged as before.
				element := l.checker.GetTypeArguments(viewed)[0]
				for _, returned := range siteResults(mapped) {
					if found := l.madeAtSiteWidened(returned, element, visited); found != nil {
						return found
					}
				}
				return nil
			}
			return l.widenedArguments(inside, viewed, false, visited)
		}
	}
	return l.widened(own, contextual, visited)
}

// freshMapCallback is a native map's callback when it's written at the call and fresh for what each call
// returns (freshFunction), or nil. Only freshValue's copies reach it, so a map here is already proved an
// Array's or a ReadonlyArray's own (#k8ed1f2, Adamic's ruling).
func (l *lowering) freshMapCallback(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.AsPropertyAccessExpression().Name().Text() != "map" || len(call.Arguments.Nodes) == 0 {
		return nil
	}
	callback := ast.SkipParentheses(call.Arguments.Nodes[0])
	if !l.freshFunction(callback) {
		return nil
	}
	return callback
}

// madeAtSiteWidened judges a value a fresh function returns (freshFunction) against target, what the slot
// says each call gives: a conditional by each branch, the first finding winning; an object literal by what
// its own slots hold, since nothing else holds the slots themselves; an array literal by its elements, as
// freshContainerWidened judges a copy's; and a new Map, Set or Array or a native copy as freshOrWidened
// judges one at its own site. The literal is judged here, against target, rather than left to its own
// sites as freshOrWidened leaves one: inside a map callback, the literal's sites see map's element type,
// which tsc takes from the callback's own return, so `holders.map(h => ({ inner: h.dogs }))` written into
// `{ inner: Animal[] }[]` would judge h.dogs as a Dog[] and let a held Dog[] out as a writable Animal[].
// Judged here, h.dogs is a view of Animal[], and refused. A slot member that isn't an object, or that
// has an index signature, is judged by widened, the walk it had before (#k8ed1f2, Adamic's ruling).
func (l *lowering) madeAtSiteWidened(value *ast.Node, target *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	value = ast.SkipParentheses(value)
	own := l.checker.GetTypeAtLocation(value)
	switch value.Kind {
	case ast.KindConditionalExpression:
		conditional := value.AsConditionalExpression()
		for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
			if found := l.madeAtSiteWidened(branch, target, visited); found != nil {
				return found
			}
		}
		return nil
	case ast.KindObjectLiteralExpression:
		if !l.freshObjectLiteral(value) {
			return l.widened(own, target, visited)
		}
		for _, member := range l.literalSlotMembers(value, target) {
			var found *widening
			if !l.structured(member) || len(l.checker.GetIndexInfosOfType(member)) > 0 {
				found = l.widened(own, member, visited)
			} else {
				found = l.widenedProperties(own, member, nil, true, visited)
			}
			if found != nil {
				return found
			}
		}
		return nil
	case ast.KindArrayLiteralExpression:
		return l.freshContainerWidened(own, target, nil, visited)
	}
	return l.freshOrWidened(value, own, target)
}

// freshValue reports whether an expression makes a container nothing else holds: a copy (slice,
// filter, map, concat, Array.from) or a new Map, Set or Array.
func (l *lowering) freshValue(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNewExpression:
		callee := node.AsNewExpression().Expression
		return l.isLibraryGlobal(callee, "Map") || l.isLibraryGlobal(callee, "Set") || l.isLibraryGlobal(callee, "Array")
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return false
		}
		access := callee.AsPropertyAccessExpression()
		method := access.Name().Text()
		if method == "from" {
			return l.isLibraryGlobal(access.Expression, "Array")
		}
		switch method {
		case "slice", "filter", "map", "concat":
			// Under ?. the receiver reads with its undefined, but the call runs only when it's there, so
			// x?.map(f) is undefined, which holds nothing, or native map's new array: the receiver is read
			// without null and undefined, and is still an Array or a ReadonlyArray or it's no copy
			// (#k8ed1f2, applying #7n03qzg).
			receiver := l.checker.GetTypeAtLocation(access.Expression)
			if access.QuestionDotToken != nil {
				receiver = l.checker.GetNonNullableType(receiver)
			}
			return l.isArrayType(receiver) || l.isLibraryType(receiver, "ReadonlyArray")
		}
	}
	return false
}

// isExpression reports whether a node is one a value can be: anything but a statement, a declaration,
// a name being declared or a type.
func (l *lowering) isExpression(node *ast.Node) bool {
	return ast.IsExpression(node) && !ast.IsPartOfTypeNode(node)
}

// hasTypeAncestor catches literals inside types that IsPartOfTypeNode does not classify, such as
// the string in X["literal"]. Those name types, not runtime reads. A statement or declaration ends
// the search so a value in its initializer or body is judged in its own context.
// IsTypeNodeKind's extra AnyKeyword, UnknownKeyword, NumberKeyword, BigIntKeyword, ObjectKeyword,
// BooleanKeyword, StringKeyword, SymbolKeyword, VoidKeyword, UndefinedKeyword, NeverKeyword and
// IntrinsicKeyword hold no expressions; neither do JSDocAllType, JSDocNullableType,
// JSDocNonNullableType, JSDocOptionalType or JSDocVariadicType. In its TypePredicate..ImportType
// range, TypePredicate, TypeReference, FunctionType, ConstructorType, TypeQuery, TypeLiteral,
// ArrayType, TupleType, OptionalType, RestType, UnionType, IntersectionType, ConditionalType,
// InferType, ParenthesizedType, ThisType, TypeOperator, IndexedAccessType, MappedType, LiteralType,
// NamedTupleMember, TemplateLiteralType, TemplateLiteralTypeSpan and ImportType are erased,
// including TypeQuery's names and LiteralType's literal expressions. ExpressionWithTypeArguments
// alone can wrap an evaluated expression: a class heritage's base. Its type arguments still erase.
func hasTypeAncestor(node *ast.Node) bool {
	child := node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindExpressionWithTypeArguments && parent.Expression() == child {
			return false
		}
		if ast.IsTypeNode(parent) {
			return true
		}
		if ast.IsStatement(parent) || ast.IsDeclaration(parent) {
			return false
		}
		child = parent
	}
	return false
}

// impliedTarget is the type tsc gave what a value with no type written for it goes into, or nil: a
// conditional's (or ??'s, ||'s, &&'s) own type for its branches, an array literal's element type for
// its elements, and a function's inferred result for what it returns.
func (l *lowering) impliedTarget(node *ast.Node) *checker.Type {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindConditionalExpression:
		if context := l.callableContext(parent); context != nil {
			return context
		}
		if conditional := parent.AsConditionalExpression(); conditional.WhenTrue == child || conditional.WhenFalse == child {
			return l.impliedExpressionTarget(parent)
		}
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindQuestionQuestionToken, ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
			if context := l.callableContext(parent); context != nil {
				return context
			}
			if contextual := l.slotType(parent); contextual != nil && l.isArrayType(l.withoutUndefined(contextual)) {
				return contextual
			}
			return l.impliedExpressionTarget(parent)
		}
	case ast.KindArrayLiteralExpression:
		return l.impliedExpressionTarget(parent)
	case ast.KindReturnStatement, ast.KindArrowFunction:
		if parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body != child {
			return nil
		}
		function := parent
		for function != nil && !ast.IsFunctionLike(function) {
			function = function.Parent
		}
		if function == nil {
			return nil
		}
		return l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(function))
	}
	return nil
}

// impliedExpressionTarget asks once per parent expression, even for nil tuple
// element targets. The checker and source tree stay fixed for this lowering;
// branches share the parent's type and array elements share its element type.
func (l *lowering) impliedExpressionTarget(parent *ast.Node) *checker.Type {
	if target, ok := l.slots.implied[parent]; ok {
		return target
	}
	query := l.impliedTargetQuery
	if query == nil {
		query = l.checker.GetTypeAtLocation
	}
	target := query(parent)
	if parent.Kind == ast.KindArrayLiteralExpression {
		literal := target
		target = nil
		if l.isArrayType(literal) {
			if arguments := l.checker.GetTypeArguments(literal); len(arguments) == 1 {
				target = arguments[0]
			}
		}
	}
	if l.slots.implied == nil {
		l.slots.implied = make(map[*ast.Node]*checker.Type)
	}
	l.slots.implied[parent] = target
	return target
}

// slotTargets are the slots a value is judged against that the checker's contextual type doesn't give.
type slotTargets struct {
	// explicit is the slot a caller set for a value the checker gives none to, which slotType returns
	// before the contextual type: a delayed Array write's element, whose parameter lib types any.
	explicit map[*ast.Node]*checker.Type
	// implied memoizes impliedExpressionTarget's answer per parent, within this checker.
	implied map[*ast.Node]*checker.Type
}

// slotType is the slot a value at node goes into: the one a caller set for it (slots.explicit), or the
// one tsc's contextual typing carries down from it to a part of the value (carriedSlot), or else the
// checker's contextual type, exactly. Every refusal rule reads its slot here and never from the checker
// (TestRefusalRulesReadTheirSlotThroughSlotType), so a caller that judges a value against a slot the
// checker doesn't give sets it and runs the whole registry over the value, and no rule is skipped
// because a second list beside the registry forgot it (#b6m00y7).
func (l *lowering) slotType(node *ast.Node) *checker.Type {
	if len(l.slots.explicit) > 0 {
		if slot, carried, _ := l.carriedSlot(node); carried {
			return slot
		}
	}
	return l.checker.GetContextualType(node, checker.ContextFlagsNone)
}

// keepSlot sets slot as node's own for the rest of the lowering. refuseInSlot only judges, but lowering
// has checks of its own at a value that read slotType (expression's weak or strong keep, a host object's
// view, a tuple where an array goes), so a value written into a slot the checker doesn't give is lowered
// with that slot kept, as what it's written into, the way a direct write's value is.
func (l *lowering) keepSlot(node *ast.Node, slot *checker.Type) {
	if l.slots.explicit == nil {
		l.slots.explicit = map[*ast.Node]*checker.Type{}
	}
	l.slots.explicit[node] = slot
}

// slotSet is whether node's slot comes from one a caller set (carriedSlot). A rule that reads a site's
// slot some other way than the contextual type, such as an argument's parameter from the call's resolved
// signature, reads slotType instead when this holds: a delayed write's call is lib's, whose parameter is
// any, and the slot the caller set is the one the value is written into.
func (l *lowering) slotSet(node *ast.Node) bool {
	if len(l.slots.explicit) == 0 {
		return false
	}
	_, carried, _ := l.carriedSlot(node)
	return carried
}

// carriedSlot is node's slot when it comes from one a caller set: node's own, or the one tsc's
// contextual typing would carry down to it from an enclosing value that has one, through parentheses, a
// non-null, an as const, a conditional's branches, a logical operand, the right of a comma, and a
// literal's property, shorthand, spread or element. carried is false when no explicit slot reaches node,
// and the checker's contextual type is its slot. readable is false when one reaches it but Adamic can't
// read node's part off it (a computed name, a property the slot doesn't declare, an element of a slot
// that isn't an array or a tuple, or a parent kind this doesn't model, such as an await's operand), and
// refuseInSlot fails closed rather than judging node by the checker's. Only the kinds tsc carries
// nothing through are let go (slotIndependent), so a kind nobody listed fails closed (Compiler's must on
// c06ed629).
func (l *lowering) carriedSlot(node *ast.Node) (slot *checker.Type, carried bool, readable bool) {
	if slot, set := l.slots.explicit[node]; set {
		return slot, true, true
	}
	parent := node.Parent
	if parent == nil {
		return nil, false, false
	}
	switch parent.Kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		return l.carriedSlot(parent)
	case ast.KindAsExpression:
		// x as const keeps the slot x goes into; any other assertion gives x the asserted type.
		if as := parent.AsAsExpression(); as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
			return l.carriedSlot(parent)
		}
	case ast.KindConditionalExpression:
		if conditional := parent.AsConditionalExpression(); conditional.WhenTrue == node || conditional.WhenFalse == node {
			return l.carriedSlot(parent)
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return l.carriedSlot(parent)
		case ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			// Only the right is the value; tsc gives the left no slot.
			if binary.Right == node {
				return l.carriedSlot(parent)
			}
			if _, carried, _ := l.carriedSlot(parent); carried {
				return nil, true, true
			}
		}
	case ast.KindPropertyAssignment:
		if parent.AsPropertyAssignment().Initializer == node {
			literal, carried, readable := l.carriedSlot(parent.Parent)
			if !carried || !readable {
				return nil, carried, readable
			}
			property, readable := l.carriedProperty(literal, parent.Name())
			return property, true, readable
		}
	case ast.KindObjectLiteralExpression:
		// A member of the literal: a spread is the literal's whole slot, and a property, a shorthand, a
		// method or an accessor is its property's.
		if node.Kind == ast.KindSpreadAssignment {
			return l.carriedSlot(parent)
		}
		if literal, carried, readable := l.carriedSlot(parent); carried {
			if !readable {
				return nil, true, false
			}
			property, readable := l.carriedProperty(literal, node.Name())
			return property, true, readable
		}
	case ast.KindShorthandPropertyAssignment:
		// { p } is the variable written as the literal's property p.
		if parent.Name() == node {
			literal, carried, readable := l.carriedSlot(parent.Parent)
			if !carried || !readable {
				return nil, carried, readable
			}
			property, readable := l.carriedProperty(literal, node)
			return property, true, readable
		}
	case ast.KindSpreadAssignment:
		return l.carriedSlot(parent.Parent)
	case ast.KindSpreadElement:
		if parent.Parent.Kind == ast.KindArrayLiteralExpression {
			return l.carriedSlot(parent.Parent)
		}
	case ast.KindArrayLiteralExpression:
		array, carried, readable := l.carriedSlot(parent)
		if !carried || !readable || array == nil {
			return nil, carried, readable
		}
		array = l.checker.GetNonNullableType(array)
		elements := parent.AsArrayLiteralExpression().Elements.Nodes
		if checker.IsTupleType(array) {
			index := slices.Index(elements, node)
			arguments := l.checker.GetTypeArguments(array)
			if index < 0 || index >= len(arguments) || slices.ContainsFunc(elements[:index], func(element *ast.Node) bool { return element.Kind == ast.KindSpreadElement }) {
				return nil, true, false
			}
			return arguments[index], true, true
		}
		if l.isArrayType(array) {
			return l.elementTypeOfArray(array), true, true
		}
		return nil, true, false
	}
	if slotIndependent(parent, node) {
		return nil, false, false
	}
	if _, carried, _ := l.carriedSlot(parent); carried {
		return nil, true, false
	}
	return nil, false, false
}

// slotIndependent is whether tsc gives node no slot from its parent's, so that whatever slot node has is
// its own: a call's callee and arguments (which take its parameters), an access's object and key, an
// operator's operands but the ones carriedSlot carries, a conditional's test, an assertion's operand
// (the asserted type), an untagged template's spans, a computed key, a property's name, a spread
// argument in a call, and any type. It names the kinds that carry nothing, never the ones that carry,
// so a kind left off it fails closed in carriedSlot.
func slotIndependent(parent, node *ast.Node) bool {
	if ast.IsPartOfTypeNode(node) {
		return true
	}
	switch parent.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindBinaryExpression, ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression,
		ast.KindConditionalExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression,
		ast.KindTemplateExpression, ast.KindTemplateSpan, ast.KindComputedPropertyName, ast.KindSpreadElement,
		ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		return true
	}
	return false
}

// carriedProperty is the slot tsc carries from a literal's slot to its property name: the property's
// declared type, or the slot's string index signature. readable is false for a computed name or one
// the slot doesn't declare, which Adamic can't read off it.
func (l *lowering) carriedProperty(literal *checker.Type, name *ast.Node) (*checker.Type, bool) {
	if literal == nil {
		// No slot reaches the literal, so none reaches its property.
		return nil, true
	}
	if name == nil || !ast.IsIdentifier(name) && !ast.IsStringLiteral(name) && !ast.IsNumericLiteral(name) {
		return nil, false
	}
	literal = l.checker.GetNonNullableType(literal)
	if property := l.checker.GetPropertyOfType(literal, name.Text()); property != nil {
		return l.checker.GetTypeOfSymbol(property), true
	}
	if index := l.checker.GetIndexTypeOfType(literal, l.checker.GetStringType()); index != nil {
		return index, true
	}
	return nil, false
}

// destructuringTarget is the array or object pattern a value is assigned into, [a, b] = value or
// { a, b } = value, or nil.
func destructuringTarget(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node && (binary.Left.Kind == ast.KindArrayLiteralExpression || binary.Left.Kind == ast.KindObjectLiteralExpression) {
			return binary.Left
		}
	}
	return nil
}

// refuseWholeWidening judges a value assigned into a pattern as one value seen as the pattern's
// type, the way any other value at a view site is judged, for a value whose parts the pattern
// doesn't read out one by one. A fresh literal, ({ x: options[name] } = { x: 3 }), keeps
// freshOrWidened's exemption: it's made as the pattern's type and held by nothing else.
func (l *lowering) refuseWholeWidening(node *ast.Node, own *checker.Type) error {
	contextual := l.slotType(node)
	if contextual == nil {
		return nil
	}
	if found := l.freshOrWidened(node, own, contextual); found != nil {
		return l.wideningRefusal(node, own, contextual, found)
	}
	return nil
}

// refuseElementWidening refuses [a, b] = tuple when an element is seen through its name's type as
// a view that can write what the element can't hold: [animals] = [dogs] puts a Dog[] in an
// Animal[]. A value that isn't a tuple is judged whole, as anywhere.
func (l *lowering) refuseElementWidening(node *ast.Node, pattern *ast.Node) error {
	own := l.checker.GetTypeAtLocation(node)
	if !checker.IsTupleType(own) {
		return l.refuseWholeWidening(node, own)
	}
	elements := l.checker.GetTypeArguments(own)
	for index, target := range pattern.AsArrayLiteralExpression().Elements.Nodes {
		if index >= len(elements) || target.Kind == ast.KindOmittedExpression {
			continue
		}
		to := l.checker.GetTypeAtLocation(target)
		if found := l.widened(elements[index], to, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{Where: l.program.Where(target), What: "a value of type " + l.checker.TypeToString(elements[index]) + " seen as " + l.checker.TypeToString(to) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read", Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"}
		}
	}
	return nil
}

// patternField is one field an object pattern reads out of a value: the value's field, the target
// it's stored into with that target's declared type, and the default stored there when the field
// is undefined, or nil.
type patternField struct {
	source       *checker.Type
	target       *ast.Node
	targetType   *checker.Type
	defaultValue *ast.Node
}

// refuseFieldWidening refuses { a, b: c } = value when a field is seen through its target's type
// as a view that can write what the field can't hold: ({ pets } = kennel()) puts kennel's Dog[] in
// an Animal[]. A default is judged against its target too. A value the pattern can't read field by
// field is judged whole, as anywhere.
func (l *lowering) refuseFieldWidening(node *ast.Node, pattern *ast.Node) error {
	own := l.checker.GetTypeAtLocation(node)
	fields, readable := l.patternFields(own, pattern)
	if !readable {
		return l.refuseWholeWidening(node, own)
	}
	for _, field := range fields {
		if found := l.widened(field.source, field.targetType, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{Where: l.program.Where(field.target), What: "a value of type " + l.checker.TypeToString(field.source) + " seen as " + l.checker.TypeToString(field.targetType) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read", Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"}
		}
		if field.defaultValue == nil {
			continue
		}
		fallback := l.checker.GetTypeAtLocation(field.defaultValue)
		if found := l.widened(fallback, field.targetType, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{Where: l.program.Where(field.defaultValue), What: "a value of type " + l.checker.TypeToString(fallback) + " seen as " + l.checker.TypeToString(field.targetType) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read", Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"}
		}
	}
	return nil
}

// patternFields lists the fields an object pattern reads out of a value of type own, each by a
// name own declares. readable is false where the pattern can't be read field by field: own is a
// union or an intersection, an element is a rest or has a computed key, a target is a nested
// pattern, an element access or carries its own default ({ a: b = 1 }), or own doesn't declare
// the name.
func (l *lowering) patternFields(own *checker.Type, pattern *ast.Node) (fields []patternField, readable bool) {
	if own.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		return nil, false
	}
	for _, property := range pattern.AsObjectLiteralExpression().Properties.Nodes {
		var field patternField
		switch property.Kind {
		case ast.KindShorthandPropertyAssignment:
			// { a } and { a = 1 }: the field goes into the variable a names.
			symbol := l.checker.GetShorthandAssignmentValueSymbol(property)
			if symbol == nil {
				return nil, false
			}
			field.target = property.Name()
			field.targetType = l.checker.GetTypeOfSymbol(symbol)
			field.defaultValue = property.AsShorthandPropertyAssignment().ObjectAssignmentInitializer
		case ast.KindPropertyAssignment:
			// { a: b } and { a: this.b }: an assignment target's type is its declared type.
			initializer := property.AsPropertyAssignment().Initializer
			if !ast.IsIdentifier(initializer) && !ast.IsPropertyAccessExpression(initializer) {
				return nil, false
			}
			field.target = initializer
			field.targetType = l.checker.GetTypeAtLocation(initializer)
		default:
			return nil, false
		}
		name := property.Name()
		if name == nil || !ast.IsIdentifier(name) && !ast.IsStringLiteral(name) && !ast.IsNumericLiteral(name) {
			return nil, false
		}
		symbol := l.checker.GetPropertyOfType(own, name.Text())
		if symbol == nil || field.targetType == nil {
			return nil, false
		}
		field.source = l.checker.GetTypeOfSymbol(symbol)
		fields = append(fields, field)
	}
	return fields, true
}

// genericFunction reports whether an expression is a function with type parameters of its own,
// which tsc instantiates to the type it's written into rather than viewing it through that type.
func (l *lowering) genericFunction(node *ast.Node) bool {
	for _, signature := range l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}
