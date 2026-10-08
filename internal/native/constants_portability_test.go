package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Diagnostics are inspected with warnings nonfatal: the assertions themselves
// must reject the portability regressions, even when execution would agree.
func TestConstantPortability(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "constants.a", Strings: []string{strings.Repeat("é🌍", 700)}, Locals: []ir.Local{
		{Name: "number", Type: ir.MaybeNumber, Global: true},
		{Name: "boolean", Type: ir.MaybeBoolean, Global: true},
	}}
	source := C(program)
	for _, compiler := range []string{"clang", "gcc"} {
		t.Run(compiler, func(t *testing.T) {
			if _, err := exec.LookPath(compiler); err != nil {
				t.Skip(err)
			}
			directory := t.TempDir()
			headers, err := runtime.ReadDir("runtime")
			if err != nil {
				t.Fatal(err)
			}
			for _, header := range headers {
				if !strings.HasSuffix(header.Name(), ".h") {
					continue
				}
				data, err := runtime.ReadFile("runtime/" + header.Name())
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(directory, header.Name()), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			compile := func(code string) string {
				t.Helper()
				path := filepath.Join(directory, "main.c")
				if err := os.WriteFile(path, []byte(code), 0644); err != nil {
					t.Fatal(err)
				}
				output, err := exec.Command(compiler, "-std=c11", "-pedantic", "-Wconversion", "-fsyntax-only", "-fsigned-char", "-I", directory, path).CombinedOutput()
				if err != nil {
					t.Fatalf("compile: %v\n%s", err, output)
				}
				return string(output)
			}
			if output := compile(source); output != "" {
				t.Fatalf("constant diagnostics:\n%s", output)
			}
			mutants := []struct {
				name, old, replacement, diagnostic string
				warnings                           int
			}{
				{"signed bytes", "static const unsigned char adamic_bytes_", "static const char adamic_bytes_", "[-Woverflow]", 4200},
				{"number compound literal", "= {false, 0.0};", "= (adamic_maybe_number){false, 0.0};", "[-Wpedantic]", 1},
				{"boolean compound literal", "= {false, false};", "= (adamic_maybe_boolean){false, false};", "[-Wpedantic]", 1},
			}
			for _, mutant := range mutants {
				// Clang accepts both constructs without these diagnostic flags;
				// GCC pins the warning counts, with no -Werror.
				if compiler == "clang" {
					continue
				}
				changed := strings.Replace(source, mutant.old, mutant.replacement, 1)
				if changed == source {
					t.Fatalf("%s: no mutant target", mutant.name)
				}
				output := compile(changed)
				if count := strings.Count(output, mutant.diagnostic); count != mutant.warnings {
					t.Fatalf("%s: got %d %s warnings, want %d", mutant.name, count, mutant.diagnostic, mutant.warnings)
				}
				t.Logf("%s caught: %d warnings", mutant.name, strings.Count(output, "warning:"))
			}
		})
	}
}
