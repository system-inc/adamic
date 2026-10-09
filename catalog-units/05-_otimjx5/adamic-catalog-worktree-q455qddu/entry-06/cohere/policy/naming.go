// Package policy is the house policy both engines read as data: what the Go engine embeds here, the Swift
// engine reads from the same file, so a decision about the house lands once and cannot drift between them.
package policy

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed RuleNaming.json
var ruleNamingFile []byte

// RuleNamingWord is one entry in a closed list, with the sentence that says what belongs under it.
type RuleNamingWord struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

// RuleNaming is RuleNaming.json: the closed lists a house rule's name is built from.
type RuleNaming struct {
	About      string           `json:"about"`
	Verbs      []RuleNamingWord `json:"verbs"`
	Categories []RuleNamingWord `json:"categories"`

	// Namespaces are the namespaces whose rules are named by a closed list rather than by the scheme:
	// `adamic/invariant-mutable` keeps Adamic's own name, so one rule has one name across both projects,
	// and the guard still covers it because a name outside the list is refused.
	Namespaces []RuleNamingNamespace `json:"namespaces"`
}

// RuleNamingNamespace is one namespace named by a closed list, and the sentence saying what belongs in it.
type RuleNamingNamespace struct {
	Name       string           `json:"name"`
	Definition string           `json:"definition"`
	Names      []RuleNamingWord `json:"names"`
}

// HouseRuleName is a house rule's name after its namespace, split into the scheme's three parts.
type HouseRuleName struct {
	Category string
	Verb     string
	Object   string
}

// Naming is the embedded file, read once. An invalid file panics at startup: a scheme read wrongly would
// let a misnamed rule through while every check reported the names clean.
var Naming = mustLoadRuleNaming(ruleNamingFile)

func mustLoadRuleNaming(data []byte) RuleNaming {
	loaded, err := loadRuleNaming(data)
	if err != nil {
		panic(fmt.Sprintf("RuleNaming.json: %v", err))
	}
	return loaded
}

// loadRuleNaming reads the file, refusing an unknown key, an empty list, a duplicate, an entry with no
// definition, and a word that is not lowercase kebab case, since each would read as something other
// than meant.
func loadRuleNaming(data []byte) (RuleNaming, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var naming RuleNaming
	if err := decoder.Decode(&naming); err != nil {
		return RuleNaming{}, err
	}
	lists := []struct {
		label string
		words []RuleNamingWord
	}{{"verbs", naming.Verbs}, {"categories", naming.Categories}}
	namespaces := make([]RuleNamingWord, 0, len(naming.Namespaces))
	for _, namespace := range naming.Namespaces {
		namespaces = append(namespaces, RuleNamingWord{Name: namespace.Name, Definition: namespace.Definition})
		lists = append(lists, struct {
			label string
			words []RuleNamingWord
		}{"namespaces: " + namespace.Name, namespace.Names})
	}
	// Checked as a list of its own, so a namespace named twice is refused the way a duplicate verb is. An
	// empty list is fine: a file with no closed namespace names every rule by the scheme.
	if len(namespaces) > 0 {
		lists = append(lists, struct {
			label string
			words []RuleNamingWord
		}{"namespaces", namespaces})
	}
	for _, list := range lists {
		if len(list.words) == 0 {
			return RuleNaming{}, fmt.Errorf("%s is empty, so no name could fit", list.label)
		}
		seen := map[string]bool{}
		for _, word := range list.words {
			if !isKebabWord(word.Name) {
				return RuleNaming{}, fmt.Errorf("%s: %q is not lowercase kebab case", list.label, word.Name)
			}
			if seen[word.Name] {
				return RuleNaming{}, fmt.Errorf("%s: %q appears twice", list.label, word.Name)
			}
			seen[word.Name] = true
			if strings.TrimSpace(word.Definition) == "" {
				return RuleNaming{}, fmt.Errorf("%s: %q has no definition", list.label, word.Name)
			}
		}
	}
	return naming, nil
}

// ParseHouseRuleName splits a house rule's name after its namespace (`consistency-no-print`, not
// `cohere-swift/consistency-no-print`) into category, verb and object, or says why it does not fit.
//
// Categories are tried longest first, so `react-hook-no-any-type` reads as `react-hook`, and a category
// that is a prefix of another can never claim the longer one's names.
func (naming RuleNaming) ParseHouseRuleName(leaf string) (HouseRuleName, error) {
	categories := make([]string, 0, len(naming.Categories))
	for _, category := range naming.Categories {
		categories = append(categories, category.Name)
	}
	sort.Slice(categories, func(left, right int) bool { return len(categories[left]) > len(categories[right]) })

	for _, category := range categories {
		rest, found := strings.CutPrefix(leaf, category+"-")
		if !found {
			continue
		}
		for _, verb := range naming.Verbs {
			object, found := strings.CutPrefix(rest, verb.Name+"-")
			if found && isKebabWord(object) {
				return HouseRuleName{Category: category, Verb: verb.Name, Object: object}, nil
			}
		}
		return HouseRuleName{}, fmt.Errorf("%q has the category %q, and what follows is not a verb from the closed list and an object", leaf, category)
	}
	return HouseRuleName{}, fmt.Errorf("%q does not begin with a category from the closed list", leaf)
}

// ParseRuleName judges a house rule's name in its namespace: against the closed list when the namespace
// has one, and by the scheme otherwise. A closed name reads as the namespace for its category and the
// whole name for its object, since it was never built from a verb.
func (naming RuleNaming) ParseRuleName(namespace string, leaf string) (HouseRuleName, error) {
	for _, closed := range naming.Namespaces {
		if closed.Name != namespace {
			continue
		}
		for _, name := range closed.Names {
			if name.Name == leaf {
				return HouseRuleName{Category: closed.Name, Object: leaf}, nil
			}
		}
		return HouseRuleName{}, fmt.Errorf("%q is not in the %s namespace's closed list in RuleNaming.json", leaf, closed.Name)
	}
	return naming.ParseHouseRuleName(leaf)
}

// isKebabWord is one or more lowercase words of letters and digits joined by single dashes.
func isKebabWord(text string) bool {
	if text == "" {
		return false
	}
	for _, word := range strings.Split(text, "-") {
		if word == "" {
			return false
		}
		for _, character := range word {
			if (character < 'a' || character > 'z') && (character < '0' || character > '9') {
				return false
			}
		}
	}
	return true
}
