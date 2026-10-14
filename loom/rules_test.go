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
	t.Logf("blocks on %v, budget %d, sizing %v; restack on %v; first %v", *rules.Blocks.On, *rules.Blocks.Budget, rules.Blocks.Sizing != nil, *rules.Restack.On, rules.First)
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
		{"first not a change id", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "first": ["chg_short"]}`, `not "chg_short"`},
		{"first twice", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "first": ["chg_2tps70xcxmasvvbzer45mfhtd4", "chg_2tps70xcxmasvvbzer45mfhtd4"]}`, "each once"},
		{"sizing field missing", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 3, "max": 16, "target": 0.8}}, "restack": {"on": false}}`, "sizing needs min, max, target and window"},
		{"sizing unknown field", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 3, "max": 16, "target": 0.8, "window": 20, "floor": 1}}, "restack": {"on": false}}`, `unknown field "floor"`},
		{"sizing min zero", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 0, "max": 16, "target": 0.8, "window": 20}}, "restack": {"on": false}}`, "1 <= min <= max"},
		{"sizing max below min", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 5, "max": 4, "target": 0.8, "window": 20}}, "restack": {"on": false}}`, "1 <= min <= max"},
		{"sizing target one", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 3, "max": 16, "target": 1, "window": 20}}, "restack": {"on": false}}`, "target above 0 and below 1"},
		{"sizing target zero", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 3, "max": 16, "target": 0, "window": 20}}, "restack": {"on": false}}`, "target above 0 and below 1"},
		{"sizing window past the queue's", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 3, "max": 16, "target": 0.8, "window": 201}}, "restack": {"on": false}}`, "window from 1 to 200"},
		{"sizing window fraction", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 3, "max": 16, "target": 0.8, "window": 2.5}}, "restack": {"on": false}}`, "cannot unmarshal number 2.5"},
		{"first not a list", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "first": "chg_2tps70xcxmasvvbzer45mfhtd4"}`, "cannot unmarshal string"},
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
