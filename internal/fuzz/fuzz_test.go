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
	directory := t.TempDir()
	for seed := uint64(1); seed <= 60; seed++ {
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
}

// Check the generated regex programs independently of lowering too.
func TestRegexProgramsPassTheChecker(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for seed := uint64(1); seed <= 30; seed++ {
		path := filepath.Join(directory, "program.a")
		source := Generate(seed).Source()
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := load.Load([]string{path}); err != nil {
			t.Errorf("seed %d: the checker refused it: %v\n%s", seed, err, source)
		}
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

// Regions are on unless named in -without. Every program carries an honest statement region and the
// near-misses that have to stay off it, and leaving the feature out drops that whole section.
func TestRegionsFeature(t *testing.T) {
	t.Parallel()
	for seed := uint64(1); seed <= 12; seed++ {
		source := Generate(seed).Source()
		for _, piece := range []string{
			"import type { Weak } from 'adamic';",
			"throw new Error(",
			"() => this.label",
			"parent: Weak<",
			"region-counted ",
			"region-kept ",
			"region-field-kept ",
			"region-index-kept ",
			"region-closure-kept ",
			"region-map-kept ",
			"region-returned ",
			"region-paired ",
			"region-both-kept ",
			"region-caught ",
			"region-weak-kept ",
			"weak-parent ",
			"region-box-later ",
			"region-labels ",
		} {
			if !strings.Contains(source, piece) {
				t.Errorf("seed %d: regions program missing %q", seed, piece)
			}
		}
		off := GenerateWithout(seed, []string{"regions"}).Source()
		if strings.Contains(off, "import type { Weak } from 'adamic'") || strings.Contains(off, "region-counted ") {
			t.Errorf("seed %d: regions stayed on when the feature was left out", seed)
		}
		if strings.Contains(off, "() => this.label") {
			t.Errorf("seed %d: a constructor capture remained with regions left out", seed)
		}
	}
}
