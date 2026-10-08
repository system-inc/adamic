package oracle

import (
	"os"
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

// Project roots are generated .ts copies of the authored .a fixtures. The ledger's
// ambient witnesses themselves are checked by internal/load, not executed.
func TestProjectIteratorBackends(t *testing.T) {
	for _, name := range []string{"iterable", "callback", "entries", "set_copy", "nested_entries", "done"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/library_overlay_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			path := filepath.Join(directory, "main.ts")
			if err := os.WriteFile(path, source, 0644); err != nil {
				t.Fatal(err)
			}
			config := `{"compilerOptions":{"strict":true,"target":"es2024","lib":["es2024"],"types":[],"noEmit":true},"files":["main.ts"]}`
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			actual, binary := natively(t, program)
			if difference := disagreement(expected, actual); difference != "" {
				t.Fatalf("native: %s", difference)
			}
			if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "" {
				t.Fatalf("JavaScript: %s", difference)
			}
			if difference := disagreement(expected, released(t, program)); difference != "" {
				t.Fatalf("release: %s", difference)
			}
			if leaked := leaks(t, program, binary); leaked != "" {
				t.Fatalf("leaks: %s", leaked)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "" {
					t.Fatalf("WASI: %s", difference)
				}
			}
		})
	}
}
