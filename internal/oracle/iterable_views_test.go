package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A for-of over a slot typed Iterable<T> or IterableIterator<T> whose every origin is a Map's or a
// Set's iterator (#qk8rztp S2b-1b): the shape tsc's mapIterator and mapDefinedIterator are called
// with. Each program in testdata/iterable_views agrees with Node natively, sanitized and on the
// JavaScript backend, and leaks nothing: entries destructured and read as a pair, from two callers
// with two maps; a Set's and a Map's values through one parameter; an iterator already stepped, so
// the view sees only the rest; a loop that adds to the map it walks, which Node's iterator visits, and
// one that deletes from it; and an Iterable<string> slot, held as a union and read through a checked
// unbox. Since S2c-3, an Iterable<T> | undefined parameter too: held as a nullable object, read past
// its undefined test through ir.Defined, which the proof follows to the read, and held as a union
// for Iterable<string>. Since S2c-2a, a hand-driven next() through an IterableIterator<T> parameter
// too, checker.ts:21636's loop among it, and since S2c-2b, through one a generator's direct call
// feeds (hand_next_generator*.a, hand_next_mixed.a, hand_next_maybe_string.a): Memory's P1 to P3, a
// generator and a collection's iterator into one parameter, and a string that may be undefined. Keys
// and values are built, never only literals. The mutants are in testdata/iterable_views/mutants.md.
func TestIterableViewsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/iterable_views"
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
	if programs != 13 {
		t.Fatalf("%d programs in %s, want 13: a program added or removed needs this count changed with it", programs, directory)
	}
}
