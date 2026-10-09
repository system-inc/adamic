//go:build unix

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A terminate sent to a mixed-repository run, the way a timeout or a caller's `kill` sends it, reaches the
// project runs it started and the engines they started, and the run ends as terminated only once none of
// them is left (#nqb3mjv). Before, the run ended alone and its Swift child went on, with its engine and
// everything the engine had started.
func TestATerminatedMixedRunLeavesNoChildBehind(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	marks := t.TempDir()
	enginePID := filepath.Join(marks, "engine.pid")
	compilerPID := filepath.Join(marks, "compiler.pid")
	// An engine that starts a long compile of its own and waits on it, as the engine's types phase waits on
	// swift build.
	engine := fakeSwiftEngine(t, "echo $$ > '"+enginePID+".partial' && mv '"+enginePID+".partial' '"+enginePID+"'\n"+
		"sleep 300 &\necho $! > '"+compilerPID+".partial' && mv '"+compilerPID+".partial' '"+compilerPID+"'\nwait")
	root := newRepository(t, nil)
	writeTree(t, filepath.Join(root, "apple", "Toy"), map[string]string{"Package.swift": discoveryPackage, "Sources/Toy/Toy.swift": "let toy = 1\n"})

	run := exec.Command(binary, "--no-fix")
	run.Dir = root
	run.Env = append(os.Environ(), "COHERE_SWIFT_ENGINE="+engine, "COHERE_VERDICT_FD=")
	if err := run.Start(); err != nil {
		t.Fatal(err)
	}
	defer run.Process.Kill()

	readPID := func(path string) int {
		deadline := time.Now().Add(2 * time.Minute)
		for time.Now().Before(deadline) {
			if contents, err := os.ReadFile(path); err == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(contents))); err == nil {
					return pid
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("%s was never written: the engine did not start", filepath.Base(path))
		return 0
	}
	engineProcess, compilerProcess := readPID(enginePID), readPID(compilerPID)

	if err := run.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	// Bounded well under the compile's 300s: a run that waits for its children to end on their own, rather
	// than ending them, also exits as terminated, but only once the compile finishes.
	waited := make(chan error, 1)
	go func() { waited <- run.Wait() }()
	var err error
	select {
	case err = <-waited:
	case <-time.After(time.Minute):
		t.Fatal("a minute after the terminate the run was still waiting on its children")
	}
	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 128+int(syscall.SIGTERM) {
		t.Errorf("a terminated run ended with %v, want exit %d, the terminate's", err, 128+int(syscall.SIGTERM))
	}

	// The engine and its compiler are gone, or going: each was signaled before the run ended, so allow
	// the kernel a moment to reap them.
	for name, pid := range map[string]int{"engine": engineProcess, "compiler": compilerProcess} {
		deadline := time.Now().Add(10 * time.Second)
		for syscall.Kill(pid, 0) == nil && time.Now().Before(deadline) {
			time.Sleep(20 * time.Millisecond)
		}
		if syscall.Kill(pid, 0) == nil {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Errorf("the %s (pid %d) outlived the terminated run", name, pid)
		}
	}
}
