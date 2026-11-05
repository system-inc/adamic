package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A method read only to test for it is an observation, not a detach (#4q1b2er, lower/method_presence.go):
// each program in testdata/method_presence agrees with Node natively, sanitized and on the JavaScript
// backend, and leaks nothing. tsc's own shapes, typeof Object.setPrototypeOf compared with "function", fill
// testing its array and repeat its string before calling them; a program class's method at typeof, !, the
// left of &&, the test of ?:, while, for and do...while, and through parentheses; a read through ?. on a
// receiver that is missing and one that isn't; and a getter that counts its calls and makes a fresh object
// each time, so a test that skipped its receiver, or ran it twice, prints another count. The mutants are in
// testdata/method_presence/mutants.md.
func TestMethodPresenceAgreesWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/method_presence"
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
	if programs != 6 {
		t.Fatalf("%d programs in %s, want 6: a program added or removed needs this count changed with it", programs, directory)
	}
}
