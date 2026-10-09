package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// newWrappersFile is where the fixtures pretend to live.
const newWrappersFile = "/repository/source/NewWrappers.ts"

// Every case here is upstream's, extracted from its test file rather than retyped.
//
// Verbatim from `eslint/tests/lib/rules/no-new-wrappers.js`: 7 valid and 4 invalid, each invalid
// case carrying exactly one `noConstructor` error with a stated line and column. The extractor
// stubbed the rule tester and captured the two arrays, and the bytes on disk here were compared
// against the bytes upstream ships rather than read.
//
// `RunTypedFiles` rather than `Run`, and that is required rather than preferred: this rule
// declares `NeedsTypeChecker`, and the plain harness supplies a nil checker, under which
// `resolvesToAGlobal` answers false for every input. Every Fires case would fail and every
// StaysSilent case would pass vacuously, so the syntax-only harness cannot prove this rule.
func TestNoNewWrappersFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"new String", "var a = new String('hello');"},
		{"new Number", "var a = new Number(10);"},
		{"new Boolean", "var a = new Boolean(false);"},
		{"the outer new String reports while the block-shadowed one does not", "\n            const a = new String('bar');\n            {\n                const String = CustomString;\n                const b = new String('foo');\n            }\n            "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTypedFiles(t, NoNewWrappers,
					map[string]string{newWrappersFile: testCase.sourceText}, newWrappersFile),
				"noNewWrappers")
		})
	}
}

// Five of upstream's seven valid cases, and four of them are shadowing.
//
// A parameter, an imported default, a `var` hoisted out of an `else` branch, and a `const` in a
// nested block each defeat a different naive implementation, which is why this rule reads the
// checker at all. The two remaining upstream valid cases are `globals: {String: "off"}` and
// `/* global Boolean:off */`; both are decided by a configuration layer we do not have, and they
// are pinned as reporting in TestNoNewWrappersReportsWhereUpstreamsGlobalsConfigWouldNot below.
//
// The import case needs a second file, since a bare unresolvable import leaves the checker with no
// declaration to find and the case would pass for the wrong reason. Its module is written here.
func TestNoNewWrappersStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"new Object is not a wrapper", "var a = new Object();"},
		{"String called without new, beside a static read off it", "var a = String('test'), b = String.fromCharCode(32);"},
		{"a parameter named Number shadows the global", "\n        function test(Number) {\n            return new Number;\n        }\n        "},
		{"a var hoisted out of the else branch shadows Boolean", "\n        if (foo) {\n            result = new Boolean(bar);\n        } else {\n            var Boolean = CustomBoolean;\n        }\n        "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTypedFiles(t, NoNewWrappers,
					map[string]string{newWrappersFile: testCase.sourceText}, newWrappersFile))
		})
	}
}

// Upstream's fourth valid case, given the module its import needs.
//
// The source text is upstream's verbatim. What is added is `./string`, because an import of a
// module that does not exist leaves the checker nothing to resolve to, `resolvesToAGlobal` answers
// false on the no-symbol path, and the case would come back clean without the shadow ever being
// the reason. That is the brief's "a passing case is not evidence about what you think": changing
// the one thing believed to make it pass has to move the verdict, and here it does, since deleting
// the import makes this report.
func TestNoNewWrappersDeclinesAnImportedShadow(t *testing.T) {
	t.Parallel()

	const stringModule = "/repository/source/string.ts"
	files := map[string]string{
		stringModule:    "export default class Str { constructor(_value: number) {} }\n",
		newWrappersFile: "\n            import String from \"./string\";\n            const str = new String(42);\n            ",
	}
	rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, NoNewWrappers, files, newWrappersFile))
}

// Upstream's remaining two valid cases, which report here, and the reasoning for why that is right.
//
// `{globals: {String: "off"}}` and `/* global Boolean:off */` both remove the name from ESLint's
// environment, so `getVariableByName` returns nothing and the rule declines. Neither mechanism
// exists in cohere: there is no globals configuration and no global comment reader, and the
// checker resolves both names to the TypeScript standard library regardless. Recording them as
// reporting is the honest pin, because the layer that makes them clean upstream is a layer above
// the rule rather than a judgment inside it.
//
// The global comment is carried verbatim so that if a global-comment reader ever lands, this test
// fails and names the case rather than the case quietly disappearing.
func TestNoNewWrappersReportsWhereUpstreamsGlobalsConfigWouldNot(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"globals String off", "new String()"},
		{"a global comment turning Boolean off", "\n        /* global Boolean:off */\n        assert(new Boolean);\n        "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTypedFiles(t, NoNewWrappers,
					map[string]string{newWrappersFile: testCase.sourceText}, newWrappersFile),
				"noNewWrappers")
		})
	}
}

// The span, asserted as text, because upstream points at the whole `new` expression.
//
// Upstream's corpus states the columns rather than implying them: `var a = new String('hello');`
// is columns 9 through 28, which is `new String('hello')` and not the callee. That is a real
// divergence from `no-new-native-nonconstructor` in this same package, which reports on the callee,
// so the two rules are not interchangeable on this point and copying its anchor would be wrong.
//
// `RunTypedFiles` writes each fixture through `strings.TrimSpace(contents)+"\n"`, so a span sliced
// out of the Go literal is offset by one on any case with a leading newline. The expectation is
// transformed the same way the harness transforms the input rather than sliced from the literal.
func TestNoNewWrappersReportsTheWholeNewExpression(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpan   string
	}{
		{"new String", "var a = new String('hello');", "new String('hello')"},
		{"new Number", "var a = new Number(10);", "new Number(10)"},
		{"new Boolean", "var a = new Boolean(false);", "new Boolean(false)"},
		{"the outer call in the block-shadow case", "\n            const a = new String('bar');\n            {\n                const String = CustomString;\n                const b = new String('foo');\n            }\n            ", "new String('bar')"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceOnDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			result := rule_testing.RunTypedFiles(t, NoNewWrappers,
				map[string]string{newWrappersFile: testCase.sourceText}, newWrappersFile)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			got := sourceOnDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.wantSpan {
				t.Fatalf("span %q, want %q", got, testCase.wantSpan)
			}
		})
	}
}

// The rendered message, asserted by equality against a literal typed here.
//
// The description interpolates the name three times, and an id assertion cannot see anything a
// format string does. Comparing against the rule's own constant would be equality that moves on
// both sides under mutation, so the expected text is written out rather than referenced.
func TestNoNewWrappersRendersTheName(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedFiles(t, NoNewWrappers,
		map[string]string{newWrappersFile: "var a = new Boolean(false);"}, newWrappersFile)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	const want = "This calls `new Boolean()`. Constructed, Boolean returns an object rather than " +
		"a primitive, so the result is always truthy, compares equal to nothing by `===`, and " +
		"reports its `typeof` as \"object\". Drop the `new` and call Boolean directly."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("description %q, want %q", got, want)
	}
	if got := result.Diagnostics[0].Message.Id; got != "noNewWrappers" {
		t.Fatalf("id %q, want %q", got, "noNewWrappers")
	}
}

// Cases upstream's corpus does not write, each measured against the installed rule before it was
// written here.
//
// Upstream writes no parenthesized form at all, and the parenthesis is where our parser and ESTree
// actually differ, so every one of these is invisible to the imported corpus. The probe drove
// `eslint@9` through the Linter API on each input; the verdicts recorded beside them are that
// measurement rather than a reading of the source.
func TestNoNewWrappersOnFormsUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantReport bool
	}{
		{"a parenthesized callee reports, because ESTree has no node for the parenthesis", "var a = new (String)('x');", true},
		{"and so does a doubly parenthesized one", "var a = new ((String))('x');", true},
		{"a new expression with no argument list still reports", "var a = new String;", true},
		{"a property access callee is silent, since upstream reads callee.name and finds none", "var a = new window.String('x');", false},
		{"globalThis.String is a property access too, and equally silent", "var a = new globalThis.String('x');", false},
		{"parentheses around the whole expression do not change the callee", "var a = (new String('x'));", true},
		{"a class shadowing the name declines", "class String {} new String();", false},
		{"a function declaration shadowing the name declines", "function Boolean() {} new Boolean();", false},
		{"Object is not one of the three", "var a = new Object();", false},
		{"Symbol belongs to no-new-native-nonconstructor, not here", "var a = new Symbol();", false},
		{"the name is case sensitive", "var a = new string('x');", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, NoNewWrappers,
				map[string]string{newWrappersFile: testCase.sourceText}, newWrappersFile)
			if testCase.wantReport {
				rule_testing.ExpectFindings(t, result, "noNewWrappers")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// One input carrying two findings, which pins that the listener runs per node rather than per file.
//
// `ExpectFindings` asserts the count as well as the ids, so a rule reporting once on a file with
// two violations fails here and passes every single-finding fixture above.
func TestNoNewWrappersReportsEachViolation(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.RunTypedFiles(t, NoNewWrappers,
			map[string]string{newWrappersFile: "var a = new Number(1); var b = new Boolean(0);"},
			newWrappersFile),
		"noNewWrappers", "noNewWrappers")
}

// The typed harness is required, asserted so a later revert to `Run` fails loudly.
//
// Under the syntax-only harness the checker is nil, `resolvesToAGlobal` takes its nil guard, and
// the rule goes completely silent. That is a vacuous green for every StaysSilent case in this file,
// which is the more dangerous failure of the two the brief names, so it is pinned rather than left
// to be rediscovered.
func TestNoNewWrappersIsSilentWithoutAChecker(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoNewWrappers, newWrappersFile, "var a = new String('hello');"))
}
