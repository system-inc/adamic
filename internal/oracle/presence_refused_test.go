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
func TestInUnionNarrowingIsRefused(test *testing.T) {
	test.Parallel()
	path := "internal/oracle/testdata/presence_refused/in_union_narrowing.a"
	absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if truth := onNode(test, absolute); truth.exitCode != 0 || len(truth.stderr) != 0 {
		test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
	}
	_, loweredError := lowered(test, absolute)
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) {
		test.Fatalf("want a refusal, got %v", loweredError)
	}
	expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
	if readError != nil {
		test.Fatal(readError)
	}
	got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
	if got != strings.TrimSpace(string(expected)) {
		test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
	}
}
