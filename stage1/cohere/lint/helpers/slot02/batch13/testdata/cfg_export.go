package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

type Slot13Node struct {
	Kind, Operator string
	A, B, C        int
	Children       []int
}
type Slot13Case struct {
	Op                               string
	Node, Builder, Mode, True, False int
}

var slot13Nodes = []*ast.Node{}
var slot13IDs = map[*ast.Node]int{}
var slot13Lists = map[*ast.NodeList]int{}

func slot13Node(n *ast.Node) int {
	if n == nil {
		return -1
	}
	if id, ok := slot13IDs[n]; ok {
		return id
	}
	id := len(slot13Nodes)
	slot13IDs[n] = id
	slot13Nodes = append(slot13Nodes, n)
	return id
}
func slot13List(n *ast.NodeList) int {
	if n == nil {
		return -1
	}
	if id, ok := slot13Lists[n]; ok {
		return id
	}
	id := len(slot13Lists)
	slot13Lists[n] = id
	return id
}
func Slot13Collect(n *ast.Node) {
	slot13Node(n)
	n.ForEachChild(func(c *ast.Node) bool { Slot13Collect(c); return false })
}

var Slot13CFGWant []string
var slot13Trace strings.Builder
var slot13Blocks []*Block[int]
var slot13Builders []*Builder[int]
var slot13Mode int

func slot13Builder[E any](b *Builder[E]) int {
	for i, p := range slot13Builders {
		if any(p) == any(b) {
			return i
		}
	}
	panic("builder identity")
}
func slot13Block[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	for i, p := range slot13Blocks {
		if any(p) == any(b) {
			return i
		}
	}
	panic("block identity")
}
func Slot13Call[E any](op string, b *Builder[E], n *ast.Node) {
	id := slot13Builder(b)
	fmt.Fprintf(&slot13Trace, "%s:%d:%d;", op, id, slot13Node(n))
	if slot13Mode == 1 && (op == "expr" || op == "hook") {
		b.cur = any(slot13Blocks[1-id]).(*Block[E])
	}
}
func Slot13List[E any](b *Builder[E], n *ast.NodeList) {
	fmt.Fprintf(&slot13Trace, "statements:%d:%d;", slot13Builder(b), slot13List(n))
}
func Slot13Link[E any](b *Builder[E], from, to *Block[E]) {
	fmt.Fprintf(&slot13Trace, "link:%d:%d:%d;", slot13Builder(b), slot13Block(from), slot13Block(to))
}
func Slot13New[E any](b *Builder[E]) *Block[E] {
	id := len(slot13Blocks)
	p := &Block[int]{}
	slot13Blocks = append(slot13Blocks, p)
	fmt.Fprintf(&slot13Trace, "new:%d:%d;", slot13Builder(b), id)
	return any(p).(*Block[E])
}
func Slot13Enter[E any](b *Builder[E], p *Block[E]) {
	fmt.Fprintf(&slot13Trace, "enter:%d:%d;", slot13Builder(b), slot13Block(p))
	b.cur = p
}
func Slot13CFG() ([]Slot13Node, []Slot13Case, string) {
	factory := ast.NodeFactory{}
	slot13Node(factory.NewBlock(nil, false))
	views := []Slot13Node{}
	for i := 0; i < len(slot13Nodes); i++ {
		n := slot13Nodes[i]
		v := Slot13Node{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), A: -1, B: -1, C: -1, Children: []int{}}
		n.ForEachChild(func(c *ast.Node) bool { v.Children = append(v.Children, slot13Node(c)); return false })
		switch n.Kind {
		case ast.KindParenthesizedExpression:
			v.A = slot13Node(n.AsParenthesizedExpression().Expression)
		case ast.KindPrefixUnaryExpression:
			x := n.AsPrefixUnaryExpression()
			v.A = slot13Node(x.Operand)
			v.Operator = strings.TrimPrefix(x.Operator.String(), "Kind")
		case ast.KindBinaryExpression:
			x := n.AsBinaryExpression()
			v.A = slot13Node(x.Left)
			v.B = slot13Node(x.Right)
			v.Operator = strings.TrimPrefix(x.OperatorToken.Kind.String(), "Kind")
		case ast.KindConditionalExpression:
			x := n.AsConditionalExpression()
			v.A = slot13Node(x.Condition)
			v.B = slot13Node(x.WhenTrue)
			v.C = slot13Node(x.WhenFalse)
		case ast.KindBlock:
			v.A = slot13List(n.AsBlock().Statements)
		case ast.KindVariableStatement:
			v.A = slot13Node(n.AsVariableStatement().DeclarationList)
		case ast.KindExpressionStatement:
			v.A = slot13Node(n.AsExpressionStatement().Expression)
		case ast.KindReturnStatement:
			v.A = slot13Node(n.AsReturnStatement().Expression)
		case ast.KindThrowStatement:
			v.A = slot13Node(n.AsThrowStatement().Expression)
		case ast.KindBreakStatement:
			v.A = slot13Node(n.AsBreakStatement().Label)
		case ast.KindContinueStatement:
			v.A = slot13Node(n.AsContinueStatement().Label)
		case ast.KindWithStatement:
			v.A = slot13Node(n.AsWithStatement().Expression)
			v.B = slot13Node(n.AsWithStatement().Statement)
		case ast.KindExportAssignment:
			v.A = slot13Node(n.AsExportAssignment().Expression)
		}
		views = append(views, v)
	}
	cases := []Slot13Case{}
	var want strings.Builder
	for node := -1; node < len(slot13Nodes); node++ {
		for _, op := range []string{"condition", "statement"} {
			for mode := 0; mode < 3; mode++ {
				for builder := 0; builder < 2; builder++ {
					slot13Blocks = []*Block[int]{{}, {}, {}, {}}
					slot13Builders = []*Builder[int]{{cur: slot13Blocks[0]}, {cur: slot13Blocks[1]}}
					slot13Mode = mode
					slot13Trace.Reset()
					b := slot13Builders[builder]
					if mode != 2 {
						b.hooks.Expression = func(b *Builder[int], n *ast.Node) { Slot13Call("hook", b, n) }
					}
					var n *ast.Node
					if node >= 0 {
						n = slot13Nodes[node]
					}
					yes, no := 2, 3
					if mode == 2 {
						yes = -1
						no = 2
					}
					var t, f *Block[int]
					if yes >= 0 {
						t = slot13Blocks[yes]
					}
					if no >= 0 {
						f = slot13Blocks[no]
					}
					if op == "condition" {
						b.condition(n, t, f)
					} else {
						b.statement(n)
					}
					line := fmt.Sprintf("%s|%d:%d", slot13Trace.String(), slot13Block(slot13Builders[0].cur), slot13Block(slot13Builders[1].cur))
					want.WriteString(line + "\n")
					Slot13CFGWant = append(Slot13CFGWant, line)
					cases = append(cases, Slot13Case{op, node, builder, mode, yes, no})
				}
			}
		}
	}
	return views, cases, want.String()
}
