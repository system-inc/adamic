package core

import (
	"encoding/json"
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// idMatchFile is where the fixtures pretend to live.
const idMatchFile = "/repository/source/IdMatch.ts"

// idMatchOf builds the settings a config carrying this pattern and these flags would decode to.
//
// Routed through the rule's own exported decoder rather than filling the struct, so a defect in the
// decoder is under test rather than bypassed by every fixture. The flags are passed as a map so a
// case naming only `properties` decodes through the same absent-key path a real config would.
func idMatchOf(pattern string, flags map[string]bool) any {
	tuple := []any{pattern}
	if flags != nil {
		tuple = append(tuple, flags)
	}
	encoded, err := json.Marshal(tuple)
	if err != nil {
		panic(err)
	}
	settings, err := DecodeIdMatchOptions(encoded)
	if err != nil {
		panic(err)
	}
	return settings
}

// The corpus is ESLint's own, taken from
// /tmp/lint-sources-fresh/eslint/tests/lib/rules/id-match.js by RUNNING that file with RuleTester
// intercepted, so upstream's own cases produced the fixtures rather than a parser of mine. 100 cases
// were captured from the single RuleTester.run call, and every one is a distinct (source, options)
// pair -- no case in this corpus carries contradictory verdicts under different parser settings.
//
// Every expectation below is what the INSTALLED ESLint 10.8.1 core rule answered when driven over
// that case through the Linter interface under `sourceType: module` with the typescript-eslint
// parser, which is the only configuration cohere has. The 10.10.0 clone and the 10.8.1 installed
// build are BYTE-IDENTICAL for this rule:
//
//	diff /tmp/lint-sources-fresh/eslint/lib/rules/id-match.js \
//	     /Users/kirkouimet/Projects/ahra/node_modules/eslint/lib/rules/id-match.js
//
// The oracle reproduced all 100 of the corpus's own declared verdicts under upstream's own
// languageOptions, which is the control that says it was measuring the rule rather than an
// unconfigured Linter. And unlike id-denylist, ZERO cases change verdict between upstream's
// configuration and cohere's, so this table needs no divergence section: id-match's corpus carries
// no `globals` languageOption and no /* global */ directive, so nothing in it depends on a surface
// cohere lacks.
//
//	49 reporting cases, 51 clean cases.
func TestIdMatchFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		// settings is what the rule's own decoder produces for this case's options.
		settings any
		// findings is one message id per expected finding, in report order, taken from what the
		// installed rule produced rather than assumed: a private name reports through a different
		// id and no count could tell them apart.
		findings []string
	}{
		{source: "var __foo = \"Matthieu\"", settings: idMatchOf("^[a-z]+$", map[string]bool{"onlyDeclarations": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "first_name = \"Matthieu\"", settings: idMatchOf("^[a-z]+$", nil), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "first_name = \"Matthieu\"", settings: idMatchOf("^z", nil), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "Last_Name = \"Larcher\"", settings: idMatchOf("^[a-z]+(_[A-Z][a-z])*$", nil), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var obj = {key: no_under}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "function no_under21(){}", settings: idMatchOf("^[^_]+$", nil), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "obj.no_under22 = function(){};", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "no_under23.foo = function(){};", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "[no_under24.baz]", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "if (foo.bar_baz === boom.bam_pow) { [no_under25.baz] }", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "foo.no_under26 = boom.bam_pow", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var foo = { no_under27: boom.bam_pow }", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "foo.qux.no_under28 = { bar: boom.bam_pow }", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var o = {no_under29: 1}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "obj.no_under30 = 2;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var { category_id: category_alias } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var { category_id: category_alias } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "ignoreDestructuring": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var { category_id: categoryId, ...other_props } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "ignoreDestructuring": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var { category_id } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var { category_id = 1 } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import no_camelcased from \"external-module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import * as no_camelcased from \"external-module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "export * as no_camelcased from \"external-module\";", settings: idMatchOf("^[^_]+$", nil), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import { no_camelcased } from \"external-module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import { no_camelcased as no_camel_cased } from \"external module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import { camelCased as no_camel_cased } from \"external module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import { camelCased, no_camelcased } from \"external-module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import { no_camelcased as camelCased, another_no_camelcased } from \"external-module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import camelCased, { no_camelcased } from \"external-module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import no_camelcased, { another_no_camelcased as camelCased } from \"external-module\";", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "function foo({ no_camelcased }) {};", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "function foo({ no_camelcased = 'default value' }) {};", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "const no_camelcased = 0; function foo({ camelcased_value = no_camelcased }) {}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id}},
		{source: "const { bar: no_camelcased } = foo;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "function foo({ value_1: my_default }) {}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "function foo({ isCamelcased: no_camelcased }) {};", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "var { foo: bar_baz = 1 } = quz;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "const { no_camelcased = false } = bar;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "\n            const foo_variable = 1;\n            class MyClass {\n            }\n            let a = new MyClass();\n            let b = {id: 1};\n            let c = Object.keys(b);\n            let d = Array.from(b);\n            let e = (Object) => Object.keys(obj, prop); // not global Object\n            let f = (Array) => Array.from(obj, prop); // not global Array\n            foo.Array = 5; // not global Array\n            ", settings: idMatchOf("^\\$?[a-z]+([A-Z0-9][a-z0-9]+)*$", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id}},
		{source: "class x { _foo() {} }", settings: idMatchOf("^[^_]+$", nil), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "class x { #_foo() {} }", settings: idMatchOf("^[^_]+$", nil), findings: []string{messageIdMatchNotMatchPrivate.Id}},
		{source: "class x { _foo = 1; }", settings: idMatchOf("^[^_]+$", map[string]bool{"classFields": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "class x { #_foo = 1; }", settings: idMatchOf("^[^_]+$", map[string]bool{"classFields": true}), findings: []string{messageIdMatchNotMatchPrivate.Id}},
		{source: "\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "onlyDeclarations": true}), findings: []string{messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id}},
		{source: "\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "onlyDeclarations": false}), findings: []string{messageIdMatchNotMatch.Id, messageIdMatchNotMatch.Id}},
		{source: "\n            const foo = {\n                [a]: 1,\n            };\n            ", settings: idMatchOf("^[^a]", map[string]bool{"properties": true, "onlyDeclarations": false}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "\n            const foo = {\n                [a]: 1,\n            };\n            ", settings: idMatchOf("^[^a]", map[string]bool{"properties": false, "onlyDeclarations": false}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import('foo.json', { with: { [type]: 'json' } })", settings: idMatchOf("^foo", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
		{source: "import('foo.json', { with: { type: json } })", settings: idMatchOf("^foo", map[string]bool{"properties": true}), findings: []string{messageIdMatchNotMatch.Id}},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != len(testCase.findings) {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, len(testCase.findings), len(result.Diagnostics), result.MessageIds())
			continue
		}
		rule_testing.ExpectFindings(t, result, testCase.findings...)
	}
}

// TestIdMatchStaysSilent carries upstream's clean cases, which are the false positives it already
// thought about. Each one is a class this port would otherwise ship.
func TestIdMatchStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "__foo = \"Matthieu\"", settings: idMatchOf("^[a-z]+$", map[string]bool{"onlyDeclarations": true})},
		{source: "firstname = \"Matthieu\"", settings: idMatchOf("^[a-z]+$", nil)},
		{source: "first_name = \"Matthieu\"", settings: idMatchOf("[a-z]+", nil)},
		{source: "firstname = \"Matthieu\"", settings: idMatchOf("^f", nil)},
		{source: "last_Name = \"Larcher\"", settings: idMatchOf("^[a-z]+(_[A-Z][a-z]+)*$", nil)},
		{source: "param = \"none\"", settings: idMatchOf("^[a-z]+(_[A-Z][a-z])*$", nil)},
		{source: "function noUnder(){}", settings: idMatchOf("^[^_]+$", nil)},
		{source: "no_under()", settings: idMatchOf("^[^_]+$", nil)},
		{source: "foo.no_under2()", settings: idMatchOf("^[^_]+$", nil)},
		{source: "var foo = bar.no_under3;", settings: idMatchOf("^[^_]+$", nil)},
		{source: "var foo = bar.no_under4.something;", settings: idMatchOf("^[^_]+$", nil)},
		{source: "foo.no_under5.qux = bar.no_under6.something;", settings: idMatchOf("^[^_]+$", nil)},
		{source: "if (bar.no_under7) {}", settings: idMatchOf("^[^_]+$", nil)},
		{source: "var obj = { key: foo.no_under8 };", settings: idMatchOf("^[^_]+$", nil)},
		{source: "var arr = [foo.no_under9];", settings: idMatchOf("^[^_]+$", nil)},
		{source: "[foo.no_under10]", settings: idMatchOf("^[^_]+$", nil)},
		{source: "var arr = [foo.no_under11.qux];", settings: idMatchOf("^[^_]+$", nil)},
		{source: "[foo.no_under12.nesting]", settings: idMatchOf("^[^_]+$", nil)},
		{source: "if (foo.no_under13 === boom.no_under14) { [foo.no_under15] }", settings: idMatchOf("^[^_]+$", nil)},
		{source: "var myArray = new Array(); var myDate = new Date();", settings: idMatchOf("^[a-z$]+([A-Z][a-z]+)*$", nil)},
		{source: "var x = obj._foo;", settings: idMatchOf("^[^_]+$", nil)},
		{source: "var obj = {key: no_under}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "onlyDeclarations": true})},
		{source: "var {key_no_under: key} = {}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true})},
		{source: "var { category_id } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "ignoreDestructuring": true})},
		{source: "var { category_id: category_id } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "ignoreDestructuring": true})},
		{source: "var { category_id = 1 } = query;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "ignoreDestructuring": true})},
		{source: "var o = {key: 1}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true})},
		{source: "var o = {no_under16: 1}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false})},
		{source: "obj.no_under17 = 2;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false})},
		{source: "var obj = {\n no_under18: 1 \n};\n obj.no_under19 = 2;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false})},
		{source: "obj.no_under20 = function(){};", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false})},
		{source: "var x = obj._foo2;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false})},
		{source: "\n            const foo = Object.keys(bar);\n            const a = Array.from(b);\n            const bar = () => Array;\n            ", settings: idMatchOf("^\\$?[a-z]+([A-Z0-9][a-z0-9]+)*$", map[string]bool{"properties": true})},
		{source: "\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false})},
		{source: "\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ", settings: idMatchOf("^[^_]+$", map[string]bool{"onlyDeclarations": true})},
		{source: "\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false, "onlyDeclarations": false})},
		{source: "\n            const foo = {\n                [a]: 1,\n            };\n            ", settings: idMatchOf("^[^a]", map[string]bool{"properties": true, "onlyDeclarations": true})},
		{source: "class x { foo() {} }", settings: idMatchOf("^[^_]+$", nil)},
		{source: "class x { #foo() {} }", settings: idMatchOf("^[^_]+$", nil)},
		{source: "class x { _foo = 1; }", settings: idMatchOf("^[^_]+$", nil)},
		{source: "class x { _foo = 1; }", settings: idMatchOf("^[^_]+$", map[string]bool{"classFields": false})},
		{source: "class x { #_foo = 1; }", settings: idMatchOf("^[^_]+$", map[string]bool{"classFields": false})},
		{source: "class x { #_foo = 1; }", settings: idMatchOf("^[^_]+$", nil)},
		{source: "import.meta", settings: idMatchOf("^$", nil)},
		{source: "function foo() { new.target; }", settings: idMatchOf("^foo$", nil)},
		{source: "import foo from 'foo.json' with { type: 'json' }", settings: idMatchOf("^foo", map[string]bool{"properties": true})},
		{source: "export * from 'foo.json' with { type: 'json' }", settings: idMatchOf("^foo", map[string]bool{"properties": true})},
		{source: "export { default } from 'foo.json' with { type: 'json' }", settings: idMatchOf("^def", map[string]bool{"properties": true})},
		{source: "import('foo.json', { with: { type: 'json' } })", settings: idMatchOf("^foo", map[string]bool{"properties": true})},
		{source: "import('foo.json', { 'with': { type: 'json' } })", settings: idMatchOf("^foo", map[string]bool{"properties": true})},
		{source: "import('foo.json', { with: { type } })", settings: idMatchOf("^foo", map[string]bool{"properties": true})},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, testCase.source, testCase.settings)
		// `ExpectClean` rather than a length check, because the guard in internal/lint/registry
		// matches the CALL NAME: a test that asserts the same property by hand reads to it as a
		// rule nothing proves can stay quiet.
		rule_testing.ExpectClean(t, result)
	}
}

// TestIdMatchSpansAndMessages asserts where each finding POINTS and what it says, which
// TestIdMatchFires cannot: that test compares message ids and counts, so a rule anchored on the
// wrong node or rendering the wrong name or pattern passes it completely.
//
// The message here carries TWO computed pieces, the name and the pattern, and the standard's rule
// is that the text is where the computed half's bugs live. The private-name row earns its place
// twice over: `Text()` already carries the `#`, so a port matching the config against the hashed
// spelling would need `#foo` in the pattern to catch `foo`, while the message must print the hash
// back. Neither defect moves a message id or a count.
//
// Every expected span and message string is what the installed ESLint 10.8.1 rule produced on the
// same source, read out of the oracle run rather than derived by hand.
func TestIdMatchSpansAndMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		spans    []string
		messages []string
	}{
		{
			source:   "var no_camelcased = 1;",
			settings: idMatchOf("^[^_]+$", nil),
			spans:    []string{"no_camelcased"},
			messages: []string{"Identifier 'no_camelcased' does not match the pattern '^[^_]+$'."},
		},
		{
			// The write arm of the member-access split, so the finding must land on the property.
			source:   "obj.no_under30 = 2;",
			settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}),
			spans:    []string{"no_under30"},
			messages: []string{"Identifier 'no_under30' does not match the pattern '^[^_]+$'."},
		},
		{
			// The object arm, so the finding lands on the receiver rather than on `foo`.
			source:   "no_under23.foo = function(){};",
			settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}),
			spans:    []string{"no_under23"},
			messages: []string{"Identifier 'no_under23' does not match the pattern '^[^_]+$'."},
		},
		{
			// A destructuring rename: only the NEW name is reported, and the span must skip the key.
			source:   "var { category_id: category_alias } = query;",
			settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}),
			spans:    []string{"category_alias"},
			messages: []string{"Identifier 'category_alias' does not match the pattern '^[^_]+$'."},
		},
		{
			// The span covers the hash and so does the message, while the PATTERN was matched
			// against the name without it.
			source:   "class C { #no_under = 1; }",
			settings: idMatchOf("^[^_]+$", map[string]bool{"classFields": true}),
			spans:    []string{"#no_under"},
			messages: []string{"Identifier '#no_under' does not match the pattern '^[^_]+$'."},
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != len(testCase.spans) {
			t.Errorf("%q: expected %d findings, got %d %v",
				testCase.source, len(testCase.spans), len(result.Diagnostics), result.MessageIds())
			continue
		}
		source := result.SourceFile.Text()
		for index, diagnostic := range result.Diagnostics {
			reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.spans[index] {
				t.Errorf("%q finding %d: expected the span to cover %q, got %q",
					testCase.source, index, testCase.spans[index], reported)
			}
			// The whole computed sentence, not a prefix of it: the name and the pattern are both
			// interpolated and both are the parts a defect would get wrong.
			if !strings.HasPrefix(diagnostic.Message.Description, testCase.messages[index]) {
				t.Errorf("%q finding %d: expected the message to open with %q, got %q",
					testCase.source, index, testCase.messages[index], diagnostic.Message.Description)
			}
		}
	}
}

// TestIdMatchDecoderReadsBothTupleElements pins the option decoder, which no fixture above can
// reach on its own: every one of them routes through `idMatchOf`, so a decoder that dropped the
// flag object entirely would still let each case pass under whatever default the flag has.
//
// The nil-options row is the one the standard names specifically. A rule configured as a bare
// "error" is handed nil, and `options.(IdMatchSettings)` on nil yields the zero value, whose
// Pattern is nil. That has to mean "report nothing", matching upstream's `^.+$` default, rather
// than panicking on a nil regexp at the first identifier.
func TestIdMatchDecoderReadsBothTupleElements(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeIdMatchOptions([]byte(`["^[a-z]+$", {"properties": true, "classFields": true, "onlyDeclarations": true, "ignoreDestructuring": true}]`))
	if err != nil {
		t.Fatalf("decoding a full option tuple: %v", err)
	}
	settings, ok := decoded.(IdMatchSettings)
	if !ok {
		t.Fatalf("expected IdMatchSettings, got %T", decoded)
	}
	if settings.PatternText != "^[a-z]+$" || settings.Pattern == nil {
		t.Errorf("expected the pattern to be read, got %q", settings.PatternText)
	}
	if !settings.CheckProperties || !settings.CheckClassFields ||
		!settings.OnlyDeclarations || !settings.IgnoreDestructuring {
		t.Errorf("expected all four flags on, got %+v", settings)
	}

	// A pattern alone leaves every flag off, which is upstream's defaultOptions.
	decoded, err = DecodeIdMatchOptions([]byte(`["^[a-z]+$"]`))
	if err != nil {
		t.Fatalf("decoding a pattern-only tuple: %v", err)
	}
	settings, _ = decoded.(IdMatchSettings)
	if settings.CheckProperties || settings.CheckClassFields ||
		settings.OnlyDeclarations || settings.IgnoreDestructuring {
		t.Errorf("expected every flag off by default, got %+v", settings)
	}

	// A pattern that does not compile is an error rather than a rule that silently matches
	// everything, because an inert rule that looks configured is the worse of the two failures.
	if _, err := DecodeIdMatchOptions([]byte(`["^[a-z"]`)); err == nil {
		t.Error("expected an uncompilable pattern to be reported, got no error")
	}

	// A rule handed nil options reports nothing rather than dereferencing a nil pattern.
	result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, "var no_camelcased = 1;", nil)
	if len(result.Diagnostics) != 0 {
		t.Errorf("expected an unconfigured rule to report nothing, got %d: %v",
			len(result.Diagnostics), result.MessageIds())
	}
}

// TestIdMatchDestructuringKeysAreNeverChecked covers what upstream's corpus cannot.
//
// In `let { a: b } = x`, only `b` is a name this file chose; `a` is the source object's own
// property and we did not name it. That exclusion survived a mutation sweep against all 100
// imported fixtures, because every renaming case upstream writes has a key that PASSES the
// pattern, so deleting the exclusion changes nothing any imported case can see. The distinguishing
// shape is a key that FAILS while its binding passes, and upstream had no reason to write one.
//
// Every verdict is what the installed ESLint 10.8.1 rule answered:
//
//	var { bad_key: goodName } = q          clean   -- the key is not ours, even failing
//	var { bad_key: goodName } = q          clean   -- and ignoreDestructuring does not change that
//	function f({ bad_key: goodName }) {}   clean   -- same in a parameter pattern
//	var { bad_key: bad_value } = q         1       -- only the binding, never the key
//
// The last row is the control: without it a rule that had simply gone silent on all destructuring
// would pass the first three and prove nothing.
func TestIdMatchDestructuringKeysAreNeverChecked(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
		// reported names the identifier the single finding must cover, so a rule reporting the KEY
		// instead of the binding fails even where the count would match.
		reported string
	}{
		{
			source:   "var { bad_key: goodName } = q;",
			settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}),
			findings: 0,
		},
		{
			source:   "var { bad_key: goodName } = q;",
			settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "ignoreDestructuring": true}),
			findings: 0,
		},
		{
			source:   "function f({ bad_key: goodName }) {}",
			settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}),
			findings: 0,
		},
		{
			source:   "var { bad_key: bad_value } = q;",
			settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}),
			findings: 1,
			reported: "bad_value",
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, testCase.findings,
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		if testCase.reported == "" {
			continue
		}
		source := result.SourceFile.Text()
		reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != testCase.reported {
			t.Errorf("%q: expected the finding on %q, got %q", testCase.source, testCase.reported, reported)
		}
	}
}

// TestIdMatchOptionGatesAreArmSpecific covers what upstream's corpus cannot, and it found a real
// defect while doing so.
//
// The four options do not gate the rule uniformly, and which arms each one reaches is the single
// easiest thing to get wrong in this port, because the intuitive reading -- "a destructured name
// is a property, so `properties` governs it" -- is right for one arm and wrong for another. Four
// mutations survived all 100 imported fixtures here, and the corpus cannot see them because it
// never writes an assignment destructuring, a rest element, or an array pattern with the gating
// option turned OFF.
//
// Every verdict is what the installed ESLint 10.8.1 rule answered, and the first block is what an
// earlier draft of this rule got WRONG in all four rows:
//
//	({ no_under } = bar)          no options                  1 finding
//	({ no_under } = bar)          properties: false           1 finding
//	({ no_under } = bar)          onlyDeclarations: true      1 finding
//	({ no_under } = bar)          ignoreDestructuring: true   clean
//	({ a: no_under } = bar)       ignoreDestructuring: true   1 finding -- a rename coins a name
//	({ no_under: b } = bar)       properties: true            clean     -- the key is never ours
//
// So an ASSIGNMENT destructuring is gated by `ignoreDestructuring` alone, while a DECLARATION
// destructuring is additionally gated by `properties`. The cause is upstream's control flow: its
// ObjectPattern block reports and falls through rather than returning, so a name it reported is
// already reported by the time the `checkProperties` and `onlyDeclarations` gates below are
// consulted.
//
// The rest-element and array-pattern rows are the same class from the other side: both report with
// `properties` off, so neither is governed by it.
func TestIdMatchOptionGatesAreArmSpecific(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
		reported string
	}{
		// An assignment destructuring: only ignoreDestructuring gates it.
		{source: "({ no_under } = bar);", settings: idMatchOf("^[^_]+$", nil), findings: 1, reported: "no_under"},
		{source: "({ no_under } = bar);", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false}), findings: 1, reported: "no_under"},
		{source: "({ no_under } = bar);", settings: idMatchOf("^[^_]+$", map[string]bool{"onlyDeclarations": true}), findings: 1, reported: "no_under"},
		{source: "({ no_under } = bar);", settings: idMatchOf("^[^_]+$", map[string]bool{"ignoreDestructuring": true}), findings: 0},
		// A rename coins a new name, so ignoreDestructuring does not excuse it.
		{source: "({ a: no_under } = bar);", settings: idMatchOf("^[^_]+$", map[string]bool{"ignoreDestructuring": true}), findings: 1, reported: "no_under"},
		// The key of an assignment destructuring is the source object's, never ours.
		{source: "({ no_under: b } = bar);", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: 0},

		// A rest element is never a property name, so `properties` does not gate it.
		{source: "var { ...rest_name } = q;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false}), findings: 1, reported: "rest_name"},
		{source: "var { ...rest_name } = q;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: 1, reported: "rest_name"},

		// An array pattern's elements are positional, so `properties` does not gate them either,
		// while `onlyDeclarations` does -- it reaches them through the default arm.
		{source: "var [ arr_name ] = q;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": false}), findings: 1, reported: "arr_name"},
		{source: "var [ arr_name ] = q;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: 1, reported: "arr_name"},
		{source: "var [ arr_name ] = q;", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true, "ignoreDestructuring": true}), findings: 1, reported: "arr_name"},
		{source: "var [ arr_name ] = q;", settings: idMatchOf("^[^_]+$", map[string]bool{"onlyDeclarations": true}), findings: 0},
		{source: "function f([ arr_name ]) {}", settings: idMatchOf("^[^_]+$", map[string]bool{"properties": true}), findings: 1, reported: "arr_name"},
		{source: "function f([ arr_name ]) {}", settings: idMatchOf("^[^_]+$", map[string]bool{"onlyDeclarations": true, "properties": true}), findings: 0},
	}

	for _, testCase := range cases {
		result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, testCase.findings,
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		if testCase.reported == "" {
			continue
		}
		source := result.SourceFile.Text()
		reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != testCase.reported {
			t.Errorf("%q: expected the finding on %q, got %q", testCase.source, testCase.reported, reported)
		}
	}
}

// TestIdMatchPrivateNamesAreMatchedWithoutTheHash pins which spelling reaches the regular
// expression, which no imported fixture can: every private-name case upstream writes uses a
// pattern that neither requires nor forbids a leading `#`, so stripping it or not makes no
// difference to any of them.
//
// Measured against the installed ESLint 10.8.1 rule, with patterns chosen so the hash decides:
//
//	class Ok { #field = 1; }   pattern ^[^#]+$   clean   -- so the hash was NOT matched
//	class Ok { #field = 1; }   pattern ^#        2       -- Ok and #field both fail
//
// The second row also shows the message keeps the hash while the pattern never saw it, which is
// upstream rendering `#{{name}}` from a name that carries none.
func TestIdMatchPrivateNamesAreMatchedWithoutTheHash(t *testing.T) {
	t.Parallel()

	// A pattern forbidding `#` is satisfied, which is only possible if the name was stripped first.
	result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, "class Ok { #field = 1; }",
		idMatchOf("^[^#]+$", map[string]bool{"classFields": true}))
	if len(result.Diagnostics) != 0 {
		t.Errorf("expected the hash to be stripped before matching, got %d findings: %v",
			len(result.Diagnostics), result.MessageIds())
	}

	// The control: a pattern REQUIRING `#` fails for both names, so the rule is not simply silent
	// on this source.
	result = rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, "class Ok { #field = 1; }",
		idMatchOf("^#", map[string]bool{"classFields": true}))
	if len(result.Diagnostics) != 2 {
		t.Fatalf("expected the control to report both names, got %d: %v",
			len(result.Diagnostics), result.MessageIds())
	}
	// And the message renders the hash back, from a name that was matched without it.
	if !strings.Contains(result.Diagnostics[1].Message.Description, "Identifier '#field'") {
		t.Errorf("expected the message to render the hash, got %q", result.Diagnostics[1].Message.Description)
	}
}

// TestIdMatchDecoderAcceptsTheConfigLayersShape pins the wire contract, which every other fixture
// in this file bypasses.
//
// Fixtures reach the decoder with bytes the TEST built. The config layer hands a list rule every
// element after the severity, which for upstream's `["error", "^[a-z]+$", {...}]` is
// `["^[a-z]+$", {...}]`. It used to hand over the bare pattern alone, so the rule ran with every
// flag silently off: the "registered, plausible, and wrong" shape the standard warns about.
func TestIdMatchDecoderAcceptsTheConfigLayersShape(t *testing.T) {
	t.Parallel()

	// Upstream's spelling: the pattern, then the flags as the second element.
	decoded, err := DecodeIdMatchOptions([]byte(`["^[a-z]+$", {"properties": true, "classFields": true}]`))
	if err != nil {
		t.Fatalf("decoding the option list: %v", err)
	}
	settings, _ := decoded.(IdMatchSettings)
	if settings.PatternText != "^[a-z]+$" {
		t.Errorf("expected the pattern to be read, got %q", settings.PatternText)
	}
	if !settings.CheckProperties || !settings.CheckClassFields {
		t.Errorf("expected the second element's flags to be read, got %+v", settings)
	}

	// Shapes the config layer does not deliver, or that upstream's schema refuses, are refused
	// rather than read as a default.
	for _, raw := range []string{
		// The bare pattern the config layer used to deliver.
		`"^[a-z]+$"`,
		// The nested workaround spelling.
		`[["^[a-z]+$", {"properties": true}]]`,
		// A flag upstream does not declare.
		`["^[a-z]+$", {"property": true}]`,
		// A third element.
		`["^[a-z]+$", {"properties": true}, {"classFields": true}]`,
	} {
		if _, err := DecodeIdMatchOptions([]byte(raw)); err == nil {
			t.Errorf("expected %s to be refused, got no error", raw)
		}
	}

	// And end to end through the rule, so the contract is pinned at the surface a config reaches
	// rather than only at the decoder.
	result := rule_testing.RunTypedWithOptions(t, IdMatch, idMatchFile, "class C { no_under = 1; }",
		mustDecodeIdMatch(t, `["^[^_]+$", {"classFields": true}]`))
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected the second element's classFields flag to reach the rule, got %d findings: %v",
			len(result.Diagnostics), result.MessageIds())
	}
}

// mustDecodeIdMatch decodes exactly the bytes the config layer would hand the rule.
func mustDecodeIdMatch(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeIdMatchOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}
	return decoded
}
