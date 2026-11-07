package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const typeofDispatchFixture = "internal/oracle/testdata/typeof_dispatch.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{typeofDispatchFixture, true, false}, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/typeof_string_literal.a", true, false})
}

// Forget only the static constructor distinction in the unified classifier. Normal objects keep
// their answer; constructors, including those held in unions, must fail Node's stdout comparison.
func TestTypeOfConstructorMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, typeofDispatchFixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_union_typeof(", "typeof_constructor_mutant(")
	if mutant == source {
		test.Fatal("unified classifier was not emitted")
	}
	mutant = `#include "adamic.h"
static adamic_string *typeof_constructor_mutant(const adamic_heap *value, bool null) {
	if (value != NULL && value->kind == adamic_kind_object) { return &adamic_typeof_object; }
	return adamic_union_typeof(value, null);
}
` + mutant
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if report := leakChecked(test, mutant, binary); report != "" {
		test.Fatalf("mutant must not leak: %s", report)
	}
	got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant must finish cleanly without leaks: exit %d, stderr %q", got.exitCode, got.stderr)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		test.Fatalf("want Node to catch stdout alone, got %q", difference)
	}
	test.Logf("caught: Node %q; mutant %q", want.stdout, got.stdout)
}

// A wrong classification of the static string must be caught by Node's output alone.
func TestTypeOfStringLiteralMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/typeof_string_literal.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_union_typeof(", "typeof_string_mutant(")
	if mutant == source {
		test.Fatal("string classifier was not emitted")
	}
	mutant = `#include "adamic.h"
static adamic_string *typeof_string_mutant(const adamic_heap *value, bool null) {
 if (value != NULL && value->kind == adamic_kind_string) { return &adamic_typeof_undefined; }
 return adamic_union_typeof(value, null);
}
` + mutant
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if report := leakChecked(test, mutant, binary); report != "" {
		test.Fatalf("mutant must not leak: %s", report)
	}
	got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant must finish cleanly: exit %d, stderr %q", got.exitCode, got.stderr)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		test.Fatalf("want Node to catch stdout alone, got %q", difference)
	}
	test.Logf("caught: Node %q; mutant %q", want.stdout, got.stdout)
}
