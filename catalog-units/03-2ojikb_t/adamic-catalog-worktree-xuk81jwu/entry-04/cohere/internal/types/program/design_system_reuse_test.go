package program_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// designSystemRule reads the design system the way the Tailwind rules do, through DesignSystemFS: it flags
// every declaration when theme.css says "flag", or when a candidate stylesheet it looked for and missed exists.
func designSystemRule(root string) rule.Rule {
	return rule.Rule{
		Name:         "fixture/design-system",
		ProgramReads: rule.ReadsDesignSystem,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
				fileSystem := ctx.Program.DesignSystemFS()
				theme, _ := fileSystem.ReadFile(tspath.RootedFilePath(filepath.Join(root, "theme.css")))
				// Asked every time, not after the theme in an ||: a question never asked is not in the read set.
				candidate := fileSystem.FileExists(tspath.RootedFilePath(filepath.Join(root, "candidate.css")))
				if strings.Contains(theme, "flag") || candidate {
					ctx.ReportNode(node, rule.Message{Id: "flagged", Description: "the design system flags this"})
				}
			}}
		},
	}
}

// The design-system rules replay on an unchanged file while every path the design system read is as it was,
// and run again on every file when one is not: a stylesheet edited, a candidate it missed created, or one
// removed (#35nqkwc). A Tailwind rule's finding on an unchanged .tsx moves with the stylesheets, so a replay
// over any of those would be a stale verdict on a file nobody touched.
//
// Each step is checked against an uncached walk of the same tree, and the count of re-runs is checked both
// ways: re-running when nothing it read changed costs the saving, and replaying when something did is the
// failure this cache must never have.
func TestDesignSystemRulesReplayUntilWhatTheDesignSystemReadChanges(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"tsconfig.json": minimalConfig,
		"theme.css":     "flag\n",
		"a.ts":          "export const a = 1;\n",
		"b.ts":          "export const b = 2;\n",
	})
	rules := []rule.Rule{designSystemRule(root)}
	write := func(name string, contents string) func() {
		return func() {
			if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	first, recorded := walkAndRecord(t, root, rules, nil)
	if len(first.Diagnostics) != 2 {
		t.Fatalf("the fixture flagged %d declarations, not 2, so what follows proves nothing", len(first.Diagnostics))
	}
	if recorded.DesignSystem == nil {
		t.Fatal("the first walk recorded no design system")
	}
	reads := map[string]bool{}
	for _, read := range recorded.DesignSystem.Reads {
		reads[filepath.Base(read.Path)] = read.Present
	}
	if present, seen := reads["theme.css"]; !seen || !present {
		t.Fatalf("theme.css is not in the recorded reads as present: %v", reads)
	}
	if present, seen := reads["candidate.css"]; !seen || present {
		t.Fatalf("candidate.css is not in the recorded reads as absent: %v", reads)
	}

	steps := []struct {
		name      string
		edit      func()
		wantRerun bool
	}{
		{"nothing changed", func() {}, false},
		{"a .ts file edited", write("a.ts", "export const a = 3;\n"), false},
		{"the stylesheet edited", write("theme.css", "plain\n"), true},
		{"nothing changed after the stylesheet edit", func() {}, false},
		{"a candidate it missed created", write("candidate.css", ""), true},
		{"that candidate removed", func() {
			if err := os.Remove(filepath.Join(root, "candidate.css")); err != nil {
				t.Fatal(err)
			}
		}, true},
	}
	for _, step := range steps {
		step.edit()
		result, next := walkAndRecord(t, root, rules, recorded)
		truth := plainWalk(t, root, rules)
		if !reflect.DeepEqual(diagnosticKeys(result.Diagnostics), diagnosticKeys(truth.Diagnostics)) {
			t.Errorf("%s: the cached walk differs from an uncached one:\n cached %v\n truth  %v",
				step.name, diagnosticKeys(result.Diagnostics), diagnosticKeys(truth.Diagnostics))
		}
		if !reflect.DeepEqual(foldedCoverage(result.Coverage), foldedCoverage(truth.Coverage)) {
			t.Errorf("%s: coverage differs from an uncached walk", step.name)
		}
		if result.FilesReplayed == 0 {
			t.Fatalf("%s: no file was replayed, so whether the design-system rules were cannot be told", step.name)
		}
		switch {
		case step.wantRerun && result.DesignSystemRerun != result.FilesReplayed:
			t.Errorf("%s: the design-system rules ran again on %d of %d replayed files; something they read changed, so all must",
				step.name, result.DesignSystemRerun, result.FilesReplayed)
		case !step.wantRerun && result.DesignSystemRerun != 0:
			t.Errorf("%s: the design-system rules ran again on %d replayed files though nothing they read changed",
				step.name, result.DesignSystemRerun)
		}
		recorded = next
	}
}

// Another program loading its design system in the middle of this walk does not make this walk's read as
// never loaded. If it did, this run's design-system findings would be recorded as independent of every
// stylesheet, and the edit below would replay them stale (#35nqkwc, @system_cohere_lint's review).
func TestAnotherProgramsDesignSystemDoesNotHideThisWalks(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"tsconfig.json": minimalConfig,
		"theme.css":     "flag\n",
		"a.ts":          "export const a = 1;\n",
		"b.ts":          "export const b = 2;\n",
	})
	elsewhere := writeProject(t, map[string]string{"tsconfig.json": minimalConfig, "c.ts": "export const c = 3;\n"})
	other, err := program.Build(program.Options{ConfigFileName: filepath.Join(elsewhere, "tsconfig.json")})
	if err != nil {
		t.Fatal(err)
	}
	inner := designSystemRule(root)
	interleaved := inner
	interleaved.Run = func(ctx rule.Context, options any) rule.Listeners {
		// A concurrent graph asking for its own design system after each time this walk asks for its own, so
		// the last program to ask before the walk ends is never this one.
		listeners := rule.Listeners{}
		for kind, listener := range inner.Run(ctx, options) {
			listeners[kind] = func(node *ast.Node) {
				listener(node)
				rule.ViewProgram(other.Program, nil, inner).DesignSystemFS()
			}
		}
		return listeners
	}
	rules := []rule.Rule{interleaved}

	_, recorded := walkAndRecord(t, root, rules, nil)
	if recorded.DesignSystem == nil || len(recorded.DesignSystem.Reads) == 0 {
		t.Fatalf("this walk's design system was recorded as never loaded: %+v", recorded.DesignSystem)
	}
	if err := os.WriteFile(filepath.Join(root, "theme.css"), []byte("plain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, _ := walkAndRecord(t, root, rules, recorded)
	truth := plainWalk(t, root, rules)
	if !reflect.DeepEqual(diagnosticKeys(result.Diagnostics), diagnosticKeys(truth.Diagnostics)) {
		t.Errorf("a stylesheet edit replayed stale design-system findings:\n cached %v\n truth  %v",
			diagnosticKeys(result.Diagnostics), diagnosticKeys(truth.Diagnostics))
	}
	if result.FilesReplayed == 0 || result.DesignSystemRerun != result.FilesReplayed {
		t.Errorf("the design-system rules ran again on %d of %d replayed files after the stylesheet edit",
			result.DesignSystemRerun, result.FilesReplayed)
	}
}
