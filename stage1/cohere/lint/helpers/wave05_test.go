package helpers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Not parallel: captures upstream consumer fixtures into this unit's named scratch directory.
func TestWave05LeadingInteger(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	logPath := filepath.Join(t.TempDir(), "leading-integer.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd := exec.CommandContext(ctx, "python3", "wave05/validate.py")
	cmd.Stdout = log
	cmd.Stderr = log
	if err = cmd.Run(); err != nil {
		data, _ := os.ReadFile(logPath)
		t.Fatalf("leadingInteger comparison: %v\n%s", err, data)
	}
	data, _ := os.ReadFile(logPath)
	t.Log(string(data))
}
