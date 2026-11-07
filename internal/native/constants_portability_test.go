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
func TestConstantPortabilityClang(test *testing.T) {
	test.Parallel()
	constantPortability(test, "clang")
}

func TestConstantPortabilityGCC(test *testing.T) {
	test.Parallel()
	constantPortability(test, "gcc")
}

func TestLongStringBytesMutant(test *testing.T) {
	test.Parallel()
	constantPortabilityMutant(
		test,
		"static const unsigned char adamic_string_",
		"static const char adamic_string_",
		"[-Woverflow]",
		4200,
	)
}

func TestStaticNumberInitializerMutant(test *testing.T) {
	test.Parallel()
	constantPortabilityMutant(test, "= {false, 0.0};", "= (adamic_maybe_number){false, 0.0};", "[-Wpedantic]", 1)
}

func TestStaticBooleanInitializerMutant(test *testing.T) {
	test.Parallel()
	constantPortabilityMutant(test, "= {false, false};", "= (adamic_maybe_boolean){false, false};", "[-Wpedantic]", 1)
}

func portabilityConstants() string {
	return C(&ir.Program{Source: "constants.a", Strings: []string{strings.Repeat("é🌍", 700)}, Locals: []ir.Local{
		{Name: "number", Type: ir.MaybeNumber, Global: true},
		{Name: "boolean", Type: ir.MaybeBoolean, Global: true},
	}})
}

func portabilityHeaders(test *testing.T) string {
	test.Helper()
	directory := test.TempDir()
	entries, err := runtime.ReadDir("runtime")
	if err != nil {
		test.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".h") {
			continue
		}
		data, err := runtime.ReadFile("runtime/" + entry.Name())
		if err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0644); err != nil {
			test.Fatal(err)
		}
	}
	return directory
}

func compilePortability(test *testing.T, compiler, source string, fatalWarnings bool) string {
	test.Helper()
	directory := portabilityHeaders(test)
	file := filepath.Join(directory, "main.c")
	if err := os.WriteFile(file, []byte(source), 0644); err != nil {
		test.Fatal(err)
	}
	flags := []string{"-std=c11", "-Wpedantic", "-Wconversion", "-fsyntax-only", "-fsigned-char", "-I", directory, file}
	if fatalWarnings {
		flags = append(flags, "-Werror")
	}
	output, err := exec.Command(compiler, flags...).CombinedOutput()
	if err != nil {
		test.Fatalf("%s compile: %v\n%s", compiler, err, output)
	}
	return string(output)
}

func constantPortability(test *testing.T, compiler string) {
	test.Helper()
	if output := compilePortability(test, compiler, portabilityConstants(), true); output != "" {
		test.Fatalf("constant diagnostics:\n%s", output)
	}
}

func constantPortabilityMutant(test *testing.T, old, replacement, diagnostic string, warnings int) {
	test.Helper()
	source := portabilityConstants()
	changed := strings.Replace(source, old, replacement, 1)
	if changed == source {
		test.Fatal("no mutant target")
	}
	output := compilePortability(test, "gcc", changed, false)
	if count := strings.Count(output, diagnostic); count != warnings {
		test.Fatalf("got %d %s warnings, want %d\n%s", count, diagnostic, warnings, output)
	}
	if count := strings.Count(output, "warning:"); count != warnings {
		test.Fatalf("unexpected additional diagnostics: %d warnings, want %d", count, warnings)
	}
	test.Logf("mutant caught with -Werror absent: %d %s warnings", warnings, diagnostic)
}
