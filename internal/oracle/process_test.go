package oracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func processExitOutput(t *testing.T, destination string, shared, backpressure bool, command ...string) run {
	t.Helper()
	arguments := []string{filepath.Join(repository, "internal/oracle/testdata/process_output.py"), destination, fmt.Sprint(shared), fmt.Sprint(backpressure)}
	result := execute(t, "python3", append(arguments, command...)...)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("output driver: %s", result.stderr)
	}
	var observed struct {
		Stdout, Stderr         []byte
		ExitCode, PipeCapacity int
	}
	if err := json.Unmarshal(result.stdout, &observed); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s shared=%t backpressure=%t: stdout=%d stderr=%d exit=%d pipe capacity=%d", filepath.Base(command[0]), shared, backpressure, len(observed.Stdout), len(observed.Stderr), observed.ExitCode, observed.PipeCapacity)
	return run{stdout: observed.Stdout, stderr: observed.Stderr, exitCode: observed.ExitCode}
}

// Exercise queued output independently of the shared panic runtime's blocking defaults.
// Module imports run before these statements, so both the control and its exit mutant
// use ordinary asynchronous Node pipes while retaining the real generated exit helper.
func processAsyncScript(t *testing.T, script string) string {
	t.Helper()
	data, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "async.mjs")
	prefix := "for (const stream of [process.stdout, process.stderr]) stream._handle?.setBlocking(false);\n"
	if err := os.WriteFile(path, append([]byte(prefix), data...), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// Runtime mutations are included in the generated translation unit with renamed public
// symbols. This exercises the real output buffer, without editing the checkout or its cache.
func processOutputProgram(t *testing.T, source, mutation string) (string, string, string) {
	return processOutputProgramWith(t, source, mutation, native.Options{Sanitize: true})
}

func processOutputProgramWith(t *testing.T, source, mutation string, options native.Options) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte("import 'node:process';\n"+source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	if mutation != "" {
		data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/adamic.c"))
		if err != nil {
			t.Fatal(err)
		}
		runtimeSource := string(data)
		before := "_Noreturn void adamic_process_exit_now(int code) {\n\tpthread_mutex_lock(&output_lock);\n\tflush();"
		after := "_Noreturn void adamic_process_exit_now(int code) {\n\tpthread_mutex_lock(&output_lock);"
		if mutation == "stderr" {
			before = "// Whatever stdout holds was written first, and stderr may be the same file.\n\t\tflush();"
			after = "// Mutant: stderr overtakes buffered stdout."
		}
		if strings.Count(runtimeSource, before) != 1 {
			t.Fatal("flush mutant anchor changed")
		}
		runtimeSource = strings.Replace(runtimeSource, before, after, 1)
		var declarations strings.Builder
		for _, name := range []string{"adamic_start", "adamic_write_line", "adamic_write_raw", "adamic_output_flush", "adamic_panic", "adamic_unreachable", "adamic_process_exit_now"} {
			fmt.Fprintf(&declarations, "#define %s mutant_%s\n", name, name)
			runtimeSource = strings.ReplaceAll(runtimeSource, name, "mutant_"+name)
			code = strings.ReplaceAll(code, name, "mutant_"+name)
		}
		code = strings.ReplaceAll(code, "adamic_process_exit(", "mutant_exit(")
		code = declarations.String() + runtimeSource + "\nstatic void mutant_exit(adamic_maybe_number code) { adamic_process_set_exit_code(code); mutant_adamic_process_exit_now(adamic_process_status()); }\n" + code
	}
	binary := filepath.Join(directory, "program")
	if err := native.Build(code, binary, options); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, binary, script
}

func TestProcessExitOutput_000(t *testing.T) { t.Parallel(); processExitOutputCase(t, 0) }
func TestProcessExitOutput_001(t *testing.T) { t.Parallel(); processExitOutputCase(t, 1) }
func TestProcessExitOutput_002(t *testing.T) { t.Parallel(); processExitOutputCase(t, 2) }
func TestProcessExitOutput_003(t *testing.T) { t.Parallel(); processExitOutputCase(t, 3) }
func TestProcessExitOutput_004(t *testing.T) { t.Parallel(); processExitOutputCase(t, 4) }

func processExitOutputCase(t *testing.T, shard int) {
	wasm := shard >= 5
	if wasm {
		shard -= 5
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			t.Skip("set ADAMIC_ORACLE_WASI=1")
		}
	}
	if runtime.GOOS != "linux" {
		t.Skip("Linux descriptor and Node pipe semantics")
	}
	large := "const line = 'x'.repeat(1023); for (let index = 0; index < 200; index += 1) { console.log(line); } process.exit(37);"
	for index, one := range []struct {
		name, source, destination, expected, mutant string
		shared                                      bool
	}{
		{"file immediate", "console.log('before chosen exit'); process.exit(37);", "file", "before chosen exit\n", "exit", false},
		{"file normal", "process.exitCode = 37; console.log('before normal return');", "file", "before normal return\n", "", false},
		{"merged pipe", "console.log('out 0'); console.error('err 0'); console.log('out 1'); console.error('err 1'); console.log('out last'); process.exit(37);", "pipe", "out 0\nerr 0\nout 1\nerr 1\nout last\n", "stderr", true},
		{"merged file", "console.log('out 0'); console.error('err 0'); console.log('out 1'); console.error('err 1'); console.log('out last'); process.exit(37);", "file", "out 0\nerr 0\nout 1\nerr 1\nout last\n", "stderr", true},
		{"large file", large, "file", strings.Repeat(strings.Repeat("x", 1023)+"\n", 200), "exit", false},
	} {
		if index != shard {
			continue
		}
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			options := native.Options{Sanitize: true}
			if wasm {
				options = native.Options{Target: "wasm32-wasi"}
			}
			path, binary, script := processOutputProgramWith(t, one.source, "", options)
			runner := filepath.Join(repository, "oracle/node.mjs")
			truth := processExitOutput(t, one.destination, one.shared, false, "node", "--disable-warning=ExperimentalWarning", runner, path)
			if truth.exitCode != 37 || string(truth.stdout) != one.expected || len(truth.stderr) != 0 {
				t.Fatal("Node did not exercise the expected output path")
			}
			if wasm {
				got := processExitOutput(t, one.destination, one.shared, false, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
				return
			}
			for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
				if difference := disagreement(truth, processExitOutput(t, one.destination, one.shared, false, command...)); difference != "" {
					t.Fatal(difference)
				}
			}
			if one.mutant != "" {
				mutationSource, destination, mutationTruth := one.source, one.destination, truth
				if destination == "file" {
					// Main writes regular files immediately. Keep that full control above,
					// and exercise the omitted flush where a buffer exists: a pipe.
					destination = "pipe"
					if one.mutant == "exit" {
						// Stay below pipe capacity so Node's immediate exit has no queued tail.
						mutationSource = "console.log('buffered exit witness'); process.exit(37);"
					}
					witness, control, _ := processOutputProgram(t, mutationSource, "")
					mutationTruth = processExitOutput(t, destination, one.shared, false, "node", "--disable-warning=ExperimentalWarning", runner, witness)
					if difference := disagreement(mutationTruth, processExitOutput(t, destination, one.shared, false, control)); difference != "" {
						t.Fatalf("buffered flush control: %s", difference)
					}
				}
				_, mutant, _ := processOutputProgram(t, mutationSource, one.mutant)
				got := processExitOutput(t, destination, one.shared, false, mutant)
				if got.exitCode != 37 || len(got.stderr) != 0 {
					t.Fatal("mutant failed outside byte comparison")
				}
				if disagreement(mutationTruth, got) != "stdout differs" {
					t.Fatal("flush mutant survived")
				}
				if one.shared {
					t.Logf("Node order %q; mutant order %q", mutationTruth.stdout, got.stdout)
				}
				t.Log("mutant compiled and ran cleanly; only Node output bytes caught it")
			}
		})
	}
}

// Node may deliver all output or lose a suffix on immediate exit. Adamic must
// deliver every byte on both backends regardless of the host's pipe behavior.
func TestProcessLargePipeExitPreservesOutput_000(t *testing.T) {
	t.Parallel()
	processLargePipeExit(t, 0)
}
func TestProcessLargePipeExitPreservesOutput_001(t *testing.T) {
	t.Parallel()
	processLargePipeExit(t, 1)
}

func processLargePipeExit(t *testing.T, shard int) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux backpressured pipe semantics")
	}
	expected := []byte(strings.Repeat(strings.Repeat("x", 1023)+"\n", 200))
	runner := filepath.Join(repository, "oracle/node.mjs")
	for index, immediate := range []bool{true, false} {
		if index != shard {
			continue
		}
		t.Run(fmt.Sprintf("immediate=%t", immediate), func(t *testing.T) {
			ending := "process.exitCode = 37;"
			if immediate {
				ending = "process.exit(37);"
			}
			path, binary, script := processOutputProgram(t, "const line = 'x'.repeat(1023); for (let index = 0; index < 200; index += 1) { console.log(line); } "+ending, "")
			script = processAsyncScript(t, script)
			truth := processExitOutput(t, "pipe", false, true, "node", "--disable-warning=ExperimentalWarning", runner, processAsyncScript(t, path))
			got := processExitOutput(t, "pipe", false, true, binary)
			backend := processExitOutput(t, "pipe", false, true, "node", "--disable-warning=ExperimentalWarning", runner, script)
			if truth.exitCode != 37 || got.exitCode != 37 || len(truth.stderr) != 0 || len(got.stderr) != 0 || !bytes.Equal(got.stdout, expected) {
				t.Fatal("unexpected exit, sanitizer failure, or incomplete native output")
			}
			if !immediate {
				if disagreement(truth, got) != "" || disagreement(truth, backend) != "" {
					t.Fatal("normal return must drain output byte for byte")
				}
			} else {
				if !bytes.HasPrefix(expected, truth.stdout) {
					t.Fatal("raw Node output is not a prefix of the writes")
				}
				if difference := disagreement(got, backend); difference != "" {
					t.Fatalf("Adamic backends must preserve all output: %s", difference)
				}
				t.Logf("Node delivered %d/%d bytes (full=%t); both Adamic backends preserve all bytes", len(truth.stdout), len(expected), bytes.Equal(truth.stdout, expected))
				data, err := os.ReadFile(script)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Count(string(data), "adamicProcessExit(") != 1 {
					t.Fatal("JS exit mutant anchor changed")
				}
				mutant := filepath.Join(t.TempDir(), "mutant.mjs")
				if err := os.WriteFile(mutant, []byte(strings.Replace(string(data), "adamicProcessExit(", "process.exit(", 1)), 0o644); err != nil {
					t.Fatal(err)
				}
				control := filepath.Join(t.TempDir(), "control.mjs")
				original, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(control, []byte(strings.Replace(string(original), ending, "process.exitCode = 37;", 1)), 0600); err != nil {
					t.Fatal(err)
				}
				drained := processExitOutput(t, "pipe", false, true, "node", "--disable-warning=ExperimentalWarning", runner, processAsyncScript(t, control))
				if difference := disagreement(drained, got); difference != "" {
					t.Fatal("Node normal-exit drain control: " + difference)
				}
				bad := processExitOutput(t, "pipe", false, true, "node", "--disable-warning=ExperimentalWarning", runner, mutant)
				if bad.exitCode != 37 || len(bad.stderr) != 0 || disagreement(drained, bad) != "stdout differs" {
					t.Fatal("JS immediate-exit drain mutant survived, or failed outside output comparison")
				}
				t.Log("JS immediate-exit drain mutant runs cleanly with exit 37; only the Node normal-exit output comparison catches it")
			}
		})
	}
}

func TestProcessExitDrainsStderr(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux backpressured pipe semantics")
	}
	expected := []byte(strings.Repeat(strings.Repeat("x", 1023)+"\n", 200))
	path, binary, script := processOutputProgram(t, "const line = 'x'.repeat(1023); for (let index = 0; index < 200; index += 1) { console.error(line); } process.exit(37);", "")
	runner := filepath.Join(repository, "oracle/node.mjs")
	truth := processExitOutput(t, "stderr-pipe", false, true, "node", "--disable-warning=ExperimentalWarning", runner, path)
	if truth.exitCode != 37 || len(truth.stdout) != 0 || !bytes.HasPrefix(expected, truth.stderr) {
		t.Fatal("raw Node stderr must be a prefix with the requested exit status")
	}
	t.Logf("Node delivered %d/%d stderr bytes (full=%t)", len(truth.stderr), len(expected), bytes.Equal(truth.stderr, expected))
	for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
		got := processExitOutput(t, "stderr-pipe", false, true, command...)
		if got.exitCode != 37 || len(got.stdout) != 0 || !bytes.Equal(got.stderr, expected) {
			t.Fatal("Adamic must drain stderr too")
		}
	}
}

func TestProcessExitDoesNotUnwind_000(t *testing.T) { t.Parallel(); processExitDoesNotUnwind(t, false) }
func TestProcessExitDoesNotUnwind_001(t *testing.T) { t.Parallel(); processExitDoesNotUnwind(t, true) }

func processExitDoesNotUnwind(t *testing.T, wasm bool) {
	options := native.Options{Sanitize: true}
	if wasm {
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			t.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		options = native.Options{Target: "wasm32-wasi"}
	}
	path, binary, script := processOutputProgramWith(t, "function end(): void { try { console.log('before'); process.exit(37); } catch { console.log('must not catch'); } finally { console.log('must not finally'); } console.log('must not continue'); } end(); console.log('must not return');", "", options)
	runner := filepath.Join(repository, "oracle/node.mjs")
	truth := processExitOutput(t, "file", false, false, "node", "--disable-warning=ExperimentalWarning", runner, path)
	if truth.exitCode != 37 || string(truth.stdout) != "before\n" || len(truth.stderr) != 0 {
		t.Fatal("unexpected Node exit control flow")
	}
	if wasm {
		got := processExitOutput(t, "file", false, false, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
		if difference := disagreement(truth, got); difference != "" {
			t.Fatal(difference)
		}
		return
	}
	for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
		if difference := disagreement(truth, processExitOutput(t, "file", false, false, command...)); difference != "" {
			t.Fatal(difference)
		}
	}
	data, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	instrumented := javascript.JavaScriptWith(program, javascript.Options{
		Enter: func(int) string { return "undefined" },
		Leave: func(int) string { return "console.log('must not leave')" },
	})
	observed := filepath.Join(t.TempDir(), "instrumented.mjs")
	if err := os.WriteFile(observed, []byte(instrumented), 0o644); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(truth, processExitOutput(t, "file", false, false, "node", "--disable-warning=ExperimentalWarning", runner, observed)); difference != "" {
		t.Fatal("exit must also skip instrumentation leave: " + difference)
	}
	for _, change := range []struct{ name, before, after, source string }{
		{"catch guard", "if (adamicProcessExiting()) throw ", "if (false) throw ", string(data)},
		{"finally guard", "if (!adamicProcessExiting()) {", "if (true) {", string(data)},
		{"leave guard", "if (!adamicProcessExiting()) { console.log('must not leave'); }", "if (true) { console.log('must not leave'); }", instrumented},
	} {
		t.Run(change.name, func(t *testing.T) {
			if strings.Count(change.source, change.before) != 1 {
				t.Fatal("exit control-flow mutant anchor changed")
			}
			mutant := filepath.Join(t.TempDir(), "mutant.mjs")
			if err := os.WriteFile(mutant, []byte(strings.Replace(change.source, change.before, change.after, 1)), 0o644); err != nil {
				t.Fatal(err)
			}
			bad := processExitOutput(t, "file", false, false, "node", "--disable-warning=ExperimentalWarning", runner, mutant)
			if bad.exitCode != 37 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				t.Fatal("control-flow mutant must fail only stdout comparison")
			}
			t.Logf("caught clean JS mutant: %q instead of %q", bad.stdout, truth.stdout)
		})
	}
}

func TestProcessExitArguments_000(t *testing.T) { t.Parallel(); processExitArgument(t, 0) }
func TestProcessExitArguments_001(t *testing.T) { t.Parallel(); processExitArgument(t, 1) }
func TestProcessExitArguments_002(t *testing.T) { t.Parallel(); processExitArgument(t, 2) }
func TestProcessExitArguments_003(t *testing.T) { t.Parallel(); processExitArgument(t, 3) }
func TestProcessExitArguments_004(t *testing.T) { t.Parallel(); processExitArgument(t, 4) }
func TestProcessExitArguments_005(t *testing.T) { t.Parallel(); processExitArgument(t, 5) }
func TestProcessExitArguments_006(t *testing.T) { t.Parallel(); processExitArgument(t, 6) }
func TestProcessExitArguments_007(t *testing.T) { t.Parallel(); processExitArgument(t, 7) }
func TestProcessExitArguments_008(t *testing.T) { t.Parallel(); processExitArgument(t, 8) }
func TestProcessExitArguments_009(t *testing.T) { t.Parallel(); processExitArgument(t, 9) }
func TestProcessExitArguments_010(t *testing.T) { t.Parallel(); processExitArgument(t, 10) }
func TestProcessExitArguments_011(t *testing.T) { t.Parallel(); processExitArgument(t, 11) }
func TestProcessExitArguments_012(t *testing.T) { t.Parallel(); processExitArgument(t, 12) }
func TestProcessExitArguments_013(t *testing.T) { t.Parallel(); processExitArgument(t, 13) }
func TestProcessExitArguments_014(t *testing.T) { t.Parallel(); processExitArgument(t, 14) }
func TestProcessExitArguments_015(t *testing.T) { t.Parallel(); processExitArgument(t, 15) }
func TestProcessExitArguments_016(t *testing.T) { t.Parallel(); processExitArgument(t, 16) }
func TestProcessExitArguments_017(t *testing.T) { t.Parallel(); processExitArgument(t, 17) }
func TestProcessExitArguments_018(t *testing.T) { t.Parallel(); processExitArgument(t, 18) }
func TestProcessExitArguments_019(t *testing.T) { t.Parallel(); processExitArgument(t, 19) }
func TestProcessExitArguments_020(t *testing.T) { t.Parallel(); processExitArgument(t, 20) }
func TestProcessExitArguments_021(t *testing.T) { t.Parallel(); processExitArgument(t, 21) }
func TestProcessExitArguments_022(t *testing.T) { t.Parallel(); processExitArgument(t, 22) }
func TestProcessExitArguments_023(t *testing.T) { t.Parallel(); processExitArgument(t, 23) }

func processExitArgument(t *testing.T, shard int) {
	for index, argument := range []string{"", "undefined", "0", "-1", "256", "257", "4294967297", "9007199254740991"} {
		if index != shard/3 {
			continue
		}
		t.Run("argument="+argument, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte("import 'node:process'; process.exitCode = 7; console.log('exit argument'); process.exit("+argument+");"), 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNodeWith(t, inputRun{}, path)
			var got run
			switch shard % 3 {
			case 0:
				got, _ = nativelyUncached(t, program)
			case 1:
				got = nodeProcessJavaScript(t, program)
			case 2:
				if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
					t.Skip("set ADAMIC_ORACLE_WASI=1")
				}
				got = onWASI(t, native.C(program))
			}
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%s: Node %d got %d", difference, truth.exitCode, got.exitCode)
			}
		})
	}
}

func TestProcessObjectsDoNotEscape(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"process", "process.env", "process.stdout", "process.stderr", "process.exit"} {
		t.Run(expression, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte("import 'node:process'; const held = "+expression+";"), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			var refused *lower.Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("want explicit NotYet rather than an uninitialized object, got %v", err)
			}
		})
	}
}

// Every mutant builds cleanly and reaches the ordinary Node comparison, never a sanitizer failure.
func TestProcessMutants_000(t *testing.T) { t.Parallel(); processStatusMutant(t, 0) }
func TestProcessMutants_001(t *testing.T) { t.Parallel(); processStatusMutant(t, 1) }
func TestProcessMutants_002(t *testing.T) { t.Parallel(); processStatusMutant(t, 2) }

func processStatusMutant(t *testing.T, shard int) {
	for index, one := range []struct{ name, fixture, before, after, helper, difference string }{
		{"normal status", "process_exit_code.a", "return adamic_process_status();", "return 0;", "", "exit codes differ"},
		{"immediate status", "process_exit.a", "adamic_process_exit(", "mutant_exit(", "static void mutant_exit(adamic_maybe_number code) { (void)code; adamic_process_exit((adamic_maybe_number){true, 0}); }", "exit codes differ"},
		{"range validation", "process_bad_code.a", "adamic_process_set_exit_code(", "mutant_set_code(", "static void mutant_set_code(adamic_maybe_number code) { if (!code.present || (isfinite(code.number) && trunc(code.number) == code.number)) adamic_process_set_exit_code(code); }", "stdout differs"},
	} {
		if index != shard {
			continue
		}
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", one.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			mutant := strings.ReplaceAll(original, one.before, one.after)
			if mutant == original {
				t.Fatal("mutant changed nothing")
			}
			if one.helper != "" {
				mutant = "#include \"adamic.h\"\n" + one.helper + "\n" + mutant
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			// LeakSanitizer is Linux's: macOS's AddressSanitizer aborts when asked for it.
			var environment []string
			if runtime.GOOS == "linux" {
				environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
			}
			got := executeWith(t, environment, binary)
			if bytes.Contains(got.stderr, []byte("Sanitizer")) || bytes.Contains(got.stderr, []byte("runtime error:")) {
				t.Fatalf("mutant failed outside comparison: %s", got.stderr)
			}
			if difference := disagreement(onNodeWith(t, inputRun{}, path), got); difference != one.difference {
				t.Fatalf("caught by %q, want %q: exit %d stdout %q stderr %q", difference, one.difference, got.exitCode, got.stdout, got.stderr)
			}
			t.Logf("compiled and ran cleanly; caught only by Node %s", one.difference)
		})
	}
}

func init() {
	for _, name := range []string{"process_bad_code", "process_exit", "process_exit_code", "process_exit_default", "process_exit_uncaught", "process_shadow"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}

func TestProcessExitAssignmentResult_000(t *testing.T) {
	t.Parallel()
	processExitAssignmentResult(t, 0)
}
func TestProcessExitAssignmentResult_001(t *testing.T) {
	t.Parallel()
	processExitAssignmentResult(t, 1)
}
func TestProcessExitAssignmentResult_002(t *testing.T) {
	t.Parallel()
	processExitAssignmentResult(t, 2)
}
func processExitAssignmentResult(t *testing.T, shard int) {
	path := filepath.Join(t.TempDir(), "main.a")
	source := "import 'node:process'; const assigned=(process.exitCode=4294967297); console.log(String(assigned)); console.log(String(process.exitCode)); const cleared=(process.exitCode=undefined); console.log(String(cleared===undefined)); console.log(String(process.exitCode));"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	nodeProcessRuntimeAgreement(t, program, onNodeWith(t, inputRun{}, path), shard)
}
func TestProcessExitNonnumericABI(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"process.exit('2');", "process.exitCode='2';", "process.exit(null);", "process.exitCode=null;", "process.exitCode=1; process.exitCode+=1;", "process.exit(...[2]);"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte("import 'node:process'; "+source), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want explicit numeric exit ABI refusal, got %v", err)
			}
		})
	}
}

func TestProcessExitOutput_005(t *testing.T) { t.Parallel(); processExitOutputCase(t, 5) }
func TestProcessExitOutput_006(t *testing.T) { t.Parallel(); processExitOutputCase(t, 6) }
func TestProcessExitOutput_007(t *testing.T) { t.Parallel(); processExitOutputCase(t, 7) }
func TestProcessExitOutput_008(t *testing.T) { t.Parallel(); processExitOutputCase(t, 8) }
func TestProcessExitOutput_009(t *testing.T) { t.Parallel(); processExitOutputCase(t, 9) }

func TestProcessExitWASIDrain_000(t *testing.T) { t.Parallel(); processExitWASIDrain(t, "log", "pipe") }
func TestProcessExitWASIDrain_001(t *testing.T) {
	t.Parallel()
	processExitWASIDrain(t, "error", "stderr-pipe")
}
func processExitWASIDrain(t *testing.T, method, destination string) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	source := "const line='x'.repeat(1023); for(let index=0;index<200;index+=1){console." + method + "(line);} process.exit(37);"
	path, binary, _ := processOutputProgramWith(t, source, "", native.Options{Target: "wasm32-wasi"})
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	control := filepath.Join(t.TempDir(), "control.mjs")
	if err := os.WriteFile(control, []byte(strings.Replace(string(original), "process.exit(37);", "process.exitCode=37;", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	truth := processExitOutput(t, destination, false, true, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), processAsyncScript(t, control))
	got := processExitOutput(t, destination, false, true, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	if difference := disagreement(truth, got); difference != "" {
		t.Fatal("Node drain control: " + difference)
	}
}

func TestProcessExitNormalizationMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/process_exit_code.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/process.c"))
	if err != nil {
		t.Fatal(err)
	}
	anchor := "exit_code.number = adamic_bitwise_or(code.number, 0);"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("normalization mutant anchor changed")
	}
	runtimeSource := strings.Replace(string(data), anchor, "exit_code.number = code.number;", 1)
	code := native.C(program)
	for _, name := range []string{"adamic_process_exit_code", "adamic_process_set_exit_code", "adamic_process_status", "adamic_process_exit", "adamic_process_is_tty", "adamic_process_environment"} {
		runtimeSource = strings.ReplaceAll(runtimeSource, name+"(", "mutant_"+name+"(")
		code = strings.ReplaceAll(code, name+"(", "mutant_"+name+"(")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(runtimeSource+"\n"+code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{}, path)
	bad := execute(t, binary)
	if bad.exitCode != truth.exitCode || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatalf("normalization mutant failed outside Node stdout: %+v", bad)
	}
	t.Log("unnormalized stored status ran cleanly; only Node stdout caught it")
}

func TestProcessExitValidationCode_000(t *testing.T) { t.Parallel(); processExitValidationCode(t, 0) }
func TestProcessExitValidationCode_001(t *testing.T) { t.Parallel(); processExitValidationCode(t, 1) }
func TestProcessExitValidationCode_002(t *testing.T) { t.Parallel(); processExitValidationCode(t, 2) }
func TestProcessExitValidationCode_003(t *testing.T) { t.Parallel(); processExitValidationCode(t, 3) }

// An explicit typed-IR runtime proof; the source cast is only the independent Node
// observation, not a claim that the frontend accepts arbitrary Error descriptors.
func processExitValidationCode(t *testing.T, shard int) {
	path := filepath.Join(t.TempDir(), "main.a")
	source := "try {process.exitCode=NaN;} catch(error) {console.log((error as {code:string}).code);}"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{}, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "ERR_OUT_OF_RANGE\n" {
		t.Fatalf("unexpected Node validation code: %+v", truth)
	}
	program := &ir.Program{Source: "process-exit-validation-code.a", Locals: []ir.Local{{Name: "error", Type: ir.Object, Function: -1}}}
	program.Main = []ir.Statement{ir.Try{HasCatch: true, CatchLocal: 0, Body: []ir.Statement{ir.Evaluate{Value: ir.ProcessCall{Operation: "setExitCode", Of: ir.MaybeNumber, Arguments: []ir.Expression{ir.MaybeOf{Of: ir.MaybeNumber, Value: ir.NumberConstant{Value: math.NaN()}}}}}}, Catch: []ir.Statement{ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "code", Of: ir.String}}}}}
	if shard < 3 {
		nodeProcessRuntimeAgreement(t, program, truth, shard)
		return
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/process.c"))
	if err != nil {
		t.Fatal(err)
	}
	anchor := "ADAMIC_STRING(\"ERR_OUT_OF_RANGE\")"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("validation code mutant anchor changed")
	}
	runtimeSource := strings.Replace(string(data), anchor, "ADAMIC_STRING(\"WRONG\")", 1)
	code := native.C(program)
	for _, name := range []string{"adamic_process_exit_code", "adamic_process_set_exit_code", "adamic_process_status", "adamic_process_exit", "adamic_process_is_tty", "adamic_process_environment"} {
		runtimeSource = strings.ReplaceAll(runtimeSource, name+"(", "mutant_"+name+"(")
		code = strings.ReplaceAll(code, name+"(", "mutant_"+name+"(")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(runtimeSource+"\n"+code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	bad := execute(t, binary)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatalf("error-code mutant failed outside Node stdout: %+v", bad)
	}
	t.Log("wrong range error code ran cleanly; only Node stdout caught it")
}

func TestProcessExitClosedStdout_000(t *testing.T) { t.Parallel(); processExitClosedStdout(t, 0) }
func TestProcessExitClosedStdout_001(t *testing.T) { t.Parallel(); processExitClosedStdout(t, 1) }
func TestProcessExitClosedStdout_002(t *testing.T) { t.Parallel(); processExitClosedStdout(t, 2) }
func TestProcessExitClosedStdout_003(t *testing.T) { t.Parallel(); processExitClosedStdout(t, 3) }
func TestProcessExitClosedStdout_004(t *testing.T) { t.Parallel(); processExitClosedStdout(t, 4) }
func TestProcessExitClosedStdout_005(t *testing.T) { t.Parallel(); processExitClosedStdout(t, 5) }

func processExitClosedStdout(t *testing.T, shard int) {
	immediate := shard < 3
	ending := "process.exitCode=37;"
	if immediate {
		ending = "process.exit(37);"
	}
	options := native.Options{Sanitize: true}
	if shard%3 == 2 {
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			t.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		options = native.Options{Target: "wasm32-wasi"}
	}
	path, binary, script := processOutputProgramWith(t, "console.log('dropped'); "+ending, "", options)
	ended := func(command ...string) run {
		t.Helper()
		reader, writer, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		reader.Close()
		defer writer.Close()
		stderr, err := os.Create(filepath.Join(t.TempDir(), "stderr"))
		if err != nil {
			t.Fatal(err)
		}
		defer stderr.Close()
		status := withStreams(t, writer, stderr, command[0], command[1:]...)
		said, err := os.ReadFile(stderr.Name())
		if err != nil {
			t.Fatal(err)
		}
		return run{exitCode: status, stderr: said}
	}
	truth := ended("node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	expected := 70
	if immediate {
		expected = 37
	}
	if truth.exitCode != expected || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node closed-pipe observation: %+v", truth)
	}
	var got run
	switch shard % 3 {
	case 0:
		got = ended(binary)
	case 1:
		got = ended("node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script)
	case 2:
		got = ended("node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	}
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("closed stdout: %s; Node %+v got %+v", difference, truth, got)
	}
	if immediate && shard%3 == 1 {
		data, err := os.ReadFile(filepath.Join(repository, "oracle/adamic.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		anchor := "stream.on('error', () => { if (!exiting) process.exit(70); });"
		if strings.Count(string(data), anchor) != 1 {
			t.Fatal("closed-stream mutant anchor changed")
		}
		runtimePath := filepath.Join(t.TempDir(), "runtime.mjs")
		if err := os.WriteFile(runtimePath, []byte(strings.Replace(string(data), anchor, "stream.on('error', () => process.exit(70));", 1)), 0600); err != nil {
			t.Fatal(err)
		}
		code, err := os.ReadFile(script)
		if err != nil {
			t.Fatal(err)
		}
		mutant := filepath.Join(t.TempDir(), "mutant.mjs")
		if err := os.WriteFile(mutant, []byte(strings.Replace(string(code), "'adamic'", fmt.Sprintf("%q", runtimePath), 1)), 0600); err != nil {
			t.Fatal(err)
		}
		bad := ended("node", "--disable-warning=ExperimentalWarning", mutant)
		if len(bad.stderr) != 0 || disagreement(truth, bad) != "exit codes differ" {
			t.Fatalf("closed-pipe mutant failed outside Node exit comparison: %+v", bad)
		}
		t.Log("old closed-stream handler ran cleanly; only Node exit status caught it")
	}
}
