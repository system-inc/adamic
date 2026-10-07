package fuzz

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
)

func TestExecuteDeadlineKillsGrandchild(t *testing.T) {
	child, heartbeat := testfixture.Tree(t, "fake-compiler")
	directory := t.TempDir()
	var result Run
	start := readyDeadlineRun(t, heartbeat, func() { result = execute(directory, nil, 200*time.Millisecond, child) })
	if !result.TimedOut {
		t.Fatalf("hung compiler did not time out: %+v", result)
	}
	testfixture.AssertStopped(t, start, heartbeat, string(result.Stderr), child)
}

func TestPrepareBuildDeadline(t *testing.T) {
	// Not parallel: inject the shared generator/reducer preparation child.
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "go")
	t.Setenv("PATH", filepath.Dir(child)+string(os.PathListSeparator)+os.Getenv("PATH"))
	work := t.TempDir()
	var err error
	start := readyDeadlineRun(t, heartbeat, func() { _, err = Prepare("../..", work) })
	if err == nil {
		t.Fatal("hung Go build passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "go")
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
