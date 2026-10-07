package oracle

import (
	"context"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// The 2048-object moves fixture takes about 98 seconds to compile for WASI
// even in isolation. Give compilation a bounded five-minute budget, as the
// parallel TSan executions have. Ordinary execution deadlines remain unchanged.
func boundedCompile(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	return command
}
