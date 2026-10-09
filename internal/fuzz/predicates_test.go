package fuzz

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func predicateGenerator(seed uint64) *generator {
	return &generator{seed: seed, random: rand.New(rand.NewPCG(seed, 0x61646d6963))}
}

func TestPredicateGeneratorShapes(t *testing.T) {
	for seed := uint64(1); seed <= 36; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			g := predicateGenerator(seed)
			for _, source := range []string{g.helperPredicateSource(), g.arrayPredicateSource()} {
				path := filepath.Join(t.TempDir(), "program.a")
				if err := os.WriteFile(path, []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
				program, err := load.Load([]string{path})
				if err != nil {
					t.Fatalf("invalid generated source: %v\n%s", err, source)
				}
				_, err = lower.Lower(context.Background(), program)
				var refusal *lower.Refused
				if err != nil && !errors.As(err, &refusal) {
					t.Fatalf("unclear compiler failure: %v\n%s", err, source)
				}
			}
		})
	}
	for _, feature := range []string{"helper-predicates", "array-predicates"} {
		marker := "predHelper"
		if feature == "array-predicates" {
			marker = "predItems"
		}
		if !strings.Contains(GenerateFeatures(1, nil, []string{feature}).Source(), marker) {
			t.Fatalf("Generate missing %s", feature)
		}
		if strings.Contains(GenerateFeatures(1, []string{feature}, []string{feature}).Source(), marker) {
			t.Fatalf("GenerateWithout retained %s", feature)
		}
	}
}
