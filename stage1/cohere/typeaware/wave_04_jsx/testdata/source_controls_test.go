package react

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
	"os"
	"testing"
)

func TestWave04JsxSourceControls(t *testing.T) {
	// Not parallel: this oracle writes one caller-supplied artifact.
	cases := []struct {
		Source  string
		Subject rule.Rule
	}{
		{"const App = () => <React.Fragment><div /></React.Fragment>;", JsxFragments},
		{"function App(){return <Ctx.Provider value={{a:1}} />;}", JsxNoConstructedContextValues},
		{"const x = <Missing />;", JsxNoUndef},
	}
	type finding struct {
		Rule, Id, Message string
		Start, End        int
	}
	type row struct {
		Source   string
		Findings []finding
	}
	rows := []row{}
	for _, c := range cases {
		result := rule_testing.RunTypedFiles(t, c.Subject, map[string]string{"/control.tsx": c.Source}, "/control.tsx")
		r := row{Source: c.Source}
		for _, d := range result.Diagnostics {
			r.Findings = append(r.Findings, finding{c.Subject.Name, d.Message.Id, d.Message.Description, d.Range.Pos(), d.Range.End()})
		}
		if len(r.Findings) == 0 {
			t.Fatalf("positive Go control silent: %s", c.Subject.Name)
		}
		rows = append(rows, r)
	}
	encoded, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("WAVE04_JSX_SOURCE"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
}
