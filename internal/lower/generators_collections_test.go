package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// new Map and new Set of a direct call to a program's generator step it by its origin, as a spread
// does (#qk8rztp S2b-1a): the Map is made from the pairs collectGenerator gave, the Set from the
// values. Mutant: newMapFromGenerator's pairs dropped from the MapNew it makes; the Map row finds no
// pairs and goes red.
func TestNewMapAndNewSetStepAGeneratorsCall(test *testing.T) {
	test.Parallel()
	program, lowerSourceError := lowerSource(
		test,
		`function* pairs(count: number): Generator<[string, number], undefined, undefined> {
	for (let index = 0; index < count; index++) {
		yield ["k" + String(index), index];
	}
}
function* words(count: number): Generator<string, undefined, undefined> {
	for (let index = 0; index < count; index++) {
		yield "w" + String(index);
	}
}
const map = new Map(pairs(2));
const set = new Set(words(2));
console.log(String(map.size) + " " + String(set.size));`,
	)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	if checkGeneratorRoadsError := checkGeneratorRoads(program); checkGeneratorRoadsError != nil {
		test.Errorf("the lowered program's roads: %v", checkGeneratorRoadsError)
	}
	made := map[string]bool{}
	walk(program.Main, func(node any) bool {
		switch node := node.(type) {
		case ir.MapNew:
			if call, ok := node.Pairs.(ir.Call); ok &&
				program.Functions[program.CallTargets(call)[0]].Name == "collect_generator" {
				made["map"] = true
			}
		case ir.SetNew:
			if call, ok := node.Values.(ir.Call); ok &&
				program.Functions[program.CallTargets(call)[0]].Name == "collect_generator" {
				made["set"] = true
			}
		}
		return true
	})
	if !made["map"] || !made["set"] {
		test.Errorf("made from a collected generator: %v; want both the Map and the Set", made)
	}
	// A generic generator's instance is consumed as a plain one is (S2a): its pairs are read from the
	// call's instantiated type.
	// A tuple's fixed indices read K and V as they are, where an array's index would read K | undefined.
	if _, lowerSourceError := lowerSource(
		test,
		`function* zip<K, V>(items: readonly (readonly [K, V])[]): Generator<[K, V], undefined, undefined> {
	for (const item of items) {
		yield [item[0], item[1]];
	}
}
function* each<T>(items: readonly T[]): Generator<T, undefined, undefined> {
	for (const item of items) {
		yield item;
	}
}
console.log(String(new Map(zip([["a", 1], ["b", 2]])).size) + " " + String(new Set(each([3, 3])).size));`,
	); lowerSourceError != nil {
		test.Errorf("new Map and new Set of generic generators' instances: %v", lowerSourceError)
	}
}

// What new Map can't take from a generator's call stops as not yet, in words that say why: a yield
// that isn't one [key, value] tuple (a union of two), and pairs held otherwise than the Map's keys and
// values, each in the words newMapFrom gives any other source. A Generator that isn't a direct call, a parameter here,
// still stops as it did before S2b-1a (TestNewMapFromAGeneratorNamesGenerators holds the same). Mutants:
// generatorPairTypes' tuple test taken out (the union row goes red), newMapFromGenerator's held-otherwise test taken
// out (that row lowers, and goes red).
func TestNewMapOfAGeneratorsCallStopsWhereItsYieldsArentItsPairs(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{
			"a union of pairs",
			`function* pairs(): Generator<readonly ["a", number] | readonly ["b", number], undefined, undefined> {
	yield ["a", 1];
	yield ["b", 2];
}
console.log(String(new Map(pairs()).size));`,
			"new Map from something that isn't [key, value] pairs",
		},
		{"pairs held otherwise", `function* pairs(): Generator<[string, number], undefined, undefined> {
	yield ["a" + "b", 1];
}
console.log(String(new Map<string, number | string>(pairs()).size));`, "new Map from pairs held otherwise than the Map's keys and values"},
		{
			"a generator that isn't a call",
			`function size(source: Generator<readonly [string, number], undefined, unknown>): number {
	return new Map(source).size;
}
console.log('built');`,
			"new Map from a generator's pairs",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var gap *NotYet
			if !errors.As(lowerSourceError, &gap) || !strings.Contains(gap.What, probe.reason) {
				test.Fatalf("got %v, want a NotYet saying %q", lowerSourceError, probe.reason)
			}
		})
	}
}
