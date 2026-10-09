package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// withFile is where the fixtures pretend to live.
const withFile = "/repository/source/With.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_with.rs`: 11 pass,
// 1 fail, and the snapshot records 1 diagnostic from that 1 fail input, so one finding per input is
// measured here rather than assumed. The extractor reported one tester block and no discrepancy.
//
// This is the smallest corpus in the lane, so it protects least, and the cases written from reading
// our own code below carry most of the weight.
func TestNoWithFires(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoWith, withFile, "with(foo) { bar() }"), "noWith")
}

// The clean cases are all the ways the four letters `with` reach a file without being the statement.
//
// Upstream's set is deliberate: a block comment, a line comment, a property name in an object
// literal, an assignment to that property, a class method, a string, a template, a computed string
// subscript, a computed key, and a destructuring rename. `with` is not a reserved word in a property
// position, so each of these parses to something other than a WithStatement, and a rule matching on
// source text rather than on node kind would report every one of them.
func TestNoWithStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an unrelated call", "foo.bar();"},
		{"the keyword inside a block comment", "/* with keyword in block comment */ foo();"},
		{"the keyword inside a line comment", "// with in line comment\nfoo();"},
		{"a property named with", "var obj = { with: 1 }; obj.with;"},
		{"an assignment to a property named with", "obj.with = 1;"},
		{"a class method named with", "class C { with() {} } new C().with();"},
		{"the keyword inside a string", "console.log('with in string');"},
		{"the keyword inside a template", "console.log(`with in template`);"},
		{"a computed string subscript", "const o = {}; o['with'] = 2;"},
		{"a computed key", "const p = { ['with']: 3 }; p.with;"},
		{"a destructuring rename", "const { with: w } = { with: 4 }; w;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoWith, withFile, testCase.sourceText))
		})
	}
}

// The span, which the message-id fixtures above structurally cannot see.
//
// This is the whole risk in this rule. Both upstream implementations deliberately report the `with`
// keyword alone and not the statement: oxc writes `Span::sized(with_statement.span.start, 4)` and
// ESLint writes `loc: sourceCode.getFirstToken(node).loc`. The snapshot agrees, underlining four
// columns. Our obvious spelling, `ctx.ReportNode(node, ...)`, spans the entire statement instead,
// and a probe confirmed it: `with(foo) { bar() }` reported [0,19) rather than [0,4). That port would
// have passed every fixture above while pointing at the wrong text, which is the defect this step
// exists to catch. The rule carries no fix, so nothing else would have surfaced it either.
func TestNoWithReportsTheKeywordAlone(t *testing.T) {
	t.Parallel()

	const sourceText = "with(foo) { bar() }"
	result := rule_testing.Run(t, NoWith, withFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noWith")

	reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "with" {
		t.Fatalf("reported span was %q, wanted the keyword %q", reported, "with")
	}
}

// Leading trivia, which is the way a first-token span goes wrong quietly.
//
// A node's `Pos()` is the position *before* its leading comments and whitespace, so a span built
// from the raw offset would underline the comment rather than the keyword. Upstream never exercises
// this because its single fail case starts at column one, where the correct and incorrect answers
// are the same number. Our code is commented and indented everywhere, so column one is the rare
// case here and this is the common one.
func TestNoWithSkipsLeadingTrivia(t *testing.T) {
	t.Parallel()

	const sourceText = "function f() {\n  /* note */ with (foo) { bar() }\n}"
	result := rule_testing.Run(t, NoWith, withFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noWith")

	reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "with" {
		t.Fatalf("reported span was %q, wanted the keyword %q", reported, "with")
	}
}

// One finding per statement rather than one per file.
//
// Upstream's corpus has a single fail input containing a single `with`, so it cannot distinguish a
// rule that reports each statement from one that reports once and stops, nor from one that reports
// the outer statement and never descends into its body. Nesting answers all three at once, and the
// inner span pins that the second finding is the inner keyword rather than a repeat of the outer.
func TestNoWithReportsEveryStatementIncludingNested(t *testing.T) {
	t.Parallel()

	const sourceText = "with (a) { with (b) { c() } }"
	result := rule_testing.Run(t, NoWith, withFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noWith", "noWith")

	first := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	second := sourceText[result.Diagnostics[1].Range.Pos():result.Diagnostics[1].Range.End()]
	if first != "with" || second != "with" {
		t.Fatalf("reported spans were %q and %q, wanted both to be the keyword", first, second)
	}
	if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
		t.Fatalf("both findings pointed at offset %d, so the inner statement was not visited",
			result.Diagnostics[0].Range.Pos())
	}
}

// The rule stays live under strict mode, which is the question worth answering here.
//
// `with` is a syntax error in strict mode and every ES module is strict, so the reasonable
// expectation is that this rule can never fire on our tree and is dead weight. That expectation is
// wrong, and a probe is what settled it rather than the reasoning. TypeScript's parser is
// error-tolerant: it builds a real WithStatement node and raises the strict-mode violation as a
// grammar diagnostic during checking instead of refusing to parse. So the node reaches a linter
// walking the AST in every context, including an explicit `'use strict'` and a genuine module.
//
// The rule is therefore redundant with the type checker rather than inert, which is what upstream's
// own note means by "not necessary in TypeScript code if `alwaysStrict` is enabled". This pins the
// behavior so that a later parser change making `with` unparseable fails here loudly rather than
// silently turning every fixture above vacuous.
func TestNoWithFiresInsideAModuleAndUnderUseStrict(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a module made strict by an export", "export const x = 1;\nwith (foo) { bar() }"},
		{"an explicit use strict directive", "'use strict';\nwith (foo) { bar() }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoWith, withFile, testCase.sourceText), "noWith")
		})
	}
}
