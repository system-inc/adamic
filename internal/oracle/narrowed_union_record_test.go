package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// typeof says object for a dictionary and a class instance alike, so a local narrowed to Box by instanceof can't be
// checked once a call has written a dictionary into it. Main compiled narrowed_union_record_tag.ts, and release native
// read the dictionary as a Box, printing 5.4e-310 where Node prints 5; sanitized and counted builds stopped only at the
// field load's record guard, as a compiler bug. Record is now among the object kinds the narrowed read stops at. The
// fixture sits in reland_refused, beside narrowed_union_object_tag.a, since it doesn't lower and the flow and fresh
// corpora lower every program in testdata. Mutant: Record dropped from that list in internal/lower/locals.go; with it
// applied, this test goes red, and ADAMIC_NARROWED_RECORD_MUTANT=1 holds the mutant's admitted program to Node on
// sanitized and release native to show the miscompile (testdata/narrowed-union-record-mutant.json).
func TestNarrowedUnionRecordTag(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/reland_refused/narrowed_union_record_tag.ts"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if disagreement(run{stdout: []byte("5\n")}, truth) != "" {
		test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, loweredError := lowered(test, path)
	if os.Getenv("ADAMIC_NARROWED_RECORD_MUTANT") == "1" {
		if loweredError != nil {
			test.Fatalf("the mutant must admit the program: %v", loweredError)
		}
		native, _ := nativelyUncached(test, program)
		for name, got := range map[string]run{"native": native, "release": releasedUncached(test, program)} {
			if disagreement(truth, got) == "" {
				test.Fatalf("the mutant's %s program agrees with Node: %+v", name, got)
			}
			test.Logf(
				"mutant admits it: %s exit %d, stdout %q, stderr %q where Node prints %q",
				name,
				got.exitCode,
				got.stdout,
				got.stderr,
				truth.stdout,
			)
		}
		return
	}
	var notYet *lower.NotYet
	if !errors.As(loweredError, &notYet) ||
		!strings.Contains(loweredError.Error(), "narrowed_union_record_tag.ts:4:5") ||
		!strings.Contains(
			loweredError.Error(),
			"stage 0 can't lower a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables yet",
		) {
		test.Fatalf("want the object-tag NotYet at the narrowing, got %v", loweredError)
	}
}
