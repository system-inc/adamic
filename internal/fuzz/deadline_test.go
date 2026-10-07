package fuzz

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
)

func TestExecuteDeadlineKillsGrandchild(t *testing.T) {
	t.Parallel()
	child, heartbeat := testfixture.Tree(t, "fake-compiler")
	start := time.Now()
	result := execute(t.TempDir(), nil, 200*time.Millisecond, child)
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
	start := time.Now()
	_, err := Prepare("../..", t.TempDir())
	if err == nil {
		t.Fatal("hung Go build passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "go")
}
