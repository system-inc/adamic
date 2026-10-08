package oracle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// CPU time excludes scheduler delays. The wall cap only kills true hangs;
// killing the process group also stops an external tool's native child.
func regExpCPUCommand(environment []string, name string, arguments []string, wallCap, cpuBudget time.Duration) (run, time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), wallCap)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Env = append(os.Environ(), environment...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		return run{}, 0, fmt.Errorf("regexp child wall cap %s exceeded: %w", wallCap, ctx.Err())
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

func TestRegExpOracleWallCap(t *testing.T) {
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
	_, _, err := regExpCPUCommand([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, nil, 250*time.Millisecond, 3*time.Second)
	if err == nil || !strings.Contains(err.Error(), "wall cap") {
		t.Fatalf("hang escaped cap: %v", err)
	}
	t.Logf("nonreturning fixture caught: %v", err)
}

func TestRegExpOracleCPUBudget(t *testing.T) {
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
	_, cpu, err := regExpCPUCommand([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, nil, 5*time.Minute, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "CPU time") {
		t.Fatalf("CPU overrun escaped budget: %v", err)
	}
	t.Logf("CPU overrun caught: CPU=%s %v", cpu, err)
}
