package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type AdamicEntry struct {
	ID    int
	Value bool
}
type AdamicNode struct {
	Tag         string
	Data, Edges []int
	EdgesBound  bool
}
type AdamicAdaptation struct {
	Nodes  []AdamicNode
	Roots  []int
	Target int
}
type AdamicCase struct {
	Name, Helper, Source, Alias       string
	Data                              []int
	LeftPresent, RightPresent, Shared bool
	Left, Right                       []AdamicEntry
	Adaptation                        AdamicAdaptation
}

func adamicInts(s string) []int {
	out := []int{}
	for _, b := range []byte(s) {
		out = append(out, int(b))
	}
	return out
}
func adamicString(v []int) string {
	out := []byte{}
	for _, b := range v {
		out = append(out, byte(b))
	}
	return string(out)
}
func adamicJoined(v []int) string {
	s := []string{}
	for _, x := range v {
		s = append(s, fmt.Sprint(x))
	}
	return strings.Join(s, ",")
}
func adamicTarget(source string) []ValueNode {
	nodes := ParseValue(source)
	if len(nodes) == 0 {
		nodes = []ValueNode{{Kind: ValueNodeKindFunction, Value: "f", Nodes: []ValueNode{{Kind: ValueNodeKindWord, Value: "old"}}}}
	}
	return nodes
}
func AdamicAdapt(c AdamicCase) AdamicAdaptation {
	a := AdamicAdaptation{Nodes: []AdamicNode{}, Roots: []int{}, Target: 0}
	var flatten func([]ValueNode) []int
	flatten = func(nodes []ValueNode) []int {
		ids := []int{}
		for _, n := range nodes {
			id := len(a.Nodes)
			a.Nodes = append(a.Nodes, AdamicNode{Tag: string(n.Kind), Data: adamicInts(n.Value), Edges: []int{}, EdgesBound: n.Nodes != nil})
			a.Nodes[id].Edges = flatten(n.Nodes)
			ids = append(ids, id)
		}
		return ids
	}
	original := adamicTarget(c.Source)
	originalRoots := flatten(original)
	if c.Alias == "self" {
		a.Roots = originalRoots[:1]
	} else if c.Alias == "children" {
		a.Roots = append([]int{}, a.Nodes[0].Edges...)
	} else {
		a.Roots = flatten(ParseValue(adamicString(c.Data)))
	}
	return a
}
func AdamicObserve(c AdamicCase) {
	if c.Helper == "union" {
		pointers := map[int]*Node{}
		idOf := map[*Node]int{}
		ptr := func(id int) *Node {
			if id == -1 {
				idOf[nil] = -1
				return nil
			}
			if p, ok := pointers[id]; ok {
				return p
			}
			p := &Node{}
			pointers[id] = p
			idOf[p] = id
			return p
		}
		makeSet := func(present bool, entries []AdamicEntry) map[*Node]bool {
			if !present {
				return nil
			}
			m := map[*Node]bool{}
			for _, e := range entries {
				m[ptr(e.ID)] = e.Value
			}
			return m
		}
		left := makeSet(c.LeftPresent, c.Left)
		right := makeSet(c.RightPresent, c.Right)
		if c.Shared {
			right = left
		}
		result := unionNodeSets(left, right)
		if result != nil {
			result[ptr(999)] = true
		}
		ids := []int{}
		for p := range result {
			ids = append(ids, idOf[p])
		}
		sort.Ints(ids)
		parts := []string{}
		for _, id := range ids {
			parts = append(parts, fmt.Sprintf("%d:%t", id, result[ptr(id)]))
		}
		fmt.Printf("%t|%s|%t|%t\n", result != nil, strings.Join(parts, ","), left[ptr(999)], right[ptr(999)])
		return
	}
	fmt.Println(isQuotedLiteral(adamicString(c.Data)))
	target := adamicTarget(c.Source)
	resolved := ParseValue(adamicString(c.Data))
	if c.Alias == "self" {
		resolved = target[:1]
	} else if c.Alias == "children" {
		resolved = target[0].Nodes
	}
	replaceValueNode(&target[0], resolved)
	n := target[0]
	fmt.Printf("%s|%s|%d|%t\n", n.Kind, adamicJoined(adamicInts(n.Value)), len(n.Nodes), n.Nodes != nil)
}
func AdamicFull() {
	raw := func(data []int) { fmt.Println(isQuotedLiteral(adamicString(data))) }
	raw([]int{})
	for a := 0; a < 256; a++ {
		raw([]int{a})
		for b := 0; b < 256; b++ {
			raw([]int{a, b})
		}
	}
}

var adamicLock sync.Mutex

func adamicRecord(c AdamicCase) {
	path := os.Getenv("ADAMIC_SLOT04_SURGERY")
	if path == "" {
		return
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(c); err != nil {
		panic(err)
	}
}
func AdamicRecordQuote(value string) {
	adamicRecord(AdamicCase{Helper: "value", Data: adamicInts(value), Source: "f(old)"})
}
func AdamicRecordReplacement(node *ValueNode, resolved []ValueNode) {
	adamicRecord(AdamicCase{Helper: "value", Data: adamicInts(ValueToCss(resolved)), Source: ValueToCss([]ValueNode{*node})})
}
func AdamicRecordUnion(left, right map[*Node]bool) {
	ids := map[*Node]int{}
	entry := func(nodes map[*Node]bool) []AdamicEntry {
		out := []AdamicEntry{}
		for p, v := range nodes {
			id, ok := ids[p]
			if !ok {
				id = len(ids)
				ids[p] = id
			}
			out = append(out, AdamicEntry{id, v})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
		return out
	}
	adamicRecord(AdamicCase{Helper: "union", LeftPresent: left != nil, RightPresent: right != nil, Left: entry(left), Right: entry(right)})
}
