package main

import (
	"runtime"
	"testing"
	"time"
)

func TestRunProgramFractionalCPUBudget(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("CPU limits require Linux or macOS")
	}
	// Observe the child's inherited limit directly, without relying on wall time
	// or spending CPU to distinguish the one-second and two-second limits.
	got := runProgram(1500*time.Millisecond, nil, "/bin/sh", "-c", "ulimit -S -t")
	want := execution{Stdout: "2\n"}
	if got != want {
		t.Fatalf("1500ms CPU budget: got %+v, want %+v (round up to two seconds)", got, want)
	}
}

func TestRunCommandExactCaptureCapacity(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		command string
		want    execution
	}{
		{"stdout", "printf abc", execution{Stdout: "abc"}},
		{"stderr", "printf abc >&2", execution{Stderr: "abc"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := runCommandWithLimit(5*time.Second, nil, 3, "/bin/sh", "-c", test.command)
			if got != test.want {
				t.Fatalf("exact capture capacity: got %+v, want %+v", got, test.want)
			}
		})
	}
}
