package oracle

import (
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
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
		code := strings.Replace(control, "resumed.number", "(resumed.number + 1)", 1)
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("only the Node stdout check must catch the mutant: %+v", result)
		}
		t.Log("wrong resumed value caught only by Node stdout comparison")
	})
	t.Run("missing-frame-parameter-retain", func(t *testing.T) {
		code := regexp.MustCompile(`(frame->cells\[[0-9]+\].value.reference = )adamic_retain\((adamic_local_[0-9]+_seed)\)`).ReplaceAllString(control, `${1}${2}`)
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
		if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-use-after-free") {
			t.Fatalf("ASan did not catch a borrowed dynamic parameter across await: %+v", result)
		}
		t.Log("missing frame parameter retain caught by ASan heap-use-after-free")
	})
}

// The source oracle covers the uncaught throw. This internal root harness consumes that rejection
// instead of panicking, so Linux LSan can check frame cleanup even though the source exits nonzero.
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
	original := "if (root->rejected) { adamic_thrown = adamic_retain(root->value.reference); adamic_uncaught(); }"
	if !strings.Contains(code, original) {
		t.Fatal("missing root cleanup hook")
	}
	code = strings.Replace(code, original, "if (!root->rejected) return 2;", 1)
	for _, mutant := range []bool{false, true} {
		t.Run(map[bool]string{false: "control", true: "missing-local-release"}[mutant], func(t *testing.T) {
			source := code
			if mutant {
				start := strings.Index(source, "adamic_release(adamic_slot_")
				end := strings.Index(source[start:], ";") + start
				if start < 0 || end < start {
					t.Fatal("mutant did not change input")
				}
				source = source[:start] + "(void)" + source[start+len("adamic_release("):end-1] + source[end:]
			}
			binary := filepath.Join(t.TempDir(), "throw")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if mutant {
				if result.exitCode == 0 || !strings.Contains(string(result.stderr), "LeakSanitizer: detected memory leaks") {
					t.Fatalf("leak mutant survived: %+v", result)
				}
				t.Log("throw frame local release mutant caught by LeakSanitizer")
			} else if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("throw root cleanup leaked: %+v", result)
			}
		})
	}
}

// Each callable and control-flow surface must expose a wrong fulfillment payload to Node.
func TestAsyncExpandedChecksCatchMutants(t *testing.T) {
	for _, fixture := range []string{"async_control", "async_operand_order", "async_methods", "async_closures", "async_conditions", "async_catch_return"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			control := native.C(program)
			code := strings.ReplaceAll(control, "resumed.number", "(resumed.number + 1)")
			if fixture == "async_conditions" {
				code = strings.ReplaceAll(code, "resumed.boolean", "(!resumed.boolean)")
			}
			if code == control {
				t.Fatal("no numeric suspension payload to mutate")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if disagreement(onNode(t, path), result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("Node output did not exclusively catch fulfillment mutant: %+v", result)
			}
		})
	}
}

func TestAsyncOperandAndRejectionMutants(t *testing.T) {
	for _, fixture := range []string{"async_operand_order", "async_catches"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			var code string
			if fixture == "async_operand_order" {
				changed := false
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name != "run" {
						continue
					}
					positions := map[string]int{}
					for position, statement := range function.Body {
						declaration, ok := statement.(ir.Declare)
						if !ok {
							continue
						}
						call, ok := declaration.Value.(ir.Call)
						if !ok {
							continue
						}
						for _, target := range program.CallTargets(call) {
							positions[program.Functions[target].Name] = position
						}
					}
					first, hasFirst := positions["first"]
					second, hasSecond := positions["second"]
					if hasFirst && hasSecond {
						function.Body[first], function.Body[second] = function.Body[second], function.Body[first]
						changed = true
					}
				}
				if !changed {
					t.Fatal("missing earlier operand snapshot")
				}
				code = native.C(program)
			} else {
				control := native.C(program)
				code = strings.ReplaceAll(control, "if (rejected)", "if (false && rejected)")
				if code == control {
					t.Fatal("missing rejection dispatch")
				}
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if disagreement(onNode(t, path), result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("Node did not exclusively catch operand/rejection mutant: %+v", result)
			}
		})
	}
}

func TestAsyncReferenceSlotDropMutants(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_live_values.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	checked := 0
	for index, function := range program.Functions {
		if function.Name != "live" {
			continue
		}
		for slot, local := range function.FrameEnvironment {
			name := program.Locals[local].Name
			if name != "text" && name != "object" && name != "box" && name != "array" {
				continue
			}
			checked++
			t.Run(name, func(t *testing.T) {
				start := strings.Index(control, fmt.Sprintf("static void adamic_async_finish_%d(", index))
				end := start + strings.Index(control[start:], "\n}\n")
				original := fmt.Sprintf("adamic_release(adamic_slot_%d);", slot)
				body := control[start:end]
				changed := strings.Replace(body, original, fmt.Sprintf("(void)adamic_slot_%d;", slot), 1)
				if changed == body {
					t.Fatal("missing owned reference slot drop")
				}
				code := control[:start] + changed + control[end:]
				binary := filepath.Join(t.TempDir(), "mutant")
				if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
				if result.exitCode == 0 || !strings.Contains(string(result.stderr), "LeakSanitizer: detected memory leaks") {
					t.Fatalf("reference slot drop mutant survived: %+v", result)
				}
			})
		}
	}
	if checked != 4 {
		t.Fatalf("checked %d reference kinds, want 4", checked)
	}
}

// A private async closure holds its parent's environment while both are suspended.
// Join the loop thread before LSan, so stale activation pointers cannot hide a cycle.
func TestAsyncGeneratedAbandonment(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_escaped_capture.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	start := strings.Index(control, "int main(")
	if start < 0 {
		t.Fatal("missing main")
	}
	control = control[:start]
	control = strings.Replace(control, "#include \"async.h\"", "#include \"async.h\"\n#include <pthread.h>\nstatic adamic_async_promise *adamic_test_pending;", 1)
	settle := regexp.MustCompile(`adamic_async_settle\((adamic_temporary_[0-9]+), \(adamic_value\)\{\.number = 0\}, false, false\);`)
	changed := settle.ReplaceAllString(control, "adamic_test_pending = $1;")
	if changed == control {
		t.Fatal("missing settled void await")
	}
	control = changed
	localCallable := false
	for _, function := range program.Functions {
		localCallable = localCallable || function.Name == "local"
	}
	if !localCallable {
		t.Fatal("missing local callable")
	}
	// The harness calls the emitted callable, whose identity is now module-qualified.
	symbols := regexp.MustCompile(`\badamic_function_[A-Za-z0-9_]*_local_[0-9a-f]{32}\b`).FindAllString(control, -1)
	target := ""
	for _, symbol := range symbols {
		if target != "" && target != symbol {
			t.Fatal("ambiguous local callable symbol")
		}
		target = symbol
	}
	if target == "" {
		t.Fatal("missing emitted local callable symbol")
	}
	for _, mode := range []string{"cancel", "exit"} {
		for _, mutant := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/mutant=%t", mode, mutant), func(t *testing.T) {
				code := control
				if mutant {
					callback := regexp.MustCompile(`(static void adamic_async_abandon_[0-9]+\(adamic_async_frame \*base\) \{ )adamic_async_finish_[0-9]+\(\(void \*\)base\);`)
					code = callback.ReplaceAllString(code, "${1}(void)base;")
					if code == control {
						t.Fatal("missing abandonment detach callback")
					}
				}
				action := ""
				if mode == "cancel" {
					action = "adamic_async_cancel(adamic_test_pending);"
				}
				code += fmt.Sprintf(`
static void *adamic_test_loop(void *unused) {
    (void)unused;
    adamic_string *seed = adamic_string_from_number(12345);
    adamic_async_promise *root = %s(seed);
    adamic_release(seed);
    if (!adamic_test_pending) return (void *)1;
    %s
    adamic_test_pending = NULL;
    adamic_release(root);
    adamic_async_run();
    adamic_heap_thread_end();
    return NULL;
}
int main(void) {
    pthread_t thread;
    if (pthread_create(&thread, NULL, adamic_test_loop, NULL) != 0) return 2;
    void *result;
    if (pthread_join(thread, &result) != 0 || result) return 2;
    return 0;
}
`, target, action)
				binary := filepath.Join(t.TempDir(), "abandon")
				if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
				if mutant {
					if result.exitCode == 0 || !strings.Contains(string(result.stderr), "LeakSanitizer: detected memory leaks") {
						t.Fatalf("abandonment detach mutant survived: %+v", result)
					}
				} else if result.exitCode != 0 || len(result.stderr) != 0 {
					t.Fatalf("generated abandonment leaked: %+v", result)
				}
			})
		}
	}
}

func TestAsyncCapturedBindingSurvivesFinishMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_escaped_capture.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	code := control
	for index, function := range program.Functions {
		if function.Name != "make" {
			continue
		}
		for slot, local := range function.FrameEnvironment {
			if program.Locals[local].Name != "held" || !program.Locals[local].Captured {
				continue
			}
			marker := fmt.Sprintf("static void adamic_async_finish_%d(adamic_generated_frame_%d *frame) {\n", index, index)
			detach := fmt.Sprintf(" void *adamic_test_captured = frame->cells[%d].value.reference; frame->cells[%d].value.reference = NULL; adamic_release(adamic_test_captured);\n", slot, slot)
			code = strings.Replace(code, marker, marker+detach, 1)
		}
	}
	if code == control {
		t.Fatal("no escaped captured binding to clear")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if result.exitCode == 0 || (!strings.Contains(string(result.stderr), "runtime error:") && !strings.Contains(string(result.stderr), "AddressSanitizer")) {
		t.Fatalf("cleared captured binding survived sanitizer: %+v", result)
	}
}

func TestAsyncInstanceAndArrowPayloadMutants(t *testing.T) {
	for _, fixture := range []string{"async_methods", "async_closures"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			constant := ir.StringConstant{Index: len(program.Strings)}
			program.Strings = append(program.Strings, "wrong captured payload")
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				selected := function.Async && !function.Closure && function.AsyncReturns == ir.String
				if fixture == "async_closures" {
					selected = function.Async && function.Closure && function.AsyncReturns == ir.String
				}
				if !selected {
					continue
				}
				for position, statement := range function.Body {
					if returned, ok := statement.(ir.Return); ok {
						returned.Value = constant
						function.Body[position] = returned
						changed = true
					}
				}
			}
			if !changed {
				t.Fatal("missing instance method or arrow return")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if disagreement(onNode(t, path), result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("Node did not exclusively catch instance/arrow payload mutant: %+v", result)
			}
		})
	}
}

// Mutate a shared ordinary typeof result so both backends can agree with each other and still be
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
	object := len(program.Strings)
	program.Strings = append(program.Strings, "object")
	for _, function := range program.Functions {
		for index, statement := range function.Body {
			declaration, ok := statement.(ir.Declare)
			if !ok {
				continue
			}
			observation, ok := declaration.Value.(ir.TypeOf)
			if ok && observation.Value.Type() == ir.Closure {
				declaration.Value = ir.StringConstant{Index: object}
				function.Body[index] = declaration
				changed = true
			}
		}
	}
	if !changed {
		t.Fatal("mutant did not change an ordinary async typeof result")
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

// Refused programs still have a source oracle. In particular the frame cycle
// must not depend on the two closures having the same callable signature.
func TestAsyncReaderProbes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, want, output string
		cycle              bool
	}{
		{"async_refuse_frame_capture_cycle", "async frame capture cycle", "seed3:held:7\n", true},
		{"async_refuse_return_thenable", "return of thenables", "7\n", false},
		{"async_refuse_arrow_thenable", "return of thenables", "7\n", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_refused", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || len(observed.stderr) != 0 || string(observed.stdout) != probe.output {
				t.Fatalf("Node witness: %+v", observed)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("named refusal: %v", err)
			}
			var refused *lower.Refused
			var notYet *lower.NotYet
			if probe.cycle {
				if !errors.As(err, &refused) {
					t.Fatalf("cycle must be Refused: %v", err)
				}
			} else if !errors.As(err, &notYet) {
				t.Fatalf("unproved semantics must be NotYet: %v", err)
			}
		})
	}
}

// Restoring preservation of the private closure slot creates exactly the
// frame -> closure -> frame graph. It still prints Node's answer; only the
// common leak check may catch it, not a build failure or an output check.
func TestAsyncFrameCapturedClosureCycleMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_frame_capture_safe.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	control := native.C(program)
	changed := false
	for index := range program.Locals {
		local := &program.Locals[index]
		if local.Name == "read" && local.Type == ir.Closure && !local.Captured {
			local.Captured = true
			changed = true
		}
	}
	if !changed {
		t.Fatal("missing private read closure slot")
	}
	mutant := native.C(program)
	if mutant == control {
		t.Fatal("mutant did not preserve the closure in its frame")
	}
	for _, test := range []struct {
		name, code string
		leaks      bool
	}{
		{"control", control, false}, {"preserved-closure-slot", mutant, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "frame-cycle")
			if err := native.Build(test.code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			observed := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
			if difference := disagreement(oracle, observed); difference != "" {
				t.Fatalf("mutant must match Node before leak check: %s: %+v", difference, observed)
			}
			report := leakcheck.Report(t, test.code, binary)
			if test.leaks {
				if !strings.Contains(report, "LeakSanitizer: detected memory leaks") && !strings.Contains(report, "allocations") {
					t.Fatalf("common leak check missed frame capture cycle: %s", report)
				}
				t.Logf("frame capture cycle caught only by leak check: %s", report)
			} else if report != "" {
				t.Fatal(report)
			}
		})
	}
}

// Keeping a header binding in one cell preserves memory safety but breaks JavaScript identity.
func TestAsyncLoopCellsCatchSharedCell(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_loop_cells.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	pattern := regexp.MustCompile(`adamic_cell \*(adamic_temporary_[0-9]+) = adamic_cell_new\(\(adamic_value\)\{\.number = \(\(adamic_cell \*\)frame->cells\[([0-9]+)\]\.value\.reference\)->value\.number\}, false\);\n\s*adamic_release\(frame->cells\[[0-9]+\]\.value\.reference\);\n\s*frame->cells\[[0-9]+\]\.value\.reference = [^;]+;`)
	code := pattern.ReplaceAllString(control, `/* shared iteration cell */`)
	if code == control {
		t.Fatal("mutant did not change input")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if disagreement(onNode(t, path), result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("Node output must catch shared-cell mutant: %+v", result)
	}
}

// Skipping an index at the continuation edge still produces clean, valid C.
// Only the source on Node can catch resuming the loop on the wrong iteration.
func TestAsyncForOfWrongIterationMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_for_of_body.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, function := range program.Functions {
		for index, statement := range function.Body {
			loop, ok := statement.(ir.Loop)
			if !ok || len(loop.Update) != 1 {
				continue
			}
			update, ok := loop.Update[0].(ir.Assign)
			if !ok || program.Locals[update.Local].Name != "suspension" {
				continue
			}
			addition, ok := update.Value.(ir.Binary)
			if !ok || addition.Operator != ir.Add {
				continue
			}
			addition.Right = ir.NumberConstant{Value: 2}
			update.Value = addition
			loop.Update[0] = update
			function.Body[index] = loop
			changed = true
		}
	}
	if !changed {
		t.Fatal("no array iterator continuation to mutate")
	}
	binary := filepath.Join(t.TempDir(), "wrong-iteration")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	oracle := onNode(t, path)
	if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("only Node stdout must catch wrong iteration: %+v", result)
	}
	t.Logf("caught: Node %q; wrong iteration %q", oracle.stdout, result.stdout)
}

// Advance the held iterator once more after the suspended body resumes. The binary must
// remain clean; the source oracle must catch its skipped code point or entry.
func TestAsyncBuiltinIteratorWrongEntryMutant(t *testing.T) {
	for _, name := range []string{"async_for_of_string", "async_for_of_map", "async_for_of_collections_mutation"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			var mutate func([]ir.Statement)
			mutate = func(body []ir.Statement) {
				for index, statement := range body {
					switch value := statement.(type) {
					case ir.Loop:
						if len(value.Body) > 0 {
							if declaration, ok := value.Body[0].(ir.Declare); ok {
								if call, ok := declaration.Value.(ir.CallClosure); ok {
									value.Update = append(value.Update, ir.Evaluate{Value: call})
									body[index] = value
									changed = true
								}
							}
						}
					case ir.Block:
						mutate(value.Body)
					}
				}
			}
			for _, function := range program.Functions {
				mutate(function.Body)
			}
			if !changed {
				t.Fatal("no held iterator step to mutate")
			}
			binary := filepath.Join(t.TempDir(), "wrong-entry")
			code := native.C(program)
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			if report := leakcheck.Report(t, code, binary); report != "" {
				t.Fatal(report)
			}
			oracle := onNode(t, path)
			for name, result := range map[string]run{"native": execute(t, binary), "JavaScript": onJavaScriptBackend(t, program)} {
				if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
					t.Fatalf("only Node stdout must catch %s wrong entry: %+v", name, result)
				}
				t.Logf("Node catches %s wrong entry", name)
			}
		})
	}
}

func TestAsyncFinallyPendingValueMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_finally_completion.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutateFallthroughStatements(reflect.ValueOf(&program.Functions).Elem(), func(statement ir.Statement) ir.Statement {
		assignment, ok := statement.(ir.Assign)
		if ok && program.Locals[assignment.Local].Name == "pending return" && assignment.Value.Type() == ir.String {
			assignment.Value = ir.StringConstant{Index: len(program.Strings)}
			program.Strings = append(program.Strings, "dropped")
			changed++
			return assignment
		}
		return statement
	})
	if changed == 0 {
		t.Fatal("no pending payload mutated")
	}
	observed, sanitized := natively(t, program)
	if observed.exitCode != 0 || len(observed.stderr) != 0 {
		t.Fatalf("mutant must run cleanly: %+v", observed)
	}
	if difference := disagreement(onNode(t, path), observed); difference != "stdout differs" {
		t.Fatalf("pending-value mutant survived: %s", difference)
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatalf("mutant leaks: %s", report)
	}
	t.Log("dropped pending return payload caught by Node stdout; sanitizer and leak checks clean")
}

func TestAsyncLabelInnerTargetMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_labels.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutateFallthroughStatements(reflect.ValueOf(&program.Functions).Elem(), func(statement ir.Statement) ir.Statement {
		jump, ok := statement.(ir.Continue)
		if ok && jump.Depth > 0 {
			jump.Depth = 0
			changed++
			return jump
		}
		return statement
	})
	if changed == 0 {
		t.Fatal("no outer continue mutated")
	}
	observed, sanitized := natively(t, program)
	if observed.exitCode != 0 || len(observed.stderr) != 0 {
		t.Fatalf("mutant must run cleanly: %+v", observed)
	}
	if difference := disagreement(onNode(t, path), observed); difference != "stdout differs" {
		t.Fatalf("inner-label mutant survived: %s", difference)
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatalf("mutant leaks: %s", report)
	}
	t.Log("inner continue target caught by Node stdout; sanitizer and leak checks clean")
}

func TestAsyncFinallyInnerLabelMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_labels_finally.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(string(source), "continue outer;", "continue inner;", 1)
	if mutant == string(source) {
		t.Fatal("no labeled target mutated")
	}
	mutantPath := filepath.Join(t.TempDir(), "inner-label.a")
	if err := os.WriteFile(mutantPath, []byte(mutant), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, mutantPath)
	if err != nil {
		t.Fatal(err)
	}
	observed, sanitized := natively(t, program)
	if observed.exitCode != 0 || len(observed.stderr) != 0 {
		t.Fatalf("mutant must run cleanly: %+v", observed)
	}
	if difference := disagreement(onNode(t, path), observed); difference != "stdout differs" {
		t.Fatalf("inner-label mutant survived: %s", difference)
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatalf("mutant leaks: %s", report)
	}
	t.Log("source jump to inner label caught by Node stdout; sanitizer and leak checks clean")
}

// Source unlabeled break must skip the synthetic default-only block switch.
// Change only that jump; label breaks to the wrapper must retain their target.
func TestLabelBlockUnlabeledBreakMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/label_block_edges.a"))
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	for _, name := range []string{"syncBreak", "asyncBreak"} {
		t.Run(name, func(t *testing.T) {
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			for index := range program.Functions {
				if program.Functions[index].Name != name {
					continue
				}
				mutateFallthroughStatements(reflect.ValueOf(&program.Functions[index].Body).Elem(), func(statement ir.Statement) ir.Statement {
					jump, ok := statement.(ir.Break)
					if ok && jump.Depth == 1 {
						jump.Depth = 0
						changed++
						return jump
					}
					return statement
				})
			}
			if changed != 1 {
				t.Fatalf("want exactly one unlabeled break mutated, got %d", changed)
			}
			backend := onJavaScriptBackend(t, program)
			if backend.exitCode != 0 || len(backend.stderr) != 0 || disagreement(oracle, backend) != "stdout differs" {
				t.Fatalf("JavaScript block-target mutant survived or failed uncleanly: %+v", backend)
			}
			observed, sanitized := natively(t, program)
			if observed.exitCode != 0 || len(observed.stderr) != 0 || disagreement(oracle, observed) != "stdout differs" {
				t.Fatalf("native block-target mutant survived or failed uncleanly: %+v", observed)
			}
			if report := leaks(t, program, sanitized); report != "" {
				t.Fatalf("mutant leaks: %s", report)
			}
			t.Log("unlabeled break to block wrapper caught by Node in both backends; sanitizer and leak checks clean")
		})
	}
}

// Reusing the numeric binding is memory safe; Node must catch the lost identity.
func TestAsyncForOfCellsCatchSharedCell(t *testing.T) {
	for _, fixture := range []string{"async_for_of_cells", "async_for_of_body_cells", "async_for_of_break_cells"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			control := native.C(program)
			pattern := regexp.MustCompile(`adamic_cell \*(adamic_temporary_[0-9]+) = adamic_cell_new\(\(adamic_value\)\{\.number = ([^\n]+)\}, false\);\n\s*adamic_release\(frame->cells\[([0-9]+)\]\.value\.reference\);\n\s*frame->cells\[[0-9]+\]\.value\.reference = [^;]+;`)
			code := pattern.ReplaceAllString(control, `adamic_cell *$1 = adamic_cell_new((adamic_value){.number = $2}, false);
if (frame->cells[$3].value.reference != NULL) {
 ((adamic_cell *)frame->cells[$3].value.reference)->value.number = $1->value.number;
 adamic_release($1);
} else { frame->cells[$3].value.reference = $1; }`)
			if code == control {
				t.Fatal("mutant did not share an iteration cell")
			}
			binary := filepath.Join(t.TempDir(), "shared-cell")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			if report := leakcheck.Report(t, code, binary); report != "" {
				t.Fatal(report)
			}
			result := execute(t, binary)
			oracle := onNode(t, path)
			if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("Node stdout must catch shared-cell mutant: %+v", result)
			}
			t.Logf("caught: Node %q; shared cells %q", oracle.stdout, result.stdout)
		})
	}
}
