package markdownblocks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/system-inc/adamic/internal/childguard"
)

func tableLayoutCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
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

// tableLayoutExecute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's
// own copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second
// after the child exited, so a run that had finished could fail its case under load (#n3gk22h's shape).
func tableLayoutExecute(ctx context.Context, environment []string, name string, args ...string) (run, error) {
	command := tableLayoutCommand(ctx, name, args...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	if ctx.Err() != nil {
		return run{}, fmt.Errorf("cooked: table layout case deadline exceeded: %w", ctx.Err())
	}
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		return run{}, fmt.Errorf("running %s: %w", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, nil
}

func tableLayoutNode(ctx context.Context, path string, args ...string) (run, error) {
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		return run{}, absolutePathError
	}
	return tableLayoutExecute(
		ctx,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...,
	)
}
