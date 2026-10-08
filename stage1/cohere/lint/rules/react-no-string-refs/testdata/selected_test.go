package lint

import (
	"strings"
	"testing"
)

func TestReactNoStringRefsSelected(t *testing.T) {
	t.Parallel()
	oracle := goOracle(t)
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "react/no-string-refs" {
			rows = append(rows, row)
		}
	}
	if len(rows) != 65 {
		t.Fatalf("upstream capture changed: got %d cases, want 65", len(rows))
	}
	t.Logf("react/no-string-refs: %d captured upstream cases", len(rows))
	for _, d := range prepareRegistry(t, ".") {
		if d.Name != "react/no-string-refs" {
			continue
		}
		witnesses := ownedWitnessRows(t, ".", d)
		t.Logf("react/no-string-refs: %d witnesses", len(witnesses))
		for _, row := range witnesses {
			count := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
			if string(count.output) == "0\n" {
				t.Fatal("witness reports no findings", row)
			}
			rows = append(rows, row)
			rows = append(rows, strings.SplitN(row, "\t", 2)[0]+"\tall")
		}
	}
	for _, row := range generated(t) {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 || fields[1] == "all" || fields[1] == "react/no-string-refs" {
			rows = append(rows, row)
		}
	}
	t.Logf("comparison manifest: %d upstream, witness and inherited rows", len(rows))
	compare(t, oracle, buildPort(t, ".", true), ".", manifest(t, recoveryRows(t, oracle, rows)))
}
