package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type AdamicEvent struct {
	Kind  string `json:"kind"`
	A     int    `json:"a"`
	B     int    `json:"b"`
	Value int    `json:"value"`
	Cur   int    `json:"cur"`
}
type AdamicMember struct {
	Node               int   `json:"node"`
	FunctionLike       bool  `json:"functionLike"`
	Parameters         []int `json:"parameters"`
	Computed           bool  `json:"computed"`
	ComputedExpression int   `json:"computedExpression"`
	Property           bool  `json:"property"`
	IndexSignature     bool  `json:"indexSignature"`
	ParameterTypes     []int `json:"parameterTypes"`
	TypeNode           int   `json:"typeNode"`
}
type AdamicCase struct {
	Op        string        `json:"op"`
	Cur       int           `json:"cur"`
	Condition int           `json:"condition"`
	WhenTrue  int           `json:"whenTrue"`
	WhenFalse int           `json:"whenFalse"`
	Target    int           `json:"target"`
	Fallback  int           `json:"fallback"`
	Member    AdamicMember  `json:"member"`
	Events    []AdamicEvent `json:"events"`
}

var AdamicCases []AdamicCase
var AdamicOutputs []string
var adamicActive bool
var adamicDepth int
var adamicEvents []AdamicEvent
var adamicTrace string
var adamicIDs = map[*ast.Node]int{nil: -1}

func adamicNodeID(n *ast.Node) int {
	if v, ok := adamicIDs[n]; ok {
		return v
	}
	v := len(adamicIDs) - 1
	adamicIDs[n] = v
	return v
}
func adamicBlockID[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	return int(b.index)
}
func adamicEvent[E any](b *Builder[E], kind string, a, barg, before, value int) {
	if adamicActive && adamicDepth == 0 {
		cur := adamicBlockID(b.cur)
		adamicEvents = append(adamicEvents, AdamicEvent{kind, a, barg, value, cur})
		adamicTrace += fmt.Sprintf("%s:%d:%d:%d>%d:%d;", kind, a, barg, before, cur, value)
	}
}
func (b *Builder[E]) expr(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawExpr(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "expr", adamicNodeID(node), -1, before, -1)
}
func (b *Builder[E]) patternBind(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawPatternBind(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "bind", adamicNodeID(node), -1, before, -1)
}
func (b *Builder[E]) decorators(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawDecorators(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "decorators", adamicNodeID(node), -1, before, -1)
}
func (b *Builder[E]) newBlock() *Block[E] {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	v := b.adamicRawNewBlock()
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "new", -1, -1, before, adamicBlockID(v))
	return v
}
func (b *Builder[E]) link(from, to *Block[E]) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawLink(from, to)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "link", adamicBlockID(from), adamicBlockID(to), before, -1)
}
func (b *Builder[E]) enter(block *Block[E]) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawEnter(block)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "enter", adamicBlockID(block), -1, before, -1)
}
func adamicObserve[E any](b *Builder[E], c AdamicCase, run func()) {
	if adamicActive || adamicDepth > 0 {
		run()
		return
	}
	c.Cur = adamicBlockID(b.cur)
	adamicTrace = ""
	adamicEvents = []AdamicEvent{}
	adamicActive = true
	run()
	adamicActive = false
	c.Events = adamicEvents
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, fmt.Sprintf("%d/%s", adamicBlockID(b.cur), adamicTrace))
}
func (b *Builder[E]) conditionalExpression(node *ast.Node) {
	cond := node.AsConditionalExpression()
	c := AdamicCase{Op: "conditional", Condition: adamicNodeID(cond.Condition), WhenTrue: adamicNodeID(cond.WhenTrue), WhenFalse: adamicNodeID(cond.WhenFalse)}
	adamicObserve(b, c, func() { b.adamicRawConditionalExpression(node) })
}
func (b *Builder[E]) bindWithDefault(target, fallback *ast.Node) {
	c := AdamicCase{Op: "default", Target: adamicNodeID(target), Fallback: adamicNodeID(fallback)}
	adamicObserve(b, c, func() { b.adamicRawBindWithDefault(target, fallback) })
}
func (b *Builder[E]) memberHeader(node *ast.Node) {
	m := AdamicMember{Node: adamicNodeID(node), FunctionLike: ast.IsFunctionLikeDeclaration(node), Parameters: []int{}, ComputedExpression: -1, ParameterTypes: []int{}, TypeNode: -1}
	if m.FunctionLike {
		for _, p := range node.Parameters() {
			m.Parameters = append(m.Parameters, adamicNodeID(p))
		}
	}
	if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
		m.Computed = true
		m.ComputedExpression = adamicNodeID(name.AsComputedPropertyName().Expression)
	}
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		m.Property = true
		m.TypeNode = adamicNodeID(node.AsPropertyDeclaration().Type)
	case ast.KindIndexSignature:
		m.IndexSignature = true
		for _, p := range node.Parameters() {
			m.ParameterTypes = append(m.ParameterTypes, adamicNodeID(p.Type()))
		}
		m.TypeNode = adamicNodeID(node.Type())
	}
	adamicObserve(b, AdamicCase{Op: "member", Member: m}, func() { b.adamicRawMemberHeader(node) })
}
