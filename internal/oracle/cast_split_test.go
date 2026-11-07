package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	regex "github.com/system-inc/adamic/internal/regexp"
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
// needs, a capture inside each lookaround (a lookahead, a quantified one, a lookbehind and both negative ones),
// a const holding a capture-free literal (semver.ts:287's shape), a new RegExp(...), and an optional call. Each
// program's whole diagnostic is pinned in its .refused file. Mutant: captureFreeSplit admitting every RegExp
// literal without counting its groups lowers every capture program here.
func TestSplitCastsWithCapturesAreRefused(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/cast_split_refused"
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
	if programs != 11 {
		test.Fatalf(
			"%d programs in %s, want 11: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}

// Some separators never reach castProof, since the checker stops at them first, each whole error pinned in its .error
// file: \1 with no group, though annex B reads it as an octal escape and Node splits on it, is TS1534 with no flag,
// with u and with v alike, and a quantified lookbehind, a syntax error in Node too, is TS1507. Should the checker
// ever pass /a\1/, TestSplitCastCaptureCountReadsAnnexB holds the count castProof would read.
func TestSplitCastSeparatorsTheCheckerStops(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/cast_split_type_error"
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
			_, loadError := load.Load([]string{absolute})
			if loadError == nil {
				test.Fatal("want the checker to stop, got a program")
			}
			expected, err2 := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".error")
			if err2 != nil {
				test.Fatal(err2)
			}
			got := strings.TrimSpace(strings.ReplaceAll(loadError.Error(), strings.TrimSuffix(absolute, path), ""))
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("error: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 4 {
		test.Fatalf(
			"%d programs in %s, want 4: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}

// The count castProof reads, from Adamic's own parser: with no flag annex B reads \1 with no group as an octal escape,
// so /a\1/ has no group (Indices holds only the whole match); with u or v it is a syntax error, which leaves the
// cast refused; and a capture inside any lookaround is counted. Mutant: CaptureFacts never reading a lookaround's body
// fails its five lookaround rows here, and the five lookaround programs of TestSplitCastsWithCapturesAreRefused.
func TestSplitCastCaptureCountReadsAnnexB(test *testing.T) {
	test.Parallel()
	for _, each := range []struct {
		pattern, flags string
		indices        int
		invalid        bool
	}{
		{`a\1`, "", 1, false},
		{`a\1`, "u", 0, true},
		{`a\1`, "v", 0, true},
		{`-(?=(b))`, "", 2, false},
		{`x(?=(y))?`, "", 2, false},
		{`(?<=(a))b`, "", 2, false},
		{`a(?!(b))`, "", 2, false},
		{`(?<!(b))a`, "", 2, false},
	} {
		facts, captureFactsError := regex.CaptureFacts(each.pattern, each.flags)
		if (captureFactsError != nil) != each.invalid {
			test.Errorf("/%s/%s: error %v, want invalid %v", each.pattern, each.flags, captureFactsError, each.invalid)
			continue
		}
		if !each.invalid && len(facts.Indices) != each.indices {
			test.Errorf("/%s/%s: %d indices, want %d", each.pattern, each.flags, len(facts.Indices), each.indices)
		}
	}
}
