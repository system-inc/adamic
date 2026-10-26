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
	}{"internal/oracle/testdata/view_logical_write_own_type.ts", true, false})
}

// &&=, ||= and ??= through a checked view read and write the field as a plain read and = through it do
// (#995z2fa), so each of these agrees with Node natively, sanitized and on the JavaScript backend: each
// operator, a number, a string and a boolean written into a union slot, an assignment lowered before the
// cast that makes its field checked, and a receiver evaluated once. Mutants: skipping the checked read
// (viewFieldAt's or checkedViewFieldAt's call) or the checked write (writeProperty's or
// checkedViewWrite's) in either backend. Each casts through an interface that extends its source with no
// tag of its own, the downcast .a refuses (#995z2fa), so each is .ts: over a tagged union, the view's
// checks would be proven away and the checked path never run.
func TestViewLogicalWritesAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/view_logical_write"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		path := directory + "/" + entry.Name()
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, path)
		})
	}
	if programs != 8 {
		t.Fatalf("%d programs in %s, want 8: a program added or removed needs this count changed with it", programs, directory)
	}
}

// Where the object doesn't hold what the view says, a logical assignment through it stops where its plain
// twin does, with the same output and exit 70 in both backends: at the checked read, which the ??= that
// writes nothing makes all the same, or at the checked write of a value that misfits the slot. Node,
// without the checks, prints. Mutants: skipping the checked read or the checked write in either backend,
// which runs on past the stop. The lie is the cast's, a downcast .a refuses, so these are .ts.
func TestViewLogicalWritesStopAsTheirPlainTwinsDo(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/view_logical_write_stops/"
	for _, pair := range []struct{ logical, plain, stop string }{
		{"and.ts", "read.ts", "field read failed: viewed.count is not a number"},
		{"nullish.ts", "read.ts", "field read failed: viewed.count is not a number"},
		{"before_cast.ts", "before_cast_read.ts", "field read failed: view.count is not a number"},
		{"misfit.ts", "misfit_equals.ts", "field read failed: <write>.count is not a string | number"},
	} {
		t.Run(pair.logical, func(t *testing.T) {
			t.Parallel()
			logical, err := lowered(t, filepath.Join(repository, directory+pair.logical))
			if err != nil {
				t.Fatal(err)
			}
			plain, err := lowered(t, filepath.Join(repository, directory+pair.plain))
			if err != nil {
				t.Fatal(err)
			}
			logicalSanitized, _ := nativelyUncached(t, logical)
			plainSanitized, _ := nativelyUncached(t, plain)
			for backend, runs := range map[string][2]run{
				"native":     {releasedUncached(t, logical), releasedUncached(t, plain)},
				"sanitized":  {logicalSanitized, plainSanitized},
				"javascript": {onJavaScriptBackend(t, logical), onJavaScriptBackend(t, plain)},
			} {
				if runs[0].exitCode != 70 || !strings.Contains(string(runs[0].stderr), pair.stop) {
					t.Errorf("%s: want exit 70 with %q, got %d, stderr %q", backend, pair.stop, runs[0].exitCode, runs[0].stderr)
				}
				if diff := disagreement(runs[1], runs[0]); diff != "" {
					t.Errorf("%s: %s from %s, stderr %q, want %q", backend, diff, pair.plain, runs[0].stderr, runs[1].stderr)
				}
			}
		})
	}
}

// The four stops are ruled negative witnesses (#995z2fa, Adamic, Oct 25), listed in
// cloud/admission-corpus/negative-witnesses.json, and each one's repair, which stores a value of the
// declared type, agrees with Node natively, sanitized and on JavaScript.
func TestViewLogicalWriteWitnessRepairsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"and", "nullish", "before_cast", "misfit"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, "internal/oracle/testdata/view_logical_write_stops/"+name+"_repaired.ts")
		})
	}
}

// &&=, ||= and ??= through a checked view by a computed key are refused, by the names its key can be: a
// plain read through the view by a computed key is refused and a plain = isn't lowered, so there is no
// checked form to share. The program's whole diagnostic is pinned in its .refused file, and its first
// line names the refusal. Mutant: checkViewLogicalWrites returning nil, which lowers it.
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
	if programs != 1 {
		t.Fatalf("%d programs in %s, want 1: a program added or removed needs this count changed with it", programs, directory)
	}
}
