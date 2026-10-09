package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ignoreClassWithStaticInitBlock, ignoreUsingDeclarations and reportUsedIgnorePattern, replayed through
 * upstream (#e1zk9s0). The corpus is every typescript-eslint 8.71.0 row that sets one of the three, plus
 * edge rows for the order of upstream's gates, each with the verdict the installed rule gave on the same
 * bytes under the same tsconfig; no_unused_vars_options_corpus_data_test.go says how it was built.
 */

// noUnusedVarsUpstreamIds maps this rule's message ids to typescript-eslint's, the one place they differ.
var noUnusedVarsUpstreamIds = map[string]string{"noUnusedVars": "unusedVar"}

// noUnusedVarsOptionsCorpusDivergences are the upstream rows cohere answers differently, by row, with the
// reason and cohere's answer, so the difference is asserted rather than skipped and cannot drift.
var noUnusedVarsOptionsCorpusDivergences = map[int]struct {
	reason string
	cohere []noUnusedVarsOptionsCorpusFinding
}{
	// `namespace _Foo {}` read as a value is TypeScript's error TS2708: a namespace with no body has no
	// value, so the checker resolves the read to nothing and this rule, which counts reads by symbol,
	// never sees it. Upstream's scope analysis counts the reference anyway. The gap is in read detection,
	// on code the compiler rejects, and predates these options.
	25: {"a value read of an uninstantiated namespace (TS2708) resolves to no symbol", []noUnusedVarsOptionsCorpusFinding{}},
}

func TestNoUnusedVarsOptionsUpstreamCorpus(t *testing.T) {
	t.Parallel()

	upstream, pinned := 0, 0
	for _, row := range noUnusedVarsOptionsCorpus {
		if row.edge != "" {
			continue
		}
		upstream++
		if len(row.pinned) > 0 {
			pinned++
		}
	}
	if upstream != noUnusedVarsOptionsCorpusUpstreamRows || pinned != noUnusedVarsOptionsCorpusUpstreamPinned {
		t.Fatalf("the corpus holds %d upstream rows, %d pinned, and upstream has %d, %d pinned",
			upstream, pinned, noUnusedVarsOptionsCorpusUpstreamRows, noUnusedVarsOptionsCorpusUpstreamPinned)
	}

	for _, row := range noUnusedVarsOptionsCorpus {
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
			result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoUnusedVars, map[string]string{"file.ts": row.source}, "file.ts", options,
				func(directory string) {
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(noUnusedVarsOptionsCorpusTsconfig), 0o644); err != nil {
						t.Fatal(err)
					}
				})

			diagnostics := result.Diagnostics
			sort.SliceStable(diagnostics, func(first, second int) bool {
				return diagnostics[first].Range.Pos() < diagnostics[second].Range.Pos()
			})
			reported := []noUnusedVarsOptionsCorpusFinding{}
			for _, diagnostic := range diagnostics {
				id := diagnostic.Message.Id
				if upstreamId, renamed := noUnusedVarsUpstreamIds[id]; renamed {
					id = upstreamId
				}
				reported = append(reported, noUnusedVarsOptionsCorpusFinding{
					id:   id,
					text: row.source[diagnostic.Range.Pos():diagnostic.Range.End()],
				})
			}
			if divergence, diverges := noUnusedVarsOptionsCorpusDivergences[row.index]; diverges && row.edge == "" {
				if fmt.Sprint(reported) != fmt.Sprint(divergence.cohere) {
					t.Fatalf("row %d is a recorded divergence (%s) where cohere reports %q, and it now reports %q",
						row.index, divergence.reason, divergence.cohere, reported)
				}
				rule_testing.RecordAssertedCase(t, result)
				return
			}
			if fmt.Sprint(reported) != fmt.Sprint(row.findings) {
				t.Fatalf("the rule reports %q, and typescript-eslint 8.71.0 reports %q", reported, row.findings)
			}
			rule_testing.RecordAssertedCase(t, result)
		})
	}
}

// usedIgnoredVar names the pattern that set the binding aside, by the option that holds it, so a reader
// can tell which of four patterns to fix. Asserted on one row per kind rather than read off the constant.
func TestNoUnusedVarsUsedIgnoredMessageNamesItsPattern(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		options string
		want    string
	}{
		{"variable", "const _a = 1;\nconsole.log(_a);\nexport {};", `{"reportUsedIgnorePattern": true, "varsIgnorePattern": "^_"}`, "'_a' matches varsIgnorePattern ^_, which marks a variable"},
		{"argument", "export function f(_a: number) {\n  return _a;\n}", `{"reportUsedIgnorePattern": true, "argsIgnorePattern": "^_"}`, "'_a' matches argsIgnorePattern ^_, which marks an argument"},
		{"caught error", "try {\n} catch (_e) {\n  console.log(_e);\n}\nexport {};", `{"reportUsedIgnorePattern": true, "caughtErrorsIgnorePattern": "^_"}`, "'_e' matches caughtErrorsIgnorePattern ^_, which marks a caught error"},
		{"array element", "const [_a] = [1];\nconsole.log(_a);\nexport {};", `{"reportUsedIgnorePattern": true, "destructuredArrayIgnorePattern": "^_"}`, "'_a' matches destructuredArrayIgnorePattern ^_, which marks an element of array destructuring"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options, err := DecodeNoUnusedVarsOptions(json.RawMessage(testCase.options))
			if err != nil {
				t.Fatal(err)
			}
			result := rule_testing.RunTypedWithOptions(t, NoUnusedVars, "/repository/source/Case.ts", testCase.source, options)
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Message.Id != "usedIgnoredVar" {
				t.Fatalf("got %+v, want one usedIgnoredVar", result.Diagnostics)
			}
			description := result.Diagnostics[0].Message.Description
			if len(description) < len(testCase.want) || description[:len(testCase.want)] != testCase.want {
				t.Errorf("the message reads %q, want it to begin %q", description, testCase.want)
			}
		})
	}
}
