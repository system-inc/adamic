package control_flow_graph

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"slices"
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
	Joins  []int
}

func (b *Builder[E]) adamicSnapshot() AdamicSnapshot {
	s := AdamicSnapshot{Cur: b.cur.Index(), Blocks: []AdamicBlockRow{}, Broken: []bool{}, Joins: []int{}}
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
	for _, j := range b.chainJoins {
		s.Joins = append(s.Joins, j.Index())
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
	adamicEffects[fmt.Sprintf("expr:%d", adamicID(n))] = b.adamicSnapshot()
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

func (b *Builder[E]) adamicDisconnected(n *Block[E]) {
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("disconnected:%d:%t;", n.Index(), b.cur.Reachable)
	}
	b.enterDisconnected(n)
}
func adamicDeclaration(n *ast.Node) bool {
	v := n.Kind == ast.KindVariableDeclarationList
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("declaration:%d:%t;", adamicID(n), v)
	}
	return v
}
func adamicNames(n *ast.Node) {
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("names:%d;", adamicID(n))
	}
}
func (b *Builder[E]) adamicVariables(n *ast.Node) {
	if adamicMuted {
		b.variableDeclarationList(n)
		return
	}
	adamicTrace += fmt.Sprintf("variables:%d:%d;", adamicID(n), b.cur.Index())
	adamicMuted = true
	b.variableDeclarationList(n)
	adamicMuted = false
	adamicEffects["variables"] = b.adamicSnapshot()
}
func (b *Builder[E]) adamicBind(n *ast.Node) {
	if adamicMuted {
		b.patternBind(n)
		return
	}
	adamicTrace += fmt.Sprintf("bind:%d:%d;", adamicID(n), b.cur.Index())
	adamicMuted = true
	b.patternBind(n)
	adamicMuted = false
	adamicEffects[fmt.Sprintf("bind:%d", adamicID(n))] = b.adamicSnapshot()
}
func adamicOptional(n *ast.Node) bool {
	v := ast.IsOptionalChain(n)
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("optional:%d:%t;", adamicID(n), v)
	}
	return v
}
func adamicOutermost(n *ast.Node) bool {
	v := ast.IsOutermostOptionalChain(n)
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("outermost:%d:%t;", adamicID(n), v)
	}
	return v
}
func adamicKind(n *ast.Node) ast.Kind {
	if !adamicMuted {
		adamicTrace += fmt.Sprintf("project:%d;", adamicID(n))
	}
	return n.Kind
}
func (b *Builder[E]) adamicForkOptional(n *ast.Node) {
	if adamicMuted {
		b.forkOptionalChain(n)
		return
	}
	adamicTrace += fmt.Sprintf("optionalFork:%d:%d:%d;", adamicID(n), b.cur.Index(), len(b.chainJoins))
	adamicMuted = true
	b.forkOptionalChain(n)
	adamicMuted = false
	adamicEffects["optionalFork"] = b.adamicSnapshot()
}
func (b *Builder[E]) adamicTypes(n *ast.Node) {
	if adamicMuted {
		b.typeArguments(n)
		return
	}
	adamicTrace += fmt.Sprintf("types:%d:%d;", adamicID(n), b.cur.Index())
	adamicMuted = true
	b.typeArguments(n)
	adamicMuted = false
	adamicEffects["types"] = b.adamicSnapshot()
}
func (b *Builder[E]) adamicThrowable() {
	if adamicMuted {
		b.firstThrowableFork()
		return
	}
	adamicTrace += fmt.Sprintf("throwable:%d:%d;", b.cur.Index(), len(b.chainJoins))
	adamicMuted = true
	b.firstThrowableFork()
	adamicMuted = false
	adamicEffects["throwable"] = b.adamicSnapshot()
}

type AdamicRow struct {
	Node, Initializer, Condition, Incrementor, Expression, Argument, Template, Statement, Prefix int
	Truthy, Declaration, Optional, Outermost                                                     bool
	Property, Element, Call, Construct, Tagged                                                   bool
	Arguments, Names                                                                             []int
	Labels                                                                                       []string
	Start                                                                                        AdamicSnapshot
	Effects                                                                                      map[string]AdamicSnapshot
}

func AdamicObserve(mode string, n *ast.Node, reachable bool, prefix, chainDepth int, handler bool) (AdamicRow, string) {
	b := &Builder[int]{}
	b.cur = b.newBlock()
	b.cur.Reachable = reachable
	b.cur.hasIncoming = reachable
	adamicIDs = map[*ast.Node]int{}
	adamicID(n)
	if handler {
		target := b.newBlock()
		b.tryStack = append(b.tryStack, &tryFrame[int]{position: posTry, catchEntry: target})
	}
	for i := 0; i < chainDepth; i++ {
		b.chainJoins = append(b.chainJoins, b.newBlock())
	}
	for i := 0; i < prefix; i++ {
		b.jumps = append(b.jumps, jumpTarget[int]{labels: []string{fmt.Sprintf("seed%d", i)}, breakTo: b.cur, continueTo: b.cur, loop: n, breakable: i%2 == 0, broken: i%2 == 1})
	}
	adamicTrace = ""
	adamicMuted = false
	adamicEffects = map[string]AdamicSnapshot{}
	r := AdamicRow{Node: adamicID(n), Initializer: -1, Condition: -1, Incrementor: -1, Expression: -1, Argument: -1, Template: -1, Statement: -1, Prefix: prefix, Start: b.adamicSnapshot(), Labels: labelsOf(n), Arguments: []int{}, Names: []int{}}
	var init, expr, stmt *ast.Node
	switch mode {
	case "for":
		s := n.AsForStatement()
		init, stmt = s.Initializer, s.Statement
		r.Condition = adamicID(s.Condition)
		r.Incrementor = adamicID(s.Incrementor)
		r.Truthy = s.Condition != nil && isAlwaysTruthyTest(s.Condition)
	case "inof":
		s := n.AsForInOrOfStatement()
		init, expr, stmt = s.Initializer, s.Expression, s.Statement
	case "access":
		r.Optional = ast.IsOptionalChain(n)
		r.Outermost = ast.IsOutermostOptionalChain(n)
		switch n.Kind {
		case ast.KindPropertyAccessExpression:
			r.Property = true
			expr = n.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			r.Element = true
			s := n.AsElementAccessExpression()
			expr = s.Expression
			r.Argument = adamicID(s.ArgumentExpression)
		case ast.KindCallExpression:
			r.Call = true
			s := n.AsCallExpression()
			expr = s.Expression
			if s.Arguments != nil {
				for _, a := range s.Arguments.Nodes {
					r.Arguments = append(r.Arguments, adamicID(a))
				}
			}
		case ast.KindNewExpression:
			r.Construct = true
			s := n.AsNewExpression()
			expr = s.Expression
			if s.Arguments != nil {
				for _, a := range s.Arguments.Nodes {
					r.Arguments = append(r.Arguments, adamicID(a))
				}
			}
		case ast.KindTaggedTemplateExpression:
			r.Tagged = true
			s := n.AsTaggedTemplateExpression()
			expr = s.Tag
			r.Template = adamicID(s.Template)
		}
	}
	r.Initializer = adamicID(init)
	r.Expression = adamicID(expr)
	r.Statement = adamicID(stmt)
	if init != nil {
		r.Declaration = init.Kind == ast.KindVariableDeclarationList
		if r.Declaration {
			d := init.AsVariableDeclarationList()
			if d != nil && d.Declarations != nil {
				for _, x := range d.Declarations.Nodes {
					r.Names = append(r.Names, adamicID(x.Name()))
				}
			}
		}
	}
	switch mode {
	case "for":
		b.forStatement(n)
	case "inof":
		b.forInOfStatement(n)
	case "access":
		b.accessOrCall(n)
	}
	expectedJoins := append([]int{}, r.Start.Joins...)
	if mode == "access" && r.Optional && r.Outermost {
		expectedJoins = append(expectedJoins, len(r.Start.Blocks))
	}
	for key, snapshot := range adamicEffects {
		if !slices.Equal(snapshot.Joins, expectedJoins) {
			panic("dependency changed join stack: " + key)
		}
	}
	r.Effects = adamicEffects
	var out strings.Builder
	fmt.Fprintf(&out, "%d|%s|", b.cur.Index(), b.adamicStack())
	for _, j := range b.chainJoins {
		fmt.Fprintf(&out, "%d,", j.Index())
	}
	out.WriteString("|")
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

func (r AdamicBlockRow) MarshalJSON() ([]byte, error) {
	flags := "00"
	if r.Reachable {
		flags = "10"
	}
	if r.Incoming {
		flags = flags[:1] + "1"
	}
	var out strings.Builder
	out.WriteString(flags + "|")
	for _, e := range r.Edges {
		fmt.Fprintf(&out, "%d,", e)
	}
	return json.Marshal(out.String())
}
