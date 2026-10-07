package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

var adamicTrace string
var adamicLastLabels []string

func adamicLabelsOf(node *ast.Node) []string {
	adamicTrace += "labels:1;"
	adamicLastLabels = labelsOf(node)
	return adamicLastLabels
}
func adamicLabelText(node *ast.Node) string {
	text := node.Text()
	adamicTrace += "text:" + text + ";"
	return text
}
func adamicBlockID[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	return b.Index()
}
func (b *Builder[E]) adamicLink(from, to *Block[E]) {
	adamicTrace += fmt.Sprintf("link:%d:%d:", adamicBlockID(from), adamicBlockID(to))
	for _, j := range b.jumps {
		adamicTrace += fmt.Sprintf("%t,", j.broken)
	}
	adamicTrace += ";"
	b.link(from, to)
}
func (b *Builder[E]) adamicUnreachable() {
	adamicTrace += fmt.Sprintf("unreachable:%d;", b.cur.Index())
	b.makeUnreachable()
}

type AdamicJump struct {
	Labels    []string `json:"labels"`
	Break     int      `json:"break"`
	Continue  int      `json:"continue"`
	Loop      int      `json:"loop"`
	Breakable bool     `json:"breakable"`
	Broken    bool     `json:"broken"`
}
type AdamicSample struct {
	Jumps          []AdamicJump `json:"jumps"`
	Labels         []string     `json:"labels"`
	Break          int          `json:"break"`
	Continue       int          `json:"continue"`
	Loop           int          `json:"loop"`
	Reachable      bool         `json:"reachable"`
	HasLabel       bool         `json:"hasLabel"`
	Text           string       `json:"text"`
	AfterCur       int          `json:"afterCur"`
	AfterReachable bool         `json:"afterReachable"`
}

func AdamicJumpObserve(mode string, node, label *ast.Node, input AdamicSample) (AdamicSample, string) {
	b := &Builder[int]{}
	for i := 0; i < 5; i++ {
		b.newBlock()
	}
	b.cur = b.blocks[0]
	b.cur.Reachable = input.Reachable
	block := func(id int) *Block[int] {
		if id < 0 {
			return nil
		}
		return b.blocks[id]
	}
	loop := func(id int) *ast.Node {
		if id < 0 {
			return nil
		}
		return node
	}
	old := []jumpTarget[int]{}
	for _, j := range input.Jumps {
		old = append(old, jumpTarget[int]{labels: j.Labels, breakTo: block(j.Break), continueTo: block(j.Continue), loop: loop(j.Loop), breakable: j.Breakable, broken: j.Broken})
	}
	b.jumps = old
	adamicTrace = ""
	adamicLastLabels = nil
	switch mode {
	case "push":
		input.Labels = labelsOf(node)
		if input.Labels == nil {
			input.Labels = []string{}
		}
		b.pushJump(node, block(input.Break), block(input.Continue), loop(input.Loop))
	case "pop":
		b.popJump()
	case "break":
		b.makeBreak(label)
	default:
		panic("bad mode")
	}
	input.AfterCur = b.cur.Index()
	input.AfterReachable = b.cur.Reachable
	var out strings.Builder
	fmt.Fprintf(&out, "%d|%d:%t|", len(b.jumps), b.cur.Index(), b.cur.Reachable)
	for _, j := range b.jumps {
		lid := -1
		if j.loop != nil {
			lid = 1
			if j.loop != node {
				panic("loop identity changed")
			}
		}
		fmt.Fprintf(&out, "%s:%d:%d:%d:%t:%t;", strings.Join(j.labels, "~"), adamicBlockID(j.breakTo), adamicBlockID(j.continueTo), lid, j.breakable, j.broken)
	}
	if mode == "push" {
		added := b.jumps[len(b.jumps)-1]
		same := len(added.labels) == 0 || &added.labels[0] == &adamicLastLabels[0]
		fmt.Fprintf(&out, "labelsAlias:%t;", same)
	}
	// The original earlier targets' values survive append/truncate. Broken marking
	// is visible through an earlier saved Go slice when a matching target is found.
	out.WriteString("old:")
	for _, j := range old {
		fmt.Fprintf(&out, "%t,", j.broken)
	}
	out.WriteString("|")
	out.WriteString(adamicTrace)
	out.WriteByte('\n')
	return input, out.String()
}

func AdamicEmptyPopPanics() (panicked bool) {
	defer func() { panicked = recover() != nil }()
	b := &Builder[int]{}
	b.popJump()
	return false
}
