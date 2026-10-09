package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// callerFile is where the fixtures pretend to live.
const callerFile = "/repository/source/Caller.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_caller.rs`: 4 pass, 2 fail, and the
// snapshot records 2 diagnostics from those 2 inputs, so one finding per input holds here rather
// than being assumed.
func TestNoCallerFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"callee", "var x = arguments.callee"},
		{"caller", "var x = arguments.caller"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoCaller, callerFile, testCase.sourceText), "noCaller")
		})
	}
}

// The fourth clean case is the one that matters and it is upstream's.
//
// `arguments[caller]` reads a variable named caller and accesses whatever property that names, so
// it is a different access from `arguments.caller` and may be neither of the two. A rule matching
// on source text reports it, and one treating a computed access as equivalent to a static member
// does too.
func TestNoCallerStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an ordinary property", "var x = arguments.length"},
		{"the object itself", "var x = arguments"},
		{"a numeric subscript", "var x = arguments[0]"},
		{"a subscript through a variable", "var x = arguments[caller]"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoCaller, callerFile, testCase.sourceText))
		})
	}
}

// Cases written because somebody read our code rather than upstream's.
//
// The imported corpus never writes the property on a receiver that is not `arguments`, and never
// parenthesizes the receiver. The first would report under a rule that checked only the property
// name; the second is the behavior `isIdentifierNamed` gives us for free by skipping parentheses,
// and upstream's `is_specific_id` does the same, so it is parity rather than divergence.
func TestNoCallerReadsTheReceiver(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoCaller, callerFile,
		"declare const options: { callee: number };\nexport const a = options.callee;\n"))

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoCaller, callerFile,
		"export function f() { return (arguments).callee; }\n"), "noCaller")
}

// Where the finding lands, which the message-id fixtures above cannot see.
//
// A mutation reporting the whole `arguments.callee` rather than just `callee` compiled and changed
// no fixture. Upstream reports the property span specifically, and the difference is not cosmetic:
// a finding anchored at the start of the access sits one token earlier, and on a wrapped expression
// that can be a different line from the one an author would suppress.
//
// Asserted against the source text the range covers, rather than against offsets, because an offset
// expectation is most likely to be wrong in the same direction as the code that produced it.
func TestNoCallerReportsTheProperty(t *testing.T) {
	t.Parallel()

	const source = "var x = arguments.callee"
	result := rule_testing.Run(t, NoCaller, callerFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "callee" {
		t.Fatalf("the finding covers %q, wanted just the property name", reported)
	}
}
