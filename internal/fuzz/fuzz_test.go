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

// Every program the generator makes must be Adamic that stage 0 lowers, except a parallel refusal,
// which must be refused with a message that names the path. One the checker refuses for any other
// reason is the generator's fault, and one stage 0 can't lower tests nothing.
func TestGeneratedProgramsCheckAndLower(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for seed := uint64(1); seed <= 60; seed++ {
		path := filepath.Join(directory, "program.a")
		program := Generate(seed)
		source := program.Source()
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		loaded, err := load.Load([]string{path})
		if program.Refusal != "" {
			prefix := parallelRefusePrefix + " "
			if program.RefusalFix != "" {
				prefix = movesRefusePrefix + " "
			}
			if !strings.Contains(source, prefix+program.Refusal) {
				t.Errorf("seed %d: refusal %q is not in the source", seed, program.Refusal)
			}
			if err != nil {
				t.Errorf("seed %d: a parallel refusal must typecheck, the proof is stage 0's: %v", seed, err)
				continue
			}
			_, err = lower.Lower(context.Background(), loaded)
			if err == nil {
				t.Errorf("seed %d: compiler accepted a program it must refuse (%s)", seed, program.Refusal)
				continue
			}
			var move *lower.Refused
			if program.RefusalFix != "" && (!errors.As(err, &move) || move.What != program.Refusal || move.Fix != program.RefusalFix) {
				t.Errorf("seed %d: wrong move refusal: %v", seed, err)
			}
			if !strings.Contains(err.Error(), program.Refusal) {
				t.Errorf("seed %d: refusal did not name the path %q:\n%v", seed, program.Refusal, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("seed %d: the checker refused it: %v", seed, err)
			continue
		}
		if _, err := lower.Lower(context.Background(), loaded); err != nil {
			t.Errorf("seed %d: stage 0 didn't lower it: %v\n%s", seed, err, source)
		}
	}
}

// The parallel feature is on unless asked otherwise, and a run of seeds covers every shape and
// every refusal the runner knows how to check.
func TestParallelFeatureCoverage(t *testing.T) {
	t.Parallel()
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
				t.Errorf("seed %d: unexpected refusal %q", seed, program.Refusal)
			}
			if !strings.Contains(source, "parallelMap(") {
				t.Errorf("seed %d: a refusal has no parallelMap", seed)
			}
			continue
		}
		if !strings.Contains(source, "import { parallelMap } from 'adamic';") || !strings.Contains(source, "parallelMap(") {
			t.Errorf("seed %d: parallel is on by default but the program has no parallelMap", seed)
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
				t.Errorf("seed %d: string parallelMap is not long, non-ASCII, and sliced", seed)
			}
		}
		if strings.Contains(source, "// parallel-shape: fresh-map") && !strings.Contains(source, "new Map<string, number>()") {
			t.Errorf("seed %d: fresh-map shape does not make a Map inside the work", seed)
		}
		if strings.Contains(source, "// parallel-shape: nested-call") && strings.Count(source, "parallelMap(") < 2 {
			t.Errorf("seed %d: nested-call shape has one parallelMap", seed)
		}
		if strings.Contains(source, "parallelMap(") && !strings.Contains(source, ".join(") && !strings.Contains(source, ".slice(0, 5)") {
			t.Errorf("seed %d: parallel results are not read afterwards", seed)
		}
	}
	if accepted < 150 {
		t.Errorf("only %d of 240 seeds run parallelMap; refusals should be a share, not the majority", accepted)
	}
	for _, shape := range []string{"numbers", "strings", "records", "nested", "fresh-map", "nested-call"} {
		if !shapes[shape] {
			t.Errorf("no accepted program used shape %s", shape)
		}
	}
	for _, kind := range []string{"let", "mutable", "global"} {
		if !refusals[kind] {
			t.Errorf("no program refused a %s", kind)
		}
	}
	for seed := uint64(1); seed <= 20; seed++ {
		source := GenerateWithout(seed, []string{"parallel"}).Source()
		if strings.Contains(source, "parallelMap") {
			t.Errorf("seed %d with -without parallel still mentions parallelMap", seed)
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
