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
func realpathDirectory(test *testing.T) string {
	test.Helper()
	root := test.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0700); err != nil {
		test.Fatal(err)
	}
	for _, name := range []string{"file", "café-雪"} {
		if err := os.WriteFile(filepath.Join(root, "dir", name), []byte("data"), 0600); err != nil {
			test.Fatal(err)
		}
	}
	// The NUL mutant must resolve the truncated path, rather than fail ENOENT.
	if err := os.WriteFile(filepath.Join(root, "a"), []byte("data"), 0600); err != nil {
		test.Fatal(err)
	}
	for name, target := range map[string]string{
		"link":     "dir",
		"chain":    "link",
		"dangling": "absent-target",
		"loop-a":   "loop-b",
		"loop-b":   "loop-a",
	} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			test.Fatal(err)
		}
	}
	return root
}

func realpathRun(test *testing.T, cwd, name string, arguments ...string) run {
	test.Helper()
	ctx, command := bounded(test, name, arguments...)
	command.Dir = cwd
	command.Env = []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		test.Fatal(err)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

func realpathAgreement(test *testing.T, name string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join("testdata/host_realpath", name))
	if err != nil {
		test.Fatal(err)
	}
	cwd := realpathDirectory(test)
	loader, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	truth := realpathRun(test, cwd, "node", "--disable-warning=ExperimentalWarning", loader, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		test.Fatalf("Node fixture: %+v", truth)
	}
	if name == "nul.a" {
		test.Logf("Node stdout: %s", truth.stdout)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	binary := filepath.Join(test.TempDir(), "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	script := filepath.Join(test.TempDir(), "program.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
		test.Fatal(err)
	}
	for backend, got := range map[string]run{
		"native":     realpathRun(test, cwd, binary),
		"javascript": realpathRun(test, cwd, "node", "--disable-warning=ExperimentalWarning", loader, script),
	} {
		if name == "nul.a" {
			test.Logf("%s stdout: %s", backend, got.stdout)
		}
		if difference := disagreement(truth, got); difference != "" {
			test.Fatalf(
				"%s: %s; Node stdout %q stderr %q exit %d; got stdout %q stderr %q exit %d",
				backend,
				difference,
				truth.stdout,
				truth.stderr,
				truth.exitCode,
				got.stdout,
				got.stderr,
				got.exitCode,
			)
		}
	}
	if name == "errors.a" && !bytes.Contains(truth.stdout, []byte("ENOTDIR: not a directory, realpath 'dir/file/'")) {
		test.Fatal("Node trailing-slash witness changed")
	}
}
func TestHostRealpathPathsMatchNode(test *testing.T) {
	test.Parallel()
	realpathAgreement(test, "paths.a")
}
func TestHostRealpathErrorsMatchNode(test *testing.T) {
	test.Parallel()
	realpathAgreement(test, "errors.a")
}
func TestHostRealpathClosureMatchesNode(test *testing.T) {
	test.Parallel()
	realpathAgreement(test, "closure.a")
}

func TestHostRealpathNULMatchesNode(test *testing.T) {
	test.Parallel()
	realpathAgreement(test, "nul.a")
}

func TestHostRealpathWalkErrorsMatchNode(test *testing.T) {
	test.Parallel()
	realpathAgreement(test, "walk-errors.a")
}

func TestHostRealpathWalkPathsMatchNode(test *testing.T) {
	test.Parallel()
	realpathAgreement(test, "walk-paths.a")
}
