package oracle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

var asyncCountCases = []string{"stable", "reassigned", "captured", "expression_write", "parameter", "thrown", "root_error", "payload"}

func asyncCountPath(name string) string {
	return "internal/oracle/testdata/async_counts/async_counts_" + name + ".a"
}
func init() {
	for _, name := range asyncCountCases {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{asyncCountPath(name), true, false})
	}
}

// Only the outer retain call changes. All resulting mutants remain valid C.
func asyncCountDropRetain(source, needle string) string {
	start := strings.Index(source, needle)
	if start < 0 {
		panic("missing retain seam: " + needle)
	}
	start += strings.Index(needle, "adamic_retain(")
	body := start + len("adamic_retain(")
	depth := 1
	for end := body; end < len(source); end++ {
		switch source[end] {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 {
			return source[:start] + source[body:end] + source[end+1:]
		}
	}
	panic("unclosed retain")
}
func asyncCountMutant(t *testing.T, name, source string, p *ir.Program) string {
	t.Helper()
	slot := -1
	binding := "held"
	if name == "reassigned" {
		binding = "saved"
	}
	if name == "stable" || name == "parameter" || name == "payload" {
		binding = "seed"
	}
	for _, fn := range p.Functions {
		if fn.Name == "keep" {
			for position, index := range fn.FrameEnvironment {
				if p.Locals[index].Name == binding {
					slot = position
				}
			}
		}
	}
	switch name {
	case "stable", "parameter", "payload":
		return asyncCountDropRetain(source, fmt.Sprintf("frame->cells[%d].value.reference = adamic_retain(%s)", slot, emittedLocalSymbol(t, source, "adamic_local_", "seed")))
	case "reassigned":
		return asyncCountDropRetain(source, fmt.Sprintf("(&frame->cells[%d])->value.reference = adamic_retain(", slot))
	case "captured", "expression_write":
		needle := fmt.Sprintf("adamic_retain((adamic_string *)(&frame->cells[%d])->value.reference)", slot)
		at := strings.Index(source, needle)
		if at < 0 {
			t.Fatal("no owned mutable read")
		}
		line := source[strings.LastIndex(source[:at], "\n")+1 : at]
		match := regexp.MustCompile(`(adamic_temporary_[A-Za-z0-9_]+) = $`).FindStringSubmatch(line)
		if match == nil {
			t.Fatal("missing read temporary:", line)
		}
		changed := asyncCountDropRetain(source, needle)
		// A borrowed read has no statement cleanup of its own. Let the later write
		// free the value, so only its subsequent real use can catch the mutant.
		return strings.ReplaceAll(changed, "adamic_release("+match[1]+");", "")
	case "thrown":
		return asyncCountDropRetain(source, "adamic_thrown = adamic_retain(")
	case "root_error":
		needle := "frame->base.output = adamic_retain(output);"
		at := strings.LastIndex(source, needle)
		if at < 0 {
			t.Fatal("missing entry frame output owner")
		}
		return source[:at] + asyncCountDropRetain(source[at:], needle)

	}
	t.Fatal("unknown async count witness", name)
	return ""
}

func asyncCountProgramProduct(t *testing.T, name string, original bool) string {
	t.Helper()
	key := asyncCountPath(name)
	if original {
		key = "internal/oracle/testdata/" + name + ".a"
	}
	path, err := filepath.Abs(filepath.Join(repository, key))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(p)
	request, err := native.RuntimeRequestFor(native.Options{Sanitize: !original, Count: original}, native.Features(code)...)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: fmt.Sprintf("async-count-%t-%s", original, name), Files: append(compilerinputs.Lowering(), key, "internal/oracle/async_counts_test.go"), Flags: request.Portable, Toolchain: []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version"), buildcache.Tool("node", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		if original {
			return native.Build(code, filepath.Join(directory, "counted"), native.Options{Count: true})
		}
		if err := native.Build(code, filepath.Join(directory, "control"), native.Options{Sanitize: true}); err != nil {
			return err
		}
		if err := native.Build(code, filepath.Join(directory, "counted"), native.Options{Count: true}); err != nil {
			return err
		}
		return native.Build(asyncCountMutant(t, name, code, p), filepath.Join(directory, "mutant"), native.Options{Sanitize: true})
	})
}
func checkAsyncCountRow(t *testing.T, key string, got run, budget int) int {
	t.Helper()
	matches := leakcheck.CountsLine.FindSubmatch(got.stderr)
	if len(matches) != 7 {
		t.Fatalf("missing counts: %s", got.stderr)
	}
	retains, err := strconv.Atoi(string(matches[3]))
	if err != nil {
		t.Fatal(err)
	}
	if budget > 0 && retains > budget {
		t.Fatalf("unit 1 retain budget %d exceeded: %d", budget, retains)
	}
	row := "| " + key
	for _, match := range matches[1:] {
		row += " | " + string(match)
	}
	row += " |"
	recorded, err := recordedCounts()
	if err != nil {
		t.Fatal(err)
	}
	if recorded[key] != row {
		t.Fatalf("counts changed: got %s; recorded %s", row, recorded[key])
	}
	if got.exitCode == 0 {
		if leaked := leakcheck.Unbalanced(leakRun(got)); leaked != "" {
			t.Fatal(leaked)
		}
	}
	t.Log(row)
	return retains
}
func checkAsyncCountWitness(t *testing.T, name string, mutant bool) {
	t.Helper()
	directory := asyncCountProgramProduct(t, name, false)
	if mutant {
		got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"}, filepath.Join(directory, "mutant"))
		if got.exitCode == 0 || !strings.Contains(string(got.stderr), "AddressSanitizer: heap-use-after-free") {
			t.Fatalf("necessary owner mutant survived: exit %d stdout %q stderr %s", got.exitCode, got.stdout, got.stderr)
		}
		t.Log("necessary owner dropped: ASan heap-use-after-free")
		return
	}
	path, err := filepath.Abs(filepath.Join(repository, asyncCountPath(name)))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, filepath.Join(directory, "control"))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("ASan %s: %+v", difference, got)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, p)); difference != "" {
		t.Fatal("JavaScript: " + difference)
	}
	counted := execute(t, filepath.Join(directory, "counted"))
	checkAsyncCountRow(t, asyncCountPath(name), counted, 0)
	counted.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if difference := disagreement(truth, counted); difference != "" {
		t.Fatal("counted: " + difference)
	}
	t.Logf("Node, ASan, JavaScript and counted: stdout %q stderr %q exit %d", truth.stdout, truth.stderr, truth.exitCode)
}

func TestProduct_AsyncCountStable(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "stable", false)
}
func TestAsyncCountStable(t *testing.T) { t.Parallel(); checkAsyncCountWitness(t, "stable", false) }
func TestAsyncCountStableMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "stable", true)
}

func TestProduct_AsyncCountReassigned(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "reassigned", false)
}
func TestAsyncCountReassigned(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "reassigned", false)
}
func TestAsyncCountReassignedMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "reassigned", true)
}

func TestProduct_AsyncCountCaptured(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "captured", false)
}
func TestAsyncCountCaptured(t *testing.T) { t.Parallel(); checkAsyncCountWitness(t, "captured", false) }
func TestAsyncCountCapturedMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "captured", true)
}

func TestProduct_AsyncCountExpressionWrite(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "expression_write", false)
}
func TestAsyncCountExpressionWrite(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "expression_write", false)
}
func TestAsyncCountExpressionWriteMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "expression_write", true)
}

func TestProduct_AsyncCountParameter(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "parameter", false)
}
func TestAsyncCountParameter(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "parameter", false)
}
func TestAsyncCountParameterMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "parameter", true)
}

func TestProduct_AsyncCountThrown(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "thrown", false)
}
func TestAsyncCountThrown(t *testing.T) { t.Parallel(); checkAsyncCountWitness(t, "thrown", false) }
func TestAsyncCountThrownMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "thrown", true)
}

func TestProduct_AsyncCountRootError(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "root_error", false)
}
func TestAsyncCountRootError(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "root_error", false)
}
func TestAsyncCountRootErrorMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "root_error", true)
}

func TestProduct_AsyncOriginalCoverageDiscard(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_discard", true)
}
func TestAsyncOriginalCoverageDiscardRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_discard", 42)
}

func TestProduct_AsyncOriginalCoverageParameters(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_parameters", true)
}
func TestAsyncOriginalCoverageParametersRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_parameters", 18)
}

func TestProduct_AsyncOriginalCoverageRejectEager(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_reject_eager", true)
}
func TestAsyncOriginalCoverageRejectEagerRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_reject_eager", 18)
}

func TestProduct_AsyncOriginalCoverageRejectEmpty(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_reject_empty", true)
}
func TestAsyncOriginalCoverageRejectEmptyRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_reject_empty", 9)
}

func TestProduct_AsyncOriginalCoverageRejectNested(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_reject_nested", true)
}
func TestAsyncOriginalCoverageRejectNestedRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_reject_nested", 31)
}

func TestProduct_AsyncOriginalCoverageTypeof(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_typeof", true)
}
func TestAsyncOriginalCoverageTypeofRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_typeof", 24)
}

func TestProduct_AsyncOriginalCoverageUnions(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_unions", true)
}
func TestAsyncOriginalCoverageUnionsRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_unions", 60)
}

func TestProduct_AsyncOriginalCoverageValues(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_coverage_values", true)
}
func TestAsyncOriginalCoverageValuesRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_coverage_values", 126)
}

func TestProduct_AsyncOriginalNested(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_nested", true)
}
func TestAsyncOriginalNestedRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_nested", 38)
}

func TestProduct_AsyncOriginalPlain(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_plain", true)
}
func TestAsyncOriginalPlainRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_plain", 32)
}

func TestProduct_AsyncOriginalThree(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_three", true)
}
func TestAsyncOriginalThreeRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_three", 25)
}

func TestProduct_AsyncOriginalThrow(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_throw", true)
}
func TestAsyncOriginalThrowRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_throw", 16)
}

func TestProduct_AsyncOriginalTypeof(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "async_typeof", true)
}
func TestAsyncOriginalTypeofRetainBudget(t *testing.T) {
	t.Parallel()
	checkAsyncOriginalBudget(t, "async_typeof", 24)
}

func checkAsyncOriginalBudget(t *testing.T, name string, limit int) {
	t.Helper()
	directory := asyncCountProgramProduct(t, name, true)
	got := execute(t, filepath.Join(directory, "counted"))
	retains := checkAsyncCountRow(t, "internal/oracle/testdata/"+name+".a", got, limit)
	if retains > limit {
		t.Fatalf("unit 1 retain budget %d exceeded: %d", limit, retains)
	}
}

func TestProduct_AsyncCountPayload(t *testing.T) {
	t.Parallel()
	asyncCountProgramProduct(t, "payload", false)
}
func TestAsyncCountPayload(t *testing.T) { t.Parallel(); checkAsyncCountWitness(t, "payload", false) }
func TestAsyncCountPayloadFixtureMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountWitness(t, "payload", true)
}
