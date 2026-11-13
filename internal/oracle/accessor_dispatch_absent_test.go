package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A presence read of an optional field that finishAccessors rewrites into a getter dispatcher keeps the
// read's Absent in the dispatcher's fallback, so a plain literal without the field reads as absent, as
// Node does, natively, sanitized and on JavaScript (#4gpbn7w). On main 2f327160 the fallback was a bare
// Property, and the `{}` row read past the literal's allocation (ASan heap-buffer-overflow, release
// printing Node's answer by luck). Mutant (internal/lower/accessor-dispatch-absent.mutants.md): the
// fallback's Absent dropped, which reds both programs' sanitized runs.
func TestAccessorDispatchAbsentAgreesWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/accessor_dispatch_absent"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") && !strings.HasSuffix(entry.Name(), ".ts") {
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
