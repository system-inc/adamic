package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A ReadonlyMap or ReadonlySet slot is held as a map, and every operation on it is the map's, but tsc lets
// any value with the interface's members go there. On main e4c3a163 an object literal or a class instance
// shaped like one lowered (#efx3yb9): `m.has('a')` exited 70 on the JavaScript backend ("local_1_m.has is not
// a function") where Node prints true, native clang refused the C that stored an object pointer in a map
// slot, and through a field write the native build ran and crashed. Ruled by Adamic: one check keyed on
// representation, NotYet at the value on every road into the slot (container_fit.go). Each probe pins the
// whole stop, .a and .ts, one per road (Compiler's condition on #efx3yb9's proof): a ReadonlyMap literal and
// a ReadonlySet literal as initializers, a shaped value or a literal passed as an argument, returned,
// assigned, written to a field, to an array element, in an array literal and through push, given as a Map's
// value (set) and inside a Map's type arguments, as a record literal's value, spread into a typed object, as
// a class field's initializer, and through a function value whose parameter is the shape, and a class
// instance implementing ReadonlySet. Two roads stop first at their own stops, on main too, and are pinned
// with them: a spread into a typed record (records.go's NotYet) and a class instance's field spread into a
// typed object (adamic/nominal-class). Depth: a class implementing ReadonlyMap five fields down lowered on
// main (native crashed, JavaScript printed false); the map slot is judged at any depth, so depth_5 stops
// at its leaf and a real Map there (depth_5_real_map, a control) lowers. Six down the walk is past its
// bound and fails closed: depth_6 stops, and so does depth_6_real_map, a real Map six fields into a literal
// (the cost of failing closed, said so). The controls lower and agree with Node on sanitized, release and
// JavaScript: a real Map seen as a ReadonlyMap, a real Set seen as a ReadonlySet, an array seen as a
// ReadonlyArray, and a real Map five fields down.
// Mutant: containerWhereOtherGoes answering false lowers the probes that stop at it, each red; the two
// pinned at their own stops stay green.
func TestContainerInterfaceValueIsNotYetWhereAMapIsHeld(t *testing.T) {
	t.Parallel()
	const mapMembers = "\tget(key: string) { return backing.get(key); },\n" +
		"\thas(key: string) { return key === 'a'; },\n" +
		"\tforEach() {},\n" +
		"\tsize: 1,\n" +
		"\tkeys() { return backing.keys(); },\n" +
		"\tvalues() { return backing.values(); },\n" +
		"\tentries() { return backing.entries(); },\n" +
		"\t[Symbol.iterator]() { return backing.entries(); },\n"
	const backing = "const backing = new Map<string, number>([['a', 1]]);\n"
	const shaped = backing + "const shaped = {\n" + mapMembers + "};\n"
	setMembers := []string{
		"has(value: number): boolean { return value === 1; }",
		"forEach() {}",
		"keys() { return new Set<number>([1]).keys(); }",
		"values() { return new Set<number>([1]).values(); }",
		"entries() { return new Set<number>([1]).entries(); }",
		"[Symbol.iterator]() { return new Set<number>([1]).values(); }",
		"union(): Set<number> { return new Set<number>(); }",
		"intersection<U>(): Set<number & U> { throw new Error('unused'); }",
		"difference(): Set<number> { return new Set<number>(); }",
		"symmetricDifference(): Set<number> { return new Set<number>(); }",
		"isSubsetOf(): boolean { return false; }",
		"isSupersetOf(): boolean { return false; }",
		"isDisjointFrom(): boolean { return false; }",
	}
	shapedSet := "const shaped = {\nsize: 1,\n" + strings.Join(setMembers, ",\n") + ",\n};\n"
	const readSet = "if (s !== undefined) {\n\tconsole.log(String(s.has(1)));\n}\n"
	const eachSet = "for (const s of items) {\n\tconsole.log(String(s.has(1)));\n}\n"
	const containerFit = "stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet"
	const fakeMap = "class FakeMap implements ReadonlyMap<string, number> {\n" +
		"\treadonly size = 1;\n" +
		"\tget(key: string): number | undefined { return backing.get(key); }\n" +
		"\thas(key: string): boolean { return key === 'a'; }\n" +
		"\tforEach(): void {}\n" +
		"\tkeys() { return backing.keys(); }\n" +
		"\tvalues() { return backing.values(); }\n" +
		"\tentries() { return backing.entries(); }\n" +
		"\t[Symbol.iterator]() { return backing.entries(); }\n" +
		"}\n"
	for _, probe := range []struct{ name, source, where, stop string }{
		{"map_literal", backing + "const m: ReadonlyMap<string, number> = {\n" + mapMembers + "};\nconsole.log(String(m.has('a')));\n", ":2:40", containerFit},
		{"set_literal", "const s: ReadonlySet<number> = {\nsize: 1,\n" + strings.Join(setMembers, ",\n") + ",\n};\nconsole.log(String(s.has(1)));\n", ":1:32", containerFit},
		{"parameter", shaped + "function probe(m: ReadonlyMap<string, number>): boolean {\n\treturn m.has('a');\n}\nconsole.log(String(probe(shaped)));\n", ":15:26", containerFit},
		{"return", shaped + "function make(): ReadonlyMap<string, number> {\n\treturn shaped;\n}\nconsole.log(String(make().has('a')));\n", ":13:9", containerFit},
		{"field", shaped + "const holder: { m: ReadonlyMap<string, number> } = { m: new Map<string, number>() };\nholder.m = shaped;\nconsole.log(String(holder.m.has('a')));\n", ":13:12", containerFit},
		{"set_class", "class Shaped {\nreadonly size = 1;\n" + strings.Join(setMembers, "\n") + "\n}\nconst s: ReadonlySet<number> = new Shaped();\nconsole.log(String(s.has(1)));\n", ":17:32", containerFit},
		{"assignment", shapedSet + "let s: ReadonlySet<number> = new Set<number>();\ns = shaped;\nconsole.log(String(s.has(1)));\n", ":18:5", containerFit},
		{"element", shapedSet + "const items: ReadonlySet<number>[] = [new Set<number>()];\nitems[0] = shaped;\n" + eachSet, ":18:12", containerFit},
		{"push", shapedSet + "const items: ReadonlySet<number>[] = [];\nitems.push(shaped);\n" + eachSet, ":18:12", containerFit},
		{"map_value", shapedSet + "const m = new Map<string, ReadonlySet<number>>().set('k', shaped);\nconst s = m.get('k');\n" + readSet, ":17:59", containerFit},
		{"map_type_arguments", shapedSet + "const inner = new Map<string, typeof shaped>();\ninner.set('k', shaped);\nconst m: ReadonlyMap<string, ReadonlySet<number>> = inner;\nconst s = m.get('k');\n" + readSet, ":19:53", containerFit},
		{"record_value", shapedSet + "const record: Record<string, ReadonlySet<number>> = { a: shaped };\nconst s = record['a'];\n" + readSet, ":17:58", containerFit},
		{"spread_object", shapedSet + "const source = { s: shaped };\nconst holder: { s: ReadonlySet<number> } = { ...source };\nconsole.log(String(holder.s.has(1)));\n", ":18:49", containerFit},
		{"spread_record", shapedSet + "const source = { a: shaped };\nconst record: Record<string, ReadonlySet<number>> = { ...source };\nconst s = record['a'];\n" + readSet, ":18:55", "stage 0 can't lower a spread into record storage from a fixed object yet"},
		{"class_field", shapedSet + "class Holder {\n\ts: ReadonlySet<number> = shaped;\n}\nconsole.log(String(new Holder().s.has(1)));\n", ":18:27", containerFit},
		{"function_value", shapedSet + "const take = (s: ReadonlySet<number>): boolean => s.has(1);\nconst f: (s: typeof shaped) => boolean = take;\nconsole.log(String(f(shaped)));\n", ":18:42", containerFit},
		{"assignment_literal", backing + "let m: ReadonlyMap<string, number> = backing;\nm = {\n" + mapMembers + "};\nconsole.log(String(m.has('a')));\n", ":3:5", containerFit},
		{"array_literal_element", backing + "const xs: ReadonlyMap<string, number>[] = [{\n" + mapMembers + "}];\nfor (const m of xs) {\n\tconsole.log(String(m.has('a')));\n}\n", ":2:44", containerFit},
		{"push_literal", backing + "const xs: ReadonlyMap<string, number>[] = [];\nxs.push({\n" + mapMembers + "});\nfor (const m of xs) {\n\tconsole.log(String(m.has('a')));\n}\n", ":3:9", containerFit},
		{"map_value_literal", "const m = new Map<string, ReadonlySet<number>>().set('k', {\nsize: 1,\n" + strings.Join(setMembers, ",\n") + ",\n});\nconst s = m.get('k');\n" + readSet, ":1:59", containerFit},
		{"spread_class", "class FakeSet {\nreadonly size = 1;\n" + strings.Join(setMembers, "\n") + "\n}\nconst src = { m: new FakeSet() };\nconst r: { m: ReadonlySet<number> } = { ...src };\nconsole.log(String(r.m.has(1)));\n", ":18:39", "Adamic 0.1 refuses a value without nominal ancestry seen as FakeSet; construct that class or a subclass; use an interface for structural values (adamic/nominal-class)"},
		{"depth_5", backing + fakeMap + "const v: { a: { b: { c: { d: { e: FakeMap } } } } } = { a: { b: { c: { d: { e: new FakeMap() } } } } };\nconst w: { readonly a: { readonly b: { readonly c: { readonly d: { readonly e: ReadonlyMap<string, number> } } } } } = v;\nconsole.log(String(w.a.b.c.d.e.has('a')));\n", ":13:120", containerFit},
		{"depth_6", backing + fakeMap + "const v: { a: { b: { c: { d: { e: { f: FakeMap } } } } } } = { a: { b: { c: { d: { e: { f: new FakeMap() } } } } } };\nconst w: { readonly a: { readonly b: { readonly c: { readonly d: { readonly e: { readonly f: ReadonlyMap<string, number> } } } } } } = v;\nconsole.log(String(w.a.b.c.d.e.f.has('a')));\n", ":13:136", containerFit},
		{"depth_6_real_map", "const real = new Map<string, number>([['a', 1]]);\nconst v: { a: { b: { c: { d: { e: { f: Map<string, number> } } } } } } = { a: { b: { c: { d: { e: { f: real } } } } } };\nconst w: { readonly a: { readonly b: { readonly c: { readonly d: { readonly e: { readonly f: ReadonlyMap<string, number> } } } } } } = v;\nconsole.log(String(w.a.b.c.d.e.f.has('a')));\n", ":2:74", containerFit},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), probe.name+extension)
				if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if disagreement(run{stdout: []byte("true\n")}, truth) != "" {
					t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
				_, err := lowered(t, path)
				var notYet *lower.NotYet
				var refused *lower.Refused
				want := path + probe.where + ": " + probe.stop
				if !errors.As(err, &notYet) && !errors.As(err, &refused) || err.Error() != want {
					t.Fatalf("want the stop %q, got %v", want, err)
				}
			})
		}
	}
	for _, control := range []struct{ name, source, node string }{
		{"real_map", "const real = new Map<string, number>([['a', 1]]);\nconst m: ReadonlyMap<string, number> = real;\nfunction probe(view: ReadonlyMap<string, number>): boolean {\n\treturn view.has('a');\n}\nconsole.log(String(m.has('a') && probe(real)));\nfor (const [key, value] of m) {\n\tconsole.log(`${key} ${value}`);\n}\n", "true\na 1\n"},
		{"real_set", "const real = new Set<number>([1, 2]);\nconst s: ReadonlySet<number> = real;\nfunction probe(view: ReadonlySet<number>): number {\n\treturn view.size;\n}\nconsole.log(String(s.has(2) && probe(real) === 2));\nfor (const value of s) {\n\tconsole.log(String(value));\n}\n", "true\n1\n2\n"},
		{"depth_5_real_map", "const real = new Map<string, number>([['a', 1]]);\nconst v: { a: { b: { c: { d: { e: Map<string, number> } } } } } = { a: { b: { c: { d: { e: real } } } } };\nconst w: { readonly a: { readonly b: { readonly c: { readonly d: { readonly e: ReadonlyMap<string, number> } } } } } = v;\nconsole.log(String(w.a.b.c.d.e.has('a')));\n", "true\n"},
		{"array", "const xs: ReadonlyArray<number> = [10, 20];\nlet total = 0;\nfor (const x of xs) {\n\ttotal += x;\n}\nconsole.log(String(total + xs.length));\n", "32\n"},
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
