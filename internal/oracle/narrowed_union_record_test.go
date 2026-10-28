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
// checked once a call has written a dictionary into it. Main compiled narrowed_union_record_tag.ts, and native read the
// dictionary as a Box, printing 4.6e-310 where Node prints 5. Record is now among the object kinds the narrowed read
// stops at. Mutant: Record dropped from that list in internal/lower/locals.go; with it applied, this test goes red, and
// ADAMIC_NARROWED_RECORD_MUTANT=1 holds the mutant's admitted program to Node to show the miscompile
// (testdata/narrowed-union-record-mutant.json).
func TestNarrowedUnionRecordTag(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/narrowed_union_record_tag.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if disagreement(run{stdout: []byte("5\n")}, truth) != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if os.Getenv("ADAMIC_NARROWED_RECORD_MUTANT") == "1" {
		if err != nil {
			t.Fatalf("the mutant must admit the program: %v", err)
		}
		native, _ := nativelyUncached(t, program)
		if disagreement(truth, native) == "" {
			t.Fatalf("the mutant's program agrees with Node: %+v", native)
		}
		t.Logf("mutant admits it: native exit %d, stdout %q, stderr %q where Node prints %q", native.exitCode, native.stdout, native.stderr, truth.stdout)
		return
	}
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "narrowed_union_record_tag.ts:4:5") || !strings.Contains(err.Error(), "stage 0 can't lower a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables yet") {
		t.Fatalf("want the object-tag NotYet at the narrowing, got %v", err)
	}
}
