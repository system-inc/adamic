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
	if err := os.MkdirAll(folder, 0700); err != nil {
		test.Fatal(err)
	}
	input := filepath.Join(folder, "lower.go")
	if err := os.WriteFile(input, []byte("package lower\ntype lowering struct { state int }\n"), 0600); err != nil {
		test.Fatal(err)
	}
	output := filepath.Join(root, "copy.go")
	logPath := filepath.Join(root, "generation.log")
	log, err := os.Create(logPath)
	if err != nil {
		test.Fatal(err)
	}
	// The generator is built ahead, so the deadline covers its run alone, not a go run's compile.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, statecopy(test), root, input, output)
	command.Stdout, command.Stderr = log, log
	err = command.Run()
	closeErr := log.Close()
	if err != nil {
		text, _ := os.ReadFile(logPath)
		test.Fatalf("generator failed: %v\n%s", err, text)
	}
	if closeErr != nil {
		test.Fatal(closeErr)
	}
	info, err := os.Stat(output)
	if err != nil {
		test.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		test.Fatalf("published output mode = %04o, want 0600 with umask 000", got)
	}
}
