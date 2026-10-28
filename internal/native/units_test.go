package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: t.Setenv changes the process environment variable ADAMIC_GATE_UNCACHED.
func TestUnitsPreserveSharedState(t *testing.T) {
	source := `#include "adamic.h"
#include <stdio.h>
static int adamic_state = 0;
static int adamic_next(void);
static int adamic_next(void) { return ++adamic_state; }
int main(void) { int first = adamic_next(); int second = adamic_next(); printf("%d %d %d\n", first, second, adamic_state); return first == 1 && second == 2 && adamic_state == 2 ? 0 : 1; }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	nextHeader, nextUnits, err := splitC(source)
	if err != nil || header != nextHeader || len(units) != len(nextUnits) {
		t.Fatal("split is not deterministic")
	}
	for index := range units {
		if units[index] != nextUnits[index] {
			t.Fatal("split changed")
		}
	}
	for _, uncached := range []string{"0", "1"} {
		t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
		binary := filepath.Join(t.TempDir(), "program")
		if err := Build(source, binary, Options{Split: true, Sanitize: true, Jobs: 2}); err != nil {
			t.Fatal(err)
		}
		if output, err := exec.Command(binary).CombinedOutput(); err != nil || string(output) != "1 2 2\n" {
			t.Fatalf("shared state (uncached=%s): %v\n%s", uncached, err, output)
		}
	}
}

// The same bytes must be instrumented after a release compile. A flags-key mutant returns the
// release object and loses UBSan's check, while linking and executing normally. No warning kills it.
func TestUnitCacheFlagsHoldSanitizer(t *testing.T) {
	t.Parallel()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	cache := filepath.Join(directory, "cache")
	unit := compilationUnit{"probe.c", `int adamic_probe(int input) { return input + 2147483647; }` + "\n"}
	// The unit is its own first line, and nothing else is beside it.
	compile := func(flags []string) string {
		t.Helper()
		keying, err := newUnitKeying(unit.source, nil, flags, compiler, toolchain.Env())
		if err != nil {
			t.Fatal(err)
		}
		object, err := compileUnit(unit, nil, flags, compiler, keying, cache, directory, nil)
		if err != nil {
			t.Fatal(err)
		}
		return object
	}
	release := compile(Flags(Options{}))
	sanitized := compile(Flags(Options{Sanitize: true}))
	// Do not assert that paths differ: the runtime sanitizer observation is the proof.
	_ = release
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(`int adamic_probe(int); int main(void) { volatile int input = 1; volatile int result = adamic_probe(input); (void)result; return 0; }`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "probe")
	args := append(Flags(Options{Sanitize: true}), main, sanitized, "-o", binary)
	if output, err := toolchain.Command(compiler, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, output)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("signed integer overflow")) {
		t.Fatalf("sanitized rebuild reused uninstrumented object: exit=%v output=%s", err, output)
	}
	t.Log("release then sanitized: UBSan caught signed integer overflow")
}

func TestSplitTokensDoNotRewriteLiterals(t *testing.T) {
	t.Parallel()
	source := "#include \"adamic.h\"\nstatic int adamic_x = 1;\nstatic int adamic_f(void) { /* } ; */ const char *s = \"adamic_x } ; \\\"\"; (void)s; return adamic_x; }\nint main(void) { return adamic_f()-1; }\n"
	_, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(units[1].source, `"adamic_x } ; \""`) {
		t.Fatal("literal changed")
	}
	if !strings.Contains(units[1].source, "return adamic_unit_adamic_x;") {
		t.Fatal("reference not rewritten")
	}
}

// Regexp's compiled programs carry an alignment attribute after their name. Two of them in one
// program used to both be named __attribute__ and fail the split as duplicate definitions.
func TestSplitNamesAttributedDeclarations(t *testing.T) {
	t.Parallel()
	source := "#include \"adamic.h\"\nstatic const int adamic_a __attribute__((aligned(64))) = 1;\nstatic const int adamic_b __attribute__((aligned(64))) = 2;\nint main(void) { return adamic_a + adamic_b - 3; }\n"
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"adamic_unit_adamic_a", "adamic_unit_adamic_b"} {
		if !strings.Contains(header, "extern const int "+name+" __attribute__((aligned(64)));") {
			t.Fatalf("%s not declared with its attribute:\n%s", name, header)
		}
	}
	var all strings.Builder
	for _, unit := range units {
		all.WriteString(unit.source)
	}
	if strings.Contains(header+all.String(), "adamic_unit___attribute__") {
		t.Fatal("an attribute was taken for a name")
	}
	if !strings.Contains(all.String(), "return adamic_unit_adamic_a + adamic_unit_adamic_b - 3;") {
		t.Fatalf("references not rewritten:\n%s", all.String())
	}
}

// Not parallel: this check deliberately mutates a header between cache lookups.
func TestUnitSystemHeaderProvenance(t *testing.T) {
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	cache := filepath.Join(directory, "cache")
	// Clang accepts this extension in a system header, including when compiling cpp-output
	// with line markers. -P exposes it as user code and -pedantic -Werror must reject it.
	header := "#pragma clang system_header\nstruct probe_file { int (* _Nullable close)(void *); };\n#define PROBE_VALUE 1\n"
	firstLine := "#include \"probe-system.h\"\n"
	unit := compilationUnit{"probe.c", firstLine + "int adamic_probe(void) { return PROBE_VALUE; }\n"}
	flags := append(Flags(Options{}), "-Wno-system-headers")
	// The header is one of the files beside the unit, so a change to it moves the key by its file part.
	compile := func(header string, uncached bool) string {
		t.Helper()
		files := []runtimeFile{{"probe-system.h", []byte(header)}}
		var keying *unitKeying
		if !uncached {
			var err error
			if keying, err = newUnitKeying(firstLine, files, flags, compiler, toolchain.Env()); err != nil {
				t.Fatal(err)
			}
		}
		object, err := compileUnit(unit, files, flags, compiler, keying, cache, directory, nil)
		if err != nil {
			t.Fatalf("system-header provenance lost: %v", err)
		}
		return object
	}
	observe := func(object string, expected string) {
		t.Helper()
		main := filepath.Join(directory, "main.c")
		if err := os.WriteFile(main, []byte("#include <stdio.h>\nint adamic_probe(void); int main(void) { printf(\"%d\\n\", adamic_probe()); return 0; }\n"), 0644); err != nil {
			t.Fatal(err)
		}
		binary := filepath.Join(directory, "probe")
		if output, err := toolchain.Command(compiler, append(append([]string{}, flags...), main, object, "-o", binary)...).CombinedOutput(); err != nil {
			t.Fatalf("link: %v\n%s", err, output)
		}
		if output, err := exec.Command(binary).CombinedOutput(); err != nil || string(output) != expected {
			t.Fatalf("header observation: exit=%v output=%q want=%q", err, output, expected)
		}
	}
	first := compile(header, false)
	observe(first, "1\n")
	if second := compile(header, false); second != first {
		t.Fatal("unchanged header missed cache")
	}
	header = strings.Replace(header, "PROBE_VALUE 1", "PROBE_VALUE 2", 1)
	observe(compile(header, false), "2\n")
	observe(compile(header, true), "2\n")
	// The same extension must remain an error outside a system header. This prevents fixing
	// provenance by disabling nullability or pedantic diagnostics for the whole unit.
	userHeader := strings.TrimPrefix(header, "#pragma clang system_header\n")
	files := []runtimeFile{{"probe-system.h", []byte(userHeader)}}
	if _, err := compileUnit(unit, files, flags, compiler, nil, cache, directory, nil); err == nil || !strings.Contains(err.Error(), "-Wnullability-extension") {
		t.Fatalf("user-header extension must be rejected: %v", err)
	}
	t.Log("system-header extension accepted; changed header observed cached and uncached; user-header extension rejected")
}

// Attributed constants must keep distinct bindings after splitting. The mutant remains valid
// C with clean sanitizer and exit observations; only Node's stdout distinguishes it.
func TestSplitAttributedBindingsAgreeWithNode(t *testing.T) {
	t.Parallel()
	run := func(command string, args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, command, args...)
		var stderr strings.Builder
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if err != nil || stderr.Len() != 0 {
			t.Fatalf("execution: %v stderr=%s", err, &stderr)
		}
		return string(output)
	}
	expected := run("node", "-e", "const a = 1, b = 2; console.log(10 * a + b)")
	source := `#include "adamic.h"
#include <stdio.h>
static const int adamic_a __attribute__((aligned(64))) = 1;
static const int adamic_b __attribute__((aligned(64))) = 2;
int main(void) { printf("%d\n", 10 * adamic_a + adamic_b); return 0; }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	var joined strings.Builder
	joined.WriteString(header)
	for _, unit := range units {
		joined.WriteString(strings.TrimPrefix(unit.source, "#include \"units.h\"\n"))
	}
	for _, target := range []string{"native", "wasm32-wasi"} {
		t.Run(target, func(t *testing.T) {
			for _, mutant := range []bool{false, true} {
				code := joined.String()
				if target == "native" {
					code = source
				}
				if mutant {
					if target == "native" {
						code = strings.Replace(code, "10 * adamic_a + adamic_b", "10 * adamic_a + adamic_a", 1)
					} else {
						code = strings.Replace(code, "10 * adamic_unit_adamic_a + adamic_unit_adamic_b", "10 * adamic_unit_adamic_a + adamic_unit_adamic_a", 1)
					}
				}
				binary := filepath.Join(t.TempDir(), "program")
				options := Options{Split: true, Sanitize: true, Jobs: 2}
				if target == "wasm32-wasi" {
					options = Options{Target: target}
				}
				if err := Build(code, binary, options); err != nil {
					t.Fatal(err)
				}
				actual := ""
				if target == "native" {
					actual = run(binary)
				} else {
					actual = run("node", "--disable-warning=ExperimentalWarning", "../../oracle/wasi.mjs", binary)
				}
				if !mutant && actual != expected {
					t.Fatalf("control stdout=%q Node=%q", actual, expected)
				}
				if mutant && actual == expected {
					t.Fatal("binding mutant survived Node comparison")
				}
				t.Logf("mutant=%t: clean exit/stderr; stdout=%q Node=%q", mutant, actual, expected)
			}
		})
	}
}

func TestSplitRuntimePrototypes(t *testing.T) {
	t.Parallel()
	source := `#include "adamic.h"
adamic_array *adamic_object_values_checked(adamic_object *, int, const char *, bool, const adamic_value *, size_t);
static int adamic_probe(void) { return 0; }
int main(void) { return adamic_probe(); }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(header, "adamic_array *adamic_object_values_checked(") || strings.Contains(header, "adamic_unit_adamic_object_values_checked") {
		t.Fatalf("runtime prototype linkage changed: %s", header)
	}
	for _, unit := range units {
		if strings.Contains(unit.source, "adamic_object_values_checked") {
			t.Fatalf("runtime prototype became a definition: %s", unit.source)
		}
	}
	for _, rejected := range []string{
		"adamic_array *adamic_external;\n",
		"adamic_array *adamic_external = NULL;\n",
		"adamic_array *adamic_external(void) { return NULL; }\n",
	} {
		if _, _, err := splitC(rejected); err == nil {
			t.Fatalf("accepted non-static definition: %s", rejected)
		}
	}
	binary := filepath.Join(t.TempDir(), "split")
	if err := Build(source, binary, Options{Split: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil || len(output) != 0 {
		t.Fatalf("split runtime prototype: %v %s", err, output)
	}
}

// One cached build mixes hits and misses, and its binary is the uncached one: a program of three function units built
// cached, then one function in its last unit changed and built cached again, so that unit misses and the others hit.
// The mixed build's binary runs the changed body and is the same bytes as the changed program built uncached. This holds
// at a few units what TestSplitTSGoAgrees's cached build of the 4 MB checker held at 23 (#dgcy49s). Mutant: a unit's
// key without its own bytes, so the changed unit hits the old object, and the binary prints the old sum.
// Not parallel: t.Setenv changes ADAMIC_BUILD_LOG and ADAMIC_GATE_UNCACHED for the process.
func TestSplitMixesHitsAndMisses(t *testing.T) {
	// The names are this run's own, so the first cached build compiles every unit whatever a shared cache holds.
	nonce := strconv.FormatInt(time.Now().UnixNano(), 36)
	program := func(last int) string {
		var source strings.Builder
		source.WriteString("#include \"adamic.h\"\n#include <stdio.h>\n")
		for index := range 40 {
			value := index
			if index == 39 {
				value = last
			}
			fmt.Fprintf(&source, "static int adamic_%s_%d(void) { return %d; }\n", nonce, index, value)
		}
		source.WriteString("int main(void) { int sum = 0;")
		for index := range 40 {
			fmt.Fprintf(&source, " sum += adamic_%s_%d();", nonce, index)
		}
		source.WriteString(" printf(\"%d\\n\", sum); return 0; }\n")
		return source.String()
	}
	directory := t.TempDir()
	build := func(name, source, uncached string) ([]byte, map[string]string) {
		t.Helper()
		log := filepath.Join(directory, name+".log")
		t.Setenv("ADAMIC_BUILD_LOG", log)
		t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
		binary := filepath.Join(directory, name, "program")
		if err := os.MkdirAll(filepath.Dir(binary), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := Build(source, binary, Options{Split: true, Jobs: 2}); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(binary).CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v\n%s", name, err, output)
		}
		written, err := os.ReadFile(log)
		if err != nil {
			t.Fatal(err)
		}
		outcomes := map[string]string{}
		for _, line := range strings.Split(strings.TrimSuffix(string(written), "\n"), "\n") {
			if match := buildLogLine.FindStringSubmatch(line); match != nil && match[1] == "cCompile" && strings.HasPrefix(match[2], "functions_") {
				outcomes[match[2]] = match[4]
			}
		}
		contents, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		return append(output, contents...), outcomes
	}
	_, first := build("first", program(39), "0")
	if len(first) != 3 || first["functions_0000.c"] != "miss" || first["functions_0001.c"] != "miss" || first["functions_0002.c"] != "miss" {
		t.Fatalf("the first cached build's units: %v, want three misses", first)
	}
	mixed, outcomes := build("mixed", program(1000), "0")
	if outcomes["functions_0000.c"] != "hit" || outcomes["functions_0001.c"] != "hit" || outcomes["functions_0002.c"] != "miss" {
		t.Fatalf("the changed build's units: %v, want two hits and the changed unit's miss", outcomes)
	}
	uncached, _ := build("uncached", program(1000), "1")
	if want := "1741\n"; !bytes.HasPrefix(mixed, []byte(want)) {
		t.Fatalf("the mixed build printed %q, want %q", mixed[:bytes.IndexByte(mixed, '\n')+1], want)
	}
	if !bytes.Equal(mixed, uncached) {
		t.Fatal("the mixed build's binary differs from the changed program built uncached")
	}
}
