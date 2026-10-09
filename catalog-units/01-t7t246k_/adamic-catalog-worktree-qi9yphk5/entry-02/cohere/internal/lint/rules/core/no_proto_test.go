package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// protoFile is where the fixtures pretend to live.
const protoFile = "/repository/source/Proto.ts"

// The corpus is ESLint's own, extracted rather than retyped.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-proto.js` at eslint 10.8.1: 5 valid
// and 4 invalid, each invalid case naming one `unexpectedProto` and carrying its own column
// assertion. Extracted by executing that file against a stub RuleTester and rendering every string
// through a serializer, so no escape was typed on the way here. The one that matters is the embedded
// newline in the fourth clean case, which a heredoc cooks into a real newline and thereby turns the
// sharpest clean case into a different, harmless one. That happened once while writing this file.
//
// All 9 were additionally driven through the installed eslint build and agreed with the file,
// spans included.
func TestNoProtoFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a dotted access", "var a = test.__proto__;"},
		{"a string-literal subscript", "var a = test['__proto__'];"},
		{"a template-literal subscript", "var a = test[`__proto__`];"},
		{"an assignment through a template subscript", "test[`__proto__`] = function () {};"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoProto, protoFile, testCase.sourceText), "unexpectedProto")
		})
	}
}

// The clean cases are the whole discrimination, and each fails a different way.
//
// The first reads a *variable* named `__proto__`, so the property is whatever it holds. The second
// is a declaration rather than an access. The third is a different key that merely starts the same.
// The fourth contains the name followed by a newline, so any rule matching on source text reports
// it. The fifth is a private name, which cannot collide because its text carries the leading hash.
func TestNoProtoStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a computed access through a variable", "var a = test[__proto__];"},
		{"a declaration binding the name", "var __proto__ = null;"},
		{"a template naming a different key", "foo[`__proto`] = null;"},
		{"a template whose text merely contains it", "foo[`__proto__\n`] = null;"},
		{"a private name of the same spelling", "class C { #__proto__; foo() { this.#__proto__; } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoProto, protoFile, testCase.sourceText))
		})
	}
}

// The spans, which the message-id fixtures above cannot see.
//
// Taken from the corpus's own column assertions rather than from reading our own output, so they
// record upstream's choice rather than ours. The finding covers the whole member access including
// the object, which is why the assignment case starts at offset 0 and not at the bracket.
//
// `rule_testing.Run` does not trim its input, so these offsets index the literal directly. The trimming
// hazard the brief describes applies to `RunTyped`, which this rule does not use.
func TestNoProtoReportsTheWholeAccess(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantPos    int
		wantEnd    int
	}{
		{"var a = test.__proto__;", 8, 22},
		{"var a = test['__proto__'];", 8, 25},
		{"var a = test[`__proto__`];", 8, 25},
		{"test[`__proto__`] = function () {};", 0, 17},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoProto, protoFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			got := result.Diagnostics[0].Range
			if got.Pos() != testCase.wantPos || got.End() != testCase.wantEnd {
				t.Fatalf("reported [%d:%d), wanted [%d:%d) which is %q",
					got.Pos(), got.End(), testCase.wantPos, testCase.wantEnd,
					testCase.sourceText[testCase.wantPos:testCase.wantEnd])
			}
		})
	}
}

// The message, asserted against literals typed here rather than against the rule's own constants,
// which would move with it.
func TestNoProtoReportsWhyItMatters(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoProto, protoFile, "var a = test.__proto__;")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "unexpectedProto" {
		t.Fatalf("message id was %q", got)
	}
	const wantPrefix = "This reads or writes `__proto__`, which reaches an object's prototype"
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("description was %q", got)
	}
}

// Cases written from reading our code rather than upstream's.
//
// The imported corpus exercises no substituting template and no object-literal key. A port reading a
// template's cooked text without checking for substitutions would report the first; a port anchored
// on "any node naming this property" would report the second and third. All three are clean on the
// installed build, measured before they were written here.
func TestNoProtoDeclinesShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a substituting template subscript", "declare const part: string;\nexport const a = test[`__proto${part}__`];\n"},
		{"an object-literal key", "export const a = { __proto__: 1 };\n"},
		{"a class field", "export class C { __proto__ = 1; }\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoProto, protoFile, testCase.sourceText))
		})
	}
}

// Optional chaining reports, measured on the installed build before it was written here.
//
// Both spellings fire upstream. Our parser hangs both off the same two access kinds the rule already
// listens to, so this pins that no extra arm is needed rather than documenting a limit.
func TestNoProtoReportsThroughOptionalChaining(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a dotted optional access", "declare const test: any;\nexport const a = test?.__proto__;\n"},
		{"a subscripted optional access", "declare const test: any;\nexport const a = test?.[`__proto__`];\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoProto, protoFile, testCase.sourceText), "unexpectedProto")
		})
	}
}
