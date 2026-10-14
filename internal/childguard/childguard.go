// Package childguard runs Unix child process groups with output-based stall protection.
package childguard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
)

const DefaultFirstOutput = 30 * time.Minute
const DefaultStall = 2 * time.Minute
const DefaultCeiling = 60 * time.Minute

// FirstOutput is the silence allowed before the first byte; Stall applies thereafter.
// Options overrides the defaults. Zero durations select the defaults.
type Options struct {
	FirstOutput, Stall, Ceiling time.Duration
	// Unbounded disables the guard timers; command context cancellation still applies.
	Unbounded bool
}

// Error reports a guard termination, separately from a child's nonzero exit.
// Reason is "stalled" or "ceiling". Load is the one-minute load average,
// or "unavailable" on systems without /proc/loadavg.
type Error struct {
	Reason          string
	FirstOutput     bool
	Command         string
	Window, Elapsed time.Duration
	Load            string
}

func (e *Error) Error() string {
	if e.Reason == "stalled" {
		window := "no output"
		if e.FirstOutput {
			window = "no first output"
		}
		return fmt.Sprintf("stalled: %s for %s after %s, load %s (%s)", window, e.Window, e.Elapsed, e.Load, e.Command)
	}
	return fmt.Sprintf("ceiling: %s", e.Elapsed)
}

type activity struct {
	progress chan struct{}
	sync.Mutex
	last time.Time
	seen bool
}
type watched struct {
	dst      io.Writer
	activity *activity
}

func (w *watched) Write(p []byte) (int, error) {
	if len(p) > 0 {
		w.activity.Lock()
		w.activity.last = time.Now()
		w.activity.seen = true
		w.activity.Unlock()
		select {
		case w.activity.progress <- struct{}{}:
		default:
		}
	}
	return w.dst.Write(p)
}

// Run starts an unstarted command in its own process group and waits for it.
// Output is forwarded to the command's original writers (nil means discard).
// Context commands also cancel the whole process group.
func Run(cmd *exec.Cmd, options Options) error {
	if options.FirstOutput == 0 {
		options.FirstOutput = DefaultFirstOutput
	}
	if options.Stall == 0 {
		options.Stall = DefaultStall
	}
	if options.Ceiling == 0 {
		options.Ceiling = DefaultCeiling
	}
	if options.FirstOutput < 0 || options.Stall < 0 || options.Ceiling < 0 {
		return fmt.Errorf("childguard: durations must be positive")
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
	cmd.SysProcAttr.Pgid = 0
	if cmd.Cancel != nil {
		cmd.Cancel = func() error {
			if cmd.Process == nil {
				return os.ErrProcessDone
			}
			err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			if errors.Is(err, syscall.ESRCH) {
				return os.ErrProcessDone
			}
			return err
		}
	}
	originalOut, originalErr := cmd.Stdout, cmd.Stderr
	defer func() { cmd.Stdout, cmd.Stderr = originalOut, originalErr }()
	start := time.Now()
	a := &activity{last: start, progress: make(chan struct{}, 1)}
	wrap := func(dst io.Writer) io.Writer {
		if dst == nil {
			dst = io.Discard
		}
		return &watched{dst: dst, activity: a}
	}
	cmd.Stdout = wrap(originalOut)
	// Preserve exec's serialization when both streams share a writer.
	if originalOut != nil && reflect.TypeOf(originalOut).Comparable() && originalOut == originalErr {
		cmd.Stderr = cmd.Stdout
	} else {
		cmd.Stderr = wrap(originalErr)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if options.Unbounded {
		return cmd.Wait()
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timer := time.NewTimer(min(options.FirstOutput, options.Ceiling))
	defer timer.Stop()
	for {
		select {
		case err := <-done:
			return err
		case <-a.progress:
			timer.Reset(0)
		case <-timer.C:
			select {
			case err := <-done:
				return err
			default:
			}
			now := time.Now()
			a.Lock()
			idle := now.Sub(a.last)
			firstOutput := !a.seen
			window := options.Stall
			if firstOutput {
				window = options.FirstOutput
			}
			a.Unlock()
			elapsed := now.Sub(start)
			reason := ""
			if elapsed >= options.Ceiling {
				reason = "ceiling"
			} else if idle >= window {
				reason = "stalled"
			}
			if reason != "" {
				if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
					return <-done
				}
				<-done
				load := "unavailable"
				if data, err := os.ReadFile("/proc/loadavg"); err == nil {
					if fields := strings.Fields(string(data)); len(fields) > 0 {
						load = fields[0]
					}
				}
				return &Error{Reason: reason, Window: window, FirstOutput: firstOutput, Command: cmd.Path, Elapsed: elapsed, Load: load}
			}
			timer.Reset(min(window-idle, options.Ceiling-elapsed))
		}
	}
}

// Capture runs an unstarted command to completion in its own process group and returns what it wrote to stdout and to
// stderr, each read to its end. No timer decides when the output is complete: once the command exits its group is
// killed, so a process it started that still holds an output pipe (a background job, a sanitizer's helper) can't keep
// a read open, and the reads end when the last writer has. exec.Cmd's own pipes with a WaitDelay instead close the
// reads a fixed time after the exit and report exec.ErrWaitDelay, so a run that was slow to finish them, or a
// descendant that held them, is red by chance (#n3gk22h).
//
// ctx bounds it: when ctx is done the group is killed, and a writer outside the group (one that left it with setsid)
// ends the reads with ctx's error. A command bound to ctx should be exec.CommandContext with the same ctx. Its exit
// status is err's, as cmd.Run reports it.
func Capture(ctx context.Context, cmd *exec.Cmd) (stdout, stderr []byte, err error) {
	if cmd.Stdout != nil || cmd.Stderr != nil {
		return nil, nil, fmt.Errorf("childguard: output already configured")
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
	cmd.SysProcAttr.Pgid = 0
	type pipe struct {
		read, write *os.File
		output      bytes.Buffer
	}
	pipes := make([]*pipe, 2)
	closeAll := func() {
		for _, p := range pipes {
			if p != nil {
				p.read.Close()
				p.write.Close()
			}
		}
	}
	for i := range pipes {
		read, write, err := os.Pipe()
		if err != nil {
			closeAll()
			return nil, nil, err
		}
		pipes[i] = &pipe{read: read, write: write}
	}
	// *os.File writers go to the child as they are, so exec starts no copying of its own and Wait waits only for the
	// process; this process's write ends close once the child holds its own.
	cmd.Stdout, cmd.Stderr = pipes[0].write, pipes[1].write
	defer func() { cmd.Stdout, cmd.Stderr = nil, nil }()
	if err := cmd.Start(); err != nil {
		closeAll()
		return nil, nil, err
	}
	var reads sync.WaitGroup
	for _, p := range pipes {
		p.write.Close()
		reads.Add(1)
		go func() {
			defer reads.Done()
			_, _ = io.Copy(&p.output, p.read)
		}()
	}
	finished := make(chan struct{})
	go func() { reads.Wait(); close(finished) }()
	stop := context.AfterFunc(ctx, func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) })
	err = cmd.Wait()
	stop()
	// The leader is reaped, but while any member of its group lives the group's id can't be reused, so this reaches
	// exactly the processes it left behind, and none when it left none.
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	closeReads := func() {
		for _, p := range pipes {
			p.read.Close()
		}
	}
	if held := captureEnd(ctx, finished, closeReads); err == nil && held != nil {
		err = fmt.Errorf("childguard: %s's output is still held by a process outside its group: %w", cmd.Path, held)
	}
	closeReads()
	return pipes[0].output.Bytes(), pipes[1].output.Bytes(), err
}

// captureEnd waits for Capture's reads to finish, or for ctx: then it closes the reads, waits for them to stop and
// returns ctx's error. Reads that finished as ctx ended are whole, whichever a select would pick, so they return nil.
func captureEnd(ctx context.Context, finished <-chan struct{}, closeReads func()) error {
	select {
	case <-finished:
		return nil
	case <-ctx.Done():
		select {
		case <-finished:
			return nil
		default:
		}
		closeReads()
		<-finished
		return ctx.Err()
	}
}

// CombinedOutput is Run with both streams captured in one buffer.
func CombinedOutput(cmd *exec.Cmd, options Options) ([]byte, error) {
	if cmd.Stdout != nil || cmd.Stderr != nil {
		return nil, fmt.Errorf("childguard: output already configured")
	}
	var buffer bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buffer, &buffer
	err := Run(cmd, options)
	cmd.Stdout, cmd.Stderr = nil, nil
	return buffer.Bytes(), err
}
