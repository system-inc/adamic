// Raw syntax-flow graph, copied from the pinned cohere generic CFG shelf.
// MIT provenance is retained below. It contains no lint predicates or findings.
package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"math/big"
	"slices"
	"strings"
)

func (p *Program) wave27SyntaxFlow(out *fields, node *ast.Node, question string) (string, error) {
	if question != "wave-27-syntax-flow" || !syntaxFlowIsRoot(node) {
		return "", fmt.Errorf("syntax-flow requires a code path root")
	}
	type event struct {
		kind string
		node *ast.Node
	}
	graph := syntaxFlowBuild(node, syntaxFlowHooks[event]{Expression: func(b *syntaxFlowBuilder[event], n *ast.Node) { b.Emit(event{"expression", n}) }, Statement: func(b *syntaxFlowBuilder[event], n *ast.Node) { b.Emit(event{"statement", n}) }})
	out.number(uint64(len(graph.Blocks)))
	for _, block := range graph.Blocks {
		out.yes(block.Reachable)
		out.number(uint64(len(block.Events)))
		for _, e := range block.Events {
			out.text(e.kind)
			out.text(strings.TrimPrefix(e.node.Kind.String(), "Kind"))
			out.number(uint64(e.node.Pos()))
			out.number(uint64(e.node.End()))
		}
		var edges []uint64
		for _, successor := range block.Successors {
			if successor != nil {
				edges = append(edges, uint64(successor.Index()))
			}
		}
		out.ids(edges)
	}
	return out.String(), nil
}

// Original shelf: cfg.go
// Vendored from web-infra-dev/rslint, `internal/utilities/cfg`, at commit
// 44956a0c53201157aaf5bc19e1c2728896c83ec4.
//
// MIT licensed. Copyright (c) 2025 July-present Bytedance Inc and its affiliates, and
// Copyright (c) 2025 July typescript-eslint and other contributors — the dual attribution is
// upstream's own, because this package descends from tsgolint.
//
// Changed on the way in, and nothing else:
//
//   - The package is `controlflow` rather than `cfg`, matching the `internal/utilities/<domain>/`
//     shelf, where a package is named for the question it answers rather than abbreviated.
//   - Imports moved from `github.com/microsoft/typescript-go/shim/...` to
//     `github.com/microsoft/TypeScript/tsc/shim/...`. Not a fork: the upstream repository was
//     archived and moved, and the two spellings are the same shims at different pins.
//   - `helpers.go` no longer imports rslint's shared `internal/utils`. It reached for exactly two
//     functions; one was dropped as a no-op on our parser and one was inlined. The measurement
//     that separates them is recorded at `syntaxFlowIsAlwaysTruthyTest`.
//
// # Added here, and absent upstream
//
// Two files are ours entirely. They are listed separately from the changes above because the
// distinction matters to the next person diffing against upstream: nothing below modifies a
// vendored line, so a three-way merge of a new rslint commit touches none of it.
//
//   - `dominators.go` — predecessor edges and a dominator tree over the reachable subgraph.
//     Upstream computes both inside `paths.go`'s `findFinalPathDominators` and discards them; this
//     exposes them as an API. The algorithm is Cooper-Harvey-Kennedy either way, written a second
//     time here because the two run over different graphs (that one adds a synthetic exit and drops
//     thrown edges), and `TestDominatorsAgreeWithFinalPathDominators` pins them against each other
//     so neither can drift silently.
//   - `dataflow.go` — a monotone dataflow solver over a caller-supplied lattice, in either
//     direction. Upstream has none, and every consumer that needed one wrote its own loop inline.
//
// Both are additive. No vendored file lost a line to them, and the three consumers that predate
// them — `internal/unused/reachable.go`, `no_useless_assignment_analysis.go`, and
// `internal/rules/react/rules_of_hooks.go` — were measured byte-identical in their findings across
// the whole `~/Projects/ahra` tree afterwards.
//
// Three rules consume this package today. It began as a foundation lift with no callers, and the
// coupling to the rest of the tree is still deliberately zero: it imports one shim and the standard
// library, so it can be adopted by a rule without dragging anything behind it.
//
// Package controlflow builds a control-flow graph for one code path root — a source
// file, a function, a class static block, or a class field initializer.
//
// The graph mirrors the shape ESLint's code path analysis produces, because the
// observable behaviour of the rules built on it depends on it:
//
//   - the end of a `try` block flows into its `catch` block as well as past the
//     statement;
//   - the first node inside a `try` block (or inside the `catch` block of a
//     `try` that has a `finally`) that could throw forks to the handler, so an
//     assignment later in the block is bypassed on the throwing path;
//   - a `finally` block is laid out twice, once for normal completion and once
//     for the path that leaves the `try` statement through `return`, `throw`,
//     or a suspended `yield`; `break` and `continue` do not take that path;
//   - `&&`, `||`, `??`, `?:`, optional chains, and destructuring defaults fork
//     around the operand they may skip;
//   - the code after a `return`, `throw`, `break`, or `continue` is laid out in
//     blocks nothing reaches rather than dropped, keeping the edge that leads
//     into it.
//
// The graph carries no meaning of its own beyond that shape. A consumer names
// an event type and supplies syntaxFlowHooks; the builder calls them where a reference is
// read and where one is written, in evaluation order, and the consumer records
// whatever it needs into the current block. Analysis is then a plain dataflow
// over syntaxFlowGraph.Blocks.

// syntaxFlowBlock is one basic block: the events its consumer recorded, in evaluation
// order, and the blocks led to from it.
//
// Successors spans the whole graph, unreachable blocks included, so a walk that
// only wants the code that runs skips the blocks whose Reachable is false.
type syntaxFlowBlock[E any] struct {
	Events     []E
	Successors []*syntaxFlowBlock[E]

	// Reachable reports whether control can arrive here. The code after an
	// abrupt exit is laid out in unreachable blocks rather than dropped, and a
	// consumer's hooks run in them, so a rule can ask what would have run had
	// the exit not been there.
	//
	// It is settled when the block is entered, from the edges that lead into
	// it by then. Every block is entered after the edges that could make it
	// reachable exist: a block only ever gains an edge afterwards from code
	// laid out inside it, which is unreachable whenever the block itself is.
	Reachable bool

	index        int32
	hasIncoming  bool
	final        bool
	thrown       bool
	cycleBarrier []bool
	successors   [2]*syntaxFlowBlock[E]
}

const syntaxFlowBlockChunkSize = 8

type syntaxFlowBlockChunk[E any] [syntaxFlowBlockChunkSize]syntaxFlowBlock[E]

// Index reports the block's position in syntaxFlowGraph.Blocks, so a consumer can hold
// per-block analysis state in a slice rather than a map.
func (blk *syntaxFlowBlock[E]) Index() int {
	return int(blk.index)
}

// syntaxFlowGraph is the control-flow graph of one code path root, unreachable blocks
// included. Blocks[0] is the entry block; the rest are in construction order,
// which no analysis should depend on.
//
// FinalBlocks are the reachable blocks where the root completes normally,
// returns, or is suspended by a yield. ThrownBlocks are the reachable blocks
// where an uncaught throw leaves the root. A block can be in both sets when a
// yield gives the caller both ways to terminate the generator. Keeping the two
// sets separate lets consumers mirror analyses, such as ESLint code paths,
// that deliberately exclude thrown exits. Different abrupt completions can
// also share one block after a finally clause, placing that block in both sets.
//
// EndReachable reports whether control can run off the end of the root — the
// question ESLint answers with `isAnySegmentReachable(codePath.currentSegments)`
// at the root's `:exit`. It is false when every path out of the root returns,
// throws, or loops forever.
type syntaxFlowGraph[E any] struct {
	Blocks       []*syntaxFlowBlock[E]
	FinalBlocks  []*syntaxFlowBlock[E]
	ThrownBlocks []*syntaxFlowBlock[E]
	EndReachable bool
}

// syntaxFlowHooks are the points where a consumer turns a position in the walk into an
// event. Every hook may be nil, and every one is called from every position the
// walk reaches, whether or not control can arrive there — a consumer that has
// no use for unreachable code checks syntaxFlowBuilder.Current for it.
//
// Expression runs before each expression is evaluated. Read runs where an
// identifier is evaluated as a read, Write where its value is stored; a target
// that is read before it is written — a compound assignment, an update
// expression, a destructuring element — sees both, in that order.
//
// Statement runs as each statement is reached, before it is laid out. Loop runs
// where control flows back into a loop for another iteration of it.
type syntaxFlowHooks[E any] struct {
	Expression func(b *syntaxFlowBuilder[E], node *ast.Node)
	Read       func(b *syntaxFlowBuilder[E], node *ast.Node)
	Write      func(b *syntaxFlowBuilder[E], node *ast.Node)
	Statement  func(b *syntaxFlowBuilder[E], node *ast.Node)
	Loop       func(b *syntaxFlowBuilder[E], loop *ast.Node)
}

// syntaxFlowTryPosition is where in a `try` statement construction currently is.
type syntaxFlowTryPosition uint8

const (
	syntaxFlowPosTry syntaxFlowTryPosition = iota
	syntaxFlowPosCatch
)

// syntaxFlowTryFrame tracks one enclosing `try` statement while its blocks are built.
type syntaxFlowTryFrame[E any] struct {
	position     syntaxFlowTryPosition
	hasFinally   bool
	catchEntry   *syntaxFlowBlock[E]
	finallyEntry *syntaxFlowBlock[E]
	thrownForked bool
	thrownAny    bool
	implicit     []*syntaxFlowBlock[E]
	returnedAny  bool
}

// syntaxFlowJumpTarget is one `break`/`continue` destination. continueTo is nil for
// targets that only `break` can reach (switch statements, labelled blocks).
// breakable mirrors ESLint's BreakContext#breakable: it is true for loops and
// switch statements, and false for a labelled statement that wraps neither, so
// an unlabelled `break` skips past it to the enclosing loop/switch. labels
// holds every label the target answers to — a loop wrapped as
// `outer: inner: while (…)` accepts both names. loop names the loop a
// `continue` starts another iteration of, and is nil where continuing does not
// re-enter the loop's first node: a `do`…`while` continues into its test, which
// may still end the loop.
type syntaxFlowJumpTarget[E any] struct {
	labels     []string
	breakTo    *syntaxFlowBlock[E]
	continueTo *syntaxFlowBlock[E]
	loop       *ast.Node
	breakable  bool
	broken     bool
}

// syntaxFlowBuilder lays out a graph. Consumers only see it through their syntaxFlowHooks, where
// Current and Emit are the operations that matter.
type syntaxFlowBuilder[E any] struct {
	hooks syntaxFlowHooks[E]

	blocks      []*syntaxFlowBlock[E]
	cur         *syntaxFlowBlock[E]
	finals      []*syntaxFlowBlock[E]
	thrown      []*syntaxFlowBlock[E]
	blockChunks []*syntaxFlowBlockChunk[E]

	jumps      []syntaxFlowJumpTarget[E]
	tryStack   []*syntaxFlowTryFrame[E]
	chainJoins []*syntaxFlowBlock[E]
}

// syntaxFlowBuild lays out the control-flow graph of one code path root.
func syntaxFlowBuild[E any](root *ast.Node, hooks syntaxFlowHooks[E]) *syntaxFlowGraph[E] {
	b := &syntaxFlowBuilder[E]{hooks: hooks, blocks: make([]*syntaxFlowBlock[E], 0, syntaxFlowBlockChunkSize)}
	b.cur = b.newBlock()
	b.cur.hasIncoming = true
	b.cur.Reachable = true

	switch root.Kind {
	case ast.KindSourceFile:
		b.statements(root.AsSourceFile().Statements)

	case ast.KindClassStaticBlockDeclaration:
		b.statement(root.AsClassStaticBlockDeclaration().Body)

	case ast.KindPropertyDeclaration:
		b.expr(root.AsPropertyDeclaration().Initializer)

	default:
		for _, parameter := range root.Parameters() {
			b.parameter(parameter)
		}
		// A `typeof x` inside the signature's type annotations references the
		// variable even though nothing in them runs.
		b.expr(root.Type())
		if body := root.Body(); body != nil {
			if body.Kind == ast.KindBlock {
				b.statement(body)
			} else {
				b.expr(body)
			}
		}
	}

	endReachable := b.cur.Reachable
	if endReachable {
		b.markFinal(b.cur)
	}
	return &syntaxFlowGraph[E]{
		Blocks:       b.blocks,
		FinalBlocks:  b.finals,
		ThrownBlocks: b.thrown,
		EndReachable: endReachable,
	}
}

// parameter models a parameter binding: its default value is skipped when an
// argument was passed.
func (b *syntaxFlowBuilder[E]) parameter(node *ast.Node) {
	parameter := node.AsParameterDeclaration()
	if parameter == nil {
		return
	}
	b.expr(parameter.Type)
	b.bindWithDefault(parameter.Name(), parameter.Initializer)
}

// ---------------------------------------------------------------------------
// blocks
// ---------------------------------------------------------------------------

// Current returns the block being filled. Its Reachable field says whether the
// position control is at is one control can arrive at.
func (b *syntaxFlowBuilder[E]) Current() *syntaxFlowBlock[E] {
	return b.cur
}

// Emit records one event in the current block and returns where it landed, so a
// consumer can refer back to it.
func (b *syntaxFlowBuilder[E]) Emit(e E) (*syntaxFlowBlock[E], int) {
	index := len(b.cur.Events)
	b.cur.Events = append(b.cur.Events, e)
	return b.cur, index
}

func (b *syntaxFlowBuilder[E]) newBlock() *syntaxFlowBlock[E] {
	index := len(b.blocks)
	chunkIndex := index / syntaxFlowBlockChunkSize
	if chunkIndex == len(b.blockChunks) {
		b.blockChunks = append(b.blockChunks, new(syntaxFlowBlockChunk[E]))
	}
	blk := &b.blockChunks[chunkIndex][index%syntaxFlowBlockChunkSize]
	blk.index = int32(index)
	b.blocks = append(b.blocks, blk)
	return blk
}

func (b *syntaxFlowBuilder[E]) markFinal(blk *syntaxFlowBlock[E]) {
	if blk == nil || !blk.Reachable || blk.final {
		return
	}
	blk.final = true
	b.finals = append(b.finals, blk)
}

func (b *syntaxFlowBuilder[E]) markThrown(blk *syntaxFlowBlock[E]) {
	if blk == nil || !blk.Reachable || blk.thrown {
		return
	}
	blk.thrown = true
	b.thrown = append(b.thrown, blk)
}

func (b *syntaxFlowBuilder[E]) link(from, to *syntaxFlowBlock[E]) {
	b.linkWithCycleBarrier(from, to, false)
}

func (b *syntaxFlowBuilder[E]) linkWithCycleBarrier(from, to *syntaxFlowBlock[E], barrier bool) int {
	if from == nil || to == nil {
		return -1
	}
	index := len(from.Successors)
	b.appendSuccessor(from, to)
	b.setCycleBarrier(from, index, barrier)
	if from.Reachable {
		to.hasIncoming = true
	}
	return index
}

func (b *syntaxFlowBuilder[E]) appendSuccessor(from, to *syntaxFlowBlock[E]) {
	index := len(from.Successors)
	if index < len(from.successors) {
		from.successors[index] = to
		from.Successors = from.successors[:index+1]
		return
	}
	from.Successors = append(from.Successors, to)
}

func (b *syntaxFlowBuilder[E]) setCycleBarrier(from *syntaxFlowBlock[E], index int, barrier bool) {
	if from == nil || index < 0 {
		return
	}
	if from.cycleBarrier == nil {
		if !barrier {
			return
		}
		from.cycleBarrier = make([]bool, len(from.Successors))
	} else {
		for len(from.cycleBarrier) < len(from.Successors) {
			from.cycleBarrier = append(from.cycleBarrier, false)
		}
	}
	from.cycleBarrier[index] = barrier
}

// enter makes blk the current block and settles its reachability, so every edge
// that could make it reachable has to be linked before it is entered.
func (b *syntaxFlowBuilder[E]) enter(blk *syntaxFlowBlock[E]) {
	blk.Reachable = blk.hasIncoming
	b.cur = blk
}

// makeUnreachable ends the path at the current block and carries the walk on in
// a fresh block nothing reaches. The edge into it is recorded, but — unlike
// every other edge — it does not make its target something control arrives at,
// which is what leaving through it meant in the first place. That is why the
// block is never entered: its reachability is settled here, as false.
func (b *syntaxFlowBuilder[E]) makeUnreachable() {
	next := b.newBlock()
	b.appendSuccessor(b.cur, next)
	if b.cur.cycleBarrier != nil {
		b.cur.cycleBarrier = append(b.cur.cycleBarrier, false)
	}
	b.cur = next
}

// enterDisconnected makes blk current when the position the walk stands at is
// one control arrives at, rather than when an edge into blk already is. It is
// how a block whose only edge in is laid out later still gets laid out here,
// which is the order ESLint traverses one in.
func (b *syntaxFlowBuilder[E]) enterDisconnected(blk *syntaxFlowBlock[E]) {
	blk.hasIncoming = blk.hasIncoming || b.cur.Reachable
	b.enter(blk)
}

func (b *syntaxFlowBuilder[E]) read(node *ast.Node) {
	if b.hooks.Read != nil {
		b.hooks.Read(b, node)
	}
}

func (b *syntaxFlowBuilder[E]) write(node *ast.Node) {
	if b.hooks.Write != nil {
		b.hooks.Write(b, node)
	}
}

func (b *syntaxFlowBuilder[E]) reachedStatement(node *ast.Node) {
	if b.hooks.Statement != nil {
		b.hooks.Statement(b, node)
	}
}

func (b *syntaxFlowBuilder[E]) loop(node *ast.Node) {
	if b.hooks.Loop != nil && node != nil {
		b.hooks.Loop(b, node)
	}
}

// ---------------------------------------------------------------------------
// throw / return routing
// ---------------------------------------------------------------------------

// throwFrame returns the innermost `try` frame an exception raised at the
// current position lands in, or -1 when it leaves the code path entirely.
func (b *syntaxFlowBuilder[E]) throwFrame() int {
	for i := len(b.tryStack) - 1; i >= 0; i-- {
		f := b.tryStack[i]
		if f.position == syntaxFlowPosTry || (f.hasFinally && f.position == syntaxFlowPosCatch) {
			return i
		}
	}
	return -1
}

func syntaxFlowThrowTarget[E any](f *syntaxFlowTryFrame[E]) *syntaxFlowBlock[E] {
	if f.position == syntaxFlowPosTry && f.catchEntry != nil {
		return f.catchEntry
	}
	return f.finallyEntry
}

// returnFrame returns the innermost `try` frame whose `finally` block runs
// before a `return` leaves the code path, or -1 when none does.
func (b *syntaxFlowBuilder[E]) returnFrame() int {
	for i := len(b.tryStack) - 1; i >= 0; i-- {
		if b.tryStack[i].hasFinally {
			return i
		}
	}
	return -1
}

// firstThrowableFork forks to the enclosing handler the first time a node that
// could throw is completed inside a `try` block (or inside the `catch` block of
// a `try` that has a `finally`). Later throwable nodes in the same block reuse
// that first fork, matching ESLint's approximation.
func (b *syntaxFlowBuilder[E]) firstThrowableFork() {
	if !b.cur.Reachable {
		return
	}
	i := b.throwFrame()
	if i < 0 {
		return
	}
	f := b.tryStack[i]
	if f.thrownForked {
		return
	}
	f.thrownForked = true
	f.thrownAny = true
	b.link(b.cur, syntaxFlowThrowTarget(f))
	next := b.newBlock()
	b.link(b.cur, next)
	b.enter(next)
}

// snapshotForks records the "already forked" state of every enclosing `try` so
// the second copy of a `finally` block forks at the same place the first did.
func (b *syntaxFlowBuilder[E]) snapshotForks() []bool {
	snapshot := make([]bool, len(b.tryStack))
	for i, f := range b.tryStack {
		snapshot[i] = f.thrownForked
	}
	return snapshot
}

func (b *syntaxFlowBuilder[E]) restoreForks(snapshot []bool) {
	for i, forked := range snapshot {
		if i < len(b.tryStack) {
			b.tryStack[i].thrownForked = forked
		}
	}
}

// Original shelf: expressions.go

func (b *syntaxFlowBuilder[E]) expr(node *ast.Node) {
	if node == nil {
		return
	}
	if b.hooks.Expression != nil {
		b.hooks.Expression(b, node)
	}

	switch node.Kind {
	case ast.KindIdentifier:
		b.read(node)
		if syntaxFlowIsThrowableIdentifier(node) {
			b.firstThrowableFork()
		}

	case ast.KindParenthesizedExpression:
		b.expr(node.AsParenthesizedExpression().Expression)

	case ast.KindBinaryExpression:
		b.binaryExpression(node)

	case ast.KindConditionalExpression:
		b.conditionalExpression(node)

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			b.updateExpression(unary.Operand)
			return
		}
		b.expr(unary.Operand)

	case ast.KindPostfixUnaryExpression:
		b.updateExpression(node.AsPostfixUnaryExpression().Operand)

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		b.accessOrCall(node)

	case ast.KindYieldExpression:
		b.expr(node.AsYieldExpression().Expression)
		b.makeYield()

	case ast.KindClassExpression:
		b.classLike(node)

	default:
		if syntaxFlowIsRoot(node) {
			b.nestedFunction(node)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			b.visitUnknown(child)
			return false
		})
	}
}

// condition evaluates node while keeping its truthy and falsy continuations
// separate. A plain expr may merge short-circuit branches because its value is
// all the consumer needs; a control-flow condition cannot merge them before it
// decides which statement executes, or impossible paths can appear across the
// join (for example, the short branch of `a || b` reaching an `else` body).
func (b *syntaxFlowBuilder[E]) condition(node *ast.Node, whenTrue, whenFalse *syntaxFlowBlock[E]) {
	if node == nil {
		b.link(b.cur, whenTrue)
		b.link(b.cur, whenFalse)
		return
	}

	switch node.Kind {
	case ast.KindParenthesizedExpression:
		if b.hooks.Expression != nil {
			b.hooks.Expression(b, node)
		}
		b.condition(node.AsParenthesizedExpression().Expression, whenTrue, whenFalse)

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator != ast.KindExclamationToken {
			b.expr(node)
			b.link(b.cur, whenTrue)
			b.link(b.cur, whenFalse)
			return
		}
		if b.hooks.Expression != nil {
			b.hooks.Expression(b, node)
		}
		b.condition(unary.Operand, whenFalse, whenTrue)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if operator != ast.KindAmpersandAmpersandToken &&
			operator != ast.KindBarBarToken &&
			operator != ast.KindQuestionQuestionToken {
			b.expr(node)
			b.link(b.cur, whenTrue)
			b.link(b.cur, whenFalse)
			return
		}
		if b.hooks.Expression != nil {
			b.hooks.Expression(b, node)
		}
		right := b.newBlock()
		switch operator {
		case ast.KindAmpersandAmpersandToken:
			b.condition(binary.Left, right, whenFalse)
		case ast.KindBarBarToken:
			b.condition(binary.Left, whenTrue, right)
		case ast.KindQuestionQuestionToken:
			// A non-nullish left operand may itself be truthy or falsy; a
			// nullish one evaluates the right operand.
			b.expr(binary.Left)
			b.link(b.cur, whenTrue)
			b.link(b.cur, whenFalse)
			b.link(b.cur, right)
		}
		b.enter(right)
		b.condition(binary.Right, whenTrue, whenFalse)

	case ast.KindConditionalExpression:
		if b.hooks.Expression != nil {
			b.hooks.Expression(b, node)
		}
		conditional := node.AsConditionalExpression()
		trueEntry := b.newBlock()
		falseEntry := b.newBlock()
		b.condition(conditional.Condition, trueEntry, falseEntry)
		b.enter(trueEntry)
		b.condition(conditional.WhenTrue, whenTrue, whenFalse)
		b.enter(falseEntry)
		b.condition(conditional.WhenFalse, whenTrue, whenFalse)

	default:
		b.expr(node)
		b.link(b.cur, whenTrue)
		b.link(b.cur, whenFalse)
	}
}

// visitUnknown routes a child of a node the builder has no dedicated handler
// for. Statements and expressions are handled by their own walkers so control
// flow inside them is still modelled.
func (b *syntaxFlowBuilder[E]) visitUnknown(node *ast.Node) {
	if node == nil {
		return
	}
	if ast.IsStatement(node) {
		b.statement(node)
		return
	}
	b.expr(node)
}

func (b *syntaxFlowBuilder[E]) binaryExpression(node *ast.Node) {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken.Kind

	switch {
	case operator == ast.KindAmpersandAmpersandToken ||
		operator == ast.KindBarBarToken ||
		operator == ast.KindQuestionQuestionToken:
		b.expr(binary.Left)
		join := b.newBlock()
		b.link(b.cur, join)
		right := b.newBlock()
		b.link(b.cur, right)
		b.enter(right)
		b.expr(binary.Right)
		b.link(b.cur, join)
		b.enter(join)

	case ast.IsLogicalOrCoalescingAssignmentOperator(operator):
		// `a ||= b` reads a, then conditionally evaluates b and writes a.
		b.patternReads(binary.Left)
		join := b.newBlock()
		b.link(b.cur, join)
		right := b.newBlock()
		b.link(b.cur, right)
		b.enter(right)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)
		b.link(b.cur, join)
		b.enter(join)

	case operator == ast.KindEqualsToken:
		if syntaxFlowIsDestructuringTarget(binary.Left) {
			// Run-time evaluation order: the right-hand side produces the
			// value first, then the pattern binds element by element.
			b.expr(binary.Right)
			b.patternBind(binary.Left)
			return
		}
		b.patternReads(binary.Left)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)

	case ast.IsAssignmentOperator(operator):
		// Compound assignment reads the target before writing it.
		b.patternReads(binary.Left)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)

	default:
		b.expr(binary.Left)
		b.expr(binary.Right)
	}
}

func (b *syntaxFlowBuilder[E]) conditionalExpression(node *ast.Node) {
	cond := node.AsConditionalExpression()
	b.expr(cond.Condition)
	testEnd := b.cur
	join := b.newBlock()

	whenTrue := b.newBlock()
	b.link(testEnd, whenTrue)
	b.enter(whenTrue)
	b.expr(cond.WhenTrue)
	b.link(b.cur, join)

	whenFalse := b.newBlock()
	b.link(testEnd, whenFalse)
	b.enter(whenFalse)
	b.expr(cond.WhenFalse)
	b.link(b.cur, join)

	b.enter(join)
}

func (b *syntaxFlowBuilder[E]) updateExpression(operand *ast.Node) {
	b.patternReads(operand)
	b.patternWrites(operand)
}

// accessOrCall walks a member access or call and forks around each `?.` link so
// the rest of the chain can be skipped.
func (b *syntaxFlowBuilder[E]) accessOrCall(node *ast.Node) {
	outermost := ast.IsOptionalChain(node) && ast.IsOutermostOptionalChain(node)
	if outermost {
		b.chainJoins = append(b.chainJoins, b.newBlock())
	}

	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		b.expr(node.AsPropertyAccessExpression().Expression)
		b.forkOptionalChain(node)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		b.expr(access.Expression)
		b.forkOptionalChain(node)
		b.expr(access.ArgumentExpression)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		b.expr(call.Expression)
		b.forkOptionalChain(node)
		b.typeArguments(node)
		if call.Arguments != nil {
			for _, arg := range call.Arguments.Nodes {
				b.expr(arg)
			}
		}
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		b.expr(newExpr.Expression)
		b.typeArguments(node)
		if newExpr.Arguments != nil {
			for _, arg := range newExpr.Arguments.Nodes {
				b.expr(arg)
			}
		}
	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		b.expr(tagged.Tag)
		b.typeArguments(node)
		b.expr(tagged.Template)
	}

	b.firstThrowableFork()

	if outermost {
		join := b.chainJoins[len(b.chainJoins)-1]
		b.chainJoins = b.chainJoins[:len(b.chainJoins)-1]
		b.link(b.cur, join)
		b.enter(join)
	}
}

func (b *syntaxFlowBuilder[E]) forkOptionalChain(node *ast.Node) {
	if len(b.chainJoins) == 0 || !ast.IsOptionalChainRoot(node) {
		return
	}
	b.link(b.cur, b.chainJoins[len(b.chainJoins)-1])
	next := b.newBlock()
	b.link(b.cur, next)
	b.enter(next)
}

// typeArguments walks an explicit `<T>` argument list, so a `typeof x` inside
// it still counts as a read even though nothing in it runs.
func (b *syntaxFlowBuilder[E]) typeArguments(node *ast.Node) {
	for _, typeArg := range node.TypeArguments() {
		b.expr(typeArg)
	}
}

// typeParameters walks a `<T extends U = D>` declaration list, so a `typeof x`
// inside a constraint or default still counts as a read even though nothing in
// it runs.
func (b *syntaxFlowBuilder[E]) typeParameters(node *ast.Node) {
	for _, typeParam := range node.TypeParameters() {
		param := typeParam.AsTypeParameterDeclaration()
		if param == nil {
			continue
		}
		b.expr(param.Constraint)
		b.expr(param.DefaultType)
	}
}

// nestedFunction keeps evaluating the parts of a nested function that belong to
// the enclosing code path — a computed member name and decorators — while its
// parameters and body form their own code path root.
func (b *syntaxFlowBuilder[E]) nestedFunction(node *ast.Node) {
	b.memberHeader(node)
}

func (b *syntaxFlowBuilder[E]) classLike(node *ast.Node) {
	class := node.ClassLikeData()
	if class == nil {
		return
	}
	b.decorators(node)
	b.typeParameters(node)
	if class.HeritageClauses != nil {
		for _, clause := range class.HeritageClauses.Nodes {
			heritage := clause.AsHeritageClause()
			if heritage == nil || heritage.Types == nil {
				continue
			}
			for _, typeNode := range heritage.Types.Nodes {
				b.expr(typeNode)
			}
		}
	}
	if class.Members != nil {
		for _, member := range class.Members.Nodes {
			b.memberHeader(member)
		}
	}
}

// memberHeader evaluates the parts of a class or object member that belong to
// the enclosing code path: its decorators, its computed key, and its type
// annotation. A member with a body of its own — a method, accessor, or
// property initializer — forms its own code path, so its parameters and
// return type are walked there instead; an index signature has no body, so
// its parameter and return types are walked here.
func (b *syntaxFlowBuilder[E]) memberHeader(node *ast.Node) {
	b.decorators(node)
	if ast.IsFunctionLikeDeclaration(node) {
		for _, parameter := range node.Parameters() {
			b.decorators(parameter)
		}
	}
	name := node.Name()
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		b.expr(name.AsComputedPropertyName().Expression)
	}
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		b.expr(node.AsPropertyDeclaration().Type)
	case ast.KindIndexSignature:
		for _, parameter := range node.Parameters() {
			b.expr(parameter.Type())
		}
		b.expr(node.Type())
	}
}

func (b *syntaxFlowBuilder[E]) decorators(node *ast.Node) {
	for _, decorator := range node.Decorators() {
		b.expr(decorator.AsDecorator().Expression)
	}
}

// Original shelf: statements.go

func (b *syntaxFlowBuilder[E]) statements(list *ast.NodeList) {
	if list == nil {
		return
	}
	for _, stmt := range list.Nodes {
		b.statement(stmt)
	}
}

func (b *syntaxFlowBuilder[E]) statement(node *ast.Node) {
	if node == nil {
		return
	}
	b.reachedStatement(node)

	switch node.Kind {
	case ast.KindBlock:
		b.statements(node.AsBlock().Statements)

	case ast.KindVariableStatement:
		b.variableDeclarationList(node.AsVariableStatement().DeclarationList)

	case ast.KindExpressionStatement:
		b.expr(node.AsExpressionStatement().Expression)

	case ast.KindIfStatement:
		b.ifStatement(node)

	case ast.KindWhileStatement:
		b.whileStatement(node)

	case ast.KindDoStatement:
		b.doStatement(node)

	case ast.KindForStatement:
		b.forStatement(node)

	case ast.KindForInStatement, ast.KindForOfStatement:
		b.forInOfStatement(node)

	case ast.KindSwitchStatement:
		b.switchStatement(node)

	case ast.KindTryStatement:
		b.tryStatement(node)

	case ast.KindLabeledStatement:
		b.labeledStatement(node)

	case ast.KindReturnStatement:
		b.expr(node.AsReturnStatement().Expression)
		b.makeReturn()

	case ast.KindThrowStatement:
		b.expr(node.AsThrowStatement().Expression)
		b.makeThrow()

	case ast.KindBreakStatement:
		b.makeBreak(node.AsBreakStatement().Label)

	case ast.KindContinueStatement:
		b.makeContinue(node.AsContinueStatement().Label)

	case ast.KindWithStatement:
		with := node.AsWithStatement()
		b.expr(with.Expression)
		b.statement(with.Statement)

	case ast.KindEmptyStatement, ast.KindDebuggerStatement,
		ast.KindImportDeclaration, ast.KindImportEqualsDeclaration,
		ast.KindFunctionDeclaration:
		// Nothing executes here in the enclosing code path. A function
		// declaration's parameters and body are its own code path root.

	case ast.KindClassDeclaration:
		b.classLike(node)

	case ast.KindExportAssignment:
		b.expr(node.AsExportAssignment().Expression)

	case ast.KindExportDeclaration:
		// Export bindings are resolved statically; nothing is evaluated.

	default:
		node.ForEachChild(func(child *ast.Node) bool {
			b.visitUnknown(child)
			return false
		})
	}
}

func (b *syntaxFlowBuilder[E]) variableDeclarationList(list *ast.Node) {
	if list == nil {
		return
	}
	declList := list.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return
	}
	for _, decl := range declList.Declarations.Nodes {
		b.variableDeclaration(decl)
	}
}

func (b *syntaxFlowBuilder[E]) variableDeclaration(node *ast.Node) {
	if node == nil {
		return
	}
	decl := node.AsVariableDeclaration()
	if decl == nil {
		return
	}
	// A `typeof x` inside the annotation references the variable, so the
	// annotation is walked even though nothing in it runs.
	b.expr(decl.Type)
	name := decl.Name()
	if name != nil && (name.Kind == ast.KindObjectBindingPattern || name.Kind == ast.KindArrayBindingPattern) {
		// Run-time evaluation order: the initializer produces the value
		// first, then the pattern binds element by element.
		b.expr(decl.Initializer)
		b.patternBind(name)
		return
	}
	if decl.Initializer == nil {
		// `let x;` stores no value.
		b.patternReads(name)
		return
	}
	b.patternReads(name)
	b.expr(decl.Initializer)
	b.patternWrites(name)
}

func (b *syntaxFlowBuilder[E]) ifStatement(node *ast.Node) {
	stmt := node.AsIfStatement()
	after := b.newBlock()
	thenEntry := b.newBlock()
	elseEntry := after
	if stmt.ElseStatement != nil {
		elseEntry = b.newBlock()
	}
	b.condition(stmt.Expression, thenEntry, elseEntry)

	b.enter(thenEntry)
	b.statement(stmt.ThenStatement)
	b.link(b.cur, after)

	if stmt.ElseStatement != nil {
		b.enter(elseEntry)
		b.statement(stmt.ElseStatement)
		b.link(b.cur, after)
	}

	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) whileStatement(node *ast.Node) {
	stmt := node.AsWhileStatement()
	test := b.newBlock()
	b.link(b.cur, test)
	b.enter(test)
	after := b.newBlock()
	body := b.newBlock()
	falseEntry := after
	if syntaxFlowIsAlwaysTruthyTest(stmt.Expression) {
		falseEntry = nil
	}
	b.condition(stmt.Expression, body, falseEntry)

	b.pushJump(node, after, test, node)
	b.enter(body)
	b.statement(stmt.Statement)
	b.loop(node)
	b.link(b.cur, test)
	b.popJump()

	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) doStatement(node *ast.Node) {
	stmt := node.AsDoStatement()
	body := b.newBlock()
	b.link(b.cur, body)
	test := b.newBlock()
	after := b.newBlock()

	// A `continue` restarts a `do`…`while` at its test, not at its body, so it
	// is not what starts another iteration — the test passing is.
	b.pushJump(node, after, test, nil)
	b.enter(body)
	b.statement(stmt.Statement)
	b.link(b.cur, test)
	b.popJump()

	b.enter(test)
	b.expr(stmt.Expression)
	b.loop(node)
	b.link(b.cur, body)
	if !syntaxFlowIsAlwaysTruthyTest(stmt.Expression) {
		b.link(b.cur, after)
	}

	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) forStatement(node *ast.Node) {
	stmt := node.AsForStatement()
	if stmt.Initializer != nil {
		if stmt.Initializer.Kind == ast.KindVariableDeclarationList {
			b.variableDeclarationList(stmt.Initializer)
		} else {
			b.expr(stmt.Initializer)
		}
	}
	if stmt.Condition == nil && stmt.Incrementor == nil {
		// ESLint keeps the body of `for (;;) { ... }` on the loop-header
		// segment itself. Model that shape as one self-looping block: events at
		// the top of the body stay on the repeated header, while a branch in the
		// body still creates later blocks that cycle analysis can identify.
		body := b.newBlock()
		b.link(b.cur, body)
		after := b.newBlock()
		b.pushJump(node, after, body, node)
		b.enter(body)
		b.statement(stmt.Statement)
		b.loop(node)
		b.link(b.cur, body)
		b.popJump()
		b.enter(after)
		return
	}

	test := b.newBlock()
	b.link(b.cur, test)
	b.enter(test)
	after := b.newBlock()
	body := b.newBlock()
	falseEntry := after
	if stmt.Condition == nil || syntaxFlowIsAlwaysTruthyTest(stmt.Condition) {
		falseEntry = nil
	}
	b.condition(stmt.Condition, body, falseEntry)

	// The incrementor runs after the body and is laid out before it, so a body
	// that always leaves abruptly does not make it unreachable and lose the
	// fork a throwable node in it makes.
	update := test
	if stmt.Incrementor != nil {
		update = b.newBlock()
		b.enterDisconnected(update)
		b.expr(stmt.Incrementor)
		b.link(b.cur, test)
	}

	b.pushJump(node, after, update, node)
	b.enter(body)
	b.statement(stmt.Statement)
	b.loop(node)
	b.link(b.cur, update)
	b.popJump()

	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) forInOfStatement(node *ast.Node) {
	stmt := node.AsForInOrOfStatement()
	b.expr(stmt.Expression)

	head := b.newBlock()
	b.link(b.cur, head)
	b.enter(head)

	after := b.newBlock()
	b.link(head, after)
	body := b.newBlock()
	b.link(head, body)

	b.pushJump(node, after, head, node)
	b.enter(body)
	if stmt.Initializer != nil {
		// The loop variable takes its value from the iteration, so it binds
		// with no expression of its own; its computed keys and destructuring
		// defaults evaluate once per iteration.
		if stmt.Initializer.Kind == ast.KindVariableDeclarationList {
			declList := stmt.Initializer.AsVariableDeclarationList()
			if declList != nil && declList.Declarations != nil {
				for _, decl := range declList.Declarations.Nodes {
					b.patternBind(decl.Name())
				}
			}
		} else {
			b.patternBind(stmt.Initializer)
		}
	}
	b.statement(stmt.Statement)
	b.loop(node)
	b.link(b.cur, head)
	b.popJump()

	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) switchStatement(node *ast.Node) {
	stmt := node.AsSwitchStatement()
	b.expr(stmt.Expression)

	after := b.newBlock()
	var clauses []*ast.Node
	if stmt.CaseBlock != nil {
		if caseBlock := stmt.CaseBlock.AsCaseBlock(); caseBlock != nil && caseBlock.Clauses != nil {
			clauses = caseBlock.Clauses.Nodes
		}
	}
	if len(clauses) == 0 {
		b.link(b.cur, after)
		b.enter(after)
		return
	}

	bodies := make([]*syntaxFlowBlock[E], len(clauses))
	dispatchFrom := make([]*syntaxFlowBlock[E], len(clauses))
	dispatchEdge := make([]int, len(clauses))
	for i := range dispatchEdge {
		dispatchEdge[i] = -1
	}
	for i := range clauses {
		bodies[i] = b.newBlock()
	}

	defaultIndex := -1
	firstCaseIndex := -1
	testCur := b.cur
	for i, clause := range clauses {
		if clause.Kind == ast.KindDefaultClause {
			defaultIndex = i
			continue
		}
		if firstCaseIndex == -1 {
			firstCaseIndex = i
		}
		b.cur = testCur
		b.expr(clause.AsCaseOrDefaultClause().Expression)
		dispatchFrom[i] = b.cur
		dispatchEdge[i] = b.linkWithCycleBarrier(b.cur, bodies[i], false)
		nextTest := b.newBlock()
		b.link(b.cur, nextTest)
		b.enter(nextTest)
		testCur = b.cur
	}
	if defaultIndex >= 0 {
		dispatchFrom[defaultIndex] = testCur
		dispatchEdge[defaultIndex] = b.linkWithCycleBarrier(testCur, bodies[defaultIndex], false)
	} else {
		b.link(testCur, after)
	}

	b.pushJump(node, after, nil, nil)
	switchJump := len(b.jumps) - 1
	hadSwitchBreak := false
	var fallthroughFrom *syntaxFlowBlock[E]
	for i, clause := range clauses {
		beforeFirstCase := firstCaseIndex == -1 || clause.Kind == ast.KindDefaultClause && i < firstCaseIndex
		barrier := hadSwitchBreak || beforeFirstCase
		if from := dispatchFrom[i]; from != nil && dispatchEdge[i] >= 0 {
			b.setCycleBarrier(from, dispatchEdge[i], barrier)
		}
		b.linkWithCycleBarrier(fallthroughFrom, bodies[i], barrier)
		b.enter(bodies[i])
		b.statements(clause.AsCaseOrDefaultClause().Statements)
		fallthroughFrom = b.cur
		if i >= firstCaseIndex && firstCaseIndex != -1 {
			hadSwitchBreak = b.jumps[switchJump].broken
		}
	}
	b.link(fallthroughFrom, after)
	b.popJump()

	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) tryStatement(node *ast.Node) {
	stmt := node.AsTryStatement()
	hasCatch := stmt.CatchClause != nil
	hasFinally := stmt.FinallyBlock != nil

	frame := &syntaxFlowTryFrame[E]{position: syntaxFlowPosTry, hasFinally: hasFinally}
	if hasCatch {
		frame.catchEntry = b.newBlock()
	}
	if hasFinally {
		frame.finallyEntry = b.newBlock()
	}
	b.tryStack = append(b.tryStack, frame)

	b.statement(stmt.TryBlock)

	// The normal completion of a `try` statement carries on from the end of its
	// `try` block and from the end of its `catch` block, whether or not either
	// of them is somewhere control arrives.
	var normalEnds []*syntaxFlowBlock[E]
	if hasCatch {
		// ESLint routes the end of a `try` block into the `catch` block as well
		// as past the statement.
		b.link(b.cur, frame.catchEntry)
		normalEnds = append(normalEnds, b.cur)
		frame.position = syntaxFlowPosCatch
		frame.thrownForked = false
		b.enter(frame.catchEntry)
		catchClause := stmt.CatchClause.AsCatchClause()
		if catchClause.VariableDeclaration != nil {
			b.patternBind(catchClause.VariableDeclaration.Name())
		}
		b.statement(catchClause.Block)
	}
	normalEnds = append(normalEnds, b.cur)

	b.tryStack = b.tryStack[:len(b.tryStack)-1]

	after := b.newBlock()
	if !hasFinally {
		for _, end := range normalEnds {
			b.link(end, after)
		}
		b.enter(after)
		return
	}

	snapshot := b.snapshotForks()
	normalEntry := b.newBlock()
	for _, end := range normalEnds {
		b.link(end, normalEntry)
	}
	for _, source := range frame.implicit {
		b.link(source, normalEntry)
	}
	b.enter(normalEntry)
	b.statement(stmt.FinallyBlock)
	b.link(b.cur, after)

	if frame.finallyEntry.hasIncoming {
		// The `finally` block also runs on the path that leaves the statement
		// abruptly, so it is laid out a second time. Rewinding the enclosing
		// frames lets that copy fork at the same node the first copy did.
		b.restoreForks(snapshot)
		b.enter(frame.finallyEntry)
		b.statement(stmt.FinallyBlock)
		if b.cur.Reachable {
			// A `return` inside this `try`/`catch` still needs to run every
			// enclosing `finally` on its way out, not just this one — so this
			// second copy's own abrupt exit must keep propagating outward.
			if frame.returnedAny {
				if i := b.returnFrame(); i >= 0 {
					outer := b.tryStack[i]
					outer.returnedAny = true
					b.link(b.cur, outer.finallyEntry)
				} else {
					b.markFinal(b.cur)
				}
			}
			if frame.thrownAny {
				if i := b.throwFrame(); i >= 0 {
					outer := b.tryStack[i]
					outer.thrownAny = true
					b.link(b.cur, syntaxFlowThrowTarget(outer))
				} else {
					b.markThrown(b.cur)
				}
			}
		}
	}

	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) labeledStatement(node *ast.Node) {
	stmt := node.AsLabeledStatement()
	if syntaxFlowIsBreakableStatement(stmt.Statement) {
		// The label belongs to the loop or switch itself; pushJump reads it
		// back off the parent chain.
		b.statement(stmt.Statement)
		return
	}
	after := b.newBlock()
	b.jumps = append(b.jumps, syntaxFlowJumpTarget[E]{labels: []string{stmt.Label.Text()}, breakTo: after})
	b.statement(stmt.Statement)
	b.popJump()
	b.link(b.cur, after)
	b.enter(after)
}

func (b *syntaxFlowBuilder[E]) pushJump(node *ast.Node, breakTo, continueTo *syntaxFlowBlock[E], loop *ast.Node) {
	b.jumps = append(b.jumps, syntaxFlowJumpTarget[E]{labels: syntaxFlowLabelsOf(node), breakTo: breakTo, continueTo: continueTo, loop: loop, breakable: true})
}

func (b *syntaxFlowBuilder[E]) popJump() {
	b.jumps = b.jumps[:len(b.jumps)-1]
}

func (b *syntaxFlowBuilder[E]) makeBreak(label *ast.Node) {
	if !b.cur.Reachable {
		return
	}
	name := ""
	if label != nil {
		name = label.Text()
	}
	for i := len(b.jumps) - 1; i >= 0; i-- {
		target := b.jumps[i]
		if name == "" {
			if !target.breakable {
				continue
			}
		} else if !slices.Contains(target.labels, name) {
			continue
		}
		b.jumps[i].broken = true
		b.link(b.cur, target.breakTo)
		break
	}
	b.makeUnreachable()
}

func (b *syntaxFlowBuilder[E]) makeContinue(label *ast.Node) {
	if !b.cur.Reachable {
		return
	}
	name := ""
	if label != nil {
		name = label.Text()
	}
	for i := len(b.jumps) - 1; i >= 0; i-- {
		target := b.jumps[i]
		if target.continueTo == nil {
			continue
		}
		if name == "" || slices.Contains(target.labels, name) {
			b.loop(target.loop)
			b.link(b.cur, target.continueTo)
			break
		}
	}
	b.makeUnreachable()
}

func (b *syntaxFlowBuilder[E]) makeReturn() {
	if !b.cur.Reachable {
		return
	}
	if i := b.returnFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.returnedAny = true
		b.link(b.cur, frame.finallyEntry)
	} else {
		b.markFinal(b.cur)
	}
	b.makeUnreachable()
}

func (b *syntaxFlowBuilder[E]) makeThrow() {
	if !b.cur.Reachable {
		return
	}
	if i := b.throwFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.thrownAny = true
		b.link(b.cur, syntaxFlowThrowTarget(frame))
	} else {
		b.markThrown(b.cur)
	}
	b.makeUnreachable()
}

// makeYield mirrors ESLint's handling of a suspended generator: the caller may
// resume it with `.return()` or `.throw()`, so both leaving paths are recorded
// before the normal continuation resumes in a fresh block.
func (b *syntaxFlowBuilder[E]) makeYield() {
	if !b.cur.Reachable {
		return
	}
	if i := b.returnFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.returnedAny = true
		b.link(b.cur, frame.finallyEntry)
	} else {
		b.markFinal(b.cur)
	}
	if i := b.throwFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.thrownAny = true
		b.link(b.cur, syntaxFlowThrowTarget(frame))
	} else {
		b.markThrown(b.cur)
	}
	next := b.newBlock()
	b.link(b.cur, next)
	b.enter(next)
}

// Original shelf: patterns.go

// patternReads emits what an identifier or member assignment target evaluates
// before the assigned expression: the receiver and computed key of a member
// target, and the throwable fork ESLint records for naming the target inside a
// `try` block. Destructuring targets are laid out by patternBind instead.
func (b *syntaxFlowBuilder[E]) patternReads(node *ast.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		// A compound-assignment or update-expression target reads the
		// variable before writing it; a plain assignment target does not,
		// but ESLint still counts naming it as a point where the enclosing
		// `try` block can throw — so the handler sees the value the target
		// had before the assignment.
		b.read(node)
		if syntaxFlowIsThrowableIdentifier(node) {
			b.firstThrowableFork()
		}

	case ast.KindParenthesizedExpression:
		b.patternReads(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternReads(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternReads(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternReads(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternReads(node.AsTypeAssertion().Expression)

	default:
		// A member target such as `obj.x` or `obj[k]` evaluates its receiver.
		b.expr(node)
	}
}

// patternWrites emits the write of an identifier assignment target after its
// assigned expression. Member targets write no variable.
func (b *syntaxFlowBuilder[E]) patternWrites(node *ast.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		b.write(node)

	case ast.KindParenthesizedExpression:
		b.patternWrites(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternWrites(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternWrites(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternWrites(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternWrites(node.AsTypeAssertion().Expression)
	}
}

// patternBind lays out a destructuring target the way it evaluates at run
// time, after the assigned expression has produced its value: each element in
// source order evaluates its computed key, forks around its default, and
// writes its target.
func (b *syntaxFlowBuilder[E]) patternBind(node *ast.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		b.read(node)
		if syntaxFlowIsThrowableIdentifier(node) {
			b.firstThrowableFork()
		}
		b.write(node)

	case ast.KindOmittedExpression:
		// An array hole binds nothing.

	case ast.KindParenthesizedExpression:
		b.patternBind(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternBind(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternBind(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternBind(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternBind(node.AsTypeAssertion().Expression)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		pattern := node.AsBindingPattern()
		if pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			binding := element.AsBindingElement()
			if binding == nil {
				continue
			}
			if binding.PropertyName != nil && binding.PropertyName.Kind == ast.KindComputedPropertyName {
				b.expr(binding.PropertyName.AsComputedPropertyName().Expression)
			}
			b.bindWithDefault(binding.Name(), binding.Initializer)
		}

	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				assignment := property.AsPropertyAssignment()
				if name := assignment.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					b.expr(name.AsComputedPropertyName().Expression)
				}
				b.patternBind(assignment.Initializer)
			case ast.KindShorthandPropertyAssignment:
				shorthand := property.AsShorthandPropertyAssignment()
				b.bindWithDefault(shorthand.Name(), shorthand.ObjectAssignmentInitializer)
			case ast.KindSpreadAssignment:
				b.patternBind(property.AsSpreadAssignment().Expression)
			}
		}

	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			b.patternBind(element)
		}

	case ast.KindSpreadElement:
		b.patternBind(node.AsSpreadElement().Expression)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			b.bindWithDefault(binary.Left, binary.Right)
			return
		}
		b.expr(node)

	default:
		// A member target such as `obj.x` or `obj[k]` evaluates its receiver.
		b.expr(node)
	}
}

// bindWithDefault lays out one pattern element: a fork around the default
// value, which only runs when the incoming value is undefined, then the
// target's own binding.
func (b *syntaxFlowBuilder[E]) bindWithDefault(target *ast.Node, fallback *ast.Node) {
	if fallback != nil {
		join := b.newBlock()
		b.link(b.cur, join)
		withFallback := b.newBlock()
		b.link(b.cur, withFallback)

		b.enter(withFallback)
		b.expr(fallback)
		b.link(b.cur, join)

		b.enter(join)
	}
	b.patternBind(target)
}

// syntaxFlowIsDestructuringTarget reports whether an assignment's left-hand side is an
// object or array pattern.
func syntaxFlowIsDestructuringTarget(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node != nil &&
		(node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindArrayLiteralExpression)
}

// Original shelf: helpers.go

func syntaxFlowIsBreakableStatement(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindSwitchStatement, ast.KindWhileStatement, ast.KindDoStatement,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}

// syntaxFlowLabelsOf collects every label wrapped directly around a loop or switch, so
// `outer: inner: while (…)` resolves `continue outer` and `break outer` as
// well as `inner`.
//
// Keeping the outer label is a known deviation. ESLint attaches only the
// innermost one, so a `continue` naming an outer label finds no loop there and
// its back edge is dropped. It surfaces where a loop's exit hangs off that edge
// — a `do…while`, whose test runs only from it — leaving what follows the loop
// reachable here and unreachable in ESLint, which `no-useless-return`,
// `no-unreachable-loop` and `no-useless-assignment` all read.
func syntaxFlowLabelsOf(node *ast.Node) []string {
	var labels []string
	current := node
	for parent := current.Parent; parent != nil && parent.Kind == ast.KindLabeledStatement; parent = parent.Parent {
		labeled := parent.AsLabeledStatement()
		if labeled.Statement != current {
			break
		}
		labels = append(labels, labeled.Label.Text())
		current = parent
	}
	return labels
}

// syntaxFlowIsAlwaysTruthyTest mirrors ESLint's `getBooleanValueIfSimpleConstant`: only a
// literal test can mark the loop infinite, so the code after it is unreachable.
func syntaxFlowIsAlwaysTruthyTest(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindRegularExpressionLiteral:
		return true
	case ast.KindNumericLiteral:
		// Upstream runs its own `NormalizeNumericLiteral` over this text before comparing. On our
		// substrate that normalizer is a no-op, because typescript-go's scanner has already stored
		// the canonical decimal value rather than the source spelling. Probed on our parser rather
		// than assumed: `0x0`, `0.0`, `0e10`, `0b0`, `0o0`, `0_0`, `00`, and `.0` all arrive as
		// exactly "0", and `0x1`, `1e2`, `1_0`, `1e400` arrive as "1", "100", "10", "Infinity".
		// `internal/rules/core/no_compare_neg_zero.go` reached the same conclusion independently
		// and reads `.Text == "0"` the same way.
		return node.AsNumericLiteral().Text != "0"
	case ast.KindBigIntLiteral:
		// A BigInt literal is NOT normalized the way a numeric one is, which is the asymmetry that
		// makes this branch load-bearing rather than redundant. Probed on our parser: `0n` arrives
		// as "0n" and `0x0n` as "0x0n" — the `n` suffix survives, and a hex or binary spelling
		// survives with it, though `0b0n` and `0o0n` do arrive as "0n". So a bare `.Text != "0"`
		// would call every zero BigInt truthy and mark the code after `while (0n)` unreachable.
		return syntaxFlowNormalizeBigIntLiteral(node.AsBigIntLiteral().Text) != "0"
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text != ""
	}
	return false
}

// syntaxFlowIsThrowableIdentifier reports whether completing this identifier is one of
// the points ESLint treats as able to throw inside a `try` block. Names that
// only declare or label something are excluded, as are JSX names, which ESLint
// sees as a separate node type.
func syntaxFlowIsThrowableIdentifier(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if ast.IsJsxTagName(node) {
		return false
	}
	switch parent.Kind {
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement,
		ast.KindArrayBindingPattern, ast.KindImportSpecifier, ast.KindExportSpecifier,
		ast.KindImportClause, ast.KindNamespaceImport, ast.KindNamespaceExport,
		ast.KindJsxAttribute, ast.KindJsxNamespacedName:
		return false
	case ast.KindBindingElement:
		binding := parent.AsBindingElement()
		if binding.DotDotDotToken != nil || binding.PropertyName == node {
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectBindingPattern
	case ast.KindCatchClause:
		return false
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindClassDeclaration, ast.KindClassExpression, ast.KindVariableDeclaration,
		ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return parent.Name() != node
	}
	return true
}

// syntaxFlowNormalizeBigIntLiteral writes a BigInt literal's text as its decimal digits, which is what
// JavaScript's `String(value)` produces and what ESLint compares against.
//
// Inlined from rslint's `utils.NormalizeBigIntLiteral` rather than imported, because that package
// is the linter's whole shared shelf and this graph needs two lines of it. The numeric sibling was
// dropped entirely — see `syntaxFlowIsAlwaysTruthyTest` for the measurement that says why one survived and
// one did not.
//
// `big.Int` rather than `strconv.ParseInt` because BigInt is arbitrary precision by definition and
// routinely exceeds int64: `0x10000000000000000n` is 2^64 and reaches this function unnormalized.
func syntaxFlowNormalizeBigIntLiteral(text string) string {
	digits := strings.TrimSuffix(text, "n")
	if value, ok := new(big.Int).SetString(digits, 0); ok {
		return value.String()
	}
	return digits
}

// Original shelf: roots.go

// syntaxFlowIsRoot reports whether node owns a control-flow graph of its own.
func syntaxFlowIsRoot(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration:
		return true
	case ast.KindPropertyDeclaration:
		return node.AsPropertyDeclaration().Initializer != nil
	}
	return false
}

// syntaxFlowRootOf returns the code path root node executes in.
func syntaxFlowRootOf(node *ast.Node) *ast.Node {
	previous := node
	var decorator *ast.Node
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindDecorator {
			decorator = current
		}
		if current.Kind == ast.KindSourceFile {
			return current
		}
		if syntaxFlowIsRoot(current) && syntaxFlowRunsInsideRoot(current, previous, decorator) {
			return current
		}
		previous = current
	}
	return nil
}

// syntaxFlowRunsInsideRoot reports whether a direct child of a code path root runs in
// that root rather than in the surrounding one. A member's name runs where the
// member is declared, and so do the decorators on the member and on each of its
// parameters.
func syntaxFlowRunsInsideRoot(root *ast.Node, child *ast.Node, decorator *ast.Node) bool {
	switch root.Kind {
	case ast.KindPropertyDeclaration:
		return root.AsPropertyDeclaration().Initializer == child
	case ast.KindClassStaticBlockDeclaration:
		return root.AsClassStaticBlockDeclaration().Body == child
	}
	if root.Name() == child {
		return false
	}
	if decorator != nil && (decorator.Parent == root || decorator.Parent == child) {
		return false
	}
	return true
}

// syntaxFlowRootSummary is one code path root and what its own code holds, so a rule can tell whether the
// root's graph could give it anything to report before paying to build it.
//
// "Own code" stops at a nested root, which the builder lays out as a graph of its own and never
// descends into. A statement always runs in its nearest enclosing root, since a statement can only
// sit in a function body, a class static block or the file, so the nearest one is what is recorded.
type syntaxFlowRootSummary struct {
	Node *ast.Node

	// BareReturn is whether the root's own code holds a `return;` with no value.
	BareReturn bool

	// Loops are the loop statement kinds the root's own code holds, each once, in the order first met.
	Loops []ast.Kind
}

// syntaxFlowIndexRoots lists every code path root in a file, the file first and the rest in preorder, which is
// the order a rule walking the tree for syntaxFlowIsRoot meets them in, with what each one's own code holds.
//
// One walk serves every rule that builds a graph per root. Three such rules each walked the file to
// list its roots and then built a graph for every one, though each can only report from the few
// roots holding what it looks for: on a cold ahra run, 4.4% of roots hold a bare return and 7.6% a
// loop (#tmn9n27).
func syntaxFlowIndexRoots(sourceFile *ast.Node) []syntaxFlowRootSummary {
	roots := []syntaxFlowRootSummary{{Node: sourceFile}}
	var visit func(node *ast.Node, owner int)
	visit = func(node *ast.Node, owner int) {
		node.ForEachChild(func(child *ast.Node) bool {
			current := owner
			if syntaxFlowIsRoot(child) {
				roots = append(roots, syntaxFlowRootSummary{Node: child})
				current = len(roots) - 1
			}
			switch child.Kind {
			case ast.KindReturnStatement:
				if child.AsReturnStatement().Expression == nil {
					roots[current].BareReturn = true
				}
			case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindForInStatement,
				ast.KindForOfStatement:
				if !slices.Contains(roots[current].Loops, child.Kind) {
					roots[current].Loops = append(roots[current].Loops, child.Kind)
				}
			}
			visit(child, current)
			return false
		})
	}
	visit(sourceFile, 0)
	return roots
}
