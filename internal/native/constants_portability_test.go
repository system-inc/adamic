package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Diagnostics remain nonfatal for mutants: the warning assertions, rather than
// -Werror, must catch each portability regression.
func TestConstantPortabilityClang(t *testing.T) {
	t.Parallel()
	constantPortability(t, "clang")
}

func TestConstantPortabilityGCC(t *testing.T) {
	t.Parallel()
	constantPortability(t, "gcc")
}

func TestLongStringBytesMutant(t *testing.T) {
	t.Parallel()
	constantPortabilityMutant(t, "static const unsigned char adamic_bytes_", "static const char adamic_bytes_", "[-Woverflow]", 4200)
}

func TestStaticNumberInitializerMutant(t *testing.T) {
	t.Parallel()
	constantPortabilityMutant(t, "= {false, 0.0};", "= (adamic_maybe_number){false, 0.0};", "[-Wpedantic]", 1)
}

func TestStaticBooleanInitializerMutant(t *testing.T) {
	t.Parallel()
	constantPortabilityMutant(t, "= {false, false};", "= (adamic_maybe_boolean){false, false};", "[-Wpedantic]", 1)
}

func portabilityConstants() string {
	return C(&ir.Program{Source: "constants.a", Strings: []string{strings.Repeat("é🌍", 700)}, Locals: []ir.Local{
		{Name: "number", Type: ir.MaybeNumber, Global: true},
		{Name: "boolean", Type: ir.MaybeBoolean, Global: true},
	}})
}

func portabilityHeaders(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	entries, err := runtime.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".h") {
			continue
		}
		data, err := runtime.ReadFile("runtime/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func compilePortability(t *testing.T, compiler, source string, fatalWarnings bool) string {
	t.Helper()
	directory := portabilityHeaders(t)
	file := filepath.Join(directory, "main.c")
	if err := os.WriteFile(file, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	flags := []string{"-std=c11", "-Wpedantic", "-Wconversion", "-fsyntax-only", "-fsigned-char", "-I", directory, file}
	if fatalWarnings {
		flags = append(flags, "-Werror")
	}
	output, err := exec.Command(compiler, flags...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s compile: %v\n%s", compiler, err, output)
	}
	return string(output)
}

func constantPortability(t *testing.T, compiler string) {
	t.Helper()
	if output := compilePortability(t, compiler, portabilityConstants(), true); output != "" {
		t.Fatalf("constant diagnostics:\n%s", output)
	}
}

func constantPortabilityMutant(t *testing.T, old, replacement, diagnostic string, warnings int) {
	t.Helper()
	source := portabilityConstants()
	changed := strings.Replace(source, old, replacement, 1)
	if changed == source {
		t.Fatal("no mutant target")
	}
	output := compilePortability(t, "gcc", changed, false)
	if count := strings.Count(output, diagnostic); count != warnings {
		t.Fatalf("got %d %s warnings, want %d\n%s", count, diagnostic, warnings, output)
	}
	if count := strings.Count(output, "warning:"); count != warnings {
		t.Fatalf("unexpected additional diagnostics: %d warnings, want %d", count, warnings)
	}
	t.Logf("mutant caught with -Werror absent: %d %s warnings", warnings, diagnostic)
}
