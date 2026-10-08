//go:build lintoracle

// Overlay beside Go HIR. This witness exports an actual Go pass boundary.
package high_level_intermediate_representation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUnit9ScopeBoundary(t *testing.T) {
	t.Parallel()
	fn := NewFunction(nil, "scopeWitness", FunctionKind(0))
	loop := fn.NewBlock(BlockKind(0))
	body := fn.NewBlock(BlockKind(0))
	exit := fn.NewBlock(BlockKind(0))
	fn.Entry = loop.Id
	loop.Terminal = &While{Test: body.Id, Loop: body.Id, Fallthrough: exit.Id}
	body.Terminal = &Scope{Scope: 7, Block: body.Id, Fallthrough: exit.Id}
	exit.Terminal = &Unreachable{}
	before := oracleDump(fn)
	flattened := FlattenReactiveLoops(fn)
	if len(flattened) != 1 || !flattened[7] {
		t.Fatalf("flattened scopes = %v, want {7:true}", flattened)
	}
	after := oracleDump(fn)
	if before != after {
		t.Fatal("loop flattening must retain graph terminals")
	}
	dir := os.Getenv("HIR_UNIT9_OUTPUT")
	if dir == "" {
		t.Fatal("HIR_UNIT9_OUTPUT is required")
	}
	for name, text := range map[string]string{"before.hir.txt": before, "after.hir.txt": after} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	result, err := json.Marshal(flattened)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "flattened.json"), append(result, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Log("Go boundary: one scope pruned, graph unchanged")
}
