package native

import (
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The output guard stops hangs. Scheduling delays do not consume the child's
// CPU budget, which includes sanitizer startup and shutdown.
func runRegExpChild(t *testing.T, binary string, arguments []string, environment string, guard childguard.Options, cpuBudget time.Duration) ([]byte, error) {
	t.Helper()
	command := exec.Command(binary, arguments...)
	command.Env = os.Environ()
	if environment != "" {
		command.Env = append(command.Env, environment)
	}
	output, err := childguard.CombinedOutput(command, guard)
	var stalled *childguard.Error
	if errors.As(err, &stalled) {
		return output, err
	}

	if command.ProcessState == nil {
		return output, fmt.Errorf("regexp child has no CPU observation: %w", err)
	}
	cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
	t.Logf("regexp child CPU=%s budget=%s", cpu, cpuBudget)
	if cpu > cpuBudget {
		return output, fmt.Errorf("regexp child CPU time %s exceeds budget %s", cpu, cpuBudget)
	}
	return output, err
}

func TestRegExpChildStall(t *testing.T) {
	t.Parallel()
	// A real nonreturning native fixture proves the guard kills and reaps a hang.
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
	_, err := runRegExpChild(t, binary, nil, "ASAN_OPTIONS=detect_leaks=0", childguard.Options{FirstOutput: 250 * time.Millisecond}, 10*time.Second)
	if err == nil || !strings.Contains(err.Error(), "stalled") {
		t.Fatalf("hang escaped guard: %v", err)
	}
	t.Logf("nonreturning fixture caught: %v", err)
}

func TestRegExpChildCPUBudget(t *testing.T) {
	t.Parallel()
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
	_, err := runRegExpChild(t, binary, nil, "ASAN_OPTIONS=detect_leaks=0", childguard.Options{}, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "CPU time") {
		t.Fatalf("CPU overrun escaped budget: %v", err)
	}
	t.Logf("CPU overrun caught: %v", err)
}
