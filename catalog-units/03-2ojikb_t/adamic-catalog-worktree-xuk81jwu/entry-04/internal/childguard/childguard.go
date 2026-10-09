// Package childguard runs Unix child process groups with output-based stall protection.
package childguard

import (
	"bytes"
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
type Options struct{ FirstOutput, Stall, Ceiling time.Duration }

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
// Callers should use exec.Command, rather than an independent context deadline.
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
