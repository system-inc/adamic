package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
)

func TestProbeDeadlineKillsGrandchild(t *testing.T) {
	// Not parallel: cap the real probe entry point's child deadline.
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "fake-go")
	start := time.Now()
	_, err := output(child, "list", "./...")
	if err == nil {
		t.Fatal("hung Go probe passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), child)
}

func TestShardDeadlineKillsGrandchild(t *testing.T) {
	// Not parallel: inject Go through the actual shard launch constructor.
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "go")
	t.Setenv("PATH", filepath.Dir(child)+string(os.PathListSeparator)+os.Getenv("PATH"))
	command, release := shardCommand("go", "test", "./...")
	defer release()
	start := time.Now()
	_, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("hung shard Go passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "go")
}
