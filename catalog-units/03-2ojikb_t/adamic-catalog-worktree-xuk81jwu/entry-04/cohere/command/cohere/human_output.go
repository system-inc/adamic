package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

// The human view of a run, top to bottom: the files cohere rewrote, then the findings, then the footer.
//
//	🪄💅 app/os/SessionRow.tsx       prefer-const ×2, prefer-nullish-coalescing
//	  💅 modules/pensieve/Recall.ts
//	🪄   modules/data/DataSync.ts     no-useless-default-assignment
//	app/os/Session.ts:12:5 error nexus/consistency-no-abbreviated-identifier `ctx` is ...
//	✗ ☠️ 0.8s • 1 finding • 3 cohered (480 rules • 3 checked • 3,923 cached)

// textStyle colors text for a terminal, and does nothing when the output is not one or NO_COLOR is set,
// so a log, a pipe or a reader who asked for no color gets no escape codes at all.
type textStyle struct {
	color bool
}

// styleFor is the style for a stream: color only on a terminal, and never under NO_COLOR, whose
// convention is that any value, empty included, turns color off.
func styleFor(stream *os.File) textStyle {
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return textStyle{}
	}
	information, err := stream.Stat()
	if err != nil {
		return textStyle{}
	}
	return textStyle{color: information.Mode()&os.ModeCharDevice != 0}
}

func (style textStyle) wrap(text, open, close string) string {
	if !style.color || text == "" {
		return text
	}
	return open + text + close
}

func (style textStyle) bold(text string) string { return style.wrap(text, "\x1b[1m", "\x1b[22m") }
func (style textStyle) dim(text string) string  { return style.wrap(text, "\x1b[2m", "\x1b[22m") }
func (style textStyle) red(text string) string  { return style.wrap(text, "\x1b[31m", "\x1b[39m") }

// changedFilesShown is how many rewritten files the default view lists before saying how many more.
const changedFilesShown = 20

// changedFileLines lists the files the run rewrote, one per line: a 🪄 slot and a 💅 slot, each a blank
// of the same width when absent, so the paths start in one column; the paths padded to the longest, so
// the rules start in one column; and for a fixed file, the rules that fixed it, with a count past one.
// Past changedFilesShown it says how many more and where to see them.
func changedFileLines(files []changedFile, style textStyle) []string {
	sorted := append([]changedFile(nil), files...)
	sort.Slice(sorted, func(left, right int) bool { return sorted[left].Path < sorted[right].Path })
	shown := sorted
	if len(shown) > changedFilesShown {
		shown = shown[:changedFilesShown]
	}

	width := 0
	for _, file := range shown {
		width = max(width, utf8.RuneCountInString(file.Path))
	}

	// A glyph takes two columns in a terminal, so its blank is two spaces.
	const blank = "  "
	lines := make([]string, 0, len(shown)+1)
	for _, file := range shown {
		fixed, formatted := blank, blank
		if file.Fixed {
			fixed = "🪄"
		}
		if file.Formatted {
			formatted = "💅"
		}
		line := fixed + formatted + " " + file.Path
		if rules := fixingRules(file.FixedBy); rules != "" {
			line += strings.Repeat(" ", width-utf8.RuneCountInString(file.Path)) + " " + style.dim(rules)
		}
		lines = append(lines, line)
	}
	if more := len(sorted) - len(shown); more > 0 {
		lines = append(lines, fmt.Sprintf("… %d more · --verbose for all", more))
	}
	return lines
}

// fixingRules is the rules that fixed a file, the most frequent first and then by name, with ×N past one.
func fixingRules(fixedBy map[string]int) string {
	rules := make([]string, 0, len(fixedBy))
	for rule := range fixedBy {
		rules = append(rules, rule)
	}
	sort.Slice(rules, func(left, right int) bool {
		if fixedBy[rules[left]] != fixedBy[rules[right]] {
			return fixedBy[rules[left]] > fixedBy[rules[right]]
		}
		return rules[left] < rules[right]
	})
	for index, rule := range rules {
		if count := fixedBy[rule]; count > 1 {
			rules[index] = fmt.Sprintf("%s ×%d", rule, count)
		}
	}
	return strings.Join(rules, ", ")
}

// findingLine is one finding as the human view prints it: `path:line:col severity rule message`, the
// severity and rule dim, an error's severity red.
func findingLine(finding runFinding, style textStyle) string {
	severity := style.dim(finding.Severity)
	if finding.Severity == "error" {
		severity = style.red(finding.Severity)
	}
	// The rule alone, without which of its messages this is: a person reads the message, and `--json`
	// carries the message id for a program. A finding about no one file, such as a formatter that could not
	// run at all, has no location to lead with.
	if finding.Path == "" {
		return fmt.Sprintf("%s %s %s", severity, style.dim(finding.Rule), singleLineDescription(finding.Message))
	}
	return fmt.Sprintf("%s:%d:%d %s %s %s",
		finding.Path, finding.Line, finding.Column, severity, style.dim(finding.Rule), singleLineDescription(finding.Message))
}
