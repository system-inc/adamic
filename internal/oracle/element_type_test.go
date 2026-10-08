package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"generics", "boolean", "brands"} {
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
