package lower

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A slot typed Iterable<T> or IterableIterator<T> is held as an object, and nothing in the type says
// which object: an array, a generator, a string, a Map itself or a program's own iterator all fit it,
// and each steps differently. tsc's mapIterator and mapDefinedIterator take one and are handed a
// collection's iterator (`jsxOptionMap.entries()`, `fileInfos.keys()`), so a for-of over such a slot
// lowers as that iterator's own loop (libraryForOfIterator), and is proven after lowering, when every
// call that can reach the slot is in, to be handed nothing else (#qk8rztp S2b-1b, Lower's ruling on q1).
//
// The proof asks the allocation graph the view passes use (ReachingAllocations) which values reach the
// iterated expression, with an entry hook that stops the flow at the first value that isn't a road
// (a read, a call's result, a conditional's arm, a box, a narrowing, a cast). A collection's iterator,
// ir.CollectionIterator, is the one listed kind. Anything else the flow ends at is an origin the proof
// can't name, and the view is NotYet: a field, a literal, the parameter of a function used as a value
// or of an object literal's method (whose callers the graph doesn't see), a call with no known target.
// The hook stops at a field rather than projecting through it, so no projection depth bound is ever
// reached. Arrays and Maps never get this far: going where a structural type is held, they stop at the
// call (shape_fit.go).
//
// Every origin must hold its step's value as the view reads it, decided here and never converted in
// the step: number for a Map's values() read as Iterable<number>, and for an entries() pair, its key
// and its value as the tuple's two elements are held.

// iterableView is one for-of forOf lowered as a view: where it stops if unproven, the value it steps,
// and how it reads each step.
type iterableView struct {
	where    *ast.Node
	iterated ir.Expression
	held     iterableElement
}

// iterableElement is how a step's value is held: element, and when it is a [key, value] pair, how the
// pair holds each half. A view's pair says how its T, a two-element tuple, holds them. An origin's
// says an entries() iterator makes the pair each step, holding the collection's key and value as the
// collection does.
type iterableElement struct {
	element    ir.Type
	pair       bool
	key, value ir.Type
}

// iterableOrigin is a value the proof's flow stopped at: a collection's iterator, held as it says, or
// anything else, by its kind.
type iterableOrigin struct {
	name       string
	collection bool
	held       iterableElement
}

// iterableViewElement says whether a for-of's iterated value is typed the library's Iterable<T> or
// IterableIterator<T>, with T held in one boxed slot, and how T is held. TReturn and TNext don't
// matter to a for-of: it drops the return value and sends nothing.
func (l *lowering) iterableViewElement(iterated *ast.Node) (iterableElement, bool) {
	proven := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(iterated)))
	if !l.isLibraryType(proven, "Iterable", "IterableIterator") {
		return iterableElement{}, false
	}
	arguments := l.typeArguments(proven)
	if len(arguments) == 0 {
		return iterableElement{}, false
	}
	yielded := l.concrete(arguments[0])
	element, known := l.representation(yielded)
	if !known || !boxedSlot(element) {
		return iterableElement{}, false
	}
	held := iterableElement{element: element}
	if checker.IsTupleType(yielded) {
		if halves := l.typeArguments(yielded); len(halves) == 2 {
			key, keyKnown := l.representation(halves[0])
			value, valueKnown := l.representation(halves[1])
			if keyKnown && valueKnown {
				held.pair, held.key, held.value = true, key, value
			}
		}
	}
	return held, true
}

// collectionOrigin is how a collection's iterator holds what it steps: runtime/map_set.c's
// collection_next gives a key for keys() and for any Set, a value for a Map's values(), and for
// entries() a pair it makes of the key and the value.
func collectionOrigin(iterator ir.CollectionIterator) iterableOrigin {
	collection := "a Map's "
	if iterator.Set {
		collection = "a Set's "
	}
	origin := iterableOrigin{name: collection + iterator.Part + "()", collection: true}
	switch {
	case iterator.Part == "entries":
		origin.held = iterableElement{element: ir.Object, pair: true, key: iterator.Key, value: iterator.Value}
	case iterator.Part == "keys" || iterator.Set:
		origin.held = iterableElement{element: iterator.Key}
	default:
		origin.held = iterableElement{element: iterator.Value}
	}
	return origin
}

// readsAs says whether a view reading its steps as view reads an origin's steps, held as held, as
// they are: the same element, and for a pair the origin makes, the same key and value the view's
// tuple reads.
func (held iterableElement) readsAs(view iterableElement) bool {
	if held.element != view.element {
		return false
	}
	return !held.pair || view.pair && held.key == view.key && held.value == view.value
}

// checkIterableViews proves every view forOf lowered, or stops at the first it can't.
func (l *lowering) checkIterableViews() error {
	if len(l.iterableViews) == 0 {
		return nil
	}
	var origins []iterableOrigin
	numbered := map[iterableOrigin]int{}
	// The sites the hook gives are positive, so no allocation's (which are negative) is one of them.
	site := func(origin iterableOrigin) int {
		if number, seen := numbered[origin]; seen {
			return number
		}
		origins = append(origins, origin)
		numbered[origin] = len(origins)
		return len(origins)
	}
	graph := newAllocationFlowGraph(l.result)
	graph.entry = func(value ir.Expression) int {
		switch value := value.(type) {
		case ir.CollectionIterator:
			return site(collectionOrigin(value))
		case ir.Read, ir.Call, ir.CallClosure, ir.Conditional, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast:
			// A road: the graph follows it to what it holds, or says why it can't.
			return 0
		}
		return site(iterableOrigin{name: "a value made by " + strings.TrimPrefix(fmt.Sprintf("%T", value), "ir.")})
	}
	for _, view := range l.iterableViews {
		reaches := graph.ReachingAllocations(view.iterated)
		unproven := slices.Clone(reaches.Reasons)
		var differently []string
		for _, number := range reaches.Sites {
			if number < 1 || number > len(origins) {
				unproven = append(unproven, "an allocation")
				continue
			}
			origin := origins[number-1]
			if !origin.collection {
				unproven = append(unproven, origin.name)
				continue
			}
			if !origin.held.readsAs(view.held) {
				differently = append(differently, origin.name+" held as "+iterableHeldName(origin.held))
			}
		}
		if reaches.Unknown || len(unproven) != 0 {
			return l.notYet(view.where, "a view over an iterable of unproven origin ("+strings.Join(unproven, "; ")+")")
		}
		if len(differently) != 0 {
			return l.notYet(view.where, "a view over an iterable whose origins yield values held differently ("+strings.Join(differently, "; ")+", read as "+iterableHeldName(view.held)+")")
		}
	}
	return nil
}

// iterableHeldName names how a step's value is held, for a stop.
func iterableHeldName(held iterableElement) string {
	if held.pair {
		return "a pair of " + iterableTypeName(held.key) + " and " + iterableTypeName(held.value)
	}
	return iterableTypeName(held.element)
}

func iterableTypeName(of ir.Type) string {
	switch of {
	case ir.Object:
		return "an object"
	case ir.Array:
		return "an array"
	case ir.Map:
		return "a map"
	case ir.Closure:
		return "a function"
	case ir.Record:
		return "a record"
	case ir.Weak:
		return "a weak reference"
	}
	return typeName(of)
}
