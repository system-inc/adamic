package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
)

func TestCommandDeadlineKillsGrandchild(t *testing.T) {
	child, heartbeat := testfixture.Tree(t, "fake-node")
	var result execution
	start := readyDeadlineRun(t, heartbeat, func() { result = runCommand(200*time.Millisecond, nil, child) })
	if !result.TimedOut {
		t.Fatalf("hung Node did not time out: %+v", result)
	}
	testfixture.AssertStopped(t, start, heartbeat, result.Stderr, child)
}

func TestRuntimeLibraryDeadline(t *testing.T) {
	// Not parallel: inject clang into the isolated runtime builder's environment.
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "clang")
	t.Setenv("PATH", filepath.Dir(child)+string(os.PathListSeparator)+os.Getenv("PATH"))
	var err error
	start := readyDeadlineRun(t, heartbeat, func() { _, err = boundedRuntimeLibrary("../../internal/native/runtime") })
	if err == nil {
		t.Fatal("hung runtime clang passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "clang")
}

// Not parallel: install a readiness clock around the real execution entry point.
func readyDeadlineRun(t *testing.T, heartbeat string, run func()) time.Time {
	t.Helper()
	previous := boundedrun.WithTimeout
	factory, ready := testfixture.ReadyContext(t, heartbeat)
	boundedrun.WithTimeout = factory
	t.Cleanup(func() { boundedrun.WithTimeout = previous })
	done := make(chan struct{})
	go func() { run(); close(done) }()
	started := testfixture.WaitStarted(t, ready)
	select {
	case <-done:
		return started
	case <-time.After(5 * time.Second):
		t.Fatal("ready child did not stop")
		return started
	}
}
