//go:build unix

package dispatch

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A signal to the process building the Swift engine ends the build tool and every compiler it started,
// and the build reports that it was stopped. Before, the signal ended this process alone, and `swift build`
// went on compiling with nobody to read its result (#nqb3mjv).
func TestASignalDuringABuildEndsTheToolAndWhatItStarted(t *testing.T) {
	t.Parallel()
	compilerPID := filepath.Join(t.TempDir(), "compiler.pid")
	tool := exec.Command("/bin/sh", "-c",
		"sleep 300 & echo $! > '"+compilerPID+".partial' && mv '"+compilerPID+".partial' '"+compilerPID+"'; wait")

	signals := make(chan os.Signal, 1)
	result := make(chan error, 1)
	go func() { result <- runBuildToolUntil(tool, signals) }()

	compiler := 0
	for deadline := time.Now().Add(time.Minute); compiler == 0 && time.Now().Before(deadline); {
		if contents, err := os.ReadFile(compilerPID); err == nil {
			compiler, _ = strconv.Atoi(strings.TrimSpace(string(contents)))
		}
		time.Sleep(20 * time.Millisecond)
	}
	if compiler == 0 {
		t.Fatal("the tool never started its compiler")
	}

	signals <- syscall.SIGTERM
	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "stopped by terminated") {
			t.Errorf("a stopped build reported %v", err)
		}
	case <-time.After(buildToolGrace + 30*time.Second):
		t.Fatal("the build did not end after the signal")
	}

	deadline := time.Now().Add(10 * time.Second)
	for syscall.Kill(compiler, 0) == nil && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if syscall.Kill(compiler, 0) == nil {
		_ = syscall.Kill(compiler, syscall.SIGKILL)
		t.Error("the compiler the tool started outlived the signal")
	}
}

// A build that ends on its own reports its own result, untouched by the grouping.
func TestABuildToolThatEndsReportsItsOwnExit(t *testing.T) {
	t.Parallel()
	if err := runBuildToolUntil(exec.Command("/bin/sh", "-c", "exit 0"), make(chan os.Signal)); err != nil {
		t.Errorf("a tool that succeeded reported %v", err)
	}
	if err := runBuildToolUntil(exec.Command("/bin/sh", "-c", "exit 3"), make(chan os.Signal)); err == nil {
		t.Error("a tool that failed reported success")
	}
}
