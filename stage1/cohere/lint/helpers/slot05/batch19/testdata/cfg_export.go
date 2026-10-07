package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

var adamicTrace string
var adamicMuted bool
var adamicIDs map[*ast.Node]int
var adamicEffects map[string]AdamicSnapshot

func adamicID(n *ast.Node) int {
	if n == nil {
		return -1
	}
	if v, ok := adamicIDs[n]; ok {
		return v
	}
	v := len(adamicIDs)
	adamicIDs[n] = v
	return v
}
func adamicBlock[E any](n *Block[E]) int {
	if n == nil {
		return -1
	}
	return n.Index()
}

type AdamicBlockRow struct {
	Reachable, Incoming bool
	Edges               []int
}
type AdamicSnapshot struct {
	Cur    int
	Blocks []AdamicBlockRow
	Broken []bool
}

func (b *Builder[E]) adamicSnapshot() AdamicSnapshot {
	s := AdamicSnapshot{Cur: b.cur.Index(), Blocks: []AdamicBlockRow{}, Broken: []bool{}}
	for _, n := range b.blocks {
		r := AdamicBlockRow{Reachable: n.Reachable, Incoming: n.hasIncoming, Edges: []int{}}
		for _, e := range n.Successors {
			r.Edges = append(r.Edges, e.Index())
		}
		s.Blocks = append(s.Blocks, r)
	}
	for _, j := range b.jumps {
		s.Broken = append(s.Broken, j.broken)
	}
	return s
}
func (b *Builder[E]) adamicStack() string {
	var s strings.Builder
	for _, j := range b.jumps {
		fmt.Fprintf(&s, "[%s]:%d:%d:%d:%t:%t,", strings.Join(j.labels, ","), adamicBlock(j.breakTo), adamicBlock(j.continueTo), adamicID(j.loop), j.breakable, j.broken)
	}
	return s.String()
}
func (b *Builder[E]) adamicNewBlock() *Block[E] {
	n := b.newBlock()
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("new:%d;", n.Index())
	}
	return n
}
func (b *Builder[E]) adamicLink(from, to *Block[E]) {
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("link:%d:%d;", adamicBlock(from), adamicBlock(to))
	}
	b.link(from, to)
}
func (b *Builder[E]) adamicEnter(n *Block[E]) {
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("enter:%d;", n.Index())
	}
	b.enter(n)
}
func adamicTruthy(n *ast.Node) bool {
	v := isAlwaysTruthyTest(n)
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("truthy:%d:%t;", adamicID(n), v)
	}
	return v
}
func adamicBreakable(n *ast.Node) bool {
	v := isBreakableStatement(n)
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("breakable:%d:%t;", adamicID(n), v)
	}
	return v
}
func adamicText(n *ast.Node) string {
	v := n.Text()
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("text:%d:%s;", adamicID(n), v)
	}
	return v
}
func (b *Builder[E]) adamicPushJump(n *ast.Node, a, t *Block[E], l *ast.Node) {
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("push:%d:%d:%d:%d;", adamicID(n), adamicBlock(a), adamicBlock(t), adamicID(l))
	}
	b.pushJump(n, a, t, l)
}
func (b *Builder[E]) adamicPopJump() {
	if !adamicMuted {
		adamicTrace += "pop:" + b.adamicStack() + ";"
	}
	b.popJump()
}
func (b *Builder[E]) adamicStatement(n *ast.Node) {
	if adamicMuted {
		b.statement(n)
		return
	}
	adamicTrace += fmt.Sprintf("stmt:%d:%d:%s;", adamicID(n), b.cur.Index(), b.adamicStack())
	adamicMuted = true
	b.statement(n)
	adamicMuted = false
	adamicEffects["stmt"] = b.adamicSnapshot()
}
func (b *Builder[E]) adamicExpr(n *ast.Node) {
	if adamicMuted {
		b.expr(n)
		return
	}
	adamicTrace += fmt.Sprintf("expr:%d:%d;", adamicID(n), b.cur.Index())
	adamicMuted = true
	b.expr(n)
	adamicMuted = false
	adamicEffects["expr"] = b.adamicSnapshot()
}
func (b *Builder[E]) adamicCondition(n *ast.Node, y, no *Block[E]) {
	if adamicMuted {
		b.condition(n, y, no)
		return
	}
	adamicTrace += fmt.Sprintf("condition:%d:%d:%d;", adamicID(n), adamicBlock(y), adamicBlock(no))
	adamicMuted = true
	b.condition(n, y, no)
	adamicMuted = false
	adamicEffects["condition"] = b.adamicSnapshot()
}
func (b *Builder[E]) adamicLoop(n *ast.Node) {
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("loop:%d:%d;", adamicID(n), b.cur.Index())
	}
	b.loop(n)
}

type AdamicStatementRow struct {
	Node, Expression, Statement, Label, Prefix int
	Text                                       string
	Truthy, Breakable                          bool
	Labels                                     []string
	Start                                      AdamicSnapshot
	Effects                                    map[string]AdamicSnapshot
}

func AdamicStatementObserve(mode string, n *ast.Node, reachable bool, prefix int) (AdamicStatementRow, string) {
	b := &Builder[int]{}
	b.cur = b.newBlock()
	b.cur.Reachable = reachable
	b.cur.hasIncoming = reachable
	adamicIDs = map[*ast.Node]int{}
	adamicID(n)
	for i := 0; i < prefix; i++ {
		b.jumps = append(b.jumps, jumpTarget[int]{labels: []string{fmt.Sprintf("seed%d", i)}, breakTo: b.cur, continueTo: b.cur, loop: n, breakable: i%2 == 0, broken: i%2 == 1})
	}
	adamicTrace = ""
	adamicMuted = false
	adamicEffects = map[string]AdamicSnapshot{}
	r := AdamicStatementRow{Node: adamicID(n), Prefix: prefix, Expression: -1, Label: -1, Start: b.adamicSnapshot(), Labels: labelsOf(n)}
	var expr, stmt *ast.Node
	switch mode {
	case "while":
		s := n.AsWhileStatement()
		expr, stmt = s.Expression, s.Statement
	case "do":
		s := n.AsDoStatement()
		expr, stmt = s.Expression, s.Statement
	case "label":
		s := n.AsLabeledStatement()
		stmt = s.Statement
		r.Label = adamicID(s.Label)
		r.Text = s.Label.Text()
		r.Breakable = isBreakableStatement(stmt)
	}
	r.Expression = adamicID(expr)
	r.Statement = adamicID(stmt)
	r.Truthy = expr != nil && isAlwaysTruthyTest(expr)
	switch mode {
	case "while":
		b.whileStatement(n)
	case "do":
		b.doStatement(n)
	case "label":
		b.labeledStatement(n)
	}
	r.Effects = adamicEffects
	var out strings.Builder
	fmt.Fprintf(&out, "%d|%s|", b.cur.Index(), b.adamicStack())
	for _, v := range b.adamicSnapshot().Blocks {
		fmt.Fprintf(&out, "%t:%t:", v.Reachable, v.Incoming)
		for _, e := range v.Edges {
			fmt.Fprintf(&out, "%d,", e)
		}
		out.WriteString(";")
	}
	out.WriteString("|" + adamicTrace + "\n")
	return r, out.String()
}
