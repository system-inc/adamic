package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The wall cap only stops hangs. Scheduling delays do not consume the child's
// CPU budget, which includes sanitizer startup and shutdown.
func runRegExpChild(t *testing.T, binary string, arguments []string, environment string, wallCap, cpuBudget time.Duration) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), wallCap)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Env = os.Environ()
	if environment != "" {
		command.Env = append(command.Env, environment)
	}
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return output, fmt.Errorf("regexp child wall cap %s exceeded: %w", wallCap, ctx.Err())
	}
	if command.ProcessState == nil {
		return output, fmt.Errorf("regexp child has no CPU observation: %w", err)
	}
	cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
	t.Logf("regexp child CPU=%s budget=%s wall-cap=%s", cpu, cpuBudget, wallCap)
	if cpu > cpuBudget {
		return output, fmt.Errorf("regexp child CPU time %s exceeds budget %s", cpu, cpuBudget)
	}
	return output, err
}

func TestRegExpChildWallCap(t *testing.T) {
	// A real nonreturning native fixture proves the cap kills and reaps a hang.
	source := `#include "adamic.h"
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 volatile unsigned spin = 0;
 for (;;) spin++;
}
`
	binary := filepath.Join(t.TempDir(), "hang")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	_, err := runRegExpChild(t, binary, nil, "ASAN_OPTIONS=detect_leaks=0", 250*time.Millisecond, 10*time.Second)
	if err == nil || !strings.Contains(err.Error(), "wall cap") {
		t.Fatalf("hang escaped cap: %v", err)
	}
	t.Logf("nonreturning fixture caught: %v", err)
}

func TestRegExpChildCPUBudget(t *testing.T) {
	source := `#include "adamic.h"
#include <time.h>
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 clock_t start = clock();
 while ((double)(clock() - start) / CLOCKS_PER_SEC < 0.1) {}
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "cpu")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	_, err := runRegExpChild(t, binary, nil, "ASAN_OPTIONS=detect_leaks=1", 5*time.Minute, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "CPU time") {
		t.Fatalf("CPU overrun escaped budget: %v", err)
	}
	t.Logf("CPU overrun caught: %v", err)
}
