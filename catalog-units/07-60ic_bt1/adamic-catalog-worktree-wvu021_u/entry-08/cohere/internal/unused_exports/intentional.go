package unused_exports

import (
	"strings"
)

// intentionalMarker is what a person writes to say "yes, I know, keep it".
//
// The spelling follows the `cohere-disable` directive this tool already understands, because a
// codebase with two suppression vocabularies makes people guess and guessing wrong is silent.
//
// # Why a marker exists at all
//
// Some exports are unused on purpose: a public API surface, something held for an external consumer,
// work parked deliberately. Without a way to record that decision, the report is re-derived from
// scratch on every reading — the same twenty lines, the same four already-judged, and the reader has
// to remember which four. A report that cannot be answered gets read twice and abandoned, so the
// marker is not a convenience, it is what makes a second reading cheaper than the first.
//
// It changes the summary rather than hiding the line: `18 unused, 4 declared intentional` is a
// different statement from `18 unused`, and collapsing the two would lose the fact that somebody
// looked. A marked export stays countable and stays visible under `--unused-all`.
const intentionalMarker = "cohere-keep"

// intentionalReason is the text a person wrote after the marker, if any.
//
// A bare marker is accepted, because demanding a reason produces reasons like "needed" that cost
// keystrokes and carry nothing. But the reason is reported where present, since the whole value of
// the marker on a second reading is the sentence that says why.
type intentionalReason struct {
	Present bool
	Text    string
}

// findIntentionalMarker looks for the marker in the comment text immediately preceding a
// declaration.
//
// It reads raw source text rather than the comment scanner because the range handed in is already
// the declaration's leading trivia — the compiler puts every preceding comment there — and reaching
// for a second scanner would answer the same question a second way.
func findIntentionalMarker(text string) intentionalReason {
	// Only the comment lines are searched. The declaration's own source can contain the marker as
	// part of an identifier or a string, and reading that as a directive would suppress a finding
	// nobody meant to suppress.
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !isCommentLine(trimmed) {
			continue
		}
		index := strings.Index(trimmed, intentionalMarker)
		if index < 0 {
			continue
		}

		// The marker must be a whole word on its trailing edge, so `cohere-keeper` and
		// `cohere-keeping` do not count. The same discipline the suppression parser applies to
		// `eslint-disabled`, and for the same reason: a prefix match silently accepts a word nobody
		// meant as a directive.
		after := trimmed[index+len(intentionalMarker):]
		if after != "" && !isMarkerBoundary(rune(after[0])) {
			continue
		}

		reason := strings.TrimSpace(after)
		reason = strings.TrimPrefix(reason, ":")
		reason = strings.TrimSpace(reason)
		reason = strings.TrimSpace(strings.TrimSuffix(reason, "*/"))
		return intentionalReason{Present: true, Text: reason}
	}
	return intentionalReason{}
}

// isMarkerBoundary reports whether a rune ends the marker word.
func isMarkerBoundary(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return false
	case r == '-', r == '_':
		return false
	}
	return true
}

// isCommentLine reports whether a trimmed source line is a comment, which is the only place a
// marker is honoured.
func isCommentLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "//") ||
		strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*")
}
