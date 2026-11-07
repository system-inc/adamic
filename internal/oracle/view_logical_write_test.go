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
func TestViewLogicalWritesAgreeWithNode(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/view_logical_write"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		path := directory + "/" + entry.Name()
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			loomViewsDAgrees(test, path)
		})
	}
	if programs != 8 {
		test.Fatalf(
			"%d programs in %s, want 8: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}

// Where the object doesn't hold what the view says, a logical assignment through it stops where its plain
// twin does, with the same output and exit 70 in both backends: at the checked read, which the ??= that
// writes nothing makes all the same, or at the checked write of a value that misfits the slot. Node,
// without the checks, prints. Mutants: skipping the checked read or the checked write in either backend,
// which runs on past the stop. The lie is the cast's, a downcast .a refuses, so these are .ts.
func TestViewLogicalWritesStopAsTheirPlainTwinsDo(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/view_logical_write_stops/"
	for _, pair := range []struct{ logical, plain, stop string }{
		{"and.ts", "read.ts", "field read failed: viewed.count is not a number"},
		{"nullish.ts", "read.ts", "field read failed: viewed.count is not a number"},
		{"before_cast.ts", "before_cast_read.ts", "field read failed: view.count is not a number"},
		{"misfit.ts", "misfit_equals.ts", "field read failed: <write>.count is not a string | number"},
	} {
		test.Run(pair.logical, func(test *testing.T) {
			test.Parallel()
			logical, loweredError := lowered(test, filepath.Join(repository, directory+pair.logical))
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			plain, loweredError := lowered(test, filepath.Join(repository, directory+pair.plain))
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			logicalSanitized, _ := nativelyUncached(test, logical)
			plainSanitized, _ := nativelyUncached(test, plain)
			for backend, runs := range map[string][2]run{
				"native":     {releasedUncached(test, logical), releasedUncached(test, plain)},
				"sanitized":  {logicalSanitized, plainSanitized},
				"javascript": {onJavaScriptBackend(test, logical), onJavaScriptBackend(test, plain)},
			} {
				if runs[0].exitCode != 70 || !strings.Contains(string(runs[0].stderr), pair.stop) {
					test.Errorf(
						"%s: want exit 70 with %q, got %d, stderr %q",
						backend,
						pair.stop,
						runs[0].exitCode,
						runs[0].stderr,
					)
				}
				if diff := disagreement(runs[1], runs[0]); diff != "" {
					test.Errorf(
						"%s: %s from %s, stderr %q, want %q",
						backend,
						diff,
						pair.plain,
						runs[0].stderr,
						runs[1].stderr,
					)
				}
			}
		})
	}
}

// The four stops are ruled negative witnesses (#995z2fa, Adamic, Oct 25), listed in
// cloud/admission-corpus/negative-witnesses.json, and each one's repair, which stores a value of the
// declared type, agrees with Node natively, sanitized and on JavaScript.
func TestViewLogicalWriteWitnessRepairsAgreeWithNode(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"and", "nullish", "before_cast", "misfit"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			loomViewsDAgrees(test, "internal/oracle/testdata/view_logical_write_stops/"+name+"_repaired.ts")
		})
	}
}

// &&=, ||= and ??= through a checked view by a computed key are refused, by the names its key can be: a
// plain read through the view by a computed key is refused and a plain = isn't lowered, so there is no
// checked form to share. The program's whole diagnostic is pinned in its .refused file, and its first
// line names the refusal. In .a computed.a's cast is refused first, since its tag test, kind ===
// 'Holder', is one every Root passes (#q97a1rr), so that is the refusal pinned here, and the computed
// key's is pinned by its .ts twin, TestViewLogicalWriteTwinIsRefused.
func TestViewLogicalWritesAreRefused(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/view_logical_write_refused"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) {
				test.Fatalf("want a refusal, got %v", loweredError)
			}
			source, readError := os.ReadFile(absolute)
			if readError != nil {
				test.Fatal(readError)
			}
			if header := "// a-check: refused " + refused.What + "\n"; !strings.HasPrefix(string(source), header) {
				test.Fatalf("first line: want %q", header)
			}
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if readError != nil {
				test.Fatal(readError)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 1 {
		test.Fatalf(
			"%d programs in %s, want 1: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}

// computed.a as .ts, where its cast stays a checked view, so its &&= by a computed key is reached and
// refused, its whole diagnostic as computed.refused pinned it before #q97a1rr. Mutant:
// checkViewLogicalWrites returning nil, which lowers it.
func TestViewLogicalWriteTwinIsRefused(test *testing.T) {
	test.Parallel()
	twin := typeScriptTwin(
		test,
		filepath.Join(repository, "internal/oracle/testdata/view_logical_write_refused/computed.a"),
	)
	_, loweredError := lowered(test, twin)
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) {
		test.Fatalf("want a refusal, got %v", loweredError)
	}
	want := "computed.ts:10:1: Adamic 0.1 refuses a logical assignment through a checked view; write the field through the object's own type, or read it, decide, then assign with ="
	if got := strings.ReplaceAll(refused.Error(), filepath.Dir(twin)+string(filepath.Separator), ""); got != want {
		test.Fatalf("diagnostic: got %q, want %q", got, want)
	}
}

// &&=, ||= and ??= on a node array's field are not yet: lowered, native C took the array as the object
// its metadata is, and clang refused it ("incompatible pointer types initializing 'adamic_object *' with
// an expression of type 'adamic_array *'"), though JavaScript agreed with Node. Mutant: dropping the
// node array guard in lower's logicalAssignment, which lowers it.
func TestNodeArrayLogicalWritesAreNotYet(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/logical_write_notyet/node_array.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "-1|9|true|1|1|1\n" {
		test.Fatalf("Node: %+v", truth)
	}
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	want := path + ":18:1: stage 0 can't lower a logical assignment to a node array's field yet"
	if !errors.As(loweredError, &notYet) || loweredError.Error() != want {
		test.Fatalf("got %v; want %s", loweredError, want)
	}
}
