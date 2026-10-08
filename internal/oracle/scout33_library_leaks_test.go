package oracle

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var scout33Workloads = []string{"collections", "array", "number", "math", "date", "node_host", "json", "regexp", "string", "scout_22_tsc_keyword_entries"}

func init() {
	for _, name := range scout33Workloads {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/scout33/" + name + ".a", true, false})
	}
}

func scout33Family(path string) string {
	switch {
	case strings.Contains(path, "date") && !strings.Contains(path, "update"):
		return "date"
	case strings.Contains(path, "regexp") || strings.Contains(path, "regex"):
		return "regexp"
	case strings.Contains(path, "json"):
		return "json"
	case strings.Contains(path, "node_") || strings.Contains(path, "process_") || strings.Contains(path, "realpath"):
		return "node-host"
	case strings.Contains(path, "math"):
		return "math"
	case strings.Contains(path, "number"):
		return "number"
	case strings.Contains(path, "array"):
		return "array"
	case strings.Contains(path, "collections") || (strings.Contains(path, "map") && !strings.Contains(path, "case_mapping")) || strings.Contains(path, "set_") || strings.HasSuffix(path, "/sets.a"):
		return "map-set"
	case strings.Contains(path, "string") || strings.Contains(path, "utf8") || strings.Contains(path, "case_mapping"):
		return "string"
	case strings.Contains(path, "library_object") || strings.Contains(path, "scout_22_tsc_keyword_entries"):
		return "object"
	}

	// Older oracle witnesses predate family-prefixed names (for example,
	// splitting and collection-lifetime probes). Include their explicit
	// built-in calls too, rather than treating the filename as the boundary.
	source, err := os.ReadFile(filepath.Join(repository, path))
	if err != nil {
		return ""
	}
	text := string(source)
	for _, family := range []struct {
		name  string
		calls []string
	}{
		{"regexp", []string{"new RegExp", ".match(", ".matchAll(", ".exec(", ".test("}},
		{"json", []string{"JSON."}},
		{"date", []string{"new Date", "Date."}},
		{"node-host", []string{"node:"}},
		{"math", []string{"Math."}},
		{"map-set", []string{"new Map", "new Set"}},
		{"number", []string{"Number.", "parseInt(", "parseFloat(", ".toFixed(", ".toPrecision(", ".toExponential("}},
		{"array", []string{"Array.", "new Array", ".push(", ".pop(", ".map(", ".filter(", ".reduce(", ".join(", ".splice("}},
		{"string", []string{"String.", ".split(", ".replace(", ".replaceAll(", ".trim(", ".repeat(", ".padStart(", ".padEnd(", ".normalize("}},
	} {
		for _, call := range family.calls {
			if strings.Contains(text, call) {
				return family.name
			}
		}
	}
	return ""
}

// This scout makes LeakSanitizer part of the native observation, rather than
// trusting cached leak evidence. It also refuses the generic oracle's numeric
// contraction exemption: the report must say when an accepted source differs.
func TestScout33LibraryLeaks(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Fatal("scout 33 requires Linux LeakSanitizer")
	}
	families := map[string]int{}
	for _, fixture := range fixtures {
		family := scout33Family(fixture.path)
		if family == "" {
			continue
		}
		families[family]++
		t.Run(family+"/"+fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if !fixture.lowers {
				if err == nil {
					t.Fatal("declared refusal now compiles")
				}
				t.Logf("compile-time refusal: %v", err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if fixture.checked {
				t.Log("intentional checked-input witness: Node runs on; compare the inserted check between backends")
			}
			generated := onJavaScriptBackend(t, program)
			binary := filepath.Join(t.TempDir(), "lsan")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if strings.Contains(string(actual.stderr), "LeakSanitizer") {
				t.Errorf("family %s leak; fixture %s:\n%s", family, fixture.path, actual.stderr)
			}
			for _, leg := range []struct {
				name   string
				result run
			}{{"native+LSan", actual}, {"JavaScript", generated}} {
				expected := truth
				if fixture.checked {
					expected = generated
				}
				if difference := disagreement(expected, leg.result); difference != "" {
					t.Errorf("%s %s; Node exit=%d stdout=%q stderr=%q; actual exit=%d stdout=%q stderr=%q", leg.name, difference, truth.exitCode, truth.stdout, truth.stderr, leg.result.exitCode, leg.result.stdout, leg.result.stderr)
				}
			}
			t.Logf("family=%s Node exit=%d; native detect_leaks=1 exit=%d", family, truth.exitCode, actual.exitCode)
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasm := filepath.Join(t.TempDir(), "program.wasm")
				err := native.Build(native.C(program), wasm, native.Options{Target: "wasm32-wasi"})
				var refused *native.TargetRefused
				if errors.As(err, &refused) {
					t.Logf("WASI compile-time refusal: %v", refused)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				wasmRun := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), wasm)
				expected := truth
				if fixture.checked {
					expected = generated
				}
				if difference := disagreement(expected, wasmRun); difference != "" {
					t.Errorf("WASI %s; Node exit=%d stdout=%q stderr=%q; WASI exit=%d stdout=%q stderr=%q", difference, truth.exitCode, truth.stdout, truth.stderr, wasmRun.exitCode, wasmRun.stdout, wasmRun.stderr)
				}
				t.Logf("WASI agrees, exit=%d", wasmRun.exitCode)
			}

		})
	}
	t.Logf("registered family fixture census: %v", families)
}

// Keep unsupported source excerpts intact: no library approximation stands in
// for a compiler type-proof or an unimplemented borrowed intrinsic.
func TestScout33Refusals(t *testing.T) {
	for _, probe := range []struct{ name, reason string }{
		{"json_roundtrip_refused", "JSON.parse: its result's type can't be proven"},
		{"json_object_reference_refused", "structural types can hide fields"},
		{"scout_22_tsc_own_keys", "unbound-method"},
		{"number_box_refused", "new an Identifier"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout33/"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("invalid Node witness: %+v", truth)
			}
			checked, err := load.Load([]string{path})
			if err == nil {
				_, err = lower.Lower(context.Background(), checked)
			}
			if err == nil {
				t.Fatal("expected the unmodified excerpt's compile-time gap")
			}
			if probe.reason != "" && !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("expected %q, got %v", probe.reason, err)
			}
			t.Logf("Node succeeds; compile-time refusal: %v", err)
		})
	}
}

// Not parallel: -update-counts writes the shared counts table.
// Refresh only this scout's rows; keep unrelated Linux measurements and the
// registry order. Full counts remain owned by TestCountsAreRecorded.
func TestScout33Counts(t *testing.T) {
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Split(line, " | ")
		if len(fields) == 7 && fields[0] != "| Fixture" {
			rows[strings.TrimPrefix(fields[0], "| ")] = line
		}
	}
	for _, name := range scout33Workloads {
		path := "internal/oracle/testdata/scout33/" + name + ".a"
		row := counted(t, path, false, nil, false, false)
		if !*updateCounts && rows[path] != row {
			t.Errorf("%s: recorded %q; measured %q", name, rows[path], row)
		}
		if *updateCounts {
			rows[path] = row
		}
		t.Log(row)
	}
	if !*updateCounts || t.Failed() {
		return
	}
	var out strings.Builder
	out.WriteString(countsHeader)
	write := func(path string) {
		if row, ok := rows[path]; ok {
			out.WriteString(row + "\n")
			delete(rows, path)
		}
	}
	for _, fixture := range fixtures {
		if fixture.lowers && !uncounted[fixture.path] {
			write(fixture.path)
		}
	}
	for _, fixture := range slowRegExpFixtures {
		write(fixture.path)
	}
	for _, fixture := range inputFixtures {
		write(fixture.path)
	}
	for _, name := range fsFileFixtures {
		write("internal/oracle/testdata/node_fs_file_" + name + ".a")
	}
	if len(rows) != 0 {
		t.Fatalf("unregistered count rows: %v", rows)
	}
	if err := os.WriteFile(countsPath, []byte(out.String()), 0644); err != nil {
		t.Fatal(err)
	}
}

// Change a family's real operation, preserving successful execution and
// ownership. Only the untouched source on Node knows the answer changed.
func TestScout33FamilyMutants(t *testing.T) {
	cases := []struct{ name, before, after, helper string }{
		{"collections", "adamic_map_delete(", "scout33_delete(", `static bool scout33_delete(adamic_map *map, adamic_value key) { (void)map; (void)key; return false; }`},
		{"array", "adamic_array_reverse(", "scout33_reverse(", `static adamic_array *scout33_reverse(adamic_array *array) { return array; }`},
		{"number", "adamic_number_to_fixed(", "adamic_number_to_fixed(1 + ", ""},
		{"math", "adamic_math_clz32(", "adamic_math_fround(", ""},
		{"date", "adamic_fs_file_date_time(", "scout33_date_time(", `static double scout33_date_time(const adamic_object *date) { return adamic_fs_file_date_time(date) + 1; }`},
		{"node_host", "adamic_node_path_basename(", "scout33_basename(", `static adamic_string *scout33_basename(const adamic_string *path, const adamic_string *suffix) { (void)suffix; return adamic_node_path_dirname(path); }`},
		{"regexp", `ADAMIC_STRING("im")`, `ADAMIC_STRING("m")`, ""},
		{"string", "adamic_string_replace(", "scout33_replace(", `static adamic_string *scout33_replace(const adamic_string *value, const adamic_string *search, const adamic_string *replacement, bool all) { (void)all; return adamic_string_replace(value, search, replacement, false); }`},
		{"json", "", "", ""},
		{"scout_22_tsc_keyword_entries", "", "", ""},
	}
	for _, one := range cases {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout33/"+one.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if one.name == "json" {
				changed := false
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
					if call, ok := value.(ir.JSONStringify); ok && !changed && len(call.Schema.Fields) > 1 {
						call.Schema.Fields[0], call.Schema.Fields[1] = call.Schema.Fields[1], call.Schema.Fields[0]
						changed = true
					}
					return value
				})
				if !changed {
					t.Fatal("JSON schema mutation site absent")
				}
			}

			if one.name == "scout_22_tsc_keyword_entries" {
				changed := false
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
					if call, ok := value.(ir.ObjectCall); ok && call.Method == "entries" && !changed {
						changed = true
						return ir.ArrayReverse{Array: call}
					}
					return value
				})
				if !changed {
					t.Fatal("Object.entries mutation site absent")
				}
			}
			source := native.C(program)
			if one.name != "json" && one.name != "scout_22_tsc_keyword_entries" {
				changed := strings.ReplaceAll(source, one.before, one.after)
				if changed == source {
					t.Fatalf("mutation site absent: %s", one.before)
				}
				source = changed
			}
			if one.helper != "" {
				source = "#include \"adamic.h\"\n" + one.helper + "\n" + source
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant failed outside Node comparison: exit=%d stderr=%q", actual.exitCode, actual.stderr)
			}
			if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
				t.Fatalf("mutant not caught only by Node stdout: %s", difference)
			}
			t.Log("Node stdout catches mutation; exit 0 and ASan/UBSan/LSan clean")
		})
	}
}

// A real heap result, not an immortal literal, must make this scout go red
// when one owned reference is lost. Node bytes still agree; only LSan catches it.
func TestScout33LeakDetectorMutant(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smallest.a")
	if err := os.WriteFile(path, []byte("const value = 'path'.repeat(2); console.log(value.replaceAll('path', 'name'));"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	changed := strings.ReplaceAll(source, "adamic_string_replace(", "scout33_leaking_replace(")
	if changed == source {
		t.Fatal("replace mutation site absent")
	}
	helper := `#include "adamic.h"
static adamic_string *scout33_leaking_replace(const adamic_string *value, const adamic_string *search, const adamic_string *replacement, bool all) {
 adamic_string *result = adamic_string_replace(value, search, replacement, all);
 adamic_retain(result);
 return result;
}
`
	binary := filepath.Join(t.TempDir(), "leak-mutant")
	if err := native.Build(helper+changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:fast_unwind_on_malloc=0", "LSAN_OPTIONS=exitcode=23"}, binary)
	if actual.exitCode != 23 || !strings.Contains(string(actual.stderr), "LeakSanitizer") || !strings.Contains(string(actual.stderr), "Direct leak") {
		t.Fatalf("expected only a LeakSanitizer failure: exit=%d stderr=%s", actual.exitCode, actual.stderr)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(actual.stdout) != string(truth.stdout) {
		t.Fatalf("leak mutation changed Node stdout: %q vs %q", actual.stdout, truth.stdout)
	}
	t.Logf("extra retain caught only by LeakSanitizer; Node bytes agree:\n%s", actual.stderr)
}

// Not parallel: reuse the dedicated backtracking oracle's generous CPU budget;
// its Linux leak leg explicitly sets detect_leaks=1.
func TestScout33LongRegExp(t *testing.T) {
	TestRegExpLongBacktrackNode(t)
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		return
	}
	path, err := filepath.Abs(filepath.Join(repository, slowRegExpFixtures[0].path))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "backtrack.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	actual := longRegExpRun(t, nil, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	if difference := disagreement(onNode(t, path), actual); difference != "" {
		t.Fatalf("WASI backtracking: %s, stderr=%q", difference, actual.stderr)
	}
}
