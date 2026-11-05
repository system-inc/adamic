package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// containerFitStop is the stop for a value held one way going where a container is held another. A
// ReadonlyMap or ReadonlySet slot is held as ir.Map (representation), and every operation on it is the
// native map's (object.go's mapMethod and setMethod, MapSize, a for-of's MapEntries), but tsc lets any
// value with the interface's members go there: an object literal, or a class instance, both held as
// ir.Object. On main e4c3a163 such a value lowered, so `m.has('a')` called the JavaScript backend's
// Map.prototype.has on a plain object (exit 70 where Node prints true), natively clang refused the C
// that stored an object pointer in a map slot, and through a field, an element, a union or a callback's
// result the native build ran and crashed (#efx3yb9). Set and Map need [Symbol.toStringTag], a computed
// field that stops already, so only the read-only interfaces carry such a value today. ReadonlyArray
// (its [Symbol.unscopables] and number index) and Promise (its [Symbol.toStringTag]) stop before
// they're reached, so ir.Array and ir.Promise slots aren't judged here; a tuple, held as an object, has
// its own stop where an array goes (tupleWhereArrayGoes).
const containerFitStop = "a Map or Set interface implemented by a value that isn't a Map or Set"

// containerFitDepth is how many positions deep containerWhereOtherGoes walks a value against its slot.
// Past it the walk fails closed: a pair still unjudged there stops if either side can hold a map, so a
// value nested deeper than the walk can never carry a shaped value into a map slot unseen (Compiler's depth-5 probe on #efx3yb9, a class
// implementing ReadonlyMap five fields down, which crashed natively and printed false on JavaScript), or,
// for the converse (shape_fit.go), an array, so none reaches a structural slot unseen either.
const containerFitDepth = 4

// containerMentionDepth bounds mentionsHeld. Reaching it counts as a mention, so it fails closed too.
const containerMentionDepth = 16

// containerWhereOtherGoes is the stop for a value of type value, going where target is expected, that
// carries a value held one way to where a container is held another, or "" if it carries none, in either
// direction: a value held other than as a map where a map is held (#efx3yb9), or a real Map, Set or array
// where a structural object type is held, read by a member shapeFitServed doesn't serve (#nw6d7ws,
// shape_fit.go). It looks directly, as an array's elements, as a Map's keys or values, as a field of an
// object or an intersection or a value of a dictionary, or as a function's result, or the other way round
// for a function's parameters. It's keyed on representation alone, so it asks nothing of the syntax that
// made the value (Adamic's rulings on #efx3yb9 and #nw6d7ws). A value of any, never or a type
// representation can't read is some other check's to stop, and isn't judged here. When neither side can
// hold a map or an array anywhere (mentionsHeld), no position can carry one, and the pair isn't walked at
// all: that's what keeps a Promise's then, which makes a new instantiation at every step, out of the walk.
// libraryArgument says the value is a library call's own argument (libraryCallArgument), which only the
// converse's top position reads (shapeFitProtocols).
func (l *lowering) containerWhereOtherGoes(value *checker.Type, target *checker.Type, libraryArgument bool) string {
	if !l.mentionsHeld(value, containerHeld, 0, map[*checker.Type]bool{}) && !l.mentionsHeld(target, containerHeld, 0, map[*checker.Type]bool{}) {
		return ""
	}
	return l.containerPairGoes(value, target, libraryArgument, 0, map[[2]*checker.Type]bool{})
}

// containerHeld is what either direction carries: a map, and for the converse an array too.
var containerHeld = []ir.Type{ir.Map, ir.Array}

// mentionsHeld reports whether a value of type proven can hold a value held as one of held anywhere
// inside (eachReached), and past containerMentionDepth it answers yes.
func (l *lowering) mentionsHeld(proven *checker.Type, held []ir.Type, depth int, seen map[*checker.Type]bool) bool {
	return l.eachReached(proven, depth, seen, func(reached *checker.Type) bool {
		of, known := l.representation(reached)
		return known && slices.Contains(held, of)
	})
}

// eachReached calls visit with every type a value of type proven can hold, until visit answers true:
// itself, a type argument, a field, a dictionary's values, or a parameter or result of its call and
// construct signatures, read from their declarations so a generic method's own type parameters stay the
// same types at every step. A library interface or class holds only what its type arguments say (a
// Promise<number> holds no map), but an anonymous type the library declares, such as Map.forEach's
// callback instantiated for a Map<string, Map<string, number>>, holds what its signatures and fields
// do, since it has no type arguments to say it. A type parameter holds what its constraint does. Each type is read once. It answers true
// when visit does, and past containerMentionDepth too, so a caller reading true as a mention fails closed.
func (l *lowering) eachReached(proven *checker.Type, depth int, seen map[*checker.Type]bool, visit func(*checker.Type) bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	if depth > containerMentionDepth {
		return true
	}
	seen[proven] = true
	if visit(proven) {
		return true
	}
	flags := proven.Flags()
	if flags&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, member := range proven.Types() {
			if l.eachReached(member, depth+1, seen, visit) {
				return true
			}
		}
		return false
	}
	if flags&checker.TypeFlagsTypeParameter != 0 {
		return l.eachReached(l.checker.GetBaseConstraintOfType(proven), depth+1, seen, visit)
	}
	if flags&checker.TypeFlagsObject == 0 {
		return false
	}
	for _, argument := range l.typeArguments(proven) {
		if l.eachReached(argument, depth+1, seen, visit) {
			return true
		}
	}
	if l.librarySymbol(proven.Symbol()) && proven.ObjectFlags()&checker.ObjectFlagsAnonymous == 0 {
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(proven) {
		if l.eachReached(l.checker.GetTypeOfSymbol(field), depth+1, seen, visit) {
			return true
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(proven) {
		if l.eachReached(info.ValueType(), depth+1, seen, visit) {
			return true
		}
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		for _, signature := range l.checker.GetSignaturesOfType(proven, kind) {
			if declaration := signature.Declaration(); declaration != nil && ast.IsFunctionLike(declaration) {
				signature = l.checker.GetSignatureFromDeclaration(declaration)
			}
			for _, parameter := range signature.Parameters() {
				if l.eachReached(l.checker.GetTypeOfSymbol(parameter), depth+1, seen, visit) {
					return true
				}
			}
			if l.eachReached(l.checker.GetReturnTypeOfSignature(signature), depth+1, seen, visit) {
				return true
			}
		}
	}
	return false
}

// containerPairGoes is containerWhereOtherGoes' walk of one pair, depth positions in. libraryArgument
// holds only at the top and through a union's members there.
func (l *lowering) containerPairGoes(value *checker.Type, target *checker.Type, libraryArgument bool, depth int, visited map[[2]*checker.Type]bool) string {
	if value == nil || target == nil || value == target || visited[[2]*checker.Type{value, target}] {
		return ""
	}
	visited[[2]*checker.Type{value, target}] = true
	if value.Flags()&checker.TypeFlagsUnion != 0 {
		// Any member that may be there.
		for _, member := range value.Types() {
			if stop := l.containerPairGoes(member, target, libraryArgument, depth+1, visited); stop != "" {
				return stop
			}
		}
		return ""
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		// Into a union: fine if a member that takes the value takes it as it is. A member it isn't
		// assignable to (undefined, for an object) never holds it.
		stop := ""
		for _, member := range target.Types() {
			if !l.checker.IsTypeAssignableTo(value, member) {
				continue
			}
			if stop = l.containerPairGoes(value, member, libraryArgument, depth+1, visited); stop == "" {
				return ""
			}
		}
		return stop
	}
	if value.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsNever) != 0 {
		return ""
	}
	if slot, known := l.representation(target); known && slot == ir.Map {
		held, known := l.representation(value)
		if known && held != ir.Map {
			return containerFitStop
		}
		if !known {
			// An intersection Adamic holds no way at all (Fake & { tag: 1 }, a class or a collection
			// inside) can't be proven a map, so it fails closed. Any other unread type is some other
			// check's to stop.
			if value.Flags()&checker.TypeFlagsIntersection != 0 {
				return containerFitStop
			}
			return ""
		}
		if !l.isLibraryType(value, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
			return ""
		}
		// A map's keys and values go where the slot's go.
		return l.pairwiseContainerPairGoes(l.typeArguments(value), l.typeArguments(target), depth, visited)
	}
	if stop, judged := l.shapeFit(value, target, libraryArgument); judged {
		// The converse: a map or an array where a structural object type is held, a leaf that costs
		// nothing at any depth, like the map slot above.
		return stop
	}
	if depth > containerFitDepth {
		// Anything else this deep stops if either side can hold a map at all, or, for the converse, if
		// an array either side can hold could go where a shape the other side can hold is (shape_fit.go).
		if l.mentionsHeld(value, []ir.Type{ir.Map}, 0, map[*checker.Type]bool{}) || l.mentionsHeld(target, []ir.Type{ir.Map}, 0, map[*checker.Type]bool{}) {
			return containerFitStop
		}
		if stop := l.shapeFitReached(value, target); stop != "" {
			return stop
		}
		return l.shapeFitReached(target, value)
	}
	if (l.isArrayType(value) || checker.IsTupleType(value)) && (l.isArrayType(target) || checker.IsTupleType(target)) {
		values, targets := l.typeArguments(value), l.typeArguments(target)
		if l.isArrayType(target) && len(targets) == 1 {
			// Every element of a tuple goes where the array's element goes.
			for _, element := range values {
				if stop := l.containerPairGoes(element, targets[0], false, depth+1, visited); stop != "" {
					return stop
				}
			}
			return ""
		}
		return l.pairwiseContainerPairGoes(values, targets, depth, visited)
	}
	if value.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		// An intersection carries no TypeFlagsObject, but its fields and signatures read as an object's
		// do, so it's walked the same way: { m: ReadonlyMap } & { tag: 1 } fed { m: shaped, tag: 1 }, or
		// { m: Fake } & { tag: 1 } seen as { readonly m: ReadonlyMap }, both lowered and crashed natively
		// on b4478772 (#efx3yb9, Compiler's must). containerWhereOtherGoes walks only when a side can
		// hold a map or an array, so this costs nothing elsewhere.
		return ""
	}
	if value.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 && value.Target() != nil && value.Target() == target.Target() {
		// Two instantiations of one generic (Promise<Map> as Promise<ReadonlyMap>, Box<A> as Box<B>) hold
		// what their type arguments say, and an argument may flow either way (a parameter of T), so each
		// pair is judged both ways rather than walking members whose generic methods never end.
		values, targets := l.typeArguments(value), l.typeArguments(target)
		if stop := l.pairwiseContainerPairGoes(values, targets, depth, visited); stop != "" {
			return stop
		}
		return l.pairwiseContainerPairGoes(targets, values, depth, visited)
	}
	valueSignatures := l.checker.GetSignaturesOfType(value, checker.SignatureKindCall)
	targetSignatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(valueSignatures) == 1 && len(targetSignatures) == 1 {
		// A function: what it returns goes where the target's result goes, and what the target is
		// called with goes to its parameters.
		if stop := l.containerPairGoes(l.checker.GetReturnTypeOfSignature(valueSignatures[0]), l.checker.GetReturnTypeOfSignature(targetSignatures[0]), false, depth+1, visited); stop != "" {
			return stop
		}
		valueParameters, targetParameters := valueSignatures[0].Parameters(), targetSignatures[0].Parameters()
		if l.censusNeverRestSignature(targetSignatures[0]) {
			targetParameters = nil
		}
		for index := range valueParameters {
			if index < len(targetParameters) {
				if stop := l.containerPairGoes(l.checker.GetTypeOfSymbol(targetParameters[index]), l.checker.GetTypeOfSymbol(valueParameters[index]), false, depth+1, visited); stop != "" {
					return stop
				}
			}
		}
		return ""
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if inside := l.checker.GetPropertyOfType(value, field.Name); inside != nil {
			if stop := l.containerPairGoes(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(field), false, depth+1, visited); stop != "" {
				return stop
			}
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(target) {
		// A dictionary's values: the value's own dictionary values, and each field it has.
		for _, own := range l.checker.GetIndexInfosOfType(value) {
			if own.KeyType() == info.KeyType() {
				if stop := l.containerPairGoes(own.ValueType(), info.ValueType(), false, depth+1, visited); stop != "" {
					return stop
				}
			}
		}
		for _, field := range l.checker.GetPropertiesOfType(value) {
			if field.Flags&ast.SymbolFlagsMethod == 0 {
				if stop := l.containerPairGoes(l.checker.GetTypeOfSymbol(field), info.ValueType(), false, depth+1, visited); stop != "" {
					return stop
				}
			}
		}
	}
	return ""
}

func (l *lowering) pairwiseContainerPairGoes(values []*checker.Type, targets []*checker.Type, depth int, visited map[[2]*checker.Type]bool) string {
	for index := range values {
		if index < len(targets) {
			if stop := l.containerPairGoes(values[index], targets[index], false, depth+1, visited); stop != "" {
				return stop
			}
		}
	}
	return ""
}
