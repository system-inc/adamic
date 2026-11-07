package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{
		"sort_uint8array",
		"sort_uint16array",
		"sort_int32array",
		"sort_float64array",
		"integer",
		"integer_stop",
		"uint8array",
		"uint16array",
		"uint16_stop",
		"int32array",
		"float64array",
		"views",
		"order",
		"primes",
		"primes_large",
		"stats",
		"workers_stats",
		"stop",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/typed_arrays_" + name + ".a",
			true,
			name == "stop" || name == "uint16_stop" || name == "integer_stop",
		})
	}
}

// This intentional difference has independent pins for both backends and Node.
// Removing the check in both backends must still fail, even though they agree.
func TestTypedArrayWriteStopIsPinned(test *testing.T) {
	test.Parallel()
	for _, kind := range []string{"stop", "uint16_stop"} {
		test.Run(kind, func(test *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_"+kind+".a"))
			if err != nil {
				test.Fatal(err)
			}
			p, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			stopped := run{
				stdout:   []byte("before 2 7 undefined\n"),
				stderr:   []byte("adamic: panic: index 2 is outside an array of length 2\n"),
				exitCode: 70,
			}
			node := run{stdout: []byte("before 2 7 undefined\nafter 2 7 undefined\n")}
			native, _ := natively(test, p)
			for name, observed := range map[string]run{
				"native":     native,
				"release":    released(test, p),
				"JavaScript": onJavaScriptBackend(test, p),
			} {
				if difference := disagreement(stopped, observed); difference != "" {
					test.Errorf(
						"%s: %s: exit %d stdout %q stderr %q",
						name,
						difference,
						observed.exitCode,
						observed.stdout,
						observed.stderr,
					)
				}
			}
			if observed := onNode(test, path); disagreement(node, observed) != "" {
				test.Errorf("Node no longer silently drops the write: %#v", observed)
			}
		})
	}
}

// Original platform functions run on Node independently of the typed-array port.
// This pins the port's arithmetic, not just native agreement with its own source.
func TestTypedArrayWorkersStatsMatchesPlatforms(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_workers_stats.a"))
	if err != nil {
		test.Fatal(err)
	}
	reference, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/typed_arrays_workers_stats_reference.mjs"),
	)
	if err != nil {
		test.Fatal(err)
	}
	expected := onNode(test, reference)
	if expected.exitCode != 0 {
		test.Fatalf("platforms Node reference: %#v", expected)
	}
	p, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	native, _ := natively(test, p)
	for name, observed := range map[string]run{
		"source":     onNode(test, path),
		"native":     native,
		"release":    released(test, p),
		"JavaScript": onJavaScriptBackend(test, p),
	} {
		if difference := disagreement(expected, observed); difference != "" {
			test.Errorf(
				"%s: %s: exit %d stdout %q stderr %q",
				name,
				difference,
				observed.exitCode,
				observed.stdout,
				observed.stderr,
			)
		}
	}
}

// Copy the shipping inline helper into the generated unit so the mutant changes
// its actual bounds check without changing the embedded runtime or other tests.
func TestTypedArrayIntegerBoundsMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_integer.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	control := native.C(program)
	if strings.Count(control, "adamic_typed_array_get_integer(") != 13 ||
		strings.Count(control, "adamic_typed_array_set_integer(") != 4 {
		test.Fatal("all four kinds must use integer reads and writes")
	}
	header, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/adamic.h"))
	if err != nil {
		test.Fatal(err)
	}
	start := strings.Index(string(header), "static inline adamic_maybe_number adamic_typed_array_get_integer(")
	if start < 0 {
		test.Fatal("integer helper missing")
	}
	body := string(header)[start:]
	end := strings.Index(body, "\n}\n")
	if end < 0 {
		test.Fatal("integer helper end missing")
	}
	body = body[:end+3]
	check := "\tif (index < 0 || (uint64_t)index >= array->length) {\n\t\treturn (adamic_maybe_number){false, 0};\n\t}\n"
	if !strings.Contains(body, check) {
		test.Fatal("bounds mutant changed nothing")
	}
	body = strings.Replace(body, check, "", 1)
	body = strings.ReplaceAll(body, "adamic_typed_array_get_integer", "mutant_get_integer")
	code := strings.ReplaceAll(control, "adamic_typed_array_get_integer(", "mutant_get_integer(")
	code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+body, 1)
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-buffer-overflow") {
		test.Fatalf("integer bounds mutant survived: %+v", result)
	}
	test.Log("ASan caught the integer bounds mutant")
}

func TestTypedArrayIntegerWriteStopIsPinned(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_integer_stop.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	if !strings.Contains(native.C(program), "adamic_typed_array_set_integer(") {
		test.Fatal("bound write must use integer helper")
	}
	stopped := run{
		stdout:   []byte("0: 7\n1: 8\n2: undefined\n"),
		stderr:   []byte("adamic: panic: index 2 is outside an array of length 2\n"),
		exitCode: 70,
	}
	for name, observed := range map[string]run{
		"native":     released(test, program),
		"JavaScript": onJavaScriptBackend(test, program),
	} {
		if difference := disagreement(stopped, observed); difference != "" {
			test.Errorf("%s: %s: %+v", name, difference, observed)
		}
	}
	expected := run{stdout: []byte("0: 7\n1: 8\n2: undefined\nafter\n")}
	if observed := onNode(test, path); disagreement(expected, observed) != "" {
		test.Fatalf("Node: %+v", observed)
	}
}

// The indexed loop must own its original array while the body reassigns the
// source binding. Node keeps that original array through its built-in iterator.
func TestTypedArrayIterationOwnerMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_integer.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	control := native.C(program)
	if strings.Contains(control, "adamic_typed_array_iterate(") {
		test.Fatal("typed-array for-of still allocates an iterator")
	}
	holder := regexp.MustCompile(
		`(?m)^(\s*adamic_typed_array \*(adamic_temporary_[A-Za-z0-9_]+) = [^;]+;\n)(\s*for \(size_t )`,
	)
	if len(holder.FindAllStringIndex(control, -1)) < 2 {
		test.Fatal("missing live and reassigned array holders")
	}
	mutant := holder.ReplaceAllString(control, "$1\tadamic_release($2);\n$3")
	binary := filepath.Join(test.TempDir(), "unowned-iteration")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal("ownership mutant must compile:", err)
	}
	observed := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"}, binary)
	if observed.exitCode == 0 || !strings.Contains(string(observed.stderr), "heap-use-after-free") {
		test.Fatalf("original-array ownership mutant survived: %+v", observed)
	}
	test.Log("ASan caught the original array owner dropped before iteration")
}
