package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// A logical whose right operand is held as one maybe pair and whose result is held as another
// (#thwgcax). isReady(holder) && holder.value over a number | undefined field the checker narrowed
// to undefined is false | undefined, a boolean | undefined result, and fit leaves the field's number
// pair as it is, so lowering emitted C assigning the one to the other, which clang refused on a
// valid program. narrowed_number.ts now stops at the && with both representations named, and Node
// is held to its output so the program is known to run. Mutant (scratch-thw/proof.sh): the check
// removed, and narrowed_number.ts lowers and goes red at clang again.
func TestLogicalMaybeMismatchStops(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/logical_maybe_mismatch/narrowed_number.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "absent\n" {
		t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, "absent\n", truth.stderr)
	}
	program, err := lowered(t, path)
	want := "narrowed_number.ts:9:9: stage 0 can't lower a logical operand held as number | undefined where its result is held as boolean | undefined yet"
	if err == nil {
		// Built anyway, so a lost stop says what native makes of the program, clang's refusal at main.
		got := releasedUncached(t, program)
		t.Fatalf("lowered, want a stop ending %q; native: exit %d, stdout %q, stderr %q", want, got.exitCode, got.stdout, got.stderr)
	}
	if !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want a stop ending %q", err, want)
	}
}

// The two neighbors the stop must not reach, each lowered and held to Node natively, sanitized and
// on JavaScript, and leaking nothing: unnarrowed_number.ts, whose && is a union the right is boxed
// into (absent, then present), and narrowed_boolean.ts, whose boolean | undefined field is already
// held as its false | undefined result (absent).
func TestLogicalMaybeMismatchAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"internal/oracle/testdata/logical_maybe_mismatch/unnarrowed_number.ts",
		"internal/oracle/testdata/logical_maybe_mismatch/narrowed_boolean.ts",
	} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, path)
			program, err := lowered(t, repository+"/"+path)
			if err != nil {
				t.Fatal(err)
			}
			_, binary := natively(t, program)
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("leaked: %s", report)
			}
		})
	}
}
