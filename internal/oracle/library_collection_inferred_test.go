package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, family := range []string{"map", "set"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_" + family + "_inferred.a", true, false})
	}
}

// Membership mutants execute cleanly; neither compilation nor sanitizers detect the wrong answer.
func TestLibraryInferredCollectionMutants(t *testing.T) {
	for _, family := range []string{"map", "set"} {
		t.Run(family, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_"+family+"_inferred.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			mutant := strings.ReplaceAll(code, " != NULL)", " == NULL)")
			if mutant == code {
				t.Fatal("membership mutant changed no code")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must execute cleanly: exit %d stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("Node must catch mutant: %q", difference)
			}
			t.Log("clean exit 0, no sanitizer finding; only Node stdout comparison caught inverted membership")
		})
	}
}
