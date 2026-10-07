package helperslot03

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Not parallel: build each native mutant separately to bound compiler memory.
func TestThemeAddMatchesCohere(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	log, err := os.Create(filepath.Join(t.TempDir(), "validation.log"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, "python3", "validate.py", "--scratch", t.TempDir())
	cmd.Stdout = log
	cmd.Stderr = log
	runErr := cmd.Run()
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(output))
	if runErr != nil {
		t.Fatal(runErr)
	}
}
