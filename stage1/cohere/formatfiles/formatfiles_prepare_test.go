package formatfiles

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

type formatfilesPrepared struct {
	source, binary, javascript, unsanitized string
	program                                 *ir.Program
}
type formatfilesShared struct {
	casesPath, answers string
	parts              []formatfilesPart
	port               formatfilesPrepared
	mutated            map[string]formatfilesPrepared
}

var formatfilesCorpusOnce sync.Once
var formatfilesSharedState *formatfilesShared
var formatfilesPortOnce sync.Once
var formatfilesPortState formatfilesPrepared
var formatfilesUnsanitizedOnce sync.Once
var formatfilesUnsanitizedState string
var formatfilesMutantOnce = make([]sync.Once, len(mutants))
var formatfilesMutantState = make([]formatfilesPrepared, len(mutants))
var formatfilesSetupRoot string

// Keep the shared fixtures alive until every parallel leaf has finished.
// Not parallel: TestMain owns only the process-wide fixture lifetime.
func TestMain(m *testing.M) {
	var temporaryDirectoryError error
	formatfilesSetupRoot, temporaryDirectoryError = os.MkdirTemp("", "formatfiles-shared-")
	if temporaryDirectoryError != nil {
		fmt.Fprintln(os.Stderr, temporaryDirectoryError)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(formatfilesSetupRoot)
	os.Exit(code)
}

// Preparation has no test deadline. The gate's process timeout covers cold
// builds; each leaf starts its own deadline only after fetching its products.
func formatfilesPreparePort(test *testing.T, applied *mutant) formatfilesPrepared {
	test.Helper()
	prepared := formatfilesPrepared{source: formatfilesSetupPort(test, applied)}
	prepared.program = lowered(test, filepath.Join(prepared.source, "main.ts"))
	prepared.binary = formatfilesBinary(test, prepared.program, applied)
	if applied == nil {
		prepared.javascript = filepath.Join(formatfilesSetupDirectory(test), "program.mjs")
		if writeError := os.WriteFile(
			prepared.javascript,
			[]byte(javascript.JavaScript(prepared.program)),
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return prepared
}

func formatfilesPrepareShard(test *testing.T, selected int) *formatfilesShared {
	test.Helper()
	state := *formatfilesReady(test)
	state.port = formatfilesPreparedPort(test)
	if runtime.GOOS == "darwin" {
		state.port.unsanitized = formatfilesUnsanitized(test)
	}
	state.mutated = make(map[string]formatfilesPrepared)
	for index, mutant := range mutants {
		if formatfilesShard("mutant/"+mutant.name) != selected {
			continue
		}
		state.mutated[mutant.name] = formatfilesPreparedMutant(test, index)
	}
	return &state
}

// The product units and shards share these once-per-process entry points.
func formatfilesPreparedPort(test *testing.T) formatfilesPrepared {
	test.Helper()
	formatfilesPortOnce.Do(func() { formatfilesPortState = formatfilesPreparePort(test, nil) })
	if formatfilesPortState.binary == "" {
		test.Fatal("shared port preparation failed")
	}
	return formatfilesPortState
}

func formatfilesPreparedMutant(test *testing.T, index int) formatfilesPrepared {
	test.Helper()
	formatfilesMutantOnce[index].Do(func() {
		formatfilesMutantState[index] = formatfilesPreparePort(test, &mutants[index])
	})
	if formatfilesMutantState[index].binary == "" {
		test.Fatalf("mutant %q preparation failed", mutants[index].name)
	}
	return formatfilesMutantState[index]
}

func formatfilesUnsanitized(test *testing.T) string {
	test.Helper()
	formatfilesUnsanitizedOnce.Do(func() {
		source := formatfilesSetupPort(test, nil)
		program := lowered(test, filepath.Join(source, "main.ts"))
		formatfilesUnsanitizedState = formatfilesBinaryWithOptions(test, program, nil, false)
	})
	if formatfilesUnsanitizedState == "" {
		test.Fatal("unsanitized port preparation failed")
	}
	return formatfilesUnsanitizedState
}

func formatfilesSetupDirectory(test *testing.T) string {
	test.Helper()
	directory, temporaryDirectoryError := os.MkdirTemp(formatfilesSetupRoot, "fixture-")
	if temporaryDirectoryError != nil {
		test.Fatal(temporaryDirectoryError)
	}
	return directory
}

func formatfilesReady(test *testing.T) *formatfilesShared {
	test.Helper()
	formatfilesCorpusOnce.Do(func() {
		state := &formatfilesShared{}
		state.casesPath, state.answers = formatfilesAskedCases(test)
		state.parts = formatfilesPartition(test, state.casesPath, state.answers)
		formatfilesSharedState = state
	})
	if formatfilesSharedState == nil {
		test.Fatal("shared corpus preparation failed")
	}
	return formatfilesSharedState
}

// Call only after shared readiness. All children of a leaf share this deadline.
func formatfilesDeadline(test *testing.T, label string) context.Context {
	test.Helper()
	started := time.Now()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	test.Cleanup(func() {
		cancel()
		test.Logf("%s own wall %.3f s (setup wait excluded)", label, time.Since(started).Seconds())
	})
	return timeoutContext
}

func formatfilesContextCommand(callerContext context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

// formatfilesExecute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's
// own copying under a two-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF two seconds
// after the child exited, so a Node, native or leaks run that had finished could fail its shard under load (#n3gk22h's
// shape). ctx, the leaf's 90 s deadline, ends a hung run.
func formatfilesExecute(
	test *testing.T,
	callerContext context.Context,
	environment []string,
	name string,
	args ...string,
) run {
	test.Helper()
	command := formatfilesContextCommand(callerContext, name, args...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(callerContext, command)
	if callerContext.Err() != nil {
		test.Fatalf("cooked: 90 s shard deadline: %v", callerContext.Err())
	}
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

func formatfilesNode(test *testing.T, callerContext context.Context, path string, args ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return formatfilesExecute(
		test,
		callerContext,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...,
	)
}

func formatfilesLeaks(
	test *testing.T,
	callerContext context.Context,
	program *ir.Program,
	binary string,
	args ...string,
) string {
	test.Helper()
	if runtime.GOOS == "linux" {
		report := formatfilesExecute(test, callerContext, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, args...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	if runtime.GOOS == "darwin" {
		report := formatfilesExecute(
			test,
			callerContext,
			nil,
			"leaks",
			append([]string{"--atExit", "--", formatfilesUnsanitizedState}, args...)...,
		)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// Run native builds in a subprocess, outside the leaf deadline.
func init() {
	source := os.Getenv("ADAMIC_FORMATFILES_NATIVE_SOURCE")
	if source == "" {
		return
	}
	contents, err := os.ReadFile(source)
	if err == nil {
		err = native.Build(
			string(contents),
			os.Getenv("ADAMIC_FORMATFILES_NATIVE_OUTPUT"),
			native.Options{Sanitize: os.Getenv("ADAMIC_FORMATFILES_NATIVE_SANITIZE") == "true"},
		)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func formatfilesBuildNative(test *testing.T, source, output string, sanitize bool) error {
	path := filepath.Join(filepath.Dir(output), "program.c")
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		return writeError
	}
	command := formatfilesContextCommand(test.Context(), os.Args[0])
	command.Env = append(
		os.Environ(),
		"ADAMIC_FORMATFILES_NATIVE_SOURCE="+path,
		"ADAMIC_FORMATFILES_NATIVE_OUTPUT="+output,
		fmt.Sprintf("ADAMIC_FORMATFILES_NATIVE_SANITIZE=%t", sanitize),
	)
	// Read through childguard.Capture for formatfilesExecute's reason: a WaitDelay could fail a build that had
	// finished. The test's context has no deadline, so the build stays outside the leaf's; the unit's ceiling bounds
	// it.
	stdout, stderr, captureError := childguard.Capture(test.Context(), command)
	if captureError != nil {
		return fmt.Errorf("native builder: %w\n%s%s", captureError, stdout, stderr)
	}
	return nil
}
func formatfilesSetupPort(test *testing.T, applied *mutant) string {
	test.Helper()
	directory := formatfilesSetupDirectory(test)
	port := filepath.Join(directory, "formatfiles")
	gitignore := filepath.Join(directory, "gitignore")
	for _, made := range []string{port, gitignore} {
		if createDirectoryError := os.Mkdir(made, 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	for _, name := range gitignoreFiles {
		contents, readError := os.ReadFile(filepath.Join("..", "gitignore", name))
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(filepath.Join(gitignore, name), contents, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	for _, name := range portFiles {
		contents, readError := os.ReadFile(name)
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(contents)
		if applied != nil && applied.file == name {
			if strings.Count(source, applied.from) != 1 {
				test.Fatalf("the mutant %q must change exactly one place in %s", applied.name, name)
			}
			source = strings.Replace(source, applied.from, applied.to, 1)
		}
		if writeError := os.WriteFile(filepath.Join(port, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return port
}
