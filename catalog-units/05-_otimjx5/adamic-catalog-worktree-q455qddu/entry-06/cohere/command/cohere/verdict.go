package main

import (
	"os"
	"os/signal"
	"syscall"
)

// VerdictVariable names the file descriptor the dispatcher listens on for this run's exit code, when it
// returns to the caller before the engine has finished. See sendVerdict.
const VerdictVariable = "COHERE_VERDICT_FD"

// verdictFile is where this run's exit code goes the moment the report is out, nil when nobody listens.
//
// Taken at startup, marked close-on-exec and removed from the environment, so a process this one starts,
// the Swift engine or anything else, neither holds the descriptor nor finds the variable pointing at a
// number that means something else in it.
var verdictFile = takeVerdictFile()

// sendVerdict tells the dispatcher this run's exit code, once, after everything the caller will read has
// been written, so the dispatcher can return while this process finishes what nobody waits for: recording
// the run, writing the cache table, and handing back a heap of several gigabytes (#zqsdzbq, early return).
//
// It reports whether a listener got the verdict. After it does, nothing more may reach stdout or stderr,
// since the caller has moved on; see backgroundNote.
func sendVerdict(exitCode int) bool {
	if verdictFile == nil {
		return false
	}
	file := verdictFile
	verdictFile = nil
	_, err := file.Write([]byte{byte(exitCode)})
	file.Close()
	if err != nil {
		return false
	}
	verdictSent = true
	// The caller has its answer and its prompt back. An interrupt or a hangup meant for whatever runs next must
	// not stop this process between writing the table and renaming it into place; the rename keeps the old
	// table whole either way, but there is no reason to throw the new one away.
	signal.Ignore(os.Interrupt, syscall.SIGHUP)
	return true
}

// dropVerdict closes the verdict descriptor unanswered, so the dispatcher waits for this process to end
// and exits with its status, as it would have under exec.
func dropVerdict() {
	if verdictFile != nil {
		verdictFile.Close()
		verdictFile = nil
	}
}

// verdictSent is whether the caller has already been answered, so whatever this process says afterward
// must go to the next run instead of to a terminal it no longer owns.
var verdictSent bool
