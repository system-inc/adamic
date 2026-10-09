package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noLonelyIfFile is where the fixtures pretend to live.
const noLonelyIfFile = "/repository/source/NoLonelyIf.ts"

// The corpus is ESLint's own, at `tests/lib/rules/no-lonely-if.js`, copied rather than rewritten:
// all 3 clean cases and all 14 reporting ones, which is the whole upstream file. The rule declares
// `schema: []`, so no case carries options.
//
// This rule is fixable, so each reporting case carries upstream's own `output`, and five of them
// carry `output: null`, which is a decision to reproduce rather than an omission: the case reports
// and the repair is deliberately withheld. Those five are the port's hard half, since a wrong fix
// here parses and the edit engine applies it unattended.
//
// Every case string was built from a list rather than typed, so nothing could cook a newline or an
// escape on the way in, and every verdict and every output was reproduced by driving the installed
// eslint at 10.8.1 first.
func TestNoLonelyIfFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFixed  string
		fixable    bool
	}{
		{"the simplest lonely if", "if (a) {;} else { if (b) {;} }", "if (a) {;} else if (b) {;}", true},
		{"a lonely if across lines", "if (a) {\n  foo();\n} else {\n  if (b) {\n    bar();\n  }\n}", "if (a) {\n  foo();\n} else if (b) {\n    bar();\n  }", true},
		{"a comment between else and the brace, which is outside the rewrite", "if (a) {\n  foo();\n} else /* comment */ {\n  if (b) {\n    bar();\n  }\n}", "if (a) {\n  foo();\n} else /* comment */ if (b) {\n    bar();\n  }", true},
		{"a comment before the inner if, which the rewrite would drop", "if (a) {\n  foo();\n} else {\n  /* otherwise, do the other thing */ if (b) {\n    bar();\n  }\n}", "", false},
		{"a comment inside the inner if, which the rewrite carries along", "if (a) {\n  foo();\n} else {\n  if /* this comment is ok */ (b) {\n    bar();\n  }\n}", "if (a) {\n  foo();\n} else if /* this comment is ok */ (b) {\n    bar();\n  }", true},
		{"a comment after the inner if, which the rewrite would drop", "if (a) {\n  foo();\n} else {\n  if (b) {\n    bar();\n  } /* this comment will prevent this test case from being autofixed. */\n}", "", false},
		{"an unbraced consequent with a semicolon", "if (foo) {} else { if (bar) baz(); }", "if (foo) {} else if (bar) baz();", true},
		{"an unbraced consequent with no semicolon, followed on the same line", "if (foo) {} else { if (bar) baz() } qux();", "", false},
		{"the same with a semicolon, which is fixable", "if (foo) {} else { if (bar) baz(); } qux();", "if (foo) {} else if (bar) baz(); qux();", true},
		{"a following line opening with a bracket", "if (foo) {\n} else {\n  if (bar) baz()\n}\n[1, 2, 3].forEach(foo);", "", false},
		{"a consequent ending in an increment", "if (foo) {\n} else {\n  if (bar) baz++\n}\nfoo;", "", false},
		{"the same with a semicolon, which is fixable", "if (foo) {\n} else {\n  if (bar) baz++;\n}\nfoo;", "if (foo) {\n} else if (bar) baz++;\nfoo;", true},
		{"a following line opening with a template literal", "if (a) {\n  foo();\n} else {\n  if (b) bar()\n}\n`template literal`;", "", false},
		{"a whole else if chain inside the block", "if (a) {\n  foo();\n} else {\n  if (b) {\n    bar();\n  } else if (c) {\n    baz();\n  } else {\n    qux();\n  }\n}", "if (a) {\n  foo();\n} else if (b) {\n    bar();\n  } else if (c) {\n    baz();\n  } else {\n    qux();\n  }", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpectedLonelyIf")

			fixCount := 0
			for _, diagnostic := range result.Diagnostics {
				fixCount += len(diagnostic.Fixes)
			}

			if !testCase.fixable {
				// An `output: null` case. The finding fires and the repair is withheld, and
				// asserting the absence is the only thing that can see a fixer which repairs a
				// case upstream refuses to touch.
				if fixCount != 0 {
					t.Fatalf("upstream declines to fix this case, but the rule proposed %d fix(es)", fixCount)
				}
				return
			}

			if fixCount == 0 {
				t.Fatalf("upstream fixes this case and the rule proposed no repair")
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

func TestNoLonelyIfStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an else if chain already written flat", "if (a) {;} else if (b) {;}"},
		{"an else block holding two statements", "if (a) {;} else { if (b) {;} ; }"},
		{"braces that a dangling else makes necessary", "if (a) if (a) {} else { if (b) {} } else {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoLonelyIfInsertsASpaceWhenElseTouchesTheBrace covers a branch of the fixer that upstream's
// corpus cannot reach, because every case in it writes a space between `else` and `{`.
//
// Upstream's fixer prepends a space when the `else` keyword ends exactly where the opening brace
// begins. Without it, `else{ ... }` rewrites to `elseif (b)`, which is one identifier rather than
// two tokens. The edit engine would refuse that as unparseable, so the visible cost is a fix
// silently lost rather than bad code shipped, but the branch is real and no imported fixture can
// see it: a mutation removing the space survived the entire corpus.
//
// Both spellings were measured against the installed build at 10.8.1, which writes the space in
// each, including the case where the whole construct is written without any spaces at all.
func TestNoLonelyIfInsertsASpaceWhenElseTouchesTheBrace(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		wantFixed  string
	}{
		{"else directly against the brace", "if (a) {;} else{ if (b) {;} }", "if (a) {;} else if (b) {;}"},
		{"no spaces anywhere", "if(a){;}else{if(b){;}}", "if(a){;}else if(b){;}"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpectedLonelyIf")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestNoLonelyIfBracedConsequentIsExemptFromTheSemicolonInsertionTest covers the guard that opens
// upstream's automatic-semicolon-insertion check, and which no imported fixture can reach.
//
// The hazard only exists when the inner `if`'s consequent has no braces, because a braced statement
// cannot run into whatever follows it. Upstream exits the check immediately for a block, and every
// corpus case exercising that check writes an unbraced consequent, so a mutation removing the
// exemption survived the whole suite: it declined fixes upstream performs, and a fix quietly not
// offered is invisible to every assertion about which findings fired.
//
// Each case below has a braced consequent AND a following token that would be a hazard without the
// braces, so they separate the two readings. All three were measured against the installed build at
// 10.8.1, which reports and fixes each one.
func TestNoLonelyIfBracedConsequentIsExemptFromTheSemicolonInsertionTest(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		wantFixed  string
	}{
		{"a braced consequent followed on the same line", "if (foo) {} else { if (bar) { baz() } } qux();", "if (foo) {} else if (bar) { baz() } qux();"},
		{"a braced consequent followed by a line opening with a bracket", "if (foo) {\n} else {\n  if (bar) { baz() }\n}\n[1,2,3].forEach(foo);", "if (foo) {\n} else if (bar) { baz() }\n[1,2,3].forEach(foo);"},
		{"a braced consequent followed by a parenthesis", "if (a) {;} else { if (b) {;} } (x);", "if (a) {;} else if (b) {;} (x);"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpectedLonelyIf")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestNoLonelyIfDanglingElse states the guard that upstream's third clean case only hints at.
//
// `if (a) if (a) {} else { if (b) {} } else {}` is clean, and nothing in "an if alone in an else
// block" explains why. The braces are load-bearing: without them the trailing `else {}` would bind
// to the inner `if (b)` instead of the outer `if`, which is a different program. Upstream calls
// `areBracesNecessary` for this and the recursion inside it is what decides the answer.
//
// One clean case cannot pin a recursive predicate, so all four shapes were measured against the
// installed build at 10.8.1 and are asserted here. The pair that matters is the last two: both end
// in an `else if` chain and they differ only in whether the chain's final link has an `else` of its
// own, which is exactly what the recursion walks to find.
func TestNoLonelyIfDanglingElse(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		wantReport bool
	}{
		{"an alternate-less inner if followed by else", "if (a) if (a) {} else { if (b) {} } else {}", false},
		{"the same with nothing following", "if (a) if (a) {} else { if (b) {} }", true},
		{"an inner if whose chain ends in an else", "if (a) if (a) {} else { if (b) {} else {} } else {}", true},
		{"an inner if whose chain ends without one", "if (a) if (a) {} else { if (b) {} else if (c) {} } else {}", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, testCase.sourceText)
			if testCase.wantReport {
				rule_testing.ExpectFindings(t, result, "unexpectedLonelyIf")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoLonelyIfThenBranchIsNotTheSubject pins the identity test.
//
// A then branch is the same shape as an else branch: a block holding one `if`. Upstream compares
// against the enclosing statement's ALTERNATE specifically, and a kind test in its place would
// report every `if (a) { if (b) {} }` in the tree, which is ordinary code.
func TestNoLonelyIfThenBranchIsNotTheSubject(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"if (a) { if (b) {} }",
		"if (a) { if (b) {} } else {}",
		"while (a) { if (b) {} }",
		"function f() { if (b) {} }",

		// This one is here because a mutation found the others could not see the identity test at
		// all. Rewriting it from "is this block the else branch" to "does an else exist" left every
		// shape above still clean, because the dangling-else guard declined them for its own
		// reason: the inner `if` has no alternate and an `else` follows the block, so the braces
		// read as necessary. The two guards were covering for each other and a single-site mutation
		// could not separate them.
		//
		// Giving the inner `if` an alternate of its own is what separates them: the dangling-else
		// guard now passes and only the identity test declines. Measured clean upstream, and
		// measured reporting under the mutant, which is what makes this a fixture rather than a
		// guess.
		"if (a) { if (b) {} else {} } else {}",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoLonelyIfElseLookaheadIsATokenNotAPrefix guards the one place this port asks a question
// upstream asks of the token stream.
//
// Upstream reads the token after the block and compares it to `else`. Reading the text instead
// makes a bare prefix test wrong, because an identifier beginning with those four letters can
// legally appear exactly there. Neither shape is in upstream's corpus, so both were measured
// against the installed build: the rule reports in both, which is what a token comparison gives and
// what a prefix test would get wrong for the first.
func TestNoLonelyIfElseLookaheadIsATokenNotAPrefix(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
	}{
		{"an identifier starting with else, after the block", "if (a) if (a) {} else { if (b) {} }\nelsewhere;"},
		{"a plain identifier after the block", "if (a) if (a) {} else { if (b) {} }\nother;"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpectedLonelyIf")
		})
	}
}

// TestNoLonelyIfSpansTheInnerIf pins where the finding points.
//
// Upstream reports the inner `if` node itself rather than the else block or the enclosing
// statement. Getting this wrong points the reader at the outer `if`, which is the one line in the
// construct that is not the problem, and no message-id fixture can see the difference.
func TestNoLonelyIfSpansTheInnerIf(t *testing.T) {
	t.Parallel()

	const sourceText = "if (a) {;} else { if (b) {;} }"
	result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	reported := result.Diagnostics[0].Range
	if got := sourceText[reported.Pos():reported.End()]; got != "if (b) {;}" {
		t.Fatalf("expected the finding to cover the inner if, got %q", got)
	}
}

// TestNoLonelyIfMessage asserts the message.
func TestNoLonelyIfMessage(t *testing.T) {
	t.Parallel()

	if messageUnexpectedLonelyIf.Id != "unexpectedLonelyIf" {
		t.Fatalf("expected id %q, got %q", "unexpectedLonelyIf", messageUnexpectedLonelyIf.Id)
	}
	result := rule_testing.Run(t, NoLonelyIf, noLonelyIfFile, "if (a) {;} else { if (b) {;} }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Description != messageUnexpectedLonelyIf.Description {
		t.Fatalf("the reported description is not the rule's own")
	}
}
