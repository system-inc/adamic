package lower

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
)

// These scanner comparisons belong to the ahead-of-time product phase, not a
// 60-second unit. A failed comparison never publishes a successful record.
// Keep the existing opt-in Program stress policy: these products cover exactly
// the modes previously required by the equality shards, including both modes
// for objects-512 and objects-1024. Explicit stress audits still run directly.
var cycleEqualityProducts = map[string]string{
	"review/compiler/loom-implied-target/followup/objects-512.a":  "Objects512",
	"review/compiler/loom-implied-target/followup/objects-1024.a": "Objects1024",
	"review/compiler/loom-implied-target/followup/objects-2048.a": "Objects2048",
	"review/compiler/loom-implied-target/objects.a":               "Objects",
	"review/compiler/loom-implied-target/objects-sequential.a":    "ObjectsSequential",
	"internal/oracle/testdata/moves/accepted/objects.a":           "MovesObjects",
}

func cycleEqualityProduct(t *testing.T, path string) string {
	t.Helper()
	name, ok := cycleEqualityProducts[path]
	if !ok {
		t.Fatalf("no cycle equality product for %s", path)
	}
	modes := []bool{false, true}
	if cycleProgramStressInputs[path] {
		modes = []bool{false}
	}
	expected := fmt.Sprintf("%s: equal scanning/indexed outcomes in modes %v\n", path, modes)
	directory := buildcache.Product(t, buildcache.Inputs{
		Name:      "cycle-equality-" + name,
		Files:     append(compilerinputs.Lowering(), "review", "stage3/fixtures/cycles", path),
		Toolchain: []string{runtime.Version()},
	}, func(directory string) error {
		cycleCompareProgramModes(t, path, modes...)
		return os.WriteFile(filepath.Join(directory, "outcome"), []byte(expected), 0600)
	})
	data, err := os.ReadFile(filepath.Join(directory, "outcome"))
	if err != nil || string(data) != expected {
		t.Fatalf("cycle equality record for %s: %q, %v", path, data, err)
	}
	return directory
}

func cycleCompareProgramRecorded(t *testing.T, path string) {
	t.Helper()
	if _, ok := cycleEqualityProducts[path]; ok {
		cycleEqualityProduct(t, path)
		if cycleProgramStressInputs[path] && os.Getenv("ADAMIC_CYCLE_PROGRAM_STRESS") == "1" {
			cycleCompareProgramModes(t, path, true)
		}
		return
	}
	cycleCompareProgram(t, path)
}

func TestProduct_CycleEqualityObjects512(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "review/compiler/loom-implied-target/followup/objects-512.a")
}

func TestProduct_CycleEqualityObjects1024(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "review/compiler/loom-implied-target/followup/objects-1024.a")
}

func TestProduct_CycleEqualityObjects2048(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "review/compiler/loom-implied-target/followup/objects-2048.a")
}

func TestProduct_CycleEqualityObjects(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "review/compiler/loom-implied-target/objects.a")
}

func TestProduct_CycleEqualityObjectsSequential(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "review/compiler/loom-implied-target/objects-sequential.a")
}

func TestProduct_CycleEqualityMovesObjects(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "internal/oracle/testdata/moves/accepted/objects.a")
}
