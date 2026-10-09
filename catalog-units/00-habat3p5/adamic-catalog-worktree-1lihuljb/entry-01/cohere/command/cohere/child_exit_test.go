package main

import (
	"os/exec"
	"syscall"
	"testing"
)

// childExitCode is a cohere child's exit code, from the error its run returned.
//
// A child a signal ended did not choose an exit code, and Go reports it as -1. Four tests once failed
// together that way at load 168, reading "exit -1" and nothing else, and the first theory was a deadline to
// loosen. There was none: the four died within 0.6s of each other, from a kill by name somewhere on the
// machine (#wwhrpm0). So a signaled child is said to be one, by name, before the test reports its -1. It
// is an error rather than a fatal one because some of these runs are made from a subtest's goroutine
// through the enclosing test's t, where only Errorf is safe.
func childExitCode(t testing.TB, exitError *exec.ExitError) int {
	t.Helper()
	if status, isStatus := exitError.Sys().(syscall.WaitStatus); isStatus && status.Signaled() {
		t.Errorf("cohere was killed by signal %d (%s), from outside it rather than an exit it chose: rerun "+
			"the test alone, and look for a kill by name (pkill, killall) or memory pressure on the machine",
			int(status.Signal()), status.Signal())
	}
	return exitError.ExitCode()
}
