package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// A tag-proven predicate whose target has an optional field is refused for the view, which can't check
// that field yet (#1yr7rdh), and the refusal says so (#1w5bve2, Correctness's trace, Oct 29). It used to
// return the flow prover's "return is not proven", which blamed a return that is a trusted tag check:
// 287 of tsc's predicates in nodeTests.ts and its neighbours. Correctness's two programs differ by one
// field: plain is admitted, and optional names its own reason, NotYet in .ts and Refused in .a.
func TestTagPredicateOptionalFieldNamesItsReason(t *testing.T) {
	t.Parallel()
	plain, err := os.ReadFile("testdata/predicates_tag_optional/plain.ts")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNodeAs(t, string(plain), ".ts")
	optional, err := os.ReadFile("testdata/predicates_tag_optional/optional.ts")
	if err != nil {
		t.Fatal(err)
	}
	// In .ts it's a gap the checked optional read closes (#mr87ptk): NotYet, naming the field.
	_, err = lowerTypeScriptAssertionSource(t, string(optional))
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(pending.Error(), "a tag-proven predicate whose target has an optional field (jsDoc on Identifier) that the view can't check") {
		t.Fatalf("want the optional field named as not yet, got %v", err)
	}
	// In .a it needs proof, so it's refused, naming the field.
	_, err = lowerSource(t, string(optional))
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a tag-proven predicate whose target has an optional field (jsDoc on Identifier) that the view can't check" {
		t.Fatalf("want the optional field refused in .a, got %v", err)
	}
	// An asserts claim proven by its tag takes the same branch, in .a too: this row pinned the flow
	// prover's "normal return" until the reason was named (predicates_test.go's "assertion wider shape").
	_, err = lowerSource(t, `type A = { readonly kind: 'a' }; type B = { readonly kind: 'b' }; type Claimed = { readonly kind: 'a'; readonly extra?: number }; function assertA(x: A | B): asserts x is Claimed { if (x.kind !== 'a') { throw new Error('bad'); } }`)
	if !errors.As(err, &refused) || refused.What != "a tag-proven predicate whose target has an optional field (extra on Claimed) that the view can't check" {
		t.Fatalf("want the asserts claim's optional field named, got %v", err)
	}
}
