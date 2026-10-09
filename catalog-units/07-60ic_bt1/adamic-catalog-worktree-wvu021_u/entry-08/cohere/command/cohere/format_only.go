package main

import (
	"fmt"
	"io"

	"github.com/system-inc/cohere/internal/edit"
)

// `--format-only`: the fix phase narrowed to formatting, for the commit gate (#5eww5r8).
//
// The gate's question is one: would formatting change any file, in this repository or the nested
// repositories it reads? `--no-fix --format-all` answered it only beside the type and lint phases, so its
// exit was 1 on any lint finding, and www, which carries ten known ones, could not use it as a format gate
// at all. Gates scraped two human lines instead, which a rewording would have read as a confident 0.
//
// So the exit under `--no-fix --format-only` is formatting's alone. A file formatting would change is a
// finding, root and nested alike. So is a file the formatter could not read, because a gate that passes
// over a file nobody checked is the failure this exists to prevent: a walk that failed, a file that does
// not parse, a printer that broke or never settled. Skips that are not about the file, such as one
// outside the named paths or of a type the formatter does not handle, are not findings: those files
// were never the run's to check.

// uncheckedFile is a file a format-only run could not check, and why.
type uncheckedFile struct {
	FileName string
	Reason   string
}

// formatOnlyUnchecked is every file the fix phase could not check for formatting, from its summary and its
// scope, sorted by name.
func formatOnlyUnchecked(summary edit.Summary, scope formatScope) []uncheckedFile {
	unchecked := []uncheckedFile{}
	if scope.failure != nil {
		// No file to name: the walk that would have named them failed, which is the finding.
		unchecked = append(unchecked, uncheckedFile{Reason: fmt.Sprintf("the format walk failed, so no file was checked: %v", scope.failure)})
	}
	for _, file := range summary.NotTransformed {
		if file.Failed || file.Reason == unparseableSkipReason {
			unchecked = append(unchecked, uncheckedFile{FileName: file.FileName, Reason: file.Reason})
		}
	}
	// A file the engine could not process at all never reached the formatter, and one whose formatted text
	// did not parse was refused. Either way nothing about its formatting is known.
	for _, fileName := range summary.FilesFailed {
		unchecked = append(unchecked, uncheckedFile{FileName: fileName, Reason: "the file could not be processed, so the formatter never read it"})
	}
	for _, refused := range summary.FilesRefused {
		unchecked = append(unchecked, uncheckedFile{FileName: refused.FileName, Reason: edit.ReasonParseFailure})
	}
	return unchecked
}

// nestedUnchecked is the nested check's unreadable files, in the same shape, naming their repository.
func nestedUnchecked(check nestedDriftCheck) []uncheckedFile {
	unchecked := make([]uncheckedFile, 0, len(check.Unchecked))
	for _, file := range check.Unchecked {
		unchecked = append(unchecked, uncheckedFile{FileName: file.FileName, Reason: fmt.Sprintf("in nested repository %s, %s", file.Repository, file.Reason)})
	}
	return unchecked
}

// printUnchecked reports each file as a finding, in the shape every other finding prints.
func printUnchecked(out io.Writer, unchecked []uncheckedFile) {
	for _, file := range unchecked {
		if file.FileName == "" {
			printFinding(out,
				runFinding{Severity: "error", Rule: "format", MessageID: "unchecked", Message: "the formatter could not check this run: " + file.Reason},
				fmt.Sprintf("cohere: the formatter could not check this run: %s [format/unchecked]\n", file.Reason))
			continue
		}
		printFinding(out,
			runFinding{Path: file.FileName, Line: 1, Column: 1, Severity: "error", Rule: "format", MessageID: "unchecked", Message: "the formatter could not check this file: " + file.Reason},
			fmt.Sprintf("%s:1:1 - the formatter could not check this file: %s [format/unchecked]\n", file.FileName, file.Reason))
	}
}
