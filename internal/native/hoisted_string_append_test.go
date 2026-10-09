package native

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestHoistedStringAppendPreservesAliasesAndEffects(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join("testdata", "hoisted_string_append.a"))
	if err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "append")
	if err := Build(C(program), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join("..", "..", "oracle", "node.mjs"), path)
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("native %q, Node %q", got, want)
	}
}
