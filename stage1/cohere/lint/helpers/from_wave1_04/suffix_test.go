package wave104

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func suffixCases(t *testing.T, oracle string) string {
	t.Helper()
	input, err := filepath.Abs("trim_fixture_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "suffix-cases.json")
	run(t, "", oracle, "--generate-suffix", input, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		SuffixSets [][]string
		Cases      []struct{ Rule, Value string }
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
	if len(consumers) != 4 || len(corpus.SuffixSets) != 12 || len(corpus.Cases) != 101547 {
		t.Fatalf("coverage consumers=%v sets=%d cases=%d", consumers, len(corpus.SuffixSets), len(corpus.Cases))
	}
	t.Logf("%d values x %d suffix sets = %d actual Go comparisons, four consumer fixture domains", len(corpus.Cases), len(corpus.SuffixSets), len(corpus.Cases)*len(corpus.SuffixSets))
	return path
}

// Not parallel: bound native sanitizer and large output storage for all suffix populations.
func TestNumberWithSuffixMatchesGo(t *testing.T) {
	oracle := trimOracle(t)
	path := suffixCases(t, oracle)
	want := run(t, "", oracle, "--suffix", path)
	entry, js, binary := buildHelper(t, "number_with_suffix.a", "suffix_main.a", "", "")
	for side, got := range observations(t, entry, js, binary, path) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s", side, firstDifference(got, want))
		}
	}
	t.Logf("%d exact Go bytes include results, one scanner call and its unchanged argument on all three targets", len(want))
}

// Not parallel: every mutant must compile and finish with no sanitizer or stderr failure.
func TestNumberWithSuffixMutants(t *testing.T) {
	oracle := trimOracle(t)
	path := suffixCases(t, oracle)
	want := run(t, "", oracle, "--suffix", path)
	for _, change := range []struct{ name, from, to string }{{"accept a zero-length numeric match", "if(consumed === 0) { return false; }", "if(consumed === 0) { return true; }"}, {"accept a suffix prefix", "if(tail === suffix)", "if(tail.startsWith(suffix))"}, {"scan twice", "const consumed = scanNumber(value);", "scanNumber(value);\n const consumed = scanNumber(value);"}} {
		t.Run(change.name, func(t *testing.T) {
			entry, js, binary := buildHelper(t, "number_with_suffix.a", "suffix_main.a", change.from, change.to)
			for side, got := range observations(t, entry, js, binary, path) {
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side)
				}
				t.Logf("compiled and completed cleanly; only Go comparison caught %s on %s: %s", change.name, side, firstDifference(got, want))
			}
		})
	}
}
