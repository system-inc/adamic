package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// radixFile is where the fixtures pretend to live.
const radixFile = "/repository/source/Radix.ts"

// The corpus is ESLint's own, imported from eslint/tests/lib/rules/radix.js.
//
// Upstream ships 48 pass and 43 fail; 46 and 43 are expressible here. The rest need a
// configured-globals surface this tool does not have: 2 pass and 0 fail, all of them
// turning parseInt or Number off through a globals key or a /*globals*/ directive.
//
// Extracted by loading upstream's tester with a stubbed RuleTester and rendering these literals
// from that JSON, so nothing was retyped. The generator refuses any byte outside printable ASCII.
//
// Run through the TYPED harness because the shadow question is resolution: a local parseInt is
// not the global one, and only the checker can say which a name is.
func TestRadixFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		messages   []string
	}{
		{"parseInt();", []string{"missingParameters"}},
		{"parseInt(\"10\");", []string{"missingRadix"}},
		{"parseInt(\"10\",);", []string{"missingRadix"}},
		{"parseInt((0, \"10\"));", []string{"missingRadix"}},
		{"parseInt((0, \"10\"),);", []string{"missingRadix"}},
		{"parseInt(\"10\", null);", []string{"invalidRadix"}},
		{"parseInt(\"10\", undefined);", []string{"invalidRadix"}},
		{"parseInt(\"10\", true);", []string{"invalidRadix"}},
		{"parseInt(\"10\", \"foo\");", []string{"invalidRadix"}},
		{"parseInt(\"10\", \"123\");", []string{"invalidRadix"}},
		{"parseInt(\"10\", 1);", []string{"invalidRadix"}},
		{"parseInt(\"10\", 37);", []string{"invalidRadix"}},
		{"parseInt(\"10\", -1);", []string{"invalidRadix"}},
		{"parseInt(\"10\", +37);", []string{"invalidRadix"}},
		{"parseInt(\"10\", -0);", []string{"invalidRadix"}},
		{"parseInt(\"10\", +10.5);", []string{"invalidRadix"}},
		{"parseInt(\"10\", 10.5);", []string{"invalidRadix"}},
		{"parseInt(\"10\", 1, ...args);", []string{"invalidRadix"}},
		{"Number.parseInt();", []string{"missingParameters"}},
		{"Number.parseInt(\"10\");", []string{"missingRadix"}},
		{"Number.parseInt(\"10\", 1);", []string{"invalidRadix"}},
		{"Number.parseInt(\"10\", 37);", []string{"invalidRadix"}},
		{"Number.parseInt(\"10\", -1);", []string{"invalidRadix"}},
		{"Number.parseInt(\"10\", 10.5);", []string{"invalidRadix"}},
		{"Number[\"parseInt\"]();", []string{"missingParameters"}},
		{"Number[\"parseInt\"](\"10\");", []string{"missingRadix"}},
		{"Number['parseInt']('10', 1);", []string{"invalidRadix"}},
		{"Number[`parseInt`](\"10\");", []string{"missingRadix"}},
		{"Number.parseInt(\"10\", 1, ...args);", []string{"invalidRadix"}},
		{"parseInt?.(\"10\");", []string{"missingRadix"}},
		{"Number.parseInt?.(\"10\");", []string{"missingRadix"}},
		{"Number?.parseInt(\"10\");", []string{"missingRadix"}},
		{"(Number?.parseInt)(\"10\");", []string{"missingRadix"}},
		{"Number?.[\"parseInt\"](\"10\");", []string{"missingRadix"}},
		{"Number[\"parseInt\"]?.(\"10\");", []string{"missingRadix"}},
		{"parseInt();", []string{"missingParameters"}},
		{"parseInt();", []string{"missingParameters"}},
		{"parseInt(\"10\");", []string{"missingRadix"}},
		{"parseInt(\"10\");", []string{"missingRadix"}},
		{"parseInt(\"10\", 1);", []string{"invalidRadix"}},
		{"parseInt(\"10\", 1);", []string{"invalidRadix"}},
		{"Number.parseInt();", []string{"missingParameters"}},
		{"Number.parseInt();", []string{"missingParameters"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, Radix, radixFile,
				testCase.sourceText), testCase.messages...)
		})
	}
}

// The clean cases, which carry most of this rule's discrimination.
//
// Four groups matter. A radix the rule cannot evaluate is ACCEPTED rather than guessed at, which
// is why foo, +radix and ~1 are all here. A spread in either of the first two positions abandons
// the call. A shadowed parseInt or Number is not the global one. And a private name or a computed
// key that is not a literal string is not Number.parseInt at all.
func TestRadixStaysSilent(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"parseInt(\"10\", 10);",
		"parseInt(\"10\", 2);",
		"parseInt(\"10\", 36);",
		"parseInt(\"10\", 0x10);",
		"parseInt(\"10\", 1.6e1);",
		"parseInt(\"10\", 10.0);",
		"parseInt(\"10\", +10);",
		"parseInt(\"10\", +0x10);",
		"parseInt(\"10\", +1.6e1);",
		"parseInt(\"10\", +\"10\");",
		"parseInt(\"10\", ~1);",
		"parseInt(\"10\", +radix);",
		"parseInt(\"10\", -radix);",
		"parseInt(\"10\", foo);",
		"function foo(undefined) { parseInt(\"10\", undefined); }",
		"Number.parseInt(\"10\", foo);",
		"Number[\"parseInt\"](\"10\", foo);",
		"Number[`parseInt`](\"10\", 10);",
		"Number.parseInt(\"10\", +10);",
		"Number.parseInt(\"10\", +\"10\");",
		"Number.parseInt(\"10\", ~1);",
		"Number.parseInt(\"10\", +radix);",
		"Number.parseInt(\"10\", -radix);",
		"parseInt(...args);",
		"parseInt(...args, 1);",
		"parseInt(\"10\", ...args);",
		"parseInt(\"10\", 10, ...args);",
		"Number.parseInt(...args);",
		"Number.parseInt(...args, 1);",
		"Number.parseInt(\"10\", ...args);",
		"Number.parseInt(\"10\", 10, ...args);",
		"parseInt",
		"Number.foo();",
		"Number[parseInt]();",
		"class C { #parseInt; foo() { Number.#parseInt(); } }",
		"class C { #parseInt; foo() { Number.#parseInt(foo); } }",
		"class C { #parseInt; foo() { Number.#parseInt(foo, 'bar'); } }",
		"var parseInt; parseInt();",
		"var Number; Number.parseInt();",
		"let Number; Number['parseInt']();",
		"parseInt(\"10\", 10);",
		"parseInt(\"10\", 10);",
		"parseInt(\"10\", 8);",
		"parseInt(\"10\", 8);",
		"parseInt(\"10\", foo);",
		"parseInt(\"10\", foo);",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, Radix, radixFile, sourceText))
		})
	}
}

// The suggestion vectors, which are upstream's own suggestion outputs.
//
// rule_testing can apply a FIX but not a SUGGESTION, so the repair is applied by hand here. That
// is required rather than optional: a suggestion writing the right text over the wrong span
// passes every message-id assertion above, and this rule's insertion point is computed from the
// source text rather than from a node, which is exactly where such a defect would live.
//
// The trailing-comma rows are the reason the insertion reads the source at all: with a comma
// already present upstream writes " 10," rather than ", 10", so parseInt("10",) becomes
// parseInt("10", 10,) and not parseInt("10",, 10).
func TestRadixSuggestsARadixOfTen(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
	}{
		{"parseInt(\"10\");", "parseInt(\"10\", 10);"},
		{"parseInt(\"10\",);", "parseInt(\"10\", 10,);"},
		{"parseInt((0, \"10\"));", "parseInt((0, \"10\"), 10);"},
		{"parseInt((0, \"10\"),);", "parseInt((0, \"10\"), 10,);"},
		{"Number.parseInt(\"10\");", "Number.parseInt(\"10\", 10);"},
		{"Number[\"parseInt\"](\"10\");", "Number[\"parseInt\"](\"10\", 10);"},
		{"Number[`parseInt`](\"10\");", "Number[`parseInt`](\"10\", 10);"},
		{"parseInt?.(\"10\");", "parseInt?.(\"10\", 10);"},
		{"Number.parseInt?.(\"10\");", "Number.parseInt?.(\"10\", 10);"},
		{"Number?.parseInt(\"10\");", "Number?.parseInt(\"10\", 10);"},
		{"(Number?.parseInt)(\"10\");", "(Number?.parseInt)(\"10\", 10);"},
		{"Number?.[\"parseInt\"](\"10\");", "Number?.[\"parseInt\"](\"10\", 10);"},
		{"Number[\"parseInt\"]?.(\"10\");", "Number[\"parseInt\"]?.(\"10\", 10);"},
		{"parseInt(\"10\");", "parseInt(\"10\", 10);"},
		{"parseInt(\"10\");", "parseInt(\"10\", 10);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Radix, radixFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "missingRadix")
			if len(result.Diagnostics[0].Suggestions) != 1 {
				t.Fatalf("expected exactly one suggestion, got %d",
					len(result.Diagnostics[0].Suggestions))
			}
			suggestion := result.Diagnostics[0].Suggestions[0]
			if suggestion.Message.Id != "addRadixParameter10" {
				t.Errorf("expected the message id addRadixParameter10, got %q",
					suggestion.Message.Id)
			}
			// The typed harness writes strings.TrimSpace(source)+"\n", so the expectation is
			// transformed the same way the input was rather than the rule being padded to match.
			// Without this the applier looks broken on fifteen byte-correct repairs.
			wantSource := strings.TrimSpace(testCase.wantSource) + "\n"
			if applied := applyRadixSuggestion(t, result, suggestion); applied != wantSource {
				t.Errorf("applying the suggestion gave\n  %q\nwant\n  %q", applied, wantSource)
			}
			// The rule must not also propose an unattended fix. Upstream offers a suggestion
			// precisely because adding a radix changes what the call returns, and shipping the
			// same edit as a fix would have the engine apply it without anyone choosing it.
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Errorf("expected no unattended fix, got %d", len(result.Diagnostics[0].Fixes))
			}
		})
	}
}

// applyRadixSuggestion replays a suggestion's edits into the source the harness actually wrote.
//
// Hand-rolled because rule_testing has no ExpectSuggestedSource. The source is read back off the
// result rather than from the test literal, so the harness's own trimming cannot put this off by
// one the way it does for a span sliced from a Go literal.
func applyRadixSuggestion(t *testing.T, result rule_testing.Result,
	suggestion rule.Suggestion) string {
	t.Helper()

	source := result.SourceFile.Text()
	// Applied back to front so an earlier edit does not shift a later one's offsets.
	applied := source
	for index := len(suggestion.Fixes) - 1; index >= 0; index-- {
		fix := suggestion.Fixes[index]
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(applied) || start > end {
			t.Fatalf("the suggestion proposed a range [%d,%d) outside a %d byte source",
				start, end, len(applied))
		}
		applied = applied[:start] + fix.Text + applied[end:]
	}
	return applied
}

// The finding points at the whole call, not at the callee.
//
// Upstream reports the CallExpression, so `parseInt("10");` reports columns 1 to 15, which is
// `parseInt("10")` including the arguments and excluding the semicolon. Every message-id assertion
// above is satisfied by a rule reporting only the callee, so this is what records which was ported,
// and it matters here because the suggestion's insertion point is computed relative to the reported
// node's end.
func TestRadixPointsAtTheWholeCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		want       string
	}{
		{`parseInt("10");`, `parseInt("10")`},
		{`Number.parseInt("10");`, `Number.parseInt("10")`},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Radix, radixFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "missingRadix")
			source := result.SourceFile.Text()
			reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("expected the finding on %q, pointed at %q", testCase.want, reported)
			}
		})
	}
}

// Shadowing is resolved per call site here, where upstream resolves it per file.
//
// Upstream asks eslint-scope for the PROGRAM scope's `parseInt` variable and reports only if
// nothing defines it, which makes the answer whole-file. Measured against the installed rule,
// `{ let parseInt; } parseInt("10");` is CLEAN upstream, even though the block-scoped binding cannot
// reach the call and the call really is the global `parseInt` missing a radix.
//
// This port resolves the identifier at the call site, so that input reports. The divergence is
// deliberate and it runs toward catching a real defect rather than toward reporting correct code,
// which is the direction worth diverging in. Both answers are stated here so the next reader sees a
// decision rather than an accident.
//
// The rows around it pin the cases where the two agree, so this test would notice a port that had
// simply stopped checking for shadows at all.
func TestRadixResolvesShadowsPerCallSite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"a var shadow, clean in both", "var parseInt; parseInt();", nil},
		{"a parameter shadow, clean in both",
			`function f(parseInt) { parseInt("10"); }`, nil},
		{"a shadowed Number, clean in both", "var Number; Number.parseInt();", nil},
		{"the global, reporting in both", `parseInt("10");`, []string{"missingRadix"}},
		{"a block scoped binding that cannot reach the call, clean upstream and reporting here",
			`{ let parseInt; } parseInt("10");`, []string{"missingRadix"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, Radix, radixFile,
				testCase.sourceText), testCase.messages...)
		})
	}
}

// The deprecated option is accepted under both values and changes nothing, which is upstream's own
// behavior at 10.8.1: its schema still takes "always" or "as-needed" and its rule body never reads
// context.options.
//
// These are upstream's optioned rows, each run under no option and under both values. The two that
// matter most are the ones an older reading of "as-needed" would get wrong: `parseInt("10", 10)` is
// clean under it, where earlier releases reported a redundant radix, and `parseInt("10")` still
// reports, where those releases let a missing radix pass.
//
// The options go through DecodeRadixOptions rather than being built by hand, so a decoder that
// refused a value or decoded it into something the rule reacted to would fail here.
func TestRadixIgnoresItsDeprecatedOption(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		messages   []string
	}{
		{`parseInt("10", 10);`, nil},
		{`parseInt("10", 8);`, nil},
		{`parseInt("10", foo);`, nil},
		{`parseInt();`, []string{"missingParameters"}},
		{`parseInt("10");`, []string{"missingRadix"}},
		{`parseInt("10", 1);`, []string{"invalidRadix"}},
		{`Number.parseInt();`, []string{"missingParameters"}},
	}

	for _, raw := range []string{"", `"always"`, `"as-needed"`} {
		options, err := DecodeRadixOptions([]byte(raw))
		if err != nil {
			t.Fatalf("the decoder refused %q: %v", raw, err)
		}
		for _, testCase := range cases {
			t.Run(raw+" "+testCase.sourceText, func(t *testing.T) {
				t.Parallel()
				rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, Radix, radixFile,
					testCase.sourceText, options), testCase.messages...)
			})
		}
	}
}

// The decoder refuses everything upstream's schema refuses, `{enum: ["always", "as-needed"]}`.
//
// `null` is the case worth naming: it unmarshals into a Go string as the empty one without an
// error, so it is the enum check rather than the JSON decode that turns it away. A list is refused
// because the config layer hands Decode one element, never the list around it.
func TestDecodeRadixOptionsRefusesWhatTheSchemaRefuses(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		`"never"`, `"Always"`, `"as_needed"`, `""`, `null`, `10`, `true`, `{}`,
		`{"mode": "always"}`, `["always"]`,
	} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeRadixOptions([]byte(raw)); err == nil {
				t.Errorf("expected %s to be refused, it was accepted", raw)
			}
		})
	}

	for raw, want := range map[string]RadixMode{`"always"`: RadixAlways, `"as-needed"`: RadixAsNeeded} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeRadixOptions([]byte(raw))
			if err != nil {
				t.Fatalf("expected %s to be accepted, got %v", raw, err)
			}
			if decoded != want {
				t.Errorf("expected %s to decode to %q, got %#v", raw, want, decoded)
			}
		})
	}
}
