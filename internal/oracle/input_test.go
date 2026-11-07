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

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
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

	// plainNames leaves out of that directory the file whose name isn't valid UTF-8, for a fixture that
	// lists the directory for something else: wasmtime can't list a directory holding such a name
	// (fd_readdir answers EILSEQ), which walk.a pins as its host limit and the directory-entries
	// fixtures keep asking on every engine.
	plainNames bool
}{
	{"internal/oracle/testdata/read_files.a", nil, false, false, false},
	{"internal/oracle/testdata/empty-path.a", nil, false, false, false},
	{"internal/oracle/testdata/utf8_sweep.a", nil, false, false, false},
	{"internal/oracle/testdata/arguments.a", []string{
		"plain", "", "with space", "héllo 🌍", "--flag=1",
		// Invalid UTF-8, decoded as Node decodes argv: each bad sequence one U+FFFD.
		"a\xffb", "\xe2\x82", "\xc0\x80", "\xed\xa0\x80", "\xf4\x90\x80\x80", "\xef\xbb\xbfmarked", "end \xf0\x9f\x8c",
	}, false, false, false},
	{
		"internal/oracle/testdata/read_arguments.a",
		[]string{"reading/hello.txt", "reading/missing.txt", "reading"},
		true,
		false,
		false,
	},
	{"internal/oracle/testdata/write_files.a", nil, false, true, false},
	{"internal/oracle/testdata/walk.a", nil, false, true, false},
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
func executeInput(test *testing.T, how inputRun, environment []string, name string, arguments ...string) run {
	test.Helper()
	// Set the child's umask explicitly so 0644 and 0666 create modes differ.
	path, pathError := toolchain.Path(name)
	if pathError != nil {
		test.Fatal(pathError)
	}
	ctx, command := bounded(test, "/bin/sh", append([]string{"-c", `umask 0; exec "$0" "$@"`, path}, arguments...)...)
	command.Dir = how.directory
	command.SysProcAttr.Credential = how.credential
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	result := run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
	rememberRun(test, result)
	return result
}

// onNodeWith runs a program on Node through the oracle's runner, with the fixture's arguments.
func onNodeWith(test *testing.T, how inputRun, path string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return executeInput(
		test,
		how,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, how.arguments...)...,
	)
}

// sharedDirectory is a directory for what the runs need, which a run dropped to nobody can still
// reach: test.TempDir's are root's alone.
func sharedDirectory(test *testing.T) string {
	test.Helper()
	directory, temporaryDirectoryError := os.MkdirTemp("", "adamic-input-")
	if temporaryDirectoryError != nil {
		test.Fatal(temporaryDirectoryError)
	}
	test.Cleanup(func() {
		if removeError := os.RemoveAll(directory); removeError != nil {
			test.Errorf("the shared directory stays behind: %v", removeError)
		}
	})
	if changeModeError := os.Chmod(directory, 0o755); changeModeError != nil {
		test.Fatal(changeModeError)
	}
	return directory
}

func TestInputAgreesWithNode(test *testing.T) {
	test.Parallel()
	for _, fixture := range inputFixtures {
		test.Run(fixture.path, func(test *testing.T) {
			test.Parallel()
			cacheProbe(test, fixture.path, nil, inputIdentity(test), func() {

				path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture.path))
				if absolutePathError != nil {
					test.Fatal(absolutePathError)
				}
				program, loweredError := lowered(test, path)
				if loweredError != nil {
					test.Fatalf("Lower: %v", loweredError)
				}
				shared := sharedDirectory(test)
				how := inputRun{directory: filepath.Dir(path), arguments: fixture.arguments}
				if os.Geteuid() == 0 {
					how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
				}
				if fixture.unreadable {
					unreadable := filepath.Join(shared, "unreadable.txt")
					if writeError := os.WriteFile(unreadable, []byte("secret\n"), 0o000); writeError != nil {
						test.Fatal(writeError)
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
					written := writableDirectory(test, shared, fmt.Sprintf("run%d", runs), fixture.plainNames)
					return inputRun{
						directory:  how.directory,
						arguments:  append([]string{written}, how.arguments...),
						credential: how.credential,
					}
				}
				collect := func(given inputRun) map[string]string {
					if !fixture.writes {
						return nil
					}
					return snapshot(test, given.arguments[0])
				}
				nodeRun := prepared()
				oracle := onNodeWith(test, nodeRun, path)
				backendRun := prepared()
				backend := inputBackend(test, backendRun, program, shared)
				nativeRun := prepared()
				native, binary := inputNatively(test, nativeRun, program, shared)
				if fixture.writes {
					nodeLeft, backendLeft, nativeLeft := collect(nodeRun), collect(backendRun), collect(nativeRun)
					if difference := filesDiffer(nodeLeft, nativeLeft); difference != "" {
						test.Errorf("the files native wrote differ from Node's: %s", difference)
					}
					if difference := filesDiffer(nodeLeft, backendLeft); difference != "" {
						test.Errorf("the files the JavaScript backend wrote differ from Node's: %s", difference)
					}
					if !bytes.Contains(oracle.stdout, []byte(": permission denied\n")) {
						test.Errorf(
							"want Node refused to write into the locked directory, got stdout %q",
							oracle.stdout,
						)
					}
				}
				if difference := disagreement(oracle, native); difference != "" {
					test.Errorf(
						"%s\nnode:   exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q",
						difference,
						oracle.exitCode,
						oracle.stdout,
						oracle.stderr,
						native.exitCode,
						native.stdout,
						native.stderr,
					)
				}
				if difference := disagreement(oracle, backend); difference != "" {
					test.Errorf(
						"JavaScript backend: %s\nnode:    exit %d, stdout %q, stderr %q\nbackend: exit %d, stdout %q, stderr %q",
						difference,
						oracle.exitCode,
						oracle.stdout,
						oracle.stderr,
						backend.exitCode,
						backend.stdout,
						backend.stderr,
					)
				}
				if fixture.unreadable && !bytes.Contains(oracle.stdout, []byte(": permission denied\n")) {
					// Otherwise the file was read after all, and nothing here held the refusal to Node.
					test.Errorf("want Node refused the unreadable file, got stdout %q", oracle.stdout)
				}
				if oracle.exitCode != 0 {
					test.Errorf(
						"want every input fixture to finish on Node, got exit %d, stderr %q",
						oracle.exitCode,
						oracle.stderr,
					)
					return
				}
				if leaked := inputLeaks(test, prepared, program, binary); leaked != "" {
					test.Errorf("leaks:\n%s", leaked)
				}

			})
		})
	}
}

// writable makes an empty directory a run of a writing fixture may write in, whoever it runs as, with
// one directory in it, locked, that it may not, two it may not list or enter, and a file whose name
// isn't valid UTF-8.
func writable(test *testing.T, shared string, name string) string {
	test.Helper()
	return writableDirectory(test, shared, name, false)
}

// writableDirectory is writable's directory, without the file whose name isn't valid UTF-8 when
// plainNames is set (inputFixtures' plainNames).
func writableDirectory(test *testing.T, shared string, name string, plainNames bool) string {
	test.Helper()
	directory := filepath.Join(shared, name)
	if createDirectoryError := os.Mkdir(directory, 0o777); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	// Mkdir's mode passes through the umask; this one is meant whole.
	if changeModeError := os.Chmod(directory, 0o777); changeModeError != nil {
		test.Fatal(changeModeError)
	}
	if createDirectoryError := os.Mkdir(filepath.Join(directory, "locked"), 0o555); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	// A file whose name isn't valid UTF-8, which Node lists decoded, the bad byte as U+FFFD. Linux
	// keeps any bytes; macOS's file system refuses such a name, so there the file is named what Node
	// reads it back as, which lists, sorts among the rest and counts the same, and only the decoding
	// goes unasked.
	if !plainNames {
		if writeError := os.WriteFile(filepath.Join(directory, "bad\xff name"), nil, 0o644); writeError != nil {
			if writeError := os.WriteFile(filepath.Join(directory, "bad\uFFFD name"), nil, 0o644); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	// And two a walk can't go into: unlisted can be passed through but not listed, and closed not
	// even passed through. Mkdir's mode passes through the umask, so each is set whole.
	for name, mode := range map[string]os.FileMode{"unlisted": 0o311, "closed": 0o000} {
		path := filepath.Join(directory, name)
		if createDirectoryError := os.Mkdir(path, 0o700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if changeModeError := os.Chmod(path, mode); changeModeError != nil {
			test.Fatal(changeModeError)
		}
	}
	return directory
}

// snapshot is what a directory holds, every file and directory under it by its path there, with its
// permissions and, for a file, its bytes.
func snapshot(test *testing.T, directory string) map[string]string {
	test.Helper()
	found := map[string]string{}
	walkError := filepath.WalkDir(directory, func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			// A directory that can't be listed (unlisted and closed, when the test isn't root) is
			// recorded by its permissions, already, and not gone into.
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return pathError
		}
		relative, relativePathError := filepath.Rel(directory, path)
		if relativePathError != nil {
			return relativePathError
		}
		information, infoError := entry.Info()
		if infoError != nil {
			return infoError
		}
		if entry.IsDir() {
			found[relative] = fmt.Sprintf("directory %v", information.Mode().Perm())
			return nil
		}
		contents, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		found[relative] = fmt.Sprintf("file %v %q", information.Mode().Perm(), contents)
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	if value, ok := probeRuns.Load(test.Name()); ok {
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
func inputBackend(test *testing.T, how inputRun, program *ir.Program, shared string) run {
	test.Helper()
	path := filepath.Join(shared, "program.mjs")
	if writeError := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	return onNodeWith(test, how, path)
}

// inputNatively builds a lowered program under the sanitizers and runs it, as natively does.
func inputNatively(test *testing.T, how inputRun, program *ir.Program, shared string) (run, string) {
	test.Helper()
	binary := filepath.Join(shared, "program")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return executeInput(test, how, environment, binary, how.arguments...), binary
}

// inputLeaks is leaks for an input fixture: the same check, run where and as whom the fixture runs.
// prepared gives each run its own place to write.
func inputLeaks(test *testing.T, prepared func() inputRun, program *ir.Program, sanitized string) string {
	leak := inputLeaksUncached(test, prepared, program, sanitized)
	rememberLeak(test, leak)
	return leak
}
func inputLeaksUncached(test *testing.T, prepared func() inputRun, program *ir.Program, sanitized string) string {
	test.Helper()
	// Each run gets its own place to write, and runs where and as whom the fixture runs.
	var how inputRun
	report, checkError := leakcheck.Check(leakcheck.Program{
		C:         native.C(program),
		Sanitized: sanitized,
		Counted:   filepath.Join(sharedDirectory(test), "counted"),
		Arguments: func() []string {
			how = prepared()
			return how.arguments
		},
		Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
			return leakRun(executeInput(test, how, environment, name, arguments...))
		},
	})
	if checkError != nil {
		test.Fatal(checkError)
	}
	return report
}
