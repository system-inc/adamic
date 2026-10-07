package control_flow_graph

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type Slot10Case struct {
	Op                  string
	Node, Builder, Mode int
	Present, Reachable  bool
}

var Slot10Cases []Slot10Case
var Slot10Calls int
var Slot10Nodes []*ast.Node
var slot10Capturing bool
var slot10Seen = map[string]bool{}
var slot10NodeIDs = map[ast.Kind]int{}

func slot10Node(node *ast.Node) int {
	if node == nil {
		return -1
	}
	if id, ok := slot10NodeIDs[node.Kind]; ok {
		return id
	}
	id := len(Slot10Nodes)
	slot10NodeIDs[node.Kind] = id
	Slot10Nodes = append(Slot10Nodes, node)
	return id
}
func slot10Capture[E any](op string, b *Builder[E], node *ast.Node) {
	if !slot10Capturing {
		return
	}
	Slot10Calls++
	n := slot10Node(node)
	c := Slot10Case{Op: op, Node: n, Reachable: b.cur.Reachable}
	key := fmt.Sprintf("%s:%d:%t", op, n, c.Reachable)
	if !slot10Seen[key] {
		slot10Seen[key] = true
		Slot10Cases = append(Slot10Cases, c)
	}
}
func Slot10Collect(root *ast.Node) {
	slot10Capturing = true
	Build[int](root, Hooks[int]{})
	slot10Capturing = false
}
func Slot10Replay(c Slot10Case) string {
	blocks := []*Block[int]{{Reachable: c.Reachable, Events: []int{11}}, {Reachable: !c.Reachable, Events: []int{22}}}
	builders := []*Builder[int]{{cur: blocks[0]}, {cur: blocks[1]}}
	b := builders[c.Builder]
	var node *ast.Node
	if c.Node >= 0 {
		node = Slot10Nodes[c.Node]
	}
	trace := ""
	calls := 0
	hook := func(received *Builder[int], n *ast.Node) {
		id := -1
		for i, p := range builders {
			if p == received {
				id = i
			}
		}
		nid := -1
		if n != nil {
			for i, p := range Slot10Nodes {
				if p == n {
					nid = i
					break
				}
			}
		}
		calls++
		trace += fmt.Sprintf("hook:%d:%d:%d;", id, nid, calls)
		received.Emit(nid)
		if c.Mode == 1 {
			received.cur = blocks[1-id]
			received.Emit(100 + nid)
		}
	}
	if c.Present {
		switch c.Op {
		case "read":
			b.hooks.Read = hook
		case "write":
			b.hooks.Write = hook
		case "loop":
			b.hooks.Loop = hook
		}
	}
	switch c.Op {
	case "read":
		b.read(node)
	case "write":
		b.write(node)
	case "loop":
		b.loop(node)
	}
	out := fmt.Sprintf("calls:%d:%s\n", calls, trace)
	for i, p := range builders {
		current := -1
		for j, blk := range blocks {
			if p.cur == blk {
				current = j
			}
		}
		out += fmt.Sprintf("builder:%d:%d\n", i, current)
	}
	for i, p := range blocks {
		out += fmt.Sprintf("block:%d:%t:", i, p.Reachable)
		for _, e := range p.Events {
			out += fmt.Sprintf("%d,", e)
		}
		out += "\n"
	}
	return out
}
func Slot10CasesJSON() string {
	data, err := json.Marshal(Slot10Cases)
	if err != nil {
		panic(err)
	}
	return string(data)
}
