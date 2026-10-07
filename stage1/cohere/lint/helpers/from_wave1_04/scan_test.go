package wave104

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func scanCases(t *testing.T, oracle string) string {
	t.Helper()
	input, err := filepath.Abs("trim_fixture_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "scan-cases.json")
	run(t, "", oracle, "--generate-scan", input, path)
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
	names := map[string]int{}
	for _, c := range corpus.Cases {
		if c.Rule != "" {
			names[c.Rule]++
		}
	}
	if len(names) != 4 || len(corpus.Digits) != 10 || len(corpus.Cases) < 101000 {
		t.Fatalf("coverage consumers=%v digits=%v cases=%d", names, corpus.Digits, len(corpus.Cases))
	}
	for name, count := range names {
		t.Logf("%s: %d original fixture literal/field values", name, count)
	}
	t.Logf("%d cases: every BMP scalar following an ASCII prefix, every <=5-character word from +-01.eEX and 253 number/suffix combinations", len(corpus.Cases))
	return path
}

// Not parallel: hold the source and native compiler population within bounded resources.
func TestScanNumberMatchesGo(t *testing.T) {
	oracle := trimOracle(t)
	path := scanCases(t, oracle)
	want := run(t, "", oracle, "--scan", path)
	entry, js, binary := buildHelper(t, "scan_number.a", "scan_main.a", "", "")
	for side, got := range observations(t, entry, js, binary, path) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s", side, firstDifference(got, want))
		}
	}
	t.Logf("%d exact Go bytes on source Node, emitted JavaScript and ASan/UBSan native", len(want))
}

// Not parallel: each compiling mutant runs through every reference comparison.
func TestScanNumberMutants(t *testing.T) {
	oracle := trimOracle(t)
	path := scanCases(t, oracle)
	want := run(t, "", oracle, "--scan", path)
	for _, change := range []struct{ name, from, to string }{{"accept a dot without fractional digits", "if(index === digitsStart) { return 0; }", "if(index === digitsStart) { return index; }"}, {"consume an incomplete exponent", "if(exponent > digitsStart)", "if(true)"}} {
		t.Run(change.name, func(t *testing.T) {
			entry, js, binary := buildHelper(t, "scan_number.a", "scan_main.a", change.from, change.to)
			for side, got := range observations(t, entry, js, binary, path) {
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side)
				}
				t.Logf("compiled and completed cleanly; only Go comparison caught %s on %s: %s", change.name, side, firstDifference(got, want))
			}
		})
	}
}
