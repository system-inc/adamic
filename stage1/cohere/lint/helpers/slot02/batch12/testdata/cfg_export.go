package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

type Slot12Case struct {
	Node, Builder, Mode, Type, Name, Initializer int
	Pattern                                      bool
}

var Slot12Calls int
var Slot12Cases = []Slot12Case{}
var slot12Capturing, slot12Replaying bool
var slot12IDs = map[*ast.Node]int{}
var slot12Nodes = []*ast.Node{}

func slot12Node(n *ast.Node) int {
	if n == nil {
		return -1
	}
	if id, ok := slot12IDs[n]; ok {
		return id
	}
	id := len(slot12Nodes)
	slot12IDs[n] = id
	slot12Nodes = append(slot12Nodes, n)
	return id
}
func slot12Capture[E any](b *Builder[E], n *ast.Node) {
	if !slot12Capturing {
		return
	}
	Slot12Calls++
	c := Slot12Case{Node: slot12Node(n), Type: -1, Name: -1, Initializer: -1}
	if n != nil {
		d := n.AsVariableDeclaration()
		c.Type = slot12Node(d.Type)
		c.Name = slot12Node(d.Name())
		c.Initializer = slot12Node(d.Initializer)
		c.Pattern = d.Name() != nil && (d.Name().Kind == ast.KindObjectBindingPattern || d.Name().Kind == ast.KindArrayBindingPattern)
	}
	Slot12Cases = append(Slot12Cases, c)
}
func Slot12Collect(root *ast.Node) {
	slot12Capturing = true
	Build[int](root, Hooks[int]{})
	slot12Capturing = false
}

var slot12Builders []*Builder[int]
var slot12Blocks []*Block[int]
var slot12Mode, slot12Count int
var slot12Trace string

func slot12Hook[E any](op string, b *Builder[E], n *ast.Node) {
	id := -1
	for i, p := range slot12Builders {
		if any(p) == any(b) {
			id = i
		}
	}
	nid := slot12Node(n)
	slot12Count++
	slot12Trace += fmt.Sprintf("%s:%d:%d;", op, id, nid)
	b.Emit(any(nid).(E))
	if slot12Mode == 1 {
		b.cur = any(slot12Blocks[1-id]).(*Block[E])
		b.Emit(any(100 + nid).(E))
	}
}
func Slot12Expression[E any](b *Builder[E], n *ast.Node) {
	if !slot12Replaying {
		b.expr(n)
		return
	}
	slot12Hook("expr", b, n)
}
func Slot12Bind[E any](b *Builder[E], n *ast.Node) {
	if !slot12Replaying {
		b.patternBind(n)
		return
	}
	slot12Hook("bind", b, n)
}
func Slot12Reads[E any](b *Builder[E], n *ast.Node) {
	if !slot12Replaying {
		b.patternReads(n)
		return
	}
	slot12Hook("read", b, n)
}
func Slot12Writes[E any](b *Builder[E], n *ast.Node) {
	if !slot12Replaying {
		b.patternWrites(n)
		return
	}
	slot12Hook("write", b, n)
}
func Slot12CFG() ([]Slot12Case, string) {
	slot12Replaying = true
	observed := append([]Slot12Case{}, Slot12Cases...)
	observed = append(observed, Slot12Case{Node: -1, Type: -1, Name: -1, Initializer: -1})
	cases := []Slot12Case{}
	var want strings.Builder
	for _, base := range observed {
		for builder := 0; builder < 2; builder++ {
			for mode := 0; mode < 2; mode++ {
				c := base
				c.Builder = builder
				c.Mode = mode
				slot12Mode = mode
				slot12Count = 0
				slot12Trace = ""
				slot12Blocks = []*Block[int]{{Reachable: true, Events: []int{11}}, {Reachable: false, Events: []int{22}}}
				slot12Builders = []*Builder[int]{{cur: slot12Blocks[0]}, {cur: slot12Blocks[1]}}
				var n *ast.Node
				if c.Node >= 0 {
					n = slot12Nodes[c.Node]
				}
				slot12Builders[builder].variableDeclaration(n)
				want.WriteString(fmt.Sprintf("calls:%d:%s\n", slot12Count, slot12Trace))
				for i, b := range slot12Builders {
					cur := 0
					if b.cur == slot12Blocks[1] {
						cur = 1
					}
					want.WriteString(fmt.Sprintf("builder:%d:%d\n", i, cur))
				}
				for i, b := range slot12Blocks {
					want.WriteString(fmt.Sprintf("block:%d:", i))
					for _, e := range b.Events {
						want.WriteString(fmt.Sprintf("%d,", e))
					}
					want.WriteString("\n")
				}
				cases = append(cases, c)
			}
		}
	}
	return cases, want.String()
}
