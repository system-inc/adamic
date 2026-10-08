package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/element_type_array_unions.a", true, false})
}

func TestElementTypeArrayUnionFlattenMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_array_unions.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_array_push(", "element_type_discard_push(")
	if mutant == source {
		t.Fatal("flatten mutant changed nothing")
	}
	mutant = `#include "adamic.h"
static void element_type_discard_push(adamic_array *array, adamic_value value) {
    if (array->references) adamic_release(value.reference);
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
	t.Log("discarded flattened contents caught only by Node stdout")
}
