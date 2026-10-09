package formatfiles

import (
	"bytes"
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
	var err error
	formatfilesSetupRoot, err = os.MkdirTemp("", "formatfiles-shared-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(formatfilesSetupRoot)
	os.Exit(code)
}

// Preparation has no test deadline. The gate's process timeout covers cold
// builds; each leaf starts its own deadline only after fetching its products.
func formatfilesPreparePort(t *testing.T, applied *mutant) formatfilesPrepared {
	t.Helper()
	prepared := formatfilesPrepared{source: formatfilesSetupPort(t, applied)}
	prepared.program = lowered(t, filepath.Join(prepared.source, "main.ts"))
	prepared.binary = formatfilesBinary(t, prepared.program, applied)
	if applied == nil {
		prepared.javascript = filepath.Join(formatfilesSetupDirectory(t), "program.mjs")
		if err := os.WriteFile(prepared.javascript, []byte(javascript.JavaScript(prepared.program)), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return prepared
}

func formatfilesPrepareShard(t *testing.T, selected int) *formatfilesShared {
	t.Helper()
	state := *formatfilesReady(t)
	state.port = formatfilesPreparedPort(t)
	if runtime.GOOS == "darwin" {
		state.port.unsanitized = formatfilesUnsanitized(t)
	}
	state.mutated = make(map[string]formatfilesPrepared)
	for index, mutant := range mutants {
		if formatfilesShard("mutant/"+mutant.name) != selected {
			continue
		}
		state.mutated[mutant.name] = formatfilesPreparedMutant(t, index)
	}
	return &state
}

// The product units and shards share these once-per-process entry points.
func formatfilesPreparedPort(t *testing.T) formatfilesPrepared {
	t.Helper()
	formatfilesPortOnce.Do(func() { formatfilesPortState = formatfilesPreparePort(t, nil) })
	if formatfilesPortState.binary == "" {
		t.Fatal("shared port preparation failed")
	}
	return formatfilesPortState
}

func formatfilesPreparedMutant(t *testing.T, index int) formatfilesPrepared {
	t.Helper()
	formatfilesMutantOnce[index].Do(func() {
		formatfilesMutantState[index] = formatfilesPreparePort(t, &mutants[index])
	})
	if formatfilesMutantState[index].binary == "" {
		t.Fatalf("mutant %q preparation failed", mutants[index].name)
	}
	return formatfilesMutantState[index]
}

func formatfilesUnsanitized(t *testing.T) string {
	t.Helper()
	formatfilesUnsanitizedOnce.Do(func() {
		source := formatfilesSetupPort(t, nil)
		program := lowered(t, filepath.Join(source, "main.ts"))
		formatfilesUnsanitizedState = formatfilesBinaryWithOptions(t, program, nil, false)
	})
	if formatfilesUnsanitizedState == "" {
		t.Fatal("unsanitized port preparation failed")
	}
	return formatfilesUnsanitizedState
}

func formatfilesSetupDirectory(t *testing.T) string {
	t.Helper()
	directory, err := os.MkdirTemp(formatfilesSetupRoot, "fixture-")
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

func formatfilesReady(t *testing.T) *formatfilesShared {
	t.Helper()
	formatfilesCorpusOnce.Do(func() {
		state := &formatfilesShared{}
		state.casesPath, state.answers = formatfilesAskedCases(t)
		state.parts = formatfilesPartition(t, state.casesPath, state.answers)
		formatfilesSharedState = state
	})
	if formatfilesSharedState == nil {
		t.Fatal("shared corpus preparation failed")
	}
	return formatfilesSharedState
}

// Call only after shared readiness. All children of a leaf share this deadline.
func formatfilesDeadline(t *testing.T, label string) context.Context {
	t.Helper()
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(func() {
		cancel()
		t.Logf("%s own wall %.3f s (setup wait excluded)", label, time.Since(started).Seconds())
	})
	return ctx
}

func formatfilesContextCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 2 * time.Second
	return command
}

func formatfilesExecute(t *testing.T, ctx context.Context, environment []string, name string, args ...string) run {
	t.Helper()
	command := formatfilesContextCommand(ctx, name, args...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("cooked: 90 s shard deadline: %v", ctx.Err())
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

func formatfilesNode(t *testing.T, ctx context.Context, path string, args ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return formatfilesExecute(t, ctx, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...)
}

func formatfilesLeaks(t *testing.T, ctx context.Context, program *ir.Program, binary string, args ...string) string {
	t.Helper()
	if runtime.GOOS == "linux" {
		report := formatfilesExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, args...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	if runtime.GOOS == "darwin" {
		report := formatfilesExecute(t, ctx, nil, "leaks", append([]string{"--atExit", "--", formatfilesUnsanitizedState}, args...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
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
		err = native.Build(string(contents), os.Getenv("ADAMIC_FORMATFILES_NATIVE_OUTPUT"), native.Options{Sanitize: os.Getenv("ADAMIC_FORMATFILES_NATIVE_SANITIZE") == "true"})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func formatfilesBuildNative(t *testing.T, source, output string, sanitize bool) error {
	path := filepath.Join(filepath.Dir(output), "program.c")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		return err
	}
	command := formatfilesContextCommand(context.Background(), os.Args[0])
	command.Env = append(os.Environ(), "ADAMIC_FORMATFILES_NATIVE_SOURCE="+path, "ADAMIC_FORMATFILES_NATIVE_OUTPUT="+output, fmt.Sprintf("ADAMIC_FORMATFILES_NATIVE_SANITIZE=%t", sanitize))
	result, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("native builder: %w\n%s", err, result)
	}
	return nil
}
func formatfilesSetupPort(t *testing.T, applied *mutant) string {
	t.Helper()
	directory := formatfilesSetupDirectory(t)
	port := filepath.Join(directory, "formatfiles")
	gitignore := filepath.Join(directory, "gitignore")
	config := filepath.Join(directory, "config")
	for _, made := range []string{port, gitignore, config} {
		if err := os.Mkdir(made, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// enumerate imports the shared config glob implementation.
	contents, err := os.ReadFile("../config/glob.ts")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config, "glob.ts"), contents, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range gitignoreFiles {
		contents, err := os.ReadFile(filepath.Join("..", "gitignore", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(gitignore, name), contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range portFiles {
		contents, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		source := string(contents)
		if applied != nil && applied.file == name {
			if strings.Count(source, applied.from) != 1 {
				t.Fatalf("the mutant %q must change exactly one place in %s", applied.name, name)
			}
			source = strings.Replace(source, applied.from, applied.to, 1)
		}
		if err := os.WriteFile(filepath.Join(port, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return port
}
