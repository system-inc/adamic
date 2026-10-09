package rule_testing

import (
	"sort"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// expectEachFixParses applies each finding's fix on its own to the source it was proposed against,
// and fails the test when a source that parsed stops parsing.
//
// The engine refuses such a fix at write time (edit.Parses, "the rewritten file does not parse"), so
// nothing breaks on disk, but the user sees a refused fix and the rule has proposed a wrong one. Three
// rules did that on trpc and no fixture of theirs noticed, because a fixture asserts the rewrite it
// expects on the shapes its author thought of, and the broken shapes were ones nobody wrote: a type
// parameter list ending in a trailing comma, a JSDoc comment with code after it on its line, an arrow
// leading an expression statement (#kq9vtva). Asked here, the question covers every fixture of every
// rule, whatever its author was testing, so the next rule like those fails in the suite and not on
// someone else's repository.
//
// One finding's edits are applied together, as the engine judges a fix alone (breaksParseAlone);
// edits from different findings are not combined, since resolving their overlaps is the engine's job.
// A source that did not parse to begin with proves nothing either way and is skipped.
func expectEachFixParses(t *testing.T, subject *ast.SourceFile, diagnostics []rule.Diagnostic) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if len(diagnostic.Fixes) == 0 {
			continue
		}
		sourceFile := subject
		if diagnostic.SourceFile != nil {
			sourceFile = diagnostic.SourceFile
		}
		fileName := sourceFile.FileName()
		source := sourceFile.Text()
		if parses, _ := edit.Parses(fileName.AsString(), source); !parses {
			continue
		}

		edits := append([]rule.Fix(nil), diagnostic.Fixes...)
		sort.Slice(edits, func(first, second int) bool {
			return edits[first].Range.Pos() > edits[second].Range.Pos()
		})
		fixed := source
		previousStart := len(source)
		applicable := true
		for _, fix := range edits {
			// A range outside the source or overlapping its neighbour is ExpectFixedSource's finding,
			// with its own message, rather than a parse question.
			if fix.Range.Pos() < 0 || fix.Range.Pos() > fix.Range.End() || fix.Range.End() > previousStart {
				applicable = false
				break
			}
			fixed = fixed[:fix.Range.Pos()] + fix.Text + fixed[fix.Range.End():]
			previousStart = fix.Range.Pos()
		}
		if !applicable {
			continue
		}
		if parses, reason := edit.Parses(fileName.AsString(), fixed); !parses {
			t.Errorf("%s proposes a fix that does not parse (%s). The source after it:\n%s",
				diagnostic.RuleName, reason, fixed)
		}
	}
}
