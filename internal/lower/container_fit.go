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

// containerWhereOtherGoes reports whether a value of type value, going where target is expected, carries
// a value held other than as a map to where a map is held: directly, as an array's elements, as a Map's
// keys or values, as a field of an object or a value of a dictionary, or as a function's result, or the
// other way round for a function's parameters. It's keyed on representation alone, so it asks nothing
// of the syntax that made the value (Adamic's ruling on #efx3yb9). A value of any, never or a type
// representation can't read is some other check's to stop, and isn't judged here. It walks as deep as
// tupleWhereArrayGoes does, four positions, and no pair twice: a library generic's signatures (Promise's
// then) make a new instantiation at every step, so a seen set alone never ends the walk.
func (l *lowering) containerWhereOtherGoes(value *checker.Type, target *checker.Type, depth int, visited map[[2]*checker.Type]bool) bool {
	if value == nil || target == nil || value == target || depth > 4 || visited[[2]*checker.Type{value, target}] {
		return false
	}
	visited[[2]*checker.Type{value, target}] = true
	if value.Flags()&checker.TypeFlagsUnion != 0 {
		// Any member that may be there.
		for _, member := range value.Types() {
			if l.containerWhereOtherGoes(member, target, depth+1, visited) {
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
			if !l.containerWhereOtherGoes(value, member, depth+1, visited) {
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
		if !known || !l.isLibraryType(value, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
			return false
		}
		// A map's keys and values go where the slot's go.
		return l.pairwiseContainerWhereOtherGoes(l.typeArguments(value), l.typeArguments(target), depth, visited)
	}
	if (l.checker.IsArrayType(value) || checker.IsTupleType(value)) && (l.checker.IsArrayType(target) || checker.IsTupleType(target)) {
		values, targets := l.typeArguments(value), l.typeArguments(target)
		if l.checker.IsArrayType(target) && len(targets) == 1 {
			// Every element of a tuple goes where the array's element goes.
			for _, element := range values {
				if l.containerWhereOtherGoes(element, targets[0], depth+1, visited) {
					return true
				}
			}
			return false
		}
		return l.pairwiseContainerWhereOtherGoes(values, targets, depth, visited)
	}
	if value.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	valueSignatures := l.checker.GetSignaturesOfType(value, checker.SignatureKindCall)
	targetSignatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(valueSignatures) == 1 && len(targetSignatures) == 1 {
		// A function: what it returns goes where the target's result goes, and what the target is
		// called with goes to its parameters.
		if l.containerWhereOtherGoes(l.checker.GetReturnTypeOfSignature(valueSignatures[0]), l.checker.GetReturnTypeOfSignature(targetSignatures[0]), depth+1, visited) {
			return true
		}
		valueParameters, targetParameters := valueSignatures[0].Parameters(), targetSignatures[0].Parameters()
		if l.censusNeverRestSignature(targetSignatures[0]) {
			targetParameters = nil
		}
		for index := range valueParameters {
			if index < len(targetParameters) && l.containerWhereOtherGoes(l.checker.GetTypeOfSymbol(targetParameters[index]), l.checker.GetTypeOfSymbol(valueParameters[index]), depth+1, visited) {
				return true
			}
		}
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if inside := l.checker.GetPropertyOfType(value, field.Name); inside != nil && l.containerWhereOtherGoes(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(field), depth+1, visited) {
			return true
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(target) {
		// A dictionary's values: the value's own dictionary values, and each field it has.
		for _, own := range l.checker.GetIndexInfosOfType(value) {
			if own.KeyType() == info.KeyType() && l.containerWhereOtherGoes(own.ValueType(), info.ValueType(), depth+1, visited) {
				return true
			}
		}
		for _, field := range l.checker.GetPropertiesOfType(value) {
			if field.Flags&ast.SymbolFlagsMethod == 0 && l.containerWhereOtherGoes(l.checker.GetTypeOfSymbol(field), info.ValueType(), depth+1, visited) {
				return true
			}
		}
	}
	return false
}

func (l *lowering) pairwiseContainerWhereOtherGoes(values []*checker.Type, targets []*checker.Type, depth int, visited map[[2]*checker.Type]bool) bool {
	for index := range values {
		if index < len(targets) && l.containerWhereOtherGoes(values[index], targets[index], depth+1, visited) {
			return true
		}
	}
	return false
}
