package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// upstreamPasses is oxc's pass list, copied byte for byte and verified against the Rust source
// with a script rather than by eye. The tab-and-space runs are upstream's own indentation inside a
// Rust multi-line string literal and are load-bearing: rewrapping them would move every span.

var upstreamPasses = []string{
	"\n\t\t\t        import { forwardRef } from 'react'\n\t\t\t        forwardRef((props, ref) => {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import { forwardRef } from 'react'\n\t\t\t        forwardRef((props, ref) => null);\n\t\t\t      ",
	"\n\t\t\t        import { forwardRef } from 'react'\n\t\t\t        forwardRef(function (props, ref) {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import { forwardRef } from 'react'\n\t\t\t        forwardRef(function Component(props, ref) {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        React.forwardRef((props, ref) => {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        React.forwardRef((props, ref) => null);\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        React.forwardRef(function (props, ref) {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        React.forwardRef(function Component(props, ref) {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        function Component(props) {\n\t\t\t          return null;\n\t\t\t        };\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        (props) => null;\n\t\t\t      ",
	"forwardRef(() => {})",
	"forwardRef(function () {})",
	"forwardRef(function (a, b, c) {})",
}

// upstreamFails is oxc's fail list. The snapshot carries five diagnostics against five inputs, so
// each of these reports exactly once.

var upstreamFails = []string{
	"\n\t\t\t        import { forwardRef } from 'react'\n\t\t\t        forwardRef((props) => {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import { forwardRef } from 'react'\n\t\t\t        forwardRef(props => {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        React.forwardRef((props) => null);\n\t\t\t      ",
	"\n\t\t\t        import { forwardRef } from 'react'\n\t\t\t        const Component = forwardRef(function (props) {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
	"\n\t\t\t        import * as React from 'react'\n\t\t\t        React.forwardRef(function Component(props) {\n\t\t\t          return null;\n\t\t\t        });\n\t\t\t      ",
}

// TestForwardRefUsesRefFiresOnUpstreamCorpus runs oxc's five fail inputs. The snapshot gives one
// diagnostic per input, so each asserts exactly one id.
func TestForwardRefUsesRefFiresOnUpstreamCorpus(t *testing.T) {
	t.Parallel()

	for _, source := range upstreamFails {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
	}
}

// TestForwardRefUsesRefStaysSilentOnUpstreamCorpus runs oxc's thirteen pass inputs.
//
// Three of them are the whole reason to copy rather than invent. `forwardRef((props, ref) => null)`
// passes while never reading `ref`, which is what proves the rule counts parameters instead of
// asking whether the ref is used. `forwardRef(() => {})` and `forwardRef(function (a, b, c) {})`
// bracket the arity check on both sides, so a port written as "fewer than two parameters" fails the
// first and a port written as "not exactly two" fails the second.
func TestForwardRefUsesRefStaysSilentOnUpstreamCorpus(t *testing.T) {
	t.Parallel()

	for _, source := range upstreamPasses {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectClean(t, result)
	}
}

// TestForwardRefUsesRefPointsAtTheFunction asserts where the finding lands, which message ids
// cannot see. Upstream labels the function expression rather than the call, so a port reporting the
// call node passes every id assertion above while pointing at the wrong span on every finding.
func TestForwardRefUsesRefPointsAtTheFunction(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source string
		want   string
	}{
		{"React.forwardRef((props) => null);", "(props) => null"},
		{"const Component = forwardRef(function (props) { return null; });",
			"function (props) { return null; }"},
		{"forwardRef(props => null);", "props => null"},
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", testCase.source)
		rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
		reported := testCase.source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != testCase.want {
			t.Errorf("reported %q, want %q", reported, testCase.want)
		}
	}
}

// TestForwardRefUsesRefRendersItsMessage asserts the rendered description exactly rather than by
// substring, because a weaker predicate than the property it guards is not a guard.
func TestForwardRefUsesRefRendersItsMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", "forwardRef((props) => null);")
	rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
	if result.Diagnostics[0].Message.Description != messageForwardRefUsesRef.Description {
		t.Errorf("description was %q", result.Diagnostics[0].Message.Description)
	}
}

// applySuggestion returns what the source becomes when one suggestion's fixes are applied.
func applySuggestion(source string, suggestion rule.Suggestion) string {
	out := source
	for index := len(suggestion.Fixes) - 1; index >= 0; index-- {
		fix := suggestion.Fixes[index]
		out = out[:fix.Range.Pos()] + fix.Text + out[fix.Range.End():]
	}
	return out
}

// TestForwardRefUsesRefOffersUpstreamRepairs is the fix-vector block from oxc's own tester,
// asserted by applying each suggestion and comparing the resulting source.
//
// Asserting the repair rather than the id is not optional here. A rule that reports correctly while
// rewriting the wrong span satisfies every id assertion in this file, and upstream ships eight
// vectors precisely because the parameter-list rewrite has three separate normalizations in it: a
// stripped trailing paren, a stripped trailing comma with whatever whitespace follows it, and an
// inserted opening paren for a bare arrow parameter.
//
// `wantRemove` empty means upstream offers only the add-ref repair, which happens for an anonymous
// function expression in statement context, where unwrapping would not parse.
func TestForwardRefUsesRefOffersUpstreamRepairs(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source     string
		wantRemove string
		wantAddRef string
	}{
		{"forwardRef((a) => {})", "(a) => {}", "forwardRef((a, ref) => {})"},
		// Whitespace between the call's `(` and the function is the function node's leading
		// trivia, and upstream's `replace_with` writes the node's own span, so the unwrapped
		// result drops it. Confirmed by running oxlint with suggestions applied, which produced
		// `const x = function(a) {};` from both of these. A survivor on the replacement-text
		// trim is what showed no fixture could see it.
		{"const x = forwardRef(  function(a) {})", "const x = function(a) {}",
			"const x = forwardRef(  function(a, ref) {})"},
		{"const y = forwardRef(\n  function(a) {})", "const y = function(a) {}",
			"const y = forwardRef(\n  function(a, ref) {})"},
		{"forwardRef(a => {})", "a => {}", "forwardRef((a, ref) => {})"},
		{"forwardRef(function Component(a) {})", "function Component(a) {}",
			"forwardRef(function Component(a, ref) {})"},
		{"const x = forwardRef(function(a) {})", "const x = function(a) {}",
			"const x = forwardRef(function(a, ref) {})"},
		{"forwardRef(function (a) {})", "", "forwardRef(function (a, ref) {})"},
		{"forwardRef(function(a,) {})", "", "forwardRef(function(a, ref) {})"},
		{"forwardRef(function(a, ) {})", "", "forwardRef(function(a, ref) {})"},
		{"React.forwardRef(function(a) {})", "", "React.forwardRef(function(a, ref) {})"},
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", testCase.source)
		rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
		suggestions := result.Diagnostics[0].Suggestions

		wantCount := 1
		if testCase.wantRemove != "" {
			wantCount = 2
		}
		if len(suggestions) != wantCount {
			t.Errorf("%q offered %d suggestions, want %d", testCase.source, len(suggestions), wantCount)
			continue
		}

		addRef := suggestions[0]
		if testCase.wantRemove != "" {
			// Order matters: an engine applying one suggestion unattended takes the first, and
			// upstream's first is the removal.
			if suggestions[0].Message.Id != "forwardRefRemoveWrapper" {
				t.Errorf("%q offered %q first, want the removal", testCase.source, suggestions[0].Message.Id)
			}
			if got := applySuggestion(testCase.source, suggestions[0]); got != testCase.wantRemove {
				t.Errorf("%q removal produced %q, want %q", testCase.source, got, testCase.wantRemove)
			}
			addRef = suggestions[1]
		}
		if addRef.Message.Id != "forwardRefAddRefParameter" {
			t.Errorf("%q add-ref suggestion was %q", testCase.source, addRef.Message.Id)
		}
		if got := applySuggestion(testCase.source, addRef); got != testCase.wantAddRef {
			t.Errorf("%q add-ref produced %q, want %q", testCase.source, got, testCase.wantAddRef)
		}
	}
}

// TestForwardRefUsesRefMatchesTheCalleeNameOnly pins the recognition surface, all of it measured on
// the release oxlint binary because the imported corpus writes only `forwardRef` and
// `React.forwardRef`.
//
// The point of the reporting half is that upstream applies no receiver check at all, so a port
// reaching for `react.IsNamespacedMember`, whose name matches this question, silences four of these
// five. The point of the silent half is that there is no import tracking either, so an alias is
// invisible in the other direction.
func TestForwardRefUsesRefMatchesTheCalleeNameOnly(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"NotReact.forwardRef(function (e) {});",
		"obj.deep.forwardRef(function (i) {});",
		"React[\"forwardRef\"](function (d) {});",
		// A no-substitution template is a static name upstream too. This was a real defect in
		// this port, found by a surviving mutant on the element-access branch and confirmed
		// against oxlint, which reports it.
		"React[`forwardRef`](function (a) {});",
		"(React).forwardRef(function (c) {});",
		"forwardRef(function (h) {}, extra);",
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
	}

	for _, source := range []string{
		// An aliased import is silent, because the name read is the one at the call site.
		"import { forwardRef as fr } from 'react';\nfr(function (f) {});",
		"forwardRef;",
		"forwardRef();",
		"forwardRef(x);",
		"React.forwardRefs(function (a) {});",
		// A computed subscript that is not a static string answers no name, and the kind check
		// is the whole guard: `Text()` on this identifier returns `forwardRef`, so dropping the
		// check makes this report while oxlint is silent on it.
		"React[forwardRef](function (a) {});",
		"React[key](function (a) {});",
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectClean(t, result)
	}
}

// TestForwardRefUsesRefDeclinesParenthesesAtTheCalleeAndArgument is the paren measurement, and it
// is here because no imported fixture in either corpus writes a parenthesized anything, so a port
// guessing either way costs nothing at fixture time and ships a divergence.
//
// Upstream destructures the callee and the first argument directly, with no paren skipping, so
// wrapping either one makes the rule blind. All four were run against the release binary.
func TestForwardRefUsesRefDeclinesParenthesesAtTheCalleeAndArgument(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"(forwardRef)(function (b) {});",
		"(React.forwardRef)(function (a) {});",
		"forwardRef((function (g) {}));",
		"forwardRef(((a) => {}));",
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectClean(t, result)
	}
}

// TestForwardRefUsesRefSkipsParenthesesWalkingToTheStatement is the other half of the paren
// question and it falls the opposite way, which is why both exist.
//
// The statement-context test runs through `outermost_paren_parent`, which does see through parens.
// So a parenthesized call in statement position is still statement context and still withholds the
// removal suggestion. Measured by running oxlint with suggestions applied: the source came back as
// `(forwardRef(function (a, ref) {}));`, the add-ref repair, rather than unwrapped.
func TestForwardRefUsesRefSkipsParenthesesWalkingToTheStatement(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", "(forwardRef(function (a) {}));")
	rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
	if len(result.Diagnostics[0].Suggestions) != 1 {
		t.Fatalf("offered %d suggestions, want only the add-ref one",
			len(result.Diagnostics[0].Suggestions))
	}
	if result.Diagnostics[0].Suggestions[0].Message.Id != "forwardRefAddRefParameter" {
		t.Errorf("offered %q", result.Diagnostics[0].Suggestions[0].Message.Id)
	}
}

// TestForwardRefUsesRefCountsParametersTheWayOxcDoes covers the two places our parameter model and
// oxc's differ, in opposite directions, neither of which any imported fixture writes.
//
// A rest element is outside oxc's parameter list and inside ours, and upstream is silent on it
// through a separate `rest.is_some()` guard. A `this` parameter is inside ours and outside oxc's,
// and upstream reports. Counting `Parameters()` directly gets both wrong. Both measured on oxlint.
func TestForwardRefUsesRefCountsParametersTheWayOxcDoes(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"forwardRef(function (...a) {});",
		"forwardRef((a, ...b) => {});",
		"forwardRef((...a) => {});",
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectClean(t, result)
	}

	// A `this` parameter is not counted, so `a` is the only real one and this reports.
	result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", "forwardRef(function (this: any, a) {});")
	rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")

	// One parameter is the whole reportable band. A destructured or defaulted parameter is still
	// one parameter, so these report rather than being treated as a props-plus-ref shape.
	for _, source := range []string{
		"forwardRef(function ({a}) {});",
		"forwardRef(function ([a]) {});",
		"forwardRef(function (a = 1) {});",
		"forwardRef(async function (a) {});",
		"forwardRef(async (a) => {});",
		"forwardRef(function* (a) {});",
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
	}
}

// TestForwardRefUsesRefDeclinesNonFunctionArguments covers the shapes that are not a function at
// all, including the spread that upstream declines through `as_expression` returning nothing.
func TestForwardRefUsesRefDeclinesNonFunctionArguments(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"forwardRef(...x);",
		"forwardRef(class { m(a) {} });",
		"forwardRef({a: 1});",
		"forwardRef(forwardRef);",
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", source)
		rule_testing.ExpectClean(t, result)
	}
}

// TestForwardRefUsesRefWithholdsRemovalOnlyForAnonymousStatements pins the branch upstream guards
// and the wart it does not.
//
// Anonymous in statement position withholds the removal, because `function (a) {}` as a statement
// does not parse. Named in statement position does *not* withhold it, even though the result is a
// hoisted declaration rather than an expression, which is a meaning change beyond the intended one.
// That is upstream's, reproduced rather than improved on, and it is tolerable only because a
// suggestion is read by a human before it is applied. Measured by running oxlint with suggestions
// applied, which rewrote the named case to `function Named(a) {};`.
func TestForwardRefUsesRefWithholdsRemovalOnlyForAnonymousStatements(t *testing.T) {
	t.Parallel()

	anonymousStatement := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", "forwardRef(function (a) {});")
	rule_testing.ExpectFindings(t, anonymousStatement, "forwardRefUsesRef")
	if len(anonymousStatement.Diagnostics[0].Suggestions) != 1 {
		t.Errorf("anonymous statement offered %d suggestions, want 1",
			len(anonymousStatement.Diagnostics[0].Suggestions))
	}

	namedStatement := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", "forwardRef(function Named(a) {});")
	rule_testing.ExpectFindings(t, namedStatement, "forwardRefUsesRef")
	if len(namedStatement.Diagnostics[0].Suggestions) != 2 {
		t.Fatalf("named statement offered %d suggestions, want 2",
			len(namedStatement.Diagnostics[0].Suggestions))
	}
	if got := applySuggestion("forwardRef(function Named(a) {});", namedStatement.Diagnostics[0].Suggestions[0]); got != "function Named(a) {};" {
		t.Errorf("named removal produced %q", got)
	}

	// An arrow is never guarded: unwrapping leaves a valid expression statement.
	arrowStatement := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", "forwardRef((a) => {});")
	rule_testing.ExpectFindings(t, arrowStatement, "forwardRefUsesRef")
	if len(arrowStatement.Diagnostics[0].Suggestions) != 2 {
		t.Errorf("arrow statement offered %d suggestions, want 2",
			len(arrowStatement.Diagnostics[0].Suggestions))
	}
}

// TestForwardRefUsesRefRewritesAwkwardParameterLists covers parameter-list spans the upstream fix
// vectors do not reach, where the range has to be recovered from source text because our tree has
// no node for the list itself.
func TestForwardRefUsesRefRewritesAwkwardParameterLists(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source     string
		wantAddRef string
	}{
		// A type annotation sits inside the parameter's own range, so scanning to the next `)`
		// after the last parameter does not stop early on the one inside a function type.
		{"forwardRef(function (a: (x: number) => void) {})",
			"forwardRef(function (a: (x: number) => void, ref) {})"},
		// Whitespace between the name and the list, and inside it.
		{"forwardRef(function Component ( a ) {})",
			"forwardRef(function Component ( a, ref) {})"},
		// A default value containing a paren.
		{"forwardRef(function (a = f()) {})", "forwardRef(function (a = f(), ref) {})"},
		// A newline inside the list.
		{"forwardRef(function (\n  a,\n) {})", "forwardRef(function (\n  a, ref) {})"},
		// A bare arrow parameter with a space before it. `Pos()` on the first parameter starts
		// at its leading trivia, so a span taken from it directly puts the inserted opening
		// paren before the space and writes `forwardRef(( a, ref) => {})`. oxc's own span
		// starts at the parameter: its diagnostic label for `const x = forwardRef( a => {});`
		// is offset 22 length 7, which is `a => {}` rather than ` a => {}`. The space stays
		// outside the replaced span, which is why it survives into the result.
		{"const x = forwardRef( a => {})", "const x = forwardRef( (a, ref) => {})"},
		{"const y = forwardRef(  b => {})", "const y = forwardRef(  (b, ref) => {})"},
	} {
		result := rule_testing.Run(t, ForwardRefUsesRef, "input.tsx", testCase.source)
		rule_testing.ExpectFindings(t, result, "forwardRefUsesRef")
		suggestions := result.Diagnostics[0].Suggestions
		addRef := suggestions[len(suggestions)-1]
		got := applySuggestion(testCase.source, addRef)
		if got != testCase.wantAddRef {
			t.Errorf("%q add-ref produced %q, want %q", testCase.source, got, testCase.wantAddRef)
		}
		if !strings.Contains(got, ", ref)") {
			t.Errorf("%q did not gain a ref parameter", testCase.source)
		}
	}
}
