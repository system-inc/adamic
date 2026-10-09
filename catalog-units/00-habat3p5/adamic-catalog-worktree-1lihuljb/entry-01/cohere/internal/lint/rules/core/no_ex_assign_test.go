package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const exAssignFile = "/repository/source/Thing.ts"

func TestNoExAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a plain assignment", "export function run() {\n    try { work(); } catch(error) { error = new Error('x'); }\n}\ndeclare function work(): void;\n"},
		// The compound forms write to the same binding and are the ones that read as harmless.
		// `??=` in particular looks like enrichment rather than replacement.
		{"a nullish assignment", "export function run() {\n    try { work(); } catch(error) { error ??= new Error('x'); }\n}\ndeclare function work(): void;\n"},
		{"an or assignment", "export function run() {\n    try { work(); } catch(error) { error ||= new Error('x'); }\n}\ndeclare function work(): void;\n"},
		{"assignment inside a nested block", "export function run(flag: boolean) {\n    try { work(); } catch(error) { if(flag) { error = new Error('x'); } }\n}\ndeclare function work(): void;\n"},
		{"a parenthesized target", "export function run() {\n    try { work(); } catch(error) { (error) = new Error('x'); }\n}\ndeclare function work(): void;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoExAssign, exAssignFile, testCase.sourceText),
				"unexpectedExceptionAssignment")
		})
	}
}

func TestNoExAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"reading the exception", "export function run() {\n    try { work(); } catch(error) { report(error); }\n}\ndeclare function work(): void;\ndeclare function report(value: unknown): void;\n"},
		// The intended repair: a second binding rather than an overwrite.
		{"a new binding beside it", "export function run() {\n    try { work(); } catch(error) { const wrapped = String(error); report(wrapped); }\n}\ndeclare function work(): void;\ndeclare function report(value: unknown): void;\n"},
		// The decision boundary is "does this write to the binding that holds the exception".
		// A same-named binding in a different scope does not.
		{"assigning a different name", "export function run() {\n    let other = 0;\n    try { work(); } catch(error) { other = 1; void error; }\n}\ndeclare function work(): void;\n"},
		// A destructuring catch binds new names taken out of the error rather than the error
		// itself, so writing to one of them loses nothing.
		{"a destructured catch parameter", "export function run() {\n    try { work(); } catch({ message }) { message = 'x'; }\n}\ndeclare function work(): void;\n"},
		{"a catch with no parameter", "export function run() {\n    try { work(); } catch { report(1); }\n}\ndeclare function work(): void;\ndeclare function report(value: unknown): void;\n"},
		{"assigning to a property of the error", "export function run() {\n    try { work(); } catch(error) { (error as { code?: number }).code = 1; }\n}\ndeclare function work(): void;\n"},
		{"no try at all", "export function run() {\n    return 1;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoExAssign, exAssignFile, testCase.sourceText))
		})
	}
}

// An update expression is a write, and neither corpus tests one.
//
// `e++` left the binding rebound and reported nothing, while `e = 1` and `e += 1` both reported.
// Upstream asks `reference.is_write()`, which counts an update, so this was a divergence rather than
// a shared limitation, and its own corpus never exercises the shape: every case on both sides uses
// `=` or a compound operator.
//
// Found while measuring whether the four reference-blocked rules could be answered syntactically,
// not by reviewing this rule. The walker it shares with them was missing a whole category of write.
func TestNoExAssignSeesUpdateExpressions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"postfix increment", "try {} catch (error) { error++; }"},
		{"prefix increment", "try {} catch (error) { ++error; }"},
		{"postfix decrement", "try {} catch (error) { error--; }"},
		{"prefix decrement", "try {} catch (error) { --error; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoExAssign, exAssignFile, testCase.sourceText),
				"unexpectedExceptionAssignment")
		})
	}
}

// A unary that reads its operand is not a write, which is what keeps the arm above from being a
// blanket rule about unary expressions.
func TestNoExAssignIgnoresReadingUnaryOperators(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"try {} catch (error) { const a = -error; }",
		"try {} catch (error) { const a = !error; }",
		"try {} catch (error) { const a = typeof error; }",
	} {
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoExAssign, exAssignFile, sourceText))
	}
}
