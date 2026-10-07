package tailwind

import (
	"fmt"
	"reflect"
	"slices"
	"unicode/utf16"
)

func AdamicWave11Case(source, mode string, variant int) ([]any, []string, bool) {
	if reflect.TypeOf(Node{}).NumField() != 10 {
		panic("Node field inventory changed")
	}
	leaf := &Node{Kind: KindDeclaration, Selector: "leaf", Name: "name", Params: "params", Property: "property", Value: source, ValuePresent: true, Important: true, Context: map[string]string{"scope": "leaf"}}
	branch := &Node{Kind: NodeKind("unknown"), Selector: "branch", Nodes: []*Node{leaf}, Context: map[string]string{}}
	root := &Node{Kind: []NodeKind{KindRule, KindAtRule, KindDeclaration, KindComment, KindContext, KindAtRoot, NodeKind("unknown")}[variant%7], Selector: "s:" + source, Name: "n:" + source, Params: "a:" + source, Property: "p:" + source, Value: "v:" + source, ValuePresent: variant%2 == 0, Important: variant%2 == 1, Nodes: []*Node{leaf, branch, leaf}}
	switch variant % 3 {
	case 1:
		root.Context = map[string]string{}
	case 2:
		root.Context = map[string]string{"fixture": source, "scope": "root"}
	}
	roots := make([]*Node, 2, 4)
	roots[0] = root
	roots[1] = root
	removed := map[*Node]bool{}
	switch variant {
	case 0:
		roots = nil
	case 1:
		roots = make([]*Node, 0, 2)
	case 3:
		if mode != "remove" {
			root.Nodes = append(root.Nodes, nil)
		}
	case 4:
		root.Nodes = make([]*Node, 0, 2)
	case 5:
		removed[leaf] = true
	case 6:
		removed[leaf] = true
		removed[branch] = true
	case 7:
		removed[leaf] = false
		removed[root] = false
	case 8:
		removed[root] = true
	case 9:
		removed[nil] = true
		roots[1] = nil
	}
	parsed := false
	if variant == 10 {
		if nodes, err := ParseCSS(source); err == nil {
			roots = nodes
			parsed = true
		}
	}
	if mode != "remove" {
		removed = nil
	}
	originals := []*Node{}
	ids := map[*Node]int{}
	var collect func([]*Node)
	collect = func(nodes []*Node) {
		for _, n := range nodes {
			if n == nil {
				continue
			}
			if _, ok := ids[n]; ok {
				continue
			}
			ids[n] = len(originals)
			originals = append(originals, n)
			collect(n.Nodes[:cap(n.Nodes)])
		}
	}
	collect(roots[:cap(roots)])
	keys := []string{"slot-probe"}
	for _, n := range originals {
		for k := range n.Context {
			if !slices.Contains(keys, k) {
				keys = append(keys, k)
			}
		}
	}
	slices.Sort(keys)
	id := func(n *Node) int {
		if n == nil {
			return -1
		}
		if i, ok := ids[n]; ok {
			return i
		}
		return -2
	}
	listRow := func(nodes []*Node) []any {
		values := []int{}
		for _, n := range nodes[:cap(nodes)] {
			values = append(values, id(n))
		}
		return []any{nodes == nil, len(nodes), cap(nodes), values}
	}
	arena := [][]any{}
	for _, n := range originals {
		context := [][]string{}
		for _, k := range keys {
			if v, ok := n.Context[k]; ok {
				context = append(context, []string{k, v})
			}
		}
		arena = append(arena, []any{string(n.Kind), n.Selector, n.Name, n.Params, n.Property, n.Value, n.ValuePresent, n.Important, n.Context != nil, context, listRow(n.Nodes)})
	}
	removeRows := [][]any{}
	for _, n := range originals {
		if v, ok := removed[n]; ok {
			removeRows = append(removeRows, []any{id(n), v})
		}
	}
	if v, ok := removed[nil]; ok {
		removeRows = append(removeRows, []any{-1, v})
	}
	chosen := (*Node)(nil)
	if len(roots) > 0 {
		chosen = roots[0]
	}
	input := []any{arena, listRow(roots), removeRows, id(chosen), keys}
	var out []*Node
	switch mode {
	case "node":
		out = []*Node{cloneNode(chosen)}
	case "list":
		out = cloneNodes(roots)
	default:
		out = removeNodes(roots, removed)
	}
	lines := []string{}
	add := func(value string) { lines = append(lines, value) }
	units := func(value string) string {
		codes := utf16.Encode([]rune(value))
		s := fmt.Sprintf("%d:", len(codes))
		for _, c := range codes {
			s += fmt.Sprintf("%d,", c)
		}
		return s
	}
	emit := func(nodes []*Node) {
		seen := map[*Node]int{}
		var emitList func([]*Node)
		var emitNode func(*Node)
		emitNode = func(n *Node) {
			if n == nil {
				add("nil")
				return
			}
			if k, ok := seen[n]; ok {
				add(fmt.Sprintf("ref:%d", k))
				return
			}
			k := len(seen)
			seen[n] = k
			add(fmt.Sprintf("node:%d:%d", k, id(n)))
			for _, s := range []string{string(n.Kind), n.Selector, n.Name, n.Params, n.Property, n.Value} {
				add(units(s))
			}
			add(fmt.Sprintf("%t:%t:%t", n.ValuePresent, n.Important, n.Context != nil))
			count := 0
			for _, key := range keys {
				if _, ok := n.Context[key]; ok {
					count++
				}
			}
			add(fmt.Sprint(count))
			for _, key := range keys {
				if v, ok := n.Context[key]; ok {
					add(units(key))
					add(units(v))
				}
			}
			emitList(n.Nodes)
		}
		emitList = func(list []*Node) {
			add(fmt.Sprintf("%t:%d:%d", list == nil, len(list), cap(list)))
			for _, n := range list[:cap(list)] {
				add(fmt.Sprint(id(n)))
			}
			for _, n := range list {
				emitNode(n)
			}
		}
		emitList(nodes)
	}
	sourceList := make([]*Node, len(originals))
	copy(sourceList, originals)
	emit(out)
	emit(sourceList)
	if len(out) > 0 && out[0] != nil {
		n := out[0]
		n.Value = "changed"
		if n.Context != nil {
			n.Context["slot-probe"] = "written"
		}
		if len(n.Nodes) > 0 && n.Nodes[0] != nil {
			n.Nodes[0].Selector = "child-changed"
		}
	}
	emit(sourceList)
	if chosen != nil && chosen.Context != nil {
		chosen.Context["slot-probe"] = "source-write"
	}
	if len(roots) > 0 {
		roots[0] = nil
	}
	emit(out)
	return input, lines, parsed
}
