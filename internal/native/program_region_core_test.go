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

func programCoreCode(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/program_region/core.c")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func programCoreRun(environment []string, command string, arguments ...string) leakcheck.Run {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	process := exec.CommandContext(ctx, command, arguments...)
	process.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	err := process.Run()
	exit := 0
	if err != nil {
		exit = -1
		if process.ProcessState != nil {
			exit = process.ProcessState.ExitCode()
		}
	}
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: exit}
}
func programCoreLeaks(t *testing.T, code, binary, mode string, build func(string, string) error) string {
	t.Helper()
	report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary,
		Counted: filepath.Join(t.TempDir(), "counted"), BuildCounted: build,
		Arguments: func() []string { return []string{mode} }, Execute: programCoreRun})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

const programRegionCoreOptionCount = 3

func TestProgramRegionCoreASan(t *testing.T)      { t.Parallel(); programRegionCoreUnit(t, 0) }
func TestProgramRegionCoreASanSlabs(t *testing.T) { t.Parallel(); programRegionCoreUnit(t, 1) }
func TestProgramRegionCoreCounted(t *testing.T)   { t.Parallel(); programRegionCoreUnit(t, 2) }
func TestProgramRegionCoreUnion(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), "program_region_core_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[int]bool{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestProgramRegionCore") || function.Name.Name == "TestProgramRegionCoreUnion" || function.Name.Name == "TestProgramRegionCoreMutants" {
			continue
		}
		if function.Body == nil || len(function.Body.List) != 2 {
			t.Fatalf("unexpected unit body %s", function.Name.Name)
		}
		expression, ok := function.Body.List[1].(*ast.ExprStmt)
		if !ok {
			t.Fatal("unit must call core helper")
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			t.Fatal("unit must supply one option index")
		}
		helper, ok := call.Fun.(*ast.Ident)
		if !ok || helper.Name != "programRegionCoreUnit" {
			t.Fatal("unexpected unit helper")
		}
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok {
			t.Fatal("option index must be literal")
		}
		index, err := strconv.Atoi(literal.Value)
		if err != nil || index < 0 || index >= programRegionCoreOptionCount || seen[index] {
			t.Fatalf("duplicate or invalid option index %s", literal.Value)
		}
		seen[index] = true
	}
	if len(seen) != programRegionCoreOptionCount || programRegionCoreOptionCount != native.ProgramCoreAllocatorCount() {
		t.Fatal("core option inventory changed without updating top-level units")
	}
}
func programRegionCoreUnit(t *testing.T, index int) {
	t.Helper()
	code := programCoreCode(t)
	node := programCoreRun(nil, "node", "--input-type=commonjs", "-e",
		`require('node:vm').runInThisContext(require('node:fs').readFileSync(process.argv[1], 'utf8'))`,
		"testdata/program_region/oracle.a")
	if node.ExitCode != 0 || len(node.Stderr) != 0 {
		t.Fatalf("Node: exit %d stderr %s", node.ExitCode, node.Stderr)
	}

	binary := native.ProgramCoreBinary(t, index, "control")
	for _, mode := range []string{"end", "exit"} {
		run := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mode)
		if run.ExitCode != 0 {
			t.Fatalf("%s: exit %d stderr %s", mode, run.ExitCode, run.Stderr)
		}
		if !bytes.Equal(run.Stdout, node.Stdout) {
			t.Fatalf("Node differs: native %q Node %q", run.Stdout, node.Stdout)
		}
		if report := leakcheck.Unbalanced(run); report != "" {
			t.Fatal(report)
		}
		counts := leakcheck.CountsLine.FindSubmatch(run.Stderr)
		if len(counts) != 7 {
			t.Fatalf("missing counts: %s", run.Stderr)
		}
		if string(counts[1]) != "11" || string(counts[2]) != "7" || string(counts[6]) != "4" {
			t.Fatalf("unexpected membership: %s", run.Stderr)
		}
		t.Logf("%s %s", mode, strings.TrimSpace(string(run.Stderr)))
	}
	// ASan malloc owns individual storage, so it is the independent Linux leak proof.
	if index == 0 {
		if report := programCoreLeaks(t, code, binary, "end", func(code, output string) error {
			data, err := os.ReadFile(native.ProgramCoreBinary(t, 2, "control"))
			if err != nil {
				return err
			}
			return os.WriteFile(output, data, 0755)
		}); report != "" {
			t.Fatal(report)
		}
		if report := programCoreLeaks(t, code, binary, "exit", func(code, output string) error {
			data, err := os.ReadFile(native.ProgramCoreBinary(t, 2, "control"))
			if err != nil {
				return err
			}
			return os.WriteFile(output, data, 0755)
		}); report != "" {
			t.Fatal(report)
		}
		for _, mode := range []string{"share", "share-cell", "aliased-adoption", "cached-adoption", "weak-before-adoption", "weak-cell-before-adoption"} {
			refused := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, mode)
			message := "Program region members cannot cross into parallel work"
			if mode == "cached-adoption" {
				message = "Program adoption must precede canonical caching"
			}
			if strings.HasPrefix(mode, "weak-") || mode == "aliased-adoption" {
				message = "Program adoption needs a fresh counted allocation"
			}
			if refused.ExitCode != 70 || !strings.Contains(string(refused.Stderr), message) || strings.Contains(string(refused.Stderr), "ERROR: AddressSanitizer") {
				t.Fatalf("%s: exit %d stderr %s", mode, refused.ExitCode, refused.Stderr)
			}
		}
	}
}

// Not parallel: exercises the explicit environment opt-in and restores it with t.Setenv.
func TestProgramRegionTarget(t *testing.T) {
	t.Setenv("ADAMIC_PROGRAM_REGION", "")
	const refusal = "native: Program regions require a one-shot native CLI"
	for _, options := range []native.Options{{ProgramRegion: true, Target: "wasm32-wasi"}, {ProgramRegion: true, Request: true}} {
		if err := native.Build("int main(void) {return 0;}", filepath.Join(t.TempDir(), "refused"), options); err == nil || err.Error() != refusal {
			t.Fatalf("want %q, got %v", refusal, err)
		}
	}
	if !strings.Contains(strings.Join(native.Flags(native.Options{ProgramRegion: true}), " "), "-DADAMIC_PROGRAM_REGION") {
		t.Fatal("missing opt-in flag")
	}
	if strings.Contains(strings.Join(native.Flags(native.Options{}), " "), "-DADAMIC_PROGRAM_REGION") {
		t.Fatal("default enabled Program region")
	}
	t.Setenv("ADAMIC_PROGRAM_REGION", "1")
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err == nil || err.Error() != refusal {
		t.Fatalf("environment switch bypassed refusal: %v", err)
	}
}

func programRegionCoreMutant(t *testing.T, name, diagnostic string) {
	t.Helper()
	binary := native.ProgramCoreBinary(t, 0, name)
	run := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, "end")
	if run.ExitCode == 0 || !strings.Contains(string(run.Stderr), diagnostic) {
		t.Fatalf("want %s, exit %d: %s", diagnostic, run.ExitCode, run.Stderr)
	}
	t.Logf("%s caught by %s", name, diagnostic)
}
func TestProgramMemberReleaseMutant(t *testing.T) {
	t.Parallel()
	programRegionCoreMutant(t, "member-release-frees", "heap-use-after-free")
}
func TestProgramWeakInvalidationMutant(t *testing.T) {
	t.Parallel()
	programRegionCoreMutant(t, "weak-not-forgotten", "adamic_weak_target(weak) == NULL")
}
func TestProgramCellWeakInvalidationMutant(t *testing.T) {
	t.Parallel()
	programRegionCoreMutant(t, "cell-weak-not-forgotten", "adamic_weak_target(cell_weak) == NULL")
}
func TestProgramAllocationTailMutant(t *testing.T) {
	t.Parallel()
	programRegionCoreMutant(t, "short-adoption", "heap-buffer-overflow")
}

func programMembershipFault(t *testing.T, mode string) {
	t.Helper()
	binary := native.ProgramCoreBinary(t, 0, "control")
	source := programCoreRun(nil, "node", "-e", `console.log("membership " + String(12345))`)
	run := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mode)
	if source.ExitCode != 0 || run.ExitCode != 0 || !bytes.Equal(source.Stdout, run.Stdout) {
		t.Fatalf("Node %+v native %+v", source, run)
	}
	if report := leakcheck.Unbalanced(run); report != "" {
		t.Fatal(report)
	}
	counts := leakcheck.CountsLine.FindSubmatch(run.Stderr)
	if len(counts) != 7 || string(counts[6]) != "1" {
		t.Fatalf("membership must retain exactly one Program allocation: %s", run.Stderr)
	}
	t.Logf("membership fault %s agrees with Node %q; %s", mode, run.Stdout, strings.TrimSpace(string(run.Stderr)))
}
func TestProgramMembershipOverincludedMutant(t *testing.T) {
	t.Parallel()
	programMembershipFault(t, "overinclude")
}
func TestProgramMembershipUnderincludedCountedMutant(t *testing.T) {
	t.Parallel()
	programMembershipFault(t, "underinclude-counted")
}
func TestProgramMembershipUnderincludedGraphMutant(t *testing.T) {
	t.Parallel()
	programMembershipFault(t, "underinclude-graph")
}
func TestProgramMembershipFallbackRetainMutant(t *testing.T) {
	t.Parallel()
	binary := native.ProgramCoreBinary(t, 0, "fallback-retain-dropped")
	run := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"}, binary, "underinclude-graph")
	if run.ExitCode == 0 || !strings.Contains(string(run.Stderr), "heap-use-after-free") {
		t.Fatalf("missing owned edge survived: %+v", run)
	}
	t.Log("fallback retain mutant caught by ASan heap-use-after-free")
}

func programRefusalMutant(t *testing.T, name, mode, message string) {
	t.Helper()
	control := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0"}, native.ProgramCoreBinary(t, 0, "control"), mode)
	if control.ExitCode != 70 || !strings.Contains(string(control.Stderr), message) {
		t.Fatalf("missing control refusal: %+v", control)
	}
	mutant := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0"}, native.ProgramCoreBinary(t, 0, name), mode)
	if mutant.ExitCode != 0 {
		t.Fatalf("mutant must compile and reach the forbidden adoption/publication: %+v", mutant)
	}
	t.Logf("%s: refusal check catches removed guard (control exit 70, mutant exit 0)", name)
}
func TestProgramSharingRefusalMutant(t *testing.T) {
	t.Parallel()
	programRefusalMutant(t, "sharing-allowed", "share", "members cannot cross")
}
func TestProgramCanonicalAdoptionRefusalMutant(t *testing.T) {
	t.Parallel()
	programRefusalMutant(t, "cached-adoption-allowed", "cached-adoption", "must precede canonical caching")
}
func TestProgramUnmovableAdoptionRefusalMutant(t *testing.T) {
	t.Parallel()
	programRefusalMutant(t, "unmovable-adoption-allowed", "unmovable-adoption", "a number box can't be adopted into the Program region")
}
func TestProgramWeakAdoptionRefusalMutant(t *testing.T) {
	t.Parallel()
	programRefusalMutant(t, "weak-adoption-allowed", "weak-before-adoption", "fresh counted allocation")
}
func TestProgramCellWeakAdoptionRefusalMutant(t *testing.T) {
	t.Parallel()
	programRefusalMutant(t, "weak-cell-adoption-allowed", "weak-cell-before-adoption", "fresh counted allocation")
}

func TestProgramAliasedAdoptionRefusalMutant(t *testing.T) {
	t.Parallel()
	control := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0"}, native.ProgramCoreBinary(t, 0, "control"), "aliased-adoption")
	if control.ExitCode != 70 || !strings.Contains(string(control.Stderr), "fresh counted allocation") {
		t.Fatalf("missing fresh adoption refusal: %+v", control)
	}
	mutant := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"}, native.ProgramCoreBinary(t, 0, "aliased-adoption-allowed"), "aliased-adoption")
	if mutant.ExitCode == 0 || !strings.Contains(string(mutant.Stderr), "heap-use-after-free") {
		t.Fatalf("fresh adoption guard mutant survived: %+v", mutant)
	}
	t.Log("fresh adoption guard catches stale alias; removing it produces ASan heap-use-after-free")
}
