package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A real Map, Set or array is held as itself, but tsc lets it go where a structural object type is held: as
// an object, or as a union where a string would fit the type too. On main e4c3a163 it lowered there
// (#nw6d7ws): a Set as `{ has(x: string): boolean }`, a Map as `interface Lookup { get(key: string): number
// | undefined }` and an array as `{ indexOf(x: number): number }` each exited 70 on the JavaScript backend
// where Node prints, and natively clang refused the C, or, through a union-held shape, the build ran and
// crashed. Ruled by Adamic: NotYet at the value, keyed on representation, unless the slot reads it only by
// members shape_fit.go's one table serves, judged in #efx3yb9's walk (container_fit.go). Each probe pins the
// whole stop, .a and .ts: the r01, r02, r04, r06 and r07 shapes (r03 works and is a control below; r05, a
// Promise, stops at its own refusal and isn't here), a Set's size, an array's includes, an array as a shape
// reading length beside push (held as an object, where even length fails in clang), an array as a union-held
// shape that reads only its length but declares more, `name?: string` or `at(i: number)`, both agreeing with
// Node on main (Adamic's ruling: declared members, failing closed; they're the witnesses for a read-set
// proof if one is ever built), an array as a union-held `{ indexOf(x: string): number }` and a Map as
// `{ readonly size: number } | string` (both crashed natively), and the roads a Map takes there: a return, a
// field write, an element write, push (of a Map held in a name, and of a Map made right there, where push's
// T is the program's Lookup), Map.set's value where the Map's V is Lookup, an object literal's field, an
// intersection's field, written as a literal and seen through (both crashed natively, as did an array
// there), a callback's result, a cast, a Map's type arguments, a function value whose parameter is the
// shape, and a Map.forEach callback whose parameter is (it segfaulted natively). Past the walk's depth the
// converse fails closed (shapeFitReached): an array five readonly fields down stops at its leaf, and six
// down too. The two library interfaces exempt as a library call's own argument (shapeFitProtocols) each have
// a pin in a program's own slot: a Set where a program's ReadonlySetLike is held, as a parameter and an
// initializer, stops here, and so does an array where a program's Iterable<number> parameter is held, whose
// for...of lowers as a view (#qk8rztp S2b-1b, lower/iterable_views.go). The table's one entry, an array where
// `{ readonly length: number }` is held, is a control that lowers and agrees with Node on sanitized, release
// and JavaScript, through an argument, an initializer read after a push, an object literal's field, a return,
// an optional slot and five readonly fields down, and so are a plain object seen that way, a readonly array,
// an array six fields down where an array is held, Set.union and isSubsetOf with a Set (the prelude's
// ReadonlySetLike), and new Set(xs) and new Map(entries) (Iterable). Mutants (each red, the controls green):
// the converse leaf answering "" (every probe but depth_6, which shapeFitReached still stops;
// iterable_parameter then stops at its view's proof, an array being no origin the view names), the length
// entry dropped (the length controls), shapeFitProtocols emptied (set_union, set_from_array,
// map_from_entries), shapeFitReached answering "" (depth_6), eachReached cutting a library anonymous type off
// (foreach_parameter), and an intersection skipped (intersection_view_map, intersection_view_array).
func TestContainerIsNotYetWhereAShapeIsHeld(t *testing.T) {
	t.Parallel()
	const mapStop = "a Map or Set going where a plain object type is held"
	const arrayStop = "an array going where a plain object type is held"
	const lookup = "interface Lookup { get(key: string): number | undefined; }\nconst m = new Map<string, number>([['a', 1]]);\n"
	const other = "const other: Lookup = { get: (key: string) => (key === 'a' ? 2 : undefined) };\n"
	// Readonly class fields five and six deep, so the view is one the program can't write through.
	const nested = "class D { readonly e = new E(); }\nclass C { readonly d = new D(); }\nclass B { readonly c = new C(); }\nclass A { readonly b = new B(); }\n"
	const deep = "class E { readonly f: number[] = [1, 2]; }\n" + nested
	const deeper = "class F { readonly g: number[] = [1, 2]; }\nclass E { readonly f = new F(); }\n" + nested
	for _, probe := range []struct{ name, source, node, where, stop string }{
		{"r01_set_has", "const members = new Set<string>(['k']);\nfunction useShape(v: { has(x: string): boolean }): boolean { return v.has('k'); }\nconsole.log(String(useShape(members)));\n", "true\n", ":3:29", mapStop},
		{"r02_set_size_initializer", "const members = new Set<string>(['k']);\nconst v: { readonly size: number } = members;\nconsole.log(String(v.size));\n", "1\n", ":2:38", mapStop},
		{"length_beside_push", "function count(v: { readonly length: number; push(x: number): number }): number { return v.length; }\nconsole.log(String(count([1, 2, 3])));\n", "3\n", ":2:26", arrayStop},
		{"r04_array_index_of", "function first(v: { indexOf(x: number): number }): number { return v.indexOf(2); }\nconsole.log(String(first([1, 2, 3])));\n", "1\n", ":2:26", arrayStop},
		{"r06_map_size", "const m = new Map<string, number>([['a', 1]]);\nfunction sizeOf(v: { readonly size: number }): number { return v.size; }\nconsole.log(String(sizeOf(m)));\n", "1\n", ":3:27", mapStop},
		{"r07_map_get", lookup + "const v: Lookup = m;\nconsole.log(String(v.get('a')));\n", "1\n", ":3:19", mapStop},
		{"set_size", "const s = new Set<string>(['a', 'b']);\nfunction sizeOf(v: { readonly size: number }): number { return v.size; }\nconsole.log(String(sizeOf(s)));\n", "2\n", ":3:27", mapStop},
		{"length_beside_name", "function count(v: { readonly length: number; name?: string }): number { return v.length; }\nconsole.log(String(count([1, 2, 3])));\n", "3\n", ":2:26", arrayStop},
		{"length_beside_at", "function count(v: { readonly length: number; at(i: number): string | undefined }): number { return v.length; }\nconsole.log(String(count(['a', 'b'])));\n", "2\n", ":2:26", arrayStop},
		{"array_includes", "function has(v: { includes(x: number): boolean }): boolean { return v.includes(2); }\nconsole.log(String(has([1, 2, 3])));\n", "true\n", ":2:24", arrayStop},
		{"array_union_held", "function find(v: { indexOf(x: string): number }): number { return v.indexOf('b'); }\nconsole.log(String(find(['a', 'b'])));\n", "1\n", ":2:25", arrayStop},
		{"map_union_slot", "const m = new Map<string, number>([['a', 1]]);\nfunction sizeOf(v: { readonly size: number } | string): number { return typeof v === 'string' ? -1 : v.size; }\nconsole.log(String(sizeOf(m)));\n", "1\n", ":3:27", mapStop},
		{"return", lookup + "function make(): Lookup { return m; }\nconsole.log(String(make().get('a')));\n", "1\n", ":3:34", mapStop},
		{"field_write", lookup + other + "const holder: { v: Lookup } = { v: other };\nholder.v = m;\nconsole.log(String(holder.v.get('a')));\n", "1\n", ":5:12", mapStop},
		{"element_write", lookup + other + "const items: Lookup[] = [other];\nitems[0] = m;\nconsole.log(String(items[0]?.get('a')));\n", "1\n", ":5:12", mapStop},
		{"push", lookup + "const items: Lookup[] = [];\nitems.push(m);\nconsole.log(String(items[0]?.get('a')));\n", "1\n", ":4:12", mapStop},
		{"object_field", lookup + "const holder: { v: Lookup } = { v: m };\nconsole.log(String(holder.v.get('a')));\n", "1\n", ":3:36", mapStop},
		{"callback", lookup + "function run(make: () => Lookup): string { return String(make().get('a')); }\nconsole.log(run(() => m));\n", "1\n", ":4:23", mapStop},
		{"cast", lookup + "const v = m as Lookup;\nconsole.log(String(v.get('a')));\n", "1\n", ":3:11", mapStop},
		{"map_type_arguments", lookup + "const inner = new Map<string, Map<string, number>>([['k', m]]);\nconst outer: ReadonlyMap<string, Lookup> = inner;\nconsole.log(String(outer.get('k')?.get('a')));\n", "1\n", ":4:44", mapStop},
		{"foreach_parameter", lookup + "new Map<string, Map<string, number>>([['k', m]]).forEach((v: Lookup) => {\n\tconsole.log(String(v.get('a')));\n});\n", "1\n", ":3:58", mapStop},
		{"setlike_parameter", "function hasOne(s: ReadonlySetLike<number>): boolean { return s.has(1); }\nconsole.log(String(hasOne(new Set<number>([1]))));\n", "true\n", ":2:27", mapStop},
		{"setlike_initializer", "const s: ReadonlySetLike<number> = new Set<number>([1, 2]);\nconsole.log(String(s.size));\n", "2\n", ":1:36", mapStop},
		{"push_literal", "interface Lookup { get(key: string): number | undefined; }\nconst items: Lookup[] = [];\nitems.push(new Map([['a', 1]]));\nconsole.log(String(items[0]?.get('a')));\n", "1\n", ":3:12", mapStop},
		{"map_set_lookup", "interface Lookup { get(key: string): number | undefined; }\nconst m = new Map<string, Lookup>().set('k', new Map());\nconsole.log(String(m.get('k')?.get('a')));\n", "undefined\n", ":2:46", mapStop},
		{"intersection_field", lookup + "const holder: { v: Lookup } & { tag?: string } = { v: m };\nconsole.log(String(holder.v.get('a')));\n", "1\n", ":3:55", mapStop},
		{"intersection_view_map", lookup + "const source = { v: m, tag: 'x' };\nconst holder: { readonly v: Lookup } & { readonly tag: string } = source;\nconsole.log(String(holder.v.get('a')));\n", "1\n", ":4:67", mapStop},
		{"intersection_view_array", "const source = { v: [1, 2], tag: 'x' };\nconst holder: { readonly v: { indexOf(x: number): number } } & { readonly tag: string } = source;\nconsole.log(String(holder.v.indexOf(2)));\n", "1\n", ":2:91", arrayStop},
		{"depth_5", deep + "function use(v: { readonly b: { readonly c: { readonly d: { readonly e: { readonly f: { indexOf(x: number): number } } } } } }): number {\n\treturn v.b.c.d.e.f.indexOf(2);\n}\nconsole.log(String(use(new A())));\n", "1\n", ":9:24", arrayStop},
		{"depth_6", deeper + "function use(v: { readonly b: { readonly c: { readonly d: { readonly e: { readonly f: { readonly g: { indexOf(x: number): number } } } } } } }): number {\n\treturn v.b.c.d.e.f.g.indexOf(2);\n}\nconsole.log(String(use(new A())));\n", "1\n", ":10:24", arrayStop},
		{"iterable_parameter", "function total(xs: Iterable<number>): number {\n\tlet sum = 0;\n\tfor (const x of xs) {\n\t\tsum += x;\n\t}\n\treturn sum;\n}\nconsole.log(String(total([1, 2, 3])));\n", "6\n", ":8:26", arrayStop},
		{"function_value", "const take = (v: number[]): number => v.indexOf(2);\nconst f: (v: number[]) => number = take;\nconst g: (v: { indexOf(x: number): number }) => number = (v) => v.indexOf(2);\nconsole.log(String(f([1, 2]) + g([1, 2])));\n", "2\n", ":4:34", arrayStop},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), probe.name+extension)
				if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
					t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
				_, err := lowered(t, path)
				var notYet *lower.NotYet
				want := path + probe.where + ": stage 0 can't lower " + probe.stop + " yet"
				if !errors.As(err, &notYet) || err.Error() != want {
					t.Fatalf("want NotYet %q, got %v", want, err)
				}
			})
		}
	}
	for _, control := range []struct{ name, source, node string }{
		{"length_argument", "function count(v: { readonly length: number }): number { return v.length; }\nconsole.log(String(count([1, 2, 3])));\n", "3\n"},
		{"length_initializer", "const xs = [1, 2, 3];\nconst v: { readonly length: number } = xs;\nxs.push(4);\nconsole.log(String(v.length));\n", "4\n"},
		{"length_object_field", "const holder: { items: { readonly length: number } } = { items: ['a', 'b'] };\nconsole.log(String(holder.items.length));\n", "2\n"},
		{"length_return", "function make(): { length: number } { return [1, 2]; }\nconsole.log(String(make().length));\n", "2\n"},
		{"length_optional", "function count(v: { readonly length: number } | undefined): number { return v === undefined ? 0 : v.length; }\nconsole.log(String(count([1, 2, 3]) + count(undefined)));\n", "3\n"},
		{"length_plain_object", "interface Sized { readonly length: number; readonly label: string }\nfunction count(v: { readonly length: number }): number { return v.length; }\nconst s: Sized = { length: 2, label: 'x' };\nconsole.log(String(count(s) + count([1])));\n", "3\n"},
		{"readonly_array", "function count(v: readonly number[]): number { return v.length + v.indexOf(2); }\nconsole.log(String(count([1, 2, 3])));\n", "4\n"},
		{"depth_5_length", deep + "function use(v: { readonly b: { readonly c: { readonly d: { readonly e: { readonly f: { readonly length: number } } } } } }): number {\n\treturn v.b.c.d.e.f.length;\n}\nconsole.log(String(use(new A())));\n", "2\n"},
		{"depth_6_array", "const xs = [1, 2];\nconst value = { a: { b: { c: { d: { e: { f: xs } } } } } };\nconst view: { a: { b: { c: { d: { e: { f: number[] } } } } } } = value;\nconsole.log(String(view.a.b.c.d.e.f.length));\n", "2\n"},
		{"map_from_entries", "const entries: [string, number][] = [['a', 1]];\nconst m = new Map<string, number>(entries);\nconsole.log(String(m.get('a')));\n", "1\n"},
		{"set_union", "const left = new Set<number>([1, 2]);\nconst right: ReadonlySet<number> = new Set<number>([2, 3]);\nconsole.log(`${left.union(right).size} ${left.isSubsetOf(right)}`);\n", "3 false\n"},
		{"set_from_array", "const xs = [1, 2, 2];\nconst s = new Set<number>(xs);\nconsole.log(String(s.size));\n", "2\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(control.name+extension, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), control.name+extension)
				if err := os.WriteFile(path, []byte(control.source), 0o644); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if disagreement(run{stdout: []byte(control.node)}, truth) != "" {
					t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatalf("want it lowered, got %v", err)
				}
				sanitized, _ := nativelyUncached(t, program)
				for backend, got := range map[string]run{"sanitized": sanitized, "release": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
					if difference := disagreement(truth, got); difference != "" {
						t.Errorf("%s: %s, stderr %q", backend, difference, got.stderr)
					}
				}
			})
		}
	}
}
