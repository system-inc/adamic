package native_test

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func programCoreCode(test *testing.T) string {
	test.Helper()
	data, readError := os.ReadFile("testdata/program_region/core.c")
	if readError != nil {
		test.Fatal(readError)
	}
	return string(data)
}
func programCoreRun(environment []string, command string, arguments ...string) leakcheck.Run {
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	process := exec.CommandContext(timeoutContext, command, arguments...)
	process.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	runError := process.Run()
	exit := 0
	if runError != nil {
		exit = -1
		if process.ProcessState != nil {
			exit = process.ProcessState.ExitCode()
		}
	}
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: exit}
}
func programCoreLeaks(test *testing.T, code, binary, mode string, build func(string, string) error) string {
	test.Helper()
	report, checkError := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary,
		Counted: filepath.Join(test.TempDir(), "counted"), BuildCounted: build,
		Arguments: func() []string { return []string{mode} }, Execute: programCoreRun})
	if checkError != nil {
		test.Fatal(checkError)
	}
	return report
}

const programRegionCoreOptionCount = 3

func TestProgramRegionCoreASan(test *testing.T)      { test.Parallel(); programRegionCoreUnit(test, 0) }
func TestProgramRegionCoreASanSlabs(test *testing.T) { test.Parallel(); programRegionCoreUnit(test, 1) }
func TestProgramRegionCoreCounted(test *testing.T)   { test.Parallel(); programRegionCoreUnit(test, 2) }
func TestProgramRegionCoreUnion(test *testing.T) {
	test.Parallel()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "program_region_core_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	seen := map[int]bool{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok ||
			!strings.HasPrefix(function.Name.Name, "TestProgramRegionCore") ||
			function.Name.Name == "TestProgramRegionCoreUnion" ||
			function.Name.Name == "TestProgramRegionCoreMutants" {
			continue
		}
		if function.Body == nil || len(function.Body.List) != 2 {
			test.Fatalf("unexpected unit body %s", function.Name.Name)
		}
		expression, ok := function.Body.List[1].(*ast.ExprStmt)
		if !ok {
			test.Fatal("unit must call core helper")
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			test.Fatal("unit must supply one option index")
		}
		helper, ok := call.Fun.(*ast.Ident)
		if !ok || helper.Name != "programRegionCoreUnit" {
			test.Fatal("unexpected unit helper")
		}
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok {
			test.Fatal("option index must be literal")
		}
		index, parseError := strconv.Atoi(literal.Value)
		if parseError != nil || index < 0 || index >= programRegionCoreOptionCount || seen[index] {
			test.Fatalf("duplicate or invalid option index %s", literal.Value)
		}
		seen[index] = true
	}
	if len(seen) != programRegionCoreOptionCount || programRegionCoreOptionCount != native.ProgramCoreAllocatorCount() {
		test.Fatal("core option inventory changed without updating top-level units")
	}
}
func programRegionCoreUnit(test *testing.T, index int) {
	test.Helper()
	code := programCoreCode(test)
	node := programCoreRun(nil, "node", "--input-type=commonjs", "-e",
		`require('node:vm').runInThisContext(require('node:fs').readFileSync(process.argv[1], 'utf8'))`,
		"testdata/program_region/oracle.a")
	if node.ExitCode != 0 || len(node.Stderr) != 0 {
		test.Fatalf("Node: exit %d stderr %s", node.ExitCode, node.Stderr)
	}

	binary := native.ProgramCoreBinary(test, index, "control")
	for _, mode := range []string{"end", "exit"} {
		run := programCoreRun(
			[]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"},
			binary,
			mode,
		)
		if run.ExitCode != 0 {
			test.Fatalf("%s: exit %d stderr %s", mode, run.ExitCode, run.Stderr)
		}
		if !bytes.Equal(run.Stdout, node.Stdout) {
			test.Fatalf("Node differs: native %q Node %q", run.Stdout, node.Stdout)
		}
		if report := leakcheck.Unbalanced(run); report != "" {
			test.Fatal(report)
		}
		counts := leakcheck.CountsLine.FindSubmatch(run.Stderr)
		if len(counts) != 7 {
			test.Fatalf("missing counts: %s", run.Stderr)
		}
		if string(counts[1]) != "11" || string(counts[2]) != "7" || string(counts[6]) != "4" {
			test.Fatalf("unexpected membership: %s", run.Stderr)
		}
		test.Logf("%s %s", mode, strings.TrimSpace(string(run.Stderr)))
	}
	// ASan malloc owns individual storage, so it is the independent Linux leak proof.
	if index == 0 {
		if report := programCoreLeaks(test, code, binary, "end", func(code, output string) error {
			data, readError := os.ReadFile(native.ProgramCoreBinary(test, 2, "control"))
			if readError != nil {
				return readError
			}
			return os.WriteFile(output, data, 0755)
		}); report != "" {
			test.Fatal(report)
		}
		if report := programCoreLeaks(test, code, binary, "exit", func(code, output string) error {
			data, readError := os.ReadFile(native.ProgramCoreBinary(test, 2, "control"))
			if readError != nil {
				return readError
			}
			return os.WriteFile(output, data, 0755)
		}); report != "" {
			test.Fatal(report)
		}
		for _, mode := range []string{
			"share",
			"share-cell",
			"aliased-adoption",
			"cached-adoption",
			"weak-before-adoption",
			"weak-cell-before-adoption",
		} {
			refused := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, mode)
			message := "Program region members cannot cross into parallel work"
			if mode == "cached-adoption" {
				message = "Program adoption must precede canonical caching"
			}
			if strings.HasPrefix(mode, "weak-") || mode == "aliased-adoption" {
				message = "Program adoption needs a fresh counted allocation"
			}
			if refused.ExitCode != 70 ||
				!strings.Contains(string(refused.Stderr), message) ||
				strings.Contains(string(refused.Stderr), "ERROR: AddressSanitizer") {
				test.Fatalf("%s: exit %d stderr %s", mode, refused.ExitCode, refused.Stderr)
			}
		}
	}
}

// Not parallel: exercises the explicit environment opt-in and restores it with test.Setenv.
func TestProgramRegionTarget(test *testing.T) {
	test.Setenv("ADAMIC_PROGRAM_REGION", "")
	const refusal = "native: Program regions require a one-shot native CLI"
	for _, options := range []native.Options{
		{ProgramRegion: true, Target: "wasm32-wasi"},
		{ProgramRegion: true, Request: true},
	} {
		if buildError := native.Build(
			"int main(void) {return 0;}",
			filepath.Join(test.TempDir(), "refused"),
			options,
		); buildError == nil || buildError.Error() != refusal {
			test.Fatalf("want %q, got %v", refusal, buildError)
		}
	}
	if !strings.Contains(
		strings.Join(native.Flags(native.Options{ProgramRegion: true}), " "),
		"-DADAMIC_PROGRAM_REGION",
	) {
		test.Fatal("missing opt-in flag")
	}
	if strings.Contains(strings.Join(native.Flags(native.Options{}), " "), "-DADAMIC_PROGRAM_REGION") {
		test.Fatal("default enabled Program region")
	}
	test.Setenv("ADAMIC_PROGRAM_REGION", "1")
	if validateOptionsError := native.ValidateOptions(
		native.Options{Target: "wasm32-wasi"},
	); validateOptionsError == nil || validateOptionsError.Error() != refusal {
		test.Fatalf("environment switch bypassed refusal: %v", validateOptionsError)
	}
}

func programRegionCoreMutant(test *testing.T, name, diagnostic string) {
	test.Helper()
	binary := native.ProgramCoreBinary(test, 0, name)
	run := programCoreRun(
		[]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"},
		binary,
		"end",
	)
	if run.ExitCode == 0 || !strings.Contains(string(run.Stderr), diagnostic) {
		test.Fatalf("want %s, exit %d: %s", diagnostic, run.ExitCode, run.Stderr)
	}
	test.Logf("%s caught by %s", name, diagnostic)
}
func TestProgramMemberReleaseMutant(test *testing.T) {
	test.Parallel()
	programRegionCoreMutant(test, "member-release-frees", "heap-use-after-free")
}
func TestProgramWeakInvalidationMutant(test *testing.T) {
	test.Parallel()
	programRegionCoreMutant(test, "weak-not-forgotten", "adamic_weak_target(weak) == NULL")
}
func TestProgramCellWeakInvalidationMutant(test *testing.T) {
	test.Parallel()
	programRegionCoreMutant(test, "cell-weak-not-forgotten", "adamic_weak_target(cell_weak) == NULL")
}
func TestProgramAllocationTailMutant(test *testing.T) {
	test.Parallel()
	programRegionCoreMutant(test, "short-adoption", "compiler bug: adoption copies short of an object's storage")
}

func programMembershipFault(test *testing.T, mode string) {
	test.Helper()
	binary := native.ProgramCoreBinary(test, 0, "control")
	source := programCoreRun(nil, "node", "-e", `console.log("membership " + String(12345))`)
	run := programCoreRun(
		[]string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"},
		binary,
		mode,
	)
	if source.ExitCode != 0 || run.ExitCode != 0 || !bytes.Equal(source.Stdout, run.Stdout) {
		test.Fatalf("Node %+v native %+v", source, run)
	}
	if report := leakcheck.Unbalanced(run); report != "" {
		test.Fatal(report)
	}
	counts := leakcheck.CountsLine.FindSubmatch(run.Stderr)
	if len(counts) != 7 || string(counts[6]) != "1" {
		test.Fatalf("membership must retain exactly one Program allocation: %s", run.Stderr)
	}
	test.Logf("membership fault %s agrees with Node %q; %s", mode, run.Stdout, strings.TrimSpace(string(run.Stderr)))
}
func TestProgramMembershipOverincludedMutant(test *testing.T) {
	test.Parallel()
	programMembershipFault(test, "overinclude")
}
func TestProgramMembershipUnderincludedCountedMutant(test *testing.T) {
	test.Parallel()
	programMembershipFault(test, "underinclude-counted")
}
func TestProgramMembershipUnderincludedGraphMutant(test *testing.T) {
	test.Parallel()
	programMembershipFault(test, "underinclude-graph")
}
func TestProgramMembershipFallbackRetainMutant(test *testing.T) {
	test.Parallel()
	binary := native.ProgramCoreBinary(test, 0, "fallback-retain-dropped")
	run := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"}, binary, "underinclude-graph")
	if run.ExitCode == 0 || !strings.Contains(string(run.Stderr), "heap-use-after-free") {
		test.Fatalf("missing owned edge survived: %+v", run)
	}
	test.Log("fallback retain mutant caught by ASan heap-use-after-free")
}

func programRefusalMutant(test *testing.T, name, mode, message string) {
	test.Helper()
	control := programCoreRun(
		[]string{"ASAN_OPTIONS=detect_leaks=0"},
		native.ProgramCoreBinary(test, 0, "control"),
		mode,
	)
	if control.ExitCode != 70 || !strings.Contains(string(control.Stderr), message) {
		test.Fatalf("missing control refusal: %+v", control)
	}
	mutant := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0"}, native.ProgramCoreBinary(test, 0, name), mode)
	if mutant.ExitCode != 0 {
		test.Fatalf("mutant must compile and reach the forbidden adoption/publication: %+v", mutant)
	}
	test.Logf("%s: refusal check catches removed guard (control exit 70, mutant exit 0)", name)
}
func TestProgramSharingRefusalMutant(test *testing.T) {
	test.Parallel()
	programRefusalMutant(test, "sharing-allowed", "share", "members cannot cross")
}
func TestProgramCanonicalAdoptionRefusalMutant(test *testing.T) {
	test.Parallel()
	programRefusalMutant(test, "cached-adoption-allowed", "cached-adoption", "must precede canonical caching")
}
func TestProgramUnmovableAdoptionRefusalMutant(test *testing.T) {
	test.Parallel()
	programRefusalMutant(
		test,
		"unmovable-adoption-allowed",
		"unmovable-adoption",
		"a number box can't be adopted into the Program region",
	)
}
func TestProgramWeakAdoptionRefusalMutant(test *testing.T) {
	test.Parallel()
	programRefusalMutant(test, "weak-adoption-allowed", "weak-before-adoption", "fresh counted allocation")
}
func TestProgramCellWeakAdoptionRefusalMutant(test *testing.T) {
	test.Parallel()
	programRefusalMutant(test, "weak-cell-adoption-allowed", "weak-cell-before-adoption", "fresh counted allocation")
}

func TestProgramAliasedAdoptionRefusalMutant(test *testing.T) {
	test.Parallel()
	control := programCoreRun(
		[]string{"ASAN_OPTIONS=detect_leaks=0"},
		native.ProgramCoreBinary(test, 0, "control"),
		"aliased-adoption",
	)
	if control.ExitCode != 70 || !strings.Contains(string(control.Stderr), "fresh counted allocation") {
		test.Fatalf("missing fresh adoption refusal: %+v", control)
	}
	mutant := programCoreRun(
		[]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"},
		native.ProgramCoreBinary(test, 0, "aliased-adoption-allowed"),
		"aliased-adoption",
	)
	if mutant.ExitCode == 0 || !strings.Contains(string(mutant.Stderr), "heap-use-after-free") {
		test.Fatalf("fresh adoption guard mutant survived: %+v", mutant)
	}
	test.Log("fresh adoption guard catches stale alias; removing it produces ASan heap-use-after-free")
}
