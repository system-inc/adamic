package loom

import (
	"strings"
	"testing"
)

// The landed rules.json reads cleanly, so the rule.changed that names its commit carries a value the queue takes.
func TestRulesFileReads(t *testing.T) {
	t.Parallel()
	rules, err := ReadRules()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("blocks on %v, budget %d; restack on %v", *rules.Blocks.On, *rules.Blocks.Budget, *rules.Restack.On)
}

// Each mistake a hand could make in rules.json is refused, naming what's wrong.
func TestRulesRefuseMistakes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		file string
		says string
	}{
		{"unknown rule", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "speed": {"on": true}}`, `unknown field "speed"`},
		{"unknown field", `{"blocks": {"on": true, "budget": 2, "size": 3}, "restack": {"on": false}}`, `unknown field "size"`},
		{"budget zero", `{"blocks": {"on": true, "budget": 0}, "restack": {"on": false}}`, "budget is 0"},
		{"budget negative", `{"blocks": {"on": true, "budget": -1}, "restack": {"on": false}}`, "budget is -1"},
		{"budget fraction", `{"blocks": {"on": true, "budget": 2.5}, "restack": {"on": false}}`, "cannot unmarshal number 2.5"},
		{"budget string", `{"blocks": {"on": true, "budget": "2"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"on string", `{"blocks": {"on": "true", "budget": 2}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"blocks on missing", `{"blocks": {"budget": 2}, "restack": {"on": false}}`, "blocks needs on and budget"},
		{"budget missing", `{"blocks": {"on": true}, "restack": {"on": false}}`, "blocks needs on and budget"},
		{"blocks missing", `{"restack": {"on": false}}`, "blocks needs on and budget"},
		{"restack missing", `{"blocks": {"on": true, "budget": 2}}`, "restack needs on"},
		{"restack on missing", `{"blocks": {"on": true, "budget": 2}, "restack": {}}`, "restack needs on"},
		{"something after", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}} {}`, "something after the rules object"},
		{"not an object", `[]`, "cannot unmarshal array"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseRules([]byte(testCase.file))
			if err == nil {
				t.Fatalf("%s was taken", testCase.file)
			}
			if !strings.Contains(err.Error(), testCase.says) {
				t.Fatalf("%s was refused with %q, not naming %q", testCase.file, err, testCase.says)
			}
		})
	}
}
