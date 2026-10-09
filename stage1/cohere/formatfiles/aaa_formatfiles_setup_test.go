package formatfiles

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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

var formatfilesSetupReady = make(chan struct{})
var formatfilesSharedState *formatfilesShared
var formatfilesSetupRoot string
var formatfilesSetupContext context.Context

// Not parallel: owns shared fixture lifetime and ensures setup is selected before any leaf.
func TestMain(m *testing.M) {
	flag.Parse()
	pattern := flag.Lookup("test.run").Value.String()
	expression, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	needsSetup := expression.MatchString("TestThePortParsesAsGoCohereDoesUnion")
	for index := 0; index < testThePortParsesAsGoCohereDoesShards; index++ {
		needsSetup = needsSetup || expression.MatchString(fmt.Sprintf("TestThePortParsesAsGoCohereDoes_%03d", index))
	}
	if needsSetup {
		_ = flag.Set("test.run", "("+pattern+")|^TestThePortParsesAsGoCohereDoes_Setup$")
	}
	formatfilesSetupRoot, err = os.MkdirTemp("", "formatfiles-shared-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(formatfilesSetupRoot)
	os.Exit(code)
}

// This file sorts before the shard table: setup is scheduled first even with
// -parallel=1. TestMain includes it for an individually selected leaf as well.
// Not parallel: initializes shared fixtures and products before parallel shards run.
func TestThePortParsesAsGoCohereDoes_Setup(t *testing.T) {
	formatfilesSetupReady = make(chan struct{})
	formatfilesSharedState = nil
	defer close(formatfilesSetupReady)
	formatfilesSetupContext = formatfilesDeadline(t, "shared setup")
	state := &formatfilesShared{mutated: make(map[string]formatfilesPrepared)}
	state.casesPath, state.answers = formatfilesAskedCases(t)
	state.parts = formatfilesPartition(t, state.casesPath, state.answers)
	state.port.source = formatfilesSetupPort(t, nil)
	state.port.program = lowered(t, filepath.Join(state.port.source, "main.ts"))
	state.port.binary = formatfilesBinary(t, state.port.program, nil)
	if runtime.GOOS == "darwin" {
		state.port.unsanitized = formatfilesBinaryWithOptions(t, state.port.program, nil, false)
	}
	state.port.javascript = filepath.Join(formatfilesSetupDirectory(t), "program.mjs")
	if err := os.WriteFile(state.port.javascript, []byte(javascript.JavaScript(state.port.program)), 0644); err != nil {
		t.Fatal(err)
	}
	for _, mutant := range mutants {
		prepared := formatfilesPrepared{source: formatfilesSetupPort(t, &mutant)}
		prepared.program = lowered(t, filepath.Join(prepared.source, "main.ts"))
		prepared.binary = formatfilesBinary(t, prepared.program, &mutant)
		state.mutated[mutant.name] = prepared
	}
	if err := formatfilesSetupContext.Err(); err != nil {
		t.Fatalf("cooked: setup deadline: %v", err)
	}
	formatfilesSharedState = state
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
	<-formatfilesSetupReady
	if formatfilesSharedState == nil {
		t.Fatal("shared setup failed")
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
		report := formatfilesExecute(t, ctx, nil, "leaks", append([]string{"--atExit", "--", formatfilesSharedState.port.unsanitized}, args...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// A native builder subprocess lets the setup context kill both the builder and
// its compilers as one process group, even though native.Build has no context API.
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
	command := formatfilesContextCommand(formatfilesSetupContext, os.Args[0])
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
	for _, made := range []string{port, gitignore} {
		if err := os.Mkdir(made, 0o755); err != nil {
			t.Fatal(err)
		}
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
