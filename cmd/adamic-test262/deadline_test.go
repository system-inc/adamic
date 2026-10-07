package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
)

func TestCommandDeadlineKillsGrandchild(t *testing.T) {
	t.Parallel()
	child, heartbeat := testfixture.Tree(t, "fake-node")
	start := time.Now()
	result := runCommand(200*time.Millisecond, nil, child)
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
	start := time.Now()
	_, err := boundedRuntimeLibrary("../../internal/native/runtime")
	if err == nil {
		t.Fatal("hung runtime clang passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "clang")
}
