// Package rule_testing runs a rule against source text and reports what it found.
//
// Every rule in cohere ships a fixture pair: source that must produce a finding, and source that
// must produce none. Both directions, always, in the same commit as the rule.
//
// That is not ceremony. During the migration this tool replaces, a class-name filter that measured
// fifteen times faster silently stopped reporting one real collapse, and a hand-written identifier
// splitter silently stopped reporting another. Both were caught by fixtures. Neither was caught by
// reading the diff. The inverse defect, a rule that fires on correct code, had no guard at all,
// because the corpus was violations only.
//
// A rule that has never been shown to fire has not been shown to work.
package rule_testing

import (
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/docsdata/capture"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// Result is what one rule saw in one file.
type Result struct {
	Diagnostics []rule.Diagnostic
	SourceFile  *ast.SourceFile

	// capture is the case the docs capture records when an assertion passes on this result; see
	// docs_capture.go. Nil unless COHERE_DOCS_CAPTURE is set.
	capture *capturedRun
}

// MessageIds returns the ids of every finding, in the order they were reported.
func (r Result) MessageIds() []string {
	ids := make([]string, 0, len(r.Diagnostics))
	for _, diagnostic := range r.Diagnostics {
		ids = append(ids, diagnostic.Message.Id)
	}
	return ids
}

// Run parses source text and walks it with one rule, with no type information.
//
// Rules that need a checker are exercised against a real program elsewhere; this path exists for
// the syntax-only majority, where a full program would cost seconds to prove something a parse can
// prove in microseconds.
func Run(t *testing.T, subject rule.Rule, fileName string, sourceText string) Result {
	t.Helper()
	return RunWithOptions(t, subject, fileName, sourceText, nil)
}

// RunWithOptions is Run for a rule that reads configuration.
//
// A configurable rule has a failure mode an unconfigured one does not: it can be registered without
// its required option and then guard either everything or nothing. Which of those it does is a real
// decision that belongs in fixtures, so the harness has to be able to hand a rule no options at all
// as deliberately as it hands it good ones.
func RunWithOptions(t *testing.T, subject rule.Rule, fileName string, sourceText string, options any) Result {
	t.Helper()
	captured := newCapturedRun(subject, fileName, 0, options)

	// A `.jsx` fixture has to parse as JSX too, and it did not until a fixture asked it to.
	//
	// Parsing `.jsx` as plain TypeScript does not fail: `return <div />;` is read as a type
	// assertion and the tree simply contains no JSX node. So a React-gated rule that correctly
	// declines to find a component in it reports nothing, the fixture reads as a rule that does not
	// fire, and the defect looks like it is in the rule. Measured directly: the same source parsed
	// as TS finds no JSX element, and as TSX or JSX finds one.
	//
	// Both extensions are real. `FileContextFor` treats `.jsx` as a React file, so any rule gated
	// on that predicate can be reached by one, and a fixture is the only place that gets checked.
	scriptKind := core.ScriptKindTS
	switch {
	case strings.HasSuffix(fileName, ".tsx"):
		scriptKind = core.ScriptKindTSX
	case strings.HasSuffix(fileName, ".jsx"):
		scriptKind = core.ScriptKindJSX
	case strings.HasSuffix(fileName, ".js"):
		scriptKind = core.ScriptKindJS
	}

	// typescript-go panics on a relative filename: the parser stores it as an identity and expects
	// it to be normalized and absolute. A fixture names a file for readability rather than for the
	// disk, so it is rooted here instead of at every call site.
	if !tspath.IsRootedDiskPath(fileName) {
		fileName = "/" + strings.TrimPrefix(fileName, "/")
	}
	fileName = tspath.NormalizePath(fileName)

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(fileName),
		PathKey:  tspath.PathKey(fileName),
	}, sourceText, scriptKind)
	if sourceFile == nil {
		t.Fatalf("could not parse %s", fileName)
	}

	var diagnostics []rule.Diagnostic
	context := rule.Context{
		SourceFile: sourceFile,
		FileCache:  rule.NewFileCache(),
		Report: func(diagnostic rule.Diagnostic) {
			diagnostic.RuleName = subject.Name
			diagnostics = append(diagnostics, diagnostic)
		},
	}

	listeners := subject.Run(context, options)
	if listeners != nil {
		walk(sourceFile.AsNode(), listeners)
	}

	expectEachFixParses(t, sourceFile, diagnostics)
	return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}
}

// walk visits every node, calling any listener registered for its kind.
//
// One traversal serves every rule in a real run. This is the single-rule form of that walk, kept
// deliberately simple so a fixture failure points at the rule rather than at the harness.
func walk(node *ast.Node, listeners rule.Listeners) {
	if node == nil {
		return
	}
	if listener, isListening := listeners[node.Kind]; isListening {
		listener(node)
	}
	node.ForEachChild(func(child *ast.Node) bool {
		walk(child, listeners)
		return false
	})
}

// ExpectFindings asserts that a rule reported exactly these message ids, in order.
//
// Order matters because a rule reporting the right findings in the wrong place is still wrong, and
// because two defects on one line must both be reported rather than one hiding the other.
func ExpectFindings(t *testing.T, result Result, wantIds ...string) {
	t.Helper()

	gotIds := result.MessageIds()
	if len(gotIds) != len(wantIds) {
		t.Fatalf("expected %d findings %v, got %d %v", len(wantIds), wantIds, len(gotIds), gotIds)
	}
	for index, want := range wantIds {
		if gotIds[index] != want {
			t.Fatalf("finding %d: expected %q, got %q (all: %v)", index, want, gotIds[index], gotIds)
		}
	}
	recordCase(t, result, capture.OutcomeFindings, "")
}

// ExpectClean asserts that a rule stayed silent.
//
// This is the half the old corpus never had, and the half that catches a rule which fires on
// correct code. A rule with no clean fixture has been shown to detect, never to discriminate.
func ExpectClean(t *testing.T, result Result) {
	t.Helper()

	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no findings, got %d: %v", len(result.Diagnostics), result.MessageIds())
	}
	recordCase(t, result, capture.OutcomeClean, "")
}

// ExpectSameFindings asserts that two runs of a rule reported the same findings: the same message
// ids, in the same order, over the same spans. It is how a fixture shows an Adamic `.a` file is
// linted exactly as the same source named `.ts` (#kwt1htp), so the control run must report at least
// one finding, or the comparison would pass with the rule silent on both.
func ExpectSameFindings(t *testing.T, control Result, treatment Result) {
	t.Helper()

	if len(control.Diagnostics) == 0 {
		t.Fatal("the control run reported nothing, so agreeing with it proves nothing")
	}
	if len(control.Diagnostics) != len(treatment.Diagnostics) {
		t.Fatalf("expected %d findings %v, got %d %v",
			len(control.Diagnostics), control.MessageIds(), len(treatment.Diagnostics), treatment.MessageIds())
	}
	for index, want := range control.Diagnostics {
		got := treatment.Diagnostics[index]
		if got.Message.Id != want.Message.Id || got.Range != want.Range {
			t.Fatalf("finding %d: expected %q at %v, got %q at %v",
				index, want.Message.Id, want.Range, got.Message.Id, got.Range)
		}
	}
}

// ExpectFixedSource applies every fix the rule proposed and asserts the resulting source text.
//
// A fix is the one part of a rule that rewrites source, and a fixture checking only message ids
// cannot see it. Measured on this repository: corrupting the replacement text in
// `import-require-node-namespace` so its fix writes `'CORRUPTED:fs'` instead of `'node:fs'` leaves
// the whole suite green. That rule shipped as the first rule ported, and the gap sat there for the
// entire migration, because until now the harness had no way to say what a fix should produce.
//
// Fixes are applied back to front so that an earlier fix's replacement cannot move the offsets a
// later one was computed against. That is the same ordering the fix engine uses, for the same
// reason, and doing it differently here would test a rewrite nobody performs.
//
// Overlapping fixes are refused rather than resolved. The engine has an overlap policy and this is
// not the place to reimplement it: a fixture that silently applied one of two conflicting fixes
// would assert a result the real pipeline never produces.
func ExpectFixedSource(t *testing.T, result Result, wantSource string) {
	t.Helper()

	type pending struct {
		start, end int
		text       string
	}
	fixes := []pending{}
	for _, diagnostic := range result.Diagnostics {
		for _, fix := range diagnostic.Fixes {
			fixes = append(fixes, pending{start: fix.Range.Pos(), end: fix.Range.End(), text: fix.Text})
		}
	}

	if len(fixes) == 0 {
		// A rule that proposed no fix cannot have its rewrite asserted, and passing here would make
		// this function agree with any expectation at all. That is the vacuous shape the fixture
		// pair exists to refuse, arriving through the assertion meant to close it.
		t.Fatalf("the rule proposed no fixes, so there is no rewrite to check against %q", wantSource)
	}

	sort.Slice(fixes, func(first, second int) bool {
		return fixes[first].start > fixes[second].start
	})

	source := result.SourceFile.Text()
	previousStart := len(source)
	for _, fix := range fixes {
		if fix.end > previousStart {
			t.Fatalf("fixes overlap at [%d,%d): the engine resolves overlaps and a fixture must not guess which one wins",
				fix.start, fix.end)
		}
		if fix.start < 0 || fix.end > len(source) || fix.start > fix.end {
			t.Fatalf("fix range [%d,%d) is outside the source, which is a defect in the rule rather than in the test",
				fix.start, fix.end)
		}
		source = source[:fix.start] + fix.text + source[fix.end:]
		previousStart = fix.start
	}

	if source != wantSource {
		t.Errorf("the applied fixes produced:\n  %q\nwant:\n  %q", source, wantSource)
		return
	}
	recordCase(t, result, capture.OutcomeFixed, source)
}
