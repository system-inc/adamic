package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strconv"
)

type AdamicFrame struct {
	Position     int   `json:"position"`
	HasFinally   bool  `json:"hasFinally"`
	CatchEntry   int   `json:"catchEntry"`
	FinallyEntry int   `json:"finallyEntry"`
	ThrownForked bool  `json:"thrownForked"`
	ThrownAny    bool  `json:"thrownAny"`
	Implicit     []int `json:"implicit"`
	ReturnedAny  bool  `json:"returnedAny"`
}
type AdamicState struct {
	Cur       int           `json:"cur"`
	Reachable bool          `json:"reachable"`
	Incoming  []int         `json:"incoming"`
	Frames    []AdamicFrame `json:"frames"`
	Stack     []int         `json:"stack"`
}
type AdamicEvent struct {
	Kind  string      `json:"kind"`
	A     int         `json:"a"`
	B     int         `json:"b"`
	Value int         `json:"value"`
	State AdamicState `json:"state"`
}
type AdamicTry struct {
	TryBlock       int  `json:"tryBlock"`
	HasCatch       bool `json:"hasCatch"`
	HasFinally     bool `json:"hasFinally"`
	BindingPresent bool `json:"bindingPresent"`
	Binding        int  `json:"binding"`
	CatchBlock     int  `json:"catchBlock"`
	FinallyBlock   int  `json:"finallyBlock"`
}
type AdamicCase struct {
	Op     string        `json:"op"`
	State  AdamicState   `json:"state"`
	Try    AdamicTry     `json:"try"`
	Events []AdamicEvent `json:"events"`
}

var AdamicCases []AdamicCase
var AdamicOutputs []string
var active bool
var depth int
var events []AdamicEvent
var trace string
var state func() AdamicState
var frameID func(any) int
var ids = map[*ast.Node]int{nil: -1}

func nodeID(n *ast.Node) int {
	if v, ok := ids[n]; ok {
		return v
	}
	v := len(ids) - 1
	ids[n] = v
	return v
}
func blockID[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	return int(b.index)
}
func bit(v bool) int {
	if v {
		return 1
	}
	return 0
}
func ints(v []int) string {
	out := ""
	for _, i := range v {
		out += strconv.Itoa(i) + ","
	}
	return out
}
func digest(s AdamicState) string {
	out := fmt.Sprintf("%d|%d|%s|%s|", s.Cur, bit(s.Reachable), ints(s.Incoming), ints(s.Stack))
	for _, f := range s.Frames {
		out += fmt.Sprintf("%d,%d,%d,%d,%d,%d,%s%d/", f.Position, bit(f.HasFinally), f.CatchEntry, f.FinallyEntry, bit(f.ThrownForked), bit(f.ThrownAny), ints(f.Implicit), bit(f.ReturnedAny))
	}
	return out
}
func before() string {
	if active && depth == 0 {
		return digest(state())
	}
	return ""
}
func record(kind string, a, b, value int, pre string) {
	if active && depth == 0 {
		s := state()
		events = append(events, AdamicEvent{kind, a, b, value, s})
		trace += fmt.Sprintf("%s:%d:%d:%s>%s:%d;", kind, a, b, pre, digest(s), value)
	}
}
func (b *Builder[E]) tryStatement(node *ast.Node) {
	if active || depth > 0 {
		b.adamicRawTryStatement(node)
		return
	}
	frames := []*tryFrame[E]{}
	lookup := map[*tryFrame[E]]int{}
	index := func(f *tryFrame[E]) int {
		if i, ok := lookup[f]; ok {
			return i
		}
		i := len(frames)
		frames = append(frames, f)
		lookup[f] = i
		return i
	}
	frameID = func(v any) int { return index(v.(*tryFrame[E])) }
	state = func() AdamicState {
		s := AdamicState{Cur: blockID(b.cur), Reachable: b.cur.Reachable, Incoming: []int{}, Frames: []AdamicFrame{}, Stack: []int{}}
		for _, f := range b.tryStack {
			s.Stack = append(s.Stack, index(f))
		}
		for _, blk := range b.blocks {
			if blk.hasIncoming {
				s.Incoming = append(s.Incoming, blockID(blk))
			}
		}
		for _, f := range frames {
			implicit := []int{}
			for _, blk := range f.implicit {
				implicit = append(implicit, blockID(blk))
			}
			s.Frames = append(s.Frames, AdamicFrame{int(f.position), f.hasFinally, blockID(f.catchEntry), blockID(f.finallyEntry), f.thrownForked, f.thrownAny, implicit, f.returnedAny})
		}
		return s
	}
	stmt := node.AsTryStatement()
	v := AdamicTry{TryBlock: nodeID(stmt.TryBlock), HasCatch: stmt.CatchClause != nil, HasFinally: stmt.FinallyBlock != nil, Binding: -1, CatchBlock: -1, FinallyBlock: nodeID(stmt.FinallyBlock)}
	if v.HasCatch {
		c := stmt.CatchClause.AsCatchClause()
		v.CatchBlock = nodeID(c.Block)
		v.BindingPresent = c.VariableDeclaration != nil
		if v.BindingPresent {
			v.Binding = nodeID(c.VariableDeclaration.Name())
		}
	}
	c := AdamicCase{Op: "try", State: state(), Try: v}
	events = []AdamicEvent{}
	trace = ""
	active = true
	b.adamicRawTryStatement(node)
	active = false
	c.Events = events
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, digest(state())+"/"+trace)
}
func (b *Builder[E]) statement(node *ast.Node) {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	b.adamicRawStatement(node)
	if suppress {
		depth--
	}
	record("statement", nodeID(node), -1, -1, pre)
}
func (b *Builder[E]) patternBind(node *ast.Node) {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	b.adamicRawPatternBind(node)
	if suppress {
		depth--
	}
	record("bind", nodeID(node), -1, -1, pre)
}
func (b *Builder[E]) newBlock() *Block[E] {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	v := b.adamicRawNewBlock()
	if suppress {
		depth--
	}
	record("new", -1, -1, blockID(v), pre)
	return v
}
func (b *Builder[E]) link(from, to *Block[E]) {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	b.adamicRawLink(from, to)
	if suppress {
		depth--
	}
	record("link", blockID(from), blockID(to), -1, pre)
}
func (b *Builder[E]) enter(blk *Block[E]) {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	b.adamicRawEnter(blk)
	if suppress {
		depth--
	}
	record("enter", blockID(blk), -1, -1, pre)
}
func (b *Builder[E]) snapshotForks() []bool {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	v := b.adamicRawSnapshotForks()
	if suppress {
		depth--
	}
	record("snapshot", -1, -1, 0, pre)
	return v
}
func (b *Builder[E]) restoreForks(snapshot []bool) {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	b.adamicRawRestoreForks(snapshot)
	if suppress {
		depth--
	}
	record("restore", 0, -1, -1, pre)
}
func (b *Builder[E]) returnFrame() int {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	v := b.adamicRawReturnFrame()
	if suppress {
		depth--
	}
	record("returnFrame", -1, -1, v, pre)
	return v
}
func (b *Builder[E]) throwFrame() int {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	v := b.adamicRawThrowFrame()
	if suppress {
		depth--
	}
	record("throwFrame", -1, -1, v, pre)
	return v
}
func (b *Builder[E]) markFinal(blk *Block[E]) {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	b.adamicRawMarkFinal(blk)
	if suppress {
		depth--
	}
	record("final", blockID(blk), -1, -1, pre)
}
func (b *Builder[E]) markThrown(blk *Block[E]) {
	pre := before()
	suppress := active
	if suppress {
		depth++
	}
	b.adamicRawMarkThrown(blk)
	if suppress {
		depth--
	}
	record("thrown", blockID(blk), -1, -1, pre)
}
func throwTarget[E any](f *tryFrame[E]) *Block[E] {
	pre := before()
	v := adamicRawThrowTarget(f)
	if active && depth == 0 {
		record("target", frameID(f), -1, blockID(v), pre)
	}
	return v
}

// Explicit per-node controls also observe nested try helpers hidden inside an outer statement callback.
func AdamicControls(root *ast.Node) {
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n.Kind == ast.KindTryStatement {
			for mode := 0; mode < 4; mode++ {
				b := &Builder[int]{}
				b.cur = b.newBlock()
				b.cur.Reachable = true
				b.cur.hasIncoming = true
				if mode > 0 {
					f := &tryFrame[int]{position: posTry, hasFinally: true, catchEntry: b.newBlock(), finallyEntry: b.newBlock(), thrownForked: true}
					if mode == 2 {
						f.position = posCatch
					}
					if mode == 3 {
						f.hasFinally = false
						f.finallyEntry = nil
					}
					b.tryStack = append(b.tryStack, f)
				}
				b.tryStatement(n)
			}
		}
		return n.ForEachChild(walk)
	}
	walk(root)
}
