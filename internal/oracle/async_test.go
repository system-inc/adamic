package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// Generated-C mutants keep the lowering and runtime builds unchanged; warning failures are not kills.
func TestAsyncCompilerChecksCatchMutants(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_three.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	control := native.C(program)
	oracle := onNode(test, path)
	test.Run("wrong-resumed-value", func(test *testing.T) {
		code := strings.Replace(control, "= (double)resumed.number;", "= (double)resumed.number + 1;", 1)
		if code == control {
			test.Fatal("mutant did not change input")
		}
		binary := filepath.Join(test.TempDir(), "mutant")
		if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
			test.Fatal(buildError)
		}
		result := execute(test, binary)
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			test.Fatalf("only the Node stdout check must catch the mutant: %+v", result)
		}
		if report := leakcheck.Report(test, code, binary); report != "" {
			test.Fatal(report)
		}
		test.Log("wrong resumed value caught only by Node stdout comparison")
	})
	test.Run("missing-frame-parameter-retain", func(test *testing.T) {
		parameter := regexp.MustCompile(`(frame->cells\[[0-9]+\]\.value\.reference = )adamic_retain\(([^)]+)\);`)
		location := parameter.FindStringSubmatchIndex(control)
		if location == nil {
			test.Fatal("missing frame parameter store")
		}
		code := control[:location[0]] +
			parameter.ReplaceAllString(control[location[0]:location[1]], "${1}${2};") +
			control[location[1]:]
		if code == control {
			test.Fatal("mutant did not change input")
		}
		binary := filepath.Join(test.TempDir(), "mutant")
		if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
			test.Fatal(buildError)
		}
		result := execute(test, binary)
		if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-use-after-free") {
			test.Fatalf("ASan did not catch a borrowed dynamic parameter across await: %+v", result)
		}
		test.Log("missing frame parameter retain caught by ASan heap-use-after-free")
	})
}

// The source oracle covers the uncaught throw. This internal root harness consumes that rejection
// instead of panicking, so the shared leak checker can check frame cleanup on Linux and Darwin.
func TestAsyncThrowReleasesFrame(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_throw.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := native.C(program)
	begin := strings.Index(code, "if (root->rejected) {")
	if begin < 0 {
		test.Fatal("missing root cleanup hook")
	}
	end := strings.Index(code[begin:], "\n") + begin
	if end < begin {
		test.Fatal("missing root cleanup hook")
	}
	code = code[:begin] + "if (!root->rejected) return 2;" + code[end:]
	for _, mutant := range []bool{false, true} {
		test.Run(map[bool]string{false: "control", true: "missing-local-release"}[mutant], func(test *testing.T) {
			source := code
			if mutant {
				slot := regexp.MustCompile(`adamic_release\((adamic_slot_[0-9]+)\);`)
				location := slot.FindStringSubmatchIndex(source)
				if location == nil {
					test.Fatal("missing private frame slot release")
				}
				source = source[:location[0]] +
					slot.ReplaceAllString(source[location[0]:location[1]], "(void)$1;") +
					source[location[1]:]
			}
			binary := filepath.Join(test.TempDir(), "throw")
			if buildError := native.Build(source, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			report := leakcheck.Report(test, source, binary)
			// Exercise Darwin's counted rule on Linux too, using this actual frame mutant.
			counted := filepath.Join(test.TempDir(), "counted-proof")
			if buildError := native.Build(source, counted, native.Options{Count: true}); buildError != nil {
				test.Fatal(buildError)
			}
			countedReport := leakcheck.Unbalanced(leakRun(execute(test, counted)))
			if mutant {
				if !strings.Contains(report, "LeakSanitizer: detected memory leaks") &&
					!strings.Contains(report, "heap values leaked:") {
					test.Fatalf("leak mutant survived: %s", report)
				}
				if !strings.Contains(countedReport, "heap values leaked:") {
					test.Fatalf("counted rule missed local leak: %s", countedReport)
				}
				test.Logf("throw frame local release mutant caught: %s; counted rule: %s", report, countedReport)
			} else if report != "" || countedReport != "" {
				test.Fatalf("throw root cleanup leaked: %s; counted rule: %s", report, countedReport)
			}
		})
	}
}

// Mutate the shared lowered typeof operand so both backends can agree with each other and still be
// wrong. The source on Node must independently catch the mistaken async function classification.
func TestAsyncTypeOfWrongKindMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_typeof.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	oracle := onNode(test, path)
	want := "function\nfunction\nnested: function\nstring\nnumber\nshadow after await: number\ninside: function\nfunction\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		test.Fatalf("Node witness: %+v", oracle)
	}
	changed := false
	wrongKind := len(program.Strings)
	program.Strings = append(program.Strings, "object")
	var mutate func(reflect.Value) reflect.Value
	mutate = func(value reflect.Value) reflect.Value {
		if value.Kind() == reflect.Interface {
			result := reflect.New(value.Type()).Elem()
			if !value.IsNil() {
				result.Set(mutate(value.Elem()))
			}
			return result
		}
		if observation, ok := value.Interface().(ir.TypeOf); ok && observation.Value.Type() == ir.Closure {
			changed = true
			return reflect.ValueOf(ir.StringConstant{Index: wrongKind})
		}
		switch value.Kind() {
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for index := 0; index < value.NumField(); index++ {
				result.Field(index).Set(mutate(value.Field(index)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for index := 0; index < value.Len(); index++ {
				result.Index(index).Set(mutate(value.Index(index)))
			}
			return result
		}
		return value
	}
	program.Main = mutate(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for index := range program.Functions {
		program.Functions[index].Body = mutate(
			reflect.ValueOf(program.Functions[index].Body),
		).Interface().([]ir.Statement)
	}
	if !changed {
		test.Fatal("mutant did not change an async typeof operand")
	}
	binary := filepath.Join(test.TempDir(), "wrong-kind")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if report := leakcheck.Report(test, native.C(program), binary); report != "" {
		test.Fatal(report)
	}
	for name, result := range map[string]run{
		"native":     execute(test, binary),
		"JavaScript": onJavaScriptBackend(test, program),
	} {
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			test.Fatalf("only Node stdout must catch %s's wrong-kind mutant: %+v", name, result)
		}
		test.Logf("Node stdout catches %s's async function classified as object", name)
	}
}

// Keep the inherited fixtures in their registration order; new witnesses are
// explicit entries in oracle_test.go and run in the same bounded families.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/async_coverage_discard.a",
		"internal/oracle/testdata/async_coverage_parameters.a",
		"internal/oracle/testdata/async_coverage_reject_eager.a",
		"internal/oracle/testdata/async_coverage_reject_empty.a",
		"internal/oracle/testdata/async_coverage_reject_nested.a",
		"internal/oracle/testdata/async_coverage_typeof.a",
		"internal/oracle/testdata/async_coverage_unions.a",
		"internal/oracle/testdata/async_coverage_values.a",
		"internal/oracle/testdata/async_nested.a",
		"internal/oracle/testdata/async_plain.a",
		"internal/oracle/testdata/async_three.a",
		"internal/oracle/testdata/async_throw.a",
		"internal/oracle/testdata/async_typeof.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// Each suspension/capture witness is independently scheduled. The split counted
// build also holds generated declarations and frame cleanup to the source oracle.
func asyncLoopWitness(test *testing.T, name, want string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || (want != "" && string(truth.stdout) != want) {
		test.Fatalf("Node: %+v", truth)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	if diff := disagreement(truth, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatal("JavaScript: " + diff)
	}
	got, leaked, comparable := nativelyCheckingLeaks(test, program, true)
	if leaked != "" || !comparable {
		test.Fatalf("sanitized: %s", leaked)
	}
	if diff := disagreement(truth, got); diff != "" {
		test.Fatalf("native: %s: %+v", diff, got)
	}
	binary := filepath.Join(test.TempDir(), "split-counted")
	if buildError := native.Build(
		native.C(program),
		binary,
		native.Options{Split: true, Count: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
	got = execute(test, binary)
	if report := leakcheck.Unbalanced(leakRun(got)); report != "" {
		test.Fatal(report)
	}
	got.stderr = leakcheck.CountsLine.ReplaceAll(got.stderr, nil)
	if diff := disagreement(truth, got); diff != "" {
		test.Fatalf("split counted: %s: %+v", diff, got)
	}
	test.Logf("Node/native/JS/split counted: %q", truth.stdout)
}
func TestAsyncLoopsTopCatchBinding(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_top_catch_binding", "refused1\n")
}
func TestAsyncLoopsTopCatchAfterAwait(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_top_catch_after_await", "refused2\n")
}
func TestAsyncLoopsTopLoop(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_top_loop", "round 0\nround 1\nround 2\n")
}
func TestAsyncLoopsForOfBody(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_body", "1:10\n2:20\n3:30\n44\n")
}
func TestAsyncLoopsForOfMutation(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_mutation", "1\n20\n3\n4\n28\n")
}
func TestAsyncLoopsForOfReturn(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_return", "item2:found\n")
}
func TestAsyncLoopsSwitchBody(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_switch_body", "10 27 27 4\ncase1 case2 case2 case3\n")
}
func TestAsyncLoopsLoopCells(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_loop_cells", "0\n1\n2\n10\n11\n12\n12\niteration1\n0\n1\n")
}
func TestAsyncLoopsLoopBodyCapture(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_loop_body_capture", "item0 item1 item2\n")
}
func TestAsyncLoopsForOfCapture(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_capture", "10\n20\n30\n")
}
func TestAsyncLoopsForOfCells(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_cells", "10\n20\n30\n")
}
func TestAsyncLoopsForOfBodyCells(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_body_cells", "1:after1\n2:after2\n3:after3\n")
}
func TestAsyncLoopsForOfBreakCells(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_break_cells", "21\n41\n")
}
func TestAsyncLoopsMethodsSplit(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_methods", "")
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_top_catch_binding.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_top_catch_after_await.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_top_loop.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_body.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_mutation.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_return.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_switch_body.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_loop_cells.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_capture.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_cells.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_body_cells.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_break_cells.a", true, false})
}

func asyncLoopsRefusal(test *testing.T, name, want, output string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/async_refused/"+name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		test.Fatalf("Node: %+v", truth)
	}
	_, loweredError := lowered(test, path)
	if loweredError == nil || !strings.Contains(loweredError.Error(), want) {
		test.Fatalf("want %s: %v", want, loweredError)
	}
	test.Logf("Node %q; compiler %v", truth.stdout, loweredError)
}
func TestAsyncLoopsCallableReturnRemainsRefused(test *testing.T) {
	test.Parallel()
	asyncLoopsRefusal(test, "async_refuse_loop_cells_return", "hidden thenable", "2\n")
}
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_refused/async_refuse_loop_cells_return.a", false, false})
}

// The iteration mutant changes normalized control flow; the capture mutant reuses
// one counted binding. Both compile cleanly and must fail only against Node output.
func asyncLoopMutation(test *testing.T, name, kind string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	if kind == "iteration" || kind == "top-iteration" || kind == "copy" || kind == "switch" {
		var rewrite func(reflect.Value) reflect.Value
		rewrite = func(value reflect.Value) reflect.Value {
			if value.Kind() == reflect.Interface {
				out := reflect.New(value.Type()).Elem()
				if !value.IsNil() {
					out.Set(rewrite(value.Elem()))
				}
				return out
			}
			if !changed {
				if declaration, ok := value.Interface().(ir.Declare); ok && kind == "copy" && declaration.CopyCell {
					changed = true
					return reflect.ValueOf(ir.Debugger{})
				}
				if loop, ok := value.Interface().(ir.Loop); ok &&
					(kind == "iteration" || kind == "top-iteration") &&
					len(loop.Update) == 1 {
					if assign, ok := loop.Update[0].(ir.Assign); ok &&
						(program.Locals[assign.Local].Name == "suspension" ||
							(kind == "top-iteration" && program.Locals[assign.Local].Name == "round")) {
						if add, ok := assign.Value.(ir.Binary); ok && add.Operator == ir.Add {
							add.Right = ir.NumberConstant{Value: 2}
							assign.Value = add
							loop.Update[0] = assign
							changed = true
							return reflect.ValueOf(loop)
						}
					}
				}
				if branch, ok := value.Interface().(ir.Switch); ok && kind == "switch" && len(branch.Cases) > 0 {
					branch.Cases[0].Tests[0] = ir.NumberConstant{Value: 999}
					changed = true
					return reflect.ValueOf(branch)
				}
			}
			switch value.Kind() {
			case reflect.Struct:
				out := reflect.New(value.Type()).Elem()
				for index := 0; index < value.NumField(); index++ {
					out.Field(index).Set(rewrite(value.Field(index)))
				}
				return out
			case reflect.Slice:
				if value.IsNil() {
					return value
				}
				out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
				for index := 0; index < value.Len(); index++ {
					out.Index(index).Set(rewrite(value.Index(index)))
				}
				return out
			}
			return value
		}
		for index := range program.Functions {
			program.Functions[index].Body = rewrite(
				reflect.ValueOf(program.Functions[index].Body),
			).Interface().([]ir.Statement)
		}
	}
	code := native.C(program)
	if kind == "release" {
		for _, function := range program.Functions {
			if function.Name != "module" {
				continue
			}
			for slot, local := range function.FrameEnvironment {
				if program.Locals[local].Name != "error" {
					continue
				}
				old := fmt.Sprintf("adamic_release(adamic_slot_%d);", slot)
				next := strings.Replace(code, old, fmt.Sprintf("(void)adamic_slot_%d;", slot), 1)
				changed = next != code
				code = next
				break
			}
		}
	}
	if kind == "capture" {
		for _, function := range program.Functions {
			if changed || !function.Async {
				continue
			}
			for slot, local := range function.FrameEnvironment {
				binding := program.Locals[local]
				if !binding.IterationCell {
					continue
				}
				member, refs := "number", "false"
				if binding.Type.IsReference() {
					member, refs = "reference", "true"
				}
				pattern := regexp.MustCompile(fmt.Sprintf(
					`adamic_cell \*(adamic_temporary_[A-Za-z0-9_]+) = adamic_cell_new\(\(adamic_value\)\{\.%s = ([^\n]+)\}, %s\);\n\s*adamic_release\(frame->cells\[%d\]\.value\.reference\);\n\s*frame->cells\[%d\]\.value\.reference = [^;]+;`,
					member,
					refs,
					slot,
					slot,
				))
				match := pattern.FindStringSubmatch(code)
				if match == nil {
					continue
				}
				fresh, value := match[1], match[2]
				drop := ""
				if refs == "true" {
					drop = "adamic_release(" + fresh + "->value.reference);"
				}
				replacement := fmt.Sprintf(
					"adamic_cell *%s = frame->cells[%d].value.reference; if (%s == NULL) { %s = adamic_cell_new((adamic_value){.%s = %s}, %s); frame->cells[%d].value.reference = %s; } else { %s %s->value.%s = %s; %s->ready = true; }",
					fresh,
					slot,
					fresh,
					fresh,
					member,
					value,
					refs,
					slot,
					fresh,
					drop,
					fresh,
					member,
					value,
					fresh,
				)
				code = strings.Replace(code, match[0], replacement, 1)
				changed = true
				break
			}
		}
	}
	if !changed {
		test.Fatal("mutant did not change its input")
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if kind == "release" {
		if got.exitCode == 0 || !strings.Contains(string(got.stderr), "LeakSanitizer: detected memory leaks") {
			test.Fatalf("frame release mutant survived: %+v", got)
		}
	} else {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
			test.Fatalf("only Node stdout must catch %s: %+v", kind, got)
		}
		if kind != "capture" {
			if actual := onJavaScriptBackend(
				test,
				program,
			); actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(truth, actual) != "stdout differs" {
				test.Fatalf("JavaScript mutant survived: %+v", actual)
			}
		}
	}
	test.Logf("caught %s: Node %q; native %q; exit %d", kind, truth.stdout, got.stdout, got.exitCode)
}
func TestAsyncLoopsTopCatchBindingMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_top_catch_binding", "release")
}
func TestAsyncLoopsTopCatchAfterAwaitMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_top_catch_after_await", "release")
}
func TestAsyncLoopsTopLoopMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_top_loop", "top-iteration")
}
func TestAsyncLoopsForOfBodyMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_for_of_body", "iteration")
}
func TestAsyncLoopsForOfMutationMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_for_of_mutation", "iteration")
}
func TestAsyncLoopsForOfReturnMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_for_of_return", "iteration")
}
func TestAsyncLoopsSwitchBodyMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_switch_body", "switch")
}
func TestAsyncLoopsLoopCellsMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_loop_cells", "copy")
}
func TestAsyncLoopsLoopBodyCaptureMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_loop_body_capture", "capture")
}
func TestAsyncLoopsForOfCaptureMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_for_of_capture", "capture")
}
func TestAsyncLoopsForOfCellsMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_for_of_cells", "capture")
}
func TestAsyncLoopsForOfBodyCellsMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_for_of_body_cells", "capture")
}
func TestAsyncLoopsForOfBreakCellsMutant(test *testing.T) {
	test.Parallel()
	asyncLoopMutation(test, "async_for_of_break_cells", "capture")
}
