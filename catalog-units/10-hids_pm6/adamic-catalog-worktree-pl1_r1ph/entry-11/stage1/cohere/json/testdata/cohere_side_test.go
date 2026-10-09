// audit_test.go overlays this file into Go cohere. The submodule is never edited.
package javascript

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/format/formatoptions"
)

func TestAdamicJSONAudit(t *testing.T) {
	casesPath := os.Getenv("ADAMIC_JSON_CASES")
	if casesPath == "" {
		t.Skip("run by stage1/cohere/json/audit_test.go")
	}
	encoded, err := os.ReadFile(casesPath)
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(encoded, &cases); err != nil {
		t.Fatal(err)
	}
	type answer struct {
		Output string `json:"output"`
		Error  string `json:"error"`
	}
	answers := make([]answer, 0, len(cases))
	start := time.Now()
	for _, item := range cases {
		output, err := FormatJSON(item.Name, item.Text, formatoptions.PrettierDefaults())
		result := answer{Output: output}
		if err != nil {
			result.Error = err.Error()
		}
		answers = append(answers, result)
	}
	t.Logf("%d texts in %.3fs", len(cases), time.Since(start).Seconds())
	encoded, err = json.Marshal(answers)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("ADAMIC_JSON_ANSWERS"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
}
