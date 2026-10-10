package lint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/testguard"
)

func guardExecutable(t *testing.T) string {
	t.Helper()
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// The shard adapter delegates its backend to this process so all backends use
// the same CPU guard and independently diagnosed wall backstop as execute.
// Not parallel: ADAMIC_LINT_GUARD_BACKEND is unset in the process-wide environment.
func TestCompilerGuardBackend(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_GUARD_BACKEND") != "1" {
		return
	}
	index := 0
	for i, arg := range os.Args {
		if arg == "--" {
			index = i + 1
			break
		}
	}
	if index == 0 || index >= len(os.Args) {
		os.Exit(125)
	}
	os.Unsetenv("ADAMIC_LINT_GUARD_BACKEND")
	command := exec.Command(os.Args[index], os.Args[index+1:]...)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := testguard.Run(command, testguard.Budget, testguard.Ceiling); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestChildCPUHangGuard(t *testing.T) {
	t.Parallel()
	loop := filepath.Join(t.TempDir(), "planted-endless-loop.js")
	if err := os.WriteFile(loop, []byte("for (;;) {}"), 0644); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", loop)
	started := time.Now()
	err := testguard.Run(command, time.Second, testguard.Ceiling)
	if err == nil || !strings.Contains(err.Error(), "child CPU hang guard exceeded") || !strings.Contains(err.Error(), "planted-endless-loop.js") {
		t.Fatalf("loop guard: %v", err)
	}
	t.Logf("planted endless loop killed and named in %s: %v", time.Since(started), err)
}

func TestChildWallBackstop(t *testing.T) {
	t.Parallel()
	command := exec.Command("node", "-e", "setTimeout(()=>{},100000)")
	started := time.Now()
	err := testguard.Run(command, time.Minute, 200*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "wall backstop exceeded") {
		t.Fatalf("backstop: %v", err)
	}
	t.Logf("separate backstop in %s: %v", time.Since(started), err)
}

// A silent child may wait longer than its entire CPU budget and still finish.
// The full compiler comparison additionally exercises real shards under burners.
func TestChildCPUWaitGuard(t *testing.T) {
	t.Parallel()
	command := exec.Command("node", "-e", `const start=process.cpuUsage(); while(Object.values(process.cpuUsage(start)).reduce((a,b)=>a+b,0) < 200000) {} setTimeout(()=>{const again=process.cpuUsage(); while(Object.values(process.cpuUsage(again)).reduce((a,b)=>a+b,0) < 200000) {}},1200);`)
	started := time.Now()
	if err := testguard.Run(command, time.Second, testguard.Ceiling); err != nil {
		t.Fatal(err)
	}
	wall := time.Since(started)
	cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
	if wall <= time.Second || cpu >= time.Second {
		t.Fatalf("silent child: wall %s CPU %s", wall, cpu)
	}
	t.Logf("silent healthy child survived: wall %s CPU %s, budget 1s", wall, cpu)
}
