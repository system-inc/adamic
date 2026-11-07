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

func TestDebugIRFlow(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../oracle/testdata/graph_regions_list.a")
	if readError != nil {
		test.Fatal(readError)
	}
	path := filepath.Join(test.TempDir(), "debug.a")
	if writeError := os.WriteFile(path, source, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	program, lowerWithOptionsError := lower.LowerWithOptions(ctx, loaded, lower.Options{Debug: true})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	for index := -1; index < len(program.Functions); index++ {
		graph := flow.Build(program, index)
		flow.Construct(graph)
		if problems := flow.VerifySSA(graph); len(problems) > 0 {
			test.Fatal(problems)
		}
		flow.InferAliasingEffects(graph)
		flow.InferMutableRanges(graph)
	}
	flow.CheckDebugProgram(test, path, program)
	test.Logf("debug IR: main and %d functions passed SSA, ranges, graph paths and liveness", len(program.Functions))
}
