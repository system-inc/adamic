package core

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noPlusplusFile is where the fixtures pretend to live.
//
// A `.ts` name: this rule has no file gate, and nothing in its corpus is JSX.
const noPlusplusFile = "/repository/source/NoPlusplus.ts"

// noPlusplusCase is one row of upstream's corpus.
type noPlusplusCase struct {
	// name is the corpus list and index the row came from, so a failure names a case that
	// can be found in upstream's own file rather than a number local to this table.
	name string

	// source is upstream's `code`, byte for byte.
	source string

	// options is the RAW JSON of upstream's options object, routed through the rule's own
	// exported decoder rather than built as a struct, so the decoder's defaults and its
	// empty-input path are under test.
	options string

	// ids are the message ids upstream produced for this input, in order.
	ids []string
}

// noPlusplusFiresCases are the rows upstream reports on.
var noPlusplusFiresCases = []noPlusplusCase{
	{
		name:    "invalid-0",
		source:  "var foo = 0; foo++;",
		options: "",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-1",
		source:  "var foo = 0; foo--;",
		options: "",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-2",
		source:  "for (i = 0; i < l; i++) { console.log(i); }",
		options: "",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-3",
		source:  "for (i = 0; i < l; foo, i++) { console.log(i); }",
		options: "",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-4",
		source:  "var foo = 0; foo++;",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-5",
		source:  "for (i = 0; i < l; i++) { v++; }",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-6",
		source:  "for (i++;;);",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-7",
		source:  "for (;--i;);",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-8",
		source:  "for (;;) ++i;",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-9",
		source:  "for (;; i = j++);",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-10",
		source:  "for (;; i++, f(--j));",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
	{
		name:    "invalid-11",
		source:  "for (;; foo + (i++, bar));",
		options: "{\"allowForLoopAfterthoughts\": true}",
		ids:     []string{"unexpectedUnaryOp"},
	},
}

// noPlusplusSilentCases are the rows upstream is clean on.
var noPlusplusSilentCases = []noPlusplusCase{
	{
		name:    "valid-0",
		source:  "var foo = 0; foo=+1;",
		options: "",
	},
	{
		name:    "valid-1",
		source:  "var foo = 0; foo=+1;",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-2",
		source:  "for (i = 0; i < l; i++) { console.log(i); }",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-3",
		source:  "for (var i = 0, j = i + 1; j < example.length; i++, j++) {}",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-4",
		source:  "for (;; i--, foo());",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-5",
		source:  "for (;; foo(), --i);",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-6",
		source:  "for (;; foo(), ++i, bar);",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-7",
		source:  "for (;; i++, (++j, k--));",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-8",
		source:  "for (;; foo(), (bar(), i++), baz());",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-9",
		source:  "for (;; (--i, j += 2), bar = j + 1);",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
	{
		name:    "valid-10",
		source:  "for (;; a, (i--, (b, ++j, c)), d);",
		options: "{\"allowForLoopAfterthoughts\": true}",
	},
}

// decodedNoPlusplus routes a row's raw JSON through the rule's own exported decoder.
func decodedNoPlusplus(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoPlusplusOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return decoded
}

func TestNoPlusplusFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noPlusplusFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoPlusplus, noPlusplusFile, testCase.source,
				decodedNoPlusplus(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestNoPlusplusStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noPlusplusSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoPlusplus, noPlusplusFile, testCase.source,
				decodedNoPlusplus(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoPlusplusNamesTheOperator asserts the message TEXT, not just the id.
//
// The generated tables above compare message ids, so the operator this rule interpolates is never
// read and a port that always said `++` would pass all 23 cases. Mutating the operator renderer to
// return `++` unconditionally SURVIVES those tables for exactly that reason.
//
// The coverage was already paid for: upstream's corpus asserts `Unary operator '--' used.` on three
// cases. This reads it. Widening an assertion is cheaper than adding a fixture, and it is the first
// thing to check when a mutant survives a corpus that looks like it should catch it.
func TestNoPlusplusNamesTheOperator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		operator string
	}{
		{"postfix-increment", "var foo = 0; foo++;\n", "++"},
		{"postfix-decrement", "var foo = 0; foo--;\n", "--"},
		{"prefix-increment", "var foo = 0; ++foo;\n", "++"},
		{"prefix-decrement", "var foo = 0; --foo;\n", "--"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoPlusplus, noPlusplusFile, testCase.source,
				decodedNoPlusplus(t, ""))
			rule_testing.ExpectFindings(t, result, "unexpectedUnaryOp")
			if len(result.Diagnostics) != 1 {
				return
			}
			wanted := "Unary operator `" + testCase.operator + "` used."
			if !strings.HasPrefix(result.Diagnostics[0].Message.Description, wanted) {
				t.Errorf("rendered %q, expected it to name the operator as %q",
					result.Diagnostics[0].Message.Description, testCase.operator)
			}
		})
	}
}

// TestNoPlusplusSurvivesNilOptions covers the path a bare `"error"` configuration takes.
//
// A rule configured without options is handed nil rather than the decoded default, and
// `options.(T)` on nil yields the zero value silently. Every generated row above routes through
// `DecodeNoPlusplusOptions`, which always returns a real `NoPlusplusOptions`, so the type assertion
// always succeeds and the fallback beside it is never exercised by them.
//
// That is why flipping the fallback's default to `true` survives the whole 23-case corpus. The
// survivor was read as equivalent and then checked against the code rather than inferred from the
// mutation: the fallback is reachable only through nil options, which only the config layer
// produces. This test is that path.
//
// For this rule the zero value happens to be the right answer, so the consequence of getting it
// wrong is narrow. It is pinned anyway, because the failure it guards against is the inert-rule
// shape: a rule that registers on every file and silently judges with the wrong default reads
// exactly like a clean tree.
func TestNoPlusplusSurvivesNilOptions(t *testing.T) {
	t.Parallel()

	// A for-loop update, which is exempt under `allowForLoopAfterthoughts: true` and reports under
	// the default. So this input separates the two defaults rather than merely exercising the path.
	const source = "for (i = 0; i < l; i++) { console.log(i); }\n"

	result := rule_testing.RunWithOptions(t, NoPlusplus, noPlusplusFile, source, nil)
	rule_testing.ExpectFindings(t, result, "unexpectedUnaryOp")
}
