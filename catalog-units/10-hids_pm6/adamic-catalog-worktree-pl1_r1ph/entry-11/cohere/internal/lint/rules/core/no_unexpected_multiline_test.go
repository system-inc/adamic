package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, extracted by loading its test file with a stubbed RuleTester and
// serialising the captured cases, so no case here was retyped and no escape was cooked on the way.
// 39 valid and 17 invalid, each invalid case carrying exactly one finding.
//
// Four of these cases upstream could only express with custom TypeScript fixture parsers, because
// its default parser cannot read generic type arguments on a tag. Ours parses them natively, so
// they are ordinary rows here and are marked at the line.

func TestNoUnexpectedMultilineStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "valid0", source: "(x || y).aFunction()"},
		{name: "valid1", source: "[a, b, c].forEach(doSomething)"},
		{name: "valid2", source: "var a = b;\n(x || y).doSomething()"},
		{name: "valid3", source: "var a = b\n;(x || y).doSomething()"},
		{name: "valid4", source: "var a = b\nvoid (x || y).doSomething()"},
		{name: "valid5", source: "var a = b;\n[1, 2, 3].forEach(console.log)"},
		{name: "valid6", source: "var a = b\nvoid [1, 2, 3].forEach(console.log)"},
		{name: "valid7", source: "\"abc\\\n(123)\""},
		{name: "valid8", source: "var a = (\n(123)\n)"},
		{name: "valid9", source: "f(\n(x)\n)"},
		{name: "valid10", source: "(\nfunction () {}\n)[1]"},
		{name: "valid11", source: "let x = function() {};\n   `hello`"},
		{name: "valid12", source: "let x = function() {}\nx `hello`"},
		{name: "valid13", source: "String.raw `Hi\n${2+3}!`;"},
		{name: "valid14", source: "x\n.y\nz `Valid Test Case`"},
		{name: "valid15", source: "f(x\n)`Valid Test Case`"},
		{name: "valid16", source: "x.\ny `Valid Test Case`"},
		{name: "valid17", source: "(x\n)`Valid Test Case`"},
		{name: "valid18", source: "\n            foo\n            / bar /2\n        "},
		{name: "valid19", source: "\n            foo\n            / bar / mgy\n        "},
		{name: "valid20", source: "\n            foo\n            / bar /\n            gym\n        "},
		{name: "valid21", source: "\n            foo\n            / bar\n            / ygm\n        "},
		{name: "valid22", source: "\n            foo\n            / bar /GYM\n        "},
		{name: "valid23", source: "\n            foo\n            / bar / baz\n        "},
		{name: "valid24", source: "foo /bar/g"},
		{name: "valid25", source: "\n            foo\n            /denominator/\n            2\n        "},
		{name: "valid26", source: "\n            foo\n            / /abc/\n        "},
		{name: "valid27", source: "\n            5 / (5\n            / 5)\n        "},
		// Upstream needed a custom TypeScript parser for this case; ours parses it natively.
		{name: "valid28", source: "\n                tag<generic>`\n                    multiline\n                `;\n            "},
		// Upstream needed a custom TypeScript parser for this case; ours parses it natively.
		{name: "valid29", source: "\n                tag<\n                  generic\n                >`\n                    multiline\n                `;\n            "},
		// Upstream needed a custom TypeScript parser for this case; ours parses it natively.
		{name: "valid30", source: "\n                tag<\n                  generic\n                >`multiline`;\n            "},
		{name: "valid31", source: "var a = b\n  ?.(x || y).doSomething()"},
		{name: "valid32", source: "var a = b\n  ?.[a, b, c].forEach(doSomething)"},
		{name: "valid33", source: "var a = b?.\n  (x || y).doSomething()"},
		{name: "valid34", source: "var a = b?.\n  [a, b, c].forEach(doSomething)"},
		{name: "valid35", source: "class C { field1\n[field2]; }"},
		{name: "valid36", source: "class C { field1\n*gen() {} }"},
		{name: "valid37", source: "class C { field1 = () => {}\n[field2]; }"},
		{name: "valid38", source: "class C { field1 = () => {}\n*gen() {} }"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnexpectedMultiline, "file.ts", testCase.source))
		})
	}
}

func TestNoUnexpectedMultilineFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		wantIds []string
	}{
		{name: "invalid0", source: "var a = b\n(x || y).doSomething()", wantIds: []string{"function"}},
		{name: "invalid1", source: "var a = (a || b)\n(x || y).doSomething()", wantIds: []string{"function"}},
		{name: "invalid2", source: "var a = (a || b)\n(x).doSomething()", wantIds: []string{"function"}},
		{name: "invalid3", source: "var a = b\n[a, b, c].forEach(doSomething)", wantIds: []string{"property"}},
		{name: "invalid4", source: "var a = b\n    (x || y).doSomething()", wantIds: []string{"function"}},
		{name: "invalid5", source: "var a = b\n  [a, b, c].forEach(doSomething)", wantIds: []string{"property"}},
		{name: "invalid6", source: "let x = function() {}\n `hello`", wantIds: []string{"taggedTemplate"}},
		{name: "invalid7", source: "let x = function() {}\nx\n`hello`", wantIds: []string{"taggedTemplate"}},
		{name: "invalid8", source: "x\n.y\nz\n`Invalid Test Case`", wantIds: []string{"taggedTemplate"}},
		{name: "invalid9", source: "\n                foo\n                / bar /gym\n            ", wantIds: []string{"division"}},
		{name: "invalid10", source: "\n                foo\n                / bar /g\n            ", wantIds: []string{"division"}},
		{name: "invalid11", source: "\n                foo\n                / bar /g.test(baz)\n            ", wantIds: []string{"division"}},
		{name: "invalid12", source: "\n                foo\n                /bar/gimuygimuygimuy.test(baz)\n            ", wantIds: []string{"division"}},
		{name: "invalid13", source: "\n                foo\n                /bar/s.test(baz)\n            ", wantIds: []string{"division"}},
		// Upstream needed a custom TypeScript parser for this case; ours parses it natively.
		{name: "invalid14", source: "const x = aaaa<\n  test\n>/*\ntest\n*/`foo`", wantIds: []string{"taggedTemplate"}},
		{name: "invalid15", source: "class C { field1 = obj\n[field2]; }", wantIds: []string{"property"}},
		{name: "invalid16", source: "class C { field1 = function() {}\n[field2]; }", wantIds: []string{"property"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnexpectedMultiline, "file.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// Spans and message text, which ExpectFindings cannot see.
//
// Every offset here was computed from upstream's own line and column assertions rather than read
// off this rule, so a finding anchored on the node instead of the token fails here while passing
// every case above. The expected text is one character in all seventeen cases: the `(`, `[`,
// backtick or `/` that continues the previous line.
//
// rule_testing.Run does not trim its input the way RunTyped does, so these offsets index the
// fixture literal directly.
func TestNoUnexpectedMultilineSpansAndMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		wantPos  int
		wantEnd  int
		wantText string
		wantId   string
	}{
		{name: "invalid0", source: "var a = b\n(x || y).doSomething()", wantPos: 10, wantEnd: 11, wantText: "(", wantId: "function"},
		{name: "invalid1", source: "var a = (a || b)\n(x || y).doSomething()", wantPos: 17, wantEnd: 18, wantText: "(", wantId: "function"},
		{name: "invalid2", source: "var a = (a || b)\n(x).doSomething()", wantPos: 17, wantEnd: 18, wantText: "(", wantId: "function"},
		{name: "invalid3", source: "var a = b\n[a, b, c].forEach(doSomething)", wantPos: 10, wantEnd: 11, wantText: "[", wantId: "property"},
		{name: "invalid4", source: "var a = b\n    (x || y).doSomething()", wantPos: 14, wantEnd: 15, wantText: "(", wantId: "function"},
		{name: "invalid5", source: "var a = b\n  [a, b, c].forEach(doSomething)", wantPos: 12, wantEnd: 13, wantText: "[", wantId: "property"},
		{name: "invalid6", source: "let x = function() {}\n `hello`", wantPos: 23, wantEnd: 24, wantText: "`", wantId: "taggedTemplate"},
		{name: "invalid7", source: "let x = function() {}\nx\n`hello`", wantPos: 24, wantEnd: 25, wantText: "`", wantId: "taggedTemplate"},
		{name: "invalid8", source: "x\n.y\nz\n`Invalid Test Case`", wantPos: 7, wantEnd: 8, wantText: "`", wantId: "taggedTemplate"},
		{name: "invalid9", source: "\n                foo\n                / bar /gym\n            ", wantPos: 37, wantEnd: 38, wantText: "/", wantId: "division"},
		{name: "invalid10", source: "\n                foo\n                / bar /g\n            ", wantPos: 37, wantEnd: 38, wantText: "/", wantId: "division"},
		{name: "invalid11", source: "\n                foo\n                / bar /g.test(baz)\n            ", wantPos: 37, wantEnd: 38, wantText: "/", wantId: "division"},
		{name: "invalid12", source: "\n                foo\n                /bar/gimuygimuygimuy.test(baz)\n            ", wantPos: 37, wantEnd: 38, wantText: "/", wantId: "division"},
		{name: "invalid13", source: "\n                foo\n                /bar/s.test(baz)\n            ", wantPos: 37, wantEnd: 38, wantText: "/", wantId: "division"},
		{name: "invalid14", source: "const x = aaaa<\n  test\n>/*\ntest\n*/`foo`", wantPos: 34, wantEnd: 35, wantText: "`", wantId: "taggedTemplate"},
		{name: "invalid15", source: "class C { field1 = obj\n[field2]; }", wantPos: 23, wantEnd: 24, wantText: "[", wantId: "property"},
		{name: "invalid16", source: "class C { field1 = function() {}\n[field2]; }", wantPos: 33, wantEnd: 34, wantText: "[", wantId: "property"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnexpectedMultiline, "file.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly 1 finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			if finding.Range.Pos() != testCase.wantPos || finding.Range.End() != testCase.wantEnd {
				t.Errorf("span: got [%d,%d), want [%d,%d)",
					finding.Range.Pos(), finding.Range.End(), testCase.wantPos, testCase.wantEnd)
			}
			if got := testCase.source[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Errorf("reported text: got %q, want %q", got, testCase.wantText)
			}
			if finding.Message.Id != testCase.wantId {
				t.Errorf("message id: got %q, want %q", finding.Message.Id, testCase.wantId)
			}
		})
	}
}

// The four descriptions are asserted against literal strings typed here rather than against the
// rule's own constants, because comparing a diagnostic to the constant it was built from is an
// equality that moves in both directions under mutation and guards nothing.
func TestNoUnexpectedMultilineDescriptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source  string
		wantId  string
		wantOne string
	}{
		{
			source:  "var a = b\n(x || y).doSomething()",
			wantId:  "function",
			wantOne: "A newline sits between this expression and the `(` that follows it, so what reads as two statements is one call. Automatic semicolon insertion does not fire before an open paren, so the line above is being invoked rather than ended.",
		},
		{
			source:  "var a = b\n[a, b, c].forEach(doSomething)",
			wantId:  "property",
			wantOne: "A newline sits between this expression and the `[` that follows it, so what reads as an array on its own line is a computed property access on the line above. Automatic semicolon insertion does not fire before an open bracket.",
		},
		{
			source:  "let x = function() {}\n `hello`",
			wantId:  "taggedTemplate",
			wantOne: "A newline sits between this expression and the template literal that follows it, so the template is being used as a tag argument rather than standing alone. Automatic semicolon insertion does not fire before a backtick.",
		},
		{
			source:  "foo\n/ bar /gym",
			wantId:  "division",
			wantOne: "A newline sits between the numerator and the `/` that follows it, and what comes after reads as a regular expression with flags rather than a division. The two parse differently and only one of them is what the line above looks like.",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.wantId, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnexpectedMultiline, "file.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly 1 finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.wantOne {
				t.Errorf("description:\n got %q\nwant %q", got, testCase.wantOne)
			}
		})
	}
}

// Cases upstream's corpus does not write, each measured against the installed build at 10.8.1 with
// `Linter#cohere` under `sourceType: "script"` before being written here. They exist because this
// port makes a decision at each of them that no imported fixture can see.
//
// The command, so the next reader can re-take these rather than trust them:
//
//	node -e 'const {Linter}=require("eslint");console.log(JSON.stringify(
//	  new Linter().cohere(CODE,{rules:{"no-unexpected-multiline":"error"},
//	  languageOptions:{ecmaVersion:2022,sourceType:"script"}})))'
func TestNoUnexpectedMultilineMeasuredAgainstTheInstalledBuild(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name   string
		source string
		why    string
	}{
		{"flagsRunFollowedByMoreIdentifier", "foo\n/ bar /gymnasium",
			"the flag characters are a prefix of an ordinary name, so this is a division"},
		{"flagsUppercase", "foo\n/ bar /G", "the flag set is case sensitive"},
		{"flagHasIndices", "foo\n/bar/d",
			"upstream's pattern predates the d flag and calls this clean; reproduced, not corrected"},
		{"flagUnicodeSets", "foo\n/bar/v",
			"upstream's pattern predates the v flag and calls this clean; reproduced, not corrected"},
		{"callWithNoArguments", "var a = b\n()", "nothing on the next line was misread as a statement"},
		{"newExpression", "var a = new b\n(x)", "upstream listens on CallExpression alone"},
		{"optionalCall", "var a = b?.\n(x)", "the continuation is spelled explicitly"},
		{"optionalElementAccess", "var a = b?.\n[x]", "the continuation is spelled explicitly"},
		// Upstream's selector requires the LEFT child to be a `/` binary specifically, and
		// multiplicative operators share a precedence level, so `foo * bar / g` puts a `*` binary
		// there. Without the operator test the arm reports this, and no case in upstream's corpus
		// writes a mixed multiplicative chain. Found by a surviving mutant, then measured clean
		// against the installed build with a control that fired.
		{"multiplicationThenDivision", "foo\n* bar /g",
			"the left child of the division is a multiplication, not another division"},
		{"moduloThenDivision", "foo\n% bar /g",
			"same, through the other same-precedence operator"},
		{"callWithTypeArguments", "var a = foo<T>\n(x)",
			"the token after the callee is the `<`, which shares the callee's line; the same answer " +
				"upstream gives under the TypeScript parser, measured"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnexpectedMultiline, "file.ts", testCase.source))
		})
	}

	fires := []struct {
		name    string
		source  string
		wantIds []string
	}{
		{"flagsFollowedByMemberAccess", "foo\n/ bar /g.x", []string{"division"}},
		{"twoNestedCalls", "var a = b\n(c\n(d))", []string{"function", "function"}},
		// Both findings, in WALK order rather than source order: the walk is pre-order, so the
		// outer call fires before the element access nested inside its callee. Upstream reports the
		// same two findings the other way round, which is its own diagnostic sorting rather than a
		// judgment of the rule, and ExpectFindings compares the sequence.
		{"accessThenCall", "var a = b\n[c]\n(d)", []string{"function", "property"}},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnexpectedMultiline, "file.ts", testCase.source), testCase.wantIds...)
		})
	}
}
