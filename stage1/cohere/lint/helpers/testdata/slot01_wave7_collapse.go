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
func AdamicWave7Utility(source string) ([]any, []string) {
	c := &stylesheetCollector{utilityRoots: map[string]map[UtilityKind]bool{}, staticUtilityNodes: map[string][]*Node{}}
	names := []string{source, source + "-*", source, "shared", "shared-*", "shared", "-*", "foo*bar"}
	rows := [][]any{}
	queries := []string{}
	set := map[string]bool{}
	for _, name := range names {
		trim := strings.TrimSpace(name)
		root := strings.TrimSuffix(trim, "-*")
		if root != "" && !strings.Contains(root, "*") && !set[root] {
			set[root] = true
			queries = append(queries, root)
		}
	}
	sort.Strings(queries)
	var actual []string
	observe := func(err string) {
		actual = append(actual, wave7Units(err), fmt.Sprintf("%d %d", len(c.utilityRoots), len(c.utilityDefinitions)))
		for _, q := range queries {
			k := 0
			if c.utilityRoots[q][UtilityKindStatic] {
				k |= 1
			}
			if c.utilityRoots[q][UtilityKindFunctional] {
				k |= 2
			}
			body := c.staticUtilityNodes[q]
			out := fmt.Sprintf("%d %d", k, len(body))
			for _, n := range body {
				out += " " + n.Property
			}
			actual = append(actual, out)
		}
		for _, d := range c.utilityDefinitions {
			out := wave7Units(d.Name)
			for _, n := range d.Nodes {
				out += " " + n.Property
			}
			actual = append(actual, out)
		}
	}
	nodes := []*Node{}
	for index, name := range names {
		body := []*Node{{Property: fmt.Sprint(index)}}
		node := &Node{Params: name, Nodes: body}
		nodes = append(nodes, node)
		rows = append(rows, []any{name, strings.TrimSpace(name), []int{index}})
		err := c.ingestUtilityBlock(node, "fixture.css")
		message := ""
		if err != nil {
			message = err.Error()
		}
		observe(message)
	}
	for _, node := range nodes {
		node.Nodes[0] = &Node{Property: "99"}
	}
	observe("")
	return []any{rows, queries}, actual
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
