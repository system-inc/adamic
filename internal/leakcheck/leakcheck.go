// Package leakcheck is the one leak check every test of a native program runs: the oracle's fixtures,
// its input fixtures, and each stage-1 port. No garbage collector means every reference the compiler
// hands out has to come back, and this is where a missing release shows. Only a program that finished
// is asked; one that panicked stopped holding what it held. Explicit process.exit is intentional teardown: the runtime reports it and releases nothing.
//
// It's a package of its own, not a _test.go file, because tests in one package can't import another
// package's tests, and every caller has to run the same check, so a fix to it lands once.
//
// On Linux the check is LeakSanitizer, run on the sanitized binary the comparison already built. On
// macOS it's the counted build (native.Options.Count), run twice: on its own, where a finished
// program's allocations must be its frees and its values in regions (runtime/count.h), the rule the
// oracle's counts table is read by; then under leaks --atExit. The counts are the check for values:
// outside the sanitizers every small value lives in a chunk of the runtime's size-class allocator
// (heap.c), and every chunk stays reachable from the runtime's own table of them, so to leaks a value
// never let go of still reads as reachable. leaks is the check for what the runtime takes from malloc
// outside the counts: an array's elements, a map's table, a region's blocks.
package leakcheck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// CountsLine is the line a counted build writes last to stderr (runtime/count.c).
var CountsLine = regexp.MustCompile(`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n\z`)

// IntentionalExitLine is emitted only by the runtime exit primitive in counted/leak-check runs.
var IntentionalExitLine = regexp.MustCompile(`(?m)^adamic: intentional exit: status (\d+)\n`)

func intentionalExit(run Run) bool {
	match := IntentionalExitLine.FindSubmatch(run.Stderr)
	if match == nil {
		return false
	}
	status, err := strconv.Atoi(string(match[1]))
	if err != nil || status != run.ExitCode {
		return false
	}
	// Only a terminal marker, optionally followed by the runtime's complete counts, is accepted.
	tail := run.Stderr[bytes.Index(run.Stderr, match[0])+len(match[0]):]
	if len(tail) == 0 {
		return true
	}
	// Graph counters are a second, fixed runtime record, never arbitrary diagnostics.
	graph := regexp.MustCompile(`^adamic: graph counts: regions \d+ merges \d+\n`)
	tail = graph.ReplaceAll(tail, nil)
	location := CountsLine.FindIndex(tail)
	return location != nil && location[0] == 0 && location[1] == len(tail)
}

// Run is one execution's observable behavior.
type Run struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Program is a finished native program to check, and how its caller runs things.
type Program struct {
	// C is the program, as native.C wrote it; macOS builds it counted.
	C string

	// Sanitized is the binary the comparison built under the sanitizers; Linux runs it again with leak
	// detection on.
	Sanitized string

	// Counted is where macOS writes the counted build.
	Counted string

	// BuildCounted optionally builds the counted program against a changed runtime.
	// Nil uses native.Build and the embedded production runtime.
	BuildCounted func(code, output string) error

	// Arguments gives the program's arguments, once for each run, so a program that writes can be
	// given a fresh place to write in every time. Nil is no arguments.
	Arguments func() []string

	// Execute runs a command with environment added to the caller's own, as the caller runs its
	// program: its deadline, its directory, its user.
	Execute func(environment []string, name string, arguments ...string) Run
}

// Check returns a report of everything the program never let go of, or "" when it let go of
// everything. The error is the check failing to run at all.
func Check(program Program) (string, error) {
	arguments := func() []string {
		if program.Arguments == nil {
			return nil
		}
		return program.Arguments()
	}
	switch runtime.GOOS {
	case "darwin":
		build := program.BuildCounted
		if build == nil {
			build = func(code, output string) error { return native.Build(code, output, native.Options{Count: true}) }
		}
		if err := build(program.C, program.Counted); err != nil {
			return "", err
		}
		counted := program.Execute(nil, program.Counted, arguments()...)
		if report := Unbalanced(counted); report != "" {
			return report, nil
		}
		if intentionalExit(counted) {
			return "", nil
		}
		report := program.Execute(nil, "leaks", append([]string{"--atExit", "--", program.Counted}, arguments()...)...)
		if report.ExitCode == 0 {
			return "", nil
		}
		return string(report.Stdout), nil
	case "linux":
		report := program.Execute([]string{"ASAN_OPTIONS=detect_leaks=1", "ADAMIC_LEAK_CHECK=1"}, program.Sanitized, arguments()...)
		if intentionalExit(report) || report.ExitCode == 0 {
			return "", nil
		}
		return fmt.Sprintf("exit %d\n%s", report.ExitCode, report.Stderr), nil
	}
	return "", fmt.Errorf("leakcheck: no leak check for %s: there's macOS's counted build and Linux's LeakSanitizer", runtime.GOOS)
}

// Unbalanced reads a counted run's counts and returns a report when the program finished holding heap
// values, or "" when its allocations are its frees and its values in regions.
func Unbalanced(counted Run) string {
	match := CountsLine.FindSubmatch(counted.Stderr)
	if match != nil && intentionalExit(counted) {
		return ""
	}
	if counted.ExitCode != 0 || match == nil {
		return fmt.Sprintf("the counted build didn't finish with its counts: exit %d, stderr %q", counted.ExitCode, counted.Stderr)
	}
	// Signed, so frees past allocations read as a negative leak. The digits matched \d+, so only a
	// count past int64 fails to parse, and that is reported as what it is.
	var counts [3]int64
	for index, group := range []int{1, 2, 6} {
		count, err := strconv.ParseInt(string(match[group]), 10, 64)
		if err != nil {
			return fmt.Sprintf("unreadable counts: %v", err)
		}
		counts[index] = count
	}
	allocations, frees, regions := counts[0], counts[1], counts[2]
	if allocations == frees+regions {
		return ""
	}
	return fmt.Sprintf("heap values leaked: %d (allocations %d, frees %d, in regions %d)", allocations-frees-regions, allocations, frees, regions)
}

// Report is Check for a test that runs its program plainly from its own directory: every command
// bounded by ten minutes, in a process group of its own that's killed when the deadline passes or the
// test ends, and the counted build in the test's temporary directory. A check that can't run fails the
// test.
func Report(t testing.TB, code string, sanitized string, arguments ...string) string {
	t.Helper()
	report, err := Check(Program{
		C:         code,
		Sanitized: sanitized,
		Counted:   filepath.Join(t.TempDir(), "counted"),
		Arguments: func() []string { return arguments },
		Execute: func(environment []string, name string, arguments ...string) Run {
			return bounded(t, environment, name, arguments...)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

// bounded runs a command that can't outlive its test.
func bounded(t testing.TB, environment []string, name string, arguments ...string) Run {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
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
	return Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
}
