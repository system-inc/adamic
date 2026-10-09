package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const dupeElseIfFile = "/repository/source/Branch.ts"

func TestNoDupeElseIfFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the same identifier twice", "declare const a: boolean;\nexport function run() { if(a) {} else if(a) {} }\n"},
		{"a duplicate later in the chain", "declare const a: boolean, b: boolean;\nexport function run() { if(a) {} else if(b) {} else if(a) {} }\n"},
		{"the same call expression", "declare function f(): boolean;\nexport function run() { if(f()) {} else if(f()) {} }\n"},
		{"the same member access", "declare const o: { a: boolean };\nexport function run() { if(o.a) {} else if(o.a) {} }\n"},

		// Subsumption rather than equality, which is most of the rule: everything the second
		// condition matches, the first already matched.
		{"an or covering a later single", "declare const a: boolean, b: boolean;\nexport function run() { if(a || b) {} else if(a) {} }\n"},
		{"an or covering a later or", "declare const a: boolean, b: boolean, c: boolean;\nexport function run() { if(a || b || c) {} else if(b || c) {} }\n"},
		{"a single covering a later and", "declare const a: boolean, b: boolean;\nexport function run() { if(a) {} else if(a && b) {} }\n"},

		// Commutativity, which is sound only because these are read for truthiness.
		//
		// The first two cases below do NOT exercise the commutative branch in conditionsAreEqual,
		// which is worth stating because I thought they did. Splitting `a && b` and `b && a` gives
		// the operand sets {a,b} and {b,a}, and the subset test is order-independent, so they match
		// without any commutative comparison. Found by dropping the commutative branch and watching
		// both stay green.
		//
		// The branch is reached only when a logical expression is compared as a whole, which
		// happens when it is nested inside another operator so the split does not take it apart.
		// The parenthesized cases are the ones that measure it.
		{"and operands reordered", "declare const a: boolean, b: boolean;\nexport function run() { if(a && b) {} else if(b && a) {} }\n"},
		{"or operands reordered", "declare const a: boolean, b: boolean;\nexport function run() { if(a || b) {} else if(b || a) {} }\n"},
		{
			"a nested or with its operands reordered",
			"declare const a: boolean, b: boolean, c: boolean;\n" +
				"export function run() { if((a || b) && c) {} else if((b || a) && c) {} }\n",
		},
		{
			"a nested and with its operands reordered",
			"declare const a: boolean, b: boolean, c: boolean;\n" +
				"export function run() { if((a && b) || c) {} else if((b && a) || c) {} }\n",
		},

		// Flattening. `a || b || c` is a left-nested tree, so a rule reading only the top node
		// sees two operands rather than three.
		{"a middle operand of a flattened or", "declare const a: boolean, b: boolean, c: boolean;\nexport function run() { if(a || b || c) {} else if(b) {} }\n"},

		// Two earlier branches together covering a later one, which is why the walk accumulates
		// rather than comparing against the immediately preceding branch only.
		{"two earlier branches covering one later", "declare const a: boolean, b: boolean;\nexport function run() { if(a) {} else if(b) {} else if(a || b) {} }\n"},

		// A later `&&` extending an earlier one. Fires, but it does NOT measure the conjunct
		// expansion: `a && b && c` splits into the single operand set {a,b,c}, which the earlier
		// {a,b} is already a subset of. Kept as a case in its own right.
		{
			"a later and extending an earlier one",
			"declare const a: boolean, b: boolean, c: boolean;\n" +
				"export function run() { if(a && b) {} else if(a && b && c) {} }\n",
		},
		// These two do measure it, and finding them took three attempts. The expansion matters when
		// a conjunct is itself an OR, because the conjunct then has its own OR-groups that earlier
		// branches can cover separately. `(a || b) && c` as a whole is one group {(a||b), c} that
		// nothing covers, while the conjunct `a || b` is two groups that `if(a)` and `if(b)`
		// between them eliminate.
		//
		// The first attempt (`a && b` then `a && b && c`) fires without the expansion, and the
		// second (`if(a)` then `(a || b) && c`) does not fire at all, since one earlier branch
		// covers only one of the two groups. Both were written believing they measured this.
		{
			"two earlier branches covering the or inside a later and",
			"declare const a: boolean, b: boolean, c: boolean;\n" +
				"export function run() { if(a) {} else if(b) {} else if((a || b) && c) {} }\n",
		},
		{
			"an earlier or covering the or inside a later and",
			"declare const a: boolean, b: boolean, c: boolean;\n" +
				"export function run() { if(a || b) {} else if((a || b) && c) {} }\n",
		},
		{"whitespace differences are ignored", "declare const a: boolean, b: boolean;\nexport function run() { if(a &&  b) {} else if(a&&b) {} }\n"},
		{"parentheses are ignored", "declare const a: boolean;\nexport function run() { if(a) {} else if((a)) {} }\n"},
		{"a comment between does not matter", "declare const a: boolean;\nexport function run() { if(a) {} else if(/* same */ a) {} }\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoDupeElseIf, dupeElseIfFile, testCase.sourceText),
				"unexpected")
		})
	}
}

func TestNoDupeElseIfStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"different conditions", "declare const a: boolean, b: boolean;\nexport function run() { if(a) {} else if(b) {} }\n"},
		{"a longer chain of distinct conditions", "declare const a: boolean, b: boolean, c: boolean;\nexport function run() { if(a) {} else if(b) {} else if(c) {} }\n"},

		// The direction that must not be reversed. An earlier `a && b` does not cover a later `a`:
		// the later branch is reachable whenever a holds and b does not.
		{"an and followed by one of its operands", "declare const a: boolean, b: boolean;\nexport function run() { if(a && b) {} else if(a) {} }\n"},
		{"a single followed by an unrelated or", "declare const a: boolean, b: boolean;\nexport function run() { if(a) {} else if(a || b) {} }\n"},

		// Not the same expression despite looking close. These are what the token oracle is for.
		{"optional chaining is a different expression", "declare const o: { a?: boolean } | undefined;\nexport function run() { if(o!.a) {} else if(o?.a) {} }\n"},
		{"different literals", "declare const a: number;\nexport function run() { if(a === 1) {} else if(a === 2) {} }\n"},
		{"different negations", "declare const a: number;\nexport function run() { if(a === -1) {} else if(a === +1) {} }\n"},
		{"different members of the same object", "declare const o: { a: boolean; b: boolean };\nexport function run() { if(o.a) {} else if(o.b) {} }\n"},

		// A nested if inside a consequent is not part of this chain, which is what the alternate
		// check enforces. Without it the inner condition would be compared to the outer one.
		//
		// The braced cases do NOT measure that check, which is worth stating because I assumed they
		// did. A braced consequent makes the inner if's parent a Block, so the walk stops at the
		// kind test one line earlier and never reaches the alternate test. Found by dropping the
		// alternate check and watching both stay green.
		//
		// The brace-less case is the one that measures it: there the inner if's parent really is
		// the outer IfStatement, and only the alternate test distinguishes "this is the else branch"
		// from "this is the then branch".
		{"a repeated condition inside the consequent", "declare const a: boolean;\nexport function run() { if(a) { if(a) {} } }\n"},
		{"a repeated condition in a nested else", "declare const a: boolean, b: boolean;\nexport function run() { if(a) { if(b) {} else if(a) {} } }\n"},
		{
			"a repeated condition in a brace-less consequent",
			"declare const a: boolean;\nexport function run() { if(a) if(a) {} }\n",
		},
		{
			"a covered condition in a brace-less consequent",
			"declare const a: boolean, b: boolean;\nexport function run() { if(a || b) if(a) {} }\n",
		},

		{"a lone if", "declare const a: boolean;\nexport function run() { if(a) {} }\n"},
		{"an if with a plain else", "declare const a: boolean;\nexport function run() { if(a) {} else {} }\n"},
		{"no if at all", "export const value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDupeElseIf, dupeElseIfFile, testCase.sourceText))
		})
	}
}

// One report per unreachable branch, not one per covering condition.
//
// A branch covered by three earlier ones is still one dead branch and one edit. Worth pinning
// because the walk continues up the chain after a match unless it returns, and a missing return
// would produce a finding per ancestor.
func TestNoDupeElseIfReportsOncePerBranch(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoDupeElseIf, dupeElseIfFile,
		"declare const a: boolean;\nexport function run() { if(a) {} else if(a) {} else if(a) {} }\n"),
		"unexpected", "unexpected")
}
