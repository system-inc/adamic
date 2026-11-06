package lower

import (
	"errors"
	"strings"
	"testing"
)

// handNumbers fills a Map and a Set of numbers whose values() and keys() feed the probes below.
const handNumbers = "const sizes = new Map<string, number>();\nconst odd = new Set<number>();\nfor (let index = 0; index < 3; index++) { sizes.set(`size ${String(index)}`, index + 7); odd.add(index * 2 + 1); }\n"

// handFirst is a function whose IterableIterator<number> parameter is stepped by hand, its value read
// where done has narrowed the result.
const handFirst = "function first(numbers: IterableIterator<number>): number { const status = numbers.next(); if (status.done) { return -1; } return status.value; }\n"

// A hand-driven next() through an IterableIterator<T> parameter lowers when every origin is a
// collection's iterator (#qk8rztp S2c-2a, Adamic's (a) and (b)): checker.ts:21636's loop, a status
// kept across a later next(), and the `| undefined` form. internal/oracle's hand_next.a holds them
// to Node.
func TestHandDrivenNextLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"one next, value read where done narrows it", handNumbers + handFirst + "console.log(String(first(sizes.values())) + String(first(odd.keys())));"},
		{"checker.ts:21636's loop", handNumbers + "function total(numbers: IterableIterator<number>): number { let sum = 0; for (let status = numbers.next(); !status.done; status = numbers.next()) { sum += status.value; } return sum; }\nconsole.log(String(total(sizes.values())));"},
		{"a status kept across a later next", handNumbers + "function two(numbers: IterableIterator<number>): number { const a = numbers.next(); const b = numbers.next(); return (a.done ? 0 : a.value) + (b.done ? 0 : b.value); }\nconsole.log(String(two(odd.values())));"},
		{"the | undefined form, tested first", handNumbers + "function first(numbers: IterableIterator<number> | undefined): number { if (numbers === undefined) { return -1; } const status = numbers.next(); return status.done ? -2 : status.value; }\nconsole.log(String(first(sizes.values())));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
		})
	}
}

// Every other road stops (Adamic's ruling on #qk8rztp S2c-2a). Not yet: next(x), an origin the proof
// can't name (the parameter of an arrow kept in a local, whose callers the graph doesn't see), a
// generator, whose step returns its value, not a result (S2c-2b), and an optional call of next or
// of return(), as any call through ?. is. Still refused as main refuses it: a detached next, by the
// unbound-method rule. The value read unnarrowed, or in the
// done arm, is any, which stops at the any rules (Adamic's void pin). Mutants
// (internal/oracle/testdata/iterable_views/mutants.md): the origin check dropped admits the arrow
// row; the next(x) stop dropped admits its row.
func TestHandDrivenNextStopsOnEveryOtherRoad(t *testing.T) {
	t.Parallel()
	const generator = "function* counted(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n"
	for _, probe := range []struct{ name, source, reason string }{
		{"next(x)", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next(1); return status.done ? -1 : status.value; }\nconsole.log(String(first(sizes.values())));", "next(x) through a view"},
		{"an arrow's parameter, its callers unseen", handNumbers + "const first = (numbers: IterableIterator<number>): number => { const status = numbers.next(); return status.done ? -1 : status.value; };\nconsole.log(String(first(sizes.values())));", "a view over an iterable of unproven origin"},
		{"a generator", generator + handFirst + "console.log(String(first(counted(3))));", "which a hand-driven next() reads"},
		{"an optional call of next", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next?.(); return status === undefined || status.done ? -1 : status.value; }\nconsole.log(String(first(sizes.values())));", "a call through ?. (an optional call)"},
		{"return()", handNumbers + "function first(numbers: IterableIterator<number>): number { numbers.return?.(); return 0; }\nconsole.log(String(first(sizes.values())));", "a call through ?. (an optional call)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				t.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
	// A for-of beside a hand-driven next over the same parameter, in .a and in .ts (Compiler's must).
	for _, extension := range []string{".a", ".ts"} {
		mixed := handNumbers + "function both(numbers: IterableIterator<number>): number { let sum = 0; const first = numbers.next(); if (!first.done) { sum += first.value; } for (const value of numbers) { sum += value; } return sum; }\nconsole.log(String(both(sizes.values())));"
		t.Run("a for-of beside a next in "+extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, mixed, extension)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "a for-of beside a hand-driven next() over the same parameter") {
				t.Fatalf("got %v, want a NotYet saying a for-of beside a hand-driven next()", err)
			}
		})
	}
	for _, probe := range []struct{ name, source, reason string }{
		{"a detached next", handNumbers + "function first(numbers: IterableIterator<number>): number { const step = numbers.next; const status = step(); return status.done ? -1 : status.value; }\nconsole.log(String(first(sizes.values())));", "a method read as a value"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want a refusal saying %q", err, probe.reason)
			}
		})
	}
	// The value unnarrowed, or in the done arm, is any. Which any rule stops it, refused or not yet,
	// is the any rules' to say; that it stops, and names any, is the pin.
	for _, probe := range []struct{ name, source string }{
		{"value unnarrowed", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next(); return status.value; }\nconsole.log(String(first(sizes.values())));"},
		{"value in the done arm", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next(); if (status.done) { return status.value; } return 0; }\nconsole.log(String(first(sizes.values())));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), "any") {
				t.Fatalf("got %v, want it stopped as a value of type any", err)
			}
		})
	}
}
