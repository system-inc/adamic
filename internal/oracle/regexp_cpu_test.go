package oracle

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
)

// CPU time excludes scheduler delays. The shared output guard kills stalled
// process groups independently of the semantic CPU budget.
func regExpCPUCommand(environment []string, name string, arguments []string, guard childguard.Options, cpuBudget time.Duration) (run, time.Duration, error) {
	command := exec.Command(name, arguments...)
	command.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := runChildWith(command, guard)
	var stopped *childguard.Error
	if errors.As(err, &stopped) {
		return run{}, 0, err
	}
	if command.ProcessState == nil {
		return run{}, 0, fmt.Errorf("regexp child has no CPU observation: %w", err)
	}
	cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
	result := run{stdout.Bytes(), stderr.Bytes(), command.ProcessState.ExitCode()}
	if cpuBudget > 0 && cpu > cpuBudget {
		return result, cpu, fmt.Errorf("regexp child CPU time %s exceeds budget %s", cpu, cpuBudget)
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return result, cpu, err
	}
	return result, cpu, nil
}

func TestRegExpOracleStall(t *testing.T) {
	t.Parallel()
	source := `#include "adamic.h"
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 volatile unsigned spin = 0;
 for (;;) spin++;
}
`
	binary := filepath.Join(t.TempDir(), "hang")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	_, _, err := regExpCPUCommand([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, nil, childguard.Options{FirstOutput: 250 * time.Millisecond}, 3*time.Second)
	if err == nil || !strings.Contains(err.Error(), "stalled: no first output") {
		t.Fatalf("hang escaped guard: %v", err)
	}
	t.Logf("nonreturning fixture caught: %v", err)
}

func TestRegExpOracleCPUBudget(t *testing.T) {
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
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	_, cpu, err := regExpCPUCommand([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, nil, childguard.Options{}, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "CPU time") {
		t.Fatalf("CPU overrun escaped budget: %v", err)
	}
	t.Logf("CPU overrun caught: CPU=%s %v", cpu, err)
}
