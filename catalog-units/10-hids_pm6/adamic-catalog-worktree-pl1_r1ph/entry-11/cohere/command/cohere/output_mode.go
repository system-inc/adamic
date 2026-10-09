package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/release/packaging"
)

// How a run prints. The default is the human view: the files cohere rewrote, the findings, and one
// footer line. `--verbose` is everything a run printed before the footer existed, then the footer.
// `--json` is newline-delimited JSON for a program to read. The three read the same runSummary.
type outputMode int

const (
	outputHuman outputMode = iota
	outputVerbose
	outputJSON
)

// outputSettings is the run's choice of view, set once from the flags and the project's settings.
type outputSettings struct {
	Mode outputMode
	// Phases puts where the time went first inside the footer's parentheses.
	Phases bool
	// Style colors the human view on a terminal.
	Style textStyle
}

var activeOutput = outputSettings{Mode: outputHuman}

// accountOutput is where a line that accounts for the run goes: the phase lines, the coverage block, the
// scopes and notes. Under `--verbose` that is out; in the human view and under `--json` it is nowhere,
// since the footer says what a reader needs and `--json` prints nothing a program did not ask for.
func accountOutput(out io.Writer) io.Writer {
	if activeOutput.Mode == outputVerbose {
		return out
	}
	return io.Discard
}

// activeSummary is the run's summary, filled as the phases finish and rendered at the end.
var activeSummary runSummary

// printFinding prints one finding in the view the run asked for: as it always printed under `--verbose`,
// as `path:line:col severity rule message` in the human view, as a line of JSON under `--json`.
func printFinding(out io.Writer, finding runFinding, verboseLine string) {
	switch activeOutput.Mode {
	case outputVerbose:
		fmt.Fprint(out, verboseLine)
	case outputJSON:
		writeJSONLine(out, findingAsJSON(finding))
	default:
		fmt.Fprintln(out, findingLine(finding, activeOutput.Style))
	}
}

// compilerFinding is a type error as a runFinding: its rule is its TypeScript code.
func compilerFinding(diagnostic *ast.Diagnostic) runFinding {
	finding := runFinding{
		Severity: "error",
		Rule:     fmt.Sprintf("TS%d", diagnostic.Code()),
		Message:  singleLineDescription(diagnosticMessage(diagnostic)),
	}
	if sourceFile := diagnostic.File(); sourceFile != nil {
		line, character := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, diagnostic.Loc().Pos())
		finding.Path, finding.Line, finding.Column = sourceFile.FileName().AsString(), line+1, character+1
	}
	return finding
}

// ruleFinding is a rule's diagnostic as a runFinding, with the severity the settings give the rule for
// its file. A rule the settings do not name for the file reports as an error, as it fails the run.
func ruleFinding(diagnostic rule.Diagnostic, lintConfig *configuration.Config) runFinding {
	finding := runFinding{
		Severity:  "error",
		Rule:      diagnostic.RuleName,
		MessageID: diagnostic.Message.Id,
		Message:   singleLineDescription(diagnostic.Message.Description),
	}
	if sourceFile := diagnostic.SourceFile; sourceFile != nil {
		finding.Path, finding.Line, finding.Column = diagnostic.Location()
		if lintConfig != nil {
			if setting, named := lintConfig.Resolve(sourceFile.FileName().AsString()).Rules[diagnostic.RuleName]; named &&
				setting.Severity == configuration.SeverityWarn {
				finding.Severity = "warn"
			}
		}
	}
	return finding
}

// writeRunEnd ends the run's output with what the view calls for: under `--verbose` the phase line and its
// account, then the footer; in the human view the footer; under `--json` the summary line. The footer is
// written as describing this invocation, since its time is this run's, so the run cache never replays it
// as recorded; a replay renders its own from the stored summary.
func writeRunEnd(report *pipelineReport, out io.Writer) {
	summary := activeSummary
	summary.Phases = report.records
	summary.Graph = report.graph
	if !report.processStart.IsZero() {
		summary.Total = time.Since(report.processStart)
	}
	summary.Cache.Off = report.cacheOff
	if session := activeRunCache; session != nil {
		summary.Cache.MissReason = session.missed
		summary.Cache.FindingsMissReason = session.findingsMissed
		if summary.Cache.FindingsMissReason == "" {
			summary.Cache.FindingsMissReason = session.findings.Misses()
		}
	}
	// Decided here, once the phases are known, and kept: a replay prints this count rather than judging the
	// skips again against phases it did not run.
	summary.Gaps.RulesSkippingEverything = summary.uncoveredSkips()
	summary.Gaps.ModifiedBuild = release.Current().SourceTreeModified || report.engineSourceTreeModified
	// The phase line's sentence points at notes only --verbose prints, so the footer says the gap itself:
	// each path that finds one sets Gaps with its count. A gap nothing counted still shows, saying where
	// to look, so no run falls short in silence.
	if report.incompleteBeyondPhases != "" && summary.Gaps.Unread == "" && summary.Gaps.CrashedFiles == 0 {
		summary.Gaps.Unread = "the run fell short of checking everything; --verbose says where"
	}
	if report.filesInScope > 0 && report.filesInScope < report.filesInProgram {
		summary.Gaps.ProgramFiles = report.filesInProgram
	}
	// In a repository with several projects each is its own run, told its path, the root's as ".".
	if label := os.Getenv(projectLabelVariable); label != "" {
		summary.Label = label
	}
	recordRunSummary(summary)

	switch activeOutput.Mode {
	case outputVerbose:
		report.Write(out)
		if line := summary.Cache.missLine(); line != "" {
			fmt.Fprintln(invocationOutput(out), line)
		}
		fmt.Fprintln(invocationOutput(out), footer(summary, activeOutput.Style, footerOptions{Phases: activeOutput.Phases, Verbose: true}))
	case outputJSON:
		writeJSONLine(invocationOutput(out), summaryAsJSON(summary))
	default:
		fmt.Fprintln(invocationOutput(out), footer(summary, activeOutput.Style, footerOptions{Phases: activeOutput.Phases}))
	}
}

// changedFilesFrom is the edit engine's account of what it rewrote, as the summary keeps it: each path
// relative to the project root, whether fixes landed, whether the formatter changed it, and the fixes by
// rule. The engine names the formatter among a file's changers as "format".
func changedFilesFrom(files []edit.ChangedFile, root string) []changedFile {
	changed := make([]changedFile, 0, len(files))
	for _, file := range files {
		path := file.FileName
		if relative, err := filepath.Rel(root, file.FileName); err == nil && !strings.HasPrefix(relative, "..") {
			path = relative
		}
		entry := changedFile{Path: path, FixedBy: file.FixesByRule}
		for _, changer := range file.Changers {
			if changer == "format" {
				entry.Formatted = true
			} else {
				entry.Fixed = true
			}
		}
		changed = append(changed, entry)
	}
	return changed
}

// printChangedFiles lists the files the fix phase rewrote, above the findings, in the human view, and as
// lines of JSON under `--json`. `--verbose` has the fix phase's own line for it instead.
func printChangedFiles(out io.Writer, changed []changedFile) {
	switch activeOutput.Mode {
	case outputJSON:
		for _, line := range changedFilesAsJSON(changed) {
			writeJSONLine(out, line)
		}
	case outputHuman:
		for _, line := range changedFileLines(changed, activeOutput.Style) {
			fmt.Fprintln(out, line)
		}
	}
}
