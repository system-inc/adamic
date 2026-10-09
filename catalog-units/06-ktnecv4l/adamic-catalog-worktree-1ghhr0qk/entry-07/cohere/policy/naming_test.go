package policy

import (
	"strings"
	"testing"
)

// minimalRuleNaming is a file every refusal below starts from, changed in one place, so each refusal is
// the one under test rather than some other.
const minimalRuleNaming = `{"about": "a", "verbs": [{"name": "no", "definition": "d"}], "categories": [{"name": "consistency", "definition": "d"}]}`

// TestTheLoaderRefusesWhatWouldReadAsSomethingElse: the minimal file loads, and each one-place change that
// would misread does not.
func TestTheLoaderRefusesWhatWouldReadAsSomethingElse(t *testing.T) {
	t.Parallel()
	if _, err := loadRuleNaming([]byte(minimalRuleNaming)); err != nil {
		t.Fatalf("the minimal file is refused (%v), so no refusal below would mean anything", err)
	}
	for _, refused := range []struct {
		name string
		old  string
		new  string
	}{
		{"an unknown key", `"about": "a"`, `"about": "a", "aliases": []`},
		{"an empty verb list", `[{"name": "no", "definition": "d"}]`, `[]`},
		{"an empty category list", `[{"name": "consistency", "definition": "d"}]`, `[]`},
		{"a duplicate", `{"name": "no", "definition": "d"}`, `{"name": "no", "definition": "d"}, {"name": "no", "definition": "e"}`},
		{"a word with no definition", `{"name": "no", "definition": "d"}`, `{"name": "no", "definition": " "}`},
		{"a word that is not kebab case", `"name": "consistency"`, `"name": "Consistency"`},
		{"a word with an empty segment", `"name": "consistency"`, `"name": "consistency-"`},
	} {
		if !strings.Contains(minimalRuleNaming, refused.old) {
			t.Fatalf("%s: the anchor %q is not in the minimal file, so the change would not apply", refused.name, refused.old)
		}
		if _, err := loadRuleNaming([]byte(strings.Replace(minimalRuleNaming, refused.old, refused.new, 1))); err == nil {
			t.Errorf("%s loaded", refused.name)
		}
	}
}

// TestParseHouseRuleNameSplitsTheThreeParts: names that fit split into category, verb and object, those
// that don't are refused, each for its own reason.
func TestParseHouseRuleNameSplitsTheThreeParts(t *testing.T) {
	t.Parallel()
	for leaf, want := range map[string]HouseRuleName{
		"consistency-no-print":                    {"consistency", "no", "print"},
		"react-hook-no-any-type":                  {"react-hook", "no", "any-type"},
		"toolchain-require-swift-6-language-mode": {"toolchain", "require", "swift-6-language-mode"},
	} {
		got, err := Naming.ParseHouseRuleName(leaf)
		if err != nil || got != want {
			t.Errorf("%s: %+v, %v; want %+v", leaf, got, err, want)
		}
	}
	for leaf, because := range map[string]string{
		"force-unwrapping":             "does not begin with a category",
		"consistency-organize-imports": "is not a verb",
		"consistency-no":               "is not a verb",
		"consistency-no-":              "is not a verb",
		"consistency-no-Print":         "is not a verb",
		"consistency":                  "does not begin with a category",
	} {
		if _, err := Naming.ParseHouseRuleName(leaf); err == nil || !strings.Contains(err.Error(), because) {
			t.Errorf("%s: %v; want an error saying it %s", leaf, err, because)
		}
	}
}

// TestALongerCategoryWinsOverItsPrefix: with `react` listed before `react-hook`, a `react-hook-` name still
// reads as `react-hook`. In file order it would read as `react` and fail on the verb `hook`, so the case
// tells longest-first from file order, which a list with no shared prefix cannot.
func TestALongerCategoryWinsOverItsPrefix(t *testing.T) {
	t.Parallel()
	naming := RuleNaming{
		Verbs:      []RuleNamingWord{{Name: "no", Definition: "d"}},
		Categories: []RuleNamingWord{{Name: "react", Definition: "d"}, {Name: "react-hook", Definition: "d"}},
	}
	if got, err := naming.ParseHouseRuleName("react-no-anchor"); err != nil || got.Category != "react" {
		t.Fatalf("the shorter category's own name: %+v, %v", got, err)
	}
	if got, err := naming.ParseHouseRuleName("react-hook-no-any-type"); err != nil || got.Category != "react-hook" {
		t.Errorf("react-hook-no-any-type: %+v, %v; want the category react-hook", got, err)
	}
}

// TestAClosedNamespaceAcceptsOnlyItsList: an adamic name on the list reads as the namespace's, one off the
// list is refused by name, and a namespace with no closed list still reads by the scheme. The refusal is
// the guard: without it a misspelled adamic rule would publish because nothing judged it.
func TestAClosedNamespaceAcceptsOnlyItsList(t *testing.T) {
	t.Parallel()
	if got, err := Naming.ParseRuleName("adamic", "invariant-mutable"); err != nil || got.Category != "adamic" || got.Object != "invariant-mutable" {
		t.Errorf("adamic/invariant-mutable: %+v, %v; want the namespace as its category and the name as its object", got, err)
	}
	if _, err := Naming.ParseRuleName("adamic", "invariant-mutables"); err == nil || !strings.Contains(err.Error(), "closed list") {
		t.Errorf("adamic/invariant-mutables: %v; want a refusal naming the closed list", err)
	}
	if _, err := Naming.ParseRuleName("adamic", "correctness-no-any"); err == nil {
		t.Error("adamic/correctness-no-any fits the scheme but not the list, and loaded: the list must win in its namespace")
	}
	if got, err := Naming.ParseRuleName("nexus", "consistency-no-print"); err != nil || got.Verb != "no" {
		t.Errorf("nexus/consistency-no-print: %+v, %v; want the scheme's reading", got, err)
	}
}

// TestTheLoaderRefusesAMisreadNamespace: each one-place change to a closed namespace that would misread is
// refused, as the verbs' and categories' are.
func TestTheLoaderRefusesAMisreadNamespace(t *testing.T) {
	t.Parallel()
	const withNamespace = `{"about": "a", "verbs": [{"name": "no", "definition": "d"}], "categories": [{"name": "consistency", "definition": "d"}], "namespaces": [{"name": "adamic", "definition": "d", "names": [{"name": "single-spread", "definition": "d"}]}]}`
	if _, err := loadRuleNaming([]byte(withNamespace)); err != nil {
		t.Fatalf("a file with one closed namespace is refused (%v), so no refusal below would mean anything", err)
	}
	for _, refused := range []struct {
		name string
		old  string
		new  string
	}{
		{"an empty name list", `[{"name": "single-spread", "definition": "d"}]`, `[]`},
		{"a duplicate name", `{"name": "single-spread", "definition": "d"}`, `{"name": "single-spread", "definition": "d"}, {"name": "single-spread", "definition": "e"}`},
		{"a name with no definition", `{"name": "single-spread", "definition": "d"}`, `{"name": "single-spread", "definition": ""}`},
		{"a name that is not kebab case", `"name": "single-spread"`, `"name": "singleSpread"`},
		{"a namespace with no definition", `"name": "adamic", "definition": "d"`, `"name": "adamic", "definition": ""`},
		{"a namespace named twice", `"namespaces": [{"name": "adamic", "definition": "d", "names": [{"name": "single-spread", "definition": "d"}]}]`, `"namespaces": [{"name": "adamic", "definition": "d", "names": [{"name": "single-spread", "definition": "d"}]}, {"name": "adamic", "definition": "d", "names": [{"name": "nominal-class", "definition": "d"}]}]`},
	} {
		if !strings.Contains(withNamespace, refused.old) {
			t.Fatalf("%s: the anchor %q is not in the file, so the change would not apply", refused.name, refused.old)
		}
		if _, err := loadRuleNaming([]byte(strings.Replace(withNamespace, refused.old, refused.new, 1))); err == nil {
			t.Errorf("%s loaded", refused.name)
		}
	}
}
