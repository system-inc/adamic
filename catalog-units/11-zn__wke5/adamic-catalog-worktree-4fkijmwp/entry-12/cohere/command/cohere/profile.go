package main

import (
	"fmt"
	"os"
	"runtime/pprof"
)

// profileFile is the open `--profile` file while the CPU profile is running, nil otherwise.
var profileFile *os.File

// startProfile begins writing a Go CPU profile of the rest of the run to path, for `go tool pprof`.
func startProfile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating the profile: %w", err)
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return fmt.Errorf("starting the profile: %w", err)
	}
	profileFile = file
	// A profiled run never returns early: the profile is something the caller reads, and it is complete
	// only once this process stops and closes it, after the verdict would have gone. Closing the
	// descriptor unanswered tells the dispatcher to wait for this process instead. See sendVerdict.
	dropVerdict()
	return nil
}

// exitProcess ends the process with code, finishing the profile first. os.Exit runs no deferred call, so
// every exit goes through here, or a profiled run would leave a truncated profile behind.
func exitProcess(code int) {
	if profileFile != nil {
		pprof.StopCPUProfile()
		if err := profileFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "cohere: closing the profile: %v\n", err)
		}
		profileFile = nil
	}
	// Every exit answers the caller, if a run cache has not already: what follows, keeping what the run read
	// in the content pack and the heap's teardown, is nothing the caller waits for. See sendVerdict.
	sendVerdict(code)
	saveContentPack()
	os.Exit(code)
}
