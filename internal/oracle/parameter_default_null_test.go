package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A parameter default is taken only for undefined, never for a passed null, as JavaScript does
// (functions.go's default prologue). On main be2242c1 every default lowered to ?? (ir.Coalesce), so
// f(null) for f(x: number | null = 1) printed 1 where Node prints null, on every backend (Codex
// 01a1c449). number | null, string | null, a default with an effect that must run once, for undefined
// only, and a control whose parameters hold no null each agree with Node natively, sanitized and on
// JavaScript. The mutant in internal/lower/parameter-default-null.mutants.md, ?? restored, turns the
// first three red.
func TestParameterDefaultNullAgrees(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/parameter_default_null"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 4 {
		t.Fatalf("%d programs in %s, want 4: a program added or removed needs this count changed with it", programs, directory)
	}
}
