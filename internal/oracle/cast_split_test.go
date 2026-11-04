package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A capture-free RegExp literal's split cast to string[] (#b5w3ycg), in .a and its .ts twin: each lowers and
// agrees with Node natively, through the JavaScript backend and the wasm families that run the oracle's fixtures.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/cast_split_capture_free.a",
		"internal/oracle/testdata/cast_split_capture_free.ts",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// A split cast to string[] stays a cast the runtime can't check whenever its separator might push undefined or
// isn't a literal written at the call (#b5w3ycg): an unnamed capture, a named one, a capture a backreference
// needs, a const holding a capture-free literal (semver.ts:287's shape), a new RegExp(...), and an optional call.
// Each program's whole diagnostic is pinned in its .refused file. Mutant: captureFreeSplit admitting every RegExp
// literal without counting its groups lowers capture, named_capture and backreference.
func TestSplitCastsWithCapturesAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/cast_split_refused"
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
	if programs != 6 {
		t.Fatalf("%d programs in %s, want 6: a program added or removed needs this count changed with it", programs, directory)
	}
}
