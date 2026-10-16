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
func TestAsyncCompilerChecksCatchMutants(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_three.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	oracle := onNode(t, path)
	t.Run("wrong-resumed-value", func(t *testing.T) {
		code := strings.Replace(control, "= (double)resumed.number;", "= (double)resumed.number + 1;", 1)
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, binary)
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("only the Node stdout check must catch the mutant: %+v", result)
		}
		if report := leakcheck.Report(t, code, binary); report != "" {
			t.Fatal(report)
		}
		t.Log("wrong resumed value caught only by Node stdout comparison")
	})
	t.Run("missing-frame-parameter-retain", func(t *testing.T) {
		parameter := regexp.MustCompile(`(frame->cells\[[0-9]+\]\.value\.reference = )adamic_retain\(([^)]+)\);`)
		location := parameter.FindStringSubmatchIndex(control)
		if location == nil {
			t.Fatal("missing frame parameter store")
		}
		code := control[:location[0]] + parameter.ReplaceAllString(control[location[0]:location[1]], "${1}${2};") + control[location[1]:]
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, binary)
		if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-use-after-free") {
			t.Fatalf("ASan did not catch a borrowed dynamic parameter across await: %+v", result)
		}
		t.Log("missing frame parameter retain caught by ASan heap-use-after-free")
	})
}

// The source oracle covers the uncaught throw. This internal root harness consumes that rejection
// instead of panicking, so the shared leak checker can check frame cleanup on Linux and Darwin.
func TestAsyncThrowReleasesFrame(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_throw.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	begin := strings.Index(code, "if (root->rejected) {")
	if begin < 0 {
		t.Fatal("missing root cleanup hook")
	}
	end := strings.Index(code[begin:], "\n") + begin
	if end < begin {
		t.Fatal("missing root cleanup hook")
	}
	code = code[:begin] + "if (!root->rejected) return 2;" + code[end:]
	for _, mutant := range []bool{false, true} {
		t.Run(map[bool]string{false: "control", true: "missing-local-release"}[mutant], func(t *testing.T) {
			source := code
			if mutant {
				slot := regexp.MustCompile(`adamic_release\((adamic_slot_[0-9]+)\);`)
				location := slot.FindStringSubmatchIndex(source)
				if location == nil {
					t.Fatal("missing private frame slot release")
				}
				source = source[:location[0]] + slot.ReplaceAllString(source[location[0]:location[1]], "(void)$1;") + source[location[1]:]
			}
			binary := filepath.Join(t.TempDir(), "throw")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			report := leakcheck.Report(t, source, binary)
			// Exercise Darwin's counted rule on Linux too, using this actual frame mutant.
			counted := filepath.Join(t.TempDir(), "counted-proof")
			if err := native.Build(source, counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			countedReport := leakcheck.Unbalanced(leakRun(execute(t, counted)))
			if mutant {
				if !strings.Contains(report, "LeakSanitizer: detected memory leaks") && !strings.Contains(report, "heap values leaked:") {
					t.Fatalf("leak mutant survived: %s", report)
				}
				if !strings.Contains(countedReport, "heap values leaked:") {
					t.Fatalf("counted rule missed local leak: %s", countedReport)
				}
				t.Logf("throw frame local release mutant caught: %s; counted rule: %s", report, countedReport)
			} else if report != "" || countedReport != "" {
				t.Fatalf("throw root cleanup leaked: %s; counted rule: %s", report, countedReport)
			}
		})
	}
}

// Mutate the shared lowered typeof operand so both backends can agree with each other and still be
// wrong. The source on Node must independently catch the mistaken async function classification.
func TestAsyncTypeOfWrongKindMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_typeof.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	want := "function\nfunction\nnested: function\nstring\nnumber\nshadow after await: number\ninside: function\nfunction\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		t.Fatalf("Node witness: %+v", oracle)
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
		program.Functions[index].Body = mutate(reflect.ValueOf(program.Functions[index].Body)).Interface().([]ir.Statement)
	}
	if !changed {
		t.Fatal("mutant did not change an async typeof operand")
	}
	binary := filepath.Join(t.TempDir(), "wrong-kind")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if report := leakcheck.Report(t, native.C(program), binary); report != "" {
		t.Fatal(report)
	}
	for name, result := range map[string]run{
		"native":     execute(t, binary),
		"JavaScript": onJavaScriptBackend(t, program),
	} {
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("only Node stdout must catch %s's wrong-kind mutant: %+v", name, result)
		}
		t.Logf("Node stdout catches %s's async function classified as object", name)
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
func asyncLoopWitness(t *testing.T, name, want string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || (want != "" && string(truth.stdout) != want) {
		t.Fatalf("Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(truth, onJavaScriptBackend(t, program)); diff != "" {
		t.Fatal("JavaScript: " + diff)
	}
	got, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" || !comparable {
		t.Fatalf("sanitized: %s", leaked)
	}
	if diff := disagreement(truth, got); diff != "" {
		t.Fatalf("native: %s: %+v", diff, got)
	}
	binary := filepath.Join(t.TempDir(), "split-counted")
	if err := native.Build(native.C(program), binary, native.Options{Split: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	got = execute(t, binary)
	if report := leakcheck.Unbalanced(leakRun(got)); report != "" {
		t.Fatal(report)
	}
	got.stderr = leakcheck.CountsLine.ReplaceAll(got.stderr, nil)
	if diff := disagreement(truth, got); diff != "" {
		t.Fatalf("split counted: %s: %+v", diff, got)
	}
	t.Logf("Node/native/JS/split counted: %q", truth.stdout)
}
func TestAsyncLoopsTopCatchBinding(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_top_catch_binding", "refused1\n")
}
func TestAsyncLoopsTopCatchAfterAwait(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_top_catch_after_await", "refused2\n")
}
func TestAsyncLoopsTopLoop(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_top_loop", "round 0\nround 1\nround 2\n")
}
func TestAsyncLoopsForOfBody(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_body", "1:10\n2:20\n3:30\n44\n")
}
func TestAsyncLoopsForOfMutation(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_mutation", "1\n20\n3\n4\n28\n")
}
func TestAsyncLoopsForOfReturn(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_return", "item2:found\n")
}
func TestAsyncLoopsSwitchBody(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_switch_body", "10 27 27 4\ncase1 case2 case2 case3\n")
}
func TestAsyncLoopsLoopCells(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_loop_cells", "0\n1\n2\n10\n11\n12\n12\niteration1\n0\n1\n")
}
func TestAsyncLoopsLoopBodyCapture(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_loop_body_capture", "item0 item1 item2\n")
}
func TestAsyncLoopsForOfCapture(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_capture", "10\n20\n30\n")
}
func TestAsyncLoopsForOfCells(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_cells", "10\n20\n30\n")
}
func TestAsyncLoopsForOfBodyCells(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_body_cells", "1:after1\n2:after2\n3:after3\n")
}
func TestAsyncLoopsForOfBreakCells(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_break_cells", "21\n41\n")
}
func TestAsyncLoopsMethodsSplit(t *testing.T) { t.Parallel(); asyncLoopWitness(t, "async_methods", "") }

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

func asyncLoopsRefusal(t *testing.T, name, want, output string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_refused/"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		t.Fatalf("Node: %+v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("want %s: %v", want, err)
	}
	t.Logf("Node %q; compiler %v", truth.stdout, err)
}
func TestAsyncLoopsStringRemainsRefused(t *testing.T) {
	t.Parallel()
	asyncLoopsRefusal(t, "async_refuse_for_of_string", "async for-of over this iterable or pattern", "a\nb\n")
}
func TestAsyncLoopsCallableReturnRemainsRefused(t *testing.T) {
	t.Parallel()
	asyncLoopsRefusal(t, "async_refuse_loop_cells_return", "hidden thenable", "2\n")
}
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_refused/async_refuse_for_of_string.a", false, false}, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_refused/async_refuse_loop_cells_return.a", false, false})
}

// The iteration mutant changes normalized control flow; the capture mutant reuses
// one counted binding. Both compile cleanly and must fail only against Node output.
func asyncLoopMutation(t *testing.T, name, kind string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
				if loop, ok := value.Interface().(ir.Loop); ok && (kind == "iteration" || kind == "top-iteration") && len(loop.Update) == 1 {
					if assign, ok := loop.Update[0].(ir.Assign); ok && (program.Locals[assign.Local].Name == "suspension" || (kind == "top-iteration" && program.Locals[assign.Local].Name == "round")) {
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
				for i := 0; i < value.NumField(); i++ {
					out.Field(i).Set(rewrite(value.Field(i)))
				}
				return out
			case reflect.Slice:
				if value.IsNil() {
					return value
				}
				out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
				for i := 0; i < value.Len(); i++ {
					out.Index(i).Set(rewrite(value.Index(i)))
				}
				return out
			}
			return value
		}
		for i := range program.Functions {
			program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
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
				pattern := regexp.MustCompile(fmt.Sprintf(`adamic_cell \*(adamic_temporary_[0-9]+) = adamic_cell_new\(\(adamic_value\)\{\.%s = ([^\n]+)\}, %s\);\n\s*adamic_release\(frame->cells\[%d\]\.value\.reference\);\n\s*frame->cells\[%d\]\.value\.reference = [^;]+;`, member, refs, slot, slot))
				match := pattern.FindStringSubmatch(code)
				if match == nil {
					continue
				}
				fresh, value := match[1], match[2]
				drop := ""
				if refs == "true" {
					drop = "adamic_release(" + fresh + "->value.reference);"
				}
				replacement := fmt.Sprintf("adamic_cell *%s = frame->cells[%d].value.reference; if (%s == NULL) { %s = adamic_cell_new((adamic_value){.%s = %s}, %s); frame->cells[%d].value.reference = %s; } else { %s %s->value.%s = %s; %s->ready = true; }", fresh, slot, fresh, fresh, member, value, refs, slot, fresh, drop, fresh, member, value, fresh)
				code = strings.Replace(code, match[0], replacement, 1)
				changed = true
				break
			}
		}
	}
	if !changed {
		t.Fatal("mutant did not change its input")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if kind == "release" {
		if got.exitCode == 0 || !strings.Contains(string(got.stderr), "LeakSanitizer: detected memory leaks") {
			t.Fatalf("frame release mutant survived: %+v", got)
		}
	} else {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("only Node stdout must catch %s: %+v", kind, got)
		}
		if kind != "capture" {
			if actual := onJavaScriptBackend(t, program); actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(truth, actual) != "stdout differs" {
				t.Fatalf("JavaScript mutant survived: %+v", actual)
			}
		}
	}
	t.Logf("caught %s: Node %q; native %q; exit %d", kind, truth.stdout, got.stdout, got.exitCode)
}
func TestAsyncLoopsTopCatchBindingMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_top_catch_binding", "release")
}
func TestAsyncLoopsTopCatchAfterAwaitMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_top_catch_after_await", "release")
}
func TestAsyncLoopsTopLoopMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_top_loop", "top-iteration")
}
func TestAsyncLoopsForOfBodyMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_for_of_body", "iteration")
}
func TestAsyncLoopsForOfMutationMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_for_of_mutation", "iteration")
}
func TestAsyncLoopsForOfReturnMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_for_of_return", "iteration")
}
func TestAsyncLoopsSwitchBodyMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_switch_body", "switch")
}
func TestAsyncLoopsLoopCellsMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_loop_cells", "copy")
}
func TestAsyncLoopsLoopBodyCaptureMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_loop_body_capture", "capture")
}
func TestAsyncLoopsForOfCaptureMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_for_of_capture", "capture")
}
func TestAsyncLoopsForOfCellsMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_for_of_cells", "capture")
}
func TestAsyncLoopsForOfBodyCellsMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_for_of_body_cells", "capture")
}
func TestAsyncLoopsForOfBreakCellsMutant(t *testing.T) {
	t.Parallel()
	asyncLoopMutation(t, "async_for_of_break_cells", "capture")
}
