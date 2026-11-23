package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// dictionaryViewDeclarations is tsc's MapLike (core.ts) and an OptionsBase-shaped dictionary whose values are held
// as a Union, for each probe to cast to.
const dictionaryViewDeclarations = "interface MapLike<T> { [index: string]: T } interface OptionsBase { [option: string]: string | number | boolean | undefined } "

// dictionaryViewOwnKeys is tsc's hasOwnProperty alias and getOwnKeys (core.ts:1287), a for...in and an own-key test,
// for the probes that hand the view on to it.
const dictionaryViewOwnKeys = "const hasOwnProperty = Object.prototype.hasOwnProperty; function getOwnKeys<T>(map: MapLike<T>): string[] { const keys: string[] = []; for (const key in map) { if (hasOwnProperty.call(map, key)) { keys.push(key); } } return keys; } "

// dictionaryViewStop is the NotYet's words at a cast of object to MapLike<unknown> the view doesn't take.
const dictionaryViewStop = "a checked dictionary view of MapLike<unknown> over object"

// TestDictionaryViewCasts holds where the checked dictionary view is taken (#2v3xtgm, re-ruled Nov 23) and where a
// cast keeps castIndexSignature's NotYet at the cast: only object, undefined beside it where the target keeps it,
// cast to a dictionary of exactly unknown values, whose result is only ever read (dictionaryViewUse). An admitted cast
// lowers to its one helper, a checked narrow to a record that also tests the record holds references, in the two
// run-time NotYet's words. The rows a type keeps stopped read the view by key, so the use test alone would admit
// them. Mutant (M7): the use test answers true for every use, and every row a use keeps stopped goes red, admitted.
func TestDictionaryViewCasts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		// stop is the NotYet's words at the cast, or "" where the view is admitted.
		stop string
	}{
		{"maplike of unknown", "function f(x: object): unknown { return (x as MapLike<unknown>)['a']; }", ""},
		{"record of unknown", "function f(x: object): unknown { return (x as Record<string, unknown>)['a']; }", ""},
		{"undefined kept", "function f(x: object | undefined): unknown { const v = x as MapLike<unknown> | undefined; return v !== undefined ? v['a'] : undefined; }", ""},
		{"walked by its own keys", "const hasOwnProperty = Object.prototype.hasOwnProperty; function f(x: object): number { let count = 0; for (const key in x as MapLike<unknown>) { if (hasOwnProperty.call(x as MapLike<unknown>, key)) { count++; } } return count; }", ""},
		{"handed to an own keys helper", dictionaryViewOwnKeys + "function f(x: object): string[] { return getOwnKeys(x as MapLike<unknown>); }", ""},
		// moduleNameResolver.ts:2573's shape, which hands the view on to getOwnKeys.
		{"handed on to an own keys helper", dictionaryViewOwnKeys + "function allKeysStartWithDot(obj: MapLike<unknown>): boolean { for (const k of getOwnKeys(obj)) { if (!k.startsWith('.')) { return false; } } return true; } function f(x: object): boolean { return allKeysStartWithDot(x as MapLike<unknown>); }", ""},
		// core.ts:1370's shape: ===, !, for...in, hasOwnProperty.call and reads by key, both arguments views.
		{"compared and walked as equal own properties", "const hasOwnProperty = Object.prototype.hasOwnProperty; function equalOwnProperties(left: MapLike<unknown> | undefined, right: MapLike<unknown> | undefined): boolean { if (left === right) { return true; } if (!left || !right) { return false; } for (const key in left) { if (hasOwnProperty.call(left, key)) { if (!hasOwnProperty.call(right, key) || left[key] !== right[key]) { return false; } } } for (const key in right) { if (hasOwnProperty.call(right, key) && !hasOwnProperty.call(left, key)) { return false; } } return true; } function f(x: object, y: object): boolean { return equalOwnProperties(x as MapLike<unknown>, y as MapLike<unknown>); }", ""},
		{"undefined dropped", "function f(x: object | undefined): unknown { return (x as MapLike<unknown>)['a']; }",
			"a checked dictionary view of MapLike<unknown> over object | undefined"},
		{"number values", "function f(x: object): number | undefined { return (x as Record<string, number>)['a']; }",
			"a checked dictionary view of Record<string, number> over object"},
		{"union-held values", "function f(x: object): unknown { return (x as OptionsBase)['a']; }",
			"a checked dictionary view of OptionsBase over object"},
		{"returned", "function f(x: object): MapLike<unknown> { return x as MapLike<unknown>; }", dictionaryViewStop},
		{"stored in a const then written by key", "function f(x: object): void { const v = x as MapLike<unknown>; v['a'] = 1; }", dictionaryViewStop},
		{"handed to a helper that writes", "function f(x: object): void { g(x as MapLike<unknown>); } function g(m: MapLike<unknown>): void { m['a'] = 1; }", dictionaryViewStop},
		{"handed on to a helper that writes", "function f(x: object): void { g(x as MapLike<unknown>); } function g(m: MapLike<unknown>): void { h(m); } function h(m: MapLike<unknown>): void { m['a'] = 1; }", dictionaryViewStop},
		// Only reads, but the helper hands the view to itself, a pair met again while it's answered, which answers no.
		{"handed to a self-recursive helper", "function f(x: object): unknown { return walk(x as MapLike<unknown>, 'a'); } function walk(m: MapLike<unknown>, k: string): unknown { if (m[k] === undefined && k.length < 3) { return walk(m, k + 'a'); } return m[k]; }", dictionaryViewStop},
		{"assigned to a let", "function f(x: object): unknown { let v = x as MapLike<unknown>; return v['a']; }", dictionaryViewStop},
		// Wave 2's write and delete went through another function's MapLike<unknown> parameter, which the cast's
		// result never reached, and the cast was returned, which the use test stops alone. Each writes through the
		// view's own value now, so only the write keeps the NotYet.
		{"written through a dictionary", "function f(x: object): void { (x as MapLike<unknown>)['a'] = 1; }", dictionaryViewStop},
		{"deleted through a dictionary", "function f(x: object, k: string): void { delete (x as MapLike<unknown>)[k]; }", dictionaryViewStop},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, dictionaryViewDeclarations+probe.source, ".ts")
			if probe.stop != "" {
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "stage 0 can't lower "+probe.stop+" yet") {
					t.Fatalf("want the NotYet %q at the cast, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want the view admitted, got %v", err)
			}
			found := false
			for _, function := range program.Functions {
				if function.Name != "dictionary_view" {
					continue
				}
				found = true
				returned, isReturn := function.Body[len(function.Body)-1].(ir.Return)
				narrow, isNarrow := returned.Value.(ir.Narrow)
				if !isReturn || !isNarrow || !narrow.Checked || narrow.To != ir.Record || narrow.Message != ir.DictionaryViewRecordStop || narrow.ReferencesMessage != ir.DictionaryViewReferencesStop {
					t.Fatalf("want the helper to return a checked narrow to a record with both stops, got %#v", function.Body)
				}
			}
			if !found {
				t.Fatal("no dictionary_view helper lowered")
			}
		})
	}
}

// A source with fields of its own keeps a stop, whichever rule makes it: the view takes object alone, so a fixed
// object's static type never reaches a record's read through it (#mft3zc6). Package is a type literal, whose implicit
// index signature makes the cast an upcast tsc takes: an interface would be TS2352 before any rule is reached.
func TestDictionaryViewKeepsAFixedSourceStopped(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, dictionaryViewDeclarations+"type Package = { readonly name: string }; function f(x: Package): MapLike<unknown> { return x as MapLike<unknown>; }", ".ts")
	var notYet *NotYet
	var refused *Refused
	if !errors.As(err, &notYet) && !errors.As(err, &refused) {
		t.Fatalf("want the fixed source stopped, NotYet or Refused, got %v", err)
	}
}

// .a keeps main's refusal for the view's own cast, as item 8 refuses every checked downcast there.
func TestDictionaryViewRefusedInAdamic(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, dictionaryViewDeclarations+"function f(x: object): MapLike<unknown> { return x as MapLike<unknown>; }", ".a")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "Adamic 0.1 refuses a cast the runtime can't check; ") {
		t.Fatalf(".a: want main's refusal word for word, got %v", err)
	}
}
