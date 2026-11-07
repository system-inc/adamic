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
func TestEntriesAcceptance(test *testing.T) {
	test.Parallel()
	directory := os.Getenv("ADAMIC_ENTRIES_ACCEPTANCE")
	if directory == "" {
		directory = filepath.Join(repository, "stage3/fixtures/entries")
	}
	contents, err := os.ReadFile(filepath.Join(directory, "expectations.json"))
	if errors.Is(err, os.ErrNotExist) {
		test.Skip(
			"awaits compiler/entries-fixtures: typescript's Object.entries acceptance programs (8d864f9c) and the provenance fix they test (compiler/entries-provenance 05a60066), stage3/fixtures/entries, #tsdh6nb",
		)
	}
	if err != nil {
		test.Fatal(err)
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
		test.Fatal(err)
	}
	for _, contract := range contracts {
		source, err := os.ReadFile(filepath.Join(directory, contract.File))
		if err != nil {
			test.Fatal(err)
		}
		if strings.Count(string(source), contract.Mutant.Find) != 1 {
			test.Fatal("mutant site must occur once")
		}
		for _, variant := range []struct {
			name, source string
			expected     entriesAcceptanceContract
		}{
			{"baseline", string(source), contract.Expected},
			{
				"mutant",
				strings.Replace(string(source), contract.Mutant.Find, contract.Mutant.Replace, 1),
				contract.Mutant.Expected,
			},
		} {
			test.Run(contract.File+"/"+variant.name, func(test *testing.T) {
				test.Parallel()
				path := filepath.Join(test.TempDir(), contract.File)
				if err := os.WriteFile(path, []byte(variant.source), 0600); err != nil {
					test.Fatal(err)
				}
				node := onNode(test, path)
				if node.exitCode != 0 || len(node.stderr) != 0 {
					test.Fatalf("source Node: %#v", node)
				}
				program, err := lowered(test, path)
				if variant.expected.Outcome == "Refused" {
					var refusal *lower.Refused
					if !errors.As(err, &refusal) {
						test.Fatalf("want Refused, got %v", err)
					}
					reason := strings.ToLower(refusal.What + " " + refusal.Fix)
					for _, accepted := range variant.expected.Diagnostic {
						if strings.Contains(reason, accepted) {
							return
						}
					}
					test.Fatalf("refusal must name excluded property form: %v", err)
				}
				if err != nil {
					test.Fatal(err)
				}
				checks := 0
				inspect := func(value any) any {
					if call, ok := value.(ir.ObjectCall); ok && call.Method == "entries" {
						checks++
						if call.Checked != (variant.expected.Reflection == "checked") {
							test.Errorf("reflection=%s checked=%t", variant.expected.Reflection, call.Checked)
						}
					}
					if call, ok := value.(ir.RecordCall); ok && call.Method == "entries" {
						checks++
						// A homogeneous record proves its slot representation at writes;
						// open object views still require ObjectCall.Checked above.
						if variant.expected.Reflection != "proven" || call.Element != ir.Number {
							test.Errorf(
								"record entries: reflection=%s element=%v",
								variant.expected.Reflection,
								call.Element,
							)
						}
					}
					return value
				}
				mutateReadiness(program.Main, inspect)
				for _, function := range program.Functions {
					mutateReadiness(function.Body, inspect)
				}
				if checks != 1 {
					test.Fatalf("want one entries call, got %d", checks)
				}
				actual, _ := nativelyUncached(test, program)
				javascript := onJavaScriptBackend(test, program)
				want := node
				if expected := variant.expected.Runtime; expected.Exit == 70 {
					want = run{exitCode: expected.Exit, stdout: []byte(expected.Stdout), stderr: actual.stderr}
					if !strings.HasPrefix(string(actual.stderr), expected.Prefix) ||
						!strings.HasSuffix(string(actual.stderr), "\n") ||
						len(strings.TrimSpace(string(actual.stderr))) <= len(strings.TrimSpace(expected.Prefix)) {
						test.Fatalf("missing checked diagnostic: %#v", actual)
					}
				} else if node.stdout != nil && string(node.stdout) != expected.Stdout {
					test.Fatalf("source Node differs from acceptance golden: %#v", node)
				}
				for _, got := range []run{actual, releasedUncached(test, program), javascript} {
					if difference := disagreement(want, got); difference != "" {
						test.Fatalf("%s: want %#v got %#v", difference, want, got)
					}
				}
			})
		}
	}
}
