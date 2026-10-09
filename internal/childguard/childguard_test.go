package childguard

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestChild(t *testing.T) {
	t.Parallel()
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
	}
	os.Exit(0)
}
func child(mode string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestChild$")
	cmd.Env = append(os.Environ(), "CHILDGUARD_TEST="+mode, "GORACE=atexit_sleep_ms=0")
	return cmd
}

// Not parallel: child-process scheduling on this machine determines the 5s first-output, 2s stall and 12s ceiling windows.
func TestProgress(t *testing.T) {
	var out bytes.Buffer
	cmd := child("progress")
	cmd.Stdout = &out
	if err := Run(cmd, Options{FirstOutput: 5 * time.Second, Stall: 2 * time.Second, Ceiling: 12 * time.Second}); err != nil {
		t.Fatalf("progressing child killed: %v", err)
	}
	if strings.Count(out.String(), "progress\n") != 5 {
		t.Fatalf("output lost: %q", out.String())
	}
}

// Not parallel: child-process scheduling on this machine determines the 200ms stall and 1200ms elapsed-time bound.
func TestStalled(t *testing.T) {
	var out bytes.Buffer
	cmd := child("stall")
	cmd.Stderr = &out
	start := time.Now()
	err := Run(cmd, Options{FirstOutput: 5 * time.Second, Stall: 200 * time.Millisecond, Ceiling: 5 * time.Second})
	var guard *Error
	if !errors.As(err, &guard) || guard.Reason != "stalled" || guard.FirstOutput || guard.Window != 200*time.Millisecond || !strings.Contains(err.Error(), "stalled: no output for 200ms after") {
		t.Fatalf("want stalled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 200*time.Millisecond || elapsed > 1200*time.Millisecond {
		t.Fatalf("stall took %s", elapsed)
	}
	if out.String() != "once\n" {
		t.Fatalf("stderr lost: %q", out.String())
	}
}

// Not parallel: child-process scheduling on this machine determines the 400ms ceiling and 1400ms elapsed-time bound.
func TestCeiling(t *testing.T) {
	start := time.Now()
	err := Run(child("ceiling"), Options{Stall: 200 * time.Millisecond, Ceiling: 400 * time.Millisecond})
	var guard *Error
	if !errors.As(err, &guard) || guard.Reason != "ceiling" || !strings.HasPrefix(err.Error(), "ceiling: ") {
		t.Fatalf("want ceiling, got %v", err)
	}
	if time.Since(start) > 1400*time.Millisecond {
		t.Fatal("ceiling was late")
	}
}
func TestExitIsNotGuardError(t *testing.T) {
	t.Parallel()
	err := Run(child("exit"), Options{})
	var guard *Error
	var exit *exec.ExitError
	if errors.As(err, &guard) || !errors.As(err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("wrong exit error: %v", err)
	}
}

// A surviving descendant would keep the output pipes open and Run would not return.
// Not parallel: child-process scheduling on this machine determines the 200ms stall and 2s process-group kill deadline.
func TestKillsProcessGroup(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 60 & echo started; wait")
	done := make(chan error, 1)
	go func() { done <- Run(cmd, Options{Stall: 200 * time.Millisecond, Ceiling: time.Second}) }()
	select {
	case err := <-done:
		var guard *Error
		if !errors.As(err, &guard) || guard.Reason != "stalled" {
			t.Fatalf("want stalled group, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("descendant kept output pipes open after group kill")
	}
}

// Not parallel: child-process scheduling on this machine determines the 300ms first-output and 1300ms elapsed-time bound.
func TestNoFirstOutput(t *testing.T) {
	start := time.Now()
	err := Run(child("silent"), Options{FirstOutput: 300 * time.Millisecond, Stall: 100 * time.Millisecond, Ceiling: 5 * time.Second})
	var guard *Error
	if !errors.As(err, &guard) || guard.Reason != "stalled" || !guard.FirstOutput || guard.Window != 300*time.Millisecond || !strings.Contains(err.Error(), "no first output for 300ms") {
		t.Fatalf("want first-output stall, got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 300*time.Millisecond || elapsed > 1300*time.Millisecond {
		t.Fatalf("first-output stall took %s", elapsed)
	}
}

// Leave Stall unset to exercise the default window after the first output.
func TestChildFinishesWithinDefaultStallWindow(t *testing.T) {
	t.Parallel()
	out, err := CombinedOutput(child("finish"), Options{FirstOutput: 5 * time.Second, Ceiling: 5 * time.Second})
	if err != nil {
		t.Fatalf("child finishing within its window reported an error: %v (output %q)", err, out)
	}
	if string(out) != "ready\ndone\n" {
		t.Fatalf("output lost: got %q, want ready and done", out)
	}
}

func TestRunRestoresStdout(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	cmd := child("finish")
	cmd.Stdout = &out
	if err := Run(cmd, Options{}); err != nil {
		t.Fatal(err)
	}
	if cmd.Stdout != &out {
		t.Fatalf("stdout was not restored: got %T, want original buffer", cmd.Stdout)
	}
}

func TestStalledReportsSystemLoad(t *testing.T) {
	t.Parallel()
	readLoad := func() string {
		t.Helper()
		data, err := os.ReadFile("/proc/loadavg")
		if os.IsNotExist(err) {
			t.Skip("system has no /proc/loadavg load source")
		}
		if err != nil {
			t.Fatal(err)
		}
		fields := strings.Fields(string(data))
		if len(fields) == 0 {
			t.Fatal("empty /proc/loadavg")
		}
		return fields[0]
	}
	// The kernel updates load periodically. Compare only runs bracketed by
	// matching source values so an update during Run cannot make this flaky.
	for attempt := 0; attempt < 5; attempt++ {
		before := readLoad()
		err := Run(child("stall"), Options{FirstOutput: 5 * time.Second, Stall: 50 * time.Millisecond, Ceiling: 5 * time.Second})
		var guard *Error
		if !errors.As(err, &guard) || guard.Reason != "stalled" {
			t.Fatalf("want stalled error, got %v", err)
		}
		after := readLoad()
		if before != after {
			continue
		}
		if guard.Load != before {
			t.Fatalf("reported load %q, system load source gives %q", guard.Load, before)
		}
		return
	}
	t.Fatal("system load changed during every guard run")
}
