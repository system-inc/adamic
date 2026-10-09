package fuzz

import (
	"context"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// interfaceOmitted is an optional argument left out of a call through an interface.
var interfaceOmitted = regexp.MustCompile(`maybeCarrier\([0-9]\)\.pick\(\)|maybeSole\.pick\(\)`)

// Within forty seeds the scene carries every kind of value through every shape, and every program it
// makes checks and lowers. The interface's left-out optional argument stays out until it's asked for.
func TestUndefinedNumbersShapes(t *testing.T) {
	t.Parallel()
	programs := map[uint64]string{}
	for seed := uint64(1); seed <= 40; seed++ {
		// Moves and parallel programs can be refused by design; every program here must lower.
		programs[seed] = GenerateWithout(seed, []string{"moves", "parallel"}).Source()
	}

	// The literal ranges are also the gate's independently selectable work units.
	for _, seeds := range []struct {
		name        string
		first, last uint64
	}{
		{"vocabulary", 0, 0},
		{"seeds-001-005", 1, 5},
		{"seeds-006-010", 6, 10},
		{"seeds-011-015", 11, 15},
		{"seeds-016-020", 16, 20},
		{"seeds-021-025", 21, 25},
		{"seeds-026-030", 26, 30},
		{"seeds-031-035", 31, 35},
		{"seeds-036-040", 36, 40},
	} {
		t.Run(seeds.name, func(t *testing.T) {
			t.Parallel()
			if seeds.name == "vocabulary" {
				checkUndefinedNumbersVocabulary(t, programs)
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

// Keep the aggregate vocabulary check separate from the gate's literal seed table.
func checkUndefinedNumbersVocabulary(t *testing.T, programs map[uint64]string) {
	shapes := map[string]bool{}
	sources := map[string]bool{}
	for seed := uint64(1); seed <= 40; seed++ {
		source := programs[seed]
		for _, line := range strings.Split(source, "\n") {
			if !strings.HasPrefix(line, "console.log(`maybeValue") {
				continue
			}
			chain := strings.Fields(line)[1]
			steps := strings.Split(chain, ">")
			sources[steps[0]] = true
			for _, step := range steps[1:] {
				shapes[step] = true
			}
			if interfaceOmitted.MatchString(line) {
				t.Errorf("seed %d left an argument out through an interface without the opt-in: %s", seed, line)
			}
		}
	}
	for _, shape := range (&generator{random: rand.New(rand.NewPCG(1, 2))}).maybeShapes() {
		if shape.optIn != "" {
			if shapes[shape.name] {
				t.Errorf("%s ran without its opt-in, %s", shape.name, shape.optIn)
			}
			continue
		}
		if !shapes[shape.name] {
			t.Errorf("40 seeds never carried a value through %s", shape.name)
		}
	}
	for _, source := range []string{"omitted", "undefined", "NaN", "madeNaN", "hole", "held", "number"} {
		if !sources[source] {
			t.Errorf("40 seeds never started from %s", source)
		}
	}

	var optedIn bool
	for seed := uint64(1); seed <= 40 && !optedIn; seed++ {
		source := GenerateFeatures(seed, nil, []string{interfaceOmittedOptional}).Source()
		optedIn = interfaceOmitted.MatchString(source)
	}
	if !optedIn {
		t.Error("with the opt-in, 40 seeds never left an argument out through an interface")
	}
	if strings.Contains(GenerateWithout(1, []string{"undefined-numbers"}).Source(), "MaybeSlot") {
		t.Error("leaving undefined-numbers out still wrote the scene")
	}
}
