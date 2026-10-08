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
	"strconv"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
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

// By default, eight GiB for the formerly serial tests leaves eight GiB on the measured
// sixteen-GiB box for layoutSlots' work, other parallel tests and retained heaps.
// Each unit is 512 MiB; weights round 125% of the isolated process-tree peak up.
const markdownMemoryDefaultUnits = 16
const markdownMemoryLargestWeight = 8

var markdownMemory = newMarkdownMemoryBudget(markdownMemoryDefaultUnits)
var markdownMemoryConfigurationOnce sync.Once
var markdownMemoryConfigurationError error

// Check before t.Parallel so the first selected test reports configuration
// errors by name. Every top-level test participates, including filtered runs.
func configureMarkdownMemory(t *testing.T) {
	t.Helper()
	first := false
	markdownMemoryConfigurationOnce.Do(func() {
		first = true
		units := markdownMemoryDefaultUnits
		if value, set := os.LookupEnv("ADAMIC_MARKDOWNBLOCKS_MEMORY_UNITS"); set {
			var err error
			units, err = strconv.Atoi(value)
			if err != nil || units < markdownMemoryLargestWeight {
				markdownMemoryConfigurationError = fmt.Errorf("ADAMIC_MARKDOWNBLOCKS_MEMORY_UNITS=%q: must be a positive integer of at least %d (512 MiB units; largest test weight)", value, markdownMemoryLargestWeight)
				return
			}
		}
		markdownMemory.capacity = units
		t.Logf("markdownblocks memory capacity: %d units (512 MiB each)", units)
	})
	if markdownMemoryConfigurationError != nil {
		if first {
			t.Fatal(markdownMemoryConfigurationError)
		}
		// Abort later tests too, before any work or admission waits. The first
		// test owns the diagnostic; do not repeat it in concurrent output.
		t.FailNow()
	}
}

func parallelMarkdown(t *testing.T) {
	t.Helper()
	configureMarkdownMemory(t)
	t.Parallel()
}

type markdownMemoryBudget struct {
	mu       sync.Mutex
	changed  *sync.Cond
	capacity int
	used     int
}

func newMarkdownMemoryBudget(capacity int) *markdownMemoryBudget {
	budget := &markdownMemoryBudget{capacity: capacity}
	budget.changed = sync.NewCond(&budget.mu)
	return budget
}

func (budget *markdownMemoryBudget) acquire(weight int) {
	budget.mu.Lock()
	defer budget.mu.Unlock()
	for budget.used+weight > budget.capacity {
		budget.changed.Wait()
	}
	budget.used += weight
}

func (budget *markdownMemoryBudget) release(weight int) {
	budget.mu.Lock()
	budget.used -= weight
	budget.changed.Broadcast()
	budget.mu.Unlock()
}

func parallelMarkdownMemory(t *testing.T, weight int) {
	t.Helper()
	configureMarkdownMemory(t)
	// KEEP modes export to caller-chosen directories. Leave the newly parallel
	// tests in the serial phase whenever an export is requested, even if callers
	// assign the same directory to different KEEP variables.
	for _, name := range []string{
		"ADAMIC_MARKDOWNAST_KEEP", "ADAMIC_MDAST_KEEP", "ADAMIC_PATH_KEEP",
		"ADAMIC_MARKDOWNBLOCKS_KEEP", "ADAMIC_MARKDOWNLISTS_KEEP",
	} {
		if os.Getenv(name) != "" {
			return
		}
	}
	t.Parallel()
	if weight > markdownMemory.capacity {
		t.Fatalf("%s: markdownblocks memory weight %d exceeds capacity %d (512 MiB units)", t.Name(), weight, markdownMemory.capacity)
	}
	markdownMemory.acquire(weight)
	t.Cleanup(func() { markdownMemory.release(weight) })
}

type artifactKey struct {
	source   [32]byte
	sanitize bool
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
		if output, err := combinedOutput(command); err != nil {
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
	key := artifactKey{source: sha256.Sum256([]byte(source)), sanitize: options.Sanitize}
	pending := &artifactBuild{done: make(chan struct{})}
	actual, loaded := artifacts.LoadOrStore(key, pending)
	build := actual.(*artifactBuild)
	if !loaded {
		build.binary = filepath.Join(artifactDirectory, fmt.Sprintf("%x-%t", key.source, key.sanitize))
		build.err = native.Build(source, build.binary, options)
		close(build.done)
	}
	<-build.done
	return build.binary, build.err
}

func TestNativeBuildModesAreDistinct(t *testing.T) {
	parallelMarkdown(t)
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

// bounded prepares a child for the shared output-based hang guard.
// Silent builds and buffered children use its 30-minute first-output window.
// With ten CPU burners, the longest output gap was 5.7s; the default
// two-minute Stall leaves more than three times that gap as headroom.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	return exec.Command(name, arguments...)
}

func combinedOutput(command *exec.Cmd) ([]byte, error) {
	return childguard.CombinedOutput(command, childguard.Options{})
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
	err := childguard.Run(command, childguard.Options{})
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
	if output, err := combinedOutput(bounded(t, "clang", flags...)); err != nil {
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

// leaks returns a report of everything the finished port never let go of, or "": the leak check the
// oracle runs on every fixture (internal/leakcheck), on the sanitized binary and the port's C.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	return leakcheck.Report(t, native.C(program), sanitized, arguments...)
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
