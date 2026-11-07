package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"numeric.a", "string.a", "reexport.a", "fallthrough.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/imported_const_cases/" + name, true, false})
	}
}

func TestImportedNonliteralConstCaseIsNotYet(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/imported_const_cases/nonliteral.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "constants initialized\nwide\n" {
		test.Fatalf("Node: %+v", truth)
	}
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	want := path +
		":3:7: stage 0 can't lower a case label using const WIDE with type number (not a number or string literal type) yet"
	if !errors.As(loweredError, &notYet) || loweredError.Error() != want {
		test.Fatalf("got %v; want %s", loweredError, want)
	}
}
