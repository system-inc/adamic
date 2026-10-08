package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Register this unit's fixtures without changing the shared oracle list.
func init() {
	for _, name := range []string{
		"library_math_metadata.a",
		"library_math_function_own.a",
		"library_number_immediate.a",
		"library_number_math_edges.a",
		"library_number_own_names.a",
		"library_number_convert_sources.a",
		"library_number_optional_formats.a",
		"library_number_global_predicates.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Each mutant changes valid generated C and finishes cleanly. Only the independent Node
// stdout comparison kills it, so a compiler or sanitizer failure cannot claim the check.
func TestLibraryNumberMathOracleCatchesMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct{ name, fixture, before, after string }{
		{"function_own", "library_math_function_own.a", "adamic_string_equal(", "!adamic_string_equal("},
		{"boxed_slot", "library_number_immediate.a", "adamic_number_to_radix(", "adamic_number_to_radix(1 + "},
		{"undefined_digits", "library_number_optional_formats.a", ", 0.0, false)", ", 0.0, true)"},
		{"global_predicate", "library_number_global_predicates.a", "isfinite(", "isnan("},
		{"metadata_arity", "library_math_metadata.a", "0x1p+01", "0x1.8p+01"},
		{"metadata_type", "library_math_metadata.a", "object", "function"},
		{"metadata_name", "library_math_metadata.a", "max", "min"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", mutant.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			stringMutant := mutant.name == "metadata_type" || mutant.name == "metadata_name"
			if stringMutant {
				found := false
				for index, text := range program.Strings {
					if text == mutant.before {
						program.Strings[index] = mutant.after
						found = true
					}
				}
				if !found {
					t.Fatalf("string mutant target %q absent", mutant.before)
				}
			}
			source := native.C(program)
			if !stringMutant && !strings.Contains(source, mutant.before) {
				t.Fatalf("mutant target %q absent", mutant.before)
			}
			if mutant.name == "undefined_digits" {
				source = regexp.MustCompile(`(adamic_number_to_exponential\([^
]*, 0\.0, )false\)`).ReplaceAllString(source, `${1}true)`)
			} else if !stringMutant {
				source = strings.ReplaceAll(source, mutant.before, mutant.after)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := execute(t, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d, stderr %q", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("got %q, want stdout differs", difference)
			}
			t.Log("Node caught the mutant: stdout differs; sanitizer clean")
		})
	}
}
