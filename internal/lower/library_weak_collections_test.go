package lower

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A WeakSet lowers whole, on its own representation, and what it can't hold stops by name: a key type a
// primitive fits (Runtime's ruling 1 on #hcgy751: lib's Object and String interfaces, and any shape a
// primitive satisfies, which tsc lets take a primitive without a cast), a WeakSet seen as one of another
// key type (invariance), a WeakSet going where a structural type is held and the converse, a method read as
// a value, and an iterable handed to the constructor.
// Mutants: weakCollectionKey without its noPrimitiveFits test, and the String and Object rows lower or stop
// in other words; containers() without WeakMap and WeakSet, and the widening rows no longer stop in
// invariance's words; containerPairGoes without its WeakCollection slot, and an object held where a WeakSet
// is held no longer stops.
func TestWeakSetStopsByName(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		reasons      []string
	}{
		{"a String key, which a string fits", `const s = new WeakSet<String>(); s.add('abc');`, []string{"a WeakSet keyed by String, which a primitive fits"}},
		{"an Object key, which a number fits", `const s = new WeakSet<Object>(); s.add(1);`, []string{"a WeakSet keyed by Object, which a primitive fits"}},
		{"a shape a string fits", `interface Lengthy { readonly length: number } const s = new WeakSet<Lengthy>(); console.log(String(s.has({ length: 1 })));`, []string{"a WeakSet keyed by Lengthy, which a primitive fits"}},
		{"a String key reached through a parameter", `function keep(s: WeakSet<String>): void { s.add('abc'); }`, []string{"a WeakSet keyed by String, which a primitive fits"}},
		{"a WeakSet seen as one of a wider key", `interface Keyed { name: string } function put(s: WeakSet<object>, k: object): void { s.add(k); } const s = new WeakSet<Keyed>(); put(s, {});`, []string{"WeakSet<Keyed> seen as WeakSet<object>, which can write object where Keyed is read"}},
		{"a WeakSet assigned to one of a wider key", `interface Keyed { name: string } const s = new WeakSet<Keyed>(); const wide: WeakSet<object> = s; wide.add([]);`, []string{"WeakSet<Keyed> seen as WeakSet<object>, which can write object where Keyed is read"}},
		{"a WeakSet where a structural has is held", `interface Lookup { has(k: object): boolean } const s = new WeakSet<object>(); const seen: Lookup = s; console.log(String(seen.has({})));`, []string{"a WeakMap or WeakSet going where a plain object type is held", "a view naming has, which lib declares on WeakSet"}},
		{"a WeakSet where a structural add is held", `interface Adder { add(k: object): unknown } const s = new WeakSet<object>(); const a: Adder = s; a.add({});`, []string{"a WeakMap or WeakSet going where a plain object type is held", "a view naming add, which lib declares on WeakSet"}},
		// The converse: a value that isn't a WeakSet where one is held. A class's add must return this, which
		// stops first; an object literal needs [Symbol.toStringTag], a computed key.
		{"a class implementing WeakSet where one is held", `class Fake { add(k: object): this { return this; } has(k: object): boolean { return false; } delete(k: object): boolean { return false; } readonly [Symbol.toStringTag]: string = 'WeakSet'; } const s: WeakSet<object> = new Fake(); console.log(String(s.has({})));`, []string{"a function returning this"}},
		{"an object literal implementing WeakSet where one is held", `const fake: WeakSet<object> = { add(k: object): WeakSet<object> { return fake; }, has(k: object): boolean { return false; }, delete(k: object): boolean { return false; }, [Symbol.toStringTag]: 'WeakSet' }; console.log(String(fake.has({})));`, []string{"a WeakMap or WeakSet interface implemented by a value that isn't one", "symbol", "Symbol", "computed"}},
		// The road with no incidental stop: an interface extending WeakSet is the program's own type, held as
		// an object, and tsc lets a value of it go where a WeakSet is held. Only the representation fit
		// (container_fit.go's WeakCollection slot) stops it there, and only shapeFit's WeakCollection arm, or
		// the view rule, stops a real WeakSet going where it is held.
		{"an object held where a WeakSet is held", `interface Marks extends WeakSet<object> { readonly label: string } function keep(marks: Marks): boolean { const plain: WeakSet<object> = marks; return plain.has({}); }`, []string{"a WeakMap or WeakSet interface implemented by a value that isn't one"}},
		{"a WeakSet held where an interface extending it is held", `interface Marks extends WeakSet<object> { readonly label?: string } function label(marks: Marks): string { return marks.label ?? 'none'; } console.log(label(new WeakSet<object>()));`, []string{"a WeakMap or WeakSet going where a plain object type is held", "a view naming", "label"}},
		{"a method read as a value", `const s = new WeakSet<object>(); const add = s.add;`, []string{"a method read as a value"}},
		{"an iterable handed to the constructor", `const k = {}; const s = new WeakSet<object>([k]);`, []string{"new WeakSet from an iterable"}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil {
				t.Fatalf("lowered; want a stop naming one of %q", probe.reasons)
			}
			for _, reason := range probe.reasons {
				if strings.Contains(err.Error(), reason) {
					return
				}
			}
			t.Fatalf("want a stop naming one of %q, got %v", probe.reasons, err)
		})
	}
}

// A WeakSet or a WeakMap an unknown or an object reaches stops by name, held directly or through a field, since
// neither backend's unknown can answer a run-time read of one as Node does (String gives [object WeakSet],
// `in` gives false). Runtime's must on W1: before this, each compiled and stopped or panicked at run
// time. W2 adds a WeakMap's three rows, which lower now that its get and set are built. Mutant:
// unknownView's weakCollectionType test taken out, and every row lowers.
func TestWeakSetViewedAsUnknownStops(t *testing.T) {
	t.Parallel()
	const reason = "a WeakMap or WeakSet viewed as unknown or object (weak collection metadata)"
	for _, probe := range []struct{ name, source string }{
		{"unknown holding a WeakSet", `const u: unknown = new WeakSet<object>(); console.log(String(u));`},
		{"object holding a WeakSet", `const o: object = new WeakSet<object>(); console.log(String('x' in o));`},
		{"unknown holding a WeakSet through a field", `interface Holder { readonly marks: WeakSet<object> } const h: Holder = { marks: new WeakSet<object>() }; const u: unknown = h; console.log(String(u));`},
		{"object holding a WeakSet through a field", `interface Holder { readonly marks: WeakSet<object> } const h: Holder = { marks: new WeakSet<object>() }; const o: object = h; console.log(String('marks' in o));`},
		{"a WeakSet passed where unknown is taken", `function show(value: unknown): string { return String(value); } console.log(show(new WeakSet<object>()));`},
		{"unknown holding a WeakMap", `const u: unknown = new WeakMap<object, number>(); console.log(String(u));`},
		{"object holding a WeakMap through a field", `interface Holder { readonly scores: WeakMap<object, number> } const h: Holder = { scores: new WeakMap<object, number>() }; const o: object = h; console.log(String('scores' in o));`},
		{"a WeakMap passed where unknown is taken", `function show(value: unknown): string { return String(value); } console.log(show(new WeakMap<object, string>()));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil {
				t.Fatalf("lowered; want a stop naming %q", reason)
			}
			if !strings.Contains(err.Error(), reason) {
				t.Fatalf("want a stop naming %q, got %v", reason, err)
			}
		})
	}
}

// Each key kind lowers to its own representation, and a WeakSet's operations to its own nodes, never a
// Map's: the set is ir.WeakCollection, add a WeakSetAdd that keeps no write site, has and delete their own.
func TestWeakSetLowersOnItsOwnType(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `const s = new WeakSet<object>();
const k = { x: 1 };
s.add(k);
console.log(String(s.has(k)));
console.log(String(s.delete(k)));
const arrays = new WeakSet<number[]>();
arrays.add([1]);
const buffers = new WeakSet<Uint8Array>();
buffers.add(new Uint8Array(1));
`)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]int{}
	keys := map[ir.Type]bool{}
	walk(program.Main, func(node any) bool {
		switch node := node.(type) {
		case ir.WeakCollectionNew:
			found["new"]++
			keys[node.Key] = true
			if !node.Set {
				t.Errorf("a WeakSet made as a WeakMap: %+v", node)
			}
		case ir.WeakSetAdd:
			found["add"]++
		case ir.WeakCollectionHas:
			found["has"]++
		case ir.WeakCollectionDelete:
			found["delete"]++
		case ir.MapNew, ir.SetNew, ir.SetAdd, ir.MapHas, ir.MapDelete:
			t.Errorf("a WeakSet lowered to a Map's node: %T", node)
		}
		return true
	})
	if found["new"] != 3 || found["add"] != 3 || found["has"] != 1 || found["delete"] != 1 {
		t.Errorf("found %v, want 3 new, 3 add, 1 has and 1 delete", found)
	}
	for _, want := range []ir.Type{ir.Union, ir.Array, ir.Uint8Array} {
		if !keys[want] {
			t.Errorf("no WeakSet keyed as %v among %v", want, keys)
		}
	}
}

// A WeakMap lowers on W1's weak collection with the values Runtime's ruling 2 on #hcgy751 admits, numbers,
// booleans and strings, and what it can't hold stops by name: a WeakMap seen as one of a wider value
// (invariance, by assignment and through a parameter, as tsc lets either through), a key type a primitive
// fits (k5), a value that may be missing (in the words a Map of one stops in), an object, an array, a
// union or a closure value (Lower's admission, #q2ybjr1), a method read as a value, and an iterable handed
// to the constructor. Mutant: weakCollectionAdmitted admits every value, and the object rows lower.
func TestWeakMapStopsByName(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		reasons      []string
	}{
		{"a WeakMap assigned to one of unknown values", `const m = new WeakMap<object, string>(); const wide: WeakMap<object, unknown> = m;`, []string{"WeakMap<object, string> seen as WeakMap<object, unknown>, which can write"}},
		{"a WeakMap passed as one of unknown values", `function keep(m: WeakMap<object, unknown>): void {} const m = new WeakMap<object, string>(); keep(m);`, []string{"WeakMap<object, string> seen as WeakMap<object, unknown>, which can write"}},
		{"a WeakMap assigned to one of a wider union value", `const m = new WeakMap<object, string>(); const wide: WeakMap<object, string | object> = m;`, []string{"WeakMap<object, string> seen as WeakMap<object, string | object>, which can write", "WeakMap<object, string> seen as WeakMap<object, object | string>, which can write"}},
		{"a WeakMap passed as one of a wider union value", `function keep(m: WeakMap<object, string | object>): void {} const m = new WeakMap<object, string>(); keep(m);`, []string{"WeakMap<object, string> seen as WeakMap<object, string | object>, which can write", "WeakMap<object, string> seen as WeakMap<object, object | string>, which can write"}},
		{"an Object key, which a number fits", `const m = new WeakMap<Object, number>(); m.set(1, 1);`, []string{"a WeakMap keyed by Object, which a primitive fits"}},
		{"an Object key reached through a parameter", `function put(m: WeakMap<Object, number>): void { m.set(1, 1); }`, []string{"a WeakMap keyed by Object, which a primitive fits"}},
		{"a string or null value", `const m = new WeakMap<object, string | null>();`, []string{"a WeakMap of string | null"}},
		{"a number or undefined value", `const m = new WeakMap<object, number | undefined>();`, []string{"a WeakMap of number | undefined"}},
		{"an object value", `const m = new WeakMap<object, { x: number }>();`, []string{"#q2ybjr1"}},
		{"an array value through a parameter", `function count(m: WeakMap<object, number[]>, k: object): number { return m.get(k)?.length ?? 0; }`, []string{"#q2ybjr1"}},
		{"a union value", `const m = new WeakMap<object, string | number>();`, []string{"#q2ybjr1"}},
		{"a closure value", `const m = new WeakMap<object, () => number>();`, []string{"#q2ybjr1"}},
		{"a method read as a value", `const m = new WeakMap<object, number>(); const get = m.get;`, []string{"a method read as a value"}},
		{"an iterable handed to the constructor", `const k = {}; const m = new WeakMap<object, number>([[k, 1]]);`, []string{"new WeakMap from an iterable"}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil {
				t.Fatalf("lowered; want a stop naming one of %q", probe.reasons)
			}
			for _, reason := range probe.reasons {
				if strings.Contains(err.Error(), reason) {
					return
				}
			}
			t.Fatalf("want a stop naming one of %q, got %v", probe.reasons, err)
		})
	}
}

// A WeakMap lowers to its own nodes on W1's weak collection, never a Map's: new is a WeakCollectionNew
// holding its values' representation, set a WeakMapSet with a write site, get a WeakMapGet, and has and
// delete W1's own.
func TestWeakMapLowersOnItsOwnType(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `const m = new WeakMap<object, number>();
const k = { x: 1 };
m.set(k, 1).set([1], 2);
console.log(String(m.get(k)));
console.log(String(m.has(k)));
console.log(String(m.delete(k)));
const flags = new WeakMap<number[], boolean>();
flags.set([1], true);
const names = new WeakMap<object, string>();
names.set(k, 'k');
console.log(names.get(k) ?? 'none');
`)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]int{}
	values := map[ir.Type]bool{}
	walk(program.Main, func(node any) bool {
		switch node := node.(type) {
		case ir.WeakCollectionNew:
			found["new"]++
			values[node.Value] = true
			if node.Set {
				t.Errorf("a WeakMap made as a WeakSet: %+v", node)
			}
		case ir.WeakMapSet:
			found["set"]++
			if node.Site == 0 {
				t.Errorf("a WeakMap's set with no write site: %+v", node)
			}
		case ir.WeakMapGet:
			found["get"]++
		case ir.WeakCollectionHas:
			found["has"]++
		case ir.WeakCollectionDelete:
			found["delete"]++
		case ir.MapNew, ir.MapSet, ir.MapGet, ir.MapHas, ir.MapDelete, ir.SetNew, ir.SetAdd, ir.WeakSetAdd:
			t.Errorf("a WeakMap lowered to another collection's node: %T", node)
		}
		return true
	})
	if found["new"] != 3 || found["set"] != 4 || found["get"] != 2 || found["has"] != 1 || found["delete"] != 1 {
		t.Errorf("found %v, want 3 new, 4 set, 2 get, 1 has and 1 delete", found)
	}
	for _, want := range []ir.Type{ir.Number, ir.Boolean, ir.String} {
		if !values[want] {
			t.Errorf("no WeakMap holding %v among %v", want, values)
		}
	}
}
