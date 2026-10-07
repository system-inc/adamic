package tailwind

import (
	"fmt"
	"sort"
	"strings"
)

type AdamicMembership struct {
	Namespace string   `json:"namespace"`
	Keys      []string `json:"keys"`
}
type AdamicBatch7Case struct {
	Mode       string             `json:"mode"`
	Source     string             `json:"source"`
	Prefix     string             `json:"prefix"`
	Entries    []string           `json:"entries"`
	Membership []AdamicMembership `json:"membership"`
	Normalized []string           `json:"normalized"`
}

func AdamicBatch7(sample *AdamicBatch7Case) string {
	var out strings.Builder
	print := func(x any) { fmt.Fprintln(&out, x) }
	switch sample.Mode {
	case "roots":
		sample.Source = "root-" + sample.Source
		old := &Descriptor{Root: "original", TypeList: []DataType{"color"}, PerDeclaration: false}
		table := &Table{Descriptors: map[string]*Descriptor{sample.Source: old, "untouched": old}}
		alias := table.Descriptors
		system := &LoadedDesignSystem{utilityRoots: map[string]map[UtilityKind]bool{sample.Source: {UtilityKindFunctional: true}, "static-only": {UtilityKindStatic: true}, "disabled": {UtilityKindFunctional: false}}}
		table.addRepositoryFunctionalRoots(system)
		first := table.Descriptors[sample.Source]
		print(alias[sample.Source] == first)
		print(first == old)
		print(table.Descriptors["untouched"] == old)
		print(first.Root)
		print(first.PerDeclaration)
		print(first.TypeList == nil)
		print(first.ByLiteral == nil)
		for _, axis := range []AxisReadings{first.Absent, first.Alpha, first.Themed} {
			print(axis.ByType == nil)
			print(axis.ByNamespace == nil)
			print(len(axis.Fallback.Order))
			print(axis.Fallback.Count)
			print(axis.Empty == nil)
		}
		_, hasStatic := table.Descriptors["static-only"]
		_, hasDisabled := table.Descriptors["disabled"]
		print(hasStatic)
		print(hasDisabled)
		table.addRepositoryFunctionalRoots(system)
		print(first == table.Descriptors[sample.Source])
		print(old.Root)
		print(len(old.TypeList))
	case "namespace":
		theme := NewTheme()
		theme.Prefix = sample.Prefix
		for _, key := range []string{"--a-x", "--a-long-y", "--z-x", "--é-x", "--😀-x", "--aaaa-x", "--a--x", "--a-x--nested", "--font-weight-bold", "--font-size-sm", "--font-size-small", "--a-" + sample.Source} {
			if err := theme.Add(key, "value", 0); err != nil {
				panic(err)
			}
		}
		sample.Entries = []string{}
		sample.Membership = []AdamicMembership{}
		candidates := map[string]bool{}
		for _, entry := range theme.Entries() {
			sample.Entries = append(sample.Entries, entry.Key)
			segments := splitThemeKey(entry.Key)
			for n := 1; n < len(segments); n++ {
				candidates["--"+joinSegments(segments[:n])] = true
			}
		}
		sorted := []string{}
		for key := range candidates {
			sorted = append(sorted, key)
		}
		sort.Strings(sorted)
		for _, key := range sorted {
			sample.Membership = append(sample.Membership, AdamicMembership{key, theme.KeysInNamespaces([]string{key})})
		}
		oldList := []string{"stale"}
		oldMap := map[string]map[string]bool{"stale": {"old": true}}
		table := &Table{Namespaces: oldList, KeysByNamespace: oldMap}
		table.addThemeNamespaces(theme)
		print(len(table.Namespaces))
		for _, ns := range table.Namespaces {
			print(ns)
			keys := []string{}
			for k := range table.KeysByNamespace[ns] {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			print(len(keys))
			for _, k := range keys {
				print(k)
			}
		}
		table.KeysByNamespace["mutated"] = map[string]bool{}
		print(len(oldMap))
		print(oldList[0])
	case "evaluator":
		sample.Source = "root-" + sample.Source
		makeDef := func(name, value string) *UtilityDefinition {
			return &UtilityDefinition{Name: name, Nodes: []*Node{{Kind: KindDeclaration, Property: "--test", Value: value, ValuePresent: true}}}
		}
		defs := []*UtilityDefinition{makeDef(sample.Source, "--value(--spacing)"), makeDef("other", sample.Source), makeDef(sample.Source, "--modifier( --color )"), makeDef("", "--value( --spacing-\\* )")}
		theme := NewTheme()
		evaluator := NewUtilityEvaluator(theme, defs)
		sample.Normalized = []string{}
		for _, d := range defs {
			sample.Normalized = append(sample.Normalized, d.Nodes[0].Value)
			print(d.Nodes[0].Value)
		}
		print(evaluator.Theme == theme)
		print(evaluator.Definitions[sample.Source] == defs[2])
		print(evaluator.Definitions["other"] == defs[1])
		print(len(evaluator.Definitions))
		defs[2].Nodes[0].Value = "alias mutation"
		print(evaluator.Definitions[sample.Source].Nodes[0].Value)
		evaluator.Definitions["new"] = defs[0]
		second := NewUtilityEvaluator(theme, []*UtilityDefinition{})
		print(len(second.Definitions))
		print(second.Theme == theme)
		nilTheme := NewUtilityEvaluator(nil, []*UtilityDefinition{})
		print(nilTheme.Theme == nil)
	default:
		panic("unsupported mode")
	}
	return out.String()
}
