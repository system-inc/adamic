package parser

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Pin one burner to every available CPU. The healthy child shares the first
// CPU, so it continues to make CPU progress despite losing wall-clock time.
func TestParserChildGuard(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "guard-proof.c")
	binary := filepath.Join(directory, "guard-proof")
	const program = `#include <stdio.h>
#include <string.h>
#include <time.h>
static double cpu(void) { struct timespec t; clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &t); return t.tv_sec + t.tv_nsec / 1e9; }
int main(int argc, char **argv) {
 if (strcmp(argv[1], "burn") == 0) { puts("ready"); fflush(stdout); for (;;) {} }
 if (strcmp(argv[1], "hang") == 0) { for (;;) {} }
 double start = cpu(); volatile unsigned long work = 0;
 while (cpu() - start < 2.4) { for (int i = 0; i < 10000; ++i) ++work; }
 puts(argv[2]); return 0;
}
`
	if err := os.WriteFile(source, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := parserGuardOutput(exec.Command("clang", "-O2", source, "-o", binary), parserBuildCPUBudget); err != nil {
		t.Fatalf("proof build: %v %s", err, output)
	}
	name := "testdata/recovery/corePublic.ts-cut-37.ts.txt"
	t.Run("hang", func(t *testing.T) {
		command := exec.Command(binary, "hang", name)
		started := time.Now()
		err := parserRunGuard(command, time.Second)
		if err == nil || !strings.Contains(err.Error(), "stalled:") || !strings.Contains(err.Error(), name) {
			t.Fatalf("hang must be named as stalled: %v", err)
		}
		t.Logf("hang killed: wall %s, CPU %s; %v", time.Since(started), command.ProcessState.UserTime()+command.ProcessState.SystemTime(), err)
	})
	t.Run("loaded", func(t *testing.T) {
		status, err := os.ReadFile("/proc/self/status")
		if err != nil {
			t.Fatal(err)
		}
		var cpus []string
		for _, line := range strings.Split(string(status), "\n") {
			if !strings.HasPrefix(line, "Cpus_allowed_list:") {
				continue
			}
			for _, part := range strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "Cpus_allowed_list:")), ",") {
				bounds := strings.Split(part, "-")
				first, err := strconv.Atoi(bounds[0])
				if err != nil {
					t.Fatal(err)
				}
				last := first
				if len(bounds) == 2 {
					last, err = strconv.Atoi(bounds[1])
					if err != nil {
						t.Fatal(err)
					}
				}
				for cpu := first; cpu <= last; cpu++ {
					cpus = append(cpus, strconv.Itoa(cpu))
				}
			}
		}
		if len(cpus) == 0 {
			t.Fatal("no available CPUs")
		}
		for _, cpu := range cpus {
			burner := exec.Command("taskset", "-c", cpu, binary, "burn")
			ready, err := burner.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := burner.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { burner.Process.Kill(); burner.Wait() })
			if line, err := bufio.NewReader(ready).ReadString('\n'); err != nil || line != "ready\n" {
				t.Fatalf("burner CPU %s: %q %v", cpu, line, err)
			}
		}
		command := exec.Command("taskset", "-c", cpus[0], binary, "healthy", name)
		started := time.Now()
		output, err := parserGuardOutput(command, 3*time.Second)
		if err != nil || strings.TrimSpace(string(output)) != name {
			t.Fatalf("healthy child under load: %v %q", err, output)
		}
		t.Logf("healthy survived %d pinned CPU burners (NumCPU %d): wall %s, CPU %s, budget 3s; case %s", len(cpus), runtime.NumCPU(), time.Since(started), command.ProcessState.UserTime()+command.ProcessState.SystemTime(), name)
	})
	fmt.Fprintln(os.Stdout, "parser guard proofs complete")
}
