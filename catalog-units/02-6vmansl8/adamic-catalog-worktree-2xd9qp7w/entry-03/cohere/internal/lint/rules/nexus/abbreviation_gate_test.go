package nexus

import (
	"os"
	"strings"
	"testing"
)

// The gate must be a superset of what the vocabulary reports, and this is the only thing that proves
// it.
//
// It runs the vocabulary itself with no gate in front of it, which is stronger than the regular
// expression this used to compare against: that pattern was a hand-written description of the
// branches, and the vocabulary is the branches.
//
// The asymmetry is the whole point and it is stated in the rule's own comment: a false positive here
// costs one cheap traversal, while a false negative silently stops the rule firing, and a rule that
// stops firing looks exactly like a clean run. So over-approximation passes and under-approximation
// fails loudly.
//
// The corpus is every distinct identifier in the ahra tree rather than a hand-written list, because
// a hand-written list is exactly where this goes wrong. Two independent attempts at this gate lost
// `elapsedMs` to the same subtle cause: splitting a name at every uppercase letter puts `Ms` in its
// own word and discards the lowercase letter before it, and that boundary is precisely what the
// `[a-z]Ms($|[A-Z])` arm tests. A list somebody wrote from reading the pattern does not contain the
// case they misread.
func TestAbbreviationGateHasNoFalseNegatives(t *testing.T) {
	t.Parallel()

	corpus := loadIdentifierCorpus(t)

	falseNegatives := []string{}
	for _, name := range corpus {
		if _, reported := vocabulary.find(name); reported && !isAbbreviationCandidate(name) {
			falseNegatives = append(falseNegatives, name)
		}
	}

	if len(falseNegatives) != 0 {
		shown := falseNegatives
		if len(shown) > 20 {
			shown = shown[:20]
		}
		t.Fatalf("gate misses %d names the vocabulary reports, which would silence the rule for each: %v",
			len(falseNegatives), shown)
	}
}

// Over-approximation is allowed, but a gate that admits everything is not a gate.
//
// This is the vacuous-pass guard: a gate returning true unconditionally has zero false negatives and
// is worthless, and the test above would call it correct. Measuring the over-approximation rate
// turns "no false negatives" into a claim about a real filter.
func TestAbbreviationGateStaysSelective(t *testing.T) {
	t.Parallel()

	corpus := loadIdentifierCorpus(t)

	admitted := 0
	reported := 0
	for _, name := range corpus {
		if isAbbreviationCandidate(name) {
			admitted++
		}
		if _, found := vocabulary.find(name); found {
			reported++
		}
	}

	// The vocabulary reports a small fraction of real identifiers. The gate may admit more, but not
	// by a margin that would put the expensive work back.
	const maximumAdmittedFraction = 0.10
	if float64(admitted) > float64(len(corpus))*maximumAdmittedFraction {
		t.Fatalf("gate admits %d of %d names (%.1f%%), which is too coarse to be worth having",
			admitted, len(corpus), 100*float64(admitted)/float64(len(corpus)))
	}
	t.Logf("corpus %d, vocabulary reports %d, gate admits %d", len(corpus), reported, admitted)
}

// The arms the two failed attempts lost, pinned by name so a regression names itself.
//
// Every one of these is a real spelling from the tree that the vocabulary reports. They are here in
// addition to the corpus test, not instead of it: the corpus proves the gate today, and these say
// which shapes were historically easy to drop.
func TestAbbreviationGateCoversTheArmsThatWereLost(t *testing.T) {
	t.Parallel()

	cases := []string{
		// The `[a-z]Ms($|[A-Z])` arm. Splitting at uppercase discards the lowercase letter before
		// `Ms`, which is the boundary the arm tests.
		"elapsedMs", "timeoutMs", "delayMs", "elapsedMsTotal",
		// Mid-name word segments, the `(Cwd|Dir|Env|...)($|[A-Z0-9])` arm.
		"workingCwd", "outputDir", "processEnv", "outputDirName", "maxLen", "nextSeq",
		// Whole-word names.
		"props", "params", "ctx", "opts", "queryFn",
		// Prefix arm.
		"ctxValue", "optsBag", "idxOf", "configName",
		// Suffix arm.
		"eventProps", "routeParams", "elementRef", "handlerFn",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, reported := vocabulary.find(name); !reported {
				t.Fatalf("the vocabulary does not report %q, so it proves nothing", name)
			}
			if !isAbbreviationCandidate(name) {
				t.Fatalf("gate misses %q, which the vocabulary reports", name)
			}
		})
	}
}

// loadIdentifierCorpus reads every distinct identifier collected from the ahra tree.
func loadIdentifierCorpus(t *testing.T) []string {
	t.Helper()

	data, err := os.ReadFile("testdata/identifiers.txt")
	if err != nil {
		t.Fatalf("could not read the identifier corpus: %v", err)
	}
	corpus := strings.Split(strings.TrimSpace(string(data)), "\n")
	// A corpus that failed to load would make every assertion above pass vacuously.
	if len(corpus) < 10000 {
		t.Fatalf("corpus holds only %d identifiers, which is too few to have been read correctly", len(corpus))
	}
	return corpus
}
