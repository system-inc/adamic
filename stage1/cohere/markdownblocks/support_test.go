package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."

var artifactDirectory string
var artifacts sync.Map
var formatterOnce sync.Once
var formatterBinary string
var formatterError error

type artifactKey struct {
	source   [32]byte
	sanitize bool
	split    bool
}

type artifactBuild struct {
	done   chan struct{}
	binary string
	err    error
}

func TestMain(m *testing.M) {
	var err error
	artifactDirectory, err = os.MkdirTemp("", "adamic-markdown-build-")
	if err != nil {
		panic(err)
	}
	code := m.Run()
	if err := os.RemoveAll(artifactDirectory); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

// TestMain keeps the formatter alive for every regeneration check in this package run.
func cohereFormatter(t *testing.T) string {
	t.Helper()
	formatterOnce.Do(func() {
		formatterBinary = filepath.Join(artifactDirectory, "cohere")
		cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
		if err != nil {
			formatterError = err
			return
		}
		command := bounded(t, "go", "build", "-o", formatterBinary, "./command/cohere")
		command.Dir = cohere
		if output, err := command.CombinedOutput(); err != nil {
			formatterError = fmt.Errorf("build cohere formatter: %w\n%s", err, output)
		}
	})
	if formatterError != nil {
		t.Fatal(formatterError)
	}
	t.Logf("regeneration formatter: %s", formatterBinary)
	return formatterBinary
}

// Reuse compilation only within this package execution, never observations.
// All corpora, oracle comparisons and leak checks execute afresh even uncached.
func nativeBinary(t *testing.T, source string, sanitize bool) string {
	t.Helper()
	binary, err := nativeBinaryResult(source, native.Options{Sanitize: sanitize})
	if err != nil {
		t.Fatal(err)
	}
	return binary
}

func nativeBinaryResult(source string, options native.Options) (string, error) {
	key := artifactKey{source: sha256.Sum256([]byte(source)), sanitize: options.Sanitize, split: options.Split || os.Getenv("ADAMIC_NATIVE_SPLIT") == "1"}
	pending := &artifactBuild{done: make(chan struct{})}
	actual, loaded := artifacts.LoadOrStore(key, pending)
	build := actual.(*artifactBuild)
	if !loaded {
		build.binary = filepath.Join(artifactDirectory, fmt.Sprintf("%x-%t-%t", key.source, key.sanitize, key.split))
		build.err = native.Build(source, build.binary, options)
		close(build.done)
	}
	<-build.done
	return build.binary, build.err
}

func TestNativeBuildModesAreDistinct(t *testing.T) {
	t.Parallel()
	const source = `#include <stdio.h>
int main(void) {
#if __has_feature(address_sanitizer)
    puts("sanitized");
#else
    puts("release");
#endif
    return 0;
}
`
	for _, mode := range []struct {
		sanitize bool
		want     string
	}{{true, "sanitized\n"}, {false, "release\n"}} {
		answer := execute(t, nil, nativeBinary(t, source, mode.sanitize))
		clean(t, "clang build mode", answer)
		equal(t, "clang build mode", answer.stdout, []byte(mode.want))
	}
}

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// lowered checks and lowers a program, failing the test with stage 0's refusal if it can't.
func lowered(t *testing.T, path string) *ir.Program {
	t.Helper()
	result, err := loweredResult(path)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func loweredResult(path string) (*ir.Program, error) {
	program, err := load.Load([]string{path})
	if err != nil {
		return nil, fmt.Errorf("Load: %v", err)
	}
	result, err := lower.Lower(context.Background(), program)
	if err != nil {
		return nil, fmt.Errorf("Lower: %v", err)
	}
	return result, nil
}

// bounded is a command that can't outlive its test: it has a deadline, it runs in a process group of
// its own, and when the deadline passes or the test ends, the whole group is killed.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	return command
}

func execute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	result, err := executeResult(t, environment, name, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func executeResult(t *testing.T, environment []string, name string, arguments ...string) (run, error) {
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
		return run{}, fmt.Errorf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, nil
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	result, err := onNodeResult(t, path, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func onNodeResult(t *testing.T, path string, arguments ...string) (run, error) {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		return run{}, err
	}
	return executeResult(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path, arguments...)
}

// nativelyRun is natively's run alone.
func nativelyRun(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	result, _ := natively(t, program, arguments...)
	return result
}

// Output-only mutants need sanitizer checks, not repeated optimizer work on the
// full formatter. The unchanged port is checked at -O1 and release -O2 separately.
func nativeMutant(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	directory := t.TempDir()
	options := native.Options{Sanitize: true}
	library, err := native.RuntimeLibrary("", options)
	if err != nil {
		t.Fatal(err)
	}
	source, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "mutant")
	write(t, source, []byte(native.C(program)))
	flags := native.Flags(options)
	for index, flag := range flags {
		if flag == "-O1" {
			flags[index] = "-O0"
		}
	}
	flags = append(flags, "-I", filepath.Dir(library), "-o", binary, source)
	flags = append(flags, native.RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := bounded(t, "clang", flags...).CombinedOutput(); err != nil {
		t.Fatalf("mutant build: %v\n%s", err, output)
	}
	return execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, arguments...)
}

func releaseRun(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	binary := nativeBinary(t, native.C(program), false)
	return execute(t, nil, binary, arguments...)
}

// natively builds the lowered port under the address and undefined-behavior sanitizers and runs it,
// returning the binary too, for the leak check. Leak detection is off here, as in the oracle; leaks is
// its own run.
func natively(t *testing.T, program *ir.Program, arguments ...string) (run, string) {
	t.Helper()
	binary := nativeBinary(t, native.C(program), true)
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(t, environment, binary, arguments...), binary
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(t.TempDir(), "port")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		report := execute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
}
func clean(t *testing.T, name string, r run) {
	t.Helper()
	if r.exitCode != 0 || len(r.stderr) > 0 {
		t.Fatalf("%s exit %d stderr %s", name, r.exitCode, r.stderr)
	}
}
func equal(t *testing.T, name string, a, b []byte) {
	t.Helper()
	if !bytes.Equal(a, b) {
		i := 0
		for i < len(a) && i < len(b) && a[i] == b[i] {
			i++
		}
		t.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, i, len(a), len(b))
	}
}

// A fixture task collects its result without calling Fatal from a worker. The caller
// joins every task before returning, then reports errors and comparisons on the test
// goroutine, keeping each failure's message intact.
type fixtureTask[T any] struct {
	done  chan struct{}
	value T
	err   error
}

func startFixtureTask[T any](workers *sync.WaitGroup, work func() (T, error)) *fixtureTask[T] {
	task := &fixtureTask[T]{done: make(chan struct{})}
	workers.Add(1)
	go func() {
		defer workers.Done()
		defer close(task.done)
		task.value, task.err = work()
	}()
	return task
}

func (task *fixtureTask[T]) result() (T, error) {
	<-task.done
	return task.value, task.err
}

func (task *fixtureTask[T]) await(t *testing.T) T {
	t.Helper()
	value, err := task.result()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
