package typescript

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * The corpus is typescript-eslint 8.71.0's own dot-notation rows plus edge rows, each with the verdict
 * the installed rule gave on the same bytes under the same tsconfig; dot_notation_corpus_data_test.go
 * says how it was built. Each row runs over a real program, so the private, protected and
 * index-signature questions are asked of a checker rather than passing vacuously on a nil one.
 */

// dotNotationTsconfig is the corpus tsconfig, with noPropertyAccessFromIndexSignature turned on for a
// row that sets flag, as upstream's tsconfig.noPropertyAccessFromIndexSignature.json does
func dotNotationTsconfig(t *testing.T, flag bool) string {
	t.Helper()
	if !flag {
		return dotNotationCorpusTsconfig
	}
	var configuration struct {
		CompilerOptions map[string]any `json:"compilerOptions"`
		Include         []string       `json:"include"`
	}
	if err := json.Unmarshal([]byte(dotNotationCorpusTsconfig), &configuration); err != nil {
		t.Fatal(err)
	}
	configuration.CompilerOptions["noPropertyAccessFromIndexSignature"] = true
	encoded, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func TestDotNotationUpstreamCorpus(t *testing.T) {
	t.Parallel()

	valid, invalid := 0, 0
	for _, row := range dotNotationCorpus {
		if row.edge != "" {
			continue
		}
		if len(row.findings) == 0 {
			valid++
		} else {
			invalid++
		}
	}
	if valid != dotNotationCorpusUpstreamValid || invalid != dotNotationCorpusUpstreamInvalid {
		t.Fatalf("the corpus holds %d valid and %d invalid upstream rows, and upstream has %d and %d",
			valid, invalid, dotNotationCorpusUpstreamValid, dotNotationCorpusUpstreamInvalid)
	}

	for _, row := range dotNotationCorpus {
		name := fmt.Sprintf("upstream-%d", row.index)
		if row.edge != "" {
			name = row.edge
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options, err := DecodeDotNotationOptions([]byte(row.options))
			if err != nil {
				t.Fatalf("decoding %q: %v", row.options, err)
			}
			configuration := dotNotationTsconfig(t, row.flag)
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, DotNotation, map[string]string{"file.ts": row.source}, "file.ts", options,
				func(directory string) {
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
						t.Fatal(err)
					}
				})

			diagnostics := result.Diagnostics
			sort.SliceStable(diagnostics, func(first, second int) bool {
				return diagnostics[first].Range.Pos() < diagnostics[second].Range.Pos()
			})
			reported := []dotNotationCorpusFinding{}
			for _, diagnostic := range diagnostics {
				reported = append(reported, dotNotationCorpusFinding{
					id:   diagnostic.Message.Id,
					text: row.source[diagnostic.Range.Pos():diagnostic.Range.End()],
				})
			}
			if fmt.Sprint(reported) != fmt.Sprint(row.findings) {
				t.Fatalf("the rule reports %q, and typescript-eslint 8.71.0 reports %q", reported, row.findings)
			}

			if row.output != "" {
				rule_testing.ExpectFixedSource(t, result, row.output)
				return
			}
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) > 0 {
					t.Errorf("the rule proposes a fix typescript-eslint 8.71.0 does not: %+v", diagnostic.Fixes)
				}
			}
		})
	}
}

// The fixture pair through the default harness, on api's shape: a test reaching a protected hook through
// brackets, which the dot form would make a type error (#qv7evaa)
func TestDotNotationFixturePair(t *testing.T) {
	t.Parallel()

	source := `
class Entity {
    protected clearChangedFields() {}
}
const entity = new Entity();
entity['clearChangedFields']();
`
	unconfigured, err := DecodeDotNotationOptions(nil)
	if err != nil {
		t.Fatal(err)
	}
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, DotNotation, "Case.ts", source, unconfigured), "useDot")

	allowed, err := DecodeDotNotationOptions([]byte(`{"allowProtectedClassPropertyAccess": true}`))
	if err != nil {
		t.Fatal(err)
	}
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, DotNotation, "Case.ts", source, allowed))
}

// Upstream's schema refuses a key it does not name, and an empty configuration is the defaults:
// keywords allowed and every typed exemption off
func TestDotNotationOptionsDecode(t *testing.T) {
	t.Parallel()

	if _, err := DecodeDotNotationOptions([]byte(`{"allowProtectedPropertyAccess": true}`)); err == nil {
		t.Error("a misspelled option decoded, so it would silently do nothing")
	}
	decoded, err := DecodeDotNotationOptions(nil)
	if err != nil {
		t.Fatal(err)
	}
	options := decoded.(DotNotationOptions)
	if options.AllowKeywords == nil || !*options.AllowKeywords ||
		options.AllowIndexSignaturePropertyAccess || options.AllowPrivateClassPropertyAccess || options.AllowProtectedClassPropertyAccess {
		t.Errorf("the defaults are %+v", options)
	}
}
