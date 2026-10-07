package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

type AdamicCss struct {
	Kind     string
	Children []int
}
type AdamicValue struct {
	Kind            string
	Bytes, Children []int
}
type AdamicCase struct {
	Name       string
	Helper     string
	Css        []AdamicCss
	Roots      []int
	Values     []AdamicValue
	ValueRoots []int
}

func adamicInts(s string) []int {
	out := []int{}
	for _, b := range []byte(s) {
		out = append(out, int(b))
	}
	return out
}
func adamicText(v []int) string {
	out := []byte{}
	for _, b := range v {
		out = append(out, byte(b))
	}
	return string(out)
}
func adamicJoined(v []int) string {
	out := []string{}
	for _, x := range v {
		out = append(out, fmt.Sprint(x))
	}
	return strings.Join(out, ",")
}
func AdamicTreeCase(css []*Node, values []ValueNode) AdamicCase {
	c := AdamicCase{Name: "control:call", Css: []AdamicCss{}, Roots: []int{}, Values: []AdamicValue{}, ValueRoots: []int{}}
	ids := map[*Node]int{}
	var nodeID func(*Node) int
	nodeID = func(n *Node) int {
		if n == nil {
			panic("nil CSS adapter")
		}
		if id, ok := ids[n]; ok {
			return id
		}
		id := len(c.Css)
		ids[n] = id
		c.Css = append(c.Css, AdamicCss{Kind: string(n.Kind), Children: []int{}})
		children := []int{}
		for _, child := range n.Nodes {
			children = append(children, nodeID(child))
		}
		c.Css[id].Children = children
		return id
	}
	for _, n := range css {
		c.Roots = append(c.Roots, nodeID(n))
	}
	var valueID func(ValueNode) int
	valueID = func(n ValueNode) int {
		id := len(c.Values)
		c.Values = append(c.Values, AdamicValue{Kind: string(n.Kind), Bytes: adamicInts(n.Value), Children: []int{}})
		children := []int{}
		for _, child := range n.Nodes {
			children = append(children, valueID(child))
		}
		c.Values[id].Children = children
		return id
	}
	for _, n := range values {
		c.ValueRoots = append(c.ValueRoots, valueID(n))
	}
	return c
}
func AdamicRawTree(input []byte) AdamicCase {
	b := adamicInts(string(input))
	return AdamicCase{Css: []AdamicCss{{"rule", []int{1, 2}}, {"context", []int{3}}, {"declaration", []int{3}}, {"at-root", []int{4}}, {"at-rule", []int{5}}, {"comment", []int{}}}, Roots: []int{0, 2}, Values: []AdamicValue{{"word", b, []int{}}, {"separator", []int{32}, []int{}}, {"function", []int{102}, []int{3}}, {"word", b, []int{}}, {"foreign", b, []int{3}}}, ValueRoots: []int{0, 1, 2, 4}}
}
func adamicAction(mode, id int) WalkAction {
	if mode == 1 && id == 0 {
		return WalkSkip
	}
	if mode == 2 && id == 1 {
		return WalkStop
	}
	if mode == 3 && id == 0 {
		return WalkAction(99)
	}
	if mode == 4 && id == 0 {
		return WalkStop
	}
	if mode == 5 && id%2 == 0 {
		return WalkSkip
	}
	return WalkContinue
}
func AdamicObserveTree(c AdamicCase) {
	arena := make([]*Node, len(c.Css))
	ids := map[*Node]int{}
	for i, n := range c.Css {
		arena[i] = &Node{Kind: NodeKind(n.Kind)}
		ids[arena[i]] = i
	}
	for i, n := range c.Css {
		for _, child := range n.Children {
			arena[i].Nodes = append(arena[i].Nodes, arena[child])
		}
	}
	roots := []*Node{}
	for _, id := range c.Roots {
		roots = append(roots, arena[id])
	}
	for mode := 0; mode < 6; mode++ {
		trace := []int{}
		visit := func(n *Node) WalkAction { id := ids[n]; trace = append(trace, id); return adamicAction(mode, id) }
		result := walkNodes(roots, visit)
		fmt.Printf("nodes %d %t|%s\n", mode, result, adamicJoined(trace))
		trace = []int{}
		Walk(roots, visit)
		fmt.Printf("walk %d %s\n", mode, adamicJoined(trace))
	}
	var value func(int) ValueNode
	value = func(id int) ValueNode {
		v := c.Values[id]
		out := ValueNode{Kind: ValueNodeKind(v.Kind), Value: adamicText(v.Bytes)}
		for _, child := range v.Children {
			out.Nodes = append(out.Nodes, value(child))
		}
		return out
	}
	values := []ValueNode{}
	for _, id := range c.ValueRoots {
		values = append(values, value(id))
	}
	var builder strings.Builder
	builder.Write([]byte{137, 0, 255})
	writeValueCss(&builder, values)
	fmt.Printf("css %s\n", adamicJoined(adamicInts(builder.String())))
	writeValueCss(&builder, values)
	fmt.Printf("again %s\n", adamicJoined(adamicInts(builder.String())))
}

var adamicLock sync.Mutex

func AdamicRecordTree(helper string, css []*Node, values []ValueNode) {
	path := os.Getenv("ADAMIC_SLOT04_TREES")
	if path == "" {
		return
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	c := AdamicTreeCase(css, values)
	c.Helper = helper
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(c); err != nil {
		panic(err)
	}
}
