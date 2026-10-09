package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/presence_refused/optional_field_write.a", false, false})
}

// in that narrows a union of object types is refused, not left for later: Node runs the source, and
// the whole diagnostic is pinned, so a lowering that trusts the narrowed member can't slip back in.
func TestInUnionNarrowingIsRefused(t *testing.T) {
	t.Parallel()
	path := "internal/oracle/testdata/presence_refused/in_union_narrowing.a"
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	if truth := onNode(t, absolute); truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
	}
	_, err = lowered(t, absolute)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want a refusal, got %v", err)
	}
	expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
	if got != strings.TrimSpace(string(expected)) {
		t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
	}
}
