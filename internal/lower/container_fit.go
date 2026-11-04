package lower

import (
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
// implementing ReadonlyMap five fields down, which crashed natively and printed false on JavaScript).
const containerFitDepth = 4

// containerMentionDepth bounds mentionsMap. Reaching it counts as a mention, so it fails closed too.
const containerMentionDepth = 16

// containerWhereOtherGoes reports whether a value of type value, going where target is expected, carries
// a value held other than as a map to where a map is held: directly, as an array's elements, as a Map's
// keys or values, as a field of an object or a value of a dictionary, or as a function's result, or the
// other way round for a function's parameters. It's keyed on representation alone, so it asks nothing
// of the syntax that made the value (Adamic's ruling on #efx3yb9). A value of any, never or a type
// representation can't read is some other check's to stop, and isn't judged here. When neither side can
// hold a map anywhere (mentionsMap), no position can carry one, and the pair isn't walked at all: that's
// what keeps a Promise's then, which makes a new instantiation at every step, out of the walk.
func (l *lowering) containerWhereOtherGoes(value *checker.Type, target *checker.Type) bool {
	if !l.mentionsMap(value, 0, map[*checker.Type]bool{}) && !l.mentionsMap(target, 0, map[*checker.Type]bool{}) {
		return false
	}
	return l.containerPairGoes(value, target, 0, map[[2]*checker.Type]bool{})
}

// mentionsMap reports whether a value of type proven can hold a map anywhere inside: itself, a type
// argument, a field, a dictionary's values, or a parameter or result of its call and construct signatures,
// read from their declarations so a generic method's own type parameters stay the same types at every
// step. A library type holds only what its type arguments say (a Promise<number> holds no map), and a type
// parameter what its constraint does. Each type is read once, and past containerMentionDepth it answers yes.
func (l *lowering) mentionsMap(proven *checker.Type, depth int, seen map[*checker.Type]bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	if depth > containerMentionDepth {
		return true
	}
	seen[proven] = true
	if held, known := l.representation(proven); known && held == ir.Map {
		return true
	}
	flags := proven.Flags()
	if flags&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, member := range proven.Types() {
			if l.mentionsMap(member, depth+1, seen) {
				return true
			}
		}
		return false
	}
	if flags&checker.TypeFlagsTypeParameter != 0 {
		return l.mentionsMap(l.checker.GetBaseConstraintOfType(proven), depth+1, seen)
	}
	if flags&checker.TypeFlagsObject == 0 {
		return false
	}
	for _, argument := range l.typeArguments(proven) {
		if l.mentionsMap(argument, depth+1, seen) {
			return true
		}
	}
	if l.librarySymbol(proven.Symbol()) {
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(proven) {
		if l.mentionsMap(l.checker.GetTypeOfSymbol(field), depth+1, seen) {
			return true
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(proven) {
		if l.mentionsMap(info.ValueType(), depth+1, seen) {
			return true
		}
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		for _, signature := range l.checker.GetSignaturesOfType(proven, kind) {
			if declaration := signature.Declaration(); declaration != nil && ast.IsFunctionLike(declaration) {
				signature = l.checker.GetSignatureFromDeclaration(declaration)
			}
			for _, parameter := range signature.Parameters() {
				if l.mentionsMap(l.checker.GetTypeOfSymbol(parameter), depth+1, seen) {
					return true
				}
			}
			if l.mentionsMap(l.checker.GetReturnTypeOfSignature(signature), depth+1, seen) {
				return true
			}
		}
	}
	return false
}

// containerPairGoes is containerWhereOtherGoes' walk of one pair, depth positions in.
func (l *lowering) containerPairGoes(value *checker.Type, target *checker.Type, depth int, visited map[[2]*checker.Type]bool) bool {
	if value == nil || target == nil || value == target || visited[[2]*checker.Type{value, target}] {
		return false
	}
	visited[[2]*checker.Type{value, target}] = true
	if value.Flags()&checker.TypeFlagsUnion != 0 {
		// Any member that may be there.
		for _, member := range value.Types() {
			if l.containerPairGoes(member, target, depth+1, visited) {
				return true
			}
		}
		return false
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		// Into a union: fine if a member that takes the value takes it as it is. A member it isn't
		// assignable to (undefined, for an object) never holds it.
		taken := false
		for _, member := range target.Types() {
			if !l.checker.IsTypeAssignableTo(value, member) {
				continue
			}
			taken = true
			if !l.containerPairGoes(value, member, depth+1, visited) {
				return false
			}
		}
		return taken
	}
	if value.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsNever) != 0 {
		return false
	}
	if slot, known := l.representation(target); known && slot == ir.Map {
		held, known := l.representation(value)
		if known && held != ir.Map {
			return true
		}
		if !known {
			// An intersection Adamic holds no way at all (Fake & { tag: 1 }, a class or a collection
			// inside) can't be proven a map, so it fails closed. Any other unread type is some other
			// check's to stop.
			return value.Flags()&checker.TypeFlagsIntersection != 0
		}
		if !l.isLibraryType(value, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
			return false
		}
		// A map's keys and values go where the slot's go.
		return l.pairwiseContainerPairGoes(l.typeArguments(value), l.typeArguments(target), depth, visited)
	}
	if depth > containerFitDepth {
		// A map slot is judged above at any depth, since that leaf costs nothing. Anything else this deep
		// stops if either side can hold a map at all.
		return l.mentionsMap(value, 0, map[*checker.Type]bool{}) || l.mentionsMap(target, 0, map[*checker.Type]bool{})
	}
	if (l.checker.IsArrayType(value) || checker.IsTupleType(value)) && (l.checker.IsArrayType(target) || checker.IsTupleType(target)) {
		values, targets := l.typeArguments(value), l.typeArguments(target)
		if l.checker.IsArrayType(target) && len(targets) == 1 {
			// Every element of a tuple goes where the array's element goes.
			for _, element := range values {
				if l.containerPairGoes(element, targets[0], depth+1, visited) {
					return true
				}
			}
			return false
		}
		return l.pairwiseContainerPairGoes(values, targets, depth, visited)
	}
	if value.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		// An intersection carries no TypeFlagsObject, but its fields and signatures read as an object's
		// do, so it's walked the same way: { m: ReadonlyMap } & { tag: 1 } fed { m: shaped, tag: 1 }, or
		// { m: Fake } & { tag: 1 } seen as { readonly m: ReadonlyMap }, both lowered and crashed natively
		// on b4478772 (#efx3yb9, Compiler's must). containerWhereOtherGoes walks only when a side can
		// hold a map, so this costs nothing elsewhere.
		return false
	}
	if value.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 && value.Target() != nil && value.Target() == target.Target() {
		// Two instantiations of one generic (Promise<Map> as Promise<ReadonlyMap>, Box<A> as Box<B>) hold
		// what their type arguments say, and an argument may flow either way (a parameter of T), so each
		// pair is judged both ways rather than walking members whose generic methods never end.
		values, targets := l.typeArguments(value), l.typeArguments(target)
		return l.pairwiseContainerPairGoes(values, targets, depth, visited) || l.pairwiseContainerPairGoes(targets, values, depth, visited)
	}
	valueSignatures := l.checker.GetSignaturesOfType(value, checker.SignatureKindCall)
	targetSignatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(valueSignatures) == 1 && len(targetSignatures) == 1 {
		// A function: what it returns goes where the target's result goes, and what the target is
		// called with goes to its parameters.
		if l.containerPairGoes(l.checker.GetReturnTypeOfSignature(valueSignatures[0]), l.checker.GetReturnTypeOfSignature(targetSignatures[0]), depth+1, visited) {
			return true
		}
		valueParameters, targetParameters := valueSignatures[0].Parameters(), targetSignatures[0].Parameters()
		if l.censusNeverRestSignature(targetSignatures[0]) {
			targetParameters = nil
		}
		for index := range valueParameters {
			if index < len(targetParameters) && l.containerPairGoes(l.checker.GetTypeOfSymbol(targetParameters[index]), l.checker.GetTypeOfSymbol(valueParameters[index]), depth+1, visited) {
				return true
			}
		}
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if inside := l.checker.GetPropertyOfType(value, field.Name); inside != nil && l.containerPairGoes(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(field), depth+1, visited) {
			return true
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(target) {
		// A dictionary's values: the value's own dictionary values, and each field it has.
		for _, own := range l.checker.GetIndexInfosOfType(value) {
			if own.KeyType() == info.KeyType() && l.containerPairGoes(own.ValueType(), info.ValueType(), depth+1, visited) {
				return true
			}
		}
		for _, field := range l.checker.GetPropertiesOfType(value) {
			if field.Flags&ast.SymbolFlagsMethod == 0 && l.containerPairGoes(l.checker.GetTypeOfSymbol(field), info.ValueType(), depth+1, visited) {
				return true
			}
		}
	}
	return false
}

func (l *lowering) pairwiseContainerPairGoes(values []*checker.Type, targets []*checker.Type, depth int, visited map[[2]*checker.Type]bool) bool {
	for index := range values {
		if index < len(targets) && l.containerPairGoes(values[index], targets[index], depth+1, visited) {
			return true
		}
	}
	return false
}
