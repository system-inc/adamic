package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func arrayNarrowingSource(test *testing.T, name string) string {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join("testdata", "array_narrowing", name+".a"))
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

func refusesArrayNarrowing(test *testing.T, name, reason string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join("testdata", "array_narrowing", name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := Lower(context.Background(), checked)
	method := "every"
	if name == "find" {
		method = "find"
	} else if name == "find_last" {
		method = "findLast"
	}
	marker := "items." + method + "("
	if name == "property" {
		marker = "obj." + marker
	}
	source := arrayNarrowingSource(test, name)
	position := strings.Index(source, marker)
	if position < 0 {
		test.Fatal("fixture has no narrowing site")
	}
	wantWhere := fmt.Sprintf(
		"%s:%d:%d",
		path,
		strings.Count(source[:position], "\n")+1,
		position-strings.LastIndex(source[:position], "\n"),
	)
	var refusal *Refused
	var notYet *NotYet
	where := ""
	if errors.As(lowerError, &refusal) {
		where = refusal.Where
	} else if errors.As(lowerError, &notYet) {
		where = notYet.Where
	}
	if where != wantWhere {
		test.Fatalf("want refusal at narrowing site %s, got %v", wantWhere, lowerError)
	}
	if lowerError == nil ||
		!strings.Contains(lowerError.Error(), reason) ||
		!strings.Contains(lowerError.Error(), path+":") {
		test.Fatalf("want refusal at exact fixture path with %q, got %v", reason, lowerError)
	}
	if reason == "array element narrowing across a union layout is not yet sound" {
		var refusal *Refused
		if !errors.As(lowerError, &refusal) ||
			refusal.Fix != "read the elements through the union type, or copy them into a new S[]" {
			test.Fatalf("want actionable layout refusal, got %v", lowerError)
		}
	}
}

func TestArrayNarrowingEveryRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "every", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingInferredRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "inferred", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingReadonlyRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "readonly", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingEarlyReturnRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "early_return", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingPropertyRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "property", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingForwardRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "forward", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingCatRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "cat", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingFindRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "find", "find predicate result representation conversion")
}

func TestArrayNarrowingFindLastRefused(test *testing.T) {
	test.Parallel()
	refusesArrayNarrowing(test, "find_last", "findLast predicate result representation conversion")
}

func TestArrayNarrowingOptionalNumberAgrees(test *testing.T) {
	test.Parallel()
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(test, "ArrayNarrowingOptionalNumber")
}

func TestArrayNarrowingOptionalStringAgrees(test *testing.T) {
	test.Parallel()
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(test, "ArrayNarrowingOptionalString")
}

func TestArrayNarrowingSomeNegatedAgrees(test *testing.T) {
	test.Parallel()
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(test, "ArrayNarrowingSomeNegated")
}

func TestArrayNarrowingFindLastIndexAgrees(test *testing.T) {
	test.Parallel()
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(test, "ArrayNarrowingFindLastIndex")
}

func TestArrayNarrowingPreservedUnionAgrees(test *testing.T) {
	test.Parallel()
	// Only native exposes whether the union layout is retained.
	lowersAndAgreesWithNodeNative(test, "ArrayNarrowingPreservedUnion")
}

func TestArrayNarrowingBooleanCallbackAgrees(test *testing.T) {
	test.Parallel()
	// Only native exposes whether the union layout is retained.
	lowersAndAgreesWithNodeNative(test, "ArrayNarrowingBooleanCallback")
}
