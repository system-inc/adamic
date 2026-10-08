package lint

import (
	"strings"
	"testing"
)

func TestReactNoUnusedStateSelected(t *testing.T) {
	oracle := goOracle(t)
	rows := []string{}
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "react/no-unused-state" {
			rows = append(rows, row)
		}
	}
	if len(rows) != 161 {
		t.Fatalf("upstream count %d", len(rows))
	}
	t.Logf("react/no-unused-state: %d upstream cases", len(rows))
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == "react/no-unused-state" {
			for _, row := range ownedWitnessRows(t, ".", d) {
				answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
				if string(answer.output) == "0\n" {
					t.Fatal("silent owned witness")
				}
				rows = append(rows, row)
				rows = append(rows, strings.SplitN(row, "\t", 2)[0]+"\tall")
			}
		}
	}
	compare(t, oracle, buildPort(t, ".", true), ".", manifest(t, rows))
}
