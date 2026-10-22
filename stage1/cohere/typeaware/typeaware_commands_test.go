package typeaware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
)

const typeAwareChildLimit = 90 * time.Second

var typeAwareCommandGroups sync.Map

// A fresh Unix process group includes the command's compiler descendants. This
// works on macOS and Linux without an external timeout program or /proc.
//
// No WaitDelay: exec reads a buffered stream to its end, where one closed the reads a second after the child exited and
// failed a build or probe that had finished but was slow to drain under load (#n3gk22h's shape). The commands
// typeAwareRunCommandContext runs aren't childguard.Capture's, which has no Started: typeAwareDeadline's cooked kills the
// groups it registers before it exits the process, and a group it couldn't register would outlive the exit. go and
// stage 0 wait for the compilers they start before they exit, so nothing holds a pipe after them, and the unit's
// ceiling ends a wait on one that did.
func typeAwareContextCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return command
}

func typeAwareRunCommand(original *exec.Cmd) error {
	ctx, cancel := context.WithTimeout(context.Background(), typeAwareChildLimit)
	defer cancel()
	return typeAwareRunCommandContext(ctx, original)
}

// Build-cache misses belong to shared setup; the outer unit limit bounds them.
func typeAwareRunBuildCommand(original *exec.Cmd) error {
	return typeAwareRunCommandContext(context.Background(), original)
}

func typeAwareRunCommandContext(ctx context.Context, original *exec.Cmd) error {
	command := typeAwareContextCommand(ctx, original.Path, original.Args[1:]...)
	command.Args = append([]string(nil), original.Args...)
	command.Dir, command.Env = original.Dir, original.Env
	command.Stdin, command.Stdout, command.Stderr = original.Stdin, original.Stdout, original.Stderr
	command.ExtraFiles = original.ExtraFiles
	if original.SysProcAttr != nil {
		attributes := *original.SysProcAttr
		attributes.Setpgid, attributes.Pgid = true, 0
		command.SysProcAttr = &attributes
	}
	if err := command.Start(); err != nil {
		return err
	}
	pid := command.Process.Pid
	typeAwareCommandGroups.Store(pid, struct{}{})
	defer typeAwareCommandGroups.Delete(pid)
	return command.Wait()
}
func typeAwareKillCommandGroups() {
	typeAwareCommandGroups.Range(func(key, _ any) bool {
		_ = syscall.Kill(-key.(int), syscall.SIGKILL)
		return true
	})
}

// BuildTSGo creates clang internally. Running that API in a child test binary
// gives the unchanged production build pipeline the same cancellable group.
func typeAwareNativeBuild(source, output, archive string, sanitize bool) error {
	directory, err := os.MkdirTemp(productDirectory, "native-command-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	request := typeAwareNativeRequest{source, output, archive, sanitize}
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "request.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	command := exec.Command(os.Args[0], "-test.run=^TestTypeAwareNativeBuildChild$")
	command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_NATIVE_REQUEST="+path)
	var outputText bytes.Buffer
	command.Stdout, command.Stderr = &outputText, &outputText
	if err := typeAwareRunBuildCommand(command); err != nil {
		return fmt.Errorf("native build: %w\n%s", err, outputText.String())
	}
	return nil
}

type typeAwareNativeRequest struct {
	Source, Output, Archive string
	Sanitize                bool
}

// Not parallel: native build child owns the requested shared product output.
func TestTypeAwareNativeBuildChild(t *testing.T) {
	path := os.Getenv("ADAMIC_TYPEAWARE_NATIVE_REQUEST")
	if path == "" {
		t.Skip("native build subprocess only")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var request typeAwareNativeRequest
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	if err := native.BuildTSGo(request.Source, request.Output, request.Archive, native.Options{Sanitize: request.Sanitize}); err != nil {
		t.Fatal(err)
	}
}

// Exercise Cancel itself, rather than only the enclosing unit watchdog. The
// child starts a descendant in its inherited group, as go and clang do.
// Not parallel: parent and descendants share the process group and marker path.
func TestTypeAwareContextKillsCompilerGroup(t *testing.T) {
	mode := os.Getenv("ADAMIC_TYPEAWARE_GROUP_PROBE")
	marker := os.Getenv("ADAMIC_TYPEAWARE_GROUP_MARKER")
	if mode == "marker" {
		time.Sleep(time.Second)
		if err := os.WriteFile(marker, []byte("survived"), 0600); err != nil {
			t.Fatal(err)
		}
		return
	}
	if mode == "leader" {
		command := exec.Command(os.Args[0], "-test.run=^TestTypeAwareContextKillsCompilerGroup$")
		command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_GROUP_PROBE=marker")
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
		return
	}
	marker = filepath.Join(t.TempDir(), "marker")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	command := typeAwareContextCommand(ctx, os.Args[0], "-test.run=^TestTypeAwareContextKillsCompilerGroup$")
	command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_GROUP_PROBE=leader", "ADAMIC_TYPEAWARE_GROUP_MARKER="+marker)
	// Not childguard.Capture: its own group kill when ctx ends would do Cancel's work and hide a Cancel that kills only
	// the leader. The output is only the failure's message, and only the leader holds the pipe (the marker child's
	// streams are the null device), so with no WaitDelay the read ends when the leader dies.
	output, err := command.CombinedOutput()
	if err == nil || !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("context did not kill leader: %v %s", err, output)
	}
	time.Sleep(time.Second)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("context left compiler descendant alive: %v", err)
	}
}

// Inventory probes use the same bound as builds, including subprocess cleanup: they are six's products' key-time work,
// shared setup the unit's own budget bounds, never typeAwareChildLimit's 90 s for case work (#h4hrxpe).
func typeAwareCommandOutput(command *exec.Cmd) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := typeAwareRunBuildCommand(command)
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		exit.Stderr = append([]byte(nil), stderr.Bytes()...)
	}
	return stdout.Bytes(), err
}
