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
// a value, an iterable handed to the constructor, and every WeakMap, until its get and set are built.
// Mutants: weakCollectionKey without its noPrimitiveFits test, and the String and Object rows lower;
// containers() without WeakMap and WeakSet, and the widening rows no longer stop in invariance's words.
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
		{"a method read as a value", `const s = new WeakSet<object>(); const add = s.add;`, []string{"a method read as a value"}},
		{"an iterable handed to the constructor", `const k = {}; const s = new WeakSet<object>([k]);`, []string{"new WeakSet from an iterable"}},
		{"a WeakMap", `const m = new WeakMap<object, number>();`, []string{"a WeakMap (its get and set aren't built yet, #hcgy751)"}},
		{"a WeakMap's method through a parameter", `function has(m: WeakMap<object, number>, k: object): boolean { return m.has(k); }`, []string{"a WeakMap (its get and set aren't built yet, #hcgy751)"}},
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
