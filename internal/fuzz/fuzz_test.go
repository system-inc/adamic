package fuzz

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// One seed makes one program, every time: that's what makes a finding reproduce.
func TestOneSeedOneProgram(test *testing.T) {
	test.Parallel()
	for seed := uint64(1); seed <= 20; seed++ {
		if first, second := Generate(seed).Source(), Generate(seed).Source(); first != second {
			test.Fatalf("seed %d made two different programs", seed)
		}
	}
	if Generate(1).Source() == Generate(2).Source() {
		test.Fatal("seeds 1 and 2 made the same program")
	}
}

// Every generated program must typecheck and lower, except an intentional task refusal,
// which must name the expected path. Move refusals must match the complete diagnostic and fix.
func TestGeneratedProgramsCheckAndLower(test *testing.T) {
	test.Parallel()
	// The literal ranges are also the gate's independently selectable work units.
	for _, seeds := range []struct {
		name        string
		first, last uint64
	}{
		{"seeds-001-005", 1, 5},
		{"seeds-006-010", 6, 10},
		{"seeds-011-015", 11, 15},
		{"seeds-016-020", 16, 20},
		{"seeds-021-025", 21, 25},
		{"seeds-026-030", 26, 30},
		{"seeds-031-035", 31, 35},
		{"seeds-036-040", 36, 40},
		{"seeds-041-045", 41, 45},
		{"seeds-046-050", 46, 50},
		{"seeds-051-055", 51, 55},
		{"seeds-056-060", 56, 60},
	} {
		test.Run(seeds.name, func(test *testing.T) {
			test.Parallel()
			directory := test.TempDir()
			for seed := seeds.first; seed <= seeds.last; seed++ {
				path := filepath.Join(directory, "program.a")
				generated := Generate(seed)
				source := generated.Source()
				if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
					test.Fatal(writeError)
				}
				program, loadError := load.Load([]string{path})
				if loadError != nil {
					test.Errorf("seed %d: the checker refused it: %v", seed, loadError)
					continue
				}
				if generated.Refusal != "" {
					prefix := parallelRefusePrefix + " "
					if generated.RefusalFix != "" {
						prefix = movesRefusePrefix + " "
					}
					if !strings.Contains(source, prefix+generated.Refusal) {
						test.Errorf("seed %d: refusal %q is not in the source", seed, generated.Refusal)
					}
					_, lowerError := lower.Lower(context.Background(), program)
					if lowerError == nil {
						test.Errorf("seed %d: compiler accepted a program it must refuse (%s)", seed, generated.Refusal)
						continue
					}
					var move *lower.Refused
					if generated.RefusalFix != "" &&
						(!errors.As(lowerError, &move) ||
							move.What != generated.Refusal ||
							move.Fix != generated.RefusalFix) {
						test.Errorf("seed %d: wrong move refusal: %v", seed, lowerError)
					}
					if !strings.Contains(lowerError.Error(), generated.Refusal) {
						test.Errorf(
							"seed %d: refusal did not name the path %q: %v",
							seed,
							generated.Refusal,
							lowerError,
						)
					}
					continue
				}
				if _, lowerError := lower.Lower(context.Background(), program); lowerError != nil {
					test.Errorf("seed %d: stage 0 didn't lower it: %v", seed, lowerError)
				}
			}
		})
	}
}

// The parallel feature is on unless asked otherwise, and a run of seeds covers every shape and
// every refusal the runner knows how to check.
func TestParallelFeatureCoverage(test *testing.T) {
	test.Parallel()
	shapes := map[string]bool{}
	refusals := map[string]bool{}
	accepted := 0
	for seed := uint64(1); seed <= 240; seed++ {
		program := GenerateWithout(seed, []string{"moves"})
		source := program.Source()
		if program.Refusal != "" {
			switch {
			case strings.Contains(program.Refusal, "not an immutable binding"):
				refusals["let"] = true
			case strings.Contains(program.Refusal, "mutable"):
				refusals["mutable"] = true
			case strings.Contains(program.Refusal, "writes the global"):
				refusals["global"] = true
			default:
				test.Errorf("seed %d: unexpected refusal %q", seed, program.Refusal)
			}
			if !strings.Contains(source, "parallelMap(") {
				test.Errorf("seed %d: a refusal has no parallelMap", seed)
			}
			continue
		}
		if !strings.Contains(source, "import { parallelMap } from 'adamic';") ||
			!strings.Contains(source, "parallelMap(") {
			test.Errorf("seed %d: parallel is on by default but the program has no parallelMap", seed)
			continue
		}
		accepted++
		for _, shape := range []string{"numbers", "strings", "records", "nested", "fresh-map", "nested-call"} {
			if strings.Contains(source, "// parallel-shape: "+shape) {
				shapes[shape] = true
			}
		}
		if strings.Contains(source, "// parallel-shape: strings") {
			if !strings.Contains(source, "世界") || !strings.Contains(source, ".slice(") {
				test.Errorf("seed %d: string parallelMap is not long, non-ASCII, and sliced", seed)
			}
		}
		if strings.Contains(source, "// parallel-shape: fresh-map") &&
			!strings.Contains(source, "new Map<string, number>()") {
			test.Errorf("seed %d: fresh-map shape does not make a Map inside the work", seed)
		}
		if strings.Contains(source, "// parallel-shape: nested-call") && strings.Count(source, "parallelMap(") < 2 {
			test.Errorf("seed %d: nested-call shape has one parallelMap", seed)
		}
		if strings.Contains(source, "parallelMap(") &&
			!strings.Contains(source, ".join(") &&
			!strings.Contains(source, ".slice(0, 5)") {
			test.Errorf("seed %d: parallel results are not read afterwards", seed)
		}
	}
	if accepted < 150 {
		test.Errorf("only %d of 240 seeds run parallelMap; refusals should be a share, not the majority", accepted)
	}
	for _, shape := range []string{"numbers", "strings", "records", "nested", "fresh-map", "nested-call"} {
		if !shapes[shape] {
			test.Errorf("no accepted program used shape %s", shape)
		}
	}
	for _, kind := range []string{"let", "mutable", "global"} {
		if !refusals[kind] {
			test.Errorf("no program refused a %s", kind)
		}
	}
	for seed := uint64(1); seed <= 20; seed++ {
		source := GenerateWithout(seed, []string{"parallel"}).Source()
		if strings.Contains(source, "parallelMap") {
			test.Errorf("seed %d with -without parallel still mentions parallelMap", seed)
		}
	}
}

// Check the generated regex programs independently of lowering too.
func TestRegexProgramsPassTheChecker(test *testing.T) {
	test.Parallel()
	// The literal ranges are also the gate's independently selectable work units.
	for _, seeds := range []struct {
		name        string
		first, last uint64
	}{
		{"seeds-001-005", 1, 5},
		{"seeds-006-010", 6, 10},
		{"seeds-011-015", 11, 15},
		{"seeds-016-020", 16, 20},
		{"seeds-021-025", 21, 25},
		{"seeds-026-030", 26, 30},
	} {
		test.Run(seeds.name, func(test *testing.T) {
			test.Parallel()
			directory := test.TempDir()
			for seed := seeds.first; seed <= seeds.last; seed++ {
				path := filepath.Join(directory, "program.a")
				source := Generate(seed).Source()
				if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
					test.Fatal(writeError)
				}
				if _, loadError := load.Load([]string{path}); loadError != nil {
					test.Errorf("seed %d: the checker refused it: %v\n%s", seed, loadError, source)
				}
			}
		})
	}
}

// The October vocabulary shows up: a hierarchy, bitwise shifts, and NaN keys. One seed is not
// required to use all of them.
func TestOctoberFeaturesAppear(test *testing.T) {
	test.Parallel()
	seen := map[string]bool{}
	needles := []string{
		"extends ",
		"super.bump",
		">>>",
		"numbers.get(NaN",
		"new Set<number>",
		"flags.has(-0)",
		"/a/g.exec",
	}
	for seed := uint64(1); seed <= 200; seed++ {
		source := Generate(seed).Source()
		for _, needle := range needles {
			if strings.Contains(source, needle) {
				seen[needle] = true
			}
		}
	}
	for _, needle := range needles {
		if !seen[needle] {
			test.Errorf("200 seeds never wrote %s", needle)
		}
	}
}

// The shrinker keeps only what the failure needs. Here a program "fails" while it still deletes
// from the table, so all that should be left is one delete.
func TestShrinkKeepsOnlyWhatFails(test *testing.T) {
	test.Parallel()
	var program *Program
	for seed := uint64(1); program == nil; seed++ {
		if candidate := Generate(seed); strings.Contains(candidate.Source(), "table.delete(") {
			program = candidate
		}
	}
	tries := 0
	shrunk := Shrink(program, "deletes", func(candidate *Program) Outcome {
		tries++
		if strings.Contains(candidate.Source(), "table.delete(") {
			return Outcome{Verdict: Finding, Key: "deletes"}
		}
		return Outcome{Verdict: Agreed}
	})
	source := shrunk.Source()
	if !strings.Contains(source, "table.delete(") {
		test.Fatalf("the shrunk program no longer fails:\n%s", source)
	}
	if lines := strings.Count(source, "\n"); lines > 3 {
		test.Errorf("shrunk to %d lines after %d tries, want at most 3:\n%s", lines, tries, source)
	}
	if program.Source() == source {
		test.Error("shrinking changed the original program instead of a copy")
	}
}

// Every shared-slice scene the probe can write really does share, by the runtime's rule, and the
// generator puts one in a program: a long owner, a slice, and an append to a local.
func TestSharedSliceCutsShare(test *testing.T) {
	test.Parallel()
	cuts := allShareCuts()
	if len(cuts) < 12 {
		test.Fatalf(
			"got %d shared-slice scenes, want at least a repeat, a literal, a template, a join and a nested slice",
			len(cuts),
		)
	}
	sawNested, sawConstant, sawEnd := false, false, false
	for index, cut := range cuts {
		cut.check()
		if cut.pieceEnd == cut.ownerUnits {
			sawEnd = true
		}
		if cut.nested {
			sawNested = true
		}
		if cut.constant {
			sawConstant = true
		}
		if cut.pieceEnd-cut.pieceStart < 48 {
			test.Errorf(
				"cut %d's slice is only %d units, too short to reach a position-index checkpoint",
				index,
				cut.pieceEnd-cut.pieceStart,
			)
		}
	}
	if !sawNested || !sawConstant || !sawEnd {
		test.Errorf("scenes missing a kind: nested %v constant %v to-the-end %v", sawNested, sawConstant, sawEnd)
	}
	var sawProbe, sawLong, sawSlice bool
	for seed := uint64(1); seed <= 24; seed++ {
		source := Generate(seed).Source()
		if strings.Contains(source, "function share") &&
			strings.Contains(source, " += ") &&
			strings.Contains(source, ".slice(") &&
			strings.Contains(source, ".indexOf(") &&
			strings.Contains(source, ".charCodeAt(") {
			sawProbe = true
		}
		if strings.Contains(source, ".repeat(") || strings.Count(source, "é") > 40 || strings.Count(source, "a") > 70 {
			sawLong = true
		}
		if strings.Contains(source, ".slice(16, 112)") ||
			strings.Contains(source, ".slice(8)") ||
			strings.Contains(source, ".slice(16, 176)") {
			sawSlice = true
		}
	}
	if !sawProbe || !sawLong || !sawSlice {
		test.Fatalf("vocabulary missing: probe %v long string %v sharing slice %v", sawProbe, sawLong, sawSlice)
	}
	without := GenerateWithout(1, []string{"shared-slices"}).Source()
	if strings.Contains(without, "function share") {
		test.Fatal("leaving shared slices out still wrote the probe")
	}
}

// The ownership scenes are the shapes the memory passes get wrong, one per seed on a cadence of
// eight, and the family around them is in every program. Leaving the feature out drops all of it.
// Every one of those programs has to lower: a scene the checker refuses tests nothing.
func TestOwnershipShapes(test *testing.T) {
	test.Parallel()
	want := []string{
		"ownShow(ownGlobal, ownResetGlobal()",
		"ownShow(ownSlot.box, ownReplaceSlot()",
		"ownShow(ownLocal, ownDrop()",
		"ownWorker.run(items)",
		"super.run(items)",
		"...ownTree, tag: ownTree.keep()",
		"ownGrow(ownTree.left), tag: ownTree.peek()",
		"new OwnMarked(",
		"ownFamilyMix(ownFamilyText, ownFamilySet())",
		"ownFamilyReset();",
		"ownFamilyMap()",
		"...ownFamilyTree, tag: ownFamilyTree.keep()",
		"ownFamilyKeep(new OwnFamilyHeld(",
	}
	found := map[string]bool{}
	directory := test.TempDir()
	for seed := uint64(1); seed <= 8; seed++ {
		source := GenerateWithout(seed, []string{"moves", "parallel"}).Source()
		for _, marker := range want {
			if strings.Contains(source, marker) {
				found[marker] = true
			}
		}
		if !strings.Contains(source, "interface OwnFamilyBox") {
			test.Errorf("seed %d: ownership left out the family", seed)
		}
		path := filepath.Join(directory, "program.a")
		if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		program, loadError := load.Load([]string{path})
		if loadError != nil {
			test.Errorf("seed %d: the checker refused it: %v\n%s", seed, loadError, source)
			continue
		}
		if _, lowerError := lower.Lower(context.Background(), program); lowerError != nil {
			test.Errorf("seed %d: stage 0 didn't lower it: %v\n%s", seed, lowerError, source)
		}
	}
	for _, marker := range want {
		if !found[marker] {
			test.Errorf("seeds 1 to 8 never generated %s", marker)
		}
	}
	without := GenerateWithout(1, []string{"ownership"}).Source()
	if strings.Contains(without, "OwnFamilyBox") ||
		strings.Contains(without, "OwnMarked") ||
		strings.Contains(without, "ownGlobal") {
		test.Fatal("leaving ownership out still wrote a scene")
	}
}
