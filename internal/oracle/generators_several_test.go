package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// One iterable view fed by generators of several functions (#qk8rztp S3c-2): its dispatch's one step
// resumes whichever frame it holds through the frame's own code, and tests for a pending throw when
// any function feeding it may leave one. Each program in testdata/generators_several agrees with
// Node natively, sanitized and on the JavaScript backend, and leaks nothing: tsc's
// elaborateElementwise shape, four nested generator functions into one nested callee's parameter,
// one of them never yielding; a view where only one of the functions throws, caught around the
// for-of, with the thrower listed first in one view and second in another; and one view fed by two
// functions in turn, each holding a built string across its yields, broken out of early so the
// view's block closes a frame still holding it. The mutants are in
// testdata/generators_several/mutants.md.
func TestGeneratorsSeveralAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/generators_several"
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
	if programs != 3 {
		t.Fatalf("%d programs in %s, want 3: a program added or removed needs this count changed with it", programs, directory)
	}
}
