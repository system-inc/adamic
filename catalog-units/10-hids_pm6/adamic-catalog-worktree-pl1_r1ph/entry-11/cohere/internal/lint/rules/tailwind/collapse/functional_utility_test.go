package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// functionalFixture is one probe captured by internal/lint/rules/tailwind/tools/generate_functional/enumerate.mjs.
type functionalFixture struct {
	ProbeCount    int `json:"probeCount"`
	RejectedCount int `json:"rejectedCount"`
	Probes        []struct {
		ClassName    string `json:"className"`
		Branch       string `json:"branch"`
		Declarations []struct {
			Property string  `json:"property"`
			Value    *string `json:"value"`
		} `json:"declarations"`
	} `json:"probes"`
}

func loadFunctionalFixture(t *testing.T, name string) functionalFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	var fixture functionalFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	if fixture.ProbeCount == 0 {
		t.Fatalf("%s holds no probes, so every assertion over it passes by comparing nothing", name)
	}
	return fixture
}

// TestFunctionalResolutionRejectsWhatTheEngineRejects is the half of the fixture that is unambiguous.
//
// Resolution is an intermediate and the fixture observes it through the declarations a class
// compiles to, which is a weaker signal than the other generators produce: a handler can transform
// the value on its way out. The rejections are not weak in that way. A class the engine compiles to
// no CSS at all is a class this pipeline must refuse to resolve, with nothing to interpret, so these
// carry the test.
//
// Three of the four are the modifier-cancels rule in its three positions, which is the detail most
// likely to be dropped in a port: upstream spells it three separate times and each one guards a
// different branch.
func TestFunctionalResolutionRejectsWhatTheEngineRejects(t *testing.T) {
	t.Parallel()
	fixture := loadFunctionalFixture(t, "functional_fixtures.json")

	if fixture.RejectedCount < 4 {
		t.Fatalf("the fixture carries %d rejected probes; expected at least 4, so the branches that must produce nothing are under-exercised", fixture.RejectedCount)
	}

	// Each rejected probe is reproduced as the candidate shape the parser produces for it, so the
	// assertion is against this pipeline rather than against the parser, which has its own tests.
	byClassName := map[string]func() (ParsedCandidate, *FunctionalUtilityDescription, bool){
		"bg/50":      rejectedValuelessWithModifier,
		"w-[3px]/50": rejectedArbitraryWithModifier,
		"grow-2/3":   rejectedBareWithModifier,
		"w-1.5/2":    rejectedNonIntegerFraction,
	}

	rejections, covered := 0, 0
	for _, probe := range fixture.Probes {
		if probe.Declarations != nil {
			continue
		}
		rejections++
		build, known := byClassName[probe.ClassName]
		if !known {
			t.Errorf("%s (%s) is rejected by the engine and this test has no case for it, so the rejection is unchecked", probe.ClassName, probe.Branch)
			continue
		}
		covered++
		candidate, description, negative := build()
		if _, produced := ResolveFunctionalUtilityValue(&candidate, description, NewTheme(), negative); produced {
			t.Errorf("%s (%s): the engine compiles it to no CSS and the pipeline resolved it to a value", probe.ClassName, probe.Branch)
		}
	}
	if covered != rejections {
		t.Errorf("covered %d of %d rejections; an unchecked rejection is a branch this file claims to test and does not", covered, rejections)
	}
	t.Logf("checked %d rejections of %d probes", rejections, fixture.ProbeCount)
}

func bareValueDescription() *FunctionalUtilityDescription {
	return &FunctionalUtilityDescription{
		HandleBareValue: func(value *ParsedValue) (string, bool) { return value.Value, true },
	}
}

func rejectedValuelessWithModifier() (ParsedCandidate, *FunctionalUtilityDescription, bool) {
	return ParsedCandidate{
		Kind:     ParsedCandidateKindFunctional,
		Root:     "bg",
		Modifier: &ParsedModifier{Kind: ParsedModifierKindNamed, Value: "50"},
	}, &FunctionalUtilityDescription{DefaultValue: "1", DefaultValuePresent: true}, false
}

func rejectedArbitraryWithModifier() (ParsedCandidate, *FunctionalUtilityDescription, bool) {
	return ParsedCandidate{
		Kind:     ParsedCandidateKindFunctional,
		Root:     "w",
		Value:    &ParsedValue{Kind: ParsedValueKindArbitrary, Value: "3px"},
		Modifier: &ParsedModifier{Kind: ParsedModifierKindNamed, Value: "50"},
	}, &FunctionalUtilityDescription{}, false
}

func rejectedBareWithModifier() (ParsedCandidate, *FunctionalUtilityDescription, bool) {
	return ParsedCandidate{
		Kind:     ParsedCandidateKindFunctional,
		Root:     "grow",
		Value:    &ParsedValue{Kind: ParsedValueKindNamed, Value: "2"},
		Modifier: &ParsedModifier{Kind: ParsedModifierKindNamed, Value: "3"},
	}, bareValueDescription(), false
}

func rejectedNonIntegerFraction() (ParsedCandidate, *FunctionalUtilityDescription, bool) {
	return ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "w",
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "1.5", Fraction: "1.5/2"},
	}, &FunctionalUtilityDescription{SupportsFractions: true, ThemeKeys: []string{"--width"}}, false
}

// TestFunctionalResolutionModifierCancelsBareValue pins the rule three upstream branches share.
//
// `if (!value?.includes('/') && candidate.modifier) return`. A bare handler that returns a value
// without a slash cancels the whole utility when a modifier is present, which is what stops
// `grow-2/3` from reading as `grow` carrying a stray modifier. Dropping the test leaves those
// classes resolving to something, which sorts, so it would not surface as an error anywhere.
func TestFunctionalResolutionModifierCancelsBareValue(t *testing.T) {
	t.Parallel()
	theme := NewTheme()
	description := &FunctionalUtilityDescription{
		HandleBareValue: func(value *ParsedValue) (string, bool) { return value.Value, true },
	}

	withModifier := ParsedCandidate{
		Kind:     ParsedCandidateKindFunctional,
		Root:     "grow",
		Value:    &ParsedValue{Kind: ParsedValueKindNamed, Value: "2"},
		Modifier: &ParsedModifier{Kind: ParsedModifierKindNamed, Value: "3"},
	}
	if _, produced := ResolveFunctionalUtilityValue(&withModifier, description, theme, false); produced {
		t.Error("a bare value without a slash resolved while a modifier was present; the engine produces nothing")
	}

	withSlash := withModifier
	withSlash.Value = &ParsedValue{Kind: ParsedValueKindNamed, Value: "2/3"}
	if _, produced := ResolveFunctionalUtilityValue(&withSlash, description, theme, false); !produced {
		t.Error("a bare value containing a slash was cancelled by its modifier; only slashless values are")
	}
}

// TestFunctionalResolutionNegatesThroughMathOperatorSpacing pins the string, not only the arithmetic.
//
// The engine emits `calc(10 * -1)` for `-z-10`, measured in the fixture. Emitting `calc(10*-1)` is
// the same CSS and a different string, and since this port compares readings rather than text it
// would pass every assertion in this package while disagreeing with the engine anywhere the text
// reaches a data-type inference.
func TestFunctionalResolutionNegatesThroughMathOperatorSpacing(t *testing.T) {
	t.Parallel()
	theme := NewTheme()
	description := &FunctionalUtilityDescription{
		HandleBareValue: func(value *ParsedValue) (string, bool) { return value.Value, true },
	}
	candidate := ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "z",
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "10"},
	}

	resolved, produced := ResolveFunctionalUtilityValue(&candidate, description, theme, true)
	if !produced {
		t.Fatal("a negative bare value produced nothing")
	}
	if resolved.Value != "calc(10 * -1)" {
		t.Errorf("negation produced %q, the engine produces %q", resolved.Value, "calc(10 * -1)")
	}
}

// TestFunctionalResolutionFractionMatchesTheEngineExactly pins branch 4 against the measured string.
//
// The fixture records `w-1/2` compiling to `calc(1 / 2 * 100%)`. The spacing is upstream's own and a
// port emitting `calc(1/2 * 100%)` would be the same CSS and a different string, in a value that
// reaches `inferDataType` downstream.
func TestFunctionalResolutionFractionMatchesTheEngineExactly(t *testing.T) {
	t.Parallel()
	theme := NewTheme()
	description := &FunctionalUtilityDescription{SupportsFractions: true, ThemeKeys: []string{"--width"}}
	candidate := ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "w",
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "1", Fraction: "1/2"},
	}

	resolved, produced := ResolveFunctionalUtilityValue(&candidate, description, theme, false)
	if !produced {
		t.Fatal("a supported fraction produced nothing")
	}
	if resolved.Value != "calc(1 / 2 * 100%)" {
		t.Errorf("fraction produced %q, the engine produces %q", resolved.Value, "calc(1 / 2 * 100%)")
	}

	// A non-integer half returns rather than falling through, which is upstream returning inside the
	// branch. Falling through would reach the bare handler and resolve `1.5` as an ordinary value.
	nonInteger := candidate
	nonInteger.Value = &ParsedValue{Kind: ParsedValueKindNamed, Value: "1.5", Fraction: "1.5/2"}
	if _, produced := ResolveFunctionalUtilityValue(&nonInteger, description, theme, false); produced {
		t.Error("a fraction with a non-integer numerator resolved; the engine compiles w-1.5/2 to nothing")
	}
}

// TestFunctionalResolutionValuelessModifierProducesNothing pins branches 1 and 2.
//
// A candidate with no value and a modifier, and an arbitrary value with a modifier, both produce
// nothing. Measured: `bg/50` and `w-[3px]/50` compile to no CSS.
func TestFunctionalResolutionValuelessModifierProducesNothing(t *testing.T) {
	t.Parallel()
	theme := NewTheme()
	theme.Add("--radius", "0.25rem", 0)
	description := &FunctionalUtilityDescription{
		ThemeKeys:           []string{"--radius"},
		DefaultValue:        "1",
		DefaultValuePresent: true,
	}
	modifier := &ParsedModifier{Kind: ParsedModifierKindNamed, Value: "50"}

	valueless := ParsedCandidate{Kind: ParsedCandidateKindFunctional, Root: "bg", Modifier: modifier}
	if _, produced := ResolveFunctionalUtilityValue(&valueless, description, theme, false); produced {
		t.Error("a valueless candidate with a modifier resolved; the engine compiles bg/50 to nothing")
	}

	arbitrary := ParsedCandidate{
		Kind:     ParsedCandidateKindFunctional,
		Root:     "w",
		Value:    &ParsedValue{Kind: ParsedValueKindArbitrary, Value: "3px"},
		Modifier: modifier,
	}
	if _, produced := ResolveFunctionalUtilityValue(&arbitrary, description, theme, false); produced {
		t.Error("an arbitrary value with a modifier resolved; the engine compiles w-[3px]/50 to nothing")
	}
}

// TestFunctionalResolutionThemeBeforeStaticValues is refinement 2 in its own test.
//
// A name that is both a theme key and a `staticValues` entry must resolve through the theme, since
// the theme is branch 3 and staticValues is branch 5. Getting this backwards is invisible on every
// name that is only one of the two, which is nearly all of them.
func TestFunctionalResolutionThemeBeforeStaticValues(t *testing.T) {
	t.Parallel()
	theme := NewTheme()
	theme.Add("--width-full", "60rem", 0)

	description := &FunctionalUtilityDescription{
		ThemeKeys:        []string{"--width"},
		StaticValueNames: map[string]bool{"full": true},
	}
	candidate := ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "w",
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "full"},
	}

	resolved, produced := ResolveFunctionalUtilityValue(&candidate, description, theme, false)
	if !produced {
		t.Fatal("a name in both the theme and staticValues produced nothing")
	}
	if resolved.IsStaticValue {
		t.Error("resolution ended at staticValues for a name the theme declares; the theme is branch 3 and staticValues is branch 5")
	}
	if resolved.Value != "var(--width-full)" {
		t.Errorf("resolved to %q, expected the theme's own var() reference", resolved.Value)
	}
}

// TestIsPositiveIntegerMatchesTheEngineIncludingPrecision pins the predicate against measured answers.
//
// Upstream is `Number.isInteger(num) && num >= 0 && String(num) === String(value)`, and the
// round-trip clause is the whole predicate. Every case below was run against that JavaScript rather
// than reasoned about, because the interesting ones are where a digits-only scan agrees by accident.
//
// `9007199254740993` is the one that separates them. It is past 2^53, so `Number(value)` rounds it
// and `String(num)` prints the rounded form, and upstream returns false where a digit scan returns
// true. It is the only disagreement in this corpus, which is exactly why it needs to be here: a port
// that dropped the round-trip would pass every other case in this test.
func TestIsPositiveIntegerMatchesTheEngineIncludingPrecision(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"0": true, "1": true, "10": true,
		"007": false, "1.0": false, "1e3": false, " 1": false, "": false,
		"1 ": false, "+1": false, "-1": false, "0.5": false,
		"9007199254740993": false, "1_0": false, "Infinity": false, "0x10": false,
	}
	for input, expected := range cases {
		if actual := isPositiveInteger(input); actual != expected {
			t.Errorf("isPositiveInteger(%q) = %v, the engine says %v", input, actual, expected)
		}
	}

	// The strict form differs from the loose one on exactly one input, so that is what is asserted
	// rather than the whole corpus a second time.
	if isStrictPositiveInteger("0") {
		t.Error(`isStrictPositiveInteger("0") = true, the engine says false`)
	}
	if !isStrictPositiveInteger("1") {
		t.Error(`isStrictPositiveInteger("1") = false, the engine says true`)
	}
}
