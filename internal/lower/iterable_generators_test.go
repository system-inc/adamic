package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// iterableCounted is a generator whose call the probes below hand into an Iterable<number> parameter.
// Its TReturn is undefined, not void, so a road the one-read rule closes reaches lowering: a road of a
// Generator<T, void> outside a stepped site is refused earlier by the void rules.
const iterableCounted = "function* counted(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n"

// iterableTotal is a function whose one use of its Iterable<number> parameter is one for-of over it.
const iterableTotal = "function total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\n"

// A generator's direct call passed straight into a one-body function's Iterable<T> parameter, whose
// only use is one for-of, lowers (#qk8rztp S2c-1, Adamic's ruling): a module function, through
// IterableIterator<T> too, a generic instance, and the shapes Language named that must stay
// admitted, a type-only typeof and a nested block's own name. The oracle holds what these run to
// (testdata/generators); this holds that they lower at all.
func TestGeneratorIntoAnIterableParameterLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"a module function", iterableCounted + iterableTotal + "console.log(String(total(counted(3))));"},
		// steppedSite's arm: a void TReturn is never read through the parameter's one for-of.
		{"a void TReturn", "function* counted(limit: number): Generator<number, void, unknown> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n" + iterableTotal + "console.log(String(total(counted(3))));"},
		{"an IterableIterator parameter", iterableCounted + "function total(items: IterableIterator<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
		{"a generic instance", iterableCounted + "function weigh<T>(items: Iterable<T>, by: (item: T) => number): number { let sum = 0; for (const item of items) { sum += by(item); } return sum; }\nconsole.log(String(weigh(counted(3), (value: number): number => value + 1)));"},
		{"one function at two call sites", iterableCounted + iterableTotal + "console.log(String(total(counted(2)) + total(counted(4))));"},
		{"an outer loop over the for-of", iterableCounted + "function firsts(items: Iterable<number>): number { let sum = 0; for (let pass = 0; pass < 2; pass++) { for (const value of items) { sum += value; break; } } return sum; }\nconsole.log(String(firsts(counted(3))));"},
		{"a type-only typeof", iterableCounted + "function total(items: Iterable<number>): number { const none: typeof items | undefined = undefined; let sum = none === undefined ? 0 : 1; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
		{"a nested block's own name", iterableCounted + "function total(items: Iterable<number>): number { let sum = 0; { const items = 4; sum += items; } for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
		})
	}
}

// Every road the one-read rule closes stops as not yet built, each in its own words (Adamic's ruling
// on #qk8rztp S2c-1, conditions 2 and 3, Compiler's roads and Language's shapes): a second read of the
// parameter in the lowered IR (a hand-driven next, a typeof or an identity test, a read after a
// labeled break out of the loop, a default that names it, a next inside the loop's own body), a
// closure that captures it, never called, passing it on, assigning it, reading arguments, and a callee
// the call can't name as one body (an async function, a function value, a method), or a generator
// that isn't the argument's own call (one stored first). And what S2c-1 doesn't build yet: two
// generator functions into one parameter, a generator beside a collection's iterator, an
// Iterable<string> parameter, held as a union, and yields held otherwise than the view reads them.
// Spreading or destructuring the parameter stops before the proof, in its own words, so those rows
// hold only that it doesn't compile. Mutants (internal/oracle/testdata/generators/mutants.md): the
// one-read test made false admits the second-read rows; the call site taking any generator's call as
// an Iterable<T> argument turns the stored row into a compiler bug.
func TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad(t *testing.T) {
	t.Parallel()
	const readTwice = "read 2 times"
	const other = "used other than by a for-of"
	function := func(body string) string {
		return iterableCounted + "function total(items: IterableIterator<number>): number { let sum = 0; " + body + " return sum; }\nconsole.log(String(total(counted(3))));"
	}
	for _, probe := range []struct{ name, source, reason string }{
		{"next through the parameter", function("const first = items.next(); if (!first.done) { sum += first.value; } for (const value of items) { sum += value; }"), readTwice},
		{"typeof it", function("if (typeof items === 'object') { sum += 1; } for (const value of items) { sum += value; }"), readTwice},
		{"next inside the loop", function("for (const value of items) { items.next(); sum += value; }"), readTwice},
		{"a read after a labeled break", function("outer: for (let pass = 0; pass < 2; pass++) { for (const value of items) { sum += value; break outer; } } if (typeof items === 'object') { sum += 1; }"), readTwice},
		{"it === other", iterableCounted + "function total(items: Iterable<number>, other: Iterable<number>): number { let sum = items === other ? 1 : 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2), counted(3))));", readTwice},
		{"a default naming it", iterableCounted + "function total(items: Iterable<number>, base: number = typeof items === 'object' ? 1 : 0): number { let sum = base; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2))));", readTwice},
		{"a closure capturing it, never called", function("const keep = (): Iterable<number> => items; for (const value of items) { sum += value; }"), "which a closure captures"},
		{"passing it on", iterableCounted + iterableTotal + "function outer(items: Iterable<number>): number { return total(items); }\nconsole.log(String(outer(counted(3))));", "whose one use isn't a for-of over it"},
		{"reassigning it", iterableCounted + "function total(items: Iterable<number>, other: Iterable<number>): number { let sum = 0; items = other; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2), counted(3))));", "which the function assigns"},
		{"arguments", function("sum += arguments.length; for (const value of items) { sum += value; }"), "reads arguments"},
		{"an async callee", iterableCounted + "async function total(items: Iterable<number>): Promise<number> { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(await total(counted(2))));", other},
		{"a call through a function value", iterableCounted + iterableTotal + "const add = total;\nconsole.log(String(add(counted(2))));", other},
		{"a method", iterableCounted + "class Summer { total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; } }\nconsole.log(String(new Summer().total(counted(2))));", other},
		{"a stored generator passed in", iterableCounted + iterableTotal + "const made = counted(2);\nconsole.log(String(total(made)));", other},
		{"two generator functions", iterableCounted + iterableTotal + "function* halved(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index / 2; } }\nconsole.log(String(total(counted(2)) + total(halved(2))));", "generators of more than one function"},
		{"a generator beside a collection's iterator", iterableCounted + iterableTotal + "const sizes = new Map<string, number>([[`a${String(1)}`, 1]]);\nconsole.log(String(total(counted(2)) + total(sizes.values())));", "both a generator and a collection's iterator"},
		{"an Iterable<string> parameter", "function* words(limit: number): Generator<string, undefined, undefined> { for (let index = 0; index < limit; index++) { yield `w${String(index)}`; } }\nfunction joined(items: Iterable<string>): string { let all = ''; for (const word of items) { all += word; } return all; }\nconsole.log(joined(words(2)));", "held as a union"},
		{"yields held otherwise", iterableCounted + "function show(items: Iterable<number | boolean>): void { for (const value of items) { console.log(String(value)); } }\nshow(counted(2));", "held differently"},
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
	for _, probe := range []struct{ name, source string }{
		{"spreading it", function("const all = [...items]; sum += all.length; for (const value of items) { sum += value; }")},
		{"destructuring it", function("const [first] = items; sum += first ?? 0; for (const value of items) { sum += value; }")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			var refused *Refused
			if !errors.As(err, &gap) && !errors.As(err, &refused) {
				t.Fatalf("got %v, want it stopped", err)
			}
		})
	}
}

// The view a generator feeds is rewritten into a per-step dispatch on the heap kind (Lower's (b)): its
// block closes what it holds at every way out, its loop tests the kind, panics as a compiler bug on
// any kind the proof didn't list, and steps the generator by GeneratorNext and GeneratorDone of the
// object it holds, never by a library iterator's next. And the generator crosses into the parameter
// through IterableOf, as the call's own argument. Mutant: the kind test made always true, and every
// generator fixture that reaches a view panics through the loud default.
func TestGeneratorViewDispatchesOnTheHeapKind(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, iterableCounted+iterableTotal+"console.log(String(total(counted(3))));")
	if err != nil {
		t.Fatal(err)
	}
	closes, kinds, steps, iterables := 0, 0, 0, 0
	inspect := func(node any) bool {
		switch node := node.(type) {
		case ir.Block:
			if node.Closes != 0 {
				closes++
			}
		case ir.IsGenerator:
			kinds++
		case ir.GeneratorNext:
			if node.Generator.Type() == ir.Object && program.Functions[node.Function].Generator {
				steps++
			}
		case ir.Call:
			for _, argument := range node.Arguments {
				if _, isIterable := argument.(ir.IterableOf); isIterable {
					iterables++
				}
			}
		case ir.Property:
			if node.Name == "next" {
				t.Errorf("the dispatch reads a library iterator's next: %+v", node)
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if closes != 1 || kinds != 1 || steps != 1 || iterables != 1 {
		t.Fatalf("%d closing blocks, %d kind tests, %d generator steps of an object and %d IterableOf arguments, want one each", closes, kinds, steps, iterables)
	}
}
