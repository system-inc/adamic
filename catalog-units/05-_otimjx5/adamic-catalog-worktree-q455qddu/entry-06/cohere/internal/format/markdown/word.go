package markdown

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/word.js. printWordLegacy is mdx's and not ported.

// isFakeSetextHeader is /^(?:=+|-+)$/.test(text).
func isFakeSetextHeader(text string) bool {
	return text != "" && (strings.Trim(text, "=") == "" || strings.Trim(text, "-") == "")
}

func printWord(path *astPath, options *options) doc.Doc {
	node := currentNode(path)
	emphasisOrStrong, found := path.FindAncestor(func(node *Node) bool {
		return node.NodeType == "emphasis" || node.NodeType == "strong"
	})
	text := node.Value
	if !found {
		if settingsOf(options).proseWrap == "preserve" &&
			parentNode(path).NodeType == "sentence" &&
			isFakeSetextHeader(text) &&
			isNewLine(previousNode(path)) &&
			(path.IsLast() || isNewLine(nextNode(path))) {
			// escape indented pseudo setext header, e.g. `Previous line↵␣␣␣␣===`
			return doc.Text(`\` + text)
		}

		return doc.Text(text)
	}

	// escape leading `*` or `_` if it's the first character in an emphasis/strong
	if grandparent, _ := path.Grandparent(); path.IsFirst() &&
		(strings.HasPrefix(text, "*") || strings.HasPrefix(text, "_")) &&
		printingCallParentIsFirst(path) &&
		grandparent == emphasisOrStrong {
		text = `\` + text
	}

	// escape internal `*` or `_` that can open or close emphasis/strong
	var previousValue, nextValue string
	if previous := previousNode(path); previous != nil {
		previousValue = previous.Value
	}
	if next := nextNode(path); next != nil {
		nextValue = next.Value
	}
	return doc.Text(escapeDelimiterRuns(text, previousValue, nextValue))
}

// printingCallParentIsFirst is path.callParent(() => path.isFirst): the callback ignores its argument and
// reads the same path, which callParent has moved to the parent, so it asks whether the parent is first.
func printingCallParentIsFirst(path *astPath) bool {
	return printing.CallParent(path, func(*astPath) bool { return path.IsFirst() }, 0)
}

// escapeDelimiterRuns is text.replaceAll(/(\\+|^|.)(\*+|_+)($|.)/g, ...). The pattern has no u flag, so
// it runs over UTF-16 units, and `.` matches any unit but a line terminator; the matcher below tries the
// alternatives in the regex's order, backtracking the same way.
func escapeDelimiterRuns(text string, previousValue string, nextValue string) string {
	// Without a `*` or `_` nothing matches, and the text comes back as it went in: decoding its units
	// rebuilds it exactly when it is valid UTF-8, so only then is the walk skipped (#vbjv3d6).
	if !strings.ContainsAny(text, "*_") && utf8.ValidString(text) {
		return text
	}
	units := utf16.Encode([]rune(text))
	var output []uint16
	position := 0
	for position < len(units) {
		precedingEnd, runEnd, followingEnd, matched := matchDelimiterRun(units, position)
		if !matched {
			output = append(output, units[position])
			position++
			continue
		}

		preceding := units[position:precedingEnd]
		delimiterRun := units[precedingEnd:runEnd]
		following := units[runEnd:followingEnd]
		match := units[position:followingEnd]
		position = followingEnd

		if allBackslashes(preceding) && len(preceding)%2 == 1 {
			// already escaped
			output = append(output, match...)
			continue
		}

		precedingUnit := rune(-1)
		if len(preceding) > 0 {
			precedingUnit = rune(preceding[len(preceding)-1])
		} else {
			precedingUnit = lastUnit(previousValue)
		}
		followingUnit := rune(-1)
		if len(following) > 0 {
			followingUnit = rune(following[0])
		} else {
			followingUnit = firstUnit(nextValue)
		}
		if canOpenOrCloseStrongOrEmphasis(precedingUnit, rune(delimiterRun[0]), followingUnit) {
			output = append(output, preceding...)
			output = append(output, '\\')
			output = append(output, delimiterRun...)
			output = append(output, following...)
			continue
		}
		output = append(output, match...)
	}
	return string(utf16.Decode(output))
}

// allBackslashes is [...preceding].every((c) => c === "\\"), true for an empty preceding.
func allBackslashes(units []uint16) bool {
	for _, unit := range units {
		if unit != '\\' {
			return false
		}
	}
	return true
}

// isLineTerminatorUnit is what `.` refuses without the s flag.
func isLineTerminatorUnit(unit uint16) bool {
	return unit == '\n' || unit == '\r' || unit == 0x2028 || unit == 0x2029
}

// matchDelimiterRun tries /(\\+|^|.)(\*+|_+)($|.)/ at position, returning where each group ends.
func matchDelimiterRun(units []uint16, position int) (int, int, int, bool) {
	// (\*+|_+)($|.) from start, the run greedy and backtracking.
	rest := func(start int) (int, int, bool) {
		if start >= len(units) {
			return 0, 0, false
		}
		marker := units[start]
		if marker != '*' && marker != '_' {
			return 0, 0, false
		}
		runEnd := start
		for runEnd < len(units) && units[runEnd] == marker {
			runEnd++
		}
		for end := runEnd; end > start; end-- {
			if end == len(units) {
				return end, end, true
			}
			if !isLineTerminatorUnit(units[end]) {
				return end, end + 1, true
			}
		}
		return 0, 0, false
	}

	// \\+, greedy and backtracking.
	backslashEnd := position
	for backslashEnd < len(units) && units[backslashEnd] == '\\' {
		backslashEnd++
	}
	for end := backslashEnd; end > position; end-- {
		if runEnd, followingEnd, matched := rest(end); matched {
			return end, runEnd, followingEnd, true
		}
	}

	// ^
	if position == 0 {
		if runEnd, followingEnd, matched := rest(0); matched {
			return 0, runEnd, followingEnd, true
		}
	}

	// .
	if position < len(units) && !isLineTerminatorUnit(units[position]) {
		if runEnd, followingEnd, matched := rest(position + 1); matched {
			return position + 1, runEnd, followingEnd, true
		}
	}

	return 0, 0, 0, false
}

// canOpenOrCloseStrongOrEmphasis takes one UTF-16 unit on either side, -1 where upstream's is undefined.
// Upstream returns null, which is falsy, when it cannot determine.
func canOpenOrCloseStrongOrEmphasis(preceding rune, indicator rune, following rune) bool {
	if preceding < 0 || following < 0 {
		return false // cannot determine
	}

	// https://spec.commonmark.org/0.31.2/#emphasis-and-strong-emphasis
	followedByWhitespace := inClass(spaceSeparatorRanges, following)
	precededByWhitespace := inClass(spaceSeparatorRanges, preceding)
	followedByPunctuation := isPunctuationUnit(following)
	precededByPunctuation := isPunctuationUnit(preceding)

	isLeftFlanking := !followedByWhitespace &&
		(!followedByPunctuation || (followedByPunctuation && (precededByWhitespace || precededByPunctuation)))
	isRightFlanking := !precededByWhitespace &&
		(!precededByPunctuation || (precededByPunctuation && (followedByWhitespace || followedByPunctuation)))

	if indicator == '*' {
		return isLeftFlanking || isRightFlanking
	}

	if isLeftFlanking {
		return !isRightFlanking || precededByPunctuation
	}

	if isRightFlanking {
		return !isLeftFlanking || followedByPunctuation
	}

	return false
}
