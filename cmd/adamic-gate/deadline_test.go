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

func TestJSONCommandDeadlineKillsGrandchild(t *testing.T) {
	t.Setenv("ADAMIC_CHILD_DEADLINE", "0.2")
	child, heartbeat := testfixture.Tree(t, "fake-go-json")
	command, release := shardCommand(child, "test")
	defer release()
	start := time.Now()
	err := runJSONCommand(command, filepath.Join(t.TempDir(), "test.jsonl"))
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
	start := time.Now()
	_, err := registeredFixtures(source, "fixtures")
	if err == nil {
		t.Fatal("hung fixture discovery passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), child)
}
