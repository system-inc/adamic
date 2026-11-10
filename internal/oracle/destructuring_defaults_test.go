package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An object pattern's default lowers with JavaScript's meaning (#smqr4vn, lower's destructuredDefault):
// the field is read once, the default is evaluated only when that read is undefined, never for a
// present null, in source order, after the names before it, and the name holds its own type. Version.with's
// shape from tsc's semver.ts:124, a default with an effect, a default reading an earlier name, an
// unknown field given null, a let pattern and a destructured parameter each agree with Node natively, sanitized and on
// JavaScript. A field whose type holds null stops before its default is judged, pinned in internal/lower's
// TestDestructuringDefaultNullStops. The mutants in internal/lower/destructuring-defaults.mutants.md: the
// default evaluated eagerly turns default_effect.a red.
func TestDestructuringDefaultsAgree(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/destructuring_defaults"
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
	if programs != 6 {
		t.Fatalf("%d programs in %s, want 6: a program added or removed needs this count changed with it", programs, directory)
	}
}
