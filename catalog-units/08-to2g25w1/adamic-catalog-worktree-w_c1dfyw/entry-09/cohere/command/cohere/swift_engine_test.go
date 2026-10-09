//go:build unix

// The stand-in engines here are /bin/sh scripts, and the process-group behavior they prove exists only
// on Unix, so the file builds there.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// fakeSwiftEngine writes a shell script standing in for cohere-swift, so the process boundary can be driven
// through every way an engine ends without building the real one.
func fakeSwiftEngine(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cohere-swift")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func runFakeSwiftEngine(t *testing.T, body string, mode swiftMode) (string, string, int, error) {
	t.Helper()
	var out, standardError bytes.Buffer
	exitCode, err := runEngineBinary(fakeSwiftEngine(t, body), []string{"--contract", "1"}, newSwiftRun(&out, mode, "", time.Now()), &standardError)
	return out.String(), standardError.String(), exitCode, err
}

// Not parallel: its subtests drive check runs through swiftRun, whose record acceptors set the package-level
// activeSummary
func TestRunEngineBinaryEndings(t *testing.T) {
	fixtures, err := filepath.Abs(filepath.Join("..", "..", "swift", "Contract"))
	if err != nil {
		t.Fatal(err)
	}

	// Not parallel: it drives a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("a real stream streams and exits with its summary", func(t *testing.T) {
		out, _, exitCode, err := runFakeSwiftEngine(t, "cat '"+filepath.Join(fixtures, "Findings.jsonl")+"'; exit 1", swiftModeCheck)
		if err != nil || exitCode != 1 {
			t.Fatalf("exit %d, err %v\n%s", exitCode, err, out)
		}
		if !strings.Contains(out, "[cohere-swift/force-unwrapping/forceUnwrap]") || !strings.Contains(out, "phases: ") {
			t.Errorf("the stream was not rendered:\n%s", out)
		}
	})

	// Not parallel: it drives a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("the control: a clean stream exits 0", func(t *testing.T) {
		_, _, exitCode, err := runFakeSwiftEngine(t, "cat '"+filepath.Join(fixtures, "Clean.jsonl")+"'", swiftModeCheck)
		if err != nil || exitCode != 0 {
			t.Fatalf("a clean engine run exited %d, err %v", exitCode, err)
		}
	})

	// Not parallel: it drives a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("exit 2 passes stderr through and says nothing was checked", func(t *testing.T) {
		_, standardError, exitCode, err := runFakeSwiftEngine(t, "echo 'cohere-swift: --timing is not implemented for Swift yet' >&2; exit 2", swiftModeCheck)
		if err == nil || exitCode != 1 || !strings.Contains(err.Error(), "did not run, so nothing was checked") {
			t.Fatalf("exit %d, err %v", exitCode, err)
		}
		if !strings.Contains(standardError, "--timing is not implemented") {
			t.Errorf("the engine's own sentence was not passed through: %q", standardError)
		}
	})

	// Not parallel: it drives a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("a signal is named as a signal, never as an exit code", func(t *testing.T) {
		out, _, exitCode, err := runFakeSwiftEngine(t, "head -2 '"+filepath.Join(fixtures, "Clean.jsonl")+"'; kill -9 $$", swiftModeCheck)
		if err == nil || exitCode != 1 {
			t.Fatalf("a killed engine exited %d, err %v", exitCode, err)
		}
		if !strings.Contains(err.Error(), "was killed by") {
			t.Errorf("the signal was not named: %v", err)
		}
		if strings.Contains(err.Error(), "exited -1") {
			t.Errorf("printed an exit code the process never chose: %v", err)
		}
		if !strings.Contains(out, "did not run (the Swift engine was killed") {
			t.Errorf("the phase line did not say what was not reached:\n%s", out)
		}
	})

	// Not parallel: it drives a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("a broken record kills the engine rather than waiting on it", func(t *testing.T) {
		start := time.Now()
		_, _, exitCode, err := runFakeSwiftEngine(t, "head -1 '"+filepath.Join(fixtures, "Clean.jsonl")+"'; echo 'Compiling swift-syntax'; sleep 30", swiftModeCheck)
		if err == nil || exitCode != 1 || !strings.Contains(err.Error(), "the Swift engine is broken") {
			t.Fatalf("exit %d, err %v", exitCode, err)
		}
		if elapsed := time.Since(start); elapsed > 10*time.Second {
			t.Errorf("waited %s for an engine that had already broken its contract", elapsed)
		}
	})

	// The child is started before the broken record, so it is always alive when the engine is killed.
	// Killing only the engine left it holding the pipes, and the front door waited out its 30 seconds.
	// Not parallel: it drives a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("a broken engine's children die with it", func(t *testing.T) {
		childFile := filepath.Join(t.TempDir(), "child")
		start := time.Now()
		_, _, exitCode, err := runFakeSwiftEngine(t,
			"sleep 30 & echo $! > '"+childFile+"'; head -1 '"+filepath.Join(fixtures, "Clean.jsonl")+"'; echo 'Compiling swift-syntax'; wait",
			swiftModeCheck)
		if err == nil || exitCode != 1 || !strings.Contains(err.Error(), "the Swift engine is broken") {
			t.Fatalf("exit %d, err %v", exitCode, err)
		}
		if elapsed := time.Since(start); elapsed > 10*time.Second {
			t.Errorf("waited %s on a child of an engine that had already broken its contract", elapsed)
		}
		requireProcessEnded(t, childFile)
	})

	// SIGTERM rather than an interrupt, because a POSIX shell starts a background job with interrupts
	// ignored, and the stand-in's child must be one the signal can end.
	// Not parallel: it drives a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("a signal cohere receives is passed to the engine and everything it started", func(t *testing.T) {
		childFile := filepath.Join(t.TempDir(), "child")
		interrupts := make(chan os.Signal, 1)
		var out, standardError bytes.Buffer
		run := newSwiftRun(&out, swiftModeCheck, "", time.Now())
		go func() {
			waitForFile(t, childFile)
			interrupts <- syscall.SIGTERM
		}()
		start := time.Now()
		exitCode, err := runEngineBinaryUntil(
			fakeSwiftEngine(t, "sleep 30 & echo $! > '"+childFile+"'; head -2 '"+filepath.Join(fixtures, "Clean.jsonl")+"'; wait"),
			[]string{"--contract", "1"}, run, &standardError, interrupts)
		if err == nil || exitCode != 1 || !strings.Contains(err.Error(), "was killed by") {
			t.Fatalf("an engine signalled through cohere exited %d, err %v\n%s", exitCode, err, out.String())
		}
		if elapsed := time.Since(start); elapsed > 10*time.Second {
			t.Errorf("the signal took %s to end the engine, so it was not passed on", elapsed)
		}
		requireProcessEnded(t, childFile)
	})
}

// waitForFile waits for a stand-in engine to write a file, failing the test after ten seconds.
func waitForFile(t *testing.T, path string) {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if contents, err := os.ReadFile(path); err == nil && len(contents) > 0 {
			return
		}
	}
	t.Errorf("the stand-in engine never wrote %s", path)
}

// requireProcessEnded fails unless the process whose id a stand-in engine wrote to path is gone. An
// orphan is reaped by launchd or init a moment after it dies, so it is given two seconds.
func requireProcessEnded(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the stand-in engine never wrote its child's id: %v", err)
	}
	child, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		t.Fatalf("the child id %q is not a number: %v", contents, err)
	}
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if syscall.Kill(child, 0) == syscall.ESRCH {
			return
		}
	}
	_ = syscall.Kill(child, syscall.SIGKILL)
	t.Errorf("the engine's child %d outlived it", child)
}
