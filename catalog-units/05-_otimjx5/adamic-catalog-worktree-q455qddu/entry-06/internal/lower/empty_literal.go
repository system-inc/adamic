package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// literalArrayElement finds the element destination of a fresh empty literal.
// A union contributes only array members: do not invent slots for a mixed
// array/object context, or turn a tuple into an array.
func (l *lowering) literalArrayElement(proven *checker.Type) *checker.Type {
	if proven == nil {
		return nil
	}
	proven = l.concrete(proven)
	if l.checker.IsArrayType(proven) {
		return l.checker.GetElementTypeOfArrayType(proven)
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	elements := []*checker.Type{}
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		element := l.literalArrayElement(member)
		if element == nil {
			return nil
		}
		elements = append(elements, element)
	}
	if len(elements) == 0 {
		return nil
	}
	// [] satisfies every array member. Choose a concrete member layout, not
	// a union element layout: array unions do not mean arrays of unions.
	for _, element := range elements {
		if held, known := l.kept(element); known && !slotless(held) {
			return element
		}
	}
	return nil
}
