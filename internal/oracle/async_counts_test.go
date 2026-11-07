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

var asyncCountCases = []string{
	"stable",
	"reassigned",
	"captured",
	"expression_write",
	"parameter",
	"thrown",
	"root_error",
	"payload",
}

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
func asyncCountMutant(test *testing.T, name, source string, p *ir.Program) string {
	test.Helper()
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
		return asyncCountDropRetain(source, fmt.Sprintf(
			"frame->cells[%d].value.reference = adamic_retain(%s)",
			slot,
			emittedLocalSymbol(test, source, "adamic_local_", "seed"),
		))
	case "reassigned":
		return asyncCountDropRetain(source, fmt.Sprintf("(&frame->cells[%d])->value.reference = adamic_retain(", slot))
	case "captured", "expression_write":
		needle := fmt.Sprintf("adamic_retain((adamic_string *)(&frame->cells[%d])->value.reference)", slot)
		at := strings.Index(source, needle)
		if at < 0 {
			test.Fatal("no owned mutable read")
		}
		line := source[strings.LastIndex(source[:at], "\n")+1 : at]
		match := regexp.MustCompile(`(adamic_temporary_[A-Za-z0-9_]+) = $`).FindStringSubmatch(line)
		if match == nil {
			test.Fatal("missing read temporary:", line)
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
			test.Fatal("missing entry frame output owner")
		}
		return source[:at] + asyncCountDropRetain(source[at:], needle)

	}
	test.Fatal("unknown async count witness", name)
	return ""
}

func asyncCountProgramProduct(test *testing.T, name string, original bool) string {
	test.Helper()
	key := asyncCountPath(name)
	if original {
		key = "internal/oracle/testdata/" + name + ".a"
	}
	path, err := filepath.Abs(filepath.Join(repository, key))
	if err != nil {
		test.Fatal(err)
	}
	p, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	code := native.C(p)
	request, err := native.RuntimeRequestFor(
		native.Options{Sanitize: !original, Count: original},
		native.Features(code)...,
	)
	if err != nil {
		test.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:  fmt.Sprintf("async-count-%t-%s", original, name),
		Files: append(compilerinputs.Lowering(), key, "internal/oracle/async_counts_test.go"),
		Flags: request.Portable,
		Toolchain: []string{
			buildcache.Tool(request.Compiler, "--version"),
			buildcache.Tool(request.Archiver, "--version"),
			buildcache.Tool("node", "--version"),
		},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		if original {
			return native.Build(code, filepath.Join(directory, "counted"), native.Options{Count: true})
		}
		if err := native.Build(code, filepath.Join(directory, "control"), native.Options{Sanitize: true}); err != nil {
			return err
		}
		if err := native.Build(code, filepath.Join(directory, "counted"), native.Options{Count: true}); err != nil {
			return err
		}
		return native.Build(
			asyncCountMutant(test, name, code, p),
			filepath.Join(directory, "mutant"),
			native.Options{Sanitize: true},
		)
	})
}
func checkAsyncCountRow(test *testing.T, key string, got run, budget int) int {
	test.Helper()
	matches := leakcheck.CountsLine.FindSubmatch(got.stderr)
	if len(matches) != 7 {
		test.Fatalf("missing counts: %s", got.stderr)
	}
	retains, err := strconv.Atoi(string(matches[3]))
	if err != nil {
		test.Fatal(err)
	}
	if budget > 0 && retains > budget {
		test.Fatalf("unit 1 retain budget %d exceeded: %d", budget, retains)
	}
	row := "| " + key
	for _, match := range matches[1:] {
		row += " | " + string(match)
	}
	row += " |"
	recorded, err := recordedCounts()
	if err != nil {
		test.Fatal(err)
	}
	if recorded[key] != row {
		test.Fatalf("counts changed: got %s; recorded %s", row, recorded[key])
	}
	if got.exitCode == 0 {
		if leaked := leakcheck.Unbalanced(leakRun(got)); leaked != "" {
			test.Fatal(leaked)
		}
	}
	test.Log(row)
	return retains
}
func checkAsyncCountWitness(test *testing.T, name string, mutant bool) {
	test.Helper()
	directory := asyncCountProgramProduct(test, name, false)
	if mutant {
		got := executeWith(
			test,
			[]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1"},
			filepath.Join(directory, "mutant"),
		)
		if got.exitCode == 0 || !strings.Contains(string(got.stderr), "AddressSanitizer: heap-use-after-free") {
			test.Fatalf(
				"necessary owner mutant survived: exit %d stdout %q stderr %s",
				got.exitCode,
				got.stdout,
				got.stderr,
			)
		}
		test.Log("necessary owner dropped: ASan heap-use-after-free")
		return
	}
	path, err := filepath.Abs(filepath.Join(repository, asyncCountPath(name)))
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	got := executeWith(
		test,
		[]string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"},
		filepath.Join(directory, "control"),
	)
	if difference := disagreement(truth, got); difference != "" {
		test.Fatalf("ASan %s: %+v", difference, got)
	}
	p, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	if difference := disagreement(truth, onJavaScriptBackend(test, p)); difference != "" {
		test.Fatal("JavaScript: " + difference)
	}
	counted := execute(test, filepath.Join(directory, "counted"))
	checkAsyncCountRow(test, asyncCountPath(name), counted, 0)
	counted.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if difference := disagreement(truth, counted); difference != "" {
		test.Fatal("counted: " + difference)
	}
	test.Logf(
		"Node, ASan, JavaScript and counted: stdout %q stderr %q exit %d",
		truth.stdout,
		truth.stderr,
		truth.exitCode,
	)
}

func TestProduct_AsyncCountStable(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "stable", false)
}
func TestAsyncCountStable(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "stable", false)
}
func TestAsyncCountStableMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "stable", true)
}

func TestProduct_AsyncCountReassigned(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "reassigned", false)
}
func TestAsyncCountReassigned(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "reassigned", false)
}
func TestAsyncCountReassignedMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "reassigned", true)
}

func TestProduct_AsyncCountCaptured(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "captured", false)
}
func TestAsyncCountCaptured(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "captured", false)
}
func TestAsyncCountCapturedMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "captured", true)
}

func TestProduct_AsyncCountExpressionWrite(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "expression_write", false)
}
func TestAsyncCountExpressionWrite(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "expression_write", false)
}
func TestAsyncCountExpressionWriteMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "expression_write", true)
}

func TestProduct_AsyncCountParameter(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "parameter", false)
}
func TestAsyncCountParameter(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "parameter", false)
}
func TestAsyncCountParameterMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "parameter", true)
}

func TestProduct_AsyncCountThrown(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "thrown", false)
}
func TestAsyncCountThrown(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "thrown", false)
}
func TestAsyncCountThrownMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "thrown", true)
}

func TestProduct_AsyncCountRootError(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "root_error", false)
}
func TestAsyncCountRootError(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "root_error", false)
}
func TestAsyncCountRootErrorMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "root_error", true)
}

func TestProduct_AsyncOriginalCoverageDiscard(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_discard", true)
}
func TestAsyncOriginalCoverageDiscardRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_discard", 42)
}

func TestProduct_AsyncOriginalCoverageParameters(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_parameters", true)
}
func TestAsyncOriginalCoverageParametersRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_parameters", 18)
}

func TestProduct_AsyncOriginalCoverageRejectEager(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_reject_eager", true)
}
func TestAsyncOriginalCoverageRejectEagerRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_reject_eager", 18)
}

func TestProduct_AsyncOriginalCoverageRejectEmpty(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_reject_empty", true)
}
func TestAsyncOriginalCoverageRejectEmptyRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_reject_empty", 9)
}

func TestProduct_AsyncOriginalCoverageRejectNested(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_reject_nested", true)
}
func TestAsyncOriginalCoverageRejectNestedRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_reject_nested", 31)
}

func TestProduct_AsyncOriginalCoverageTypeof(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_typeof", true)
}
func TestAsyncOriginalCoverageTypeofRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_typeof", 24)
}

func TestProduct_AsyncOriginalCoverageUnions(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_unions", true)
}
func TestAsyncOriginalCoverageUnionsRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_unions", 60)
}

func TestProduct_AsyncOriginalCoverageValues(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_coverage_values", true)
}
func TestAsyncOriginalCoverageValuesRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_coverage_values", 126)
}

func TestProduct_AsyncOriginalNested(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_nested", true)
}
func TestAsyncOriginalNestedRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_nested", 38)
}

func TestProduct_AsyncOriginalPlain(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_plain", true)
}
func TestAsyncOriginalPlainRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_plain", 32)
}

func TestProduct_AsyncOriginalThree(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_three", true)
}
func TestAsyncOriginalThreeRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_three", 25)
}

func TestProduct_AsyncOriginalThrow(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_throw", true)
}
func TestAsyncOriginalThrowRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_throw", 16)
}

func TestProduct_AsyncOriginalTypeof(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "async_typeof", true)
}
func TestAsyncOriginalTypeofRetainBudget(test *testing.T) {
	test.Parallel()
	checkAsyncOriginalBudget(test, "async_typeof", 24)
}

func checkAsyncOriginalBudget(test *testing.T, name string, limit int) {
	test.Helper()
	directory := asyncCountProgramProduct(test, name, true)
	got := execute(test, filepath.Join(directory, "counted"))
	retains := checkAsyncCountRow(test, "internal/oracle/testdata/"+name+".a", got, limit)
	if retains > limit {
		test.Fatalf("unit 1 retain budget %d exceeded: %d", limit, retains)
	}
}

func TestProduct_AsyncCountPayload(test *testing.T) {
	test.Parallel()
	asyncCountProgramProduct(test, "payload", false)
}
func TestAsyncCountPayload(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "payload", false)
}
func TestAsyncCountPayloadFixtureMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountWitness(test, "payload", true)
}
