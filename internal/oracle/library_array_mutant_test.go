package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// These mutants compile, finish without sanitizer errors or leaks, and fail only Node's stdout.
func TestArrayFamilyMutants(test *testing.T) {
	test.Parallel()
	cases := []struct {
		name, fixture string
		mutate        func(string) string
	}{
		{"iterator index", "library_array_iterators.a", func(source string) string {
			pattern := regexp.MustCompile(`(adamic_local_\d+_array_iterator_index \+ \()0x1p\+00`)
			return pattern.ReplaceAllString(source, `${1}0x1p+01`)
		}},
		{"join separator", "library_array_join.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_join_nested(", "array_mutant_join_nested(")
		}},
		{"method metadata", "library_array_metadata.a", func(source string) string {
			return strings.ReplaceAll(
				source,
				"ADAMIC_STRING_UNITS(\"indexOf\", 7)",
				"ADAMIC_STRING_UNITS(\"wrong\", 5)",
			)
		}},
		{"with replacement", "library_array_with.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_set(", "array_mutant_set(")
		}},
		{"flatMap order", "library_array_flat_map.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_push(", "array_mutant_push(")
		}},
		{"spliced copy", "library_array_spliced.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_remove(", "array_mutant_remove(")
		}},
		{"flat order", "library_array_flat.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_push(", "array_mutant_push(")
		}},
		{"copyWithin write", "library_array_copy_within.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_set(", "array_mutant_set(")
		}},
		{"search direction", "library_array_search.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_search_from(", "array_mutant_search(")
		}},
		{"copy reversal", "library_array_copy.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_array_reverse(", "array_mutant_reverse(")
		}},
		{"default sort ordering", "library_array_copy.a", func(source string) string {
			return strings.ReplaceAll(source, "adamic_string_compare(", "adamic_string_equal(")
		}},
		{"reverse callback order", "library_array_find_last.a", func(source string) string {
			pattern := regexp.MustCompile(`for \(size_t (\w+) = (\w+); \w+-- > 0;\)`)
			return pattern.ReplaceAllString(source, `for (size_t $1 = 0; $1 < $2; $1++)`)
		}},
	}
	for _, one := range cases {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", one.fixture))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			original := native.C(program)
			mutant := one.mutate(original)
			if mutant == original {
				test.Fatal("mutant changed nothing")
			}
			switch one.name {
			case "join separator":
				mutant = `#include "adamic.h"
static adamic_string *array_mutant_join_nested(const adamic_array *array, const adamic_string *separator, enum adamic_join kind, size_t depth) {
 static const adamic_string comma = ADAMIC_STRING(",");
 (void)separator;
 return adamic_array_join_nested(array, &comma, kind, depth);
}
` + mutant
			case "spliced copy":
				mutant = `#include "adamic.h"
static void array_mutant_remove(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items) {
 adamic_array_remove(array, start, 0, has_count, item_count, items);
}
` + mutant
			case "flat order", "flatMap order":
				mutant = `#include "adamic.h"
static void array_mutant_push(adamic_array *array, adamic_value value) { adamic_array_push(array, value); adamic_array_reverse(array); }
` + mutant
			case "copyWithin write", "with replacement":
				mutant = `#include "adamic.h"
static void array_mutant_set(adamic_array *array, double index, adamic_value value) {
 if (array->references) adamic_release(value.reference);
 (void)index;
}
` + mutant
			case "search direction":
				mutant = `#include "adamic.h"
static double array_mutant_search(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool includes, double from, bool has_from, bool last) {
	(void)last;
	return adamic_array_search_from(array, value, equality, includes, from, has_from, false);
}
` + mutant
			case "copy reversal":
				mutant = `#include "adamic.h"
static adamic_array *array_mutant_reverse(adamic_array *array) { return array; }
` + mutant
			}
			binary := filepath.Join(test.TempDir(), "mutant")
			if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			if report := leakChecked(test, mutant, binary); report != "" {
				test.Fatalf("mutant must not leak: %s", report)
			}
			got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
			truth := onNode(test, path)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("mutant failed outside comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				test.Fatalf("mutant caught by %q, want stdout differs", difference)
			}
			test.Log("caught only by stdout comparison with Node")
		})
	}
}

func TestArrayWithBoundsMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_with.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "array_with" {
			continue
		}
		body := []ir.Statement{}
		for _, statement := range function.Body {
			switch node := statement.(type) {
			case ir.If:
				node.Condition = ir.BooleanConstant{Value: false}
				statement = node
				changed = true
			case ir.SetIndex:
				continue // Avoid the generic index panic: only the RangeError comparison may catch this mutant.
			}
			body = append(body, statement)
		}
		function.Body = body
	}
	if !changed {
		test.Fatal("bounds mutant changed nothing")
	}
	got, binary := natively(test, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
	if difference := disagreement(onNode(test, path), got); difference != "stdout differs" {
		test.Fatalf("bounds mutant caught by %q", difference)
	}
	test.Log("removed range check caught only by Node stdout")
}
