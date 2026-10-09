package lint

import (
	"strings"
	"testing"
)

func TestReactStateInConstructorSelected(t *testing.T) {
	oracle := goOracle(t)
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "react/state-in-constructor" {
			rows = append(rows, row)
		}
	}
	if len(rows) != 82 {
		t.Fatalf("upstream case count drift: %d", len(rows))
	}
	t.Logf("react/state-in-constructor: %d captured upstream cases", len(rows))
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == "react/state-in-constructor" {
			rows = append(rows, recoveryRows(t, oracle, ownedWitnessRows(t, ".", d))...)
		}
	}
	compare(t, oracle, buildPort(t, ".", true), ".", manifest(t, rows))
}
