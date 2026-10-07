package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

var adamicTrace string
var adamicIDs map[*ast.Node]int

func adamicID(n *ast.Node) int {
	if n == nil {
		return -1
	}
	if id, ok := adamicIDs[n]; ok {
		return id
	}
	id := len(adamicIDs)
	adamicIDs[n] = id
	return id
}
func (b *Builder[E]) adamicReads(n *ast.Node) {
	adamicTrace += fmt.Sprintf("read:%d;", adamicID(n))
	outer := adamicTrace
	b.patternReads(n)
	adamicTrace = outer
}
func (b *Builder[E]) adamicWrites(n *ast.Node) {
	adamicTrace += fmt.Sprintf("write:%d;", adamicID(n))
	outer := adamicTrace
	b.patternWrites(n)
	adamicTrace = outer
}
func (b *Builder[E]) adamicExpr(n *ast.Node) {
	adamicTrace += fmt.Sprintf("expr:%d;", adamicID(n))
	outer := adamicTrace
	b.expr(n)
	adamicTrace = outer
}
func (b *Builder[E]) adamicBind(n, i *ast.Node) {
	adamicTrace += fmt.Sprintf("bind:%d:%d;", adamicID(n), adamicID(i))
	outer := adamicTrace
	b.bindWithDefault(n, i)
	adamicTrace = outer
}

type AdamicNodeRow struct{ Node, Type, Name, Initializer int }

func AdamicNodeObserve(mode string, n *ast.Node) (AdamicNodeRow, string) {
	b := &Builder[int]{}
	b.cur = b.newBlock()
	b.cur.Reachable = true
	b.cur.hasIncoming = true
	adamicIDs = map[*ast.Node]int{nil: -1}
	row := AdamicNodeRow{Node: adamicID(n), Type: -1, Name: -1, Initializer: -1}
	if mode == "parameter" {
		p := n.AsParameterDeclaration()
		row.Type = adamicID(p.Type)
		row.Name = adamicID(p.Name())
		row.Initializer = adamicID(p.Initializer)
	}
	adamicTrace = ""
	if mode == "parameter" {
		b.parameter(n)
	} else {
		b.updateExpression(n)
	}
	return row, adamicTrace + "\n"
}

type AdamicFrame struct {
	Position                int
	HasFinally, Forked, Any bool
	Catch, Finally          int
}
type AdamicForkRow struct {
	Reachable     bool
	Frames        []AdamicFrame
	Index, Target int
}

func (b *Builder[E]) adamicThrowFrame() int {
	i := b.throwFrame()
	adamicTrace += fmt.Sprintf("frame:%d;", i)
	return i
}
func adamicThrowTarget[E any](f *tryFrame[E]) *Block[E] {
	t := throwTarget(f)
	adamicTrace += fmt.Sprintf("target:%d;", adamicBlockID(t))
	return t
}
func adamicBlockID[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	return b.Index()
}
func (b *Builder[E]) adamicLink(from, to *Block[E]) {
	adamicTrace += fmt.Sprintf("link:%d:%d:", adamicBlockID(from), adamicBlockID(to))
	for _, f := range b.tryStack {
		adamicTrace += fmt.Sprintf("%t:%t,", f.thrownForked, f.thrownAny)
	}
	adamicTrace += ";"
	b.link(from, to)
}
func (b *Builder[E]) adamicNewBlock() *Block[E] {
	n := b.newBlock()
	adamicTrace += fmt.Sprintf("new:%d;", n.Index())
	return n
}
func (b *Builder[E]) adamicEnter(n *Block[E]) {
	adamicTrace += fmt.Sprintf("enter:%d;", n.Index())
	b.enter(n)
}
func AdamicForkObserve(row AdamicForkRow, repeats int) (AdamicForkRow, string) {
	b := &Builder[int]{}
	for i := 0; i < 6; i++ {
		b.newBlock()
	}
	b.cur = b.blocks[0]
	b.cur.Reachable = row.Reachable
	block := func(id int) *Block[int] {
		if id < 0 {
			return nil
		}
		return b.blocks[id]
	}
	for _, f := range row.Frames {
		b.tryStack = append(b.tryStack, &tryFrame[int]{position: tryPosition(f.Position), hasFinally: f.HasFinally, catchEntry: block(f.Catch), finallyEntry: block(f.Finally), thrownForked: f.Forked, thrownAny: f.Any})
	}
	row.Index = b.throwFrame()
	row.Target = -1
	if row.Index >= 0 {
		row.Target = adamicBlockID(throwTarget(b.tryStack[row.Index]))
	}
	adamicTrace = ""
	for i := 0; i < repeats; i++ {
		b.firstThrowableFork()
	}
	var out strings.Builder
	fmt.Fprintf(&out, "%d:%t|", b.cur.Index(), b.cur.Reachable)
	for _, f := range b.tryStack {
		fmt.Fprintf(&out, "%t:%t,", f.thrownForked, f.thrownAny)
	}
	out.WriteString("|" + adamicTrace + "\n")
	return row, out.String()
}
