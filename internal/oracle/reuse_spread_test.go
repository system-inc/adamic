package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A spread runs its source's getters natively (runtime/object.c's adamic_object_copy_checked), so
// native/reuse.go's reachesGlobal and worker/purity.go judge an object literal by the getters
// lowering says its spread may run (ir.ObjectLiteral's SpreadGetters), a set read by flow off the
// accessor registry, never off the source's type (Memory's and Compiler's ruling on #qk8rztp S2c-2b).
// Each program in testdata/reuse_spread agrees with Node natively, sanitized and on the JavaScript
// backend, and leaks nothing: an object literal's getter that writes the global a call was handed,
// and the control without a getter. Mutants (h3-proof.sh, and testdata/iterable_views' mutants.md):
// the spread's case dropped from reachesGlobal, and spreadGetters' call made nil; each lets reuse lend
// or move the global across the getter's write.
func TestReuseSpreadAgreesWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/reuse_spread"
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
			// generatorAgrees holds any program to Node on every backend and to leaking nothing.
			generatorAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 2 {
		t.Fatalf("%d programs in %s, want 2: a program added or removed needs this count changed with it", programs, directory)
	}
}
