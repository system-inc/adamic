package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type AdamicFrame struct {
	Index        int  `json:"index"`
	FinallyEntry int  `json:"finallyEntry"`
	Target       int  `json:"target"`
	ReturnedAny  bool `json:"returnedAny"`
	ThrownAny    bool `json:"thrownAny"`
}
type AdamicCase struct {
	Op             string        `json:"op"`
	CurIndex       int           `json:"curIndex"`
	Reachable      bool          `json:"reachable"`
	ReturnFrame    int           `json:"returnFrame"`
	ThrowFrame     int           `json:"throwFrame"`
	Frames         []AdamicFrame `json:"frames"`
	FreshIndex     int           `json:"freshIndex"`
	AfterIndex     int           `json:"afterIndex"`
	AfterReachable bool          `json:"afterReachable"`
}

var AdamicCases []AdamicCase
var AdamicOutputs []string
var adamicActive bool
var adamicDepth int
var adamicTrace string
var adamicFlags func() string
var adamicFrameIDs = map[any]int{}

func adamicBlockID[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	return int(b.index)
}
func adamicFrameFlags[E any](b *Builder[E]) string {
	s := ""
	for _, f := range b.tryStack {
		s += fmt.Sprintf("%t:%t,", f.returnedAny, f.thrownAny)
	}
	return s
}
func adamicRecord(s string) {
	if adamicActive && adamicDepth == 0 {
		adamicTrace += s + "/" + adamicFlags() + ";"
	}
}
func (b *Builder[E]) returnFrame() int {
	adamicDepth++
	v := b.adamicRawReturnFrame()
	adamicDepth--
	adamicRecord(fmt.Sprintf("return:%d", v))
	return v
}
func (b *Builder[E]) throwFrame() int {
	adamicDepth++
	v := b.adamicRawThrowFrame()
	adamicDepth--
	adamicRecord(fmt.Sprintf("throw:%d", v))
	return v
}
func throwTarget[E any](f *tryFrame[E]) *Block[E] {
	adamicRecord(fmt.Sprintf("target:%d", adamicFrameIDs[f]))
	adamicDepth++
	v := adamicRawThrowTarget(f)
	adamicDepth--
	return v
}
func (b *Builder[E]) link(from, to *Block[E]) {
	adamicRecord(fmt.Sprintf("link:%d:%d", adamicBlockID(from), adamicBlockID(to)))
	adamicDepth++
	b.adamicRawLink(from, to)
	adamicDepth--
}
func (b *Builder[E]) markFinal(blk *Block[E]) {
	adamicRecord(fmt.Sprintf("final:%d", adamicBlockID(blk)))
	adamicDepth++
	b.adamicRawMarkFinal(blk)
	adamicDepth--
}
func (b *Builder[E]) markThrown(blk *Block[E]) {
	adamicRecord(fmt.Sprintf("thrown:%d", adamicBlockID(blk)))
	adamicDepth++
	b.adamicRawMarkThrown(blk)
	adamicDepth--
}
func (b *Builder[E]) makeUnreachable() {
	adamicRecord(fmt.Sprintf("unreachable:%d", adamicBlockID(b.cur)))
	adamicDepth++
	b.adamicRawMakeUnreachable()
	adamicDepth--
}
func (b *Builder[E]) newBlock() *Block[E] {
	adamicRecord("new")
	adamicDepth++
	v := b.adamicRawNewBlock()
	adamicDepth--
	return v
}
func (b *Builder[E]) enter(blk *Block[E]) {
	adamicRecord(fmt.Sprintf("enter:%d", adamicBlockID(blk)))
	adamicDepth++
	b.adamicRawEnter(blk)
	adamicDepth--
}
func adamicObserve[E any](b *Builder[E], op string, run func()) {
	c := AdamicCase{Op: op, CurIndex: adamicBlockID(b.cur), Reachable: b.cur.Reachable, ReturnFrame: b.adamicRawReturnFrame(), ThrowFrame: b.adamicRawThrowFrame(), FreshIndex: len(b.blocks), Frames: []AdamicFrame{}}
	for i, f := range b.tryStack {
		adamicFrameIDs[f] = i
		c.Frames = append(c.Frames, AdamicFrame{i, adamicBlockID(f.finallyEntry), adamicBlockID(adamicRawThrowTarget(f)), f.returnedAny, f.thrownAny})
	}
	adamicFlags = func() string { return adamicFrameFlags(b) }
	adamicTrace = ""
	adamicDepth = 0
	adamicActive = true
	run()
	adamicActive = false
	c.AfterIndex = adamicBlockID(b.cur)
	c.AfterReachable = b.cur.Reachable
	AdamicCases = append(AdamicCases, c)
	AdamicOutputs = append(AdamicOutputs, fmt.Sprintf("%d:%t/%s/%s", c.AfterIndex, c.AfterReachable, adamicFrameFlags(b), adamicTrace))
}
func (b *Builder[E]) makeReturn() { adamicObserve(b, "return", func() { b.adamicRawMakeReturn() }) }
func (b *Builder[E]) makeThrow()  { adamicObserve(b, "throw", func() { b.adamicRawMakeThrow() }) }
func (b *Builder[E]) makeYield()  { adamicObserve(b, "yield", func() { b.adamicRawMakeYield() }) }
func AdamicControls(node *ast.Node) {
	for depth := 0; depth < 4; depth++ {
		for mask := 0; mask < 64; mask++ {
			for _, op := range []string{"return", "throw", "yield"} {
				for reachable := 0; reachable < 2; reachable++ {
					b := &Builder[int]{}
					b.cur = b.newBlock()
					b.cur.Reachable = reachable == 1
					for i := 0; i < depth; i++ {
						f := &tryFrame[int]{position: tryPosition((mask >> uint(i)) & 1), hasFinally: mask&8 != 0, returnedAny: mask&16 != 0, thrownAny: mask&32 != 0}
						if mask&2 != 0 {
							f.catchEntry = b.newBlock()
						}
						if mask&4 != 0 {
							f.finallyEntry = b.newBlock()
						}
						b.tryStack = append(b.tryStack, f)
					}
					switch op {
					case "return":
						b.makeReturn()
					case "throw":
						b.makeThrow()
					case "yield":
						b.makeYield()
					}
				}
			}
		}
	}
}
