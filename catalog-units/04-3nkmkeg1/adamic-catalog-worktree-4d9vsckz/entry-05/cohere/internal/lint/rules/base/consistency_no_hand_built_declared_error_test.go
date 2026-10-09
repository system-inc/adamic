package base

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func consistencyNoHandBuiltDeclaredErrorCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoHandBuiltDeclaredErrorStaysSilent covers every shape the original declines.
//
// This is one of Kirk's own rules rather than an upstream port, so there is no corpus to import and
// every case here was invented. The brief warns that an invented fixture encodes the same belief as
// the port, which would make this list worthless as evidence. The mitigation is that the original
// rule is executable: each source below was driven through it, loaded out of api-phi-health with the
// `@nexus` alias mapped by hand, and the verdict recorded here is what the original produced rather
// than what I expected.
//
// Four of these are places the original KNOWINGLY misses, all four measured rather than reasoned
// about: a computed key, a string-literal key, a spread, and a namespaced constructor. Each plainly
// carries or names the option, and widening any of them would report inputs the original is silent
// on.
func TestNoHandBuiltDeclaredErrorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fileName   string
		sourceText string
	}{
		// No options object at all.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom');\n",
		},
		// Options naming something else. A BaseError carrying no identifier is a different thing and
		// stays legal, which is why the rule keys on the option rather than on the class.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { statusCode: 500 });\n",
		},
		// The options passed as a variable. The original inspects the literal's properties, so an
		// object it cannot see through is not examined. A knowing gap, reproduced.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst o = { identifier: 'X' };\nconst e = new BaseError('boom', o);\n",
		},
		// A COMPUTED key. The original tests `property.key.type === 'Identifier'`, which a computed key
		// is not, so it is silent even though the option is plainly there. Measured.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { ['identifier']: 'X' });\n",
		},
		// A STRING-LITERAL key, silent for the same reason. Both are places the rule knowingly misses,
		// and widening either would report inputs the original does not.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { 'identifier': 'X' });\n",
		},
		// A spread carrying the option. Not a Property, so not counted.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst rest = { identifier: 'X' };\nconst e = new BaseError('boom', { ...rest });\n",
		},
		// A NAMESPACED constructor. The original matches a bare `BaseError` identifier callee only, so
		// `Errors.BaseError` is silent. Measured rather than assumed.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare namespace Errors { class BaseError { constructor(m: string, o?: object); } }\nconst e = new Errors.BaseError('boom', { identifier: 'X' });\n",
		},
		// The two exempt files. The exemption is keyed on the filename so it is stated once where it is
		// true and cannot travel to a call site by someone pasting a comment.
		{
			fileName:   "/repository/source/a/b/BaseError.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { identifier: 'X' });\n",
		},
		{
			fileName:   "/repository/source/a/b/CreateBaseErrors.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { identifier: 'X' });\n",
		},
		// Another class taking the same option shape. The rule keys on the CLASS as well as on the
		// option, and no other case here constructs anything but a BaseError, so a mutant dropping
		// the callee-name test survived the whole suite. Measured clean against the original.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class OtherError { constructor(m: string, o?: object); }\nconst e = new OtherError('boom', { identifier: 'X' });\n",
		},
	}
	for index, testCase := range cases {
		t.Run(consistencyNoHandBuiltDeclaredErrorCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ConsistencyNoHandBuiltDeclaredError,
				testCase.fileName, testCase.sourceText))
		})
	}
}

// TestNoHandBuiltDeclaredErrorFires covers the shapes the original reports, with the span it uses.
//
// The span is the whole `new` expression rather than the option that triggered the finding, which is
// the original's `node`. Asserting it matters because pointing at the option instead would satisfy
// every message-id assertion while underlining something the reader was not shown.
func TestNoHandBuiltDeclaredErrorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fileName   string
		sourceText string
		wantSpan   string
	}{
		// The motivating shape, using the identifier from the original's own account example.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { identifier: 'AccountNotFound' });\n",
			wantSpan:   "new BaseError('boom', { identifier: 'AccountNotFound' })",
		},
		// Both options written by hand, which is the pairing the original says nothing checks.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { identifier: 'X', statusCode: 404 });\n",
			wantSpan:   "new BaseError('boom', { identifier: 'X', statusCode: 404 })",
		},
		// A SHORTHAND property. estree makes this a Property whose key is an Identifier named
		// `identifier`, so the original counts it; our parser gives it its own node kind, which is why
		// the predicate names both spellings. Measured reporting against the original.
		{
			fileName:   "/repository/source/Probe.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst identifier = 'X';\nconst e = new BaseError('boom', { identifier });\n",
			wantSpan:   "new BaseError('boom', { identifier })",
		},
		// A file whose name ENDS in BaseError.ts without being it, which REPORTS. The suffix includes
		// the separator, so the exemption does not leak to a neighbour.
		//
		// This row was nearly wrong in the worst way. The first measurement of all three filename
		// cases came back clean, and all three were false: an absolute path did not match the
		// probe config's files glob, so the linter declined the file and returned 'No matching
		// configuration found' instead of a verdict. That reads exactly like a clean result. The
		// probe captured non-rule messages alongside the findings, which is the only reason it was
		// caught; relative paths fixed it.
		{
			fileName:   "/repository/source/a/b/MyBaseError.ts",
			sourceText: "declare class BaseError { constructor(m: string, o?: object); }\nconst e = new BaseError('boom', { identifier: 'X' });\n",
			wantSpan:   "new BaseError('boom', { identifier: 'X' })",
		},
	}
	for index, testCase := range cases {
		t.Run(consistencyNoHandBuiltDeclaredErrorCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoHandBuiltDeclaredError, testCase.fileName,
				testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "consistencyNoHandBuiltDeclaredError")

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]
			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
			if diagnostic.Message.Description != messageNoHandBuiltDeclaredError().Description {
				t.Fatalf("message: expected %q, got %q",
					messageNoHandBuiltDeclaredError().Description, diagnostic.Message.Description)
			}
		})
	}
}
