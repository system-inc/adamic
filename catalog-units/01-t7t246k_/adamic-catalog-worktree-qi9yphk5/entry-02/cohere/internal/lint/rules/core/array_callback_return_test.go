package core

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// arrayCallbackReturnCorpusProvenance records where the imported cases came from and how they were
// checked, because a fixture table is only worth what its provenance is.
//
// Source: ESLint's own tests at `tests/lib/rules/array-callback-return.js`, extracted by hooking its
// `RuleTester` and capturing the case objects rather than by parsing the file, so no case was
// retyped and no escape passed through a shell. 97 clean and 133 reporting cases, 139 findings, 24
// suggestions.
//
// Checked twice before being trusted:
//
//	every extracted case was replayed against the INSTALLED eslint 10.8.1 through the Linter API,
//	  and all 230 reproduced their expected verdict with zero mismatches, which is what proves the
//	  extraction is faithful rather than merely well-formed
//	every Go literal in the corpus file was parsed back out and compared byte against byte with the
//	  extracted JSON, which is what catches a cooked escape or a smart quote
//
// Ten upstream errors carry only rendered message text rather than a `messageId`. Those were mapped
// back to their id by the sentence upstream's own `meta.messages` associates with it, which is a
// one-to-one mapping across the four ids this rule can report.
const arrayCallbackReturnCorpusProvenance = "eslint/tests/lib/rules/array-callback-return.js"

// TestArrayCallbackReturnStaysSilent runs every clean case upstream ships.
//
// These are the false positives upstream already thought about. They are the half of the corpus that
// catches a port matching too many callee shapes, which is the direction this rule fails in: the
// method-name set is permissive by design, so any widening of the argument-position or async tests
// reports on correct code while every reporting case stays green.
func TestArrayCallbackReturnStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range arrayCallbackReturnCleanCases {
		result := rule_testing.RunWithOptions(
			t, ArrayCallbackReturn, "clean.ts", testCase.source, testCase.options)
		if len(result.Diagnostics) != 0 {
			t.Errorf("reported %v on a clean case\n%s", result.MessageIds(), testCase.source)
		}
	}
}

// TestArrayCallbackReturnFires runs every reporting case upstream ships, asserting the message ids
// in order and their count.
//
// Order is asserted rather than just membership, and it is load-bearing here.
// `foo.every(function() { if (a) return; })` reports twice, and upstream orders the function-level
// finding before the per-return one because ESLint sorts its messages by position. A port emitting
// in traversal order gets the reverse and every count assertion still passes.
func TestArrayCallbackReturnFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range arrayCallbackReturnReportingCases {
		result := rule_testing.RunWithOptions(
			t, ArrayCallbackReturn, "fires.ts", testCase.source, testCase.options)
		got := result.MessageIds()
		if len(got) != len(testCase.wantIds) {
			t.Errorf("reported %v, want %v\n%s", got, testCase.wantIds, testCase.source)
			continue
		}
		for index, want := range testCase.wantIds {
			if got[index] != want {
				t.Errorf("finding %d was %q, want %q\n%s",
					index, got[index], want, testCase.source)
			}
		}
	}
}

// TestArrayCallbackReturnSuggestions applies every suggestion upstream asserts and compares the
// whole resulting source.
//
// This is the half of the port that no message-id fixture can see. A suggestion writing the right
// text over the wrong span passes a text comparison and is a real defect, so the repair is applied
// and the file compared rather than the fix inspected.
//
// It caught two things while being written. The `void` prefix needs a leading space when `return`
// and the expression are adjacent, because `returnvoid x` is one identifier, which upstream computes
// from its token stream and the corpus pins with `foo.forEach((x) => { return(x); })`. And a
// concise arrow under `checkForEach` offers one suggestion or two depending on `allowVoid`, so the
// count per finding is asserted rather than only the ids.
func TestArrayCallbackReturnSuggestions(t *testing.T) {
	t.Parallel()

	checked := 0
	for _, testCase := range arrayCallbackReturnReportingCases {
		if testCase.suggestions == nil {
			continue
		}
		result := rule_testing.RunWithOptions(
			t, ArrayCallbackReturn, "suggest.ts", testCase.source, testCase.options)
		written := result.SourceFile.Text()

		if len(result.Diagnostics) != len(testCase.suggestions) {
			t.Errorf("reported %d findings, want %d suggestion groups\n%s",
				len(result.Diagnostics), len(testCase.suggestions), testCase.source)
			continue
		}
		for index, want := range testCase.suggestions {
			got := result.Diagnostics[index].Suggestions
			if len(got) != len(want) {
				t.Errorf("finding %d offers %d suggestions, want %d\n%s",
					index, len(got), len(want), testCase.source)
				continue
			}
			for offset, wantSuggestion := range want {
				if got[offset].Message.Id != wantSuggestion.id {
					t.Errorf("suggestion %d is %q, want %q\n%s",
						offset, got[offset].Message.Id, wantSuggestion.id, testCase.source)
					continue
				}
				applied := applyArrayCallbackFixes(written, got[offset].Fixes)
				if applied != wantSuggestion.output {
					t.Errorf("applying %q produced\n  %q\nwant\n  %q\nfor %s",
						wantSuggestion.id, applied, wantSuggestion.output, testCase.source)
					continue
				}
				checked++
			}
		}
	}
	// The count is pinned so a corpus regeneration that silently drops the suggestion field fails
	// here rather than passing vacuously.
	if checked != 24 {
		t.Errorf("checked %d suggestions, want 24", checked)
	}
}

// applyArrayCallbackFixes applies a suggestion's edits to a source string.
//
// Applied back to front so an earlier edit does not shift the offsets a later one was computed
// against. `rule_testing` can apply a Fix but has no suggestion applier, so this is the hand-rolled half
// the brief warns to budget for.
func applyArrayCallbackFixes(source string, fixes []rule.Fix) string {
	ordered := make([]rule.Fix, len(fixes))
	copy(ordered, fixes)
	sort.SliceStable(ordered, func(left, right int) bool {
		return ordered[left].Range.Pos() > ordered[right].Range.Pos()
	})
	for _, fix := range ordered {
		source = source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]
	}
	return source
}

// TestArrayCallbackReturnSpan asserts WHERE each finding points.
//
// Upstream passes an explicit `loc: getFunctionHeadLoc(node, sourceCode)` for the function-level
// findings, so a callback body hundreds of lines long does not become the highlighted range. The
// expected columns below are upstream's own, taken from the `column` field its corpus asserts, and
// the spans are what those columns name.
//
// A port reporting on the whole function passes every id fixture in this file while highlighting the
// entire callback.
func TestArrayCallbackReturnSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name: "an arrow reports on its arrow token",
			// Upstream column 14, which is the `=>`.
			source: `foo.every(() => {})`,
			want:   []string{"=>"},
		},
		{
			name: "an anonymous function reports on its head",
			// Upstream column 11, the `function` keyword, through the parameter list.
			source: `foo.every(function() { if (a) return true; })`,
			want:   []string{"function()"},
		},
		{
			name:   "a named function keeps its name inside the span",
			source: `foo.every(function cb() { if (a) return true; })`,
			want:   []string{"function cb()"},
		},
		{
			name: "a callback returned from a wrapper reports at the inner function",
			// Upstream column 30, which is the inner function rather than the wrapper.
			source: `foo.every(function(){ return function() {}; }())`,
			want:   []string{"function()"},
		},
		{
			name:   "a bare return reports at the return statement, not the function",
			source: `foo.filter(function() { return; })`,
			want:   []string{"return;"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "span.ts", testCase.source, nil)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("reported %d findings, want %d: %v",
					len(result.Diagnostics), len(testCase.want), result.MessageIds())
			}
			written := result.SourceFile.Text()
			for index, want := range testCase.want {
				diagnostic := result.Diagnostics[index]
				got := written[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Errorf("finding %d spans %q, want %q", index, got, want)
				}
			}
		})
	}
}

// TestArrayCallbackReturnMessageText asserts the rendered text, exactly.
//
// The per-finding half names the method and the callback, and it is built with `Sprintf`, so a
// mutation moving only that text leaves every id and count fixed. The two values are upstream's own
// `arrayMethodName` and `name`, spelled the same way so a reader moving between the linters sees the
// same words. Asserted against literals typed here rather than against the rule's message constants,
// because comparing to the constant moves both sides together under mutation.
func TestArrayCallbackReturnMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "a prototype method is qualified through the prototype",
			source: `foo.every(function() {})`,
			want:   " Here Array.prototype.every() is being passed function.",
		},
		{
			name:   "a named function is named in single quotes",
			source: `foo.every(function cb() {})`,
			want:   " Here Array.prototype.every() is being passed function 'cb'.",
		},
		{
			name:   "an arrow is named by its kind",
			source: `foo.every(() => {})`,
			want:   " Here Array.prototype.every() is being passed arrow function.",
		},
		{
			name: "a static method is qualified on Array itself",
			// `Array.prototype.from` does not exist, so getting this wrong renders a method nobody
			// can look up.
			source: `Array.from(x, function() {})`,
			want:   " Here Array.from() is being passed function.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "text.ts", testCase.source, nil)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("reported %d findings, want 1", len(result.Diagnostics))
			}
			got := result.Diagnostics[0].Message.Description
			if !strings.HasSuffix(got, testCase.want) {
				t.Errorf("description tail is wrong\n got %q\nwant suffix %q", got, testCase.want)
			}
		})
	}

	t.Run("message ids are the four upstream names", func(t *testing.T) {
		t.Parallel()
		ids := []string{
			messageArrayCallbackExpectedAtEnd.Id,
			messageArrayCallbackExpectedInside.Id,
			messageArrayCallbackExpectedReturnValue.Id,
			messageArrayCallbackExpectedNoReturnValue.Id,
			messageArrayCallbackWrapBraces.Id,
			messageArrayCallbackPrependVoid.Id,
		}
		want := []string{
			"expectedAtEnd", "expectedInside", "expectedReturnValue",
			"expectedNoReturnValue", "wrapBraces", "prependVoid",
		}
		for index := range want {
			if ids[index] != want[index] {
				t.Errorf("message %d is %q, want %q", index, ids[index], want[index])
			}
		}
	})
}

// TestArrayCallbackReturnOptionsDecode routes every option through the rule's own exported decoder.
//
// Building the struct directly would leave the three json tags and the defaults untested, and those
// are the lines with no upstream counterpart. The nil case is the one that has shipped a broken rule
// in this tree before: a rule configured as a bare severity is handed nil, and a port relying on the
// zero value arriving by accident cannot tell that from a decoder that never ran.
func TestArrayCallbackReturnOptionsDecode(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[ArrayCallbackReturnOptions]()

	t.Run("allowImplicit accepts a bare return", func(t *testing.T) {
		t.Parallel()
		decoded, err := decode(json.RawMessage(`{"allowImplicit": true}`))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !decoded.(ArrayCallbackReturnOptions).AllowImplicit {
			t.Fatalf("decoded %#v", decoded)
		}
		result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts",
			`foo.filter(function() { return; })`, decoded)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("without allowImplicit a bare return reports", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts",
			`foo.filter(function() { return; })`, nil)
		rule_testing.ExpectFindings(t, result, "expectedReturnValue")
	})

	t.Run("checkForEach turns the judgment around", func(t *testing.T) {
		t.Parallel()
		decoded, err := decode(json.RawMessage(`{"checkForEach": true}`))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts",
			`foo.forEach(x => x)`, decoded)
		rule_testing.ExpectFindings(t, result, "expectedNoReturnValue")
	})

	t.Run("without checkForEach a forEach returning a value is clean", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts", `foo.forEach(x => x)`, nil)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("allowVoid accepts void and adds a second suggestion", func(t *testing.T) {
		t.Parallel()
		decoded, err := decode(json.RawMessage(`{"checkForEach": true, "allowVoid": true}`))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		clean := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts",
			`foo.forEach(x => void x)`, decoded)
		rule_testing.ExpectClean(t, clean)

		// The option changes what is OFFERED, not only what is reported, and only asserting the
		// finding would miss half of it.
		offered := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts",
			`foo.forEach(x => x)`, decoded)
		rule_testing.ExpectFindings(t, offered, "expectedNoReturnValue")
		if len(offered.Diagnostics[0].Suggestions) != 2 {
			t.Fatalf("offered %d suggestions under allowVoid, want 2",
				len(offered.Diagnostics[0].Suggestions))
		}
	})

	t.Run("nil options bypass the decoder and keep every default off", func(t *testing.T) {
		t.Parallel()
		// This is the shape a bare `"error"` produces. It reaches the rule without passing through
		// the decoder above, so no fixture routed through `decode` can see it.
		result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts", `foo.forEach(x => x)`, nil)
		rule_testing.ExpectClean(t, result)

		reports := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "o.ts",
			`foo.filter(function() { return; })`, nil)
		rule_testing.ExpectFindings(t, reports, "expectedReturnValue")
	})

	t.Run("an empty object decodes to all three defaults off", func(t *testing.T) {
		t.Parallel()
		decoded, err := decode(json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		settings := decoded.(ArrayCallbackReturnOptions)
		if settings.AllowImplicit || settings.CheckForEach || settings.AllowVoid {
			t.Fatalf("empty object did not default every option off: %#v", settings)
		}
	})
}

// TestArrayCallbackReturnReachability pins the judgment that needs the control-flow graph.
//
// "Can this callback run off its end" is not syntactic, and the pair below differs only in whether
// the `if` has an `else`. Upstream asks `isAnySegmentReachable` at the function's exit; this asks
// `control_flow_graph.Graph.EndReachable`, which is documented as answering that question.
//
// The third case is the one a naive "does the body end with a return" test gets wrong.
func TestArrayCallbackReturnReachability(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "a branch that can fall through reports",
			source: `foo.every(function() { if (a) return true; })`,
			want:   []string{"expectedAtEnd"},
		},
		{
			name:   "both branches returning is clean",
			source: `foo.every(function() { if (a) return true; else return false; })`,
		},
		{
			name: "a throw also ends the path",
			// Nothing here returns and the body still cannot fall off its end.
			source: `foo.every(function() { throw new Error('x'); })`,
		},
		{
			name:   "no return anywhere reports the other message",
			source: `foo.every(function() { doSomething(); })`,
			want:   []string{"expectedInside"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "reach.ts", testCase.source, nil)
			if len(testCase.want) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

// TestArrayCallbackReturnCalleeShapes pins which callee positions count.
//
// Each pair is a shape and its control, because the walk up from the callback is the half of this
// rule that decides whether to judge at all, and a widening there reports on correct code while
// every reporting fixture stays green.
func TestArrayCallbackReturnCalleeShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "a logical expression choosing a callback is walked through",
			source: `foo.every(cb || function() {})`,
			want:   []string{"expectedInside"},
		},
		{
			name:   "a conditional choosing a callback reports on each arm",
			source: `foo.every(a ? function() {} : function() {})`,
			want:   []string{"expectedInside", "expectedInside"},
		},
		{
			name:   "an immediately invoked wrapper's returned callback is judged",
			source: `foo.every(function(){ return function() {}; }())`,
			want:   []string{"expectedInside"},
		},
		{
			name: "a wrapper that is NOT called is not walked through",
			// The control for the row above. Without the trailing call the returned function is an
			// ordinary return value rather than the callback, so the inner function is not judged.
			//
			// The OUTER function is the callback here and it is clean, which is worth stating
			// because the intuitive reading is that it should report: it returns on its only path,
			// so nothing is missing. Measured against the installed rule, which is silent on it.
			source: `foo.every(function(){ return function() {}; })`,
		},
		{
			name:   "a parenthesized optional-chain callee still resolves",
			source: `(foo?.filter)(() => { console.log('hello') })`,
			want:   []string{"expectedInside"},
		},
		{
			name:   "a template-literal key names the method",
			source: "foo[`every`](function() {})",
			want:   []string{"expectedInside"},
		},
		{
			name: "a typed array constructor is an Array.from receiver",
			// Upstream's pattern is `/Array$/u`, any identifier ENDING in Array, so every typed
			// array constructor matches.
			source: `Int32Array.from(x, function() {})`,
			want:   []string{"expectedInside"},
		},
		{
			name: "a second argument to a prototype method is not the callback",
			// The control for the argument-position test: these methods take their callback first.
			source: `foo.every(cb, function() {})`,
		},
		{
			name: "a first argument to Array.from is not the callback",
			// The mirror control: Array.from takes its callback second.
			source: `Array.from(function() {})`,
		},
		{
			name:   "a method outside the set is not checked",
			source: `foo.forEachEntry(function() {})`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "callee.ts", testCase.source, nil)
			if len(testCase.want) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

// TestArrayCallbackReturnAsyncAndGenerator pins the asymmetry upstream ships deliberately.
//
// A generator is never checked. An async function is checked only for `Array.fromAsync`, because
// upstream's `!node.async` guard wraps the `Array.from` and prototype-method arms and leaves the
// `fromAsync` arm outside it. Measured against the installed rule in all four directions.
func TestArrayCallbackReturnAsyncAndGenerator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "an async callback to a prototype method is not checked",
			source: `foo.map(async function() {})`,
		},
		{
			name:   "a generator callback is never checked",
			source: `foo.every(function*() {})`,
		},
		{
			name:   "an async callback to Array.fromAsync IS checked",
			source: `Array.fromAsync(x, async function() {})`,
			want:   []string{"expectedInside"},
		},
		{
			name: "a plain callback to a prototype method is checked",
			// The control that keeps the three rows above from passing on a rule that checks
			// nothing at all.
			source: `foo.map(function() {})`,
			want:   []string{"expectedInside"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayCallbackReturn, "async.ts", testCase.source, nil)
			if len(testCase.want) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

// TestArrayCallbackReturnVoidParenthesization pins a branch the imported corpus cannot reach.
//
// Upstream parenthesizes when prepending `void` to an expression that binds looser than a unary
// operator, because `void x + 1` is `(void x) + 1` and means something else. Every `prependVoid` case
// in upstream's corpus is either already parenthesized or a tight unary, so nothing there
// distinguishes a port that always parenthesizes, never parenthesizes, or gets it right. A mutant
// forcing the branch off survived the whole corpus.
//
// The expectations below are measured rather than reasoned: each was produced by applying the
// installed rule's own `prependVoid` fix to that source through the Linter API, since the corpus's
// `output` field is not populated on this path.
//
//	foo.forEach((x) => { return x + 1; })      -> return void (x + 1);
//	foo.forEach(x => x + 1)                    -> x => void (x + 1)
//	foo.forEach((x) => { return a ? b : c; })  -> return void (a ? b : c);
//	foo.forEach((x) => { return x; })          -> return void x;
//
// The last row is the control: a tight expression must NOT gain parentheses, which is what keeps a
// port that always wraps from passing the first three.
func TestArrayCallbackReturnVoidParenthesization(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[ArrayCallbackReturnOptions]()
	options, err := decode(json.RawMessage(`{"checkForEach": true, "allowVoid": true}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "a binary expression gains parentheses",
			source: `foo.forEach((x) => { return x + 1; })`,
			want:   `foo.forEach((x) => { return void (x + 1); })`,
		},
		{
			name:   "a concise arrow body gains them too",
			source: `foo.forEach(x => x + 1)`,
			want:   `foo.forEach(x => void (x + 1))`,
		},
		{
			name:   "a conditional gains them",
			source: `foo.forEach((x) => { return a ? b : c; })`,
			want:   `foo.forEach((x) => { return void (a ? b : c); })`,
		},
		{
			name: "a tight expression does not",
			// The control. Without it, a port that always parenthesizes passes every row above.
			source: `foo.forEach((x) => { return x; })`,
			want:   `foo.forEach((x) => { return void x; })`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(
				t, ArrayCallbackReturn, "void.ts", testCase.source, options)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("reported %d findings, want 1", len(result.Diagnostics))
			}
			var prependVoid *rule.Suggestion
			for index := range result.Diagnostics[0].Suggestions {
				if result.Diagnostics[0].Suggestions[index].Message.Id == "prependVoid" {
					prependVoid = &result.Diagnostics[0].Suggestions[index]
				}
			}
			if prependVoid == nil {
				t.Fatalf("no prependVoid suggestion was offered")
			}
			applied := applyArrayCallbackFixes(result.SourceFile.Text(), prependVoid.Fixes)
			if applied != testCase.want {
				t.Errorf("applying prependVoid produced\n  %q\nwant\n  %q", applied, testCase.want)
			}
		})
	}
}
