package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A value typed T, T's constraint holding undefined or null, narrowed past === undefined or === null, at an
// instance whose T holds nothing else (#rnbvya6): the checker narrows T by its constraint and calls the
// read a string, but here it's never, and lowering's defined (narrowed.go) holds what follows its check as
// the narrowed type. Each program in testdata/narrowed_never agrees with Node natively, sanitized and on the
// JavaScript backend, and leaks nothing (generatorAgrees). Before, return.ts's bare return handed clang the
// missing object the declaration is held as, in a function returning a string, and clang refused it. The
// mutant is in testdata/narrowed_never/mutants.md.
func TestNarrowedNeverReadsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/narrowed_never"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 4 {
		t.Fatalf("%d programs in %s, want 4: a program added or removed needs this count changed with it", programs, directory)
	}
}
