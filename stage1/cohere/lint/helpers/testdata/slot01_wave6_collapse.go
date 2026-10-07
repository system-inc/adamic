package tailwind

import (
	"fmt"
	"unicode/utf16"
)

var wave6Trace []string

func AdamicWave6Definition(source string, bound bool, mode string) ([][]any, []string, []string) {
	nodes := []*Node{
		{Kind: KindRule, Value: source, ValuePresent: true},
		{Kind: KindDeclaration, Value: source, ValuePresent: true},
		{Kind: KindComment, Value: source, ValuePresent: true},
		{Kind: KindDeclaration, Value: "calc(--modifier(--text --line-height) * 2)", ValuePresent: true},
		{Kind: KindDeclaration, Value: source, ValuePresent: false},
		{Kind: KindDeclaration, Value: "", ValuePresent: true},
	}
	nodes[0].Nodes = []*Node{nodes[1], nodes[2]}
	nodes[1].Nodes = []*Node{nodes[3]}
	nodes[2].Nodes = []*Node{nodes[4]}
	children := [][]int{{1, 2}, {3}, {4}, {}, {}, {}}
	rows := make([][]any, len(nodes))
	for index, node := range nodes {
		single := &Node{Kind: KindDeclaration, Value: node.Value, ValuePresent: true}
		normalizeValueFunctionArguments([]*Node{single})
		rows[index] = []any{string(node.Kind), node.ValuePresent, node.Value, single.Value, children[index]}
	}
	roots := []*Node{nodes[0], nodes[5]}
	normalizeValueFunctionArguments(roots)
	for index, node := range nodes {
		rows[index] = append(rows[index], node.Value)
		node.Value = rows[index][2].(string)
	}
	wave6Trace = nil
	if bound {
		normalizeUtilityDefinition(&UtilityDefinition{Nodes: roots})
	} else {
		normalizeUtilityDefinition(nil)
	}
	actual := make([]string, len(nodes))
	for index, node := range nodes {
		actual[index] = node.Value
	}
	return rows, actual, wave6Trace
}

func AdamicWave6Framework(source string, last int) ([]any, []string) {
	initial := []FrameworkVariantRegistration{{Name: "existing", Order: 3, Kind: ParsedVariantKindStatic}}
	incoming := []FrameworkVariantRegistration{{Name: source, Order: 2, Kind: ParsedVariantKindFunctional}, {Name: source, Order: 99, Kind: ParsedVariantKindCompound}, {Name: "shared-a", Order: 5, Kind: ParsedVariantKindStatic}, {Name: "shared-b", Order: 5, Kind: ParsedVariantKindArbitrary}, {Name: "negative", Order: -3, Kind: ParsedVariantKindFunctional}, {Name: "existing", Order: 1000, Kind: ParsedVariantKindFunctional}}
	queries := []string{"existing", source, "shared-a", "shared-b", "negative"}
	if source == "adamic:actual-framework-table" {
		incoming = append([]FrameworkVariantRegistration(nil), FrameworkVariantRegistrations...)
		queries = []string{"existing"}
		for _, row := range incoming {
			queries = append(queries, row.Name)
		}
	}
	registry := NewVariantRegistry()
	registry.lastOrder = last
	first := [][]any{}
	next := [][]any{}
	for _, row := range initial {
		registry.registrations[row.Name] = VariantRegistration{Name: row.Name, Order: row.Order, Kind: row.Kind}
		first = append(first, []any{row.Name, row.Order, string(row.Kind)})
	}
	for _, row := range incoming {
		next = append(next, []any{row.Name, row.Order, string(row.Kind)})
	}
	var observations []string
	units := func(value string) string {
		list := utf16.Encode([]rune(value))
		output := fmt.Sprintf("%d:", len(list))
		for _, unit := range list {
			output += fmt.Sprintf("%d,", unit)
		}
		return output
	}
	observe := func() {
		observations = append(observations, fmt.Sprintf("%d %d", registry.lastOrder, len(registry.registrations)))
		for _, name := range queries {
			row := registry.registrations[name]
			observations = append(observations, fmt.Sprint(row.Order), units(row.Name), units(string(row.Kind)))
		}
	}
	registry.RegisterFrameworkVariants(incoming)
	observe()
	for index := range incoming {
		incoming[index].Kind = "tampered"
	}
	observe()
	registry.RegisterFrameworkVariants(incoming)
	observe()
	return []any{last, first, next, queries}, observations
}
