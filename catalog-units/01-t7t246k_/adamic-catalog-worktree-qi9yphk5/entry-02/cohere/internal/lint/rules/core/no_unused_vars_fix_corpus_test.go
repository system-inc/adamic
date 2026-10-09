package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * enableAutofixRemoval, and the default mode's suggestion, replayed through upstream (#e1zk9s0). The corpus
 * is every typescript-eslint 8.71.0 row that sets enableAutofixRemoval or asserts suggestions, plus edge
 * rows, each with the verdict the installed rule gave on the same bytes; no_unused_vars_fix_corpus_data_test.go
 * says how it was built.
 */

// noUnusedVarsFixCorpusDivergences are the rows cohere answers differently, by row index, with the reason
// and cohere's answer, so each difference is asserted rather than skipped.
var noUnusedVarsFixCorpusDivergences = map[int]struct {
	reason string
	cohere []noUnusedVarsFixCorpusFinding
	output string
}{
	// `import React` beside `import { h }` under the `h` pragma: upstream reads the program's jsxFactory and
	// reports React. This rule exempts an import named React or h in every file holding JSX, the
	// deliberate breadth isJsxFactoryImportInJsxFile documents, which predates these options.
	17: {"React is exempt in every JSX file, whatever the pragma", []noUnusedVarsFixCorpusFinding{}, ""},
	// Upstream attaches the removal to the usedIgnoredVar finding and deletes an import the file reads,
	// leaving `console.log(_a)` reading nothing. See no_unused_vars_fix.go.
	53: {"upstream's autofix deletes a used ignored import", []noUnusedVarsFixCorpusFinding{{"usedIgnoredVar", "_a", nil}}, ""},
	// Upstream removes a namespace import's whole declaration, deleting the used default beside it.
	55: {"upstream's autofix deletes a used default beside an unused namespace", []noUnusedVarsFixCorpusFinding{{"unusedVar", "Unused", nil}}, ""},
}

// noUnusedVarsFixTsconfig is the corpus tsconfig, for a JSX row with the pragma and fragment name upstream's
// parser options gave and the .tsx file in its include.
func noUnusedVarsFixTsconfig(t *testing.T, row noUnusedVarsFixCorpusCase) string {
	t.Helper()
	if !row.jsx {
		return noUnusedVarsOptionsCorpusTsconfig
	}
	var configuration struct {
		CompilerOptions map[string]any `json:"compilerOptions"`
		Include         []string       `json:"include"`
	}
	if err := json.Unmarshal([]byte(noUnusedVarsOptionsCorpusTsconfig), &configuration); err != nil {
		t.Fatal(err)
	}
	configuration.CompilerOptions["jsx"] = "react"
	if row.jsxFactory != "" {
		configuration.CompilerOptions["jsxFactory"] = row.jsxFactory
	}
	if row.jsxFragmentFactory != "" {
		configuration.CompilerOptions["jsxFragmentFactory"] = row.jsxFragmentFactory
	}
	configuration.Include = []string{"file.tsx"}
	encoded, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

// applyFixesOnce is ESLint's one pass: fixes sorted by start, then end, and one that starts at or before
// the last applied end is skipped. The edit engine resolves overlaps the same way, earliest claim first.
func applyFixesOnce(source string, fixes []rule.Fix) string {
	sorted := append([]rule.Fix{}, fixes...)
	sort.SliceStable(sorted, func(first, second int) bool {
		if sorted[first].Range.Pos() != sorted[second].Range.Pos() {
			return sorted[first].Range.Pos() < sorted[second].Range.Pos()
		}
		return sorted[first].Range.End() < sorted[second].Range.End()
	})
	output := ""
	cursor, last := 0, -1
	for _, fix := range sorted {
		if fix.Range.Pos() <= last {
			continue
		}
		output += source[cursor:fix.Range.Pos()] + fix.Text
		cursor = fix.Range.End()
		last = fix.Range.End()
	}
	return output + source[cursor:]
}

func TestNoUnusedVarsFixUpstreamCorpus(t *testing.T) {
	t.Parallel()

	upstream := 0
	for _, row := range noUnusedVarsFixCorpus {
		if row.edge == "" {
			upstream++
		}
	}
	if upstream != noUnusedVarsFixCorpusUpstreamRows {
		t.Fatalf("the corpus holds %d upstream rows, and upstream has %d", upstream, noUnusedVarsFixCorpusUpstreamRows)
	}

	for _, row := range noUnusedVarsFixCorpus {
		name := fmt.Sprintf("upstream-%d", row.index)
		if row.edge != "" {
			name = row.edge
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options, err := DecodeNoUnusedVarsOptions(json.RawMessage(row.options))
			if err != nil {
				t.Fatalf("decoding %q: %v", row.options, err)
			}
			fileName := "file.ts"
			if row.jsx {
				fileName = "file.tsx"
			}
			configuration := noUnusedVarsFixTsconfig(t, row)
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoUnusedVars, map[string]string{fileName: row.source}, fileName, options,
				func(directory string) {
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
						t.Fatal(err)
					}
				})

			diagnostics := result.Diagnostics
			sort.SliceStable(diagnostics, func(first, second int) bool {
				return diagnostics[first].Range.Pos() < diagnostics[second].Range.Pos()
			})
			reported := []noUnusedVarsFixCorpusFinding{}
			var fixes []rule.Fix
			for _, diagnostic := range diagnostics {
				id := diagnostic.Message.Id
				if upstreamId, renamed := noUnusedVarsUpstreamIds[id]; renamed {
					id = upstreamId
				}
				finding := noUnusedVarsFixCorpusFinding{id: id, text: row.source[diagnostic.Range.Pos():diagnostic.Range.End()]}
				for _, suggestion := range diagnostic.Suggestions {
					finding.suggestions = append(finding.suggestions, noUnusedVarsFixCorpusSuggestion{
						id:     suggestion.Message.Id,
						output: applyFixesOnce(row.source, suggestion.Fixes),
					})
				}
				reported = append(reported, finding)
				fixes = append(fixes, diagnostic.Fixes...)
			}
			output := ""
			if len(fixes) > 0 {
				output = applyFixesOnce(row.source, fixes)
			}

			want, wantOutput := row.findings, row.output
			if divergence, diverges := noUnusedVarsFixCorpusDivergences[row.index]; diverges {
				want, wantOutput = divergence.cohere, divergence.output
			}
			if fmt.Sprint(reported) != fmt.Sprint(want) {
				t.Fatalf("the rule reports %q, and the corpus records %q", reported, want)
			}
			if output != wantOutput {
				t.Fatalf("one pass of the fixes writes %q, and the corpus records %q", output, wantOutput)
			}
			rule_testing.RecordAssertedCase(t, result)
		})
	}
}
