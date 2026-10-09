package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/string_lastindexof_position.a", true, false})
}

func TestStringLastIndexOfPositionMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_lastindexof_position.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	mutant := strings.ReplaceAll(code, "adamic_string_last_index_of_from(", "last_index_mutant(")
	if mutant == code {
		t.Fatal("mutant changed no code")
	}
	mutant = insertCollectionMutant(mutant, `static double last_index_mutant(const adamic_string *text, const adamic_string *search, double position) {
 return adamic_string_last_index_of_from(text, search, isnan(position) ? 0 : position);
}`)
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
	t.Log("NaN treated as zero: clean exit 0, no sanitizer finding; only Node stdout comparison caught it")
}
