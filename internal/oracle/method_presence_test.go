package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
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

// A presence test whose receiver a checked cast let in is not yet: the real object, whose tag passed,
// may lack the method, and testdata/method_presence/unproven/view_lacks_method.ts's does, so Node prints
// "undefined" where a folded "function" would be a silent miscompile. Should it ever lower, it is held to
// Node on every backend, so a fold shows as the disagreement it is. Mutant: methodPresence's guard
// dropped (testdata/method_presence/mutants.md).
func TestMethodPresenceThroughAViewIsNotYet(t *testing.T) {
	t.Parallel()
	relative := "internal/oracle/testdata/method_presence/unproven/view_lacks_method.ts"
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	if want := onNode(t, path); string(want.stdout) != "undefined\nfunction\n" || want.exitCode != 0 {
		t.Fatalf("Node: exit %d, stdout %q, want the view's real object to lack area", want.exitCode, want.stdout)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if errors.As(err, &notYet) && strings.HasSuffix(err.Error(), "a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it yet") {
		return
	}
	if err != nil {
		t.Fatalf("got %v, want a not-yet naming the unproven presence", err)
	}
	generatorAgrees(t, relative)
	t.Error("lowered, want a not-yet naming the unproven presence")
}
