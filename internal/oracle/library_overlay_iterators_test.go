package oracle

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The ledger helpers are ambient declarations. Their exact checking witnesses live in
// internal/load/testdata/overlay-iterators; these executable counterparts use the concrete
// collection iterator bodies currently supported by lowering.
func init() {
	for _, name := range []string{"iterable", "callback", "entries", "set_copy", "nested_entries", "done"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_overlay_" + name + ".a", true, false})
	}
}

func TestLibraryIteratorDoneMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_overlay_done.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	fallback := regexp.MustCompile(`(} else \{\n\s*adamic_temporary_[0-9]+ = )false;`)
	mutant := fallback.ReplaceAllString(code, "${1}true;")
	if code == mutant {
		t.Fatal("mutant changed no optional done fallback")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant must run cleanly: exit %d, stderr %s", actual.exitCode, actual.stderr)
	}
	if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
		t.Fatalf("Node did not catch mutant: %q", difference)
	}
	t.Log("omitted done treated as true: clean exit 0, no sanitizer finding, caught only by Node stdout comparison")
}
