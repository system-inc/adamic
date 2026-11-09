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

// This witness must stop at its registered construct, with the same reason.
func TestAbsentOptionalOwnFieldWriteRefusalPin(t *testing.T) {
	t.Parallel()
	path := "internal/oracle/testdata/presence_refused/optional_field_write.a"
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, absolute)
	const reason = "writing a possibly absent optional own field"
	const location = "optional_field_write.a:9:"
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), reason) || !strings.Contains(notYet.Where, location) {
		t.Fatalf("want located refusal %q at %s: %v", reason, location, err)
	}
	expected, readErr := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
	if readErr != nil {
		t.Fatal(readErr)
	}
	got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), "")
	if got != strings.TrimSpace(string(expected)) {
		t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
	}
	t.Logf("%v", err)
}
