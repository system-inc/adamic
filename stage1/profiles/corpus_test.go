package profiles

import (
	"encoding/json"
	"os"
	"testing"
)

type corpusRecord struct {
	Path string `json:"path"`
	Hash string `json:"sha256"`
}
type corpusList struct {
	Files []corpusRecord `json:"files"`
}

func disjoint(training, benchmarks []corpusRecord) bool {
	names, hashes := map[string]bool{}, map[string]bool{}
	for _, row := range benchmarks {
		names[row.Path] = true
		hashes[row.Hash] = true
	}
	seen := map[string]bool{}
	for _, row := range training {
		if names[row.Path] || hashes[row.Hash] || seen[row.Path] {
			return false
		}
		seen[row.Path] = true
	}
	return true
}
func TestTrainingNeverIncludesBenchmarks(t *testing.T) {
	var training, benchmarks corpusList
	for _, row := range []struct {
		name string
		list *corpusList
	}{{"training.json", &training}, {"benchmarks.json", &benchmarks}} {
		data, err := os.ReadFile(row.name)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, row.list); err != nil {
			t.Fatal(err)
		}
	}
	if len(training.Files) == 0 || len(benchmarks.Files) != 77 || !disjoint(training.Files, benchmarks.Files) {
		t.Fatal("training overlaps benchmark files or corpus is incomplete")
	}
	mutant := append(append([]corpusRecord{}, training.Files...), benchmarks.Files[0])
	if disjoint(mutant, benchmarks.Files) {
		t.Fatal("benchmark-file training mutant survived")
	}
	renamed := append(append([]corpusRecord{}, training.Files...), corpusRecord{"typescript/src/services/renamed.ts", benchmarks.Files[0].Hash})
	if disjoint(renamed, benchmarks.Files) {
		t.Fatal("renamed benchmark-content mutant survived")
	}
	t.Log("actual benchmark path and renamed benchmark-content mutants caught")
}
