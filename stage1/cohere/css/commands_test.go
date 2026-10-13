package css

import (
	"bytes"
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// cssParserCommand bounds each owned child, including its compiler descendants.
func cssParserCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	return cssParserCommandContext(t, ctx, cancel, name, arguments...)
}

// Shared setup has no deadline; cancellation still kills the process group.
func cssParserSetupCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	return cssParserCommandContext(t, ctx, cancel, name, arguments...)
}

// cssParserGoCommand is cssParserSetupCommand for go as toolchain resolves it, run in directory under toolchain.Env,
// the environment whose compilers the Go oracle's key names.
func cssParserGoCommand(t *testing.T, directory string, arguments ...string) *exec.Cmd {
	t.Helper()
	resolved := toolchain.CommandIn(directory, "go", arguments...)
	if resolved.Err != nil {
		return resolved
	}
	command := cssParserSetupCommand(t, resolved.Path, arguments...)
	command.Dir, command.Env = resolved.Dir, resolved.Env
	return command
}

// cssCompilers are the C and C++ compilers go builds a Go oracle with, and the C compiler's report.
type cssCompilers struct{ c, cPlusPlus, report string }

// cssGoCompilers reads the compilers toolchain.Env gives go: the caller's CC and CXX, or clang and clang++ when it
// names neither.
func cssGoCompilers() cssCompilers {
	var compilers cssCompilers
	for _, setting := range toolchain.Env() {
		switch name, value, _ := strings.Cut(setting, "="); name {
		case "CC":
			compilers.c = value
		case "CXX":
			compilers.cPlusPlus = value
		}
	}
	compilers.report = "no C compiler"
	if fields := strings.Fields(compilers.c); len(fields) > 0 {
		compilers.report = buildcache.Tool(fields[0], "--version")
	}
	return compilers
}

func cssParserCommandContext(t *testing.T, ctx context.Context, cancel context.CancelFunc, name string, arguments ...string) *exec.Cmd {
	t.Helper()
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
	t.Cleanup(func() {
		cancel()
		if command.Process != nil && command.ProcessState == nil {
			_ = command.Cancel()
		}
	})
	return command
}
func cssParserExecute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := cssParserCommand(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}
func cssParserOnNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return cssParserExecute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}
