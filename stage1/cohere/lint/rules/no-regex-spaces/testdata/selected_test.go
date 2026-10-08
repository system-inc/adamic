package lint

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOwnedRuleGoAgreement(t *testing.T) {
	oracle := goOracle(t)
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "no-regex-spaces" {
			rows = append(rows, row)
		}
	}
	data, err := os.ReadFile("helpers/literal/testdata/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var coverage map[string]struct {
		UpstreamCases int `json:"upstream_cases"`
	}
	if err = json.Unmarshal(data, &coverage); err != nil {
		t.Fatal(err)
	}
	if coverage["no-regex-spaces"].UpstreamCases != 153 {
		t.Fatal("upstream case count drift", coverage)
	}
	// The shared capture runs every upstream test and deduplicates identical inputs.
	if len(rows) != 95 {
		t.Fatalf("upstream count drift: got %d, want 95", len(rows))
	}
	t.Logf("no-regex-spaces: 153 upstream cases, %d unique source/rule/options combinations", len(rows))
	original := append([]string{}, rows...)
	for _, row := range original {
		rows = append(rows, strings.SplitN(row, "\t", 2)[0]+"\tall")
	}
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == "no-regex-spaces" {
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
