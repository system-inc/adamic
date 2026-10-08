package oracle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func runtimeConstructorMutant(t *testing.T, source string) run {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside Node comparison: %+v", actual)
	}
	return actual
}

func TestRuntimeConstructorFamilyMutants(t *testing.T) {
	for _, row := range []struct{ name, file, old, new string }{
		{"enums", "regexp_compile_bytecode.c", "base.assertion=node->kind;", "base.assertion=2;"},
		{"option_errors", "regexp_compile_bytecode.c", "base.range_count=set.range_count;", "base.range_count=0;"},
		{"schema", "regexp_compile_bytecode.c", "base.range_count=set.range_count;", "base.range_count=0;"},
		{"error_messages", "regexp_compile_bytecode.c", "base.range_count=set.range_count;", "base.range_count=0;"},
		{"mobile_detect", "regexp_compile_parser.c", `case 'i': flag = REGEX_FLAG_I; break;`, `case 'i': flag = 0; break;`},
		{"emoji", "regexp_compile_bytecode.c", "base.range_count=set.range_count;", "base.range_count=0;"},
	} {
		t.Run(row.name, func(t *testing.T) {
			path, _ := filepath.Abs(filepath.Join(repository, runtimeConstructorsDirectory, row.name+".a"))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", row.file))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Count(runtime, []byte(row.old)) != 1 {
				t.Fatal("mutation site moved")
			}
			source := native.C(program)
			changed := strings.Replace(source, `#include "`+row.file+`"`, strings.Replace(string(runtime), row.old, row.new, 1), 1)
			if changed == source {
				t.Fatal("runtime compiler was not linked")
			}
			expected := onNode(t, path)
			actual := runtimeConstructorMutant(t, changed)
			if d := disagreement(expected, actual); d != "stdout differs" {
				t.Fatalf("mutant survived Node or failed another check: %s", d)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasi := onWASI(t, changed)
				if wasi.exitCode != 0 || len(wasi.stderr) != 0 || disagreement(expected, wasi) != "stdout differs" {
					t.Fatalf("WASI mutant not caught only by Node: %+v", wasi)
				}
			}
			t.Logf("%s: clean sanitizer exit, caught only by Node stdout", row.old)
		})
	}
}

func TestRuntimeConstructorFailureMutants(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, runtimeConstructorsDirectory, "construction_failure.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_compile_runtime.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	for _, row := range []struct{ name, pattern, flags, old, new string }{
		{"Node syntax mislabeled Error", "[", "u", `memcpy((char *)name->bytes,"SyntaxError",11);`, `memcpy((char *)name->bytes,"OtherError!",11);`},
		{"V8 refusal made catchable", "[\\q{a}]", "iv", `if(result.status!=1){`, `if(result.status!=1&&result.status!=3){`},
		{"missing feature made catchable", "a{18446744073709551616}", "", `if(result.status!=1){`, `if(result.status!=1&&result.status!=4){`},
	} {
		t.Run(row.name, func(t *testing.T) {
			if bytes.Count(runtime, []byte(row.old)) != 1 {
				t.Fatal("mutation site moved")
			}
			changed := strings.Replace(source, `#include "regexp_compile_runtime.c"`, strings.Replace(string(runtime), row.old, row.new, 1), 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, row.pattern, row.flags)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant failed outside Node: %+v", actual)
			}
			expected := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path, row.pattern, row.flags)
			if d := disagreement(expected, actual); d != "stdout differs" {
				t.Fatalf("mutant not caught only by Node: %s", d)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasm := filepath.Join(t.TempDir(), "mutant.wasm")
				buildWASI(t, changed, wasm)
				wasi := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), wasm, row.pattern, row.flags)
				if wasi.exitCode != 0 || len(wasi.stderr) != 0 || disagreement(expected, wasi) != "stdout differs" {
					t.Fatalf("WASI failure mutant survived Node: %+v", wasi)
				}
			}
			t.Log("clean sanitizer exit, caught only by Node stdout")
		})
	}
}

func TestRuntimeConstructorBudgetStop(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, runtimeConstructorsDirectory, "budget.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if string(expected.stdout) != "true\nfinished\n" {
		t.Fatalf("Node witness changed: %+v", expected)
	}
	source := native.C(program)
	old := "adamic_start(argc, argv);"
	if strings.Count(source, old) != 1 {
		t.Fatal("main initialization moved")
	}
	source = strings.Replace(source, old, old+"adamic_regex_set_step_limit(1);", 1)
	binary := filepath.Join(t.TempDir(), "stop")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if actual.exitCode != 70 || len(actual.stdout) != 0 || !bytes.Contains(actual.stderr, []byte("regexp: instruction step limit exceeded /^a{12}$/")) {
		t.Fatalf("budget was catchable, unnamed, or returned an answer: %+v", actual)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if d := disagreement(actual, onWASI(t, source)); d != "" {
			t.Fatal("WASI: " + d)
		}
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp.c"))
	if err != nil {
		t.Fatal(err)
	}
	mutation := `regex_budget_stop(workspace,"regexp: instruction step limit exceeded");`
	if bytes.Count(runtime, []byte(mutation)) != 1 {
		t.Fatal("budget mutation site moved")
	}
	// Returning 'no match' builds and exits cleanly; only Node's true answer
	// catches it. The unreachable call keeps diagnostic code referenced.
	changed := strings.Replace(string(runtime), mutation, `if (*steps >= limit) goto finished; `+mutation, 1) + "\n" + source
	// Keep the archive's matcher symbols independent; its regular-engine
	// helpers can still link without colliding with this instrumented copy.
	for _, name := range []string{"set_regular_enabled", "set_regular_mode", "set_step_limit", "read", "canonical", "word", "contains", "new_owned", "new", "done", "group_lookup", "property", "exec", "test", "match_all", "match", "iterator_step", "next", "search", "split", "replace"} {
		changed = strings.ReplaceAll(changed, "adamic_regex_"+name+"(", "adamic_regex_mutant_"+name+"(")
	}
	mutant := runtimeConstructorMutant(t, changed)
	if d := disagreement(expected, mutant); d != "stdout differs" {
		t.Fatalf("budget mutant survived Node: %s", d)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		wasi := onWASI(t, changed)
		if wasi.exitCode != 0 || len(wasi.stderr) != 0 || disagreement(expected, wasi) != "stdout differs" {
			t.Fatalf("WASI budget mutant survived Node: %+v", wasi)
		}
	}
	t.Log("budget returning false: clean sanitizer exit, caught only by Node stdout")
}

// Allocation/set/emission failures share the native-only stop, even if the
// parser cannot allocate a Node message. Fault injection makes each reachable
// without asking the machine to exhaust memory or allocate INT_MAX opcodes.
func TestRuntimeConstructorInternalFailurePaths(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, runtimeConstructorsDirectory, "construction_failure.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_compile_runtime.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	site := `adamic_regex_program *program=result.status==0?adamic_regex_compile_checked(&result):NULL;`
	if bytes.Count(runtime, []byte(site)) != 1 {
		t.Fatal("compilation site moved")
	}
	for _, row := range []struct {
		name   string
		status int
		reason string
	}{
		{"allocation", 2, "RegExp runtime compiler: out of memory"},
		{"set compiler", 4, "unsupported class syntax"},
		{"emission", 4, "native regexp instruction count exceeds int"},
	} {
		t.Run(row.name, func(t *testing.T) {
			injection := fmt.Sprintf("result.status=%d;result.reference_reason=%q;", row.status, row.reason)
			body := strings.Replace(string(runtime), site, injection+site, 1)
			changed := strings.Replace(source, `#include "regexp_compile_runtime.c"`, body, 1)
			binary := filepath.Join(t.TempDir(), "stop")
			if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, "abc", "i")
			if actual.exitCode != 70 || len(actual.stdout) != 0 || !bytes.Contains(actual.stderr, []byte("for /abc/i: "+row.reason)) {
				t.Fatalf("internal failure entered catch: %+v", actual)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasm := filepath.Join(t.TempDir(), "stop.wasm")
				buildWASI(t, changed, wasm)
				wasi := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), wasm, "abc", "i")
				if d := disagreement(actual, wasi); d != "" {
					t.Fatal("WASI internal stop: " + d)
				}
			}
			mutant := strings.Replace(body, `if(result.status!=1){`, `if(result.status!=1&&result.status!=`+fmt.Sprint(row.status)+`){`, 1)
			changed = strings.Replace(source, `#include "regexp_compile_runtime.c"`, mutant, 1)
			binary = filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, "abc", "i")
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant failed outside Node: %+v", actual)
			}
			expected := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path, "abc", "i")
			if d := disagreement(expected, actual); d != "stdout differs" {
				t.Fatalf("internal failure mutant survived Node: %s", d)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasm := filepath.Join(t.TempDir(), "mutant.wasm")
				buildWASI(t, changed, wasm)
				wasi := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), wasm, "abc", "i")
				if wasi.exitCode != 0 || len(wasi.stderr) != 0 || disagreement(expected, wasi) != "stdout differs" {
					t.Fatalf("WASI internal failure mutant survived Node: %+v", wasi)
				}
			}
			t.Log("native-only failure made catchable: clean sanitizer exit, caught only by Node stdout")
		})
	}
}
