package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// uselessCallFile is where the fixtures pretend to live.
const uselessCallFile = "/repository/source/UselessCall.ts"

// The corpus is upstream's, extracted from its tester rather than retyped.
//
// `/tmp/lint-sources/eslint/tests/lib/rules/no-useless-call.js` carries 19 valid cases and 25
// invalid ones, every invalid case naming exactly one `unnecessaryCall`. Twenty-four are reproduced
// below; the twenty-fifth is a stated divergence and has its own test.
//
// The clean list is where the discrimination lives. Four of its entries are the `.apply()` split
// against `prefer-spread`: `foo.apply(null, args)` passes a variable rather than an array literal,
// so the rewrite would need spread syntax and belongs to that rule. Two are computed accesses
// through a variable key. Four are calls with no arguments at all. And the last is a private
// identifier, which merely spells the same word.
func TestNoUselessCallStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"foo.apply(obj, 1, 2);"},
		{"obj.foo.apply(null, 1, 2);"},
		{"obj.foo.apply(otherObj, 1, 2);"},
		{"a.b(x, y).c.foo.apply(a.b(x, z).c, 1, 2);"},
		{"foo.apply(obj, [1, 2]);"},
		{"obj.foo.apply(null, [1, 2]);"},
		{"obj.foo.apply(otherObj, [1, 2]);"},
		{"a.b(x, y).c.foo.apply(a.b(x, z).c, [1, 2]);"},
		{"a.b.foo.apply(a.b.c, [1, 2]);"},
		{"foo.apply(null, args);"},
		{"obj.foo.apply(obj, args);"},
		{"var call; foo[call](null, 1, 2);"},
		{"var apply; foo[apply](null, [1, 2]);"},
		{"foo.call();"},
		{"obj.foo.call();"},
		{"foo.apply();"},
		{"obj.foo.apply();"},
		{"obj?.foo.bar.call(obj.foo, 1, 2);"},                            // lang={"ecmaVersion":2020}
		{"class C { #call; wrap(foo) { foo.#call(undefined, 1, 2); } }"}, // lang={"ecmaVersion":2022}
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUselessCall, uselessCallFile, testCase.sourceText))
		})
	}
}

// The twenty-four reporting cases this port reproduces exactly.
//
// Three of them are why the receiver comparison must be over tokens rather than over source text:
// `a.b(x, y).c.foo.call(a.b(x, y).c, 1, 2)` matches across a call expression,
// `abc.get("foo", 0).concat.apply(abc . get("foo",  0 ), [1, 2])` matches across differing
// whitespace, and `[].concat.apply([ ], [1, 2])` matches an empty array literal written two ways,
// which is the edge case `hasSameTokens` carries from `prefer-spread`.
func TestNoUselessCallFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"foo.call(undefined, 1, 2);"},
		{"foo.call(void 0, 1, 2);"},
		{"foo.call(null, 1, 2);"},
		{"obj.foo.call(obj, 1, 2);"},
		{"a.b.c.foo.call(a.b.c, 1, 2);"},
		{"a.b(x, y).c.foo.call(a.b(x, y).c, 1, 2);"},
		{"foo.apply(undefined, [1, 2]);"},
		{"foo.apply(void 0, [1, 2]);"},
		{"foo.apply(null, [1, 2]);"},
		{"obj.foo.apply(obj, [1, 2]);"},
		{"a.b.c.foo.apply(a.b.c, [1, 2]);"},
		{"a.b(x, y).c.foo.apply(a.b(x, y).c, [1, 2]);"},
		{"[].concat.apply([ ], [1, 2]);"},
		{"[].concat.apply([\n/*empty*/\n], [1, 2]);"},
		{"abc.get(\"foo\", 0).concat.apply(abc . get(\"foo\",  0 ), [1, 2]);"},
		{"foo.call?.(undefined, 1, 2);"},
		{"foo?.call(undefined, 1, 2);"},
		{"(foo?.call)(undefined, 1, 2);"},
		{"obj.foo.call?.(obj, 1, 2);"},
		{"obj?.foo.call(obj, 1, 2);"},
		{"(obj?.foo).call(obj, 1, 2);"},
		{"(obj?.foo.call)(obj, 1, 2);"},
		{"obj?.foo.bar.call(obj?.foo, 1, 2);"},
		{"obj.foo?.bar.call(obj.foo, 1, 2);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUselessCall, uselessCallFile, testCase.sourceText), "unnecessaryCall")
		})
	}
}

// The one upstream reporting case this port deliberately does not reproduce.
//
// `(obj?.foo).bar` and `obj?.foo.bar` are not the same expression: the first throws when `obj` is
// null and the second short-circuits to undefined. ESTree has no node for parentheses, so upstream
// reads both sides of this comparison as the same tokens and reports; our parser keeps the node and
// the two sides differ.
//
// This is the same divergence `prefer-spread` records at its own line, reached through the same
// `hasSameTokens` oracle, so the two rules agree with each other rather than each inventing an
// answer. Measured with a signature probe over all nine optional-chain shapes in this corpus: this
// is the only one of them where our verdict differs from upstream's.
//
// Recorded as a passing case with the reasoning rather than deleted, because a deleted case reads
// as an oversight and a reader cannot tell it was considered.
//
// The control is the unparenthesized form, which is upstream's neighbouring case and does report,
// so the silence here is a measurement about the parentheses rather than about the shape.
func TestNoUselessCallKeepsParenthesizedOptionalChainsDistinct(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"(obj?.foo).bar.call(obj?.foo, 1, 2);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessCall, uselessCallFile, testCase.sourceText))
		})
	}

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUselessCall, uselessCallFile,
		"obj?.foo.bar.call(obj?.foo, 1, 2);"), "unnecessaryCall")
}

// A computed access with a STRING LITERAL key, which the imported corpus never writes.
//
// Upstream requires `computed === false`, so `foo["call"](undefined, 1, 2)` is clean. Measured on
// the installed rule for both method names. The corpus cannot see this: its two computed cases use a
// variable key, which any port declines for a different reason, so a port reusing this package's
// `isApplyMemberAccess` would pass all 44 imported cases while reporting a class upstream passes.
//
// That helper accepts a string-literal subscript on purpose, because for `prefer-spread`
// `foo['apply'](null, args)` really is the same call. Here it is the wrong answer, which is why this
// rule tests the access shape itself.
//
// The control is the dotted form, which reports.
func TestNoUselessCallDeclinesAComputedAccess(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"foo[\"call\"](undefined, 1, 2);"},
		{"foo[\"apply\"](null, [1, 2]);"},
		{"obj[\"foo\"][\"call\"](obj, 1, 2);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessCall, uselessCallFile, testCase.sourceText))
		})
	}

	// The control, so the silence above is about the brackets rather than about the receiver.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUselessCall, uselessCallFile,
		"obj[\"foo\"].call(obj, 1, 2);"), "unnecessaryCall")
}

// A method that is neither `call` nor `apply`, which the imported corpus never writes.
//
// Every clean case upstream ships either has no matching receiver or is already declined for
// another reason, so nothing in the corpus separates "the method name is one of two" from "the
// receiver matches". A mutant deleting the default arm let every method name through and survived
// all 49 fixtures; with it removed, `obj.foo.bar(obj, 1, 2)` reaches the receiver comparison,
// matches, and reports.
//
// Driven on the installed rule to fix which way these fall: `obj.foo.bar(obj, 1, 2)`,
// `obj.foo.bind(obj, 1, 2)` and `foo.reduce(null, 1, 2)` are all clean, and the neighbouring
// `obj.foo.call(obj, 1, 2)` reports. `bind` is in the list on purpose, because it is the nearest
// neighbour by meaning and belongs to `no-extra-bind` rather than to this rule.
//
// The control is the same shape spelled `call`, so the silence is about the name rather than about
// the receiver.
func TestNoUselessCallDeclinesOtherMethods(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"obj.foo.bar(obj, 1, 2);"},
		{"obj.foo.bind(obj, 1, 2);"},
		{"foo.reduce(null, 1, 2);"},
		{"obj.foo.applyTo(obj, [1, 2]);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessCall, uselessCallFile, testCase.sourceText))
		})
	}

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUselessCall, uselessCallFile,
		"obj.foo.call(obj, 1, 2);"), "unnecessaryCall")
}

// The rendered message, which the id assertions above cannot see.
//
// The Description interpolates the method name twice, and a `rule.Message` carries no rendering
// layer, so nothing between the format string and the reader checks it. A `.call()` finding whose
// sentence says `.apply()` has the right id, the right span, and tells the reader to look for
// something that is not there.
func TestNoUselessCallNamesTheMethod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantPrefix string
	}{
		{"foo.call(undefined, 1, 2);", "This `.call()` sets a `this`"},
		{"foo.apply(null, [1, 2]);", "This `.apply()` sets a `this`"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessCall, uselessCallFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			description := result.Diagnostics[0].Message.Description
			if !strings.HasPrefix(description, testCase.wantPrefix) {
				t.Fatalf("the message reads %q, which does not name the method it reports", description)
			}
		})
	}
}

// The reported span is the whole call, which no id fixture above can see.
//
// Upstream passes `node`, the CallExpression. A port anchoring on the callee or on the method name
// satisfies every assertion above while pointing past the arguments that make the finding true.
func TestNoUselessCallReportsTheWholeCall(t *testing.T) {
	t.Parallel()

	const sourceText = "obj.foo.call(obj, 1, 2);"
	result := rule_testing.Run(t, NoUselessCall, uselessCallFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	written := result.SourceFile.Text()
	reported := written[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "obj.foo.call(obj, 1, 2)" {
		t.Fatalf("reported %q, wanted the whole call", reported)
	}
}

// The boundary between this rule and `prefer-spread`, which the two corpora share syntax across.
//
// `.apply()` with a non-literal second argument belongs to `prefer-spread`, and with an array
// literal it belongs here. Neither rule may report both, or the same line gets two findings telling
// the reader to make two different edits. Upstream's clean case `foo.apply(null, args)` is the first
// half; the second is asserted with the array literal in place.
func TestNoUselessCallLeavesVariadicApplyToPreferSpread(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessCall, uselessCallFile,
		"declare const args: number[];\nfoo.apply(null, args);\n"))
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUselessCall, uselessCallFile,
		"foo.apply(null, [1, 2]);"), "unnecessaryCall")

	// And the mirror, so the partition is asserted from both sides rather than assumed.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, PreferSpread, uselessCallFile,
		"declare const args: number[];\nfoo.apply(null, args);\n"), "preferSpread")
	rule_testing.ExpectClean(t, rule_testing.Run(t, PreferSpread, uselessCallFile,
		"foo.apply(null, [1, 2]);"))
}
