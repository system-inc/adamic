package json

import (
	"fmt"
	"github.com/system-inc/adamic/internal/gatesample"
	"os"
	"strings"
	"testing"
)

// Prettier is a separate upstream report. Only the nine named disagreements are known;
// an added difference or a closed difference requires updating the report explicitly.
func TestUpstreamRepositoryCorpusParity(t *testing.T) {
	if err := gatesample.Validate(); err != nil {
		t.Fatalf("%s: %v", t.Name(), err)
	}
	t.Parallel()
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		t.Skip("set ADAMIC_JSON_PRETTIER for the separate upstream report")
	}
	cases := sampledCorpusCases(t, 8)
	goAnswers, prettierAnswers := oracleAnswers(t, cases)
	differences := 0
	var report strings.Builder
	for index, item := range cases {
		if sameAnswer(goAnswers[index], prettierAnswers[index]) {
			continue
		}
		differences++
		known := item.Name == "stage1/cohere/json/gaps/numeric-separators.json" || strings.HasPrefix(item.Name, "generated/12/") || strings.HasPrefix(item.Name, "generated/13/") || strings.HasPrefix(item.Name, "generated/14/") || strings.HasPrefix(item.Name, "generated/16/")
		if !known {
			t.Errorf("unexpected upstream difference: %s", item.Name)
		}
		fmt.Fprintf(&report, "%s\nGo: %q error=%q\nPrettier: %q error=%q\n", item.Name,
			goAnswers[index].Output, goAnswers[index].Error, prettierAnswers[index].Output, prettierAnswers[index].Error)
		if differences <= 20 {
			t.Logf("difference: %s (Go error=%q, Prettier error=%q)", item.Name, goAnswers[index].Error, prettierAnswers[index].Error)
		}
	}
	if path := os.Getenv("ADAMIC_JSON_REPORT"); path != "" {
		if err := os.WriteFile(path, []byte(report.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	knownReport, err := os.ReadFile("known-upstream-differences.txt")
	if err != nil {
		t.Fatal(err)
	}
	if report.String() != string(knownReport) {
		t.Error("upstream difference identities or answers changed; update the checked-in report")
	}
	if differences != 9 {
		t.Fatalf("known upstream report changed: %d differences, want exactly 9", differences)
	}
	t.Logf("exactly nine known upstream differences in %d texts", len(cases))
}
