package program

import (
	"crypto/sha256"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// TestFilesThatResolveAlikeShareOneSelectionAndStillCountPerFile covers the selection memo (#9jpmqm9).
// Files that resolved the same way share one selection, so it is made once; a file that resolved
// differently gets its own; and the scoped-off and unconfigured counts, which coverage reports per file,
// are still counted once for every file rather than once for every selection.
func TestFilesThatResolveAlikeShareOneSelectionAndStillCountPerFile(t *testing.T) {
	t.Parallel()
	graph := &Graph{LintConfig: &configuration.Config{
		Rules: map[string]configuration.RuleSetting{
			"runs": {Severity: configuration.SeverityError},
			"off":  {Severity: configuration.SeverityOff},
		},
		Overrides: []configuration.Override{
			{Files: []string{"special/**"}, Rules: map[string]configuration.RuleSetting{"runs": {Severity: configuration.SeverityOff}}},
		},
	}}
	rules := []rule.Rule{{Name: "runs"}, {Name: "off"}, {Name: "nobody-configured"}}
	selections := map[any]*ruleSelection{}
	scopedOff, unconfigured := map[string]int{}, map[string]int{}

	first, _ := graph.rulesFor("source/a.ts", rules, selections, scopedOff, unconfigured)
	second, _ := graph.rulesFor("source/b.ts", rules, selections, scopedOff, unconfigured)
	special, _ := graph.rulesFor("special/c.ts", rules, selections, scopedOff, unconfigured)

	if first != second {
		t.Fatal("two files that resolved alike made two selections")
	}
	if special == first {
		t.Fatal("a file an override changed shared the selection of files it did not change")
	}
	if len(first.applicable) != 1 || first.applicable[0].Name != "runs" || len(special.applicable) != 0 {
		t.Fatalf("the selections apply %v and %v, want [runs] and none", ruleNames(first.applicable), ruleNames(special.applicable))
	}
	if scopedOff["off"] != 3 || scopedOff["runs"] != 1 || unconfigured["nobody-configured"] != 3 {
		t.Fatalf("counts are scoped off %v and unconfigured %v, want off 3, runs 1 and nobody-configured 3",
			scopedOff, unconfigured)
	}
	if len(selections) != 2 {
		t.Fatalf("three files in two resolutions made %d selections", len(selections))
	}
}

// Selections that give a derived rule different options fingerprint it apart, since an option can choose what
// the rule reads, and the walk computes each rule and options pair once, however many selections and workers
// ask (#s9k38p3). Sixteen workers, each with selections of its own as a walk's are, resolve three files: two
// overrides give the rule red, so their selections differ and their options do not.
func TestSelectionsFingerprintADerivedRuleOncePerOptionsForTheWholeWalk(t *testing.T) {
	t.Parallel()
	setting := func(marker string) configuration.RuleSetting {
		return configuration.RuleSetting{Severity: configuration.SeverityError, Options: []json.RawMessage{json.RawMessage(`"` + marker + `"`)}}
	}
	graph := &Graph{
		LintConfig: &configuration.Config{
			Rules: map[string]configuration.RuleSetting{"derived": setting("red")},
			Overrides: []configuration.Override{
				{Files: []string{"special/**"}, Rules: map[string]configuration.RuleSetting{"derived": setting("blue")}},
				{Files: []string{"again/**"}, Rules: map[string]configuration.RuleSetting{"derived": setting("red")}},
			},
		},
		RuleOptions: configuration.OptionsRegistry{"derived": {Decode: func(raw json.RawMessage) (any, error) {
			var marker string
			err := json.Unmarshal(raw, &marker)
			return marker, err
		}}},
	}
	var calls atomic.Int32
	rules := []rule.Rule{{Name: "derived", ProgramReads: rule.ReadsOtherFiles, ProgramFingerprint: func(_ rule.Program, options any) [sha256.Size]byte {
		calls.Add(1)
		return sha256.Sum256([]byte(options.(string)))
	}}}
	memo := &programFingerprintMemo{entries: map[programFingerprintKey]*programFingerprintEntry{}}

	const workers = 16
	var wait sync.WaitGroup
	var failures sync.Map
	for worker := range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			selections := map[any]*ruleSelection{}
			scopedOff, unconfigured := map[string]int{}, map[string]int{}
			for file, want := range map[string]string{"source/a.ts": "red", "special/c.ts": "blue", "again/d.ts": "red"} {
				selection, _ := graph.rulesFor(file, rules, selections, scopedOff, unconfigured)
				if got := graph.programFingerprints(selection, memo)["derived"]; got != sha256.Sum256([]byte(want)) {
					failures.Store(worker, file+" was not fingerprinted under "+want)
				}
			}
			if len(selections) != 3 {
				failures.Store(worker, "three files in three resolutions did not make three selections")
			}
		}()
	}
	wait.Wait()
	failures.Range(func(worker, failure any) bool {
		t.Errorf("worker %d: %s", worker, failure)
		return true
	})
	if calls.Load() != 2 {
		t.Fatalf("%d workers with 3 selections each over 2 options computed the fingerprint %d times, want once per options", workers, calls.Load())
	}
}
