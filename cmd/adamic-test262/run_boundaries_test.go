package main

import (
	"runtime"
	"testing"
	"time"
)

func TestRunProgramFractionalCPUBudget(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		test.Skip("CPU limits require Linux or macOS")
	}
	// Observe the child's inherited limit directly, without relying on wall time
	// or spending CPU to distinguish the one-second and two-second limits.
	got := runProgram(1500*time.Millisecond, nil, "/bin/sh", "-c", "ulimit -S -t")
	want := execution{Stdout: "2\n"}
	if got != want {
		test.Fatalf("1500ms CPU budget: got %+v, want %+v (round up to two seconds)", got, want)
	}
}

func TestRunCommandExactCaptureCapacity(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		name    string
		command string
		want    execution
	}{
		{"stdout", "printf abc", execution{Stdout: "abc"}},
		{"stderr", "printf abc >&2", execution{Stderr: "abc"}},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			got := runCommandWithLimit(5*time.Second, nil, 3, "/bin/sh", "-c", testCase.command)
			if got != testCase.want {
				test.Fatalf("exact capture capacity: got %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestRunCommandCaptureLimit(test *testing.T) {
	test.Parallel()
	for _, one := range []struct {
		name, command string
		want          execution
	}{
		{
			"stdout",
			"printf abcd",
			execution{Stdout: "abc", Stderr: "command output exceeded capture limit\n", Exit: -1},
		},
		{"stderr", "printf abcd >&2", execution{Stderr: "command output exceeded capture limit\n", Exit: -1}},
	} {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			if got := runCommandWithLimit(5*time.Second, nil, 3, "/bin/sh", "-c", one.command); got != one.want {
				test.Fatalf("capture overflow: got %+v, want %+v", got, one.want)
			}
		})
	}
}
