package typeaware

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// A compiler can spawn clang or other compiler children. Cancel the whole
// process group at the context deadline on both Linux and macOS.
func volumeProfileCommand(callerContext context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}
