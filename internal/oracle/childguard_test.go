package oracle

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

func runChild(command *exec.Cmd) error {
	return runChildWith(command, childguard.Options{})
}

func runChildWith(command *exec.Cmd, options childguard.Options) error {
	if filepath.Base(command.Path) != "python3" {
		return childguard.Run(command, options)
	}
	// Python drivers own their child's pipes, so report actual bytes separately
	// from the JSON observation. This is output activity, not a keepalive.
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	defer reader.Close()
	defer writer.Close()
	descriptor := 3 + len(command.ExtraFiles)
	command.ExtraFiles = append(command.ExtraFiles, writer)
	if command.Env == nil {
		command.Env = os.Environ()
	}
	root, err := filepath.Abs(filepath.Join(repository, "oracle"))
	if err != nil {
		return err
	}
	command.Env = append(command.Env, "ADAMIC_CHILDGUARD_PROGRESS_FD="+strconv.Itoa(descriptor), "PYTHONPATH="+root+string(os.PathListSeparator)+os.Getenv("PYTHONPATH"), "PYTHONDONTWRITEBYTECODE=1")
	var progress atomic.Uint64
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		buffer := make([]byte, 4096)
		for {
			n, err := reader.Read(buffer)
			progress.Add(uint64(n))
			if err != nil {
				return
			}
		}
	}()
	options.Progress = progress.Load
	err = childguard.Run(command, options)
	writer.Close()
	reader.Close()
	<-drained
	return err
}

func combinedChildOutput(command *exec.Cmd) ([]byte, error) {
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err := runChild(command)
	return output.Bytes(), err
}

func checkChild(t *testing.T, err error) {
	t.Helper()
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		t.Fatal(err)
	}
}

// Start an interactive child under the same guard. The notification synchronizes
// Process access with Start; the caller consumes the result instead of Wait.
func startChild(t *testing.T, command *exec.Cmd, options childguard.Options) <-chan error {
	started := make(chan struct{})
	done := make(chan error, 1)
	finished := make(chan struct{})
	var running atomic.Bool
	options.Started = func() { running.Store(true); close(started) }
	go func() {
		err := runChildWith(command, options)
		running.Store(false)
		close(finished)
		if command.Process == nil {
			close(started)
		}
		done <- err
	}()
	<-started
	t.Cleanup(func() {
		if running.Load() {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
		<-finished
	})
	return done
}

func TestLibraryChildguardHangWorker(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_LIBRARY_CHILD_HANG") != "1" {
		return
	}
	// A real Python child emits progress on its observation channel, then hangs
	// with a descendant holding its output pipes. The assertion must go red.
	command := exec.Command("python3", "-c", `from child_progress import progress
import subprocess,time
progress(b'ready')
subprocess.Popen(['sleep','3600'])
time.sleep(3600)`)
	err := runChildWith(command, childguard.Options{FirstOutput: 5 * time.Second, Stall: 300 * time.Millisecond, Ceiling: 10 * time.Second})
	if err != nil {
		t.Fatalf("running python3: %v", err)
	}
}

func TestLibraryChildguardRejectsHang(t *testing.T) {
	t.Parallel()
	command := exec.Command(os.Args[0], "-test.run=^TestLibraryChildguardHangWorker$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_LIBRARY_CHILD_HANG=1")
	output, err := combinedChildOutput(command)
	var exited *exec.ExitError
	if !errors.As(err, &exited) || !strings.Contains(string(output), "running python3: stalled: no output for 300ms") || !strings.Contains(string(output), "FAIL: TestLibraryChildguardHangWorker") {
		t.Fatalf("real hang did not fail via oracle guard: %v\n%s", err, output)
	}
	t.Log("real Python hang failed via output-based stall guard; process group killed without WaitDelay")
}
