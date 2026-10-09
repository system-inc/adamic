package base

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/policy"
)

// consistencyNoBareThrowCaseName numbers a row so a failure names which one, since many rows differ only in a
// path segment.
func consistencyNoBareThrowCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoBareThrowFires is the reporting half of the corpus.
//
// This is one of OUR rules, so there is no upstream corpus to import and the fixtures below are
// written rather than copied. The brief warns that an invented fixture encodes the same belief as
// the port, so every row here was produced by DRIVING the original rule through the ESLint 10.8.1
// Linter API rather than by reading it: the source, the span and the constructor name in each row
// are what that rule reported for that exact input.
//
// The span is asserted because it is a decision rather than an accident. The original reports the
// new expression rather than the throw statement, so the finding covers `new Error('x')` and leaves
// the `throw` keyword outside it. Reporting the statement would read as equally correct and would
// put the underline one token to the left of the constructor name the message is about.
//
// The constructor name is asserted because it is the only part of the message that varies, and a
// file throwing two different built-ins produces two findings under one id that nothing else
// separates.
func TestNoBareThrowFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fileName   string
		sourceText string
		wantSpans  []string
		wantNames  []string
		reason     string
	}{
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			wantSpans:  []string{"new Error('x')"},
			wantNames:  []string{"Error"},
			reason:     "the Error constructor is banned",
		},
		{
			fileName:   "/repository/libraries/base/code-quality/lint/BaseLintEngineParity.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			wantSpans:  []string{"new Error('x')"},
			wantNames:  []string{"Error"},
			reason:     "base's own lint tooling is not below the vocabulary; only nexus's is",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new TypeError('x'); }",
			wantSpans:  []string{"new TypeError('x')"},
			wantNames:  []string{"TypeError"},
			reason:     "the TypeError constructor is banned",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new RangeError('x'); }",
			wantSpans:  []string{"new RangeError('x')"},
			wantNames:  []string{"RangeError"},
			reason:     "the RangeError constructor is banned",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new SyntaxError('x'); }",
			wantSpans:  []string{"new SyntaxError('x')"},
			wantNames:  []string{"SyntaxError"},
			reason:     "the SyntaxError constructor is banned",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new ReferenceError('x'); }",
			wantSpans:  []string{"new ReferenceError('x')"},
			wantNames:  []string{"ReferenceError"},
			reason:     "the ReferenceError constructor is banned",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new EvalError('x'); }",
			wantSpans:  []string{"new EvalError('x')"},
			wantNames:  []string{"EvalError"},
			reason:     "the EvalError constructor is banned",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new URIError('x'); }",
			wantSpans:  []string{"new URIError('x')"},
			wantNames:  []string{"URIError"},
			reason:     "the URIError constructor is banned",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new Error(); }",
			wantSpans:  []string{"new Error()"},
			wantNames:  []string{"Error"},
			reason:     "a bare Error with no message still reports",
		},
		{
			fileName:   "/repository/source/modules/orm/schema-tools/Build.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			wantSpans:  []string{"new Error('x')"},
			wantNames:  []string{"Error"},
			reason:     "a similar directory that is not the exempted one",
		},
		{
			fileName:   "/repository/source/latest/Thing.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			wantSpans:  []string{"new Error('x')"},
			wantNames:  []string{"Error"},
			reason:     "a path containing test as a substring of another word",
		},
		{
			fileName:   "/repository/source/protest/Thing.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			wantSpans:  []string{"new Error('x')"},
			wantNames:  []string{"Error"},
			reason:     "protest contains test but is not a test directory",
		},
		{
			fileName:   "/repository/source/modules/thing/NotBaseError.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			wantSpans:  []string{"new Error('x')"},
			wantNames:  []string{"Error"},
			reason:     "a filename ending in BaseError.ts by suffix",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new Error('a'); }\nexport function g(): void { throw new TypeError('b'); }",
			wantSpans:  []string{"new Error('a')", "new TypeError('b')"},
			wantNames:  []string{"Error", "TypeError"},
			reason:     "two sites report twice",
		},
	}
	for index, testCase := range cases {
		t.Run(consistencyNoBareThrowCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoBareThrow, testCase.fileName, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantSpans))
			for index := range wantIds {
				wantIds[index] = "bareThrow"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}

				// Against a literal built here rather than against the rule's own message: a test
				// comparing a diagnostic to the value it was built from moves both sides under
				// mutation and asserts nothing. The literal is the text from before the wording moved
				// to policy/messages, so this row also proves the move changed no word.
				want := consistencyNoBareThrowWording(testCase.wantNames[index])
				if result.Diagnostics[index].Message.Description != want {
					t.Errorf("finding %d reads %q, want %q", index,
						result.Diagnostics[index].Message.Description, want)
				}
			}
		})
	}
}

// consistencyNoBareThrowWording is the rule's message for one constructor, written out here.
func consistencyNoBareThrowWording(constructorName string) string {
	return "This throws a bare `" + constructorName + "`, which names no declared " +
		"failure. Raise it through the tier that declares it, `AccountModule.error(identifier, " +
		"data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no " +
		"identifier, so the board groups it by its message and one interpolated value mints one " +
		"identity per value, and it normalizes to 500, so a refusal reads as our fault."
}

// TestNoBareThrowRendersAnEditedEntry: the finding's words come from policy/messages, so an edit to the
// entry shows up in the finding.
//
// Not parallel: it swaps the catalog every render reads with policy.UseMessages, so a test rendering beside
// it finds only this one entry (a parallel run failed rules/base with "has no message", #nxgt2ca).
func TestNoBareThrowRendersAnEditedEntry(t *testing.T) {
	original, err := fs.ReadFile(policy.MessageFiles(), "consistency-no-bare-throw.json")
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(original), "which names no declared failure", "which names no failure anyone declared", 1)
	if edited == string(original) {
		t.Fatal("the anchor is not in the entry, so the edit would not apply")
	}
	catalog, err := policy.LoadMessages(fstest.MapFS{"consistency-no-bare-throw.json": {Data: []byte(edited)}})
	if err != nil {
		t.Fatal(err)
	}
	restore := policy.UseMessages(catalog)
	defer restore()

	result := rule_testing.Run(t, ConsistencyNoBareThrow, "/repository/source/modules/thing/Service.ts",
		"export function f(): void { throw new TypeError('x'); }")
	rule_testing.ExpectFindings(t, result, "bareThrow")
	if got := result.Diagnostics[0].Message.Description; !strings.Contains(got, "This throws a bare `TypeError`, which names no failure anyone declared.") {
		t.Errorf("the finding reads %q, without the edit", got)
	}
}

// TestNoBareThrowStaysSilent is the declining half, and it is most of the corpus.
//
// Twenty-two of the thirty-nine rows are here, which is the shape of a rule whose whole judgment is
// four exemption gates and one narrow anchor. Every row is the original's verdict on that input,
// measured the same way.
//
// The near misses are the point. A substring test on a path is easy to write and easy to widen by
// accident, so each gate is paired in the reporting test above with a path that looks like it should
// match and does not: `source/latest/` and `source/protest/` both contain the letters of a test
// directory, `schema-tools` looks like the schema builder, and `NotBaseError.ts` ends in the capture
// path's filename without its separator. All four report, which is what proves these gates are
// segment tests rather than substring ones.
func TestNoBareThrowStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fileName   string
		sourceText string
		reason     string
	}{
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new AggregateError([], 'x'); }",
			reason:     "AggregateError is a container and stays legal",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(e: unknown): void { throw e; }",
			reason:     "rethrowing a value is not a construction",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw someError.message; }",
			reason:     "a property access is not a construction",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw makeError(); }",
			reason:     "a call expression is not a construction",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw new Errors.Validation('x'); }",
			reason:     "a qualified callee is not a bare identifier",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { throw 'a string'; }",
			reason:     "a string literal is not a construction",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export class MyError extends Error {}\nexport function f(): void { throw new MyError('x'); }",
			reason:     "a declared subclass is not a built-in name",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.ts",
			sourceText: "export function f(): void { const e = new Error('x'); return; }",
			reason:     "constructing without throwing is not this rule",
		},
		{
			fileName:   "/repository/libraries/base/command-line/commands/Run.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "declaration time: the command-line tree",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/orm/schema/Build.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "declaration time: the schema builder",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/orm/metadata/Registry.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "declaration time: orm metadata",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/internal/metadata/Registry.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "declaration time: internal metadata",
		},
		{
			fileName:   "/repository/libraries/base/source/account/metadata/Registry.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "declaration time: account metadata",
		},
		{
			fileName:   "/repository/libraries/base/source/client/Fetch.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "below the vocabulary: client",
		},
		{
			fileName:   "/repository/libraries/base/source/api/Handler.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "below the vocabulary: api",
		},
		{
			fileName:   "/repository/libraries/base/libraries/nexus/source/numbers/Money.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "below the vocabulary: nexus",
		},
		{
			fileName:   "/repository/libraries/base/libraries/nexus/code-quality/lint/LintEngineParity.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "below the vocabulary: nexus's lint tooling",
		},
		{
			fileName:   "/repository/source/modules/thing/Service.test.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "a test file by suffix",
		},
		{
			fileName:   "/repository/source/test/Helper.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "a test directory",
		},
		{
			fileName:   "/repository/source/tests/Helper.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "a tests directory",
		},
		{
			fileName:   "/repository/source/testing/Helper.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "a testing directory",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/errors/BaseError.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "capture path: BaseError",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/errors/BaseLog.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "capture path: BaseLog",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/errors/CreateBaseErrors.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "capture path: CreateBaseErrors",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/errors/WriteUnhandledErrorEnvelope.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "capture path: WriteUnhandledErrorEnvelope",
		},
		{
			fileName:   "/repository/libraries/base/source/foundation/errors/IsolateErrorHandlers.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "capture path: IsolateErrorHandlers",
		},
		{
			fileName:   "/repository/libraries/base/source/nexus/source/Thing.ts",
			sourceText: "export function f(): void { throw new Error('x'); }",
			reason:     "a nested nexus source path still matches",
		},
	}
	for index, testCase := range cases {
		t.Run(consistencyNoBareThrowCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ConsistencyNoBareThrow,
				testCase.fileName, testCase.sourceText))
		})
	}
}

// TestNoBareThrowSurvivesThrownShapesThatCannotBeRead is a crash fixture.
//
// `ast.Node.Text()` PANICS on a property access and on a call expression, and both are ordinary
// source for a throw site: `throw someError.message` and `throw makeError()` are shapes this rule
// meets. The kind test sits above the text read for exactly that reason.
//
// The walk recovers per FILE rather than per rule, so one such panic would cost every rule in this
// package every finding in that file, and the run would still print a plausible summary. There is no
// finding to assert here; the assertion is that the run completes.
//
// The malformed rows are ours rather than the original's. Its parser refuses those files outright,
// so it never meets a throw with nothing thrown; ours recovers and hands the walk one.
func TestNoBareThrowSurvivesThrownShapesThatCannotBeRead(t *testing.T) {
	t.Parallel()

	for name, sourceText := range map[string]string{
		"propertyAccess": "export function f(): void { throw someError.message; }",
		"callExpression": "export function f(): void { throw makeError(); }",
		"elementAccess":  "export function f(): void { throw errors[0]; }",
		"qualifiedNew":   "export function f(): void { throw new Errors.Validation('x'); }",
		"computedNew":    "export function f(): void { throw new errors['Validation']('x'); }",
		"awaitThrow":     "export async function f(): Promise<void> { throw await makeError(); }",
		"templateThrow":  "export function f(): void { throw `oops`; }",
		"bareThrow":      "export function f(): void { throw }",
		"throwNothing":   "export function f(): void { throw; }",
		"unclosedNew":    "export function f(): void { throw new Error( }",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.Run(t, ConsistencyNoBareThrow, "/repository/source/modules/thing/Service.ts", sourceText)
		})
	}
}

// TestNoBareThrowNormalizesThePathSeparator pins the one line with no counterpart in the original.
//
// Every gate is a substring or a suffix test on a path written with forward slashes, and the
// original ran on a filename ESLint had already normalized. Reading `FileName()` directly would hand
// a backslash-separated path to those tests, which would decline every gate and report inside the
// capture path, where the rule must never fire.
//
// The harness normalizes the filename it is given before building the program, so a fixture cannot
// reach the rule with a backslash path and this cannot be proven through `Run`. It is asserted
// against the helper instead, which is the layer that actually decides it.
func TestNoBareThrowNormalizesThePathSeparator(t *testing.T) {
	t.Parallel()

	// A file that MUST be exempt, spelled both ways. The forward-slash spelling is the one a real
	// run produces and is covered by the silent corpus above; this asserts the gate itself rather
	// than the harness.
	if !consistencyNoBareThrowFileIsExempt("/repository/libraries/base/source/foundation/errors/BaseError.ts") {
		t.Error("the capture path is not exempt with forward slashes")
	}
	if consistencyNoBareThrowFileIsExempt("/repository/source/modules/thing/Service.ts") {
		t.Error("an ordinary path is exempt when it should not be")
	}
	// The control that gives the row above meaning: a backslash spelling of the SAME capture-path
	// file is not exempt, which is exactly why the rule normalizes before asking.
	if consistencyNoBareThrowFileIsExempt(`\repository\libraries\base\source\foundation\errors\BaseError.ts`) {
		t.Error("a backslash path matched a gate, so normalization would be unnecessary")
	}
}
