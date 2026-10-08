package oracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func processStreams(t *testing.T, stdout, stderr bool, environment map[string]string, command ...string) run {
	t.Helper()
	encoded, err := json.Marshal(environment)
	if err != nil {
		t.Fatal(err)
	}
	arguments := []string{filepath.Join(repository, "internal/oracle/testdata/process_tty.py"), fmt.Sprint(stdout), fmt.Sprint(stderr), string(encoded)}
	result := execute(t, "python3", append(arguments, command...)...)
	if result.exitCode != 0 {
		t.Fatalf("terminal driver: %s", result.stderr)
	}
	var observed struct {
		Stdout, Stderr []byte
		ExitCode       int
	}
	if err := json.Unmarshal(result.stdout, &observed); err != nil {
		t.Fatal(err)
	}
	return run{stdout: observed.Stdout, stderr: observed.Stderr, exitCode: observed.ExitCode}
}

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
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
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
		before := "_Noreturn void adamic_process_exit_now(int code) {\n\tflush();"
		after := "_Noreturn void adamic_process_exit_now(int code) {"
		if mutation == "stderr" {
			before = "// Whatever stdout holds was written first, and stderr may be the same file.\n\t\tflush();"
			after = "// Mutant: stderr overtakes buffered stdout."
		}
		if strings.Count(runtimeSource, before) != 1 {
			t.Fatal("flush mutant anchor changed")
		}
		runtimeSource = strings.Replace(runtimeSource, before, after, 1)
		for _, name := range []string{"adamic_start", "adamic_write_line", "adamic_write_raw", "adamic_output_flush", "adamic_panic", "adamic_unreachable", "adamic_process_exit_now"} {
			runtimeSource = strings.ReplaceAll(runtimeSource, name, "mutant_"+name)
			code = strings.ReplaceAll(code, name, "mutant_"+name)
		}
		code = strings.ReplaceAll(code, "adamic_process_exit(", "mutant_exit(")
		code = runtimeSource + "\nstatic void mutant_exit(adamic_maybe_number code) { adamic_process_set_exit_code(code); mutant_adamic_process_exit_now(adamic_process_status()); }\n" + code
	}
	binary := filepath.Join(directory, "program")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, binary, script
}

func TestProcessExitOutput(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux descriptor and Node pipe semantics")
	}
	t.Parallel()
	large := "const line = 'x'.repeat(1023); for (let index = 0; index < 200; index += 1) { console.log(line); } process.exit(37);"
	for _, one := range []struct {
		name, source, destination, expected, mutant string
		shared                                      bool
	}{
		{"file immediate", "console.log('before chosen exit'); process.exit(37);", "file", "before chosen exit\n", "exit", false},
		{"file normal", "process.exitCode = 37; console.log('before normal return');", "file", "before normal return\n", "", false},
		{"merged pipe", "console.log('out 0'); console.error('err 0'); console.log('out 1'); console.error('err 1'); console.log('out last'); process.exit(37);", "pipe", "out 0\nerr 0\nout 1\nerr 1\nout last\n", "stderr", true},
		{"merged file", "console.log('out 0'); console.error('err 0'); console.log('out 1'); console.error('err 1'); console.log('out last'); process.exit(37);", "file", "out 0\nerr 0\nout 1\nerr 1\nout last\n", "stderr", true},
		{"large file", large, "file", strings.Repeat(strings.Repeat("x", 1023)+"\n", 200), "exit", false},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, binary, script := processOutputProgram(t, one.source, "")
			runner := filepath.Join(repository, "oracle/node.mjs")
			truth := processExitOutput(t, one.destination, one.shared, false, "node", "--disable-warning=ExperimentalWarning", runner, path)
			if truth.exitCode != 37 || string(truth.stdout) != one.expected || len(truth.stderr) != 0 {
				t.Fatal("Node did not exercise the expected output path")
			}
			for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
				if difference := disagreement(truth, processExitOutput(t, one.destination, one.shared, false, command...)); difference != "" {
					t.Fatal(difference)
				}
			}
			if one.mutant != "" {
				_, mutant, _ := processOutputProgram(t, one.source, one.mutant)
				got := processExitOutput(t, one.destination, one.shared, false, mutant)
				if got.exitCode != 37 || len(got.stderr) != 0 {
					t.Fatal("mutant failed outside byte comparison")
				}
				if disagreement(truth, got) != "stdout differs" {
					t.Fatal("flush mutant survived")
				}
				if one.shared {
					t.Logf("Node order %q; mutant order %q", truth.stdout, got.stdout)
				}
				t.Log("mutant compiled and ran cleanly; only Node output bytes caught it")
			}
		})
	}
}

// Node may deliver all output or lose a suffix on immediate exit. Adamic must
// deliver every byte on both backends regardless of the host's pipe behavior.
func TestProcessLargePipeExitPreservesOutput(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux backpressured pipe semantics")
	}
	t.Parallel()
	expected := []byte(strings.Repeat(strings.Repeat("x", 1023)+"\n", 200))
	runner := filepath.Join(repository, "oracle/node.mjs")
	for _, immediate := range []bool{true, false} {
		t.Run(fmt.Sprintf("immediate=%t", immediate), func(t *testing.T) {
			ending := "process.exitCode = 37;"
			if immediate {
				ending = "process.exit(37);"
			}
			path, binary, script := processOutputProgram(t, "const line = 'x'.repeat(1023); for (let index = 0; index < 200; index += 1) { console.log(line); } "+ending, "")
			script = processAsyncScript(t, script)
			truth := processExitOutput(t, "pipe", false, true, "node", "--disable-warning=ExperimentalWarning", path)
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
				if err := os.WriteFile(mutant, []byte("console.log = () => {};\n"+string(data)), 0o644); err != nil {
					t.Fatal(err)
				}
				bad := processExitOutput(t, "pipe", false, true, "node", "--disable-warning=ExperimentalWarning", runner, mutant)
				if bad.exitCode != 37 || len(bad.stderr) != 0 || disagreement(got, bad) != "stdout differs" {
					t.Fatal("JS dropped-output mutant survived, or failed outside output comparison")
				}
				t.Log("JS dropped-output mutant runs cleanly with exit 37; only the full-output comparison catches it")
			}
		})
	}
}

func TestProcessExitDrainsStderr(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux backpressured pipe semantics")
	}
	t.Parallel()
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

func TestProcessExitDoesNotUnwind(t *testing.T) {
	t.Parallel()
	path, binary, script := processOutputProgram(t, "function end(): void { try { console.log('before'); process.exit(37); } catch { console.log('must not catch'); } finally { console.log('must not finally'); } console.log('must not continue'); } end(); console.log('must not return');", "")
	runner := filepath.Join(repository, "oracle/node.mjs")
	truth := processExitOutput(t, "file", false, false, "node", "--disable-warning=ExperimentalWarning", runner, path)
	if truth.exitCode != 37 || string(truth.stdout) != "before\n" || len(truth.stderr) != 0 {
		t.Fatal("unexpected Node exit control flow")
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

func TestProcessTerminalsAndEnvironment(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/process_observations.a")
	runner := filepath.Join(repository, "oracle/node.mjs")
	for _, stdout := range []bool{false, true} {
		for _, stderr := range []bool{false, true} {
			for _, environment := range []map[string]string{
				{}, {"NO_COLOR": "", "ADAMIC_PROCESS_TEST": ""},
				{"NO_COLOR": "1", "ADAMIC_PROCESS_TEST": "hé世界"},
				{"FORCE_COLOR": "0", "ADAMIC_PROCESS_TEST": "line\nnext"},
				{"FORCE_COLOR": "3", "ADAMIC_PROCESS_TEST": "value"},
			} {
				t.Run(fmt.Sprintf("stdout=%t/stderr=%t/env=%v", stdout, stderr, environment), func(t *testing.T) {
					truth := processStreams(t, stdout, stderr, environment, "node", "--disable-warning=ExperimentalWarning", runner, path)
					if truth.exitCode != 0 || !bytes.HasPrefix(truth.stdout, []byte(fmt.Sprintf("stdout %s, stderr %s", ttyText(stdout), ttyText(stderr)))) {
						t.Fatalf("Node observation: %+v", truth)
					}
					for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
						got := processStreams(t, stdout, stderr, environment, command...)
						if difference := disagreement(truth, got); difference != "" {
							t.Errorf("%s: %s; Node %q/%q, actual %q/%q", command[0], difference, truth.stdout, truth.stderr, got.stdout, got.stderr)
						}
					}
				})
			}
		}
	}
}

func ttyText(terminal bool) string {
	if terminal {
		return "true"
	}
	return "undefined"
}

func TestProcessExitArguments(t *testing.T) {
	t.Parallel()
	for _, argument := range []string{"", "undefined", "0", "-1", "256", "257", "4294967297", "9007199254740991"} {
		t.Run("argument="+argument, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte("process.exitCode = 7; console.log('exit argument'); process.exit("+argument+");"), 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, _ := natively(t, program)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%s: Node %d native %d", difference, truth.exitCode, got.exitCode)
			}
			if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
				t.Fatal(difference)
			}
		})
	}
}

func TestProcessObjectsDoNotEscape(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"process", "process.env", "process.stdout", "process.stderr", "process.exit"} {
		t.Run(expression, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte("const held = "+expression+";"), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want explicit NotYet rather than an uninitialized object, got %v", err)
			}
		})
	}
}

// Every mutant builds cleanly and reaches the ordinary Node comparison, never a sanitizer failure.
func TestProcessMutants(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, fixture, before, after, helper, difference string }{
		{"normal status", "process_exit_code.a", "return adamic_process_status();", "return 0;", "", "exit codes differ"},
		{"immediate status", "process_exit.a", "adamic_process_exit(", "mutant_exit(", "static void mutant_exit(adamic_maybe_number code) { (void)code; adamic_process_exit((adamic_maybe_number){true, 0}); }", "exit codes differ"},
		{"missing environment", "process_observations.a", "adamic_process_environment(", "mutant_env(", "static adamic_string *mutant_env(const adamic_string *key) { (void)key; return adamic_retain(&adamic_string_empty); }", "stdout differs"},
		{"pipe tty", "process_observations.a", "adamic_process_is_tty(", "mutant_tty(", "static adamic_maybe_boolean mutant_tty(enum adamic_stream stream) { (void)stream; return (adamic_maybe_boolean){true, true}; }", "stdout differs"},
		{"range validation", "process_bad_code.a", "adamic_process_set_exit_code(", "mutant_set_code(", "static void mutant_set_code(adamic_maybe_number code) { if (!code.present || (isfinite(code.number) && trunc(code.number) == code.number)) adamic_process_set_exit_code(code); }", "stdout differs"},
	} {
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
			if difference := disagreement(onNode(t, path), got); difference != one.difference {
				t.Fatalf("caught by %q, want %q: exit %d stdout %q stderr %q", difference, one.difference, got.exitCode, got.stdout, got.stderr)
			}
			t.Logf("compiled and ran cleanly; caught only by Node %s", one.difference)
		})
	}
}
