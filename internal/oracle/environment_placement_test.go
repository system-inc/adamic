package oracle

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"direct", "callbacks", "loop", "siblings", "returned", "field", "array", "map", "set", "global", "capture", "keeping_call", "unknown_call", "local_call", "callback_escape", "virtual_call", "unknown_callback", "exits", "large", "captured_parameters", "captured_reassigned", "captured_early_return", "async"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/environment_" + name + ".a", true, false})
	}
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

// Removing the cell's retain instead of the redundant parameter owner leaves
// the caller holding a reference that cell destruction has already released.
func TestCapturedParameterCellRetainMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/environment_captured_early_return.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	initialization := regexp.MustCompile(`(adamic_local_[0-9]+_text_cell->value.reference = )adamic_retain\((adamic_local_[0-9]+_text)\);`)
	if matches := initialization.FindAllString(source, -1); len(matches) != 1 {
		t.Fatalf("want one captured text initialization, got %d", len(matches))
	}
	parameter := initialization.FindStringSubmatch(source)[2]
	extraOwner := regexp.MustCompile(`(?m)^\s*adamic_(?:retain|release)\(` + regexp.QuoteMeta(parameter) + `\);`)
	if extraOwner.MatchString(source) {
		t.Fatal("captured parameter still has an owner outside its cell")
	}
	source = initialization.ReplaceAllString(source, "${1}${2};")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	report := leakChecked(t, source, binary)
	if report == "" {
		t.Fatal("oracle leak check accepted the missing cell retain")
	}
	if runtime.GOOS == "linux" && !strings.Contains(report, "AddressSanitizer: heap-use-after-free") {
		t.Fatalf("want under-retained cell caught by ASan in leak check, got %s", report)
	}
	t.Logf("missing cell retain rejected by oracle leak check:\n%s", report)
}
