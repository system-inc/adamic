package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The fixture corpus is the record of what the engine actually does, captured from a real Tailwind
// 4.3.3 design system built from the real theme.css, and checked in as JSON next to this test.
//
// It is generated rather than transcribed, because a hand-copied expectation is a claim about the
// engine rather than a measurement of it, and every one of the three approaches that lost findings
// during the migration looked correct when read.

// fixtureCorpus is the shape of testdata/gate_fixtures.json.
type fixtureCorpus struct {
	TailwindVersion string        `json:"tailwindVersion"`
	Fixtures        []gateFixture `json:"fixtures"`
}

// gateFixture is one literal, what the engine said about it, and the bucket key per class.
type gateFixture struct {
	Literal string   `json:"literal"`
	Classes []string `json:"classes"`
	// EngineCollapses is what the real engine reported. Empty means the engine stayed silent.
	EngineCollapses []engineCollapse `json:"engineCollapses"`
	// BucketKeys is the key the JavaScript gate computed per class, which the Go gate must match
	// exactly. Matching the verdict through different keys would be luck rather than a port.
	BucketKeys map[string]string `json:"bucketKeys"`
	// ParsedCandidates feeds the recorded parser below, so the Go gate is exercised against the
	// engine's real readings rather than against a Go reimplementation of parsing.
	ParsedCandidates map[string][]Candidate `json:"parsedCandidates"`
	// DeclaredValues is what each class compiles to, for the static bucketing path.
	DeclaredValues map[string][]string `json:"declaredValues"`
	// DeclaredProperties is the CSS property names a class declares, which is a different thing
	// from its values and is what the third rejected approach keyed on. Carried only so the
	// known-dirty control below can reproduce that mistake exactly: `w-8` declares `width` and `h-8`
	// declares `height`, which is why filtering on property family stopped reporting `size-8`.
	DeclaredProperties map[string][]string `json:"declaredProperties"`
}

type engineCollapse struct {
	Input  []string `json:"input"`
	Output string   `json:"output"`
}

func (f gateFixture) engineFoundCollapse() bool {
	return len(f.EngineCollapses) > 0
}

// recordedParser answers from what the real engine said, which is what makes this a test of the
// gate's bucketing rather than a test of a Go parser that does not exist.
//
// Parsing a class correctly needs the theme, so the gate deliberately takes the parse as input. The
// question these fixtures answer is the one the gate actually owns: given true readings, does the
// bucketing keep every collapse?
type recordedParser struct {
	candidates map[string][]Candidate
	declared   map[string][]string
}

func (p recordedParser) ParseCandidates(className string) []Candidate {
	return p.candidates[className]
}

func (p recordedParser) DeclaredValues(className string) []string {
	return p.declared[className]
}

func loadFixtures(t *testing.T) fixtureCorpus {
	t.Helper()

	path := filepath.Join("testdata", "gate_fixtures.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	var corpus fixtureCorpus
	if err := json.Unmarshal(contents, &corpus); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	// A fixture file that loaded nothing would let every assertion below pass for the wrong reason.
	// This is the "prove the input is real" half: a vacuous corpus and a clean one are
	// indistinguishable from inside the test.
	if len(corpus.Fixtures) < 20 {
		t.Fatalf("expected a real fixture corpus, got %d fixtures", len(corpus.Fixtures))
	}

	return corpus
}

func gateForFixture(fixture gateFixture) *Gate {
	return NewGate(recordedParser{candidates: fixture.ParsedCandidates, declared: fixture.DeclaredValues})
}

// TestCorpusDiscriminates proves the corpus can fail before anything is measured on it.
//
// A fixture set where nothing collapses would let a gate that never asks the engine score perfectly.
// The tree this rule guards is already canonical, which is exactly the condition under which a
// broken gate looks correct, so the corpus has to carry known-collapsing cases and this has to
// confirm they are there.
func TestCorpusDiscriminates(t *testing.T) {
	t.Parallel()
	corpus := loadFixtures(t)

	collapsing := 0
	silent := 0
	for _, fixture := range corpus.Fixtures {
		if fixture.engineFoundCollapse() {
			collapsing++
		} else {
			silent++
		}
	}

	if collapsing == 0 {
		t.Fatal("no fixture collapses, so a gate that never asks the engine would pass every test here")
	}
	if silent == 0 {
		t.Fatal("no fixture stays silent, so a gate that always asks the engine would pass every test here")
	}

	t.Logf("corpus: %d collapsing, %d silent, tailwind %s", collapsing, silent, corpus.TailwindVersion)
}

// TestGateNeverLosesACollapse is the assertion the whole package exists to satisfy.
//
// For every literal the real engine found a collapse in, the gate must send it to the engine. A
// failure here is the silent kind: the rule keeps running, the tree keeps looking clean, and a real
// finding stops being reported.
func TestGateNeverLosesACollapse(t *testing.T) {
	t.Parallel()
	corpus := loadFixtures(t)

	checked := 0
	for _, fixture := range corpus.Fixtures {
		if !fixture.engineFoundCollapse() {
			continue
		}
		checked++

		gate := gateForFixture(fixture)
		if !gate.NeedsEngine(fixture.Classes, nil) {
			t.Errorf("gate suppressed %q, but the engine collapses it into %v",
				fixture.Literal, collapseOutputs(fixture))
		}
	}

	if checked == 0 {
		t.Fatal("checked no collapsing fixtures, so this test proved nothing")
	}
}

// TestGateMatchesRecordedBucketKeys checks the mechanism, not just the verdict.
//
// A gate can reach the right answer through wrong keys and stay right only by accident. The keys are
// what a future change breaks first, so they are asserted directly against what the JavaScript gate
// computed on the same classes.
func TestGateMatchesRecordedBucketKeys(t *testing.T) {
	t.Parallel()
	corpus := loadFixtures(t)

	compared := 0
	for _, fixture := range corpus.Fixtures {
		gate := gateForFixture(fixture)
		for className, wantKey := range fixture.BucketKeys {
			gotKey := gate.GroupKey(className)
			if gotKey != wantKey {
				t.Errorf("%s: bucket key %q, want %q", className, gotKey, wantKey)
			}
			compared++
		}
	}

	if compared == 0 {
		t.Fatal("compared no bucket keys, so this test proved nothing")
	}
	t.Logf("compared %d bucket keys", compared)
}

// TestGateSkipsWhatItCan records the gate's value rather than only its safety.
//
// Safety alone is satisfiable by a gate that always says yes, which would be correct and useless.
// This asserts the silent fixtures are actually skipped, so a regression toward always-ask shows up
// as a failure rather than as a slow run nobody notices.
func TestGateSkipsWhatItCan(t *testing.T) {
	t.Parallel()
	corpus := loadFixtures(t)

	skipped := 0
	asked := 0
	for _, fixture := range corpus.Fixtures {
		if fixture.engineFoundCollapse() {
			continue
		}

		gate := gateForFixture(fixture)
		if gate.NeedsEngine(fixture.Classes, nil) {
			asked++
			t.Logf("gate still asks about silent literal %q (safe, but costs a call)", fixture.Literal)
		} else {
			skipped++
		}
	}

	if skipped == 0 {
		t.Fatal("the gate skipped nothing, so it is not a gate")
	}
	t.Logf("silent fixtures: %d skipped, %d still asked", skipped, asked)
}

// TestKnownWrongGatesLoseFindings is the known-dirty control.
//
// Each of these three shortcuts was tried during the migration, measured faster, and silently
// stopped reporting a real collapse. If any of them now passes the corpus, the corpus has stopped
// discriminating and every other test in this file is worth less than it appears.
//
// A guard that has never returned a positive has not been shown to work, so these deliberately
// broken gates are run against the same fixtures and required to fail.
func TestKnownWrongGatesLoseFindings(t *testing.T) {
	t.Parallel()
	corpus := loadFixtures(t)

	testCases := []struct {
		name string
		// key is the broken bucketing being demonstrated.
		key func(fixture gateFixture, className string) string
		// wantLoses names a literal this shortcut is known to lose.
		wantLoses string
	}{
		{
			// A dash-splitter reads `border-l` as root `border` value `l` and `border-r` as root
			// `border` value `r`, so their values differ and they never share a bucket.
			name:      "dash splitting for the root",
			key:       dashSplitKey,
			wantLoses: "border-l border-r",
		},
		{
			// Each static in its own bucket separates the two overflow axes.
			name:      "one bucket per static name",
			key:       staticNameKey,
			wantLoses: "overflow-x-hidden overflow-y-hidden",
		},
		{
			// Requiring a shared declared-property family separates `width` from `height`.
			name:      "shared declared property family",
			key:       declaredPropertyFamilyKey,
			wantLoses: "w-8 h-8",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			lost := brokenGateLosses(corpus, testCase.key)

			if len(lost) == 0 {
				t.Fatalf("the %q shortcut lost nothing on this corpus, so the corpus no longer "+
					"discriminates and the passing tests above prove less than they appear to",
					testCase.name)
			}

			if !contains(lost, testCase.wantLoses) {
				t.Errorf("expected the %q shortcut to lose %q, it lost %v",
					testCase.name, testCase.wantLoses, lost)
			}

			// The real gate must keep exactly what the broken one drops.
			for _, literal := range lost {
				fixture := fixtureByLiteral(corpus, literal)
				gate := gateForFixture(fixture)
				if !gate.NeedsEngine(fixture.Classes, nil) {
					t.Errorf("the real gate also loses %q, which the %q shortcut was supposed to "+
						"be alone in losing", literal, testCase.name)
				}
			}
		})
	}
}

// brokenGateLosses returns the collapsing literals a bucketing scheme would suppress.
func brokenGateLosses(corpus fixtureCorpus, key func(gateFixture, string) string) []string {
	var lost []string

	for _, fixture := range corpus.Fixtures {
		if !fixture.engineFoundCollapse() {
			continue
		}

		// A single class that rewrites alone is not what these shortcuts break, and including it
		// would let a broken pair-bucketing scheme look correct because an unrelated path caught it.
		if len(fixture.Classes) < 2 {
			continue
		}

		seen := map[string]bool{}
		shares := false
		for _, className := range fixture.Classes {
			bucketKey := key(fixture, className)
			if seen[bucketKey] {
				shares = true
				break
			}
			seen[bucketKey] = true
		}
		if !shares {
			lost = append(lost, fixture.Literal)
		}
	}

	return lost
}

// dashSplitKey buckets on everything after the first dash, which is the hand-written splitter.
func dashSplitKey(fixture gateFixture, className string) string {
	for index := 0; index < len(className); index++ {
		if className[index] == '-' {
			return "value:" + className[index+1:]
		}
	}
	return "value:" + className
}

// staticNameKey gives every class its own bucket keyed on its name.
func staticNameKey(_ gateFixture, className string) string {
	return "name:" + className
}

// declaredPropertyFamilyKey buckets on the CSS property names a class declares.
//
// Property names, not values, and the distinction is the entire point of this control. `w-8` and
// `h-8` declare the same value (`calc(var(--spacing) * 8)`) under different properties (`width`,
// `height`), so a control keyed on values would not reproduce the mistake and would report that the
// shortcut is safe.
func declaredPropertyFamilyKey(fixture gateFixture, className string) string {
	properties := fixture.DeclaredProperties[className]
	if len(properties) == 0 {
		return "none:" + className
	}
	return "property:" + properties[0]
}

func fixtureByLiteral(corpus fixtureCorpus, literal string) gateFixture {
	for _, fixture := range corpus.Fixtures {
		if fixture.Literal == literal {
			return fixture
		}
	}
	return gateFixture{}
}

func collapseOutputs(fixture gateFixture) []string {
	outputs := make([]string, 0, len(fixture.EngineCollapses))
	for _, collapse := range fixture.EngineCollapses {
		outputs = append(outputs, collapse.Output)
	}
	return outputs
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// TestUnparseableClassesNeverGroup guards the case that has no engine answer at all.
//
// A project-specific class such as `background--2/50` is not something Tailwind can read. Grouping
// every unreadable class under one empty key would make them send each other to the engine forever,
// and grouping them with a real class would let one suppress a question about the other.
func TestUnparseableClassesNeverGroup(t *testing.T) {
	t.Parallel()
	parser := recordedParser{
		candidates: map[string][]Candidate{
			"mystery-one": {},
			"mystery-two": {},
		},
		declared: map[string][]string{},
	}
	gate := NewGate(parser)

	if gate.GroupKey("mystery-one") == gate.GroupKey("mystery-two") {
		t.Fatal("two unparseable classes shared a bucket, so each would send the other to the engine")
	}
	if gate.CouldRewriteAlone("mystery-one") {
		t.Fatal("an unparseable class cannot be canonicalized, so it is not worth an engine call")
	}
	if !gate.CannotCollapse([]string{"mystery-one", "mystery-two"}) {
		t.Fatal("two unparseable classes cannot collapse into each other")
	}
}

// TestRecordedSingleRewriteForcesAnEngineCall covers the feedback path.
//
// The gate can see that `z-[1]` is worth asking about but not that it is `z-1`. Once an authority
// says so, every later literal containing it must reach the engine, including ones where no two
// classes share a bucket.
func TestRecordedSingleRewriteForcesAnEngineCall(t *testing.T) {
	t.Parallel()
	parser := recordedParser{
		candidates: map[string][]Candidate{
			"z-[1]": {{Kind: CandidateKindFunctional, Root: "z", Value: Value{Kind: ValueKindArbitrary, Value: "1"}}},
			"flex":  {{Kind: CandidateKindStatic, Root: "flex"}},
		},
		declared: map[string][]string{"flex": {"flex"}},
	}
	gate := NewGate(parser)

	classes := []string{"z-[1]", "flex"}
	if !gate.NeedsEngine(classes, nil) {
		t.Fatal("an arbitrary value is always worth one engine call, because only the theme knows if it has a name")
	}

	// Once asked, it should not be asked again by the singles path.
	alreadyChecked := map[string]bool{"z-[1]": true, "flex": true}
	if len(gate.SingleClassesWorthChecking(classes, alreadyChecked)) != 0 {
		t.Fatal("a class already asked about should not be asked again")
	}

	// But a recorded rewrite must still force the call, because the literal genuinely has a finding.
	gate.RecordSingleRewrites([]string{"z-[1]"})
	if !gate.NeedsEngine(classes, alreadyChecked) {
		t.Fatal("a class known to rewrite alone must always reach the engine")
	}
}

// TestDuplicateClassesDoNotFakeACollapse guards a false positive rather than a false negative.
//
// `flex flex` is a different rule's finding. If the gate treated the repeat as two classes sharing a
// bucket, every duplicate in the tree would buy an engine call and report nothing.
func TestDuplicateClassesDoNotFakeACollapse(t *testing.T) {
	t.Parallel()
	parser := recordedParser{
		candidates: map[string][]Candidate{
			"flex": {{Kind: CandidateKindStatic, Root: "flex"}},
		},
		declared: map[string][]string{"flex": {"flex"}},
	}
	gate := NewGate(parser)

	if gate.NeedsEngine([]string{"flex", "flex"}, map[string]bool{"flex": true}) {
		t.Fatal("a repeated class is a duplicate, not a collapse")
	}
}
