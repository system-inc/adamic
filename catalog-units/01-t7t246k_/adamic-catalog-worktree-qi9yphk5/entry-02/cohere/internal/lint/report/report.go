// Package report turns findings into what a person reads, and says what was covered.
//
// The coverage line is the reason this package exists. The gate cohere replaces printed a green
// checkmark over zero files linted, for days, because a missing binary produced an empty file list
// and an empty list is indistinguishable from a clean tree. Silence looked exactly like success.
//
// So every run states what it touched. A run that checked nothing must not be able to print the
// same thing as a run that checked everything and found it clean.
package report

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// Coverage is what a run actually did, as distinct from what it found.
type Coverage struct {
	FilesInProgram int
	FilesChecked   int
	RulesRun       int
	GraphWarm      bool
	Elapsed        time.Duration
}

// String renders the coverage summary that follows every verdict.
func (c Coverage) String() string {
	graph := "graph cold"
	if c.GraphWarm {
		graph = "graph warm"
	}
	return fmt.Sprintf("%s · %s · %s · %s",
		formatDuration(c.Elapsed),
		pluralize(c.FilesChecked, "file", "files"),
		pluralize(c.RulesRun, "rule", "rules"),
		graph,
	)
}

// Write renders a whole run: every finding, then the verdict and what was covered.
//
// It returns whether the run passed, so the caller sets an exit code from the same value the
// reader sees rather than recomputing it and risking the two disagreeing.
func Write(out io.Writer, diagnostics []rule.Diagnostic, coverage Coverage) bool {
	sorted := make([]rule.Diagnostic, len(diagnostics))
	copy(sorted, diagnostics)
	sort.SliceStable(sorted, func(first, second int) bool {
		firstName := fileNameOf(sorted[first])
		secondName := fileNameOf(sorted[second])
		if firstName != secondName {
			return firstName < secondName
		}
		return sorted[first].Range.Pos() < sorted[second].Range.Pos()
	})

	for _, diagnostic := range sorted {
		line, column := positionOf(diagnostic)
		fmt.Fprintf(out, "%s:%d:%d\n  %s  %s\n\n",
			fileNameOf(diagnostic), line, column,
			diagnostic.RuleName,
			diagnostic.Message.Description,
		)
	}

	passed := len(sorted) == 0
	if passed {
		fmt.Fprintf(out, "✓ cohere (%s)\n", coverage)
		return true
	}

	fmt.Fprintf(out, "✗ cohere (%s)  %s\n",
		coverage,
		pluralize(len(sorted), "finding", "findings"),
	)
	return false
}

// WriteNothingChecked is the shape a run takes when it could not do its job.
//
// This is deliberately not a clean result. A tool that cannot check must say so in the same breath
// it would have said "clean", because the two are otherwise identical from outside.
func WriteNothingChecked(out io.Writer, reason string) {
	fmt.Fprintf(out, "✗ cohere checked nothing: %s\n", reason)
}

func fileNameOf(diagnostic rule.Diagnostic) string {
	if diagnostic.SourceFile == nil {
		return "<unknown>"
	}
	return diagnostic.SourceFile.FileName().AsString()
}

// positionOf converts a byte offset into the line and column a person can navigate to.
func positionOf(diagnostic rule.Diagnostic) (int, int) {
	if diagnostic.SourceFile == nil {
		return 0, 0
	}
	text := diagnostic.SourceFile.Text()
	offset := diagnostic.Range.Pos()
	if offset < 0 || offset > len(text) {
		return 0, 0
	}
	line := 1 + strings.Count(text[:offset], "\n")
	lineStart := strings.LastIndex(text[:offset], "\n") + 1
	return line, offset - lineStart + 1
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// formatDuration reads at a glance: milliseconds under a second, seconds above.
func formatDuration(elapsed time.Duration) string {
	if elapsed < time.Second {
		return fmt.Sprintf("%dms", elapsed.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", elapsed.Seconds())
}
