// Loaded through an overlay in cohere's react package. No tracked test edit.
package react

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
	"os"
	"testing"
)

func TestWave04ExportReactBlockers(t *testing.T) {
	cases := []struct {
		Name, Source, Declarations string
		Subject                    rule.Rule
	}{
		{"memo", reassignedContextCapture, preserveManualMemoizationReactDeclarations, PreserveManualMemoization},
		{"purity", purityImpureFunctionsInRender, "", Purity},
		{"refs", refsImport + refsCorpus[3].Source, refsReactDeclarations, Refs},
	}
	type finding struct {
		Rule, Id, Message string
		Start, End        int
	}
	type record struct {
		Name, Source, Declarations string
		Findings                   []finding
	}
	var records []record
	for _, c := range cases {
		result := rule_testing.RunTypedFiles(t, c.Subject, map[string]string{"/fixture.tsx": c.Source, "/react.d.ts": c.Declarations}, "/fixture.tsx")
		r := record{Name: c.Name, Source: c.Source, Declarations: c.Declarations}
		for _, d := range result.Diagnostics {
			r.Findings = append(r.Findings, finding{c.Subject.Name, d.Message.Id, d.Message.Description, d.Range.Pos(), d.Range.End()})
		}
		if len(r.Findings) == 0 {
			t.Fatalf("positive Go control silent: %s", c.Name)
		}
		records = append(records, r)
	}
	r := record{Name: "refs-reporting-arms"}
	for _, kind := range []refsFindingKind{refsFindingValueAccess, refsFindingPassedToFunction, refsFindingUpdate, refsFindingFunctionAccessesRef, refsFindingDidNotConverge} {
		m := refsMessageFor(kind)
		r.Findings = append(r.Findings, finding{Refs.Name, m.Id, m.Description, 0, 1})
	}
	records = append(records, r)
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("WAVE04_REACT_RECORDS"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
