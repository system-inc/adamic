package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// This is the task's exact program, imported alone from runtime's draft.
func TestRegExpViewSpreadIsRefused(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/library_spread_refused/coverage_regexp_view.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "\n" || len(truth.stderr) != 0 {
		test.Fatalf("source Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, loweredError := lowered(test, path)
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) {
		test.Fatalf("want library spread refusal, got %v", loweredError)
	}
	want := path +
		":3:19: Adamic 0.1 refuses an object spread through a view that may hold library class RegExp (its internal fields are not own enumerable properties); copy the fields you mean by name"
	if refused.Error() != want {
		test.Fatalf("got %q, want %q", refused.Error(), want)
	}
	test.Log(strings.TrimPrefix(refused.Error(), filepath.Dir(path)+"/"))
}
