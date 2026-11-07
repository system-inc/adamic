package executable

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWriteLeavesNoWritableDescriptorForSiblingForksToExec(test *testing.T) {
	test.Parallel()
	dir := test.TempDir()
	// The forks run this test binary with no test selected, so the pressure needs no host tool.
	self, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
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
				if runError := exec.CommandContext(
					ctx,
					self,
					"-test.run=^$",
				).Run(); runError != nil && ctx.Err() == nil {
					failures <- runError
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
			test.Fatal("fork stress deadline", err)
		}
		path := filepath.Join(dir, "compiler")
		if writeError := Write(path, []byte("#!/bin/sh\necho ready\n")); writeError != nil {
			test.Fatal(writeError)
		}
		output, combinedOutputError := exec.CommandContext(ctx, path).CombinedOutput()
		if combinedOutputError != nil || string(output) != "ready\n" {
			test.Fatalf("executable %d under concurrent forks: %v, output %q", program, combinedOutputError, output)
		}
	}
	cancel()
	forks.Wait()
	close(failures)
	for err := range failures {
		test.Errorf("fork pressure failed: %v", err)
	}
}
