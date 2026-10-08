package javascript

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"os"
	"testing"
)

func TestScoutFormatterWitnesses(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(os.Getenv("ADAMIC_SCOUT_FIXTURES"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ Label, Source, Go, Prettier, Reason string }
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("empty scout fixtures")
	}
	for i := range cases {
		path := "witness.ts"
		if cases[i].Label == "jsx" {
			path = "witness.tsx"
		}
		text, err := Format(path, cases[i].Source, formatoptions.Default(), nil)
		if err != nil {
			t.Fatalf("%s: %v", cases[i].Label, err)
		}
		if cases[i].Go != "" && cases[i].Go != text {
			t.Fatalf("%s Go bytes changed: %q", cases[i].Label, text)
		}
		cases[i].Go = text
	}
	data, err = json.MarshalIndent(cases, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("ADAMIC_SCOUT_ANSWERS"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}
