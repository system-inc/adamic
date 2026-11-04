package native

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A type with no C representation is a compiler bug said out loud, never a string pointer: before
// #tky5538 cType's fallback gave every unnamed type `adamic_string *`, so a new kind (S1's generator)
// or a slot tag reaching it compiled silently as a string. The string itself keeps its own case.
// Mutant: the fallback put back, and the unnamed type returns a string pointer instead of panicking.
func TestCTypeRefusesATypeWithNoCase(t *testing.T) {
	t.Parallel()
	if got := cType(ir.String); got != "adamic_string *" {
		t.Fatalf("cType(String) = %q", got)
	}
	for _, unnamed := range []ir.Type{0, ir.NullRepresentation, ir.UndefinedRepresentation, ir.Uint16Array + 100} {
		func() {
			defer func() {
				failure, isString := recover().(string)
				if !isString || !strings.Contains(failure, "no C type for ir.Type") {
					t.Errorf("cType(%d) gave no compiler-bug panic: %v", unnamed, failure)
				}
			}()
			got := cType(unnamed)
			t.Errorf("cType(%d) = %q, want a panic", unnamed, got)
		}()
	}
}
