package lint

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Isolation children deliberately include cold product preparation.
func lintBuildPhaseChild(t *testing.T, name string, args ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func TestShardsAgree_Setup(t *testing.T) {
	t.Parallel()
	testShardsAgreePrepare(t)
}

func TestShardsAgree_SetupRequired(t *testing.T) {
	t.Parallel()
	// Keep build-phase dependencies warm while forcing the selected leaf to
	// create the prepared corpus itself, without selecting Setup.
	ready := testShardsAgreeReady(t)
	cache := t.TempDir()
	entries, err := os.ReadDir(filepath.Dir(ready.Root))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == filepath.Base(ready.Root) {
			continue
		}
		if err := os.Symlink(filepath.Join(filepath.Dir(ready.Root), entry.Name()), filepath.Join(cache, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
	command := lintBuildPhaseChild(t, os.Args[0], "-test.run=^TestShardsAgree_000$", "-test.timeout=600s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_CACHE=on", "ADAMIC_SHARDS_AGREE_LEAF=0")
	output, err := command.CombinedOutput()
	if err != nil || !bytes.Contains(output, []byte("--- PASS: TestShardsAgree_000")) || !bytes.Contains(output, []byte("build shards-agree-prepared-v1 ")) || !bytes.Contains(output, []byte(" miss ")) {
		t.Fatalf("standalone shard must prepare its corpus: %v\n%s", err, output)
	}
}

func TestProfileCompilation_Setup(t *testing.T) {
	t.Parallel()
	started := time.Now()
	compilationPrepare(t)
	t.Logf("setup wall=%s", time.Since(started))
}

func TestProfileCompilationBuildLower(t *testing.T) {
	t.Parallel()
	started := time.Now()
	defer func() { t.Logf("build wall=%s", time.Since(started)) }()
	compilationLowered(t, compilationInputs(t))
}

func TestProfileCompilationBuildC(t *testing.T) {
	t.Parallel()
	started := time.Now()
	defer func() { t.Logf("build wall=%s", time.Since(started)) }()
	compilationEmission(t, compilationInputs(t), "c")
}

func TestProfileCompilationBuildJavaScript(t *testing.T) {
	t.Parallel()
	started := time.Now()
	defer func() { t.Logf("build wall=%s", time.Since(started)) }()
	compilationEmission(t, compilationInputs(t), "javascript")
}

func TestProfileCompilationBuildNative(t *testing.T) {
	t.Parallel()
	started := time.Now()
	defer func() { t.Logf("build wall=%s", time.Since(started)) }()
	compilationProducts(t)
}

func TestProfileCompilationPlantedFailure(t *testing.T) {
	t.Parallel()
	compilationPrepare(t)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := lintBuildPhaseChild(t, executable, "-test.run=^TestProfileCompilation_000$", "-test.v", "-test.timeout=600s")
	command.Env = append(os.Environ(), "ADAMIC_PROFILE_COMPILATION_PLANT=1")
	output, err := command.CombinedOutput()
	if command.ProcessState == nil {
		t.Fatalf("planted-failure child did not complete: %v", err)
	}
	if err == nil || !strings.Contains(string(output), "planted profile disagreement") || strings.Count(string(output), "--- FAIL: TestProfileCompilation_000") != 1 {
		t.Fatalf("wrong planted-failure attribution: %v\n%s", err, output)
	}
	t.Log("planted single-case disagreement caught by TestProfileCompilation_000")
}
