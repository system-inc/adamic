package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/gitignore"
)

/*
 * Nested repositories: a library checked out inside a project, such as libraries/structure in ahra
 * and libraries/structure/libraries/nexus inside it.
 *
 * The rule (@system_cohere, 2026-10-03): writes stay inside one repository, reads reach everywhere.
 *
 * A library has no tsconfig of its own, so a run started inside it finds the project above it and
 * checks that project's program, which holds the library's code. Writing is a different question. A
 * run inside a library formats that library, rooted at its own `.git`, and writes nothing outside it,
 * so the commit it leaves belongs to the library. A run in the project never writes into a library,
 * but its check reads every library's files and reports the ones the library's own run would rewrite,
 * so drift there fails the project's check rather than staying invisible to every gate.
 *
 * Before this, no gate formatted Structure or Nexus. A run in the project skipped them as somebody
 * else's tree, and a run inside them formatted the project above, 3,086 of ahra's files and none of
 * their own: 8 of Structure's 1,683 files and 8 of Nexus's 323 had drifted with nothing saying so.
 */

// writeRepositoryRoot is the repository a run started at start may write: the nearest directory at or
// above start that is a repository of its own, when that is below the project root, and the project
// root otherwise.
//
// Only a boundary strictly inside the project counts. The project's own `.git` is its root, and a
// project that is not a repository at all (a test fixture, a tarball) writes where it always did.
func writeRepositoryRoot(start string, projectRoot string) string {
	projectRoot = filepath.Clean(projectRoot)
	for directory := filepath.Clean(start); ; directory = filepath.Dir(directory) {
		relative, err := filepath.Rel(projectRoot, directory)
		if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return projectRoot
		}
		if formatfiles.HasOwnRepository(directory) {
			return directory
		}
	}
}

// unwritableRepository says why a run writing repositoryRoot may not write fileName, as the phrase a
// note prints: "nested repository X" for a repository below the root that holds it, or "the project
// outside X" when the file lies outside the root altogether, as the project's files do for a run inside
// a library. Empty means the file is this run's to write.
func unwritableRepository(repositoryRoot string, fileName string) string {
	relative, err := filepath.Rel(filepath.Clean(repositoryRoot), filepath.Clean(fileName))
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "the project outside " + repositoryRoot
	}
	if nested := formatfiles.NestedRepositoryContaining(repositoryRoot, fileName); nested != "" {
		return "nested repository " + nested
	}
	return ""
}

// nestedDrift is one file in a nested repository that the repository's own run would rewrite, or, in
// Unchecked, one it could not read, with Reason saying why.
type nestedDrift struct {
	// Repository is the nested repository's root, relative to the project.
	Repository string
	FileName   string
	Reason     string
}

// nestedDriftCheck is what reading the nested repositories found.
type nestedDriftCheck struct {
	Drift []nestedDrift

	// Repositories and Files are what was read, so a check that found nothing still says it looked.
	Repositories int
	Files        int

	// Unchecked is each nested file the formatter was asked about and could not read, such as one that does
	// not parse, so it was neither found to drift nor found clean.
	Unchecked []nestedDrift

	// Checked is how many of those files went through the formatter: under `--format-all` every one, and
	// otherwise the ones not on their repository's record as formatted at their current bytes. The rest
	// were taken on the record's word, and the summary says so rather than counting them as read.
	Checked int

	// Walks is each repository's walk, for the caller to declare to the run cache, and Notes is what the
	// caller prints to stderr. Both are handed back rather than acted on here, because the check runs
	// beside the fix phase (see startNestedCheck) and neither the run cache nor stderr is the check's to
	// touch from another goroutine.
	Walks []formatfiles.Enumeration
	Notes []string
}

// nestedFormatRecord is a nested repository's own format record, kept in its own cache table and never in
// the parent's, so it is the same record a run started inside that repository reads and writes.
//
// Only where that repository's cohere cache exists: made by a run inside it, or at the start of a run above it
// (prepareNestedCacheDirectories). The check itself creates nothing there, since creating the directory mid-run
// would move the repository's root after the run cache started its clock, which it rightly refuses to record
// (see program.RecordRunCache's readSince), so the first check after it would never replay. A library whose
// .gitignore does not ignore .cache gets no cache, and is formatted whole each check, as before the record.
func nestedFormatRecord(repositoryRoot string) *formatRecord {
	if cacheOff {
		return formatRecordOff("the cache is off (--no-cache)")
	}
	if information, err := os.Stat(cacheDirectory(repositoryRoot)); err != nil || !information.IsDir() {
		return formatRecordOff("no cohere cache is kept in " + repositoryRoot)
	}
	return loadFormatRecord(repositoryRoot)
}

// prepareNestedCacheDirectories creates the cohere cache of every repository root declares, and every one those
// declare, that keeps none yet, at the start of a run, where prepareCacheDirectory creates the project's own.
//
// The nested check reads each library's own format record, and only where its cache exists, so a library
// nobody had run cohere in kept no record, and every check formatted every one of its files: 2,074 files and
// 0.44 CPU-s of each edit run on an ahra clone, for every fresh clone, CI machine and new laptop, every run
// (#v0etgwx). Made here, before the run cache starts its clock and before discovery lists the tree, a
// directory created moves no root a recorded run has read.
//
// Only in a repository whose own ignore rules ignore it, so a run never leaves an untracked directory in a
// library. The record in it is the library's own, the one a run inside the library reads and writes, and an
// entry still skips only bytes it proved are a fixed point, so drift in a library is reported as before.
func prepareNestedCacheDirectories(root string) {
	pending, err := declaredBelow(root, "")
	if err != nil {
		return
	}
	for len(pending) > 0 {
		relative := pending[0]
		pending = pending[1:]
		repositoryRoot := filepath.Join(root, relative)
		if !formatfiles.HasOwnRepository(repositoryRoot) {
			continue
		}
		if inner, err := declaredBelow(root, relative); err == nil {
			pending = append(pending, inner...)
		}
		if information, err := os.Stat(cacheDirectory(repositoryRoot)); err == nil && information.IsDir() {
			continue
		}
		matcher, err := gitignore.New(repositoryRoot)
		if err != nil {
			continue
		}
		if ignored, _, err := matcher.IgnoredPath(".cache/cohere/findings.gob", false); err != nil || !ignored {
			continue
		}
		// A directory that could not be made costs this library what it cost before: a whole check.
		_ = os.MkdirAll(cacheDirectory(repositoryRoot), 0o755)
	}
}

// checkNestedRepositories reads every submodule root declares, and every submodule those declare, and
// reports each file the submodule's own run would rewrite. It writes no source. The one thing it keeps is
// each submodule's format record, in that submodule's own cache table (never the parent's), only where a
// run inside the submodule already made one (see nestedFormatRecord), and only when the record learned
// something, so an edit in the project does not reformat every library file each run.
//
// Only declared submodules, read from each repository's own `.gitmodules`: a walk also finds
// repositories cloned into ignored directories (ahra holds six under `projects/`), and those belong to
// nobody running here. Each submodule is walked as a run rooted there would walk it, with its own ignore
// layers. Only the formatter is asked: a fix in a library's code is a finding the project's lint already
// reports, at its position, so repeating it here would count one problem twice.
func checkNestedRepositories(engine formatEngine, root string, formatAll bool) (nestedDriftCheck, error) {
	var check nestedDriftCheck

	pending, err := declaredBelow(root, "")
	if err != nil {
		return check, err
	}
	for len(pending) > 0 {
		relative := pending[0]
		pending = pending[1:]
		repositoryRoot := filepath.Join(root, relative)
		if !formatfiles.HasOwnRepository(repositoryRoot) {
			// Declared and not checked out: there is nothing on disk to read.
			continue
		}
		check.Repositories++

		inner, err := declaredBelow(root, relative)
		if err != nil {
			return check, err
		}
		pending = append(pending, inner...)

		enumeration, err := engine.Enumerate(repositoryRoot)
		if err != nil {
			return check, fmt.Errorf("reading nested repository %s: %w", relative, err)
		}
		check.Walks = append(check.Walks, enumeration)
		check.Files += len(enumeration.Files)

		// The repository's own record, the one a run inside it reads and writes, so a file either run saw
		// formatted at its current bytes is not formatted again here. An entry proves its bytes are a fixed
		// point of this formatter under these options, so skipping it cannot hide drift.
		record := nestedFormatRecord(repositoryRoot)
		transform := record.observe(formatTransform(engine), engine.OptionsFingerprint)
		// `--format-all` is every file, nested ones included: the record says nothing about which files it
		// reads, and only learns from what the formatter sees, as at the top level.
		unformatted := enumeration.Files
		if !formatAll {
			unformatted = record.unformatted(enumeration.Files, engine.OptionsFingerprint)
		}
		check.Checked += len(unformatted)
		for index, outcome := range formatNestedFiles(transform, unformatted) {
			if outcome.err != nil {
				return check, fmt.Errorf("formatting %s in nested repository %s: %w", unformatted[index], relative, outcome.err)
			}
			if outcome.drifted {
				check.Drift = append(check.Drift, nestedDrift{Repository: relative, FileName: unformatted[index]})
			}
			if outcome.unchecked != "" {
				check.Unchecked = append(check.Unchecked, nestedDrift{Repository: relative, FileName: unformatted[index], Reason: outcome.unchecked})
			}
		}

		// Not writing it costs the next check a format of files already formatted, never a skip.
		if err := record.saveChanged(enumeration.Files); err != nil {
			check.Notes = append(check.Notes, fmt.Sprintf("note: the format record of nested repository %s could not be written: %v", relative, firstLine(err.Error())))
		}
	}

	sort.Slice(check.Drift, func(first, second int) bool {
		return check.Drift[first].FileName < check.Drift[second].FileName
	})
	sort.Slice(check.Unchecked, func(first, second int) bool {
		return check.Unchecked[first].FileName < check.Unchecked[second].FileName
	})
	return check, nil
}

// nestedCheckResult is checkNestedRepositories' answer, carried from the goroutine it ran on.
type nestedCheckResult struct {
	check nestedDriftCheck
	err   error
}

// startNestedCheck runs checkNestedRepositories beside the project's fix phase and hands back where its
// answer will arrive. The check reads only the nested repositories and writes only their records, and a
// run that checks writes nothing at all, so the two share no state; run one after the other, the check
// was a second walk and format the whole run waited for.
func startNestedCheck(engine formatEngine, root string, formatAll bool) <-chan nestedCheckResult {
	answer := make(chan nestedCheckResult, 1)
	go func() {
		check, err := checkNestedRepositories(engine, root, formatAll)
		answer <- nestedCheckResult{check: check, err: err}
	}()
	return answer
}

// nestedOutcome is one nested file's answer from the formatter: whether its own run would rewrite it, or
// why it could not be read or formatted.
type nestedOutcome struct {
	drifted bool
	err     error

	// unchecked is why the formatter declined the file, which for a file the walk offered it means the file
	// could not be read as its language (see unparseableSkipReason).
	unchecked string
}

// formatNestedFiles asks the formatter about each file at once, through as many workers as the fix
// phase's parallel pass uses, and answers in the files' order, so the report and the first error are the
// same as one file at a time. Formatting is a parse and a print, pure per file, and no rule runs here.
func formatNestedFiles(transform edit.Transform, fileNames []string) []nestedOutcome {
	outcomes := make([]nestedOutcome, len(fileNames))
	next := make(chan int, len(fileNames))
	for index := range fileNames {
		next <- index
	}
	close(next)

	var workers sync.WaitGroup
	for range max(1, parallelFormatWorkers()) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range next {
				contents, err := os.ReadFile(fileNames[index])
				if err != nil {
					outcomes[index] = nestedOutcome{err: err}
					continue
				}
				formatted, err := transform(fileNames[index], string(contents), nil)
				switch {
				case errors.Is(err, edit.ErrSkipped):
					outcomes[index] = nestedOutcome{unchecked: strings.TrimPrefix(err.Error(), edit.ErrSkipped.Error()+": ")}
				case err != nil:
					outcomes[index] = nestedOutcome{err: err}
				default:
					outcomes[index] = nestedOutcome{drifted: formatted != string(contents)}
				}
			}
		}()
	}
	workers.Wait()
	return outcomes
}

// declaredBelow is the submodules the repository at root/relative declares, each relative to root,
// sorted so the read and its report are in a stable order.
func declaredBelow(root string, relative string) ([]string, error) {
	declared, err := declaredSubmodules(filepath.Join(root, relative))
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(declared))
	for path := range declared {
		paths = append(paths, filepath.Join(relative, filepath.FromSlash(path)))
	}
	sort.Strings(paths)
	return paths, nil
}

// nestedSummary is the check's line: what it read, split into the files the formatter checked and the
// files taken on their record's word, with what would change stated only of the checked ones. A check
// that took every file on the record's word says so, never a bare "0 would change" about files it did not
// read.
func nestedSummary(check nestedDriftCheck) string {
	line := fmt.Sprintf("nested repositories: %d read, %d files: ", check.Repositories, check.Files)
	onRecord := check.Files - check.Checked
	switch {
	case check.Checked == 0 && onRecord > 0:
		line += fmt.Sprintf("none checked by the formatter, all %d taken on their record's word as formatted at their current bytes", onRecord)
	case onRecord == 0:
		line += fmt.Sprintf("%d checked by the formatter, %d would change under their own run", check.Checked, len(check.Drift))
	default:
		line += fmt.Sprintf("%d checked by the formatter, %d would change under their own run; %d taken on their record's word as formatted at their current bytes",
			check.Checked, len(check.Drift), onRecord)
	}
	if len(check.Unchecked) > 0 {
		line += fmt.Sprintf("; %d the formatter could not read", len(check.Unchecked))
	}
	return line
}

// printNestedDrift reports each drifted file as a finding, in the shape every other finding prints,
// naming the repository whose own run fixes it.
func printNestedDrift(out io.Writer, check nestedDriftCheck) {
	for _, drift := range check.Drift {
		message := fmt.Sprintf("the formatter would rewrite this file in nested repository %s, which a run here never writes: run cohere there", drift.Repository)
		printFinding(out,
			runFinding{Path: drift.FileName, Line: 1, Column: 1, Severity: "error", Rule: "format", MessageID: "nested-drift", Message: message},
			fmt.Sprintf("%s:1:1 - %s [format/nested-drift]\n", drift.FileName, message))
	}
}
