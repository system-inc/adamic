package typescript

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * The corpus is typescript-eslint 8.71.0's own no-misused-promises rows that set checksConditionals, all
 * 34, as a boolean and as an object holding flagUnions, plus edge rows, each with the verdict the
 * installed rule gave on the same bytes; no_misused_promises_conditionals_corpus_data_test.go says how
 * it was built.
 */

func TestNoMisusedPromisesConditionalsUpstreamCorpus(t *testing.T) {
	t.Parallel()

	valid, invalid := 0, 0
	for _, row := range noMisusedPromisesConditionalsCorpus {
		if row.edge != "" {
			continue
		}
		if row.valid {
			valid++
		} else {
			invalid++
		}
	}
	if valid != noMisusedPromisesConditionalsCorpusUpstreamValid || invalid != noMisusedPromisesConditionalsCorpusUpstreamInvalid {
		t.Fatalf("the corpus holds %d valid and %d invalid upstream rows, and upstream has %d and %d",
			valid, invalid, noMisusedPromisesConditionalsCorpusUpstreamValid, noMisusedPromisesConditionalsCorpusUpstreamInvalid)
	}

	decode := rule.DecodeOptionsInto[NoMisusedPromisesOptions]()
	for _, row := range noMisusedPromisesConditionalsCorpus {
		name := fmt.Sprintf("upstream-%d", row.index)
		if row.edge != "" {
			name = row.edge
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options, err := decode([]byte(row.options))
			if err != nil {
				t.Fatalf("decoding %q: %v", row.options, err)
			}
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoMisusedPromises, map[string]string{"file.ts": row.source}, "file.ts", options,
				func(directory string) {
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(noMisusedPromisesConditionalsCorpusTsconfig), 0o644); err != nil {
						t.Fatal(err)
					}
				})

			diagnostics := result.Diagnostics
			sort.SliceStable(diagnostics, func(first, second int) bool {
				return diagnostics[first].Range.Pos() < diagnostics[second].Range.Pos()
			})
			reported := []noMisusedPromisesConditionalsCorpusFinding{}
			for _, diagnostic := range diagnostics {
				reported = append(reported, noMisusedPromisesConditionalsCorpusFinding{
					id:   diagnostic.Message.Id,
					text: row.source[diagnostic.Range.Pos():diagnostic.Range.End()],
				})
			}
			if fmt.Sprint(reported) != fmt.Sprint(row.findings) {
				t.Fatalf("the rule reports %q, and typescript-eslint 8.71.0 reports %q", reported, row.findings)
			}
			rule_testing.RecordAssertedCase(t, result)
		})
	}
}

// The object spelling refuses what upstream's schema refuses, before any run: an unknown flagUnions
// value, and a key beside flagUnions.
func TestNoMisusedPromisesConditionalsObjectRefusesOtherShapes(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[NoMisusedPromisesOptions]()
	for _, raw := range []string{
		`{"checksConditionals": {"flagUnions": "some"}}`,
		`{"checksConditionals": {"flagUnion": "all"}}`,
		`{"checksConditionals": "all"}`,
	} {
		if _, err := decode([]byte(raw)); err == nil {
			t.Errorf("%s decoded, and upstream's schema refuses it", raw)
		}
	}
}
