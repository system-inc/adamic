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
	// Started runs synchronously after a successful start, before waiting or arming guard timers.
	// Callers can register the exact child PID with an existing hard-deadline owner.
	Started func(int)

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

func (guard *Error) Error() string {
	if guard.Reason == "stalled" {
		window := "no output"
		if guard.FirstOutput {
			window = "no first output"
		}
		return fmt.Sprintf(
			"stalled: %s for %s after %s, load %s (%s)",
			window,
			guard.Window,
			guard.Elapsed,
			guard.Load,
			guard.Command,
		)
	}
	return fmt.Sprintf("ceiling: %s", guard.Elapsed)
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

func (watched *watched) Write(p []byte) (int, error) {
	if len(p) > 0 {
		watched.activity.Lock()
		watched.activity.last = time.Now()
		watched.activity.seen = true
		watched.activity.Unlock()
		select {
		case watched.activity.progress <- struct{}{}:
		default:
		}
	}
	return watched.dst.Write(p)
}

// Run starts an unstarted command in its own process group and waits for it.
// Output is forwarded to the command's original writers (nil means discard).
// Context commands also cancel the whole process group.
func Run(command *exec.Cmd, options Options) error {
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
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	command.SysProcAttr.Setpgid = true
	command.SysProcAttr.Pgid = 0
	if command.Cancel != nil {
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
	}
	originalOut, originalError := command.Stdout, command.Stderr
	defer func() { command.Stdout, command.Stderr = originalOut, originalError }()
	start := time.Now()
	activity := &activity{last: start, progress: make(chan struct{}, 1)}
	wrap := func(destination io.Writer) io.Writer {
		if destination == nil {
			destination = io.Discard
		}
		return &watched{dst: destination, activity: activity}
	}
	command.Stdout = wrap(originalOut)
	// Preserve exec's serialization when both streams share a writer.
	if originalOut != nil && reflect.TypeOf(originalOut).Comparable() && originalOut == originalError {
		command.Stderr = command.Stdout
	} else {
		command.Stderr = wrap(originalError)
	}
	if startError := command.Start(); startError != nil {
		return startError
	}
	if options.Started != nil {
		options.Started(command.Process.Pid)
	}
	if options.Unbounded {
		return command.Wait()
	}

	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	timer := time.NewTimer(min(options.FirstOutput, options.Ceiling))
	defer timer.Stop()
	for {
		select {
		case err := <-done:
			return err
		case <-activity.progress:
			timer.Reset(0)
		case <-timer.C:
			select {
			case err := <-done:
				return err
			default:
			}
			now := time.Now()
			activity.Lock()
			idle := now.Sub(activity.last)
			firstOutput := !activity.seen
			window := options.Stall
			if firstOutput {
				window = options.FirstOutput
			}
			activity.Unlock()
			elapsed := now.Sub(start)
			reason := ""
			if elapsed >= options.Ceiling {
				reason = "ceiling"
			} else if idle >= window {
				reason = "stalled"
			}
			if reason != "" {
				if killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL); killError != nil {
					return <-done
				}
				<-done
				load := "unavailable"
				if data, readError := os.ReadFile("/proc/loadavg"); readError == nil {
					if fields := strings.Fields(string(data)); len(fields) > 0 {
						load = fields[0]
					}
				}
				return &Error{
					Reason:      reason,
					Window:      window,
					FirstOutput: firstOutput,
					Command:     command.Path,
					Elapsed:     elapsed,
					Load:        load,
				}
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
func Capture(callerContext context.Context, command *exec.Cmd) (stdout, stderr []byte, captureError error) {
	var out, diagnostic bytes.Buffer
	captureToError := CaptureTo(callerContext, command, &out, &diagnostic)
	return out.Bytes(), diagnostic.Bytes(), captureToError
}

// CaptureTo is Capture with caller-provided destinations, allowing bounded output
// retention without a timer on either drain. Nil destinations discard output.
// Destinations are written concurrently and must not share unsynchronized state.
// A destination error is reported after draining the remainder to discard.
func CaptureTo(callerContext context.Context, command *exec.Cmd, stdout, stderr io.Writer) (captureToError error) {
	if command.Stdout != nil || command.Stderr != nil {
		return fmt.Errorf("childguard: output already configured")
	}
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	command.SysProcAttr.Setpgid = true
	command.SysProcAttr.Pgid = 0
	type pipe struct {
		read, write *os.File
		output      io.Writer
		copyError   error
	}
	pipes := make([]*pipe, 2)
	closeAll := func() {
		for _, pipe := range pipes {
			if pipe != nil {
				pipe.read.Close()
				pipe.write.Close()
			}
		}
	}
	for index := range pipes {
		read, write, pipeError := os.Pipe()
		if pipeError != nil {
			closeAll()
			return pipeError
		}
		destination := []io.Writer{stdout, stderr}[index]
		if destination == nil {
			destination = io.Discard
		}
		pipes[index] = &pipe{read: read, write: write, output: destination}
	}
	// *os.File writers go to the child as they are, so exec starts no copying of its own and Wait waits only for the
	// process; this process's write ends close once the child holds its own.
	command.Stdout, command.Stderr = pipes[0].write, pipes[1].write
	defer func() { command.Stdout, command.Stderr = nil, nil }()
	if startError := command.Start(); startError != nil {
		closeAll()
		return startError
	}
	var reads sync.WaitGroup
	for _, pipe := range pipes {
		pipe.write.Close()
		reads.Add(1)
		go func() {
			defer reads.Done()
			_, pipe.copyError = io.Copy(pipe.output, pipe.read)
			if pipe.copyError != nil {
				// A failing destination must not leave the child blocked on a full pipe.
				_, _ = io.Copy(io.Discard, pipe.read)
			}
		}()
	}
	finished := make(chan struct{})
	go func() { reads.Wait(); close(finished) }()
	stop := context.AfterFunc(callerContext, func() { _ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL) })
	captureToError = command.Wait()
	stop()
	// The leader is reaped, but while any member of its group lives the group's id can't be reused, so this reaches
	// exactly the processes it left behind, and none when it left none.
	_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	closeReads := func() {
		for _, pipe := range pipes {
			pipe.read.Close()
		}
	}
	if held := captureEnd(callerContext, finished, closeReads); captureToError == nil && held != nil {
		captureToError = fmt.Errorf(
			"childguard: %s's output is still held by a process outside its group: %w",
			command.Path,
			held,
		)
	}
	closeReads()
	for _, pipe := range pipes {
		if captureToError == nil {
			captureToError = pipe.copyError
		}
	}
	return captureToError
}

// captureEnd waits for Capture's reads to finish, or for ctx: then it closes the reads, waits for them to stop and
// returns ctx's error. Reads that finished as ctx ended are whole, whichever a select would pick, so they return nil.
func captureEnd(callerContext context.Context, finished <-chan struct{}, closeReads func()) error {
	select {
	case <-finished:
		return nil
	case <-callerContext.Done():
		select {
		case <-finished:
			return nil
		default:
		}
		closeReads()
		<-finished
		return callerContext.Err()
	}
}

// CombinedOutput is Run with both streams captured in one buffer.
func CombinedOutput(command *exec.Cmd, options Options) ([]byte, error) {
	if command.Stdout != nil || command.Stderr != nil {
		return nil, fmt.Errorf("childguard: output already configured")
	}
	var buffer bytes.Buffer
	command.Stdout, command.Stderr = &buffer, &buffer
	runError := Run(command, options)
	command.Stdout, command.Stderr = nil, nil
	return buffer.Bytes(), runError
}
