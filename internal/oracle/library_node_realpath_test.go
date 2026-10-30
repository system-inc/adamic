package oracle

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Both source and backends observe exactly this filesystem, cwd and environment.
func realpathDirectory(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"file", "café-雪"} {
		if err := os.WriteFile(filepath.Join(root, "dir", name), []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	// The NUL mutant must resolve the truncated path, rather than fail ENOENT.
	if err := os.WriteFile(filepath.Join(root, "a"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	for name, target := range map[string]string{"link": "dir", "chain": "link", "dangling": "absent-target", "loop-a": "loop-b", "loop-b": "loop-a"} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func realpathRun(t *testing.T, cwd, name string, arguments ...string) run {
	t.Helper()
	ctx, command := bounded(t, name, arguments...)
	command.Dir = cwd
	command.Env = []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		t.Fatal(err)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

func realpathAgreement(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata/host_realpath", name))
	if err != nil {
		t.Fatal(err)
	}
	cwd := realpathDirectory(t)
	loader, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	truth := realpathRun(t, cwd, "node", "--disable-warning=ExperimentalWarning", loader, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node fixture: %+v", truth)
	}
	if name == "nul.a" {
		t.Logf("Node stdout: %s", truth.stdout)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	for backend, got := range map[string]run{"native": realpathRun(t, cwd, binary), "javascript": realpathRun(t, cwd, "node", "--disable-warning=ExperimentalWarning", loader, script)} {
		if name == "nul.a" {
			t.Logf("%s stdout: %s", backend, got.stdout)
		}
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s; Node stdout %q stderr %q exit %d; got stdout %q stderr %q exit %d", backend, difference, truth.stdout, truth.stderr, truth.exitCode, got.stdout, got.stderr, got.exitCode)
		}
	}
	if name == "errors.a" && !bytes.Contains(truth.stdout, []byte("ENOTDIR: not a directory, realpath 'dir/file/'")) {
		t.Fatal("Node trailing-slash witness changed")
	}
}
func TestHostRealpathPathsMatchNode(t *testing.T)  { t.Parallel(); realpathAgreement(t, "paths.a") }
func TestHostRealpathErrorsMatchNode(t *testing.T) { t.Parallel(); realpathAgreement(t, "errors.a") }
func TestHostRealpathClosureMatchesNode(t *testing.T) {
	t.Parallel()
	realpathAgreement(t, "closure.a")
}

func TestHostRealpathNULMatchesNode(t *testing.T) {
	t.Parallel()
	realpathAgreement(t, "nul.a")
}

func TestHostRealpathWalkErrorsMatchNode(t *testing.T) {
	t.Parallel()
	realpathAgreement(t, "walk-errors.a")
}

func TestHostRealpathWalkPathsMatchNode(t *testing.T) {
	t.Parallel()
	realpathAgreement(t, "walk-paths.a")
}
