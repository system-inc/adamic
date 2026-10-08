package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"generics", "boolean", "brands", "unions"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/element_type_" + name + ".a", true, false})
	}
}

func TestElementTypeGenericIndexMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_generics.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, ", (0x0p+00))", ", (0x1p+00))")
	if mutant == source {
		t.Fatal("index mutant changed nothing")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("wrong specialized array index caught only by Node stdout")
}

func TestElementTypeBooleanUnpackMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_boolean.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_maybe_boolean_unpack(", "element_type_mutant_unpack(")
	if mutant == source {
		t.Fatal("unpack mutant changed nothing")
	}
	mutant = `#include "adamic.h"
static adamic_maybe_boolean element_type_mutant_unpack(uint8_t value) {
 adamic_maybe_boolean result = adamic_maybe_boolean_unpack(value);
 if (result.present) result.boolean = true;
 return result;
}
` + mutant
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("false unpacked as true caught only by Node stdout")
}

func TestElementTypeBooleanSortMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_boolean.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_array_sort_maybe_boolean(", "adamic_array_sort(")
	if mutant == source {
		t.Fatal("sort mutant changed nothing")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("passing undefined to comparator caught only by Node stdout")
}

func TestElementTypeUnionSearchMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_unions.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_union_equal(", "element_type_mutant_equal(")
	if mutant == source {
		t.Fatal("search mutant changed nothing")
	}
	mutant = `#include "adamic.h"
static bool element_type_mutant_equal(const adamic_heap *left, const adamic_heap *right) { return left == right; }
` + mutant
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("number boxes compared by address caught only by Node stdout")
}

func TestElementTypeUnionSortMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_unions.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_array_sort_union(", "adamic_array_sort(")
	if mutant == source {
		t.Fatal("sort mutant changed nothing")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("undefined union passed to comparator caught only by Node stdout")
}

func TestElementTypeUnknownOnNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_unknown.a"))
	if err != nil {
		t.Fatal(err)
	}
	got := onNode(t, path)
	if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != "object undefined\n" {
		t.Fatalf("unknown null/undefined witness: %+v", got)
	}
}

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/element_type_unknown.a", false, false})
}

// TestElementTypeBrandedLengthMutant observes both inhabitable and empty required brands.
func TestElementTypeBrandedLengthMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_brands.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := regexp.MustCompile(`([A-Za-z_][A-Za-z_0-9]*)->length`).ReplaceAllString(source, `($1->length + 1)`)
	if mutant == source {
		t.Fatal("length mutant changed nothing")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("wrong branded array length caught only by Node stdout")
}
