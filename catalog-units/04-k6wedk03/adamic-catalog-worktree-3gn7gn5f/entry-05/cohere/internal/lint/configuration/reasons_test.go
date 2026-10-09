package configuration

import (
	"path/filepath"
	"testing"
)

// Every off states its reason where coverage prints it (#2qq4yr7): a rule turned off with no sentence
// saying why is an allowance, the same failure a departure with no reason was. A file says why under
// `reasons`, and the list is held to the guard `departures` has, so a stale entry cannot pass as a
// decision.

func TestAReasonedOffLoadsAndIsRecorded(t *testing.T) {
	t.Parallel()
	directory := writeConfigs(t, map[string]string{
		"CohereSettings.json": `{"rules": {"no-continue": "off"}, "reasons": {"no-continue": "  style, with no bug class behind it  "}}`,
	})
	loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
	offReason, recorded := loaded.OffReasons["no-continue"]
	if !recorded || offReason.Reason != "style, with no bug class behind it" {
		t.Fatalf("recorded %+v (%v), want the file's sentence, trimmed", offReason, recorded)
	}
	if offReason.File != filepath.Join(directory, "CohereSettings.json") {
		t.Errorf("the reason names %s, want the file that gave it", offReason.File)
	}
}

func TestAReasonThatExplainsNothingIsRefused(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name:  "a reason for a rule the file turns on",
			files: map[string]string{"CohereSettings.json": `{"rules": {"no-continue": "error"}, "reasons": {"no-continue": "left over from when it was off"}}`},
			want:  `names "no-continue" under "reasons", but it does not turn that rule off`,
		},
		{
			name:  "a reason for a rule the file never writes",
			files: map[string]string{"CohereSettings.json": `{"reasons": {"no-continue": "stale"}}`},
			want:  `names "no-continue" under "reasons", but it does not turn that rule off`,
		},
		{
			// The off is the base's, so its reason belongs in the base.
			name: "a reason for an off another file wrote",
			files: map[string]string{
				"base.json":           `{"rules": {"no-continue": "off"}}`,
				"CohereSettings.json": `{"extends": "./base.json", "reasons": {"no-continue": "explained from the wrong file"}}`,
			},
			want: `names "no-continue" under "reasons", but it does not turn that rule off`,
		},
		{
			name:  "a reason with no sentence",
			files: map[string]string{"CohereSettings.json": `{"rules": {"no-continue": "off"}, "reasons": {"no-continue": "   "}}`},
			want:  `names "no-continue" under "reasons" and gives no reason`,
		},
		{
			name: "a reason beside the departure that already says why",
			files: map[string]string{
				"base.json": `{"rules": {"no-var": "error"}}`,
				"CohereSettings.json": `{"extends": "./base.json", "rules": {"no-var": "off"},
					"departures": {"no-var": "a legacy script keeps var"}, "reasons": {"no-var": "said twice"}}`,
			},
			want: `names "no-var" under both "reasons" and "departures"`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			directory := writeConfigs(t, testCase.files)
			refusedWith(t, filepath.Join(directory, "CohereSettings.json"), testCase.want)
		})
	}
}

// A reason follows its rule up the chain: a project inheriting a reasoned off inherits the reason, a
// restated off keeps it, an off that departs from an inherited ruling carries the departure's reason,
// and a project turning the rule back on drops it.
func TestAReasonFollowsTheRuleUpTheChain(t *testing.T) {
	t.Parallel()
	base := `{"rules": {"no-continue": "off", "no-plusplus": "off", "no-var": "error"},
		"reasons": {"no-continue": "style", "no-plusplus": "style too"}}`
	directory := writeConfigs(t, map[string]string{
		"base.json": base,
		"CohereSettings.json": `{"extends": "./base.json",
			"rules": {"no-plusplus": "error", "no-continue": "off", "no-var": "off"},
			"departures": {"no-plusplus": "this project counts by hand", "no-var": "a legacy script keeps var"}}`,
	})
	loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))

	if offReason := loaded.OffReasons["no-continue"]; offReason.Reason != "style" || offReason.File != filepath.Join(directory, "base.json") {
		t.Errorf("a restated off recorded %+v, want the base's reason and file", offReason)
	}
	if offReason := loaded.OffReasons["no-var"]; offReason.Reason != "a legacy script keeps var" {
		t.Errorf("an off departing from an inherited error recorded %+v, want the departure's reason", offReason)
	}
	if offReason, recorded := loaded.OffReasons["no-plusplus"]; recorded {
		t.Errorf("a rule the project turned back on kept a reason for being off: %+v", offReason)
	}
}

// Once every set and repository carried its reasons, an off with no reason anywhere in its chain became
// a refusal at load, so the standard cannot rot back. Both ways: the unreasoned off is refused, naming the
// rule and the file that turned it off, wherever in the chain that file sits; the same off with a reason,
// its own or its departure's, loads.
// Not parallel: it swaps the package's setFiles for a stand-in tier set through withOurTierForTest
// (withSetFiles), which every test that reads a set would see.
func TestAnOffWithNoReasonIsRefusedAndAReasonedOneLoads(t *testing.T) {
	withOurTierForTest(t)
	t.Run("an unreasoned off in the project's own file is refused", func(t *testing.T) {
		t.Parallel()
		directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "cohere:system-inc/test", "rules": {"no-continue": "off"}}`})
		refusedWith(t, filepath.Join(directory, "CohereSettings.json"),
			`turns off "no-continue" in `+filepath.Join(directory, "CohereSettings.json")+` with no reason`)
	})

	t.Run("an unreasoned off in a base is refused, naming the base", func(t *testing.T) {
		t.Parallel()
		directory := writeConfigs(t, map[string]string{
			"base.json":           `{"extends": "cohere:system-inc/test", "rules": {"no-continue": "off"}}`,
			"CohereSettings.json": `{"extends": "./base.json"}`,
		})
		refusedWith(t, filepath.Join(directory, "CohereSettings.json"),
			`turns off "no-continue" in `+filepath.Join(directory, "base.json")+` with no reason`)
	})

	t.Run("the same off with a reason loads", func(t *testing.T) {
		t.Parallel()
		directory := writeConfigs(t, map[string]string{
			"CohereSettings.json": `{"rules": {"no-continue": "off"}, "reasons": {"no-continue": "style"}}`,
		})
		loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
		if !loaded.TurnsOffAtTopLevel("no-continue") {
			t.Error("TurnsOffAtTopLevel did not see the off")
		}
	})

	t.Run("an off its departure explains loads", func(t *testing.T) {
		t.Parallel()
		directory := writeConfigs(t, map[string]string{
			"base.json": `{"rules": {"no-var": "error"}}`,
			"CohereSettings.json": `{"extends": "./base.json", "rules": {"no-var": "off"},
				"departures": {"no-var": "a legacy script keeps var"}}`,
		})
		loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
	})
}

// withOurTierForTest carries an empty tier set, `cohere:system-inc/test`, in place of the embedded sets,
// so a test's chain can be one of ours by extending it, and be held to reasons.
func withOurTierForTest(t *testing.T) {
	t.Helper()
	withSetFiles(t, map[string]string{"system-inc/test": `{}`})
}

// Outside our tiers a project goes its own way (#bfxz13m): an off or a departure in its own files needs
// no reason, and loads. Each case is a refusal above, made again in a chain that reaches no tier set.
func TestOutsideOurTiersAProjectsOwnOffsAndDeparturesNeedNoReason(t *testing.T) {
	t.Parallel()
	cases := map[string]map[string]string{
		"an unreasoned off in the project's own file": {"CohereSettings.json": `{"rules": {"no-continue": "off"}}`},
		"an unreasoned off in a base": {
			"base.json":           `{"rules": {"no-continue": "off"}}`,
			"CohereSettings.json": `{"extends": "./base.json"}`,
		},
		"an undeclared departure": {
			"base.json":           `{"rules": {"no-implied-eval": "error"}}`,
			"CohereSettings.json": `{"extends": "./base.json", "rules": {"no-implied-eval": "off"}}`,
		},
		"an undeclared whole-tree override": {
			"base.json":           `{"rules": {"no-const-assign": "error"}}`,
			"CohereSettings.json": `{"extends": "./base.json", "overrides": [{"files": ["**/*.ts"], "rules": {"no-const-assign": "off"}}]}`,
		},
	}
	for name, files := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			directory := writeConfigs(t, files)
			loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
			if len(loaded.Departures) != 0 {
				t.Errorf("a departure with no reason was recorded as one: %+v", loaded.Departures)
			}
		})
	}

	// The unreasoned off is not given a reason it never had: coverage counts it as an off with no reason.
	directory := writeConfigs(t, cases["an undeclared departure"])
	loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
	if _, reasoned := loaded.OffReasonFor("no-implied-eval"); reasoned || !loaded.TurnsOffAtTopLevel("no-implied-eval") {
		t.Error("an outsider's unreasoned off must load as an off with no reason")
	}
}

// A set cohere carries is ours wherever it is read: an off in it with no reason is refused even in a
// chain that reaches no tier set.
// Not parallel: it swaps the package's setFiles for its own sets through withSetFiles, which every test
// that reads a set would see.
func TestASetsOwnUnreasonedOffIsRefusedInAnyChain(t *testing.T) {
	withSetFiles(t, map[string]string{"typescript": `{"rules": {"no-continue": "off"}}`})
	directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "cohere:typescript"}`})
	refusedWith(t, filepath.Join(directory, "CohereSettings.json"), `turns off "no-continue" in cohere:typescript with no reason`)
}

// The loader decides tier membership by InOurTiers over the whole chain, the predicate the format
// resolver asks too, so a tier set reached through a nested extends holds the project to reasons exactly
// as one named directly does. The same project with the tier set dropped from the chain loads.
// Not parallel: it swaps the package's setFiles for a stand-in tier set through withOurTierForTest
// (withSetFiles), which every test that reads a set would see.
func TestTierMembershipIsTheWholeChainsAndThePredicateFormatAsks(t *testing.T) {
	withOurTierForTest(t)
	inTier := writeConfigs(t, map[string]string{
		"middle.json":         `{"extends": "cohere:system-inc/test"}`,
		"base.json":           `{"extends": "./middle.json"}`,
		"CohereSettings.json": `{"extends": "./base.json", "rules": {"no-continue": "off"}}`,
	})
	path := filepath.Join(inTier, "CohereSettings.json")
	sources, err := SourcesOf(path)
	if err != nil || !InOurTiers(sources) {
		t.Fatalf("a nested tier set must make the chain ours by the shared predicate: %v %v", sources, err)
	}
	refusedWith(t, path, `turns off "no-continue" in `+path+` with no reason`)

	outside := writeConfigs(t, map[string]string{
		"middle.json":         `{}`,
		"base.json":           `{"extends": "./middle.json"}`,
		"CohereSettings.json": `{"extends": "./base.json", "rules": {"no-continue": "off"}}`,
	})
	loadOrFail(t, filepath.Join(outside, "CohereSettings.json"))
}
