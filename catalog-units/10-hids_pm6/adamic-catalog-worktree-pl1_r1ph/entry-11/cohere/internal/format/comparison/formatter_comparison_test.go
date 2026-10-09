package formatter_comparison

import (
	"strings"
	"testing"
)

// TestFormatsCrampedSource is the known-dirty control. A differ that has never returned a positive
// has not been shown to work, so this proves the formatter is actually reached and actually edits.
func TestFormatsCrampedSource(t *testing.T) {
	t.Parallel()
	formatted, ok, _, _ := FormatFile("/probe.ts", "const   x=1;\n")
	if !ok {
		t.Fatal("could not parse the probe")
	}
	if formatted == "const   x=1;\n" {
		t.Fatal("the formatter made no change to obviously cramped source, so it is not running")
	}
	if formatted != "const x = 1;\n" {
		t.Fatalf("unexpected formatting: %q", formatted)
	}
}

// TestLeavesFormattedSourceAlone is the clean control, and the other half of proving the differ
// works: a detector that flags everything is as useless as one that flags nothing.
func TestLeavesFormattedSourceAlone(t *testing.T) {
	t.Parallel()
	source := "const x = 1;\n"
	formatted, ok, _, _ := FormatFile("/clean.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if formatted != source {
		t.Fatalf("already-formatted source was changed to %q", formatted)
	}
}

// TestDoesNotWrapLongLines pins the structural finding: FormatCodeSettings has no print width, and
// the formatter never breaks a line to fit one. Prettier's entire layout algorithm is line-width
// driven, so this is the gap that decides the phase.
func TestDoesNotWrapLongLines(t *testing.T) {
	t.Parallel()
	// Well past printWidth 120, and trivially breakable by Prettier at every comma.
	source := "const value = someFunction(" + strings.Repeat("argument, ", 30) + "last);\n"
	formatted, ok, _, _ := FormatFile("/long.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if strings.Count(formatted, "\n") != 1 {
		t.Fatalf("expected the long line to be left on one line, got:\n%s", formatted)
	}
}

// TestDoesNotJoinShortLines is the mirror: Prettier collapses a needlessly broken call onto one
// line when it fits. The formatter preserves the author's breaks instead.
func TestDoesNotJoinShortLines(t *testing.T) {
	t.Parallel()
	source := "const value = f(\n    1,\n    2,\n);\n"
	formatted, ok, _, _ := FormatFile("/short.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if !strings.Contains(formatted, "\n    1,") {
		t.Fatalf("expected the author's line breaks to be preserved, got:\n%s", formatted)
	}
}

// TestDoesNotNormalizeQuotes pins the second unmapped Prettier option. singleQuote has no field on
// FormatCodeSettings, and the formatter does not touch string literals at all.
func TestDoesNotNormalizeQuotes(t *testing.T) {
	t.Parallel()
	source := "const name = \"double\";\n"
	formatted, ok, _, _ := FormatFile("/quotes.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if !strings.Contains(formatted, "\"double\"") {
		t.Fatalf("expected double quotes to survive, got:\n%s", formatted)
	}
}

// TestApplyIsBackToFront guards the edit application itself. Applying changes in forward order
// silently corrupts every range after the first length change, and the corruption looks like a
// formatting difference rather than a bug in the harness.
func TestApplyIsBackToFront(t *testing.T) {
	t.Parallel()
	formatted, ok, _, _ := FormatFile("/multi.ts", "const  a=1;\nconst  b=2;\nconst  c=3;\n")
	if !ok {
		t.Fatal("could not parse")
	}
	want := "const a = 1;\nconst b = 2;\nconst c = 3;\n"
	if formatted != want {
		t.Fatalf("got %q, want %q", formatted, want)
	}
}

// The tests below pin the measured divergences against our Prettier fork. Each one is a fact about
// the formatter established by the corpus sweep over 1,542 already-formatted files, kept here so a
// typescript-go bump that changes any of them fails loudly rather than silently reshaping the tree.

// TestKeywordSpaceIsSettable pins the largest divergence class and the setting that closes it. Our
// house style is `if(x)`; the formatter's default is `if (x)`, and one boolean settles it.
func TestKeywordSpaceIsSettable(t *testing.T) {
	t.Parallel()
	source := "if(x) {\n    y();\n}\n"
	formatted, ok, _, _ := FormatFile("/keyword.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if formatted != source {
		t.Fatalf("expected our tight-keyword style to survive, got:\n%s", formatted)
	}
}

// TestDoesNotAddTypeMemberSemicolons pins the second class. With Semicolons "insert" the formatter
// rewrites `{ a: number }` to `{ a: number; }`, which our Prettier never does; "ignore" is correct
// for a codebase whose semicolons are already right.
func TestDoesNotAddTypeMemberSemicolons(t *testing.T) {
	t.Parallel()
	source := "const r: Record<string, { label: string }> = {};\n"
	formatted, ok, _, _ := FormatFile("/members.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if formatted != source {
		t.Fatalf("expected no semicolon insertion, got:\n%s", formatted)
	}
}

// TestKeepsEmptyBracesTight pins the third, which is also what makes our fork's anonymous-function
// customization survive: `function() {}` must not become `function() { }`.
func TestKeepsEmptyBracesTight(t *testing.T) {
	t.Parallel()
	source := "const a = function() {};\n"
	formatted, ok, _, _ := FormatFile("/empty.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if formatted != source {
		t.Fatalf("expected tight empty braces, got:\n%s", formatted)
	}
}

// TestPreservesForkCatchAndFinallyPlacement pins our fork's other customization. Both keywords sit
// on their own line after the closing brace, and the formatter must leave that alone.
func TestPreservesForkCatchAndFinallyPlacement(t *testing.T) {
	t.Parallel()
	source := "try {\n    a();\n}\ncatch(error) {\n    b();\n}\nfinally {\n    c();\n}\n"
	formatted, ok, _, _ := FormatFile("/try.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if formatted != source {
		t.Fatalf("expected our catch/finally placement to survive, got:\n%s", formatted)
	}
}

// TestFlattensBinaryContinuationInCallArguments pins the largest remaining gap, 237 of 1,542 files.
// Prettier indents the trailing operands of a wrapped binary expression one level past the first;
// the formatter flattens them all to the first operand's column. It reproduces only inside a call
// argument, which is why the plain `const m = 'a' + 'b'` form agrees and the corpus still diverges.
//
// Nothing in FormatCodeSettings changes this, so closing it means a post-pass or accepting the
// reflow. It is whitespace-only and does not change line count.
func TestFlattensBinaryContinuationInCallArguments(t *testing.T) {
	t.Parallel()
	source := "function f() {\n    throw new Error(\n        'a' +\n            'b' +\n            'c',\n    );\n}\n"
	want := "function f() {\n    throw new Error(\n        'a' +\n        'b' +\n        'c',\n    );\n}\n"

	formatted, ok, _, _ := FormatFile("/continuation.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if formatted != want {
		t.Fatalf("continuation indentation changed shape; the differential report needs updating.\ngot:\n%s\nwant:\n%s", formatted, want)
	}
}

// TestCollapsesTemplateExpressionBreaks pins the line-count class, 97 of 1,542 files. Prettier
// breaks a long `${...}` in a template literal onto its own lines; the formatter pulls the opening
// back up and leaves the closing brace hanging. This one changes how the file reads, not just where
// its whitespace sits.
func TestCollapsesTemplateExpressionBreaks(t *testing.T) {
	t.Parallel()
	source := "const c = `x ${\n    condition ? 'a' : 'b'\n} y`;\n"
	formatted, ok, _, _ := FormatFile("/template.tsx", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if strings.Count(formatted, "\n") > strings.Count(source, "\n") {
		t.Fatalf("expected the formatter to collapse rather than expand, got:\n%s", formatted)
	}
}

// TestForStatementEmptyClauses records a formatter bug found by the sweep: `for(;;)` becomes
// `for(; ;)`, which is not valid style anywhere and is nobody's preference. Ten files in the corpus
// hit it. Recorded rather than worked around, because a reconciliation layer needs to know.
func TestForStatementEmptyClauses(t *testing.T) {
	t.Parallel()
	source := "for(;;) {\n    a();\n}\n"
	formatted, ok, _, _ := FormatFile("/for.ts", source)
	if !ok {
		t.Fatal("could not parse")
	}
	if formatted == source {
		t.Fatal("the for(;;) spacing bug appears to be fixed upstream; update the differential report")
	}
	if !strings.Contains(formatted, "for(; ;)") {
		t.Fatalf("expected the known `for(; ;)` bug, got:\n%s", formatted)
	}
}
