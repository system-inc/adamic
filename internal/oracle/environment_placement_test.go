package oracle

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

var environmentFixtures = []string{"captured_parameter_escape", "captured_parameter_reassign", "array", "async", "callback_escape", "callbacks", "capture", "captured_parameters", "coverage_alias_join", "coverage_boundary64", "coverage_bounded_call", "coverage_callback_exits", "coverage_callback_result", "coverage_callbacks_extra", "coverage_captured_parameter", "coverage_closure_self", "coverage_collections", "coverage_initializer_throw", "coverage_local_containers", "coverage_primitives", "coverage_region_values", "coverage_returned_values", "coverage_values_exits", "direct", "exits", "field", "global", "keeping_call", "large", "local_call", "loop", "map", "returned", "set", "siblings", "unknown_call", "unknown_callback", "virtual_call"}

func init() {
	for _, name := range environmentFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/environment_" + name + ".a", true, false})
	}
}
func environmentAgrees(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if diff := disagreement(truth, onJavaScriptBackend(t, program)); diff != "" {
		t.Fatal("JavaScript backend: " + diff)
	}
	actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" {
		t.Fatal(leaked)
	}
	if !comparable {
		t.Fatal("sanitized result unavailable")
	}
	if diff := disagreement(truth, actual); diff != "" {
		t.Fatalf("native: %s %s", diff, actual.stderr)
	}
	if diff := disagreement(truth, released(t, program)); diff != "" {
		t.Fatal("release: " + diff)
	}
	code := native.C(program)
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(code, binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	command, args := pinnedStack(binary)
	counted := execute(t, command, args...)
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		t.Fatal(report)
	}
	counted.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if diff := disagreement(truth, counted); diff != "" {
		t.Fatalf("counted: %s %s", diff, counted.stderr)
	}
}
func TestEnvironmentArrayAgreesWithNode(t *testing.T) { t.Parallel(); environmentAgrees(t, "array") }
func TestEnvironmentAsyncAgreesWithNode(t *testing.T) { t.Parallel(); environmentAgrees(t, "async") }
func TestEnvironmentCallbackEscapeAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "callback_escape")
}
func TestEnvironmentCallbacksAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "callbacks")
}
func TestEnvironmentCaptureAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "capture")
}
func TestEnvironmentCapturedParametersAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "captured_parameters")
}
func TestEnvironmentCoverageAliasJoinAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_alias_join")
}
func TestEnvironmentCoverageBoundary64AgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_boundary64")
}
func TestEnvironmentCoverageBoundedCallAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_bounded_call")
}
func TestEnvironmentCoverageCallbackExitsAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_callback_exits")
}
func TestEnvironmentCoverageCallbackResultAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_callback_result")
}
func TestEnvironmentCoverageCallbacksExtraAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_callbacks_extra")
}
func TestEnvironmentCoverageCapturedParameterAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_captured_parameter")
}
func TestEnvironmentCoverageClosureSelfAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_closure_self")
}
func TestEnvironmentCoverageCollectionsAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_collections")
}
func TestEnvironmentCoverageInitializerThrowAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_initializer_throw")
}
func TestEnvironmentCoverageLocalContainersAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_local_containers")
}
func TestEnvironmentCoveragePrimitivesAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_primitives")
}
func TestEnvironmentCoverageRegionValuesAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_region_values")
}
func TestEnvironmentCoverageReturnedValuesAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_returned_values")
}
func TestEnvironmentCoverageValuesExitsAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "coverage_values_exits")
}
func TestEnvironmentDirectAgreesWithNode(t *testing.T) { t.Parallel(); environmentAgrees(t, "direct") }
func TestEnvironmentExitsAgreesWithNode(t *testing.T)  { t.Parallel(); environmentAgrees(t, "exits") }
func TestEnvironmentFieldAgreesWithNode(t *testing.T)  { t.Parallel(); environmentAgrees(t, "field") }
func TestEnvironmentGlobalAgreesWithNode(t *testing.T) { t.Parallel(); environmentAgrees(t, "global") }
func TestEnvironmentKeepingCallAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "keeping_call")
}
func TestEnvironmentLargeAgreesWithNode(t *testing.T) { t.Parallel(); environmentAgrees(t, "large") }
func TestEnvironmentLocalCallAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "local_call")
}
func TestEnvironmentLoopAgreesWithNode(t *testing.T) { t.Parallel(); environmentAgrees(t, "loop") }
func TestEnvironmentMapAgreesWithNode(t *testing.T)  { t.Parallel(); environmentAgrees(t, "map") }
func TestEnvironmentReturnedAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "returned")
}
func TestEnvironmentSetAgreesWithNode(t *testing.T) { t.Parallel(); environmentAgrees(t, "set") }
func TestEnvironmentSiblingsAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "siblings")
}
func TestEnvironmentUnknownCallAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "unknown_call")
}
func TestEnvironmentUnknownCallbackAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "unknown_callback")
}
func TestEnvironmentVirtualCallAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "virtual_call")
}

func TestEnvironmentEscapeStackMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_returned.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	allocation := regexp.MustCompile(`adamic_environment \*(adamic_temporary_[0-9]+) = adamic_environment_new\(1\);`)
	match := allocation.FindStringSubmatch(source)
	if len(match) != 2 {
		t.Fatal("escaping fixture lost its heap allocation")
	}
	environment := match[1]
	source = strings.Replace(source, match[0], "struct { adamic_environment record; adamic_cell cells[1]; } mutant_storage;\n adamic_environment *"+environment+" = &mutant_storage.record;\n adamic_environment_init("+environment+", mutant_storage.cells, 1);", 1)
	source = strings.ReplaceAll(source, "adamic_release("+environment+");", "adamic_environment_end("+environment+");")
	// Prevent inlining from converting the intended use after return into a
	// use after scope. This changes no allocation or output in the healthy build.
	source = strings.ReplaceAll(source, "static adamic_closure * adamic_function_0_make", "static __attribute__((noinline)) adamic_closure * adamic_function_0_make")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:detect_stack_use_after_return=1"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: stack-use-after-return") {
		t.Fatalf("want ASan use after return, got exit %d:\n%s", result.exitCode, result.stderr)
	}
	t.Logf("escaping environment mutant caught by ASan:\n%s", result.stderr)
}

func TestEnvironmentUnknownCallStackMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_unknown_call.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	allocation := regexp.MustCompile(`adamic_environment \*(adamic_temporary_[0-9]+) = adamic_environment_new\(1\);`)
	match := allocation.FindStringSubmatch(source)
	if len(match) != 2 {
		t.Fatal("escaping fixture lost its heap allocation")
	}
	environment := match[1]
	source = strings.Replace(source, match[0], "struct { adamic_environment record; adamic_cell cells[1]; } mutant_storage;\n adamic_environment *"+environment+" = &mutant_storage.record;\n adamic_environment_init("+environment+", mutant_storage.cells, 1);", 1)
	source = strings.ReplaceAll(source, "adamic_release("+environment+");", "adamic_environment_end("+environment+");")
	// Prevent inlining from converting the intended use after return into a
	// use after scope. This changes no allocation or output in the healthy build.
	source = strings.ReplaceAll(source, "static void adamic_function_0_fill", "static __attribute__((noinline)) void adamic_function_0_fill")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:detect_stack_use_after_return=1"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: stack-use-after-return") {
		t.Fatalf("want ASan use after return, got exit %d:\n%s", result.exitCode, result.stderr)
	}
	t.Logf("unknown-callee environment mutant caught by ASan:\n%s", result.stderr)
}

func TestEnvironmentThrowReleaseMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_exits.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	// Only the exceptional exits lose destruction. Ordinary returns keep it.
	cleanup := regexp.MustCompile(`adamic_environment_end\(adamic_temporary_[0-9]+\);(\s+return &adamic_string_empty;)`)
	if !cleanup.MatchString(source) {
		t.Fatal("throw fixture has no frame cleanup")
	}
	source = cleanup.ReplaceAllString(source, "$1")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if difference := disagreement(onNode(t, path), result); difference != "" {
		t.Fatalf("mutant must preserve output and pass sanitizers: %s", difference)
	}
	report := ""
	if runtime.GOOS == "darwin" {
		// The macOS tool needs an unsanitized build of the mutated C itself.
		// Re-emitting the original IR would silently remove this mutation.
		release := filepath.Join(t.TempDir(), "mutant-release")
		if err := native.Build(source, release, native.Options{}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, "leaks", "--atExit", "--", release)
		if result.exitCode != 0 {
			report = string(result.stdout)
		}
	} else {
		report = leaksUncached(t, program, binary)
	}
	if report == "" || (runtime.GOOS == "linux" && !strings.Contains(report, "LeakSanitizer")) {
		t.Fatalf("want leaked reference slot, got %s", report)
	}
	t.Logf("throw-path slot mutant caught only by leak check:\n%s", report)
}

// Reintroduce the old borrowing decision on the actual reader's program. This
// must reach the owned-cell store assertion, rather than clang or the runtime.
func TestEnvironmentCapturedParameterBorrowMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_captured_parameters.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	captured := 0
	for _, function := range program.Functions {
		for _, parameter := range function.Parameters {
			local := &program.Locals[parameter]
			if local.Captured && local.Type.IsReference() {
				if local.Borrowed {
					t.Fatalf("captured parameter %s is borrowed", local.Name)
				}
				captured++
			}
		}
	}
	if captured != 3 {
		t.Fatalf("want three captured reference parameters, got %d", captured)
	}
	// The healthy program must emit before we mutate its ownership metadata.
	native.C(program)
	for _, function := range program.Functions {
		for _, parameter := range function.Parameters {
			local := &program.Locals[parameter]
			if local.Captured && local.Type.IsReference() {
				local.Borrowed = true
			}
		}
	}
	defer func() {
		failure := recover()
		if failure != "native: a store into the borrowed parameter text" {
			t.Fatalf("want the original borrowed-cell panic, got %v", failure)
		}
		t.Logf("borrow mutant caught by owned-cell store assertion: %v", failure)
	}()
	native.C(program)
}

func TestEnvironmentCapturedParameterEscapeAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "captured_parameter_escape")
}
func TestEnvironmentCapturedParameterReassignAgreesWithNode(t *testing.T) {
	t.Parallel()
	environmentAgrees(t, "captured_parameter_reassign")
}

func TestEnvironmentCapturedCellOwnershipMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_captured_parameter_escape.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	retain := regexp.MustCompile(`(->value.reference = )adamic_retain\((adamic_local_[0-9]+_text)\);`)
	if len(retain.FindAllString(code, -1)) != 1 {
		t.Fatal("captured input must have exactly one cell retain")
	}
	code = retain.ReplaceAllString(code, "${1}${2};")
	binary := filepath.Join(t.TempDir(), "missing-cell-owner")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
		t.Fatalf("missing cell owner escaped ASan: %d %s", result.exitCode, result.stderr)
	}
	t.Log("missing initial cell count caught by ASan after the argument's statement")
}
func TestEnvironmentRegionThrowReleaseMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_coverage_region_values.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	cleanup := regexp.MustCompile(`adamic_region_end\(&adamic_temporary_[0-9]+\);(\s+return &adamic_string_empty;)`)
	if !cleanup.MatchString(code) {
		t.Fatal("region fixture has no exceptional region cleanup")
	}
	code = cleanup.ReplaceAllString(code, "$1")
	binary := filepath.Join(t.TempDir(), "missing-region-cleanup")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if diff := disagreement(onNode(t, path), result); diff != "" {
		t.Fatal("mutant must preserve output and sanitizer behavior: " + diff)
	}
	counted := filepath.Join(t.TempDir(), "missing-region-counted")
	if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	command, arguments := pinnedStack(counted)
	balance := execute(t, command, arguments...)
	if report := leakcheck.Unbalanced(leakRun(balance)); report == "" {
		t.Fatalf("missing exceptional region cleanup escaped allocation balance: %s", balance.stderr)
	} else {
		t.Log(report)
	}
	t.Log("missing exceptional call-region cleanup preserves Node and ASan output but fails the independent live-value balance")
}
