package oracles

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestExecutableForkIsolation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// The forks run this test binary with no test selected, so the pressure needs no host tool.
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	var forks sync.WaitGroup
	ready := make(chan struct{}, 8)
	failures := make(chan error, 8)
	for worker := 0; worker < 8; worker++ {
		forks.Add(1)
		go func() {
			defer forks.Done()
			ready <- struct{}{}
			for ctx.Err() == nil {
				// TestMain changes to the lint directory, so each child must start in oracles.
				command := exec.CommandContext(ctx, self, "-test.run=^$")
				command.Dir = filepath.Join(packageDirectory, "oracles")
				if err := command.Run(); err != nil && ctx.Err() == nil {
					failures <- err
					return
				}
			}
		}()
	}
	defer func() { cancel(); forks.Wait() }()
	for worker := 0; worker < 8; worker++ {
		<-ready
	}
	for program := 0; program < 200; program++ {
		if err := ctx.Err(); err != nil {
			t.Fatal("fork stress deadline", err)
		}
		path := filepath.Join(dir, "compiler")
		if err := writeExecutable(path, []byte("#!/bin/sh\necho ready\n")); err != nil {
			t.Fatal(err)
		}
		output, err := exec.CommandContext(ctx, path).CombinedOutput()
		if err != nil || string(output) != "ready\n" {
			t.Fatalf("executable %d under concurrent forks: %v, output %q", program, err, output)
		}
	}
	cancel()
	forks.Wait()
	close(failures)
	for err := range failures {
		t.Errorf("fork pressure failed: %v", err)
	}
}
