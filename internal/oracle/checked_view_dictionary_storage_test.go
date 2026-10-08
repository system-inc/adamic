package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

// Every source table is produced by the records lowering, including finite partial records.
func TestCheckedViewDictionaryStorage(t *testing.T) {
	for _, test := range []struct{ name, truth, stderr string }{
		{"mutable-good", "name\nundefined\n", ""},
		{"mutable-wrong", "number\n", "adamic: panic: field read failed: view.table['item']; expected string[] | undefined, found number\n"},
		{"array-shape-wrong", "object\n", "adamic: panic: field read failed: view.table['item']; expected { readonly name: string; } | undefined, found array\n"},
		{"union-shape-wrong", "object\n", "adamic: panic: field read failed: view.table['item']; expected string | number | boolean | undefined, found object\n"},
		{"number-good", "42\nundefined\n", ""},
		{"number-wrong", "number\n", "adamic: panic: field read failed: view.table['item']; expected string | undefined, found number\n"},
		{"boolean-good", "true\nundefined\n", ""},
		{"boolean-wrong", "boolean\n", "adamic: panic: field read failed: view.table['item']; expected string | undefined, found boolean\n"},
		{"string-good", "name\nundefined\n", ""},
		{"union-good", "42\nundefined\n", ""},
		{"partial-good", "42\nundefined\n", ""},
		{"partial-wrong", "number\n", "adamic: panic: field read failed: view.table['item']; expected string | undefined, found number\n"},
		{"object-good", "name\n", ""},
		{"object-wrong", "42\n", "adamic: panic: field read failed: entry.name is not a string; expected string, found number\n"},
		{"array-good", "name\n", ""},
		{"array-wrong", "42\n", "adamic: panic: element read failed: entry[0] expected string, found number\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/storage-"+test.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.truth {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if test.stderr == "" {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				} else if got.exitCode != 70 || string(got.stderr) != test.stderr {
					t.Fatalf("pinned read: %#v; want %q", got, test.stderr)
				}
			}
			if test.stderr == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// A wrong producer certificate must be rejected before scalar table slots are used.
// This program has no views, so a dictionary element check cannot mask this mutant.
func TestCheckedViewDictionaryProducerCertificateMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_operations.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	changed := strings.ReplaceAll(code, "adamic_record_new_typed(false, 1)", "adamic_record_new_typed(false, 2)")
	if changed == code {
		t.Fatal("mutant changed no producer certificate")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(changed, binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	got := execute(t, binary)
	if got.exitCode != 70 || string(got.stderr) != "adamic: panic: record operation failed: source storage does not match the declared element representation\n" {
		t.Fatalf("producer certificate mutant survived: %#v", got)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 {
		t.Fatalf("Node source: %#v", truth)
	}
	t.Log("wrong producer certificate caught only by record storage check, exit 70")
}
