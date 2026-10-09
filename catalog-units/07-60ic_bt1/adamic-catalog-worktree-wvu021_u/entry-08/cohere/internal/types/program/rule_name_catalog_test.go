package program

import "testing"

// TestRuleNameCatalogExistsIsTheTestCoversApplies pins "exists" to what a directive can actually
// resolve to, in both directions.
//
// Too strict and a real rule's directive is reported. Too loose and a directive that silences nothing
// stays silent, which is the defect the check exists for. The bare-name row is the loose mistake most
// likely to be made, since bareRuleName is right there: `consistency-no-enum` against a registered
// `nexus/consistency-no-enum` silences nothing (Covers needs the directive to be the more qualified
// side), so it must not count as existing.
func TestRuleNameCatalogExistsIsTheTestCoversApplies(t *testing.T) {
	t.Parallel()
	catalog := newRuleNameCatalog(
		[]string{"no-console", "nexus/consistency-no-enum", "@typescript-eslint/no-explicit-any"},
		[]string{"react-internal/not-ported-here"},
	)

	cases := []struct {
		name   string
		exists bool
	}{
		{"no-console", true},
		{"@typescript-eslint/no-explicit-any", true},
		{"nexus/consistency-no-enum", true},
		{"plugin/no-console", true},              // more qualified, which Covers resolves
		{"react-internal/not-ported-here", true}, // a config key proves the rule exists
		{"no-consol", false},
		{"consistency-no-enum", false},           // bare, which Covers does not resolve
		{"structure/consistency-no-enum", false}, // a stale prefix
		{"not-ported-here", false},
	}
	for _, testCase := range cases {
		if got := catalog.exists(testCase.name); got != testCase.exists {
			t.Errorf("exists(%q) = %v, want %v", testCase.name, got, testCase.exists)
		}
	}

	if got := catalog.registeredByBareName["consistency-no-enum"]; got != "nexus/consistency-no-enum" {
		t.Errorf("the suggestion for a stale prefix should name the registered rule, got %q", got)
	}
	if newRuleNameCatalog(nil, []string{"anything"}) != nil {
		t.Error("with no registry the catalog must be nil, or every name would read as unknown")
	}
}
