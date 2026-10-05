package oracle

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// inputFixtures are the programs that read from outside, through the doors 0.2 opens: their arguments
// and files. Each runs the three ways every fixture does, from its own directory, so a relative path
// names the same file on every run, and with the same arguments after the program.
var inputFixtures = []struct {
	path      string
	arguments []string

	// unreadable adds, as the last argument, the path of a file the program may not read, made here
	// since git can't keep a file's permissions.
	unreadable bool
}{
	{"internal/oracle/testdata/read_files.a", nil, false},
	{"internal/oracle/testdata/utf8_sweep.a", nil, false},
	{"internal/oracle/testdata/arguments.a", []string{
		"plain", "", "with space", "héllo 🌍", "--flag=1",
		// Invalid UTF-8, decoded as Node decodes argv: each bad sequence one U+FFFD.
		"a\xffb", "\xe2\x82", "\xc0\x80", "\xed\xa0\x80", "\xf4\x90\x80\x80", "\xef\xbb\xbfmarked", "end \xf0\x9f\x8c",
	}, false},
	{"internal/oracle/testdata/read_arguments.a", []string{"reading/hello.txt", "reading/missing.txt", "reading"}, true},
}

// inputRun is where and as whom one input fixture runs.
type inputRun struct {
	directory string
	arguments []string

	// credential, when not nil, is the user every run is made as: running as root, the file made
	// unreadable would be read anyway, so the runs drop to nobody.
	credential *syscall.Credential
}

// executeInput runs a command as an input fixture runs, with environment added to the test's own.
func executeInput(t *testing.T, how inputRun, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := bounded(t, name, arguments...)
	command.Dir = how.directory
	command.SysProcAttr.Credential = how.credential
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNodeWith runs a program on Node through the oracle's runner, with the fixture's arguments.
func onNodeWith(t *testing.T, how inputRun, path string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return executeInput(t, how, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, how.arguments...)...)
}

// sharedDirectory is a directory for what the runs need, which a run dropped to nobody can still
// reach: t.TempDir's are root's alone.
func sharedDirectory(t *testing.T) string {
	t.Helper()
	directory, err := os.MkdirTemp("", "adamic-input-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(directory) })
	if err := os.Chmod(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	return directory
}

func TestInputAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range inputFixtures {
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			shared := sharedDirectory(t)
			how := inputRun{directory: filepath.Dir(path), arguments: fixture.arguments}
			if os.Geteuid() == 0 {
				how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
			}
			if fixture.unreadable {
				unreadable := filepath.Join(shared, "unreadable.txt")
				if err := os.WriteFile(unreadable, []byte("secret\n"), 0o000); err != nil {
					t.Fatal(err)
				}
				how.arguments = append(append([]string{}, how.arguments...), unreadable)
			}
			oracle := onNodeWith(t, how, path)
			backend := inputBackend(t, how, program, shared)
			native, binary := inputNatively(t, how, program, shared)
			if difference := disagreement(oracle, native); difference != "" {
				t.Errorf("%s\nnode:   exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q",
					difference, oracle.exitCode, oracle.stdout, oracle.stderr, native.exitCode, native.stdout, native.stderr)
			}
			if difference := disagreement(oracle, backend); difference != "" {
				t.Errorf("JavaScript backend: %s\nnode:    exit %d, stdout %q, stderr %q\nbackend: exit %d, stdout %q, stderr %q",
					difference, oracle.exitCode, oracle.stdout, oracle.stderr, backend.exitCode, backend.stdout, backend.stderr)
			}
			if fixture.unreadable && !bytes.Contains(oracle.stdout, []byte(": permission denied\n")) {
				// Otherwise the file was read after all, and nothing here held the refusal to Node.
				t.Errorf("want Node refused the unreadable file, got stdout %q", oracle.stdout)
			}
			if oracle.exitCode != 0 {
				t.Errorf("want every input fixture to finish on Node, got exit %d, stderr %q", oracle.exitCode, oracle.stderr)
				return
			}
			if leaked := inputLeaks(t, how, program, binary); leaked != "" {
				t.Errorf("leaks:\n%s", leaked)
			}
		})
	}
}

// inputBackend runs a lowered program through the JavaScript backend, on Node.
func inputBackend(t *testing.T, how inputRun, program *ir.Program, shared string) run {
	t.Helper()
	path := filepath.Join(shared, "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return onNodeWith(t, how, path)
}

// inputNatively builds a lowered program under the sanitizers and runs it, as natively does.
func inputNatively(t *testing.T, how inputRun, program *ir.Program, shared string) (run, string) {
	t.Helper()
	binary := filepath.Join(shared, "program")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return executeInput(t, how, environment, binary, how.arguments...), binary
}

// inputLeaks is leaks for an input fixture: the same check, run where and as whom the fixture runs.
func inputLeaks(t *testing.T, how inputRun, program *ir.Program, sanitized string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(sharedDirectory(t), "program")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		report := executeInput(t, how, nil, "leaks", append([]string{"--atExit", "--", binary}, how.arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := executeInput(t, how, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, how.arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}
