// Rewriting a ReactiveFunction: the transform half of the walk.
//
// This is the second half of React's `visitors.ts` (`ReactiveFunctionTransform`, lines 265-575 at
// `react_conformance.UpstreamSha`). `reactive_visitor.go` shipped the read-only half first and
// deliberately left this out, on the grounds that a rewrite surface with no consumer is the shape
// this package keeps finding declared and never constructed.
//
// It has a consumer now. `pruneNonEscapingScopes` ends in a transform that replaces a scope with no
// memoized output by its own instructions, so the judgment that deferred this is spent rather than
// reversed.
//
// # What a transform returns, and why `keep` is not the same as `replace` with the same value
//
// Each hook answers with a `ReactiveTransformed`: keep the statement, remove it, replace it with
// one statement, or replace it with several. The four are upstream's exactly, and the distinction
// between `keep` and a `replace` carrying the original is load-bearing for cost rather than
// meaning: a block where every statement is kept is never reallocated, which is the ordinary case
// on a pass that rewrites a handful of scopes in a function holding hundreds of statements.
//
// # The lazy rebuild, which is upstream's and is the whole implementation
//
// A block is rewritten in place only if some statement was not kept. Until that happens no new
// slice exists; on the first non-keep the prefix is copied and every later statement appends to the
// copy. Getting this wrong in the direction of always copying is merely slow. Getting it wrong in
// the other direction -- appending to the original while iterating it -- is a corruption that
// depends on statement order, which is why the prefix copy is taken at the index rather than after
// the loop.
package high_level_intermediate_representation

// ReactiveTransformedKind is what a transform hook decided about one statement.
type ReactiveTransformedKind uint8

const (
	// ReactiveTransformKeep leaves the statement as it is. The block is not reallocated.
	ReactiveTransformKeep ReactiveTransformedKind = iota
	// ReactiveTransformRemove deletes the statement.
	ReactiveTransformRemove
	// ReactiveTransformReplace swaps the statement for exactly one other.
	ReactiveTransformReplace
	// ReactiveTransformReplaceMany swaps the statement for zero or more others.
	//
	// Distinct from Remove even when the replacement is empty, because the two say different things
	// about intent and a reader of a pass should not have to infer which was meant from a length.
	ReactiveTransformReplaceMany
)

// ReactiveTransformed is one hook's answer.
//
// A struct rather than a sum type because Go has no unions and an interface here would allocate on
// every statement of every block. `Statement` is meaningful only for Replace and `Statements` only
// for ReplaceMany; the constructors below are the intended way to build one.
type ReactiveTransformed struct {
	Kind       ReactiveTransformedKind
	Statement  ReactiveStatement
	Statements []ReactiveStatement
}

// KeepStatement leaves a statement untouched.
func KeepStatement() ReactiveTransformed {
	return ReactiveTransformed{Kind: ReactiveTransformKeep}
}

// RemoveStatement deletes a statement.
func RemoveStatement() ReactiveTransformed {
	return ReactiveTransformed{Kind: ReactiveTransformRemove}
}

// ReplaceStatement swaps a statement for one other.
func ReplaceStatement(statement ReactiveStatement) ReactiveTransformed {
	return ReactiveTransformed{Kind: ReactiveTransformReplace, Statement: statement}
}

// ReplaceStatements swaps a statement for zero or more others.
//
// This is how a scope is flattened: `pruneNonEscapingScopes` answers with the scope's own
// instructions, which splices the body into the enclosing block and drops the scope wrapper.
func ReplaceStatements(statements []ReactiveStatement) ReactiveTransformed {
	return ReactiveTransformed{Kind: ReactiveTransformReplaceMany, Statements: statements}
}

// ReactiveTransformer is the set of rewrite hooks a pass installs.
//
// Every field is optional and a nil hook means "keep, and recurse normally". A non-nil hook
// receives the node plus a `traverse` closure, exactly as `ReactiveVisitor` does: the hook decides
// whether to descend, and separately what to do with the statement it was handed.
//
// Those two decisions are independent and both matter. `pruneNonEscapingScopes` traverses a scope
// (to rewrite what is inside it) and THEN decides whether the scope itself survives.
type ReactiveTransformer struct {
	// Instruction rewrites an instruction statement.
	Instruction func(statement *ReactiveInstructionStatement, traverse func()) ReactiveTransformed
	// Scope rewrites a scope block.
	Scope func(scope *ReactiveScopeBlock, traverse func()) ReactiveTransformed
	// Terminal rewrites a terminal statement.
	Terminal func(statement *ReactiveTerminalStatement, traverse func()) ReactiveTransformed
}

// TransformReactiveFunction rewrites a function body in place.
//
// In place, matching upstream, because every consumer of this walks a function it owns and a copy
// would double the peak memory of a pass whose whole job is to make the tree smaller. The function
// is mutated; a caller wanting the original keeps its own copy.
func TransformReactiveFunction(function *ReactiveFunction, transformer ReactiveTransformer) {
	if function == nil {
		return
	}
	walker := reactiveTransformWalker{transformer: transformer}
	function.Body = walker.transformBlock(function.Body)
}

type reactiveTransformWalker struct {
	transformer ReactiveTransformer
}

// transformBlock returns the rewritten block, reallocating only if something was not kept.
func (w *reactiveTransformWalker) transformBlock(block ReactiveBlock) ReactiveBlock {
	var rebuilt ReactiveBlock
	for index, statement := range block {
		transformed := w.transformStatement(statement)

		switch transformed.Kind {
		case ReactiveTransformKeep:
			if rebuilt != nil {
				rebuilt = append(rebuilt, statement)
			}
		case ReactiveTransformRemove:
			if rebuilt == nil {
				// The prefix is copied at THIS index rather than after the loop, because every
				// later statement appends to the copy and the original must not be written while
				// it is still being read.
				rebuilt = append(ReactiveBlock{}, block[:index]...)
			}
		case ReactiveTransformReplace:
			if rebuilt == nil {
				rebuilt = append(ReactiveBlock{}, block[:index]...)
			}
			if transformed.Statement != nil {
				rebuilt = append(rebuilt, transformed.Statement)
			}
		case ReactiveTransformReplaceMany:
			if rebuilt == nil {
				rebuilt = append(ReactiveBlock{}, block[:index]...)
			}
			rebuilt = append(rebuilt, transformed.Statements...)
		}
	}
	if rebuilt == nil {
		return block
	}
	return rebuilt
}

func (w *reactiveTransformWalker) transformStatement(statement ReactiveStatement) ReactiveTransformed {
	switch shape := statement.(type) {
	case *ReactiveInstructionStatement:
		if w.transformer.Instruction != nil {
			return w.transformer.Instruction(shape, func() {})
		}
		return KeepStatement()

	case *ReactiveScopeBlock:
		if w.transformer.Scope != nil {
			return w.transformer.Scope(shape, func() {
				shape.Instructions = w.transformBlock(shape.Instructions)
			})
		}
		shape.Instructions = w.transformBlock(shape.Instructions)
		return KeepStatement()

	case *ReactiveTerminalStatement:
		if w.transformer.Terminal != nil {
			return w.transformer.Terminal(shape, func() { w.transformTerminalBlocks(shape) })
		}
		w.transformTerminalBlocks(shape)
		return KeepStatement()
	}
	return KeepStatement()
}

// transformTerminalBlocks rewrites every block a terminal contains, in place on the terminal.
//
// Each arm assigns the result back rather than discarding it, which is the difference between a
// transform and a visitor: `transformBlock` returns a possibly-new slice and a terminal holding the
// old one would silently keep the unrewritten body. A missing assignment here is invisible to any
// test that only counts what the walk reaches, so each arm is written out rather than generated.
func (w *reactiveTransformWalker) transformTerminalBlocks(statement *ReactiveTerminalStatement) {
	switch shape := statement.Terminal.(type) {
	case *ReactiveIf:
		shape.Consequent = w.transformBlock(shape.Consequent)
		if shape.Alternate != nil {
			rewritten := w.transformBlock(*shape.Alternate)
			shape.Alternate = &rewritten
		}
	case *ReactiveSwitch:
		for index := range shape.Cases {
			if shape.Cases[index].Block != nil {
				rewritten := w.transformBlock(*shape.Cases[index].Block)
				shape.Cases[index].Block = &rewritten
			}
		}
	case *ReactiveFor:
		shape.Loop = w.transformBlock(shape.Loop)
	case *ReactiveForOf:
		shape.Loop = w.transformBlock(shape.Loop)
	case *ReactiveForIn:
		shape.Loop = w.transformBlock(shape.Loop)
	case *ReactiveWhile:
		shape.Loop = w.transformBlock(shape.Loop)
	case *ReactiveDoWhile:
		shape.Loop = w.transformBlock(shape.Loop)
	case *ReactiveLabelTerminal:
		shape.Block = w.transformBlock(shape.Block)
	case *ReactiveTry:
		shape.Block = w.transformBlock(shape.Block)
		shape.Handler = w.transformBlock(shape.Handler)
	}
}
