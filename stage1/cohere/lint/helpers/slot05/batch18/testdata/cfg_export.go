package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
func adamicKind(n *ast.Node) ast.Kind {
	adamicTrace += fmt.Sprintf("kind:%d;", adamicID(n))
	if n == nil {
		return ast.KindUnknown
	}
	return n.Kind
}
func (b *Builder[E]) adamicRead(n *ast.Node) {
	adamicTrace += fmt.Sprintf("read:%d;", adamicID(n))
	b.read(n)
}
func (b *Builder[E]) adamicWrite(n *ast.Node) {
	adamicTrace += fmt.Sprintf("write:%d;", adamicID(n))
	b.write(n)
}
func adamicThrowable(n *ast.Node) bool {
	adamicTrace += fmt.Sprintf("throwable:%d;", adamicID(n))
	return isThrowableIdentifier(n)
}
func (b *Builder[E]) adamicFork() { adamicTrace += "fork;"; b.firstThrowableFork() }
func (b *Builder[E]) adamicExpr(n *ast.Node) {
	adamicTrace += fmt.Sprintf("expr:%d;", adamicID(n))
	outer := adamicTrace
	b.expr(n)
	adamicTrace = outer
}

type AdamicPatternNode struct {
	ID                             int
	Identifier, Wrapped, Throwable bool
	Expression                     int
}
type AdamicPatternRow struct {
	Root  int
	Nodes []AdamicPatternNode
}

func AdamicPatternObserve(mode string, n *ast.Node) (AdamicPatternRow, string) {
	adamicIDs = map[*ast.Node]int{nil: -1}
	row := AdamicPatternRow{Root: adamicID(n), Nodes: []AdamicPatternNode{{ID: -1, Expression: -1}}}
	for current := n; current != nil; {
		shape := AdamicPatternNode{ID: adamicID(current), Identifier: current.Kind == ast.KindIdentifier, Expression: -1}
		if shape.Identifier {
			shape.Throwable = isThrowableIdentifier(current)
		}
		var child *ast.Node
		switch current.Kind {
		case ast.KindParenthesizedExpression:
			shape.Wrapped = true
			child = current.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			shape.Wrapped = true
			child = current.AsNonNullExpression().Expression
		case ast.KindAsExpression:
			shape.Wrapped = true
			child = current.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			shape.Wrapped = true
			child = current.AsSatisfiesExpression().Expression
		case ast.KindTypeAssertionExpression:
			shape.Wrapped = true
			child = current.AsTypeAssertion().Expression
		}
		shape.Expression = adamicID(child)
		row.Nodes = append(row.Nodes, shape)
		if !shape.Wrapped {
			break
		}
		current = child
	}
	b := &Builder[int]{}
	b.cur = b.newBlock()
	b.cur.Reachable = true
	b.cur.hasIncoming = true
	adamicTrace = ""
	if mode == "reads" {
		b.patternReads(n)
	} else {
		b.patternWrites(n)
	}
	return row, adamicTrace + "\n"
}
