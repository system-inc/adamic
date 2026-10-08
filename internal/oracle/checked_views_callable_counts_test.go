package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func viewCallableCounts(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, directory := range []string{"later-ranked-callables/parameter-update", "later-ranked-callables/property-update", "later-ranked-callables/resolution-settings", "later-ranked-callables/scanner-scan", "later-ranked-callables/lexical-start", "later-ranked-callables/writer-space", "later-ranked-callables/substitution-hook", "later-ranked-callables/internal-name", "later-ranked-callables/call-update", "later-ranked-callables/class-update", "later-ranked-callables/prefix-operand", "later-ranked-callables/syntax-kind", "later-ranked-callables/static-block", "later-ranked-callables/prefix-unary", "later-ranked-callables/export-specifier", "later-ranked-callables/property-signature", "later-ranked-callables/type-check", "later-ranked-callables/lexical-end", "later-ranked-callables/module-format", "later-ranked-callables/false", "later-ranked-callables/parenthesized-type", "later-ranked-callables/variable-update-tagged", "later-ranked-callables/conditional-expression", "later-ranked-callables/binary-update", "later-ranked-callables/config-diagnostic", "later-ranked-callables/scanner-text", "later-ranked-callables/read-helpers", "later-ranked-callables/case-sensitive", "later-ranked-callables/diagnostic-newline", "later-ranked-callables/logical-and", "later-ranked-callables/named-exports", "later-ranked-callables/token-text", "later-ranked-callables/token-value", "later-ranked-callables/writer-line", "later-ranked-callables/watcher-close", "later-ranked-callables/source-file-path", "later-ranked-callables/emit-resolver", "later-ranked-callables/export-declaration", "later-ranked-callables/computed-name", "later-ranked-callables/source-file-update", "later-ranked-callables/context-diagnostic", "later-ranked-callables/token-end", "later-ranked-callables/token-full-start", "later-ranked-callables/helper-factory", "later-ranked-callables/hoist-variable", "later-ranked-callables/modifier-flags", "later-ranked-callables/resolution-path", "later-ranked-callables/performance-measure", "later-ranked-callables/binary", "later-ranked-callables/declaration-name", "later-ranked-callables/left-access", "later-ranked-callables/source-files", "later-ranked-callables/type-reference", "later-ranked-callables/program-options", "later-ranked-callables/context-options", "later-ranked-callables/system-write", "later-ranked-callables/system-exit", "later-ranked-callables/null", "later-ranked-callables/parenthesized", "later-ranked-callables/disallowed-comma", "later-ranked-callables/token-start", "later-ranked-callables/true", "later-ranked-callables/local-name", "ranked-callables/block", "ranked-callables/array-literal", "ranked-callables/object-literal", "ranked-callables/function-expression", "ranked-callables/update-block", "ranked-callables/emit-notification", "ranked-callables/substitution", "ranked-callables/return-statement", "optional-aggregates/variable-statement", "optional-aggregates/parameter-declaration", "group1", "marker", "probes", "stored-marker", "canonical", "watcher", "scanner", "performance", "boxing", "numeric-literal", "array-callables/call-expression", "array-callables/inline-expressions", "element-access", "property-access", "property-assignment", "aggregate/expression-statement", "aggregate/diagnostic-add", "aggregate/void-zero", "aggregate/this", "aggregate/emit-helper", "aggregate/node-check-flag"} {
		paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5", directory, "*.a"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			if name := filepath.Base(path); name == "write-back.a" || name == "observed.a" {
				continue
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, counted(t, relative, false, nil, false, false))
		}
	}
	for _, path := range []string{
		"stage3/interface-downcasts/untagged/fixtures/type-contract-boundary.a",
		"stage3/interface-downcasts/untagged/fixtures/base-type-boundary.a",
		"internal/lower/testdata/predicates/mixed_union_callback.a",
		"internal/lower/testdata/predicates/mixed_union_callback_variance.a",
	} {
		rows = append(rows, counted(t, path, false, nil, false, false))
	}
	return rows
}

// Not parallel: -update-counts writes only this lane's measured rows.
func TestCheckedViewCallableCounts(t *testing.T) {
	rows := viewCallableCounts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if *updateCounts {
		lines := strings.Split(strings.TrimSuffix(string(contents), "\n"), "\n")
		for _, row := range rows {
			key := strings.Split(row, " | ")[0] + " | "
			replaced := false
			for i, line := range lines {
				if strings.HasPrefix(line, key) {
					lines[i] = row
					replaced = true
					break
				}
			}
			if !replaced {
				lines = append(lines, row)
			}
		}
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, row := range rows {
		if !strings.Contains(string(contents), row+"\n") {
			t.Errorf("unrecorded callable counts: %s", row)
		}
	}
}
