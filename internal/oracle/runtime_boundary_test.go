package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func checkBoundaryRefusal(test *testing.T, fixture, output, reason string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	result := onNode(test, path)
	if result.exitCode != 0 || string(result.stdout) != output || len(result.stderr) != 0 {
		test.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	var refused *lower.Refused
	wantKind := errors.As(loweredError, &notYet)
	if runtimeBoundaryViewFixture(fixture) {
		wantKind = errors.As(loweredError, &refused)
	}
	if !wantKind || !strings.Contains(loweredError.Error(), reason) {
		test.Fatalf("want current lowering boundary %q, got %v", reason, loweredError)
	}
	test.Logf("Node stdout %q; %v", result.stdout, loweredError)
}

func TestRuntimeBoundaryNormalizeLong(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "normalize_long.a")
}

func TestRuntimeBoundaryPadInfinity(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "pad_infinity.a")
}

func TestRuntimeBoundaryPadTooLong(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "pad_too_long.a")
}

func TestRuntimeBoundaryRepeatHuge(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "repeat_huge.a")
}

func TestRuntimeBoundaryReduceUndefinedInitial(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "reduce_undefined_initial.a")
}

func TestRuntimeBoundaryCoveragePlainView(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "coverage_plain_view.a")
}

func TestRuntimeBoundaryCoverageRegexpCall(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(test, "coverage_regexp_call.a", "\n", "spreading a RegExp or its iterator")
}

func TestRuntimeBoundaryCoverageRegexpLiteral(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(test, "coverage_regexp_literal.a", "\n", "spreading a RegExp or its iterator")
}

func TestRuntimeBoundaryCoverageRegexpNew(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(test, "coverage_regexp_new.a", "\n", "spreading a RegExp or its iterator")
}

func TestRuntimeBoundaryCoverageViewArray(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(
		test,
		"coverage_view_array.a",
		"0,1\n",
		"an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)",
	)
}

func TestRuntimeBoundaryCoverageViewBoolean(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(
		test,
		"coverage_view_boolean.a",
		"\n",
		"an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)",
	)
}

func TestRuntimeBoundaryCoverageViewClosure(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(
		test,
		"coverage_view_closure.a",
		"\n",
		"an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)",
	)
}

func TestRuntimeBoundaryCoverageViewNumber(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(
		test,
		"coverage_view_number.a",
		"\n",
		"an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)",
	)
}

func TestRuntimeBoundaryCoverageViewString(test *testing.T) {
	test.Parallel()
	checkBoundaryRefusal(
		test,
		"coverage_view_string.a",
		"0,1,2,3\n",
		"an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)",
	)
}

// assertRuntimeBoundaryViewRefusal holds an empty-view spread fixture to lowering's refusal, its fix unchanged, in
// every backend's runner: a refusal is a different promise from NotYet, so the table's lowers=false never covers it.
func assertRuntimeBoundaryViewRefusal(test *testing.T, loweredError error) {
	test.Helper()
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) ||
		refused.What != "an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)" ||
		refused.Fix != "copy the fields you mean by name" {
		test.Fatalf("want library view spread refusal, got %v", loweredError)
	}
}

// Empty views may hold library objects, whose private storage cannot be spread.
func runtimeBoundaryViewFixture(path string) bool {
	switch filepath.Base(path) {
	case "coverage_view_array.a", "coverage_view_boolean.a", "coverage_view_closure.a", "coverage_view_number.a", "coverage_view_string.a":
		return true
	}
	return false
}
