// ReactiveFunction: the control-flow graph turned back into a tree.
//
// This is React's `BuildReactiveFunction` (`ReactiveScopes/BuildReactiveFunction.ts` at the pinned
// upstream SHA, 1,486 lines). oxc transcribes it at
// `oxc_react_compiler/src/react_compiler_reactive_scopes/build_reactive_function.rs`; where they
// disagree React wins, and each disagreement is recorded at the line that resolves it.
//
// # Why this exists, which is a question the earlier stages did not have to answer
//
// Every pass in this package so far has read the HIR and written a side table. This one converts the
// IR itself into a different shape, and the reason is that the remaining work is not analysis. It is
// reconstruction: `validatePreservedManualMemoization` and the eleven passes before it all run on a
// TREE that resembles the source, not on a graph. A graph says "control reached here from three
// places"; a tree says "this is a while loop whose body is these statements". Getting from the first
// to the second is what this file does.
//
// The measurement that made this stage exist: `preserve-manual-memoization` was scoped as a rule and
// declined, because the rule runs on this structure and this structure had no representation here.
// The gap was not the rule's 723 lines, it was the ~7,000 lines of pipeline underneath it.
//
// # What a ReactiveFunction is, in one paragraph
//
// A `ReactiveBlock` is a flat list of `ReactiveStatement`. A statement is one of four things: an
// instruction, a terminal (which CONTAINS its own nested blocks, so an `if` holds its arms), a scope
// block, or a pruned scope block. That containment is the whole difference from the HIR, where an
// `If` terminal names its arms by id and the arms live in a flat block table.
//
// The second difference is `ReactiveValue`. In the HIR, `a && b`, `a ? b : c`, `(a, b)` and `a?.b`
// are TERMINALS -- control-flow constructs with value blocks. Here they are VALUES, nested inside a
// single instruction. That is what lets the tree be printed back as an expression.
//
// # Three-level input check, run before this file was written
//
// The instrument this package uses is: does the type exist, is it ever constructed, and does the
// pass run on the IR we have. The third level is trivially satisfied here because this file BUILDS
// the IR in question. The initial check found four gaps; optional lowering has since closed one:
//
//	terminal        declared   constructed   how this pass handles it
//	Optional             yes           yes   rebuilds ReactiveOptionalValue
//	Sequence             yes             0   shares an arm with ternary/logical
//	MaybeThrow           yes             0   explicitly FLATTENED AWAY upstream
//	PrunedScope           no             0   shares an arm with scope
//
// `Optional` now takes upstream's `visitValueBlockTerminal` route, including recursive rebuilding of
// nested chains. `MaybeThrow` is the interesting remaining case: upstream says the reactive tree
// does not model maybe-throw semantics, so these terminals flatten away.
//
// `PrunedScope` remains undeclared. This pass would emit one only if it consumed one, and it does
// not construct any -- see `ReactiveScopeBlock.Pruned`, which exists as a field because the four
// upstream passes that would set it are downstream of here. Declaring the terminal to satisfy this
// file would create a switchable variant nothing builds.
//
// # DIVERGENCE FROM React: labels are emitted for every terminal there and here
//
// Upstream's own comment says it "naively emits labels for *all* terminals: see PruneUnusedLabels
// which removes unnecessary labels". That pruning pass is not in this tree, so this output carries
// labels a printed program would not want. Recorded as `ReactiveFunctionGapUnprunedLabels` rather than fixed
// here, because fixing it in this pass would mean diverging from upstream in a place where the next
// stage is expected to land.
//
// # Termination is a scheduled walk, not a fixpoint
//
// There is no iteration to convergence anywhere in this file. The driver walks each block once,
// and the `scheduled` set is what makes that true in the presence of joins: a block that a parent
// has already committed to emitting is not emitted again by a child, which instead emits a break.
// `emitted` is the independent check on that -- upstream keeps it purely to abort if a block is
// generated twice, and it is reproduced here for the same reason, as `ReactiveBuildResult.Emitted`.
//
// The recursion is over the block graph and is bounded by it because every visit either consumes a
// block or stops at a scheduled one. `TestReactiveBuildVisitsEachBlockOnce` pins that empirically
// rather than leaving it as an argument.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// ReactiveFunction is the tree form of a lowered function.
//
// Mirrors upstream's type, minus `env` and `directives`. `env` is an Environment, which this package
// does not have -- `Function` is the unit every pass here takes, for the reason `scopes.go` records.
// `directives` is a codegen concern and codegen is explicitly never built.
type ReactiveFunction struct {
	// Name is the function's name, empty for an anonymous function.
	Name string
	// Kind records the component/hook guess, carried from the source function.
	Kind FunctionKind
	// Params are the parameter values, in order.
	Params []Place
	// Body is the tree.
	Body ReactiveBlock
}

// ReactiveBlock is a flat sequence of statements. Nesting lives inside the statements.
type ReactiveBlock []ReactiveStatement

// ReactiveStatement is one entry in a block.
//
// The set is closed: the unexported marker method means only this package can add a variant. Four
// variants, exactly upstream's.
type ReactiveStatement interface {
	reactiveStatement()
}

// ReactiveInstructionStatement is one instruction.
type ReactiveInstructionStatement struct {
	Instruction *ReactiveInstruction
}

// ReactiveTerminalStatement is a control-flow construct and everything nested inside it.
//
// This is the variant that makes the structure a tree. The HIR's `If` names its arms by BlockId;
// this one HOLDS them.
type ReactiveTerminalStatement struct {
	Terminal ReactiveTerminal
	// Label is the block this terminal can be broken out of, and whether the label is implicit.
	//
	// Upstream emits one for every terminal and prunes later; see the package comment on
	// `ReactiveFunctionGapUnprunedLabels`. Nil when the terminal carries no break target at all.
	Label *ReactiveLabel
}

// ReactiveLabel names a break target.
type ReactiveLabel struct {
	Id static_single_assignment.BlockId
	// Implicit records that control would reach this point anyway, so a printed `break` is
	// unnecessary. Upstream's `implicit`.
	Implicit bool
}

// ReactiveScopeBlock is a reactive scope and the statements it memoizes.
//
// `Pruned` is upstream's separate `pruned-scope` variant folded into a flag. The two share every
// field and share an arm in every consumer that matters here, and a bool avoids declaring a second
// statement type that nothing in this tree constructs. Recorded as a divergence because a reader
// diffing against the source will see two types where this has one.
type ReactiveScopeBlock struct {
	Scope        ScopeId
	Range        mutation_aliasing.MutableRange
	Instructions ReactiveBlock
	// Pruned marks a scope whose memoization was discarded, by any of the four upstream passes that
	// build a `PrunedScope`: `flattenReactiveLoopsHIR` and `flattenScopesWithHooksOrUseHIR` through
	// the ids `BuildReactiveFunctionWithFlattenedScopes` is handed, `pruneUnusedScopes` and
	// `pruneAlwaysInvalidatingScopes` by setting it on the tree.
	Pruned bool

	// Merged are the scopes this one absorbed, by id.
	//
	// Upstream's `scope.merged`, written in exactly one place in the whole compiler --
	// `mergeReactiveScopesThatInvalidateTogether` -- and read by
	// `validatePreservedManualMemoization` to fold absorbed ids into the scope set it tracks. It is
	// a field here rather than a side table because it is produced and consumed as part of the
	// block, and a merge that rewrites the block would have to keep a parallel table in step.
	//
	// Nil until that pass runs, which is the ordinary state: a scope that absorbed nothing carries
	// no set rather than an empty one.
	Merged []ScopeId
}

func (*ReactiveInstructionStatement) reactiveStatement() {}
func (*ReactiveTerminalStatement) reactiveStatement()    {}
func (*ReactiveScopeBlock) reactiveStatement()           {}

// ReactiveInstruction is one instruction in the tree.
//
// `LValue` is a POINTER rather than a value, because upstream's is nullable and the distinction is
// load-bearing: an instruction whose result is discarded has no lvalue, and a zero `Place` would be
// indistinguishable from a real place naming identifier zero.
type ReactiveInstruction struct {
	Order  static_single_assignment.EvaluationOrder
	LValue *Place
	Value  ReactiveValue
}

// ReactiveValue is what an instruction computes.
//
// The set is closed. Either a plain HIR `InstructionValue`, or one of the four COMPOSITE values that
// exist only in this form -- the ones that are terminals in the graph and expressions in the tree.
type ReactiveValue interface {
	reactiveValue()
}

// ReactiveInstructionValue wraps a plain HIR instruction value.
type ReactiveInstructionValue struct {
	Value InstructionValue
}

// ReactiveLogicalValue is `a && b` or `a || b`, which the HIR holds as a `Logical` terminal.
type ReactiveLogicalValue struct {
	Operator string
	Left     ReactiveValue
	Right    ReactiveValue
}

// ReactiveTernaryValue is `a ? b : c`, which the HIR holds as a `Ternary` terminal.
type ReactiveTernaryValue struct {
	Test       ReactiveValue
	Consequent ReactiveValue
	Alternate  ReactiveValue
}

// ReactiveSequenceValue is a comma expression: instructions evaluated for effect, then a value.
//
// This is also what `wrapWithSequence` produces when a value block carries instructions before its
// result, which is the ordinary case rather than an exotic one.
type ReactiveSequenceValue struct {
	Instructions []*ReactiveInstruction
	Order        static_single_assignment.EvaluationOrder
	Value        ReactiveValue
}

// ReactiveOptionalValue is `a?.b`, which the HIR holds as an `Optional` terminal. The nested Value
// is a sequence containing the chain's test and consequent computations, matching upstream's
// `visitValueBlockTerminal` reconstruction.
type ReactiveOptionalValue struct {
	Order    static_single_assignment.EvaluationOrder
	Value    ReactiveValue
	Optional bool
}

func (*ReactiveInstructionValue) reactiveValue() {}
func (*ReactiveLogicalValue) reactiveValue()     {}
func (*ReactiveTernaryValue) reactiveValue()     {}
func (*ReactiveSequenceValue) reactiveValue()    {}
func (*ReactiveOptionalValue) reactiveValue()    {}

// ---------------------------------------------------------------------------
// Terminals
// ---------------------------------------------------------------------------

// ReactiveTerminal is a control-flow construct in the tree.
//
// Thirteen variants, exactly upstream's. Fewer than the HIR's 22, and the difference is the point of
// this pass: `goto`, `branch`, `logical`, `ternary`, `sequence`, `optional`, `maybe-throw`,
// `scope`, `pruned-scope` and `unreachable` all disappear here. Some become values, some become
// statements, and `goto` becomes a break, a continue, or nothing at all when control was going to
// fall through anyway.
//
// The set is closed: the unexported marker method means only this package can add a variant.
type ReactiveTerminal interface {
	reactiveTerminal()
}

// ReactiveTerminalTargetKind is how a break or continue names its target.
//
// `Implicit` means control reaches the target anyway and no statement need be printed. `Labeled`
// means the target is not the innermost enclosing construct, so a label is required. `Unlabeled` is
// the ordinary `break;` out of the innermost loop or switch.
type ReactiveTerminalTargetKind uint8

const (
	ReactiveTargetImplicit ReactiveTerminalTargetKind = iota
	ReactiveTargetLabeled
	ReactiveTargetUnlabeled
)

// ReactiveBreak leaves a labeled block, a loop, or a switch.
type ReactiveBreak struct {
	Target     static_single_assignment.BlockId
	TargetKind ReactiveTerminalTargetKind
	Order      static_single_assignment.EvaluationOrder
}

// ReactiveContinue jumps to the next iteration of a loop.
type ReactiveContinue struct {
	Target     static_single_assignment.BlockId
	TargetKind ReactiveTerminalTargetKind
	Order      static_single_assignment.EvaluationOrder
}

// ReactiveReturn returns a value.
type ReactiveReturn struct {
	Value Place
	Order static_single_assignment.EvaluationOrder
}

// ReactiveThrow throws a value.
type ReactiveThrow struct {
	Value Place
	Order static_single_assignment.EvaluationOrder
}

// ReactiveSwitchCase is one arm. A nil Test is the default case; a nil Block is a fallthrough arm.
//
// Both nils are meaningful and distinct, which is why they are pointers: upstream models them as
// `Place | null` and `ReactiveBlock | void`, and an empty block is a case with no statements while
// an absent block is a case that falls into the next one.
type ReactiveSwitchCase struct {
	Test  *Place
	Block *ReactiveBlock
}

// ReactiveSwitch is a switch statement.
type ReactiveSwitch struct {
	Test  Place
	Cases []ReactiveSwitchCase
	Order static_single_assignment.EvaluationOrder
}

// ReactiveDoWhile is a `do { } while ()` loop.
type ReactiveDoWhile struct {
	Loop  ReactiveBlock
	Test  ReactiveValue
	Order static_single_assignment.EvaluationOrder
}

// ReactiveWhile is a `while () { }` loop.
type ReactiveWhile struct {
	Test  ReactiveValue
	Loop  ReactiveBlock
	Order static_single_assignment.EvaluationOrder
}

// ReactiveFor is a three-clause `for` loop. Update is nil when the loop has no update clause.
type ReactiveFor struct {
	Init   ReactiveValue
	Test   ReactiveValue
	Update ReactiveValue
	Loop   ReactiveBlock
	Order  static_single_assignment.EvaluationOrder
}

// ReactiveForOf is a `for (... of ...)` loop.
type ReactiveForOf struct {
	Init  ReactiveValue
	Test  ReactiveValue
	Loop  ReactiveBlock
	Order static_single_assignment.EvaluationOrder
}

// ReactiveForIn is a `for (... in ...)` loop.
type ReactiveForIn struct {
	Init  ReactiveValue
	Loop  ReactiveBlock
	Order static_single_assignment.EvaluationOrder
}

// ReactiveIf is an if statement. Alternate is nil when there is no else arm.
type ReactiveIf struct {
	Test       Place
	Consequent ReactiveBlock
	Alternate  *ReactiveBlock
	Order      static_single_assignment.EvaluationOrder
}

// ReactiveLabel is a labeled block, the target of a labeled break.
type ReactiveLabelTerminal struct {
	Block ReactiveBlock
	Order static_single_assignment.EvaluationOrder
}

// ReactiveTry is a try/catch. HandlerBinding is nil for `catch { }` with no parameter.
type ReactiveTry struct {
	Block          ReactiveBlock
	HandlerBinding *Place
	Handler        ReactiveBlock
	Order          static_single_assignment.EvaluationOrder
}

func (*ReactiveBreak) reactiveTerminal()         {}
func (*ReactiveContinue) reactiveTerminal()      {}
func (*ReactiveReturn) reactiveTerminal()        {}
func (*ReactiveThrow) reactiveTerminal()         {}
func (*ReactiveSwitch) reactiveTerminal()        {}
func (*ReactiveDoWhile) reactiveTerminal()       {}
func (*ReactiveWhile) reactiveTerminal()         {}
func (*ReactiveFor) reactiveTerminal()           {}
func (*ReactiveForOf) reactiveTerminal()         {}
func (*ReactiveForIn) reactiveTerminal()         {}
func (*ReactiveIf) reactiveTerminal()            {}
func (*ReactiveLabelTerminal) reactiveTerminal() {}
func (*ReactiveTry) reactiveTerminal()           {}

// ReactiveFunctionGap names a rule this pass does not apply, for a caller that needs to know.
//
// Modelled on `RangeGap`, `ScopeGap`, `DependencyGap` and `ScopeTerminalsGap` deliberately and for
// the same reason: a gap answerable through the API is one a consumer can decline on, while a gap
// living only in a comment is one the next reader inherits by accident.
type ReactiveFunctionGap uint8

const (
	// ReactiveFunctionGapUnprunedLabels is upstream's own naivety, faithfully reproduced.
	//
	// `BuildReactiveFunction` emits a label for EVERY terminal that has a fallthrough, and upstream's
	// comment says so outright: "this pass naively emits labels for *all* terminals: see
	// PruneUnusedLabels which removes unnecessary labels". That pruning pass is not in this tree.
	//
	// Reproduced rather than fixed here because the fix belongs to the pass that owns it. A consumer
	// counting labels will see more than a printed program would carry; a consumer asking about
	// control flow is unaffected, since an unpruned label changes nothing about which statements
	// nest where.
	ReactiveFunctionGapUnprunedLabels ReactiveFunctionGap = iota

	// ReactiveFunctionGapPrunedScopes is the `pruned-scope` variant, which nothing here produces.
	//
	// `ReactiveScopeBlock.Pruned` exists and is always false. The four upstream passes that set it
	// are all downstream of this one and none is built. Declared as a field rather than as a second
	// statement type, and named here so that a consumer testing `Pruned` knows the answer is
	// currently a constant rather than a measurement.
	ReactiveFunctionGapPrunedScopes

	// ReactiveFunctionGapUnbuiltTerminals is the remaining HIR terminals this pass would consume but never sees.
	//
	// `Sequence` and `MaybeThrow` are declared in `terminal.go` and constructed zero times. Optional
	// used to belong to this set, but optional lowering now constructs it and this pass rebuilds its
	// reactive value. The enum name is retained because the category still exists.
	//
	// `MaybeThrow` costs nothing even in principle: upstream flattens it away because the tree does
	// not model maybe-throw semantics at all. `Sequence` still lacks a live lowering input.
	ReactiveFunctionGapUnbuiltTerminals

	// ReactiveFunctionGapValueExpressions is the remaining composite-value reconstruction gap.
	//
	// Logical, ternary and optional terminals now become their nested reactive values on the shapes
	// their lowering produces. The helper deliberately declines an unexpected value-block terminal
	// and falls back to the statement-preserving path, so labeled or scope-interleaved value blocks
	// remain outside the reconstructed subset. In particular, a Scope terminal inside a loop's init
	// or test cannot be represented by `ReactiveValue`; the corpus conservation test measures that
	// population separately from unexplained instruction loss. `Sequence` also remains until lowering
	// constructs it.
	ReactiveFunctionGapValueExpressions
)

// ReactiveFunctionGaps are the rules BuildReactiveFunction does not apply. See ReactiveFunctionGap.
//
// Returned as a value rather than documented alone so a test can assert on it, which makes closing a
// gap a visible event rather than a silent improvement.
func ReactiveFunctionGaps() []ReactiveFunctionGap {
	return []ReactiveFunctionGap{
		ReactiveFunctionGapUnprunedLabels,
		ReactiveFunctionGapPrunedScopes,
		ReactiveFunctionGapUnbuiltTerminals,
		ReactiveFunctionGapValueExpressions,
	}
}
