package loom

import (
	"strconv"
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
		{"depth one", `{"blocks": {"on": true, "budget": 2, "depth": 1}, "restack": {"on": false}}`, "from 2 to 32"},
		{"depth past the queue's", `{"blocks": {"on": true, "budget": 2, "depth": 33}, "restack": {"on": false}}`, "from 2 to 32"},
		{"depth null", `{"blocks": {"on": true, "budget": 2, "depth": null}, "restack": {"on": false}}`, "a whole number from 2 to 32"},
		{"depth fraction", `{"blocks": {"on": true, "budget": 2, "depth": 2.5}, "restack": {"on": false}}`, "cannot unmarshal number 2.5"},
		{"depth string", `{"blocks": {"on": true, "budget": 2, "depth": "3"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"door bound unknown", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"budget": 2}}`, `unknown field "budget"`},
		{"door bound zero", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"waiting": 0}}`, "a whole number from 1 to 9007199254740991"},
		{"door bound null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"heavy": null}}`, "a whole number from 1 to 9007199254740991"},
		{"door bound fraction", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"lines": 2.5}}`, "cannot unmarshal number 2.5"},
		{"door bound string", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"lines": "100"}}`, "cannot unmarshal string"},
		{"door bound past the queue's", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"files": 9007199254740992}}`, "a whole number from 1 to 9007199254740991"},
		{"door null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": null}`, "an object of bounds"},
		{"door a list", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": [2]}`, "cannot unmarshal array"},
		{"builders twice", `{"blocks": {"on": true, "budget": 2, "builders": ["Adam", "Adam"]}, "restack": {"on": false}}`, `each once, not "Adam"`},
		{"builders past the queue's", `{"blocks": {"on": true, "budget": 2, "builders": ["a", "b", "c", "d", "e", "f", "g", "h", "i"]}, "restack": {"on": false}}`, "names 9 boxes, and it names at most 8"},
		{"builders a path", `{"blocks": {"on": true, "budget": 2, "builders": ["Adam", "Eve/0"]}, "restack": {"on": false}}`, `not "Eve/0"`},
		{"builders a leading dash", `{"blocks": {"on": true, "budget": 2, "builders": ["-Adam"]}, "restack": {"on": false}}`, `not "-Adam"`},
		{"builders empty name", `{"blocks": {"on": true, "budget": 2, "builders": [""]}, "restack": {"on": false}}`, `not ""`},
		{"builders null", `{"blocks": {"on": true, "budget": 2, "builders": null}, "restack": {"on": false}}`, "a list of host names"},
		{"builders a string", `{"blocks": {"on": true, "budget": 2, "builders": "Adam"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"builders not names", `{"blocks": {"on": true, "budget": 2, "builders": [1]}, "restack": {"on": false}}`, "cannot unmarshal number"},
		{"builders a digit first", `{"blocks": {"on": true, "budget": 2, "builders": ["0Adam"]}, "restack": {"on": false}}`, `not "0Adam"`},
		{"builders past a token's length", `{"blocks": {"on": true, "budget": 2, "builders": ["A` + strings.Repeat("a", 32) + `"]}, "restack": {"on": false}}`, "each once, not"},
		{"tipFirst null", `{"blocks": {"on": true, "budget": 2, "tipFirst": null}, "restack": {"on": false}}`, "true or false"},
		{"tipFirst a string", `{"blocks": {"on": true, "budget": 2, "tipFirst": "true"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"tipFirst a number", `{"blocks": {"on": true, "budget": 2, "tipFirst": 1}, "restack": {"on": false}}`, "cannot unmarshal number"},
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

// A depth the lander takes reads here too (Loom 0.8.71, #wpcbqpd): from 2 to 32, and absent as absent.
func TestRulesTakeADepth(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "depth": 2}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.Depth == nil || *rules.Blocks.Depth != 2 {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "depth": 32}, "restack": {"on": false}}`))
	if err != nil || *rules.Blocks.Depth != 32 {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.Depth != nil {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
}

// A door the lander takes reads here too (Loom 0.8.93, #qvcm8ez): each bound a whole number from 1 to 2^53-1, a bound
// absent as absent, and an empty door as no bounds, the rollback.
func TestRulesTakeADoor(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "door": {"heavy": 100000, "files": 9007199254740991}}`))
	if err != nil || rules.Door == nil || rules.Door.Heavy == nil || *rules.Door.Heavy != 100000 || *rules.Door.Files != 9007199254740991 || rules.Door.PacedAfter != nil || rules.Door.Waiting != nil || rules.Door.Lines != nil {
		t.Fatalf("%+v, %v", rules.Door, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "door": {}}`))
	if err != nil || rules.Door == nil || rules.Door.Heavy != nil {
		t.Fatalf("%+v, %v", rules.Door, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Door != nil {
		t.Fatalf("%+v, %v", rules.Door, err)
	}
}

// Builders the lander takes read here too (Loom 0.8.375, the one-box loop, #mxbymz5): host names each once, at most 8,
// an empty list as no box leased, and absent as absent, so today's file reads as it does. Mutant: the duplicate check
// dropped (TestRulesRefuseMistakes' "builders twice").
func TestRulesTakeBuilders(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "builders": ["Adam", "Eve"]}, "restack": {"on": false}}`))
	if err != nil || len(rules.Blocks.Builders) != 2 || rules.Blocks.Builders[0] != "Adam" || rules.Blocks.Builders[1] != "Eve" {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "builders": ["a", "b", "c", "d", "e", "f", "g", "h"]}, "restack": {"on": false}}`))
	if err != nil || len(rules.Blocks.Builders) != 8 {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "builders": []}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.Builders == nil || len(rules.Blocks.Builders) != 0 {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.Builders != nil {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
}

// A tipFirst the lander takes reads here too (Loom 0.8.375, #fz4vetb): true or false, and absent as absent. Mutant: a null
// taken as absent (TestRulesRefuseMistakes' "tipFirst null").
func TestRulesTakeATipFirst(t *testing.T) {
	t.Parallel()
	for _, on := range []bool{true, false} {
		rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "tipFirst": ` + strconv.FormatBool(on) + `}, "restack": {"on": false}}`))
		if err != nil || rules.Blocks.TipFirst == nil || *rules.Blocks.TipFirst != on {
			t.Fatalf("%+v, %v", rules.Blocks, err)
		}
	}
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.TipFirst != nil {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
}
