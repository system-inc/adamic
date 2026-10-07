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
	Jumps []bool `json:"jumps"`
}
type AdamicExpression struct {
	Node        int    `json:"node"`
	Present     bool   `json:"present"`
	HookPresent bool   `json:"hookPresent"`
	Category    string `json:"category"`
	Operand     int    `json:"operand"`
	Update      bool   `json:"update"`
	Children    []int  `json:"children"`
}
type AdamicElement struct {
	Present  bool `json:"present"`
	Computed bool `json:"computed"`
	Key      int  `json:"key"`
	Target   int  `json:"target"`
	Fallback int  `json:"fallback"`
}
type AdamicProperty struct {
	Category string `json:"category"`
	Computed bool   `json:"computed"`
	Key      int    `json:"key"`
	Target   int    `json:"target"`
	Fallback int    `json:"fallback"`
}
type AdamicPattern struct {
	Node            int              `json:"node"`
	Present         bool             `json:"present"`
	Category        string           `json:"category"`
	Operand         int              `json:"operand"`
	ElementsPresent bool             `json:"elementsPresent"`
	Elements        []AdamicElement  `json:"elements"`
	Properties      []AdamicProperty `json:"properties"`
	Array           []int            `json:"array"`
	Equals          bool             `json:"equals"`
	Left            int              `json:"left"`
	Right           int              `json:"right"`
}
type AdamicClause struct {
	IsDefault  bool `json:"isDefault"`
	Expression int  `json:"expression"`
	Statements int  `json:"statements"`
}
type AdamicSwitch struct {
	Node       int            `json:"node"`
	Expression int            `json:"expression"`
	Clauses    []AdamicClause `json:"clauses"`
}
type AdamicCase struct {
	Op         string           `json:"op"`
	Cur        int              `json:"cur"`
	Jumps      []bool           `json:"jumps"`
	Expression AdamicExpression `json:"expression"`
	Pattern    AdamicPattern    `json:"pattern"`
	Switch     AdamicSwitch     `json:"switch"`
	Events     []AdamicEvent    `json:"events"`
}

var AdamicCases []AdamicCase
var AdamicOutputs []string
var adamicActive bool
var adamicDepth int
var adamicEvents []AdamicEvent
var adamicTrace string
var adamicCursor func() int
var adamicJumps func() []bool
var adamicLists = map[*ast.NodeList]int{nil: -1}

func adamicListID(n *ast.NodeList) int {
	if v, ok := adamicLists[n]; ok {
		return v
	}
	v := len(adamicLists) - 1
	adamicLists[n] = v
	return v
}
func adamicJumpView[E any](b *Builder[E]) []bool {
	out := []bool{}
	for _, j := range b.jumps {
		out = append(out, j.broken)
	}
	return out
}

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
	adamicEvents = append(adamicEvents, AdamicEvent{kind, a, b, c, value, cur, adamicJumps()})
	adamicTrace += fmt.Sprintf("%s:%d:%d:%d:%d>%d:%d;", kind, a, b, c, before, cur, value)
}
func adamicEvent[E any](b *Builder[E], kind string, a, barg, c, before, value int) {
	if adamicActive && adamicDepth == 0 {
		adamicRecord(kind, a, barg, c, before, adamicBlockID(b.cur), value)
	}
}
func adamicBool(v bool) int {
	if v {
		return 1
	}
	return 0
}
func (b *Builder[E]) read(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawRead(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "read", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) write(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawWrite(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "write", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) firstThrowableFork() {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawFirstThrowableFork()
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "fork", -1, -1, -1, before, -1)
}
func (b *Builder[E]) binaryExpression(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawBinaryExpression(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "binary", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) conditionalExpression(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawConditionalExpression(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "conditional", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) updateExpression(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawUpdateExpression(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "update", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) accessOrCall(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawAccessOrCall(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "access", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) makeYield() {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawMakeYield()
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "yield", -1, -1, -1, before, -1)
}
func (b *Builder[E]) classLike(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawClassLike(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "class", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) nestedFunction(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawNestedFunction(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "nested", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) visitUnknown(node *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawVisitUnknown(node)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "visit", adamicNodeID(node), -1, -1, before, -1)
}
func (b *Builder[E]) bindWithDefault(target, fallback *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawBindWithDefault(target, fallback)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "default", adamicNodeID(target), adamicNodeID(fallback), -1, before, -1)
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
func (b *Builder[E]) linkWithCycleBarrier(from, to *Block[E], barrier bool) int {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	v := b.adamicRawLinkWithCycleBarrier(from, to, barrier)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "barrierLink", adamicBlockID(from), adamicBlockID(to), adamicBool(barrier), before, v)
	return v
}
func (b *Builder[E]) setCycleBarrier(from *Block[E], index int, barrier bool) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawSetCycleBarrier(from, index, barrier)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "setBarrier", adamicBlockID(from), index, adamicBool(barrier), before, -1)
}
func (b *Builder[E]) pushJump(node *ast.Node, breakTo, continueTo *Block[E], loop *ast.Node) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawPushJump(node, breakTo, continueTo, loop)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "push", adamicNodeID(node), adamicBlockID(breakTo), adamicBlockID(continueTo), before, -1)
}
func (b *Builder[E]) popJump() {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawPopJump()
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "pop", -1, -1, -1, before, -1)
}
func (b *Builder[E]) statements(list *ast.NodeList) {
	before := adamicBlockID(b.cur)
	suppress := adamicActive
	if suppress {
		adamicDepth++
	}
	b.adamicRawStatements(list)
	if suppress {
		adamicDepth--
	}
	adamicEvent(b, "statements", adamicListID(list), -1, -1, before, -1)
}
func isThrowableIdentifier(node *ast.Node) bool {
	v := adamicRawisThrowableIdentifier(node)
	if adamicActive && adamicDepth == 0 {
		cur := adamicCursor()
		adamicRecord("throwable", adamicNodeID(node), -1, -1, cur, cur, adamicBool(v))
	}
	return v
}
func IsRoot(node *ast.Node) bool {
	v := adamicRawIsRoot(node)
	if adamicActive && adamicDepth == 0 {
		cur := adamicCursor()
		adamicRecord("root", adamicNodeID(node), -1, -1, cur, cur, adamicBool(v))
	}
	return v
}
func adamicObserve[E any](b *Builder[E], c AdamicCase, run func()) {
	if adamicActive || adamicDepth > 0 {
		run()
		return
	}
	c.Cur = adamicBlockID(b.cur)
	c.Jumps = adamicJumpView(b)
	adamicTrace = ""
	adamicEvents = []AdamicEvent{}
	adamicCursor = func() int { return adamicBlockID(b.cur) }
	adamicJumps = func() []bool { return adamicJumpView(b) }
	adamicActive = true
	run()
	adamicActive = false
	c.Events = adamicEvents
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, fmt.Sprintf("%d/%s", adamicBlockID(b.cur), adamicTrace))
}
func AdamicHooks() Hooks[int] {
	return Hooks[int]{Expression: func(b *Builder[int], node *ast.Node) {
		before := adamicBlockID(b.cur)
		adamicEvent(b, "hook", adamicNodeID(node), -1, -1, before, -1)
	}}
}
func adamicExpressionView[E any](b *Builder[E], node *ast.Node) AdamicExpression {
	v := AdamicExpression{Node: adamicNodeID(node), Operand: -1, Category: "other", Children: []int{}, HookPresent: b.hooks.Expression != nil}
	if node == nil {
		return v
	}
	v.Present = true
	switch node.Kind {
	case ast.KindIdentifier:
		v.Category = "identifier"
	case ast.KindParenthesizedExpression:
		v.Category = "parenthesized"
		v.Operand = adamicNodeID(node.AsParenthesizedExpression().Expression)
	case ast.KindBinaryExpression:
		v.Category = "binary"
	case ast.KindConditionalExpression:
		v.Category = "conditional"
	case ast.KindPrefixUnaryExpression:
		v.Category = "prefix"
		u := node.AsPrefixUnaryExpression()
		v.Operand = adamicNodeID(u.Operand)
		v.Update = u.Operator == ast.KindPlusPlusToken || u.Operator == ast.KindMinusMinusToken
	case ast.KindPostfixUnaryExpression:
		v.Category = "postfix"
		v.Operand = adamicNodeID(node.AsPostfixUnaryExpression().Operand)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		v.Category = "access"
	case ast.KindYieldExpression:
		v.Category = "yield"
		v.Operand = adamicNodeID(node.AsYieldExpression().Expression)
	case ast.KindClassExpression:
		v.Category = "class"
	}
	node.ForEachChild(func(child *ast.Node) bool { v.Children = append(v.Children, adamicNodeID(child)); return false })
	return v
}
func adamicComputed(node *ast.Node) (bool, int) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		return true, adamicNodeID(node.AsComputedPropertyName().Expression)
	}
	return false, -1
}
func adamicPatternView(node *ast.Node) AdamicPattern {
	v := AdamicPattern{Node: adamicNodeID(node), Category: "other", Operand: -1, Left: -1, Right: -1, Elements: []AdamicElement{}, Properties: []AdamicProperty{}, Array: []int{}}
	if node == nil {
		return v
	}
	v.Present = true
	switch node.Kind {
	case ast.KindIdentifier:
		v.Category = "identifier"
	case ast.KindOmittedExpression:
		v.Category = "omitted"
	case ast.KindParenthesizedExpression:
		v.Category = "wrapped"
		v.Operand = adamicNodeID(node.AsParenthesizedExpression().Expression)
	case ast.KindNonNullExpression:
		v.Category = "wrapped"
		v.Operand = adamicNodeID(node.AsNonNullExpression().Expression)
	case ast.KindAsExpression:
		v.Category = "wrapped"
		v.Operand = adamicNodeID(node.AsAsExpression().Expression)
	case ast.KindSatisfiesExpression:
		v.Category = "wrapped"
		v.Operand = adamicNodeID(node.AsSatisfiesExpression().Expression)
	case ast.KindTypeAssertionExpression:
		v.Category = "wrapped"
		v.Operand = adamicNodeID(node.AsTypeAssertion().Expression)
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		v.Category = "binding"
		pattern := node.AsBindingPattern()
		v.ElementsPresent = pattern.Elements != nil
		if pattern.Elements != nil {
			for _, e := range pattern.Elements.Nodes {
				binding := e.AsBindingElement()
				item := AdamicElement{Key: -1, Target: -1, Fallback: -1}
				if binding != nil {
					item.Present = true
					item.Computed, item.Key = adamicComputed(binding.PropertyName)
					item.Target = adamicNodeID(binding.Name())
					item.Fallback = adamicNodeID(binding.Initializer)
				}
				v.Elements = append(v.Elements, item)
			}
		}
	case ast.KindObjectLiteralExpression:
		v.Category = "object"
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			item := AdamicProperty{Category: "other", Key: -1, Target: -1, Fallback: -1}
			switch property.Kind {
			case ast.KindPropertyAssignment:
				a := property.AsPropertyAssignment()
				item.Category = "assignment"
				item.Computed, item.Key = adamicComputed(a.Name())
				item.Target = adamicNodeID(a.Initializer)
			case ast.KindShorthandPropertyAssignment:
				a := property.AsShorthandPropertyAssignment()
				item.Category = "shorthand"
				item.Target = adamicNodeID(a.Name())
				item.Fallback = adamicNodeID(a.ObjectAssignmentInitializer)
			case ast.KindSpreadAssignment:
				item.Category = "spread"
				item.Target = adamicNodeID(property.AsSpreadAssignment().Expression)
			}
			v.Properties = append(v.Properties, item)
		}
	case ast.KindArrayLiteralExpression:
		v.Category = "array"
		for _, e := range node.AsArrayLiteralExpression().Elements.Nodes {
			v.Array = append(v.Array, adamicNodeID(e))
		}
	case ast.KindSpreadElement:
		v.Category = "spread"
		v.Operand = adamicNodeID(node.AsSpreadElement().Expression)
	case ast.KindBinaryExpression:
		v.Category = "binary"
		a := node.AsBinaryExpression()
		v.Equals = a.OperatorToken.Kind == ast.KindEqualsToken
		v.Left = adamicNodeID(a.Left)
		v.Right = adamicNodeID(a.Right)
	}
	return v
}
func adamicSwitchView(node *ast.Node) AdamicSwitch {
	s := node.AsSwitchStatement()
	v := AdamicSwitch{Node: adamicNodeID(node), Expression: adamicNodeID(s.Expression), Clauses: []AdamicClause{}}
	if s.CaseBlock != nil {
		if block := s.CaseBlock.AsCaseBlock(); block != nil && block.Clauses != nil {
			for _, c := range block.Clauses.Nodes {
				data := c.AsCaseOrDefaultClause()
				v.Clauses = append(v.Clauses, AdamicClause{c.Kind == ast.KindDefaultClause, adamicNodeID(data.Expression), adamicListID(data.Statements)})
			}
		}
	}
	return v
}
func (b *Builder[E]) expr(node *ast.Node) {
	if adamicActive {
		before := adamicBlockID(b.cur)
		adamicDepth++
		b.adamicRawExpr(node)
		adamicDepth--
		adamicEvent(b, "expr", adamicNodeID(node), -1, -1, before, -1)
		return
	}
	adamicObserve(b, AdamicCase{Op: "expression", Expression: adamicExpressionView(b, node)}, func() { b.adamicRawExpr(node) })
}
func (b *Builder[E]) patternBind(node *ast.Node) {
	if adamicActive {
		before := adamicBlockID(b.cur)
		adamicDepth++
		b.adamicRawPatternBind(node)
		adamicDepth--
		adamicEvent(b, "bind", adamicNodeID(node), -1, -1, before, -1)
		return
	}
	adamicObserve(b, AdamicCase{Op: "pattern", Pattern: adamicPatternView(node)}, func() { b.adamicRawPatternBind(node) })
}
func (b *Builder[E]) switchStatement(node *ast.Node) {
	if adamicActive {
		before := adamicBlockID(b.cur)
		adamicDepth++
		b.adamicRawSwitchStatement(node)
		adamicDepth--
		adamicEvent(b, "switch", adamicNodeID(node), -1, -1, before, -1)
		return
	}
	adamicObserve(b, AdamicCase{Op: "switch", Switch: adamicSwitchView(node)}, func() { b.adamicRawSwitchStatement(node) })
}

// Typed nil-list controls preserve parked payloads behind absent fields.
func AdamicControls(root *ast.Node) {
	b := &Builder[int]{hooks: AdamicHooks()}
	b.cur = b.newBlock()
	b.cur.Reachable = true
	b.expr(nil)
	b.patternBind(nil)
	var pattern, switchNode *ast.Node
	all := []*ast.Node{}
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		all = append(all, n)
		if n.Kind == ast.KindObjectBindingPattern && pattern == nil {
			pattern = n
		}
		if n.Kind == ast.KindSwitchStatement && switchNode == nil {
			switchNode = n
		}
		n.ForEachChild(walk)
		return false
	}
	walk(root)
	if pattern == nil || switchNode == nil {
		panic("control lacks pattern or switch")
	}
	saved := adamicPatternView(pattern)
	data := pattern.AsBindingPattern()
	elements := data.Elements
	data.Elements = nil
	b.patternBind(pattern)
	AdamicCases[len(AdamicCases)-1].Pattern.Elements = saved.Elements
	data.Elements = elements
	// The typed nil binding is accepted by the original Go guard. Parked fields
	// make the mutant that ignores that guard produce observable operations.
	copyNode := *elements.Nodes[0]
	copyNode.AdamicNilBindingData()
	copyList := *elements
	copyList.Nodes = append([]*ast.Node(nil), elements.Nodes...)
	copyList.Nodes[0] = &copyNode
	data.Elements = &copyList
	b.patternBind(pattern)
	parked := saved.Elements[0]
	parked.Present = false
	parked.Computed = true
	parked.Key = parked.Target
	AdamicCases[len(AdamicCases)-1].Pattern.Elements[0] = parked
	data.Elements = elements
	block := switchNode.AsSwitchStatement().CaseBlock
	switchNode.AsSwitchStatement().CaseBlock = nil
	b.switchStatement(switchNode)
	switchNode.AsSwitchStatement().CaseBlock = block
	b.hooks = Hooks[int]{}
	b.expr(root)
	// Observe dispatch that is normally nested beneath a parent expression's trace.
	for _, n := range all {
		c := &Builder[int]{hooks: AdamicHooks()}
		c.cur = c.newBlock()
		c.cur.Reachable = true
		c.patternBind(n)
		c.expr(n)
	}
}
