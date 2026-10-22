package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Bound every child without an external timeout tool. A shorter parent shard
// deadline still wins. Killing the group also stops Go's compiler descendants.
// The command's context comes with it, for inventoryEngineOutput to end the reads by.
func inventoryEngineCommand(parent context.Context, name string, arguments ...string) (*exec.Cmd, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return command, ctx, cancel
}

// Cancel before go test itself exits, leaving a second for the group's kill and
// the reads to end. This prevents a killed test process from abandoning a
// compiler's separate group.
func inventoryEngineTestCommand(t *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context, context.CancelFunc) {
	t.Helper()
	parent := context.Background()
	cancelParent := func() {}
	if deadline, ok := t.Deadline(); ok {
		parent, cancelParent = context.WithDeadline(parent, deadline.Add(-time.Second))
	}
	command, ctx, cancel := inventoryEngineCommand(parent, name, arguments...)
	return command, ctx, func() { cancel(); cancelParent() }
}

// inventoryEngineOutput reads a case's output through childguard.Capture, to its end with no timer deciding when: exec's
// own copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second
// after the case exited, so a case that had finished could fail its shard under load (#n3gk22h's shape). ctx, the
// command's own deadline, still ends a hung case, and the deadlines above are unchanged. It returns stdout, then
// stderr; the callers log the output or look for a line anywhere in it.
func inventoryEngineOutput(ctx context.Context, command *exec.Cmd) ([]byte, error) {
	stdout, stderr, err := childguard.Capture(ctx, command)
	return append(stdout, stderr...), err
}
