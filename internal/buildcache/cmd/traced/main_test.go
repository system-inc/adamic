package main

import (
	"os"
	"path/filepath"
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

// A run starts from the state ADAMIC_TRACE_TREE_STATE's file holds rather than taking its own, refuses a file that is
// missing or empty, and takes its own when the variable is unset; the variable never reaches the command (#2q48btg).
// Mutants: the file's state ignored for one taken; an empty file read as a state; the variable left set.
// Serial: it sets ADAMIC_TRACE_TREE_STATE, which t.Setenv refuses beside parallel tests.
func TestARunStartsFromTheWavesTreeState(t *testing.T) {
	taken := 0
	take := func() (string, error) {
		taken++
		return "taken now\n", nil
	}
	directory := t.TempDir()
	wave, empty := filepath.Join(directory, "wave"), filepath.Join(directory, "empty")
	if err := os.WriteFile(wave, []byte("HEAD abc\nrepository /cohere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, each := range []struct {
		name, path, state string
		fails             bool
		taken             int
	}{
		{"the wave's state", wave, "HEAD abc\nrepository /cohere\n", false, 0},
		{"no variable", "", "taken now\n", false, 1},
		{"an empty file", empty, "", true, 0},
		{"a missing file", filepath.Join(directory, "gone"), "", true, 0},
	} {
		taken = 0
		t.Setenv("ADAMIC_TRACE_TREE_STATE", each.path)
		state, err := treeStateBefore(take)
		if (err != nil) != each.fails || state != each.state || taken != each.taken {
			t.Errorf("%s: state %q, err %v, taken %d times", each.name, state, err, taken)
		}
		if _, set := os.LookupEnv("ADAMIC_TRACE_TREE_STATE"); set {
			t.Errorf("%s: ADAMIC_TRACE_TREE_STATE is still set for the command", each.name)
		}
	}
}
