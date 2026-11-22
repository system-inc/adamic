package countproducts

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The run helpers here are internal/oracle's, as its counts measured with them before they moved (counts_test.go,
// oracle_test.go, input_test.go, parallel_test.go, node_fs_file_test.go and the environment and realpath harnesses),
// without a test: each returns its failure, and what it makes it removes. Oracle keeps its own for its other tests,
// which remember every run for its probe cache (rememberRun) and fail their test where these return an error.

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// pinnedStack is a command run with an 8 MiB stack, whatever the stack of whoever runs the measure: a fixture that
// recurses until the stack runs out is counted at the depth it reached, which moves with the stack's size (ulimit -s),
// and the table has to be the same on every machine. The shell's ulimit is POSIX's, on Linux and macOS alike; where the
// hard limit won't allow 8 MiB, it fails out loud.
func pinnedStack(name string, arguments ...string) (string, []string) {
	return "/bin/sh", append([]string{"-c", `ulimit -s 8192 && exec "$0" "$@"`, name}, arguments...)
}

// bounded is a command that can't outlive its run: it has a deadline, it runs in a process group of its own, and when
// the deadline passes the whole group is killed, so a fixture that loops, or a child left with nowhere to write, is
// stopped instead of orphaned. A bare name runs the program toolchain.Path finds; one Path can't find fails as it starts,
// with Path's error. It returns the deadline's context and its cancel with the command: its caller runs it with
// childguard.Capture under the same deadline (captured), so the output is read to its end rather than cut off a fixed
// time after the exit, and cancels the deadline once it's done.
func bounded(name string, arguments ...string) (context.Context, context.CancelFunc, *exec.Cmd) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	path, err := toolchain.Path(name)
	if err != nil {
		path = name
	}
	command := exec.CommandContext(ctx, path, arguments...)
	if err != nil {
		command.Err = err
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	return ctx, cancel, command
}

// captured runs command, named name, to its end under ctx and returns what it observed. An exit status is the run's;
// any other failure is captured's.
func captured(ctx context.Context, name string, command *exec.Cmd) (run, error) {
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, fmt.Errorf("running %s: %w", name, err)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, nil
}

// RunEnvironment is the whole environment every Plain and Input row runs with, and a ParallelMap row's before its pool
// and sanitizer settings: what a counted program reads from its environment decides its counts
// (node_process_os_namespace.a reads ADAMIC_PROCESS_NAMESPACE_MISSING, the runtime ADAMIC_THREADS at parallel.c's start),
// and a shard's key names no box's variables, so a count reused on every box and across landings is measured with none
// of them (adamic-loom's must on #a4frknq). It's a declared file's, so changing it moves every shard's key with Harness.
var RunEnvironment = []string{"PATH=/usr/bin:/bin"}

// execute runs a command with RunEnvironment as its whole environment.
func execute(name string, arguments ...string) (run, error) {
	ctx, cancel, command := bounded(name, arguments...)
	defer cancel()
	command.Env = RunEnvironment
	return captured(ctx, name, command)
}

// inputRun is where and as whom one input fixture runs.
type inputRun struct {
	directory string
	arguments []string

	// credential, when not nil, is the user the run is made as: running as root, the file made unreadable would be read
	// anyway, so the run drops to nobody.
	credential *syscall.Credential
}

// executeInput runs a command as an input fixture runs, with RunEnvironment as its whole environment, its umask 0, so
// 0644 and 0666 create modes differ.
func executeInput(how inputRun, name string, arguments ...string) (run, error) {
	path, err := toolchain.Path(name)
	if err != nil {
		return run{}, err
	}
	ctx, cancel, command := bounded("/bin/sh", append([]string{"-c", `umask 0; exec "$0" "$@"`, path}, arguments...)...)
	defer cancel()
	command.Dir = how.directory
	command.Env = RunEnvironment
	command.SysProcAttr.Credential = how.credential
	return captured(ctx, name, command)
}

// executeParallel runs a program that calls the pool with threads workers, RunEnvironment and the sanitizers' settings
// oracle's parallel runs give, its leak check off: a counted run is measured, never checked for leaks. Nothing of the
// process's own environment reaches it.
func executeParallel(threads string, name string, arguments ...string) (run, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Env = slices.Clone(RunEnvironment)
	if threads != "" {
		command.Env = append(command.Env, "ADAMIC_THREADS="+threads)
	}
	command.Env = append(command.Env, "TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1", "ADAMIC_TSAN_PERTURB=1")
	if runtime.GOOS == "linux" {
		command.Env = append(command.Env, "ASAN_OPTIONS=detect_leaks=0")
	}
	stdout, stderr, err := childguard.Capture(ctx, command)
	var failure *exec.ExitError
	if err != nil && !errors.As(err, &failure) {
		return run{}, fmt.Errorf("%s: %w", name, err)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, nil
}

// sharedDirectory is a directory for what an input run needs, which a run dropped to nobody can still reach: a
// temporary directory of its own is its maker's alone. Its caller removes it.
func sharedDirectory() (string, error) {
	directory, err := os.MkdirTemp("", "adamic-input-")
	if err != nil {
		return "", err
	}
	if err := os.Chmod(directory, 0o755); err != nil {
		os.RemoveAll(directory)
		return "", err
	}
	return directory, nil
}

// writable makes an empty directory a run of a writing fixture may write in, whoever it runs as, with one directory in
// it, locked, that it may not, two it may not list or enter, and a file whose name isn't valid UTF-8.
func writable(shared string, name string) (string, error) {
	directory := filepath.Join(shared, name)
	if err := os.Mkdir(directory, 0o777); err != nil {
		return "", err
	}
	// Mkdir's mode passes through the umask; this one is meant whole.
	if err := os.Chmod(directory, 0o777); err != nil {
		return "", err
	}
	if err := os.Mkdir(filepath.Join(directory, "locked"), 0o555); err != nil {
		return "", err
	}
	// A file whose name isn't valid UTF-8, which Node lists decoded, the bad byte as U+FFFD. Linux keeps any bytes;
	// macOS's file system refuses such a name, so there the file is named what Node reads it back as, which lists, sorts
	// among the rest and counts the same, and only the decoding goes unasked.
	if err := os.WriteFile(filepath.Join(directory, "bad\xff name"), nil, 0o644); err != nil {
		if err := os.WriteFile(filepath.Join(directory, "bad� name"), nil, 0o644); err != nil {
			return "", err
		}
	}
	// And two a walk can't go into: unlisted can be passed through but not listed, and closed not even passed through.
	// Mkdir's mode passes through the umask, so each is set whole.
	for entry, mode := range map[string]os.FileMode{"unlisted": 0o311, "closed": 0o000} {
		path := filepath.Join(directory, entry)
		if err := os.Mkdir(path, 0o700); err != nil {
			return "", err
		}
		if err := os.Chmod(path, mode); err != nil {
			return "", err
		}
	}
	return directory, nil
}

// fsFilePrepare makes name, in shared, the directory a node:fs file fixture runs in and is handed as its argument, and
// returns how it runs there and restore, which gives locked and closed their permissions back, so the shared directory
// can be removed: its caller calls it first.
func fsFilePrepare(shared string, name string) (inputRun, func() error, error) {
	root := filepath.Join(shared, name)
	if err := os.Mkdir(root, 0o777); err != nil {
		return inputRun{}, nil, err
	}
	if err := os.Chmod(root, 0o777); err != nil {
		return inputRun{}, nil, err
	}
	for _, directory := range []string{"directory", "locked", "closed"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o755); err != nil {
			return inputRun{}, nil, err
		}
	}
	for file, data := range map[string][]byte{
		"file": []byte("abc"), "utf8": []byte("héllo 🌍"), "bom": []byte("\xef\xbb\xbfhello"), "bad": {0xe2, 0x82, 0xff, 0xed, 0xa0, 0x80},
		"utf16le": {0xff, 0xfe, 0x61, 0, 0x3d, 0xd8, 0, 0xde}, "utf16be": {0xfe, 0xff, 0, 0x61, 0xd8, 0x3d, 0xde, 0}, "utf16be-odd": {0xfe, 0xff, 0, 0x61, 0x62}, "empty": {},
		"locked/file": []byte("locked"), "closed/file": []byte("closed"),
	} {
		if err := os.WriteFile(filepath.Join(root, file), data, 0o644); err != nil {
			return inputRun{}, nil, err
		}
	}
	for link, target := range map[string]string{"link": "file", "dangling": "missing", "directory-link": "directory"} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			return inputRun{}, nil, err
		}
	}
	restore := func() error {
		for _, directory := range []string{"locked", "closed"} {
			if err := os.Chmod(filepath.Join(root, directory), 0o755); err != nil {
				return fmt.Errorf("%s stays unremovable: %w", directory, err)
			}
		}
		return nil
	}
	if err := os.Chmod(filepath.Join(root, "locked"), 0o555); err != nil {
		return inputRun{}, nil, err
	}
	if err := os.Chmod(filepath.Join(root, "closed"), 0o000); err != nil {
		restore()
		return inputRun{}, nil, err
	}
	how := inputRun{directory: root, arguments: []string{root}}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	return how, restore, nil
}

// environmentRun runs a program with exactly variables beside the sanitizers' settings: no inherited box variables, for
// the missing key and the Object.prototype witness.
func environmentRun(variables []string, name string) (run, error) {
	ctx, cancel, command := bounded(name)
	defer cancel()
	command.Env = append([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, variables...)
	return captured(ctx, name, command)
}

// realpathLayout makes root, a directory that isn't there, the filesystem node:fs's realpath fixtures observe.
func realpathLayout(root string) error {
	if err := os.Mkdir(root, 0o700); err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o700); err != nil {
		return err
	}
	for _, name := range []string{"file", "café-雪"} {
		if err := os.WriteFile(filepath.Join(root, "dir", name), []byte("data"), 0o600); err != nil {
			return err
		}
	}
	// The NUL mutant must resolve the truncated path, rather than fail ENOENT.
	if err := os.WriteFile(filepath.Join(root, "a"), []byte("data"), 0o600); err != nil {
		return err
	}
	for name, target := range map[string]string{"link": "dir", "chain": "link", "dangling": "absent-target", "loop-a": "loop-b", "loop-b": "loop-a"} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			return err
		}
	}
	return nil
}

// realpathRun runs a program in cwd with only the sanitizers' settings for its environment.
func realpathRun(cwd string, name string) (run, error) {
	ctx, cancel, command := bounded(name)
	defer cancel()
	command.Dir = cwd
	command.Env = []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}
	return captured(ctx, name, command)
}
