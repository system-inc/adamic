package nexus

// The candidate gate, expressed as lookups rather than as a regular expression.
//
// This decides whether a name is worth spending the rest of the rule on, and it runs once per
// identifier in the tree. Measured on the ahra tree that is 667,947 calls, and the regular
// expression form cost 1,830ms of a 2,825ms total, 64.8% of all rule time. Benchmarked against the
// real pattern: 2,247 ns/op for the regex, 29.68 ns/op for this shape, a factor of 76.
//
// The regex is not badly written and the port was faithful. Go's regexp is RE2, which does not
// backtrack, and a ninety-alternative alternation has different constant factors there than under
// V8. **A regex that is cheap in JavaScript is not automatically cheap in Go**, and a per-node gate
// is where that difference gets multiplied by two million. Nothing about the rule's logic changed.
//
// Every alternative in that pattern was a literal word, or a literal word plus a case boundary, so
// there was no backtracking behavior to preserve and a map lookup answers the same question. The
// pattern itself is gone: it was kept as the specification this gate was tested against, and once
// the vocabulary moved into policy/Abbreviations.json a pattern derived from the same file would only have
// compared the file to itself.
//
// The direction of error is the whole discipline here: this must be a **superset** of what the
// vocabulary reports. A false positive costs one cheap traversal that reports nothing. A false
// negative silences the rule for that spelling, and a rule that stops firing looks exactly like a
// clean run. `abbreviation_gate_test.go` asserts the superset property by running the vocabulary
// itself, with no gate in front of it, over 82,367 distinct identifiers from the ahra tree, and
// requiring the gate to admit every name it reports. Two earlier attempts at this gate both lost
// `elapsedMs`, and a hand-written list does not contain the case its author misread.

// abbreviationGate is the vocabulary turned into the lookups below, one set per form.
//
// Derived from policy/Abbreviations.json rather than written out, so a word added to the file is admitted
// by the gate in the same edit. When these were four hand-written maps beside a hand-written list,
// the two could disagree, and a word the gate did not admit was a word the rule never judged.
type abbreviationGate struct {
	wholeWords map[string]bool
	prefixes   map[string]bool
	suffixes   map[string]bool
	segments   map[string]bool
	// millisecond is whether any entry uses the millisecond matcher, which has its own arm.
	millisecond bool
	// The longest word in each form bounds that form's scan, so a long name does not cost work
	// proportional to its length times the table size.
	longestPrefix  int
	longestSuffix  int
	longestSegment int
}

var gate = buildAbbreviationGate(vocabulary)

func buildAbbreviationGate(vocabulary *abbreviationVocabulary) abbreviationGate {
	built := abbreviationGate{
		wholeWords: map[string]bool{},
		prefixes:   map[string]bool{},
		suffixes:   map[string]bool{},
		segments:   map[string]bool{},
	}
	for name := range vocabulary.wholeByName {
		built.wholeWords[name] = true
	}
	for _, entry := range append(append([]*abbreviationEntry{}, vocabulary.earlyPrefixes...), vocabulary.latePrefixes...) {
		built.prefixes[entry.Abbreviation] = true
		built.longestPrefix = max(built.longestPrefix, len(entry.Abbreviation))
	}
	for _, entry := range vocabulary.suffixes {
		if entry.Suffix.Matcher == "millisecondWord" {
			built.millisecond = true
			continue
		}
		built.suffixes[capitalizeAbbreviation(entry.Abbreviation)] = true
		built.longestSuffix = max(built.longestSuffix, len(entry.Abbreviation))
	}
	for _, entry := range vocabulary.segments {
		built.segments[capitalizeAbbreviation(entry.Abbreviation)] = true
		built.longestSegment = max(built.longestSegment, len(entry.Abbreviation))
	}
	return built
}

// isAbbreviationCandidate reports whether a name could match any branch of the rule.
//
// A superset of the branches, never a subset. See the file comment for why that direction is not
// negotiable.
func isAbbreviationCandidate(name string) bool {
	if name == "" {
		return false
	}

	// First arm: the whole name is an abbreviation.
	if gate.wholeWords[name] {
		return true
	}

	// Second arm: an abbreviation prefix followed by an uppercase letter.
	//
	// Scanning to the first uppercase letter finds the only split the arm can match, since the
	// prefix in the pattern is anchored at the start and is entirely lowercase.
	for index := 1; index < len(name) && index <= gate.longestPrefix; index++ {
		if isUppercaseAsciiLetter(name[index]) {
			if gate.prefixes[name[:index]] {
				return true
			}
			break
		}
	}

	// Third arm: a capitalized abbreviation at the end.
	//
	// Lengths run from two, not three: `Fn` is the shortest entry in the table, and starting at
	// three silently loses every name ending in it. The corpus caught that, on `accessorFn`,
	// `LanguageFn`, `noFn`, and a bare `Fn`.
	for length := 2; length <= gate.longestSuffix && length <= len(name); length++ {
		if gate.suffixes[name[len(name)-length:]] {
			return true
		}
	}

	// Fourth arm: `[a-z]Ms($|[A-Z])`. A lowercase letter, then `Ms`, then the end or a new word.
	//
	// This arm is why the gate cannot be built by splitting the name at every uppercase letter and
	// checking the words: that split puts `Ms` in its own word and discards the lowercase letter
	// before it, which is precisely the boundary being tested. Two independent attempts lost
	// `elapsedMs` exactly here.
	for index := 1; gate.millisecond && index+2 <= len(name); index++ {
		if name[index] != 'M' || name[index+1] != 's' {
			continue
		}
		if !isLowercaseAsciiLetter(name[index-1]) {
			continue
		}
		atEnd := index+2 == len(name)
		if atEnd || isUppercaseAsciiLetter(name[index+2]) {
			return true
		}
	}

	// Fifth arm: a capitalized abbreviation segment, followed by the end, a new word, or a digit.
	//
	// The scan starts at zero rather than one, because this arm is unanchored in the pattern: a name
	// that *is* the segment, or starts with it, matches. Starting at one lost `Cwd`, `Dir`, `Env`,
	// `Len`, `Seq`, `Var`, and `DbD` against the corpus.
	for index := 0; index < len(name); index++ {
		if !isUppercaseAsciiLetter(name[index]) {
			continue
		}
		for length := 2; length <= gate.longestSegment && index+length <= len(name); length++ {
			if !gate.segments[name[index:index+length]] {
				continue
			}
			after := index + length
			if after == len(name) || isUppercaseAsciiLetter(name[after]) || isAsciiDigit(name[after]) {
				return true
			}
		}
	}

	return false
}

func isUppercaseAsciiLetter(character byte) bool {
	return character >= 'A' && character <= 'Z'
}

func isLowercaseAsciiLetter(character byte) bool {
	return character >= 'a' && character <= 'z'
}

func isAsciiDigit(character byte) bool {
	return character >= '0' && character <= '9'
}
