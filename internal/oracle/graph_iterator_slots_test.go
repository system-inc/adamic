package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A collection's iterator stored back into what the collection holds closes a cycle, whatever the slot's type
// (#66acn2a): the review programs' Iterator<string> and IterableIterator<string>, Iterable<string>, a program's own
// Cursor interface, a type alias, a Set's iterator and a Map's entries. The iterator holds its collection, which
// cycles.go's reachLinks links by assignability (libraryIteratorCaptures), so each is a graph program freed as one
// region: held to Node, and balanced by the counted build. Before the edge each counted and leaked 906 bytes.
//
// Iterator<string> is Iterator<string, any, any> by lib's defaults, which the cycle finder took for a template that
// holds nothing until an any by default stopped counting (anyByDefault). Mutants: reachLinks' default arm without
// libraryIteratorCaptures, or deriveTemplate's any test back on every argument, and every row but the control leaks
// by the counted build.
func TestGraphIteratorSlotsByAssignability(t *testing.T) {
	t.Parallel()
	holder := func(slot, made string) string {
		return "type Holder = { cursor: " + slot + " | undefined };\nconst map = new Map<string, Holder>();\nconst holder: Holder = { cursor: undefined };\nmap.set('a'.repeat(2), holder);\nholder.cursor = " + made + ";\nconsole.log(`${map.size}`);\n"
	}
	cases := []struct{ name, source string }{
		{"an Iterator<string> field", holder("Iterator<string>", "map.keys()")},
		{"an IterableIterator<string> field", holder("IterableIterator<string>", "map.keys()")},
		{"an Iterable<string> field", holder("Iterable<string>", "map.keys()")},
		{"a program's Cursor interface", "interface Step { readonly done?: boolean; readonly value: string | undefined }\ninterface Cursor { next(): Step }\n" + holder("Cursor", "map.keys()")},
		{"a type alias", "type Keys = Iterator<string>;\n" + holder("Keys", "map.keys()")},
		{"a Map's entries", holder("Iterator<[string, Holder]>", "map.entries()")},
		{"a Set's values", "type Holder = { cursor: Iterator<Holder> | undefined };\nconst set = new Set<Holder>();\nconst holder: Holder = { cursor: undefined };\nset.add(holder);\nholder.cursor = set.values();\nconsole.log(`${set.size}`);\n"},
		{"the control, no iterator", "type Holder = { size: number };\nconst map = new Map<string, Holder>();\nconst holder: Holder = { size: 0 };\nmap.set('a'.repeat(2), holder);\nholder.size = map.size;\nconsole.log(`${map.size}`);\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, _ := natively(t, program)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), counted, native_runtime.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			ran := execute(t, counted)
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
				t.Fatal(report)
			}
		})
	}
}
