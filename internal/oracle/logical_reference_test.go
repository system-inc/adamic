package oracle

import (
	"testing"
)

// A logical operand held one way and a result held another, and a return of a never call, each held
// to Node natively, sanitized and on JavaScript, and leaking nothing. Before these, each panicked in
// emission or emitted C clang refused, exit 4 or a build failure on a valid program (tsc's
// lone-compiled units, #a1yt1e6):
//   - missing_left.ts and present_left.ts, a reference left of && (#v5k6t3k). Mutant: the missing
//     arm takes the right's value, and missing_left.ts reads a field of a missing item.
//   - union_maybe.ts, a union left kept as boolean | undefined (#cme8paq). Mutant: a missing union
//     reads as present false, and the printed undefined becomes false.
//   - never_return/kinds.ts, return fail() from a number, string, string enum, array or object
//     function, on the throwing path too (#xv8kybc). Mutant: the never return lowered as any other,
//     and clang refuses the returned object pointer.
func TestLogicalReferenceOperandsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"internal/oracle/testdata/logical_reference/missing_left.ts",
		"internal/oracle/testdata/logical_reference/present_left.ts",
		"internal/oracle/testdata/logical_reference/union_maybe.ts",
		"internal/oracle/testdata/never_return/kinds.ts",
	} {
		t.Run(path, func(t *testing.T) {
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
