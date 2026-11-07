package tsprinter

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
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// firstDifference says where two outputs first differ, line by line, or "" when they don't.
func firstDifference(got string, want string) string {
	if got == want {
		return ""
	}
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want, "\n")
	for index := 0; index < len(gotLines) || index < len(wantLines); index++ {
		var gotLine, wantLine string
		if index < len(gotLines) {
			gotLine = gotLines[index]
		}
		if index < len(wantLines) {
			wantLine = wantLines[index]
		}
		if gotLine != wantLine {
			return fmt.Sprintf("line %d: %q, Go cohere %q", index+1, gotLine, wantLine)
		}
	}
	return "the same lines, not the same bytes"
}

// lowered checks and lowers a program, failing the test with stage 0's refusal if it can't.
func lowered(test *testing.T, path string) *ir.Program {
	test.Helper()
	program, err := lowerProgram(path)
	if err != nil {
		test.Fatal(err)
	}
	return program
}

func lowerProgram(path string) (*ir.Program, error) {
	program, err := load.Load([]string{path})
	if err != nil {
		return nil, fmt.Errorf("Load: %w", err)
	}
	result, err := lower.Lower(context.Background(), program)
	if err != nil {
		return nil, fmt.Errorf("Lower: %w", err)
	}
	return result, nil
}

func execute(test *testing.T, environment []string, name string, arguments ...string) oracles.Run {
	test.Helper()
	if result, ok := executeShards(test, environment, name, arguments); ok {
		return result
	}
	return executeOne(test, environment, name, arguments...)
}

// bounded prepares a child; execute and combinedOutput run it with progress-based guards.
func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	// The child guard owns hang detection, including for Go test children.
	if name == "go" && len(arguments) > 0 && arguments[0] == "test" {
		arguments = append([]string{"test", "-timeout=0"}, arguments[1:]...)
	}
	return exec.Command(name, arguments...)
}

func combinedOutput(command *exec.Cmd) ([]byte, error) {
	return childguard.CombinedOutput(command, childguard.Options{})
}

func executeOne(test *testing.T, environment []string, name string, arguments ...string) oracles.Run {
	test.Helper()
	command := bounded(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		test.Fatalf("running %s: %v", name, err)
	}
	return oracles.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(test *testing.T, path string, arguments ...string) oracles.Run {
	test.Helper()
	runner, err := filepath.Abs(filepath.Join(oracles.Repository, "oracle", "node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	return execute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(test *testing.T, program *ir.Program, arguments ...string) oracles.Run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		test.Fatal(err)
	}
	return onNode(test, path, arguments...)
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(test *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	test.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(test.TempDir(), "port")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			test.Fatal(err)
		}
		report := execute(test, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.ExitCode == 0 {
			return ""
		}
		return string(report.Stdout)
	case "linux":
		report := execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.ExitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.ExitCode, report.Stderr)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// output keeps stdout and stderr separate for git corpus queries.
func output(command *exec.Cmd) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	captureStderr := command.Stderr == nil
	if captureStderr {
		command.Stderr = &stderr
	}
	err := childguard.Run(command, childguard.Options{})
	if exitError, ok := err.(*exec.ExitError); ok && captureStderr {
		exitError.Stderr = stderr.Bytes()
	}
	return stdout.Bytes(), err
}
