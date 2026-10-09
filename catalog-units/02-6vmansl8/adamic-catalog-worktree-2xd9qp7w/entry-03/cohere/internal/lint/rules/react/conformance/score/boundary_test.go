package react_conformance_score

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// TestConfigFixturesNeedAModuleTypeProvider holds the `config` divergence in place.
//
// `statedDivergences` claims all four Config goldens depend on a `moduleTypeProvider` cohere has no
// channel to receive. A claim in a map is a label; this is what makes it a measurement. It asserts
// the mechanism (every fixture imports a module that exists only in upstream's test harness) rather
// than the symptom (the rule reports nothing), because the symptom is equally satisfied by a rule
// that is simply broken.
//
// If cohere ever grows a type-provider channel and these start reporting, this fails and the
// divergence entries have to be revisited deliberately.
func TestConfigFixturesNeedAModuleTypeProvider(t *testing.T) {
	t.Parallel()
	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	// The modules upstream's `shared-runtime-type-provider.ts` invents. Nothing resolves them.
	harnessOnly := []string{"ReactCompilerTest", "useDefaultExportNotTypedAsHook"}

	selected := SelectFixtures(fixtures, "config")
	if len(selected) != 4 {
		t.Fatalf("selected %d config fixtures, want 4", len(selected))
	}
	for _, fixture := range selected {
		found := false
		for _, module := range harnessOnly {
			if strings.Contains(fixture.Source, "'"+module+"'") {
				found = true
			}
		}
		if !found {
			t.Errorf("%s imports no harness-only module, so the stated divergence does not explain it", fixture.Name)
		}
		if _, recorded := react_conformance.StatedDivergenceNames()[fixture.Name]; !recorded {
			t.Errorf("%s is a config fixture with no statedDivergences entry", fixture.Name)
		}
	}
}

// TestOptInPragmaFixturesAreEnumerated holds the feature-flag divergences in place.
//
// Three goldens in the wired set were recorded under an upstream flag defaulting to false. The
// assertion is the ENUMERATION rather than a count: if a fourth such fixture appears, or one of
// these loses its pragma, this fails rather than the score quietly absorbing it.
//
// The defaults were read from React's own unminified bundle
// (`eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js`), not inferred from the
// corpus, because inferring "this flag must be off" from "upstream reported something" is circular.
func TestOptInPragmaFixturesAreEnumerated(t *testing.T) {
	t.Parallel()
	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	// Measured as `z.boolean().default(false)` in React 7.1.1.
	optIn := map[string]bool{
		"enableUseKeyedState":                     true,
		"enableTreatSetIdentifiersAsStateSetters": true,
	}

	want := map[string]bool{
		"error.invalid-setstate-unconditional-with-keyed-state.js":       true,
		"error.invalid-unconditional-set-state-prop-in-render.js":        true,
		"error.invalid-unconditional-set-state-hook-return-in-render.js": true,
	}

	got := map[string]bool{}
	for _, upstream := range UpstreamNames() {
		for _, fixture := range SelectFixtures(fixtures, upstream) {
			for _, pragma := range fixture.Pragmas {
				if optIn[pragma.Key] {
					got[fixture.Name] = true
				}
			}
		}
	}

	for name := range want {
		if !got[name] {
			t.Errorf("%s no longer carries an opt-in pragma; its stated divergence may be stale", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("%s carries an opt-in pragma and is not enumerated; it needs a divergence entry or a decision", name)
		}
	}
}
