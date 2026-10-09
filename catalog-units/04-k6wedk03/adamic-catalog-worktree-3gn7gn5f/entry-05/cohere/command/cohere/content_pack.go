package main

import (
	"fmt"
	"os"

	"github.com/system-inc/cohere/internal/types/program"
)

// The content pack's command half: open the project's pack before a build reads anything, and keep what
// the run read from disk on the way out, after the caller has its answer. See program.ContentPack.

// activeContentPack is the pack this run's builds read through, and activeContentPackRoot the project it
// belongs to. Nil when the cache is off or nothing was built.
var (
	activeContentPack     *program.ContentPack
	activeContentPackRoot string
)

// runCacheCheckStats is what the run cache's check statted this run, nil when no check ran. A pack opened after
// it validates against these rather than statting each file again. See program.StatSnapshot.
var runCacheCheckStats *program.StatSnapshot

// openContentPack returns the project's content pack for a build, opening it once per run, or nil with
// the cache off. A pack it cannot trust is said once and read around: every file then comes from disk.
func openContentPack(root string) *program.ContentPack {
	if cacheOff {
		return nil
	}
	if activeContentPack != nil && activeContentPackRoot == root {
		return activeContentPack
	}
	// Created before anything is read, for the reason prepareCacheDirectory gives.
	prepareCacheDirectory(root)
	pack, err := program.OpenContentPack(cacheDirectory(root))
	if err != nil {
		// A fact about this invocation, never about the tree, so a replay never says it again.
		fmt.Fprintf(accountOutput(invocationOutput(os.Stderr)), "note: %v\n", err)
	}
	pack.TrustStats(runCacheCheckStats)
	activeContentPack, activeContentPackRoot = pack, root
	return pack
}

// distrustCheckStats has the rest of the run stat what it reads afresh rather than trust the run cache's check. The
// fix phase calls it before rebuilding the graph: a file it rewrote still matches its stat from the check, taken
// before the rewrite, and so does its entry in the pack when nothing else changed it since the last run, so the
// rebuilt graph was served the bytes the fixer had just replaced (#kdee854).
func distrustCheckStats() {
	runCacheCheckStats = nil
	if activeContentPack != nil {
		activeContentPack.TrustStats(nil)
	}
}

// saveContentPack keeps what this run read from disk, under the cache directory's lock so it never
// interleaves with another run's write. It runs after the verdict, so a run that answered early pays
// nothing for it, and what goes wrong is said to the next run.
func saveContentPack() {
	pack := activeContentPack
	if pack == nil {
		return
	}
	activeContentPack = nil
	directory := cacheDirectory(activeContentPackRoot)
	if _, _, damaged := pack.Counts(); damaged > 0 {
		cacheNote(directory, fmt.Sprintf("%d files in the content pack failed their checksum or bounds and were read from disk; the pack is rewritten without them", damaged))
	}
	release := holdTableLock(directory)
	err := pack.Save()
	release()
	if err != nil {
		cacheNote(directory, fmt.Sprintf("the content pack could not be written: %v", firstLine(err.Error())))
	}
}
