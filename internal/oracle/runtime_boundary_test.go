package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func checkBoundaryRefusal(t *testing.T, fixture, output, reason string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	result := onNode(t, path)
	if result.exitCode != 0 || string(result.stdout) != output || len(result.stderr) != 0 {
		t.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	var refused *lower.Refused
	wantKind := errors.As(err, &notYet)
	if runtimeBoundaryViewFixture(fixture) {
		wantKind = errors.As(err, &refused)
	}
	if !wantKind || !strings.Contains(err.Error(), reason) {
		t.Fatalf("want current lowering boundary %q, got %v", reason, err)
	}
	t.Logf("Node stdout %q; %v", result.stdout, err)
}

func TestRuntimeBoundaryNormalizeLong(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "normalize_long.a")
}

func TestRuntimeBoundaryPadInfinity(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "pad_infinity.a")
}

func TestRuntimeBoundaryPadTooLong(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "pad_too_long.a")
}

func TestRuntimeBoundaryRepeatHuge(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "repeat_huge.a")
}

func TestRuntimeBoundaryReduceUndefinedInitial(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "reduce_undefined_initial.a")
}

func TestRuntimeBoundaryCoveragePlainView(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "coverage_plain_view.a")
}

func TestRuntimeBoundaryCoverageRegexpCall(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_regexp_call.a", "\n", "spreading a RegExp or its iterator")
}

func TestRuntimeBoundaryCoverageRegexpLiteral(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_regexp_literal.a", "\n", "spreading a RegExp or its iterator")
}

func TestRuntimeBoundaryCoverageRegexpNew(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_regexp_new.a", "\n", "spreading a RegExp or its iterator")
}

func TestRuntimeBoundaryCoverageViewArray(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_view_array.a", "0,1\n", "an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)")
}

func TestRuntimeBoundaryCoverageViewBoolean(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_view_boolean.a", "\n", "an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)")
}

func TestRuntimeBoundaryCoverageViewClosure(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_view_closure.a", "\n", "an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)")
}

func TestRuntimeBoundaryCoverageViewNumber(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_view_number.a", "\n", "an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)")
}

func TestRuntimeBoundaryCoverageViewString(t *testing.T) {
	t.Parallel()
	checkBoundaryRefusal(t, "coverage_view_string.a", "0,1,2,3\n", "an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties)")
}

// Empty views may hold library objects, whose private storage cannot be spread.
func runtimeBoundaryViewFixture(path string) bool {
	switch filepath.Base(path) {
	case "coverage_view_array.a", "coverage_view_boolean.a", "coverage_view_closure.a", "coverage_view_number.a", "coverage_view_string.a":
		return true
	}
	return false
}
