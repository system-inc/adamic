package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

var adamicTrace string

func adamicID(n *ast.Node) int {
	if n == nil {
		return -1
	}
	return 1
}
func adamicIsStatement(n *ast.Node) bool {
	adamicTrace += fmt.Sprintf("classify:%d;", adamicID(n))
	return ast.IsStatement(n)
}
func (b *Builder[E]) adamicStatement(n *ast.Node) {
	adamicTrace += fmt.Sprintf("stmt:%d;", adamicID(n))
	outer := adamicTrace
	b.statement(n)
	adamicTrace = outer
}
func (b *Builder[E]) adamicExpr(n *ast.Node) {
	adamicTrace += fmt.Sprintf("expr:%d;", adamicID(n))
	outer := adamicTrace
	b.expr(n)
	adamicTrace = outer
}

type AdamicVisitRow struct {
	Node      int
	Statement bool
}

func AdamicVisitObserve(n *ast.Node) (AdamicVisitRow, string) {
	b := &Builder[int]{}
	b.cur = b.newBlock()
	b.cur.Reachable = true
	b.cur.hasIncoming = true
	row := AdamicVisitRow{Node: adamicID(n), Statement: n != nil && ast.IsStatement(n)}
	adamicTrace = ""
	b.visitUnknown(n)
	return row, adamicTrace + "\n"
}
func adamicIsChainRoot(n *ast.Node) bool {
	adamicTrace += fmt.Sprintf("root:%d;", adamicID(n))
	return ast.IsOptionalChainRoot(n)
}
func (b *Builder[E]) adamicLink(from, to *Block[E]) {
	adamicTrace += fmt.Sprintf("link:%d:%d;", from.Index(), to.Index())
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

type AdamicChainRow struct {
	Root, Reachable bool
	Depth           int
}

func AdamicChainObserve(n *ast.Node, depth int, reachable bool) (AdamicChainRow, string) {
	b := &Builder[int]{}
	for i := 0; i < 4; i++ {
		b.newBlock()
	}
	b.cur = b.blocks[0]
	b.cur.Reachable = reachable
	for i := 0; i < depth; i++ {
		b.chainJoins = append(b.chainJoins, b.blocks[1+i])
	}
	row := AdamicChainRow{Root: ast.IsOptionalChainRoot(n), Reachable: reachable, Depth: depth}
	adamicTrace = ""
	b.forkOptionalChain(n)
	var out strings.Builder
	fmt.Fprintf(&out, "%d:%t|", b.cur.Index(), b.cur.Reachable)
	for _, j := range b.chainJoins {
		fmt.Fprintf(&out, "%d,", j.Index())
	}
	out.WriteString("|" + adamicTrace + "\n")
	return row, out.String()
}
