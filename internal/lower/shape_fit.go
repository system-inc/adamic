package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A real Map, Set or array is held as itself (ir.Map, ir.Array), but tsc lets it go anywhere its members
// fit, so a structural object type, held as an object or, where a primitive would fit it too, as a
// union, can be handed one. On main e4c3a163 such a value lowered (#nw6d7ws). `useShape(new Set(['k']))`
// with `v: { has(x: string): boolean }` and a Map seen as `interface Lookup { get(key: string): number |
// undefined }` exited 70 on the JavaScript backend (a Map or Set method called on undefined), and an array
// seen as `{ indexOf(x: number): number }` did too. Natively clang refused the C that put an adamic_map*
// or adamic_array* where an adamic_object* goes, and through a union-held shape (`{ indexOf(x: string):
// number }`, `{ readonly size: number } | string`) or a Map.forEach callback whose parameter is the shape,
// the native build ran and crashed (exit 139). The other direction, a value held other than as a map
// where a map is held, is container_fit.go's, which walks both.
const (
	shapeFitMapStop   = "a Map or Set going where a plain object type is held"
	shapeFitArrayStop = "an array going where a plain object type is held"
)

// shapeFitServed is every member a container's value is read by correctly through a structural slot, keyed
// by how the value is held and how the slot is: one table, a pin per entry (Adamic's ruling on #nw6d7ws).
// An array going where `{ readonly length: number }` is held, a union since a string fits it too, reads
// its length on every backend (TestContainerIsNotYetWhereAShapeIsHeld's length controls). Through a slot
// held as an object nothing works natively, not even length (the C hands an array to an object
// parameter), so an object-held slot has no entry, and a Map or Set has none at all.
var shapeFitServed = map[[2]ir.Type]map[string]bool{
	{ir.Array, ir.Union}: {"length": true},
}

// shapeFitProtocols are the library's and the prelude's interfaces a library call takes a container
// through, read by that call's own lowering and never kept as an object: new Set(xs)'s and new Map(xs)'s
// Iterable<T>, and the prelude's ReadonlySetLike<U> that Set's union, intersection, difference and
// predicates take (librarySetMethod takes only a concrete library Set there). They're exempt only as a
// library call's own argument; in a program's own slot they're judged like any other shape, and each has
// a pin showing a container there stops.
var shapeFitProtocols = []string{"Iterable", "ReadonlySetLike"}

// shapeFit judges a value of type value going where target is held when the value is held as a map or
// an array and the slot is a structural object type, held as an object or a union: the stop, or "" when
// every member the type declares is served for that pair. judged is false for any other pair, which the
// walk goes on into. A union-held slot declaring no members ({}) reads nothing of the value and isn't
// judged; an object-held one is handed the container as an object all the same. libraryArgument says the
// value is a library call's own argument, where a shapeFitProtocols slot is that call's.
func (l *lowering) shapeFit(value *checker.Type, target *checker.Type, libraryArgument bool) (stop string, judged bool) {
	held, valueKnown := l.representation(value)
	slot, slotKnown := l.representation(target)
	if !valueKnown || !slotKnown || (held != ir.Map && held != ir.Array) || (slot != ir.Object && slot != ir.Union) {
		return "", false
	}
	if target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || l.isArrayType(target) || checker.IsTupleType(target) {
		return "", false
	}
	if libraryArgument && l.shapeFitProtocol(target) {
		return "", true
	}
	members := l.checker.GetPropertiesOfType(target)
	if slot == ir.Union && len(members) == 0 {
		return "", true
	}
	stop = shapeFitArrayStop
	if held == ir.Map {
		stop = shapeFitMapStop
	}
	if len(members) == 0 {
		return stop, true
	}
	for _, member := range members {
		if !shapeFitServed[[2]ir.Type{held, slot}][member.Name] {
			return stop, true
		}
	}
	return "", true
}

// shapeFitReached is the converse's stop past containerFitDepth, where the walk no longer knows which
// position meets which: a map or an array a value of type value can hold anywhere (eachReached), assignable
// to a structural slot a value of type target can hold anywhere that it can't go to (shapeFit), or ""
// when none is. A mention past containerMentionDepth stops, so it fails closed. It's asked both ways
// round, since a function's parameters carry the other way.
func (l *lowering) shapeFitReached(value *checker.Type, target *checker.Type) string {
	var containers, shapes []*checker.Type
	exhausted := l.eachReached(value, 0, map[*checker.Type]bool{}, func(reached *checker.Type) bool {
		if held, known := l.representation(reached); known && (held == ir.Map || held == ir.Array) {
			containers = append(containers, reached)
		}
		return false
	})
	exhausted = l.eachReached(target, 0, map[*checker.Type]bool{}, func(reached *checker.Type) bool {
		if slot, known := l.representation(reached); known && (slot == ir.Object || slot == ir.Union) {
			shapes = append(shapes, reached)
		}
		return false
	}) || exhausted
	if exhausted {
		// What lies past the bound is unread, so it may hold a container and a shape that meet.
		for _, container := range containers {
			if held, _ := l.representation(container); held == ir.Map {
				return shapeFitMapStop
			}
		}
		return shapeFitArrayStop
	}
	for _, container := range containers {
		for _, shape := range shapes {
			if !l.checker.IsTypeAssignableTo(container, shape) {
				continue
			}
			if stop, judged := l.shapeFit(container, shape, false); judged && stop != "" {
				return stop
			}
		}
	}
	return ""
}

// shapeFitProtocol is whether a type is an instantiation of one of shapeFitProtocols as the library or
// the prelude declares it, never a program's own interface of the same name.
func (l *lowering) shapeFitProtocol(proven *checker.Type) bool {
	symbol := proven.Symbol()
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsReference == 0 || symbol == nil || len(symbol.Declarations) == 0 || !slices.Contains(shapeFitProtocols, symbol.Name) {
		return false
	}
	file := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return load.IsLibrary(file) || load.IsPrelude(file)
}
