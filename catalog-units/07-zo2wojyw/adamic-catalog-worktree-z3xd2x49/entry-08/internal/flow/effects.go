// Aliasing effects: what an instruction does to the values it reads and writes.
//
// Lifted from cohere's effects.go (high_level_intermediate_representation at 715ba94), which ports
// React's `InferMutationAliasingEffects`: its effect types and the three constructors, unchanged.
// What isn't lifted is how cohere makes effects, which is React's: a signature table keyed by the
// callee's syntactic name, with every other call assumed to mutate and capture everything. Adamic
// makes its own (infer.go), from the IR, where every runtime operation is its own node with one
// known effect.
package flow

// AliasingEffectKind is what one effect does. These are React's `AliasingEffect` variants.
//
// This is the real output type of this pass and it is NOT `Effect`. See the package comment: the
// two are different types, this one carries FROM and INTO places, and the scalar `Effect` is a
// per-operand projection of a whole list of these, computed later by the ranges pass.
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
	// effects before the ranges pass runs, and raises an invariant if one survives; this pass
	// resolves them the same way and never emits one. Kept as a constant because
	// `EffectGapInterproceduralParameters` is exactly the case where upstream would resolve one
	// and this does not.
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
// and give Into Store. See `ProjectEffects`.
func (k AliasingEffectKind) IsAliasing() bool {
	switch k {
	case AliasingEffectAssign, AliasingEffectAlias, AliasingEffectCapture,
		AliasingEffectCreateFrom, AliasingEffectMaybeAlias:
		return true
	default:
		return false
	}
}

// AliasingEffect is one effect an instruction has.
//
// From is the source of a data-flow edge and is meaningless for the mutation and create variants,
// where only Into is set. That asymmetry is upstream's: its enum gives each variant its own fields
// and Go's does not, so `HasFrom` names which variants read it rather than leaving a caller to
// infer it from a zero value. A zero IdentifierId is a real identifier, so an unset From is not
// distinguishable by value.
type AliasingEffect struct {
	Kind AliasingEffectKind

	// From is the source place of a data-flow edge. Only meaningful when HasFrom is true.
	From Place
	// Into is the target: the value created, mutated, or flowed into.
	Into Place
	// HasFrom reports whether From is set, because a zero IdentifierId is a legal value.
	HasFrom bool

	// Value is the kind created, meaningful only for AliasingEffectCreate.
	Value EffectValueKind
}

// EffectValueKind is the abstract kind a Create effect produces.
//
// This is upstream's `ValueKind` and it is deliberately NOT the same type as the six-element
// lattice `immutability.go` builds for itself. That rule's lattice is a private type with a join
// (`immutabilityMergeKinds`) tuned to what it reports; this kind also feeds the range graph's phi
// refinement. Sharing one type across the two would
// couple a rule's message selection to a substrate enum for no gain, and `immutability.go`'s
// version carries a reason bitset this one has no use for.
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

// AliasingEffects is the table this pass produces: the effect list of every instruction.
//
// Keyed by InstructionId rather than stored on the Instruction because an Instruction is shared
// through the per-file lowering cache and several rules hold the same one. Writing effects onto it
// would make one rule's pass visible to another that never asked for it, and `Construct` is not
// idempotent, so a table that has to be rebuilt is safer than a graph that has been mutated.
type AliasingEffects struct {
	byInstruction map[InstructionId][]AliasingEffect
}

// Get returns one instruction's effects, or nil when it has none.
func (e *AliasingEffects) Get(id InstructionId) []AliasingEffect {
	if e == nil {
		return nil
	}
	return e.byInstruction[id]
}

// Len is how many instructions carry at least one effect.
func (e *AliasingEffects) Len() int {
	if e == nil {
		return 0
	}
	return len(e.byInstruction)
}

// create is the Create effect, which every instruction producing a value emits first.
func create(into Place, kind EffectValueKind) AliasingEffect {
	return AliasingEffect{Kind: AliasingEffectCreate, Into: into, Value: kind}
}

// flow is one from/into effect.
func flow(kind AliasingEffectKind, from Place, into Place) AliasingEffect {
	return AliasingEffect{Kind: kind, From: from, Into: into, HasFrom: true}
}

// mutate is one mutation effect, which names only the value it writes.
func mutate(kind AliasingEffectKind, value Place) AliasingEffect {
	return AliasingEffect{Kind: kind, Into: value}
}
