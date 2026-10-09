package configuration

import (
	"path/filepath"
	"testing"
)

// TestPathsMatchingTheSameOverridesShareOneResolution covers the memo (#9jpmqm9). Two paths matching the
// same overrides resolve to one shared answer, and a path matching a different set gets its own, with the
// settings each set implies: the memo may save the copy, never change the result.
func TestPathsMatchingTheSameOverridesShareOneResolution(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Rules: map[string]RuleSetting{"no-var": {Severity: SeverityError}, "eqeqeq": {Severity: SeverityError}},
		Overrides: []Override{
			{Files: []string{"modules/**"}, Rules: map[string]RuleSetting{"no-var": {Severity: SeverityOff}}},
			{Files: []string{"**/*.test.ts"}, Rules: map[string]RuleSetting{"eqeqeq": {Severity: SeverityWarn}}},
		},
	}

	first := configuration.Resolve("modules/a.ts")
	second := configuration.Resolve("modules/b.ts")
	tested := configuration.Resolve("modules/c.test.ts")
	unmatched := configuration.Resolve("source/d.ts")

	if first.Identity() == nil || first.Identity() != second.Identity() {
		t.Fatalf("two paths matching the same override resolved apart: %v and %v", first.Identity(), second.Identity())
	}
	if tested.Identity() == first.Identity() || unmatched.Identity() == first.Identity() || tested.Identity() == unmatched.Identity() {
		t.Fatal("paths matching different overrides shared a resolution")
	}

	for _, check := range []struct {
		resolved Resolved
		rule     string
		want     Status
	}{
		{second, "no-var", StatusScopedOff},
		{second, "eqeqeq", StatusEnabled},
		{tested, "no-var", StatusScopedOff},
		{unmatched, "no-var", StatusEnabled},
	} {
		if status, _ := check.resolved.StatusOf(check.rule); status != check.want {
			t.Errorf("%s is %v, want %v", check.rule, status, check.want)
		}
	}
	if setting := tested.Rules["eqeqeq"]; setting.Severity != SeverityWarn {
		t.Errorf("the second override did not apply on top of the first: eqeqeq is %v", setting.Severity)
	}
}

// TestHouseVariantsNeverShareAResolution is the zero-config case @system_cohere_lint_sets named. A React
// file and a Node file can match the same overrides and still resolve through different house variants,
// so the memo lives per variant, after the dispatch, and the two never share an answer.
//
// Not parallel: withHouseSetsForTest swaps the package-level house sets for the length of the test.
func TestHouseVariantsNeverShareAResolution(t *testing.T) {
	withHouseSetsForTest(t)
	root := t.TempDir()
	component := filepath.Join(root, "source", "Button.tsx")
	server := filepath.Join(root, "server", "index.ts")
	loaded, err := LoadHouse(filepath.Join(root, "CohereSettings.json"), nil, HouseDetection{
		ReactFiles:      map[string]bool{component: true},
		TailwindSkipped: "no Tailwind stylesheet",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Resolved twice each, so a memo shared across variants would hand the second ask the first's answer.
	for range 2 {
		react, node := loaded.Resolve(component), loaded.Resolve(server)
		if react.Identity() == node.Identity() {
			t.Fatal("a React file and a Node file shared one resolution across house variants")
		}
		if !react.Enabled("react/no-danger") || node.Enabled("react/no-danger") {
			t.Fatalf("react/no-danger: %v in the React file and %v in the Node file, want only the first",
				react.Enabled("react/no-danger"), node.Enabled("react/no-danger"))
		}
	}
}
