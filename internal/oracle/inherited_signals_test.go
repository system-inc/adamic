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

func TestSignalTermDefaultSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGTERM, false, "Source")
}

func TestSignalTermDefaultNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGTERM, false, "Native")
}

func TestSignalTermDefaultBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGTERM, false, "Backend")
}

func TestSignalTermInheritedSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGTERM, true, "Source")
}

func TestSignalTermInheritedNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGTERM, true, "Native")
}

func TestSignalTermInheritedBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGTERM, true, "Backend")
}

func TestSignalIntDefaultSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGINT, false, "Source")
}

func TestSignalIntDefaultNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGINT, false, "Native")
}

func TestSignalIntDefaultBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGINT, false, "Backend")
}

func TestSignalIntInheritedSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGINT, true, "Source")
}

func TestSignalIntInheritedNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGINT, true, "Native")
}

func TestSignalIntInheritedBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGINT, true, "Backend")
}

func TestSignalHupDefaultSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGHUP, false, "Source")
}

func TestSignalHupDefaultNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGHUP, false, "Native")
}

func TestSignalHupDefaultBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGHUP, false, "Backend")
}

func TestSignalHupInheritedSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGHUP, true, "Source")
}

func TestSignalHupInheritedNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGHUP, true, "Native")
}

func TestSignalHupInheritedBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGHUP, true, "Backend")
}

func TestSignalQuitDefaultSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGQUIT, false, "Source")
}

func TestSignalQuitDefaultNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGQUIT, false, "Native")
}

func TestSignalQuitDefaultBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGQUIT, false, "Backend")
}

func TestSignalQuitInheritedSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGQUIT, true, "Source")
}

func TestSignalQuitInheritedNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGQUIT, true, "Native")
}

func TestSignalQuitInheritedBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGQUIT, true, "Backend")
}

func TestSignalUsr2DefaultSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGUSR2, false, "Source")
}

func TestSignalUsr2DefaultNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGUSR2, false, "Native")
}

func TestSignalUsr2DefaultBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGUSR2, false, "Backend")
}

func TestSignalUsr2InheritedSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGUSR2, true, "Source")
}

func TestSignalUsr2InheritedNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGUSR2, true, "Native")
}

func TestSignalUsr2InheritedBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGUSR2, true, "Backend")
}

func TestSignalAlrmDefaultSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGALRM, false, "Source")
}

func TestSignalAlrmDefaultNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGALRM, false, "Native")
}

func TestSignalAlrmDefaultBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGALRM, false, "Backend")
}

func TestSignalAlrmInheritedSource(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGALRM, true, "Source")
}

func TestSignalAlrmInheritedNative(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGALRM, true, "Native")
}

func TestSignalAlrmInheritedBackend(test *testing.T) {
	test.Parallel()
	signalLeg(test, syscall.SIGALRM, true, "Backend")
}

func signalLeg(test *testing.T, stop syscall.Signal, ignored bool, leg string) {
	test.Helper()
	const fixture = "internal/oracle/testdata/killed_after_output.a"
	path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	name := "node"
	arguments := []string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path}
	if leg != "Source" {
		_, binary, script := sanitized(test, fixture)
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
	timeoutContext, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	command := exec.CommandContext(timeoutContext, name, arguments...)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	var stdout, stderr lockedBuffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if startError := command.Start(); startError != nil {
		test.Fatal(startError)
	}
	defer command.Process.Kill()
	ready := func() bool {
		if leg == "Native" {
			return cpuSeconds(test, command.Process.Pid) >= 0.5
		}
		return bytes.Contains(stdout.Bytes(), []byte("\n"))
	}
	for !ready() {
		if timeoutContext.Err() != nil {
			command.Wait()
			test.Fatalf("%s never became ready: %s", leg, stderr.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if signalError := command.Process.Signal(stop); signalError != nil {
		test.Fatal(signalError)
	}
	command.Wait()
	status := command.ProcessState.Sys().(syscall.WaitStatus)
	if !status.Signaled() ||
		status.Signal() != stop ||
		stderr.String() != "" ||
		stdout.String() != "started, about to work for a long time\n" {
		test.Fatalf(
			"%s ignored=%v: want killed by %s with its complete line; got %v stdout %q stderr %q",
			leg,
			ignored,
			stop,
			status,
			stdout.String(),
			stderr.String(),
		)
	}
}
