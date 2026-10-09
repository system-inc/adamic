package typescript

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const nonNullAssertionFile = "/repository/source/Thing.ts"

func TestNoNonNullAssertionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		count      int
	}{
		{"a bare assertion", "declare const foo: any;\nexport const r = foo!;\n", 1},
		{"an assertion before a member", "declare const foo: any;\nexport const r = foo!.bar;\n", 1},
		{"an assertion before a computed member", "declare const foo: any;\ndeclare const k: any;\nexport const r = foo![k];\n", 1},
		{"an assertion before a call", "declare const foo: any;\nexport const r = foo!();\n", 1},
		{"an assertion before an optional member", "declare const foo: any;\nexport const r = foo!?.bar;\n", 1},
		{"an assertion before an optional call", "declare const foo: any;\nexport const r = foo!?.();\n", 1},
		{"an assertion in an argument", "declare const foo: any;\ndeclare const x: any;\nexport const r = foo(x!);\n", 1},
		{"an assertion in a computed index", "declare const foo: any;\ndeclare const k: any;\nexport const r = foo[k!];\n", 1},
		// Every assertion is its own finding, including redundant ones. This rule reports the
		// construct; whether one duplicates another is a different rule's judgment.
		{"two assertions", "declare const foo: any;\nexport const r = foo!!;\n", 2},
		{"three assertions", "declare const foo: any;\nexport const r = foo!!!;\n", 3},
		// The assignment shapes still report. Only the suggestion is withheld, which the suggestion
		// test below pins separately, so a rule that suppressed the finding rather than the repair
		// fails here.
		{"an assertion on an assignment target", "declare const foo: any;\nfoo!.bar = 1;\nexport const r = foo;\n", 1},
		{"an assertion under delete", "declare const foo: any;\ndelete foo!.bar;\nexport const r = foo;\n", 1},
		{"an assertion under increment", "declare const foo: any;\nfoo!.bar++;\nexport const r = foo;\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNonNullAssertion, nonNullAssertionFile, testCase.sourceText)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "noNonNull"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestNoNonNullAssertionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// Definite assignment. The same character, a different construct, and the one shape most
		// likely to be caught by a rule that reached for the token rather than the node.
		{"a definite assignment on a variable", "export let value!: number;\n"},
		{"a definite assignment on a property", "export class Thing {\n    value!: number;\n}\n"},
		// Type positions. `NonNullable<T>` is the type-level spelling of the same intent and is not
		// this rule's business.
		{"a NonNullable type", "declare const thing: string | null;\nexport const r: NonNullable<typeof thing> = 'x';\n"},
		{"a non-null type operator in a generic", "export type Thing<T> = NonNullable<T>;\n"},
		// An optional chain is the repair this rule suggests, so it must not itself report.
		{"an optional chain", "declare const foo: any;\nexport const r = foo?.bar;\n"},
		{"an optional call", "declare const foo: any;\nexport const r = foo?.();\n"},
		// Negation and inequality also spell the character, in quantity, and are the noise a
		// text-based detector drowns in. A count over this tree once returned 376 from a regex and
		// zero from a parser.
		{"a logical negation", "declare const foo: boolean;\nexport const r = !foo;\n"},
		{"a double negation", "declare const foo: any;\nexport const r = !!foo;\n"},
		{"an inequality", "declare const a: number;\ndeclare const b: number;\nexport const r = a !== b;\n"},
		{"an exclamation in a string", "export const r = 'hello!';\n"},
		{"an exclamation in a template", "export const r = `hello!${1}`;\n"},
		// A GraphQL non-null marker inside a string literal, which is what the tree this gates is
		// actually full of.
		{"a graphql non-null marker", "export const query = 'mutation M($input: Thing!) { m(input: $input) { id } }';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoNonNullAssertion, nonNullAssertionFile, testCase.sourceText))
		})
	}
}

// TestNoNonNullAssertionSuggestions pins which positions get a repair and what it rewrites to.
//
// The finding is unconditional, so a defect in the suggestion logic is invisible to the fixture pair
// above: every case there passes whether or not a suggestion is attached, and whether or not the
// suggestion would parse. This is the guard for the half that ships silently.
func TestNoNonNullAssertionSuggestions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a dot access", "declare const foo: any;\nexport const r = foo!.bar;\n", "foo?.bar"},
		{"a computed access", "declare const foo: any;\ndeclare const k: any;\nexport const r = foo![k];\n", "foo?.[k]"},
		{"a call", "declare const foo: any;\nexport const r = foo!();\n", "foo?.()"},
		{"an already-optional access", "declare const foo: any;\nexport const r = foo!?.bar;\n", "foo?.bar"},
		{"an already-optional call", "declare const foo: any;\nexport const r = foo!?.();\n", "foo?.()"},
		// A comment between the assertion and the dot. The repair edits the two tokens separately,
		// so whatever sits between them survives. Rewriting the span would eat the comment.
		{"a comment between the tokens", "declare const foo: any;\nexport const r = foo!\n    // why\n    .bar;\n", "foo\n    // why\n    ?.bar"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNonNullAssertion, nonNullAssertionFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 {
				t.Fatalf("want 1 suggestion, got %d", len(suggestions))
			}
			got := applyFixes(testCase.sourceText, suggestions[0].Fixes)
			if !containsLine(got, testCase.want) {
				t.Fatalf("want repaired source to contain %q, got:\n%s", testCase.want, got)
			}
		})
	}
}

// TestNoNonNullAssertionWithholdsSuggestionsThatWouldNotParse is the reason isAssignee was ported
// whole rather than by its common case.
//
// Each of these is a position where `?.` is a syntax error: an optional chain may not be assigned
// to, deleted, or incremented. A rule that offered a repair here would produce source that does not
// parse, which is worse than offering nothing because the suggestion reads as checked.
//
// The reference implementation consulted for this port recognizes only the first of these.
func TestNoNonNullAssertionWithholdsSuggestionsThatWouldNotParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an assignment target", "declare const foo: any;\nfoo!.bar = 1;\nexport const r = foo;\n"},
		{"a compound assignment target", "declare const foo: any;\nfoo!.bar += 1;\nexport const r = foo;\n"},
		{"a delete argument", "declare const foo: any;\ndelete foo!.bar;\nexport const r = foo;\n"},
		{"a postfix increment", "declare const foo: any;\nfoo!.bar++;\nexport const r = foo;\n"},
		{"a prefix decrement", "declare const foo: any;\n--foo!.bar;\nexport const r = foo;\n"},
		{"an array destructuring target", "declare const foo: any;\n[foo!.bar] = [0];\nexport const r = foo;\n"},
		{"a rest destructuring target", "declare const foo: any;\n[...foo!.bar] = [0];\nexport const r = foo;\n"},
		{"an object destructuring target", "declare const foo: any;\n({ k: foo!.bar } = { k: 0 });\nexport const r = foo;\n"},
		{"a target behind a type assertion", "declare const foo: any;\n(foo!.bar as number)++;\nexport const r = foo;\n"},
		{"a target behind a further assertion", "declare const foo: any;\n[...foo!.bar!] = [0];\nexport const r = foo;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNonNullAssertion, nonNullAssertionFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("want a finding, got none")
			}
			// The assertion nearest the target is the one whose suggestion would not parse. Later
			// assertions in the same statement are a different position and may carry one.
			if suggestions := result.Diagnostics[0].Suggestions; len(suggestions) != 0 {
				t.Fatalf("want no suggestion on an assignment target, got %d", len(suggestions))
			}
		})
	}
}

// TestNoNonNullAssertionWithholdsSuggestionsOutsideTheObjectPosition pins the check that an
// assertion must occupy the position the access reaches *through* before a repair is offered.
//
// An assertion inside a computed index or an argument sits under a member access or a call, so a
// guard that looked only at the parent's kind would offer to rewrite it. The rewrite is nonsense:
// `foo[k!]` does not become `foo?.[k]`, it becomes a different expression that indexes something
// else. Mutation testing found this check unpinned by every other fixture here, which is why it has
// its own.
func TestNoNonNullAssertionWithholdsSuggestionsOutsideTheObjectPosition(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an assertion in a computed index", "declare const foo: any;\ndeclare const k: any;\nexport const r = foo[k!];\n"},
		{"an assertion in an optional computed index", "declare const foo: any;\ndeclare const k: any;\nexport const r = foo?.[k!];\n"},
		{"an assertion in a call argument", "declare const foo: any;\ndeclare const x: any;\nexport const r = foo(x!);\n"},
		{"an assertion in an optional call argument", "declare const foo: any;\ndeclare const x: any;\nexport const r = foo?.(x!);\n"},
		// A bare assertion with nothing after it. There is no `?.` that means the same thing, so
		// the original attaches no suggestion either.
		{"a bare assertion", "declare const foo: any;\nexport const r = foo!;\n"},
		{"an assertion in a return position", "declare const foo: any;\nexport function get() {\n    return foo!;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNonNullAssertion, nonNullAssertionFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("want a finding, got none")
			}
			if suggestions := result.Diagnostics[0].Suggestions; len(suggestions) != 0 {
				t.Fatalf("want no suggestion outside the object position, got %d", len(suggestions))
			}
		})
	}
}

// applyFixes rewrites source with a set of proposed fixes, applying them back to front so that an
// earlier edit does not shift the offsets a later one was computed against.
func applyFixes(sourceText string, fixes []rule.Fix) string {
	ordered := make([]rule.Fix, len(fixes))
	copy(ordered, fixes)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Range.Pos() > ordered[j].Range.Pos() })
	for _, fix := range ordered {
		sourceText = sourceText[:fix.Range.Pos()] + fix.Text + sourceText[fix.Range.End():]
	}
	return sourceText
}

func containsLine(haystack string, needle string) bool {
	return strings.Contains(haystack, needle)
}

// TestNoNonNullAssertionMessageGivesTheRepairsInOrder pins the wording's shape rather than its every
// word: the finding names the four repairs in the order the house's migration waves ranked them, and
// the suggestion offers `?.` only with its condition attached. Upstream's text recommended `?.` as a
// repair outright, and a reworded message that drifted back to that would pass an id-only test.
func TestNoNonNullAssertionMessageGivesTheRepairsInOrder(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoNonNullAssertion, nonNullAssertionFile, "declare const foo: any;\nexport const r = foo!.bar;\n")
	if len(result.Diagnostics) != 1 || len(result.Diagnostics[0].Suggestions) != 1 {
		t.Fatalf("want 1 finding with 1 suggestion, got %d findings", len(result.Diagnostics))
	}
	noNonNullAssertionExpectInOrder(t, result.Diagnostics[0].Message.Description,
		"guard it", "fix the type", "restructure so the index goes away", "`required(value, 'why it holds')` from `@nexus/source/errors/Assert`")
	noNonNullAssertionExpectInOrder(t, result.Diagnostics[0].Suggestions[0].Message.Description,
		"`?.`", "only where a missing value is really possible")
}

// noNonNullAssertionExpectInOrder fails unless each part appears in text after the one before it.
func noNonNullAssertionExpectInOrder(t *testing.T, text string, parts ...string) {
	t.Helper()
	rest := text
	for _, part := range parts {
		index := strings.Index(rest, part)
		if index < 0 {
			t.Fatalf("message lacks %q after the parts before it:\n%s", part, text)
		}
		rest = rest[index+len(part):]
	}
}
