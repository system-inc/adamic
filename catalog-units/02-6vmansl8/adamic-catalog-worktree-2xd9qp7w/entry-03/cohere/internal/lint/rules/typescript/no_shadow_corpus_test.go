package typescript

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * The corpus is typescript-eslint 8.71.0's own no-shadow rows, all 239 across both test files, plus edge
 * rows, each with the verdict the installed rule gave on the same bytes under the same tsconfig;
 * no_shadow_corpus_data_test.go says how it was built. Each row runs over a real program, so the lib a
 * builtin global comes from is the program's, as it is for the installed rule under typed parsing.
 */

// noShadowShadowedPlace reads the shadowed declaration's line and column out of a message, as
// "line:column", or "global" for noShadowGlobal, which quotes none.
var noShadowShadowedPlace = regexp.MustCompile(`on line (\d+) column (\d+)\.$`)

// noShadowTsconfigFor is the corpus tsconfig with include naming the row's file.
func noShadowTsconfigFor(t *testing.T, fileName string) string {
	t.Helper()
	var configuration map[string]any
	if err := json.Unmarshal([]byte(noShadowCorpusTsconfig), &configuration); err != nil {
		t.Fatal(err)
	}
	configuration["include"] = []string{fileName}
	encoded, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func TestNoShadowUpstreamCorpus(t *testing.T) {
	t.Parallel()

	valid, invalid, pinned := 0, 0, 0
	for _, row := range noShadowCorpus {
		if row.edge != "" {
			continue
		}
		if len(row.pinned) > 0 {
			pinned++
		}
		// Upstream's direction, not the recorded verdict's: a pinned row's verdict is taken with the
		// inexpressible part out, which can turn an invalid row clean.
		if row.valid {
			valid++
		} else {
			invalid++
		}
	}
	if valid != noShadowCorpusUpstreamValid || invalid != noShadowCorpusUpstreamInvalid || pinned != noShadowCorpusUpstreamPinned {
		t.Fatalf("the corpus holds %d valid, %d invalid and %d pinned upstream rows, and upstream has %d, %d and %d",
			valid, invalid, pinned, noShadowCorpusUpstreamValid, noShadowCorpusUpstreamInvalid, noShadowCorpusUpstreamPinned)
	}

	for _, row := range noShadowCorpus {
		name := fmt.Sprintf("upstream-%d", row.index)
		if row.edge != "" {
			name = row.edge
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options, err := DecodeNoShadowOptions([]byte(row.options))
			if err != nil {
				t.Fatalf("decoding %q: %v", row.options, err)
			}
			configuration := noShadowTsconfigFor(t, row.fileName)
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoShadow, map[string]string{row.fileName: row.source}, row.fileName, options,
				func(directory string) {
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
						t.Fatal(err)
					}
				})

			diagnostics := result.Diagnostics
			sort.SliceStable(diagnostics, func(first, second int) bool {
				return diagnostics[first].Range.Pos() < diagnostics[second].Range.Pos()
			})
			reported := []noShadowCorpusFinding{}
			for _, diagnostic := range diagnostics {
				before := row.source[:diagnostic.Range.Pos()]
				line := strings.Count(before, "\n") + 1
				column := len(before) - strings.LastIndex(before, "\n")
				shadowed := "global"
				if place := noShadowShadowedPlace.FindStringSubmatch(diagnostic.Message.Description); place != nil {
					shadowed = place[1] + ":" + place[2]
				}
				reported = append(reported, noShadowCorpusFinding{
					id:       diagnostic.Message.Id,
					text:     row.source[diagnostic.Range.Pos():diagnostic.Range.End()],
					at:       fmt.Sprintf("%d:%d", line, column),
					shadowed: shadowed,
				})
			}
			if fmt.Sprint(reported) != fmt.Sprint(row.findings) {
				t.Fatalf("the rule reports %q, and typescript-eslint 8.71.0 reports %q", reported, row.findings)
			}
			rule_testing.RecordAssertedCase(t, result)
		})
	}
}

// The one thing cohere cannot express is pinned on the rows that carry it, by name, so a row cannot
// lose its pin and quietly start asserting a verdict it was never recorded under, and a second kind
// cannot appear without a reason being written for it.
func TestNoShadowCorpusPinsWhatCohereCannotExpress(t *testing.T) {
	t.Parallel()

	reasons := map[string]bool{
		// cohere carries no globals configuration, by the 1.0 contract (#bfxz13m item 4). The edge
		// rows make each pinned row's point with a global the program's lib or ESLint declares.
		"globals configuration": true,
		// There is no CommonJS function scope around a file.
		"ecmaFeatures.globalReturn": true,
	}
	counts := map[string]int{}
	for _, row := range noShadowCorpus {
		for _, feature := range row.pinned {
			if !reasons[feature] {
				t.Errorf("row %d pins %q, which has no reason here", row.index, feature)
			}
			counts[feature]++
		}
	}
	for feature := range reasons {
		if counts[feature] == 0 {
			t.Errorf("no row pins %q, so its reason above is stale", feature)
		}
	}
}
