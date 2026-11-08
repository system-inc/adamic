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
	"github.com/system-inc/adamic/internal/toolchain"
)

func guardExecutable(test *testing.T) string {
	test.Helper()
	path, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	return path
}

// The shard adapter delegates its backend to this process so all backends use
// the same CPU guard and independently diagnosed wall backstop as execute.
// Not parallel: ADAMIC_LINT_GUARD_BACKEND is unset in the process-wide environment.
func TestCompilerGuardBackend(test *testing.T) {
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
	if runError := testguard.Run(command, testguard.Budget, testguard.Ceiling); runError != nil {
		fmt.Fprintln(os.Stderr, runError)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestChildCpuHangGuard(test *testing.T) {
	test.Parallel()
	loop := filepath.Join(test.TempDir(), "planted-endless-loop.js")
	if writeError := os.WriteFile(loop, []byte("for (;;) {}"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	command := toolchain.Command("node", loop)
	started := time.Now()
	runError := testguard.Run(command, time.Second, testguard.Ceiling)
	if runError == nil ||
		!strings.Contains(runError.Error(), "child CPU hang guard exceeded") ||
		!strings.Contains(runError.Error(), "planted-endless-loop.js") {
		test.Fatalf("loop guard: %v", runError)
	}
	test.Logf("planted endless loop killed and named in %s: %v", time.Since(started), runError)
}

func TestChildWallBackstop(test *testing.T) {
	test.Parallel()
	command := toolchain.Command("node", "-e", "setTimeout(()=>{},100000)")
	started := time.Now()
	runError := testguard.Run(command, time.Minute, 200*time.Millisecond)
	if runError == nil || !strings.Contains(runError.Error(), "wall backstop exceeded") {
		test.Fatalf("backstop: %v", runError)
	}
	test.Logf("separate backstop in %s: %v", time.Since(started), runError)
}

// A silent child may wait longer than its entire CPU budget and still finish.
// The full compiler comparison additionally exercises real shards under burners.
func TestChildCpuWaitGuard(test *testing.T) {
	test.Parallel()
	command := toolchain.Command(
		"node",
		"-e",
		`const start=process.cpuUsage(); while(Object.values(process.cpuUsage(start)).reduce((a,b)=>a+b,0) < 200000) {} setTimeout(()=>{const again=process.cpuUsage(); while(Object.values(process.cpuUsage(again)).reduce((a,b)=>a+b,0) < 200000) {}},1200);`,
	)
	started := time.Now()
	if runError := testguard.Run(command, time.Second, testguard.Ceiling); runError != nil {
		test.Fatal(runError)
	}
	wall := time.Since(started)
	cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
	if wall <= time.Second || cpu >= time.Second {
		test.Fatalf("silent child: wall %s CPU %s", wall, cpu)
	}
	test.Logf("silent healthy child survived: wall %s CPU %s, budget 1s", wall, cpu)
}
