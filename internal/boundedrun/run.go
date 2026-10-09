// Package boundedrun bounds child lifetimes and kills their complete process groups.
package boundedrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Probe allows slow network filesystems and toolchain discovery. Build allows cold
// Go/toolchain work (observed setup 32.8s; historical cold setup 114s). The
// gate records a 918s package. Shard exceeds Go's existing 60m
// test timeout by 10m for compilation and reporting. Execution limits stay caller-owned.
const (
	Probe = 30 * time.Second
	Build = 10 * time.Minute
	Shard = 70 * time.Minute
)

// WithTimeout supplies the execution clock. The default is the standard immediate
// deadline; nonparallel tests may replace it to arm only after child readiness.
var WithTimeout = context.WithTimeout

// Cmd owns a child and its bounded wait.
type Cmd struct {
	*exec.Cmd
	context context.Context
}

// Command bounds execution and pipe draining. Release its context after reaping,
// including failed starts. Successful completion also stops residual descendants.
func Command(limit time.Duration, name string, args ...string) (*Cmd, func()) {
	if cap, err := strconv.ParseFloat(os.Getenv("ADAMIC_CHILD_DEADLINE"), 64); err == nil && cap > 0 && cap < limit.Seconds() {
		limit = time.Duration(cap * float64(time.Second))
	}
	ctx, cancel := WithTimeout(context.Background(), limit)
	command := CommandContext(ctx, name, args...)
	return command, cancel
}

func (cmd *Cmd) finish(err error) error {
	_ = Kill(cmd.Cmd)
	if cmd.context.Err() != nil {
		return fmt.Errorf("child %s: %w; process group killed", cmd.Path, context.Cause(cmd.context))
	}
	if errors.Is(err, exec.ErrWaitDelay) {
		return fmt.Errorf("child %s: output pipe deadline exceeded; process group killed: %w", cmd.Path, err)
	}
	return err
}
func (cmd *Cmd) Wait() error {
	done := make(chan error, 1)
	go func() { done <- cmd.Cmd.Wait() }()
	select {
	case err := <-done:
		return cmd.finish(err)
	case <-cmd.context.Done():
		_ = Kill(cmd.Cmd)
		// Reaping itself must not hold the tool forever (for example a child stuck
		// in kernel I/O). SIGKILL is issued immediately, with at most 2s to reap.
		timer := time.NewTimer(2 * time.Second)
		defer timer.Stop()
		select {
		case err := <-done:
			return cmd.finish(err)
		case <-timer.C:
			return cmd.finish(cmd.context.Err())
		}
	}
}
func (cmd *Cmd) Run() error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
func (cmd *Cmd) CombinedOutput() ([]byte, error) {
	if cmd.Stdout != nil || cmd.Stderr != nil {
		return nil, errors.New("boundedrun: output already set")
	}
	var output captureBuffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	return output.Bytes(), err
}
func (cmd *Cmd) Output() ([]byte, error) {
	if cmd.Stdout != nil {
		return nil, errors.New("boundedrun: stdout already set")
	}
	var output captureBuffer
	cmd.Stdout = &output
	err := cmd.Run()
	return output.Bytes(), err
}

func CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	if _, ok := ctx.Deadline(); !ok {
		panic("boundedrun: child deadline required")
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		fmt.Fprintf(os.Stderr, "child %s %v: deadline/cancellation expired; killing process group; descendants=%v\n", name, args, descendantNames(cmd.Process.Pid))
		return Kill(cmd)
	}
	// WaitDelay also covers a leader that exited while a descendant holds a pipe.
	cmd.WaitDelay = time.Second
	return &Cmd{cmd, ctx}
}

// Kill stops all members of the child's group, including grandchildren.
func Kill(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if err == syscall.ESRCH {
		return os.ErrProcessDone
	}
	return err
}

// Register isolates an operation whose dependency starts unbounded children. The
// enclosing group deadline covers every inherited clang/ar child without changing
// native's runtime cache keys. Registration also works in Go test executables.
func Register(name string, operation func([]string) (string, error)) {
	if len(os.Args) < 2 || os.Args[1] != "--adamic-bounded-"+name {
		return
	}
	value, err := operation(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(value)
	os.Exit(0)
}

func Operation(limit time.Duration, name string, args ...string) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	command, release := Command(limit, executable, append([]string{"--adamic-bounded-" + name}, args...)...)
	defer release()
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("child %s: %w: %s", name, err, output)
	}
	return string(output), nil
}

func descendantNames(pid int) []string {
	tasks, err := filepath.Glob(fmt.Sprintf("/proc/%d/task/*/children", pid))
	if err != nil {
		return nil
	}
	var children []byte
	for _, task := range tasks {
		data, _ := os.ReadFile(task)
		children = append(children, data...)
		children = append(children, ' ')
	}

	var names []string
	for _, child := range strings.Fields(string(children)) {
		number, err := strconv.Atoi(child)
		if err != nil {
			continue
		}
		name, _ := os.ReadFile(fmt.Sprintf("/proc/%d/comm", number))
		names = append(names, strings.TrimSpace(string(name)))
		names = append(names, descendantNames(number)...)
	}
	return names
}

// Copies may still be winding down after the bounded reap grace. Snapshot under
// a lock so a returned capture never races with a late writer.
type captureBuffer struct {
	mutex sync.Mutex
	bytes.Buffer
}

func (buffer *captureBuffer) Write(data []byte) (int, error) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	return buffer.Buffer.Write(data)
}
func (buffer *captureBuffer) Bytes() []byte {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	return append([]byte(nil), buffer.Buffer.Bytes()...)
}
