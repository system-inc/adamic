package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// multiAssignFile is where the fixtures pretend to live.
const multiAssignFile = "/repository/source/MultiAssign.ts"

// decodedMultiAssignOptions routes a fixture's options through the rule's own exported decoder
// rather than building the struct directly.
//
// The decoder is the line most likely to have no upstream counterpart, and handing
// `RunWithOptions` a struct leaves it untested. An empty string means the rule is configured as
// bare "error", which is what the live config does and what hands the rule nil.
func decodedMultiAssignOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	options, err := DecodeNoMultiAssignOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return options
}

// The corpus is ESLint's own, at `tests/lib/rules/no-multi-assign.js`, copied rather than
// rewritten: 13 clean cases and 18 reporting ones, which is the whole upstream file.
//
// Each reporting case states its own error count, so the number of ids passed below is upstream's
// data rather than an assumption. Three cases report more than once and they are the reason a
// fixture asserting one finding per input would be wrong here.
//
// Every case string was verified byte against byte against the upstream file by script, and every
// verdict was reproduced by driving the installed eslint at 10.8.1 before being written here.
func TestNoMultiAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
		wantCount  int
	}{
		{"a chain in a declarator", "var a = b = c;", "", 1},
		{"a longer chain in a declarator", "var a = b = c = d;", "", 2},
		{"a let chain", "let foo = bar = cee = 100;", "", 2},
		{"a four-link bare chain", "a=b=c=d=e", "", 3},
		{"a two-link bare chain", "a=b=c", "", 1},
		{"a chain broken across lines", "a\n=b\n=c", "", 1},
		{"parentheses on both operands", "var a = (b) = (((c)))", "", 1},
		{"doubled parentheses on the target", "var a = ((b)) = (c)", "", 1},
		{"an arithmetic right-hand side", "var a = b = ( (c * 12) + 2)", "", 1},
		{"parentheses across lines", "var a =\n((b))\n = (c)", "", 1},
		{"a string containing an equals sign", "a = b = '=' + c + 'foo';", "", 1},
		{"an arithmetic right-hand side, bare", "a = b = 7 * 12 + 5;", "", 1},
		{"a member assignment in a declarator under ignoreNonDeclaration", "const x = {};\nconst y = x.one = 1;", "{\"ignoreNonDeclaration\": true}", 1},
		{"a bare chain under empty options", "let a, b;a = b = 1", "{}", 1},
		{"a bare chain under an explicit false", "let x, y;x = y = 'baz'", "{\"ignoreNonDeclaration\": false}", 1},
		{"a declarator chain under ignoreNonDeclaration", "const a = b = 1", "{\"ignoreNonDeclaration\": true}", 1},
		{"a class field chain", "class C { field = foo = 0 }", "", 1},
		{"a class field chain under ignoreNonDeclaration", "class C { field = foo = 0 }", "{\"ignoreNonDeclaration\": true}", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "unexpectedChain"
			}
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoMultiAssign,
				multiAssignFile, testCase.sourceText, decodedMultiAssignOptions(t, testCase.options)), wantIds...)
		})
	}
}

// The clean cases are what separate a chain from several declarators, which is the whole
// discrimination: `var a = 1, b = 2` is two initializers and `var a = b = c` is one chain.
//
// The last is the sharpest and would be a false positive for any rule anchored on the assignment
// rather than on its parent: `class C { [foo = 0] = 0 }` holds an assignment whose parent is the
// computed key rather than the field, so it is not an initializer at all.
func TestNoMultiAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
	}{
		{"plain declarations across lines", "var a, b, c,\nd = 0;", ""},
		{"separate declarations", "var a = 1; var b = 2; var c = 3;\nvar d = 0;", ""},
		{"a conditional in the initializer", "var a = 1 + (b === 10 ? 5 : 4);", ""},
		{"several const declarators", "const a = 1, b = 2, c = 3;", ""},
		{"separate const declarations", "const a = 1;\nconst b = 2;\n const c = 3;", ""},
		{"two var declarators in a for head", "for(var a = 0, b = 0;;){}", ""},
		{"two let declarators in a for head", "for(let a = 0, b = 0;;){}", ""},
		{"two const declarators in a for head", "for(const a = 0, b = 0;;){}", ""},
		{"an export with no initializer", "export let a, b;", ""},
		{"an export with one initializer", "export let a,\n b = 0;", ""},
		{"member chain under ignoreNonDeclaration", "const x = {};const y = {};x.one = y.one = 1;", "{\"ignoreNonDeclaration\": true}"},
		{"a bare chain under ignoreNonDeclaration", "let a, b;a = b = 1", "{\"ignoreNonDeclaration\": true}"},
		{"an assignment inside a computed key", "class C { [foo = 0] = 0 }", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoMultiAssign,
				multiAssignFile, testCase.sourceText, decodedMultiAssignOptions(t, testCase.options)))
		})
	}
}

// Cases upstream does not write, each measured against the installed eslint at 10.8.1 before
// being recorded here rather than reasoned about.
//
// The fifteen compound operators are the largest gap between the corpus and the rule. ESTree's
// `AssignmentExpression` covers all sixteen assignment operators and upstream's selector
// therefore does too, while the corpus writes only `=`. A port taking its fixtures from the
// corpus alone would ship a rule silent on fifteen of them and every fixture would pass.
//
// The two parenthesized cases run the opposite way to how the upstream source reads: `.init` and
// `.right` look like direct field reads that must decline a parenthesized operand, and espree
// gives parentheses no node so they already see through them.
func TestNoMultiAssignFiresOnCasesBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a plus-equals chain", "var q = w += e"},
		{"a minus-equals chain", "var q = w -= e"},
		{"a times-equals chain", "var q = w *= e"},
		{"a divide-equals chain", "var q = w /= e"},
		{"a modulo-equals chain", "var q = w %= e"},
		{"an exponent-equals chain", "var q = w **= e"},
		{"a left-shift-equals chain", "var q = w <<= e"},
		{"a right-shift-equals chain", "var q = w >>= e"},
		{"an unsigned-right-shift-equals chain", "var q = w >>>= e"},
		{"a bitwise-and-equals chain", "var q = w &= e"},
		{"a bitwise-or-equals chain", "var q = w |= e"},
		{"a bitwise-xor-equals chain", "var q = w ^= e"},
		{"a logical-and-equals chain", "var q = w &&= e"},
		{"a logical-or-equals chain", "var q = w ||= e"},
		{"a nullish-equals chain", "var q = w ??= e"},
		{"a parenthesized initializer", "var a = (b = c);"},
		{"a parenthesized right-hand side", "x = (y = z)"},
		{"a compound outer and a plain inner", "a += b = c"},
		{"a plain outer and a compound inner", "a = b += c"},
		{"a chain in a for head", "for (a = b = c;;) {}"},
		{"a chain in the second declarator", "var a = b, c = d = e;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoMultiAssign,
				multiAssignFile, testCase.sourceText, nil), "unexpectedChain")
		})
	}
}

// The clean cases beyond the corpus pin the two tests the corpus cannot reach: that the slot is
// decided by which FIELD of the parent holds the node, and that the OUTER expression has to be
// an assignment too.
//
// `a, b = c` and `a || (b = c)` are both binary expressions holding an assignment, and a rule
// testing only the parent's kind reports both.
func TestNoMultiAssignStaysSilentOnCasesBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a comma expression holding an assignment", "a, b = c"},
		{"a logical expression holding an assignment", "a || (b = c)"},
		{"an assignment as a call argument", "foo(a = b)"},
		{"an assignment as an array element", "var a = [b = c]"},
		{"a declarator with no initializer", "var a;"},
		{"a class field with no initializer", "class C { field }"},
		{"an equality comparison as an initializer", "var a = b == c;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoMultiAssign,
				multiAssignFile, testCase.sourceText, nil))
		})
	}
}

// The option's whole effect is dropping the bare-chain arm, so every case here reports without
// it and is clean with it. Each is run both ways in one subtest, which is what makes this a test
// of the option rather than two tests that happen to disagree.
//
// The corpus covers this for two inputs. These four widen it across the compound operators and
// through the parenthesis skip, both of which sit on the arm the option removes.
func TestNoMultiAssignIgnoreNonDeclarationDropsOnlyTheBareChain(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare chain", "a = b = c"},
		{"a compound bare chain", "a += b += c"},
		{"a member chain", "x.one = y.one = 1"},
		{"a parenthesized right-hand side", "x = (y = z)"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoMultiAssign, multiAssignFile,
				testCase.sourceText, decodedMultiAssignOptions(t, `{"ignoreNonDeclaration": false}`)), "unexpectedChain")
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoMultiAssign, multiAssignFile,
				testCase.sourceText, decodedMultiAssignOptions(t, `{"ignoreNonDeclaration": true}`)))
		})
	}
}

// Every reporting case upstream carries line and column assertions, so the span is stated data
// rather than something recovered. These convert a representative set of those columns into the
// slice the finding names.
//
// The parenthesized rows are the reason this test is not optional. `var a = (b) = (((c)))` reports
// at column 9, which is the `(` of `(b)`, so those parentheses are INSIDE the reported span; and
// `var a = (b = c);` reports at column 10, which is the `b`, so THOSE parentheses are outside it.
// One rule, two answers, decided by whether the parenthesis belongs to the assignment's operand or
// wraps the assignment itself. Both fall out of reporting the assignment node, and neither is
// visible to a message-id assertion.
func TestNoMultiAssignSpansTheInnerAssignment(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantReported []string
	}{
		{"a chain in a declarator", "var a = b = c;", []string{"b = c"}},
		{"a longer chain reports each link", "var a = b = c = d;", []string{"b = c = d", "c = d"}},
		{"a four-link bare chain", "a=b=c=d=e", []string{"b=c=d=e", "c=d=e", "d=e"}},
		{"operand parentheses are inside the span", "var a = (b) = (((c)))", []string{"(b) = (((c)))"}},
		{"wrapping parentheses are outside the span", "var a = (b = c);", []string{"b = c"}},
		{"a class field chain", "class C { field = foo = 0 }", []string{"foo = 0"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoMultiAssign, multiAssignFile, testCase.sourceText, nil)
			if len(result.Diagnostics) != len(testCase.wantReported) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantReported), len(result.Diagnostics))
			}
			for index, want := range testCase.wantReported {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != want {
					t.Errorf("finding %d points at %q, want %q", index, reported, want)
				}
			}
		})
	}
}

// The decoder is the line with no upstream counterpart, so it is tested directly rather than only
// through the fixtures that route past it.
//
// The nil row is the one that matters. A rule configured as bare "error" is handed nil options, so
// the type assertion in the rule yields the zero value. For this rule that happens to be correct,
// because the option defaults to false; the assertion exists so that a later default change fails
// here rather than silently switching the bare-chain arm off across the whole tree.
func TestDecodeNoMultiAssignOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"empty input falls back to the default", "", false},
		{"an empty object falls back to the default", "{}", false},
		{"an explicit false", `{"ignoreNonDeclaration": false}`, false},
		{"an explicit true", `{"ignoreNonDeclaration": true}`, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoMultiAssignOptions(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("the decoder refused %q: %v", testCase.raw, err)
			}
			options, isOptions := decoded.(NoMultiAssignOptions)
			if !isOptions {
				t.Fatalf("the decoder returned %T rather than NoMultiAssignOptions", decoded)
			}
			got := options.IgnoreNonDeclaration != nil && *options.IgnoreNonDeclaration
			if got != testCase.want {
				t.Errorf("ignoreNonDeclaration decodes to %v, want %v", got, testCase.want)
			}
		})
	}
}

// A rule configured as bare "error" reaches Run with nil rather than with an options struct, and
// every fixture above that passes nil already exercises that path. This asserts it once explicitly,
// so the reason is written down: nil must behave as the default rather than as "everything off".
func TestNoMultiAssignWithNilOptionsUsesTheDefault(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, NoMultiAssign, multiAssignFile, "a = b = c", nil), "unexpectedChain")
}

// The message text is asserted against a literal typed here rather than against the rule's own
// constant, because comparing a diagnostic to the constant it was built from moves both sides
// together under mutation and asserts nothing.
func TestNoMultiAssignMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, NoMultiAssign, multiAssignFile, "var a = b = c;", nil)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "unexpectedChain" {
		t.Errorf("message id is %q, want %q", got, "unexpectedChain")
	}
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, "This chains one assignment inside another") {
		t.Errorf("message description starts %q, which is not the sentence this rule reports", got)
	}
}

// An assignment on the LEFT of another assignment, which espree cannot parse at all.
//
// `(a = b) = c` is a fatal parse error upstream, invalid assignment target, so eslint reports
// nothing and the input is not expressible in its corpus. typescript-go recovers and hands back the
// shape, which puts an assignment in a slot upstream's selector language has no word for: the
// selector is `AssignmentExpression > AssignmentExpression.right`, and this one is the left.
//
// So the decision here is ours rather than ported, and it is to stay silent, which is what the
// `.right` in upstream's selector says even though nothing upstream can exercise it. Recorded as a
// fixture rather than as a comment because a mutation dropping the right-side test survived every
// other case in this file: this is the only input that separates them.
func TestNoMultiAssignStaysSilentOnAnAssignmentTargetTheParserRecovered(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a plain assignment as a target", "(a = b) = c"},
		{"a compound assignment as a target", "(a = b) += c"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoMultiAssign,
				multiAssignFile, testCase.sourceText, nil))
		})
	}
}
