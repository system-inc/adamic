package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNumberImmediateNodeMutant(t *testing.T) {
	t.Parallel()
	numberMathMutant(t, "library_number_immediate.a", "adamic_number_from_string(", "adamic_number_parse_float(")
}
func TestNumberConversionSourcesNodeMutant(t *testing.T) {
	t.Parallel()
	numberMathMutant(t, "library_number_convert_sources.a", "adamic_number_from_string(", "adamic_number_parse_float(")
}
func TestNumberOwnNamesNodeMutant(t *testing.T) {
	t.Parallel()
	numberMathMutant(t, "library_number_own_names.a", "adamic_number_has_own_property(", "!adamic_number_has_own_property(")
}
func TestNumberMathEdgesNodeMutant(t *testing.T) {
	t.Parallel()
	numberMathMutant(t, "library_number_math_edges.a", "adamic_math_imul(", "adamic_power(")
}

func numberMathMutant(t *testing.T, file, before, after string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/number_math", file))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	if !strings.Contains(source, before) {
		t.Fatal("mutation anchor absent")
	}
	if file == "library_number_convert_sources.a" {
		after = "probe_number_parse_float("
	}
	source = strings.ReplaceAll(source, before, after)
	if file == "library_number_convert_sources.a" {
		source = "#include \"adamic.h\"\nstatic double probe_number_parse_float(const adamic_string *text) { return text == NULL ? NAN : adamic_number_parse_float(text); }\n" + source
	}
	numberMathCleanNodeMutant(t, path, source)
}
func TestNumberFormatErrorMessageNodeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_math_number_prototype.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	implementation, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/library_number_format.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(implementation)
	before := "toFixed() digits argument must be between 0 and 100"
	if strings.Count(source, before) != 1 {
		t.Fatal("mutation anchor absent")
	}
	source = strings.Replace(source, before, "wrong digits range", 1) + "\n" + native.C(program)
	for _, name := range []string{"fixed", "exponential", "precision", "radix"} {
		source = strings.ReplaceAll(source, "adamic_number_checked_"+name+"(", "probe_number_checked_"+name+"(")
	}
	numberMathCleanNodeMutant(t, path, source)
}
func numberMathCleanNodeMutant(t *testing.T, path, source string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("must finish cleanly: %+v", actual)
	}
	if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
		t.Fatalf("Node alone must catch mutant, got %q", difference)
	}
	t.Log("clean exit and sanitizers; caught only by Node stdout")
}

func TestMathContractionNodeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/number_math/contraction.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	before := "((adamic_local_0_value * (0x1.4p+03)) - (0x1p+00))"
	if strings.Count(source, before) != 1 {
		t.Fatal("contraction mutation anchor absent")
	}
	source = strings.Replace(source, before, "fma(adamic_local_0_value, 0x1.4p+03, -0x1p+00)", 1)
	numberMathCleanNodeMutant(t, path, source)
}
