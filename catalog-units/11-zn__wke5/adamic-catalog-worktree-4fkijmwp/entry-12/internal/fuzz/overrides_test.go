package fuzz

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Cover every effect with every argument, with real lowering as well as vocabulary assertions.
func TestOverridesShapesAndLower(t *testing.T) {
	t.Parallel()
	without := []string{}
	for _, feature := range Features {
		if feature != "overrides" {
			without = append(without, feature)
		}
	}
	programs := map[uint64]string{}
	for seed := uint64(0); seed < 50; seed++ {
		programs[seed] = GenerateWithout(seed, without).Source()
	}
	vocabulary := func(t *testing.T) {
		markers := []string{
			"items[0] =", "box.v =", "this.saved = box", "overrideGlobal = box",
			"overrideArray.push(box)", "overrideMap.set('kept', box)", "throw new Error(box.v)",
			"overrideClosures.push(() => box.v)", "callback();", "return box;", "items.map(",
			"worker.run(borrowed,", "worker.run(items[0] ??", "worker.run(overrideNarrowed,",
			"worker.run({ v:", "worker.run({ ...borrowed,", "super.run(borrowed,",
			"worker: OverrideBase", "worker: OverrideWorker", "extends OverrideBase",
			"class OverrideReader implements OverrideWorker", "class OverrideBase implements OverrideWorker",
			"console.log(overrideChild.saved.v)", "console.log(overrideGlobal.v)",
			"for (const kept of overrideArray)", "for (const [key, kept] of overrideMap)",
			"for (const later of overrideClosures)", "console.log(result.v)", "console.log(borrowed.v)",
		}
		seen := map[string]bool{}
		for seed := uint64(0); seed < 50; seed++ {
			source := programs[seed]
			for _, marker := range markers {
				seen[marker] = seen[marker] || strings.Contains(source, marker)
			}
		}
		for _, marker := range markers {
			if !seen[marker] {
				t.Errorf("50 seeds never generated %s", marker)
			}
		}
		if strings.Contains(GenerateWithout(1, []string{"overrides"}).Source(), "OverrideBox") {
			t.Fatal("disabled overrides still generated a scene")
		}
	}

	// The literal ranges are also the gate's independently selectable work units.
	for _, seeds := range []struct {
		name        string
		first, last uint64
	}{
		{"vocabulary", 0, 0},
		{"seeds-000-004", 0, 4},
		{"seeds-005-009", 5, 9},
		{"seeds-010-014", 10, 14},
		{"seeds-015-019", 15, 19},
		{"seeds-020-024", 20, 24},
		{"seeds-025-029", 25, 29},
		{"seeds-030-034", 30, 34},
		{"seeds-035-039", 35, 39},
		{"seeds-040-044", 40, 44},
		{"seeds-045-049", 45, 49},
	} {
		t.Run(seeds.name, func(t *testing.T) {
			t.Parallel()
			if seeds.name == "vocabulary" {
				vocabulary(t)
				return
			}
			directory := t.TempDir()
			for seed := seeds.first; seed <= seeds.last; seed++ {
				source := programs[seed]
				path := filepath.Join(directory, "program.a")
				if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
				program, err := load.Load([]string{path})
				if err != nil {
					t.Fatalf("seed %d: checker: %v\n%s", seed, err, source)
				}
				if _, err := lower.Lower(context.Background(), program); err != nil {
					t.Fatalf("seed %d: lower: %v\n%s", seed, err, source)
				}
			}
		})
	}
}
