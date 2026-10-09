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

func TestImportedNonliteralConstCaseIsNotYet(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/imported_const_cases/nonliteral.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "constants initialized\nwide\n" {
		t.Fatalf("Node: %+v", truth)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	want := path + ":3:7: stage 0 can't lower a case label using const WIDE with type number (not a number or string literal type) yet"
	if !errors.As(err, &notYet) || err.Error() != want {
		t.Fatalf("got %v; want %s", err, want)
	}
}
