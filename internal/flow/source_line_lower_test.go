package flow_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestDebugIRFlow(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/graph_regions_list.a")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "debug.a")
	if err := os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	program, err := lower.LowerWithOptions(ctx, loaded, lower.Options{Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	for index := -1; index < len(program.Functions); index++ {
		graph := flow.Build(program, index)
		flow.Construct(graph)
		if problems := flow.VerifySSA(graph); len(problems) > 0 {
			t.Fatal(problems)
		}
		flow.InferAliasingEffects(graph)
		flow.InferMutableRanges(graph)
	}
	flow.CheckDebugProgram(t, path, program)
	t.Logf("debug IR: main and %d functions passed SSA, ranges, graph paths and liveness", len(program.Functions))
}
