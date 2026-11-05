package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A phantom brand has no run-time content, so a value of a phantom extension (phantomExtension: tsc's
// `interface SortedArray<T> extends Array<T> { " __sortedArrayBrand": undefined; }`) is held as its base
// (#gb0fqaa). Lowering asks whether a type is an array, and what its elements are, through isArrayType
// and elementTypeOfArray, which see a phantom extension as its base; nothing in lowering asks the checker
// directly (TestArrayQuestionsGoThroughUnbranded). The checker's own rules keep seeing the brand: a cast
// is judged on it (castProof, brandMintOf), and so is every observation of a brand member's presence,
// which is refused before any of this runs (phantom_array_presence.go, rulings.md:31).

// unbranded is t's base when t is a phantom extension of an array, followed while that base is one too,
// or t. Only an interface whose own members are all phantom is one: a real member, a method or an own
// index entry keeps t as it is. Two more keep it, each failing closed to today's stop:
//   - A brand over anything but an array (a Map, an object): phantom_array_presence.go refuses the
//     presence observations of array brands only, and a brand is phantom only while those are refused
//     (rulings.md:31).
//   - A brand whose type arguments aren't its base's, as tsc's SortedArray<T> extends Array<T>'s are.
//     Lowering reads an array's elements off its type arguments in many places (a reference's
//     GetTypeArguments, behind an ObjectFlagsReference test), and those read the brand's own: a
//     `NumberBrand extends Array<number>` has none, and a `PairBrand<T> extends Array<[T, T]>` has T
//     where the base has [T, T]. A brand whose arguments are its base's reads the same there as its base.
//   - A brand with a member typed void: such a program fails closed under the void-container rule
//     (rulings.md, Oct 29), so it doesn't gain its base's layout here. The cast is erased all the same.
func (l *lowering) unbranded(t *checker.Type) *checker.Type {
	base := t
	for {
		next, members := l.phantomExtension(base)
		if next == nil {
			break
		}
		for _, member := range members {
			if l.voidBrandMember(base, member) {
				return t
			}
		}
		base = next
	}
	if base == t || !l.checker.IsArrayType(base) || t.ObjectFlags()&checker.ObjectFlagsReference == 0 || !slices.Equal(l.checker.GetTypeArguments(t), l.checker.GetTypeArguments(base)) {
		return t
	}
	return base
}

// voidBrandMember is whether brand's member name is typed void, or a union holding void, as
// phantomBrandMember reads a member's type.
func (l *lowering) voidBrandMember(brand *checker.Type, name string) bool {
	property := l.checker.GetPropertyOfType(brand, name)
	if property == nil {
		return false
	}
	proven := l.checker.GetTypeOfSymbol(property)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		return slices.ContainsFunc(proven.Types(), func(part *checker.Type) bool { return part.Flags()&checker.TypeFlagsVoid != 0 })
	}
	return proven.Flags()&checker.TypeFlagsVoid != 0
}

// isArrayType is the checker's array test (Array or ReadonlyArray) of t held as its base.
func (l *lowering) isArrayType(t *checker.Type) bool {
	return l.checker.IsArrayType(l.unbranded(t))
}

// elementTypeOfArray is the checker's element type of t held as its base.
func (l *lowering) elementTypeOfArray(t *checker.Type) *checker.Type {
	return l.checker.GetElementTypeOfArrayType(l.unbranded(t))
}
