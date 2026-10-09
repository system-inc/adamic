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
)

func TestCompilerWorkerMatchesSubprocess(t *testing.T) {
	t.Parallel()
	e, err := prepare("../..", "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	worker := &compilerWorker{}
	defer worker.close()
	for _, body := range []string{"assertSameValue(Math.abs(-4), 4);", "var value = 1;", "const value: number = 'bad';", "throw new Error('fixture');"} {
		source := filepath.Join(t.TempDir(), "program.a")
		if err := os.WriteFile(source, []byte(program(body)), 0600); err != nil {
			t.Fatal(err)
		}
		// The reference is our own compiler. Give its I/O drains time under load
		// without weakening the comparison of completed compiler results.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		command := exec.CommandContext(ctx, e.adamic, "c", source)
		command.WaitDelay = 10 * time.Second
		stdout, stderr := limitedBuffer{limit: 16 << 20}, limitedBuffer{limit: 16 << 20}
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		cancel()
		if errors.Is(err, exec.ErrWaitDelay) {
			t.Fatalf("reference compiler I/O still incomplete under load after 10s: %v", err)
		}
		var exitError *exec.ExitError
		if err != nil && !errors.As(err, &exitError) {
			t.Fatalf("reference compiler failed: %v", err)
		}
		if ctx.Err() == context.DeadlineExceeded || stdout.exceeded || stderr.exceeded {
			t.Fatal("reference compiler exceeded its deadline or output limit")
		}
		expected := execution{Stdout: stdout.String(), Stderr: stderr.String()}
		if exitError != nil {
			expected.Exit = exitError.ExitCode()
		}
		actual := worker.compile(source)
		expectedKind, expectedReason := compileClass(expected.Stderr, expected.Exit, expected.TimedOut)
		actualKind, actualReason := compileClass(actual.Stderr, actual.Exit, actual.TimedOut)
		if actual.Stdout != expected.Stdout || actual.Exit != expected.Exit || expectedKind != actualKind || expectedReason != actualReason {
			t.Fatalf("compiler differs: actual=%+v expected=%+v", actual, expected)
		}
	}
}

func TestCompilerHangHelper(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_TEST262_HANG_HELPER") != "1" {
		return
	}
	time.Sleep(time.Hour)
}

func TestCompilerWorkerTimeout(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestCompilerHangHelper$")
	command.Env = append(os.Environ(), "ADAMIC_TEST262_HANG_HELPER=1")
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	worker := &compilerWorker{command: command, input: input, output: output, encoder: json.NewEncoder(input), decoder: json.NewDecoder(output), timeout: 20 * time.Millisecond}
	defer worker.close()
	result := worker.compile("unused.a")
	if !result.TimedOut || result.Exit != -1 || worker.command != nil || command.ProcessState == nil {
		t.Fatalf("deadline failed to stop and reap compiler: %+v", result)
	}
}
