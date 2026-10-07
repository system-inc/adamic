package oracle

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// The same fixture registry shape, counted by the ordinary counts gate, but run
// below with a longer deadline. Ordinary fixtures retain their one-minute bound.
var slowRegExpFixtures = []struct {
	path            string
	lowers, checked bool
}{{"internal/oracle/testdata/regexp_native_long_backtrack.a", true, false}}

// Not parallel: the long failing search deliberately exercises backtracking;
// sanitizer contention must not turn its leak check into a timeout.
func TestRegExpLongBacktrackNode(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, slowRegExpFixtures[0].path))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node did not finish: %q", want.stderr)
	}
	if got := onJavaScriptBackend(t, program); disagreement(want, got) != "" {
		t.Fatalf("JavaScript backend: %s", disagreement(want, got))
	}
	sanitized := filepath.Join(t.TempDir(), "sanitized")
	if err := native.Build(native.C(program), sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := longRegExpRun(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, sanitized)
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("native: %s Node=%q native=%q stderr=%q", difference, want.stdout, got.stdout, got.stderr)
	}
	release := filepath.Join(t.TempDir(), "release")
	if err := native.Build(native.C(program), release, native.Options{}); err != nil {
		t.Fatal(err)
	}
	if got := longRegExpRun(t, nil, release); disagreement(want, got) != "" {
		t.Fatalf("release build: %s", disagreement(want, got))
	}
	var leaked run
	switch runtime.GOOS {
	case "linux":
		leaked = longRegExpRun(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized)
	case "darwin":
		leaked = longRegExpRun(t, nil, "leaks", "--atExit", "--", release)
	default:
		t.Fatalf("no leak oracle for %s", runtime.GOOS)
	}
	if leaked.exitCode != 0 {
		t.Fatalf("leaks: exit %d\n%s\n%s", leaked.exitCode, leaked.stderr, leaked.stdout)
	}
}

func longRegExpRun(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Env = append(os.Environ(), environment...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	start := time.Now()
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("long regex exceeded 3m: %v", ctx.Err())
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatal(err)
	}
	t.Logf("%s %v: %s, exit %d", filepath.Base(name), environment, time.Since(start), command.ProcessState.ExitCode())
	result := run{stdout.Bytes(), stderr.Bytes(), command.ProcessState.ExitCode()}
	rememberRun(t, result)
	return result
}
