package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The checked dictionary view (lower's dictionaryView, #2v3xtgm, re-ruled Nov 23): a value typed object cast to
// MapLike<unknown> is tested at the cast, natively and on the JavaScript backend alike, for a record whose values are
// references, which every read through the view takes as a Union as it is. These hold it against Node.

// dictionaryViewPath is a fixture of records_dictionary_view, absolute.
func dictionaryViewPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_dictionary_view", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// A record of references read through the view agrees with Node on both backends, with nothing leaked: tsc's
// moduleSpecifiers.ts shape over a record built by key of a string, an object, an array and undefined; a
// Record<string, string> whose raw string pointers read as a Union as they are; and a Record<string, string |
// undefined> whose undefined sits in a reference slot (Language's pin). Node's output is pinned so a fixture can't
// pass by printing nothing.
func TestDictionaryViewAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, printed := range map[string]string{
		"walk_by_key.ts":     "import=string ./index.mjs\ntypes=object\nfiles=object\nmissing=undefined\ndone\n",
		"string_record.ts":   "main=./index.js\ntypes=./index.d.ts\ndone\n",
		"optional_record.ts": "a=string x\nb=undefined\ndone\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := dictionaryViewPath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("object seen as MapLike<unknown> in .ts must lower: %v", err)
			}
			native, sanitized := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, optionsBaseShown(node), backend, optionsBaseShown(got))
				}
			}
			if report := leaks(t, program, sanitized); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// What the view's test stops, at the cast, on both backends alike, after the line before it, a run-time NotYet where
// Node runs on to the end: an array (Compiler's pin), an object literal's fixed object and a class instance in the
// first words, a value not held as a record; a record of numbers and one of booleans in the second, a record whose
// values aren't references. Mutants: an unchecked narrow at the cast, and the array, literal and class rows hand the
// value on as Node does, or crash natively; the references test dropped, and the number and boolean rows read a
// scalar's bits as a pointer natively; the JavaScript mark always "reference", and the number and boolean rows read
// through on the JavaScript backend.
func TestDictionaryViewStopsAtTheCast(t *testing.T) {
	t.Parallel()
	for name, stop := range map[string]struct{ node, stop string }{
		"array_cast.ts":     {"before\nstring\nafter\n", ir.DictionaryViewRecordStop},
		"fixed_literal.ts":  {"before\nstring\nafter\n", ir.DictionaryViewRecordStop},
		"class_instance.ts": {"before\nstring\nafter\n", ir.DictionaryViewRecordStop},
		"number_record.ts":  {"before\nnumber\nafter\n", ir.DictionaryViewReferencesStop},
		"boolean_record.ts": {"before\nboolean\nafter\n", ir.DictionaryViewReferencesStop},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := dictionaryViewPath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != stop.node {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, _ := natively(t, program)
			want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + stop.stop + "\n"), exitCode: 70}
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: want %s; got %s", backend, difference, optionsBaseShown(want), optionsBaseShown(got))
				}
			}
		})
	}
}
