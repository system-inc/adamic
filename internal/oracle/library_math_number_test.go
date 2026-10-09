package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// These mutants remain valid C, finish normally, and change only answers. Sanitizers and the
// compiler cannot reject them: the source run on Node has to catch the wrong result.
func TestMathNumberOracleCatchesMutants(t *testing.T) {
	t.Parallel()
	mutants := []struct{ name, fixture, before, after string }{
		{"clz32", "math", "adamic_math_clz32(", "adamic_math_fround("},
		{"fround", "math", "adamic_math_fround(", "adamic_math_sign("},
		{"imul", "math", "adamic_math_imul(", "adamic_power("},
		{"constants", "math", "0x1.26bb1bbb55516p+01", "0x1.26bb1bbb55517p+01"},
		{"conversion", "convert", "adamic_number_from_string(", "adamic_number_parse_float("},
		{"own_property", "convert", "adamic_number_has_own_property(", "!adamic_number_has_own_property("},
		{"prototype", "prototype", "adamic_number_to_fixed(", "adamic_number_to_fixed(1 + "},
		{"catchable_fixed", "prototype", "adamic_number_checked_fixed(", "adamic_number_checked_fixed(1 + "},
		{"catchable_exponential", "prototype", "adamic_number_checked_exponential(", "adamic_number_checked_exponential(1 + "},
		{"catchable_precision", "prototype", "adamic_number_checked_precision(", "adamic_number_checked_precision(1 + "},
		{"catchable_radix", "prototype", "adamic_number_checked_radix(", "adamic_number_checked_radix(1 + "},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_math_number_"+mutant.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			if !strings.Contains(source, mutant.before) {
				t.Fatalf("mutant target %q absent", mutant.before)
			}
			source = strings.ReplaceAll(source, mutant.before, mutant.after)
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
