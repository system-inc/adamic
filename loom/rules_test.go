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
		{"sizing field missing", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 3, "maximum": 16, "target": 0.8}}, "restack": {"on": false}}`, "sizing needs minimum, maximum, target and window"},
		{"sizing unknown field", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 3, "maximum": 16, "target": 0.8, "window": 20, "floor": 1}}, "restack": {"on": false}}`, `unknown field "floor"`},
		{"sizing min zero", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 0, "maximum": 16, "target": 0.8, "window": 20}}, "restack": {"on": false}}`, "1 <= minimum <= maximum"},
		{"sizing max below min", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 5, "maximum": 4, "target": 0.8, "window": 20}}, "restack": {"on": false}}`, "1 <= minimum <= maximum"},
		{"sizing target one", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 3, "maximum": 16, "target": 1, "window": 20}}, "restack": {"on": false}}`, "target above 0 and below 1"},
		{"sizing target zero", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 3, "maximum": 16, "target": 0, "window": 20}}, "restack": {"on": false}}`, "target above 0 and below 1"},
		{"sizing window past the queue's", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 3, "maximum": 16, "target": 0.8, "window": 201}}, "restack": {"on": false}}`, "window from 1 to 200"},
		{"sizing window fraction", `{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 3, "maximum": 16, "target": 0.8, "window": 2.5}}, "restack": {"on": false}}`, "cannot unmarshal number 2.5"},
		{"sizing in the old spelling", `{"blocks": {"on": true, "budget": 2, "sizing": {"min": 3, "max": 16, "target": 0.8, "window": 20}}, "restack": {"on": false}}`, `unknown field "min"`},
		{"first not a list", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "first": "chg_2tps70xcxmasvvbzer45mfhtd4"}`, "cannot unmarshal string"},
		{"depth one", `{"blocks": {"on": true, "budget": 2, "depth": 1}, "restack": {"on": false}}`, "from 2 to 32"},
		{"depth past the queue's", `{"blocks": {"on": true, "budget": 2, "depth": 33}, "restack": {"on": false}}`, "from 2 to 32"},
		{"depth null", `{"blocks": {"on": true, "budget": 2, "depth": null}, "restack": {"on": false}}`, "a whole number from 2 to 32"},
		{"depth fraction", `{"blocks": {"on": true, "budget": 2, "depth": 2.5}, "restack": {"on": false}}`, "cannot unmarshal number 2.5"},
		{"depth string", `{"blocks": {"on": true, "budget": 2, "depth": "3"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"door bound unknown", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"budget": 2}}`, `unknown field "budget"`},
		{"door bound zero", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"waiting": 0}}`, "a whole number from 1 to 9007199254740991"},
		{"door bound null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"heavy": null}}`, "a whole number from 1 to 9007199254740991"},
		{"door censusWithin zero", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"censusWithin": 0}}`, "door's censusWithin is 0"},
		{"door censusWithin null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"censusWithin": null}}`, "door's censusWithin is null"},
		{"door bound fraction", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"lines": 2.5}}`, "cannot unmarshal number 2.5"},
		{"door bound string", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"lines": "100"}}`, "cannot unmarshal string"},
		{"door bound past the queue's", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": {"files": 9007199254740992}}`, "a whole number from 1 to 9007199254740991"},
		{"door null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": null}`, "an object of bounds"},
		{"door a list", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "door": [2]}`, "cannot unmarshal array"},
		{"builders twice", `{"blocks": {"on": true, "budget": 2, "builders": ["Adam", "Adam"]}, "restack": {"on": false}}`, `each once, not "Adam"`},
		{"builders past the queue's", `{"blocks": {"on": true, "budget": 2, "builders": ["a", "b", "c", "d", "e", "f", "g", "h", "i"]}, "restack": {"on": false}}`, "names 9 boxes, and it names 1 to 8 host names"},
		{"builders empty", `{"blocks": {"on": true, "budget": 2, "builders": []}, "restack": {"on": false}}`, "names 0 boxes, and it names 1 to 8 host names"},
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
		{"stacked null", `{"blocks": {"on": true, "budget": 2, "stacked": null}, "restack": {"on": false}}`, "true or false"},
		{"stacked a string", `{"blocks": {"on": true, "budget": 2, "stacked": "true"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"stacked a number", `{"blocks": {"on": true, "budget": 2, "stacked": 1}, "restack": {"on": false}}`, "cannot unmarshal number"},
		{"lanes null", `{"blocks": {"on": true, "budget": 2, "lanes": null}, "restack": {"on": false}}`, "true or false"},
		{"lanes a string", `{"blocks": {"on": true, "budget": 2, "lanes": "true"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"lanes a number", `{"blocks": {"on": true, "budget": 2, "lanes": 1}, "restack": {"on": false}}`, "cannot unmarshal number"},
		{"heavyMaximum zero", `{"blocks": {"on": true, "budget": 2, "heavyMaximum": 0}, "restack": {"on": false}}`, "a whole number from 1 to 16"},
		{"heavyMaximum past the queue's", `{"blocks": {"on": true, "budget": 2, "heavyMaximum": 17}, "restack": {"on": false}}`, "a whole number from 1 to 16"},
		{"heavyMaximum null", `{"blocks": {"on": true, "budget": 2, "heavyMaximum": null}, "restack": {"on": false}}`, "a whole number from 1 to 16"},
		{"probes below zero", `{"blocks": {"on": true, "budget": 2, "probes": -1}, "restack": {"on": false}}`, "a whole number from 0 to 8"},
		{"probes past the queue's", `{"blocks": {"on": true, "budget": 2, "probes": 9}, "restack": {"on": false}}`, "a whole number from 0 to 8"},
		{"probes null", `{"blocks": {"on": true, "budget": 2, "probes": null}, "restack": {"on": false}}`, "a whole number from 0 to 8"},
		{"probes a string", `{"blocks": {"on": true, "budget": 2, "probes": "2"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"rentedGrants null", `{"blocks": {"on": true, "budget": 2, "rentedGrants": null}, "restack": {"on": false}}`, "true or false"},
		{"rentedGrants a number", `{"blocks": {"on": true, "budget": 2, "rentedGrants": 0}, "restack": {"on": false}}`, "cannot unmarshal number"},
		{"measure null", `{"blocks": {"on": true, "budget": 2, "measure": null}, "restack": {"on": false}}`, "true or false"},
		{"measure a string", `{"blocks": {"on": true, "budget": 2, "measure": "true"}, "restack": {"on": false}}`, "cannot unmarshal string"},
		{"liveHoleHands a capital", `{"blocks": {"on": true, "budget": 2, "liveHoleHands": ["System_adamic"]}, "restack": {"on": false}}`, `not "System_adamic"`},
		{"liveHoleHands empty", `{"blocks": {"on": true, "budget": 2, "liveHoleHands": []}, "restack": {"on": false}}`, "1 to 8 hand names"},
		{"liveHoleHands twice", `{"blocks": {"on": true, "budget": 2, "liveHoleHands": ["system_adamic", "system_adamic"]}, "restack": {"on": false}}`, `each once, not "system_adamic"`},
		{"liveHoleHands nine", `{"blocks": {"on": true, "budget": 2, "liveHoleHands": ["a", "b", "c", "d", "e", "f", "g", "h", "i"]}, "restack": {"on": false}}`, "1 to 8 hand names"},
		{"liveHoleHands null", `{"blocks": {"on": true, "budget": 2, "liveHoleHands": null}, "restack": {"on": false}}`, "1 to 8 hand names"},
		{"rentedImages not a digest", `{"blocks": {"on": true, "budget": 2, "rentedImages": ["` + strings.Repeat("A", 64) + `"]}, "restack": {"on": false}}`, "image digests of 64 lowercase hex, each once, not"},
		{"rentedImages empty", `{"blocks": {"on": true, "budget": 2, "rentedImages": []}, "restack": {"on": false}}`, "1 to 16 image digests"},
		{"rentedImages seventeen", `{"blocks": {"on": true, "budget": 2, "rentedImages": [` + strings.TrimSuffix(strings.Repeat(`"`+strings.Repeat("a", 64)+`",`, 17), ",") + `]}, "restack": {"on": false}}`, "1 to 16 image digests"},
		{"buildAhead not a box", `{"blocks": {"on": true, "budget": 2, "buildAhead": ["9box"]}, "restack": {"on": false}}`, `not "9box"`},
		{"buildAhead empty", `{"blocks": {"on": true, "budget": 2, "buildAhead": []}, "restack": {"on": false}}`, "1 to 8 box names"},
		{"witness null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": null}`, "an object {every, hours, machines}"},
		{"witness empty", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {}}`, "whole number from 1 to"},
		{"witness unknown field", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Home-f3279c"], "on": true}}`, `unknown field "on"`},
		{"witness every zero", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 0, "hours": 4, "machines": ["Home-f3279c"]}}`, "witness's every is 0"},
		{"witness every past the queue's", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 1001, "hours": 4, "machines": ["Home-f3279c"]}}`, "witness's every is 1001"},
		{"witness every fraction", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 2.5, "hours": 4, "machines": ["Home-f3279c"]}}`, "cannot unmarshal number 2.5"},
		{"witness hours missing", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "machines": ["Home-f3279c"]}}`, "witness's hours is"},
		{"witness hours past a week", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 169, "machines": ["Home-f3279c"]}}`, "witness's hours is 169"},
		{"witness hours null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": null, "machines": ["Home-f3279c"]}}`, "witness's hours is null"},
		{"witness machines missing", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4}}`, "a list of 1 to 8 box names"},
		{"witness machines empty", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": []}}`, "a list of 1 to 8 box names"},
		{"witness machines null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": null}}`, "a list of 1 to 8 box names"},
		{"witness machines past the queue's", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["a", "b", "c", "d", "e", "f", "g", "h", "i"]}}`, "a list of 1 to 8 box names"},
		{"witness machines twice", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Home", "Home"]}}`, `each once, not "Home"`},
		{"witness machines a path", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Home/0"]}}`, `not "Home/0"`},
		{"witness a list", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": [10]}`, "cannot unmarshal array"},
		{"witness at past midnight", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "at": "24:00"}}`, `witness's at is "24:00"`},
		{"witness at without its zero", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "at": "9:00"}}`, `witness's at is "9:00"`},
		{"witness at null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "at": null}}`, "witness's at is null"},
		{"witness at beside fullBlock with no machines", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 1000, "hours": 36, "at": "10:00", "fullBlock": {"box": "Eve", "hours": 24}}}`, "absent only beside fullBlock with no at"},
		{"witness machines null beside fullBlock", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 1000, "hours": 36, "machines": null, "fullBlock": {"box": "Eve", "hours": 24}}}`, "a list of 1 to 8 box names"},
		{"witness fullBlock null", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "fullBlock": null}}`, "exactly {box, hours}"},
		{"witness fullBlock without hours", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "fullBlock": {"box": "Adam"}}}`, "exactly {box, hours}"},
		{"witness fullBlock unknown field", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "fullBlock": {"box": "Adam", "hours": 24, "on": true}}}`, `unknown field "on"`},
		{"witness fullBlock a path", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "fullBlock": {"box": "Adam/0", "hours": 24}}}`, `fullBlock's box is "Adam/0"`},
		{"witness fullBlock hours zero", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "fullBlock": {"box": "Adam", "hours": 0}}}`, "fullBlock's hours is 0"},
		{"witness fullBlock hours past a week", `{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], "fullBlock": {"box": "Adam", "hours": 169}}}`, "fullBlock's hours is 169"},
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

// A censusWithin the lander takes reads here too (Loom 0.8.510, the census door, #1kz9c3d), so a rules.json that turns
// the door on is taken by adamic's own reader: a whole number of minutes from 1 to 2^53-1, beside the other bounds, and
// absent as absent. Mutant that fails it: the field dropped from DoorRule (an unknown field, refused).
func TestRulesTakeACensusWithin(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "door": {"heavy": 1000, "censusWithin": 10}}`))
	if err != nil || rules.Door == nil || rules.Door.CensusWithin == nil || *rules.Door.CensusWithin != 10 || *rules.Door.Heavy != 1000 {
		t.Fatalf("%+v, %v", rules.Door, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "door": {"heavy": 1000}}`))
	if err != nil || rules.Door.CensusWithin != nil {
		t.Fatalf("%+v, %v", rules.Door, err)
	}
}

// Builders the lander takes read here too (Loom 0.8.375, the one-box loop, #mxbymz5): 1 to 8 host names, each once, and
// absent as absent, so today's file reads as it does. An empty list is refused, as the queue and the lander refuse it
// (#hp18dwz). Mutants: the duplicate check dropped (TestRulesRefuseMistakes' "builders twice"); an empty list taken
// ("builders empty").
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

// A stacked the lander takes reads here too (Loom 0.8.678, #tyncw8b), before any rules.json names it: true or false, and
// absent as absent. Mutant: a null taken as absent (TestRulesRefuseMistakes' "stacked null").
func TestRulesTakeAStacked(t *testing.T) {
	t.Parallel()
	for _, on := range []bool{true, false} {
		rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "stacked": ` + strconv.FormatBool(on) + `}, "restack": {"on": false}}`))
		if err != nil || rules.Blocks.Stacked == nil || *rules.Blocks.Stacked != on {
			t.Fatalf("%+v, %v", rules.Blocks, err)
		}
	}
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.Stacked != nil {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
}

// A lanes the queue takes reads here too (#sera6j6, #6jrhgp7), before any rules.json names it: true or false, and absent
// as absent, which is off, today's one tower. Mutant: a null taken as absent (TestRulesRefuseMistakes' "lanes null").
func TestRulesTakeALanes(t *testing.T) {
	t.Parallel()
	for _, on := range []bool{true, false} {
		rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "lanes": ` + strconv.FormatBool(on) + `}, "restack": {"on": false}}`))
		if err != nil || rules.Blocks.Lanes == nil || *rules.Blocks.Lanes != on {
			t.Fatalf("%+v, %v", rules.Blocks, err)
		}
	}
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.Lanes != nil {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
}

// A witness the lander takes reads here too (Loom 0.8.589, #6ayc2q8, before any rules.json names one, Node's order on
// #rjnac1n): every, hours and machines, each in the queue's range, and absent as absent, which turns it off. Mutant: the
// witness check dropped from ParseRules (TestRulesRefuseMistakes' "witness every zero" and the rest).
func TestRulesTakeAWitness(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Home-f3279c"]}}`))
	if err != nil || rules.Witness == nil || rules.Witness.Every != 10 || rules.Witness.Hours != 4 || len(rules.Witness.Machines) != 1 || rules.Witness.Machines[0] != "Home-f3279c" {
		t.Fatalf("%+v, %v", rules.Witness, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "witness": {"every": 1000, "hours": 168, "machines": ["a", "b", "c", "d", "e", "f", "g", "h"]}}`))
	if err != nil || rules.Witness.Every != 1000 || rules.Witness.Hours != 168 || len(rules.Witness.Machines) != 8 {
		t.Fatalf("%+v, %v", rules.Witness, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Witness != nil {
		t.Fatalf("%+v, %v", rules.Witness, err)
	}
}

// A witness's at and fullBlock read here as the lander and the queue take them (#h9p10te, #xtymtgv, before any rules.json
// names either): at a UTC "HH:MM" at both ends of the day, fullBlock exactly {box, hours} with hours at both ends of its
// range, and each absent as absent. Mutants: the at check dropped ("witness at past midnight" taken), the fullBlock check
// dropped ("witness fullBlock hours zero" taken), and either field missing from WitnessRule (this test's files refused as
// an unknown field).
func TestRulesTakeAWitnessAtAndFullBlock(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "witness": {"every": 1000, "hours": 36, "machines": ["Pupa-9c63e0"], "at": "10:00", "fullBlock": {"box": "Adam", "hours": 24}}}`))
	if err != nil || rules.Witness == nil || rules.Witness.At == nil || *rules.Witness.At != "10:00" || rules.Witness.FullBlock == nil || rules.Witness.FullBlock.Box != "Adam" || rules.Witness.FullBlock.Hours != 24 {
		t.Fatalf("%+v, %v", rules.Witness, err)
	}
	for _, edge := range []string{`"at": "00:00", "fullBlock": {"box": "Eve", "hours": 1}`, `"at": "23:59", "fullBlock": {"box": "Eve", "hours": 168}`} {
		if _, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"], ` + edge + `}}`)); err != nil {
			t.Fatalf("%s refused: %v", edge, err)
		}
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}, "witness": {"every": 10, "hours": 4, "machines": ["Pupa"]}}`))
	if err != nil || rules.Witness.At != nil || rules.Witness.FullBlock != nil {
		t.Fatalf("%+v, %v", rules.Witness, err)
	}
}

// A sizing rule reads by its unabbreviated names, the spelling the lander posts and the queue takes (#qt36cwf).
func TestRulesReadSizingByItsNames(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 2, "sizing": {"minimum": 3, "maximum": 16, "target": 0.8, "window": 20}}, "restack": {"on": false}}`))
	if err != nil {
		t.Fatal(err)
	}
	if sizing := rules.Blocks.Sizing; sizing == nil || *sizing.Minimum != 3 || *sizing.Maximum != 16 {
		t.Fatalf("sizing read as %+v, want minimum 3 and maximum 16", sizing)
	}
}

// The full block's rule alone reads here as the lander and the queue take it (#y9xkr81, Adamic Loom's ruling on
// #r9sbvck): every, hours and fullBlock, no machines and no at, so Home leaves the rule. Mutant: machines required again
// (this file refused, "a list of 1 to 8 box names").
func TestRulesTakeAFullBlockAlone(t *testing.T) {
	t.Parallel()
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 2}, "restack": {"on": false}, "witness": {"every": 1000, "hours": 36, "fullBlock": {"box": "Eve", "hours": 24}}}`))
	if err != nil {
		t.Fatal(err)
	}
	witness := rules.Witness
	if witness == nil || witness.Every != 1000 || witness.Hours != 36 || witness.Machines != nil || witness.At != nil {
		t.Fatalf("witness read as %+v, want every 1000, hours 36, no machines and no at", witness)
	}
	if witness.FullBlock == nil || witness.FullBlock.Box != "Eve" || witness.FullBlock.Hours != 24 {
		t.Fatalf("fullBlock read as %+v, want {Eve, 24}", witness.FullBlock)
	}
}

// Every blocks key the queue took since stacked reads here as the lander takes it (#hp18dwz), so a rules.json naming one
// lands: heavyMaximum, rentedImages, rentedGrants, buildAhead, measure, liveHoleHands and probes, each at its bounds,
// and each absent as absent. Mutants: liveHoleHands dropped from BlocksRule (this file refused as an unknown field); the
// hand pattern dropped (TestRulesRefuseMistakes' "liveHoleHands a capital" taken); an empty list taken ("liveHoleHands
// empty"); probes dropped from BlocksRule (refused as an unknown field); its bound's check dropped ("probes past the
// queue's" taken).
func TestRulesTakeTheBlocksKeysSinceStacked(t *testing.T) {
	t.Parallel()
	image := strings.Repeat("a", 64)
	rules, err := ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "heavyMaximum": 16, "rentedImages": ["` + image + `"], "rentedGrants": false, "buildAhead": ["Eve"], "measure": true, "liveHoleHands": ["system_adamic", "system_adamic_loom", "system_adamic_loom_landing"], "probes": 8}, "restack": {"on": false}}`))
	if err != nil {
		t.Fatal(err)
	}
	blocks := rules.Blocks
	if blocks.HeavyMaximum == nil || *blocks.HeavyMaximum != 16 || len(blocks.RentedImages) != 1 || blocks.RentedImages[0] != image || blocks.RentedGrants == nil || *blocks.RentedGrants || len(blocks.BuildAhead) != 1 || blocks.BuildAhead[0] != "Eve" || blocks.Measure == nil || !*blocks.Measure || len(blocks.LiveHoleHands) != 3 || blocks.LiveHoleHands[2] != "system_adamic_loom_landing" || blocks.Probes == nil || *blocks.Probes != 8 {
		t.Fatalf("blocks read as %+v", blocks)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4, "heavyMaximum": 1, "probes": 0}, "restack": {"on": false}}`))
	if err != nil || *rules.Blocks.HeavyMaximum != 1 || rules.Blocks.Probes == nil || *rules.Blocks.Probes != 0 {
		t.Fatalf("%+v, %v", rules.Blocks, err)
	}
	rules, err = ParseRules([]byte(`{"blocks": {"on": true, "budget": 4}, "restack": {"on": false}}`))
	if err != nil || rules.Blocks.HeavyMaximum != nil || rules.Blocks.RentedImages != nil || rules.Blocks.RentedGrants != nil || rules.Blocks.BuildAhead != nil || rules.Blocks.Measure != nil || rules.Blocks.LiveHoleHands != nil || rules.Blocks.Probes != nil {
		t.Fatalf("a file without them read %+v, %v", rules.Blocks, err)
	}
}

// The landed rules.json names the hands that may mark a live-hole fix (#hp18dwz step 3, Adamic Loom's order), so the
// queue's join can be turned on.
func TestRulesFileNamesTheLiveHoleHands(t *testing.T) {
	t.Parallel()
	rules, err := ReadRules()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"system_adamic", "system_adamic_loom", "system_adamic_loom_landing"}
	if strings.Join(rules.Blocks.LiveHoleHands, ",") != strings.Join(want, ",") {
		t.Fatalf("rules.json names live-hole hands %q, want %q", rules.Blocks.LiveHoleHands, want)
	}
}
