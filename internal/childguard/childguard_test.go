package childguard

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
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

func TestChild(test *testing.T) {
	test.Parallel()
	mode := os.Getenv("CHILDGUARD_TEST")
	if mode == "" {
		return
	}
	switch mode {
	case "silent":
		time.Sleep(time.Hour)
	case "progress":
		time.Sleep(3 * time.Second)
		for i := 0; i < 5; i++ {
			fmt.Fprintln(os.Stdout, "progress")
			time.Sleep(time.Second)
		}
	case "stall":
		fmt.Fprintln(os.Stderr, "once")
		time.Sleep(time.Hour)
	case "ceiling":
		for {
			fmt.Fprintln(os.Stdout, "progress")
			time.Sleep(20 * time.Millisecond)
		}
	case "finish":
		fmt.Fprintln(os.Stdout, "ready")
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(os.Stdout, "done")
	case "exit":
		os.Exit(7)
	case "descendant", "escaped":
		// A silent descendant that inherits this process's stdout and stderr and outlives it; "escaped" leaves the
		// group.
		grandchild := child("silent")
		grandchild.Stdout, grandchild.Stderr = os.Stdout, os.Stderr
		if mode == "escaped" {
			grandchild.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		}
		if err := grandchild.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "parent %d\n", grandchild.Process.Pid)
	case "large":
		line := strings.Repeat("x", 1023) + "\n"
		done := make(chan struct{})
		go func() {
			for range 1024 {
				fmt.Fprint(os.Stderr, line)
			}
			close(done)
		}()
		for range 1024 {
			fmt.Fprint(os.Stdout, line)
		}
		<-done
	}
	os.Exit(0)
}
func child(mode string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestChild$")
	cmd.Env = append(os.Environ(), "CHILDGUARD_TEST="+mode, "GORACE=atexit_sleep_ms=0")
	return cmd
}

// Not parallel: child-process scheduling on this machine determines the 5s first-output, 2s stall and 12s ceiling windows.
func TestProgress(test *testing.T) {
	var out bytes.Buffer
	cmd := child("progress")
	cmd.Stdout = &out
	if err := Run(
		cmd,
		Options{FirstOutput: 5 * time.Second, Stall: 2 * time.Second, Ceiling: 12 * time.Second},
	); err != nil {
		test.Fatalf("progressing child killed: %v", err)
	}
	if strings.Count(out.String(), "progress\n") != 5 {
		test.Fatalf("output lost: %q", out.String())
	}
}

// Not parallel: child-process scheduling on this machine determines the 200ms stall and 1200ms elapsed-time bound.
func TestStalled(test *testing.T) {
	var out bytes.Buffer
	cmd := child("stall")
	cmd.Stderr = &out
	start := time.Now()
	err := Run(cmd, Options{FirstOutput: 5 * time.Second, Stall: 200 * time.Millisecond, Ceiling: 5 * time.Second})
	var guard *Error
	if !errors.As(err, &guard) ||
		guard.Reason != "stalled" ||
		guard.FirstOutput ||
		guard.Window != 200*time.Millisecond ||
		!strings.Contains(err.Error(), "stalled: no output for 200ms after") {
		test.Fatalf("want stalled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 200*time.Millisecond || elapsed > 1200*time.Millisecond {
		test.Fatalf("stall took %s", elapsed)
	}
	if out.String() != "once\n" {
		test.Fatalf("stderr lost: %q", out.String())
	}
}

// Not parallel: child-process scheduling on this machine determines the 400ms ceiling and 1400ms elapsed-time bound.
func TestCeiling(test *testing.T) {
	start := time.Now()
	err := Run(child("ceiling"), Options{Stall: 200 * time.Millisecond, Ceiling: 400 * time.Millisecond})
	var guard *Error
	if !errors.As(err, &guard) || guard.Reason != "ceiling" || !strings.HasPrefix(err.Error(), "ceiling: ") {
		test.Fatalf("want ceiling, got %v", err)
	}
	if time.Since(start) > 1400*time.Millisecond {
		test.Fatal("ceiling was late")
	}
}
func TestExitIsNotGuardError(test *testing.T) {
	test.Parallel()
	err := Run(child("exit"), Options{})
	var guard *Error
	var exit *exec.ExitError
	if errors.As(err, &guard) || !errors.As(err, &exit) || exit.ExitCode() != 7 {
		test.Fatalf("wrong exit error: %v", err)
	}
}

// A surviving descendant would keep the output pipes open and Run would not return.
// Not parallel: child-process scheduling on this machine determines the 200ms stall and 2s process-group kill deadline.
func TestKillsProcessGroup(test *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 60 & echo started; wait")
	done := make(chan error, 1)
	go func() { done <- Run(cmd, Options{Stall: 200 * time.Millisecond, Ceiling: time.Second}) }()
	select {
	case err := <-done:
		var guard *Error
		if !errors.As(err, &guard) || guard.Reason != "stalled" {
			test.Fatalf("want stalled group, got %v", err)
		}
	case <-time.After(2 * time.Second):
		test.Fatal("descendant kept output pipes open after group kill")
	}
}

// Not parallel: child-process scheduling on this machine determines the 300ms first-output and 1300ms elapsed-time bound.
func TestNoFirstOutput(test *testing.T) {
	start := time.Now()
	err := Run(
		child("silent"),
		Options{FirstOutput: 300 * time.Millisecond, Stall: 100 * time.Millisecond, Ceiling: 5 * time.Second},
	)
	var guard *Error
	if !errors.As(err, &guard) ||
		guard.Reason != "stalled" ||
		!guard.FirstOutput ||
		guard.Window != 300*time.Millisecond ||
		!strings.Contains(err.Error(), "no first output for 300ms") {
		test.Fatalf("want first-output stall, got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 300*time.Millisecond || elapsed > 1300*time.Millisecond {
		test.Fatalf("first-output stall took %s", elapsed)
	}
}

// Leave Stall unset to exercise the default window after the first output.
func TestChildFinishesWithinDefaultStallWindow(test *testing.T) {
	test.Parallel()
	out, err := CombinedOutput(child("finish"), Options{FirstOutput: 5 * time.Second, Ceiling: 5 * time.Second})
	if err != nil {
		test.Fatalf("child finishing within its window reported an error: %v (output %q)", err, out)
	}
	if string(out) != "ready\ndone\n" {
		test.Fatalf("output lost: got %q, want ready and done", out)
	}
}

func TestRunRestoresStdout(test *testing.T) {
	test.Parallel()
	var out bytes.Buffer
	cmd := child("finish")
	cmd.Stdout = &out
	if err := Run(cmd, Options{}); err != nil {
		test.Fatal(err)
	}
	if cmd.Stdout != &out {
		test.Fatalf("stdout was not restored: got %T, want original buffer", cmd.Stdout)
	}
}

func TestStalledReportsSystemLoad(test *testing.T) {
	test.Parallel()
	readLoad := func() string {
		test.Helper()
		data, err := os.ReadFile("/proc/loadavg")
		if os.IsNotExist(err) {
			test.Skip("system has no /proc/loadavg load source")
		}
		if err != nil {
			test.Fatal(err)
		}
		fields := strings.Fields(string(data))
		if len(fields) == 0 {
			test.Fatal("empty /proc/loadavg")
		}
		return fields[0]
	}
	// The kernel updates load periodically. Compare only runs bracketed by
	// matching source values so an update during Run cannot make this flaky.
	for attempt := 0; attempt < 5; attempt++ {
		before := readLoad()
		err := Run(
			child("stall"),
			Options{FirstOutput: 5 * time.Second, Stall: 50 * time.Millisecond, Ceiling: 5 * time.Second},
		)
		var guard *Error
		if !errors.As(err, &guard) || guard.Reason != "stalled" {
			test.Fatalf("want stalled error, got %v", err)
		}
		after := readLoad()
		if before != after {
			continue
		}
		if guard.Load != before {
			test.Fatalf("reported load %q, system load source gives %q", guard.Load, before)
		}
		return
	}
	test.Fatal("system load changed during every guard run")
}

// A descendant in the command's group that holds its output past its exit is killed with the group, so Capture
// returns what the command wrote, with no error, as soon as the command exits, not when the descendant would have.
func TestCaptureReadsPastADescendant(test *testing.T) {
	test.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	start := time.Now()
	stdout, stderr, err := Capture(ctx, child("descendant"))
	if err != nil || !strings.HasPrefix(string(stdout), "parent ") || len(stderr) != 0 {
		test.Fatalf("err %v, stdout %q, stderr %q", err, stdout, stderr)
	}
	if elapsed := time.Since(start); elapsed > 30*time.Second {
		test.Fatalf("Capture waited %s on the descendant", elapsed)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(string(stdout), "parent ")))
	if err != nil {
		test.Fatal(err)
	}
	// Killed with the group: the signal reaches it at once, so it is gone or a zombie waiting for its new parent.
	for deadline := time.Now().Add(10 * time.Second); syscall.Kill(pid, 0) == nil && !zombie(pid); {
		if time.Now().After(deadline) {
			test.Fatalf("descendant %d still runs after Capture returned", pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// zombie reports whether pid has exited and waits to be reaped, from /proc (false off Linux).
func zombie(pid int) bool {
	stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	_, rest, found := strings.Cut(string(stat), ") ")
	return found && strings.HasPrefix(rest, "Z")
}

// A descendant that left the group (setsid) can't be killed with it; ctx then ends the reads, with what was read and
// an error naming why.
func TestCaptureEndsWithAWriterOutsideTheGroup(test *testing.T) {
	test.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stdout, _, err := Capture(ctx, child("escaped"))
	if pid, parsed := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(string(stdout), "parent "))); parsed == nil {
		defer syscall.Kill(pid, syscall.SIGKILL)
	}
	if !errors.Is(err, context.DeadlineExceeded) ||
		!strings.Contains(err.Error(), "outside its group") ||
		!strings.HasPrefix(string(stdout), "parent ") {
		test.Fatalf("err %v, stdout %q", err, stdout)
	}
}

// Reads that finished as ctx ended are whole: with both ready, a bare select would pick ctx's branch half the time and
// name an outside writer for a run whose reads had ended, so a thousand tries catch it. Reads still open when ctx ends
// are closed and named.
func TestCaptureEndReadsFinishedAsCtxEnded(test *testing.T) {
	test.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	finished := make(chan struct{})
	close(finished)
	for try := range 1000 {
		if err := captureEnd(ctx, finished, func() { test.Fatal("closed reads that had finished") }); err != nil {
			test.Fatalf("try %d: finished reads reported as held: %v", try, err)
		}
	}
	open := make(chan struct{})
	closed := false
	if err := captureEnd(
		ctx,
		open,
		func() { closed = true; close(open) },
	); !errors.Is(err, context.Canceled) || !closed {
		test.Fatalf("open reads at ctx's end: err %v, closed %t", err, closed)
	}
}

func TestCaptureReportsTheExit(test *testing.T) {
	test.Parallel()
	_, _, err := Capture(context.Background(), child("exit"))
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 7 {
		test.Fatalf("wrong exit error: %v", err)
	}
}

// Both streams are read at once, so a command filling both pipes never blocks on the one not being read.
func TestCaptureReadsBothStreamsWhole(test *testing.T) {
	test.Parallel()
	stdout, stderr, err := Capture(context.Background(), child("large"))
	if err != nil || len(stdout) != 1<<20 || len(stderr) != 1<<20 {
		test.Fatalf("err %v, %d stdout bytes, %d stderr bytes, want %d each", err, len(stdout), len(stderr), 1<<20)
	}
}

func TestCaptureRefusesConfiguredOutput(test *testing.T) {
	test.Parallel()
	cmd := child("finish")
	cmd.Stdout = &bytes.Buffer{}
	if _, _, err := Capture(
		context.Background(),
		cmd,
	); err == nil || !strings.Contains(err.Error(), "already configured") {
		test.Fatalf("configured output accepted: %v", err)
	}
}

// Not parallel: scheduling determines the cancellation and process-group cleanup bound.
func TestContextCancellationKillsProcessGroup(test *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	cmd := toolchain.CommandContextIn(ctx, "", "sh", "-c", "sleep 60 & echo started; wait")
	var output bytes.Buffer
	cmd.Stdout = &output
	done := make(chan error, 1)
	go func() { done <- Run(cmd, Options{Unbounded: true}) }()
	select {
	case err := <-done:
		if err == nil || ctx.Err() != context.DeadlineExceeded {
			test.Fatalf("want cancelled child, got %v, context %v", err, ctx.Err())
		}
		if output.String() != "started\n" {
			test.Fatalf("child did not start: %q", output.String())
		}
	case <-time.After(2 * time.Second):
		test.Fatal("cancelled process group kept its output pipe open")
	}
}

func TestUnboundedIgnoresGuardTimers(test *testing.T) {
	test.Parallel()
	output, err := CombinedOutput(
		child("finish"),
		Options{FirstOutput: time.Millisecond, Stall: time.Millisecond, Ceiling: time.Millisecond, Unbounded: true},
	)
	if err != nil || string(output) != "ready\ndone\n" {
		test.Fatalf("unbounded child: %v, %q", err, output)
	}
}

func TestStartedRegistersOnlySuccessfulStarts(test *testing.T) {
	test.Parallel()
	command := child("finish")
	registered := 0
	if err := Run(command, Options{Unbounded: true, Started: func(pid int) {
		if command.Process == nil || command.Process.Pid != pid {
			test.Fatalf("registered %d before the child was available", pid)
		}
		registered = pid
	}}); err != nil {
		test.Fatal(err)
	}
	if registered <= 0 || registered != command.Process.Pid {
		test.Fatalf("registered pid %d, want %d", registered, command.Process.Pid)
	}
	missing := exec.Command(filepath.Join(test.TempDir(), "missing-child"))
	if err := Run(
		missing,
		Options{Unbounded: true, Started: func(int) { test.Fatal("registered a child that did not start") }},
	); err == nil {
		test.Fatal("missing child started")
	}
}
