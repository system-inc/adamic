//go:build linux

package childguard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Use one core and an equal-priority busy sibling to exercise scheduler delay
// without needing privileged cgroup writes. The child must progress for several
// complete silence windows, both before and after its first output.
func TestCPUProgressUnderLoad(t *testing.T) {
	for _, first := range []bool{true, false} {
		t.Run(fmt.Sprintf("first-output-%t", first), func(t *testing.T) {
			status, err := os.ReadFile("/proc/self/status")
			if err != nil {
				t.Fatal(err)
			}
			core := ""
			for _, line := range strings.Split(string(status), "\n") {
				if strings.HasPrefix(line, "Cpus_allowed_list:") {
					core = strings.Fields(line)[1]
					core = strings.Split(strings.Split(core, ",")[0], "-")[0]
				}
			}
			if _, err := exec.LookPath("taskset"); err != nil {
				t.Fatal("load simulation requires taskset:", err)
			}
			sibling := exec.Command("taskset", "-c", core, "sh", "-c", "while :; do :; done")
			if err := sibling.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = sibling.Process.Kill(); _ = sibling.Wait() }()
			cmd := exec.Command("taskset", "-c", core, os.Args[0], "-test.run=^TestCPUBusyChild$")
			cmd.Env = append(os.Environ(), "CHILDGUARD_CPU_BUSY=1", "CHILDGUARD_CPU_FIRST="+strconv.FormatBool(first), "GORACE=atexit_sleep_ms=0")
			start := time.Now()
			if err := Run(cmd, Options{FirstOutput: 400 * time.Millisecond, Stall: 400 * time.Millisecond, Ceiling: 10 * time.Second}); err != nil {
				t.Fatalf("silent CPU progress killed under load: %v", err)
			}
			cpu := cmd.ProcessState.UserTime() + cmd.ProcessState.SystemTime()
			if time.Since(start) < 1600*time.Millisecond || cpu < 100*time.Millisecond {
				t.Fatalf("load probe did not exercise CPU progress: wall %s CPU %s", time.Since(start), cpu)
			}
			t.Logf("silent child survived: wall %s CPU %s, window 400ms, busy sibling on core %s", time.Since(start), cpu, core)
		})
	}
}

func TestCPUBusyChild(t *testing.T) {
	if os.Getenv("CHILDGUARD_CPU_BUSY") != "1" {
		return
	}
	if os.Getenv("CHILDGUARD_CPU_FIRST") == "false" {
		fmt.Fprintln(os.Stdout, "started")
	}
	deadline := time.Now().Add(1800 * time.Millisecond)
	for time.Now().Before(deadline) {
		runtime.KeepAlive(deadline)
	}
	os.Exit(0)
}

func TestCPUSilentSleep(t *testing.T) {
	err := Run(child("silent"), Options{FirstOutput: 400 * time.Millisecond, Stall: 400 * time.Millisecond, Ceiling: 5 * time.Second})
	var guard *Error
	if !errors.As(err, &guard) || guard.Reason != "stalled" || guard.CPUProgress >= minimumCPUProgress || guard.CPUError != "" {
		t.Fatalf("want CPU-stalled sleeper, got %v", err)
	}
	t.Log(guard)
}

// The leader sleeps while its descendant consumes CPU. wait then transfers the
// descendant's CPU into the leader's cutime/cstime: both must count as progress.
func TestCPUGroupAndReapedChildren(t *testing.T) {
	cmd := exec.Command("sh", "-c", `"$1" -test.run=^TestCPUBusyChild$ & wait; sleep 0.1`, "sh", os.Args[0])
	cmd.Env = append(os.Environ(), "CHILDGUARD_CPU_BUSY=1", "CHILDGUARD_CPU_FIRST=true", "GORACE=atexit_sleep_ms=0")
	if err := Run(cmd, Options{FirstOutput: 400 * time.Millisecond, Stall: 400 * time.Millisecond, Ceiling: 10 * time.Second}); err != nil {
		t.Fatal(err)
	}
	// Assert accounting after an actual wait, while the parent is still alive.
	probe := exec.Command("sh", "-c", `"$1" -test.run=^TestCPUBusyChild$ & wait; sleep 5`, "sh", os.Args[0])
	probe.Env = cmd.Env
	probe.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := probe.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = syscall.Kill(-probe.Process.Pid, syscall.SIGKILL); _ = probe.Wait() }()
	deadline := time.Now().Add(4 * time.Second)
	for {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", probe.Process.Pid))
		if err != nil {
			t.Fatal(err)
		}
		fields := strings.Fields(string(data)[strings.LastIndexByte(string(data), ')')+1:])
		reaped, err := strconv.ParseInt(fields[13], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if reaped > 0 {
			cpu, err := processGroupCPU(probe.Process.Pid)
			if err != nil || cpu < time.Duration(reaped)*10*time.Millisecond {
				t.Fatalf("lost reaped CPU: %s %v", cpu, err)
			}
			t.Logf("reaped descendant CPU retained: %s", cpu)
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("descendant not reaped")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestStatCPU(t *testing.T) {
	group, cpu, err := statCPU("123 (name with ) parentheses) S 1 123 1 0 -1 0 0 0 0 0 2 3 5 7")
	if err != nil || group != 123 || cpu != 170*time.Millisecond {
		t.Fatalf("group=%d CPU=%s err=%v", group, cpu, err)
	}
	for _, data := range []string{"", "123 (x) S", "123 (x) S 1 nope 1 0 -1 0 0 0 0 0 2 3 5 7"} {
		if _, _, err := statCPU(data); err == nil {
			t.Fatalf("accepted malformed stat %q", data)
		}
	}
}
