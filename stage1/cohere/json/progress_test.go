package json

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type progressPolicy struct{ startup, stall, overall time.Duration }

// Startup accommodates a heavily contended build. Once the child writes, only
// a lack of output triggers the stall guard. The overall cap is a backstop.
var jsonProgressPolicy = progressPolicy{30 * time.Minute, 5 * time.Minute, 6 * time.Hour}

type progressCommand struct {
	*exec.Cmd
	policy progressPolicy
	cancel context.CancelFunc
}

func newProgressCommand(policy progressPolicy, name string, arguments ...string) *progressCommand {
	ctx, cancel := context.WithCancel(context.Background())
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 5 * time.Second
	return &progressCommand{Cmd: command, policy: policy, cancel: cancel}
}

type progressWriter struct{ write func([]byte) (int, error) }

func (w progressWriter) Write(data []byte) (int, error) { return w.write(data) }

func (c *progressCommand) Run() error {
	defer c.cancel()
	started := time.Now()
	last := started
	seen := false
	reason := ""
	var mutex sync.Mutex
	progress := make(chan struct{}, 1)
	wrap := func(destination io.Writer) io.Writer {
		if destination == nil {
			destination = io.Discard
		}
		return progressWriter{write: func(data []byte) (int, error) {
			mutex.Lock()
			defer mutex.Unlock()
			count, err := destination.Write(data)
			if count > 0 {
				seen, last = true, time.Now()
				select {
				case progress <- struct{}{}:
				default:
				}
			}
			return count, err
		}}
	}
	c.Stdout, c.Stderr = wrap(c.Stdout), wrap(c.Stderr)
	if err := c.Cmd.Start(); err != nil {
		return err
	}
	done, watched := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(watched)
		timer := time.NewTimer(c.policy.startup)
		overall := time.NewTimer(c.policy.overall)
		defer timer.Stop()
		defer overall.Stop()
		for {
			select {
			case <-done:
				return
			case <-progress:
				timer.Reset(c.policy.stall)
			case <-timer.C:
				mutex.Lock()
				remaining := c.policy.startup - time.Since(started)
				message := "never started"
				if seen {
					remaining = c.policy.stall - time.Since(last)
					message = "stalled without output"
				}
				if remaining <= 0 {
					reason = fmt.Sprintf("%s: %s", c.Path, message)
				}
				mutex.Unlock()
				if remaining > 0 {
					timer.Reset(remaining)
					continue
				}
				c.cancel()
				return
			case <-overall.C:
				mutex.Lock()
				reason = fmt.Sprintf("%s: overall backstop exceeded", c.Path)
				mutex.Unlock()
				c.cancel()
				return
			}
		}
	}()
	err := c.Cmd.Wait()
	close(done)
	<-watched
	if reason != "" && err != nil {
		return fmt.Errorf("%s", reason)
	}
	return err
}

func (c *progressCommand) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil || c.Stderr != nil {
		return nil, fmt.Errorf("exec: output already set")
	}
	var output bytes.Buffer
	c.Stdout, c.Stderr = &output, &output
	err := c.Run()
	return output.Bytes(), err
}

func TestProgressGuard(t *testing.T) {
	t.Parallel()
	policy := progressPolicy{500 * time.Millisecond, 200 * time.Millisecond, 5 * time.Second}
	for _, fixture := range []struct{ name, source, message string }{
		{"never prints", "exec sleep 20", "never started"},
		{"never returns", "printf ready; exec sleep 20", "stalled without output"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			command := newProgressCommand(policy, "sh", "-c", fixture.source)
			_, err := command.CombinedOutput()
			if err == nil || !strings.Contains(err.Error(), fixture.message) || !strings.Contains(err.Error(), "sh") {
				t.Fatalf("guard failed by name: %v", err)
			}
			t.Logf("caught %s: %v", fixture.name, err)
		})
	}
	t.Run("progress outlasts startup", func(t *testing.T) {
		command := newProgressCommand(policy, "sh", "-c", "for i in 1 2 3 4 5 6 7 8; do printf x; sleep 0.1; done")
		output, err := command.CombinedOutput()
		if err != nil || string(output) != "xxxxxxxx" {
			t.Fatalf("progress was killed: %q %v", output, err)
		}
	})
	t.Run("stderr is progress", func(t *testing.T) {
		command := newProgressCommand(policy, "sh", "-c", "for i in 1 2 3 4 5 6 7 8; do printf x >&2; sleep 0.1; done")
		output, err := command.CombinedOutput()
		if err != nil || string(output) != "xxxxxxxx" {
			t.Fatalf("stderr progress was killed: %q %v", output, err)
		}
	})
	t.Run("overall backstop", func(t *testing.T) {
		short := policy
		short.overall = 300 * time.Millisecond
		command := newProgressCommand(short, "sh", "-c", "while :; do printf x; sleep 0.05; done")
		_, err := command.CombinedOutput()
		if err == nil || !strings.Contains(err.Error(), "overall backstop") {
			t.Fatalf("backstop survived: %v", err)
		}
	})
}
