//go:build unix

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// errorRecorder is a testing.TB that keeps what Errorf was handed instead of failing.
type errorRecorder struct {
	testing.TB
	messages []string
}

func (recorder *errorRecorder) Helper() {}

func (recorder *errorRecorder) Errorf(format string, arguments ...any) {
	recorder.messages = append(recorder.messages, fmt.Sprintf(format, arguments...))
}

func exitErrorOf(t *testing.T, script string) *exec.ExitError {
	t.Helper()
	var exitError *exec.ExitError
	if err := exec.Command("sh", "-c", script).Run(); !errors.As(err, &exitError) {
		t.Fatalf("sh -c %q ended with %v, want an exit error", script, err)
	}
	return exitError
}

// A child a signal ended is named as killed, by its signal, and a child that chose its exit is not.
func TestChildExitCodeNamesTheSignalThatEndedAChild(t *testing.T) {
	t.Parallel()

	killed := &errorRecorder{TB: t}
	if code := childExitCode(killed, exitErrorOf(t, "kill -KILL $$")); code != -1 {
		t.Errorf("a killed child read as exit %d, want -1", code)
	}
	if len(killed.messages) != 1 || !strings.Contains(killed.messages[0], "killed by signal 9 (killed)") {
		t.Errorf("a killed child was reported as %q", killed.messages)
	}

	chose := &errorRecorder{TB: t}
	if code := childExitCode(chose, exitErrorOf(t, "exit 3")); code != 3 || len(chose.messages) != 0 {
		t.Errorf("a child that exited 3 read as %d with %q", code, chose.messages)
	}
}
