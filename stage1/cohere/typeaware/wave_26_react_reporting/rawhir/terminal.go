// The terminal set.
//
// 21 variants. Upstream has 22; the one absent is `PrunedScope`, which four later passes
// (`flattenReactiveLoopsHIR`, `flattenScopesWithHooksOrUseHIR`, `pruneUnusedScopes`,
// `pruneAlwaysInvalidatingScopes`) construct. All four are ported, and each records its decision as
// `ReactiveScopeBlock.Pruned` rather than as a terminal, so nothing here needs the variant. `Scope` was absent for
// the same reason until `BuildReactiveScopeTerminals` landed to construct it; see that terminal's
// comment for why the cost was paid at that moment and not earlier. See high_level_intermediate_representation.go's package comment for
// the rule that puts JSX in the core instruction set and kept these outside it.
//
// # Fallthrough, and why every structured terminal carries one
//
// A terminal names its successors by id. Most also name a `Fallthrough`: the block where control
// resumes after the whole construct completes, whichever way it went.
//
// That field is not redundant with the successor edges, and it is the single most important thing
// to understand about this IR. It is what distinguishes an `If` from two arbitrary jumps that
// happen to reconverge. Without it, the graph records that control forked and later merged, and
// recovering "this was an if-statement, and this block is what comes after it" means finding the
// immediate post-dominator and hoping it corresponds to a syntactic construct. With it, the
// structure the source had is preserved through every pass that maintains the field.
//
// Two consequences the next reader should hold:
//
//   - Fallthrough is NOT a control-flow edge. Control reaches it through the arms, not from the
//     terminal directly. `EachSuccessor` therefore does not yield it by default, and
//     `MarkPredecessors` does not lay an edge to it. A dataflow analysis that treated it as an edge
//     would propagate values along a path that does not exist.
//   - It is nonetheless reachable-from in the graph sense, so a pass that walks only real edges
//     still visits it, via whichever arm reaches it.
package high_level_intermediate_representation

// Terminal is how control leaves a basic block.
//
// The set is closed: the unexported marker method means only this package can add a variant.
type Terminal interface {
	terminal()
}

// ---------------------------------------------------------------------------
// Leaving the function
// ---------------------------------------------------------------------------

// Return returns a value.
type Return struct {
	Value Place
	Order EvaluationOrder
}

// Throw throws a value.
type Throw struct {
	Value Place
	Order EvaluationOrder
}

// Unreachable is a block control cannot arrive at.
//
// Lowering lays out the code after an abrupt exit rather than dropping it, matching what
// `controlflow` does and for the same reason: a rule may want to ask what would have run.
type Unreachable struct{ Order EvaluationOrder }

// Unsupported terminates a block whose construct lowering does not model.
type Unsupported struct{ Order EvaluationOrder }

// ---------------------------------------------------------------------------
// Unconditional
// ---------------------------------------------------------------------------

// Goto jumps to a block.
type Goto struct {
	Block   BlockId
	Variant GotoVariant
	Order   EvaluationOrder
}

// GotoVariant records why a jump exists, which a later pass reconstructing structure needs.
type GotoVariant uint8

const (
	// GotoVariantBreak leaves a construct.
	GotoVariantBreak GotoVariant = iota
	// GotoVariantContinue begins the next iteration of a loop.
	GotoVariantContinue
	// GotoVariantTry leaves a try block.
	GotoVariantTry
)

// ---------------------------------------------------------------------------
// Branching
// ---------------------------------------------------------------------------

// If is a statement-level conditional.
type If struct {
	Test        Place
	Consequent  BlockId
	Alternate   BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// Branch is a conditional inside a value-producing construct.
//
// Distinct from If because the arms produce a value that the fallthrough consumes. A pass that
// rebuilds an expression cares which it is; a pass that only walks control flow can treat them
// alike, and `EachSuccessor` does.
type Branch struct {
	Test        Place
	Consequent  BlockId
	Alternate   BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// Switch dispatches on a value.
//
// Cases are in source order. A case with a nil Test is the `default`. Fallthrough between cases is
// represented by the case blocks' own terminals, not here, so a case that falls through ends in a
// Goto to the next case block and one that breaks ends in a Goto to Fallthrough.
type Switch struct {
	Test        Place
	Cases       []SwitchCase
	Fallthrough BlockId
	Order       EvaluationOrder
}

// SwitchCase is one arm. Test is nil for `default`.
type SwitchCase struct {
	Test  *Place
	Block BlockId
}

// ---------------------------------------------------------------------------
// Loops
// ---------------------------------------------------------------------------

// While is a `while` loop: test, then body.
type While struct {
	Test        BlockId
	Loop        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// DoWhile is a `do...while` loop: body, then test.
type DoWhile struct {
	Loop        BlockId
	Test        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// For is a C-style `for` loop.
//
// Update is InvalidBlock when the loop has no update clause.
type For struct {
	Init        BlockId
	Test        BlockId
	Update      BlockId
	Loop        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// ForOf is a `for...of` loop.
//
// The iterator protocol is in the blocks: Init holds the GetIterator, Test holds the IteratorNext
// and the done check. Keeping them as instructions rather than folding them into the terminal is
// what lets an effect pass see that iterating mutates the iterator.
type ForOf struct {
	Init        BlockId
	Test        BlockId
	Loop        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// ForIn is a `for...in` loop.
type ForIn struct {
	Init        BlockId
	Loop        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// ---------------------------------------------------------------------------
// Value-producing control flow
// ---------------------------------------------------------------------------

// Logical is `&&`, `||`, or `??`.
//
// These are terminals rather than BinaryExpression instructions because they short-circuit: the
// right operand is not evaluated on every path, which is control flow by definition.
type Logical struct {
	Operator    string
	Test        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// Ternary is `a ? b : c`.
type Ternary struct {
	Test        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// Optional is an optional chain link, `a?.b` or `a?.()`.
//
// Optional reports whether this specific link is the one that tests, since `a?.b.c` has one testing
// link and one that merely rides the short circuit.
type Optional struct {
	Optional    bool
	Test        BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// Sequence is one element of a comma expression.
type Sequence struct {
	Block       BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// ---------------------------------------------------------------------------
// Labels and exceptions
// ---------------------------------------------------------------------------

// Label is a labeled statement, the target of a labeled break or continue.
type Label struct {
	Block       BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// Try enters a try statement.
//
// HandlerBinding is the catch parameter, nil for `catch {}` or a bare `try...finally`.
//
// # The finally divergence, stated where it will be found
//
// cohere's `controlflow` lays a `finally` body out TWICE - once for normal completion, once for the
// abrupt path - and both copies carry the same source positions. This IR does NOT. The finally body
// is lowered once, and the paths that must run it jump to it.
//
// The divergence is deliberate and it is the main reason this package does not reuse that graph.
// Two layouts of one syntactic body give a value two definitions that are not a merge of each
// other; single-assignment construction over that mints a phi at a point the source has no join,
// and every effect computed downstream inherits the error. ESLint's shape is right for asking
// whether an event occurs on every path, which is what that graph's consumers ask. It is wrong for
// naming values, which is what this one does.
//
// The cost is that a pass wanting the ESLint answer cannot get it from here, and must use the other
// graph. That is the correct outcome: two graphs, two questions, neither bent.
type Try struct {
	Block          BlockId
	HandlerBinding *Place
	Handler        BlockId
	Fallthrough    BlockId
	Order          EvaluationOrder
}

// MaybeThrow marks a point where an exception can leave the block.
//
// Handler is InvalidBlock when nothing in this function catches, meaning the exception leaves the
// function.
//
// # Why this exists rather than an edge from every throwable instruction
//
// `controlflow` forks to the handler at the FIRST throwable node in a try body, so later statements
// are bypassed on the throwing path, and gives a body that cannot throw NO edge to its catch at all.
// Both are right for reachability and wrong for a mutation lattice, which needs the handler to
// observe the state the body actually left behind - including from a body the analysis believes
// cannot throw, because "cannot throw" is a judgement that a later, more precise pass may revise.
//
// So this IR places one MaybeThrow at the end of a run of throwable instructions, and the handler
// always has an edge. A block ending in MaybeThrow has two successors: Continuation on the normal
// path, Handler on the exceptional one.
type MaybeThrow struct {
	Continuation BlockId
	Handler      BlockId
	Order        EvaluationOrder
}

// ---------------------------------------------------------------------------
// Reactive scopes
// ---------------------------------------------------------------------------

// Scope begins a reactive scope: the values in it are memoized together as one unit.
//
// Built by `BuildReactiveScopeTerminals`, which is the only thing that constructs one. Before that
// pass runs a function contains none, and every consumer written against a freshly lowered graph
// may still assume the terminal set it knew.
//
// # Why this variant exists now, when the package comment said it might never
//
// `high_level_intermediate_representation.go` records the rule: a variant belongs in the core set when LOWERING must produce it, and
// behind a seam when only a LATER PASS would, and it names the moment to pay the cost as "if
// reactive scopes are ever built". They are built. Four stages before this one declined to add the
// variant early and were right to: a terminal nothing constructs is worse than no terminal, because
// consumers switch on it and get an arm that never fires. `Optional` is the standing example in this
// package. This variant is added in the same commit as its producer, and `TestScopeTerminalsAreBuilt`
// pins the constructed count so it can never quietly become a second one.
//
// # The shape, and the one divergence from React
//
// Upstream's terminal embeds the scope OBJECT (`scope: ReactiveScope`); oxc holds a `ScopeId`
// (`react_compiler_hir/mod.rs:452`). This holds the id, matching oxc, for a reason that is about
// this tree rather than a preference between them: reactive scopes live in a side table here
// (`ReactiveScopes`, `AlignedScopes`, `MergedScopes`) precisely so a widening is one write rather
// than a fan-out, and embedding the object would reintroduce the aliasing that `scopes.go` declined
// and put two copies of a range in the graph. A consumer wanting the range asks the table it already
// holds. Recorded as a divergence because a reader diffing against the bundle will see it.
//
// `Block` is where the scope's body begins and is a real successor. `Fallthrough` is where control
// resumes after the scope completes, and is NOT an edge, exactly as for every other structured
// terminal. See this file's header.
type Scope struct {
	Scope       ScopeId
	Block       BlockId
	Fallthrough BlockId
	Order       EvaluationOrder
}

// ---------------------------------------------------------------------------

func (*Return) terminal()      {}
func (*Throw) terminal()       {}
func (*Unreachable) terminal() {}
func (*Unsupported) terminal() {}
func (*Goto) terminal()        {}
func (*If) terminal()          {}
func (*Branch) terminal()      {}
func (*Switch) terminal()      {}
func (*While) terminal()       {}
func (*DoWhile) terminal()     {}
func (*For) terminal()         {}
func (*ForOf) terminal()       {}
func (*ForIn) terminal()       {}
func (*Logical) terminal()     {}
func (*Ternary) terminal()     {}
func (*Optional) terminal()    {}
func (*Sequence) terminal()    {}
func (*Label) terminal()       {}
func (*Try) terminal()         {}
func (*MaybeThrow) terminal()  {}
func (*Scope) terminal()       {}
