package flow

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestDebuggerHasNoFlowInstruction(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Main: []ir.Statement{ir.Debugger{}, ir.Evaluate{Value: ir.NumberConstant{Value: 42}}}}
	graph := Build(program, -1)
	instructions := 0
	for _, block := range graph.Blocks {
		instructions += len(block.Instructions)
	}
	if instructions != 1 {
		t.Fatalf("debugger changed the flow: %d instructions, want 1", instructions)
	}
}
