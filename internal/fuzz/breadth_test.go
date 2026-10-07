package fuzz

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Every construct and variant the breadth scene can write appears, and each program it makes checks
// and lowers. The opt-ins are lowered only where stage 0 can: objects/frozen not at all, statics
// beside constructs that make no structural calls or spreads (math and long-sort).
func TestBreadthWritesEveryConstruct(t *testing.T) {
	t.Parallel()
	var without []string
	for _, feature := range Features {
		if feature != breadthFeature {
			without = append(without, feature)
		}
	}
	uniform := map[string]float64{}
	seen := map[string]bool{}
	directory := t.TempDir()
	for seed := uint64(1); seed <= 250; seed++ {
		program, chosen := GenerateWeighted(seed, without, []string{breadthStatics}, uniform)
		for _, name := range chosen {
			seen[name] = true
		}
		if slices.Contains(chosen, "objects/frozen") || slices.Contains(chosen, "static") {
			continue
		}
		checkAndLower(t, directory, seed, chosen, program)
	}
	statics := map[string]float64{"static": 1e9, "math": 1e6, "long-sort": 1e6}
	for seed := uint64(1); seed <= 12; seed++ {
		program, chosen := GenerateWeighted(seed, without, []string{breadthStatics}, statics)
		if slices.Contains(chosen, "static") && slices.Contains(chosen, "math") && slices.Contains(chosen, "long-sort") {
			checkAndLower(t, directory, seed, chosen, program)
		}
	}
	for _, name := range BreadthChoices() {
		if !seen[name] && name != "objects/frozen" {
			t.Errorf("250 seeds never chose %s", name)
		}
	}
}

func checkAndLower(t *testing.T, directory string, seed uint64, chosen []string, program *Program) {
	t.Helper()
	path := filepath.Join(directory, "program.a")
	if err := os.WriteFile(path, []byte(program.Source()), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("seed %d (%v): checker: %v\n%s", seed, chosen, err, program.Source())
	}
	if _, err := lower.Lower(context.Background(), loaded); err != nil {
		t.Fatalf("seed %d (%v): lower: %v\n%s", seed, chosen, err, program.Source())
	}
}

// The weights steer: a construct weighted far above the rest is in every program, one far below is
// in almost none, an opt-in construct is in none without its opt-in, and one seed with one table
// always chooses the same.
func TestBreadthWeightsSteer(t *testing.T) {
	t.Parallel()
	weights := map[string]float64{"math": 1000, "json": 0.0001}
	jsons := 0
	for seed := uint64(1); seed <= 100; seed++ {
		_, chosen := GenerateWeighted(seed, nil, nil, weights)
		if !slices.Contains(chosen, "math") {
			t.Fatalf("seed %d didn't choose math, weighted 1000: %v", seed, chosen)
		}
		if slices.Contains(chosen, "json") {
			jsons++
		}
		if slices.Contains(chosen, "static") {
			t.Fatalf("seed %d chose static without its opt-in", seed)
		}
		if _, again := GenerateWeighted(seed, nil, nil, weights); !slices.Equal(chosen, again) {
			t.Fatalf("seed %d chose %v, then %v", seed, chosen, again)
		}
	}
	if jsons > 10 {
		t.Errorf("json, weighted 0.0001, was chosen in %d of 100 programs", jsons)
	}
	if strings.Contains(GenerateWithout(1, []string{breadthFeature}).Source(), "breadth") {
		t.Error("the breadth scene was written with breadth left out")
	}
}

// The checked-in table parses and names every choice, and a table written out reads back the same.
func TestBreadthWeightTable(t *testing.T) {
	t.Parallel()
	weights := BreadthWeights()
	for _, name := range BreadthChoices() {
		if _, known := weights[name]; !known {
			t.Errorf("breadth_weights.txt has no weight for %s", name)
		}
	}
	again, err := ParseWeights(FormatWeights(weights, "a header\nof two lines"))
	if err != nil {
		t.Fatal(err)
	}
	for name, weight := range weights {
		if again[name] != weight {
			t.Errorf("%s: wrote %v, read %v", name, weight, again[name])
		}
	}
	if _, err := ParseWeights("switch -1\n"); err == nil {
		t.Error("a negative weight was read")
	}
}

// Lines of source become statements with blocks, so the shrinker can take them apart, and print as
// the same program.
func TestBreadthParsesBlocks(t *testing.T) {
	t.Parallel()
	statements := parseStatements([]string{
		"function f(value: number): string {",
		"try {",
		"if (value > 1) {",
		"return 'big';",
		"} else {",
		"return 'small';",
		"}",
		"} finally {",
		"console.log('done');",
		"}",
		"}",
		"console.log(f(2));",
	})
	if len(statements) != 2 {
		t.Fatalf("want 2 statements, got %d", len(statements))
	}
	program := &Program{Block: Block{Statements: statements}}
	want := "function f(value: number): string {\n\ttry {\n\t\tif (value > 1) {\n\t\t\treturn 'big';\n\t\t} else {\n\t\t\treturn 'small';\n\t\t}\n\t} finally {\n\t\tconsole.log('done');\n\t}\n}\nconsole.log(f(2));\n"
	if got := program.Source(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
