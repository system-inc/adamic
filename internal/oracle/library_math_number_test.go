package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// These mutants remain valid C, finish normally, and change only answers. Sanitizers and the
// compiler cannot reject them: the source run on Node has to catch the wrong result.
func TestMathNumberOracleCatchesMutants(test *testing.T) {
	test.Parallel()
	mutants := []struct{ name, fixture, before, after string }{
		{"clz32", "math", "adamic_math_clz32(", "adamic_math_fround("},
		{"fround", "math", "adamic_math_fround(", "adamic_math_sign("},
		{"imul", "math", "adamic_math_imul(", "adamic_power("},
		{"constants", "math", "0x1.26bb1bbb55516p+01", "0x1.26bb1bbb55517p+01"},
		{"conversion", "convert", "adamic_number_from_string(", "adamic_number_parse_float("},
		{"own_property", "convert", "adamic_number_has_own_property(", "!adamic_number_has_own_property("},
		{"prototype", "prototype", "adamic_number_to_fixed(", "adamic_number_to_fixed(1 + "},
	}
	for _, mutant := range mutants {
		test.Run(mutant.name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/library_math_number_"+mutant.fixture+".a"),
			)
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			source := native.C(program)
			if !strings.Contains(source, mutant.before) {
				test.Fatalf("mutant target %q absent", mutant.before)
			}
			source = strings.ReplaceAll(source, mutant.before, mutant.after)
			binary := filepath.Join(test.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				test.Fatal(err)
			}
			result := execute(test, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				test.Fatalf("mutant must finish cleanly: exit %d, stderr %q", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(test, path), result); difference != "stdout differs" {
				test.Fatalf("got %q, want stdout differs", difference)
			}
			test.Log("Node caught the mutant: stdout differs; sanitizer clean")
		})
	}
}
