package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

type AdamicJump struct {
	Labels     []string `json:"labels"`
	BreakTo    int      `json:"breakTo"`
	ContinueTo int      `json:"continueTo"`
	Loop       int      `json:"loop"`
	Breakable  bool     `json:"breakable"`
	Broken     bool     `json:"broken"`
}
type AdamicBlock struct {
	Index        int    `json:"index"`
	Successors   []int  `json:"successors"`
	CycleBarrier []bool `json:"cycleBarrier"`
	Reachable    bool   `json:"reachable"`
	HasIncoming  bool   `json:"hasIncoming"`
}
type AdamicCase struct {
	Op         string       `json:"op"`
	Jumps      []AdamicJump `json:"jumps"`
	Labels     []string     `json:"labels"`
	Node       int          `json:"node"`
	BreakTo    int          `json:"breakTo"`
	ContinueTo int          `json:"continueTo"`
	Loop       int          `json:"loop"`
	Cur        AdamicBlock  `json:"cur"`
	Fresh      AdamicBlock  `json:"fresh"`
}

var AdamicCases []AdamicCase
var AdamicOutputs []string
var adamicNodes = map[*ast.Node]int{}
var adamicTrace string

func adamicNode(n *ast.Node) int {
	if n == nil {
		return -1
	}
	if v, ok := adamicNodes[n]; ok {
		return v
	}
	v := len(adamicNodes)
	adamicNodes[n] = v
	return v
}
func adamicBlockID[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	return int(b.index)
}
func adamicJumps[E any](b *Builder[E]) []AdamicJump {
	out := []AdamicJump{}
	for _, j := range b.jumps {
		labels := append([]string{}, j.labels...)
		out = append(out, AdamicJump{labels, adamicBlockID(j.breakTo), adamicBlockID(j.continueTo), adamicNode(j.loop), j.breakable, j.broken})
	}
	return out
}
func adamicBlock[E any](b *Block[E]) AdamicBlock {
	out := AdamicBlock{Index: int(b.index), Successors: []int{}, Reachable: b.Reachable, HasIncoming: b.hasIncoming}
	for _, v := range b.Successors {
		out.Successors = append(out.Successors, int(v.index))
	}
	if b.cycleBarrier != nil {
		out.CycleBarrier = append([]bool{}, b.cycleBarrier...)
	}
	return out
}
func adamicSummaryJumps(jumps []AdamicJump) string {
	s := fmt.Sprint(len(jumps))
	for _, j := range jumps {
		s += fmt.Sprintf("/%d:%d:%d:%t:%t:%s", j.BreakTo, j.ContinueTo, j.Loop, j.Breakable, j.Broken, strings.Join(j.Labels, ","))
	}
	return s
}
func adamicSummaryBlock(b AdamicBlock) string {
	s := fmt.Sprintf("%d:%t:%t/", b.Index, b.Reachable, b.HasIncoming)
	for _, v := range b.Successors {
		s += fmt.Sprintf("%d,", v)
	}
	s += "/"
	if b.CycleBarrier == nil {
		s += "nil"
	} else {
		for _, v := range b.CycleBarrier {
			s += fmt.Sprintf("%t,", v)
		}
	}
	return s
}
func labelsOf(n *ast.Node) []string {
	adamicTrace += fmt.Sprintf("labels:%d;", adamicNode(n))
	return adamicRawLabelsOf(n)
}
func (b *Builder[E]) pushJump(n *ast.Node, br, co *Block[E], loop *ast.Node) {
	c := AdamicCase{Op: "push", Jumps: adamicJumps(b), Labels: append([]string{}, adamicRawLabelsOf(n)...), Node: adamicNode(n), BreakTo: adamicBlockID(br), ContinueTo: adamicBlockID(co), Loop: adamicNode(loop)}
	adamicTrace = ""
	b.adamicRawPushJump(n, br, co, loop)
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, adamicSummaryJumps(adamicJumps(b))+"/"+adamicTrace)
}
func (b *Builder[E]) popJump() {
	c := AdamicCase{Op: "pop", Jumps: adamicJumps(b)}
	b.adamicRawPopJump()
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, adamicSummaryJumps(adamicJumps(b))+"/")
}
func (b *Builder[E]) makeUnreachable() {
	old := b.cur
	c := AdamicCase{Op: "unreachable", Cur: adamicBlock(old), Fresh: AdamicBlock{Index: len(b.blocks), Successors: []int{}}}
	b.adamicRawMakeUnreachable()
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, adamicSummaryBlock(adamicBlock(old))+"|"+adamicSummaryBlock(adamicBlock(b.cur)))
}
func AdamicControls(node *ast.Node) {
	for depth := 0; depth < 8; depth++ {
		b := &Builder[int]{}
		br := b.newBlock()
		co := b.newBlock()
		for i := 0; i <= depth; i++ {
			b.pushJump(node, br, co, node)
			b.jumps[len(b.jumps)-1].broken = i%2 == 0
		}
		for len(b.jumps) > 0 {
			b.popJump()
		}
	}
	for count := 0; count < 10; count++ {
		for flags := 0; flags < 8; flags++ {
			b := &Builder[int]{}
			b.cur = b.newBlock()
			b.cur.Reachable = flags&1 != 0
			b.cur.hasIncoming = flags&2 != 0
			for i := 0; i < count; i++ {
				b.appendSuccessor(b.cur, b.newBlock())
			}
			if flags&4 != 0 {
				b.cur.cycleBarrier = make([]bool, count)
				for i := range b.cur.cycleBarrier {
					b.cur.cycleBarrier[i] = i%2 == 0
				}
			}
			for i := 0; i < 3; i++ {
				b.makeUnreachable()
			}
		}
	}
}

func AdamicEmptyPopRejected() (rejected bool) {
	defer func() { rejected = recover() != nil }()
	b := &Builder[int]{}
	b.adamicRawPopJump()
	return false
}
