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
