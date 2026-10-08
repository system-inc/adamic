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
	"github.com/system-inc/adamic/internal/leakcheck"
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

	// writes gives the program, as its first argument, an empty directory of its own on every run, with
	// a directory in it named locked that it may not write into, and unlisted and closed, which it may
	// not list and may not enter. What each run leaves there, every
	// file's name, bytes and permissions, must agree too.
	writes bool
}{
	{"internal/oracle/testdata/node_process_host.a", []string{"plain", "", "with space", "héllo 🌍", "--prof", "bad\xff"}, false, false},
	{"internal/oracle/testdata/node_process_performance.a", nil, false, false},
	{"internal/oracle/testdata/node_process_performance_core.a", nil, false, false},
	{"internal/oracle/testdata/node_process_system.a", []string{"--noEmit", "tiny.a"}, false, false},
	{"internal/oracle/testdata/read_files.a", nil, false, false},
	{"internal/oracle/testdata/realpath.a", nil, false, false},
	{"internal/oracle/testdata/empty-path.a", nil, false, false},
	{"internal/oracle/testdata/utf8_sweep.a", nil, false, false},
	{"internal/oracle/testdata/arguments.a", []string{
		"plain", "", "with space", "héllo 🌍", "--flag=1",
		// Invalid UTF-8, decoded as Node decodes argv: each bad sequence one U+FFFD.
		"a\xffb", "\xe2\x82", "\xc0\x80", "\xed\xa0\x80", "\xf4\x90\x80\x80", "\xef\xbb\xbfmarked", "end \xf0\x9f\x8c",
	}, false, false},
	{"internal/oracle/testdata/read_arguments.a", []string{"reading/hello.txt", "reading/missing.txt", "reading"}, true, false},
	{"internal/oracle/testdata/write_files.a", nil, false, true},
	{"internal/oracle/testdata/walk.a", nil, false, true},
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
	// Set the child's umask explicitly. With 022, 0644 and 0666 create modes look identical.
	command := bounded(t, "/bin/sh", append([]string{"-c", `umask 0; exec "$0" "$@"`, name}, arguments...)...)
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
	result := run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
	rememberRun(t, result)
	return result
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
			cacheProbe(t, fixture.path, nil, inputIdentity(t), func() {

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
				// Each run gets its own directory to write in, when the fixture writes, so what each leaves
				// there can be compared.
				runs := 0
				prepared := func() inputRun {
					if !fixture.writes {
						return how
					}
					runs++
					written := writable(t, shared, fmt.Sprintf("run%d", runs))
					return inputRun{directory: how.directory, arguments: append([]string{written}, how.arguments...), credential: how.credential}
				}
				collect := func(given inputRun) map[string]string {
					if !fixture.writes {
						return nil
					}
					return snapshot(t, given.arguments[0])
				}
				nodeRun := prepared()
				oracle := onNodeWith(t, nodeRun, path)
				backendRun := prepared()
				backend := inputBackend(t, backendRun, program, shared)
				nativeRun := prepared()
				native, binary := inputNatively(t, nativeRun, program, shared)
				if fixture.writes {
					nodeLeft, backendLeft, nativeLeft := collect(nodeRun), collect(backendRun), collect(nativeRun)
					if difference := filesDiffer(nodeLeft, nativeLeft); difference != "" {
						t.Errorf("the files native wrote differ from Node's: %s", difference)
					}
					if difference := filesDiffer(nodeLeft, backendLeft); difference != "" {
						t.Errorf("the files the JavaScript backend wrote differ from Node's: %s", difference)
					}
					if !bytes.Contains(oracle.stdout, []byte(": permission denied\n")) {
						t.Errorf("want Node refused to write into the locked directory, got stdout %q", oracle.stdout)
					}
				}
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
				if leaked := inputLeaks(t, prepared, program, binary); leaked != "" {
					t.Errorf("leaks:\n%s", leaked)
				}

			})
		})
	}
}

// writable makes an empty directory a run of a writing fixture may write in, whoever it runs as, with
// one directory in it, locked, that it may not, two it may not list or enter, and a file whose name
// isn't valid UTF-8.
func writable(t *testing.T, shared string, name string) string {
	t.Helper()
	directory := filepath.Join(shared, name)
	if err := os.Mkdir(directory, 0o777); err != nil {
		t.Fatal(err)
	}
	// Mkdir's mode passes through the umask; this one is meant whole.
	if err := os.Chmod(directory, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(directory, "locked"), 0o555); err != nil {
		t.Fatal(err)
	}
	// A file whose name isn't valid UTF-8, which Node lists decoded, the bad byte as U+FFFD. Linux
	// keeps any bytes; macOS's file system refuses such a name, so there the file is named what Node
	// reads it back as, which lists, sorts among the rest and counts the same, and only the decoding
	// goes unasked.
	if err := os.WriteFile(filepath.Join(directory, "bad\xff name"), nil, 0o644); err != nil {
		if err := os.WriteFile(filepath.Join(directory, "bad\uFFFD name"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// And two a walk can't go into: unlisted can be passed through but not listed, and closed not
	// even passed through. Mkdir's mode passes through the umask, so each is set whole.
	for name, mode := range map[string]os.FileMode{"unlisted": 0o311, "closed": 0o000} {
		path := filepath.Join(directory, name)
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

// snapshot is what a directory holds, every file and directory under it by its path there, with its
// permissions and, for a file, its bytes.
func snapshot(t *testing.T, directory string) map[string]string {
	t.Helper()
	found := map[string]string{}
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			// A directory that can't be listed (unlisted and closed, when the test isn't root) is
			// recorded by its permissions, already, and not gone into.
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return err
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		information, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			found[relative] = fmt.Sprintf("directory %v", information.Mode().Perm())
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		found[relative] = fmt.Sprintf("file %v %q", information.Mode().Perm(), contents)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := probeRuns.Load(t.Name()); ok {
		var files []recordedFile
		for path, contents := range found {
			files = append(files, recordedFile{[]byte(path), []byte(contents)})
		}
		value.(*probeEvidence).Files = append(value.(*probeEvidence).Files, files)
	}
	return found
}

// filesDiffer says how two snapshots differ, or "" when they don't.
func filesDiffer(want map[string]string, got map[string]string) string {
	for name, held := range want {
		if got[name] != held {
			return fmt.Sprintf("%s: want %.200s, got %.200s", name, held, got[name])
		}
	}
	for name, held := range got {
		if _, isWanted := want[name]; !isWanted {
			return fmt.Sprintf("%s: not wanted, got %.200s", name, held)
		}
	}
	return ""
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
// prepared gives each run its own place to write.
func inputLeaks(t *testing.T, prepared func() inputRun, program *ir.Program, sanitized string) string {
	leak := inputLeaksUncached(t, prepared, program, sanitized)
	rememberLeak(t, leak)
	return leak
}
func inputLeaksUncached(t *testing.T, prepared func() inputRun, program *ir.Program, sanitized string) string {
	t.Helper()
	// Each run gets its own place to write, and runs where and as whom the fixture runs.
	var how inputRun
	report, err := leakcheck.Check(leakcheck.Program{
		C:         native.C(program),
		Sanitized: sanitized,
		Counted:   filepath.Join(sharedDirectory(t), "counted"),
		Arguments: func() []string {
			how = prepared()
			return how.arguments
		},
		Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
			return leakRun(executeInput(t, how, environment, name, arguments...))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return report
}
