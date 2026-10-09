package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus ran every case as `.tsx`; the snapshot header names
// `no_unnecessary_type_constraint.tsx`. Keeping that extension matters rather than being
// cosmetic, because the extension is an input to the repair: the same source loses its
// constraint as `<T>` in a `.ts` file and as `<T,>` in a `.tsx` one.
const unnecessaryTypeConstraintFile = "/repository/source/no_unnecessary_type_constraint.tsx"

// TestNoUnnecessaryTypeConstraintFires is upstream's twenty failing inputs, verbatim.
//
// Nineteen report once and one reports twice. The extractor flags the mismatch as a
// DISCREPANCY (21 diagnostics against 20 inputs) and it is real rather than the
// multiple-snapshot artifact: this rule has one snapshot file with 21 location headers, and
// aligning them against the inputs in order shows `<T extends any, U extends any>` printed
// twice, once for each parameter. That is the whole of the discrepancy.
func TestNoUnnecessaryTypeConstraintFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		count      int
	}{
		{"function data<T extends any>() {}", 1},
		{"function data<T extends any, U>() {}", 1},
		{"function data<T, U extends any>() {}", 1},
		{"function data<T extends any, U extends T>() {}", 1},
		{"const data = <T extends any>() => {};", 1},
		{"const data = <T extends any,>() => {};", 1},
		{"const data = <T extends any, >() => {};", 1},
		{"const data = <T extends any ,>() => {};", 1},
		{"const data = <T extends any , >() => {};", 1},
		{"const data = <T extends any = unknown>() => {};", 1},
		{"const data = <T extends any, U extends any>() => {};", 2},
		{"function data<T extends unknown>() {}", 1},
		{"const data = <T extends any>() => {};", 1},
		{"const data = <T extends unknown>() => {};", 1},
		{"class Data<T extends unknown> {}", 1},
		{"const Data = class<T extends unknown> {};", 1},
		{"class Data { member<T extends unknown>() {} }", 1},
		{"const Data = class { member<T extends unknown>() {} };", 1},
		{"interface Data<T extends unknown> {}", 1},
		{"type Data<T extends unknown> = {};", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryTypeConstraint, unnecessaryTypeConstraintFile, testCase.sourceText)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "noUnnecessaryTypeConstraint"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoUnnecessaryTypeConstraintStaysSilent is upstream's twelve passing inputs, verbatim,
// plus the constraint kinds neighbouring `any` that a reader expects to behave the same way.
func TestNoUnnecessaryTypeConstraintStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"function data() {}",
		"function data<T>() {}",
		"function data<T, U>() {}",
		"function data<T extends number>() {}",
		"function data<T extends number | string>() {}",
		"function data<T extends any | number>() {}",
		"type X = any; function data<T extends X>() {}",
		"const data = () => {};",
		"const data = <T, >() => {};",
		"const data = <T, U>() => {};",
		"const data = <T extends number>() => {};",
		"const data = <T extends number | string>() => {};",

		// Beyond upstream's corpus. Each of these is a constraint a reader expects to be as empty as
		// `unknown`, and each was measured silent against the release binary before being written here
		// rather than reasoned about. Without them the rule's two-kind switch could be widened to any
		// keyword-shaped constraint and every imported fixture would stay green.
		"function data<T extends {}>() {}",
		"function data<T extends object>() {}",
		"function data<T extends never>() {}",
		"function data<T extends void>() {}",

		// The parenthesis measurement, and the reason it is a fixture rather than a comment.
		// `(any)` parses to a parenthesized type on both sides, so upstream's direct destructure
		// declines it and so does ours. Adding a parenthesis skip would read as a free correctness
		// improvement and would ship a divergence no upstream case can see. Pinned silent against the
		// release binary.
		"function data<T extends (any)>() {}",
		"function data<T extends (unknown)>() {}",
	}
	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnnecessaryTypeConstraint, unnecessaryTypeConstraintFile, sourceText))
		})
	}
}

// TestNoUnnecessaryTypeConstraintSuggestionsWriteWhatTheyClaim is upstream's own twenty-four fix
// vectors, applied and compared as source text.
//
// This test is not optional and it is not covered by the two above. The rule proposes a repair, and
// a repair is the one part of a rule that rewrites a file; an assertion on the message id is
// satisfied by a correct detection carrying a suggestion that deletes the wrong range. Widening the
// replaced span here leaves every id and every span assertion in this file green, which was
// confirmed by mutation rather than assumed.
//
// The vectors are also the artifact the extractor cannot see. It recognizes the tuple form of
// upstream's fix table and prints nothing for the `ExpectFixTestCase` form this rule uses, so its
// summary reports no fix count at all and the rule reads as repair-free. They were pulled out of the
// Rust source separately.
//
// # These are suggestions rather than fixes, which is a decision and not a formality
//
// oxc calls `ctx.diagnostic_with_suggestion` and declares `suggestion` in its metadata, so the
// engine does not apply this unattended. Confirmed against the release binary rather than read off
// the source: running `oxlint --fix` over every shape below changes not one byte, and
// `--fix-suggestions` rewrites all of them.
//
// The distinction is earned. Removing `extends any` cannot change what a type parameter means, so
// it reads like a fix, but the repair has to add a comma in some files and not others precisely
// because deleting the constraint can change how the file *parses*. A repair whose correctness
// depends on the filename is one a human should look at.
//
// `rule_testing.ExpectFixedSource` applies fixes and this rule proposes none, so the repair is applied
// here by hand. The application is deliberately trivial: one suggestion carrying one fix, sliced
// into the source. Anything cleverer would be testing the applier rather than the rule.
func TestNoUnnecessaryTypeConstraintSuggestionsWriteWhatTheyClaim(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fileName   string
		sourceText string
		wantSource string
	}{
		{unnecessaryTypeConstraintFile, "function data<T extends any>() {}", "function data<T>() {}"},
		{unnecessaryTypeConstraintFile, "function data<T extends any, U>() {}", "function data<T, U>() {}"},
		{unnecessaryTypeConstraintFile, "function data<T, U extends any>() {}", "function data<T, U>() {}"},
		{unnecessaryTypeConstraintFile, "function data<T extends any, U extends T>() {}", "function data<T, U extends T>() {}"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any>() => {};", "const data = <T,>() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any,>() => {};", "const data = <T,>() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any, >() => {};", "const data = <T, >() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any ,>() => {};", "const data = <T ,>() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any , >() => {};", "const data = <T , >() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any = unknown>() => {};", "const data = <T = unknown>() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any, U extends any>() => {};", "const data = <T, U>() => {};"},
		{unnecessaryTypeConstraintFile, "function data<T extends unknown>() {}", "function data<T>() {}"},
		{"/repository/source/no_unnecessary_type_constraint.ts", "const data = <T extends any>() => {};", "const data = <T>() => {};"},
		{"/repository/source/no_unnecessary_type_constraint.mts", "const data = <T extends any>() => {};", "const data = <T,>() => {};"},
		{"/repository/source/no_unnecessary_type_constraint.cts", "const data = <T extends any>() => {};", "const data = <T,>() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any /* comment */,>() => {};", "const data = <T /* comment */,>() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends any /* comment */>() => {};", "const data = <T, /* comment */>() => {};"},
		{unnecessaryTypeConstraintFile, "const data = <T extends unknown>() => {};", "const data = <T,>() => {};"},
		{unnecessaryTypeConstraintFile, "class Data<T extends unknown> {}", "class Data<T> {}"},
		{unnecessaryTypeConstraintFile, "const Data = class<T extends unknown> {};", "const Data = class<T> {};"},
		{unnecessaryTypeConstraintFile, "class Data { member<T extends unknown>() {} }", "class Data { member<T>() {} }"},
		{unnecessaryTypeConstraintFile, "const Data = class { member<T extends unknown>() {} };", "const Data = class { member<T>() {} };"},
		{unnecessaryTypeConstraintFile, "interface Data<T extends unknown> {}", "interface Data<T> {}"},
		{unnecessaryTypeConstraintFile, "type Data<T extends unknown> = {};", "type Data<T> = {};"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText+" in "+testCase.fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryTypeConstraint, testCase.fileName, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("wanted at least one diagnostic to carry a suggestion, got none")
			}

			// Applied back to front so an earlier rewrite cannot move the offsets a later one was
			// computed against. Two of these inputs carry two suggestions.
			source := testCase.sourceText
			for index := len(result.Diagnostics) - 1; index >= 0; index-- {
				suggestions := result.Diagnostics[index].Suggestions
				if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 {
					t.Fatalf("wanted one suggestion carrying one fix, got %d suggestions", len(suggestions))
				}
				fix := suggestions[0].Fixes[0]
				source = source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]
			}

			if source != testCase.wantSource {
				t.Errorf("the repair wrote\n  %q\nwant\n  %q", source, testCase.wantSource)
			}
		})
	}
}

// TestNoUnnecessaryTypeConstraintPointsAtTheParameterName asserts where the finding lands and what
// it says, neither of which any assertion above can see.
//
// The span matters here because there are two defensible answers and upstream carries both: oxc
// attaches a primary label to the parameter's name and a secondary one to the constraint type. We
// report one range per finding, so choosing the name is a decision, and a fixture is the only place
// that decision is recorded. A rule pointing at the constraint instead would leave every message id
// assertion in this file green.
//
// The message is compared against a literal typed here rather than against
// `messageNoUnnecessaryTypeConstraint`, deliberately. Comparing a diagnostic to the constant the
// rule reports with is equality that looks correct and cannot fail: both sides move together under
// any edit to the constant, so the assertion tracks the rule instead of guarding it.
func TestNoUnnecessaryTypeConstraintPointsAtTheParameterName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantTexts  []string
	}{
		{"function data<T extends any>() {}", []string{"T"}},
		{"function data<T, U extends any>() {}", []string{"U"}},
		{"const data = <Element extends unknown>() => {};", []string{"Element"}},
		// Two findings in one list, in source order, each naming its own parameter.
		{"const data = <T extends any, U extends any>() => {};", []string{"T", "U"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryTypeConstraint, unnecessaryTypeConstraintFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("wanted %d diagnostics, got %d", len(testCase.wantTexts), len(result.Diagnostics))
			}
			for index, wantText := range testCase.wantTexts {
				diagnostic := result.Diagnostics[index]
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != wantText {
					t.Errorf("finding %d points at %q, want %q", index, reported, wantText)
				}
				if diagnostic.Message.Id != "noUnnecessaryTypeConstraint" {
					t.Errorf("finding %d has id %q", index, diagnostic.Message.Id)
				}
				if len(diagnostic.Suggestions) != 1 ||
					diagnostic.Suggestions[0].Message.Id != "removeTheConstraint" {
					t.Errorf("finding %d does not carry the removal suggestion", index)
				}
				// A suggestion is not a fix, and the difference is the whole reason this rule does
				// not rewrite files unattended. If the repair ever moves into Fixes, the engine
				// starts applying a comma-dependent rewrite without anyone agreeing to it.
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d carries %d applicable fixes; the repair must stay a suggestion",
						index, len(diagnostic.Fixes))
				}
			}
		})
	}
}
