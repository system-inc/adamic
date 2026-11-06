package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An iterator step result's keys enumerate value then done, as Node's { value, done } do (#7nw4kgj).
// Each program in testdata/iterator_result_order agrees with Node natively, released and sanitized,
// and on the JavaScript backend, and leaks nothing. The mutant, each runtime shape's old done-first
// order, is in testdata/iterator_result_order/mutants.md.
func TestIteratorResultKeysAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/iterator_result_order"
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
			generatorAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 1 {
		t.Fatalf("%d programs in %s, want 1: a program added or removed needs this count changed with it", programs, directory)
	}
}
