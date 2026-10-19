package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Writes through a union that fit every member still compile and agree with Node (#yawc0z5).
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/union_write_fits.a", true, false})
}

// A write through a union whose value doesn't fit every member's slot is refused, in .a and .ts,
// for every write form tsc accepts: Language's H1 probes, each of which prints NaN or throws on
// Node where native read the wrong bits (ruled Oct 18). Each program's whole diagnostic is pinned in
// its .refused file. Mutant: unionSlotWriteRefusal returning nil, which lowers every one of them.
func TestUnionSlotWritesAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/union_write_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") && !strings.HasSuffix(name, ".ts") {
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
			expected, err := os.ReadFile(strings.TrimSuffix(strings.TrimSuffix(absolute, ".a"), ".ts") + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 13 {
		t.Fatalf("%d programs in %s, want 13: a program added or removed needs this count changed with it", programs, directory)
	}
}
