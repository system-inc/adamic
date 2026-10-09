package lower

import (
	"os"
	"strings"
	"testing"
)

func TestPredicateNestedIfNeedsElse(t *testing.T) {
	source, err := os.ReadFile("../fuzz/testdata/predicates/oct8_predicates_p26_assert_nested_if_without_else.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	if err == nil || !strings.Contains(err.Error(), "return is not proven") {
		t.Fatalf("p26: want unproven-predicate refusal, got %v", err)
	}
}

func TestPredicateOverloadReboundParameter(t *testing.T) {
	source, err := os.ReadFile("../fuzz/testdata/predicates/oct8_predicates_p27_overload_parameter_rebound.a")
	if err != nil {
		t.Fatal(err)
	}
	result, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	if result.PredicateChecks.Checked == 0 {
		t.Fatalf("p27: rebound parameter must retain a runtime predicate check: %+v", result.PredicateChecks)
	}
}
