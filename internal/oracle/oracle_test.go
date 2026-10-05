// Package oracle holds Adamic's differential test: every fixture runs twice, from source on Node and
// as a native binary stage 0 built, and the two must agree byte for byte on stdout and stderr and
// exactly on the exit code. A disagreement is a miscompile until proven otherwise.
package oracle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../.."

// fixtures is every program the oracle knows, and whether stage 0 lowers it yet. A program that
// doesn't lower yet must still be refused loudly, with where and what, so all of them are run every
// time and the list says exactly how far stage 0 has come.
var fixtures = []struct {
	path   string
	lowers bool

	// checked is a program whose inserted check fires: there, Node running the source doesn't have
	// the check, so the native binary is held to the JavaScript backend, which does.
	checked bool
}{
	{"dedication/dedication.a", true, false},
	{"internal/load/testdata/0.1/compile/01_hello.ts", true, false},
	{"internal/load/testdata/0.1/compile/02_fizzbuzz.ts", true, false},
	{"internal/load/testdata/0.1/compile/03_shapes.ts", true, false},
	{"internal/load/testdata/0.1/compile/04_closures.ts", true, false},
	{"internal/load/testdata/0.1/compile/05_wordcount.ts", true, false},
	{"internal/load/testdata/0.1/compile/06_stack.ts", true, false},
	{"internal/load/testdata/0.1/compile/07_modules/main.ts", true, false},
	{"internal/load/testdata/0.1/compile/08_results.ts", true, false},
	{"internal/load/testdata/0.1/compile/09_tree.ts", true, false},
	{"internal/load/testdata/0.1/compile/10_unicode.ts", true, false},
	{"internal/oracle/testdata/strings.a", true, false},
	{"internal/oracle/testdata/numbers.a", true, false},
	{"internal/oracle/testdata/loops.a", true, false},
	{"internal/oracle/testdata/booleans.a", true, false},
	{"internal/oracle/testdata/shadowing.a", true, false},
	{"internal/oracle/testdata/functions.a", true, false},
	{"internal/oracle/testdata/effects.a", true, false},
	{"internal/oracle/testdata/dead_zone.a", true, false},
	{"internal/oracle/testdata/objects.a", true, false},
	{"internal/oracle/testdata/modules/main.a", true, false},
	{"internal/oracle/testdata/panic.a", true, false},
	{"internal/oracle/testdata/maps_and_text.a", true, false},
	{"internal/oracle/testdata/lone_surrogates.a", true, false},
	{"internal/oracle/testdata/sorting.a", true, false},
	{"internal/oracle/testdata/classes.a", true, false},
	{"internal/oracle/testdata/closures.a", true, false},
	// A local console is the program's, and lowering it as the prelude's would print what the program
	// never asked to print.
	{"internal/oracle/testdata/local_console.a", true, false},
	{"internal/oracle/testdata/indexing.a", true, false},
	{"internal/oracle/testdata/writes.a", true, false},
	{"internal/oracle/testdata/writes_past_end.a", true, true},
	{"internal/oracle/testdata/casts.a", true, false},
	{"internal/oracle/testdata/cast_fails.a", true, true},
	{"internal/oracle/testdata/updates.a", true, false},
	{"internal/oracle/testdata/read_order.a", true, false},
	{"internal/oracle/testdata/string_index.a", true, false},
	{"internal/oracle/testdata/visits.a", true, false},
	{"internal/oracle/testdata/searches.a", true, false},
	{"internal/oracle/testdata/spreads.a", true, false},
	{"internal/oracle/testdata/maybe_numbers.a", true, false},
	{"internal/oracle/testdata/defaults.a", true, false},
	{"internal/oracle/testdata/search_halves.a", true, false},
	{"internal/oracle/testdata/strings_more.a", true, false},
	{"internal/oracle/testdata/number_parsing.a", true, false},
	// Every Math function ported from V8, printed in full, so a last bit that differs from Node shows.
	{"internal/oracle/testdata/navigation.a", true, false},
	{"internal/oracle/testdata/number_formats.a", true, false},
	{"internal/oracle/testdata/precision_range.a", true, false},
	{"internal/oracle/testdata/radixes.a", true, false},
	{"internal/oracle/testdata/radix_range.a", true, false},
	{"internal/oracle/testdata/optional_numbers.a", true, false},
	{"internal/oracle/testdata/map_iteration.a", true, false},
	{"internal/oracle/testdata/sorts.a", true, false},
	{"internal/oracle/testdata/sort_releases.a", true, false},
	// The same sorts at the top level, where only the globals' release at exit lets the leak check see.
	{"internal/oracle/testdata/sort_top_level.a", true, false},
	{"internal/oracle/testdata/splices.a", true, false},
	{"internal/oracle/testdata/fills.a", true, false},
	{"internal/oracle/testdata/fill_length.a", true, false},
	{"internal/oracle/testdata/array_from.a", true, false},
	{"internal/oracle/testdata/array_from_length.a", true, false},
	{"internal/oracle/testdata/array_from_undefined.a", true, false},
	{"internal/oracle/testdata/maybe_booleans.a", true, false},
	{"internal/oracle/testdata/maybe_boolean_panic.a", true, false},
	{"internal/oracle/testdata/unions.a", true, false},
	{"internal/oracle/testdata/maybe_number_slots.a", true, false},
	{"internal/oracle/testdata/case_mapping.a", true, false},
	{"internal/oracle/testdata/undefined_elements.a", true, false},
	{"internal/oracle/testdata/map_zero_keys.a", true, false},
	{"internal/oracle/testdata/string_limits.a", true, false},
	{"internal/oracle/testdata/string_too_long.a", true, false},
	{"internal/oracle/testdata/pad_too_long.a", true, false},
	{"internal/oracle/testdata/stack_overflow.a", true, false},
	{"internal/oracle/testdata/adversarial_order.a", true, false},
	{"internal/oracle/testdata/adversarial_exits.a", true, false},
	{"internal/oracle/testdata/adversarial_iteration.a", true, false},
	{"internal/oracle/testdata/number_edges.a", true, false},
	{"internal/oracle/testdata/long_chain.a", true, false},
	{"internal/oracle/testdata/write_after_shrink.a", true, true},
	{"internal/oracle/testdata/normalize.a", true, false},
	{"internal/oracle/testdata/normalize_form.a", true, false},
	{"internal/oracle/testdata/string_positions.a", true, false},
	// Borrowed parameters: a reassigned one has to stay owned, and so does a closure's, which map hands
	// an element it may overwrite. Each breaks under ASan if it's borrowed.
	{"internal/oracle/testdata/borrow_reassigned.a", true, false},
	{"internal/oracle/testdata/borrow_map_overwrite.a", true, false},
	// A spread is read before its fields' values, as JavaScript reads it.
	{"internal/oracle/testdata/spread_snapshot.a", true, false},
	// Reuse in place: taken only where nothing can tell, and where something could, never.
	{"internal/oracle/testdata/reuse.a", true, false},
	// Output beyond native's stdout buffer, and the points where it must be flushed (adamic.c).
	{"internal/oracle/testdata/large_output.a", true, false},
	{"internal/oracle/testdata/output_then_panic.a", true, false},
	{"internal/oracle/testdata/interleaved.a", true, false},
	// sin, cos and tan where reducing by pi / 2 cancels the most bits, as no sweep input does.
	{"internal/oracle/testdata/trig_reduction.a", true, false},
}

// run is one execution's observable behavior: what the oracle compares.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

func execute(t *testing.T, name string, arguments ...string) run {
	t.Helper()
	return executeWith(t, nil, name, arguments...)
}

// executeWith runs a command with environment added to the test's own; a later value for the same
// name wins, so a sanitizer setting here can't be overridden by one inherited from the shell.
func executeWith(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := bounded(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// bounded is a command that can't outlive its test: it has a deadline, it runs in a process group of
// its own, and when the deadline passes or the test ends, the whole group is killed. A fixture that
// loops, or a child left with nowhere to write, is stopped instead of orphaned.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	return command
}

// onNode runs a program's source on Node: the oracle.
func onNode(t *testing.T, path string) run {
	t.Helper()
	return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
}

// onJavaScriptBackend runs a lowered program through the JavaScript backend, on Node.
func onJavaScriptBackend(t *testing.T, program *ir.Program) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}

// lowered checks and lowers a program, or returns stage 0's refusal.
func lowered(t *testing.T, path string) (*ir.Program, error) {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return lower.Lower(context.Background(), program)
}

// natively builds a lowered program under the sanitizers and runs it. It returns the binary too, so
// the leak check on Linux can run the same one again.
//
// On Linux, ASan carries LeakSanitizer and runs it at exit by default. This run is the comparison, and
// a program that panics exits 70 holding what it held, which isn't a leak, so leak detection is off
// here and the leak check is a run of its own. macOS's ASan has no leak detection to turn off.
func natively(t *testing.T, program *ir.Program) (run, string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return executeWith(t, environment, binary), binary
}

// leaks returns a report of everything a finished program never let go of, or "" when it let go of
// everything. No garbage collector means every reference the compiler hands out has to come back;
// this is where a missing release shows. Only programs Node finishes with exit 0 are asked.
//
// macOS has the leaks tool; Linux has LeakSanitizer, part of ASan there, run on the sanitized binary
// the comparison already built.
func leaks(t *testing.T, program *ir.Program, sanitized string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		return leaksTool(t, program)
	case "linux":
		return leakSanitizer(t, sanitized)
	}
	t.Fatalf("no leak check for %s: the oracle knows macOS's leaks tool and Linux's LeakSanitizer", runtime.GOOS)
	return ""
}

// leaksTool builds a lowered program without sanitizers (they and macOS's leaks tool don't mix), runs
// it under leaks --atExit, and returns its report when anything leaked.
func leaksTool(t *testing.T, program *ir.Program) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	report := execute(t, "leaks", "--atExit", "--", binary)
	if report.exitCode == 0 {
		return ""
	}
	return string(report.stdout)
}

// leakSanitizer runs a sanitized binary again with leak detection on, and returns LeakSanitizer's
// report when anything leaked. The program finished with exit 0 on the comparison run, so any other
// exit here is the sanitizer's.
func leakSanitizer(t *testing.T, binary string) string {
	t.Helper()
	report := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if report.exitCode == 0 {
		return ""
	}
	return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
}

// disagreement says how two runs differ, or "" when they don't.
func disagreement(oracle run, native run) string {
	switch {
	case oracle.exitCode != native.exitCode:
		return "exit codes differ"
	case !bytes.Equal(oracle.stdout, native.stdout):
		return "stdout differs"
	case !bytes.Equal(oracle.stderr, native.stderr):
		return "stderr differs"
	}
	return ""
}

func TestNativeAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range fixtures {
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if !fixture.lowers {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("want stage 0 to refuse with where and what, got %v", err)
				}
				t.Logf("not yet: %v", notYet)
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			oracle, backend := onNode(t, path), onJavaScriptBackend(t, program)
			native, sanitized := natively(t, program)
			if fixture.checked {
				// The check fires, so the source on Node goes on where Adamic stops: hold native to the
				// backend that carries the same check, and make sure the check really did fire.
				if difference := disagreement(backend, native); difference != "" {
					t.Errorf("%s\nbackend: exit %d, stdout %q, stderr %q\nnative:  exit %d, stdout %q, stderr %q",
						difference, backend.exitCode, backend.stdout, backend.stderr, native.exitCode, native.stdout, native.stderr)
				}
				if native.exitCode != 70 || oracle.exitCode == 70 {
					t.Errorf("want the inserted check to fire natively (exit 70) where the source on Node runs on: native %d, Node %d", native.exitCode, oracle.exitCode)
				}
				return
			}
			if difference := disagreement(oracle, native); difference != "" {
				t.Errorf("%s\nnode:   exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q",
					difference, oracle.exitCode, oracle.stdout, oracle.stderr, native.exitCode, native.stdout, native.stderr)
			}
			// The JavaScript backend runs the same IR the native one compiled, with nothing of C, so
			// where it differs from the source the fault is in lowering.
			if difference := disagreement(oracle, backend); difference != "" {
				t.Errorf("JavaScript backend: %s\nnode:    exit %d, stdout %q, stderr %q\nbackend: exit %d, stdout %q, stderr %q",
					difference, oracle.exitCode, oracle.stdout, oracle.stderr, backend.exitCode, backend.stdout, backend.stderr)
			}
			// A program that panicked stopped where it stood, as Node's does, so what it held then
			// isn't a leak; every program that finishes must have let go of everything.
			if oracle.exitCode == 0 {
				if leaked := leaks(t, program, sanitized); leaked != "" {
					t.Errorf("leaks:\n%s", leaked)
				}
			}
		})
	}
}

// The oracle has to be able to fail. One byte added to the dedication's string, after lowering, must
// read as a disagreement.
func TestTheOracleCatchesOneByte(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "dedication", "dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	program.Strings[0] += "!"
	native, _ := natively(t, program)
	if difference := disagreement(onNode(t, path), native); difference != "stdout differs" {
		t.Errorf("got %q, want the mutant caught as \"stdout differs\"", difference)
	}
}
