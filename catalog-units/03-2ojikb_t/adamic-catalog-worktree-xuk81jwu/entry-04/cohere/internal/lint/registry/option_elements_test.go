package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * Every option element a config writes reaches the rule, through the real config layer.
 *
 * This file exists because no rule fixture can see the defect it guards. A fixture hands a decoder
 * bytes the test built; the config layer builds different bytes. `parseRuleSetting` used to keep
 * `tuple[1]` and drop the rest, so `["error", {"object": true}, {"enforceForRenamedProperties":
 * true}]` loaded clean and reported 0 where the option exists to report 1, while every
 * prefer-destructuring fixture stayed green. PortingARule.md section 7c is that blindness.
 *
 * So every row here starts from config TEXT: it is written to a CohereSettings.json, loaded by
 * `configuration.Load`, resolved, handed through `Options().Decode` exactly as the lint walk does,
 * and the decoded value runs the rule. Each row is a pair. The control writes the rule without the
 * element under test, the treatment writes upstream's spelling with it, and the two must disagree.
 * Against the pre-repair config layer every treatment collapses onto its control, so every row
 * fails; that was checked by restoring `tuple[1]` and watching them fail, not assumed.
 */

// optionElementRules are the registered rules whose upstream `meta.schema` takes more than one
// option element, measured by loading each installed rule from ahra's node_modules and reading its
// schema's length or `anyOf` arms rather than by reading source. Each must register with
// DecodeOptionList; with Decode the config layer would refuse its second element.
var optionElementRules = []string{
	"arrow-body-style",
	"consistent-this",
	"eqeqeq",
	"func-name-matching",
	"grouped-accessor-pairs",
	"id-denylist",
	"id-match",
	"logical-assignment-operators",
	"no-inner-declarations",
	"no-restricted-globals",
	"no-restricted-imports",
	"no-restricted-properties",
	"object-shorthand",
	"prefer-destructuring",
	"yoda",
	"@typescript-eslint/init-declarations",
	"react/jsx-no-script-url",
	"react/static-property-placement",
}

// TestEveryMultiElementRuleReadsItsWholeOptionList pins the arity each of those rules declares.
func TestEveryMultiElementRuleReadsItsWholeOptionList(t *testing.T) {
	t.Parallel()
	registrations := map[string]rule.Registration{}
	for _, registration := range rule.Registered() {
		registrations[registration.Rule.Name] = registration
	}
	for _, name := range optionElementRules {
		registration, registered := registrations[name]
		if !registered {
			t.Errorf("%s is listed as taking several option elements and is not registered under that name", name)
			continue
		}
		if registration.DecodeOptionList == nil {
			t.Errorf("%s takes several option elements upstream and does not register DecodeOptionList, "+
				"so the config layer refuses every element after its first", name)
		}
	}
}

// optionElementCase is one control-and-treatment pair, both written as config text.
type optionElementCase struct {
	rule     string
	fileName string
	source   string

	// control is the rule's config value without the element under test, and treatment is upstream's
	// spelling with it. The two finding counts must differ, or the element changed nothing.
	control        string
	controlCount   int
	treatment      string
	treatmentCount int
}

var optionElementCases = []optionElementCase{
	{
		rule: "grouped-accessor-pairs", source: "export interface I { get a(): any, x: 1, set a(v: any): void }\n",
		control:   `["error", "anyOrder"]`,
		treatment: `["error", "anyOrder", {"enforceForTSTypes": true}]`, treatmentCount: 1,
	},
	{
		rule: "prefer-destructuring", source: "var foobar = object.bar;\n",
		control:   `["error", {"object": true}]`,
		treatment: `["error", {"object": true}, {"enforceForRenamedProperties": true}]`, treatmentCount: 1,
	},
	{
		rule: "eqeqeq", source: "declare const a: unknown;\nexport const b = a == null;\n",
		control: `["error", "always"]`, controlCount: 1,
		treatment: `["error", "always", {"null": "ignore"}]`,
	},
	{
		rule: "arrow-body-style", source: "export const foo = () => ({});\n",
		control:   `["error", "as-needed"]`,
		treatment: `["error", "as-needed", {"requireReturnForObjectLiteral": true}]`, treatmentCount: 1,
	},
	{
		rule: "consistent-this", source: "export function f(this: unknown) { const vm = this; return vm; }\n",
		control: `["error", "self"]`, controlCount: 1,
		treatment: `["error", "self", "vm"]`,
	},
	{
		rule: "func-name-matching", source: "declare const module: { exports: unknown };\nmodule.exports = function foo() {};\n",
		control:   `["error", "always"]`,
		treatment: `["error", "always", {"includeCommonJSModuleExports": true}]`, treatmentCount: 1,
	},
	{
		rule: "id-denylist", source: "export const data = 1;\nexport const err = 2;\n",
		control: `["error", "data"]`, controlCount: 1,
		treatment: `["error", "data", "err"]`, treatmentCount: 2,
	},
	{
		rule: "id-match", source: "export const settings = { my_pref: 1 };\n",
		control:   `["error", "^[^_]+$"]`,
		treatment: `["error", "^[^_]+$", {"properties": true}]`, treatmentCount: 1,
	},
	{
		rule: "logical-assignment-operators", source: "declare let a: unknown;\ndeclare const b: unknown;\nif (!a) a = b;\n",
		control:   `["error", "always"]`,
		treatment: `["error", "always", {"enforceForIfStatements": true}]`, treatmentCount: 1,
	},
	{
		rule: "no-inner-declarations", source: "'use strict';\ndeclare const test: boolean;\nif (test) { function doSomething() {} }\n",
		control:   `["error", "functions"]`,
		treatment: `["error", "functions", {"blockScopedFunctions": "disallow"}]`, treatmentCount: 1,
	},
	{
		rule: "no-restricted-globals", source: "export const a = event;\nexport const b = fdescribe;\n",
		control: `["error", "event"]`, controlCount: 1,
		treatment: `["error", "event", "fdescribe"]`, treatmentCount: 2,
	},
	{
		rule: "no-restricted-imports", source: "import fs from 'fs';\nimport os from 'os';\nexport { fs, os };\n",
		control: `["error", "fs"]`, controlCount: 1,
		treatment: `["error", "fs", "os"]`, treatmentCount: 2,
	},
	{
		rule: "no-restricted-properties", source: "declare const foo: any;\ndeclare const baz: any;\nfoo.bar;\nbaz.qux;\n",
		control: `["error", {"object": "foo", "property": "bar"}]`, controlCount: 1,
		treatment: `["error", {"object": "foo", "property": "bar"}, {"object": "baz", "property": "qux"}]`, treatmentCount: 2,
	},
	{
		rule: "object-shorthand", source: "export const x = { ConstructorFunction: function () {} };\n",
		control: `["error", "always"]`, controlCount: 1,
		treatment: `["error", "always", {"ignoreConstructors": true}]`,
	},
	{
		rule: "yoda", source: "declare const x: number;\nif (0 < x && x <= 1) {}\n",
		control: `["error", "never"]`, controlCount: 1,
		treatment: `["error", "never", {"exceptRange": true}]`,
	},
	{
		rule: "@typescript-eslint/init-declarations", source: "for (var index = 0; index < 1; index++) {}\n",
		control: `["error", "never"]`, controlCount: 1,
		treatment: `["error", "never", {"ignoreForLoopInit": true}]`,
	},
	{
		rule: "react/static-property-placement", fileName: "Probe.tsx",
		source:    "import React from 'react';\nexport class MyComponent extends React.Component {\n  static displayName = 'Hello';\n  render() { return null; }\n}\n",
		control:   `["error", "static public field"]`,
		treatment: `["error", "static public field", {"displayName": "static getter"}]`, treatmentCount: 1,
	},
}

/*
 * The second element changes what the rule reports, measured through the config layer.
 *
 * `react/jsx-no-script-url` is not in this table because its second element's only key is one this
 * tree cannot honour (`includeFromSettings`), so its `true` is refused rather than enforced;
 * `TestASecondElementTheRuleCannotHonourIsRefused` covers it.
 */
func TestTheSecondOptionElementReachesTheRule(t *testing.T) {
	t.Parallel()

	rules := map[string]rule.Rule{}
	for _, subject := range All() {
		rules[subject.Name] = subject
	}
	covered := map[string]bool{}

	for _, testCase := range optionElementCases {
		covered[testCase.rule] = true
		t.Run(testCase.rule, func(t *testing.T) {
			t.Parallel()
			subject, registered := rules[testCase.rule]
			if !registered {
				t.Fatalf("%s is not registered", testCase.rule)
			}
			if testCase.controlCount == testCase.treatmentCount {
				t.Fatalf("the row's two counts are equal, so it cannot show the element did anything")
			}
			fileName := "/repository/source/Probe.ts"
			if testCase.fileName != "" {
				fileName = "/repository/source/" + testCase.fileName
			}

			for _, arm := range []struct {
				name   string
				config string
				want   int
			}{
				{"control", testCase.control, testCase.controlCount},
				{"treatment", testCase.treatment, testCase.treatmentCount},
			} {
				options, err := optionsThroughTheConfigLayer(t, testCase.rule, arm.config)
				if err != nil {
					t.Fatalf("%s %s was refused: %v", arm.name, arm.config, err)
				}
				result := rule_testing.RunTypedWithOptions(t, subject, fileName, testCase.source, options)
				if len(result.Diagnostics) != arm.want {
					t.Errorf("%s %s reported %d, want %d: %v",
						arm.name, arm.config, len(result.Diagnostics), arm.want, result.MessageIds())
				}
			}
		})
	}

	for _, name := range optionElementRules {
		if !covered[name] && name != "react/jsx-no-script-url" {
			t.Errorf("%s takes several option elements and has no control-and-treatment row here", name)
		}
	}
}

// TestASecondElementTheRuleCannotHonourIsRefused covers the rule whose second element's only key is
// unimplemented here. Its default is accepted, because that is what the rule already does,
// and `true` is refused by name, because accepting it would be an option read and never honoured.
func TestASecondElementTheRuleCannotHonourIsRefused(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		rule, accepted, refused, key string
	}{
		{
			rule:     "react/jsx-no-script-url",
			accepted: `["error", [{"name": "Link", "props": ["to"]}], {"includeFromSettings": false}]`,
			refused:  `["error", [{"name": "Link", "props": ["to"]}], {"includeFromSettings": true}]`,
			key:      "includeFromSettings",
		},
	} {
		if _, err := optionsThroughTheConfigLayer(t, testCase.rule, testCase.accepted); err != nil {
			t.Errorf("%s refused its second element at the default: %v", testCase.rule, err)
		}
		_, err := optionsThroughTheConfigLayer(t, testCase.rule, testCase.refused)
		if err == nil || !strings.Contains(err.Error(), testCase.key) {
			t.Errorf("%s: want a refusal naming %s, got %v", testCase.rule, testCase.key, err)
		}
	}
}

/*
 * An extra element on a rule that takes one, or any element on a rule that takes none, is a loud
 * configuration error naming the rule and the element.
 *
 * Through real registrations rather than a test registry, so a rule that quietly switches to
 * DecodeOptionList, or one whose decoder starts tolerating a list, fails here.
 */
func TestAnElementTheRuleDoesNotTakeIsALoudError(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		rule, config string
		mustName     []string
	}{
		// array-callback-return takes one object; a second is refused.
		{
			rule:     "array-callback-return",
			config:   `["error", {"allowImplicit": true}, {"checkForEach": true}]`,
			mustName: []string{"array-callback-return", `element 2 {"checkForEach": true}`},
		},
		// default-case-last takes none; any option is refused.
		{
			rule:     "default-case-last",
			config:   `["error", {"ignored": true}]`,
			mustName: []string{"default-case-last", `element 1 {"ignored": true}`},
		},
		// A list rule given more elements than upstream's schema allows.
		{
			rule:     "prefer-destructuring",
			config:   `["error", {"object": true}, {"enforceForRenamedProperties": true}, {"array": true}]`,
			mustName: []string{"prefer-destructuring", `element 3 {"array":true}`},
		},
	} {
		_, err := optionsThroughTheConfigLayer(t, testCase.rule, testCase.config)
		if err == nil {
			t.Errorf("%s %s loaded, and the extra element would never be read", testCase.rule, testCase.config)
			continue
		}
		for _, want := range testCase.mustName {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: the refusal does not name %q: %v", testCase.rule, want, err)
			}
		}
	}

	// The control: the same single-element rule with one element loads, so the refusals above are
	// about the count rather than about the rule refusing everything.
	if _, err := optionsThroughTheConfigLayer(t, "array-callback-return", `["error", {"allowImplicit": true}]`); err != nil {
		t.Errorf("array-callback-return refused its one element: %v", err)
	}
}

/*
 * A null anywhere in a rule's options is refused, naming the element and the path to the null.
 *
 * Go's decoder reads a null as "leave the field alone", so before #pd2chkx `{"props": null}` loaded
 * as no-self-assign's default and `{"allowConstructorFlags": null}` as no-invalid-regexp's, while
 * ESLint's validator refuses both: no ported rule's upstream schema admits null. The rows span the
 * three decoder shapes: one built on rule.UnmarshalOptions, one list rule that reads its elements
 * with json.Unmarshal, and a null standing for a whole element.
 */
func TestANullInARulesOptionsIsRefused(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		rule, accepted, refused, mustName string
	}{
		{
			rule:     "no-self-assign",
			accepted: `["error", {"props": false}]`,
			refused:  `["error", {"props": null}]`,
			mustName: "element 1 at props is null",
		},
		{
			rule:     "no-invalid-regexp",
			accepted: `["error", {"allowConstructorFlags": ["u"]}]`,
			refused:  `["error", {"allowConstructorFlags": ["u", null]}]`,
			mustName: "element 1 at allowConstructorFlags.1 is null",
		},
		{
			rule:     "eqeqeq",
			accepted: `["error", "always", {"null": "ignore"}]`,
			refused:  `["error", "always", {"null": null}]`,
			mustName: "element 2 at null is null",
		},
		{
			rule:     "eqeqeq",
			accepted: `["error", "always"]`,
			refused:  `["error", null]`,
			mustName: "element 1 is null",
		},
	} {
		if _, err := optionsThroughTheConfigLayer(t, testCase.rule, testCase.accepted); err != nil {
			t.Errorf("%s refused %s, so the refusal below would not be about the null: %v", testCase.rule, testCase.accepted, err)
		}
		_, err := optionsThroughTheConfigLayer(t, testCase.rule, testCase.refused)
		if err == nil {
			t.Errorf("%s %s loaded, and the null would have read as the default", testCase.rule, testCase.refused)
			continue
		}
		for _, want := range []string{testCase.rule, testCase.mustName} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: the refusal does not name %q: %v", testCase.rule, want, err)
			}
		}
	}
}

// optionsThroughTheConfigLayer writes one rule's config value to a CohereSettings.json, loads and
// resolves it, and decodes the result through the registry exactly as the lint walk does.
func optionsThroughTheConfigLayer(t *testing.T, ruleName string, value string) (any, error) {
	t.Helper()
	if !json.Valid([]byte(value)) {
		t.Fatalf("the row's config value is not JSON: %s", value)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "CohereSettings.json")
	contents := `{"rules": {"` + ruleName + `": ` + value + `}}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	loaded, err := configuration.Load(path)
	if err != nil {
		t.Fatalf("loading %s: %v", contents, err)
	}
	resolved := loaded.Resolve(filepath.Join(directory, "Probe.ts"))
	return Options().Decode(ruleName, resolved.RawOptionsFor(ruleName))
}
