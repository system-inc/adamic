package rule

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// TestFixHelpersDoNotEatLeadingTrivia is the guard for the sharpest edge in this API.
//
// `node.Loc.Pos()` is the position before leading trivia, so a fix built from Loc replaces the
// whitespace or comment preceding the node along with the node itself:
//
//	import * as X from 'fs'   ->   import * as X from'node:fs'
//
// What makes this worse than cosmetic is that the corrupted output still parses. The edit engine
// refuses a rewrite that breaks syntax, and that guard never fires here, so the damage lands with
// every downstream check green. A range wider than the rule intended is damage no later validation
// can detect.
//
// Found by @system_cohere_lint_fix running real rules through the engine rather than reasoning
// about ranges.
func TestFixHelpersDoNotEatLeadingTrivia(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{"space", "import * as X from 'fs';", "import * as X from 'node:fs';"},
		{"newline", "import * as X from\n\t'fs';", "import * as X from\n\t'node:fs';"},
		{"comment", "import * as X from /* pinned */ 'fs';", "import * as X from /* pinned */ 'node:fs';"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceFile, specifier := parseModuleSpecifier(t, testCase.source)
			context := Context{SourceFile: sourceFile}

			fix := context.ReplaceNode(specifier, "'node:fs'")
			applied := testCase.source[:fix.Range.Pos()] + fix.Text + testCase.source[fix.Range.End():]

			if applied != testCase.want {
				t.Fatalf("applying the fix produced\n  %q\nwant\n  %q", applied, testCase.want)
			}
		})
	}
}

// TestRemoveNodeLeavesTriviaBehind pins a deliberate choice rather than an accident.
//
// Removing a node and removing the blank line above it are different intentions, and a helper that
// guessed would be wrong half the time. A rule that wants the surrounding whitespace gone says so
// with ReplaceRange over a range it computed.
func TestRemoveNodeLeavesTriviaBehind(t *testing.T) {
	t.Parallel()
	source := "import * as X from /* pinned */ 'fs';"
	sourceFile, specifier := parseModuleSpecifier(t, source)
	context := Context{SourceFile: sourceFile}

	fix := context.RemoveNode(specifier)
	applied := source[:fix.Range.Pos()] + source[fix.Range.End():]

	if applied != "import * as X from /* pinned */ ;" {
		t.Fatalf("RemoveNode swept up trivia: %q", applied)
	}
}

// TestInsertBeforeLandsOnTheTokenNotItsComment covers the third helper that reads Pos().
//
// Inserting before a node's trivia rather than before the node puts a modifier above the comment
// that documents the declaration, which is never what a rule means.
func TestInsertBeforeLandsOnTheTokenNotItsComment(t *testing.T) {
	t.Parallel()
	source := "import * as X from /* pinned */ 'fs';"
	sourceFile, specifier := parseModuleSpecifier(t, source)
	context := Context{SourceFile: sourceFile}

	fix := context.InsertBefore(specifier, "!")
	applied := source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]

	if applied != "import * as X from /* pinned */ !'fs';" {
		t.Fatalf("InsertBefore landed before the trivia: %q", applied)
	}
}

// TestInsertAfterNeedsNoTrimming confirms the one helper that was already correct, so a later
// refactor cannot quietly break it while fixing the others.
func TestInsertAfterNeedsNoTrimming(t *testing.T) {
	t.Parallel()
	source := "import * as X from 'fs';"
	sourceFile, specifier := parseModuleSpecifier(t, source)
	context := Context{SourceFile: sourceFile}

	fix := context.InsertAfter(specifier, "!")
	applied := source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]

	if applied != "import * as X from 'fs'!;" {
		t.Fatalf("InsertAfter moved: %q", applied)
	}
}

// TestTokenRangeSurvivesANilSourceFile keeps the helper from panicking in a harness that parsed no
// file. Untrimmed is wrong, but a crash inside a linter is worse.
func TestTokenRangeSurvivesANilSourceFile(t *testing.T) {
	t.Parallel()
	_, specifier := parseModuleSpecifier(t, "import * as X from 'fs';")

	if got := TokenRange(nil, specifier); got != specifier.Loc {
		t.Fatalf("expected the untrimmed Loc as a fallback, got %+v", got)
	}
}

func parseModuleSpecifier(t *testing.T, source string) (*ast.SourceFile, *ast.Node) {
	t.Helper()

	fileName := tspath.NormalizePath("/Fixture.ts")
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(fileName), PathKey: tspath.PathKey(fileName),
	}, source, core.ScriptKindTS)
	if sourceFile == nil {
		t.Fatalf("could not parse %q", source)
	}

	var specifier *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || specifier != nil {
			return
		}
		if node.Kind == ast.KindImportDeclaration {
			specifier = node.AsImportDeclaration().ModuleSpecifier
			return
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sourceFile.AsNode())

	if specifier == nil {
		t.Fatalf("no module specifier in %q", source)
	}
	return sourceFile, specifier
}

// TestReportNodeAnchorsOnTheTokenNotItsComment is the fifth symptom of one root cause, and the one
// that makes a finding unsuppressable.
//
// The case is real, from modules/mcp/McpApi.ts in the tree this tool gates:
//
//	const {
//	    method,
//	    // eslint-disable-next-line nexus/consistency-no-abbreviated-identifier
//	    params,
//	} = message;
//
// The suppression sits on the line directly above `params` and correctly covers it. Anchored on
// node.Loc the finding lands on the comment's line or earlier, and a `-next-line` directive only
// matches the line after itself, so no suppression the author could write would silence it.
//
// This is worse than an offset because it is invisible in a count. A finding at the wrong line
// still reads as a real finding.
func TestReportNodeAnchorsOnTheTokenNotItsComment(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"const {",
		"    method,",
		"    // a comment above the binding",
		"    params,",
		"} = message;",
	}, "\n")

	sourceFile, binding := parseNamedBinding(t, source, "params")
	context := Context{SourceFile: sourceFile}

	var reported Diagnostic
	context.Report = func(diagnostic Diagnostic) { reported = diagnostic }
	context.ReportNode(binding, Message{Id: "probe", Description: "probe"})

	line := 1 + strings.Count(source[:reported.Range.Pos()], "\n")
	if line != 4 {
		t.Fatalf("the finding anchored on line %d, want 4 (the binding). Anchoring earlier puts it on or above the suppression comment, where no -next-line directive can reach it", line)
	}
}

// TestReportNodeWithFixesAnchorsOnTheToken covers the variant that matters twice over: a finding at
// the wrong line carrying a fix means the repair lands somewhere the reader was never shown.
func TestReportNodeWithFixesAnchorsOnTheToken(t *testing.T) {
	t.Parallel()
	source := "const {\n    // a comment\n    params,\n} = message;"

	sourceFile, binding := parseNamedBinding(t, source, "params")
	context := Context{SourceFile: sourceFile}

	var reported Diagnostic
	context.Report = func(diagnostic Diagnostic) { reported = diagnostic }
	context.ReportNodeWithFixes(binding, Message{Id: "probe", Description: "probe"},
		context.ReplaceNode(binding, "parameters"))

	if got := source[reported.Range.Pos():reported.Range.End()]; got != "params" {
		t.Fatalf("the finding covers %q, want exactly the token", got)
	}
	if len(reported.Fixes) != 1 {
		t.Fatalf("expected one fix, got %d", len(reported.Fixes))
	}
	if got := source[reported.Fixes[0].Range.Pos():reported.Fixes[0].Range.End()]; got != "params" {
		t.Fatalf("the fix covers %q, want exactly the token", got)
	}
}

// parseNamedBinding finds an identifier by text, which is how these fixtures name a subject without
// depending on the shape of the tree around it.
func parseNamedBinding(t *testing.T, source string, want string) (*ast.SourceFile, *ast.Node) {
	t.Helper()

	fileName := tspath.NormalizePath("/Fixture.ts")
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(fileName), PathKey: tspath.PathKey(fileName),
	}, source, core.ScriptKindTS)
	if sourceFile == nil {
		t.Fatalf("could not parse %q", source)
	}

	var found *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found != nil {
			return
		}
		if node.Kind == ast.KindIdentifier && node.Text() == want {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sourceFile.AsNode())

	if found == nil {
		t.Fatalf("no identifier %q in %q", want, source)
	}
	return sourceFile, found
}
