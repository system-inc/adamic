package lint

import (
	"strings"
	"testing"
)

func TestRequireNamedExportSelected(t *testing.T) {
	t.Parallel()
	oracle := goOracle(t)
	var rows []string
	count := 0
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "structure/react-component-require-named-export" {
			rows = append(rows, row)
			count++
		}
	}
	if count == 0 {
		t.Fatal("no captured cases")
	}
	t.Logf("captured upstream cases: %d", count)
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == "structure/react-component-require-named-export" {
			for _, row := range ownedWitnessRows(t, ".", d) {
				rows = append(rows, row)
				rows = append(rows, strings.Split(row, "\t")[0]+"\tall")
			}
		}
	}
	for _, row := range generated(t) {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 || fields[1] == "all" {
			rows = append(rows, row)
		}
	}
	compare(t, oracle, buildPort(t, ".", true), ".", manifest(t, recoveryRows(t, oracle, rows)))
}
