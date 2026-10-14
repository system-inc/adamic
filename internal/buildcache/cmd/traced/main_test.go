package main

import (
	"strings"
	"testing"
)

// strace's own lines are kept from the stderr it shares with the command, however the writes split them, and the
// command's status replaces strace's only when every one of them is PTRACE_LISTEN failing on a tracee already gone
// (home-7's dry run, Oct 11). Mutants: any complaint accepted; no complaint accepted as one.
func TestOnlyAFailedListenLetsTheCommandsStatusStand(t *testing.T) {
	t.Parallel()
	listen := "strace: ptrace(PTRACE_LISTEN,pid:572077,sig:0): Input/output error\n"
	for _, each := range []struct {
		name   string
		writes []string
		only   bool
	}{
		{"one failed listen among the command's output", []string{"ok  \tgithub.com/x\t1.0s\n", listen[:20], listen[20:], "--- PASS\n"}, true},
		{"two failed listens", []string{listen, listen}, true},
		{"nothing from strace", []string{"ok  \tgithub.com/x\t1.0s\n"}, false},
		{"another complaint", []string{listen, "strace: Process 12 detached\n"}, false},
		{"a listen failing otherwise", []string{"strace: ptrace(PTRACE_LISTEN,pid:5,sig:0): No such process\n"}, false},
	} {
		complaints := &straceComplaints{}
		for _, write := range each.writes {
			if _, err := complaints.Write([]byte(write)); err != nil {
				t.Fatal(err)
			}
		}
		if got := complaints.onlyListen(); got != each.only {
			t.Errorf("%s: onlyListen %v (lines %q)", each.name, got, complaints.lines)
		}
	}
}

// A run strace complained and failed under before the trace held the command's exit is lost to the tracer (#j4hwdks),
// never a settled-nothing fill; a run whose command exited, one strace passed, and one strace failed silently (the
// command killed) aren't. Mutants: the command's exit ignored; strace's status ignored; a failure with no complaint
// read as lost.
func TestARunStraceFailedUnderIsLostToTheTracer(t *testing.T) {
	t.Parallel()
	listen := &straceComplaints{lines: []string{"strace: ptrace(PTRACE_LISTEN,pid:3683302,sig:0): Input/output error"}}
	for _, each := range []struct {
		name       string
		exited     bool
		status     int
		complaints *straceComplaints
		lost       bool
	}{
		{"strace failed before the command exited", false, 1, listen, true},
		{"the command exited in the trace", true, 1, listen, false},
		{"strace exited 0", false, 0, listen, false},
		{"strace failed with nothing to say", false, 137, &straceComplaints{}, false},
	} {
		why := tracerLost(each.exited, each.status, each.complaints)
		if (why != "") != each.lost || each.lost && !strings.Contains(why, "PTRACE_LISTEN") {
			t.Errorf("%s: lost %q", each.name, why)
		}
	}
}
