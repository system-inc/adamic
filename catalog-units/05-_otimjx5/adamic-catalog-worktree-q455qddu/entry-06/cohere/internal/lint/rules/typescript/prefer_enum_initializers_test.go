package typescript

import (
	"sort"
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// preferEnumInitializersFile names the fixture file. The rule reads no path and gates on no
// extension: upstream registers a bare TSEnumDeclaration visitor with no source-type test.
const preferEnumInitializersFile = "/repository/source/Direction.ts"

// preferEnumInitializersCaseName numbers a row so a failure names which one.
func preferEnumInitializersCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// preferEnumInitializersSuggestion is one offered repair: what it says, and what it writes.
//
// Both halves are asserted because they can disagree. The description is interpolated text and the
// applied source is a range plus a replacement, so a suggestion can name the right value and write
// it over the wrong span, which reads as correct in every message assertion.
type preferEnumInitializersSuggestion struct {
	desc    string
	applied string
}

type preferEnumInitializersFinding struct {
	wantSpan        string
	wantMessage     string
	wantSuggestions []preferEnumInitializersSuggestion
}

// applyPreferEnumInitializersSuggestion rewrites source with one suggestion's fixes.
//
// The harness has no suggestion support at all: ExpectFixedSource covers fixes, and a rule shipping
// only suggestions has nothing that applies them. So this is hand-rolled rather than skipped, since
// asserting the message id of a rule whose entire repair surface is three suggestions would leave
// every one of them unchecked.
//
// Back to front so an earlier edit cannot move a later one's offsets, matching what the fix engine
// does. Each suggestion here carries exactly one fix, so the ordering is not load-bearing today; it
// is written correctly anyway rather than relying on that staying true.
func applyPreferEnumInitializersSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	fixes := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.Slice(fixes, func(first, second int) bool { return fixes[first].Range.Pos() > fixes[second].Range.Pos() })
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("the suggestion proposes an out-of-range edit [%d,%d) over %d bytes", start, end, len(source))
		}
		source = source[:start] + fix.Text + source[end:]
	}
	return source
}

// TestPreferEnumInitializersStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All four of upstream's passing inputs, extracted from the clone's test file by parsing it with the
// TypeScript compiler rather than by reading it. Every one was additionally run through the
// installed 8.x build, which reported nothing and produced no parse error on any of them.
func TestPreferEnumInitializersStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nenum Direction {}\n    ",
		"\nenum Direction {\n  Up = 1,\n}\n    ",
		"\nenum Direction {\n  Up = 1,\n  Down = 2,\n}\n    ",
		"\nenum Direction {\n  Up = 'Up',\n  Down = 'Down',\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(preferEnumInitializersCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, PreferEnumInitializers,
				preferEnumInitializersFile, sourceText))
		})
	}
}

// TestPreferEnumInitializersFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// Four inputs carrying five findings and fifteen suggestion outputs between them. Every field is
// measured rather than transcribed: the span is upstream's reported line and column range converted
// to an offset slice of the same input, the message is what the installed build rendered, and each
// suggestion's applied source is that build's own fix range and text replayed against the input.
//
// The fifteen applied results were separately confirmed equal to the `output` fields the corpus
// records, so the running rule and the checked-in corpus agree and there is no drift to record.
//
// All three assertion layers run on every row. A finding at the right span carrying a suggestion
// that writes the wrong text passes a message-id check, and a suggestion naming the right value
// while replacing the wrong range passes a description check.
func TestPreferEnumInitializersFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []preferEnumInitializersFinding
	}{
		{
			sourceText: "\nenum Direction {\n  Up,\n}\n      ",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "\nenum Direction {\n  Up = 0,\n}\n      "},
						{desc: "Can be fixed to Up = 1", applied: "\nenum Direction {\n  Up = 1,\n}\n      "},
						{desc: "Can be fixed to Up = 'Up'", applied: "\nenum Direction {\n  Up = 'Up',\n}\n      "},
					},
				},
			},
		},
		{
			sourceText: "\nenum Direction {\n  Up,\n  Down,\n}\n      ",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "\nenum Direction {\n  Up = 0,\n  Down,\n}\n      "},
						{desc: "Can be fixed to Up = 1", applied: "\nenum Direction {\n  Up = 1,\n  Down,\n}\n      "},
						{desc: "Can be fixed to Up = 'Up'", applied: "\nenum Direction {\n  Up = 'Up',\n  Down,\n}\n      "},
					},
				},
				{
					wantSpan:    "Down",
					wantMessage: "The value of the member 'Down' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Down = 1", applied: "\nenum Direction {\n  Up,\n  Down = 1,\n}\n      "},
						{desc: "Can be fixed to Down = 2", applied: "\nenum Direction {\n  Up,\n  Down = 2,\n}\n      "},
						{desc: "Can be fixed to Down = 'Down'", applied: "\nenum Direction {\n  Up,\n  Down = 'Down',\n}\n      "},
					},
				},
			},
		},
		{
			sourceText: "\nenum Direction {\n  Up = 'Up',\n  Down,\n}\n      ",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Down",
					wantMessage: "The value of the member 'Down' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Down = 1", applied: "\nenum Direction {\n  Up = 'Up',\n  Down = 1,\n}\n      "},
						{desc: "Can be fixed to Down = 2", applied: "\nenum Direction {\n  Up = 'Up',\n  Down = 2,\n}\n      "},
						{desc: "Can be fixed to Down = 'Down'", applied: "\nenum Direction {\n  Up = 'Up',\n  Down = 'Down',\n}\n      "},
					},
				},
			},
		},
		{
			sourceText: "\nenum Direction {\n  Up,\n  Down = 'Down',\n}\n      ",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "\nenum Direction {\n  Up = 0,\n  Down = 'Down',\n}\n      "},
						{desc: "Can be fixed to Up = 1", applied: "\nenum Direction {\n  Up = 1,\n  Down = 'Down',\n}\n      "},
						{desc: "Can be fixed to Up = 'Up'", applied: "\nenum Direction {\n  Up = 'Up',\n  Down = 'Down',\n}\n      "},
					},
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(preferEnumInitializersCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferEnumInitializers, preferEnumInitializersFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position := range wantIds {
				wantIds[position] = "defineInitializer"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage, diagnostic.Message.Description)
				}
				if diagnostic.Message.Id != "defineInitializer" {
					t.Fatalf("finding %d message id: got %q", position, diagnostic.Message.Id)
				}

				// The repair is offered, never applied. A rule shipping this as a fix would rewrite
				// an enum value unattended, which is the one thing this rule exists to say nobody
				// can guess.
				if len(diagnostic.Fixes) != 0 {
					t.Fatalf("finding %d proposes %d fixes; the repair must be suggestions only",
						position, len(diagnostic.Fixes))
				}
				if len(diagnostic.Suggestions) != len(want.wantSuggestions) {
					t.Fatalf("finding %d: expected %d suggestions, got %d",
						position, len(want.wantSuggestions), len(diagnostic.Suggestions))
				}
				for order, wantSuggestion := range want.wantSuggestions {
					got := diagnostic.Suggestions[order]
					if got.Message.Description != wantSuggestion.desc {
						t.Fatalf("finding %d suggestion %d: expected %q, got %q",
							position, order, wantSuggestion.desc, got.Message.Description)
					}
					if got.Message.Id != "defineInitializerSuggestion" {
						t.Fatalf("finding %d suggestion %d id: got %q", position, order, got.Message.Id)
					}
					applied := applyPreferEnumInitializersSuggestion(t, testCase.sourceText, got)
					if applied != wantSuggestion.applied {
						t.Fatalf("finding %d suggestion %d applied: expected %q, got %q",
							position, order, wantSuggestion.applied, applied)
					}
				}
			}
		})
	}
}

// TestPreferEnumInitializersStaysSilentOnShapesTheCorpusDoesNotWrite covers clean divergences.
//
// Upstream's corpus has four passing inputs and none of them carries a computed initializer, so
// nothing in it establishes that the rule tests only whether an initializer is PRESENT rather than
// what it evaluates to. Both verdicts below were measured against the installed 8.x build.
func TestPreferEnumInitializersStaysSilentOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{
			// Fully initialized, so silent.
			sourceText: "enum D {\n  Up = 'Up',\n  Down = 'Down',\n}",
		},
		{
			// A computed initializer is still an initializer.
			sourceText: "enum D {\n  A = 1 << 0,\n}",
		},
		{
			// So is a call-bearing one; the rule tests presence, never the shape of the value.
			sourceText: "enum D {\n  A = 'a'.length,\n}",
		},
	}
	for index, testCase := range cases {
		t.Run(preferEnumInitializersCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, PreferEnumInitializers,
				preferEnumInitializersFile, testCase.sourceText))
		})
	}
}

// TestPreferEnumInitializersFiresOnShapesTheCorpusDoesNotWrite covers reporting divergences.
//
// The corpus writes only plain identifier members inside a plain enum, so it cannot see three
// things this rule has to get right: that the suggested number is the member's POSITION rather than
// the value TypeScript computes for it, that a quoted key renders its quotes twice, and that const,
// ambient, exported and nested enums are all reached.
//
// The position case is the one worth naming. In `enum D { A = 1, B, C }` the member B evaluates to
// 2, and upstream offers `B = 1`. A port that reached for the real value would pass every message-id
// fixture and write a different edit than the corpus asserts.
//
// Every span, message and applied suggestion below was measured against the installed build.
func TestPreferEnumInitializersFiresOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []preferEnumInitializersFinding
	}{
		{
			// A quoted key: `getText(member)` includes the quotes, so upstream renders the name doubled in the message and emits a syntactically broken third suggestion. Reproduced rather than corrected; the differential compares the message text.
			sourceText: "enum D {\n  'quoted',\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "'quoted'",
					wantMessage: "The value of the member ''quoted'' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to 'quoted' = 0", applied: "enum D {\n  'quoted' = 0,\n}"},
						{desc: "Can be fixed to 'quoted' = 1", applied: "enum D {\n  'quoted' = 1,\n}"},
						{desc: "Can be fixed to 'quoted' = ''quoted''", applied: "enum D {\n  'quoted' = ''quoted'',\n}"},
					},
				},
			},
		},
		{
			// A const enum reports like any other.
			sourceText: "const enum D {\n  Up,\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "const enum D {\n  Up = 0,\n}"},
						{desc: "Can be fixed to Up = 1", applied: "const enum D {\n  Up = 1,\n}"},
						{desc: "Can be fixed to Up = 'Up'", applied: "const enum D {\n  Up = 'Up',\n}"},
					},
				},
			},
		},
		{
			// An ambient enum reports too, even though it declares rather than defines.
			sourceText: "declare enum D {\n  Up,\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "declare enum D {\n  Up = 0,\n}"},
						{desc: "Can be fixed to Up = 1", applied: "declare enum D {\n  Up = 1,\n}"},
						{desc: "Can be fixed to Up = 'Up'", applied: "declare enum D {\n  Up = 'Up',\n}"},
					},
				},
			},
		},
		{
			// A nested enum is reached: the listener anchors on the declaration wherever it sits.
			sourceText: "namespace N {\n  enum D {\n    Up,\n  }\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "namespace N {\n  enum D {\n    Up = 0,\n  }\n}"},
						{desc: "Can be fixed to Up = 1", applied: "namespace N {\n  enum D {\n    Up = 1,\n  }\n}"},
						{desc: "Can be fixed to Up = 'Up'", applied: "namespace N {\n  enum D {\n    Up = 'Up',\n  }\n}"},
					},
				},
			},
		},
		{
			// The suggested numbers are POSITIONS, not computed values. C is offered 2 and 3.
			sourceText: "enum D {\n  A,\n  B = 1,\n  C,\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "A",
					wantMessage: "The value of the member 'A' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to A = 0", applied: "enum D {\n  A = 0,\n  B = 1,\n  C,\n}"},
						{desc: "Can be fixed to A = 1", applied: "enum D {\n  A = 1,\n  B = 1,\n  C,\n}"},
						{desc: "Can be fixed to A = 'A'", applied: "enum D {\n  A = 'A',\n  B = 1,\n  C,\n}"},
					},
				},
				{
					wantSpan:    "C",
					wantMessage: "The value of the member 'C' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to C = 2", applied: "enum D {\n  A,\n  B = 1,\n  C = 2,\n}"},
						{desc: "Can be fixed to C = 3", applied: "enum D {\n  A,\n  B = 1,\n  C = 3,\n}"},
						{desc: "Can be fixed to C = 'C'", applied: "enum D {\n  A,\n  B = 1,\n  C = 'C',\n}"},
					},
				},
			},
		},
		{
			// The case that separates position from value: B's computed value is 2 and upstream offers 1, because index is the member's place in the list.
			sourceText: "enum D {\n  A = 1,\n  B,\n  C,\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "B",
					wantMessage: "The value of the member 'B' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to B = 1", applied: "enum D {\n  A = 1,\n  B = 1,\n  C,\n}"},
						{desc: "Can be fixed to B = 2", applied: "enum D {\n  A = 1,\n  B = 2,\n  C,\n}"},
						{desc: "Can be fixed to B = 'B'", applied: "enum D {\n  A = 1,\n  B = 'B',\n  C,\n}"},
					},
				},
				{
					wantSpan:    "C",
					wantMessage: "The value of the member 'C' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to C = 2", applied: "enum D {\n  A = 1,\n  B,\n  C = 2,\n}"},
						{desc: "Can be fixed to C = 3", applied: "enum D {\n  A = 1,\n  B,\n  C = 3,\n}"},
						{desc: "Can be fixed to C = 'C'", applied: "enum D {\n  A = 1,\n  B,\n  C = 'C',\n}"},
					},
				},
			},
		},
		{
			// A single-line enum with no trailing comma.
			sourceText: "enum D { Up }",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "enum D { Up = 0 }"},
						{desc: "Can be fixed to Up = 1", applied: "enum D { Up = 1 }"},
						{desc: "Can be fixed to Up = 'Up'", applied: "enum D { Up = 'Up' }"},
					},
				},
			},
		},
		{
			// An exported enum reports.
			sourceText: "export enum D {\n  Up,\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "Up",
					wantMessage: "The value of the member 'Up' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to Up = 0", applied: "export enum D {\n  Up = 0,\n}"},
						{desc: "Can be fixed to Up = 1", applied: "export enum D {\n  Up = 1,\n}"},
						{desc: "Can be fixed to Up = 'Up'", applied: "export enum D {\n  Up = 'Up',\n}"},
					},
				},
			},
		},
		{
			// Four members in one enum, so four findings in source order, each with its own index.
			sourceText: "enum D {\n  A,\n  B,\n  C,\n  D2,\n}",
			wantFindings: []preferEnumInitializersFinding{
				{
					wantSpan:    "A",
					wantMessage: "The value of the member 'A' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to A = 0", applied: "enum D {\n  A = 0,\n  B,\n  C,\n  D2,\n}"},
						{desc: "Can be fixed to A = 1", applied: "enum D {\n  A = 1,\n  B,\n  C,\n  D2,\n}"},
						{desc: "Can be fixed to A = 'A'", applied: "enum D {\n  A = 'A',\n  B,\n  C,\n  D2,\n}"},
					},
				},
				{
					wantSpan:    "B",
					wantMessage: "The value of the member 'B' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to B = 1", applied: "enum D {\n  A,\n  B = 1,\n  C,\n  D2,\n}"},
						{desc: "Can be fixed to B = 2", applied: "enum D {\n  A,\n  B = 2,\n  C,\n  D2,\n}"},
						{desc: "Can be fixed to B = 'B'", applied: "enum D {\n  A,\n  B = 'B',\n  C,\n  D2,\n}"},
					},
				},
				{
					wantSpan:    "C",
					wantMessage: "The value of the member 'C' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to C = 2", applied: "enum D {\n  A,\n  B,\n  C = 2,\n  D2,\n}"},
						{desc: "Can be fixed to C = 3", applied: "enum D {\n  A,\n  B,\n  C = 3,\n  D2,\n}"},
						{desc: "Can be fixed to C = 'C'", applied: "enum D {\n  A,\n  B,\n  C = 'C',\n  D2,\n}"},
					},
				},
				{
					wantSpan:    "D2",
					wantMessage: "The value of the member 'D2' should be explicitly defined.",
					wantSuggestions: []preferEnumInitializersSuggestion{
						{desc: "Can be fixed to D2 = 3", applied: "enum D {\n  A,\n  B,\n  C,\n  D2 = 3,\n}"},
						{desc: "Can be fixed to D2 = 4", applied: "enum D {\n  A,\n  B,\n  C,\n  D2 = 4,\n}"},
						{desc: "Can be fixed to D2 = 'D2'", applied: "enum D {\n  A,\n  B,\n  C,\n  D2 = 'D2',\n}"},
					},
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(preferEnumInitializersCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferEnumInitializers, preferEnumInitializersFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position := range wantIds {
				wantIds[position] = "defineInitializer"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]
				gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage, diagnostic.Message.Description)
				}
				if len(diagnostic.Suggestions) != len(want.wantSuggestions) {
					t.Fatalf("finding %d: expected %d suggestions, got %d", position, len(want.wantSuggestions), len(diagnostic.Suggestions))
				}
				for order, wantSuggestion := range want.wantSuggestions {
					got := diagnostic.Suggestions[order]
					if got.Message.Description != wantSuggestion.desc {
						t.Fatalf("finding %d suggestion %d: expected %q, got %q", position, order, wantSuggestion.desc, got.Message.Description)
					}
					applied := applyPreferEnumInitializersSuggestion(t, testCase.sourceText, got)
					if applied != wantSuggestion.applied {
						t.Fatalf("finding %d suggestion %d applied: expected %q, got %q", position, order, wantSuggestion.applied, applied)
					}
				}
			}
		})
	}
}

// TestPreferEnumInitializersReportsAComputedKeyUpstreamCannotParse pins a stated divergence.
//
// A computed enum key is illegal TypeScript, and upstream's parser refuses the file outright with
// "Computed property names are not allowed in enums", so the rule never runs and reports nothing.
// That silence is a fact about the parser rather than about the rule, and the tell is that every
// other diagnostic disappears with it.
//
// Our parser recovers and hands back a real enum member, so this port reports it and interpolates
// the member's whole text as the name. The divergence is recorded here rather than papered over
// with a computed-key skip, because a skip would be a judgment upstream never made and this input
// cannot occur in source that compiles.
func TestPreferEnumInitializersReportsAComputedKeyUpstreamCannotParse(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, PreferEnumInitializers, preferEnumInitializersFile,
		"enum D {\n  ['computed'],\n}")
	rule_testing.ExpectFindings(t, result, "defineInitializer")
	if got := result.Diagnostics[0].Message.Description; got != "The value of the member '['computed']' should be explicitly defined." {
		t.Fatalf("message: got %q", got)
	}
}
