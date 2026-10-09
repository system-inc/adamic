package react_conformance

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureRoot is the vendored corpus.
//
// Vendored rather than fetched, and the tradeoff was real on both sides. Fetching keeps the corpus
// current; vendoring makes the suite hermetic, reviewable in a diff, and runnable with no network.
// 650 files and 504 KiB is a cheap price for a score that means the same thing on every machine on
// every day, and the currency that fetching buys is the thing a pinned sha deliberately gives up:
// an expectation that changes under the implementation is a diff a human should read, not a number
// that moves. `internal/lint/rules/react/tools/vendor_fixtures` re-pulls at a chosen sha when the pin is bumped.
const fixtureRoot = "testdata/fixtures"

func load(t *testing.T) []Fixture {
	t.Helper()
	fixtures, err := Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the vendored corpus: %v", err)
	}
	return fixtures
}

// TestCorpusIsTheSizeItClaimsToBe anchors the counts.
//
// The number is asserted rather than derived because a derived number agrees with whatever it
// found. The research pass this suite comes from reported 221 error fixtures because its grep was
// anchored at the fixture root and so excluded every subdirectory — wrong by a third, and it read
// exactly like a clean result. It was caught only because a parallel investigation returned a
// different number.
//
// So the assertion is the literal 325, established twice by different means before it was written
// here: once from the GitHub git-tree API at the pinned sha with `truncated: false`, and once by
// walking the extracted tarball on disk. Both said 325, 222 at the top level and 103 across nine
// subdirectories. The depth split is asserted too, because that is the exact axis the wrong answer
// moved along: a run reporting 222 here is the original defect reproducing itself.
// The corpus also holds a second population, and the split is asserted rather than the total alone.
// `ExpectedCleanFixtureCount` fixtures carry the preserve-memoization pragma and expect no error;
// they are the only false-positive oracle here, since over-reporting on an error fixture produces
// more findings and reads as success. A total-only assertion would let the two populations trade
// against each other silently, which is the shape of defect this whole file guards.
func TestCorpusIsTheSizeItClaimsToBe(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	errorNamed, clean := 0, 0
	for _, fixture := range fixtures {
		if base := filepath.Base(fixture.Name); strings.HasPrefix(base, "error.") ||
			strings.HasPrefix(base, "todo.error.") {
			errorNamed++
			continue
		}
		clean++
	}
	if errorNamed != ExpectedErrorFixtureCount {
		t.Errorf("corpus holds %d error fixtures, want %d at sha %s",
			errorNamed, ExpectedErrorFixtureCount, UpstreamSha)
	}
	if clean != ExpectedCleanFixtureCount {
		t.Errorf("corpus holds %d clean fixtures, want %d at sha %s; these are the only population "+
			"where a finding is a defect, so losing them silently removes the ability to measure "+
			"over-reporting at all", clean, ExpectedCleanFixtureCount, UpstreamSha)
	}

	// The depth split below is about the error fixtures specifically, because 222 + 103 is the
	// number that moved and the axis it moved along.
	topLevel, nested := 0, 0
	subdirectories := map[string]int{}
	for _, fixture := range fixtures {
		base := filepath.Base(fixture.Name)
		if !strings.HasPrefix(base, "error.") && !strings.HasPrefix(base, "todo.error.") {
			continue
		}
		directory := filepath.Dir(fixture.Name)
		if directory == "." {
			topLevel++
			continue
		}
		nested++
		subdirectories[directory]++
	}

	// 222 + 103, not 221 + 104. The brief this task came from carried the latter; both halves were
	// off by one and the total was right, which is how a transcription slip survives a total check.
	if topLevel != 222 {
		t.Errorf("top-level fixtures = %d, want 222", topLevel)
	}
	if nested != 103 {
		t.Errorf("nested fixtures = %d, want 103", nested)
	}
	if len(subdirectories) != 9 {
		t.Errorf("subdirectories holding error fixtures = %d, want 9; got %v", len(subdirectories), subdirectories)
	}

	// Guard the trap directly: a filter that only saw the top level would report 222 and pass every
	// other assertion in this file.
	if nested == 0 {
		t.Fatal("no nested fixtures found at all, which is the exact shape of the depth-anchored grep defect")
	}
}

// TestEveryFixtureHasAParsedExpectation proves the goldens were read, not merely opened.
//
// Load succeeding says the files exist. This says they parsed into something with content in it, so
// an expectation parser that silently produced empty results cannot pass — which it otherwise
// would, since an empty expectation compared against an empty result is a pass, and 325 of those is
// a perfect score over a corpus nobody read.
func TestEveryFixtureHasAParsedExpectation(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	withDeclaredCount, withLocation, totalErrors := 0, 0, 0
	for _, fixture := range fixtures {
		// A clean fixture has no `## Error` block, and that absence IS its expectation rather than a
		// parse failure. Only an error-named fixture with an empty block is the defect this catches.
		base := filepath.Base(fixture.Name)
		if !strings.HasPrefix(base, "error.") && !strings.HasPrefix(base, "todo.error.") {
			if fixture.Expected.Raw != "" {
				t.Errorf("%s: clean fixture carries an `## Error` block", fixture.Name)
			}
			continue
		}
		if fixture.Expected.Raw == "" {
			t.Errorf("%s: empty `## Error` block", fixture.Name)
		}
		if fixture.Expected.HasDeclaredCount {
			withDeclaredCount++
		}
		for _, expectedError := range fixture.Expected.Errors {
			totalErrors++
			if expectedError.Message == "" {
				t.Errorf("%s: parsed an error with an empty message", fixture.Name)
			}
			if expectedError.HasLocation {
				withLocation++
			}
		}
	}

	// 311 carry `Found N error(s)`; the other 14 are parse errors and pipeline exceptions with a
	// different shape entirely (`Missing semicolon. (7:24)`, `unexpected error`). Asserting the
	// split keeps a parser that started ignoring the header from passing quietly.
	if withDeclaredCount != 311 {
		t.Errorf("fixtures with a `Found N error` header = %d, want 311", withDeclaredCount)
	}

	if totalErrors == 0 {
		t.Fatal("parsed zero individual errors across the whole corpus, which is what a broken parser looks like from a passing suite")
	}

	// 429 diagnostics, 398 with a primary location. Both pinned, because both are load-bearing and
	// both have a plausible wrong neighbour: counting raw `path:line:col` lines instead of primary
	// spans gives 427, since 31 diagnostics carry a secondary span and 31 different ones carry no
	// span at all. The two errors nearly cancel, which is what makes 427 dangerous rather than
	// obviously wrong.
	if totalErrors != 429 {
		t.Errorf("parsed %d individual errors, want 429 at sha %s", totalErrors, UpstreamSha)
	}
	if withLocation != 398 {
		t.Errorf("errors carrying a primary location = %d, want 398", withLocation)
	}
	t.Logf("parsed %d individual errors, %d carrying a location, across %d fixtures", totalErrors, withLocation, len(fixtures))
}

// TestDeclaredCountMatchesParsedErrors cross-checks the parser against the goldens' own arithmetic.
//
// Each golden states how many errors it holds, and the parser counts them independently. Two
// measurements of the same quantity that were not derived from each other: if the parser drops an
// `Error:` line or invents one, the sum stops matching the headers and says so.
func TestDeclaredCountMatchesParsedErrors(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	declaredTotal, parsedTotal := 0, 0
	for _, fixture := range fixtures {
		if !fixture.Expected.HasDeclaredCount {
			continue
		}
		declaredTotal += fixture.Expected.DeclaredCount
		parsedTotal += len(fixture.Expected.Errors)

		if fixture.Expected.DeclaredCount != len(fixture.Expected.Errors) {
			t.Errorf("%s: golden declares %d errors, parser found %d",
				fixture.Name, fixture.Expected.DeclaredCount, len(fixture.Expected.Errors))
		}
	}

	if declaredTotal != parsedTotal {
		t.Errorf("declared %d errors in total, parsed %d", declaredTotal, parsedTotal)
	}
	t.Logf("declared and parsed agree at %d individual errors across the 311 headered fixtures", declaredTotal)
}

// TestFlowFixtureCountIsUpstreamsOwnTest pins the Flow split.
//
// 35, not 72. The brief carried 72 and therefore concluded 253 fixtures need no Flow parser; the
// real figure is 290. The gap is worth pinning because it is the difference between two plausible
// tests: `@flow` anywhere in the source, which is upstream's (`packages/snap/src/compiler.ts:36`,
// `source.indexOf('@flow') !== -1`), gives 35, while a first-line-pragma test gives 20. Using the
// narrower one would hand 15 Flow-syntax files to a TypeScript parser and read the parse failures
// as rule failures.
func TestFlowFixtureCountIsUpstreamsOwnTest(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	flowCount := 0
	for _, fixture := range fixtures {
		if fixture.RequiresFlow() {
			flowCount++
		}
	}

	if flowCount != ExpectedFlowFixtureCount {
		t.Errorf("fixtures requiring a Flow parser = %d, want %d at sha %s",
			flowCount, ExpectedFlowFixtureCount, UpstreamSha)
	}
	if got, want := len(fixtures)-flowCount, ExpectedFixtureCount-ExpectedFlowFixtureCount; got != want {
		t.Errorf("fixtures needing no Flow parser = %d, want %d", got, want)
	}
}

// TestEveryPragmaInTheCorpusIsModelled is the fail-loud requirement, run over the real corpus.
//
// If this fails, the corpus moved and this package's model of it is stale. That is the intended
// outcome: a directive nobody has looked at should stop the suite, because an unmodelled directive
// changes what a fixture expects, and a runner that skips it reports "not applicable" for a case it
// silently got wrong.
func TestEveryPragmaInTheCorpusIsModelled(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	seen := map[string]int{}
	for _, fixture := range fixtures {
		for _, pragma := range fixture.Pragmas {
			seen[pragma.Key]++
		}
		if err := CheckPragmas(fixture.Name, fixture.Pragmas); err != nil {
			t.Errorf("%v", err)
		}
	}

	// 34 distinct directives at the pinned sha, up from 31 when the corpus held only error
	// fixtures. Asserted so that a corpus bump that adds one is visible here even if the new
	// directive happened to already be in a map. The three the clean population added are
	// `panicThreshold`, `loggerTestOnly` and `expectNothingCompiled`.
	if len(seen) != 34 {
		t.Errorf("distinct pragmas across the corpus = %d, want 34; got %v", len(seen), seen)
	}
	t.Logf("modelled %d directives, deliberately ignore %d", len(KnownPragmaKeys()), len(IgnoredPragmaKeys()))
}

// TestUnknownPragmaIsRefusedNotSkipped proves the refusal actually fires.
//
// The test over the real corpus above passes when every directive is modelled, which is also what a
// CheckPragmas that always returned nil would do. So this hands it a directive that is in neither
// map and requires an error, which is the only way to tell a working guard from an absent one.
func TestUnknownPragmaIsRefusedNotSkipped(t *testing.T) {
	t.Parallel()
	pragmas := ParsePragmas("// @validateRefAccessDuringRender @enableSomethingNobodyModelled")

	err := CheckPragmas("error.hypothetical.js", pragmas)
	if err == nil {
		t.Fatal("CheckPragmas accepted an unmodelled directive; the whole point of this package is that it does not")
	}

	var unknown *UnknownPragmaError
	if !errors.As(err, &unknown) {
		t.Fatalf("got %T, want *UnknownPragmaError", err)
	}
	if unknown.Pragma.Key != "enableSomethingNobodyModelled" {
		t.Errorf("refused %q, want the unmodelled directive", unknown.Pragma.Key)
	}
	if !strings.Contains(err.Error(), "error.hypothetical.js") {
		t.Errorf("the refusal does not name the fixture: %v", err)
	}
}

// TestRunRefusesTheWholeCorpusOnOneUnknownPragma checks the refusal aborts rather than annotates.
//
// One unmodelled directive means the model of the corpus is wrong, so the other 324 verdicts were
// computed under a model already known to be incomplete. Marking a single fixture and carrying on
// would produce exactly the scoreboard this package exists to prevent.
func TestRunRefusesTheWholeCorpusOnOneUnknownPragma(t *testing.T) {
	t.Parallel()
	fixtures := []Fixture{
		{Name: "error.fine.js", Pragmas: ParsePragmas("// @validateNoSetStateInRender")},
		{Name: "error.bad.js", Pragmas: ParsePragmas("// @somethingUnmodelled")},
	}

	if _, err := Run(fixtures, NothingImplemented{}, Options{}); err == nil {
		t.Fatal("Run scored a corpus containing an unmodelled directive")
	}
}

// TestPragmaParserReproducesUpstreamsSplit including the sharp edge it is supposed to reproduce.
//
// `@compilationMode(infer)` must parse to the key `compilationMode(infer)`, because upstream's
// `splitPragma` only splits values on `:`. A parser that normalised it would be kinder than
// upstream and would score three fixtures under a configuration upstream never applied when it
// recorded their expectations.
func TestPragmaParserReproducesUpstreamsSplit(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		firstLine string
		want      []Pragma
	}{
		{"function Component() {", nil},
		{"// @flow", []Pragma{{Key: "flow"}}},
		{"//@flow", []Pragma{{Key: "flow"}}},
		{
			"// @validateRefAccessDuringRender @compilationMode:\"infer\"",
			[]Pragma{
				{Key: "validateRefAccessDuringRender"},
				{Key: "compilationMode", Value: "\"infer\"", HasValue: true},
			},
		},
		{
			"// @flow @compilationMode(infer)",
			[]Pragma{{Key: "flow"}, {Key: "compilationMode(infer)"}},
		},
		{
			"// @validatePreserveExistingMemoizationGuarantees @enablePreserveExistingMemoizationGuarantees:false",
			[]Pragma{
				{Key: "validatePreserveExistingMemoizationGuarantees"},
				{Key: "enablePreserveExistingMemoizationGuarantees", Value: "false", HasValue: true},
			},
		},
	} {
		got := ParsePragmas(testCase.firstLine)
		if len(got) != len(testCase.want) {
			t.Errorf("%q: parsed %d pragmas (%v), want %d", testCase.firstLine, len(got), got, len(testCase.want))
			continue
		}
		for index := range got {
			if got[index] != testCase.want[index] {
				t.Errorf("%q: pragma %d = %+v, want %+v", testCase.firstLine, index, got[index], testCase.want[index])
			}
		}
	}
}

// TestExpectationParserReadsARealGolden checks the parse against a file read by hand.
//
// One fixture, verified by eye against the file on disk, so the parser is anchored to something
// outside its own output at least once.
func TestExpectationParserReadsARealGolden(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	var subject *Fixture
	for index := range fixtures {
		if fixtures[index].Name == "error.assign-global-in-jsx-children.js" {
			subject = &fixtures[index]
			break
		}
	}
	if subject == nil {
		t.Fatal("error.assign-global-in-jsx-children.js is not in the vendored corpus")
	}

	if !subject.Expected.HasDeclaredCount || subject.Expected.DeclaredCount != 1 {
		t.Errorf("declared count = %d (present %v), want 1", subject.Expected.DeclaredCount, subject.Expected.HasDeclaredCount)
	}
	if len(subject.Expected.Errors) != 1 {
		t.Fatalf("parsed %d errors, want 1", len(subject.Expected.Errors))
	}

	parsed := subject.Expected.Errors[0]
	if parsed.Heading != "Error" {
		t.Errorf("heading = %q, want %q", parsed.Heading, "Error")
	}
	if want := "Cannot reassign variables declared outside of the component/hook"; parsed.Message != want {
		t.Errorf("message = %q, want %q", parsed.Message, want)
	}
	if parsed.File != "error.assign-global-in-jsx-children.ts" || parsed.Line != 3 || parsed.Column != 4 {
		t.Errorf("location = %s:%d:%d, want error.assign-global-in-jsx-children.ts:3:4", parsed.File, parsed.Line, parsed.Column)
	}
}

// perfectImplementation answers every fixture with exactly what its golden expects.
//
// It is a fake, and it is the only thing in this package that can prove the comparison is capable
// of returning true.
type perfectImplementation struct{}

func (perfectImplementation) Analyze(fixture Fixture) (Result, error) {
	result := Result{}
	for _, expectedError := range fixture.Expected.Errors {
		result.Errors = append(result.Errors, ReportedError{
			Heading: expectedError.Heading,
			Message: expectedError.Message,
			Line:    expectedError.Line,
		})
	}
	return result, nil
}

// TestRunnerCanDetectAPass is the calibration, and it is the most important test in this file.
//
// Everything else here confirms the runner reports zero. A runner that always reported zero — a
// Compare that never returns Passed, a loader that finds nothing, a scorer that never increments —
// would satisfy every one of those tests and would be worthless, because the number it prints would
// not be a measurement of anything.
//
// So this feeds the scorer a hand-built implementation that returns each golden's own expected
// errors and requires 325 passes. If the comparison cannot recognise a correct answer, this is the
// test that says so, and it says so today rather than on the day a real rule lands and its first
// correct result is scored as a failure.
func TestRunnerCanDetectAPass(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	score, err := Run(fixtures, perfectImplementation{}, Options{})
	if err != nil {
		t.Fatalf("scoring a deliberately-correct implementation: %v", err)
	}

	if score.Passed != len(fixtures) {
		t.Errorf("a correct implementation scored %d / %d; the comparison cannot recognise a right answer", score.Passed, score.Considered)
		for _, fixtureScore := range score.Fixtures {
			if fixtureScore.Outcome != OutcomePassed {
				t.Logf("  %s: %s — %s", fixtureScore.Name, fixtureScore.Outcome, fixtureScore.Detail)
			}
		}
	}
	if score.Failed != 0 || score.Declined != 0 || score.Refused != 0 {
		t.Errorf("unexpected non-passes: %s", score.Summary())
	}
	t.Logf("calibration: %s", score.Summary())
}

// TestRunnerCanDetectAFailure is the other half of the calibration.
//
// A comparison that always returned Passed would satisfy the test above just as well as a correct
// one. This gives it a deliberately wrong answer and requires a failure, so that Passed and Failed
// are both shown to be reachable.
func TestRunnerCanDetectAFailure(t *testing.T) {
	t.Parallel()
	expectation := Expectation{
		Errors: []ExpectedError{{Heading: "Error", Message: "Cannot reassign variables", Line: 3}},
	}

	if verdict := Compare(expectation, Result{Errors: []ReportedError{{Heading: "Error", Message: "Cannot reassign variables", Line: 3}}}); verdict.Outcome != OutcomePassed {
		t.Errorf("an exactly-right result scored %s", verdict.Outcome)
	}

	// Right message, wrong line.
	if verdict := Compare(expectation, Result{Errors: []ReportedError{{Heading: "Error", Message: "Cannot reassign variables", Line: 9}}}); verdict.Outcome != OutcomeFailed {
		t.Errorf("a result on the wrong line scored %s", verdict.Outcome)
	}

	// Right line, wrong message.
	if verdict := Compare(expectation, Result{Errors: []ReportedError{{Heading: "Error", Message: "Something else", Line: 3}}}); verdict.Outcome != OutcomeFailed {
		t.Errorf("a result with the wrong message scored %s", verdict.Outcome)
	}

	// Right message and line, wrong heading. `Compilation Skipped` and `Error` are different
	// verdicts about the same source and a port that confuses them has got a real thing wrong.
	if verdict := Compare(expectation, Result{Errors: []ReportedError{{Heading: "Todo", Message: "Cannot reassign variables", Line: 3}}}); verdict.Outcome != OutcomeFailed {
		t.Errorf("a result under the wrong heading scored %s", verdict.Outcome)
	}

	// Nothing at all. This is the case an unimplemented engine would produce if it returned an
	// empty result rather than declining, and it must not read as a pass.
	if verdict := Compare(expectation, Result{}); verdict.Outcome != OutcomeFailed {
		t.Errorf("an empty result against a golden expecting an error scored %s", verdict.Outcome)
	}

	// One finding must not satisfy two expectations. 87 of the 325 expect more than one error.
	twoOfTheSame := Expectation{Errors: []ExpectedError{
		{Heading: "Error", Message: "Same message", Line: 4},
		{Heading: "Error", Message: "Same message", Line: 4},
	}}
	if verdict := Compare(twoOfTheSame, Result{Errors: []ReportedError{{Heading: "Error", Message: "Same message", Line: 4}}}); verdict.Outcome != OutcomeFailed {
		t.Errorf("one finding satisfied two identical expectations, so the comparison is a set rather than a multiset")
	}
}

// TestFirstHonestScoreIsZeroOfThreeHundredTwentyFive is the state of the world today.
//
// Nothing in cohere implements a React compiler rule, so the honest score is zero. What this test
// actually guards is the denominator: a zero over 325 considered is a measurement, and a zero over
// nothing considered is an empty suite wearing the same digit. The assertions on Considered and
// Declined are what separate them, and `Score.Check` proves the buckets partition the corpus rather
// than merely summing to something.
func TestFirstHonestScoreIsZeroOfThreeHundredTwentyFive(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	score, err := Run(fixtures, NothingImplemented{}, Options{})
	if err != nil {
		t.Fatalf("scoring the unimplemented engine: %v", err)
	}

	if score.Considered != ExpectedFixtureCount {
		t.Errorf("considered %d fixtures, want %d — a zero over a smaller denominator is not the same measurement",
			score.Considered, ExpectedFixtureCount)
	}
	if score.Passed != 0 {
		t.Errorf("passed %d, want 0; nothing implements these rules yet", score.Passed)
	}
	// Declined, not Failed. NothingImplemented says "I do not handle this" rather than "I found no
	// errors", and those are different facts about the same absent implementation.
	if score.Declined != ExpectedFixtureCount {
		t.Errorf("declined %d, want %d", score.Declined, ExpectedFixtureCount)
	}
	if score.Refused != 0 {
		t.Errorf("refused %d, want 0; a non-zero refusal count is a defect in this package, not a score", score.Refused)
	}

	t.Logf("first honest score: %s", score.Summary())
}

// TestFlowExclusionIsCountedNotDropped checks the excluded fixtures stay in the denominator.
//
// The failure mode being guarded is a suite that reports "290 / 290 declined" and reads like full
// coverage. Excluding must move a fixture between columns, never out of the total.
func TestFlowExclusionIsCountedNotDropped(t *testing.T) {
	t.Parallel()
	fixtures := load(t)

	score, err := Run(fixtures, NothingImplemented{}, Options{SkipFlowFixtures: true})
	if err != nil {
		t.Fatalf("scoring with Flow fixtures excluded: %v", err)
	}

	if score.Considered != ExpectedFixtureCount {
		t.Errorf("considered %d, want %d; excluding must not shrink the denominator",
			score.Considered, ExpectedFixtureCount)
	}
	if score.Excluded != ExpectedFlowFixtureCount {
		t.Errorf("excluded %d, want %d", score.Excluded, ExpectedFlowFixtureCount)
	}
	if score.Declined != ExpectedFixtureCount-ExpectedFlowFixtureCount {
		t.Errorf("declined %d, want %d", score.Declined, ExpectedFixtureCount-ExpectedFlowFixtureCount)
	}
	t.Logf("with Flow excluded: %s", score.Summary())
}

// TestVendoredCorpusPairsAreComplete guards the vendored tree against silent drift.
//
// Vendoring buys hermeticity and pays for it with the risk that the checked-in tree stops matching
// the sha it claims. A pair going missing — a bad merge, a partial re-vendor, a file deleted by
// something walking the tree — would shrink the denominator, and every other test here is written
// against counts, so most of them would fail loudly. This one names the failure directly rather
// than leaving it to be inferred from an arithmetic mismatch.
//
// It checks structure rather than content: every input has its expectation, every expectation has
// its input, and nothing else is in the directory. Content is guaranteed by `Load` parsing each
// golden, and against upstream by re-running `internal/lint/rules/react/tools/vendor_fixtures` and diffing, which is how
// the current tree was verified byte-for-byte when it was written.
func TestVendoredCorpusPairsAreComplete(t *testing.T) {
	t.Parallel()
	var inputs, expectations []string
	err := filepath.WalkDir(fixtureRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".expect.md") {
			expectations = append(expectations, path)
			return nil
		}
		inputs = append(inputs, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walking the vendored corpus: %v", err)
	}

	const wantPairs = ExpectedErrorFixtureCount + ExpectedCleanFixtureCount
	if len(inputs) != wantPairs || len(expectations) != wantPairs {
		t.Errorf("vendored tree holds %d inputs and %d expectations, want %d of each",
			len(inputs), len(expectations), wantPairs)
	}

	haveExpectation := map[string]bool{}
	for _, expectation := range expectations {
		haveExpectation[expectation] = true
	}
	for _, input := range inputs {
		want := strings.TrimSuffix(input, filepath.Ext(input)) + ".expect.md"
		if !haveExpectation[want] {
			t.Errorf("%s has no expectation; upstream pairs these 1:1", input)
		}
		delete(haveExpectation, want)
	}
	for orphan := range haveExpectation {
		t.Errorf("%s has no input", orphan)
	}
}

// TestKnownAndIgnoredPragmasDoNotOverlap keeps the two maps from disagreeing about a directive.
//
// A key in both would resolve by whichever map is consulted first, which is an implementation
// detail rather than a decision, and it would silently mean "modelled" and "deliberately inert" at
// once. CheckPragmas consults knownPragmas first, so an overlap would make the ignoredPragmas entry
// and the evidence written next to it unreachable.
func TestKnownAndIgnoredPragmasDoNotOverlap(t *testing.T) {
	t.Parallel()
	ignored := map[string]bool{}
	for _, key := range IgnoredPragmaKeys() {
		ignored[key] = true
	}
	for _, key := range KnownPragmaKeys() {
		if ignored[key] {
			t.Errorf("%q is both modelled and deliberately ignored; one of the two entries is unreachable", key)
		}
	}

	// 25 modelled and 6 ignored, summing to the 31 distinct directives the corpus uses. Every
	// modelled key was checked against upstream's own schemas when it was written — the 15
	// PluginOptions keys in `Entrypoint/Options.ts` and the 40 in `EnvironmentConfigSchema` — so
	// none of them is a directive this package invented a meaning for.
	if got := len(KnownPragmaKeys()) + len(IgnoredPragmaKeys()); got != 34 {
		t.Errorf("modelled plus ignored = %d, want 34 to match the corpus's distinct directives", got)
	}
}
