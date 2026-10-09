package tsprinter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// Mutation tests need witnesses, while the agreement tests exercise every
// corpus input on unchanged ports. Select independent Go-oracle cases from
// the relevant generated families, keeping runs small even as tracked sources
// grow. Evenly spaced cases retain different operators, widths and contexts.
var mutantFamilies = map[string][]string{
	"program loses its statement separator":               {"program-sequence", "program-boundary"},
	"declaration loses its function keyword":              {"program-sequence", "program-boundary"},
	"assignment chain ignores statement boundaries":       {"statement-composition-regression", "variable-statement-edge"},
	"variable statement loses its keyword":                {"variable-statement-composition", "variable-statement-edge"},
	"await loses its keyword":                             {"await-composition", "await-yield-edge"},
	"yield loses delegation":                              {"yield-composition", "await-yield-edge"},
	"tagged template loses its tag":                       {"tagged-template-composition"},
	"template preview loses optional stopping boundaries": {"template-optional-boundary-composition"},
	"computed key forgets its clean text":                 {"computed-short-key-boundary"},
	"method loses its key":                                {"object-method-composition", "object-accessor-composition"},
	"function loses its keyword":                          {"named-function-composition"},
	"optional chain loses its stopping parentheses":       {"optional-chain-boundary"},
	"return and throw lose their keyword":                 {"named-function-composition", "arrow-composition"},
	"statement expression loses its semicolon":            {"named-function-composition", "statement-composition-regression"},
	"body loses its opening brace":                        {"named-function-composition", "arrow-composition"},
	"simple statement loses debugger":                     {"named-function-composition", "arrow-composition"},
	"arrow loses its arrow token":                         {"arrow-composition", "arrow-chain"},
	"template loses its interpolation dollar":             {"template-interpolation", "template-width"},
	"member chain loses its method dot":                   {"member-call-composition", "member-call-width"},
	"arguments lose their trailing comma":                 {"expanded-arguments", "argument-width"},
	"object loses its property colon":                     {"object-composition", "object-width"},
	"conditional swaps its separator":                     {"conditional-composition", "conditional-nesting"},
	"assignment loses its operator":                       {"assignment-composition", "assignment-chain"},
	"sequence loses a comma":                              {"sequence-composition", "sequence-width", "sequence"},
	"expression loses required parentheses":               {"operator-pair", "edge"},
}

func mutantCorpus(t *testing.T, change mutation, path, want string) (string, string) {
	t.Helper()
	families, ok := mutantFamilies[change.name]
	if !ok {
		t.Fatalf("mutant %q has no proving corpus families", change.name)
	}
	data, err := os.ReadFile(strings.TrimSuffix(path, ".txt") + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var specs []printerCase
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
			t.Fatalf("mutant %q: missing proving family %q", change.name, family)
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
	t.Logf("%s: %d Go-oracle witnesses from %v", change.name, len(selected), families)
	return subset, expected.String()
}

// Preparation belongs to the parent, not to timed shard subtests. Limit heavy
// compiler work to four workers and report all errors from the test goroutine.
func buildMutants(t *testing.T, paths, directories []string) []string {
	t.Helper()
	binaries := make([]string, len(paths))
	errors := make([]error, len(paths))
	workers := make(chan struct{}, min(runtime.GOMAXPROCS(0), 4))
	var group sync.WaitGroup
	for i := range paths {
		workers <- struct{}{}
		group.Add(1)
		go func(i int) {
			defer group.Done()
			defer func() { <-workers }()
			program, err := lowerProgram(paths[i])
			if err == nil {
				binaries[i], err = buildNative(program, directories[i])
			}
			if err != nil {
				errors[i] = fmt.Errorf("mutant %d (%s): %w", i, mutations[i].name, err)
			}
		}(i)
	}
	group.Wait()
	for _, err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
	return binaries
}
