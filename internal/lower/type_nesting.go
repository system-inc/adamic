package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// A generic that instantiates itself deeper at each step (Grow<T> { next?: Grow<Grow<T>> }, or
// G<T> { next?: G<T[]> }, or G<{ v: T }>) is a new type at every level, so a structural walk whose only
// stop is a set of the types it has seen never ends: collectViewedElements ran until Go's stack
// overflowed, and prepareUntaggedStructuralRead after it (#k031ey0). typeNesting counts as tsc's
// isDeeplyNestedType counts: types open on the walk's path that share a recursion identity (a
// reference's target, else an object type's symbol), each with a higher type id than the one before,
// since a deeper instantiation is created after the one it grows from. Past typeNestingLimit of them,
// enter says no, and each walk answers as it does wherever it can't see.
type typeNesting struct {
	open map[any]typeNestingOpen
}

type typeNestingOpen struct {
	count  int
	lastID int
}

// typeNestingLimit is how many ever-newer types of one recursion identity may be open on a walk's path.
const typeNestingLimit = 5

// enter reports whether a walk may go into target, and leave closes it again; the walk calls leave
// when it comes back out, on every path.
func (nesting *typeNesting) enter(target *checker.Type) (ok bool, leave func()) {
	identity := typeRecursionIdentity(target)
	if identity == nil {
		return true, func() {}
	}
	if nesting.open == nil {
		nesting.open = map[any]typeNestingOpen{}
	}
	before := nesting.open[identity]
	if int(target.Id()) <= before.lastID {
		return true, func() {}
	}
	if before.count >= typeNestingLimit {
		return false, func() {}
	}
	nesting.open[identity] = typeNestingOpen{count: before.count + 1, lastID: int(target.Id())}
	return true, func() { nesting.open[identity] = before }
}

// typeRecursionIdentity is what tsc's isDeeplyNestedType calls a type's recursion identity: a
// reference's generic target, or an object type's symbol, which an instantiated anonymous type shares
// with every other instantiation of it, or nil.
func typeRecursionIdentity(proven *checker.Type) any {
	if proven == nil || proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		return proven.Target()
	}
	if symbol := proven.Symbol(); symbol != nil {
		return symbol
	}
	return nil
}
