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

func arrayNarrowingSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "array_narrowing", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func refusesArrayNarrowing(t *testing.T, name, reason string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", "array_narrowing", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), checked)
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
	source := arrayNarrowingSource(t, name)
	position := strings.Index(source, marker)
	if position < 0 {
		t.Fatal("fixture has no narrowing site")
	}
	wantWhere := fmt.Sprintf("%s:%d:%d", path, strings.Count(source[:position], "\n")+1, position-strings.LastIndex(source[:position], "\n"))
	var refusal *Refused
	var notYet *NotYet
	where := ""
	if errors.As(err, &refusal) {
		where = refusal.Where
	} else if errors.As(err, &notYet) {
		where = notYet.Where
	}
	if where != wantWhere {
		t.Fatalf("want refusal at narrowing site %s, got %v", wantWhere, err)
	}
	if err == nil || !strings.Contains(err.Error(), reason) || !strings.Contains(err.Error(), path+":") {
		t.Fatalf("want refusal at exact fixture path with %q, got %v", reason, err)
	}
	if reason == "array element narrowing across a union layout is not yet sound" {
		var refusal *Refused
		if !errors.As(err, &refusal) || refusal.Fix != "read the elements through the union type, or copy them into a new S[]" {
			t.Fatalf("want actionable layout refusal, got %v", err)
		}
	}
}

func TestArrayNarrowingEveryRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "every", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingInferredRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "inferred", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingReadonlyRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "readonly", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingEarlyReturnRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "early_return", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingPropertyRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "property", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingForwardRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "forward", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingCatRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "cat", "array element narrowing across a union layout is not yet sound")
}

func TestArrayNarrowingFindRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "find", "find predicate result representation conversion")
}

func TestArrayNarrowingFindLastRefused(t *testing.T) {
	t.Parallel()
	refusesArrayNarrowing(t, "find_last", "findLast predicate result representation conversion")
}

func TestArrayNarrowingOptionalNumberAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "optional_number")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingOptionalStringAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "optional_string")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingSomeNegatedAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "some_negated")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingFindLastIndexAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "find_last_index")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingPreservedUnionAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "preserved_union")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether the union layout is retained.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingBooleanCallbackAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "boolean_callback")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether the union layout is retained.
	lowersAndAgreesWithNodeNative(t, source)
}
