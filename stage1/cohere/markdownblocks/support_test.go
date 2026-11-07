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
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."

var artifactDirectory string
var artifacts sync.Map

// By default, eight GiB for the formerly serial tests leaves eight GiB on the measured
// sixteen-GiB box for the layout shards' work, other parallel tests and retained heaps.
// Each unit is 512 MiB; weights round 125% of the isolated process-tree peak up.
const markdownMemoryDefaultUnits = 16
const markdownMemoryLargestWeight = 8

var markdownMemory = newMarkdownMemoryBudget(markdownMemoryDefaultUnits)
var markdownMemoryConfigurationOnce sync.Once
var markdownMemoryConfigurationError error

// Check before test.Parallel so the first selected test reports configuration
// errors by name. Every top-level test participates, including filtered runs.
func configureMarkdownMemory(test *testing.T) {
	test.Helper()
	first := false
	markdownMemoryConfigurationOnce.Do(func() {
		first = true
		units := markdownMemoryDefaultUnits
		if value, set := os.LookupEnv("ADAMIC_MARKDOWNBLOCKS_MEMORY_UNITS"); set {
			var parseError error
			units, parseError = strconv.Atoi(value)
			if parseError != nil || units < markdownMemoryLargestWeight {
				markdownMemoryConfigurationError = fmt.Errorf(
					"ADAMIC_MARKDOWNBLOCKS_MEMORY_UNITS=%q: must be a positive integer of at least %d (512 MiB units; largest test weight)",
					value,
					markdownMemoryLargestWeight,
				)
				return
			}
		}
		markdownMemory.capacity = units
		test.Logf("markdownblocks memory capacity: %d units (512 MiB each)", units)
	})
	if markdownMemoryConfigurationError != nil {
		if first {
			test.Fatal(markdownMemoryConfigurationError)
		}
		// Abort later tests too, before any work or admission waits. The first
		// test owns the diagnostic; do not repeat it in concurrent output.
		test.FailNow()
	}
}

func parallelMarkdown(test *testing.T) {
	test.Helper()
	configureMarkdownMemory(test)
	test.Parallel()
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

func parallelMarkdownMemory(test *testing.T, weight int) {
	test.Helper()
	configureMarkdownMemory(test)
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
	test.Parallel()
	if weight > markdownMemory.capacity {
		test.Fatalf(
			"%s: markdownblocks memory weight %d exceeds capacity %d (512 MiB units)",
			test.Name(),
			weight,
			markdownMemory.capacity,
		)
	}
	markdownMemory.acquire(weight)
	test.Cleanup(func() { markdownMemory.release(weight) })
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

func TestMain(suite *testing.M) {
	var temporaryDirectoryError error
	artifactDirectory, temporaryDirectoryError = os.MkdirTemp("", "adamic-markdown-build-")
	if temporaryDirectoryError != nil {
		panic(temporaryDirectoryError)
	}
	code := suite.Run()
	if removeError := os.RemoveAll(artifactDirectory); removeError != nil {
		fmt.Fprintln(os.Stderr, removeError)
		code = 1
	}
	os.Exit(code)
}

// Reuse compilation only within this package execution, never observations.
// All corpora, oracle comparisons and leak checks execute afresh even uncached.
func nativeBinary(test *testing.T, source string, sanitize bool) string {
	test.Helper()
	binary, nativeBinaryResultError := nativeBinaryResult(source, native.Options{Sanitize: sanitize})
	if nativeBinaryResultError != nil {
		test.Fatal(nativeBinaryResultError)
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

// Not parallel: shared markdownMemory configuration and artifacts build cache; existing helper controls parallel execution.
func TestNativeBuildModesAreDistinct(test *testing.T) {
	parallelMarkdown(test)
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
		answer := execute(test, nil, nativeBinary(test, source, mode.sanitize))
		clean(test, "clang build mode", answer)
		equal(test, "clang build mode", answer.stdout, []byte(mode.want))
	}
}

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// lowered checks and lowers a program, failing the test with stage 0's refusal if it can't.
func lowered(test *testing.T, path string) *ir.Program {
	test.Helper()
	result, loweredResultError := loweredResult(path)
	if loweredResultError != nil {
		test.Fatal(loweredResultError)
	}
	return result
}

func loweredResult(path string) (*ir.Program, error) {
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		return nil, fmt.Errorf("Load: %v", loadError)
	}
	result, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		return nil, fmt.Errorf("Lower: %v", lowerError)
	}
	return result, nil
}

// bounded prepares a child for the shared output-based hang guard.
// Silent builds and buffered children use its 30-minute first-output window.
// With ten CPU burners, the longest output gap was 5.7s; the default
// two-minute Stall leaves more than three times that gap as headroom.
func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	return exec.Command(name, arguments...)
}

func combinedOutput(command *exec.Cmd) ([]byte, error) {
	return childguard.CombinedOutput(command, childguard.Options{})
}

func execute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	result, executeResultError := executeResult(test, environment, name, arguments...)
	if executeResultError != nil {
		test.Fatal(executeResultError)
	}
	return result
}

func executeResult(test *testing.T, environment []string, name string, arguments ...string) (run, error) {
	test.Helper()
	command := bounded(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if runError != nil && !errors.As(runError, &exitError) {
		return run{}, fmt.Errorf("running %s: %v", name, runError)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, nil
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	result, onNodeResultError := onNodeResult(test, path, arguments...)
	if onNodeResultError != nil {
		test.Fatal(onNodeResultError)
	}
	return result
}

func onNodeResult(test *testing.T, path string, arguments ...string) (run, error) {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		return run{}, absolutePathError
	}
	return executeResult(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(test *testing.T, program *ir.Program, arguments ...string) run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if writeError := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	return onNode(test, path, arguments...)
}

// Output-only mutants need sanitizer checks, not repeated optimizer work on the
// full formatter. The unchanged port is checked at -O1 and release -O2 separately.
func nativeMutant(test *testing.T, program *ir.Program, arguments ...string) run {
	test.Helper()
	directory := test.TempDir()
	options := native.Options{Sanitize: true}
	library, runtimeLibraryError := native.RuntimeLibrary("", options)
	if runtimeLibraryError != nil {
		test.Fatal(runtimeLibraryError)
	}
	source, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "mutant")
	write(test, source, []byte(native.C(program)))
	flags := native.Flags(options)
	for index, flag := range flags {
		if flag == "-O1" {
			flags[index] = "-O0"
		}
	}
	flags = append(flags, "-I", filepath.Dir(library), "-o", binary, source)
	flags = append(flags, native.RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, combinedOutputError := combinedOutput(bounded(test, "clang", flags...)); combinedOutputError != nil {
		test.Fatalf("mutant build: %v\n%s", combinedOutputError, output)
	}
	return execute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, arguments...)
}

func releaseRun(test *testing.T, program *ir.Program, arguments ...string) run {
	test.Helper()
	binary := nativeBinary(test, native.C(program), false)
	return execute(test, nil, binary, arguments...)
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(test *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	test.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(test.TempDir(), "port")
		if buildError := native.Build(native.C(program), binary, native.Options{}); buildError != nil {
			test.Fatal(buildError)
		}
		report := execute(test, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

func write(test *testing.T, path string, b []byte) {
	test.Helper()
	if writeError := os.WriteFile(path, b, 0644); writeError != nil {
		test.Fatal(writeError)
	}
}
func clean(test *testing.T, name string, run run) {
	test.Helper()
	if run.exitCode != 0 || len(run.stderr) > 0 {
		test.Fatalf("%s exit %d stderr %s", name, run.exitCode, run.stderr)
	}
}
func equal(test *testing.T, name string, left, right []byte) {
	test.Helper()
	if !bytes.Equal(left, right) {
		index := 0
		for index < len(left) && index < len(right) && left[index] == right[index] {
			index++
		}
		test.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, index, len(left), len(right))
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

func (task *fixtureTask[T]) await(test *testing.T) T {
	test.Helper()
	value, resultError := task.result()
	if resultError != nil {
		test.Fatal(resultError)
	}
	return value
}
