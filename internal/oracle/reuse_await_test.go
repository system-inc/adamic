package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A global read before an await and written after it, while another task writes it during the
// suspension (Memory's question 1 on #qk8rztp S2c-2b). Each program in testdata/reuse_await agrees
// with Node natively, sanitized and on the JavaScript backend, and leaks nothing. Natively
// flow.NormalizeAsync snapshots every operand before a suspension into a frame cell that owns its
// value, so the program is safe whatever native/reuse.go's reachesGlobal says of an await. Mutant
// (h3-proof.sh, and testdata/iterable_views' mutants.md), expected to survive by that reading: Await put on reachesGlobal's not-reaching side, main's behavior. A heap-use-after-free under
// it would mean the snapshot doesn't hold the read, and this program is the witness after all.
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
