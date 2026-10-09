package program_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// TestWalkReturnsFindingsInFileThenPositionOrder pins the order a walk returns its findings in, and
// the set it returns, as two separate assertions.
//
// # Why both, and why the order is not left to the scheduler
//
// Files are walked in parallel and each worker merges under a mutex in whichever order it finishes,
// so before this the output carried goroutine scheduling: three runs over one unchanged tree printed
// three orderings of one identical set (#zevtfkq). Two runs could not be diffed, and anything
// comparing them had to sort first or report a change nobody made.
//
// Order alone is one instrument with a blind side. Byte-identical output is satisfied by a correct
// sort and by a change that stabilises order by dropping findings, so the exact set is asserted as
// well, built here from the fixture rather than read back from the walk.
//
// The fixture is shaped so the unsorted walk fails deterministically rather than by scheduling luck:
// one rule reports a file's declarations last to first from the file node, which no merge order can
// put in position order, and the rule that sorts first by name is dispatched second.
func TestWalkReturnsFindingsInFileThenPositionOrder(t *testing.T) {
	t.Parallel()
	const fileCount = 12
	const declarationCount = 3

	files := map[string]string{"tsconfig.json": minimalConfig}
	for fileIndex := range fileCount {
		var text strings.Builder
		for declarationIndex := range declarationCount {
			fmt.Fprintf(&text, "export const value%d = %d;\n", declarationIndex, declarationIndex)
		}
		files[fmt.Sprintf("file%02d.ts", fileIndex)] = text.String()
	}
	directory := writeProject(t, files)

	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}

	declarationName := func(node *ast.Node) string {
		return node.AsVariableDeclaration().Name().Text()
	}

	// Dispatched first, sorts last: reports every declaration twice, message `b` before `a`, so the
	// message id is the only thing separating the pair.
	forward := rule.Rule{
		Name: "zeta-forward",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
				ctx.ReportNode(node, rule.Message{Id: "b", Description: declarationName(node)})
				ctx.ReportNode(node, rule.Message{Id: "a", Description: declarationName(node)})
			}}
		},
	}
	// Dispatched second, sorts first: reports from the file node, last declaration first.
	reverse := rule.Rule{
		Name: "alpha-reverse",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
				var declarations []*ast.Node
				var collect func(*ast.Node)
				collect = func(current *ast.Node) {
					if current.Kind == ast.KindVariableDeclaration {
						declarations = append(declarations, current)
					}
					current.ForEachChild(func(child *ast.Node) bool {
						collect(child)
						return false
					})
				}
				collect(node)
				slices.Reverse(declarations)
				for _, declaration := range declarations {
					ctx.ReportNode(declaration, rule.Message{Id: "only", Description: declarationName(declaration)})
				}
			}}
		},
	}

	// What the walk must return, built from the fixture: files by name, declarations by position,
	// then rule name, then message id.
	var want []string
	for fileIndex := range fileCount {
		for declarationIndex := range declarationCount {
			name := fmt.Sprintf("value%d", declarationIndex)
			prefix := fmt.Sprintf("file%02d.ts %s ", fileIndex, name)
			want = append(want, prefix+"alpha-reverse/only", prefix+"zeta-forward/a", prefix+"zeta-forward/b")
		}
	}

	describe := func(result program.Result) []string {
		described := make([]string, 0, len(result.Diagnostics))
		for _, diagnostic := range result.Diagnostics {
			fileName := diagnostic.SourceFile.FileName()
			described = append(described, fmt.Sprintf("%s %s %s/%s",
				fileName[strings.LastIndexByte(fileName.AsString(), '/')+1:], diagnostic.Message.Description,
				diagnostic.RuleName, diagnostic.Message.Id))
		}
		return described
	}

	var first []string
	for run := range 5 {
		result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{forward, reverse})
		if err != nil {
			t.Fatalf("walking: %v", err)
		}
		got := describe(result)

		// Preservation, compared as a set so it is independent of the order assertion below.
		sortedGot, sortedWant := slices.Clone(got), slices.Clone(want)
		slices.Sort(sortedGot)
		slices.Sort(sortedWant)
		if !slices.Equal(sortedGot, sortedWant) {
			t.Fatalf("run %d returned %d findings, want the %d the fixture produces; the set changed, "+
				"which no ordering fix is allowed to do", run, len(got), len(want))
		}

		// Order: the stated order, not merely the same order twice.
		if !slices.Equal(got, want) {
			t.Fatalf("run %d returned findings out of file-then-position order\n got: %s\nwant: %s",
				run, strings.Join(got[:min(len(got), 9)], " | "), strings.Join(want[:9], " | "))
		}

		// Reproducibility across runs, with its own vacuity control: two empty runs agree trivially.
		if run == 0 {
			if len(got) == 0 {
				t.Fatal("the first run returned no findings, so comparing runs would prove nothing")
			}
			first = got
		} else if !slices.Equal(got, first) {
			t.Fatalf("run %d returned a different sequence from run 0", run)
		}
	}
}
