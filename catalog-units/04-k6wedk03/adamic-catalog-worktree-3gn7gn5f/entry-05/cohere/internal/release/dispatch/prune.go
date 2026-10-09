package dispatch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Nothing pruned the binary cache. Measured 2026-10-02: 15 GB across 303 entries, 298 of them cohere
// binaries going back to 6 September, built under the old launcher's one-binary-per-working-tree-edit
// naming, which nothing resolves any more. Every committed build adds one, so the cache only grows.
//
// The policy keeps what a run can still reach and removes the rest one named file at a time:
//
//   - The launcher execs only the binary for the commit it just resolved, and a gate that resolved
//     before a newer commit landed execs its own within milliseconds. So recency covers every live
//     reference: the newest few of each kind, and anything touched within the last hour, which is far
//     past any build or resolve and covers a build that is still writing its `.partial-` file.
//   - Files other members own are kept by name: `cohere-swift-current` is a hand-built engine their
//     proving-ground checks exec directly, and the launcher itself lives here.
//   - Anything the prune does not recognise is kept and reported. A name it cannot classify is a name
//     it has no business deleting.
//
// The compiler extractions follow the same policy. Each is about 400 MB, and they used to be reported
// and never removed, on the reasoning that removing one is a walk of 66,000 files and a pin bump is
// rare. The disk filled on 2026-10-02 with a stale one reported on every build since 06:41 and nobody
// removing it, which is what a report with no owner turns into. So one not in use by the build just
// made, and not used within the hour, is removed. A cache hit refreshes its extraction's modification
// time (see ensureCompiler), so the hour counts from the last build that read it rather than from when
// it was extracted, and a build still compiling against an older pin keeps it.
//
// Snapshot directories are removed by the build that made them. One survives only when that build was
// killed before its deferred removal ran, so one older than the hour has no writer and is removed too.

// pruneKeepNewest is how many of each kind of hash-named binary survive by recency. Generous on
// purpose: the files are the cheap part, and an older one costs a rebuild only if it is ever asked for.
const pruneKeepNewest = 8

// pruneInFlightWindow protects anything touched recently: a build still writing, or a binary a run
// resolved a moment ago.
const pruneInFlightWindow = time.Hour

// prunedByName are the files in the binary directory that are never removed.
var prunedByName = map[string]string{
	developmentBinaryName:           "the --dev binary",
	developmentBinaryName + ".hash": "the --dev binary's recorded inputs",
	"cohere-dispatch":               "the launcher",
	"cohere-swift-current":          "@system_cohere_swift's proving-ground engine",
	"cohere-swift-current.tmp":      "@system_cohere_swift's engine mid-install",
}

// PrunePlan is what a prune would do, decided without touching anything.
type PrunePlan struct {
	// Examined is how many entries the binary directory held. A plan that examined nothing says so,
	// so a prune pointed at the wrong directory cannot read as a tidy cache.
	Examined int

	// Remove are the files to delete, each with its size.
	Remove []PrunedFile

	// Unrecognised are names the prune kept because it could not classify them.
	Unrecognised []string

	// StaleCompilers are compiler extractions to remove: not the current build's, and unused for the
	// in-flight window. Bytes is the extraction's total size.
	StaleCompilers []PrunedFile

	// StaleSnapshots are snapshot directories a killed build left behind, older than the window.
	StaleSnapshots []PrunedFile
}

// PrunedFile is one file, or one directory of the compiler or snapshot cache, that a prune removes.
type PrunedFile struct {
	Name  string
	Bytes int64
}

// Bytes is the total a plan reclaims.
func (plan PrunePlan) Bytes() int64 {
	total := int64(0)
	for _, list := range [][]PrunedFile{plan.Remove, plan.StaleCompilers, plan.StaleSnapshots} {
		for _, file := range list {
			total += file.Bytes
		}
	}
	return total
}

// PlanPrune decides what to remove from the binary directory.
//
// keep names binaries that must survive whatever their age, such as the one just built. currentCompiler
// is the compiler extraction in use, or empty when it is not known.
func PlanPrune(paths Paths, keep []string, currentCompiler string, now time.Time) (PrunePlan, error) {
	entries, err := os.ReadDir(paths.BinaryDirectory())
	if err != nil {
		return PrunePlan{}, fmt.Errorf("reading the binary cache %s: %w", paths.BinaryDirectory(), err)
	}

	plan := PrunePlan{Examined: len(entries)}
	kept := map[string]bool{}
	for _, path := range keep {
		kept[filepath.Base(path)] = true
	}

	type candidate struct {
		name     string
		bytes    int64
		modified time.Time
	}
	kinds := map[string][]candidate{}
	cohereKind := platformBinaryPrefix()
	swiftKind := fmt.Sprintf("cohere-swift-%s-%s-", runtime.GOOS, runtime.GOARCH)

	for _, entry := range entries {
		name := entry.Name()
		if _, named := prunedByName[name]; named || kept[name] {
			continue
		}
		information, err := entry.Info()
		if err != nil {
			// Gone between the listing and the stat, most likely renamed into place by a build.
			continue
		}
		if !information.Mode().IsRegular() {
			plan.Unrecognised = append(plan.Unrecognised, name)
			continue
		}
		if now.Sub(information.ModTime()) < pruneInFlightWindow {
			continue
		}

		switch {
		case strings.Contains(name, ".partial-"):
			// Older than any build takes, so its writer is gone.
			plan.Remove = append(plan.Remove, PrunedFile{Name: name, Bytes: information.Size()})
		case strings.HasPrefix(name, swiftKind):
			kinds[swiftKind] = append(kinds[swiftKind], candidate{name, information.Size(), information.ModTime()})
		case strings.HasPrefix(name, cohereKind):
			kinds[cohereKind] = append(kinds[cohereKind], candidate{name, information.Size(), information.ModTime()})
		default:
			plan.Unrecognised = append(plan.Unrecognised, name)
		}
	}

	for _, candidates := range kinds {
		sort.Slice(candidates, func(first int, second int) bool {
			if !candidates[first].modified.Equal(candidates[second].modified) {
				return candidates[first].modified.After(candidates[second].modified)
			}
			return candidates[first].name > candidates[second].name
		})
		for index, candidate := range candidates {
			if index < pruneKeepNewest {
				continue
			}
			plan.Remove = append(plan.Remove, PrunedFile{Name: candidate.name, Bytes: candidate.bytes})
		}
	}
	sort.Slice(plan.Remove, func(first int, second int) bool { return plan.Remove[first].Name < plan.Remove[second].Name })
	sort.Strings(plan.Unrecognised)

	// Without the current compiler named, nothing says which extraction a build is reading, so none is
	// removed.
	if currentCompiler != "" {
		plan.StaleCompilers, err = planStaleDirectories(paths.CompilerDirectory(), filepath.Base(currentCompiler), now)
		if err != nil {
			return PrunePlan{}, fmt.Errorf("reading the compiler cache: %w", err)
		}
	}
	plan.StaleSnapshots, err = planStaleDirectories(paths.SnapshotDirectory(), "", now)
	if err != nil {
		return PrunePlan{}, fmt.Errorf("reading the snapshot cache: %w", err)
	}
	return plan, nil
}

// planStaleDirectories lists the directories under parent that are not current and were not touched
// within the in-flight window, with each one's size. A missing parent holds nothing to remove.
func planStaleDirectories(parent string, current string, now time.Time) ([]PrunedFile, error) {
	entries, err := os.ReadDir(parent)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	stale := []PrunedFile{}
	for _, entry := range entries {
		if entry.Name() == current || !entry.IsDir() {
			continue
		}
		information, err := entry.Info()
		if err != nil || now.Sub(information.ModTime()) < pruneInFlightWindow {
			continue
		}
		stale = append(stale, PrunedFile{Name: entry.Name(), Bytes: directoryBytes(filepath.Join(parent, entry.Name()))})
	}
	return stale, nil
}

// directoryBytes is the total size of the files under a directory, for the log. A file it cannot stat
// counts as nothing, since the number reports what was reclaimed and decides nothing.
func directoryBytes(directory string) int64 {
	total := int64(0)
	filepath.WalkDir(directory, func(_ string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if information, err := entry.Info(); err == nil {
			total += information.Size()
		}
		return nil
	})
	return total
}

// ApplyPrune removes the plan's files, one named file at a time, and returns what it removed.
//
// Each removal is the binary directory joined with a name read from that directory's own listing,
// never a pattern and never a walk. A file already gone is skipped, not an error: a concurrent prune
// got there first.
func ApplyPrune(paths Paths, plan PrunePlan) ([]PrunedFile, error) {
	removed := []PrunedFile{}
	for _, file := range plan.Remove {
		if file.Name == "" || file.Name != filepath.Base(file.Name) {
			return removed, fmt.Errorf("refusing to remove %q from the binary cache: it is not a plain file name", file.Name)
		}
		err := os.Remove(filepath.Join(paths.BinaryDirectory(), file.Name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return removed, fmt.Errorf("removing %s from the binary cache: %w", file.Name, err)
		}
		removed = append(removed, file)
	}

	// A directory goes whole, which is the one walk the prune takes, and only by a name read from its
	// parent's own listing: the same plain-name check as a file, joined to the one parent it was planned
	// from.
	for _, group := range []struct {
		parent string
		list   []PrunedFile
	}{
		{paths.CompilerDirectory(), plan.StaleCompilers},
		{paths.SnapshotDirectory(), plan.StaleSnapshots},
	} {
		for _, directory := range group.list {
			if directory.Name == "" || directory.Name == "." || directory.Name == ".." || directory.Name != filepath.Base(directory.Name) {
				return removed, fmt.Errorf("refusing to remove %q from %s: it is not a plain directory name", directory.Name, group.parent)
			}
			path := filepath.Join(group.parent, directory.Name)
			information, err := os.Lstat(path)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			// Asked again at the moment of removal. The plan's size walk takes a moment, and a build that
			// started meanwhile has marked the extraction in use; its compile must not lose it.
			if err == nil && time.Since(information.ModTime()) < pruneInFlightWindow {
				continue
			}
			if err := os.RemoveAll(path); err != nil {
				return removed, fmt.Errorf("removing %s: %w", path, err)
			}
			removed = append(removed, PrunedFile{Name: filepath.Join(filepath.Base(group.parent), directory.Name), Bytes: directory.Bytes})
		}
	}
	return removed, nil
}

// PruneLogPath is where every prune records what it examined and removed.
func (paths Paths) PruneLogPath() string {
	return filepath.Join(paths.CacheDirectory, "prune.log")
}

// RecordPrune appends one prune's outcome to the log, names included, so a removal that matters later
// can be traced to the run that made it.
func RecordPrune(paths Paths, plan PrunePlan, removed []PrunedFile, now time.Time) error {
	bytes := int64(0)
	for _, file := range removed {
		bytes += file.Bytes
	}
	var entry strings.Builder
	fmt.Fprintf(&entry, "%s examined %d, removed %d (%d bytes)\n", now.Format(time.RFC3339), plan.Examined, len(removed), bytes)
	for _, file := range removed {
		fmt.Fprintf(&entry, "  removed %s (%d bytes)\n", file.Name, file.Bytes)
	}
	for _, name := range plan.Unrecognised {
		fmt.Fprintf(&entry, "  kept unrecognised %s\n", name)
	}

	log, err := os.OpenFile(paths.PruneLogPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening the prune log: %w", err)
	}
	defer log.Close()
	_, err = log.WriteString(entry.String())
	return err
}
