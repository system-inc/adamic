package scanall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func scanAllSources(t *testing.T) ([]byte, []byte) {
	t.Helper()
	a, e := os.ReadFile("../../../internal/lower/lower.go")
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile("../../../internal/lower/refusals.go")
	if e != nil {
		t.Fatal(e)
	}
	return a, b
}
func TestScanAllCensusMainAnchorsStillMatch(t *testing.T) {
	t.Parallel()
	a, b := scanAllSources(t)
	if err := ValidateRuleInputs("../../.."); err != nil {
		t.Fatal(err)
	}
	sources, err := Rewrite(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 9 {
		t.Fatalf("unexpected overlay paths: %v", sources)
	}
	if !strings.Contains(sources["internal/lower/lower.go"], "measurement only; no IR output") {
		t.Fatal("output guard missing")
	}
	for _, text := range []string{"latentFullEnabled()", "latentBodyDiagnostics", "latentFindingOwner", "latentNestedDeclarations"} {
		if strings.Contains(sources["internal/lower/refusals.go"], text) {
			t.Fatalf("per-unit diagnostic exclusion remains: %s", text)
		}
	}
}
func TestScanAllCensusDriftNamesItsAnchor(t *testing.T) {
	t.Parallel()
	a, b := scanAllSources(t)
	for _, name := range []string{"lower.go", "refusals.go"} {
		aa, bb := a, b
		if name == "lower.go" {
			aa = append(append([]byte{}, a...), '\n')
		} else {
			bb = append(append([]byte{}, b...), '\n')
		}
		got, e := Rewrite(aa, bb)
		if got != nil || e == nil || !strings.Contains(e.Error(), name+" snapshot") {
			t.Fatalf("%s drift did not fail closed: %v", name, e)
		}
	}
}
func TestScanAllCensusScratchCannotBeProduction(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "repo")
	for _, out := range []string{root, filepath.Join(root, "build"), filepath.Join(root, "internal", "lower")} {
		if OutsideRepository(root, out) {
			t.Fatalf("accepted production path %s", out)
		}
	}
	if !OutsideRepository(root, filepath.Join(filepath.Dir(root), "scratch")) {
		t.Fatal("rejected outside scratch")
	}
}

func TestScanAllCensusRuleHeadBaseDriftFailsLoudly(t *testing.T) {
	t.Parallel()
	names := []string{"function_bind.go", "cycles.go", "graph_records.go", "records.go"}
	for _, mutated := range names {
		repo := t.TempDir()
		folder := filepath.Join(repo, "internal/lower")
		if err := os.MkdirAll(folder, 0700); err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			data, err := os.ReadFile(filepath.Join("../../../internal/lower", name))
			if err != nil {
				t.Fatal(err)
			}
			if name == mutated {
				data = append(data, '\n')
			}
			if err := os.WriteFile(filepath.Join(folder, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
		if err := ValidateRuleInputs(repo); err == nil || !strings.Contains(err.Error(), mutated+" snapshot") {
			t.Fatalf("%s drift did not name its anchor: %v", mutated, err)
		}
	}
}
