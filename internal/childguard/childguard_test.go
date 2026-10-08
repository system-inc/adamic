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
	err := Run(child("exit"), Options{})
	var guard *Error
	var exit *exec.ExitError
	if errors.As(err, &guard) || !errors.As(err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("wrong exit error: %v", err)
	}
}

// A surviving descendant would keep the output pipes open and Run would not return.
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

// File descriptor type is part of the oracle observation. Wrapping this output
// in a pipe would make the otherwise valid child print the wrong answer.
func TestKeepFilesReportsRealOutput(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	command := exec.Command("sh", "-c", `if [ -f /dev/stdout ]; then printf 'regular\n'; else printf 'pipe\n'; fi; sleep 3600`)
	command.Stdout = file
	started := false
	err = Run(command, Options{KeepFiles: true, Started: func() { started = true }, FirstOutput: 5 * time.Second, Stall: 300 * time.Millisecond})
	var stopped *Error
	if !errors.As(err, &stopped) || stopped.FirstOutput || stopped.Reason != "stalled" || !started {
		t.Fatalf("file output escaped stall guard: %v", err)
	}
	data, err := os.ReadFile(file.Name())
	if err != nil || string(data) != "regular\n" {
		t.Fatalf("descriptor semantics changed: %q %v", data, err)
	}
}
