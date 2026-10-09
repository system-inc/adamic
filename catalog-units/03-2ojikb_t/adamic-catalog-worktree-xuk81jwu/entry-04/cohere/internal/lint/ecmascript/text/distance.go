// Package text holds string measurements that are not specific to any language we lint.
//
// It exists because the one caller that needed an edit distance found three implementations already
// in reach and every one of them answers a different question. That is the whole reason to write a
// fourth rather than to reuse: a distance function's *weights* are its meaning, and a helper whose
// name matches your question while its weights do not will pass every fixture you write from it.
package text

// MinimumEditDistance returns the Levenshtein distance between two strings.
//
// Unit cost for an insertion, a deletion, or a substitution. **No transposition**, so swapping two
// adjacent characters costs two rather than one. That is not an omission: it is what oxc's
// `min_edit_distance` computes, and it decides real cases. `getSatticProps` is one transposition
// from `getStaticProps` and therefore sits at distance two, which is outside the threshold every
// caller here uses, so it does not read as a typo. A Damerau variant would report it.
//
// Compared over runes rather than bytes, matching oxc's `chars()` walk. For the ASCII identifiers
// this currently serves the two agree, but a byte walk would count a multi-byte rune as several
// edits, which silently inflates the distance for any non-ASCII name.
//
// Two rows rather than a full matrix, which is oxc's shape and costs O(min(len)) memory.
//
// # Why not the two implementations already in this binary
//
// **`core.GetSpellingSuggestionForStrings`, reachable through the shim, is not Levenshtein and the
// difference is not subtle.** TypeScript weights a substitution that differs only in case at 0.1 and
// any other substitution at 2, against 1 for an insertion or a deletion, then accepts anything
// within 0.4 of the name's length. Measured against this function's three-candidate caller: it
// answers a suggestion for `getServerSidePropsss` and for `getstatisPath`, both of which upstream
// ships as *passing* cases, and for `getSatticProps`, the transposition described above. A port
// built on it reports on clean code. The name is the closest match on any shelf here and the body is
// the wrong answer, which is why this comment names it rather than leaving the next reader to find
// out by shipping it.
func MinimumEditDistance(a string, b string) int {
	first := []rune(a)
	second := []rune(b)

	// Keep the shorter string on the inner axis so the rows allocated are the smaller of the two.
	// oxc spells this as a recursive swap on byte length; the effect is the same and the comparison
	// is on rune length here for the same reason the walk is.
	if len(first) < len(second) {
		first, second = second, first
	}

	width := len(second)
	previous := make([]int, width+1)
	current := make([]int, width+1)
	for index := range previous {
		previous[index] = index
	}

	for firstIndex, firstRune := range first {
		current[0] = firstIndex + 1
		for secondIndex, secondRune := range second {
			substitution := previous[secondIndex]
			if firstRune != secondRune {
				substitution++
			}
			deletion := previous[secondIndex+1] + 1
			insertion := current[secondIndex] + 1

			best := substitution
			if deletion < best {
				best = deletion
			}
			if insertion < best {
				best = insertion
			}
			current[secondIndex+1] = best
		}
		previous, current = current, previous
	}

	return previous[width]
}

// BestMatch returns the candidate nearest to needle within threshold, and whether one was found.
//
// **An exact match is not a match here.** A needle equal to some candidate returns false, because
// every caller is asking "is this a near miss of one of these names", and a name that IS one of
// these names is correct rather than nearly correct. Getting this backwards turns the rule into one
// that flags every correct spelling, which is the loudest possible failure and therefore the one
// most likely to be caught; the quieter risk is a caller that expects a hit and reads the false as
// "nothing was close".
//
// The exact-match decline short-circuits the whole search rather than skipping the one candidate,
// which is oxc's behaviour and is observable in principle: a needle exactly equal to one candidate
// and within threshold of a later one answers false rather than naming the later one. With the
// candidate sets in use here no such pair exists, so it is reproduced for fidelity rather than
// because a case reaches it.
//
// On a tie, the FIRST candidate in slice order wins, so the caller's ordering is part of its meaning.
//
// The length prefilter is oxc's and is preserved. It skips the distance computation when the lengths
// alone put a candidate out of range, which is sound: a distance can never be less than the
// difference in lengths.
//
// **It is a pure optimization and removing it cannot change an answer**, which is stated here
// because a mutation sweep reports its removal as a surviving mutant and the next reader will want
// to know whether that is a blind spot. It is not: measured exhaustively over every needle and
// every ordered candidate pair drawn from the 121 strings of length 0 to 4 over a three-letter
// alphabet, at thresholds 0 through 3, which is 7,086,244 comparisons, the filtered and unfiltered
// searches agreed on every one. The exact-match decline is the case worth checking by hand, and it
// is safe for a structural reason rather than a measured one: an exact match has zero length
// difference, so the filter can never skip the candidate that would have triggered it. oxc compares byte lengths there while computing over runes, which for
// non-ASCII input would prefilter away a candidate the distance would have accepted. Rune lengths
// are used here on both sides instead. That is a deliberate narrow divergence toward internal
// consistency, and it cannot change an answer for any caller whose candidates are ASCII.
func BestMatch(needle string, candidates []string, threshold int) (string, bool) {
	needleLength := len([]rune(needle))

	bestCandidate := ""
	bestDistance := 0
	found := false

	for _, candidate := range candidates {
		candidateLength := len([]rune(candidate))

		difference := candidateLength - needleLength
		if difference < 0 {
			difference = -difference
		}
		if difference > threshold {
			continue
		}

		distance := MinimumEditDistance(candidate, needle)
		if distance == 0 {
			return "", false
		}
		if distance > threshold {
			continue
		}
		if found && distance >= bestDistance {
			continue
		}
		bestCandidate, bestDistance, found = candidate, distance, true
	}

	return bestCandidate, found
}
