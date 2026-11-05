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
func TestIterableViewsOverCollectionIteratorsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"entries destructured", iterableViewCounts + "function show(items: Iterable<[string, number]>): void { for (const [name, count] of items) { console.log(`${name} ${String(count)}`); } }\nshow(counts.entries());\nshow(new Map<string, number>([['other'.repeat(2), 4]]).entries());"},
		{"entries as a pair", iterableViewCounts + "function show(items: Iterable<[string, number]>): void { for (const pair of items) { console.log(`${pair[0]} ${String(pair[1])}`); } }\nshow(counts.entries());"},
		{"a Map's and a Set's values", iterableViewCounts + iterableViewTotal + "const seen = new Set<number>([1, 2]);\nconsole.log(String(total(counts.values()) + total(seen.values())));"},
		{"IterableIterator", iterableViewCounts + "function show(items: IterableIterator<number>): void { for (const value of items) { console.log(String(value)); } }\nshow(counts.values());"},
		{"a local", iterableViewCounts + "const view: Iterable<number> = counts.values();\nfor (const value of view) { console.log(String(value)); }"},
		{"through a call's result", iterableViewCounts + iterableViewTotal + "function values(): Iterable<number> { return counts.values(); }\nconsole.log(String(total(values())));"},
		{"Iterable<string>, held as a union", iterableViewCounts + iterableViewWords + iterableViewShout + "shout(words.values());\nshout(counts.keys());"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it to lower, got %v", err)
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
// and a string handed to an Iterable<string> slot is a second kind, not built until a dispatch is.
// Mutants (testdata/iterable_views/mutants.md in internal/oracle): the unproven test made false turns
// every unproven row but the array's into a program that compiles; the held test made false does that
// to the held-differently rows; the pass not run, or the view not recorded, does it to both.
func TestIterableViewsOfOtherOriginsAreNotYet(t *testing.T) {
	t.Parallel()
	const unproven = "a view over an iterable of unproven origin"
	const differently = "a view over an iterable whose origins yield values held differently"
	for _, probe := range []struct{ name, source, reason string }{
		{"an array at a second call site", iterableViewCounts + iterableViewTotal + "console.log(String(total(counts.values())));\nconsole.log(String(total([1, 2])));", "an array going where a plain object type is held"},
		{"a program's own iterable object", iterableViewCounts + iterableViewTotal + "const wrapped = { [Symbol.iterator](): MapIterator<number> { return counts.values(); } };\nconsole.log(String(total(counts.values())));\nconsole.log(String(total(wrapped)));", unproven},
		{"the function as a value", iterableViewCounts + iterableViewTotal + "const add = total;\nconsole.log(String(total(counts.values())));\nconsole.log(String(add(counts.values())));", unproven},
		{"an object literal's method", iterableViewCounts + "const summer = { total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; } };\nconsole.log(String(summer.total(counts.values())));", unproven},
		{"a field", iterableViewCounts + "const box: { items: Iterable<number> } = { items: counts.values() };\nfor (const value of box.items) { console.log(String(value)); }", unproven},
		{"a host callback's parameter", iterableViewCounts + "const lists: Iterable<number>[] = [counts.values()];\nlists.forEach((items: Iterable<number>): void => { for (const value of items) { console.log(String(value)); } });", unproven},
		{"a record's value", iterableViewCounts + "const table: Record<string, Iterable<number>> = { first: counts.values() };\nconst got = table['first'];\nif (got !== undefined) { for (const value of got) { console.log(String(value)); } }", unproven},
		{"a Map's value", iterableViewCounts + "const byName = new Map<string, Iterable<number>>();\nbyName.set(`name ${String(1)}`, counts.values());\nconst got = byName.get('name 1');\nif (got !== undefined) { for (const value of got) { console.log(String(value)); } }", unproven},
		{"a string where Iterable<string> is held", iterableViewWords + iterableViewShout + "shout(words.values());\nshout(`text ${String(words.size)}`);", unproven},
		{"values held as a number and a boolean, read as a union", iterableViewCounts + "const flags = new Map<string, boolean>();\nflags.set(`flag ${String(1)}`, true);\nfunction show(items: Iterable<number | boolean>): void { for (const value of items) { console.log(String(value)); } }\nshow(counts.values());\nshow(flags.values());", differently},
		{"entries whose value is read as a union", iterableViewCounts + "function show(items: Iterable<readonly [string, number | boolean]>): void { for (const [name, value] of items) { console.log(`${name} ${String(value)}`); } }\nshow(counts.entries());", differently},
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
}

// An Iterable<string> slot is held as a union, and its view reads it as the collection's iterator only
// through a checked unbox (Lower's ruling on #qk8rztp): if the origin proof were ever wrong, the tag
// test stops the program instead of stepping a string as an iterator. No proven program can reach the
// test's failing arm, so this pins that the loop has it. Mutant: the unbox made unchecked, and this
// goes red.
func TestIterableViewHeldAsAUnionIsReadThroughACheckedUnbox(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, iterableViewWords+iterableViewShout+"shout(words.values());")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	inspect := func(node any) bool {
		if narrow, isNarrow := node.(ir.Narrow); isNarrow && narrow.To == ir.Object && narrow.Checked && strings.Contains(narrow.Message, "a view over an iterable held something other than a collection's iterator") {
			checked++
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if checked != 1 {
		t.Fatalf("%d checked unboxes of the view's slot, want 1", checked)
	}
}

// Through a view, only the for-of is built: next with an argument, return() and throw() through an
// IterableIterator<T> parameter whose only origin is a Map's iterator don't compile (Lower's ruling on
// q1, condition 6, and q10's fourth). This head builds none of them and changes none of their roads;
// the rows hold that they stay stopped.
func TestIterableViewProtocolCallsDontLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, call string }{
		{"next with an argument", "items.next(1);"},
		{"return", "items.return?.(undefined);"},
		{"throw", "items.throw?.(new Error(`stop ${String(1)}`));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := iterableViewCounts + "function show(items: IterableIterator<number>): void { " + probe.call + " for (const value of items) { console.log(String(value)); } }\nshow(counts.values());"
			_, err := lowerSource(t, source)
			var gap *NotYet
			var refused *Refused
			if !errors.As(err, &gap) && !errors.As(err, &refused) {
				t.Fatalf("got %v, want it stopped", err)
			}
		})
	}
}
