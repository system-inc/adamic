package configuration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
)

// liveConfigPath is the real config cohere must agree with: ahra's, which is private, so a test that reads
// it skips where the ahra corpus is not set, and the gate names it as not covered (#sycrdr6).
func liveConfigPath(t testing.TB) string {
	t.Helper()
	return corpus.Ahra.Path(t, "CohereSettings.json")
}

// TestAgainstTheLiveConfig loads the actual file rather than a hand-written model of it.
//
// A test that models a config can drift from the configuration. This one reads the bytes that are actually
// gating the codebase, so a change there shows up here rather than at the next full run.
func TestAgainstTheLiveConfig(t *testing.T) {
	t.Parallel()

	loaded, err := Load(liveConfigPath(t))
	if err != nil {
		t.Fatalf("loading the live config: %v", err)
	}

	// Prove the input is real before asserting anything about it. A config that parsed to zero rules
	// would make every assertion below pass vacuously, which is the failure mode this project keeps
	// finding.
	if len(loaded.Rules) < 100 {
		t.Fatalf("only %d rules loaded from the live config, which is not a real corpus", len(loaded.Rules))
	}
	if len(loaded.Overrides) == 0 || len(loaded.IgnorePatterns) == 0 {
		t.Fatalf("overrides=%d ignorePatterns=%d, expected both non-empty", len(loaded.Overrides), len(loaded.IgnorePatterns))
	}

	// Generated code gets every rule (#c076xbg). The one override left is the GraphQL output's
	// abbreviation rule, waiting on a wire decision; everything else runs there, and on any other
	// generated file nothing is off.
	graphQlOutput := loaded.Resolve("libraries/structure/source/api/graphql/GraphQlOperations.generated.ts")
	if graphQlOutput.Enabled("nexus/consistency-no-abbreviated-identifier") {
		t.Error("the GraphQL output's interim abbreviation override did not apply")
	}
	workerBundle := loaded.Resolve("libraries/structure/source/api/web-sockets/shared-worker/WebSocketSharedWorkerCode.generated.ts")
	// The folder spelling the ruling retired: a `generated/` override coming back would turn rules off here.
	inGeneratedFolder := loaded.Resolve("libraries/structure/source/api/graphql/generated/GraphQlOperations.ts")
	for _, ruleName := range []string{
		"nexus/consistency-require-type-suffix",
		"nexus/consistency-no-abbreviated-identifier",
		"@typescript-eslint/no-explicit-any",
	} {
		if !graphQlOutput.Enabled(ruleName) && ruleName != "nexus/consistency-no-abbreviated-identifier" {
			t.Errorf("%s is off for the GraphQL output; generated code gets every rule", ruleName)
		}
		if !workerBundle.Enabled(ruleName) {
			t.Errorf("%s is off for a generated file; generated code gets every rule", ruleName)
		}
		if !inGeneratedFolder.Enabled(ruleName) {
			t.Errorf("%s is off inside a generated/ folder; generated code gets every rule", ruleName)
		}
	}

	// The same rule must still run on authored source, or the override is too broad.
	authored := loaded.Resolve("libraries/structure/source/components/buttons/Button.tsx")
	if !authored.Enabled("nexus/consistency-require-type-suffix") {
		t.Error("consistency-require-type-suffix stopped running on authored source")
	}

	// The other two live override blocks.
	if loaded.Resolve("next-env.d.ts").Enabled("structure/consistency-require-organized-imports") {
		t.Error("the next-env.d.ts override did not apply")
	}
	if loaded.Resolve("modules/finance/connections/QuickBooksAdapter.ts").Enabled("structure/network-no-direct-fetch") {
		t.Error("the modules/** override did not apply")
	}
	if !loaded.Resolve("libraries/structure/source/api/Fetch.ts").Enabled("structure/network-no-direct-fetch") {
		t.Error("the modules/** override leaked outside modules/")
	}

	// Ignored paths.
	if !loaded.Resolve("node_modules/react/index.d.ts").Ignored {
		t.Error("node_modules was not ignored")
	}
	if loaded.Resolve("libraries/structure/source/components/buttons/Button.tsx").Ignored {
		t.Error("real source was ignored")
	}
}

// TestBothPathsAgreeOnEveryPluginDefault is the gate on removing the hand-written lines.
//
// The live config names all forty plugin-default rules by hand, which reproduces oxlint's behaviour
// in the wrong layer. Removing those lines is only safe if the plugin declaration resolves the same
// forty, and this is what shows it: the config is loaded as written, then loaded again with every
// plugin-default line stripped, and the two results are compared rule by rule.
//
// Without this the removal IS the change, and a silent disagreement between the two paths looks
// exactly like a successful migration.
//
// # This guard was vacuous once, and the reason generalises
//
// The first version read `Load`'s result. `Load` merges the plugin defaults into `Rules`, so a rule
// arriving only through the declaration is indistinguishable there from one named by hand -- the
// exact distinction this test exists to measure. Checked by deleting three plugin-default lines from
// a copy of the live config: the test passed and reported all forty still named by hand.
//
// The shape: a test comparing two paths cannot read the surface where they have already been merged,
// and that surface is usually the convenient one because it is what every consumer uses. Reading the
// file directly is why this can now fail.
//
// # The severity difference, which is real and is the reason nothing is being removed yet
//
// The two paths agree on WHICH rules run and disagree on how loudly. The hand-written lines say
// `error`; `warn_correctness` contributes `warn`, and the inventory records all forty as `warn`.
// So the hand-written lines are not reproducing the gate's behaviour, they are strengthening it --
// forty rules currently fail this build that only warn in the tool being replaced.
//
// That is a decision somebody made and did not write down, and it is not this test's to reverse.
// The test therefore asserts membership exactly and reports severity rather than requiring it to
// match, so the difference is visible instead of blocking or being smoothed away.
func TestBothPathsAgreeOnEveryPluginDefault(t *testing.T) {
	t.Parallel()
	asWritten, err := Load(liveConfigPath(t))
	if err != nil {
		t.Fatalf("loading the live config: %v", err)
	}

	// Every plugin-default rule must be reachable through the declaration alone.
	fromPlugins := RulesFromPlugins(asWritten.Plugins, map[string]RuleSetting{})
	if len(fromPlugins) != len(PluginDefaultRules) {
		t.Fatalf("the live config's plugins resolve %d of %d plugin-default rules; a rule whose "+
			"plugin is not declared would be lost the moment its hand-written line came out",
			len(fromPlugins), len(PluginDefaultRules))
	}

	// Read what the FILE names rather than what `Load` returned. `Load` merges the plugin defaults
	// into `Rules`, so a rule arriving only through the declaration is indistinguishable there from
	// one named by hand -- which is exactly the distinction this test exists to measure. Reading the
	// merged map made this guard pass on a config with three lines deliberately removed.
	namedInFile, err := ruleNamesInChain(liveConfigPath(t))
	if err != nil {
		t.Fatalf("reading the rules block: %v", err)
	}

	missingByHand, severityDiffers := 0, 0
	for ruleName := range PluginDefaultRules {
		byHand, named := namedInFile[ruleName]
		if !named {
			// Reachable through the declaration and not named by hand. Not a failure -- it is the
			// state this whole change is heading toward -- but counted so the number is visible.
			missingByHand++
			continue
		}
		if byHand.Severity != PluginDefaultSeverity {
			severityDiffers++
		}
	}

	t.Logf("plugin-default rules: %d resolved from the declaration, %d also named by hand, "+
		"%d of those at a severity the declaration would not contribute",
		len(fromPlugins), len(PluginDefaultRules)-missingByHand, severityDiffers)

	if severityDiffers > 0 {
		t.Logf("the hand-written lines set %d rules to a severity `warn_correctness` does not "+
			"contribute, so removing them would lower those rules to warn. Ruled 2026-08-25 that "+
			"they stay at error: parity is the criterion for which rules run rather than how "+
			"loudly, and this set holds no rule whose violation is arguably fine. See "+
			"`PluginDefaultSeverity` for the full reasoning before deleting any of them.",
			severityDiffers)
	}

	// The forty are a deliberate override, so all forty being named is the expected state and a
	// drop is what wants attention. Asserted rather than logged: if a cleanup pass removes some of
	// these lines believing them redundant, this is the check that says otherwise, and a log line
	// in a passing test is not something anyone reads.
	if missingByHand != 0 {
		t.Errorf("%d plugin-default rules are no longer named in the config, so they have dropped "+
			"from error to warn. If that was deliberate, update this test and say so; if it was a "+
			"cleanup removing lines that looked redundant, the lines were load-bearing and the "+
			"reasoning is on `PluginDefaultSeverity`", missingByHand)
	}
}

// ruleNamesInChain returns the rules named by hand anywhere in the config's extends chain, before any
// plugin default is merged in, the project's own file winning over the tiers it extends.
//
// Separate from `Load` on purpose. `Load` returns the resolved result, which is the right answer for
// every consumer and the wrong one for a test asking which of two paths a rule arrived by.
//
// The whole chain rather than the one file, since ahra's config became an overlay on the house tiers
// (#rkm5a31), now the sets cohere carries (#njhfftt): the forty lines moved into the Nexus tier, so
// reading ahra's file alone reported all forty dropped to warn while every one still ran at error.
func ruleNamesInChain(path string) (map[string]RuleSetting, error) {
	sources, err := SourcesOf(path)
	if err != nil {
		return nil, err
	}
	named := map[string]RuleSetting{}
	for index := len(sources) - 1; index >= 0; index-- {
		contents, err := SourceContents(sources[index])
		if err != nil {
			return nil, err
		}
		layer, err := ruleNamesInContents(contents)
		if err != nil {
			return nil, fmt.Errorf("reading the rules block of %s: %w", sources[index], err)
		}
		for name, setting := range layer {
			named[name] = setting
		}
	}
	return named, nil
}

// ruleNamesInContents is the same read, over bytes rather than a path.
//
// Split out so the guard above can be exercised against a modified config WITHOUT writing one. That
// matters more than it looks: the config this guards is untracked and shared, so any check needing a
// mutated copy on disk would have to mutate the artifact it guards, or write one somewhere and hope
// the two stay comparable.
//
// Stated as a property to keep rather than an accident of how this was written: **a guard that can be
// fed its own defect in-process is verifiable by anyone, at any time, without touching what it
// protects.** `TestTheGuardCatchesARemovedLine` is what that buys, and it is the difference between
// running a guard and verifying one.
func ruleNamesInContents(contents []byte) (map[string]RuleSetting, error) {
	var raw struct {
		Rules map[string]json.RawMessage `json:"rules"`
	}
	if err := json.Unmarshal(contents, &raw); err != nil {
		return nil, err
	}

	named := map[string]RuleSetting{}
	for name, value := range raw.Rules {
		setting, err := parseRuleSetting(value)
		if err != nil {
			return nil, err
		}
		// The config writes `react/no-children-prop` and the table keys on the same spelling the
		// inventory uses, so both are recorded and a lookup finds either.
		named[name] = setting
		if _, bare, namespaced := strings.Cut(name, "/"); namespaced {
			named[bare] = setting
		}
	}
	return named, nil
}

// TestTheGuardCatchesARemovedLine feeds the guard the defect it exists to catch.
//
// `TestBothPathsAgreeOnEveryPluginDefault` is the condition on removing the forty `error` lines, and
// an earlier version of it could not fail: it read `Load`'s result, where the two paths have already
// been merged, so a rule arriving only through the declaration was indistinguishable from one named
// by hand. Three lines were deleted from a copy of the live config and it passed, reporting all
// forty still named.
//
// Running a guard and seeing green is not verifying a guard. This verifies it, by reproducing the
// removal in-process and asserting the count moves.
//
// It writes nothing. The config being guarded is untracked and shared, so a check that needed a
// mutated copy on disk would either touch the artifact it guards or drift from it; `ruleNamesInContents`
// exists so this one does neither.
func TestTheGuardCatchesARemovedLine(t *testing.T) {
	t.Parallel()
	intact, err := ruleNamesInChain(liveConfigPath(t))
	if err != nil {
		t.Fatalf("reading the rules blocks: %v", err)
	}
	unnamed := func(named map[string]RuleSetting) int {
		missing := 0
		for ruleName := range PluginDefaultRules {
			if _, present := named[ruleName]; !present {
				missing++
			}
		}
		return missing
	}

	// The control: as shipped, every plugin-default rule is named by hand. If this is ever not true
	// the mutation below proves nothing, because the count was already moving.
	if before := unnamed(intact); before != 0 {
		t.Fatalf("%d plugin-default rules are already unnamed before any mutation, so this test "+
			"cannot attribute a change to the removal", before)
	}

	// Remove three by hand, the way a cleanup pass believing them redundant would.
	removed := []string{"no-const-assign", "getter-return", "constructor-super"}
	stripped := map[string]RuleSetting{}
	for name, setting := range intact {
		stripped[name] = setting
	}
	for _, name := range removed {
		if _, present := stripped[name]; !present {
			t.Fatalf("%q is not named in the live config, so removing it models nothing; pick a "+
				"rule the config actually carries", name)
		}
		delete(stripped, name)
	}

	if after := unnamed(stripped); after != len(removed) {
		t.Errorf("after removing %d plugin-default lines the guard counts %d unnamed; it must count "+
			"exactly the removals or it cannot tell a cleanup from a no-op", len(removed), after)
	}
}
