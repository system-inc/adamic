package oracle

import (
	"os/exec"
	"testing"
)

// Silent compilation uses childguard's first-output window rather than a
// scheduler-sensitive five-minute deadline.
func boundedCompile(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	return bounded(t, name, arguments...)
}
