package docsdata

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/system-inc/cohere/internal/lint/configuration"
)

// RuleSet is one rule set cohere carries, as a project that extends it alone sees it: the loader's own
// resolution, extends chain followed and plugin defaults applied, never a re-reading of the file.
type RuleSet struct {
	Name string `json:"name"`
	// Extends is what the set's own file names, in its order.
	Extends []string `json:"extends,omitempty"`
	// Plugins is every namespace the resolved chain declares.
	Plugins []string `json:"plugins,omitempty"`
	// RulesOn counts the resolved rules at warn or error, plugin defaults and the chain's included. It is
	// not CHANGELOG.md's per-set count, which counts the lines the set's own file turns on.
	RulesOn int `json:"rulesOn"`
	// Rules is every rule the resolved set names, off included, since an off is a ruling too.
	Rules     []SetRule     `json:"rules"`
	Overrides []SetOverride `json:"overrides,omitempty"`
}

// SetRule is one rule's resolved setting in a set.
type SetRule struct {
	Name     string            `json:"name"`
	Severity string            `json:"severity"`
	Options  []json.RawMessage `json:"options,omitempty"`
}

// SetOverride is one glob-scoped block in a set's resolved chain.
type SetOverride struct {
	// DefinedIn is the set in the chain that wrote the block.
	DefinedIn string    `json:"definedIn"`
	Files     []string  `json:"files"`
	Reason    string    `json:"reason,omitempty"`
	Rules     []SetRule `json:"rules"`
}

// buildSets resolves every set, in name order.
func buildSets(inputs Inputs) ([]RuleSet, error) {
	sets := make([]RuleSet, 0, len(inputs.SetNames))
	for _, name := range inputs.SetNames {
		contents, err := inputs.SetContents(name)
		if err != nil {
			return nil, err
		}
		extends, err := ownExtends(contents)
		if err != nil {
			return nil, fmt.Errorf("docsdata: %s: %w", name, err)
		}
		resolved, err := inputs.LoadSet(name)
		if err != nil {
			return nil, fmt.Errorf("docsdata: resolving %s: %w", name, err)
		}

		set := RuleSet{
			Name:    name,
			Extends: extends,
			Plugins: append([]string(nil), resolved.Plugins...),
			Rules:   setRules(resolved.Rules),
		}
		sort.Strings(set.Plugins)
		for _, setRule := range set.Rules {
			if setRule.Severity != configuration.SeverityOff.String() {
				set.RulesOn++
			}
		}
		for _, override := range resolved.Overrides {
			set.Overrides = append(set.Overrides, SetOverride{
				DefinedIn: override.File,
				Files:     override.Files,
				Reason:    override.Reason,
				Rules:     setRules(override.Rules),
			})
		}
		sets = append(sets, set)
	}
	return sets, nil
}

// setRules lists rule settings by name.
func setRules(settings map[string]configuration.RuleSetting) []SetRule {
	rules := make([]SetRule, 0, len(settings))
	for name, setting := range settings {
		rules = append(rules, SetRule{Name: name, Severity: setting.Severity.String(), Options: setting.Options})
	}
	sort.Slice(rules, func(left, right int) bool { return rules[left].Name < rules[right].Name })
	return rules
}

// ownExtends reads a set file's `extends`, one name or a list.
func ownExtends(contents []byte) ([]string, error) {
	var file struct {
		Extends json.RawMessage `json:"extends"`
	}
	if err := json.Unmarshal(contents, &file); err != nil {
		return nil, err
	}
	if len(file.Extends) == 0 {
		return nil, nil
	}
	var single string
	if err := json.Unmarshal(file.Extends, &single); err == nil {
		return []string{single}, nil
	}
	var many []string
	if err := json.Unmarshal(file.Extends, &many); err != nil {
		return nil, fmt.Errorf("extends is neither a name nor a list of names: %w", err)
	}
	return many, nil
}
