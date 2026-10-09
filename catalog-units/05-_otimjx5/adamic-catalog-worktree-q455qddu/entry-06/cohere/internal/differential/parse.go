package differential

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// The two gates print findings in two shapes, and both are parsed here rather than by asking
// either one for JSON.
//
// That choice is deliberate and it costs something, so it is worth stating. Parsing the human
// format means the harness reads exactly what a human reads, which is the artifact both sides
// actually ship. Asking cohere for a machine format would mean adding an output mode that exists
// only for this harness, and a format nobody looks at is a format that drifts from the one
// everybody looks at until the harness is measuring a surface that no longer matches the tool.
//
// The risk is the opposite one: a format change breaks parsing. That is why parse failures are
// counted and surfaced rather than skipped. A line that does not parse is not nothing — it is a
// finding the harness cannot see, and silently dropping it is precisely how a diff comes back
// clean because it failed to read one side.

var (
	// cohereFindingPattern matches cohere's rule output:
	//   /abs/path/File.ts:240:16 - Message text here [rule-name/messageId]
	cohereFindingPattern = regexp.MustCompile(`^(.+?):(\d+):(\d+) - (.*) \[([^\]]+)\]\s*$`)

	// cohereTypeDiagnosticPattern matches cohere's type output, which the lint-only harness should
	// recognize as "not a rule finding" rather than as an unparsed line:
	//   /abs/path/File.ts:10:5 - error TS2304: Cannot find name 'x'.
	cohereTypeDiagnosticPattern = regexp.MustCompile(`^(.+?):(\d+):(\d+) - error TS(\d+): (.*)$`)

	// gateFindingPattern matches RunCachedOxlint.ts output:
	//   relative/path/File.tsx:198:1: error structure(rule-name): Message text here
	gateFindingPattern = regexp.MustCompile(`^(.+?):(\d+):(\d+): (\w+) ([^:]+): (.*)$`)

	// gateFindingWithoutPositionPattern matches the same without a span, which oxlint emits when a
	// diagnostic has no labels. Dropping these would silently lose whole-file findings.
	gateFindingWithoutPositionPattern = regexp.MustCompile(`^(.+?): (\w+) ([^:]+): (.*)$`)
)

// ParseResult is what a parse produced, including what it could not read.
type ParseResult struct {
	Findings []Finding

	// UnparsedLines are lines that looked like findings but did not match. Surfaced rather than
	// dropped: an unparsed finding is invisible to the diff, which turns a format change into a
	// false clean.
	UnparsedLines []string

	// SummaryLines are the coverage and verdict lines, kept so the caller can read counts out of
	// them rather than recomputing.
	SummaryLines []string
}

// ParseCohere reads cohere's lint output into findings.
//
// root is the directory paths are made relative to, so cohere's absolute paths and the gate's
// relative ones compare.
func ParseCohere(output string, root string) (ParseResult, error) {
	result := ParseResult{Findings: []Finding{}, UnparsedLines: []string{}, SummaryLines: []string{}}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimRight(line, "\r")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}

		// Coverage notes and the verdict line are indented or prefixed, never findings.
		if isVerifySummaryLine(trimmed) {
			result.SummaryLines = append(result.SummaryLines, strings.TrimSpace(trimmed))
			continue
		}

		// A type diagnostic is a real finding but not a rule finding, and this harness compares
		// rules. Recognized explicitly so it does not land in UnparsedLines and read as a parse
		// failure.
		if cohereTypeDiagnosticPattern.MatchString(trimmed) {
			continue
		}

		match := cohereFindingPattern.FindStringSubmatch(trimmed)
		if match == nil {
			result.UnparsedLines = append(result.UnparsedLines, trimmed)
			continue
		}

		line, column, err := parsePosition(match[2], match[3])
		if err != nil {
			result.UnparsedLines = append(result.UnparsedLines, trimmed)
			continue
		}

		result.Findings = append(result.Findings, Finding{
			File:    relativeTo(root, match[1]),
			Line:    line,
			Column:  column,
			Rule:    NormalizeRuleName(match[5]),
			Message: strings.TrimSpace(match[4]),
		})
	}

	return result, nil
}

// ParseGate reads RunCachedOxlint.ts output into findings.
func ParseGate(output string, root string) (ParseResult, error) {
	result := ParseResult{Findings: []Finding{}, UnparsedLines: []string{}, SummaryLines: []string{}}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimRight(line, "\r")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}

		if isGateSummaryLine(trimmed) {
			result.SummaryLines = append(result.SummaryLines, strings.TrimSpace(trimmed))
			continue
		}

		if match := gateFindingPattern.FindStringSubmatch(trimmed); match != nil {
			line, column, err := parsePosition(match[2], match[3])
			if err != nil {
				result.UnparsedLines = append(result.UnparsedLines, trimmed)
				continue
			}
			result.Findings = append(result.Findings, Finding{
				File:    relativeTo(root, match[1]),
				Line:    line,
				Column:  column,
				Rule:    NormalizeRuleName(match[5]),
				Message: strings.TrimSpace(match[6]),
			})
			continue
		}

		if match := gateFindingWithoutPositionPattern.FindStringSubmatch(trimmed); match != nil {
			// A finding with no span is reported at the top of the file rather than dropped. Position
			// 0:0 marks it as span-less so it never accidentally matches a real 1:1 finding.
			result.Findings = append(result.Findings, Finding{
				File:    relativeTo(root, match[1]),
				Line:    0,
				Column:  0,
				Rule:    NormalizeRuleName(match[3]),
				Message: strings.TrimSpace(match[4]),
			})
			continue
		}

		result.UnparsedLines = append(result.UnparsedLines, trimmed)
	}

	return result, nil
}

// isVerifySummaryLine recognizes cohere's coverage and verdict output.
func isVerifySummaryLine(line string) bool {
	if strings.HasPrefix(line, "  ") {
		return true
	}
	// `phases:` says which phases ran and which were skipped, so a lint-only run cannot be mistaken
	// for a full one. Added to cohere after this parser was written, and caught by the harness's
	// refusal to compare when a line does not parse rather than by anyone noticing — which is the
	// whole argument for that refusal being an error instead of a note.
	for _, prefix := range []string{"graph built in ", "lint: ", "types: ", "fix: ", "format: ", "phases: "} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

// isGateSummaryLine recognizes the runner's own progress and stats output.
func isGateSummaryLine(line string) bool {
	return strings.HasPrefix(line, "[oxlint-files]") ||
		strings.HasPrefix(line, "[phase]") ||
		strings.HasPrefix(line, "Found ") ||
		strings.HasPrefix(line, "Finished in ")
}

func parsePosition(lineText string, columnText string) (int, int, error) {
	line, err := strconv.Atoi(lineText)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing line number %q: %w", lineText, err)
	}
	column, err := strconv.Atoi(columnText)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing column number %q: %w", columnText, err)
	}
	return line, column, nil
}

// relativeTo normalizes a path against the tree root so both gates' paths compare.
//
// The two sides genuinely differ here — cohere prints absolute paths and the gate prints paths
// relative to the project — and this is the single most likely place for a silent total mismatch,
// because a path difference makes every finding look one-sided and the diff still comes back
// well-formed. It is the reason the controls exist.
func relativeTo(root string, path string) string {
	cleaned := filepath.Clean(strings.TrimSpace(path))
	if root == "" {
		return cleaned
	}
	if relative, err := filepath.Rel(filepath.Clean(root), cleaned); err == nil && !strings.HasPrefix(relative, "..") {
		return relative
	}
	return cleaned
}
