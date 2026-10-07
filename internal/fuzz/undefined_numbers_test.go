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
	shapes := map[string]bool{}
	sources := map[string]bool{}
	directory := t.TempDir()
	for seed := uint64(1); seed <= 40; seed++ {
		source := Generate(seed).Source()
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
