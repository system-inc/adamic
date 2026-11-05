package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A generic function* lowers once per instantiation, where it's called (#qk8rztp S2a), and its
// yielded type is judged at the instance, so each over strings yields a string slot and each over
// numbers a number slot, functions with their own frames. A generic generator called by another
// generic function is instantiated at that function's own instance. Every consumer finds the
// instance: a for-of, a spread, Array.from and a dropped call. Mutant: generatorCallee's generic arm
// taken out; the for-of reaches call()'s guard and stops, and the test goes red.
//
// Read by testing: generatorSignature reads the result through concrete, under the instance's type
// mapper, but representation reads a bare T through the instance's substitution too, so the two
// agree on every read instance. concrete is what names an unread one's yield (unknown, not T), and
// the mutant without it is red only on TestGenericGeneratorRoadsAreNotYet's unread row.
func TestGenericGeneratorsLowerPerInstance(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function* each<T>(items: readonly T[]) {
	for (const item of items) {
		yield item;
	}
}
function collect<U>(items: readonly U[]): U[] {
	return [...each(items)];
}
for (const word of each(["a", "b"])) {
	console.log(word);
}
console.log([...each([1, 2])].join(","));
console.log(Array.from(each(["c"])).join(","));
console.log(collect([true]).join(","));
each([3]);`)
	if err != nil {
		t.Fatal(err)
	}
	yields := map[ir.Type]int{}
	for _, function := range program.Functions {
		if function.Generator {
			if !strings.HasPrefix(function.Name, "each_") || function.Returns != ir.Generator {
				t.Errorf("generator %s returns %d; want an instance of each returning a Generator", function.Name, function.Returns)
			}
			yields[function.Yields]++
		}
	}
	if len(yields) != 3 || yields[ir.String] != 1 || yields[ir.Number] != 1 || yields[ir.Boolean] != 1 {
		t.Errorf("instances by what they yield: %v; want one each of string, number and boolean", yields)
	}
	if err := checkGeneratorRoads(program); err != nil {
		t.Errorf("the lowered program's roads: %v", err)
	}
}

// A generic generator's instance takes every road a non-generic one's call takes, and stops on each
// that isn't a consumer exactly as that one does: stored, passed, returned, put in an array, read
// through a view or as a value, destructuring, a mapper, an optional or spread call, a consumer
// inside a generator. new Set and new Map of one are consumers (generators_collections_test.go). An instance whose yielded type isn't held in one present
// slot, a union or unknown (an unread instance), stops at its declaration. Each row's reason is a
// word of its stop. Mutant: call()'s generator guard at the generic branch taken out; the five
// value rows stop for other reasons or reach checkGeneratorRoads' compiler bug, and go red.
func TestGenericGeneratorRoadsAreNotYet(t *testing.T) {
	t.Parallel()
	// TReturn is undefined, as in TestGeneratorFormsThisSliceDoesntLowerAreNotYet: a void one is
	// refused on these roads by the void rules first.
	const each = "function* each<T>(items: readonly T[]): Generator<T, undefined, undefined> { for (const item of items) { yield item; } }\n"
	for _, probe := range []struct{ name, source, reason string }{
		{"stored in a variable", each + "const held = each([1]); console.log(String(held.next().value));", "used other than by a for-of"},
		{"passed as an argument", each + "function take(source: object): number { return source === undefined ? 0 : 1; } console.log(String(take(each([2]))));", "used other than by a for-of"},
		{"returned", each + "function make(): Generator<number, undefined, undefined> { return each([2]); } for (const value of make()) { console.log(String(value)); }", "used other than by a for-of"},
		{"put in an array", each + "const all = [each([1]), each([2])]; console.log(String(all.length));", "used other than by a for-of"},
		{"read through a view", each + "const view: Iterable<number> = each([2]); for (const value of view) { console.log(String(value)); }", "used other than by a for-of"},
		{"the function as a value", each + "const make = each; for (const value of make([2])) { console.log(String(value)); }", "used as a value"},
		{"Array.from with a mapper", each + "console.log(Array.from(each([3]), (value: number): number => value * 2).join(','));", "Array.from with a mapper"},
		{"destructured", each + "const [first] = each([3]); console.log(String(first));", "consumed by destructuring"},
		{"called optionally", each + "for (const value of each?.([1])) { console.log(String(value)); }", "called optionally"},
		{"called with a spread argument", each + "const lists: [number[]] = [[1]]; for (const value of each(...lists)) { console.log(String(value)); }", "spread argument"},
		{"consumed inside a generator", each + "function* outer(): Generator<number, undefined, undefined> { for (const value of each([1])) { yield value; } return undefined; } for (const value of outer()) { console.log(String(value)); }", "inside an async function or a generator"},
		{"a union instance", each + "for (const value of each<number | string>([1, 'a'])) { console.log(String(value)); }", "one present slot"},
		{"an unread instance", each + "for (const value of each<unknown>([1])) { console.log(String(value)); }", "yielding unknown"},
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

// A generic generator called through a namespace import's member routes by its origin to its
// instance and stops as a non-generic one does there: its callee could run code first.
func TestGenericGeneratorCalledOtherThanByItsNameIsNotYet(t *testing.T) {
	t.Parallel()
	library := "export function* each<T>(items: readonly T[]): SetIterator<T> { for (const item of items) { yield item; } }\n"
	for _, probe := range []struct{ name, main string }{
		{"namespace for-of", "import * as library from './library';\nfor (const value of library.each([1])) console.log(String(value));\n"},
		{"namespace spread", "import * as library from './library';\nconsole.log([...library.each(['a'])].join(','));\n"},
		{"namespace Array.from", "import * as library from './library';\nconsole.log(Array.from(library.each(['a'])).join(','));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for name, source := range map[string]string{"library.ts": library, "main.ts": probe.main} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			_, err = Lower(context.Background(), program)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "called other than by its name") {
				t.Fatalf("got %v, want the generator's called-other-than-by-its-name stop", err)
			}
		})
	}
}

// tsc's own mapIterator (core.ts:332-336). Its for-of over iter is a view since S2b-1b
// (iterable_views.go), so the body lowers, and instantiated over an array it stops at the array:
// expression() judges every value against its contextual slot (containerWhereOtherGoes), and an
// argument of a generic call has the instantiated parameter's, Iterable<number>, where an array is
// shape_fit.go's stop. That fires while the call is lowered, before the view's proof runs after
// lowering, which would stop the same program at 2:21. Handed a Map's keys(), mapIterator lowers.
func TestTscMapIteratorOverAnArrayStops(t *testing.T) {
	t.Parallel()
	const mapIterator = `export function* mapIterator<T, U>(iter: Iterable<T>, mapFn: (x: T) => U): Generator<U, void, unknown> {
    for (const x of iter) {
        yield mapFn(x);
    }
}
`
	_, err := lowerSource(t, mapIterator+`console.log([...mapIterator([1, 2, 3], x => x * 2)].join(","));`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.Where, "main.a:6:29") || !strings.Contains(gap.What, "an array going where a plain object type is held") {
		t.Fatalf("got %v, want a NotYet at the array, 6:29", err)
	}
	keys := "const ids = new Map<number, string>();\nfor (let index = 1; index < 4; index++) { ids.set(index * 10, `id ${String(index)}`); }\n"
	if _, err := lowerSource(t, mapIterator+keys+`console.log([...mapIterator(ids.keys(), x => x + 1)].join(","));`); err != nil {
		t.Fatalf("over a Map's keys(): %v", err)
	}
}
