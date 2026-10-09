package configuration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheThreeHundredThirtySixCase is the finding that produced this package.
//
// `consistency-require-type-suffix` fired 336 times across
// `libraries/structure/source/api/graphql/generated/`, which the gate cohere replaces correctly
// stays silent on. If this test ever passes in the wrong direction, those 336 come back.
func TestTheThreeHundredThirtySixCase(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Rules: map[string]RuleSetting{
			"consistency-require-type-suffix": {Severity: SeverityError},
		},
		Overrides: []Override{{
			Files: []string{"**/generated/**/*.{ts,tsx}"},
			Rules: map[string]RuleSetting{
				"consistency-require-type-suffix": {Severity: SeverityOff},
			},
		}},
	}

	generated := configuration.Resolve("libraries/structure/source/api/graphql/generated/GraphQlOperations.ts")
	if generated.Enabled("consistency-require-type-suffix") {
		t.Fatal("the rule is still enabled inside generated/, which is the 336 findings")
	}

	// The other half, and the one a too-broad override would break silently: the rule must still run
	// everywhere else. An override that disabled it globally would pass the assertion above.
	authored := configuration.Resolve("libraries/structure/source/components/buttons/Button.tsx")
	if !authored.Enabled("consistency-require-type-suffix") {
		t.Fatal("the override leaked outside generated/, disabling the rule on authored code")
	}
}

// TestDoubleStarCrossesDirectoriesAndStarDoesNot is the distinction that decides whether an
// approximation loses findings.
//
// Treating `**` as `*` makes the generated override miss every nested file. Treating `*` as
// crossing slashes makes `modules/*` swallow a whole subtree and scope rules off files nobody
// excluded. The first is loud, the second is invisible.
func TestDoubleStarCrossesDirectoriesAndStarDoesNot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**/generated/**/*.ts", "a/b/generated/c/d/File.ts", true},
		{"**/generated/**/*.ts", "generated/File.ts", true},
		{"**/generated/**/*.ts", "a/generated/File.ts", true},
		{"**/generated/**/*.ts", "a/b/File.ts", false},
		{"**/generated/**/*.ts", "a/generated/c/File.tsx", false},

		{"modules/*", "modules/finance", true},
		{"modules/*", "modules/finance/Deep.ts", false},
		{"modules/**", "modules/finance/Deep.ts", true},
		{"modules/**", "modules", true},

		{"*.ts", "File.ts", true},
		{"*.ts", "nested/File.ts", false},
	}

	for _, testCase := range cases {
		if got := Match(testCase.pattern, testCase.path); got != testCase.want {
			t.Errorf("Match(%q, %q) = %v, want %v", testCase.pattern, testCase.path, got, testCase.want)
		}
	}
}

// TestBraceExpansion covers `*.{ts,tsx}`, which is in the live configuration. A matcher that ignored braces
// would match neither extension while looking like it worked.
func TestBraceExpansion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**/generated/**/*.{ts,tsx}", "a/generated/b/File.ts", true},
		{"**/generated/**/*.{ts,tsx}", "a/generated/b/File.tsx", true},
		{"**/generated/**/*.{ts,tsx}", "a/generated/b/File.js", false},
		{"*.{ts,tsx}", "File.tsx", true},
		{"{a,b}/*.ts", "b/File.ts", true},
		{"{a,b}/*.ts", "c/File.ts", false},
		// Nested alternation, and an unbalanced brace that must not crash.
		{"{a,{b,c}}/File.ts", "c/File.ts", true},
		{"{unclosed/File.ts", "{unclosed/File.ts", true},
	}

	for _, testCase := range cases {
		if got := Match(testCase.pattern, testCase.path); got != testCase.want {
			t.Errorf("Match(%q, %q) = %v, want %v", testCase.pattern, testCase.path, got, testCase.want)
		}
	}
}

// TestEveryLivePatternBehaves runs the actual patterns from .oxlintrc.json.
//
// Hand-written patterns in a test can drift from the config they claim to model. These are the real
// strings.
func TestEveryLivePatternBehaves(t *testing.T) {
	t.Parallel()
	ignore := []string{
		"code-quality/fixtures/**", "node_modules/**", "public/**", "**/.next/**",
		"**/.open-next/**", "**/.worker-next/**", "**/.wrangler/**", "**/build/**",
		"**/dist/**", "**/*.code.js", ".vscode/**", ".claude/worktrees/**",
		"data/**", "projects/**",
	}

	excluded := []string{
		"node_modules/react/index.d.ts",
		"code-quality/fixtures/Violation.ts",
		"libraries/structure/.next/types/route.ts",
		"apps/web/dist/bundle.ts",
		"scripts/Worker.code.js",
		"data/conversations/log.ts",
		".claude/worktrees/scratch/File.ts",
	}
	for _, path := range excluded {
		if !MatchAny(ignore, path) {
			t.Errorf("%q should be ignored by the live patterns but is not", path)
		}
	}

	// The half that matters more: real source must not be swept up by an ignore pattern.
	kept := []string{
		"libraries/structure/source/components/buttons/Button.tsx",
		"modules/finance/connections/QuickBooksAdapter.ts",
		"app/(os-layout)/os/wisdom/page.tsx",
		"libraries/structure/source/utilities/Data.ts",
	}
	for _, path := range kept {
		if MatchAny(ignore, path) {
			t.Errorf("%q is real source and must not be ignored", path)
		}
	}
}

// TestLaterOverridesWin pins the precedence both ESLint and oxlint use. Getting this backwards makes
// cohere disagree with the gate about which rules were supposed to run, which is the thing that
// blocks an honest acceptance diff.
func TestLaterOverridesWin(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Rules: map[string]RuleSetting{"a-rule": {Severity: SeverityError}},
		Overrides: []Override{
			{Files: []string{"modules/**"}, Rules: map[string]RuleSetting{"a-rule": {Severity: SeverityOff}}},
			{Files: []string{"modules/finance/**"}, Rules: map[string]RuleSetting{"a-rule": {Severity: SeverityError}}},
		},
	}

	if configuration.Resolve("modules/other/File.ts").Enabled("a-rule") {
		t.Fatal("the first override should have disabled the rule under modules/")
	}
	if !configuration.Resolve("modules/finance/File.ts").Enabled("a-rule") {
		t.Fatal("the later override should have re-enabled the rule under modules/finance/")
	}
}

// An override that re-states a rule with a bare severity keeps the options already in force for the
// file, as ESLint does and as a bare severity across `extends` already did. Found on api: its TS-file
// override `"prefer-const": "error"` dropped the Nexus tier's ignoreReadBeforeAssign in cohere only, so
// the two engines judged prefer-const differently on every TS file (#na0hgjz). An override that writes
// its own options still replaces them.
func TestABareSeverityInAnOverrideKeepsTheOptionsInForce(t *testing.T) {
	t.Parallel()
	inherited := []json.RawMessage{json.RawMessage(`{"ignoreReadBeforeAssign":true}`)}
	written := []json.RawMessage{json.RawMessage(`{"destructuring":"all"}`)}
	configuration := &Config{
		Rules: map[string]RuleSetting{"prefer-const": {Severity: SeverityError, Options: inherited}},
		Overrides: []Override{
			{Files: []string{"**/*.ts"}, Rules: map[string]RuleSetting{"prefer-const": {Severity: SeverityError}}},
			{Files: []string{"modules/written/**"}, Rules: map[string]RuleSetting{"prefer-const": {Severity: SeverityError, Options: written}}},
		},
	}

	if options := configuration.Resolve("modules/File.ts").RawOptionsFor("prefer-const"); len(options) != 1 || string(options[0]) != string(inherited[0]) {
		t.Errorf("a bare severity in an override dropped the options in force: got %s", options)
	}
	if options := configuration.Resolve("modules/written/File.ts").RawOptionsFor("prefer-const"); len(options) != 1 || string(options[0]) != string(written[0]) {
		t.Errorf("an override's own options did not replace the ones in force: got %s", options)
	}
	if options := configuration.Resolve("modules/File.js").RawOptionsFor("prefer-const"); len(options) != 1 || string(options[0]) != string(inherited[0]) {
		t.Errorf("a file no override matches lost its options: got %s", options)
	}
}

// TestOneFilesOverridesDoNotLeakIntoTheNext guards a real aliasing bug: sharing the base rule map
// across files makes the first override permanent for every file resolved afterward.
func TestOneFilesOverridesDoNotLeakIntoTheNext(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Rules: map[string]RuleSetting{"a-rule": {Severity: SeverityError}},
		Overrides: []Override{{
			Files: []string{"**/generated/**"},
			Rules: map[string]RuleSetting{"a-rule": {Severity: SeverityOff}},
		}},
	}

	configuration.Resolve("src/generated/File.ts")
	if !configuration.Resolve("src/authored/File.ts").Enabled("a-rule") {
		t.Fatal("resolving a generated file disabled the rule for a later authored file")
	}
}

// TestIgnoredFileRunsNoRulesAndSaysWhy proves exclusion is reportable rather than merely silent. A
// file skipped by ignorePatterns and a file with no findings are identical output otherwise.
func TestIgnoredFileRunsNoRulesAndSaysWhy(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Rules:          map[string]RuleSetting{"a-rule": {Severity: SeverityError}},
		IgnorePatterns: []string{"node_modules/**"},
	}

	resolved := configuration.Resolve("node_modules/thing/index.ts")
	if !resolved.Ignored {
		t.Fatal("an ignored path did not resolve as ignored")
	}
	if resolved.IgnoredBy != "node_modules/**" {
		t.Fatalf("the excluding pattern was not reported: %q", resolved.IgnoredBy)
	}
	if resolved.Enabled("a-rule") {
		t.Fatal("a rule ran on an ignored file")
	}
}

// TestAnUnconfiguredRuleDoesNotRun keeps adding a rule to the registry from silently enabling it
// across the whole tree.
func TestAnUnconfiguredRuleDoesNotRun(t *testing.T) {
	t.Parallel()
	configuration := &Config{Rules: map[string]RuleSetting{"known": {Severity: SeverityError}}}

	if configuration.Resolve("File.ts").Enabled("never-configured") {
		t.Fatal("a rule the config never mentions was treated as enabled")
	}
}

// TestBothRuleShapesLoad covers the 173 bare severities and the 9 that carry options.
func TestBothRuleShapesLoad(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, ".oxlintrc.json")
	contents := `{
		"rules": {
			"plain-rule": "error",
			"disabled-rule": "off",
			"configured-rule": ["error", {"ignoreRestArgs": true}]
		},
		"reasons": {"disabled-rule": "a fixture's off"},
		"ignorePatterns": ["dist/**"],
		"overrides": [{"files": ["**/generated/**/*.{ts,tsx}"], "rules": {"plain-rule": "off"}}]
	}`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if loaded.Rules["plain-rule"].Severity != SeverityError {
		t.Fatal("a bare severity did not load as error")
	}
	if loaded.Rules["disabled-rule"].Severity != SeverityOff {
		t.Fatal("an off rule did not load as off")
	}
	if loaded.Rules["configured-rule"].Severity != SeverityError {
		t.Fatal("a [severity, options] rule did not load its severity")
	}
	if len(loaded.Rules["configured-rule"].Options) == 0 {
		t.Fatal("a rule's options were dropped, which is how a rule ends up guarding nothing")
	}
	if len(loaded.Rules["plain-rule"].Options) != 0 {
		t.Fatal("a bare rule invented options it was never given")
	}
	if len(loaded.Overrides) != 1 || len(loaded.IgnorePatterns) != 1 {
		t.Fatalf("overrides or ignorePatterns did not load: %+v", loaded)
	}
}

// TestAnUnreadableConfigIsAnErrorNotAnEmptyConfig is the loudness guard. An empty config lints
// everything with nothing configured, which looks exactly like a clean run.
func TestAnUnreadableConfigIsAnErrorNotAnEmptyConfig(t *testing.T) {
	t.Parallel()
	if _, err := Load(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Fatal("a missing config file loaded successfully")
	}

	directory := t.TempDir()
	path := filepath.Join(directory, "broken.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("writing: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("a malformed config loaded successfully")
	}
}

// TestAnUnknownSeverityIsRefused keeps a typo from silently disabling a rule tree-wide.
func TestAnUnknownSeverityIsRefused(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "config.json")
	if err := os.WriteFile(path, []byte(`{"rules":{"a-rule":"errrror"}}`), 0o644); err != nil {
		t.Fatalf("writing: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("an unknown severity was accepted, which would disable the rule silently")
	}
}

// TestUnimplementedTopLevelKeyIsRefused proves the front-door guard fires.
//
// `encoding/json` drops an unlisted key with no error, so before this guard a config could declare
// something and have it do nothing while every check passed. Three keys were in that state at once:
// `plugins`, `jsPlugins` and `settings`. None was found by a mechanism -- `plugins` surfaced while
// someone investigated why 40 rules ran on zero files, and the other two only because that
// investigation prompted a read of the whole file.
//
// The fixture uses a key nobody would add by accident, so a future config gaining a real key does
// not make this test wrong. What it asserts is the mechanism, not a particular key.
func TestUnimplementedTopLevelKeyIsRefused(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "CohereSettings.json")
	contents := `{"rules": {"a-rule": "error"}, "notAKeyThisLoaderKnows": {"anything": 1}}`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("a config declaring an unimplemented top-level key loaded successfully, so the key " +
			"was discarded silently and whatever it configures would never take effect")
	}
	if !strings.Contains(err.Error(), "notAKeyThisLoaderKnows") {
		t.Errorf("the error does not name the offending key, so a reader cannot act on it: %v", err)
	}

	// The control. The same config without the key must load, or the test above passes for the
	// wrong reason and this guard would reject every config equally.
	controlPath := filepath.Join(directory, "control.json")
	if err := os.WriteFile(controlPath, []byte(`{"rules": {"a-rule": "error"}}`), 0o644); err != nil {
		t.Fatalf("writing the control config: %v", err)
	}
	if _, err := Load(controlPath); err != nil {
		t.Fatalf("the control config failed to load, so the guard rejects more than it should: %v", err)
	}
}

// TestIgnoredKeyWithoutAReasonIsRefused is the guard on the guard.
//
// The escape hatch for an unimplemented key is to record it as deliberately ignored. That hatch is
// only worth anything if using it costs a sentence of justification -- otherwise the fix for this
// error is to add a name to a map, which is the same silence with one more step.
//
// The map is package state rather than a parameter, so this exercises the check by adding an entry
// with an empty reason and removing it again. That is a mutation of shared state in a test, and it
// is done here rather than by restructuring the maps because the alternative is threading a
// parameter through `Load` for the sole benefit of this assertion.
// Not parallel: it adds an entry to the package's ignoredTopLevelKeys map, which Load reads for every
// top-level key it does not parse.
func TestIgnoredKeyWithoutAReasonIsRefused(t *testing.T) {
	const key = "keyRecordedWithNoReason"
	ignoredTopLevelKeys[key] = ""
	t.Cleanup(func() { delete(ignoredTopLevelKeys, key) })

	directory := t.TempDir()
	path := filepath.Join(directory, "CohereSettings.json")
	contents := `{"rules": {"a-rule": "error"}, "keyRecordedWithNoReason": true}`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("a key listed as ignored with no reason loaded successfully; an entry with no " +
			"reason says only that somebody wanted the config to load")
	}
	if !strings.Contains(err.Error(), "no reason") {
		t.Errorf("the error does not say the reason is what is missing: %v", err)
	}
}

// TestEveryIgnoredKeyCarriesAReason holds the map itself to the standard the guard enforces.
//
// The test above proves an empty reason is refused at load time, which only helps if someone
// actually loads a config carrying that key. This asserts the invariant directly, so an entry added
// with no reason fails immediately rather than whenever a config happens to use it.
func TestEveryIgnoredKeyCarriesAReason(t *testing.T) {
	t.Parallel()
	if len(ignoredTopLevelKeys) == 0 {
		t.Fatal("no ignored keys are recorded, so this test asserts nothing")
	}
	for key, reason := range ignoredTopLevelKeys {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("ignored key %q carries no reason", key)
		}
	}
}

// TestPluginsIsParsedRatherThanIgnored records that one of the three dropped keys is now implemented.
//
// `plugins` was discarded, then recorded as deliberately ignored, and is now parsed and acted on.
// Each of those is a different state and only the last is correct, so this asserts the current one
// rather than leaving the transition implicit. It also documents why `plugins` is absent from the
// known-dirty control below: it left that population by being fixed.
func TestPluginsIsParsedRatherThanIgnored(t *testing.T) {
	t.Parallel()
	if !parsedTopLevelKeys["plugins"] {
		t.Error("`plugins` is not parsed; it was implemented under #0ympke3 and a config declaring " +
			"it would be refused rather than honoured")
	}
	if _, ignored := ignoredTopLevelKeys["plugins"]; ignored {
		t.Error("`plugins` is recorded as deliberately ignored AND parsed; one of the two is stale " +
			"and a reader cannot tell which describes the behaviour")
	}
}

// TestTheThreeRealKeysWouldHaveBeenCaught is the known-dirty control.
//
// The refusal test above uses a synthetic key, which proves the mechanism and not that it would have
// caught the defect that motivated it. This runs the guard against the three keys that were actually
// being discarded from the live config -- `plugins`, `jsPlugins` and `settings` -- with each one
// temporarily removed from the ignored set, so a reader can see the guard produce the finding rather
// than trust that it would have.
//
// Without a control like this, the guard and a guard that fires only on names nobody uses look
// identical from a green suite.
// Not parallel: its subtests delete entries from the package's ignoredTopLevelKeys map, which Load reads
// for every top-level key it does not parse.
func TestTheThreeRealKeysWouldHaveBeenCaught(t *testing.T) {
	// `plugins` and `settings` were two of these and are no longer: each is implemented now, so it is
	// parsed rather than ignored and the decay guard below correctly refused to keep testing it
	// (`settings` since #gj5nm6e, read for better-tailwindcss). Removed here rather than by weakening
	// the guard, which is the whole point of the guard.
	for _, key := range []string{"jsPlugins"} {
		// Not parallel: each one deletes and restores an entry in the package's ignoredTopLevelKeys map, so
		// two at once would write the same map.
		t.Run(key, func(t *testing.T) {
			reason, recorded := ignoredTopLevelKeys[key]
			if !recorded {
				t.Fatalf("%q is no longer recorded as ignored, so this control is testing a key "+
					"that is not the one that was being dropped; either it was implemented, in "+
					"which case remove it from this list, or the entry was deleted", key)
			}
			delete(ignoredTopLevelKeys, key)
			t.Cleanup(func() { ignoredTopLevelKeys[key] = reason })

			directory := t.TempDir()
			path := filepath.Join(directory, "CohereSettings.json")
			contents := `{"rules": {"a-rule": "error"}, "` + key + `": ["something"]}`
			if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
				t.Fatalf("writing the config: %v", err)
			}

			_, err := Load(path)
			if err == nil {
				t.Fatalf("a config declaring %q loaded successfully with the key unaccounted for; "+
					"this is the exact state the live config was in and nothing reported it", key)
			}
			if !strings.Contains(err.Error(), key) {
				t.Errorf("the error does not name %q: %v", key, err)
			}
		})
	}
}

// TestPluginDeclarationEnablesItsRules is the mechanism, on a config naming none of them.
func TestPluginDeclarationEnablesItsRules(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "CohereSettings.json")
	contents := `{"plugins": ["react"], "rules": {"some-named-rule": "error"}}`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	// A react rule nobody named must now be enabled, at warn.
	setting, enabled := loaded.Rules["react/no-children-prop"]
	if !enabled {
		t.Fatal("declaring the react plugin did not enable react/no-children-prop, which the real " +
			"oxlint binary reports on a tree whose config never names it")
	}
	if setting.Severity != SeverityWarn {
		t.Errorf("react/no-children-prop resolved at severity %v, want warn: `warn_correctness` "+
			"inserts Warn and reporting it louder than the gate does is a divergence", setting.Severity)
	}

	// A typescript rule must NOT be, since that plugin is not declared. Without this the test
	// passes for a resolver that enables everything.
	if _, enabled := loaded.Rules["typescript/no-extra-non-null-assertion"]; enabled {
		t.Error("a typescript rule was enabled by a config declaring only react, so the plugin " +
			"filter is not discriminating")
	}

	// Core rules arrive whatever is declared, because `ESLINT` is an empty bitflag upstream.
	if _, enabled := loaded.Rules["no-const-assign"]; !enabled {
		t.Error("no-const-assign was not enabled; core correctness rules are contributed " +
			"unconditionally and upstream states there is no way to disable them")
	}
}

// TestAnExplicitLineBeatsAPluginDefault pins the precedence.
//
// An explicit entry is a decision and a default is not. Getting this backwards would let a plugin
// declaration silently re-enable a rule someone deliberately turned off, which is the loudest way
// this feature could go wrong.
func TestAnExplicitLineBeatsAPluginDefault(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "CohereSettings.json")
	contents := `{"plugins": ["react"], "rules": {"react/no-children-prop": "off"}, "reasons": {"react/no-children-prop": "turned off on purpose"}}`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("loading: %v", err)
	}
	if loaded.Rules["react/no-children-prop"].Severity != SeverityOff {
		t.Error("a plugin default overrode an explicit `off`, so declaring a plugin would silently " +
			"re-enable rules somebody decided to disable")
	}
}

// TestTheFormatBlockLoadsAndAnUnknownKeyBesideItStillDoesNot: the formatter's options live in the
// same file as the rules, under "format", and the linter must load around them. Accepting that one key
// must not loosen the guard, so an unknown key next to it is still refused and still named.
func TestTheFormatBlockLoadsAndAnUnknownKeyBesideItStillDoesNot(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "CohereSettings.json")

	accepted := `{"rules": {"a-rule": "error"}, "format": {"printWidth": 120, "tabWidth": 4}}`
	if err := os.WriteFile(path, []byte(accepted), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("a config carrying the formatter's options did not load: %v", err)
	}

	refused := `{"rules": {"a-rule": "error"}, "format": {"printWidth": 120}, "formatter": {}}`
	if err := os.WriteFile(path, []byte(refused), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("an unknown top-level key beside \"format\" loaded, so accepting format loosened the guard")
	}
	if !strings.Contains(err.Error(), "formatter") {
		t.Errorf("the refusal does not name the unknown key: %v", err)
	}
}
