package oracle

import (
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
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
	target := ""
	for index, function := range program.Functions {
		if function.Name == "local" {
			target = fmt.Sprintf("adamic_function_%d_local", index)
		}
	}
	if target == "" {
		t.Fatal("missing local callable")
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
		{"async_refuse_for_of_string", "async for-of over this iterable", "a\nb\n", false},
		{"async_refuse_loop_body_capture", "async per-iteration captured cells", "item0 item1 item2\n", false},
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
