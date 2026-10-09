package react_conformance

import (
	"fmt"
	"sort"
	"strings"
)

// Implementation is what a candidate rule engine has to provide to be scored.
//
// One method, and it hands back `{message, line}` pairs. That is upstream's own backend-comparison
// contract from `RustBackendComparison-test.ts`, and it is the narrowest surface that can still
// distinguish a right answer from a wrong one. Nothing implements it yet, which is the point: the
// instrument goes down before the thing it measures.
//
// An implementation that cannot handle a fixture returns ErrUnsupported rather than an empty
// result. An empty result is a claim ("I found no errors here"), and for a corpus where every
// fixture expects at least one error, that claim is a wrong answer. Declining is a different fact
// and gets counted in a different column.
type Implementation interface {
	// Analyze reports the diagnostics found in one fixture.
	Analyze(fixture Fixture) (Result, error)
}

// ErrUnsupported is what an Implementation returns for a fixture it cannot handle.
type ErrUnsupported struct{ Reason string }

func (e *ErrUnsupported) Error() string { return "unsupported: " + e.Reason }

// Outcome is what happened to one fixture.
type Outcome string

const (
	// OutcomePassed means every expected `{message, line}` was reported, and nothing else was.
	OutcomePassed Outcome = "Passed"

	// OutcomeFailed means the implementation ran and disagreed with the golden.
	OutcomeFailed Outcome = "Failed"

	// OutcomeDeclined means the implementation said it could not handle the fixture. Distinct from
	// Failed because "I do not implement this yet" and "I implement it wrongly" are different
	// facts, and collapsing them makes progress unreadable.
	OutcomeDeclined Outcome = "Declined"

	// OutcomeExcluded means the runner itself declined the fixture before the implementation saw
	// it — today only for requiring a Flow parser. A stated exclusion with a count, never a skip.
	OutcomeExcluded Outcome = "Excluded"

	// OutcomeRefused means an unrecognised pragma stopped the fixture from being scored at all.
	// Non-zero here is a defect in this package, not a score.
	OutcomeRefused Outcome = "Refused"
)

// FixtureScore is one fixture's verdict.
type FixtureScore struct {
	Name    string
	Outcome Outcome

	// Detail explains a non-pass in one line.
	Detail string

	ExpectedErrors []ExpectedError
	ReportedErrors []ReportedError
}

// Score is the whole run.
//
// The counters exist because "0 passed" and "found nothing to run" print the same digit. Considered
// is the denominator the score is honest about, and every fixture lands in exactly one outcome
// bucket, asserted rather than assumed — see `Score.Check`.
type Score struct {
	// Considered is every fixture loaded, including ones never handed to the implementation.
	Considered int

	Passed   int
	Failed   int
	Declined int
	Excluded int
	Refused  int

	Fixtures []FixtureScore
}

// Check verifies the counters partition the corpus.
//
// This is the guard on the guard. If a future outcome is added and its counter is not summed here,
// the arithmetic stops adding up and says so, rather than quietly under-reporting the denominator
// and making a partial run look like a complete one.
func (s Score) Check() error {
	sum := s.Passed + s.Failed + s.Declined + s.Excluded + s.Refused
	if sum != s.Considered {
		return fmt.Errorf(
			"score does not partition the corpus: considered %d but outcomes sum to %d (passed %d, failed %d, declined %d, excluded %d, refused %d)",
			s.Considered, sum, s.Passed, s.Failed, s.Declined, s.Excluded, s.Refused,
		)
	}
	if s.Considered != len(s.Fixtures) {
		return fmt.Errorf("score considered %d fixtures but recorded %d verdicts", s.Considered, len(s.Fixtures))
	}
	return nil
}

// Summary is the one-line scoreboard.
//
// It always states the denominator and always states what was not scored, because "0 / 325" and
// "0 / 0" differ by exactly the fact this suite is built to keep visible.
func (s Score) Summary() string {
	return fmt.Sprintf(
		"%d / %d passed (failed %d, declined %d, excluded %d, refused %d) against facebook/react@%s",
		s.Passed, s.Considered, s.Failed, s.Declined, s.Excluded, s.Refused, UpstreamSha[:12],
	)
}

// Options configure a run.
type Options struct {
	// SkipFlowFixtures excludes the 35 fixtures whose source contains `@flow`. They are counted as
	// Excluded with a reason, never dropped from the denominator.
	SkipFlowFixtures bool
}

// Run scores an Implementation against the corpus.
//
// The pragma check happens before the implementation is consulted, and a refusal aborts the whole
// run rather than marking one fixture. An unrecognised directive means this package's model of the
// corpus is wrong, and continuing would produce a scoreboard whose other 324 entries were computed
// under a model already known to be incomplete.
func Run(fixtures []Fixture, implementation Implementation, options Options) (Score, error) {
	score := Score{Considered: len(fixtures)}

	for _, fixture := range fixtures {
		if err := CheckPragmas(fixture.Name, fixture.Pragmas); err != nil {
			return Score{}, err
		}
	}

	for _, fixture := range fixtures {
		if options.SkipFlowFixtures && fixture.RequiresFlow() {
			score.Excluded++
			score.Fixtures = append(score.Fixtures, FixtureScore{
				Name:    fixture.Name,
				Outcome: OutcomeExcluded,
				Detail:  "requires a Flow parser (`@flow` in source, upstream's own test)",
			})
			continue
		}

		result, err := implementation.Analyze(fixture)
		if err != nil {
			var unsupported *ErrUnsupported
			if asUnsupported(err, &unsupported) {
				score.Declined++
				score.Fixtures = append(score.Fixtures, FixtureScore{
					Name:           fixture.Name,
					Outcome:        OutcomeDeclined,
					Detail:         unsupported.Reason,
					ExpectedErrors: fixture.Expected.Errors,
				})
				continue
			}
			return Score{}, fmt.Errorf("%s: %w", fixture.Name, err)
		}

		verdict := Compare(fixture.Expected, result)
		verdict.Name = fixture.Name
		switch verdict.Outcome {
		case OutcomePassed:
			score.Passed++
		default:
			score.Failed++
		}
		score.Fixtures = append(score.Fixtures, verdict)
	}

	if err := score.Check(); err != nil {
		return Score{}, err
	}
	return score, nil
}

func asUnsupported(err error, target **ErrUnsupported) bool {
	for err != nil {
		if unsupported, ok := err.(*ErrUnsupported); ok {
			*target = unsupported
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

// Compare decides whether a Result satisfies an Expectation.
//
// The comparison is on `{message-first-line, line}` as an unordered multiset, and each half of that
// choice is deliberate.
//
// Only the first line of the message, because the golden's remaining paragraphs are prose carrying
// react.dev URLs. Upstream edits them for clarity without touching compiler behaviour, and a runner
// asserting on the full block would report a rule regression for a documentation change. This is
// what upstream's own `RustBackendComparison-test.ts` asserts on for the same reason.
//
// Line but not column, because the caret span in the golden is rendered from a byte offset into a
// source excerpt, and reproducing the exact column requires agreeing with upstream's span
// convention for every node kind before any rule exists to disagree about. Line is the coarser
// claim and it is the one that can be met honestly today. Columns are parsed and carried on
// ExpectedError so a later pass can tighten this without re-reading the corpus.
//
// Unordered, because diagnostic emission order follows pass order in the pipeline, and requiring a
// Go port to match Rust pass ordering would fail correct implementations for a reason unrelated to
// correctness. Multiset rather than set, because 87 of the 325 expect more than one error and
// collapsing duplicates would let one finding satisfy two expectations.
func Compare(expected Expectation, result Result) FixtureScore {
	type pair struct {
		Heading string
		Message string
		Line    int
	}

	want := map[pair]int{}
	for _, expectedError := range expected.Errors {
		want[pair{expectedError.Heading, expectedError.Message, expectedError.Line}]++
	}
	got := map[pair]int{}
	for _, reportedError := range result.Errors {
		got[pair{reportedError.Heading, reportedError.Message, reportedError.Line}]++
	}

	var missing, unexpected []string
	for key, count := range want {
		if got[key] < count {
			missing = append(missing, fmt.Sprintf("%s: %s (line %d) x%d", key.Heading, key.Message, key.Line, count-got[key]))
		}
	}
	for key, count := range got {
		if want[key] < count {
			unexpected = append(unexpected, fmt.Sprintf("%s: %s (line %d) x%d", key.Heading, key.Message, key.Line, count-want[key]))
		}
	}
	sort.Strings(missing)
	sort.Strings(unexpected)

	verdict := FixtureScore{
		ExpectedErrors: expected.Errors,
		ReportedErrors: result.Errors,
	}
	if len(missing) == 0 && len(unexpected) == 0 {
		verdict.Outcome = OutcomePassed
		return verdict
	}

	verdict.Outcome = OutcomeFailed
	var detail []string
	if len(missing) > 0 {
		detail = append(detail, "missing: "+strings.Join(missing, "; "))
	}
	if len(unexpected) > 0 {
		detail = append(detail, "unexpected: "+strings.Join(unexpected, "; "))
	}
	verdict.Detail = strings.Join(detail, " | ")
	return verdict
}

// NothingImplemented is the Implementation used until a real one exists.
//
// It declines every fixture. It does NOT return an empty result, and the difference is the whole
// reason the type is written out rather than left as a nil check: an empty result asserts "no
// errors here", which is a wrong answer on a corpus where every fixture expects at least one, and
// would score as 325 failures — a number that looks like a working implementation doing badly
// rather than an absent one.
type NothingImplemented struct{}

// Analyze declines.
func (NothingImplemented) Analyze(Fixture) (Result, error) {
	return Result{}, &ErrUnsupported{Reason: "no React compiler rules are implemented in cohere yet"}
}
