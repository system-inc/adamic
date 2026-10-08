package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// These reductions isolate the receiver assertion from unrelated whole-project
// checker errors. Original positions and owners are retained in the report.
func TestReceiverAssertionRouting(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ id, function, source string }{
		{"D212", "transformGenerators.hasImmediateContainingLabeledBlock", `function hasImmediateContainingLabeledBlock(): number { const blockStack: number[] = [7]; const j = 0; const containingBlock: number = blockStack![j]; return containingBlock; } console.log(hasImmediateContainingLabeledBlock());`},
		{"D220", "transformGenerators.tryEnterOrLeaveBlock", `function tryEnterOrLeaveBlock(): number { const blockOffsets: number[] = [7]; const blockIndex = 0; const offset: number = blockOffsets![blockIndex]; return offset; } console.log(tryEnterOrLeaveBlock());`},
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
			_, err = lower.Lower(context.Background(), checked)
			const message = "Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case"
			if err == nil || !strings.Contains(err.Error(), message) {
				t.Fatalf("want named assertion refusal, got %v", err)
			}
			t.Logf("ROUTE %s file=src/compiler/transformers/generators.ts function=%s reduction=%v", fixture.id, fixture.function, err)
		})
	}
}
