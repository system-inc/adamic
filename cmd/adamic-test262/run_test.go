package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestProgramCPUDeadline(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("CPU limits require Linux or macOS")
	}
	// clock() measures this child's CPU time, independent of scheduling delays.
	directory := t.TempDir()
	source := filepath.Join(directory, "cpu.c")
	binary := filepath.Join(directory, "cpu")
	if err := os.WriteFile(source, []byte(`#include <stdio.h>
#include <time.h>
int main(int argc, char **argv) {
    clock_t start = clock();
    for (;;) {
        if (argc > 1 && clock() - start >= CLOCKS_PER_SEC) {
            puts("completed");
            return 0;
        }
    }
}
`), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("clang", "-O2", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("build CPU helper: %v\n%s", err, output)
	}
	t.Run("spin", func(t *testing.T) {
		start := time.Now()
		result := runProgram(time.Second, nil, binary)
		if !result.TimedOut || result.Exit != -1 || result.Signal != "" {
			t.Fatalf("spinning child: %+v, want TimedOut with Exit -1", result)
		}
		if time.Since(start) >= time.Minute {
			t.Fatal("spinning child reached the wall backstop instead of the CPU limit")
		}
	})
	t.Run("saturated", func(t *testing.T) {
		previous := runtime.GOMAXPROCS(2 * runtime.NumCPU())
		defer runtime.GOMAXPROCS(previous)
		var stop atomic.Bool
		var ready, workers sync.WaitGroup
		for range 2 * runtime.NumCPU() {
			ready.Add(1)
			workers.Go(func() {
				ready.Done()
				for !stop.Load() {
				}
			})
		}
		defer func() { stop.Store(true); workers.Wait() }()
		ready.Wait()
		const budget = 2 * time.Second
		start := time.Now()
		result := runProgram(budget, nil, binary, "one-second")
		elapsed := time.Since(start)
		if result.TimedOut || result.Exit != 0 || result.Stdout != "completed\n" {
			t.Fatalf("1 CPU second under load: %+v after %s; want successful completion", result, elapsed)
		}
		if elapsed <= budget {
			t.Fatalf("load did not extend wall time beyond %s: %s", budget, elapsed)
		}
		t.Logf("1 CPU second completed in %s with a %s CPU budget", elapsed, budget)
	})
}
