package lint

import (
	"strings"
	"testing"
)

func TestReactRequireOptimizationSelected(t *testing.T) {
	oracle := goOracle(t)
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "react/require-optimization" {
			rows = append(rows, row)
		}
	}
	if len(rows) != 84 {
		t.Fatalf("upstream capture changed: got %d cases, want 84", len(rows))
	}
	t.Logf("react/require-optimization: %d unique captured upstream cases", len(rows))
	for _, d := range prepareRegistry(t, ".") {
		if d.Name != "react/require-optimization" {
			continue
		}
		for _, row := range ownedWitnessRows(t, ".", d) {
			count := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
			if string(count.output) == "0\n" {
				t.Fatal("witness reports no findings", row)
			}
			rows = append(rows, row)
		}
	}
	// Include both selected and all-rule witness runs, plus inherited rows that can reach this rule.
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == "react/require-optimization" {
			for _, row := range ownedWitnessRows(t, ".", d) {
				path := strings.SplitN(row, "\t", 2)[0]
				rows = append(rows, path+"\tall")
			}
		}
	}
	for _, row := range generated(t) {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 || fields[1] == "all" || fields[1] == "react/require-optimization" {
			rows = append(rows, row)
		}
	}
	t.Logf("comparison manifest: %d upstream, witness and inherited rows", len(rows))
	compare(t, oracle, buildPort(t, ".", true), ".", manifest(t, recoveryRows(t, oracle, rows)))
}
