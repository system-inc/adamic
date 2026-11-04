package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A real Map, Set or array is held as itself (ir.Map, ir.Array), but tsc lets it go anywhere its members
// fit, so a structural object type, held as an object or, where a primitive would fit it too, as a
// union, can be handed one. On main e4c3a163 such a value lowered (#nw6d7ws). `useShape(new Set(['k']))`
// with `v: { has(x: string): boolean }` and a Map seen as `interface Lookup { get(key: string): number |
// undefined }` exited 70 on the JavaScript backend (a Map or Set method called on undefined), and an array
// seen as `{ indexOf(x: number): number }` did too. Natively clang refused the C that put an adamic_map*
// or adamic_array* where an adamic_object* goes, and through a union-held shape (`{ indexOf(x: string):
// number }`, `{ readonly size: number } | string`) the native build ran and crashed (exit 139).
const (
	shapeFitMapStop   = "a Map or Set going where a plain object type is held"
	shapeFitArrayStop = "an array going where a plain object type is held"
)

// shapeFitServed is every member a container's value is read by correctly through a structural slot, keyed
// by how the value is held and how the slot is: one table, a pin per entry (Adamic's ruling on #nw6d7ws).
// An array going where `{ readonly length: number }` is held, a union since a string fits it too, reads
// its length on every backend. Through a slot held as an object nothing works natively, not even length
// (the C hands an array to an object parameter), so an object-held slot has no entry, and a Map or Set has
// none at all.
var shapeFitServed = map[[2]ir.Type]map[string]bool{
	{ir.Array, ir.Union}: {"length": true},
}

// shapeFitProtocols are the library's interfaces a library call takes a container through (new Set(xs)'s
// Iterable, Array.from's ArrayLike), read by that call's own lowering and never as a kept object. A
// program's own Iterable parameter stops where it's read (for...of over an object).
var shapeFitProtocols = []string{"Iterable", "IterableIterator", "Iterator", "AsyncIterable", "AsyncIterableIterator", "ArrayLike", "PromiseLike"}

// containerWhereShapeGoes is the stop for a value of type value going where target is expected that
// carries a value held as a map or an array to where a structural object type is held, reading it by a
// member shapeFitServed doesn't serve, or "" if it carries none. It looks where tupleWhereArrayGoes does:
// directly, as an array's elements, as a Map's keys or values, as a field of an object or a value of a
// dictionary, or as a function's result, or the other way round for a function's parameters. It's keyed
// on representation alone, never on the syntax that made the value. A union value stops if any member
// does; a union slot stops only if every member that takes the value does. It walks four positions deep
// with a seen set: a library generic's signatures make a new instantiation at every step.
func (l *lowering) containerWhereShapeGoes(value *checker.Type, target *checker.Type, depth int, visited map[[2]*checker.Type]bool) string {
	if value == nil || target == nil || value == target || depth > 4 || visited[[2]*checker.Type{value, target}] {
		return ""
	}
	visited[[2]*checker.Type{value, target}] = true
	if value.Flags()&checker.TypeFlagsUnion != 0 {
		// Any member that may be there.
		for _, member := range value.Types() {
			if stop := l.containerWhereShapeGoes(member, target, depth+1, visited); stop != "" {
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
			if stop = l.containerWhereShapeGoes(value, member, depth+1, visited); stop == "" {
				return ""
			}
		}
		return stop
	}
	if value.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsNever) != 0 {
		return ""
	}
	held, valueKnown := l.representation(value)
	slot, slotKnown := l.representation(target)
	if valueKnown && slotKnown && (held == ir.Map || held == ir.Array) && (slot == ir.Object || slot == ir.Union) && target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0 && !l.checker.IsArrayType(target) && !checker.IsTupleType(target) && !l.isLibraryType(target, shapeFitProtocols...) {
		// A structural slot: every member it declares must be one the value is served by. A union-held
		// slot declaring none ({}, object) reads nothing of it, but an object-held one is handed the
		// container as an object all the same.
		members := l.checker.GetPropertiesOfType(target)
		if slot == ir.Union && len(members) == 0 {
			return ""
		}
		stop := shapeFitArrayStop
		if held == ir.Map {
			stop = shapeFitMapStop
		}
		if len(members) == 0 {
			return stop
		}
		for _, member := range members {
			if !shapeFitServed[[2]ir.Type{held, slot}][member.Name] {
				return stop
			}
		}
		return ""
	}
	if valueKnown && held == ir.Map && l.isLibraryType(value, "Map", "ReadonlyMap", "Set", "ReadonlySet") && l.isLibraryType(target, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		// A map's keys and values go where the slot's go.
		return l.pairwiseContainerWhereShapeGoes(l.typeArguments(value), l.typeArguments(target), depth, visited)
	}
	if (l.checker.IsArrayType(value) || checker.IsTupleType(value)) && (l.checker.IsArrayType(target) || checker.IsTupleType(target)) {
		values, targets := l.typeArguments(value), l.typeArguments(target)
		if l.checker.IsArrayType(target) && len(targets) == 1 {
			// Every element of a tuple goes where the array's element goes.
			for _, element := range values {
				if stop := l.containerWhereShapeGoes(element, targets[0], depth+1, visited); stop != "" {
					return stop
				}
			}
			return ""
		}
		return l.pairwiseContainerWhereShapeGoes(values, targets, depth, visited)
	}
	if value.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 {
		return ""
	}
	valueSignatures := l.checker.GetSignaturesOfType(value, checker.SignatureKindCall)
	targetSignatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(valueSignatures) == 1 && len(targetSignatures) == 1 {
		// A function: what it returns goes where the target's result goes, and what the target is
		// called with goes to its parameters.
		if stop := l.containerWhereShapeGoes(l.checker.GetReturnTypeOfSignature(valueSignatures[0]), l.checker.GetReturnTypeOfSignature(targetSignatures[0]), depth+1, visited); stop != "" {
			return stop
		}
		valueParameters, targetParameters := valueSignatures[0].Parameters(), targetSignatures[0].Parameters()
		if l.censusNeverRestSignature(targetSignatures[0]) {
			targetParameters = nil
		}
		for index := range valueParameters {
			if index < len(targetParameters) {
				if stop := l.containerWhereShapeGoes(l.checker.GetTypeOfSymbol(targetParameters[index]), l.checker.GetTypeOfSymbol(valueParameters[index]), depth+1, visited); stop != "" {
					return stop
				}
			}
		}
		return ""
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if inside := l.checker.GetPropertyOfType(value, field.Name); inside != nil {
			if stop := l.containerWhereShapeGoes(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(field), depth+1, visited); stop != "" {
				return stop
			}
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(target) {
		// A dictionary's values: the value's own dictionary values, and each field it has.
		for _, own := range l.checker.GetIndexInfosOfType(value) {
			if own.KeyType() == info.KeyType() {
				if stop := l.containerWhereShapeGoes(own.ValueType(), info.ValueType(), depth+1, visited); stop != "" {
					return stop
				}
			}
		}
		for _, field := range l.checker.GetPropertiesOfType(value) {
			if field.Flags&ast.SymbolFlagsMethod == 0 {
				if stop := l.containerWhereShapeGoes(l.checker.GetTypeOfSymbol(field), info.ValueType(), depth+1, visited); stop != "" {
					return stop
				}
			}
		}
	}
	return ""
}

func (l *lowering) pairwiseContainerWhereShapeGoes(values []*checker.Type, targets []*checker.Type, depth int, visited map[[2]*checker.Type]bool) string {
	for index := range values {
		if index < len(targets) {
			if stop := l.containerWhereShapeGoes(values[index], targets[index], depth+1, visited); stop != "" {
				return stop
			}
		}
	}
	return ""
}
