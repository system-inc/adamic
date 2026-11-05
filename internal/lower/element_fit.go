package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An array, a map or a set stores its elements as they're held where it's made, and every read and write
// through a slot works them as the slot's element type holds them, with nothing between to convert. So a
// container whose elements are stored one way can't go where they're held another (#b8fwyfx): [0, 1]
// stores numbers, and written into "string" | (number | string)[], an object literal's field, a
// declaration, an assignment, an argument or a return, it was pushed a boxed string and read back as
// boxes, which printed undefined and segfaulted natively where Node printed 0, 1, c. A copy
// (kinds.slice()) and a new Map did the same in a plain (number | string)[] or Map<string, number | string>
// slot, read only too. invariant-mutable judges whether a view can be written through; this judges
// whether the container can be held there at all, so it's asked of every value at its slot, readonly
// ones included.
//
// The storage is decided where the value is made, and it's what every slot that holds the value reads,
// as #fsb6n82 found for a literal's tuple. A union slot is any member the value is assignable to, and
// narrowing can pick any of them, so every such container member must read the elements as they're
// stored, not one.

// elementsHeldOtherwise is the stop for the value at node going into slot when a container in it stores
// its elements otherwise than a container slot reads them, or "". It's the one judgment, made where
// expression hands every lowered value to its slot (TestElementsHeldOtherwiseIsJudgedOnce).
func (l *lowering) elementsHeldOtherwise(node *ast.Node, slot *checker.Type) string {
	node = ast.SkipParentheses(node)
	own := l.checker.GetTypeAtLocation(node)
	var built ir.Type
	literal := node.Kind == ast.KindArrayLiteralExpression
	if literal {
		if len(node.AsArrayLiteralExpression().Elements.Nodes) == 0 || l.tupleType(node) != nil {
			// [] stores nothing to be read otherwise, and a tuple is stored as its own type's slots
			// (tupleLiteral), which the reads of it follow (#fsb6n82).
			return ""
		}
		// A literal's elements are stored as elementType builds them: the slot's, where that's an array
		// whose elements are boxed, or the literal's own. What's inside them is judged at each element,
		// a value at its own slot, so only the literal's own elements are judged here.
		element, err := l.elementType(node)
		if err != nil {
			// Lowering the literal met this already, and its stop is the one a program hears.
			return ""
		}
		built = element
	}
	stored, read, found := l.elementsSeenAs(own, slot, built, literal, map[[2]*checker.Type]bool{})
	if !found {
		return ""
	}
	return "a " + l.checker.TypeToString(own) + " where " + l.checker.TypeToString(slot) + " goes, its elements stored as " + typeName(stored) + " and read there as " + typeName(read) + " (build it as the slot's type where it's made, or copy its elements into one)"
}

// elementsSeenAs finds a container in a value of type stored, seen as type viewed, whose elements (an
// array's, a map's keys and values, a set's) are stored otherwise than a container member of viewed
// that takes it reads them, all the way down through elements. built, when it's set, is how the top
// array stores its elements, and shallow leaves what's inside them to their own sites.
func (l *lowering) elementsSeenAs(stored *checker.Type, viewed *checker.Type, built ir.Type, shallow bool, visited map[[2]*checker.Type]bool) (ir.Type, ir.Type, bool) {
	if stored == nil || viewed == nil || visited[[2]*checker.Type{stored, viewed}] {
		return 0, 0, false
	}
	visited[[2]*checker.Type{stored, viewed}] = true
	for _, storedMember := range l.heldMembers(stored) {
		for _, inside := range l.containers(storedMember) {
			for _, viewedMember := range l.heldMembers(viewed) {
				if !l.checker.IsTypeAssignableTo(storedMember, viewedMember) {
					continue
				}
				for _, seen := range l.containers(viewedMember) {
					if !l.sameContainer(inside, seen) || checker.IsTupleType(inside) != checker.IsTupleType(seen) {
						// A tuple seen as an array has its own stop (tupleSeenAsArray).
						continue
					}
					storedArguments, viewedArguments := l.typeArguments(inside), l.typeArguments(seen)
					for index := 0; index < len(storedArguments) && index < len(viewedArguments); index++ {
						held, heldKnown := l.kept(storedArguments[index])
						if built != 0 && index == 0 && l.checker.IsArrayType(inside) {
							held, heldKnown = built, true
						}
						read, readKnown := l.kept(viewedArguments[index])
						if heldKnown && readKnown && !elementReadAs(held, read) {
							return held, read, true
						}
						if shallow {
							continue
						}
						if held, read, found := l.elementsSeenAs(storedArguments[index], viewedArguments[index], 0, false, visited); found {
							return held, read, true
						}
					}
				}
			}
		}
	}
	return 0, 0, false
}

// heldMembers is a type's members that can hold a value, a union's but null and undefined, or the type.
func (l *lowering) heldMembers(proven *checker.Type) []*checker.Type {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return []*checker.Type{proven}
	}
	var members []*checker.Type
	for _, member := range proven.Types() {
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			members = append(members, member)
		}
	}
	return members
}

// elementReadAs reports whether an element stored as held reads correctly as read: the same, or a number
// read as number | undefined, which packs undefined into a reserved NaN and every other number as itself.
func elementReadAs(held ir.Type, read ir.Type) bool {
	return held == read || held == ir.Number && read == ir.MaybeNumber
}
