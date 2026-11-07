//go:build unix

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// Not parallel: changes the process-wide umask inherited by the generator.
func TestPublishedCopyModeIsPrivate(test *testing.T) {
	previous := syscall.Umask(0)
	defer syscall.Umask(previous)
	root := test.TempDir()
	folder := filepath.Join(root, "internal", "lower")
	if createDirectoryError := os.MkdirAll(folder, 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	input := filepath.Join(folder, "lower.go")
	if writeError := os.WriteFile(
		input,
		[]byte("package lower\ntype lowering struct { state int }\n"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	output := filepath.Join(root, "copy.go")
	logPath := filepath.Join(root, "generation.log")
	log, createError := os.Create(logPath)
	if createError != nil {
		test.Fatal(createError)
	}
	// The generator is built ahead, so the deadline covers its run alone, not a go run's compile.
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(timeoutContext, statecopy(test), root, input, output)
	command.Stdout, command.Stderr = log, log
	runError := command.Run()
	closeError := log.Close()
	if runError != nil {
		text, _ := os.ReadFile(logPath)
		test.Fatalf("generator failed: %v\n%s", runError, text)
	}
	if closeError != nil {
		test.Fatal(closeError)
	}
	info, statError := os.Stat(output)
	if statError != nil {
		test.Fatal(statError)
	}
	if got := info.Mode().Perm(); got != 0600 {
		test.Fatalf("published output mode = %04o, want 0600 with umask 000", got)
	}
}
