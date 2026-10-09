package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// These reductions isolate the receiver assertion from unrelated whole-project
// checker errors. Original positions and owners are retained in the report.
func TestReceiverAssertionRouting(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ id, function, source string }{
		{"D212", "transformGenerators.hasImmediateContainingLabeledBlock", `function hasImmediateContainingLabeledBlock(): number { const blockStack: number[] = [7]; const j = 0; const containingBlock: number = blockStack![j]; return containingBlock; } console.log(String(hasImmediateContainingLabeledBlock()));`},
		{"D220", "transformGenerators.tryEnterOrLeaveBlock", `function tryEnterOrLeaveBlock(): number { const blockOffsets: number[] = [7]; const blockIndex = 0; const offset: number = blockOffsets![blockIndex]; return offset; } console.log(String(tryEnterOrLeaveBlock()));`},
	} {
		t.Run(fixture.id, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "generators.ts")
			for name, source := range map[string]string{"generators.ts": fixture.source, "tsconfig.json": `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":["generators.ts"]}`} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			checks := ir.InsertedChecks(lowered)
			count := 0
			for _, check := range checks {
				if check.Kind == "indexed-presence" {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("want one indexed check after redundant assertion, got %+v", checks)
			}
			t.Logf("ROUTE %s file=src/compiler/transformers/generators.ts function=%s redundant assertion lowers; indexed-presence=%d", fixture.id, fixture.function, count)
		})
	}
}
