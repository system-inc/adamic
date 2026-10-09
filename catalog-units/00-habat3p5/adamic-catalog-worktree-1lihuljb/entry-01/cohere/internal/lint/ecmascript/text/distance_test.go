package text

import "testing"

// Verbatim from oxc's own `edit_distance.rs` test module, which is the imported corpus for this
// helper rather than for any rule. `sitting`/`kitten` is the textbook Levenshtein case and pins the
// three-operation recurrence: one substitution, one substitution, one insertion.
func TestMinimumEditDistanceMatchesUpstream(t *testing.T) {
	t.Parallel()
	cases := []struct {
		first  string
		second string
		want   int
	}{
		{"", "", 0},
		{"a", "a", 0},
		{"abc", "abc", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "def", 3},
		{"sitting", "kitten", 3},
	}
	for _, testCase := range cases {
		if got := MinimumEditDistance(testCase.first, testCase.second); got != testCase.want {
			t.Fatalf("MinimumEditDistance(%q, %q) = %d, want %d",
				testCase.first, testCase.second, got, testCase.want)
		}
	}
}

// The argument order must not change the answer. oxc gets this from a recursive swap and this
// implementation from an in-place one, so it is worth asserting rather than assuming: a swap written
// on the wrong axis produces a function that is right in one direction and wrong in the other, and
// every caller here happens to pass its arguments the same way round.
func TestMinimumEditDistanceIsSymmetric(t *testing.T) {
	t.Parallel()
	pairs := [][2]string{
		{"getStaticProps", "getStaticPropss"},
		{"getServerSideProps", "getServurSideProps"},
		{"abc", ""},
		{"sitting", "kitten"},
	}
	for _, pair := range pairs {
		forward := MinimumEditDistance(pair[0], pair[1])
		backward := MinimumEditDistance(pair[1], pair[0])
		if forward != backward {
			t.Fatalf("%q/%q: %d forward against %d backward", pair[0], pair[1], forward, backward)
		}
	}
}

// A transposition costs two, not one. This is the single distinction that separates Levenshtein from
// Damerau, it is the one a reader is most likely to "fix", and it decides whether `getSatticProps`
// reports. Upstream does not report it.
func TestMinimumEditDistanceChargesTwoForATransposition(t *testing.T) {
	t.Parallel()
	if got := MinimumEditDistance("getSatticProps", "getStaticProps"); got != 2 {
		t.Fatalf("a transposition measured %d, want 2 (Damerau would say 1)", got)
	}
}

// Runes rather than bytes. A byte walk counts a two-byte rune as two edits, so this pair would
// measure 2 instead of 1. No caller currently passes non-ASCII, which is exactly why an unasserted
// rune walk would rot without anyone noticing.
func TestMinimumEditDistanceCountsRunesNotBytes(t *testing.T) {
	t.Parallel()
	if got := MinimumEditDistance("café", "cafe"); got != 1 {
		t.Fatalf("one accented substitution measured %d, want 1", got)
	}
	if got := MinimumEditDistance("é", ""); got != 1 {
		t.Fatalf("one accented rune against empty measured %d, want 1", got)
	}
}

// Case matters, and at full unit cost. TypeScript's spelling suggester charges 0.1 for a
// case-only substitution and this charges 1, which is the whole reason this function exists rather
// than a call into the shim.
func TestMinimumEditDistanceTreatsCaseAsAFullEdit(t *testing.T) {
	t.Parallel()
	if got := MinimumEditDistance("getStaticpaths", "getStaticPaths"); got != 1 {
		t.Fatalf("one case change measured %d, want 1", got)
	}
}

// Verbatim from oxc's `test_best_match`.
func TestBestMatchMatchesUpstream(t *testing.T) {
	t.Parallel()
	candidates := []string{"apple", "banana", "cherry"}

	if _, ok := BestMatch("apple", candidates, 2); ok {
		t.Fatal("an exact match must decline")
	}
	if got, ok := BestMatch("aple", candidates, 2); !ok || got != "apple" {
		t.Fatalf(`BestMatch("aple") = %q/%v, want "apple"/true`, got, ok)
	}
	if got, ok := BestMatch("banan", candidates, 2); !ok || got != "banana" {
		t.Fatalf(`BestMatch("banan") = %q/%v, want "banana"/true`, got, ok)
	}
	if _, ok := BestMatch("xyz", candidates, 2); ok {
		t.Fatal("nothing within threshold must decline")
	}
	if _, ok := BestMatch("test", nil, 2); ok {
		t.Fatal("an empty candidate list must decline")
	}
}

// The tie-break: first in slice order wins. Unreachable at the threshold the only caller uses, and
// asserted anyway because it is the behaviour a reader would otherwise have to derive from the loop.
func TestBestMatchKeepsTheFirstCandidateOnATie(t *testing.T) {
	t.Parallel()
	if got, _ := BestMatch("bat", []string{"cat", "hat"}, 1); got != "cat" {
		t.Fatalf("tie resolved to %q, want the first candidate", got)
	}
	if got, _ := BestMatch("bat", []string{"hat", "cat"}, 1); got != "hat" {
		t.Fatalf("reversed tie resolved to %q, want the first candidate", got)
	}
}

// A nearer candidate beats a further one regardless of order, so the tie-break above is genuinely a
// tie-break rather than a first-wins search.
func TestBestMatchPrefersTheNearestCandidate(t *testing.T) {
	t.Parallel()
	if got, _ := BestMatch("bats", []string{"bat", "batsy"}, 1); got != "bat" {
		t.Fatalf("got %q; both are within threshold and both are distance 1", got)
	}
	if got, _ := BestMatch("cars", []string{"carts", "car"}, 2); got != "carts" {
		t.Fatalf("got %q, want the distance-1 candidate over the distance-1 one at index 1", got)
	}
}

// The exact-match decline short-circuits rather than skipping the candidate. `car` is exact and
// `care` is one away, so a skip-this-one implementation would answer `care`.
func TestBestMatchExactMatchShortCircuitsTheWholeSearch(t *testing.T) {
	t.Parallel()
	if got, ok := BestMatch("car", []string{"car", "care"}, 1); ok {
		t.Fatalf("got %q; an exact match must end the search rather than skip one candidate", got)
	}
}

// The length prefilter must never change an answer, only skip work. A candidate whose length is
// within threshold has to survive it.
func TestBestMatchLengthPrefilterDoesNotHideACandidate(t *testing.T) {
	t.Parallel()
	if got, ok := BestMatch("getStaticPropss", []string{"getStaticProps"}, 1); !ok || got != "getStaticProps" {
		t.Fatalf("BestMatch = %q/%v; one insertion is within both the prefilter and the threshold", got, ok)
	}
	if _, ok := BestMatch("getServerSidePropsss", []string{"getServerSideProps"}, 1); ok {
		t.Fatal("two insertions is outside threshold 1 and must decline")
	}
}
