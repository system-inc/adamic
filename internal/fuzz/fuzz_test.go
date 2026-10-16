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

// One seed makes one program, every time: that's what makes a finding reproduce.
func TestOneSeedOneProgram(t *testing.T) {
	t.Parallel()
	for seed := uint64(1); seed <= 20; seed++ {
		if first, second := Generate(seed).Source(), Generate(seed).Source(); first != second {
			t.Fatalf("seed %d made two different programs", seed)
		}
	}
	if Generate(1).Source() == Generate(2).Source() {
		t.Fatal("seeds 1 and 2 made the same program")
	}
}

// Every program the generator makes must be Adamic that stage 0 lowers: one the checker refuses is
// the generator's fault, and one stage 0 can't lower tests nothing. Native regex is available on
// the integration head, so this includes every generated feature.
func TestGeneratedProgramsCheckAndLower(t *testing.T) {
	t.Parallel()
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
		t.Run(seeds.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for seed := seeds.first; seed <= seeds.last; seed++ {
				path := filepath.Join(directory, "program.a")
				source := Generate(seed).Source()
				if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
				program, err := load.Load([]string{path})
				if err != nil {
					t.Errorf("seed %d: the checker refused it: %v", seed, err)
					continue
				}
				if _, err := lower.Lower(context.Background(), program); err != nil {
					t.Errorf("seed %d: stage 0 didn't lower it: %v", seed, err)
				}
			}
		})
	}
}

// Check the generated regex programs independently of lowering too.
func TestRegexProgramsPassTheChecker(t *testing.T) {
	t.Parallel()
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
		t.Run(seeds.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for seed := seeds.first; seed <= seeds.last; seed++ {
				path := filepath.Join(directory, "program.a")
				source := Generate(seed).Source()
				if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
				if _, err := load.Load([]string{path}); err != nil {
					t.Errorf("seed %d: the checker refused it: %v\n%s", seed, err, source)
				}
			}
		})
	}
}

// The October vocabulary shows up: a hierarchy, bitwise shifts, and NaN keys. One seed is not
// required to use all of them.
func TestOctoberFeaturesAppear(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	needles := []string{"extends ", "super.bump", ">>>", "numbers.get(NaN", "new Set<number>", "flags.has(-0)", "/a/g.exec"}
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
			t.Errorf("200 seeds never wrote %s", needle)
		}
	}
}

// The shrinker keeps only what the failure needs. Here a program "fails" while it still deletes
// from the table, so all that should be left is one delete.
func TestShrinkKeepsOnlyWhatFails(t *testing.T) {
	t.Parallel()
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
		t.Fatalf("the shrunk program no longer fails:\n%s", source)
	}
	if lines := strings.Count(source, "\n"); lines > 3 {
		t.Errorf("shrunk to %d lines after %d tries, want at most 3:\n%s", lines, tries, source)
	}
	if program.Source() == source {
		t.Error("shrinking changed the original program instead of a copy")
	}
}

// Every shared-slice scene the probe can write really does share, by the runtime's rule, and the
// generator puts one in a program: a long owner, a slice, and an append to a local.
func TestSharedSliceCutsShare(t *testing.T) {
	t.Parallel()
	cuts := allShareCuts()
	if len(cuts) < 12 {
		t.Fatalf("got %d shared-slice scenes, want at least a repeat, a literal, a template, a join and a nested slice", len(cuts))
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
			t.Errorf("cut %d's slice is only %d units, too short to reach a position-index checkpoint", index, cut.pieceEnd-cut.pieceStart)
		}
	}
	if !sawNested || !sawConstant || !sawEnd {
		t.Errorf("scenes missing a kind: nested %v constant %v to-the-end %v", sawNested, sawConstant, sawEnd)
	}
	var sawProbe, sawLong, sawSlice bool
	for seed := uint64(1); seed <= 24; seed++ {
		source := Generate(seed).Source()
		if strings.Contains(source, "function share") && strings.Contains(source, " += ") && strings.Contains(source, ".slice(") && strings.Contains(source, ".indexOf(") && strings.Contains(source, ".charCodeAt(") {
			sawProbe = true
		}
		if strings.Contains(source, ".repeat(") || strings.Count(source, "é") > 40 || strings.Count(source, "a") > 70 {
			sawLong = true
		}
		if strings.Contains(source, ".slice(16, 112)") || strings.Contains(source, ".slice(8)") || strings.Contains(source, ".slice(16, 176)") {
			sawSlice = true
		}
	}
	if !sawProbe || !sawLong || !sawSlice {
		t.Fatalf("vocabulary missing: probe %v long string %v sharing slice %v", sawProbe, sawLong, sawSlice)
	}
	without := GenerateWithout(1, []string{"shared-slices"}).Source()
	if strings.Contains(without, "function share") {
		t.Fatal("leaving shared slices out still wrote the probe")
	}
}

// The ownership scenes are the shapes the memory passes get wrong, one per seed on a cadence of
// eight, and the family around them is in every program. Leaving the feature out drops all of it.
// Every one of those programs has to lower: a scene the checker refuses tests nothing.
func TestOwnershipShapes(t *testing.T) {
	t.Parallel()
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
	directory := t.TempDir()
	for seed := uint64(1); seed <= 8; seed++ {
		source := Generate(seed).Source()
		for _, marker := range want {
			if strings.Contains(source, marker) {
				found[marker] = true
			}
		}
		if !strings.Contains(source, "interface OwnFamilyBox") {
			t.Errorf("seed %d: ownership left out the family", seed)
		}
		path := filepath.Join(directory, "program.a")
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Errorf("seed %d: the checker refused it: %v\n%s", seed, err, source)
			continue
		}
		if _, err := lower.Lower(context.Background(), program); err != nil {
			t.Errorf("seed %d: stage 0 didn't lower it: %v\n%s", seed, err, source)
		}
	}
	for _, marker := range want {
		if !found[marker] {
			t.Errorf("seeds 1 to 8 never generated %s", marker)
		}
	}
	without := GenerateWithout(1, []string{"ownership"}).Source()
	if strings.Contains(without, "OwnFamilyBox") || strings.Contains(without, "OwnMarked") || strings.Contains(without, "ownGlobal") {
		t.Fatal("leaving ownership out still wrote a scene")
	}
}
