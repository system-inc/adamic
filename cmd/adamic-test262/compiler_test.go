package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

func TestCompilerWorkerMatchesSubprocess(test *testing.T) {
	test.Parallel()
	e, prepareError := prepare(test, "testdata/mini", test.TempDir())
	if prepareError != nil {
		test.Fatal(prepareError)
	}
	worker := &compilerWorker{}
	defer worker.close()
	for _, body := range []string{
		"assertSameValue(Math.abs(-4), 4);",
		"var value = 1;",
		"const value: number = 'bad';",
		"throw new Error('fixture');",
	} {
		source := filepath.Join(test.TempDir(), "program.a")
		if writeError := os.WriteFile(source, []byte(program(body)), 0600); writeError != nil {
			test.Fatal(writeError)
		}
		// The reference is our own compiler. Drain both streams completely while
		// retaining the deadline and bounded diagnostic capture.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		command := exec.CommandContext(ctx, e.adamic, "c", source)
		stdout, stderr := limitedBuffer{limit: 16 << 20}, limitedBuffer{limit: 16 << 20}
		captureToError := childguard.CaptureTo(ctx, command, &stdout, &stderr)
		cancel()
		var exitError *exec.ExitError
		if captureToError != nil && !errors.As(captureToError, &exitError) {
			test.Fatalf("reference compiler failed: %v", captureToError)
		}
		if ctx.Err() == context.DeadlineExceeded || stdout.exceeded || stderr.exceeded {
			test.Fatal("reference compiler exceeded its deadline or output limit")
		}
		expected := execution{Stdout: stdout.String(), Stderr: stderr.String()}
		if exitError != nil {
			expected.Exit = exitError.ExitCode()
		}
		actual := worker.compile(source)
		expectedKind, expectedReason := compileClass(expected.Stderr, expected.Exit, expected.TimedOut)
		actualKind, actualReason := compileClass(actual.Stderr, actual.Exit, actual.TimedOut)
		if actual.Stdout != expected.Stdout ||
			actual.Exit != expected.Exit ||
			expectedKind != actualKind ||
			expectedReason != actualReason {
			test.Fatalf("compiler differs: actual=%+v expected=%+v", actual, expected)
		}
	}
}

func TestCompilerHangHelper(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TEST262_HANG_HELPER") != "1" {
		return
	}
	time.Sleep(time.Hour)
}

func TestCompilerWorkerTimeout(test *testing.T) {
	test.Parallel()
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	command := exec.Command(executable, "-test.run=^TestCompilerHangHelper$")
	command.Env = append(os.Environ(), "ADAMIC_TEST262_HANG_HELPER=1")
	input, stdinPipeError := command.StdinPipe()
	if stdinPipeError != nil {
		test.Fatal(stdinPipeError)
	}
	output, stdoutPipeError := command.StdoutPipe()
	if stdoutPipeError != nil {
		test.Fatal(stdoutPipeError)
	}
	if startError := command.Start(); startError != nil {
		test.Fatal(startError)
	}
	worker := &compilerWorker{
		command: command,
		input:   input,
		output:  output,
		encoder: json.NewEncoder(input),
		decoder: json.NewDecoder(output),
		timeout: 20 * time.Millisecond,
	}
	defer worker.close()
	result := worker.compile("unused.a")
	if !result.TimedOut || result.Exit != -1 || worker.command != nil || command.ProcessState == nil {
		test.Fatalf("deadline failed to stop and reap compiler: %+v", result)
	}
}
