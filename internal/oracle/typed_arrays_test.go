package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"sort_uint8array", "sort_uint16array", "sort_int32array", "sort_float64array", "uint8array", "uint16array", "uint16_stop", "int32array", "float64array", "views", "order", "primes", "primes_large", "stats", "workers_stats", "stop", "integer", "integer_stop"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/typed_arrays_" + name + ".a", true, name == "stop" || name == "uint16_stop" || name == "integer_stop",
		})
	}
}

// This intentional difference has independent pins for both backends and Node.
// Removing the check in both backends must still fail, even though they agree.
func TestTypedArrayWriteStopIsPinned(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"stop", "uint16_stop"} {
		t.Run(kind, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_"+kind+".a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			stopped := run{stdout: []byte("before 2 7 undefined\n"), stderr: []byte("adamic: panic: index 2 is outside an array of length 2\n"), exitCode: 70}
			node := run{stdout: []byte("before 2 7 undefined\nafter 2 7 undefined\n")}
			native, _ := natively(t, p)
			for name, observed := range map[string]run{"native": native, "release": released(t, p), "JavaScript": onJavaScriptBackend(t, p)} {
				if difference := disagreement(stopped, observed); difference != "" {
					t.Errorf("%s: %s: exit %d stdout %q stderr %q", name, difference, observed.exitCode, observed.stdout, observed.stderr)
				}
			}
			if observed := onNode(t, path); disagreement(node, observed) != "" {
				t.Errorf("Node no longer silently drops the write: %#v", observed)
			}
		})
	}
}

// Original platform functions run on Node independently of the typed-array port.
// This pins the port's arithmetic, not just native agreement with its own source.
func TestTypedArrayWorkersStatsMatchesPlatforms(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_workers_stats.a"))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_workers_stats_reference.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, reference)
	if expected.exitCode != 0 {
		t.Fatalf("platforms Node reference: %#v", expected)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := natively(t, p)
	for name, observed := range map[string]run{"source": onNode(t, path), "native": native, "release": released(t, p), "JavaScript": onJavaScriptBackend(t, p)} {
		if difference := disagreement(expected, observed); difference != "" {
			t.Errorf("%s: %s: exit %d stdout %q stderr %q", name, difference, observed.exitCode, observed.stdout, observed.stderr)
		}
	}
}

// Copy the shipping inline helper into the generated unit so the mutant changes
// its actual bounds check without changing the embedded runtime or other tests.
func TestTypedArrayIntegerBoundsMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_integer.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	if strings.Count(control, "adamic_typed_array_get_integer(") != 8 || strings.Count(control, "adamic_typed_array_set_integer(") != 4 {
		t.Fatal("all four kinds must use integer reads and writes")
	}
	header, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/adamic.h"))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(header), "static inline adamic_maybe_number adamic_typed_array_get_integer(")
	if start < 0 {
		t.Fatal("integer helper missing")
	}
	body := string(header)[start:]
	end := strings.Index(body, "\n}\n")
	if end < 0 {
		t.Fatal("integer helper end missing")
	}
	body = body[:end+3]
	check := "\tif (index < 0 || (uint64_t)index >= array->length) {\n\t\treturn (adamic_maybe_number){false, 0};\n\t}\n"
	if !strings.Contains(body, check) {
		t.Fatal("bounds mutant changed nothing")
	}
	body = strings.Replace(body, check, "", 1)
	body = strings.ReplaceAll(body, "adamic_typed_array_get_integer", "mutant_get_integer")
	code := strings.ReplaceAll(control, "adamic_typed_array_get_integer(", "mutant_get_integer(")
	code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+body, 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-buffer-overflow") {
		t.Fatalf("integer bounds mutant survived: %+v", result)
	}
	t.Log("ASan caught the integer bounds mutant")
}

func TestTypedArrayIntegerWriteStopIsPinned(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_integer_stop.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(native.C(program), "adamic_typed_array_set_integer(") {
		t.Fatal("bound write must use integer helper")
	}
	stopped := run{stdout: []byte("0: 7\n1: 8\n2: undefined\n"), stderr: []byte("adamic: panic: index 2 is outside an array of length 2\n"), exitCode: 70}
	for name, observed := range map[string]run{"native": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(stopped, observed); difference != "" {
			t.Errorf("%s: %s: %+v", name, difference, observed)
		}
	}
	expected := run{stdout: []byte("0: 7\n1: 8\n2: undefined\nafter\n")}
	if observed := onNode(t, path); disagreement(expected, observed) != "" {
		t.Fatalf("Node: %+v", observed)
	}
}
