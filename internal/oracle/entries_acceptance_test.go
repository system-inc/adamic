package oracle

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

type entriesAcceptanceContract struct {
	Outcome    string   `json:"outcome"`
	Reflection string   `json:"reflection"`
	Diagnostic []string `json:"diagnostic_any_of"`
	Runtime    *struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Prefix string `json:"stderr_prefix"`
		Exit   int    `json:"exit"`
	} `json:"runtime"`
}

// The acceptance bucket is a separate worker's dependency. Do not merge or
// rewrite it: run its pinned source and Node observations when supplied.
func TestEntriesAcceptance(t *testing.T) {
	t.Parallel()
	directory := os.Getenv("ADAMIC_ENTRIES_ACCEPTANCE")
	if directory == "" {
		directory = filepath.Join(repository, "stage3/fixtures/entries")
	}
	contents, err := os.ReadFile(filepath.Join(directory, "expectations.json"))
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("awaits codex/step12-entries-fixtures: typescript's Object.entries acceptance programs (8d864f9c), stage3/fixtures/entries")
	}
	if err != nil {
		t.Fatal(err)
	}
	var contracts []struct {
		File     string                    `json:"file"`
		Expected entriesAcceptanceContract `json:"expected"`
		Mutant   struct {
			Find     string                    `json:"find"`
			Replace  string                    `json:"replace"`
			Expected entriesAcceptanceContract `json:"expected"`
		} `json:"mutant"`
	}
	if err := json.Unmarshal(contents, &contracts); err != nil {
		t.Fatal(err)
	}
	for _, contract := range contracts {
		source, err := os.ReadFile(filepath.Join(directory, contract.File))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), contract.Mutant.Find) != 1 {
			t.Fatal("mutant site must occur once")
		}
		for _, variant := range []struct {
			name, source string
			expected     entriesAcceptanceContract
		}{
			{"baseline", string(source), contract.Expected},
			{"mutant", strings.Replace(string(source), contract.Mutant.Find, contract.Mutant.Replace, 1), contract.Mutant.Expected},
		} {
			t.Run(contract.File+"/"+variant.name, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), contract.File)
				if err := os.WriteFile(path, []byte(variant.source), 0600); err != nil {
					t.Fatal(err)
				}
				node := onNode(t, path)
				if node.exitCode != 0 || len(node.stderr) != 0 {
					t.Fatalf("source Node: %#v", node)
				}
				program, err := lowered(t, path)
				if variant.expected.Outcome == "Refused" {
					var refusal *lower.Refused
					if !errors.As(err, &refusal) {
						t.Fatalf("want Refused, got %v", err)
					}
					reason := strings.ToLower(refusal.What + " " + refusal.Fix)
					for _, accepted := range variant.expected.Diagnostic {
						if strings.Contains(reason, accepted) {
							return
						}
					}
					t.Fatalf("refusal must name excluded property form: %v", err)
				}
				if err != nil {
					t.Fatal(err)
				}
				checks := 0
				inspect := func(value any) any {
					if call, ok := value.(ir.ObjectCall); ok && call.Method == "entries" {
						checks++
						if call.Checked != (variant.expected.Reflection == "checked") {
							t.Errorf("reflection=%s checked=%t", variant.expected.Reflection, call.Checked)
						}
					}
					return value
				}
				mutateReadiness(program.Main, inspect)
				for _, function := range program.Functions {
					mutateReadiness(function.Body, inspect)
				}
				if checks != 1 {
					t.Fatalf("want one entries call, got %d", checks)
				}
				actual, _ := nativelyUncached(t, program)
				javascript := onJavaScriptBackend(t, program)
				want := node
				if expected := variant.expected.Runtime; expected.Exit == 70 {
					want = run{exitCode: expected.Exit, stdout: []byte(expected.Stdout), stderr: actual.stderr}
					if !strings.HasPrefix(string(actual.stderr), expected.Prefix) || !strings.HasSuffix(string(actual.stderr), "\n") || len(strings.TrimSpace(string(actual.stderr))) <= len(strings.TrimSpace(expected.Prefix)) {
						t.Fatalf("missing checked diagnostic: %#v", actual)
					}
				} else if node.stdout != nil && string(node.stdout) != expected.Stdout {
					t.Fatalf("source Node differs from acceptance golden: %#v", node)
				}
				for _, got := range []run{actual, releasedUncached(t, program), javascript} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s: want %#v got %#v", difference, want, got)
					}
				}
			})
		}
	}
}
