package wave104

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func vectorCases(t *testing.T, oracle string) string {
	t.Helper()
	input, err := filepath.Abs("trim_fixture_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "vector-cases.json")
	run(t, "", oracle, "--generate-vector", input, path)
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
	if len(consumers) != 4 || len(corpus.Digits) != 10 || len(corpus.Cases) != 170369 {
		t.Fatalf("coverage consumers=%v digits=%d cases=%d", consumers, len(corpus.Digits), len(corpus.Cases))
	}
	t.Logf("%d actual Go comparisons across four consumer fixture domains, numeric triples and all BMP scalar separators", len(corpus.Cases))
	return path
}

// Not parallel: bound native sanitizer memory for the complete population.
func TestVectorMatchesGo(t *testing.T) {
	oracle := trimOracle(t)
	path := vectorCases(t, oracle)
	want := run(t, "", oracle, "--vector", path)
	entry, js, binary := buildHelper(t, "is_vector.a", "vector_main.a", "", "")
	for side, got := range observations(t, entry, js, binary, path) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s", side, firstDifference(got, want))
		}
	}
	t.Logf("%d exact Go bytes include result and lazy scanner call count and arguments on three targets", len(want))
}

// Not parallel: a mutant counts only when it compiles and completes cleanly.
func TestVectorMutants(t *testing.T) {
	oracle := trimOracle(t)
	path := vectorCases(t, oracle)
	want := run(t, "", oracle, "--vector", path)
	for _, change := range []struct{ name, from, to string }{
		{"accept two components", "component === 2", "component === 1"},
		{"accept a tab separator", "rest.charCodeAt(spaces) === 32", "(rest.charCodeAt(spaces) === 32 || rest.charCodeAt(spaces) === 9)"},
		{"ignore the trailing remainder", "return rest === '';", "return true;"},
		{"scan twice", "const consumed = scanNumber(rest);", "scanNumber(rest);\n        const consumed = scanNumber(rest);"},
	} {
		t.Run(change.name, func(t *testing.T) {
			entry, js, binary := buildHelper(t, "is_vector.a", "vector_main.a", change.from, change.to)
			for side, got := range observations(t, entry, js, binary, path) {
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side)
				}
				t.Logf("compiled and completed cleanly; only Go comparison caught %s on %s: %s", change.name, side, firstDifference(got, want))
			}
		})
	}
}
