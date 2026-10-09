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
func TestPublishedCopyModeIsPrivate(t *testing.T) {
	previous := syscall.Umask(0)
	defer syscall.Umask(previous)
	root := t.TempDir()
	folder := filepath.Join(root, "internal", "lower")
	if err := os.MkdirAll(folder, 0700); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(folder, "lower.go")
	if err := os.WriteFile(input, []byte("package lower\ntype lowering struct { state int }\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "copy.go")
	logPath := filepath.Join(root, "generation.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 85*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", "main.go", root, input, output)
	command.Env = append(os.Environ(), "GOWORK=off")
	command.Stdout, command.Stderr = log, log
	err = command.Run()
	closeErr := log.Close()
	if err != nil {
		text, _ := os.ReadFile(logPath)
		t.Fatalf("generator failed: %v\n%s", err, text)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("published output mode = %04o, want 0600 with umask 000", got)
	}
}
