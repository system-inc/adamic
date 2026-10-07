package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

type AdamicEntry struct {
	Key, Value string
	Options    int
}
type AdamicCss struct {
	Kind, Value string
	Present     bool
	Edges       []int
	Values      []AdamicValue
	Roots       []int
}
type AdamicValue struct {
	Kind, Value, Prefix, Suffix string
	Edges                       []int
	Resolution                  AdamicResolution
}
type AdamicResolution struct {
	Text      string
	Ratio, OK bool
}
type AdamicCase struct {
	Name, Source, Helper, Prefix string
	Value                        *ParsedValue
	Modifier                     *ParsedModifier
	Entries                      []AdamicEntry
	Nodes                        []AdamicCss
	Roots                        []int
	Flags                        int
	SeedSets                     bool
}

func adamicState(c AdamicCase, nodes []*Node) *utilityEvaluation {
	t := NewTheme()
	t.Prefix = c.Prefix
	for _, e := range c.Entries {
		t.values[e.Key] = themeValue{value: e.Value, options: ThemeOptions(e.Options)}
	}
	s := &utilityEvaluation{evaluator: &UtilityEvaluator{Theme: t}, value: c.Value, modifier: c.Modifier, usedValueFunction: c.Flags&1 != 0, resolvedValueFunction: c.Flags&2 != 0, usedModifierFunction: c.Flags&4 != 0, resolvedModifierFunction: c.Flags&8 != 0, resolvedRatioValue: c.Flags&16 != 0}
	if c.SeedSets {
		s.nonRatioDeclarations = map[*Node]bool{}
		s.droppedDeclarations = map[*Node]bool{}
		if len(nodes) > 0 {
			s.nonRatioDeclarations[nodes[0]] = false
			s.droppedDeclarations[nodes[0]] = false
		}
	}
	return s
}
func adamicNodes(c AdamicCase) ([]*Node, []*Node) {
	nodes := make([]*Node, len(c.Nodes))
	for i, n := range c.Nodes {
		nodes[i] = &Node{Kind: NodeKind(n.Kind), Value: n.Value, ValuePresent: n.Present}
	}
	for i, n := range c.Nodes {
		for _, edge := range n.Edges {
			nodes[i].Nodes = append(nodes[i].Nodes, nodes[edge])
		}
	}
	roots := []*Node{}
	for _, id := range c.Roots {
		roots = append(roots, nodes[id])
	}
	return nodes, roots
}
func adamicValueNodes(nodes []ValueNode, s *utilityEvaluation, arena *[]AdamicValue) []int {
	ids := []int{}
	for _, node := range nodes {
		id := len(*arena)
		ids = append(ids, id)
		v := AdamicValue{Kind: string(node.Kind), Value: node.Value, Edges: []int{}, Prefix: ValueToCss([]ValueNode{node})}
		if node.Kind == ValueNodeKindFunction {
			v.Prefix = node.Value + "("
			v.Suffix = ")"
			var val *ParsedValue
			if node.Value == "--modifier" {
				val = s.modifierAsValue()
			} else {
				val = s.value
			}
			resolved, ratio, ok := s.resolveValueFunction(val, &node)
			v.Resolution = AdamicResolution{ValueToCss(resolved), ratio, ok}
		}
		*arena = append(*arena, v)
		edges := adamicValueNodes(node.Nodes, s, arena)
		(*arena)[id].Edges = edges
	}
	return ids
}
func AdamicAdapt(c AdamicCase) AdamicCase {
	if c.Nodes == nil {
		c.Nodes = []AdamicCss{{Kind: "declaration", Value: c.Source, Present: true, Edges: []int{}}}
		c.Roots = []int{0}
	}
	s := adamicState(c, nil)
	for i := range c.Nodes {
		c.Nodes[i].Values = []AdamicValue{}
		c.Nodes[i].Roots = adamicValueNodes(ParseValue(c.Nodes[i].Value), s, &c.Nodes[i].Values)
		if c.Nodes[i].Edges == nil {
			c.Nodes[i].Edges = []int{}
		}
	}
	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase { return []AdamicCase{AdamicAdapt(c)} }
func AdamicObserve(c AdamicCase) {
	for mode := 0; mode < 3; mode++ {
		nodes, roots := adamicNodes(c)
		s := adamicState(c, nodes)
		direct := make([]bool, len(nodes))
		if mode == 0 {
			s.walk(roots)
		} else {
			for id, node := range nodes {
				if mode == 1 {
					s.resolveDeclaration(node)
				} else {
					parsed := ParseValue(node.Value)
					direct[id] = s.resolveValueFunctions(parsed, node)
					node.Value = ValueToCss(parsed)
				}
			}
		}
		fmt.Printf("%t|%t|%t|%t|%t|%t|%t\n", s.usedValueFunction, s.resolvedValueFunction, s.usedModifierFunction, s.resolvedModifierFunction, s.resolvedRatioValue, s.nonRatioDeclarations != nil, s.droppedDeclarations != nil)
		for id, node := range nodes {
			a, ap := s.nonRatioDeclarations[node]
			b, bp := s.droppedDeclarations[node]
			fmt.Printf("%t|%t|%t|%t|%t|%t|%s\n", node.ValuePresent, ap, a, bp, b, direct[id], node.Value)
		}
	}
}

var adamicLock sync.Mutex

func AdamicRecordNodes(nodes []*Node, s *utilityEvaluation, helper string) {
	path := os.Getenv("ADAMIC_SLOT04_INTEGER")
	if path == "" {
		return
	}
	c := AdamicCase{Helper: helper, Value: s.value, Modifier: s.modifier, Prefix: s.evaluator.Theme.Prefix, Entries: []AdamicEntry{}, Nodes: []AdamicCss{}, Roots: []int{}}
	keys := []string{}
	for k := range s.evaluator.Theme.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := s.evaluator.Theme.values[k]
		c.Entries = append(c.Entries, AdamicEntry{k, v.value, int(v.options)})
	}
	ids := map[*Node]int{}
	var visit func(*Node) int
	visit = func(n *Node) int {
		if id, ok := ids[n]; ok {
			return id
		}
		id := len(c.Nodes)
		ids[n] = id
		c.Nodes = append(c.Nodes, AdamicCss{Kind: string(n.Kind), Value: n.Value, Present: n.ValuePresent, Edges: []int{}})
		edges := []int{}
		for _, child := range n.Nodes {
			edges = append(edges, visit(child))
		}
		c.Nodes[id].Edges = edges
		return id
	}
	for _, n := range nodes {
		c.Roots = append(c.Roots, visit(n))
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e = json.NewEncoder(f).Encode(c); e != nil {
		panic(e)
	}
}
