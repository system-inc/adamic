package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func mutantCorpus(test *testing.T, change oracles.Mutation, path, want string) (string, string) {
	test.Helper()
	families, ok := oracles.MutantFamilies[change.Name]
	if !ok {
		test.Fatalf("mutant %q has no proving corpus families", change.Name)
	}
	data, readError := os.ReadFile(strings.TrimSuffix(path, ".txt") + ".json")
	if readError != nil {
		test.Fatal(readError)
	}
	var specs []oracles.PrinterCase
	if decodeError := json.Unmarshal(data, &specs); decodeError != nil {
		test.Fatal(decodeError)
	}
	data, readError = os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	rows := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	answers := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(rows) != len(specs) || len(answers) != len(specs) {
		test.Fatal("mutant corpus lost input/answer correspondence")
	}
	selected := make(map[int]bool)
	for _, family := range families {
		var indices []int
		for index, spec := range specs {
			if spec.Label == family {
				indices = append(indices, index)
			}
		}
		if len(indices) == 0 {
			test.Fatalf("mutant %q: missing proving family %q", change.Name, family)
		}
		count := min(len(indices), 128)
		for index := 0; index < count; index++ {
			position := 0
			if count > 1 {
				position = index * (len(indices) - 1) / (count - 1)
			}
			selected[indices[position]] = true
		}
	}
	var input, expected strings.Builder
	for index := range rows {
		if selected[index] {
			input.WriteString(rows[index] + "\n")
			expected.WriteString(answers[index] + "\n")
		}
	}
	// A distinct basename bypasses executeShards: these already are small units.
	subset := filepath.Join(test.TempDir(), "witnesses.txt")
	if writeError := os.WriteFile(subset, []byte(input.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	test.Logf("%s: %d Go-oracle witnesses from %v", change.Name, len(selected), families)
	return subset, expected.String()
}
