package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A spread runs its source's getters natively (runtime/object.c's adamic_object_copy_checked), so
// native/reuse.go's reachesGlobal and worker/purity.go judge an object literal by the getters
// lowering says its spread may run (ir.ObjectLiteral's SpreadGetters), a set read off the accessor
// registry, every registered class's getters whose instances are assignable to the type the spread
// sees, never a test of the source's type (Memory's and Compiler's ruling on #qk8rztp S2c-2b). Each
// program in testdata/reuse_spread agrees with Node natively, sanitized and on the JavaScript
// backend, and leaks nothing: a getter that writes the global a call was handed, an object literal's
// own (which a spread runs), a class's and one inherited from a base (which it doesn't, the set
// over-approximating), and the control without a getter. Each reads the copy by its named fields.
// Mutants (h3-proof.sh, and testdata/iterable_views' mutants.md): the spread's case dropped from
// reachesGlobal, and spreadGetters' call made nil; each lets reuse lend or move the global across
// the getter's write, which only the object literal's row runs.
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
	if programs != 4 {
		t.Fatalf("%d programs in %s, want 4: a program added or removed needs this count changed with it", programs, directory)
	}
}
