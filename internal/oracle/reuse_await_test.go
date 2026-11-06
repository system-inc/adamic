package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A global moved into an assignment's evaluation across an await (Memory's ruling on #qk8rztp
// S2c-2b, question 1): native/reuse.go's reachesGlobal takes an await, which has no case of its own
// and isn't ir.Inert, as reaching the global, since every other ready task runs before the
// expression resumes. Each program in testdata/reuse_await agrees with Node natively, sanitized and
// on the JavaScript backend, and leaks nothing. Mutant (h3-proof.sh, and testdata/iterable_views'
// mutants.md): Await put on reachesGlobal's not-reaching side, main's behavior, and the sanitized
// run reports heap-use-after-free on the resumed concatenation.
func TestReuseAwaitAgreesWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/reuse_await"
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
	if programs != 1 {
		t.Fatalf("%d programs in %s, want 1: a program added or removed needs this count changed with it", programs, directory)
	}
}
