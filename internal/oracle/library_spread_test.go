package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// This is the task's exact program, imported alone from runtime's draft.
func TestRegExpViewSpreadIsRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_spread_refused/coverage_regexp_view.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want library spread refusal, got %v", err)
	}
	want := path + ":3:19: Adamic 0.1 refuses an object spread through a view that may hold library class RegExp (its internal fields are not own enumerable properties); copy the fields you mean by name"
	if refused.Error() != want {
		t.Fatalf("got %q, want %q", refused.Error(), want)
	}
	t.Log(strings.TrimPrefix(refused.Error(), filepath.Dir(path)+"/"))
}
