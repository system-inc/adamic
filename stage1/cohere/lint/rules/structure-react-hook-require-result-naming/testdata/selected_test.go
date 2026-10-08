package lint

import (
	"strings"
	"testing"
)

func TestOwnedRuleGoAgreement(t *testing.T) {
	oracle := goOracle(t)
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "structure/react-hook-require-result-naming" {
			rows = append(rows, row)
		}
	}
	if len(rows) != 51 {
		t.Fatalf("upstream count drift: got %d, want 51", len(rows))
	}
	t.Logf("structure/react-hook-require-result-naming: %d unique upstream cases, including optional-node guards", len(rows))
	original := append([]string{}, rows...)
	for _, row := range original {
		rows = append(rows, strings.SplitN(row, "\t", 2)[0]+"\tall")
	}
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == "structure/react-hook-require-result-naming" {
			for _, row := range ownedWitnessRows(t, ".", d) {
				if string(execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count").output) == "0\n" {
					t.Fatal("witness has no Go finding", row)
				}
				rows = append(rows, row, strings.SplitN(row, "\t", 2)[0]+"\tall")
			}
		}
	}
	path := manifest(t, rows)
	want := execute(t, "", oracle, "--manifest", path).output
	if diff := difference(node(t, ".", path, false).output, want); diff != "" {
		t.Fatal("Node source differs from Go:", diff)
	}
	compare(t, oracle, buildPort(t, ".", true), ".", path)
}
