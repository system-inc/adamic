package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCfgIsDestructuringTargetMatchesGoWithMutant(t *testing.T) {
	// Not parallel: each invocation builds two sanitizer-enabled compiler artifacts.
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	log, err := os.Create(filepath.Join(scratch, "test.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	command := exec.Command("python3", filepath.Join(directory, "validate_pattern.py"), "--scratch", scratch, "--replay")
	command.Stdout = log
	command.Stderr = log
	if err := command.Run(); err != nil {
		data, _ := os.ReadFile(log.Name())
		t.Fatalf("comparison failed: %v\n%s", err, data)
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
