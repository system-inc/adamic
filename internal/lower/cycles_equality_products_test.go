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
// Parser/lint ports and large Map fixtures also have products: their reference
// comparisons dominate loaded shard time even when each has its own shard.
var cycleEqualityProducts = map[string]string{
	"stage1/cohere/parse/parse.a":                                    "ParseDriver",
	"stage1/cohere/lint/rules/no-unsafe-negation/profile.a":          "LintNoUnsafeNegation",
	"stage1/cohere/parse/main.a":                                     "ParseDriverMain",
	"stage1/cohere/lint/rules/no-unsafe-optional-chaining/profile.a": "LintNoUnsafeOptionalChaining",
	"stage1/cohere/lint/rules/no-underscore-dangle/profile.a":        "LintNoUnderscoreDangle",
	"stage1/cohere/lint/rules/typescript-no-this-alias/profile.a":    "LintTypescriptNoThisAlias",
	"internal/oracle/testdata/program_region/parse.a":                "ProgramRegionParse",
	"stage3/drivers/parser/witness.a":                                "ParserWitness",
	"internal/oracle/testdata/library_map_set.a":                     "LibraryMapSet",
	"internal/oracle/testdata/library_map_set_setops.a":              "LibraryMapSetSetops",
	"review/compiler/loom-implied-target/followup/objects-512.a":     "Objects512",
	"review/compiler/loom-implied-target/followup/objects-1024.a":    "Objects1024",
	"review/compiler/loom-implied-target/followup/objects-2048.a":    "Objects2048",
	"review/compiler/loom-implied-target/objects.a":                  "Objects",
	"review/compiler/loom-implied-target/objects-sequential.a":       "ObjectsSequential",
	"internal/oracle/testdata/moves/accepted/objects.a":              "MovesObjects",
}

// Own the entry's sibling modules as well as the shared source roots. Keep this
// separate from compilerinputs: fixture imports must remain keyed even if the
// compiler's input list becomes narrower.
func cycleEqualityProgramFiles(path string) []string {
	return []string{"review", "stage1", "stage3/fixtures/cycles", path, filepath.Dir(path)}
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
		Files:     append(compilerinputs.Lowering(), cycleEqualityProgramFiles(path)...),
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

func TestProduct_CycleEqualityParseDriver(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "stage1/cohere/parse/parse.a")
}

func TestProduct_CycleEqualityLintNoUnsafeNegation(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "stage1/cohere/lint/rules/no-unsafe-negation/profile.a")
}

func TestProduct_CycleEqualityParseDriverMain(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "stage1/cohere/parse/main.a")
}

func TestProduct_CycleEqualityLintNoUnsafeOptionalChaining(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "stage1/cohere/lint/rules/no-unsafe-optional-chaining/profile.a")
}

func TestProduct_CycleEqualityLintNoUnderscoreDangle(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "stage1/cohere/lint/rules/no-underscore-dangle/profile.a")
}

func TestProduct_CycleEqualityLintTypescriptNoThisAlias(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "stage1/cohere/lint/rules/typescript-no-this-alias/profile.a")
}

func TestProduct_CycleEqualityProgramRegionParse(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "internal/oracle/testdata/program_region/parse.a")
}

func TestProduct_CycleEqualityParserWitness(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "stage3/drivers/parser/witness.a")
}

func TestProduct_CycleEqualityLibraryMapSet(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "internal/oracle/testdata/library_map_set.a")
}

func TestProduct_CycleEqualityLibraryMapSetSetops(t *testing.T) {
	t.Parallel()
	cycleEqualityProduct(t, "internal/oracle/testdata/library_map_set_setops.a")
}
