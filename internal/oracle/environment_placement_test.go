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

var environmentFixtures = []string{
	"captured_parameter_escape",
	"captured_parameter_reassign",
	"array",
	"async",
	"callback_escape",
	"callbacks",
	"capture",
	"captured_parameters",
	"coverage_alias_join",
	"coverage_boundary64",
	"coverage_bounded_call",
	"coverage_callback_exits",
	"coverage_callback_result",
	"coverage_callbacks_extra",
	"coverage_captured_parameter",
	"coverage_closure_self",
	"coverage_collections",
	"coverage_initializer_throw",
	"coverage_local_containers",
	"coverage_primitives",
	"coverage_region_values",
	"coverage_returned_values",
	"coverage_values_exits",
	"direct",
	"exits",
	"field",
	"global",
	"keeping_call",
	"large",
	"local_call",
	"loop",
	"map",
	"returned",
	"set",
	"siblings",
	"unknown_call",
	"unknown_callback",
	"virtual_call",
}

func init() {
	for _, name := range environmentFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/environment_" + name + ".a", true, false})
	}
}
func environmentAgrees(test *testing.T, name string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_"+name+".a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	if diff := disagreement(truth, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatal("JavaScript backend: " + diff)
	}
	actual, leaked, comparable := nativelyCheckingLeaks(test, program, true)
	if leaked != "" {
		test.Fatal(leaked)
	}
	if !comparable {
		test.Fatal("sanitized result unavailable")
	}
	if diff := disagreement(truth, actual); diff != "" {
		test.Fatalf("native: %s %s", diff, actual.stderr)
	}
	if diff := disagreement(truth, released(test, program)); diff != "" {
		test.Fatal("release: " + diff)
	}
	code := native.C(program)
	binary := filepath.Join(test.TempDir(), "counted")
	if err := native.Build(code, binary, native.Options{Count: true}); err != nil {
		test.Fatal(err)
	}
	command, args := pinnedStack(binary)
	counted := execute(test, command, args...)
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		test.Fatal(report)
	}
	counted.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if diff := disagreement(truth, counted); diff != "" {
		test.Fatalf("counted: %s %s", diff, counted.stderr)
	}
}
func TestEnvironmentArrayAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "array")
}
func TestEnvironmentAsyncAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "async")
}
func TestEnvironmentCallbackEscapeAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "callback_escape")
}
func TestEnvironmentCallbacksAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "callbacks")
}
func TestEnvironmentCaptureAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "capture")
}
func TestEnvironmentCapturedParametersAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "captured_parameters")
}
func TestEnvironmentCoverageAliasJoinAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_alias_join")
}
func TestEnvironmentCoverageBoundary64AgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_boundary64")
}
func TestEnvironmentCoverageBoundedCallAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_bounded_call")
}
func TestEnvironmentCoverageCallbackExitsAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_callback_exits")
}
func TestEnvironmentCoverageCallbackResultAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_callback_result")
}
func TestEnvironmentCoverageCallbacksExtraAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_callbacks_extra")
}
func TestEnvironmentCoverageCapturedParameterAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_captured_parameter")
}
func TestEnvironmentCoverageClosureSelfAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_closure_self")
}
func TestEnvironmentCoverageCollectionsAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_collections")
}
func TestEnvironmentCoverageInitializerThrowAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_initializer_throw")
}
func TestEnvironmentCoverageLocalContainersAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_local_containers")
}
func TestEnvironmentCoveragePrimitivesAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_primitives")
}
func TestEnvironmentCoverageRegionValuesAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_region_values")
}
func TestEnvironmentCoverageReturnedValuesAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_returned_values")
}
func TestEnvironmentCoverageValuesExitsAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "coverage_values_exits")
}
func TestEnvironmentDirectAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "direct")
}
func TestEnvironmentExitsAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "exits")
}
func TestEnvironmentFieldAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "field")
}
func TestEnvironmentGlobalAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "global")
}
func TestEnvironmentKeepingCallAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "keeping_call")
}
func TestEnvironmentLargeAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "large")
}
func TestEnvironmentLocalCallAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "local_call")
}
func TestEnvironmentLoopAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "loop")
}
func TestEnvironmentMapAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "map")
}
func TestEnvironmentReturnedAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "returned")
}
func TestEnvironmentSetAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "set")
}
func TestEnvironmentSiblingsAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "siblings")
}
func TestEnvironmentUnknownCallAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "unknown_call")
}
func TestEnvironmentUnknownCallbackAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "unknown_callback")
}
func TestEnvironmentVirtualCallAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "virtual_call")
}

func TestEnvironmentEscapeStackMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_returned.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	source := native.C(program)
	allocation := regexp.MustCompile(
		`adamic_environment \*(adamic_temporary_[A-Za-z0-9_]+) = adamic_environment_new\(1\);`,
	)
	match := allocation.FindStringSubmatch(source)
	if len(match) != 2 {
		test.Fatal("escaping fixture lost its heap allocation")
	}
	environment := match[1]
	source = strings.Replace(
		source,
		match[0],
		"struct { adamic_environment record; adamic_cell cells[1]; } mutant_storage;\n adamic_environment *"+
			environment+
			" = &mutant_storage.record;\n adamic_environment_init("+
			environment+
			", mutant_storage.cells, 1);",
		1,
	)
	source = strings.ReplaceAll(source, "adamic_release("+environment+");", "adamic_environment_end("+environment+");")
	// Prevent inlining from converting the intended use after return into a
	// use after scope. This changes no allocation or output in the healthy build.
	source = strings.ReplaceAll(
		source,
		"static adamic_closure * "+emittedFunctionSymbol(test, source, "make"),
		"static __attribute__((noinline)) adamic_closure * "+emittedFunctionSymbol(test, source, "make"),
	)
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0:detect_stack_use_after_return=1"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: stack-use-after-return") {
		test.Fatalf("want ASan use after return, got exit %d:\n%s", result.exitCode, result.stderr)
	}
	test.Logf("escaping environment mutant caught by ASan:\n%s", result.stderr)
}

func TestEnvironmentUnknownCallStackMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_unknown_call.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	source := native.C(program)
	allocation := regexp.MustCompile(
		`adamic_environment \*(adamic_temporary_[A-Za-z0-9_]+) = adamic_environment_new\(1\);`,
	)
	match := allocation.FindStringSubmatch(source)
	if len(match) != 2 {
		test.Fatal("escaping fixture lost its heap allocation")
	}
	environment := match[1]
	source = strings.Replace(
		source,
		match[0],
		"struct { adamic_environment record; adamic_cell cells[1]; } mutant_storage;\n adamic_environment *"+
			environment+
			" = &mutant_storage.record;\n adamic_environment_init("+
			environment+
			", mutant_storage.cells, 1);",
		1,
	)
	source = strings.ReplaceAll(source, "adamic_release("+environment+");", "adamic_environment_end("+environment+");")
	// Prevent inlining from converting the intended use after return into a
	// use after scope. This changes no allocation or output in the healthy build.
	source = strings.ReplaceAll(
		source,
		"static void "+emittedFunctionSymbol(test, source, "fill"),
		"static __attribute__((noinline)) void "+emittedFunctionSymbol(test, source, "fill"),
	)
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0:detect_stack_use_after_return=1"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: stack-use-after-return") {
		test.Fatalf("want ASan use after return, got exit %d:\n%s", result.exitCode, result.stderr)
	}
	test.Logf("unknown-callee environment mutant caught by ASan:\n%s", result.stderr)
}

func TestEnvironmentThrowReleaseMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_exits.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	source := native.C(program)
	// Only the exceptional exits lose destruction. Ordinary returns keep it.
	cleanup := regexp.MustCompile(
		`adamic_environment_end\(adamic_temporary_[A-Za-z0-9_]+\);(\s+return &adamic_string_empty;)`,
	)
	if !cleanup.MatchString(source) {
		test.Fatal("throw fixture has no frame cleanup")
	}
	source = cleanup.ReplaceAllString(source, "$1")
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if difference := disagreement(onNode(test, path), result); difference != "" {
		test.Fatalf("mutant must preserve output and pass sanitizers: %s", difference)
	}
	report := ""
	if runtime.GOOS == "darwin" {
		// The macOS tool needs an unsanitized build of the mutated C itself.
		// Re-emitting the original IR would silently remove this mutation.
		release := filepath.Join(test.TempDir(), "mutant-release")
		if err := native.Build(source, release, native.Options{}); err != nil {
			test.Fatal(err)
		}
		result := execute(test, "leaks", "--atExit", "--", release)
		if result.exitCode != 0 {
			report = string(result.stdout)
		}
	} else {
		report = leaksUncached(test, program, binary)
	}
	if report == "" || (runtime.GOOS == "linux" && !strings.Contains(report, "LeakSanitizer")) {
		test.Fatalf("want leaked reference slot, got %s", report)
	}
	test.Logf("throw-path slot mutant caught only by leak check:\n%s", report)
}

// Reintroduce the old borrowing decision on the actual reader's program. This
// must reach the owned-cell store assertion, rather than clang or the runtime.
func TestEnvironmentCapturedParameterBorrowMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_captured_parameters.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	captured := 0
	for _, function := range program.Functions {
		for _, parameter := range function.Parameters {
			local := &program.Locals[parameter]
			if local.Captured && local.Type.IsReference() {
				if local.Borrowed {
					test.Fatalf("captured parameter %s is borrowed", local.Name)
				}
				captured++
			}
		}
	}
	if captured != 3 {
		test.Fatalf("want three captured reference parameters, got %d", captured)
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
			test.Fatalf("want the original borrowed-cell panic, got %v", failure)
		}
		test.Logf("borrow mutant caught by owned-cell store assertion: %v", failure)
	}()
	native.C(program)
}

func TestEnvironmentCapturedParameterEscapeAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "captured_parameter_escape")
}
func TestEnvironmentCapturedParameterReassignAgreesWithNode(test *testing.T) {
	test.Parallel()
	environmentAgrees(test, "captured_parameter_reassign")
}

func TestEnvironmentCapturedCellOwnershipMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/environment_captured_parameter_escape.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	code := native.C(program)
	retain := regexp.MustCompile(`(->value.reference = )adamic_retain\((adamic_local_[0-9]+_text)\);`)
	if len(retain.FindAllString(code, -1)) != 1 {
		test.Fatal("captured input must have exactly one cell retain")
	}
	code = retain.ReplaceAllString(code, "${1}${2};")
	binary := filepath.Join(test.TempDir(), "missing-cell-owner")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
		test.Fatalf("missing cell owner escaped ASan: %d %s", result.exitCode, result.stderr)
	}
	test.Log("missing initial cell count caught by ASan after the argument's statement")
}
func TestEnvironmentRegionThrowReleaseMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/environment_coverage_region_values.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	code := native.C(program)
	cleanup := regexp.MustCompile(
		`adamic_region_end\(&adamic_temporary_[A-Za-z0-9_]+\);(\s+return &adamic_string_empty;)`,
	)
	if !cleanup.MatchString(code) {
		test.Fatal("region fixture has no exceptional region cleanup")
	}
	code = cleanup.ReplaceAllString(code, "$1")
	binary := filepath.Join(test.TempDir(), "missing-region-cleanup")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if diff := disagreement(onNode(test, path), result); diff != "" {
		test.Fatal("mutant must preserve output and sanitizer behavior: " + diff)
	}
	counted := filepath.Join(test.TempDir(), "missing-region-counted")
	if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
		test.Fatal(err)
	}
	command, arguments := pinnedStack(counted)
	balance := execute(test, command, arguments...)
	if report := leakcheck.Unbalanced(leakRun(balance)); report == "" {
		test.Fatalf("missing exceptional region cleanup escaped allocation balance: %s", balance.stderr)
	} else {
		test.Log(report)
	}
	test.Log(
		"missing exceptional call-region cleanup preserves Node and ASan output but fails the independent live-value balance",
	)
}
