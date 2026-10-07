package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

var adamicTrace string

func adamicLabelText(node *ast.Node) string {
	text := node.Text()
	if adamicStatementDepth == 0 {
		adamicTrace += "text:" + text + ";"
	}
	return text
}
func adamicBlockID[E any](b *Block[E]) int {
	if b == nil {
		return -1
	}
	return b.Index()
}
func (b *Builder[E]) adamicLink(from, to *Block[E]) {
	if adamicStatementDepth == 0 {
		adamicTrace += fmt.Sprintf("link:%d:%d:", adamicBlockID(from), adamicBlockID(to))
		for _, j := range b.jumps {
			adamicTrace += fmt.Sprintf("%t,", j.broken)
		}
		adamicTrace += ";"
	}
	b.link(from, to)
}
func (b *Builder[E]) adamicLoop(n *ast.Node) {
	id := -1
	if n != nil {
		id = 1
	}
	if adamicStatementDepth == 0 {
		adamicTrace += fmt.Sprintf("loop:%d;", id)
	}
	b.loop(n)
}
func (b *Builder[E]) adamicUnreachable() {
	if adamicStatementDepth == 0 {
		adamicTrace += fmt.Sprintf("unreachable:%d;", b.cur.Index())
	}
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
	if mode != "continue" {
		panic("bad mode")
	}
	b.makeContinue(label)

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

var adamicNodeIDs map[*ast.Node]int
var adamicStatementDepth int
var adamicStatementCalls int

func (b *Builder[E]) adamicStatement(n *ast.Node) {
	if adamicStatementDepth == 0 {
		adamicTrace += fmt.Sprintf("stmt:%d;", adamicNodeIDs[n])
		adamicStatementCalls++
	}
	adamicStatementDepth++
	b.statement(n)
	adamicStatementDepth--
}

type AdamicStatementSample struct {
	HasList bool  `json:"hasList"`
	Nodes   []int `json:"nodes"`
}

func AdamicStatementsObserve(list *ast.NodeList) (AdamicStatementSample, string) {
	b := &Builder[int]{}
	b.cur = b.newBlock()
	b.cur.Reachable = true
	b.cur.hasIncoming = true
	adamicNodeIDs = map[*ast.Node]int{nil: -1}
	row := AdamicStatementSample{HasList: list != nil, Nodes: []int{}}
	if list != nil {
		for _, n := range list.Nodes {
			if _, ok := adamicNodeIDs[n]; !ok {
				adamicNodeIDs[n] = len(adamicNodeIDs)
			}
			row.Nodes = append(row.Nodes, adamicNodeIDs[n])
		}
	}
	adamicTrace = ""
	adamicStatementDepth = 0
	adamicStatementCalls = 0
	b.statements(list)
	return row, fmt.Sprintf("%d|%s\n", adamicStatementCalls, adamicTrace)
}
