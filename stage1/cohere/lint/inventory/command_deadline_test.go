package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// Bound every child without an external timeout tool. A shorter parent shard
// deadline still wins. Killing the group also stops Go's compiler descendants.
func inventoryEngineCommand(parent context.Context, name string, arguments ...string) (*exec.Cmd, context.CancelFunc) {
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
	command.WaitDelay = time.Second
	return command, cancel
}

// Cancel before go test itself exits, leaving a second for pipe cleanup. This
// prevents a killed test process from abandoning a compiler's separate group.
func inventoryEngineTestCommand(t *testing.T, name string, arguments ...string) (*exec.Cmd, context.CancelFunc) {
	t.Helper()
	parent := context.Background()
	cancelParent := func() {}
	if deadline, ok := t.Deadline(); ok {
		parent, cancelParent = context.WithDeadline(parent, deadline.Add(-time.Second))
	}
	command, cancel := inventoryEngineCommand(parent, name, arguments...)
	return command, func() { cancel(); cancelParent() }
}
