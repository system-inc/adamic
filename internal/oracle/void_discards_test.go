package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/void_discards_refused/methods.a", false, false})
}

// A void call through an interface whose class methods of the name disagree, one returning an object
// and one a number, has no one way to let go of what it gives back: a class's method carries no
// discard, as a closure does. Mutant: lowering's voidDiscards left out, which lowers the program and
// leaks the Box Maker.run returns.
func TestVoidDiscardMethodsThatDisagreeAreNotYet(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/void_discards_refused/methods.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ran 2\n" || len(truth.stderr) != 0 {
		test.Fatalf("source Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	if !errors.As(loweredError, &notYet) {
		test.Fatalf("want not yet, got %v", loweredError)
	}
	want := "methods.a: stage 0 can't lower a void call through an interface at the module's top level where the class methods of its name disagree on whether their result is a reference to release; give the methods one result type, or call through a field yet"
	if notYet.Error() != want {
		test.Fatalf("got %q, want %q", notYet.Error(), want)
	}
}
