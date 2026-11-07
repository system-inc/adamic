package lint

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Isolation children deliberately include cold product preparation. The command's context comes with it, for
// childguard.Capture to end the reads by: every caller reads the child's output through Capture, to its end with no
// WaitDelay deciding when, where exec's copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the
// reads hadn't reached EOF a second after the child exited, so an isolation child that had passed could fail under
// load (#n3gk22h's shape). Capture has no stall guard to lose, and the 600 s deadline or the unit's own ceiling,
// whichever comes first, ends a hung child.
func lintBuildPhaseChild(test *testing.T, name string, args ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	test.Cleanup(cancel)
	command := exec.CommandContext(timeoutContext, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	return command, timeoutContext
}

func TestShardsAgree_Setup(test *testing.T) {
	test.Parallel()
	testShardsAgreePrepare(test)
	testShardsAgreeReferencesReady(test)
}

func TestShardsAgree_SetupRequired(test *testing.T) {
	test.Parallel()
	// The selected leaf, without Setup, looks up the prepared corpus itself, inheriting this run's products (a
	// runner's are Workshop's, read and never built; TestProduct_ proves it builds alone).
	command, ctx := lintBuildPhaseChild(
		test,
		os.Args[0],
		"-test.run=^TestShardsAgree_000$",
		"-test.timeout=600s",
		"-test.v",
	)
	command.Env = append(os.Environ(), "ADAMIC_SHARDS_AGREE_LEAF=0")
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	if captureError != nil ||
		!bytes.Contains(output, []byte("--- PASS: TestShardsAgree_000")) ||
		!productLookedUp(output, "shards-agree-prepared-v1") ||
		!productLookedUp(output, "shards-agree-references-v1") {
		test.Fatalf("standalone shard must look up its corpus: %v\n%s", captureError, output)
	}
}

func TestProfileCompilation_Setup(test *testing.T) {
	test.Parallel()
	started := time.Now()
	compilationPrepare(test)
	test.Logf("setup wall=%s", time.Since(started))
}

func TestProfileCompilationBuildC(test *testing.T) {
	test.Parallel()
	started := time.Now()
	defer func() { test.Logf("build wall=%s", time.Since(started)) }()
	compilationEmission(test, compilationInputs(test), "c")
}

func TestProfileCompilationBuildJavaScript(test *testing.T) {
	test.Parallel()
	started := time.Now()
	defer func() { test.Logf("build wall=%s", time.Since(started)) }()
	compilationEmission(test, compilationInputs(test), "javascript")
}

func TestProfileCompilationBuildNative(test *testing.T) {
	test.Parallel()
	started := time.Now()
	defer func() { test.Logf("build wall=%s", time.Since(started)) }()
	compilationProducts(test)
}

func TestProfileCompilationPlantedFailure(test *testing.T) {
	test.Parallel()
	compilationPrepare(test)
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	command, ctx := lintBuildPhaseChild(
		test,
		executable,
		"-test.run=^TestProfileCompilation_000$",
		"-test.v",
		"-test.timeout=600s",
	)
	command.Env = append(os.Environ(), "ADAMIC_PROFILE_COMPILATION_PLANT=1")
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	if command.ProcessState == nil {
		test.Fatalf("planted-failure child did not complete: %v", captureError)
	}
	if captureError == nil ||
		!strings.Contains(string(output), "planted profile disagreement") ||
		strings.Count(string(output), "--- FAIL: TestProfileCompilation_000") != 1 {
		test.Fatalf("wrong planted-failure attribution: %v\n%s", captureError, output)
	}
	test.Log("planted single-case disagreement caught by TestProfileCompilation_000")
}
