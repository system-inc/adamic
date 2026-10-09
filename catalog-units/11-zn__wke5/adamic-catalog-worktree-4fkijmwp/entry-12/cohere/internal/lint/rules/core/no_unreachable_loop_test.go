package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// unreachableLoopFile is where the fixtures pretend to live.
const unreachableLoopFile = "/repository/source/Loops.ts"

// The corpus is ESLint's own, generated the way upstream generates it and then driven through the
// installed eslint 10.8.1 build to record what it actually reports on each input.
//
// Upstream's tester builds its cases from a cross product: 28 loop templates against 28 clean
// bodies and 29 exiting bodies, plus 21 hand-written cases and 12 carrying an `ignore` option.
// 1,629 cases in total, and the installed build agreed with the corpus's own stated verdict on
// every one of them, so nothing here rests on reading the rule source.
//
// The tables live in no_unreachable_loop_cases_test.go, emitted from that measurement rather than
// typed, and verified byte for byte back against the corpus.

// TestNoUnreachableLoopFires runs every corpus input upstream reports on.
//
// The finding count per input is upstream's rather than assumed: a loop nested in an invalid loop
// can make both invalid, and eleven cases in the corpus report more than once.
func TestNoUnreachableLoopFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range unreachableLoopFiringCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "invalid"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnreachableLoop, unreachableLoopFile, testCase.source), wantIds...)
		})
	}
}

// TestNoUnreachableLoopStaysSilent runs every corpus input upstream leaves alone.
//
// This is the half that matters. A rule reporting "the body exits" from the syntax alone passes
// most of the firing table and fails here on `while (a) { if (foo) break; }`, on every body holding
// a `continue`, and on the three shapes upstream deliberately declines to evaluate.
func TestNoUnreachableLoopStaysSilent(t *testing.T) {
	t.Parallel()

	for _, source := range unreachableLoopCleanCases {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnreachableLoop, unreachableLoopFile, source))
		})
	}
}

// TestNoUnreachableLoopOptions runs the corpus's `ignore` cases through the rule's own decoder.
//
// Routed through `rule.DecodeOptionsInto` rather than by building the struct, because the wire
// shape and the field name are the two lines with no upstream counterpart, and a fixture handing
// the rule a struct leaves both untested.
//
// The JSON here is the bare object. Upstream writes `[{...}]` in its corpus and cohere's config
// layer unwraps the severity tuple before dispatch, so the array wrapper never reaches a decoder.
func TestNoUnreachableLoopOptions(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[NoUnreachableLoopOptions]()

	for _, testCase := range unreachableLoopOptionCases {
		t.Run(testCase.source+" "+testCase.options, func(t *testing.T) {
			t.Parallel()
			decoded, err := decode([]byte(testCase.options))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.options, err)
			}
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "invalid"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, NoUnreachableLoop, unreachableLoopFile, testCase.source, decoded),
				wantIds...)
		})
	}
}

// TestNoUnreachableLoopHandlesNilOptions pins the bare-severity path, which no corpus case reaches.
//
// A rule configured as `"no-unreachable-loop": "error"` is handed nil, and `options.(T)` on nil
// yields the zero value rather than reporting failure. That is right here, since upstream's default
// ignores nothing, but it is right by accident unless something asserts it: a rule whose defaults
// inverted under nil would pass every option fixture in the table above, because each of those
// arrives through the decoder with a real value.
func TestNoUnreachableLoopHandlesNilOptions(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, NoUnreachableLoop, unreachableLoopFile, "while (a) break;", nil),
		"invalid")

	rule_testing.ExpectClean(t,
		rule_testing.RunWithOptions(t, NoUnreachableLoop, unreachableLoopFile, "while (a) { bar(); }", nil))
}

// TestNoUnreachableLoopSpans asserts where the finding points, which no message-id fixture can see.
//
// Upstream reports on the loop statement, so the span is the whole loop from its first token to its
// last. Measured against the installed build, which gives `while (a) break;` columns 1 through 17,
// the whole 16-character statement.
//
// The nested case is the one worth having twice over: it pins both spans AND their order, and a
// port emitting findings as its per-root walk discovers them would order these the other way.
func TestNoUnreachableLoopSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		reports []string
	}{
		{
			"the whole while statement",
			"while (a) break;",
			[]string{"while (a) break;"},
		},
		{
			"the whole do-while statement, whose last token is the closing paren",
			"do break; while (a)",
			[]string{"do break; while (a)"},
		},
		{
			"a loop indented inside another, pointing past the leading whitespace",
			"for (;;) { for (var i = 1; i < 10; i ++) break; if (foo) break; continue; }",
			[]string{"for (var i = 1; i < 10; i ++) break;"},
		},
		{
			"outer before inner, in source order rather than discovery order",
			"for (a in b) { while (foo) { if(baz) { break; } else { break; } } break; }",
			[]string{
				"for (a in b) { while (foo) { if(baz) { break; } else { break; } } break; }",
				"while (foo) { if(baz) { break; } else { break; } }",
			},
		},
		{
			"a loop inside a function, reported at the loop and not at the function",
			"function foo() { for (var i = 0; i < 10; i++) { do { return; } while(i) } }",
			[]string{
				"for (var i = 0; i < 10; i++) { do { return; } while(i) }",
				"do { return; } while(i)",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnreachableLoop, unreachableLoopFile, testCase.source)
			if len(result.Diagnostics) != len(testCase.reports) {
				t.Fatalf("expected %d findings, got %d", len(testCase.reports), len(result.Diagnostics))
			}
			for index, want := range testCase.reports {
				diagnostic := result.Diagnostics[index]
				// rule_testing.Run does not trim the source, unlike RunTyped, so the file on disk and
				// the literal here are the same bytes and this slice needs no adjustment.
				got := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Errorf("finding %d span: got %q, want %q", index, got, want)
				}
			}
		})
	}
}

// TestNoUnreachableLoopMessage asserts the message the rule carries.
//
// `rule.Message` is `{Id, Description}` with no interpolation, so there is nothing to render and
// this is a direct equality on both fields. Asserted against literals typed here rather than
// against the rule's own constant, because comparing a diagnostic to the constant it was built from
// is an equality that moves in both directions under mutation and cannot fail.
func TestNoUnreachableLoopMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUnreachableLoop, unreachableLoopFile, "while (a) break;")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
	}
	message := result.Diagnostics[0].Message

	if message.Id != "invalid" {
		t.Errorf("message id: got %q, want %q", message.Id, "invalid")
	}
	// Upstream's whole message is "Invalid loop. Its body allows only one iteration." Ours says why
	// rather than restating the rule name, so the text is not upstream's and the assertion is on
	// the prefix actually meant rather than on a substring that a longer wrong message would also
	// satisfy.
	const wantPrefix = "This loop can never run a second iteration, because every path through its body leaves the loop."
	if !strings.HasPrefix(message.Description, wantPrefix) {
		t.Errorf("message description: got %q, want it to start with %q", message.Description, wantPrefix)
	}
}

// TestNoUnreachableLoopIgnoringEveryKindDisablesTheRule pins the empty-selector arm.
//
// Upstream returns an empty visitor when every loop kind is ignored, rather than a visitor that
// matches nothing, and the corpus asserts it with one case. Kept as its own test because the arm is
// a whole-rule short circuit rather than a per-loop decision, and a port that dropped it would
// still pass that corpus case through the per-loop filter.
func TestNoUnreachableLoopIgnoringEveryKindDisablesTheRule(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[NoUnreachableLoopOptions]()
	decoded, err := decode([]byte(`{"ignore":["WhileStatement","DoWhileStatement","ForStatement","ForInStatement","ForOfStatement"]}`))
	if err != nil {
		t.Fatalf("decoding: %v", err)
	}

	// Five loops, one of every kind, every one of them invalid without the option.
	const source = "while (a) break; do break; while (b); for (;;) break; for (c in d) break; for (e of f) break;"
	rule_testing.ExpectClean(t,
		rule_testing.RunWithOptions(t, NoUnreachableLoop, unreachableLoopFile, source, decoded))
}

// TestNoUnreachableLoopJudgesEachRootSeparately covers a shape the corpus does not write.
//
// The corpus nests loops inside functions but never puts a loop in a function that is itself
// unreachable, and the two reachability questions are different: the function is never called, yet
// the loop inside it is perfectly reachable within the function's own graph. Upstream gets this
// from having one code path per function; here it comes from judging a loop only in the root
// returned by `control_flow_graph.RootOf`, and without that guard the outer graph's walk past the nested
// body would answer the wrong question.
//
// Measured against the installed eslint 10.8.1 build: both loops report.
func TestNoUnreachableLoopJudgesEachRootSeparately(t *testing.T) {
	t.Parallel()

	const source = "function outer() { return; function inner() { while (a) break; } } while (b) break;"
	result := rule_testing.Run(t, NoUnreachableLoop, unreachableLoopFile, source)
	rule_testing.ExpectFindings(t, result, "invalid", "invalid")

	wantSpans := []string{"while (a) break;", "while (b) break;"}
	for index, want := range wantSpans {
		got := source[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
		if got != want {
			t.Errorf("finding %d span: got %q, want %q", index, got, want)
		}
	}
}

// TestNoUnreachableLoopInNonFunctionRoots covers the two code path roots that are neither a file
// nor a function, which the corpus has no cases for because ESLint's tester runs at ES2018 and a
// class static block is ES2022.
//
// Both are roots in `control_flow_graph.IsRoot`, so a rule enumerating only files and functions would go
// silent on them. Measured against the installed eslint 10.8.1 build at ecmaVersion 2022: both
// report.
func TestNoUnreachableLoopInNonFunctionRoots(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			"a class static block",
			"class C { static { while (a) break; } }",
			"while (a) break;",
		},
		{
			"a property initializer, whose loop lives in an arrow inside it",
			"class C { field = (() => { while (a) break; })(); }",
			"while (a) break;",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnreachableLoop, unreachableLoopFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "invalid")
			got := testCase.source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if got != testCase.want {
				t.Errorf("span: got %q, want %q", got, testCase.want)
			}
		})
	}
}
