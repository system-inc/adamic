package childguard

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
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
	case "descriptors":
		for _, file := range []*os.File{os.Stdout, os.Stderr} {
			info, err := file.Stat()
			if err != nil {
				os.Exit(2)
			}
			if info.Mode().IsRegular() {
				fmt.Fprintln(file, "regular")
			} else {
				fmt.Fprintln(file, "pipe")
			}
		}
	case "started":
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() || scanner.Text() != "started" {
			os.Exit(3)
		}
		fmt.Fprintln(os.Stdout, scanner.Text())
	case "external":
		for i := 0; i < 6; i++ {
			fmt.Fprintln(os.Stdout, "progress")
			time.Sleep(100 * time.Millisecond)
		}
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

// File descriptor type is part of the oracle observation. Wrapping this output
// in a pipe would make the otherwise valid child print the wrong answer.
func TestKeepFilesReportsRealOutput(t *testing.T) {
	t.Parallel()
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

// Node observes the same real descriptors independently of the guard. A wrapping
// mutant exits cleanly and changes only the bytes compared with Node.
func TestPreservedDescriptorsAgreeWithNode(t *testing.T) {
	t.Parallel()
	files := func() (*os.File, *os.File) {
		t.Helper()
		out, err := os.CreateTemp(t.TempDir(), "stdout")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { out.Close() })
		stderr, err := os.CreateTemp(t.TempDir(), "stderr")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { stderr.Close() })
		return out, stderr
	}
	expectedOut, expectedErr := files()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	node := exec.CommandContext(ctx, "node", "-e", "const fs = require('node:fs'); console.log(fs.fstatSync(1).isFile() ? 'regular' : 'pipe'); console.error(fs.fstatSync(2).isFile() ? 'regular' : 'pipe');")
	node.Stdout, node.Stderr = expectedOut, expectedErr
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	actualOut, actualErr := files()
	command := child("descriptors")
	command.Stdout, command.Stderr = actualOut, actualErr
	if err := Run(command, Options{KeepFiles: true, Ceiling: 5 * time.Second}); err != nil {
		t.Fatalf("descriptor child must exit cleanly: %v", err)
	}
	for i, pair := range [][2]*os.File{{actualOut, expectedOut}, {actualErr, expectedErr}} {
		actual, err := os.ReadFile(pair[0].Name())
		if err != nil {
			t.Fatal(err)
		}
		expected, err := os.ReadFile(pair[1].Name())
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, expected) {
			t.Fatalf("descriptor %d differs from Node: got %q want %q", i, actual, expected)
		}
	}
}

func TestStartedCallbackRunsAfterStart(t *testing.T) {
	t.Parallel()
	command := child("started")
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	var output bytes.Buffer
	command.Stdout = &output
	started := false
	err = Run(command, Options{FirstOutput: time.Second, Ceiling: 2 * time.Second, Started: func() {
		if command.Process == nil {
			t.Error("callback ran before Start")
		}
		started = true
		if _, err := io.WriteString(input, "started\n"); err != nil {
			t.Error(err)
		}
		input.Close()
	}})
	if err != nil || !started || output.String() != "started\n" {
		t.Fatalf("startup callback: started=%t output=%q err=%v", started, output.String(), err)
	}
}

func TestExternalProgressKeepsOutputAlive(t *testing.T) {
	t.Parallel()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	var count atomic.Uint64
	output := make(chan []byte, 1)
	go func() {
		var all []byte
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer)
			if n > 0 {
				all = append(all, buffer[:n]...)
				count.Add(uint64(n))
			}
			if err != nil {
				break
			}
		}
		output <- all
	}()
	command := child("external")
	command.Stdout = writer
	err = Run(command, Options{KeepFiles: true, Progress: count.Load, FirstOutput: 250 * time.Millisecond, Stall: 250 * time.Millisecond, Ceiling: 2 * time.Second})
	writer.Close()
	data := <-output
	if err != nil || string(data) != strings.Repeat("progress\n", 6) {
		t.Fatalf("external output lost or killed: %q %v", data, err)
	}
}

func TestUnchangedExternalProgressStillStalls(t *testing.T) {
	t.Parallel()
	err := Run(child("silent"), Options{Progress: func() uint64 { return 1 }, FirstOutput: time.Second, Stall: 200 * time.Millisecond, Ceiling: 2 * time.Second})
	var stopped *Error
	if !errors.As(err, &stopped) || stopped.Reason != "stalled" || stopped.FirstOutput {
		t.Fatalf("unchanged external count kept a child alive: %v", err)
	}
}
