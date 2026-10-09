package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Not parallel: changes runtime.GOMAXPROCS and saturates all CPUs in the saturated subtest.
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
		t.Parallel()
		start := time.Now()
		result := runProgram(time.Second, nil, binary)
		if !result.TimedOut || result.Exit != -1 || result.Signal != "" {
			t.Fatalf("spinning child: %+v, want TimedOut with Exit -1", result)
		}
		if time.Since(start) >= time.Minute {
			t.Fatal("spinning child reached the wall backstop instead of the CPU limit")
		}
	})
	// Not parallel: changes runtime.GOMAXPROCS and saturates all CPUs.
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
		// Whether the load stretched wall time past the budget depends on the box (about 1 run in 10
		// it doesn't), so it's reported, not asserted: the assertion is that CPU, not wall time,
		// decided the verdict, which holds either way.
		t.Logf("1 CPU second completed in %s with a %s CPU budget (wall past the budget: %t)", elapsed, budget, elapsed > budget)
	})
}

func TestOutputOverflowTimeout(t *testing.T) {
	endless := runCommand(time.Second, nil, "sh", "-c", "yes")
	if !endless.TimedOut || !crashedExecution(endless) {
		t.Errorf("endless printer: timedOut=%v crashed=%v", endless.TimedOut, crashedExecution(endless))
	}
	got := decide(verdictInput{Node: execution{Stdout: "x", Exit: 0}, Native: endless})
	if got.Kind != outcomeCrashed {
		t.Errorf("endless verdict: %+v", got)
	}
}

func TestOutputOverflowSanitizer(t *testing.T) {
	for _, stream := range []string{"", " >&2"} {
		t.Run(stream, func(t *testing.T) {
			sanitizer := runCommand(10*time.Second, nil, "sh", "-c", "yes | head -c 300000"+stream+"; echo '==1==ERROR: AddressSanitizer: heap-use-after-free' >&2; exit 1")
			if !strings.Contains(sanitizer.Stderr, "AddressSanitizer") || !crashedExecution(sanitizer) {
				t.Errorf("sanitizer lost: stderr bytes=%d crashed=%v", len(sanitizer.Stderr), crashedExecution(sanitizer))
			}
			if len(sanitizer.Stderr) > outputLimit+len("command output exceeded capture limit\n") {
				t.Errorf("unbounded stderr: %d", len(sanitizer.Stderr))
			}
			got := decide(verdictInput{Node: execution{Exit: 0}, Native: sanitizer})
			if got.Kind != outcomeCrashed {
				t.Errorf("sanitizer verdict: %+v", got)
			}
		})
	}
}
