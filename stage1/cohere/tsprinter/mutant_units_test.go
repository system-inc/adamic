package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func mutantCorpus(t *testing.T, change oracles.Mutation, path, want string) (string, string) {
	t.Helper()
	families, ok := oracles.MutantFamilies[change.Name]
	if !ok {
		t.Fatalf("mutant %q has no proving corpus families", change.Name)
	}
	data, err := os.ReadFile(strings.TrimSuffix(path, ".txt") + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var specs []oracles.PrinterCase
	if err := json.Unmarshal(data, &specs); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	answers := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(rows) != len(specs) || len(answers) != len(specs) {
		t.Fatal("mutant corpus lost input/answer correspondence")
	}
	selected := make(map[int]bool)
	for _, family := range families {
		var indices []int
		for i, spec := range specs {
			if spec.Label == family {
				indices = append(indices, i)
			}
		}
		if len(indices) == 0 {
			t.Fatalf("mutant %q: missing proving family %q", change.Name, family)
		}
		count := min(len(indices), 128)
		for i := 0; i < count; i++ {
			position := 0
			if count > 1 {
				position = i * (len(indices) - 1) / (count - 1)
			}
			selected[indices[position]] = true
		}
	}
	var input, expected strings.Builder
	for i := range rows {
		if selected[i] {
			input.WriteString(rows[i] + "\n")
			expected.WriteString(answers[i] + "\n")
		}
	}
	// A distinct basename bypasses executeShards: these already are small units.
	subset := filepath.Join(t.TempDir(), "witnesses.txt")
	if err := os.WriteFile(subset, []byte(input.String()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: %d Go-oracle witnesses from %v", change.Name, len(selected), families)
	return subset, expected.String()
}
