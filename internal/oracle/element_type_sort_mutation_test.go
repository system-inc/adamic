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
	}{"internal/oracle/testdata/element_type_sort_mutation.a", true, false})
}

func TestElementTypeSortMutationTailMutant(t *testing.T) {
	elementTypeSortMutationMutant(t, `#include "adamic.h"
static void element_type_sort_drop_tail(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
 size_t original_length = array->length;
 adamic_array_sort_maybe_boolean(array, compare, context);
 array->length = original_length;
}
`)
}

func TestElementTypeSortMutationShortenMutant(t *testing.T) {
	elementTypeSortMutationMutant(t, `#include "adamic.h"
typedef struct {
 adamic_array *array;
 int (*compare)(adamic_value, adamic_value, void *);
 void *context;
 size_t shortest;
} element_type_mutation_context;
static int element_type_mutation_compare(adamic_value left, adamic_value right, void *context) {
 element_type_mutation_context *state = context;
 int result = state->compare(left, right, state->context);
 if (state->array->length < state->shortest) state->shortest = state->array->length;
 return result;
}
static void element_type_sort_drop_tail(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
 size_t original_length = array->length;
 element_type_mutation_context state = {array, compare, context, original_length};
 adamic_array_sort_maybe_boolean(array, element_type_mutation_compare, &state);
 if (state.shortest < original_length) array->length = state.shortest;
}
`)
}

func elementTypeSortMutationMutant(t *testing.T, shim string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_sort_mutation.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_array_sort_maybe_boolean(", "element_type_sort_drop_tail(")
	if mutant == source {
		t.Fatal("tail mutant changed nothing")
	}
	mutant = shim + mutant
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
	t.Log("discarding required sort writeback caught by Node stdout")
}
