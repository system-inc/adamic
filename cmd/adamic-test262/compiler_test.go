package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
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
		expected := runCommandWithLimit(2*time.Minute, nil, 16<<20, e.adamic, "c", source)
		actual := worker.compile(source)
		expectedKind, expectedReason := compileClass(expected.Stderr, expected.Exit, expected.TimedOut)
		actualKind, actualReason := compileClass(actual.Stderr, actual.Exit, actual.TimedOut)
		if actual.Stdout != expected.Stdout || actual.Exit != expected.Exit || expectedKind != actualKind || expectedReason != actualReason {
			t.Fatalf("compiler differs: actual=%+v expected=%+v", actual, expected)
		}
	}
}

func TestCompilerHangHelper(t *testing.T) {
	if os.Getenv("ADAMIC_TEST262_HANG_HELPER") != "1" {
		return
	}
	time.Sleep(time.Hour)
}

func TestCompilerWorkerTimeout(t *testing.T) {
	t.Parallel()
	child, heartbeat := testfixture.Tree(t, "fake-worker")
	command, release := boundedrun.Command(boundedrun.Build, child)
	defer release()

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
	worker := &compilerWorker{command: command.Cmd, input: input, output: output, encoder: json.NewEncoder(input), decoder: json.NewDecoder(output), timeout: 200 * time.Millisecond}
	defer worker.close()
	started := testfixture.WaitTree(t, heartbeat)
	result := worker.compile("unused.a")
	testfixture.AssertStopped(t, started, heartbeat, result.Stderr, child)
	if !result.TimedOut || result.Exit != -1 || worker.command != nil || command.ProcessState == nil {
		t.Fatalf("deadline failed to stop and reap compiler: %+v", result)
	}
}
