package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, 29 valid and 62 invalid cases, extracted by loading its test file
// with a stubbed rule tester so nothing was retyped, then verified byte against byte afterwards.
//
// Twenty seven of the invalid cases carry `output: null`, so the declines are nearly half this
// rule rather than an edge. Each is asserted as reporting AND offering no fix, because a fixer that
// repaired one of them would be a defect no message assertion could see: the finding would look
// right and the repair would change the program's value.

// preferNumericLiteralsDeclinesToFix marks a case upstream reports and deliberately does not
// repair.
const preferNumericLiteralsDeclinesToFix = "\x00declines"

func TestPreferNumericLiteralsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "valid0", source: "parseInt(1);"},
		{name: "valid1", source: "parseInt(1, 3);"},
		{name: "valid2", source: "Number.parseInt(1);"},
		{name: "valid3", source: "Number.parseInt(1, 3);"},
		{name: "valid4", source: "0b111110111 === 503;"},
		{name: "valid5", source: "0o767 === 503;"},
		{name: "valid6", source: "0x1F7 === 503;"},
		{name: "valid7", source: "a[parseInt](1,2);"},
		{name: "valid8", source: "parseInt(foo);"},
		{name: "valid9", source: "parseInt(foo, 2);"},
		{name: "valid10", source: "Number.parseInt(foo);"},
		{name: "valid11", source: "Number.parseInt(foo, 2);"},
		{name: "valid12", source: "parseInt(11, 2);"},
		{name: "valid13", source: "Number.parseInt(1, 8);"},
		{name: "valid14", source: "parseInt(1e5, 16);"},
		{name: "valid15", source: "parseInt('11', '2');"},
		{name: "valid16", source: "Number.parseInt('11', '8');"},
		{name: "valid17", source: "parseInt(/foo/, 2);"},
		{name: "valid18", source: "parseInt(`11${foo}`, 2);"},
		{name: "valid19", source: "parseInt('11', 2n);"},
		{name: "valid20", source: "Number.parseInt('11', 8n);"},
		{name: "valid21", source: "parseInt('11', 16n);"},
		{name: "valid22", source: "parseInt(`11`, 16n);"},
		{name: "valid23", source: "parseInt(1n, 2);"},
		{name: "valid24", source: "class C { #parseInt; foo() { Number.#parseInt(\"111110111\", 2); } }"},
		{name: "valid25", source: "function foo(parseInt) { parseInt(\"111110111\", 2); }"},
		{name: "valid26", source: "function foo() { var parseInt; parseInt(\"111110111\", 2); }"},
		{name: "valid27", source: "function foo(Number) { Number.parseInt(\"111110111\", 2); }"},
		{name: "valid28", source: "function foo() { var Number; Number.parseInt(\"111110111\", 2); }"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts", testCase.source))
		})
	}
}

func TestPreferNumericLiteralsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		wantFixed string
	}{
		{name: "invalid0", source: "parseInt(\"111110111\", 2) === 503;", wantFixed: "0b111110111 === 503;\n"},
		{name: "invalid1", source: "parseInt(\"767\", 8) === 503;", wantFixed: "0o767 === 503;\n"},
		{name: "invalid2", source: "parseInt(\"1F7\", 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid3", source: "Number.parseInt(\"111110111\", 2) === 503;", wantFixed: "0b111110111 === 503;\n"},
		{name: "invalid4", source: "Number.parseInt(\"767\", 8) === 503;", wantFixed: "0o767 === 503;\n"},
		{name: "invalid5", source: "Number.parseInt(\"1F7\", 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid6", source: "parseInt('7999', 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid7", source: "parseInt('1234', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid8", source: "parseInt('1234.5', 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid9", source: "parseInt('1\ufe0f\u20e33\ufe0f\u20e33\ufe0f\u20e37\ufe0f\u20e3', 16);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid10", source: "Number.parseInt('7999', 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid11", source: "Number.parseInt('1234', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid12", source: "Number.parseInt('1234.5', 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid13", source: "Number.parseInt('1\ufe0f\u20e33\ufe0f\u20e33\ufe0f\u20e37\ufe0f\u20e3', 16);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid14", source: "parseInt(`111110111`, 2) === 503;", wantFixed: "0b111110111 === 503;\n"},
		{name: "invalid15", source: "parseInt(`767`, 8) === 503;", wantFixed: "0o767 === 503;\n"},
		{name: "invalid16", source: "parseInt(`1F7`, 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid17", source: "parseInt('', 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid18", source: "parseInt(``, 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid19", source: "parseInt(`7999`, 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid20", source: "parseInt(`1234`, 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid21", source: "parseInt(`1234.5`, 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid22", source: "parseInt('11', 2)", wantFixed: "0b11\n"},
		{name: "invalid23", source: "Number.parseInt('67', 8)", wantFixed: "0o67\n"},
		{name: "invalid24", source: "5+parseInt('A', 16)", wantFixed: "5+0xA\n"},
		{name: "invalid25", source: "function *f(){ yield(Number).parseInt('11', 2) }", wantFixed: "function *f(){ yield 0b11 }\n"},
		{name: "invalid26", source: "function *f(){ yield(Number.parseInt)('67', 8) }", wantFixed: "function *f(){ yield 0o67 }\n"},
		{name: "invalid27", source: "function *f(){ yield(parseInt)('A', 16) }", wantFixed: "function *f(){ yield 0xA }\n"},
		{name: "invalid28", source: "function *f(){ yield Number.parseInt('11', 2) }", wantFixed: "function *f(){ yield 0b11 }\n"},
		{name: "invalid29", source: "function *f(){ yield/**/Number.parseInt('67', 8) }", wantFixed: "function *f(){ yield/**/0o67 }\n"},
		{name: "invalid30", source: "function *f(){ yield(parseInt('A', 16)) }", wantFixed: "function *f(){ yield(0xA) }\n"},
		{name: "invalid31", source: "parseInt('11', 2)+5", wantFixed: "0b11+5\n"},
		{name: "invalid32", source: "Number.parseInt('17', 8)+5", wantFixed: "0o17+5\n"},
		{name: "invalid33", source: "parseInt('A', 16)+5", wantFixed: "0xA+5\n"},
		{name: "invalid34", source: "parseInt('11', 2)in foo", wantFixed: "0b11 in foo\n"},
		{name: "invalid35", source: "Number.parseInt('17', 8)in foo", wantFixed: "0o17 in foo\n"},
		{name: "invalid36", source: "parseInt('A', 16)in foo", wantFixed: "0xA in foo\n"},
		{name: "invalid37", source: "parseInt('11', 2) in foo", wantFixed: "0b11 in foo\n"},
		{name: "invalid38", source: "Number.parseInt('17', 8)/**/in foo", wantFixed: "0o17/**/in foo\n"},
		{name: "invalid39", source: "(parseInt('A', 16))in foo", wantFixed: "(0xA)in foo\n"},
		{name: "invalid40", source: "/* comment */Number.parseInt('11', 2);", wantFixed: "/* comment */0b11;\n"},
		{name: "invalid41", source: "Number/**/.parseInt('11', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid42", source: "Number//\n.parseInt('11', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid43", source: "Number./**/parseInt('11', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid44", source: "Number.parseInt(/**/'11', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid45", source: "Number.parseInt('11', /**/2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid46", source: "Number.parseInt('11', 2)/* comment */;", wantFixed: "0b11/* comment */;\n"},
		{name: "invalid47", source: "parseInt/**/('11', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid48", source: "parseInt(//\n'11', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid49", source: "parseInt('11'/**/, 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid50", source: "parseInt(`11`/**/, 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid51", source: "parseInt('11', 2 /**/);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid52", source: "parseInt('11', 2)//comment\n;", wantFixed: "0b11//comment\n;\n"},
		{name: "invalid53", source: "parseInt?.(\"1F7\", 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid54", source: "Number?.parseInt(\"1F7\", 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid55", source: "Number?.parseInt?.(\"1F7\", 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid56", source: "(Number?.parseInt)(\"1F7\", 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid57", source: "(Number?.parseInt)?.(\"1F7\", 16) === 255;", wantFixed: "0x1F7 === 255;\n"},
		{name: "invalid58", source: "parseInt('1_0', 2);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid59", source: "Number.parseInt('5_000', 8);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid60", source: "parseInt('0_1', 16);", wantFixed: preferNumericLiteralsDeclinesToFix},
		{name: "invalid61", source: "Number.parseInt('0_0', 16);", wantFixed: preferNumericLiteralsDeclinesToFix},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, "useLiteral")
			if testCase.wantFixed == preferNumericLiteralsDeclinesToFix {
				if len(result.Diagnostics) > 0 && len(result.Diagnostics[0].Fixes) > 0 {
					t.Errorf("upstream declines to fix this case and this offered %d fixes; the "+
						"repair would change what the program computes",
						len(result.Diagnostics[0].Fixes))
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// The span and the interpolated function name, neither of which ExpectFindings can see. The name is
// read out of the source rather than being a constant, so each callee spelling names itself.
func TestPreferNumericLiteralsSpanAndMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		source       string
		wantText     string
		wantSystem   string
		wantFunction string
	}{
		{"plainCall", "parseInt('11', 2);", "parseInt('11', 2)", "binary", "parseInt"},
		{"memberCall", "Number.parseInt('767', 8);", "Number.parseInt('767', 8)", "octal",
			"Number.parseInt"},
		{"hexadecimal", "parseInt('1F7', 16);", "parseInt('1F7', 16)", "hexadecimal", "parseInt"},
		// The callee text is whatever was written, optional chaining and all.
		{"optionalChain", "Number?.parseInt('11', 2);", "Number?.parseInt('11', 2)", "binary",
			"Number?.parseInt"},
		{"parenthesizedCallee", "(Number.parseInt)('11', 2);", "(Number.parseInt)('11', 2)",
			"binary", "(Number.parseInt)"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			written := strings.TrimSpace(testCase.source) + "\n"
			if got := written[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Errorf("reported text: got %q, want %q", got, testCase.wantText)
			}
			wantPrefix := "Use " + testCase.wantSystem + " literals instead of " +
				testCase.wantFunction + "(). "
			if got := finding.Message.Description; !strings.HasPrefix(got, wantPrefix) {
				t.Errorf("message:\n got %q\nwant prefix %q", got, wantPrefix)
			}
		})
	}
}

// The global-reference guard, which is the half of this rule that does not fall out of the syntax.
// Five of upstream's passing cases are shadows and nothing about the syntax separates them from the
// real thing.
func TestPreferNumericLiteralsGlobalGuard(t *testing.T) {
	t.Parallel()

	shadowed := []string{
		"function foo(parseInt: any) { parseInt('111110111', 2); }",
		"function foo() { var parseInt: any; parseInt('111110111', 2); }",
		"function foo(Number: any) { Number.parseInt('111110111', 2); }",
		"function foo() { var Number: any; Number.parseInt('111110111', 2); }",
		"{ const parseInt = (a: string, b: number) => 0; parseInt('11', 2); }",
	}
	for _, source := range shadowed {
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts", source))
	}

	// A control, so five clean verdicts mean something.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts",
		"parseInt('111110111', 2);"), "useLiteral")
}

// The value comparison, which is what decides most of the declines. Each pair here writes an input
// that reports with a fix beside one that reports without, differing only in the thing being
// measured, so a comparison that answered the same for both would fail rather than half-pass.
func TestPreferNumericLiteralsValueComparison(t *testing.T) {
	t.Parallel()

	fixed := []struct {
		name      string
		source    string
		wantFixed string
	}{
		{"validOctalDigits", "parseInt('777', 8);", "0o777;"},
		{"validBinaryDigits", "parseInt('1010', 2);", "0b1010;"},
		{"validHexadecimalDigits", "parseInt('ff', 16);", "0xff;"},
		{"leadingZeroIsFine", "parseInt('0011', 2);", "0b0011;"},
	}
	for _, testCase := range fixed {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed+"\n")
		})
	}

	declined := []struct {
		name   string
		source string
		why    string
	}{
		{"digitOutOfRangeForOctal", "parseInt('7999', 8);",
			"parseInt reads 7 and stops; 0o7999 is not a literal"},
		{"digitOutOfRangeForBinary", "parseInt('1234', 2);",
			"parseInt reads 1 and stops; 0b1234 is not a literal"},
		{"fractionIsTruncated", "parseInt('1234.5', 8);",
			"parseInt stops at the dot; the literal would carry it"},
		{"numericSeparator", "parseInt('1_0', 2);",
			"parseInt reads 1 and stops at the underscore, while 0b1_0 is 2"},
		{"emptyString", "parseInt('', 8);", "parseInt is NaN and 0o is not a literal"},
		{"signIsNotALiteralPrefix", "parseInt('-11', 2);",
			"parseInt is -3 while 0b-11 is not a literal"},
	}
	for _, testCase := range declined {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
			}
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Errorf("offered a fix where upstream declines: %s", testCase.why)
			}
		})
	}
}

// Cases upstream's corpus does not write, each found by a surviving mutant and then measured
// against the installed build at 10.8.1 with a control that fired.
//
// Upstream gets both of these for free from a selector, `CallExpression[arguments.length=2]` and
// `isSpecificId(calleeNode, "parseInt")`, so neither needed a test there. Expressed as code here,
// each becomes a line that can be wrong on its own.
func TestPreferNumericLiteralsArgumentCountAndCalleeName(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name   string
		source string
	}{
		// The argument count is exactly two. A third argument means the call is not the one this
		// rule knows how to rewrite, and dropping it to "at least two" reports all of these.
		{"threeArguments", "parseInt('11', 2, 3);"},
		{"fourArguments", "parseInt('11', 2, 3, 4);"},
		{"threeArgumentsThroughNumber", "Number.parseInt('11', 2, 3);"},

		// The callee name is `parseInt` specifically. Every other global resolves to ambient
		// declarations exactly as it does, so without the name test each of these reports and
		// offers to rewrite a call to a different function.
		{"anotherGlobalTakingAString", "isNaN('11', 2);"},
		{"parseFloat", "parseFloat('11', 2);"},
		{"numberParseFloat", "Number.parseFloat('11', 2);"},
		{"numberIsInteger", "Number.isInteger('11', 2);"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts", testCase.source))
		})
	}

	// Controls, so seven clean verdicts mean something.
	t.Run("controlTwoArgumentsReports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts",
			"parseInt('11', 2);"), "useLiteral")
	})
	t.Run("controlNumberParseIntReports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferNumericLiterals, "file.ts",
			"Number.parseInt('11', 2);"), "useLiteral")
	})
}
