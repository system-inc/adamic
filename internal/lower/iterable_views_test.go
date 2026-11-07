package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// iterableViewCounts is a Map whose keys and values are built, for the probes below to take iterators of.
const iterableViewCounts = "const counts = new Map<string, number>();\nfor (let index = 0; index < 3; index++) { counts.set(`key ${String(index)}`, index * 2); }\n"

// iterableViewWords is a Set of built strings, whose iterators an Iterable<string> slot is handed.
const iterableViewWords = "const words = new Set<string>();\nfor (let index = 0; index < 3; index++) { words.add(`word ${String(index)}`); }\n"

// iterableViewShout is a for-of over an Iterable<string> parameter, held as a union.
const iterableViewShout = "function shout(items: Iterable<string>): void { for (const word of items) { console.log(word.toUpperCase()); } }\n"

// iterableViewTotal is a for-of over an Iterable<number> parameter, the view every probe hands origins to.
const iterableViewTotal = "function total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\n"

// A view whose every origin is a collection's iterator lowers (#qk8rztp S2b-1b): entries read as a
// pair, values and keys from a Map and a Set, through a parameter typed Iterable<T> or
// IterableIterator<T> and a local typed Iterable<T>. The oracle holds what these run to
// (testdata/iterable_views); this holds that they lower at all, so a mutant that stops every view
// goes red here too.
func TestIterableViewsOverCollectionIteratorsLower(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"entries destructured", iterableViewCounts +
			"function show(items: Iterable<[string, number]>): void { for (const [name, count] of items) { console.log(`${name} ${String(count)}`); } }\nshow(counts.entries());\nshow(new Map<string, number>([['other'.repeat(2), 4]]).entries());"},
		{"entries as a pair", iterableViewCounts +
			"function show(items: Iterable<[string, number]>): void { for (const pair of items) { console.log(`${pair[0]} ${String(pair[1])}`); } }\nshow(counts.entries());"},
		{"a Map's and a Set's values", iterableViewCounts +
			iterableViewTotal +
			"const seen = new Set<number>([1, 2]);\nconsole.log(String(total(counts.values()) + total(seen.values())));"},
		{"IterableIterator", iterableViewCounts +
			"function show(items: IterableIterator<number>): void { for (const value of items) { console.log(String(value)); } }\nshow(counts.values());"},
		{"a local", iterableViewCounts +
			"const view: Iterable<number> = counts.values();\nfor (const value of view) { console.log(String(value)); }"},
		{"through a call's result", iterableViewCounts +
			iterableViewTotal +
			"function values(): Iterable<number> { return counts.values(); }\nconsole.log(String(total(values())));"},
		{"Iterable<string>, held as a union", iterableViewCounts +
			iterableViewWords +
			iterableViewShout +
			"shout(words.values());\nshout(counts.keys());"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			if _, err := lowerSource(test, probe.source); err != nil {
				test.Fatalf("want it to lower, got %v", err)
			}
		})
	}
}

// A view any value but a collection's iterator may reach stops as not yet built, in its own words
// (Lower's ruling on q1, conditions 1 to 3): an origin the proof can't name, and origins that hold
// their steps otherwise than the view reads them. Each row's reason is a word of its stop, so a row
// that stops for another reason goes red too. An array never reaches the view's proof: going where a
// structural type is held, it stops at the call (shape_fit.go), and that row holds it stays so. A
// host callback's parameter, a record's value and a Map's value are origins the walk doesn't follow,
// and a string handed to an Iterable<string> slot is a second kind, not built until a dispatch is. A
// program arrow kept in a local and called through it is a function value: the graph marks its
// parameters unknown (view_flow.go's MakeClosure arm), whatever its one call site hands it, since
// the closure may be called from anywhere it reaches.
// Mutants (testdata/iterable_views/mutants.md in internal/oracle): the unproven test made false turns
// every unproven row but the array's into a program that compiles; the held test made false does that
// to the held-differently rows; the pass not run, or the view not recorded, does it to both.
func TestIterableViewsOfOtherOriginsAreNotYet(test *testing.T) {
	test.Parallel()
	const unproven = "a view over an iterable of unproven origin"
	const differently = "a view over an iterable whose origins yield values held differently"
	for _, probe := range []struct{ name, source, reason string }{
		{
			"an array at a second call site",
			iterableViewCounts +
				iterableViewTotal +
				"console.log(String(total(counts.values())));\nconsole.log(String(total([1, 2])));",
			"an array going where a plain object type is held",
		},
		{
			"a program's own iterable object",
			iterableViewCounts +
				iterableViewTotal +
				"const wrapped = { [Symbol.iterator](): MapIterator<number> { return counts.values(); } };\nconsole.log(String(total(counts.values())));\nconsole.log(String(total(wrapped)));",
			unproven,
		},
		{
			"the function as a value",
			iterableViewCounts +
				iterableViewTotal +
				"const add = total;\nconsole.log(String(total(counts.values())));\nconsole.log(String(add(counts.values())));",
			unproven,
		},
		{
			"a program arrow called through its local",
			iterableViewCounts +
				"const total = (it: Iterable<number>): number => { let sum = 0; for (const x of it) { sum += x; } return sum; };\nconsole.log(String(total(counts.values())));",
			unproven,
		},
		{
			"an object literal's method",
			iterableViewCounts +
				"const summer = { total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; } };\nconsole.log(String(summer.total(counts.values())));",
			unproven,
		},
		{
			"a field",
			iterableViewCounts +
				"const box: { items: Iterable<number> } = { items: counts.values() };\nfor (const value of box.items) { console.log(String(value)); }",
			unproven,
		},
		{
			"a host callback's parameter",
			iterableViewCounts +
				"const lists: Iterable<number>[] = [counts.values()];\nlists.forEach((items: Iterable<number>): void => { for (const value of items) { console.log(String(value)); } });",
			unproven,
		},
		{
			"a record's value",
			iterableViewCounts +
				"const table: Record<string, Iterable<number>> = { first: counts.values() };\nconst got = table['first'];\nif (got !== undefined) { for (const value of got) { console.log(String(value)); } }",
			unproven,
		},
		{
			"a Map's value",
			iterableViewCounts +
				"const byName = new Map<string, Iterable<number>>();\nbyName.set(`name ${String(1)}`, counts.values());\nconst got = byName.get('name 1');\nif (got !== undefined) { for (const value of got) { console.log(String(value)); } }",
			unproven,
		},
		{
			"a string where Iterable<string> is held",
			iterableViewWords + iterableViewShout + "shout(words.values());\nshout(`text ${String(words.size)}`);",
			unproven,
		},
		{
			"values held as a number and a boolean, read as a union",
			iterableViewCounts +
				"const flags = new Map<string, boolean>();\nflags.set(`flag ${String(1)}`, true);\nfunction show(items: Iterable<number | boolean>): void { for (const value of items) { console.log(String(value)); } }\nshow(counts.values());\nshow(flags.values());",
			differently,
		},
		{
			"entries whose value is read as a union",
			iterableViewCounts +
				"function show(items: Iterable<readonly [string, number | boolean]>): void { for (const [name, value] of items) { console.log(`${name} ${String(value)}`); } }\nshow(counts.entries());",
			differently,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				test.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
}

// An Iterable<string> slot is held as a union, and its view reads it as the collection's iterator only
// through a checked unbox (Lower's ruling on #qk8rztp): if the origin proof were ever wrong, the tag
// test stops the program instead of stepping a string as an iterator. No proven program can reach the
// test's failing arm, so this pins that the loop has it. Mutant: the unbox made unchecked, and this
// goes red.
func TestIterableViewHeldAsAUnionIsReadThroughACheckedUnbox(test *testing.T) {
	test.Parallel()
	program, err := lowerSource(test, iterableViewWords+iterableViewShout+"shout(words.values());")
	if err != nil {
		test.Fatal(err)
	}
	checked := 0
	inspect := func(node any) bool {
		if narrow, isNarrow := node.(ir.Narrow); isNarrow &&
			narrow.To == ir.Object &&
			narrow.Checked &&
			strings.Contains(
				narrow.Message,
				"a view over an iterable held something other than a collection's iterator",
			) {
			checked++
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if checked != 1 {
		test.Fatalf("%d checked unboxes of the view's slot, want 1", checked)
	}
}

// Through a view, only the for-of is built: next with an argument, return() and throw() through an
// IterableIterator<T> parameter whose only origin is a Map's iterator don't compile (Lower's ruling on
// q1, condition 6, and q10's fourth). This head builds none of them and changes none of their roads;
// the rows hold that they stay stopped.
func TestIterableViewProtocolCallsDontLower(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, call string }{
		{"next with an argument", "items.next(1);"},
		{"return", "items.return?.(undefined);"},
		{"throw", "items.throw?.(new Error(`stop ${String(1)}`));"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			source := iterableViewCounts +
				"function show(items: IterableIterator<number>): void { " +
				probe.call +
				" for (const value of items) { console.log(String(value)); } }\nshow(counts.values());"
			_, err := lowerSource(test, source)
			var gap *NotYet
			var refused *Refused
			if !errors.As(err, &gap) && !errors.As(err, &refused) {
				test.Fatalf("got %v, want it stopped", err)
			}
		})
	}
}
