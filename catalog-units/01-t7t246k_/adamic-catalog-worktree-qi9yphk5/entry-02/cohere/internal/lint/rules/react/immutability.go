package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

// Immutability flags mutating a value React requires to stay frozen.
//
//	valid:   function C() { const o = {a: 0}; o.a = 1; return <div>{o.a}</div>; }
//	valid:   function C() { const o = {a: 0}; o.a = 1; useFoo(o); return <div />; }
//	valid:   function C(props) { let x; if (props.c) { x = {a: 1}; } else { x = {}; } x.q = 1; return <div />; }
//	invalid: function C(props) { props.a = 1; return <div />; }
//	invalid: function C() { const [s] = useState({}); s.a = 1; return <div />; }
//	invalid: function C() { const o = {a: 0}; useFoo(o); o.a = 1; return <div />; }
//	invalid: function C() { let l = 0; const cb = () => { l = 1; }; return <div onClick={cb} />; }
//
// React re-renders by comparing values, not by observing writes. A value it has already read -- a
// prop, a piece of state, something handed to a hook, something interpolated into JSX -- is
// therefore frozen from that moment: React has taken its answer and will not look again. Writing
// through it afterwards changes what the program sees while leaving what React sees untouched, so
// the interface keeps rendering the old value and only corrects itself when some unrelated update
// happens to schedule a render. The symptom is a control that visibly does nothing until you click
// something else, which reads as a state-management bug rather than as what it is.
//
// # Where this comes from
//
// React's `InferMutationAliasingEffects` and two validators, transcribed by oxc at
// `oxc_react_compiler/src/react_compiler_inference/infer_mutation_aliasing_effects.rs` and
// `react_compiler_validation/`. oxc's *linter* rule at `rules/react/immutability.rs` is a four-line
// dispatcher into `run_react_compiler_rule(ctx, ErrorCategory::Immutability)`, so reading only the
// rule file makes a large port look trivial. Six diagnostic sites carry this category: three inside
// the inference pass and three inside the two validators.
//
// # This rule DOES need the abstract interpretation, and that was measured rather than inherited
//
// Three rules shipped onto this representation before this one found their scoping measurement
// overturned, so the claim that this one genuinely needs a value lattice was treated as a
// hypothesis and tested against React's own executable before any of it was written.
//
// The hypothesis holds, and the input that settles it is an ORDERING pair. Both of these compile,
// both mutate the same locally-created object, and they differ only in which statement comes first:
//
//	const o = {a: 0}; useFoo(o); o.a = 1;    REPORTS
//	const o = {a: 0}; o.a = 1; useFoo(o);    CLEAN
//
// So freezing is not a property of where a value came from, and a rule keyed on origin -- "props
// and state are immutable, locals are not" -- gets exactly half of that pair wrong while looking
// entirely reasonable. Freezing is a fact that a value ACQUIRES at a program point and carries
// forward, which is a dataflow fact and nothing cheaper. The same probe run through an alias
// (`const alias = o; useFoo(alias); o.a = 1;`) also reports, so the fact propagates through
// aliasing as well as through order, which rules out a purely per-binding table.
//
// # The lattice, and why its join cannot be an ordering
//
// Six elements: Mutable, Frozen, Primitive, MaybeFrozen, Global, Context. `internal/utilities/hir/
// high_level_intermediate_representation.go` carries a warning written for this rule, and it is correct: the `Effect` constants
// declared beside it look like a lattice and are not one, and a pass whose join is "take the larger
// Effect" would be self-consistent, pass its own tests, and be wrong.
//
// The join here is `immutabilityMergeKinds`, transcribed from upstream's `merge_value_kinds`, and
// the thing that makes it not an ordering is that `Frozen` joined with `Mutable` is `MaybeFrozen`,
// a THIRD element that is neither operand. No ordering of six constants expresses that, so the
// meet cannot be recovered by sorting them differently.
//
// Verified against React's executable rather than transcribed on faith, one probe per join edge,
// each with the compilation gate held constant so the gate could not explain the answer:
//
//	if (c) { x = useHook(); } else { x = {}; }   x.q = 1   REPORTS   Frozen  join Mutable   = MaybeFrozen
//	if (c) { x = {a:1};     } else { x = {}; }   x.q = 1   CLEAN     Mutable join Mutable   = Mutable
//	if (c) { x = useHook(); } else { x = h2;  }  x.q = 1   REPORTS   Frozen  join Frozen    = Frozen
//	if (c) { x = useHook(); } else { x = 3;   }  x.q = 1   REPORTS   Frozen  join Primitive = Frozen
//	if (c) { x = aGlobal;   } else { x = {}; }   x.q = 1   CLEAN     Global  join Mutable   = Mutable
//
// The fifth row is the one worth keeping: joining a global with a local produces something this
// rule is silent on, even though the global ALONE reports. A lattice built by intuition would make
// Global absorbing and would report there.
//
// # Convergence equality is not identity equality
//
// `refs.go` records this hazard from the neighbouring validator and it applies here for a different
// reason, so it is restated rather than cross-referenced. Upstream's abstract value is a kind plus
// a REASON SET, and the reason set only ever grows. If the fixpoint's "did anything move" test
// compared whole abstract values with Go's `==`, a value whose kind had settled but whose reason
// set was still absorbing bits from a back edge would report movement on every round.
//
// `immutabilityStateChanged` therefore compares the KIND for change and treats the reason set as
// monotone bookkeeping: a round counts as having moved only when some value's kind moved or some
// binding gained a new value, never when a reason bit arrived at a value whose kind is already
// final. Reasons still merge -- they select the message text -- they simply do not vote on
// termination. Without that split the sweep runs its full bound on any function with a loop.
//
// The reason set is a bitset rather than a slice for the same reason: two sets that merged the same
// bits in a different order must compare equal, and a slice would not.
//
// # Why a bounded whole-function sweep rather than `control_flow_graph.Solve`
//
// `control_flow_graph.Solve[V,E]` is mechanically unreachable from this representation: `control_flow_graph.
// Block`'s `index`, `final` and `thrown` fields are unexported and set only by `control_flow_graph.Build`,
// so a caller holding a `high_level_intermediate_representation.Function` cannot construct one. `ssa.go`, `postdominator.go` and
// `refs.go` all already record this. That is a mechanical fact and it is not the reason this rule
// does not use it.
//
// The reason is that upstream is not a per-block entry/exit solve either. It is a worklist over
// blocks carrying one `InferenceState` each, and the state is keyed by VALUE identity rather than
// by binding, with an aliasing side map that lets several bindings share one value. Reproduced here
// as a bounded whole-function sweep in reverse postorder, which is the same shape `refs.go` uses
// and which settles an acyclic region in a single round because `Function.Blocks` is already in
// reverse postorder. The bound is ten rounds, matching upstream's own. Measured on Kirk's tree,
// every function settled in at most two rounds; see the test file.
//
// # Nothing here may depend on phi operand order
//
// `high_level_intermediate_representation.Phi.Operands` is sorted by predecessor block id, so ranging
// it is deterministic, but the answer here does not rely on that. A sibling rule shipped a message
// that named a different builtin between runs because an operand walk's order varied. The join used
// here is commutative and associative, so the resulting KIND is order-independent by construction;
// the reason set is a bitset union, which is likewise. The message text is selected from the merged
// reason set by a fixed priority order and never from "whichever operand arrived first", so no
// output of this rule can vary between runs.
//
// # The component gate is the whole rule for a third of the corpus
//
// Twelve of the fifteen imported fixtures that React reports nothing on are silent for one reason:
// the function creates no JSX and calls no hook, so React never compiles it and no validator runs.
// Adding a single hook call to those twelve makes all twelve report. oxc's compiler test harness
// compiles every fixture unconditionally and therefore snapshots findings React's linter does not
// produce, which is a harness difference rather than a rule difference and is recorded here because
// it makes the imported snapshot counts misleading for this rule.
//
// The gate lives in `unsupported_syntax.go` and is reused whole. Its parameter test is load-bearing
// and not obvious: a component may take at most two parameters and a second one must be named for a
// ref, so `function Component(props, cond) { props.a = 1; return <div />; }` is SILENT while the
// same body with one parameter reports. That was measured here independently before the shared
// predicate was found, and an earlier reading of the same evidence produced a confident wrong
// theory that conditional mutation of props was exempt -- every probe supporting it had quietly
// introduced a second parameter. Holding the gate constant and re-running collapsed the theory: with
// arity fixed, conditionality makes no difference to any of the six freeze sources.
//
// # One span where upstream has two
//
// Upstream's diagnostics carry a primary span plus a secondary label pointing at the value's
// origin. Our `Diagnostic` carries a single `Range`, so the primary span is kept where upstream puts
// it and the origin label is dropped rather than folded into the message.
var Immutability = rule.Rule{
	Name:             "react-hooks/immutability",
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// Taken at the file level for the reason static_components.go and refs.go both give:
			// lowering already descends into nested functions through the function arena, so
			// listening per function kind would lower every inner function twice, once as a child
			// and once standalone with no enclosing context.
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// A file that cannot hold a component or a hook is not lowered at all; see
				// high_level_intermediate_representation.MayHoldComponentOrHook for why that is exact.
				if !high_level_intermediate_representation.MayHoldComponentOrHook(ctx) {
					return
				}
				immutabilityForEachCompiledFunction(node, func(functionNode *ast.Node) {
					lowered := high_level_intermediate_representation.ForFunction(ctx, functionNode)
					if lowered == nil {
						return
					}
					immutabilityAnalyzeCompilationUnit(ctx, lowered, functionNode)
				})
			},
		}
	},
}

// immutabilityForEachCompiledFunction visits the outermost function-like node on each branch.
//
// Identical in shape to the walkers in refs.go and static_components.go, and deliberately not
// shared: two porters collided on unprefixed helper names in this package and both then removed
// their copies at once, turning a redeclaration error into an undefined-symbol error pointing at
// nobody. Every helper here carries the rule's name. This is the THIRD copy of this traversal, which
// is the evidence threshold task `#g29j9ab` names for lifting it onto the shelf; see the report for
// what the shared shape actually is now that three exist.
func immutabilityForEachCompiledFunction(root *ast.Node, visit func(*ast.Node)) {
	if root == nil {
		return
	}
	root.ForEachChild(func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			visit(node)
			return false
		}
		immutabilityForEachCompiledFunction(node, visit)
		return false
	})
}

// immutabilityAnalyzeCompilationUnit runs the sweep over the functions React would have compiled.
//
// A function that is not a compilation unit is walked past into the functions it contains, because
// one of those may be a unit even when its wrapper is not.
func immutabilityAnalyzeCompilationUnit(ctx rule.Context, function *high_level_intermediate_representation.Function, node *ast.Node) {
	if function == nil {
		return
	}
	if !immutabilityIsCompilationUnit(function, node) {
		for index, nested := range function.Functions {
			_ = index
			immutabilityAnalyzeCompilationUnit(ctx, nested, nested.Node)
		}
		return
	}
	// A unit found nested in a function that is not one is analyzed as React Compiler compiles it,
	// on its own; see AsCompilationUnit.
	function = high_level_intermediate_representation.AsCompilationUnit(ctx, function,
		high_level_intermediate_representation.ForFunction)
	for _, finding := range immutabilitySweepFunction(ctx, function) {
		immutabilityReport(ctx, function, finding)
	}
}

// immutabilityIsCompilationUnit reports whether React would compile this function.
//
// Delegates wholly to the shared gate in unsupported_syntax.go, which is React's
// `getComponentOrHookLike`: the name must look like a component or a hook, the body must actually
// create JSX or call a hook, and a COMPONENT additionally has to have component-shaped parameters.
// The parameter half is what makes a two-parameter component silent, measured above.
//
// `high_level_intermediate_representation.Function.Kind` is deliberately NOT the gate here even though refs.go uses it. That field is a
// name-only guess recorded at lowering time and it does not ask the parameter question, so a
// two-parameter component passes it. Measured: keying on the field alone reports four inputs React
// is silent on.
func immutabilityIsCompilationUnit(function *high_level_intermediate_representation.Function, node *ast.Node) bool {
	if function == nil || node == nil {
		return false
	}
	return utilsreact.IsComponentOrHookLike(node)
}

// immutabilityInstructionAt returns one instruction by id, or nil when the id is out of range.
func immutabilityInstructionAt(function *high_level_intermediate_representation.Function, id high_level_intermediate_representation.InstructionId) *high_level_intermediate_representation.Instruction {
	if function == nil || int(id) >= len(function.Instructions) {
		return nil
	}
	return function.Instructions[id]
}

// immutabilityIdentifierNode returns the syntax a value came from, or nil for a pure temporary.
func immutabilityIdentifierNode(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) *ast.Node {
	if function == nil || int(id) >= len(function.Identifiers) {
		return nil
	}
	identifier := function.Identifiers[id]
	if identifier == nil {
		return nil
	}
	return identifier.Node
}

// immutabilityIdentifierName returns a value's source binding name, empty for a temporary.
func immutabilityIdentifierName(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) string {
	if function == nil || int(id) >= len(function.Identifiers) {
		return ""
	}
	identifier := function.Identifiers[id]
	if identifier == nil {
		return ""
	}
	return identifier.Name
}

// immutabilityIsHookName reports whether a name is a hook by React's own syntactic test.
func immutabilityIsHookName(name string) bool {
	if len(name) < 4 || name[:3] != "use" {
		return false
	}
	initial := name[3]
	return initial >= 'A' && initial <= 'Z'
}

// ---------------------------------------------------------------------------
// The lattice
// ---------------------------------------------------------------------------

// immutabilityValueKind is one of the six abstract values a value can hold.
//
// These are upstream's `ValueKind`, and the declaration order carries no meaning: the join below is
// not a max over it. `internal/utilities/hir/high_level_intermediate_representation.go` documents why that distinction matters, in a
// comment addressed to this rule.
type immutabilityValueKind uint8

const (
	// immutabilityMutable is a value this code created and may still write to.
	immutabilityMutable immutabilityValueKind = iota
	// immutabilityFrozen is a value React has already read. Writing to it is the finding.
	immutabilityFrozen
	// immutabilityPrimitive is a number, string or boolean. Writing through one is meaningless
	// rather than wrong, so it is silent.
	immutabilityPrimitive
	// immutabilityMaybeFrozen is the join of a frozen value with a mutable one. It reports, because
	// the write is wrong on at least one path and React cannot tell which path ran.
	immutabilityMaybeFrozen
	// immutabilityGlobal is a value declared outside the component. Writing to one reports with a
	// different reason, and joining one with a mutable value produces a MUTABLE result rather than
	// an absorbing global. Measured; see the rule comment.
	immutabilityGlobal
	// immutabilityContext is a local captured and written by a nested function. Mutable for the
	// purpose of this rule, and kept distinct because the join treats it differently from Mutable
	// when it meets Frozen.
	immutabilityContext
)

// immutabilityMergeKinds is upstream's `merge_value_kinds`, and it is not a max over an order.
//
// `Frozen` joined with `Mutable` is `MaybeFrozen`, a third element that is neither operand, so no
// reordering of the constants above can express this as a comparison. The branch order below is
// upstream's and is load-bearing: MaybeFrozen absorbs everything, then Mutable, then Context, and
// only after those does Frozen win over the two elements that carry no mutability of their own.
//
// Every edge was checked against React's own executable with the compilation gate held constant.
// The rows are in the rule comment.
func immutabilityMergeKinds(a, b immutabilityValueKind) immutabilityValueKind {
	if a == b {
		return a
	}
	if a == immutabilityMaybeFrozen || b == immutabilityMaybeFrozen {
		return immutabilityMaybeFrozen
	}
	if a == immutabilityMutable || b == immutabilityMutable {
		if a == immutabilityFrozen || b == immutabilityFrozen {
			return immutabilityMaybeFrozen
		}
		if a == immutabilityContext || b == immutabilityContext {
			return immutabilityContext
		}
		return immutabilityMutable
	}
	if a == immutabilityContext || b == immutabilityContext {
		if a == immutabilityFrozen || b == immutabilityFrozen {
			return immutabilityMaybeFrozen
		}
		return immutabilityContext
	}
	if a == immutabilityFrozen || b == immutabilityFrozen {
		return immutabilityFrozen
	}
	if a == immutabilityGlobal || b == immutabilityGlobal {
		return immutabilityGlobal
	}
	return immutabilityPrimitive
}

// immutabilityReason is why a value is frozen, which selects the message rather than the verdict.
//
// The kind decides WHETHER to report; the reason decides WHAT IT SAYS. Conflating the two is the
// most natural mistake here, because the reason names are the ones a reader recognizes.
type immutabilityReason uint16

const (
	immutabilityReasonGlobal immutabilityReason = 1 << iota
	immutabilityReasonJsxCaptured
	immutabilityReasonContext
	immutabilityReasonKnownReturnSignature
	immutabilityReasonReactiveFunctionArgument
	immutabilityReasonState
	immutabilityReasonReducerState
	immutabilityReasonEffect
	immutabilityReasonHookCaptured
	immutabilityReasonHookReturn
	immutabilityReasonOther
)

// immutabilityAbstractValue is one value's kind plus the set of reasons that produced it.
//
// The reason set is a BITSET rather than a slice, so two sets that absorbed the same reasons in a
// different order compare equal. That keeps the merged value independent of the order phi operands
// arrive in, which a slice-backed set would make it depend on.
type immutabilityAbstractValue struct {
	Kind    immutabilityValueKind
	Reasons immutabilityReason
}

// immutabilityMergeValues joins two abstract values.
func immutabilityMergeValues(a, b immutabilityAbstractValue) immutabilityAbstractValue {
	return immutabilityAbstractValue{
		Kind:    immutabilityMergeKinds(a.Kind, b.Kind),
		Reasons: a.Reasons | b.Reasons,
	}
}

// immutabilityMutationResult is what writing to a value of a given kind does.
//
// Transcribed from upstream's `mutate_with_span`, which is the ACTUAL gate on all three inference
// diagnostics. The scoping measurement for this rule described that gate as
// `abstract_value.kind == ValueKind::Frozen`, and that is the branch which picks the message, not
// the branch which decides to report: `MaybeFrozen` reports through the same path and `Global`
// reports through a different one. Reading only the equality would have shipped a rule blind to
// every joined value and to every global write.
type immutabilityMutationResult uint8

const (
	// immutabilityMutationNone is a permitted write.
	immutabilityMutationNone immutabilityMutationResult = iota
	// immutabilityMutationAllowed is a write to a value this code owns.
	immutabilityMutationAllowed
	// immutabilityMutationFrozen is a write to a frozen or maybe-frozen value.
	immutabilityMutationFrozen
	// immutabilityMutationGlobal is a write to a value declared outside the component.
	immutabilityMutationGlobal
)

// immutabilityMutate answers what writing to a value of this kind does.
//
// The `Primitive` arm is not an optimization: writing a property onto a number is silently
// discarded at runtime rather than being an immutability violation, and React says nothing about
// it. Verified: `let x = 3; x.q = 1;` inside a compiled component is CLEAN.
func immutabilityMutate(kind immutabilityValueKind) immutabilityMutationResult {
	switch kind {
	case immutabilityMutable, immutabilityContext:
		return immutabilityMutationAllowed
	case immutabilityPrimitive:
		return immutabilityMutationNone
	case immutabilityFrozen, immutabilityMaybeFrozen:
		return immutabilityMutationFrozen
	case immutabilityGlobal:
		return immutabilityMutationGlobal
	}
	return immutabilityMutationNone
}

// ---------------------------------------------------------------------------
// The abstract state
// ---------------------------------------------------------------------------

// immutabilityFindingKind is which of the four diagnostics a finding is.
type immutabilityFindingKind uint8

const (
	// immutabilityFindingImmutableValue is writing to a value React has frozen. The dominant case.
	immutabilityFindingImmutableValue immutabilityFindingKind = iota
	// immutabilityFindingReassignedAfterRender is reassigning a local from a callback that outlives
	// render.
	immutabilityFindingReassignedAfterRender
	// immutabilityFindingReassignedInAsync is reassigning a local from an async callback.
	immutabilityFindingReassignedInAsync
	// immutabilityFindingKnownMutableFunction is freezing a function that reassigns a local, by
	// passing it somewhere React will hold onto.
	immutabilityFindingKnownMutableFunction
)

// immutabilityFinding is one diagnostic, held as data until the sweep settles.
//
// Deferred rather than reported inline for the reason refs.go gives: the sweep runs the same
// instructions up to ten times, so reporting at the moment a check fails would emit each finding
// once per round. Findings are recomputed per round and only the last round's set is kept.
type immutabilityFinding struct {
	Kind    immutabilityFindingKind
	Reasons immutabilityReason
	// Variable is the source name to interpolate, empty when the value is a temporary.
	Variable string
	// Node is where the finding points.
	Node *ast.Node
	// order keeps findings in instruction order so output is deterministic.
	order int
}

// immutabilityState is the abstract interpretation's state.
//
// Keyed by VALUE rather than by binding, with `aliases` mapping a binding onto the value it names.
// That indirection is upstream's `variables`/`values` split and it is what makes freezing propagate
// through an alias: `const alias = o; useFoo(alias); o.a = 1;` reports because `alias` and `o` name
// one value, and freezing reaches the value rather than the name.
type immutabilityState struct {
	// values holds each value's abstract kind and reasons, keyed by the identifier that introduced
	// it. A binding reaches its value through `aliases`.
	values map[static_single_assignment.IdentifierId]immutabilityAbstractValue
	// aliases maps a binding onto the value it ultimately names.
	aliases map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId
	// declarationValues recovers a value whose single-assignment numbering did not unify, keyed by
	// the source binding. refs.go documents the same representation gap and works around it the
	// same way.
	declarations  map[static_single_assignment.IdentifierId]static_single_assignment.DeclarationId
	byDeclaration map[static_single_assignment.DeclarationId]static_single_assignment.IdentifierId
	// names carries a source binding name onto the temporary that loaded it.
	names map[static_single_assignment.IdentifierId]string
	// functionsReassigning records, per value, which locals calling it would reassign. A function
	// value carrying a non-empty set is a "known mutable function".
	functionsReassigning map[static_single_assignment.IdentifierId][]immutabilityReassignment
	// functionCaptures records what a closure closes over, so freezing the closure freezes them.
	functionCaptures map[static_single_assignment.IdentifierId][]static_single_assignment.IdentifierId
	// aggregateElements records what an object or array literal holds, so freezing the aggregate
	// freezes its contents. Kept separate from functionCaptures because an object holding a value
	// and a closure capturing one are different relations that happen to propagate the same way.
	aggregateElements map[static_single_assignment.IdentifierId][]static_single_assignment.IdentifierId
	// changed is the fixpoint's only termination signal.
	changed bool
	// insideLoop is true while the sweep is walking a block that a back edge can re-enter.
	insideLoop bool
	// frozenInsideLoop records values frozen at such a point, which are the only ones a back edge
	// may carry into the next round. See the sweep.
	frozenInsideLoop map[static_single_assignment.IdentifierId]bool
}

// immutabilityReassignment is one local a nested function reassigns.
type immutabilityReassignment struct {
	Variable string
	Node     *ast.Node
	IsAsync  bool
	// Target is the DECLARATION written. Compared by declaration rather than by identifier because
	// the capture edge and the store name the same binding under two different single-assignment
	// numberings; see the caller.
	Target static_single_assignment.DeclarationId
	// Direct is true when the write happened in the immediate closure rather than in one nested
	// inside it. Only a direct write can be matched against the capture edge, because a deeper
	// closure's identifier space is a third one again.
	Direct bool
}

// immutabilityFilterOwned keeps only the reassignments that write a binding this function owns.
//
// A write from the immediate closure is matched against the capture edge. A write from a DEEPER
// closure is dropped rather than matched, because its identifier ids live in yet another space and
// translating them would need the capture edge chained through every level. That is a real
// narrowing and it is stated rather than hidden: a two-level closure reassigning a render local is
// silent here and reports upstream. See the report.
func immutabilityFilterOwned(reassignments []immutabilityReassignment, owned map[static_single_assignment.DeclarationId]bool) []immutabilityReassignment {
	kept := reassignments[:0]
	for _, reassignment := range reassignments {
		if reassignment.Direct && owned[reassignment.Target] {
			kept = append(kept, reassignment)
		}
	}
	return kept
}

func newImmutabilityState() *immutabilityState {
	return &immutabilityState{
		values:               map[static_single_assignment.IdentifierId]immutabilityAbstractValue{},
		aliases:              map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{},
		declarations:         map[static_single_assignment.IdentifierId]static_single_assignment.DeclarationId{},
		byDeclaration:        map[static_single_assignment.DeclarationId]static_single_assignment.IdentifierId{},
		names:                map[static_single_assignment.IdentifierId]string{},
		functionsReassigning: map[static_single_assignment.IdentifierId][]immutabilityReassignment{},
		functionCaptures:     map[static_single_assignment.IdentifierId][]static_single_assignment.IdentifierId{},
		aggregateElements:    map[static_single_assignment.IdentifierId][]static_single_assignment.IdentifierId{},
		frozenInsideLoop:     map[static_single_assignment.IdentifierId]bool{},
	}
}

// resolve follows a binding to the value it names.
//
// Bounded rather than a bare loop: an alias chain is acyclic by construction in single-assignment
// form, but this pass also seeds aliases from a declaration fallback, and a defensive bound costs
// nothing next to a hang in a linter.
func (state *immutabilityState) resolve(id static_single_assignment.IdentifierId) static_single_assignment.IdentifierId {
	current := id
	for round := 0; round < 32; round++ {
		next, found := state.aliases[current]
		if !found || next == current {
			return current
		}
		current = next
	}
	return current
}

// alias records that one binding names the same value as another.
func (state *immutabilityState) alias(from static_single_assignment.IdentifierId, to static_single_assignment.IdentifierId) {
	resolved := state.resolve(to)
	if resolved == from {
		return
	}
	if existing, found := state.aliases[from]; found && existing == resolved {
		return
	}
	state.aliases[from] = resolved
	state.changed = true
}

// kindOf is a value's current abstract value, defaulting to Mutable.
//
// Mutable is the default rather than a bottom element because an unseen value is one this code has
// not been shown to have frozen, and the rule reports only on freezing. Defaulting to anything
// else would report values nothing had frozen.
func (state *immutabilityState) kindOf(id static_single_assignment.IdentifierId) immutabilityAbstractValue {
	resolved := state.resolve(id)
	if found, ok := state.values[resolved]; ok {
		return found
	}
	// Fall back to whatever the same source BINDING holds elsewhere in this function, which
	// recovers a value whose single-assignment numbering did not unify with the one that was
	// written. refs.go measures the same gap on its own corpus.
	if declaration, ok := state.declarations[resolved]; ok {
		if other, ok := state.byDeclaration[declaration]; ok && other != resolved {
			if found, ok := state.values[other]; ok {
				return found
			}
		}
	}
	return immutabilityAbstractValue{Kind: immutabilityMutable}
}

// setKind writes a value's abstract value, recording whether the KIND moved.
//
// This is the convergence test and it deliberately ignores the reason set. See the rule comment:
// reasons only ever grow, so counting a reason arrival as movement makes any function with a loop
// burn the whole bound.
func (state *immutabilityState) setKind(id static_single_assignment.IdentifierId, value immutabilityAbstractValue) {
	resolved := state.resolve(id)
	previous, existed := state.values[resolved]
	merged := value
	if existed {
		merged = immutabilityMergeValues(previous, value)
	}
	if !existed || previous.Kind != merged.Kind {
		state.changed = true
	}
	state.values[resolved] = merged
	if declaration, ok := state.declarations[resolved]; ok {
		state.byDeclaration[declaration] = resolved
	}
}

// freeze marks a value frozen for a reason, which is the whole engine of this rule.
//
// A value that is already Frozen, Global or Primitive is left alone, matching upstream's `freeze`:
// re-freezing an already-frozen value must not count as movement or the fixpoint never settles.
func (state *immutabilityState) freeze(id static_single_assignment.IdentifierId, reason immutabilityReason) {
	current := state.kindOf(id)
	switch current.Kind {
	case immutabilityFrozen, immutabilityGlobal, immutabilityPrimitive:
		// Already settled. Absorb the reason without moving the kind.
		if current.Reasons&reason == 0 {
			resolved := state.resolve(id)
			current.Reasons |= reason
			state.values[resolved] = current
		}
		return
	}
	state.setKind(id, immutabilityAbstractValue{Kind: immutabilityFrozen, Reasons: current.Reasons | reason})
	// Only a freeze that happened INSIDE a loop body may travel around the back edge into the next
	// round. See the sweep for why carrying every freeze forward reports ordinary accumulator loops.
	if state.insideLoop {
		state.frozenInsideLoop[state.resolve(id)] = true
	}
	// Freezing a closure freezes what it captured, which is upstream's transitive freeze. Guarded
	// against re-entry through the captures map itself so a closure capturing another closure that
	// captures it cannot recurse forever.
	state.freezeInward(id, reason)
}

// freezeInward distributes a freeze to what a value holds: a closure's captures, an aggregate's
// elements. Both maps are removed for the duration of the descent rather than guarded by a visited
// set, so a structure that contains itself cannot recurse forever, and are restored afterwards so a
// second freeze for a different reason still reaches inside.
func (state *immutabilityState) freezeInward(id static_single_assignment.IdentifierId, reason immutabilityReason) {
	resolved := state.resolve(id)
	captures, hasCaptures := state.functionCaptures[resolved]
	elements, hasElements := state.aggregateElements[resolved]
	if !hasCaptures && !hasElements {
		return
	}
	delete(state.functionCaptures, resolved)
	delete(state.aggregateElements, resolved)
	for _, inner := range captures {
		state.freeze(inner, reason)
	}
	for _, inner := range elements {
		state.freeze(inner, reason)
	}
	if hasCaptures {
		state.functionCaptures[resolved] = captures
	}
	if hasElements {
		state.aggregateElements[resolved] = elements
	}
}

// noteDeclaration records which source binding a value belongs to.
func (state *immutabilityState) noteDeclaration(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) {
	if function == nil || int(id) >= len(function.Identifiers) {
		return
	}
	if identifier := function.Identifiers[id]; identifier != nil && identifier.Name != "" {
		state.declarations[id] = identifier.Declaration
		if _, seen := state.byDeclaration[identifier.Declaration]; !seen {
			state.byDeclaration[identifier.Declaration] = id
		}
	}
}

// nameOf is the value's own binding name, or the one a load carried onto it.
func (state *immutabilityState) nameOf(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) string {
	// `function` is nil at one call site, where only the side maps are wanted. `immutability
	// IdentifierName` already declines a nil function rather than panicking, so this is safe and is
	// stated here because a nil receiver in a name lookup reads like an oversight.
	if name := immutabilityIdentifierName(function, id); name != "" {
		return name
	}
	if name, ok := state.names[id]; ok {
		return name
	}
	return state.names[state.resolve(id)]
}

// ---------------------------------------------------------------------------
// The sweep
// ---------------------------------------------------------------------------

// immutabilitySweepFunction runs the bounded ten-round sweep over one function.
//
// Seed the parameters, collect the aliasing side map once, then iterate the whole function in block
// order until the state stops moving or ten rounds pass. `Blocks` is in reverse postorder, which is
// what lets a single forward pass settle an acyclic region in one round; a loop needs a second
// round, which is what the bound is for.
//
// Unlike refs.go this does NOT return early once a finding exists. That rule reproduces upstream's
// early return, which is sound there because its diagnostics are per-function. Here upstream keeps
// analysing and reports every distinct mutation, and the imported corpus contains functions with
// two and three findings, so returning early would drop them.
func immutabilitySweepFunction(ctx rule.Context, function *high_level_intermediate_representation.Function) []immutabilityFinding {
	state := newImmutabilityState()
	immutabilityCollectAliases(function, state)

	// `entry` is the state a round STARTS from, and it is what carries a fact across a back edge.
	// The working state is reset to it at the top of every round, which is the single most important
	// line in this sweep and the one a straightforward reading gets wrong.
	//
	// Without the reset, a freeze discovered late in round one is still present when round two
	// re-executes the instructions ABOVE it, and this rule's entire content is that a write before a
	// freeze is legal while the same write after it is not. Measured: with the state carried
	// forward, `const o = {}; o.a = 1; useThing(o);` reports on round two, which is the exact input
	// React is silent on and the exact input that distinguishes this rule from an origin-based one.
	// Both ordering fixtures failed this way on the first run.
	//
	// What legitimately crosses a round is a fact that reached the entry through a loop back edge,
	// which is what `entry` holds and what the merge below accumulates.
	// Parameters are seeded into the ENTRY state rather than into the working state, because the
	// working state is cleared at the top of every round and a seed written into it would be wiped
	// before the first instruction ran. That is the whole seeding bug: every props fixture went
	// silent at once and every ordering fixture kept passing, which reads as a props-specific defect
	// and is really a lifetime one.
	entry := map[static_single_assignment.IdentifierId]immutabilityAbstractValue{}
	immutabilitySeedParameters(function, entry)
	// A function with no back edge settles in ONE round: `Function.Blocks` is in reverse postorder,
	// so a single forward pass already visits every definition before every use. Iterating such a
	// function again cannot discover anything and can only do harm, because round two re-executes
	// the writes that round one correctly permitted while holding the freezes round one discovered
	// after them. Bounding the sweep to the shape that needs it is therefore a correctness measure
	// rather than a cost one.
	rounds := 1
	if immutabilityHasBackEdge(function) {
		rounds = 10
	}
	var findings []immutabilityFinding
	for iteration := 0; iteration < rounds; iteration++ {
		state.values = map[static_single_assignment.IdentifierId]immutabilityAbstractValue{}
		for id, value := range entry {
			state.values[id] = value
		}
		state.changed = false
		findings = immutabilityRunOneRound(ctx, function, state)
		// Accumulate into the entry state ONLY what was frozen at a point that can reach the back
		// edge, which is what `state.frozenInsideLoop` records.
		//
		// Carrying every conclusion forward is the obvious thing and it is wrong in a way the
		// fixtures cannot see. A value frozen AFTER the loop -- the accumulator built in a `for`
		// and then handed to a hook on the next line -- would arrive at round two already frozen,
		// and round two re-executes the loop body and reports the writes that built it. That was 3
		// of the last 10 findings on Kirk's tree, all of them ordinary
		// `for (...) { totals[key] = value; }` followed by a hook call, and React is silent on all
		// of them: measured, `for (...) { acc[u.id] = 1; } useThing(acc);` is CLEAN while
		// `for (...) { useThing(acc); acc[u.id] = 1; }` REPORTS. The freeze has to be inside the
		// loop for the back edge to carry it.
		moved := false
		for id := range state.frozenInsideLoop {
			value, known := state.values[id]
			if !known {
				continue
			}
			previous, existed := entry[id]
			merged := value
			if existed {
				merged = immutabilityMergeValues(previous, value)
			}
			if !existed || previous.Kind != merged.Kind {
				moved = true
			}
			entry[id] = merged
		}
		if !moved {
			break
		}
	}
	return immutabilityDeduplicate(findings)
}

// immutabilitySeedParameters gives every parameter its starting kind.
//
// A component's or a hook's parameters are its props and hook arguments, which React has already
// read by the time the body runs. They are Frozen from entry, with the reason that selects the
// "component props or hook arguments" message.
//
// Destructuring is handled by the ordinary instruction arms rather than here: a destructured
// parameter lowers to a temporary in `Params` plus destructuring instructions in the entry block,
// so freezing the temporary reaches every bound name through `Destructure`.
func immutabilitySeedParameters(function *high_level_intermediate_representation.Function, entry map[static_single_assignment.IdentifierId]immutabilityAbstractValue) {
	for _, param := range function.Params {
		entry[param.Identifier] = immutabilityAbstractValue{
			Kind:    immutabilityFrozen,
			Reasons: immutabilityReasonReactiveFunctionArgument,
		}
	}
}

// immutabilityCollectAliases builds the binding-to-value side map before the sweep runs.
//
// Gathered once rather than per round because aliasing is a syntactic fact that does not depend on
// the abstract state, and because a use can precede its definition across a back edge.
func immutabilityCollectAliases(function *high_level_intermediate_representation.Function, state *immutabilityState) {
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := immutabilityInstructionAt(function, instructionId)
			if instruction == nil {
				continue
			}
			state.noteDeclaration(function, instruction.LValue.Identifier)
			high_level_intermediate_representation.EachInstructionPlace(instruction, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
				state.noteDeclaration(function, place.Identifier)
			})
			switch value := instruction.Value.(type) {
			case *high_level_intermediate_representation.LoadLocal:
				state.alias(instruction.LValue.Identifier, value.Place.Identifier)
				if name := immutabilityIdentifierName(function, value.Place.Identifier); name != "" {
					state.names[instruction.LValue.Identifier] = name
				}
			case *high_level_intermediate_representation.LoadContext:
				state.alias(instruction.LValue.Identifier, value.Place.Identifier)
				if name := immutabilityIdentifierName(function, value.Place.Identifier); name != "" {
					state.names[instruction.LValue.Identifier] = name
				}
			case *high_level_intermediate_representation.StoreLocal:
				state.alias(instruction.LValue.Identifier, value.Value.Identifier)
				state.alias(value.LValue.Identifier, value.Value.Identifier)
			case *high_level_intermediate_representation.StoreContext:
				state.alias(value.LValue.Identifier, value.Value.Identifier)
				state.alias(instruction.LValue.Identifier, value.Value.Identifier)
			case *high_level_intermediate_representation.TypeCastExpression:
				// A cast is a no-op at runtime, so the cast value and its operand are one value.
				state.alias(instruction.LValue.Identifier, value.Value.Identifier)
			case *high_level_intermediate_representation.LoadGlobal:
				state.names[instruction.LValue.Identifier] = value.Name
			}
		}
	}
	// The alias map is a syntactic fact settled before the first round. Movement recorded while
	// building it is not movement of the ABSTRACT state, so clearing the flag here keeps the
	// fixpoint's termination signal about the lattice rather than about the side map.
	state.changed = false
}

// immutabilityRunOneRound is one pass over every block, phi and instruction.
//
// A hook's RETURN deliberately does not freeze what it hands back, and that is a removal rather than
// an omission. Freezing at the return terminal was the obvious transcription of "the caller owns it
// now", and it was the last false positive on Kirk's tree: `useTranslations` builds a local `result`
// in a loop and returns it, and because every return terminal was visited before the instruction
// walk, the freeze landed on the accumulator before the loop that fills it had run.
//
// Measured before removing it rather than after: both `useThing(d) { const r = {}; for (...) { r[k]
// = d[k]; } return r; }` and the same shape behind an early return are CLEAN under React, against a
// control that reports. A hook building a value and returning it is what hooks are for; the freeze
// that matters is the one the CALLER applies when it receives the value, and that is already
// modelled at the call site by `immutabilityHookResultKind`.
func immutabilityRunOneRound(ctx rule.Context, function *high_level_intermediate_representation.Function, state *immutabilityState) []immutabilityFinding {
	var findings []immutabilityFinding
	order := 0

	loopBlocks := immutabilityLoopBlocks(function)
	// Phis are resolved over every block first, then instructions are walked in INSTRUCTION-ID
	// order rather than in block order, and that ordering is the correctness point rather than a
	// tidiness one.
	//
	// `Function.Blocks` is in reverse postorder, which is the right order for a lattice to converge
	// in and the WRONG order for a rule whose whole content is "did the freeze happen before the
	// write". Measured by dumping the representation for
	// `const acc = {}; for (const u of list) { acc[u.id] = 1; } useThing(acc);`: reverse postorder
	// visits b5, holding the `useThing` call at instruction 14, BEFORE b4, holding the loop body's
	// write at instruction 11. So the freeze was applied before the write was checked, on the very
	// first round, and the accumulator loop reported. Instruction ids are assigned in lowering
	// order, which is source order, so walking them recovers the order a reader sees.
	//
	// The first theory for this was a multi-round leak and it was wrong: instrumenting the sweep
	// showed the finding present in ROUND ZERO, which no amount of round bookkeeping could explain.
	// That measurement is what redirected this from the fixpoint to the walk order.
	// Phis are resolved when the walk REACHES their block rather than up front, because a phi's
	// operands are defined by instructions and resolving every phi first would read them before they
	// held anything. Keyed by the block's first instruction so the resolution lands in source order
	// with everything else.
	phiAt := map[high_level_intermediate_representation.InstructionId][]*high_level_intermediate_representation.Phi{}
	for _, block := range function.Blocks {
		if len(block.Phis) == 0 || len(block.Instructions) == 0 {
			continue
		}
		phiAt[block.Instructions[0]] = block.Phis
	}
	// Which block each instruction belongs to, so the loop flag still follows the instruction.
	blockOfInstruction := make(map[high_level_intermediate_representation.InstructionId]static_single_assignment.BlockId, len(function.Instructions))
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			blockOfInstruction[instructionId] = block.Id
		}
	}
	for instructionId := 0; instructionId < len(function.Instructions); instructionId++ {
		instruction := immutabilityInstructionAt(function, high_level_intermediate_representation.InstructionId(instructionId))
		if instruction == nil {
			continue
		}
		blockId, reachable := blockOfInstruction[high_level_intermediate_representation.InstructionId(instructionId)]
		if !reachable {
			// An instruction in no block is unreachable code that lowering left in the table.
			continue
		}
		state.insideLoop = loopBlocks[blockId]
		for _, phi := range phiAt[high_level_intermediate_representation.InstructionId(instructionId)] {
			// `Phi.Operands` ranges in ascending predecessor order, but the merge does not rely on
			// it: the join is commutative and associative and the reason set is a bitset union, so
			// the merged result is order-independent by construction. Nothing here may read "the
			// first operand".
			merged := immutabilityAbstractValue{Kind: immutabilityPrimitive}
			first := true
			for _, entry := range phi.Operands {
				operandValue := state.kindOf(entry.Place.Identifier)
				if first {
					merged = operandValue
					first = false
					continue
				}
				merged = immutabilityMergeValues(merged, operandValue)
			}
			if !first {
				state.setKind(phi.Place.Identifier, merged)
			}
		}
		findings = immutabilityTransfer(ctx, function, state, instruction, findings, &order)
	}

	return findings
}

// immutabilityDeduplicate collapses findings that name the same site.
//
// The sweep recomputes findings every round and keeps the last round's set, so duplication within
// one round is what this handles: an aliased value can reach the same mutation through two
// bindings. Keyed on the reported node and the finding kind, and it preserves first-seen order so
// the output does not depend on map iteration.
func immutabilityDeduplicate(findings []immutabilityFinding) []immutabilityFinding {
	type key struct {
		node *ast.Node
		kind immutabilityFindingKind
	}
	seen := map[key]bool{}
	result := make([]immutabilityFinding, 0, len(findings))
	for _, finding := range findings {
		identity := key{node: finding.Node, kind: finding.Kind}
		if seen[identity] {
			continue
		}
		seen[identity] = true
		result = append(result, finding)
	}
	return result
}

// ---------------------------------------------------------------------------
// The transfer function
// ---------------------------------------------------------------------------

// immutabilityTransfer applies one instruction to the state and collects any findings.
//
// The arms fall into three groups and it is worth naming them, because reading the switch as one
// flat list hides the structure:
//
//   - Arms that FREEZE. A hook call freezes its arguments, JSX freezes what it interpolates, a hook
//     return freezes what it hands back. These are the sources of every finding this rule produces.
//   - Arms that WRITE. A property store, a computed store, a delete, an update operator. Each asks
//     the lattice what writing to that value does, and reports when the answer is not "allowed".
//   - Arms that PROPAGATE. Object and array construction, destructuring, calls. These move kinds
//     around without deciding anything.
//
// The default arm deliberately does nothing. Unlike refs.go, whose default arm reports
// conservatively, an unmodelled instruction here must NOT report: this rule's finding is a write,
// and inventing writes from instructions we do not model would report ordinary code. Failing toward
// silence is the safe direction for a mutation rule and toward reporting is the safe direction for
// an access rule, which is why the two neighbours differ.
func immutabilityTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	state *immutabilityState,
	instruction *high_level_intermediate_representation.Instruction,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	target := instruction.LValue.Identifier

	switch value := instruction.Value.(type) {

	// --- Freezing ---------------------------------------------------------

	// JSX freezes every value it interpolates: React has read it to render, so a later write
	// changes what the program sees and not what React saw. Measured: the same mutation BEFORE the
	// JSX is clean and after it reports.
	case *high_level_intermediate_representation.JsxExpression:
		for _, attribute := range value.Props {
			// A callback handed to JSX is held by React past the end of this render, exactly as one
			// handed to a hook is, so the known-mutable-function check belongs here too. Measured:
			// `let l = 0; const cb = () => { l = 1; }; return <div onClick={cb} />;` reports TWICE
			// under React, once at the reassignment and once at the attribute. Checking only hook
			// arguments left both findings missing and the rule silent on the single most common
			// spelling of this mistake.
			findings = immutabilityCheckFrozenFunction(state, attribute.Value.Identifier, instruction.Node, findings, order)
			state.freeze(attribute.Value.Identifier, immutabilityReasonJsxCaptured)
		}
		for _, child := range value.Children {
			findings = immutabilityCheckFrozenFunction(state, child.Identifier, instruction.Node, findings, order)
			state.freeze(child.Identifier, immutabilityReasonJsxCaptured)
		}
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityFrozen, Reasons: immutabilityReasonJsxCaptured})

	case *high_level_intermediate_representation.JsxFragment:
		for _, child := range value.Children {
			state.freeze(child.Identifier, immutabilityReasonJsxCaptured)
		}
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityFrozen, Reasons: immutabilityReasonJsxCaptured})

	// --- Writing ----------------------------------------------------------

	case *high_level_intermediate_representation.PropertyStore:
		findings = immutabilityCheckWrite(ctx, function, state, value.Object.Identifier, instruction.Node, findings, order)

	case *high_level_intermediate_representation.ComputedStore:
		findings = immutabilityCheckWrite(ctx, function, state, value.Object.Identifier, instruction.Node, findings, order)

	case *high_level_intermediate_representation.PropertyDelete:
		findings = immutabilityCheckWrite(ctx, function, state, value.Object.Identifier, instruction.Node, findings, order)

	case *high_level_intermediate_representation.ComputedDelete:
		findings = immutabilityCheckWrite(ctx, function, state, value.Object.Identifier, instruction.Node, findings, order)

	// `props.count++` and `props.count += 1` are deliberately NOT handled here, and the reason is a
	// property of the lowering rather than of React. Dumping the representation for `props.count++`
	// shows five instructions: a load of `props`, a property load of `count`, the update itself,
	// another load of `props`, and a PROPERTY STORE writing the result back. The update's operand is
	// the loaded NUMBER, not the object, so an arm here would ask the lattice about a primitive and
	// answer nothing, while the store that follows is the real write and is already handled above.
	//
	// Both spellings report through that store, measured, with the span on the member access. An arm
	// here reported zero findings on `props.count++` and reading it as a missing feature would have
	// produced a second finding on every compound assignment in the tree.

	// A write to a global binding is upstream's `global_reassignment`, which carries the Globals
	// category rather than this one, so it is deliberately not reported here. The GLOBAL VALUE's
	// kind is still recorded, because a property store THROUGH a global does belong to this rule.
	case *high_level_intermediate_representation.StoreGlobal:
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityGlobal, Reasons: immutabilityReasonGlobal})

	// --- Propagating ------------------------------------------------------

	// A global read is a value declared outside the component. Writing through one reports with the
	// "defined outside a component or hook" reason.
	case *high_level_intermediate_representation.LoadGlobal:
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityGlobal, Reasons: immutabilityReasonGlobal})

	// A property load off a frozen value is itself frozen, which is what makes `props.a.q = 1`
	// report: the intermediate `props.a` inherits the freeze from `props`.
	case *high_level_intermediate_representation.PropertyLoad:
		state.setKind(target, immutabilityPropagate(state, value.Object.Identifier))
		if name := immutabilityIdentifierName(function, value.Object.Identifier); name != "" {
			state.names[target] = name
		} else if carried, ok := state.names[value.Object.Identifier]; ok {
			state.names[target] = carried
		}

	case *high_level_intermediate_representation.ComputedLoad:
		state.setKind(target, immutabilityPropagate(state, value.Object.Identifier))

	// Destructuring distributes the source's kind to every bound name, which is how a destructured
	// parameter's members become frozen without a per-member rule.
	case *high_level_intermediate_representation.Destructure:
		carried := immutabilityPropagate(state, value.Value.Identifier)
		high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
			if role == high_level_intermediate_representation.PlaceRoleDefine {
				state.setKind(place.Identifier, carried)
			}
		})

	// A literal is a fresh mutable value: this code just made it and nothing has read it yet.
	// A literal is a fresh mutable value: this code just made it and nothing has read it yet. What it
	// HOLDS is recorded so that freezing the aggregate later reaches its contents, which is what
	// makes `useMemo(() => 1, [o]); o.a = 1;` report: the freeze lands on the dependency array and
	// has to travel one hop inward to reach `o`. Without that hop the array froze and the value it
	// named did not, and the fixture read as a missing hook rather than as a missing edge.
	case *high_level_intermediate_representation.ObjectExpression, *high_level_intermediate_representation.ArrayExpression:
		elements := []static_single_assignment.IdentifierId{}
		high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
			elements = append(elements, place.Identifier)
		})
		state.aggregateElements[state.resolve(target)] = elements
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityMutable})

	case *high_level_intermediate_representation.Primitive, *high_level_intermediate_representation.BinaryExpression, *high_level_intermediate_representation.TemplateLiteral:
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityPrimitive})

	// A nested function is analysed as an inner subject: which locals it reassigns is recorded on
	// the VALUE, so the finding lands where the function is handed to React rather than at the
	// reassignment. That is what makes the two reassignment diagnostics point where they do.
	case *high_level_intermediate_representation.FunctionExpression:
		findings = immutabilityNestedFunctionTransfer(ctx, function, state, value, target, findings, order)

	case *high_level_intermediate_representation.CallExpression:
		findings = immutabilityCallTransfer(ctx, function, state, instruction, value.Callee.Identifier, value.Args, target, findings, order)

	case *high_level_intermediate_representation.MethodCall:
		findings = immutabilityMethodCallTransfer(ctx, function, state, instruction, value, target, findings, order)

	default:
		// Deliberately silent. See the doc comment: an unmodelled instruction must not invent a
		// write.
	}

	return findings
}

// immutabilityPropagate is what a value loaded out of another value inherits.
//
// A member of a frozen value is frozen; a member of a mutable value is mutable. The reasons travel
// with it so the message still names the original source.
//
// A member of a PRIMITIVE is treated as mutable rather than primitive: reading a property off a
// number yields undefined, not another primitive, and carrying `Primitive` forward would make a
// later write silent for the wrong reason.
func immutabilityPropagate(state *immutabilityState, object static_single_assignment.IdentifierId) immutabilityAbstractValue {
	current := state.kindOf(object)
	if current.Kind == immutabilityPrimitive {
		return immutabilityAbstractValue{Kind: immutabilityMutable}
	}
	return current
}

// immutabilityCheckWrite asks the lattice what writing to a value does and reports when it must.
func immutabilityCheckWrite(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	state *immutabilityState,
	object static_single_assignment.IdentifierId,
	node *ast.Node,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	// A REF is exempt before the lattice is consulted at all, which is upstream's `MutateRef` arm:
	// `mutate_with_span` asks `is_ref_or_ref_value(ty)` FIRST and returns before it looks at any
	// kind. Writing `something.current` is the sanctioned way to hold a value across renders, and a
	// ref is deliberately outside React's data flow, so the whole premise of this rule does not
	// apply to it.
	//
	// This was 180 of 281 findings on Kirk's tree, every single one an ordinary
	// `handlersReference.current = handlers` inside an effect, and it is the reason step 10 exists.
	// The 40 imported fixtures were all green over it, because none of them writes through a ref: a
	// rule can be right about every case its corpus contains and wrong about the most common shape
	// in the repository it actually runs on. Measured on the executable with a live control in the
	// same batch, so the silence is a real exemption rather than a probe that could see nothing:
	// `r.current = 1` is clean in render, in an effect and in a callback, while a non-ref value from
	// the same hook shape reports.
	//
	// `refs.go` OWNS the render-time half of this question and reports a ref access during render
	// under its own rule, so the exemption here loses no coverage; it routes the finding to the rule
	// whose name matches it.
	if immutabilityIsRefValue(ctx, function, state, object) || immutabilityWriteTargetIsRef(ctx, node) {
		return findings
	}
	current := state.kindOf(object)
	switch immutabilityMutate(current.Kind) {
	case immutabilityMutationFrozen, immutabilityMutationGlobal:
		*order++
		return append(findings, immutabilityFinding{
			Kind:     immutabilityFindingImmutableValue,
			Reasons:  current.Reasons,
			Variable: state.nameOf(function, object),
			Node:     immutabilityMutatedObjectNode(node, object, function),
			order:    *order,
		})
	}
	return findings
}

// immutabilityMutatedObjectNode is the syntax React underlines for a write.
//
// React points at the OBJECT BEING MUTATED rather than at the assignment or at the member access,
// and the difference is one node in each direction, so a span assertion is the only fixture that can
// see it. Measured on the executable, reading the rendered underline for each shape:
//
//	props.a = 1        underlines `props`      not `props.a` and not `props.a = 1`
//	props.a.q = 1      underlines `props.a`    the object of the write, one level in
//	delete props.a     underlines `props`
//	props.a = 1 inside a callback   underlines `props`
//
// So the node wanted is the LEFT-HAND side of the member expression being written, which is the
// instruction node's own object. Falling back to the instruction node keeps a finding rather than
// dropping one when the shape is not a member expression.
func immutabilityMutatedObjectNode(node *ast.Node, object static_single_assignment.IdentifierId, function *high_level_intermediate_representation.Function) *ast.Node {
	if node == nil {
		return immutabilityIdentifierNode(function, object)
	}
	target := node
	// A write lowers with the assignment as its node in some shapes and the member access in
	// others, so the assignment is unwrapped first and the member access second.
	if target.Kind == ast.KindBinaryExpression {
		if left := target.AsBinaryExpression().Left; left != nil {
			target = left
		}
	}
	if target.Kind == ast.KindDeleteExpression {
		if inner := target.Expression(); inner != nil {
			target = inner
		}
	}
	switch target.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if expression := target.Expression(); expression != nil {
			return expression
		}
	}
	if resolved := immutabilityIdentifierNode(function, object); resolved != nil {
		return resolved
	}
	return node
}

// immutabilityCallTransfer handles a call, which is where most freezing happens.
//
// A HOOK call freezes every argument, because React holds onto what it is given across renders. It
// also decides the result's kind: `useState` and `useReducer` return frozen state, `useContext`
// returns a frozen context value, and any other hook returns a frozen hook return. Those three
// reasons select three different messages over one verdict.
//
// A plain call freezes nothing. That is not a simplification: measured on React's executable,
// `const o = {}; doThing(o); o.a = 1;` is CLEAN while the same code with `useThing` reports. React
// only freezes at boundaries it controls.
func immutabilityCallTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	state *immutabilityState,
	instruction *high_level_intermediate_representation.Instruction,
	callee static_single_assignment.IdentifierId,
	args []high_level_intermediate_representation.Argument,
	target static_single_assignment.IdentifierId,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	name := state.nameOf(function, callee)
	if !immutabilityIsHookName(name) {
		// A non-hook call leaves everything alone, but a function value that reassigns a local is
		// still being CALLED here rather than handed to React, and calling it during render is
		// ordinary. Nothing to do.
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityMutable})
		return findings
	}

	for _, argument := range args {
		// Handing a function that reassigns a local to a hook is the "known mutable function"
		// finding: React will hold the function past render and calling it later writes to a
		// variable whose value the next render has already forgotten.
		findings = immutabilityCheckFrozenFunction(state, argument.Place.Identifier, instruction.Node, findings, order)
		if immutabilityDependencyArrayIsExempt(name) && immutabilityIsDependencyArray(function, state, argument.Place.Identifier) {
			continue
		}
		state.freeze(argument.Place.Identifier, immutabilityReasonHookCaptured)
	}

	state.setKind(target, immutabilityHookResultKind(name))
	return findings
}

// immutabilityDependencyArrayIsExempt reports whether this hook leaves its dependencies unread.
//
// The exemption is NOT a property of the array-literal shape alone, which is where the first version
// of this went wrong. Measured across eight hooks with the identical `hook(() => 1, [o]); o.a = 1;`
// body, so the shape is held constant and only the callee varies:
//
//	useEffect             CLEAN      useMemo               REPORTS
//	useLayoutEffect       CLEAN      useCallback           REPORTS
//	useInsertionEffect    CLEAN
//	useImperativeHandle   CLEAN
//	useDeferredValue      CLEAN
//	useThing (unknown)    CLEAN
//
// The split is exactly the two MEMOIZATION hooks, and the reason is that their callbacks run during
// render, so React reads the dependencies on the spot; every other hook defers past the end of
// render and only stores them. So the test is on the hook rather than on the argument, and an
// unknown hook falls on the exempt side because that is where every measured unknown falls.
//
// This is the arm that no amount of reading produced. The obvious rule -- "an array literal in
// argument position is a dependency list and dependency lists are not captures" -- is right for six
// of these eight and silently wrong for the two most common memo spellings in any real codebase.
func immutabilityDependencyArrayIsExempt(name string) bool {
	return name != "useMemo" && name != "useCallback"
}

// immutabilityIsDependencyArray reports whether an argument is an inline array literal.
//
// A dependency array is NOT a capture, and this is the one arm of the hook-freezing rule that
// reading the source would not have produced. Measured on React's executable with a live control in
// the same batch:
//
//	useThing(o);                  o.a = 1    REPORTS   the value is handed to the hook
//	useThing(() => 1, [o]);       o.a = 1    CLEAN     the value is only named as a dependency
//	useEffect(() => {}, [o]);     o.a = 1    CLEAN     same shape, the common spelling of it
//	useEffect(() => { read(o); }, []); o.a=1 REPORTS   the value is captured by the callback
//
// The distinction is not about which hook it is. `useMemo(() => 1, [o])` REPORTS, and it has the
// identical shape, because a memo callback runs during render and React reads what it closes over.
// So the array itself is frozen either way; what differs is whether anything reads the ELEMENTS,
// and only an inline array literal in argument position is exempt from distributing its freeze
// inward. Anything reached some other way, including an array built into a variable first, is not
// covered by this and freezes normally, which is upstream's behaviour and is why the test is on the
// argument's own instruction rather than on the value's kind.
//
// Note this is a lookup on the DEFINING instruction rather than on syntax: a value whose defining
// instruction is an array literal is the exemption, and a value that merely happens to hold an
// array is not.
func immutabilityIsDependencyArray(function *high_level_intermediate_representation.Function, state *immutabilityState, id static_single_assignment.IdentifierId) bool {
	resolved := state.resolve(id)
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := immutabilityInstructionAt(function, instructionId)
			if instruction == nil {
				continue
			}
			if state.resolve(instruction.LValue.Identifier) != resolved {
				continue
			}
			if _, isArray := instruction.Value.(*high_level_intermediate_representation.ArrayExpression); isArray {
				return true
			}
		}
	}
	return false
}

// immutabilityHookResultKind is what a hook hands back.
//
// The three reasons here are the three different messages a reader sees for one verdict, and they
// were each confirmed against React's executable rather than read off the reason enum:
//
//	const [s] = useState({});   s.a = 1     "returned from 'useState()'"
//	const [s] = useReducer(..); s.a = 1     "returned from 'useReducer()'"
//	const c = useContext(C);    c.a = 1     "returned from 'useContext()'"
//	const h = useThing();       h.a = 1     "returned from a hook"
func immutabilityHookResultKind(name string) immutabilityAbstractValue {
	reason := immutabilityReasonHookReturn
	switch name {
	case "useState":
		reason = immutabilityReasonState
	case "useReducer":
		reason = immutabilityReasonReducerState
	case "useContext":
		reason = immutabilityReasonContext
	case "useEffect", "useLayoutEffect", "useInsertionEffect":
		reason = immutabilityReasonEffect
	}
	return immutabilityAbstractValue{Kind: immutabilityFrozen, Reasons: reason}
}

// immutabilityMethodCallTransfer handles `object.method(...)`.
//
// The hook test reads the PROPERTY name, so `React.useEffect(...)` behaves like the bare import.
// refs.go measured this as 235 false positives on Kirk's tree when it was missing, and the same
// shape is live here: a value frozen inside a `React.useCallback` callback must be frozen by the
// same rule as one frozen inside `useCallback`.
//
// A method call on a value does NOT freeze the receiver and does not report a mutation through it.
// Measured: `props.list.push(1)` and `Object.assign(props, {a:1})` are both CLEAN under React, which
// is upstream declining to model an unknown method's effects rather than deciding the write is
// fine. Reproduced as silence deliberately; a rule that helpfully reported `props.list.push(1)`
// would disagree with React on ordinary code.
func immutabilityMethodCallTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	state *immutabilityState,
	instruction *high_level_intermediate_representation.Instruction,
	value *high_level_intermediate_representation.MethodCall,
	target static_single_assignment.IdentifierId,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	name := immutabilityMethodName(instruction)
	if !immutabilityIsHookName(name) {
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityMutable})
		return findings
	}
	for _, argument := range value.Args {
		findings = immutabilityCheckFrozenFunction(state, argument.Place.Identifier, instruction.Node, findings, order)
		if immutabilityDependencyArrayIsExempt(name) && immutabilityIsDependencyArray(function, state, argument.Place.Identifier) {
			continue
		}
		state.freeze(argument.Place.Identifier, immutabilityReasonHookCaptured)
	}
	state.setKind(target, immutabilityHookResultKind(name))
	return findings
}

// immutabilityMethodName is the property being called, read from the syntax.
//
// The representation gives a method call a Place for the property, but the property's NAME lives on
// the syntax rather than on the value, so it is read back from the node. Identical in shape to
// refs.go's helper and kept separate for the naming reason given at the top of this file.
func immutabilityMethodName(instruction *high_level_intermediate_representation.Instruction) string {
	if instruction == nil || instruction.Node == nil {
		return ""
	}
	node := instruction.Node
	if node.Kind == ast.KindCallExpression {
		if expression := node.Expression(); expression != nil {
			node = expression
		}
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// immutabilityNestedFunctionTransfer records which locals a nested function reassigns.
//
// The finding does not land inside the nested function. It lands where that function is handed to
// React, which is upstream's design and is why the two reassignment messages talk about "after
// render completes" rather than about the write itself: the write is fine, the problem is that it
// happens at a time when nothing will re-render.
//
// The inner function is examined through its OWN identifier space. Identifier ids are per-function
// and they overlap, which refs.go measured as a wrong-finding rather than a missing one, so the
// inner walk here reads only names and syntax and never indexes the outer state by an inner id.
func immutabilityNestedFunctionTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	state *immutabilityState,
	value *high_level_intermediate_representation.FunctionExpression,
	target static_single_assignment.IdentifierId,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	inner := immutabilityNestedFunction(function, value.Function)
	if inner == nil {
		state.setKind(target, immutabilityAbstractValue{Kind: immutabilityMutable})
		return findings
	}
	// A closure that WRITES THROUGH a captured value is a different finding from one that reassigns
	// a captured binding, and it lands at the write rather than where the closure is handed over.
	// Measured: `const cb = () => { props.a = 1; };` reports on line 2, inside the callback, while
	// `const cb = () => { local = 1; };` reports on the line that hands `cb` to JSX. The two look
	// alike in source and are opposite in where they point.
	findings = immutabilityCapturedWrites(ctx, state, inner, value, findings, order)
	// A reassignment only matters when the binding belongs to THIS function, which is the render
	// whose end outlives it. `Function.Context` names what the closure captured from here, so a
	// write to anything else is a write to a binding some intermediate scope owns and the render has
	// no stake in it.
	//
	// Three false-positive classes on Kirk's tree came from omitting this, 30 findings between them,
	// and all three are shapes React is silent on (measured, against a control that reports twice):
	//
	//	useEffect(() => { let cancelled = false; ...; return () => { cancelled = true; }; }, [])
	//	useEffect(() => { let id; function h() { id = setTimeout(...); } ... }, [])
	//	const cb = () => { let v; if (c) { v = 1; } else { v = 2; } return v; }
	//
	// Every one declares its variable inside a callback and writes it from a deeper callback, so the
	// binding never belonged to the render at all. The cleanup-function idiom is the most common of
	// the three and it is the single most idiomatic thing in React effects.
	// Compared by DECLARATION rather than by identifier id, which is the trap here. The context
	// value and the store's lvalue are two single-assignment numberings of the SAME binding and
	// carry different ids: measured on `let local = 0; const cb = () => { local = 1; };`, the
	// capture arrives as context id 2 while the store writes id 6. An identifier comparison answers
	// false for every real case and silences the rule completely, which is what it did on the first
	// run of this check. `high_level_intermediate_representation.Identifier.Declaration` is the identity that survives the renumbering.
	owned := map[static_single_assignment.DeclarationId]bool{}
	for index := range value.Captures {
		if index >= len(inner.Context) {
			continue
		}
		contextId := inner.Context[index].Identifier
		if int(contextId) < len(inner.Identifiers) {
			if identifier := inner.Identifiers[contextId]; identifier != nil {
				owned[identifier.Declaration] = true
			}
		}
	}
	reassignments := immutabilityFilterOwned(immutabilityReassignmentsIn(inner, inner.IsAsync), owned)
	if len(reassignments) > 0 {
		state.functionsReassigning[state.resolve(target)] = reassignments
	}
	// The values this closure captures are recorded on the function value, so that freezing the
	// FUNCTION freezes what it closes over. That is upstream's
	// `enableTransitivelyFreezeFunctionExpressions`, which defaults to TRUE, and it is what makes
	// `const o = {}; useEffect(() => { read(o); }, []); o.a = 1;` report while the same value named
	// only in the dependency array stays clean. Both were measured; the pair is the fixture
	// `f_useEffect_callback_captures` against `s_useEffect_dependency_array`.
	captured := make([]static_single_assignment.IdentifierId, 0, len(value.Captures))
	for _, capture := range value.Captures {
		captured = append(captured, capture.Identifier)
	}
	state.functionCaptures[state.resolve(target)] = captured
	state.setKind(target, immutabilityAbstractValue{Kind: immutabilityMutable})
	return findings
}

// immutabilityWriteTargetIsRef asks the checker whether the thing being written to is a ref.
//
// Asked through the WRITE'S OWN SYNTAX rather than through the value's identifier, and that is the
// whole trick. A property store's object is a temporary produced by a load, so
// `high_level_intermediate_representation.Identifier.Node` is nil for it and every type question asked that way answers nothing. The
// member expression on the left of the assignment still names the object in source, and handing
// THAT to the checker resolves.
//
// Measured, and the measurement is the reason this function exists in this shape. Asking through
// the identifier returned a nil symbol for all 281 findings, which reads exactly like "the checker
// cannot see React's types here" and would have justified giving up on the exemption. Asking
// through the write's syntax resolved `RefObject` for 215 of 223 on the same tree in the same run.
// One phrasing of the same question is blind and the other is not.
//
// A hand-written `react.d.ts` stub cannot reproduce this either way, which is why it was measured on
// the real tree rather than in a fixture: `rule_testing`'s tsconfig sets `types: []` and resolves no
// `node_modules`, so `@types/react` never resolves there and every symbol is nil. The fixtures
// therefore cannot see this exemption at all, and the two ref fixtures below pin the NAME half only.
func immutabilityWriteTargetIsRef(ctx rule.Context, node *ast.Node) bool {
	if ctx.TypeChecker == nil || node == nil {
		return false
	}
	target := node
	if target.Kind == ast.KindBinaryExpression {
		target = target.AsBinaryExpression().Left
	}
	if target != nil && target.Kind == ast.KindDeleteExpression {
		target = target.Expression()
	}
	if target == nil {
		return false
	}
	if target.Kind != ast.KindPropertyAccessExpression && target.Kind != ast.KindElementAccessExpression {
		return false
	}
	// Walk OUT through the whole member chain rather than testing only the immediate object.
	//
	// `reference.current.style.height = ''` writes to the DOM node a ref holds, and the immediate
	// object is a `CSSStyleDeclaration` rather than a `RefObject`, so a test on one level answers
	// false and reports. It was 7 of the 10 remaining findings on Kirk's tree and every one was
	// ordinary imperative DOM code inside an effect, which React is silent on: measured, both
	// `r.current.style.height = ''` and `r.current.options = {}` are clean, against a control that
	// reports.
	//
	// The walk is bounded rather than recursive because a member chain is short and a bound cannot
	// hang a linter on a pathological input.
	expression := target.Expression()
	for depth := 0; expression != nil && depth < 16; depth++ {
		valueType := ctx.TypeChecker.GetTypeAtLocation(expression)
		if valueType != nil {
			// Compared positively rather than checked for non-nil: the shim is a hand-mirrored
			// struct read through `unsafe.Pointer` and has returned a silently wrong type before,
			// so asserting that what came back is what was asked for is the only check that can see
			// that failure.
			if symbol := shimchecker.Type_symbol(valueType); symbol != nil {
				if symbol.Name == "RefObject" || symbol.Name == "MutableRefObject" {
					return true
				}
			}
		}
		if expression.Kind != ast.KindPropertyAccessExpression && expression.Kind != ast.KindElementAccessExpression {
			return false
		}
		expression = expression.Expression()
	}
	return false
}

// immutabilityCapturedWrites reports a nested function writing THROUGH a captured frozen value.
//
// The write happens inside the closure and is reported there, which is upstream's own placement and
// is measurably different from where a captured-binding REASSIGNMENT reports. The two are easy to
// conflate because both are "a closure touching an outer value".
//
// The capture edge is what makes this answerable: `FunctionExpression.Captures` at the call site and
// `Function.Context` inside are positionally paired, so the Nth capture is the Nth context value.
// That translation is required rather than convenient, because identifier ids are per-function and
// OVERLAP, so indexing the outer state with an inner id silently reads an unrelated value. refs.go
// measured that failure as a wrong finding rather than a missing one.
//
// Recurses one level only. A closure inside a closure writing through a doubly-captured value is a
// real shape and it is NOT covered here; see the report for what that costs.
func immutabilityCapturedWrites(
	ctx rule.Context,
	state *immutabilityState,
	inner *high_level_intermediate_representation.Function,
	value *high_level_intermediate_representation.FunctionExpression,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	// Translate the outer kinds into the inner function's identifier space through the capture edge.
	innerKinds := map[static_single_assignment.IdentifierId]immutabilityAbstractValue{}
	innerNames := map[static_single_assignment.IdentifierId]string{}
	for index, capture := range value.Captures {
		if index >= len(inner.Context) {
			continue
		}
		contextId := inner.Context[index].Identifier
		innerKinds[contextId] = state.kindOf(capture.Identifier)
		innerNames[contextId] = state.nameOf(nil, capture.Identifier)
	}
	// An alias built inside the closure carries the captured value's kind along with it.
	aliases := map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{}
	resolve := func(id static_single_assignment.IdentifierId) static_single_assignment.IdentifierId {
		for round := 0; round < 32; round++ {
			next, found := aliases[id]
			if !found || next == id {
				return id
			}
			id = next
		}
		return id
	}
	for _, block := range inner.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := immutabilityInstructionAt(inner, instructionId)
			if instruction == nil {
				continue
			}
			switch instructionValue := instruction.Value.(type) {
			case *high_level_intermediate_representation.LoadContext:
				aliases[instruction.LValue.Identifier] = instructionValue.Place.Identifier
			case *high_level_intermediate_representation.LoadLocal:
				aliases[instruction.LValue.Identifier] = instructionValue.Place.Identifier
			case *high_level_intermediate_representation.PropertyLoad:
				// A member of a captured frozen value is frozen too, which is what reaches
				// `props.a.q = 1` written inside a callback.
				if carried, known := innerKinds[resolve(instructionValue.Object.Identifier)]; known {
					innerKinds[instruction.LValue.Identifier] = carried
				}
			case *high_level_intermediate_representation.PropertyStore:
				findings = immutabilityCapturedWriteFinding(ctx, inner, innerKinds, innerNames, resolve(instructionValue.Object.Identifier), instruction, findings, order)
			case *high_level_intermediate_representation.ComputedStore:
				findings = immutabilityCapturedWriteFinding(ctx, inner, innerKinds, innerNames, resolve(instructionValue.Object.Identifier), instruction, findings, order)
			case *high_level_intermediate_representation.PropertyDelete:
				findings = immutabilityCapturedWriteFinding(ctx, inner, innerKinds, innerNames, resolve(instructionValue.Object.Identifier), instruction, findings, order)
			case *high_level_intermediate_representation.ComputedDelete:
				findings = immutabilityCapturedWriteFinding(ctx, inner, innerKinds, innerNames, resolve(instructionValue.Object.Identifier), instruction, findings, order)
			}
		}
	}
	return findings
}

// immutabilityCapturedWriteFinding reports one write through a captured value when it is frozen.
func immutabilityCapturedWriteFinding(
	ctx rule.Context,
	inner *high_level_intermediate_representation.Function,
	kinds map[static_single_assignment.IdentifierId]immutabilityAbstractValue,
	names map[static_single_assignment.IdentifierId]string,
	object static_single_assignment.IdentifierId,
	instruction *high_level_intermediate_representation.Instruction,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	current, known := kinds[object]
	if !known {
		return findings
	}
	// The same ref exemption as the outer write check. A ref captured by a callback and written
	// there is the single most common React idiom there is, and it accounted for most of this
	// rule's first dry run.
	if names[object] != "" && refsIsRefLikeName(names[object]) {
		return findings
	}
	if immutabilityWriteTargetIsRef(ctx, instruction.Node) {
		return findings
	}
	switch immutabilityMutate(current.Kind) {
	case immutabilityMutationFrozen, immutabilityMutationGlobal:
		*order++
		return append(findings, immutabilityFinding{
			Kind:     immutabilityFindingImmutableValue,
			Reasons:  current.Reasons,
			Variable: names[object],
			Node:     immutabilityMutatedObjectNode(instruction.Node, object, nil),
			order:    *order,
		})
	}
	return findings
}

// immutabilityNestedFunction resolves a function id within its parent.
func immutabilityNestedFunction(function *high_level_intermediate_representation.Function, id high_level_intermediate_representation.FunctionId) *high_level_intermediate_representation.Function {
	if function == nil || int(id) >= len(function.Functions) {
		return nil
	}
	return function.Functions[id]
}

// immutabilityReassignmentsIn collects the captured locals a function writes to.
//
// A write to a CONTEXT value is the tell: lowering emits `StoreContext` for a write to a binding
// the function closed over, and `StoreLocal` for one it owns. That distinction is exactly the
// question "does this write escape the function", and it comes free from the representation rather
// than needing a capture analysis.
//
// Recurses into further nested functions, because a callback inside a callback that reassigns an
// outer local is the same finding. `error.invalid-nested-function-reassign-local-variable-in-effect`
// is that shape.
func immutabilityReassignmentsIn(function *high_level_intermediate_representation.Function, isAsync bool) []immutabilityReassignment {
	return immutabilityReassignmentsInner(function, isAsync, true)
}

// immutabilityReassignmentsInner carries whether this level is the immediate closure.
func immutabilityReassignmentsInner(function *high_level_intermediate_representation.Function, isAsync bool, direct bool) []immutabilityReassignment {
	if function == nil {
		return nil
	}
	var found []immutabilityReassignment
	seen := map[string]bool{}
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := immutabilityInstructionAt(function, instructionId)
			if instruction == nil {
				continue
			}
			store, isStore := instruction.Value.(*high_level_intermediate_representation.StoreContext)
			if !isStore {
				continue
			}
			name := immutabilityIdentifierName(function, store.LValue.Identifier)
			if name == "" || seen[name] {
				continue
			}
			declarationOf := static_single_assignment.DeclarationId(0)
			if int(store.LValue.Identifier) < len(function.Identifiers) {
				if identifier := function.Identifiers[store.LValue.Identifier]; identifier != nil {
					declarationOf = identifier.Declaration
				}
			}
			seen[name] = true
			found = append(found, immutabilityReassignment{
				Variable: name,
				Node:     instruction.Node,
				IsAsync:  isAsync,
				Target:   declarationOf,
				Direct:   direct,
			})
		}
	}
	for _, nested := range function.Functions {
		for _, reassignment := range immutabilityReassignmentsInner(nested, isAsync || nested.IsAsync, false) {
			if seen[reassignment.Variable] {
				continue
			}
			seen[reassignment.Variable] = true
			found = append(found, reassignment)
		}
	}
	return found
}

// immutabilityCheckFrozenFunction reports handing React a function that reassigns a local.
//
// Two findings come out of one condition, split on whether the reassigning function is async, which
// is upstream's own split between `reassigned_after_render` and `reassigned_in_async_function`.
// Both are reported at the reassignment, plus one `known_mutable_function` at the site where the
// function is handed over, which is why React emits two diagnostics for a single such callback.
func immutabilityCheckFrozenFunction(
	state *immutabilityState,
	id static_single_assignment.IdentifierId,
	node *ast.Node,
	findings []immutabilityFinding,
	order *int,
) []immutabilityFinding {
	reassignments, found := state.functionsReassigning[state.resolve(id)]
	if !found || len(reassignments) == 0 {
		return findings
	}
	for _, reassignment := range reassignments {
		kind := immutabilityFindingReassignedAfterRender
		if reassignment.IsAsync {
			kind = immutabilityFindingReassignedInAsync
		}
		*order++
		findings = append(findings, immutabilityFinding{
			Kind:     kind,
			Variable: reassignment.Variable,
			Node:     reassignment.Node,
			order:    *order,
		})
	}
	*order++
	findings = append(findings, immutabilityFinding{
		Kind:     immutabilityFindingKnownMutableFunction,
		Variable: reassignments[0].Variable,
		Node:     node,
		order:    *order,
	})
	return findings
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

const (
	messageImmutabilityImmutableValueId        = "immutableValue"
	messageImmutabilityReassignedAfterRenderId = "reassignedAfterRender"
	messageImmutabilityReassignedInAsyncId     = "reassignedInAsyncFunction"
	messageImmutabilityKnownMutableFunctionId  = "knownMutableFunction"
)

// immutabilityWhyFrozenValuesMatter is the shared explanation, held once so a single edit moves
// every finding and the descriptions cannot drift apart.
const immutabilityWhyFrozenValuesMatter = "React re-renders by comparing values rather than by " +
	"observing writes, so a value it has already read is fixed from that moment. Writing through " +
	"it afterwards changes what this code sees while leaving what React sees untouched, and the " +
	"interface keeps showing the old value until some unrelated update happens to schedule a " +
	"render."

// immutabilityReasonText is the sentence that names WHY a value is frozen.
//
// The reason set is a bitset and several bits can be present at once, so the arms below are a fixed
// PRIORITY ORDER rather than a switch on a single value. That order is upstream's own
// `get_write_error_reason`, read top to bottom, and reproducing it exactly is what makes the
// message deterministic: without a fixed order, a value frozen by both JSX and a hook would name
// whichever bit happened to be tested first.
func immutabilityReasonText(reasons immutabilityReason) string {
	switch {
	case reasons&immutabilityReasonGlobal != 0:
		return "This variable is declared outside the component or hook, so React has no way to " +
			"know it changed. Update it from an effect instead."
	case reasons&immutabilityReasonJsxCaptured != 0:
		return "This value was already used in JSX, so React has read it and will not read it " +
			"again. Move the change above the JSX that uses it."
	case reasons&immutabilityReasonContext != 0:
		return "This value came from `useContext`, and it belongs to whichever provider supplied " +
			"it. Change it where the provider owns it."
	case reasons&immutabilityReasonKnownReturnSignature != 0:
		return "This value came from a function whose result is not meant to be changed."
	case reasons&immutabilityReasonReactiveFunctionArgument != 0:
		return "This value is a prop or a hook argument, so it belongs to the caller. Change it " +
			"where it is owned and pass an update callback down, or copy it into local state."
	case reasons&immutabilityReasonState != 0:
		return "This value came from `useState`, and React compares the whole value to decide " +
			"whether to re-render. Call the setter with a new value instead of writing into this one."
	case reasons&immutabilityReasonReducerState != 0:
		return "This value came from `useReducer`. Dispatch an action instead of writing into it."
	case reasons&immutabilityReasonEffect != 0:
		return "This value was already handed to an effect, so React has captured it. Make the " +
			"change before the effect that uses it."
	case reasons&immutabilityReasonHookCaptured != 0:
		return "This value was already passed to a hook, so React has captured it. Make the " +
			"change before the hook call that receives it."
	case reasons&immutabilityReasonHookReturn != 0:
		return "This value came from a hook, so the hook owns it. Make the change inside the hook " +
			"where the value is built."
	}
	return "React treats this value as fixed once it has been read."
}

// immutabilityMessageFor renders the finding.
//
// Every description is a fixed string rather than an interpolation, which is deliberate: a
// `rule.Message` is `{Id, Description}` with no rendering layer, so there is nothing to guard
// against a format-string defect and nothing that can render a doubled token. The variable name is
// carried on the finding for the span rather than for the text.
func immutabilityMessageFor(finding immutabilityFinding) rule.Message {
	switch finding.Kind {
	case immutabilityFindingReassignedAfterRender:
		return rule.Message{
			Id: messageImmutabilityReassignedAfterRenderId,
			Description: "This assignment runs after the render that created the variable has " +
				"finished, so the value it writes is already gone by the time anything reads it " +
				"again. The next render rebuilds the variable from scratch and the write is lost, " +
				"which looks like a value that silently resets. Hold it in state instead.",
		}
	case immutabilityFindingReassignedInAsync:
		return rule.Message{
			Id: messageImmutabilityReassignedInAsyncId,
			Description: "This assignment happens inside an async function, so it lands some time " +
				"after the render that created the variable has finished. The value is written " +
				"into a variable nothing will read again, and no re-render is scheduled, so the " +
				"result never reaches the screen. Hold it in state instead.",
		}
	case immutabilityFindingKnownMutableFunction:
		return rule.Message{
			Id: messageImmutabilityKnownMutableFunctionId,
			Description: "This function is handed to React, which keeps it past the end of this " +
				"render, and it writes to a local variable of this component. When React calls it " +
				"later, that variable belongs to a render that has already finished, so the write " +
				"goes nowhere and schedules nothing. Keep the value in state instead.",
		}
	}
	return rule.Message{
		Id:          messageImmutabilityImmutableValueId,
		Description: immutabilityReasonText(finding.Reasons) + " " + immutabilityWhyFrozenValuesMatter,
	}
}

// immutabilityReport points a finding at the syntax it came from.
//
// Through `ctx.ReportNode` rather than a Place's range, because a Place's range is the range of the
// node that produced the INSTRUCTION, and reporting through the node routes via `rule.TokenRange`,
// which trims leading trivia at the harness.
func immutabilityReport(ctx rule.Context, function *high_level_intermediate_representation.Function, finding immutabilityFinding) {
	message := immutabilityMessageFor(finding)
	if finding.Node != nil {
		ctx.ReportNode(finding.Node, message)
		return
	}
	if function.Node != nil {
		ctx.ReportNode(function.Node, message)
	}
}

// immutabilityHasBackEdge reports whether control can flow from a block to an earlier one.
//
// `Function.Blocks` is in reverse postorder, so a predecessor appearing at or after its successor in
// that ordering is a back edge and the function contains a loop. Asked by position rather than by
// walking terminals, because every terminal shape would otherwise have to be enumerated and a new
// one would silently answer false.
func immutabilityHasBackEdge(function *high_level_intermediate_representation.Function) bool {
	position := make(map[static_single_assignment.BlockId]int, len(function.Blocks))
	for index, block := range function.Blocks {
		position[block.Id] = index
	}
	for index, block := range function.Blocks {
		for _, predecessor := range block.Predecessors {
			if at, known := position[predecessor]; known && at >= index {
				return true
			}
		}
	}
	return false
}

// immutabilityIsRefValue reports whether a value is a React ref, which is exempt from this rule.
//
// Two signals, and BOTH are needed for the same reason refs.go documents at length: the checker
// sees a `useRef` call and cannot see a ref arriving as an untyped prop, while the name test sees
// the prop and would over-match on its own.
//
//   - The checker. `useRef` returns `RefObject`, identified by the TYPE'S SYMBOL rather than its
//     alias, because `RefObject` is an interface and its alias is correctly nil. That asymmetry is
//     measured in refs.go against the opposite case, where `Dispatch` lands on the alias.
//   - The name. React's `enableTreatRefLikeIdentifiersAsRefs` defaults to true and matches
//     `/^(?:[a-zA-Z$_][\w$]*)Ref$|^ref$/` against the binding whose `current` is accessed. Kirk's
//     tree spells these `handlersReference` rather than `handlersRef`, so the NAME half does not
//     fire on most of them and the checker half is what actually carries this exemption here.
//
// The property name is checked too, because the exemption is for writing a ref's `current` and not
// for writing arbitrary properties onto a ref object.
func immutabilityIsRefValue(ctx rule.Context, function *high_level_intermediate_representation.Function, state *immutabilityState, id static_single_assignment.IdentifierId) bool {
	if ctx.TypeChecker == nil || function == nil {
		return false
	}
	for _, candidate := range [2]static_single_assignment.IdentifierId{id, state.resolve(id)} {
		node := immutabilityIdentifierNode(function, candidate)
		if node == nil {
			continue
		}
		valueType := ctx.TypeChecker.GetTypeAtLocation(node)
		if valueType == nil {
			continue
		}
		// Compared positively rather than checked for non-nil: the shim is a hand-mirrored struct
		// read through `unsafe.Pointer` and has returned a silently wrong type before, so asserting
		// that what came back is what was asked for is the only check that can see that failure.
		if symbol := shimchecker.Type_symbol(valueType); symbol != nil {
			if symbol.Name == "RefObject" || symbol.Name == "MutableRefObject" {
				return true
			}
		}
	}
	if refsIsRefLikeName(immutabilityIdentifierName(function, id)) {
		return true
	}
	return refsIsRefLikeName(state.nameOf(function, id))
}

// immutabilityLoopBlocks is the set of blocks a back edge can re-enter.
//
// A back edge is an edge from a LATCH to a HEADER, found by reverse postorder position: a
// predecessor at or after its successor in `Function.Blocks` is one. The blocks in the loop are then
// the header plus everything that can REACH the latch through predecessors without leaving through
// the header, which is the standard natural-loop body.
//
// Position ranges are NOT sufficient and that was measured rather than assumed. For
// `for (const u of list) { acc[u.id] = 1; } useThing(acc);` the reverse postorder is b1 b2 b3 b5 b4,
// with b4 the latch and b5 the block AFTER the loop; taking the contiguous run from header to latch
// swept b5 in, so the `useThing` freeze counted as happening inside the loop and travelled back into
// the body on the next round, reporting the writes that built the accumulator. The predecessor walk
// gets b5 right because b5 cannot reach the latch.
func immutabilityLoopBlocks(function *high_level_intermediate_representation.Function) map[static_single_assignment.BlockId]bool {
	position := make(map[static_single_assignment.BlockId]int, len(function.Blocks))
	byId := make(map[static_single_assignment.BlockId]*high_level_intermediate_representation.BasicBlock, len(function.Blocks))
	for index, block := range function.Blocks {
		position[block.Id] = index
		byId[block.Id] = block
	}
	inside := map[static_single_assignment.BlockId]bool{}
	for index, header := range function.Blocks {
		for _, predecessor := range header.Predecessors {
			at, known := position[predecessor]
			if !known || at < index {
				continue
			}
			// `predecessor` is a latch and `header` is the loop header. Walk backwards from the
			// latch through predecessors, stopping at the header, to collect the loop body.
			inside[header.Id] = true
			pending := []static_single_assignment.BlockId{predecessor}
			for len(pending) > 0 {
				current := pending[len(pending)-1]
				pending = pending[:len(pending)-1]
				if current == header.Id || inside[current] {
					continue
				}
				inside[current] = true
				if block, found := byId[current]; found {
					pending = append(pending, block.Predecessors...)
				}
			}
		}
	}
	return inside
}
