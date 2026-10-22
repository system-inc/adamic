package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A logical assignment through the object's own type still compiles and agrees with Node, with a
// checked view of the same object elsewhere (#a3y38ng).
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/view_logical_write_own_type.a", true, false})
}

// &&=, ||= and ??= through a checked view are refused until the logical assignment emitters check and
// box the write as a plain one does (#a3y38ng): each of these segfaulted natively, or would have, where
// Node prints the number written. The computed-key form is refused by the names its key can be, and an
// assignment lowered before the cast that makes its field checked is refused all the same. Each
// program's whole diagnostic is pinned in its .refused file, and its first line names the refusal.
// Mutant: checkViewLogicalWrites returning nil, which lowers every one of them.
func TestViewLogicalWritesAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/view_logical_write_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want a refusal, got %v", err)
			}
			source, err := os.ReadFile(absolute)
			if err != nil {
				t.Fatal(err)
			}
			if header := "// a-check: refused " + refused.What + "\n"; !strings.HasPrefix(string(source), header) {
				t.Fatalf("first line: want %q", header)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 4 {
		t.Fatalf("%d programs in %s, want 4: a program added or removed needs this count changed with it", programs, directory)
	}
}
