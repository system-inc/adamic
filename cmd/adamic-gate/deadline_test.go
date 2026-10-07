package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
)

func TestProbeDeadlineKillsGrandchild(t *testing.T) {
	// Not parallel: cap the real probe entry point's child deadline.
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "fake-go")
	start, err := readyDeadlineRun(t, heartbeat, func() error { _, err := output(child, "list", "./..."); return err })
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
	start, err := readyDeadlineRun(t, heartbeat, func() error {
		command, release := shardCommand("go", "test", "./...")
		defer release()
		_, err := command.CombinedOutput()
		return err
	})
	if err == nil {
		t.Fatal("hung shard Go passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "go")
}

func TestJSONCommandDeadlineKillsGrandchild(t *testing.T) {
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "fake-go-json")
	log := filepath.Join(t.TempDir(), "test.jsonl")
	start, err := readyDeadlineRun(t, heartbeat, func() error {
		command, release := shardCommand(child, "test")
		defer release()
		return runJSONCommand(command, log)
	})
	if err == nil {
		t.Fatal("hung JSON command passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), child)
}

func TestFixtureDiscoveryDeadlineKillsGrandchild(t *testing.T) {
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "go")
	t.Setenv("PATH", filepath.Dir(child)+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	source := filepath.Join(root, "oracle_test.go")
	if err := os.WriteFile(source, []byte("package oracle\nimport \"testing\"\nvar fixtures=[]struct{path string}{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	start, err := readyDeadlineRun(t, heartbeat, func() error { _, err := registeredFixtures(source, "fixtures"); return err })
	if err == nil {
		t.Fatal("hung fixture discovery passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), child)
}

// Not parallel: substitute the execution clock, never the bounded command or
// process-group kill implementation exercised by these real entry points.
func readyDeadlineRun(t *testing.T, heartbeat string, run func() error) (time.Time, error) {
	t.Helper()
	previous := boundedrun.WithTimeout
	factory, ready := testfixture.ReadyContext(t, heartbeat)
	boundedrun.WithTimeout = factory
	t.Cleanup(func() { boundedrun.WithTimeout = previous })
	done := make(chan error, 1)
	go func() { done <- run() }()
	started := testfixture.WaitStarted(t, ready)
	select {
	case err := <-done:
		return started, err
	case <-time.After(5 * time.Second):
		t.Fatal("ready child did not stop within bounded wait")
		return started, nil
	}
}
