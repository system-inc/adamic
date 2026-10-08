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
	for _, directory := range []string{"later-ranked-callables/token-end", "later-ranked-callables/token-full-start", "later-ranked-callables/helper-factory", "later-ranked-callables/hoist-variable", "later-ranked-callables/modifier-flags", "later-ranked-callables/resolution-path", "later-ranked-callables/performance-measure", "later-ranked-callables/binary", "later-ranked-callables/declaration-name", "later-ranked-callables/left-access", "later-ranked-callables/source-files", "later-ranked-callables/type-reference", "later-ranked-callables/program-options", "later-ranked-callables/context-options", "later-ranked-callables/system-write", "later-ranked-callables/system-exit", "later-ranked-callables/null", "later-ranked-callables/parenthesized", "later-ranked-callables/disallowed-comma", "later-ranked-callables/token-start", "later-ranked-callables/true", "later-ranked-callables/local-name", "ranked-callables/block", "ranked-callables/array-literal", "ranked-callables/object-literal", "ranked-callables/function-expression", "ranked-callables/update-block", "ranked-callables/emit-notification", "ranked-callables/substitution", "ranked-callables/return-statement", "optional-aggregates/variable-statement", "optional-aggregates/parameter-declaration", "group1", "marker", "probes", "stored-marker", "canonical", "watcher", "scanner", "performance", "boxing", "numeric-literal", "array-callables/call-expression", "array-callables/inline-expressions", "element-access", "property-access", "property-assignment", "aggregate/expression-statement", "aggregate/diagnostic-add", "aggregate/void-zero", "aggregate/this", "aggregate/emit-helper", "aggregate/node-check-flag"} {
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
