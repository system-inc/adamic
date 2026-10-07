package wave104

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func fractionCases(t *testing.T, oracle string) string {
	t.Helper()
	input, err := filepath.Abs("trim_fixture_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "fraction-cases.json")
	run(t, "", oracle, "--generate-fraction", input, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Digits []int
		Cases  []struct{ Rule, Value string }
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	consumers := map[string]int{}
	for _, c := range corpus.Cases {
		if c.Rule != "" {
			consumers[c.Rule]++
		}
	}
	if len(consumers) != 4 || len(corpus.Digits) != 10 || len(corpus.Cases) != 233079 {
		t.Fatalf("coverage consumers=%v digits=%d cases=%d", consumers, len(corpus.Digits), len(corpus.Cases))
	}
	t.Logf("%d actual Go comparisons across four consumer fixture domains, numeric ratios and all BMP scalar slash-adjacent points", len(corpus.Cases))
	return path
}

// Not parallel: bound native sanitizer memory for the complete population.
func TestFractionMatchesGo(t *testing.T) {
	oracle := trimOracle(t)
	path := fractionCases(t, oracle)
	want := run(t, "", oracle, "--fraction", path)
	entry, js, binary := buildHelper(t, "is_fraction.a", "fraction_main.a", "", "")
	for side, got := range observations(t, entry, js, binary, path) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s", side, firstDifference(got, want))
		}
	}
	t.Logf("%d exact Go bytes include result and ordered math, scanner and trim calls on three targets", len(want))
}

// Not parallel: a mutant counts only when it compiles and completes cleanly.
func TestFractionMutants(t *testing.T) {
	oracle := trimOracle(t)
	path := fractionCases(t, oracle)
	want := run(t, "", oracle, "--fraction", path)
	for _, change := range []struct{ name, from, to string }{
		{"ignore math functions", "if(hasMathFunction(value))", "if(false && hasMathFunction(value))"},
		{"accept an empty denominator", "denominator > 0", "denominator >= 0"},
		{"omit numerator whitespace trim", "rest = trimLeadingJavaScriptSpace(rest);", "rest = rest;"},
		{"scan before the math short circuit", "if(hasMathFunction(value))", "scanNumber(value); if(hasMathFunction(value))"},
	} {
		t.Run(change.name, func(t *testing.T) {
			entry, js, binary := buildHelper(t, "is_fraction.a", "fraction_main.a", change.from, change.to)
			for side, got := range observations(t, entry, js, binary, path) {
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side)
				}
				t.Logf("compiled and completed cleanly; only Go comparison caught %s on %s: %s", change.name, side, firstDifference(got, want))
			}
		})
	}
}
