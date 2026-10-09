// Package mutation_aliasing is React's mutation and aliasing model, shared by two intermediate
// representations: cohere's high-level IR (the React Compiler's, in
// internal/lint/ecmascript/high_level_intermediate_representation) and Adamic's flow graph.
//
// It holds two things, both named after React's own passes, InferMutationAliasingEffects and
// InferMutationAliasingRanges:
//
//   - The effect vocabulary: AliasingEffect and its kinds, what an instruction does to the values it
//     reads and writes. Each IR makes its own effects, cohere from React's signature table keyed by
//     the callee's name and Adamic from its IR, where every runtime operation has one known effect.
//     Both speak this vocabulary.
//   - Mutable ranges: the alias graph those effects build, the mutate worklist over it, and the range
//     of evaluation over which each value is still being written.
//
// The ranges are React Compiler semantics, held to parity with upstream, so cohere's behaviour is
// the default. Where the two IRs differ, the difference is a method on Graph or a field of Options
// rather than a fork of the algorithm. Graph says which method is which seam.
//
// It is a module of its own, beside static_single_assignment rather than inside it, because that
// module holds no IR's meaning and this one is React's. Adamic consumes both by reference through its
// cohere submodule, since Go forbids importing cohere's internal/ from another module.
package mutation_aliasing

import "github.com/system-inc/cohere/static_single_assignment"

// AliasingEffectKind is what one effect does. These are React's `AliasingEffect` variants.
//
// This is the real output type of the effect inference and it is NOT `Effect`. See cohere's
// effects.go: the two are different types, this one carries FROM and INTO places, and the scalar
// `Effect` is a per-operand projection of a whole list of these, computed later from the ranges.
//
// The declaration order carries no meaning and there is no join over it. Unlike `ValueKind`, which
// `immutability.go` merges with a real lattice operation, these are not merged at all: an
// instruction produces a LIST of effects and they are all applied in order. A pass that tried to
// reduce the list to one element would lose the from/into pairing that is the entire content.
type AliasingEffectKind uint8

const (
	// AliasingEffectCreate makes a new value of a given kind at Into. From is unset.
	AliasingEffectCreate AliasingEffectKind = iota
	// AliasingEffectCreateFrom makes a new value at Into with the same kind as From.
	AliasingEffectCreateFrom
	// AliasingEffectAssign is `Into = From`, a direct assignment.
	AliasingEffectAssign
	// AliasingEffectAlias means mutating Into implies mutating From. Direct aliasing.
	AliasingEffectAlias
	// AliasingEffectMaybeAlias is a potential aliasing relationship, used for unknown callees.
	AliasingEffectMaybeAlias
	// AliasingEffectCapture is information flow from From into Into without aliasing.
	AliasingEffectCapture
	// AliasingEffectImmutableCapture is data flow that escape analysis sees and mutable ranges do
	// not. Projects to Read rather than to Capture.
	AliasingEffectImmutableCapture
	// AliasingEffectMutate mutates the value and its direct aliases.
	AliasingEffectMutate
	// AliasingEffectMutateConditionally mutates only if the value is mutable.
	AliasingEffectMutateConditionally
	// AliasingEffectMutateTransitive mutates the value and everything it transitively captured.
	AliasingEffectMutateTransitive
	// AliasingEffectMutateTransitiveConditionally is the conditional form, and is the DEFAULT
	// applied to every operand of a call with no known signature.
	AliasingEffectMutateTransitiveConditionally
	// AliasingEffectFreeze marks the value and its direct aliases as frozen.
	AliasingEffectFreeze
	// AliasingEffectApply is an unresolved call. Upstream replaces every Apply with more precise
	// effects before the ranges pass runs, and raises an invariant if one survives; cohere's
	// inference resolves them the same way and never emits one. Kept as a constant because
	// `EffectGapInterproceduralParameters` is exactly the case where upstream would resolve one
	// and cohere does not.
	AliasingEffectApply
)

func (k AliasingEffectKind) String() string {
	switch k {
	case AliasingEffectCreate:
		return "create"
	case AliasingEffectCreateFrom:
		return "create-from"
	case AliasingEffectAssign:
		return "assign"
	case AliasingEffectAlias:
		return "alias"
	case AliasingEffectMaybeAlias:
		return "maybe-alias"
	case AliasingEffectCapture:
		return "capture"
	case AliasingEffectImmutableCapture:
		return "immutable-capture"
	case AliasingEffectMutate:
		return "mutate"
	case AliasingEffectMutateConditionally:
		return "mutate-conditionally"
	case AliasingEffectMutateTransitive:
		return "mutate-transitive"
	case AliasingEffectMutateTransitiveConditionally:
		return "mutate-transitive-conditionally"
	case AliasingEffectFreeze:
		return "freeze"
	case AliasingEffectApply:
		return "apply"
	default:
		return "<unknown>"
	}
}

// IsMutation reports whether this effect widens a value's mutable range.
//
// This is the exact predicate the ranges pass reads to build its `mutations` list, at
// `infer_mutation_aliasing_ranges.rs` where `mutations` is populated from effects matching
// `Mutate | MutateConditionally | MutateTransitive | MutateTransitiveConditionally`. Exposed here
// rather than restated there so the two passes cannot drift.
func (k AliasingEffectKind) IsMutation() bool {
	switch k {
	case AliasingEffectMutate, AliasingEffectMutateConditionally,
		AliasingEffectMutateTransitive, AliasingEffectMutateTransitiveConditionally:
		return true
	default:
		return false
	}
}

// IsAliasing reports whether this effect creates a from/into data-flow edge.
//
// These are the five variants the ranges pass treats identically when projecting to a scalar
// `Effect`: all five give From either Capture or Read depending on whether Into is still mutable,
// and give Into Store. See cohere's `ProjectEffects`.
func (k AliasingEffectKind) IsAliasing() bool {
	switch k {
	case AliasingEffectAssign, AliasingEffectAlias, AliasingEffectCapture,
		AliasingEffectCreateFrom, AliasingEffectMaybeAlias:
		return true
	default:
		return false
	}
}

// AliasingEffect is one effect an instruction has, over an IR's place type P.
//
// From is the source of a data-flow edge and is meaningless for the mutation and create variants,
// where only Into is set. That asymmetry is upstream's: its enum gives each variant its own fields
// and Go's does not, so `HasFrom` names which variants read it rather than leaving a caller to
// infer it from a zero value. A zero IdentifierId is a real identifier, so an unset From is not
// distinguishable by value.
type AliasingEffect[P any] struct {
	Kind AliasingEffectKind

	// From is the source place of a data-flow edge. Only meaningful when HasFrom is true.
	From P
	// Into is the target: the value created, mutated, or flowed into.
	Into P
	// HasFrom reports whether From is set, because a zero IdentifierId is a legal value.
	HasFrom bool

	// Value is the kind created, meaningful only for AliasingEffectCreate.
	Value EffectValueKind
}

// CreateEffect is the Create effect, which every instruction producing a value emits first.
func CreateEffect[P any](into P, kind EffectValueKind) AliasingEffect[P] {
	return AliasingEffect[P]{Kind: AliasingEffectCreate, Into: into, Value: kind}
}

// FlowEffect is one from/into effect.
func FlowEffect[P any](kind AliasingEffectKind, from P, into P) AliasingEffect[P] {
	return AliasingEffect[P]{Kind: kind, From: from, Into: into, HasFrom: true}
}

// MutationEffect is one mutation effect, which names only the value it writes.
func MutationEffect[P any](kind AliasingEffectKind, value P) AliasingEffect[P] {
	return AliasingEffect[P]{Kind: kind, Into: value}
}

// EffectValueKind is the abstract kind a Create effect produces.
//
// This is upstream's `ValueKind` and it is deliberately NOT the same type as the six-element
// lattice cohere's `immutability.go` builds for itself. That rule's lattice is a private type with a
// join (`immutabilityMergeKinds`) tuned to what it reports; this kind also feeds the range graph's
// phi refinement. Sharing one type across the two would couple a rule's message selection to a
// substrate enum for no gain, and `immutability.go`'s version carries a reason bitset this one has no
// use for.
type EffectValueKind uint8

const (
	// EffectValueMutable is a freshly created value this code may still write to.
	EffectValueMutable EffectValueKind = iota
	// EffectValuePrimitive is a number, string, boolean or similar.
	EffectValuePrimitive
	// EffectValueFrozen is a value that must not be mutated from here on.
	EffectValueFrozen
	EffectValueMaybeFrozen
	EffectValueGlobal
)

func (k EffectValueKind) String() string {
	switch k {
	case EffectValuePrimitive:
		return "primitive"
	case EffectValueFrozen:
		return "frozen"
	case EffectValueMaybeFrozen:
		return "maybe-frozen"
	case EffectValueGlobal:
		return "global"
	default:
		return "mutable"
	}
}

// Graph is what the alias graph and the ranges need of an intermediate representation: everything
// single assignment needs, and one method more per seam where cohere's IR and Adamic's differ.
//
// An IR implements it on a type of its own, which carries the IR's effect table, and calls a pass
// with that value and its function. The passes hold no IR type; every read of one goes through here.
type Graph[F any, B comparable, P any] interface {
	static_single_assignment.Graph[F, B, P]

	// InstructionOrder is the evaluation order of a block's instruction at index, and false when the
	// block names an instruction the function does not hold, which every pass skips.
	InstructionOrder(function F, block B, index int) (static_single_assignment.EvaluationOrder, bool)
	// TerminalOrder is the evaluation order of a block's terminal.
	TerminalOrder(block B) static_single_assignment.EvaluationOrder

	// Effects are the effects of a block's instruction at index, in the order they apply. This is the
	// effects input: cohere's come from its signature-table inference, Adamic's from its own.
	Effects(function F, block B, index int) []AliasingEffect[P]

	// ParametersFrozen reports whether the function's parameters arrive Frozen. React's answer is a
	// component's or a hook's, since React owns the arguments it passes them; Adamic's is never.
	ParametersFrozen(function F) bool
	// Context are the values a nested function captures, defined on entry like its parameters. An IR
	// that keeps captured variables out of its graph has none.
	Context(function F) []P
	// ReturnValue is the value a block's Return terminal returns, which the pass aliases into the
	// Returns place, and false for a block that does not return one. Asked only of an IR whose Returns
	// is not nil.
	ReturnValue(block B) (P, bool)
	// StoredContextValue is the value a block's instruction at index stores into a captured binding,
	// when it is such a store: React widens its range from the instruction's shape alone.
	StoredContextValue(function F, block B, index int) (static_single_assignment.IdentifierId, bool)
	// Closure is asked at every Create of a block's instruction at index whose Into is into. When the
	// instruction makes a closure whose value is into, it returns the closure's captures, which a
	// later Freeze of the closure freezes, and whether React creates the closure Frozen: every capture
	// immutable and its body only reading them. kinds are the abstract kinds the pass holds at that
	// point, an immutable kind for each value that is not Mutable, and are only read. An IR with no
	// such rule returns nil and false.
	Closure(function F, block B, index int, into static_single_assignment.IdentifierId,
		kinds map[static_single_assignment.IdentifierId]EffectValueKind) (captures []P, frozen bool)
}

// Options are what a caller sets on the pass, rather than what its IR answers.
type Options struct {
	// ParametersDefinedOnEntry opens every parameter's range at the function's first instruction,
	// rather than at its first read, and gives one nothing widened that instruction alone.
	//
	// Adamic's rule (its 4748a636), and false by default, so cohere keeps React's: there, a
	// component's or hook's parameters are Frozen and nothing the body does mutates them. An Adamic
	// parameter is an owned or borrowed value its body may mutate, and one mutated before its first
	// read (through another value that reaches it: this.list and a node of it, say) would otherwise be
	// mutated outside its range. A parameter then always has a range, so a mutation the graph missed
	// is outside it rather than on a range the pass declined to set.
	ParametersDefinedOnEntry bool

	// ContextKinds seeds the abstract kind of captured values, for a caller that knows them: cohere's
	// read-only-closure probe passes the kinds its enclosing function holds for the captures. A value
	// absent here, or Mutable, is created Mutable.
	ContextKinds map[static_single_assignment.IdentifierId]EffectValueKind
}
