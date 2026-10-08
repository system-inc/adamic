package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// A union of readonly arrays has one array kind and a union of element promises.
// Selecting an element is lazy; choosing one array alternative would require a
// scan. Mutable alternatives cannot use this erasure because their writes differ.
func (l *lowering) viewReadonlyArrayUnionBase(target *checker.Type, seen map[*checker.Type]bool) *checker.Type {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	var common *checker.Type
	for _, member := range target.Types() {
		base := l.viewArrayBaseSeen(member, seen)
		if base == nil || !l.isLibraryType(base, "ReadonlyArray") {
			return nil
		}
		if common == nil {
			common = base
		}
	}
	return common
}

func (l *lowering) viewArrayUnionElementType(target *checker.Type) *checker.Type {
	if target.Flags()&checker.TypeFlagsUnion == 0 || l.viewArrayBase(target) == nil {
		return nil
	}
	var elements []*checker.Type
	for _, member := range target.Types() {
		elements = append(elements, l.viewArrayElementType(member))
	}
	return l.checker.GetUnionType(elements)
}

// Keep unsupported mutable array unions as named lazy read obligations instead
// of sending an array descriptor through the tagged-object union emitter.
func (l *lowering) viewMutableArrayUnion(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	count, mutable := 0, false
	for _, member := range target.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		base := l.viewArrayBase(member)
		if base == nil {
			return false
		}
		count++
		mutable = mutable || !l.isLibraryType(base, "ReadonlyArray")
	}
	return count > 1 && mutable
}
