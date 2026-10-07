package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

type Slot11Case struct {
	Op                  string
	Node, Builder, Mode int
	IsList, HasList     bool
	Declarations        []int
}

var Slot11Calls int
var Slot11Cases = []Slot11Case{}
var slot11Capturing, slot11Replaying bool
var slot11IDs = map[*ast.Node]int{}
var slot11Nodes = []*ast.Node{}

func slot11Node(n *ast.Node) int {
	if n == nil {
		return -1
	}
	if id, ok := slot11IDs[n]; ok {
		return id
	}
	id := len(slot11Nodes)
	slot11IDs[n] = id
	slot11Nodes = append(slot11Nodes, n)
	return id
}
func slot11Capture[E any](op string, b *Builder[E], n *ast.Node) {
	if !slot11Capturing {
		return
	}
	Slot11Calls++
	c := Slot11Case{Op: op, Node: slot11Node(n), Declarations: []int{}}
	if n != nil && n.Kind == ast.KindVariableDeclarationList {
		c.IsList = true
		list := n.AsVariableDeclarationList()
		if list.Declarations != nil {
			c.HasList = true
			for _, d := range list.Declarations.Nodes {
				c.Declarations = append(c.Declarations, slot11Node(d))
			}
		}
	}
	Slot11Cases = append(Slot11Cases, c)
}
func Slot11Collect(root *ast.Node) {
	slot11Capturing = true
	Build[int](root, Hooks[int]{})
	slot11Capturing = false
}

var slot11Builders []*Builder[int]
var slot11Blocks []*Block[int]
var slot11Mode, slot11Count int
var slot11Trace string

func slot11Hook[E any](op string, b *Builder[E], n *ast.Node) {
	id := -1
	for i, p := range slot11Builders {
		if any(p) == any(b) {
			id = i
		}
	}
	nid := slot11Node(n)
	slot11Count++
	slot11Trace += fmt.Sprintf("%s:%d:%d;", op, id, nid)
	b.Emit(any(nid).(E))
	if slot11Mode == 1 {
		b.cur = any(slot11Blocks[1-id]).(*Block[E])
		b.Emit(any(100 + nid).(E))
	}
}
func Slot11Header[E any](b *Builder[E], n *ast.Node) {
	if !slot11Replaying {
		b.memberHeader(n)
		return
	}
	slot11Hook("header", b, n)
}
func Slot11Declaration[E any](b *Builder[E], n *ast.Node) {
	if !slot11Replaying {
		b.variableDeclaration(n)
		return
	}
	slot11Hook("decl", b, n)
}
func Slot11CFG() ([]Slot11Case, string) {
	slot11Replaying = true
	factory := ast.NodeFactory{}
	nilList := factory.NewVariableDeclarationList(nil, 0)
	slot11Capturing = true
	slot11Capture("list", &Builder[int]{}, nilList)
	slot11Capturing = false
	observed := append([]Slot11Case{}, Slot11Cases...)
	observed = append(observed, Slot11Case{Op: "nested", Node: -1, Declarations: []int{}}, Slot11Case{Op: "list", Node: -1, Declarations: []int{}})
	cases := []Slot11Case{}
	var want strings.Builder
	for _, base := range observed {
		for builder := 0; builder < 2; builder++ {
			for mode := 0; mode < 2; mode++ {
				c := base
				c.Builder = builder
				c.Mode = mode
				slot11Mode = mode
				slot11Count = 0
				slot11Trace = ""
				slot11Blocks = []*Block[int]{{Reachable: true, Events: []int{11}}, {Reachable: false, Events: []int{22}}}
				slot11Builders = []*Builder[int]{{cur: slot11Blocks[0]}, {cur: slot11Blocks[1]}}
				var n *ast.Node
				if c.Node >= 0 {
					n = slot11Nodes[c.Node]
				}
				if c.Op == "nested" {
					slot11Builders[builder].nestedFunction(n)
				} else {
					slot11Builders[builder].variableDeclarationList(n)
				}
				want.WriteString(fmt.Sprintf("calls:%d:%s\n", slot11Count, slot11Trace))
				for i, b := range slot11Builders {
					cur := 0
					if b.cur == slot11Blocks[1] {
						cur = 1
					}
					want.WriteString(fmt.Sprintf("builder:%d:%d\n", i, cur))
				}
				for i, b := range slot11Blocks {
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
