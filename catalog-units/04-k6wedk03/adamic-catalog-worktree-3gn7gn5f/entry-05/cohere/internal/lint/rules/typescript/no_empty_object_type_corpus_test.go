package typescript

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * The corpus is typescript-eslint 8.71.0's own no-empty-object-type rows, all 42, plus edge rows, each
 * with the verdict and suggestions the installed rule gave on the same bytes;
 * no_empty_object_type_corpus_data_test.go says how it was built.
 */

func TestNoEmptyObjectTypeUpstreamCorpus(t *testing.T) {
	t.Parallel()

	valid, invalid := 0, 0
	for _, row := range noEmptyObjectTypeCorpus {
		if row.edge != "" {
			continue
		}
		if row.valid {
			valid++
		} else {
			invalid++
		}
	}
	if valid != noEmptyObjectTypeCorpusUpstreamValid || invalid != noEmptyObjectTypeCorpusUpstreamInvalid {
		t.Fatalf("the corpus holds %d valid and %d invalid upstream rows, and upstream has %d and %d",
			valid, invalid, noEmptyObjectTypeCorpusUpstreamValid, noEmptyObjectTypeCorpusUpstreamInvalid)
	}

	for _, row := range noEmptyObjectTypeCorpus {
		name := fmt.Sprintf("upstream-%d", row.index)
		if row.edge != "" {
			name = row.edge
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options, err := DecodeNoEmptyObjectTypeOptions([]byte(row.options))
			if err != nil {
				t.Fatalf("decoding %q: %v", row.options, err)
			}
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoEmptyObjectType, map[string]string{"file.ts": row.source}, "file.ts", options,
				func(directory string) {
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(noEmptyObjectTypeCorpusTsconfig), 0o644); err != nil {
						t.Fatal(err)
					}
				})

			diagnostics := result.Diagnostics
			sort.SliceStable(diagnostics, func(first, second int) bool {
				return diagnostics[first].Range.Pos() < diagnostics[second].Range.Pos()
			})
			reported := []noEmptyObjectTypeCorpusFinding{}
			for _, diagnostic := range diagnostics {
				if len(diagnostic.Fixes) > 0 {
					t.Fatalf("the rule proposes a fix, and upstream only ever suggests")
				}
				finding := noEmptyObjectTypeCorpusFinding{id: diagnostic.Message.Id, text: row.source[diagnostic.Range.Pos():diagnostic.Range.End()]}
				for _, suggestion := range diagnostic.Suggestions {
					finding.suggestions = append(finding.suggestions, noEmptyObjectTypeCorpusSuggestion{
						id:     suggestion.Message.Id,
						output: applySuggestion(t, row.source, suggestion),
					})
				}
				reported = append(reported, finding)
			}
			if fmt.Sprint(reported) != fmt.Sprint(row.findings) {
				t.Fatalf("the rule reports %q, and typescript-eslint 8.71.0 reports %q", reported, row.findings)
			}
			rule_testing.RecordAssertedCase(t, result)
		})
	}
}
