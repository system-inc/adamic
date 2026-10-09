package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const tripleSlashFile = "/repository/source/Thing.ts"

// decodeTripleSlashOptions runs a configuration through the real decoder rather than building the
// options struct directly.
//
// The decoder holds three defaults and a per-key value filter, so a fixture that constructed
// `TripleSlashReferenceOptions` by hand would test the rule while leaving the half of the port that
// reads the config unexercised. Every case below names its configuration the way a user writes it.
func decodeTripleSlashOptions(t *testing.T, configuration string) any {
	t.Helper()

	options, err := DecodeTripleSlashReferenceOptions([]byte(configuration))
	if err != nil {
		t.Fatalf("could not decode %s: %v", configuration, err)
	}
	return options
}

// TestTripleSlashReferenceStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All sixteen of upstream's passing inputs, extracted by the fixture tool and written out by a
// script so no escape sequence passed through a shell or a keyboard on the way here.
func TestTripleSlashReferenceStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
	}{
		{"{ \"lib\": \"never\", \"path\": \"never\", \"types\": \"never\" }", "\n                    // <reference path=\"foo\" />\n                    // <reference types=\"bar\" />\n                    // <reference lib=\"baz\" />\n                    import * as foo from 'foo';\n                    import * as bar from 'bar';\n                    import * as baz from 'baz';\n                  "},
		{"{ \"lib\": \"never\", \"path\": \"never\", \"types\": \"never\" }", "\n                    // <reference path=\"foo\" />\n                    // <reference types=\"bar\" />\n                    // <reference lib=\"baz\" />\n                    import foo = require('foo');\n                    import bar = require('bar');\n                    import baz = require('baz');\n                  "},
		{"{ \"lib\": \"always\", \"path\": \"always\", \"types\": \"always\" }", "\n                    /// <reference path=\"foo\" />\n                    /// <reference types=\"bar\" />\n                    /// <reference lib=\"baz\" />\n                    import * as foo from 'foo';\n                    import * as bar from 'bar';\n                    import * as baz from 'baz';\n                  "},
		{"{ \"lib\": \"always\", \"path\": \"always\", \"types\": \"always\" }", "\n                    /// <reference path=\"foo\" />\n                    /// <reference types=\"bar\" />\n                    /// <reference lib=\"baz\" />\n                    import foo = require('foo');\n                    import bar = require('bar');\n                    import baz = require('baz');\n                  "},
		{"{ \"lib\": \"always\", \"path\": \"always\", \"types\": \"always\" }", "\n                    /// <reference path=\"foo\" />\n                    /// <reference types=\"bar\" />\n                    /// <reference lib=\"baz\" />\n                    import foo = foo;\n                    import bar = bar;\n                    import baz = baz;\n                  "},
		{"{ \"lib\": \"always\", \"path\": \"always\", \"types\": \"always\" }", "\n                    /// <reference path=\"foo\" />\n                    /// <reference types=\"bar\" />\n                    /// <reference lib=\"baz\" />\n                    import foo = foo.foo;\n                    import bar = bar.bar.bar.bar;\n                    import baz = baz.baz;\n                  "},
		{"{ \"path\": \"never\" }", "import * as foo from 'foo';"},
		{"{ \"path\": \"never\" }", "import foo = require('foo');"},
		{"{ \"types\": \"never\" }", "import * as foo from 'foo';"},
		{"{ \"types\": \"never\" }", "import foo = require('foo');"},
		{"{ \"lib\": \"never\" }", "import * as foo from 'foo';"},
		{"{ \"lib\": \"never\" }", "import foo = require('foo');"},
		{"{ \"types\": \"prefer-import\" }", "import * as foo from 'foo';"},
		{"{ \"types\": \"prefer-import\" }", "import foo = require('foo');"},
		{"{ \"types\": \"prefer-import\" }", "\n                    /// <reference types=\"foo\" />\n                    import * as bar from 'bar';\n                  "},
		{"{ \"lib\": \"never\", \"path\": \"never\", \"types\": \"never\" }", "\n                    /*\n                    /// <reference types=\"foo\" />\n                    */\n                    import * as foo from 'foo';\n                  "},
	}
	for index, testCase := range cases {
		t.Run(testCase.configuration, func(t *testing.T) {
			t.Parallel()
			_ = index
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, TripleSlashReference,
				tripleSlashFile, testCase.sourceText,
				decodeTripleSlashOptions(t, testCase.configuration)))
		})
	}
}

// TestTripleSlashReferenceFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// The snapshot records five diagnostics against these five inputs, one each, so the extractor
// reported no discrepancy and every count below is one.
func TestTripleSlashReferenceFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		count         int
	}{
		{"{ \"types\": \"prefer-import\" }", "\n            /// <reference types=\"foo\" />\n            import * as foo from 'foo';\n                  ", 1},
		{"{ \"types\": \"prefer-import\" }", "\n            /// <reference types=\"foo\" />\n            import foo = require('foo');\n                  ", 1},
		{"{ \"path\": \"never\" }", "/// <reference path=\"foo\" />", 1},
		{"{ \"types\": \"never\" }", "/// <reference types=\"foo\" />", 1},
		{"{ \"lib\": \"never\" }", "/// <reference lib=\"foo\" />", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.configuration, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
				testCase.sourceText, decodeTripleSlashOptions(t, testCase.configuration))
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "tripleSlashReference"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestTripleSlashReferenceFiresOnCasesUpstreamDoesNotCover exercises the decisions our own code
// makes that the imported corpus never reaches.
//
// Every case here was measured against the release oxlint binary before it was written, because
// reading the upstream source settles almost none of them: the attribute scan, the position cutoff
// and the prefer-import counting each have several plausible readings and the corpus separates
// none of them.
func TestTripleSlashReferenceFiresOnCasesUpstreamDoesNotCover(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		configuration string
		sourceText    string
		count         int
	}{
		{"a directive above the first statement is still a directive", "{}", "/// <reference path=\"foo\" />\nconst a = 1;\n", 1},
		{"four slashes, because the test is on the third character", "{}", "//// <reference path=\"foo\" />\n", 1},
		{"a block comment whose content opens with a slash", "{}", "/*/ <reference path=\"foo\" /> */\n", 1},
		{"attributes read in whitespace order, path first", "{}", "/// <reference path=\"bar\" lib=\"foo\" />\n", 1},
		{"an unquoted value", "{}", "/// <reference path=foo />\n", 1},
		{"no space before the closing slash", "{}", "/// <reference path=\"foo\"/>\n", 1},
		{"text around the reference", "{}", "/// junk <reference path=\"foo\" /> junk\n", 1},
		{"two imports of one module report the one directive twice", "{}", "/// <reference types=\"foo\" />\nimport * as a from 'foo';\nimport * as b from 'foo';\n", 2},
		{"a type-only import counts", "{}", "/// <reference types=\"foo\" />\nimport type { T } from 'foo';\n", 1},
		{"a bare side-effect import counts", "{}", "/// <reference types=\"foo\" />\nimport 'foo';\n", 1},
		{"types never reports without any import", "{\"types\": \"never\"}", "/// <reference types=\"foo\" />\n", 1},
		{"lib never", "{\"lib\": \"never\"}", "/// <reference lib=\"foo\" />\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
				testCase.sourceText, decodeTripleSlashOptions(t, testCase.configuration))
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "tripleSlashReference"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestTripleSlashReferenceStaysSilentOnCasesUpstreamDoesNotCover is the other half, and it is the
// half that catches a rule which is merely eager.
//
// The position cases are the most valuable here: a scanner that matched a directive anywhere in the
// file would pass every fires case above and every upstream fail case, and would report on real
// source that TypeScript does not treat as a directive at all.
func TestTripleSlashReferenceStaysSilentOnCasesUpstreamDoesNotCover(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		configuration string
		sourceText    string
	}{
		{"a directive after the first statement is not a directive", "{}", "const a = 1;\n/// <reference path=\"foo\" />\n"},
		{"a directive between two statements", "{}", "const a = 1;\n/// <reference path=\"foo\" />\nconst b = 2;\n"},
		{"a directive trailing code on the same line", "{}", "const a = 1; /// <reference path=\"foo\" />\n"},
		{"a double slash is not a directive", "{}", "// <reference path=\"foo\" />\n"},
		{"a block comment whose content opens with a space", "{}", "/* <reference path=\"foo\" /> */\n"},
		{"a jsdoc block, whose content opens with a star", "{}", "/** <reference path=\"foo\" /> */\n"},
		{"attributes read in whitespace order, lib first and lib defaults to always", "{}", "/// <reference lib=\"foo\" path=\"bar\" />\n"},
		{"a value containing an equals sign declines the whole comment", "{}", "/// <reference path=\"a=b\" />\n"},
		{"no-default-lib is not an attribute this rule reads", "{}", "/// <reference no-default-lib=\"true\" />\n"},
		// A malformed first attribute declines the WHOLE comment rather than falling through to a
		// later well-formed one. Upstream returns from `get_attr_key_and_value` on the first match
		// whose split is not exactly two parts, and the loop never continues. Measured on the release
		// binary: both of these are silent even under a configuration that would report their second
		// attribute. This is the fixture that catches turning that return into a continue.
		{"a malformed first attribute declines the comment rather than falling through", "{\"types\": \"never\"}", "/// <reference path=\"a=b\" types=\"foo\" />\n"},
		{"a malformed lib attribute declines a following path attribute", "{}", "/// <reference lib=\"a=b\" path=\"foo\" />\n"},
		{"lib defaults to always", "{}", "/// <reference lib=\"foo\" />\n"},
		{"types defaults to prefer-import, so no import means no finding", "{}", "/// <reference types=\"foo\" />\n"},
		{"a dynamic import is not one of the two shapes", "{}", "/// <reference types=\"foo\" />\nconst p = import('foo');\n"},
		{"a require call is not an import-equals declaration", "{}", "/// <reference types=\"foo\" />\nconst q = require('foo');\n"},
		{"an export-from is not an import", "{}", "/// <reference types=\"foo\" />\nexport * from 'foo';\n"},
		{"an import nested in a namespace is not a top-level statement", "{}", "/// <reference types=\"foo\" />\ndeclare module M { import foo = require('foo'); }\n"},
		{"an import-equals naming a qualified name rather than a module", "{}", "/// <reference types=\"foo\" />\nimport foo = foo.bar;\n"},
		{"path always", "{\"path\": \"always\"}", "/// <reference path=\"foo\" />\n"},
		{"prefer-import is refused for path, so the default of never is kept and this reports elsewhere", "{\"lib\": \"prefer-import\"}", "/// <reference lib=\"foo\" />\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, TripleSlashReference,
				tripleSlashFile, testCase.sourceText,
				decodeTripleSlashOptions(t, testCase.configuration)))
		})
	}
}

// TestTripleSlashReferencePointsAtTheDirective asserts where each finding lands.
//
// A comment is reported through ReportRange rather than ReportNode, so nothing in the harness is
// trimming trivia on this rule's behalf and the span is entirely the rule's own arithmetic. Every
// message-id fixture above stays green over a span off by the two leading slashes.
func TestTripleSlashReferencePointsAtTheDirective(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		configuration string
		sourceText    string
		reported      string
	}{
		{"the whole comment including the leading slashes", "{}", "/// <reference path=\"foo\" />\n", "/// <reference path=\"foo\" />"},
		{"four slashes are all inside the span", "{}", "//// <reference path=\"foo\" />\n", "//// <reference path=\"foo\" />"},
		{"a block comment reports its closing delimiter too", "{}", "/*/ <reference path=\"foo\" /> */\n", "/*/ <reference path=\"foo\" /> */"},
		{"indentation is not part of the span", "{}", "\n    /// <reference path=\"foo\" />\n", "/// <reference path=\"foo\" />"},
		{"a prefer-import finding points at the directive, not at the import", "{}", "/// <reference types=\"foo\" />\nimport * as foo from 'foo';\n", "/// <reference types=\"foo\" />"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
				testCase.sourceText, decodeTripleSlashOptions(t, testCase.configuration))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.reported {
				t.Fatalf("expected the finding to cover %q, got %q", testCase.reported, reported)
			}
		})
	}
}

// TestTripleSlashReferenceNamesTheModuleInItsMessage asserts the rendered text.
//
// The module name is concatenated into the description, so it is the one part of this rule a
// message-id assertion structurally cannot see. Several of these names are upstream's own trimming
// quirks rather than anything a reader would call correct, and each was measured on the release
// binary: `path='foo'` really does render with the single quotes attached.
//
// The expected strings are literals typed here rather than references to
// messageTripleSlashReference, because comparing a finding against the constant it was built from
// moves both sides together under mutation and asserts nothing.
func TestTripleSlashReferenceNamesTheModuleInItsMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		configuration string
		sourceText    string
		moduleName    string
	}{
		{"the module name is the trimmed attribute value", "{}", "/// <reference path=\"foo\" />\n", "foo"},
		{"single quotes survive, because only the double quote is trimmed", "{}", "/// <reference path='foo' />\n", "'foo'"},
		{"an empty value renders an empty name", "{}", "/// <reference path=\"\" />\n", ""},
		{"trailing slashes are stripped from inside the value", "{}", "/// <reference path=\"foo///\"/>\n", "foo"},
		{"the second attribute is the one read when it is the first match", "{}", "/// <reference path=\"bar\" lib=\"foo\" />\n", "bar"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
				testCase.sourceText, decodeTripleSlashOptions(t, testCase.configuration))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			want := "This file pulls in `" + testCase.moduleName + "` with a triple-slash reference " +
				"directive. A directive is a global instruction to the compiler with no binding and no " +
				"local name, so nothing in the file records that the dependency is used here and no " +
				"tool can follow it the way it follows an import. Write an `import` instead, which " +
				"names what it brings in and participates in module resolution."
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("expected message\n  %q\ngot\n  %q", want, got)
			}
			if got := result.Diagnostics[0].Message.Id; got != "tripleSlashReference" {
				t.Fatalf("expected id %q, got %q", "tripleSlashReference", got)
			}
		})
	}
}

// TestTripleSlashReferenceKeepsTheLastOfTwoIdenticalDirectives pins the bookkeeping choice.
//
// Upstream records pending prefer-import directives in a map keyed by the module name, and a map
// insert overwrites. So two directives naming one module leave only the second span, and three
// leave only the third. That is a behavioral decision hiding inside what reads as a data-structure
// choice, and no fixture asserting message ids can see it: the count is one either way and only the
// position moves.
//
// Measured on the release binary before this was written. A file with three identical directives
// and one import reports once, against line three.
func TestTripleSlashReferenceKeepsTheLastOfTwoIdenticalDirectives(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantColumn int
	}{
		{
			"two identical directives report against the second",
			"/// <reference types=\"foo\" />\n/// <reference types=\"foo\" />\nimport * as a from 'foo';\n",
			30,
		},
		{
			"three identical directives report against the third",
			"/// <reference types=\"foo\" />\n/// <reference types=\"foo\" />\n/// <reference types=\"foo\" />\nimport * as a from 'foo';\n",
			60,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
				testCase.sourceText, decodeTripleSlashOptions(t, "{}"))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Range.Pos(); got != testCase.wantColumn {
				t.Fatalf("expected the finding to start at %d, got %d", testCase.wantColumn, got)
			}
		})
	}
}

// TestTripleSlashReferenceReportsTwoDirectivesInSourceOrder is the other side of the map.
//
// Two directives naming different modules each keep their own entry, so both report. Written
// alongside the overwrite test so the two cases that separate "keyed by name" from "keyed by
// position" sit together.
func TestTripleSlashReferenceReportsTwoDirectivesInSourceOrder(t *testing.T) {
	t.Parallel()

	const source = "/// <reference types=\"a\" />\n/// <reference types=\"b\" />\nimport * as y from 'b';\nimport * as x from 'a';\n"

	result := rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile, source,
		decodeTripleSlashOptions(t, "{}"))
	rule_testing.ExpectFindings(t, result, "tripleSlashReference", "tripleSlashReference")

	// The findings arrive in the order the import loop produces them rather than in source order,
	// because the second pass walks statements. The import of `b` is written first, so `b`'s
	// directive is reported first even though it sits second in the file.
	first := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if first != "/// <reference types=\"b\" />" {
		t.Fatalf("expected the first finding to be b's directive, got %q", first)
	}
}

// TestTripleSlashReferenceDeclinesJavaScript pins the source-type gate.
//
// Upstream's `should_run` reads `ctx.source_type().is_typescript()`, so the same directive reports
// in a `.ts` file and is silent in a `.js` one. Dropping the gate would report every triple-slash
// directive in every JavaScript file in a tree upstream leaves alone, and JavaScript is where these
// directives are most common.
func TestTripleSlashReferenceDeclinesJavaScript(t *testing.T) {
	t.Parallel()

	const source = "/// <reference path=\"foo\" />\n"

	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, TripleSlashReference, "/repository/source/Thing.ts", source,
			decodeTripleSlashOptions(t, "{}")),
		"tripleSlashReference")

	for _, fileName := range []string{"/repository/source/Thing.js", "/repository/source/Thing.jsx", "/repository/source/Thing.mjs"} {
		rule_testing.ExpectClean(t,
			rule_testing.RunWithOptions(t, TripleSlashReference, fileName, source,
				decodeTripleSlashOptions(t, "{}")))
	}

	// The TypeScript extensions the gate does admit, including the definition file where a
	// directive is most likely to be written deliberately.
	for _, fileName := range []string{"/repository/source/Thing.tsx", "/repository/source/Thing.mts", "/repository/source/Thing.d.ts"} {
		rule_testing.ExpectFindings(t,
			rule_testing.RunWithOptions(t, TripleSlashReference, fileName, source,
				decodeTripleSlashOptions(t, "{}")),
			"tripleSlashReference")
	}
}

// TestDecodeTripleSlashReferenceOptions asserts the three defaults and the per-key value filter.
//
// The defaults are the part of this rule most likely to be quietly wrong, because none of them is
// the Go zero value and a struct built by hand would carry three empty strings that match no arm.
// Our own rule inventory records this rule as taking no options at all, which is what makes writing
// the surface down here worth more than usual.
func TestDecodeTripleSlashReferenceOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		configuration string
		want          TripleSlashReferenceOptions
	}{
		{
			"an empty object is upstream's documented default",
			"{}",
			TripleSlashReferenceOptions{Lib: "always", Path: "never", Types: "prefer-import"},
		},
		{
			"each key set explicitly",
			`{"lib": "never", "path": "always", "types": "never"}`,
			TripleSlashReferenceOptions{Lib: "never", Path: "always", Types: "never"},
		},
		{
			"one key set leaves the other two at their defaults",
			`{"path": "always"}`,
			TripleSlashReferenceOptions{Lib: "always", Path: "always", Types: "prefer-import"},
		},
		{
			// The release binary refuses this configuration outright. There is no error channel
			// that reaches a user here, so the key keeps its default rather than falling into a
			// fourth state the rule never tests for.
			"prefer-import is not accepted for path",
			`{"path": "prefer-import"}`,
			TripleSlashReferenceOptions{Lib: "always", Path: "never", Types: "prefer-import"},
		},
		{
			"prefer-import is not accepted for lib",
			`{"lib": "prefer-import"}`,
			TripleSlashReferenceOptions{Lib: "always", Path: "never", Types: "prefer-import"},
		},
		{
			// Upstream's values are kebab-case and serde matches them exactly.
			"a wrongly-cased value keeps the default",
			`{"path": "Never"}`,
			TripleSlashReferenceOptions{Lib: "always", Path: "never", Types: "prefer-import"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeTripleSlashReferenceOptions([]byte(testCase.configuration))
			if err != nil {
				t.Fatalf("could not decode: %v", err)
			}
			got, ok := decoded.(TripleSlashReferenceOptions)
			if !ok {
				t.Fatalf("expected TripleSlashReferenceOptions, got %T", decoded)
			}
			if got != testCase.want {
				t.Fatalf("expected %+v, got %+v", testCase.want, got)
			}
		})
	}
}

// TestTripleSlashReferenceUsesDefaultsWhenTheConfigNamesNoOptions is the fixture for the defect the
// dry run found, and it is the only test here that does not go through the decoder.
//
// Every other fixture calls `DecodeTripleSlashReferenceOptions`, which is exactly why none of them
// could see this. A rule configured as a bare `"error"` is handed a nil `options` by the config
// layer: `rule.DecodeOptionsInto` reports an error on empty input, and
// `configuration.OptionsRegistry.Decode` turns that into nil for a rule whose options only tune it. A bare
// type assertion on nil yields the Go zero value, which for this rule is three empty settings
// matching no arm, so the rule declines every file.
//
// Measured before the fix: 3,407 files offered, 3,407 registrations, zero findings on the real tree
// and zero on a probe tree seeded with directives that must report. The registration count is what
// separated this from an unwired rule; the finding count is what said it was broken anyway.
//
// The nil case is the one the config actually produces. A value of the wrong type must not silently
// disable the rule either, and since #qmvkf83 it cannot: rule.OptionsAs panics on it, which
// TestTripleSlashReferenceRefusesOptionsOfTheWrongType holds.
func TestTripleSlashReferenceUsesDefaultsWhenTheConfigNamesNoOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options any
	}{
		{"nil, which is what a bare severity in the config produces", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// path defaults to never, so this must report.
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
					"/// <reference path=\"foo\" />\n", testCase.options),
				"tripleSlashReference")

			// lib defaults to always, so this must not. Both halves are needed: a fallback that
			// set every key to never would pass the first assertion alone.
			rule_testing.ExpectClean(t,
				rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
					"/// <reference lib=\"foo\" />\n", testCase.options))

			// types defaults to prefer-import, which is neither of the other two settings.
			rule_testing.ExpectClean(t,
				rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
					"/// <reference types=\"foo\" />\n", testCase.options))
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
					"/// <reference types=\"foo\" />\nimport * as foo from 'foo';\n", testCase.options),
				"tripleSlashReference")
		})
	}
}

// TestTripleSlashReferenceScansEveryDirectiveBeforeTheCutoff pins the loop's early exit.
//
// The scan breaks on the first comment at or past the cutoff rather than filtering, which is
// correct only because `comments.ForFile` returns comments in source order. That ordering is the
// shelf's contract and it is asserted in the shelf's own tests, but a `break` written against an
// unsorted list would drop every directive after the first out-of-order one and would do it
// silently, so the property is pinned from this side too.
//
// Four directives above the first statement, each reporting, is what an early exit fires too soon
// would break; the fifth below the statement is what a missing exit would wrongly report.
func TestTripleSlashReferenceScansEveryDirectiveBeforeTheCutoff(t *testing.T) {
	t.Parallel()

	const source = "/// <reference path=\"a\" />\n" +
		"/// <reference path=\"b\" />\n" +
		"/// <reference path=\"c\" />\n" +
		"/// <reference path=\"d\" />\n" +
		"export const value = 1;\n" +
		"/// <reference path=\"e\" />\n"

	result := rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile, source,
		decodeTripleSlashOptions(t, "{}"))
	rule_testing.ExpectFindings(t, result, "tripleSlashReference", "tripleSlashReference",
		"tripleSlashReference", "tripleSlashReference")

	// The four that report are the four above the statement, in source order, and `e` is absent.
	wantSpans := []string{
		"/// <reference path=\"a\" />",
		"/// <reference path=\"b\" />",
		"/// <reference path=\"c\" />",
		"/// <reference path=\"d\" />",
	}
	for index, want := range wantSpans {
		got := source[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
		if got != want {
			t.Fatalf("finding %d: expected %q, got %q", index, want, got)
		}
	}
}

// Options of a type the decoder never returns are a disagreement between the registration and the
// rule, and that fails loudly rather than defaulting: these used to fall back to the defaults, which
// is the silence that let unified-signatures ignore its options (#qmvkf83).
func TestTripleSlashReferenceRefusesOptionsOfTheWrongType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options any
	}{
		{"a value of some other type", "error"},
		{"a pointer rather than the value", &TripleSlashReferenceOptions{}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recover() == nil {
					t.Fatal("expected options of the wrong type to panic rather than default")
				}
			}()
			rule_testing.RunWithOptions(t, TripleSlashReference, tripleSlashFile,
				"/// <reference path=\"foo\" />\n", testCase.options)
		})
	}
}
