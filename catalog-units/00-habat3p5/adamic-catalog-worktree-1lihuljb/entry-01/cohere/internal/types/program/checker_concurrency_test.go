package program_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// A rule that asks the checker a question must not crash the run, and this needs enough files to
// actually run several workers.
//
// The failure this guards is not a panic. `dispatchFileSafely` recovers a panicking rule into one
// lost file with a coverage note, and that recover does nothing here: calling the checker from the
// parallel walk produced `fatal error: concurrent map read and map write` inside upstream's own
// `core.LinkStore.Get`, and a concurrent map access is a runtime fatal error rather than a panic. The
// process died with no coverage line, no phase line, and no partial result.
//
// It went unnoticed because **no rule in the catalog had ever called the checker.** Ninety-three
// rules, zero type queries, so the type-aware path had never executed. Every existing guard here
// exercises only code that exists.
//
// The cause was `GetTypeCheckerForFile` rather than `GetTypeCheckerForFileExclusive`. Upstream
// documents the difference and we were on the wrong side of it: the exclusive form is "locked to the
// current thread to prevent data races from multiple threads accessing the same checker". The walk
// already acquired per file and released after, which looked correct and was not, because ownership
// is by position in the program's full file list while the walk strides over a filtered subset.
//
// The existing checker test uses one file, so it runs one worker and could never have caught this.
// **The file count is the instrument here**, which is why it is written out rather than trimmed to
// something tidier.
func TestCheckerSurvivesTheParallelWalk(t *testing.T) {
	t.Parallel()
	files := map[string]string{"tsconfig.json": minimalConfig}

	// Enough files, cross-importing, that several workers run at once and reach across checkers.
	// Sixty-four standalone files was not enough: the fix reverted, that version passed five of five,
	// because tiny independent files finish before two workers can collide on one checker.
	const fileCount = 400
	for index := range fileCount {
		files[fmt.Sprintf("file%03d.ts", index)] = fmt.Sprintf(
			"import type { Value%03d } from './file%03d';\n"+
				"export type Value%03d = { count: number; label: string };\n"+
				"export const value%03d: Value%03d = { count: %d, label: 'ahra' };\n"+
				"export function use%03d(other: Value%03d): string { return other.label; }\n",
			(index+1)%fileCount, (index+1)%fileCount,
			index, index, index, index, index, (index+1)%fileCount,
		)
	}

	directory := writeProject(t, files)
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}

	if graph.Workers() < 2 {
		// One worker cannot race, so a pass here would prove nothing. This is the same shape as a
		// sweep whose probe never ran.
		t.Skipf("this machine gave the walk %d worker, so nothing ran concurrently", graph.Workers())
	}

	// Counted atomically: the listeners run on every worker at once, which is the point of the test.
	var queries atomic.Int64
	probe := rule.Rule{
		Name: "test-checker-under-concurrency",

		// Declared, because the walk hands out a checker only to rules that ask for one. Without
		// this the probe is dispatched with a nil checker and the test fails on its own first
		// assertion — which is the guard working, and is why that assertion is worth keeping.
		NeedsTypeChecker: true,

		Run: func(ctx rule.Context, options any) rule.Listeners {
			if ctx.TypeChecker == nil {
				t.Error("a rule was dispatched without a checker")
				return nil
			}
			return rule.Listeners{
				// Every identifier, deliberately. A rule would not query this indiscriminately, and
				// that is the point: it maximizes the chance of two workers reaching one checker.
				ast.KindIdentifier: func(node *ast.Node) {
					if ctx.TypeChecker.GetTypeAtLocation(node) != nil {
						queries.Add(1)
					}
				},
			}
		},
	}

	projectFiles := graph.ProjectFiles()
	if len(projectFiles) < 2 {
		t.Fatalf("expected the project to hold many files, got %d", len(projectFiles))
	}

	if _, err := graph.Walk(context.Background(), projectFiles, []rule.Rule{probe}); err != nil {
		t.Fatalf("walking with a type-querying rule: %v", err)
	}

	if queries.Load() == 0 {
		// Surviving a walk that asked nothing is exactly the vacuous pass this file exists to avoid.
		t.Fatal("the probe rule completed without answering a single type query, so nothing was proven")
	}
}
