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
func TestTypeOnlyImportCycleCompiles(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("testdata/type_import_cycle/entry.a")
	if err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	want := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	if string(want) != "1\n" {
		test.Fatalf("Node cycle result %q", want)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	if _, err := lower.Lower(context.Background(), program); err != nil {
		test.Fatalf("type-only cycle must compile after 779ff9d: %v", err)
	}
	binary := gapNative(
		test,
		"testdata/type_import_cycle/entry.a",
		"testdata/type_import_cycle",
		native.Options{Sanitize: true},
	)
	got := execute(test, "", binary).output
	if !bytes.Equal(got, want) {
		test.Fatalf("native %q, Node %q", got, want)
	}
}
