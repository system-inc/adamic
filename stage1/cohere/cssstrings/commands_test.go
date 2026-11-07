package cssstrings

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/childguard"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// stringsCommand bounds each owned child, including its compiler descendants. It returns the command's context with it,
// for childguard.Capture to end the reads by.
func stringsCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	return stringsCommandContext(test, ctx, cancel, name, arguments...), ctx
}

// Shared setup has no deadline; cancellation still kills the process group.
func stringsSetupCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	return stringsCommandContext(test, ctx, cancel, name, arguments...), ctx
}

func stringsCommandContext(
	test *testing.T,
	ctx context.Context,
	cancel context.CancelFunc,
	name string,
	arguments ...string,
) *exec.Cmd {
	test.Helper()
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
	test.Cleanup(func() {
		cancel()
		if command.Process != nil && command.ProcessState == nil {
			_ = command.Cancel()
		}
	})
	return command
}

// stringsExecute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after
// the child exited, so a port, Node, leaks or probe run that had finished could fail its shard under load (#n3gk22h's
// shape). Capture has no stall guard, and the runs lose nothing by it: childguard.Run's windows (two minutes of
// silence, thirty before the first byte) are longer than the command's 90 s deadline, which ends a hung run sooner.
func stringsExecute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command, ctx := stringsCommand(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		test.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}
func stringsOnNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	return stringsExecute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}
