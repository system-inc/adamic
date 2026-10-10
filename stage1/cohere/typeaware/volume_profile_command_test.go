package typeaware

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// A compiler can spawn clang or other compiler children. Cancel the whole
// process group at the context deadline on both Linux and macOS.
func volumeProfileCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
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
	return command
}
