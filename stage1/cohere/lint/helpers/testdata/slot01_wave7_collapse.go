package tailwind

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf16"
)

func wave7Units(value string) string {
	units := utf16.Encode([]rune(value))
	out := fmt.Sprintf("%d:", len(units))
	for _, unit := range units {
		out += fmt.Sprintf("%d,", unit)
	}
	return out
}

var wave7Actions []int
var wave7Adds []string

func wave7ThemeAdd(theme *Theme, key, value string, options ThemeOptions) error {
	wave7Adds = append(wave7Adds, fmt.Sprintf("%s %s %d", wave7Units(key), wave7Units(value), options))
	return theme.Add(key, value, options)
}
func AdamicWave7Theme(source string) ([]any, []string) {
	params := source
	if source == "foo" {
		params = "inline prefix(ok)"
	}
	if source == "foo-*" {
		params = "prefix(BAD)"
	}
	if source == "*" {
		params = "prefix()"
	}
	options, prefix := parseThemeOptions(params)
	parsed, _ := ParseCSS(source)
	wildcardValue := source
	if source == "initial" || source == "not-custom" {
		wildcardValue = "initial"
	}
	nodes := []*Node{{Kind: KindComment}, {Kind: KindAtRule, Name: "@keyframes", Nodes: []*Node{{Kind: KindDeclaration, Property: "color", Value: "red"}}}, {Kind: KindRule, Nodes: []*Node{{Kind: KindDeclaration, Property: `--color-\61`, Value: source}}}, {Kind: KindDeclaration, Property: "--color-*", Value: wildcardValue}}
	nodes = append(nodes, parsed...)
	nodes = append(nodes, &Node{Kind: KindDeclaration, Property: source, Value: "unexpected"}, &Node{Kind: KindDeclaration, Property: "--after", Value: "later"})
	rows := [][]any{}
	var flatten func(*Node) int
	flatten = func(node *Node) int {
		index := len(rows)
		rows = append(rows, nil)
		children := []int{}
		for _, child := range node.Nodes {
			children = append(children, flatten(child))
		}
		key := unescapeCSSIdentifier(node.Property)
		err := NewTheme().Add(key, node.Value, options)
		message := ""
		if err != nil {
			message = err.Error()
		}
		rows[index] = []any{string(node.Kind), node.Name, node.Property, node.Value, children, key, message, fmt.Sprintf("%q", node.Property)}
		return index
	}
	roots := []int{}
	for _, node := range nodes {
		roots = append(roots, flatten(node))
	}
	theme := NewTheme()
	theme.Prefix = "old"
	c := &stylesheetCollector{theme: theme}
	wave7Actions = nil
	wave7Adds = nil
	err := c.ingestThemeBlock(&Node{Params: params, Nodes: nodes}, "fixture.css")
	message := ""
	if err != nil {
		message = err.Error()
	}
	actions := []string{}
	for _, a := range wave7Actions {
		actions = append(actions, fmt.Sprint(a))
	}
	actual := []string{wave7Units(message), wave7Units(theme.Prefix), strings.Join(actions, ",")}
	actual = append(actual, wave7Adds...)
	return []any{params, int(options), prefix, isValidThemePrefix(prefix), fmt.Sprintf("%q", prefix), rows, roots}, actual
}

var wave7TableTrace []string
var wave7Readings map[string]*Reading

func wave7StaticReading(name string) (Reading, bool) {
	r, ok := FrameworkStaticReading(name)
	copy := r
	wave7Readings[name] = &copy
	return r, ok
}
func wave7ReadingRow(name string, r Reading) []any {
	return []any{name, r.Order, len(r.Order), cap(r.Order), r.Count}
}
func AdamicWave7TableMetadata() []any {
	roots := []string{}
	for name := range baseDescriptors {
		roots = append(roots, name)
	}
	sort.Strings(roots)
	names := []string{}
	for name := range FrameworkStaticDeclarations {
		names = append(names, name)
	}
	sort.Strings(names)
	statics := [][]any{}
	for _, name := range names {
		r, ok := FrameworkStaticReading(name)
		if !ok {
			panic("missing static reading")
		}
		statics = append(statics, wave7ReadingRow(name, r))
	}
	properties := []string{}
	for name := range PropertyOrder {
		properties = append(properties, name)
	}
	sort.Strings(properties)
	rows := [][]any{}
	for _, name := range properties {
		rows = append(rows, []any{name, PropertyOrder[name]})
	}
	return []any{"metadata", roots, statics, rows}
}
func AdamicWave7Table(source string, bound bool) ([]any, []string) {
	theme := NewTheme()
	_ = theme.Add("--color-wave", source, ThemeOptionInline)
	_ = theme.Add("--text-sm", "1rem", ThemeOptionNone)
	system := &LoadedDesignSystem{TailwindVersion: "test-version", theme: theme, utilityRoots: map[string]map[UtilityKind]bool{source: {UtilityKindFunctional: true}}, staticUtilityNodes: map[string][]*Node{source: {{Kind: KindDeclaration, Property: "color", Value: source, ValuePresent: true}}}}
	scratch := &Table{Descriptors: map[string]*Descriptor{}, Statics: map[string]Reading{}}
	scratch.addThemeNamespaces(theme)
	scratch.addRepositoryStatics(system)
	scratch.addRepositoryFunctionalRoots(system)
	keys := [][]any{}
	for _, name := range scratch.Namespaces {
		values := []string{}
		for key := range scratch.KeysByNamespace[name] {
			values = append(values, key)
		}
		sort.Strings(values)
		keys = append(keys, []any{name, values})
	}
	repo := [][]any{}
	for name, r := range scratch.Statics {
		repo = append(repo, wave7ReadingRow(name, r))
	}
	roots := []string{}
	for name := range scratch.Descriptors {
		roots = append(roots, name)
	}
	sort.Strings(roots)
	wave7TableTrace = nil
	wave7Readings = map[string]*Reading{}
	var table *Table
	if bound {
		table = NewTable(system)
	} else {
		table = NewTable(nil)
	}
	metadata := AdamicWave7TableMetadata()
	baseNames := metadata[1].([]string)
	first := baseNames[0]
	queryNames := []string{first, "block", source}
	var actual []string
	observe := func() {
		if table == nil {
			actual = append(actual, wave7Units(""), "false 0 0 0 0", "", "", "-1", "-1", "-1", "-1", "-1", "-1", "-1")
			return
		}
		actual = append(actual, wave7Units(table.TailwindVersion), fmt.Sprintf("true %d %d %d 7", len(table.Descriptors), len(table.Statics), len(table.PropertyOrder)), strings.Join(wave7TableTrace, ","), strings.Join(table.Namespaces, ","))
		for _, name := range queryNames {
			if d := table.Descriptors[name]; d != nil {
				actual = append(actual, wave7Units(d.Root)+fmt.Sprintf(" %t", d.PerDeclaration))
			} else {
				actual = append(actual, "-1")
			}
			if r, ok := table.Statics[name]; ok {
				out := fmt.Sprintf("%d %d %d", r.Count, len(r.Order), cap(r.Order))
				for _, order := range r.Order {
					out += fmt.Sprintf(" %d", order)
				}
				actual = append(actual, out)
			} else {
				actual = append(actual, "-1")
			}
		}
		actual = append(actual, fmt.Sprint(table.PropertyOrder["color"]))
	}
	observe()
	if bound {
		originalDescriptor := baseDescriptors[first]
		delete(baseDescriptors, first)
		originalProperty := PropertyOrder["color"]
		PropertyOrder["color"] = 99999
		reading := wave7Readings["block"]
		if reading != nil {
			if len(reading.Order) > 0 {
				reading.Order[0] = 99999
			}
			reading.Order = reading.Order[:0:0]
			reading.Count = 99999
		}
		observe()
		baseDescriptors[first] = originalDescriptor
		PropertyOrder["color"] = originalProperty
	} else {
		observe()
	}
	return []any{source, bound, scratch.Namespaces, keys, repo, roots, first}, actual
}
