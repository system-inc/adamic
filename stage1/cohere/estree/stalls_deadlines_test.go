package estree

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Context cancellation kills the entire child group, including descendants that
// retain stdout/stderr or run compilers. This uses Go and Unix APIs on macOS/Linux.
// Callers read its output through childguard.Capture, to its end with no timer
// deciding when: exec's own copying under a two-second WaitDelay reported
// exec.ErrWaitDelay whenever the reads hadn't reached EOF two seconds after the
// child exited, so a run that had finished could fail its shard under load
// (#n3gk22h's shape). Each caller's 90 s deadline ends a hung run.
func estreeScopedCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	return command
}

func estreeScopedExecute(t *testing.T, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	output, stderr, err := childguard.Capture(ctx, estreeScopedCommand(ctx, name, args...))
	if ctx.Err() != nil {
		t.Fatalf("cooked: %s %v exceeded its 90s child deadline", name, args)
	}
	if err != nil || len(stderr) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, stderr)
	}
	return output
}

func estreeScopedNode(t *testing.T, program, path string) []byte {
	t.Helper()
	return estreeScopedExecute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), program, path)
}

// A refusal held to its calibrated CPU budget (estreeRefusalBudget) through the declared Go test executable, while also
// bounding startup and wall time with a 90s context/group deadline.
func estreeScopedRefusal(t *testing.T, argv []string, diagnostic string) {
	t.Helper()
	budget := estreeRefusalBudget(t, argv)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	args := append([]string{"-test.run=^TestDeadlineChild$", "-test.timeout=75s", "--", strconv.FormatUint(estreeBudgetSeconds(budget), 10)}, argv...)
	command := estreeScopedCommand(ctx, os.Args[0], args...)
	command.Env = append(os.Environ(), "ADAMIC_ESTREE_DEADLINE_CHILD=1")
	command.ExtraFiles = []*os.File{writer}
	// Captured, not exec's copying under a WaitDelay: exec.ErrWaitDelay is an error with a ProcessState, so a child
	// that exited cleanly could pass the err check below as a refusal, and one that refused could lose its diagnostic.
	output, stderr, err := childguard.Capture(ctx, command)
	writer.Close()
	ready, readErr := io.ReadAll(reader)
	if ctx.Err() != nil {
		t.Fatalf("cooked: %v exceeded its 90s child deadline", argv)
	}
	if readErr != nil || string(ready) != "started\n" {
		t.Fatalf("%v: child never started: %v %q\n%s", argv, readErr, ready, stderr)
	}
	cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
	status := command.ProcessState.Sys().(syscall.WaitStatus)
	timedOut := cpu >= budget || (status.Signaled() && status.Signal() == syscall.SIGXCPU)
	if timedOut || err == nil || len(output) != 0 || !strings.Contains(string(stderr), diagnostic) {
		t.Fatalf("%v: timeout=%v cpu=%s budget=%s exit=%v stdout=%d stderr=%s", argv, timedOut, cpu, budget, err, len(output), stderr)
	}
}

// A grandchild holds the output pipe open. Killing only its parent would leave
// that pipe open until WaitDelay; group cancellation must close it promptly.
func TestEstreeScopedProcessGroupDeadline(t *testing.T) {
	t.Parallel()
	marker := "ADAMIC_ESTREE_GROUP_DEADLINE"
	switch os.Getenv(marker) {
	case "grandchild":
		fmt.Fprintln(os.Stdout, "grandchild-ready")
		time.Sleep(90 * time.Second)
		return
	case "child":
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		// This descendant deliberately inherits its parent's group, as compilers do.
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestEstreeScopedProcessGroupDeadline$", "-test.timeout=75s")
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, marker+"=") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, marker+"=grandchild")
		command.Stdout, command.Stderr = os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command := estreeScopedCommand(ctx, os.Args[0], "-test.run=^TestEstreeScopedProcessGroupDeadline$", "-test.timeout=75s")
	command.Env = append(os.Environ(), marker+"=child")
	if command.SysProcAttr == nil || !command.SysProcAttr.Setpgid {
		t.Fatal("child must have its own process group")
	}
	// WaitDelay: it bounds the failure this test exists to catch. Were Cancel to kill only the parent, the grandchild
	// would hold CombinedOutput's pipe and this delay would end the read at five seconds, which the four-second check
	// below reports; with none the read would wait out the grandchild's 75 s timeout.
	command.WaitDelay = 2 * time.Second
	defer func() {
		if command.Process != nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
	}()
	started := time.Now()
	output, err := command.CombinedOutput()
	if err == nil || ctx.Err() != context.DeadlineExceeded || !bytes.Contains(output, []byte("grandchild-ready")) || time.Since(started) >= 4*time.Second {
		t.Fatalf("context must kill parent and grandchild promptly: %v context=%v elapsed=%s\n%s", err, ctx.Err(), time.Since(started), output)
	}
}
