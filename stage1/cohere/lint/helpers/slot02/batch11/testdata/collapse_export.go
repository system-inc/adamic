package tailwind

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Slot11Node struct {
	AtRule, Container bool
	Name, Params      string
	Children          []int
}
type Slot11Ingest struct {
	Nodes []Slot11Node
	Roots []int
	Path  string
	Fail  int
}

var slot11IDs map[*Node]int
var slot11Fail, slot11Calls int
var slot11Trace string
var slot11Error = errors.New("dependency failure")

func slot11Call(op string, n *Node, path string) error {
	id := -1
	if n != nil {
		id = slot11IDs[n]
	}
	slot11Trace += fmt.Sprintf("%s:%d:%s;", op, id, path)
	slot11Calls++
	if slot11Calls == slot11Fail {
		return slot11Error
	}
	return nil
}
func Slot11Resolve(c *stylesheetCollector, n *Node, p string) (string, error) {
	e := slot11Call("resolve", n, p)
	return "resolved:" + n.Params, e
}
func Slot11Load(c *stylesheetCollector, p string) error           { return slot11Call("load", nil, p) }
func Slot11Theme(c *stylesheetCollector, n *Node, p string) error { return slot11Call("theme", n, p) }
func Slot11Utility(c *stylesheetCollector, n *Node, p string) error {
	return slot11Call("utility", n, p)
}
func Slot11Custom(c *stylesheetCollector, n *Node) { slot11Call("custom", n, "") }
func Slot11CSS(texts []string) ([]Slot11Ingest, string) {
	cases := []Slot11Ingest{}
	var want strings.Builder
	texts = append(texts, `@config "a";@plugin "b";.x{@layer t{@theme{--x:1}@utility foo{color:red}@import "x";@custom-variant v (&)}}@config "last";`, "")
	for _, text := range texts {
		roots, e := ParseCSS(text)
		if e != nil {
			continue
		}
		nodes := []Slot11Node{}
		slot11IDs = map[*Node]int{}
		var add func(*Node) int
		add = func(n *Node) int {
			id := len(nodes)
			slot11IDs[n] = id
			nodes = append(nodes, Slot11Node{AtRule: n.Kind == KindAtRule, Container: n.IsContainer(), Name: n.Name, Params: n.Params, Children: []int{}})
			children := []int{}
			for _, child := range n.Nodes {
				children = append(children, add(child))
			}
			nodes[id].Children = children
			return id
		}
		ids := []int{}
		for _, n := range roots {
			ids = append(ids, add(n))
		}
		for fail := 0; fail < 7; fail++ {
			slot11Fail = fail
			slot11Calls = 0
			slot11Trace = ""
			c := &stylesheetCollector{}
			err := c.ingest(roots, "/fixture/é😀.css")
			result := -1
			if err != nil {
				if err != slot11Error {
					panic("error identity")
				}
				result = 7
			}
			want.WriteString(fmt.Sprintf("result:%d:%s\n", result, slot11Trace))
			for _, s := range c.skipped {
				want.WriteString(fmt.Sprintf("skip:%s:%s:%s\n", s.Name, s.Params, s.Path))
			}
			cases = append(cases, Slot11Ingest{Nodes: nodes, Roots: ids, Path: "/fixture/é😀.css", Fail: fail})
		}
	}
	return cases, want.String()
}
func Slot11Marshal(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return string(b)
}
