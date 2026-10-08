package control_flow_graph

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Keep this workload unchanged across the numeric-to-wrapper migration so the
// runtime counters expose the price of allocating an index for each block.
func TestArenaAllocationCounts(t *testing.T) {
	for _, blocks := range []int{0, 1, 1000} {
		t.Run(fmt.Sprintf("blocks_%d", blocks), func(t *testing.T) {
			dir := cfgCopy(t)
			entry := filepath.Join(dir, "arena_counts.a")
			cfgWrite(t, entry, []byte(fmt.Sprintf("import { BlockArena } from './arena.a';\nconst arena=new BlockArena();for(let i=0;i<%d;i++)arena.allocate();console.log(`${arena.length}`);arena.dispose();\n", blocks)))
			program, err := load.Load([]string{entry})
			if err != nil {
				t.Fatal(err)
			}
			ir, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "arena_counts")
			if err := native.Build(native.C(ir), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary)
			var stderr bytes.Buffer
			command.Stderr = &stderr
			out, err := command.Output()
			if err != nil || string(out) != fmt.Sprintf("%d\n", blocks) || !strings.HasPrefix(stderr.String(), "adamic: counts: allocations ") {
				t.Fatalf("arena counter workload failed: %v %s %s", err, out, stderr.String())
			}
			t.Logf("%d allocated blocks, disposed as one arena: %s", blocks, strings.TrimSpace(stderr.String()))
		})
	}
}
