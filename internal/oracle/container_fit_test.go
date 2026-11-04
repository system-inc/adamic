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
// whole stop, .a and .ts: a ReadonlyMap literal and a ReadonlySet literal written in place, a shaped value
// passed as a parameter, returned, and written to a field, and a class instance implementing ReadonlySet. The
// controls lower and agree with Node on sanitized, release and JavaScript: a real Map seen as a ReadonlyMap,
// a real Set seen as a ReadonlySet, and an array seen as a ReadonlyArray.
// Mutant: containerWhereOtherGoes answering false lowers every probe again, and each goes red.
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
	for _, probe := range []struct{ name, source, where string }{
		{"map_literal", backing + "const m: ReadonlyMap<string, number> = {\n" + mapMembers + "};\nconsole.log(String(m.has('a')));\n", ":2:40"},
		{"set_literal", "const s: ReadonlySet<number> = {\nsize: 1,\n" + strings.Join(setMembers, ",\n") + ",\n};\nconsole.log(String(s.has(1)));\n", ":1:32"},
		{"parameter", shaped + "function probe(m: ReadonlyMap<string, number>): boolean {\n\treturn m.has('a');\n}\nconsole.log(String(probe(shaped)));\n", ":15:26"},
		{"return", shaped + "function make(): ReadonlyMap<string, number> {\n\treturn shaped;\n}\nconsole.log(String(make().has('a')));\n", ":13:9"},
		{"field", shaped + "const holder: { m: ReadonlyMap<string, number> } = { m: new Map<string, number>() };\nholder.m = shaped;\nconsole.log(String(holder.m.has('a')));\n", ":13:12"},
		{"set_class", "class Shaped {\nreadonly size = 1;\n" + strings.Join(setMembers, "\n") + "\n}\nconst s: ReadonlySet<number> = new Shaped();\nconsole.log(String(s.has(1)));\n", ":17:32"},
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
				want := path + probe.where + ": stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet"
				if !errors.As(err, &notYet) || err.Error() != want {
					t.Fatalf("want NotYet %q, got %v", want, err)
				}
			})
		}
	}
	for _, control := range []struct{ name, source, node string }{
		{"real_map", "const real = new Map<string, number>([['a', 1]]);\nconst m: ReadonlyMap<string, number> = real;\nfunction probe(view: ReadonlyMap<string, number>): boolean {\n\treturn view.has('a');\n}\nconsole.log(String(m.has('a') && probe(real)));\nfor (const [key, value] of m) {\n\tconsole.log(`${key} ${value}`);\n}\n", "true\na 1\n"},
		{"real_set", "const real = new Set<number>([1, 2]);\nconst s: ReadonlySet<number> = real;\nfunction probe(view: ReadonlySet<number>): number {\n\treturn view.size;\n}\nconsole.log(String(s.has(2) && probe(real) === 2));\nfor (const value of s) {\n\tconsole.log(String(value));\n}\n", "true\n1\n2\n"},
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
