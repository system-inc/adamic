// Classifying every instruction value by how strongly it needs memoizing.
//
// This is the level half of React's `computeMemoizationInputs`
// (`ReactiveScopes/PruneNonEscapingScopes.ts` at `react_conformance.UpstreamSha`). The rvalue half --
// which operands a value carries -- is `EachPlace`, which this package already has.
//
// # It is a transcription, which took measuring to establish
//
// This was scoped as the largest judgment call in the pipeline on the belief that our 67 instruction
// values had to be mapped onto upstream's smaller set. That 67 was a miscount: it came from the
// `EachPlace` switch, which covers instruction values, terminals and patterns together.
//
// The real numbers are 43 instruction values here against 47 upstream arms, and the difference is
// exactly the four composite values this tree models as `ReactiveValue` variants -- ternary,
// logical, sequence, optional -- plus one spelling difference, `JSXText` against `JsxText`.
// Normalised and diffed both ways: 47 and 47, nothing uncovered in either direction.
//
// So every arm below has one upstream counterpart and the classification is transcribed rather than
// argued.
//
// # The three option-dependent arms, which are where the judgment actually was
//
// Upstream computes two options from environment flags:
//
//	memoizeJsxElements     = !enableForest
//	forceMemoizePrimitives =  enableForest || enablePreserveExistingMemoizationGuarantees
//
// Schema defaults at `HIR/Environment.ts:205,240` are `enableForest: false` and
// `enablePreserveExistingMemoizationGuarantees: true`, so both options are true, which decides:
//
//	JsxExpression, JsxFragment  ->  Memoized     (Unmemoized under Forest)
//	the 13-kind primitive group ->  Conditional  (Never without the force)
//
// A mechanical extraction read the primitive group as `Conditional` by taking the first level in the
// arm. That is the forced branch; the default branch is `Never`. The extraction was right by
// accident, and anyone re-deriving this must resolve the flags rather than read the first branch.
package high_level_intermediate_representation

// MemoizationLevelOf returns how strongly a value needs memoizing.
//
// Upstream's per-kind classification, transcribed. The shared arms are kept as shared arms rather
// than expanded, so the grouping stays visible: the 13-kind primitive group in particular is one
// upstream decision and reads as one here.
//
// An unrecognised value answers `MemoizationNever`, matching upstream's `UnsupportedNode` arm. That
// is the only default and it is safe: a value nothing knows about is not a value worth holding.
func MemoizationLevelOf(value InstructionValue) MemoizationLevel {
	switch value.(type) {
	// Allocating shapes: each produces a fresh identity every evaluation, so anything comparing
	// across renders sees a change unless the value is held.
	case *ArrayExpression, *ObjectExpression, *FunctionExpression, *ObjectMethod,
		*NewExpression, *RegExpLiteral, *PropertyStore:
		return MemoizationMemoized

	// Calls. Upstream reaches the same level here and additionally consults a signature to decide
	// whether operands escape at all; see `MemoizationInputsGapCallSignatures`.
	case *CallExpression, *MethodCall, *TaggedTemplateExpression:
		return MemoizationMemoized

	// JSX, under `memoizeJsxElements`. True by default; `Unmemoized` only under Forest.
	case *JsxExpression, *JsxFragment:
		return MemoizationMemoized

	// Context writes allocate a binding that outlives the instruction.
	case *DeclareContext, *StoreContext:
		return MemoizationMemoized

	// The primitive group, under `forceMemoizePrimitives`. True by default; `Never` without it.
	//
	// Thirteen kinds in one upstream arm, kept together here for the same reason: the grouping is
	// the decision. Its comment is that these produce primitives and are usually not escape points,
	// but that forcing memoization walks their rvalues so transitively reachable scopes are still
	// considered.
	case *NextPropertyOf, *StartMemoize, *FinishMemoize, *Debugger, *ComputedDelete,
		*PropertyDelete, *LoadGlobal, *MetaProperty, *TemplateLiteral, *Primitive,
		*JsxText, *BinaryExpression, *UnaryExpression:
		return MemoizationConditional

	// Propagating shapes: they produce no new identity, so whether they need memoizing is entirely
	// a question about what flows through them.
	case *LoadLocal, *LoadContext, *PropertyLoad, *ComputedLoad, *ComputedStore,
		*StoreLocal, *Destructure, *GetIterator, *IteratorNext,
		*PrefixUpdate, *PostfixUpdate, *Await, *TypeCastExpression:
		return MemoizationConditional

	// Declared but not held: not comparable with `Object.is`, and left alone unless forced.
	case *DeclareLocal, *StoreGlobal:
		return MemoizationUnmemoized

	default:
		return MemoizationNever
	}
}

// MemoizationLevelOfReactiveValue returns the level of a value in the tree, composites included.
//
// The four composite values are terminals in the graph and values in the tree, so they have no
// `InstructionValue` to classify. Upstream classifies all four as `Conditional` in separate arms
// with the same answer, which is consistent with what they are: none produces an identity of its
// own, so each is memoized exactly when what flows through it is.
func MemoizationLevelOfReactiveValue(value ReactiveValue) MemoizationLevel {
	switch shape := value.(type) {
	case *ReactiveInstructionValue:
		if shape.Value == nil {
			return MemoizationNever
		}
		return MemoizationLevelOf(shape.Value)
	case *ReactiveLogicalValue, *ReactiveTernaryValue, *ReactiveSequenceValue,
		*ReactiveOptionalValue:
		return MemoizationConditional
	}
	return MemoizationNever
}

// MemoizationInputsGap names a rule this classification does not apply.
type MemoizationInputsGap uint8

const (
	// MemoizationInputsGapCallSignatures is the `noAlias` consultation, not performed.
	//
	// Upstream's call arms ask `getFunctionCallSignature` whether the callee aliases its arguments;
	// a `noAlias` callee returns no rvalues at all, so its operands do not escape through it. That
	// lookup is keyed by type and is total -- every call has a signature.
	//
	// This tree has `lookupSignature`, keyed by callee name against two static tables holding 29
	// entries between them, against 57,446 corpus call sites. A miss is the overwhelming case.
	//
	// That table is correct where it is used today, because a miss falls through to
	// `effectsForUnknownCall`, which treats every operand pessimistically -- missing costs precision
	// and cannot cost correctness. `noAlias` inverts that: answering false too often keeps operands
	// that should not escape, which keeps a scope upstream prunes, which the rule then validates a
	// memoization against. A default safe under one consumer is a defect under another.
	//
	// So the consultation is declined rather than approximated. Closing it needs a type-keyed
	// lookup through the resident checker, which upstream does not have and which
	// `IsAlwaysInvalidatingType` already shows is smaller than upstream's apparatus suggests.
	MemoizationInputsGapCallSignatures MemoizationInputsGap = iota

	// MemoizationInputsGapOperandMutability is the per-operand mutability test, approximated.
	//
	// Upstream filters a call's operands by `isMutableEffect` to decide which are themselves
	// memoization outputs. That reads `Place.Effect`, which is unpopulated here -- 51,945 of 51,945
	// corpus places are `EffectUnknown` -- so the aliasing model answers instead: does any mutating
	// effect on this instruction name this operand as its target.
	//
	// The translation is sound and measured in both directions: over 200 corpus files, 3,864
	// mutating effects on call-shaped instructions and 3,864 attributed operands, with zero effects
	// targeting anything the instruction does not name. No effect goes unplaced and no attribution
	// is phantom.
	//
	// What it is not is precise. Split by kind, 3,750 of those 3,864 are
	// `mutate-transitive-conditionally`, the documented default for a call with no known signature;
	// only 114 are a precise `mutate`. Upstream counts conditional forms too, so that is faithful
	// rather than distorted -- but upstream reaches its states through the total signature lookup
	// above and this defaults there. Roughly 97 percent of call-operand mutability answers come from
	// the no-signature default, and closing the gap above closes this one too.
	MemoizationInputsGapOperandMutability
)

// MemoizationInputsGaps are the rules this classification does not apply. See MemoizationInputsGap.
//
// Returned as a value rather than documented alone so a test can assert on it, which makes closing a
// gap a visible event rather than a silent improvement.
func MemoizationInputsGaps() []MemoizationInputsGap {
	return []MemoizationInputsGap{
		MemoizationInputsGapCallSignatures,
		MemoizationInputsGapOperandMutability,
	}
}
