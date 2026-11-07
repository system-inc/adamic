package css

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

// cssParserExecute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after
// the child exited, so a port, Node, leaks, sanitizer or probe run that had finished could fail its shard under load
// (#n3gk22h's shape). Capture has no stall guard, and the runs lose nothing by it: childguard.Run's windows (two
// minutes of silence, thirty before the first byte) are longer than the command's 90 s deadline, which ends a hung run
// sooner.
func cssParserExecute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command, ctx := cssParserCommand(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}
func cssParserOnNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return cssParserExecute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

// cssParserCommand bounds each owned child, including its compiler descendants. It returns the command's context with
// it, for childguard.Capture to end the reads by.
func cssParserCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	return cssParserCommandContext(test, timeoutContext, cancel, name, arguments...), timeoutContext
}

func cssParserCommandContext(
	test *testing.T,
	callerContext context.Context,
	cancel context.CancelFunc,
	name string,
	arguments ...string,
) *exec.Cmd {
	test.Helper()
	command := toolchain.CommandContextIn(callerContext, "", name, arguments...)
	command.Env = os.Environ()
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
	test.Cleanup(func() {
		cancel()
		if command.Process != nil && command.ProcessState == nil {
			_ = command.Cancel()
		}
	})
	return command
}

// Setup retains process-group cancellation without a deadline of its own. It returns the command's context with it, for
// childguard.Capture to end the reads by.
func compositionSetupCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	cancelContext, cancel := context.WithCancel(context.Background())
	test.Cleanup(cancel)
	return compositionCommandContext(cancelContext, name, arguments...), cancelContext
}

func compositionCommandContext(callerContext context.Context, name string, arguments ...string) *exec.Cmd {
	cmd := toolchain.CommandContextIn(callerContext, "", name, arguments...)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		killError := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return cmd
}
