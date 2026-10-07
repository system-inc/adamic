package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/e4eec87_u03_discriminated_undefined.a",
		"internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// Restore the old numeric read and the zero bits of an undefined reference slot. This is valid
// C and runs cleanly under the sanitizers; only Node's output exposes the lost undefined.
func TestLiteralUndefinedOracleCatchesMutant(t *testing.T) {
	for _, fixture := range []string{"e4eec87_u03_discriminated_undefined.a", "e4eec87_u01_undefined_field_widened.a"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			read := regexp.MustCompile(`adamic_object_maybe_number\(([^,]+), ([^,]+), (&[^)]+)\)`)
			mutant := read.ReplaceAllString(source, "adamic_maybe_number_unpack(adamic_object_field($1, $2, $3)->number)")
			mutant = strings.ReplaceAll(mutant, "adamic_maybe_number_pack((adamic_maybe_number){false, 0.0})", "0.0")
			if mutant == source {
				t.Fatal("mutant target absent")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := execute(t, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d, stderr %q", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("got %q, want stdout differs", difference)
			}
			t.Logf("Node caught restored undefined slot bug: stdout %q", result.stdout)
		})
	}
}
