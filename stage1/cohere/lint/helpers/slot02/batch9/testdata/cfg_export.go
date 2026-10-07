package control_flow_graph

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type Slot02Block struct {
	Successors                           []int
	Reachable, Incoming, BarriersPresent bool
	Barriers                             []bool
}
type Slot02Case struct {
	Op              string
	From, To, Index int
	Barrier         bool
	Blocks          []Slot02Block
}

var Slot02Cases []Slot02Case
var slot02Capturing bool
var slot02Tracing bool
var Slot02Trace string
var slot02IDs map[*Block[int]]int

func slot02Capture[E any](op string, from, to *Block[E], index int, barrier bool) {
	if !slot02Capturing {
		return
	}
	pointers := []*Block[E]{}
	ids := map[*Block[E]]int{}
	add := func(p *Block[E]) int {
		if p == nil {
			return -1
		}
		if id, ok := ids[p]; ok {
			return id
		}
		id := len(pointers)
		ids[p] = id
		pointers = append(pointers, p)
		return id
	}
	c := Slot02Case{Op: op, From: add(from), To: add(to), Index: index, Barrier: barrier}
	if from != nil {
		for _, p := range from.Successors {
			add(p)
		}
	}
	if to != nil {
		for _, p := range to.Successors {
			add(p)
		}
	}
	for _, p := range pointers {
		v := Slot02Block{Reachable: p.Reachable, Incoming: p.hasIncoming, BarriersPresent: p.cycleBarrier != nil, Successors: []int{}, Barriers: append([]bool{}, p.cycleBarrier...)}
		for _, q := range p.Successors {
			if id, ok := ids[q]; ok {
				v.Successors = append(v.Successors, id)
			} else {
				v.Successors = append(v.Successors, -2)
			}
		}
		c.Blocks = append(c.Blocks, v)
	}
	Slot02Cases = append(Slot02Cases, c)
}
func slot02TraceCall[E any](op string, from, to *Block[E], index int, barrier bool) {
	if !slot02Tracing {
		return
	}
	// Only replay uses Builder[int]. Pointer identity is encoded independently of graph index.
	f, t := -1, -1
	for p, id := range slot02IDs {
		if any(p) == any(from) {
			f = id
		}
		if any(p) == any(to) {
			t = id
		}
	}
	if op == "with" {
		Slot02Trace += fmt.Sprintf("with:%d:%d:%t;", f, t, barrier)
	} else if op == "append" {
		Slot02Trace += fmt.Sprintf("append:%d:%d;", f, t)
	} else {
		Slot02Trace += fmt.Sprintf("barrier:%d:%d:%t;", f, index, barrier)
	}
}
func Slot02Collect(root *ast.Node) {
	slot02Capturing = true
	Build[int](root, Hooks[int]{})
	slot02Capturing = false
}
func Slot02Replay(c Slot02Case) string {
	blocks := make([]*Block[int], len(c.Blocks))
	slot02IDs = map[*Block[int]]int{}
	for i, v := range c.Blocks {
		blocks[i] = &Block[int]{Reachable: v.Reachable, hasIncoming: v.Incoming}
		slot02IDs[blocks[i]] = i
	}
	for i, v := range c.Blocks {
		for _, id := range v.Successors {
			if id < 0 {
				blocks[i].Successors = append(blocks[i].Successors, nil)
			} else {
				blocks[i].Successors = append(blocks[i].Successors, blocks[id])
			}
		}
		for j, q := range blocks[i].Successors {
			if j < 2 {
				blocks[i].successors[j] = q
			}
		}
		if len(blocks[i].Successors) <= 2 {
			blocks[i].Successors = blocks[i].successors[:len(blocks[i].Successors)]
		}
		if v.BarriersPresent {
			blocks[i].cycleBarrier = append([]bool{}, v.Barriers...)
		}
	}
	get := func(id int) *Block[int] {
		if id < 0 {
			return nil
		}
		return blocks[id]
	}
	b := &Builder[int]{}
	Slot02Trace = ""
	slot02Tracing = true
	answer := -2
	switch c.Op {
	case "set":
		b.setCycleBarrier(get(c.From), c.Index, c.Barrier)
	case "with":
		answer = b.linkWithCycleBarrier(get(c.From), get(c.To), c.Barrier)
	case "link":
		b.link(get(c.From), get(c.To))
	}
	slot02Tracing = false
	out := fmt.Sprintf("result:%d:%s\n", answer, Slot02Trace)
	for _, p := range blocks {
		out += fmt.Sprintf("block:%t:%t:%t:", p.Reachable, p.hasIncoming, p.cycleBarrier != nil)
		for _, q := range p.Successors {
			id := -2
			if q != nil {
				id = slot02IDs[q]
			}
			out += fmt.Sprintf("%d,", id)
		}
		out += ":"
		for _, v := range p.cycleBarrier {
			out += fmt.Sprintf("%t,", v)
		}
		out += "\n"
	}
	return out
}
func Slot02Unique() []Slot02Case {
	seen := map[string]bool{}
	out := []Slot02Case{}
	for _, c := range Slot02Cases {
		b, _ := json.Marshal(c)
		key := string(b)
		if !seen[key] {
			seen[key] = true
			out = append(out, c)
		}
	}
	return out
}
