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
		for index := 0; index < 5; index++ {
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
		if startError := grandchild.Start(); startError != nil {
			fmt.Fprintln(os.Stderr, startError)
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
	command := exec.Command(os.Args[0], "-test.run=^TestChild$")
	command.Env = append(os.Environ(), "CHILDGUARD_TEST="+mode, "GORACE=atexit_sleep_ms=0")
	return command
}

// Not parallel: child-process scheduling on this machine determines the 5s first-output, 2s stall and 12s ceiling windows.
func TestProgress(test *testing.T) {
	var out bytes.Buffer
	command := child("progress")
	command.Stdout = &out
	if runError := Run(
		command,
		Options{FirstOutput: 5 * time.Second, Stall: 2 * time.Second, Ceiling: 12 * time.Second},
	); runError != nil {
		test.Fatalf("progressing child killed: %v", runError)
	}
	if strings.Count(out.String(), "progress\n") != 5 {
		test.Fatalf("output lost: %q", out.String())
	}
}

// Not parallel: child-process scheduling on this machine determines the 200ms stall and 1200ms elapsed-time bound.
func TestStalled(test *testing.T) {
	var out bytes.Buffer
	command := child("stall")
	command.Stderr = &out
	start := time.Now()
	runError := Run(
		command,
		Options{FirstOutput: 5 * time.Second, Stall: 200 * time.Millisecond, Ceiling: 5 * time.Second},
	)
	var guard *Error
	if !errors.As(runError, &guard) ||
		guard.Reason != "stalled" ||
		guard.FirstOutput ||
		guard.Window != 200*time.Millisecond ||
		!strings.Contains(runError.Error(), "stalled: no output for 200ms after") {
		test.Fatalf("want stalled, got %v", runError)
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
	runError := Run(child("ceiling"), Options{Stall: 200 * time.Millisecond, Ceiling: 400 * time.Millisecond})
	var guard *Error
	if !errors.As(runError, &guard) || guard.Reason != "ceiling" || !strings.HasPrefix(runError.Error(), "ceiling: ") {
		test.Fatalf("want ceiling, got %v", runError)
	}
	if time.Since(start) > 1400*time.Millisecond {
		test.Fatal("ceiling was late")
	}
}
func TestExitIsNotGuardError(test *testing.T) {
	test.Parallel()
	runError := Run(child("exit"), Options{})
	var guard *Error
	var exit *exec.ExitError
	if errors.As(runError, &guard) || !errors.As(runError, &exit) || exit.ExitCode() != 7 {
		test.Fatalf("wrong exit error: %v", runError)
	}
}

// A surviving descendant would keep the output pipes open and Run would not return.
// Not parallel: child-process scheduling on this machine determines the 200ms stall and 2s process-group kill deadline.
func TestKillsProcessGroup(test *testing.T) {
	command := exec.Command("sh", "-c", "sleep 60 & echo started; wait")
	done := make(chan error, 1)
	go func() { done <- Run(command, Options{Stall: 200 * time.Millisecond, Ceiling: time.Second}) }()
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
	runError := Run(
		child("silent"),
		Options{FirstOutput: 300 * time.Millisecond, Stall: 100 * time.Millisecond, Ceiling: 5 * time.Second},
	)
	var guard *Error
	if !errors.As(runError, &guard) ||
		guard.Reason != "stalled" ||
		!guard.FirstOutput ||
		guard.Window != 300*time.Millisecond ||
		!strings.Contains(runError.Error(), "no first output for 300ms") {
		test.Fatalf("want first-output stall, got %v", runError)
	}
	if elapsed := time.Since(start); elapsed < 300*time.Millisecond || elapsed > 1300*time.Millisecond {
		test.Fatalf("first-output stall took %s", elapsed)
	}
}

// Leave Stall unset to exercise the default window after the first output.
func TestChildFinishesWithinDefaultStallWindow(test *testing.T) {
	test.Parallel()
	out, combinedOutputError := CombinedOutput(
		child("finish"),
		Options{FirstOutput: 5 * time.Second, Ceiling: 5 * time.Second},
	)
	if combinedOutputError != nil {
		test.Fatalf("child finishing within its window reported an error: %v (output %q)", combinedOutputError, out)
	}
	if string(out) != "ready\ndone\n" {
		test.Fatalf("output lost: got %q, want ready and done", out)
	}
}

func TestRunRestoresStdout(test *testing.T) {
	test.Parallel()
	var out bytes.Buffer
	command := child("finish")
	command.Stdout = &out
	if runError := Run(command, Options{}); runError != nil {
		test.Fatal(runError)
	}
	if command.Stdout != &out {
		test.Fatalf("stdout was not restored: got %T, want original buffer", command.Stdout)
	}
}

func TestStalledReportsSystemLoad(test *testing.T) {
	test.Parallel()
	readLoad := func() string {
		test.Helper()
		data, readError := os.ReadFile("/proc/loadavg")
		if os.IsNotExist(readError) {
			test.Skip("system has no /proc/loadavg load source")
		}
		if readError != nil {
			test.Fatal(readError)
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
		runError := Run(
			child("stall"),
			Options{FirstOutput: 5 * time.Second, Stall: 50 * time.Millisecond, Ceiling: 5 * time.Second},
		)
		var guard *Error
		if !errors.As(runError, &guard) || guard.Reason != "stalled" {
			test.Fatalf("want stalled error, got %v", runError)
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
	timeoutContext, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	start := time.Now()
	stdout, stderr, captureError := Capture(timeoutContext, child("descendant"))
	if captureError != nil || !strings.HasPrefix(string(stdout), "parent ") || len(stderr) != 0 {
		test.Fatalf("err %v, stdout %q, stderr %q", captureError, stdout, stderr)
	}
	if elapsed := time.Since(start); elapsed > 30*time.Second {
		test.Fatalf("Capture waited %s on the descendant", elapsed)
	}
	pid, parseError := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(string(stdout), "parent ")))
	if parseError != nil {
		test.Fatal(parseError)
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
	stat, readError := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if readError != nil {
		return false
	}
	_, rest, found := strings.Cut(string(stat), ") ")
	return found && strings.HasPrefix(rest, "Z")
}

// A descendant that left the group (setsid) can't be killed with it; ctx then ends the reads, with what was read and
// an error naming why.
func TestCaptureEndsWithAWriterOutsideTheGroup(test *testing.T) {
	test.Parallel()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stdout, _, captureError := Capture(timeoutContext, child("escaped"))
	if pid, parsed := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(string(stdout), "parent "))); parsed == nil {
		defer syscall.Kill(pid, syscall.SIGKILL)
	}
	if !errors.Is(captureError, context.DeadlineExceeded) ||
		!strings.Contains(captureError.Error(), "outside its group") ||
		!strings.HasPrefix(string(stdout), "parent ") {
		test.Fatalf("err %v, stdout %q", captureError, stdout)
	}
}

// Reads that finished as ctx ended are whole: with both ready, a bare select would pick ctx's branch half the time and
// name an outside writer for a run whose reads had ended, so a thousand tries catch it. Reads still open when ctx ends
// are closed and named.
func TestCaptureEndReadsFinishedAsCtxEnded(test *testing.T) {
	test.Parallel()
	cancelContext, cancel := context.WithCancel(context.Background())
	cancel()
	finished := make(chan struct{})
	close(finished)
	for try := range 1000 {
		if captureEndError := captureEnd(
			cancelContext,
			finished,
			func() { test.Fatal("closed reads that had finished") },
		); captureEndError != nil {
			test.Fatalf("try %d: finished reads reported as held: %v", try, captureEndError)
		}
	}
	open := make(chan struct{})
	closed := false
	if captureEndError := captureEnd(
		cancelContext,
		open,
		func() { closed = true; close(open) },
	); !errors.Is(captureEndError, context.Canceled) || !closed {
		test.Fatalf("open reads at ctx's end: err %v, closed %t", captureEndError, closed)
	}
}

func TestCaptureReportsTheExit(test *testing.T) {
	test.Parallel()
	_, _, captureError := Capture(context.Background(), child("exit"))
	var exit *exec.ExitError
	if !errors.As(captureError, &exit) || exit.ExitCode() != 7 {
		test.Fatalf("wrong exit error: %v", captureError)
	}
}

// Both streams are read at once, so a command filling both pipes never blocks on the one not being read.
func TestCaptureReadsBothStreamsWhole(test *testing.T) {
	test.Parallel()
	stdout, stderr, captureError := Capture(context.Background(), child("large"))
	if captureError != nil || len(stdout) != 1<<20 || len(stderr) != 1<<20 {
		test.Fatalf(
			"err %v, %d stdout bytes, %d stderr bytes, want %d each",
			captureError,
			len(stdout),
			len(stderr),
			1<<20,
		)
	}
}

func TestCaptureRefusesConfiguredOutput(test *testing.T) {
	test.Parallel()
	command := child("finish")
	command.Stdout = &bytes.Buffer{}
	if _, _, captureError := Capture(
		context.Background(),
		command,
	); captureError == nil || !strings.Contains(captureError.Error(), "already configured") {
		test.Fatalf("configured output accepted: %v", captureError)
	}
}

// Not parallel: scheduling determines the cancellation and process-group cleanup bound.
func TestContextCancellationKillsProcessGroup(test *testing.T) {
	timeoutContext, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	command := toolchain.CommandContextIn(timeoutContext, "", "sh", "-c", "sleep 60 & echo started; wait")
	var output bytes.Buffer
	command.Stdout = &output
	done := make(chan error, 1)
	go func() { done <- Run(command, Options{Unbounded: true}) }()
	select {
	case err := <-done:
		if err == nil || timeoutContext.Err() != context.DeadlineExceeded {
			test.Fatalf("want cancelled child, got %v, context %v", err, timeoutContext.Err())
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
	output, combinedOutputError := CombinedOutput(
		child("finish"),
		Options{FirstOutput: time.Millisecond, Stall: time.Millisecond, Ceiling: time.Millisecond, Unbounded: true},
	)
	if combinedOutputError != nil || string(output) != "ready\ndone\n" {
		test.Fatalf("unbounded child: %v, %q", combinedOutputError, output)
	}
}

func TestStartedRegistersOnlySuccessfulStarts(test *testing.T) {
	test.Parallel()
	command := child("finish")
	registered := 0
	if runError := Run(command, Options{Unbounded: true, Started: func(pid int) {
		if command.Process == nil || command.Process.Pid != pid {
			test.Fatalf("registered %d before the child was available", pid)
		}
		registered = pid
	}}); runError != nil {
		test.Fatal(runError)
	}
	if registered <= 0 || registered != command.Process.Pid {
		test.Fatalf("registered pid %d, want %d", registered, command.Process.Pid)
	}
	missing := exec.Command(filepath.Join(test.TempDir(), "missing-child"))
	if runError := Run(
		missing,
		Options{Unbounded: true, Started: func(int) { test.Fatal("registered a child that did not start") }},
	); runError == nil {
		test.Fatal("missing child started")
	}
}
