package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/generic-optional-callback-result.a", "internal/oracle/testdata/generic_optional_callback_results.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestGenericOptionalUndefinedMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generic_optional_callback_results.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "(adamic_maybe_number){false, 0.0}", "(adamic_maybe_number){true, 0.0}")
	if mutant == source {
		t.Fatal("undefined mutant target absent")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant did not finish cleanly: %d %s", result.exitCode, result.stderr)
	}
	if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
		t.Fatalf("want stdout differs, got %q", difference)
	}
	t.Logf("Node caught zero instead of undefined: %q", result.stdout)
}
