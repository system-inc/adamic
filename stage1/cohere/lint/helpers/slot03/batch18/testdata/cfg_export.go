package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type AdamicEvent struct {
	Kind  string `json:"kind"`
	A     int    `json:"a"`
	B     int    `json:"b"`
	C     int    `json:"c"`
	Value int    `json:"value"`
	Cur   int    `json:"cur"`
}
type AdamicBinary struct {
	Left              int  `json:"left"`
	Right             int  `json:"right"`
	ShortCircuit      bool `json:"shortCircuit"`
	LogicalAssignment bool `json:"logicalAssignment"`
	Equals            bool `json:"equals"`
	Assignment        bool `json:"assignment"`
}
type AdamicHeritage struct {
	Present      bool  `json:"present"`
	TypesPresent bool  `json:"typesPresent"`
	Types        []int `json:"types"`
}
type AdamicClass struct {
	Node            int              `json:"node"`
	Present         bool             `json:"present"`
	HeritagePresent bool             `json:"heritagePresent"`
	Heritages       []AdamicHeritage `json:"heritages"`
	MembersPresent  bool             `json:"membersPresent"`
	Members         []int            `json:"members"`
}
type AdamicIf struct {
	Expression int `json:"expression"`
	Then       int `json:"thenStatement"`
	Else       int `json:"elseStatement"`
}
type AdamicCase struct {
	Op     string        `json:"op"`
	Cur    int           `json:"cur"`
	Binary AdamicBinary  `json:"binary"`
	Class  AdamicClass   `json:"class"`
	If     AdamicIf      `json:"if"`
	Events []AdamicEvent `json:"events"`
}

var AdamicCases []AdamicCase
var AdamicOutputs []string
var adamicActive bool
var adamicDepth int
var adamicEvents []AdamicEvent
var adamicTrace string
var adamicCursor func() int
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
func adamicRecord(kind string, a, b, c, before, cur, value int) {
	adamicEvents = append(adamicEvents, AdamicEvent{kind, a, b, c, value, cur})
	adamicTrace += fmt.Sprintf("%s:%d:%d:%d:%d>%d:%d;", kind, a, b, c, before, cur, value)
}
func adamicEvent[E any](b *Builder[E], kind string, a, barg, c, before, value int) {
	if adamicActive && adamicDepth == 0 {
		adamicRecord(kind, a, barg, c, before, adamicBlockID(b.cur), value)
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
	adamicEvent(b, "expr", adamicNodeID(node), -1, -1, before, -1)
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
	adamicEvent(b, "bind", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) patternReads(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawPatternReads(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "reads", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) patternWrites(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawPatternWrites(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "writes", adamicNodeID(node), -1, -1, before, -1)
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
	adamicEvent(b, "decorators", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) typeParameters(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawTypeParameters(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "typeParameters", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) memberHeader(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawMemberHeader(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "memberHeader", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) statement(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawStatement(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "statement", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) condition(node *ast.Node, whenTrue, whenFalse *Block[E]) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawCondition(node, whenTrue, whenFalse)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "condition", adamicNodeID(node), adamicBlockID(whenTrue), adamicBlockID(whenFalse), before, -1)
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
	adamicEvent(b, "new", -1, -1, -1, before, adamicBlockID(v))
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
	adamicEvent(b, "link", adamicBlockID(from), adamicBlockID(to), -1, before, -1)
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
	adamicEvent(b, "enter", adamicBlockID(block), -1, -1, before, -1)
}
func isDestructuringTarget(node *ast.Node) bool {
	v := adamicRawIsDestructuringTarget(node)
	if adamicActive && adamicDepth == 0 {
		value := 0
		if v {
			value = 1
		}
		cur := adamicCursor()
		adamicRecord("destructure", adamicNodeID(node), -1, -1, cur, cur, value)
	}
	return v
}
func adamicObserve[E any](b *Builder[E], c AdamicCase, run func()) {
	if adamicActive || adamicDepth > 0 {
		run()
		return
	}
	c.Cur = adamicBlockID(b.cur)
	adamicTrace = ""
	adamicEvents = []AdamicEvent{}
	adamicCursor = func() int { return adamicBlockID(b.cur) }
	adamicActive = true
	run()
	adamicActive = false
	c.Events = adamicEvents
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, fmt.Sprintf("%d/%s", adamicBlockID(b.cur), adamicTrace))
}
func (b *Builder[E]) binaryExpression(node *ast.Node) {
	binary := node.AsBinaryExpression()
	op := binary.OperatorToken.Kind
	v := AdamicBinary{adamicNodeID(binary.Left), adamicNodeID(binary.Right), op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken || op == ast.KindQuestionQuestionToken, ast.IsLogicalOrCoalescingAssignmentOperator(op), op == ast.KindEqualsToken, ast.IsAssignmentOperator(op)}
	adamicObserve(b, AdamicCase{Op: "binary", Binary: v}, func() { b.adamicRawBinaryExpression(node) })
}
func adamicClassView(node *ast.Node) AdamicClass {
	v := AdamicClass{Node: adamicNodeID(node), Heritages: []AdamicHeritage{}, Members: []int{}}
	class := node.ClassLikeData()
	if class == nil {
		return v
	}
	v.Present = true
	v.HeritagePresent = class.HeritageClauses != nil
	v.MembersPresent = class.Members != nil
	if class.HeritageClauses != nil {
		for _, clause := range class.HeritageClauses.Nodes {
			heritage := clause.AsHeritageClause()
			h := AdamicHeritage{Types: []int{}}
			h.Present = heritage != nil
			if heritage != nil {
				h.TypesPresent = heritage.Types != nil
				if heritage.Types != nil {
					for _, typ := range heritage.Types.Nodes {
						h.Types = append(h.Types, adamicNodeID(typ))
					}
				}
			}
			v.Heritages = append(v.Heritages, h)
		}
	}
	if class.Members != nil {
		for _, member := range class.Members.Nodes {
			v.Members = append(v.Members, adamicNodeID(member))
		}
	}
	return v
}
func (b *Builder[E]) classLike(node *ast.Node) {
	adamicObserve(b, AdamicCase{Op: "class", Class: adamicClassView(node)}, func() { b.adamicRawClassLike(node) })
}
func (b *Builder[E]) ifStatement(node *ast.Node) {
	stmt := node.AsIfStatement()
	v := AdamicIf{adamicNodeID(stmt.Expression), adamicNodeID(stmt.ThenStatement), adamicNodeID(stmt.ElseStatement)}
	adamicObserve(b, AdamicCase{Op: "if", If: v}, func() { b.adamicRawIfStatement(node) })
}

// Absent pointer fields leave their flattened payload ignored, including parked arena entries.
func AdamicClassControls(root *ast.Node) {
	var found *ast.Node
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n.ClassLikeData() != nil {
			found = n
			return true
		}
		return n.ForEachChild(walk)
	}
	walk(root)
	if found == nil {
		panic("class control needs a parsed class")
	}
	saved := adamicClassView(found)
	data := found.ClassLikeData()
	heritages, members := data.HeritageClauses, data.Members
	builder := func() *Builder[int] { b := &Builder[int]{}; b.cur = b.newBlock(); b.cur.Reachable = true; return b }
	data.HeritageClauses = nil
	data.Members = nil
	builder().classLike(found)
	c := &AdamicCases[len(AdamicCases)-1]
	c.Class.Heritages = saved.Heritages
	c.Class.Members = saved.Members
	data.HeritageClauses = heritages
	data.Members = members
	builder().classLike(root)
	c = &AdamicCases[len(AdamicCases)-1]
	c.Class.Heritages = saved.Heritages
	c.Class.Members = saved.Members
	// A real heritage clause with nil Types is skipped by Go.
	if len(heritages.Nodes) > 0 {
		original := heritages.Nodes
		heritage := original[0].AsHeritageClause()
		types := heritage.Types
		heritage.Types = nil
		builder().classLike(found)
		c = &AdamicCases[len(AdamicCases)-1]
		for i := range c.Class.Heritages {
			if !c.Class.Heritages[i].Present || !c.Class.Heritages[i].TypesPresent {
				c.Class.Heritages[i].Types = saved.Heritages[0].Types
			}
		}
		heritage.Types = types
		heritages.Nodes = original
	}
}
