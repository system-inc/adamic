package registry

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/lint/configuration"
)

// The live config, which is the only thing that decides whether a registered rule ever runs.
// liveConfigPath is ahra's config, which is private, so the test that reads it skips where the ahra corpus
// is not set, and the gate names it as not covered (#sycrdr6).
func liveConfigPath(t testing.TB) string {
	t.Helper()
	return corpus.Ahra.Path(t, "CohereSettings.json")
}

// A rule's fixture proves it works. The config decides whether it runs, and nothing else connects
// the two: `rule_testing` never reads `CohereSettings.json`, so a rule can pass both directions of its own
// pair and be inert on every real file.
//
// That happened. The first `@next/next` rule registered as `next-no-assign-module-variable` while
// the config said `nextjs/no-assign-module-variable`, which resolves on a `/` boundary to
// `no-assign-module-variable`. The names did not match, the rule ran on nothing, and its tests were
// green. Only the coverage line caught it, and only because a second sentence happened to print.
//
// This is the mechanical version: every registered rule must be a name the live config can resolve,
// or be listed below as deliberately not enabled.
func TestEveryRegisteredRuleIsReachableFromTheLiveConfig(t *testing.T) {
	t.Parallel()

	loaded, err := configuration.Load(liveConfigPath(t))
	if err != nil {
		t.Fatalf("loading the live config: %v", err)
	}
	resolved := loaded.Resolve("app/Probe.tsx")

	// Rules cohere implements that the live config deliberately leaves unconfigured. Each needs a reason,
	// because an entry here silences the guard for that rule permanently, and the reverse check below
	// fails an entry the config has since configured or that names no rule.
	//
	// It held 70 more until #sycrdr6 found every one of them configured or naming nothing, so silencing
	// nothing while reading as if it did. Their enabling-cost counts and porting notes are archived
	// verbatim on #e4cn6wq (deliberately-not-enabled-archive.md); the decisions themselves live in the
	// sets' reasons and departures.
	deliberatelyNotEnabled := map[string]string{
		// Ported and registered without being enabled, because enabling it is a decision with work
		// attached rather than a wiring step. The audit measured 54 violations and the rule has no
		// fixer, so every one is a hand edit; the config has never named it under either spelling.
		"@typescript-eslint/no-deprecated": "ported and registered; the audit measured 54 violations and the rule has no fixer, so enabling is a decision for whoever takes that cleanup",
	}

	rules := All()
	if len(rules) == 0 {
		t.Fatal("the registry is empty, so this test proves nothing")
	}

	var unreachable []string
	for _, subject := range rules {
		if _, excused := deliberatelyNotEnabled[subject.Name]; excused {
			continue
		}
		// `StatusOf` rather than `Enabled`, because Enabled collapses two different worlds into
		// one false. A rule the config turns off is a decision somebody made and recorded; a rule
		// the config never mentions is a wiring gap. Twelve rules landed deliberately unenabled
		// tonight, each honouring a standing `off`, and this guard was reporting three of them as
		// running on nothing alongside genuine gaps. That is a false positive on correct work, and
		// a guard that cries wolf on the right answer gets ignored on the wrong one.
		status, _ := resolved.StatusOf(subject.Name)
		if status == configuration.StatusUnconfigured {
			unreachable = append(unreachable, subject.Name)
		}
	}

	// An exemption is a claim about the world, and the world moves, so each is checked rather than trusted.
	// One that names no registered rule matches nothing and silences nothing it says it does; one the live
	// config now configures is a decision the config has already made. Either way the entry outlived its
	// premise. The last one found was import-require-path-alias: keyed by a name no rule has had since the
	// naming scheme, its reason checked against an oxlint plugin that left with the gate, and the check
	// skipped silently once the plugin was gone (#sycrdr6).
	registered := map[string]bool{}
	for _, subject := range rules {
		registered[subject.Name] = true
	}
	var staleExemptions []string
	for name := range deliberatelyNotEnabled {
		if !registered[name] {
			staleExemptions = append(staleExemptions, name+": no registered rule has this name")
			continue
		}
		if status, _ := resolved.StatusOf(name); status != configuration.StatusUnconfigured {
			staleExemptions = append(staleExemptions, name+": the live config configures it")
		}
	}
	sort.Strings(staleExemptions)
	if len(staleExemptions) > 0 {
		t.Errorf("%d entries in deliberatelyNotEnabled no longer hold, so remove each one:\n  %s",
			len(staleExemptions), strings.Join(staleExemptions, "\n  "))
	}

	if len(unreachable) > 0 {
		t.Errorf(
			"%d registered rules are not mentioned by the live config, so they run on nothing and "+
				"nobody has said whether they should: %v\n"+
				"a rule whose name the config cannot resolve passes its own fixtures and lints no "+
				"files. A rule the config explicitly turns off is not this: that is a decision, and "+
				"it is reported separately by the run itself as scoped off",
			len(unreachable), unreachable,
		)
	}
}
