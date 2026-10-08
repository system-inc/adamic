package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCompileCPUBudgets(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("requires CPU limits")
	}
	dir := t.TempDir()
	source, binary := filepath.Join(dir, "fake.c"), filepath.Join(dir, "fake")
	// A fake compiler/clang uses clock(), so only its own CPU advances the work.
	code := `#include <stdio.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
int main(int argc, char **argv) {
 int worker = strstr(argv[1], "worker") != NULL;
 do {
  if (worker) { char line[4096]; if (!fgets(line, sizeof(line), stdin)) return 0; }
  if (strstr(argv[1], "block")) { for (;;) pause(); }
  clock_t start = clock();
  for (;;) {
   if ((strcmp(argv[1], "work") == 0 || strcmp(argv[1], "worker-work") == 0) && clock() - start >= CLOCKS_PER_SEC) break;
  }
  if (worker) puts("{\"Stdout\":\"Y29tcGxldGVkCg==\"}"); else puts("completed");
  fflush(stdout);
 } while (worker);
 return 0;
}
`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("clang", "-O2", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("build fake compiler: %v\n%s", err, out)
	}
	for _, kind := range []string{"compile", "clang", "worker"} {
		t.Run(kind, func(t *testing.T) {
			if kind == "worker" && runtime.GOOS != "linux" {
				t.Skip("Darwin retains the worker wall deadline")
			}
			run := func(mode string, budget, wall time.Duration) execution {
				if kind != "worker" {
					limit := outputLimit
					if kind == "compile" {
						limit = 16 << 20
					}
					if mode == "block" {
						return runCPUCommand(budget, wall, nil, limit, binary, mode)
					}
					if kind == "compile" {
						return runCommandWithLimit(budget, nil, limit, binary, mode)
					}
					return runCommand(budget, nil, binary, mode)
				}
				command := exec.Command(binary, "worker-"+mode)
				input, err := command.StdinPipe()
				if err != nil {
					t.Fatal(err)
				}
				output, err := command.StdoutPipe()
				if err != nil {
					t.Fatal(err)
				}
				if err := command.Start(); err != nil {
					t.Fatal(err)
				}
				worker := &compilerWorker{command: command, input: input, output: output, encoder: json.NewEncoder(input), decoder: json.NewDecoder(output), cpuLimit: budget, wallTimeout: wall}
				defer worker.close()
				result := worker.compile("unused.a")
				if result.TimedOut && (worker.command != nil || command.ProcessState == nil) {
					t.Fatal("worker not stopped and reaped")
				}
				// Two successful requests must each get a fresh budget, not a process lifetime limit.
				if mode == "work" && !result.TimedOut && result.Exit == 0 {
					result = worker.compile("unused.a")
				}
				return result
			}
			t.Run("spin", func(t *testing.T) {
				start := time.Now()
				result := run("spin", time.Second, time.Minute)
				if !result.TimedOut || result.Exit != -1 || result.Signal != "" {
					t.Fatalf("spin: %+v", result)
				}
				if time.Since(start) >= time.Minute {
					t.Fatal("spin reached wall backstop")
				}
			})
			t.Run("block", func(t *testing.T) {
				start := time.Now()
				result := run("block", time.Minute, 100*time.Millisecond)
				if !result.TimedOut || result.Exit != -1 {
					t.Fatalf("block: %+v", result)
				}
				if time.Since(start) >= 5*time.Second {
					t.Fatal("blocking child not promptly stopped")
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
				const budget = 1500 * time.Millisecond
				start := time.Now()
				result := run("work", budget, 2*time.Minute)
				elapsed := time.Since(start)
				if result.TimedOut || result.Exit != 0 || result.Stdout != "completed\n" {
					t.Fatalf("one CPU second under load: %+v after %s", result, elapsed)
				}
				// As in TestProgramCPUDeadline, scheduling determines whether wall exceeds
				// the budget on a particular box; report it without a flaky assertion.
				t.Logf("CPU budget %s, wall %s (past budget: %t)", budget, elapsed, elapsed > budget)
			})
		})
	}
}
