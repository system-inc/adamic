package parser

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Import-cycles 779ff9d closed the type-only import cycle gap. Keep Node and
// sanitized native output as the positive witness of that ruling.
func TestTypeOnlyImportCycleCompiles(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/type_import_cycle/entry.a")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	want := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	if string(want) != "1\n" {
		t.Fatalf("Node cycle result %q", want)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatalf("type-only cycle must compile after 779ff9d: %v", err)
	}
	binary := filepath.Join(t.TempDir(), "cycle")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := execute(t, "", binary).output
	if !bytes.Equal(got, want) {
		t.Fatalf("native %q, Node %q", got, want)
	}
}
