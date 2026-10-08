package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstantTablesMatchNode(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "tables.a")
	source, err := os.ReadFile("testdata/constant_tables.a")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, source, 0644); err != nil {
		t.Fatal(err)
	}
	code := C(namedProgram(t, entry, nil))
	if !strings.Contains(code, "static const double adamic_number_table_") || !strings.Contains(code, "static const adamic_value adamic_record_table_") {
		t.Fatal("fixture did not emit constant templates")
	}
	oracle, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", oracle, entry)
	for _, split := range []bool{false, true} {
		for _, sanitize := range []bool{false, true} {
			binary := filepath.Join(directory, "tables")
			if err := Build(code, binary, Options{Split: split, Jobs: 3, Sanitize: sanitize}); err != nil {
				t.Fatal(err)
			}
			if got := runWithInput(t, "", binary); got != want {
				t.Fatalf("split=%t sanitize=%t: native %q Node %q", split, sanitize, got, want)
			}
		}
	}
}
