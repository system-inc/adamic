package configuration

import "testing"

/*
 * A bare registry name reachable from two prefixed entries resolves as unconfigured, not as a guess.
 *
 * The reconciliation is a convenience and it cannot be one when the config is genuinely ambiguous.
 * Returning unconfigured routes this to the caller that already reports "nobody has said whether
 * this rule should run", which is true, rather than picking whichever key the map yielded first.
 */
func TestSettingForDeclinesAnAmbiguousBareName(t *testing.T) {
	t.Parallel()

	config := &Config{
		Rules: map[string]RuleSetting{
			"@typescript-eslint/no-shadow": {Severity: SeverityError},
			"nexus/no-shadow":              {Severity: SeverityOff},
		},
		Root: "/",
	}

	for attempt := 0; attempt < 200; attempt++ {
		if config.Resolve("/file.ts").Enabled("no-shadow") {
			t.Fatalf("attempt %d: an ambiguous bare name resolved as enabled; two prefixed entries "+
				"reach it and the config has not said which one is meant", attempt)
		}
	}
}

/*
 * The reconciliation this whole mechanism exists for still works.
 *
 * Kept beside the two guards above because they constrain it, and a fix that satisfied them by
 * removing the fallback entirely would make every one of the 170 bare registry names unconfigured,
 * therefore disabled, which is the failure the doc comment on `settingFor` records as having
 * printed 0 findings over 3,408 files with exit 0.
 */
func TestSettingForStillReconcilesABarePluginRule(t *testing.T) {
	t.Parallel()

	config := &Config{
		Rules: map[string]RuleSetting{"nexus/consistency-no-enum": {Severity: SeverityError}},
		Root:  "/",
	}
	if !config.Resolve("/file.ts").Enabled("consistency-no-enum") {
		t.Error("a bare registry name no longer finds its prefixed config entry")
	}
}

/*
 * Every option element survives parsing, in order, rather than the first alone.
 *
 * eslint's wire format is `[severity, ...options]`, and several core rules use more than one
 * element: `eqeqeq` is `['error', 'always', { null: 'ignore' }]`. `parseRuleSetting` kept `tuple[1]`
 * and discarded the rest, so such an entry was accepted and silently did nothing beyond its first
 * option. Measured before the first repair: a config carrying `["error", "always", {"null":
 * "ignore"}]` resolved its options to the string `"always"` alone.
 *
 * The first repair kept the remainder in a separate `AdditionalOptions` field that nothing outside
 * this test read, so the option was still dropped one layer further in, which is why the field is
 * gone and the whole list is `Options`. Whether a rule may take that many is
 * `OptionsRegistry.Decode`'s decision, and decode_test.go pins it.
 *
 * The single-option and bare-severity rows are the load-bearing half: a single element must still
 * arrive as itself, and a bare severity as nil rather than an empty list a decoder might read as
 * "configured with nothing".
 */
func TestParseRuleSettingKeepsEveryOptionElement(t *testing.T) {
	t.Parallel()

	for name, testCase := range map[string]struct {
		body        string
		wantOptions []string
	}{
		"twoOptionElements": {
			body:        `["error", "always", {"null": "ignore"}]`,
			wantOptions: []string{`"always"`, `{"null": "ignore"}`},
		},
		"threeOptionElements": {
			body:        `["error", "self", "vm", "that"]`,
			wantOptions: []string{`"self"`, `"vm"`, `"that"`},
		},
		"oneOptionElement": {
			body:        `["error", {"allow": ["warn"]}]`,
			wantOptions: []string{`{"allow": ["warn"]}`},
		},
		"bareSeverity": {
			body: `"error"`,
		},
		"severityAloneInAnArray": {
			body: `["error"]`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			setting, err := parseRuleSetting([]byte(testCase.body))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if testCase.wantOptions == nil && setting.Options != nil {
				t.Fatalf("Options = %s, want nil for a rule with no option elements", setting.Options)
			}
			if len(setting.Options) != len(testCase.wantOptions) {
				t.Fatalf("Options has %d elements, want %d", len(setting.Options), len(testCase.wantOptions))
			}
			for index, want := range testCase.wantOptions {
				if got := string(setting.Options[index]); got != want {
					t.Errorf("Options[%d] = %s, want %s", index, got, want)
				}
			}
		})
	}
}
