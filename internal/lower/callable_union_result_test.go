package lower

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestCallableUnionResultAgreement(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/callable_union_result_both.a")
	if err != nil {
		t.Fatal(err)
	}
	// Only native exposes the numeric result-slot interpretation; JavaScript's
	// ordinary invocation already has the producer's number or string value.
	lowersAndAgreesWithNodeNative(t, string(source))
}

func TestCallableUnionResultReferenceMutant(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/callable_union_result_zero.a")
	if err != nil {
		t.Fatal(err)
	}
	message := recordNativeAgreementFailure(t, string(source), func(code string) string {
		slots := regexp.MustCompile(`adamic_box_number\((adamic_temporary_[0-9]+)\.number\)`)
		if !slots.MatchString(code) {
			t.Fatal("mutant found no numeric producer boxing")
		}
		// Numeric zero becomes NULL when read as .reference. This is an executable
		// stdout mismatch (undefined versus 0), with no crash or clang rejection.
		return slots.ReplaceAllString(code, "(adamic_heap *)$1.reference")
	})
	if !strings.Contains(message, "Native backend stdout") {
		t.Fatalf("reference-slot mutant escaped Node comparison: %q", message)
	}
	t.Log(message)
}
