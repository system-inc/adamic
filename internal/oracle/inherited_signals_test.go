package oracle

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestSignalTERMDefaultSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGTERM, false, "Source")
}

func TestSignalTERMDefaultNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGTERM, false, "Native")
}

func TestSignalTERMDefaultBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGTERM, false, "Backend")
}

func TestSignalTERMInheritedSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGTERM, true, "Source")
}

func TestSignalTERMInheritedNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGTERM, true, "Native")
}

func TestSignalTERMInheritedBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGTERM, true, "Backend")
}

func TestSignalINTDefaultSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGINT, false, "Source")
}

func TestSignalINTDefaultNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGINT, false, "Native")
}

func TestSignalINTDefaultBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGINT, false, "Backend")
}

func TestSignalINTInheritedSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGINT, true, "Source")
}

func TestSignalINTInheritedNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGINT, true, "Native")
}

func TestSignalINTInheritedBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGINT, true, "Backend")
}

func TestSignalHUPDefaultSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGHUP, false, "Source")
}

func TestSignalHUPDefaultNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGHUP, false, "Native")
}

func TestSignalHUPDefaultBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGHUP, false, "Backend")
}

func TestSignalHUPInheritedSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGHUP, true, "Source")
}

func TestSignalHUPInheritedNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGHUP, true, "Native")
}

func TestSignalHUPInheritedBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGHUP, true, "Backend")
}

func TestSignalQUITDefaultSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGQUIT, false, "Source")
}

func TestSignalQUITDefaultNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGQUIT, false, "Native")
}

func TestSignalQUITDefaultBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGQUIT, false, "Backend")
}

func TestSignalQUITInheritedSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGQUIT, true, "Source")
}

func TestSignalQUITInheritedNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGQUIT, true, "Native")
}

func TestSignalQUITInheritedBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGQUIT, true, "Backend")
}

func TestSignalUSR2DefaultSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGUSR2, false, "Source")
}

func TestSignalUSR2DefaultNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGUSR2, false, "Native")
}

func TestSignalUSR2DefaultBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGUSR2, false, "Backend")
}

func TestSignalUSR2InheritedSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGUSR2, true, "Source")
}

func TestSignalUSR2InheritedNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGUSR2, true, "Native")
}

func TestSignalUSR2InheritedBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGUSR2, true, "Backend")
}

func TestSignalALRMDefaultSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGALRM, false, "Source")
}

func TestSignalALRMDefaultNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGALRM, false, "Native")
}

func TestSignalALRMDefaultBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGALRM, false, "Backend")
}

func TestSignalALRMInheritedSource(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGALRM, true, "Source")
}

func TestSignalALRMInheritedNative(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGALRM, true, "Native")
}

func TestSignalALRMInheritedBackend(t *testing.T) {
	t.Parallel()
	signalLeg(t, syscall.SIGALRM, true, "Backend")
}

func signalLeg(t *testing.T, stop syscall.Signal, ignored bool, leg string) {
	t.Helper()
	const fixture = "internal/oracle/testdata/killed_after_output.a"
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	name := "node"
	arguments := []string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path}
	if leg != "Source" {
		_, binary, script := sanitized(t, fixture)
		if leg == "Native" {
			name, arguments = binary, nil
		} else {
			arguments[len(arguments)-1] = script
		}
	}
	if ignored {
		arguments = append([]string{"-c", `trap "" TERM INT HUP QUIT USR2 ALRM; exec "$0" "$@"`, name}, arguments...)
		name = "sh"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	var stdout, stderr lockedBuffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer command.Process.Kill()
	ready := func() bool {
		if leg == "Native" {
			return cpuSeconds(t, command.Process.Pid) >= 0.5
		}
		return bytes.Contains(stdout.Bytes(), []byte("\n"))
	}
	for !ready() {
		if ctx.Err() != nil {
			command.Wait()
			t.Fatalf("%s never became ready: %s", leg, stderr.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := command.Process.Signal(stop); err != nil {
		t.Fatal(err)
	}
	command.Wait()
	status := command.ProcessState.Sys().(syscall.WaitStatus)
	if !status.Signaled() || status.Signal() != stop || stderr.String() != "" || stdout.String() != "started, about to work for a long time\n" {
		t.Fatalf("%s ignored=%v: want killed by %s with its complete line; got %v stdout %q stderr %q", leg, ignored, stop, status, stdout.String(), stderr.String())
	}
}
