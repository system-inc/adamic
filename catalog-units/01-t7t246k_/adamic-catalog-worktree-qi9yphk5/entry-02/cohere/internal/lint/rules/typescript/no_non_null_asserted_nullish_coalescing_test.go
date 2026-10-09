package typescript

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const noNonNullAssertedNullishCoalescingFile = "/repository/source/Coalescing.ts"

func noNonNullAssertedNullishCoalescingCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// applyNoNonNullAssertedNullishCoalescingSuggestion rewrites source with one suggestion's fixes.
//
// The harness applies fixes and has no suggestion support, and this rule's ENTIRE repair surface is
// a suggestion, so without this the thing it offers would go completely unasserted.
func applyNoNonNullAssertedNullishCoalescingSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
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

// TestNoNonNullAssertedNullishCoalescingStaysSilentOnUpstreamPassCases is the clean corpus, verbatim.
//
// All eighteen of upstream's passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler. Every one was additionally run through the installed 8.x build, which
// reported nothing and produced no parse error on any of them.
//
// Six of the eighteen are the assignment test: a variable declared without a value, asserted on the
// left of a nullish coalesce, is clean because the assertion may genuinely be doing something. Those
// six are most of what this rule decides, and none of them can be reached without resolving the
// identifier to its binding.
func TestNoNonNullAssertedNullishCoalescingStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"foo ?? bar;",
		"foo ?? bar!;",
		"foo.bazz ?? bar;",
		"foo.bazz ?? bar!;",
		"foo!.bazz ?? bar;",
		"foo!.bazz ?? bar!;",
		"foo() ?? bar;",
		"foo() ?? bar!;",
		"(foo ?? bar)!;",
		"\nlet x: string;\nx! ?? '';\n    ",
		"\nlet x: string;\nx ?? '';\n    ",
		"\nlet x!: string;\nx ?? '';\n    ",
		"\nlet x: string;\nfoo(x);\nx! ?? '';\n    ",
		"\nlet x: string;\nx! ?? '';\nx = foo();\n    ",
		"\nlet x: string;\nfoo(x);\nx! ?? '';\nx = foo();\n    ",
		"\nlet x = foo();\nx ?? '';\n    ",
		"\nfunction foo() {\n  let x: string;\n  return x ?? '';\n}\n    ",
		"\nlet x: string;\nfunction foo() {\n  return x ?? '';\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(noNonNullAssertedNullishCoalescingCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoNonNullAssertedNullishCoalescing,
				noNonNullAssertedNullishCoalescingFile, sourceText))
		})
	}
}

// TestNoNonNullAssertedNullishCoalescingFiresOnUpstreamFailCases is the reporting corpus, verbatim.
//
// All fifteen of upstream's failing inputs, each carrying one finding and one suggestion. The span
// is measured from the installed build's reported range, and each expected rewrite is that build's
// own fix replayed against the input; all fifteen were separately confirmed equal to the `output`
// fields the corpus records, so the running rule and the checked-in corpus agree.
//
// The `x  !` row is the one that pins how the mark is located. Its assertion spans the name, two
// spaces and the mark, and the repair must delete only the mark, so a rule computing the token from
// the node's end by arithmetic writes the wrong bytes while satisfying every message assertion.
//
// The harness writes each fixture as `strings.TrimSpace(source)+"\n"`, so both the span slices and
// the expected rewrites are against that trimmed text rather than the Go literal.
func TestNoNonNullAssertedNullishCoalescingFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText     string
		wantSpan       string
		wantSuggestion string
	}{
		{
			sourceText:     "foo! ?? bar;",
			wantSpan:       "foo!",
			wantSuggestion: "foo ?? bar;\n",
		},
		{
			sourceText:     "foo! ?? bar!;",
			wantSpan:       "foo!",
			wantSuggestion: "foo ?? bar!;\n",
		},
		{
			sourceText:     "foo.bazz! ?? bar;",
			wantSpan:       "foo.bazz!",
			wantSuggestion: "foo.bazz ?? bar;\n",
		},
		{
			sourceText:     "foo.bazz! ?? bar!;",
			wantSpan:       "foo.bazz!",
			wantSuggestion: "foo.bazz ?? bar!;\n",
		},
		{
			sourceText:     "foo!.bazz! ?? bar;",
			wantSpan:       "foo!.bazz!",
			wantSuggestion: "foo!.bazz ?? bar;\n",
		},
		{
			sourceText:     "foo!.bazz! ?? bar!;",
			wantSpan:       "foo!.bazz!",
			wantSuggestion: "foo!.bazz ?? bar!;\n",
		},
		{
			sourceText:     "foo()! ?? bar;",
			wantSpan:       "foo()!",
			wantSuggestion: "foo() ?? bar;\n",
		},
		{
			sourceText:     "foo()! ?? bar!;",
			wantSpan:       "foo()!",
			wantSuggestion: "foo() ?? bar!;\n",
		},
		{
			sourceText:     "\nlet x!: string;\nx! ?? '';\n      ",
			wantSpan:       "x!",
			wantSuggestion: "let x!: string;\nx ?? '';\n",
		},
		{
			sourceText:     "\nlet x: string;\nx = foo();\nx! ?? '';\n      ",
			wantSpan:       "x!",
			wantSuggestion: "let x: string;\nx = foo();\nx ?? '';\n",
		},
		{
			sourceText:     "\nlet x: string;\nx = foo();\nx! ?? '';\nx = foo();\n      ",
			wantSpan:       "x!",
			wantSuggestion: "let x: string;\nx = foo();\nx ?? '';\nx = foo();\n",
		},
		{
			sourceText:     "\nlet x = foo();\nx! ?? '';\n      ",
			wantSpan:       "x!",
			wantSuggestion: "let x = foo();\nx ?? '';\n",
		},
		{
			sourceText:     "\nfunction foo() {\n  let x!: string;\n  return x! ?? '';\n}\n      ",
			wantSpan:       "x!",
			wantSuggestion: "function foo() {\n  let x!: string;\n  return x ?? '';\n}\n",
		},
		{
			sourceText:     "\nlet x!: string;\nfunction foo() {\n  return x! ?? '';\n}\n      ",
			wantSpan:       "x!",
			wantSuggestion: "let x!: string;\nfunction foo() {\n  return x ?? '';\n}\n",
		},
		{
			sourceText:     "\nlet x = foo();\nx  ! ?? '';\n      ",
			wantSpan:       "x  !",
			wantSuggestion: "let x = foo();\nx   ?? '';\n",
		},
	}
	for index, testCase := range cases {
		t.Run(noNonNullAssertedNullishCoalescingCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoNonNullAssertedNullishCoalescing,
				noNonNullAssertedNullishCoalescingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noNonNullAssertedNullishCoalescing")

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]

			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
			if diagnostic.Message.Description != "The nullish coalescing operator is designed to handle undefined and null - using a non-null assertion is not needed." {
				t.Fatalf("message: got %q", diagnostic.Message.Description)
			}
			if diagnostic.Message.Id != "noNonNullAssertedNullishCoalescing" {
				t.Fatalf("message id: got %q", diagnostic.Message.Id)
			}

			// The repair is offered, never applied. Removing the assertion can WIDEN the resulting
			// type and break a return that used to compile, which is why upstream ships a
			// suggestion and why a fix here would be a defect rather than an improvement.
			if len(diagnostic.Fixes) != 0 {
				t.Fatalf("the repair must be a suggestion, got %d fixes", len(diagnostic.Fixes))
			}
			if len(diagnostic.Suggestions) != 1 {
				t.Fatalf("expected one suggestion, got %d", len(diagnostic.Suggestions))
			}
			suggestion := diagnostic.Suggestions[0]
			if suggestion.Message.Description != "Remove the non-null assertion." {
				t.Fatalf("suggestion text: got %q", suggestion.Message.Description)
			}
			if suggestion.Message.Id != "suggestRemovingNonNull" {
				t.Fatalf("suggestion id: got %q", suggestion.Message.Id)
			}
			applied := applyNoNonNullAssertedNullishCoalescingSuggestion(t, onDisk, suggestion)
			if applied != testCase.wantSuggestion {
				t.Fatalf("suggestion applied: expected %q, got %q", testCase.wantSuggestion, applied)
			}
		})
	}
}

// TestNoNonNullAssertedNullishCoalescingStaysSilentOnShapesTheCorpusDoesNotWrite covers the clean
// divergences.
//
// The corpus writes no other logical operator, no shadowed binding, and no parameter, so nothing in
// it establishes that the rule is about `??` specifically, that the assignment test resolves to the
// right binding, or that a parameter is not a variable declaration. The shadowing pair is the one
// worth reading twice: resolving to the wrong binding would flip both rows, and each row alone would
// still look correct.
func TestNoNonNullAssertedNullishCoalescingStaysSilentOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{
			// The assertion sits on a RIGHT operand of both, so neither binary matches.
			sourceText: "foo ?? bar! ?? baz;",
		},
		{
			// A different logical operator entirely.
			sourceText: "foo! || bar;",
		},
		{
			// And the other one.
			sourceText: "foo! && bar;",
		},
		{
			// The same write BELOW the node does not count. This is the pair that makes the end-position comparison load-bearing rather than decorative.
			sourceText: "let x: string;\nx! ?? '';\nfunction g() { x = 'a'; }",
		},
		{
			// A write to a SHADOWED binding of the same name. The outer `x` is never assigned, so
			// the rule must stay silent, and it can only know that by comparing symbols: the walk
			// pre-filters on the identifier's text, so name matching alone would see the inner
			// write and report.
			//
			// Added after a mutant dropping the symbol comparison survived the whole corpus. The
			// corpus has shadowing cases, but in all of them the inner binding is never WRITTEN,
			// so the pre-filter and the symbol test agree on every one. Three shapes, measured
			// clean on the installed build, paired with a reporting control in the firing test.
			sourceText: "let x: string;\n{ let x = 1; x = 2; }\nx! ?? '';",
		},
		{
			sourceText: "let x: string;\nfunction g() { let x = 1; x = 2; }\nx! ?? '';",
		},
		{
			// The shadow is a PARAMETER rather than a declaration, which reaches the same
			// comparison through a different node kind.
			sourceText: "let x: string;\nfunction g(x: number) { x = 2; }\nx! ?? '';",
		},
		{
			// A declaration that carries a value but sits BELOW the assertion. Upstream compares
			// end positions on the definition half as well as the reference half, so a hoisted
			// `var` initialized later does not count as an assignment before this point.
			//
			// Added after a mutant dropping that position test survived the whole corpus: upstream
			// writes no case where the initialized declaration comes after the node, so nothing
			// imported can see the comparison at all. Measured clean on the installed build.
			sourceText: "x! ?? '';\nvar x = 1;",
		},
		{
			sourceText: "function f() { return x! ?? ''; }\nvar x = 1;",
		},
		{
			sourceText: "function f() { return x! ?? ''; }\nlet x!: string;",
		},
		{
			// A parameter is not a variable declaration, so neither half of the assignment test fires and the rule stays silent.
			sourceText: "class K { m(x?: string) { return x! ?? ''; } }",
		},
		{
			// The same shape as a plain function parameter, which is upstream's own example in the comment explaining why the repair is a suggestion.
			sourceText: "function f(x?: string) { return x! ?? ''; }",
		},
		{
			// A shadowing inner binding with no value. Resolving to the OUTER one would see its initializer and report, so this pins that the symbol comparison picks the right binding.
			sourceText: "let x = 1;\n{ let x: string; x! ?? ''; }",
		},
		{
			// The mirror: an initialized binding in an inner block must not satisfy the outer name's test.
			sourceText: "let x: string;\n{ let x = 1; }\nx! ?? '';",
		},
	}
	for index, testCase := range cases {
		t.Run(noNonNullAssertedNullishCoalescingCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoNonNullAssertedNullishCoalescing,
				noNonNullAssertedNullishCoalescingFile, testCase.sourceText))
		})
	}
}

// TestNoNonNullAssertedNullishCoalescingFiresOnShapesTheCorpusDoesNotWrite covers the reporting
// divergences.
//
// The parenthesized row is here because the rule was wrong about it. The doc comment argued that a
// direct-child selector cannot match through parentheses, which is true of a tree that HAS a
// parenthesized node and false of estree, so upstream reports and this port was silently missing it.
// The same mistake in the same batch cost a sibling rule a false negative, which is why both now
// carry a measurement rather than an argument.
//
// The hoisted-write row pairs with its write-below twin in the clean test above, and together they
// are what make the end-position comparison a real discrimination rather than a detail.
func TestNoNonNullAssertedNullishCoalescingFiresOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText     string
		wantSpan       string
		wantSuggestion string
	}{
		{
			// A parenthesized left side. estree has no paren node, so upstream's child selector sees through it and reports. This is the case the rule got wrong from a confident argument before it was measured.
			sourceText:     "(foo!) ?? bar;",
			wantSpan:       "foo!",
			wantSuggestion: "(foo) ?? bar;\n",
		},
		{
			// Chained coalescing: the assertion is on the left of the inner one.
			sourceText:     "foo! ?? bar ?? baz;",
			wantSpan:       "foo!",
			wantSuggestion: "foo ?? bar ?? baz;\n",
		},
		{
			// A write inside a hoisted function above the node still counts, which is why the search is over the whole file rather than the enclosing scope.
			sourceText:     "let x: string;\nfunction g() { x = 'a'; }\nx! ?? '';",
			wantSpan:       "x!",
			wantSuggestion: "let x: string;\nfunction g() { x = 'a'; }\nx ?? '';\n",
		},
		{
			// A doubled assertion: the outer one is the left operand, and the span covers both.
			sourceText:     "declare const foo: any;\nfoo!! ?? 1;",
			wantSpan:       "foo!!",
			wantSuggestion: "declare const foo: any;\nfoo! ?? 1;\n",
		},
		{
			// The control for the three shadowing rows above: the OUTER binding really is written
			// before the node, so the same shape reports. Without this, those three would pass
			// against a rule that had simply stopped finding writes at all.
			sourceText:     "let x: string;\nx = 'a';\nx! ?? '';",
			wantSpan:       "x!",
			wantSuggestion: "let x: string;\nx = 'a';\nx ?? '';\n",
		},
		{
			// A newline between the assertion and the operator.
			sourceText:     "foo!\n ?? bar;",
			wantSpan:       "foo!",
			wantSuggestion: "foo\n ?? bar;\n",
		},
	}
	for index, testCase := range cases {
		t.Run(noNonNullAssertedNullishCoalescingCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoNonNullAssertedNullishCoalescing,
				noNonNullAssertedNullishCoalescingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noNonNullAssertedNullishCoalescing")

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]

			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
			if len(diagnostic.Suggestions) != 1 {
				t.Fatalf("expected one suggestion, got %d", len(diagnostic.Suggestions))
			}
			applied := applyNoNonNullAssertedNullishCoalescingSuggestion(t, onDisk, diagnostic.Suggestions[0])
			if applied != testCase.wantSuggestion {
				t.Fatalf("suggestion applied: expected %q, got %q", testCase.wantSuggestion, applied)
			}
		})
	}
}

// TestNoNonNullAssertedNullishCoalescingNeedsTheTypedHarness pins the checker declaration.
//
// Upstream reads its scope manager rather than the type checker, so seeing no `getParserServices`
// in the rule file is not evidence this question can be answered without types HERE. The assignment
// test resolves an identifier to its binding, which is a checker call in this tree.
//
// Under the plain harness the checker is nil and the assignment test answers false, so every
// identifier case goes silent. That is the dangerous direction: the clean fixtures would all pass
// having proven nothing.
func TestNoNonNullAssertedNullishCoalescingNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoNonNullAssertedNullishCoalescing.NeedsTypeChecker {
		t.Fatal("the assignment test resolves an identifier to its binding, so the rule must declare NeedsTypeChecker")
	}

	// Handed no checker, the rule must stay silent rather than panic. The non-identifier shapes
	// still report, because they never consult the checker at all, which is why this asserts an
	// identifier case specifically.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoNonNullAssertedNullishCoalescing,
		noNonNullAssertedNullishCoalescingFile, "let x = foo();\nx! ?? '';"))
}
