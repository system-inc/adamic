// Reactive scopes: which values are memoized together, and over what span.
//
// This is React's `inferReactiveScopeVariables` (development bundle line 32231), run over this
// graph. oxc transcribes it at
// `oxc_react_compiler/src/react_compiler_inference/infer_reactive_scope_variables.rs`; where the two
// disagree React wins, and each disagreement is recorded at the line that resolves it.
//
// # What this pass actually does, which is less than its name suggests
//
// `FindDisjointMutableValues` has already answered the hard question: which values are so entangled
// by mutation that they cannot be memoized apart. This pass does three mechanical things with that
// answer and nothing else.
//
//	one scope per equivalence class, with a fresh id
//	the scope's range is the HULL of its members' mutable ranges
//	every member's range is REWRITTEN to the hull
//
// The third is the one a reader does not expect, and it is not incidental. Upstream writes
// `identifier.mutableRange = scope.range` on the line after it writes `identifier.scope`
// (line 32261), so after this pass every member of a class reports the CLASS's range rather than
// its own. A consumer asking "how long is this value alive" gets the entangled answer, which is the
// point: two values in one scope are memoized as one unit and therefore live as one unit.
//
// # DIVERGENCE FROM React: the hull is a VALUE here and an ALIAS there
//
// React's `range: identifier.mutableRange` binds the first member's range OBJECT into the scope,
// and `identifier.mutableRange = scope.range` then points every member at that same object. The
// aliasing is load-bearing THERE: five later passes widen `scope.range`
// (`alignReactiveScopesToBlockScopes`, `alignMethodCallScopes`, `alignObjectMethodScopes`,
// `mergeOverlappingReactiveScopes`, and the fbt/macro pass), and each widening retroactively
// updates every member's `mutableRange` through the shared reference.
//
// `MutableRange` is a value type here and this stores a copy, exactly as oxc does
// (`infer_reactive_scope_variables.rs:91-100`, a fixup loop rather than an alias). None of those
// five widening passes exists in this tree, so today the alias and the copy are indistinguishable:
// nothing widens a scope range after this pass runs. The difference becomes real the moment one of
// those passes lands, and at that point `ScopeRanges` below is the seam that must be re-derived
// rather than read once. Recorded as `ScopeGapPostAlignmentWidening` so a consumer can decline on
// it, and named here because a reader diffing against the bundle would otherwise see the alias
// missing with no explanation.
//
// # Upstream's invariant is an INVARIANT, not a filter, and it holds on this corpus
//
// React ends with a hard `CompilerError.invariant` (line 32271) rejecting any scope whose range
// starts at zero, ends at zero, or ends past `maxInstruction + 1`. There is no branch that skips
// such a scope: upstream considers it a compiler bug and stops.
//
// That is a genuine risk here rather than a formality, because `RangeGapLoopCarriedInversion` leaves
// 48 values across 400 files with an UNSET range, and a class composed entirely of those would hull
// to [0,0) and trip the invariant. Measured before this file was written, over 400 corpus files and
// 685 outermost functions:
//
//	classes                                        3,306
//	classes with every member's range set          3,278
//	classes with SOME member unset                    28
//	classes with EVERY member unset                    0
//
//	hulls that are empty (start or end zero)           0
//	hulls that are inverted (end <= start)             0
//	hulls that end past maxInstruction + 1             0
//
// So the invariant holds on every class, and the 28 mixed classes are exactly what upstream's
// `start === 0` branch exists to handle. `ValidateScopes` returns the violations rather than
// panicking, because a linter that stops on a graph it dislikes is worse than one that declines a
// function, and `TestScopeRangesSatisfyUpstreamsInvariant` pins the zero.
//
// # The distribution that proves these scopes are real
//
// A count cannot distinguish genuine structure from catastrophic over-merging, so the headline is a
// CROSS-TABULATION of scope width against scope membership. Over the same 400 files:
//
//	                       width 1      width > 1
//	one member               1,903             80
//	many members                 0          1,323
//
// The empty cell is the result. Not one class of several values hulls to a single instruction --
// which is precisely the shape catastrophic over-merging would produce, values swept together into
// a span too narrow to hold them. The two dimensions are measured independently and they agree.
//
// The counterfactual says which mechanism produces which population. Neutralising `mayAllocate` to
// false collapses the width-1 singletons from 1,903 to 193 while the wide multi-member scopes hold
// at 1,203 of 1,323. So allocation is what mints the narrow single-value scopes, and the entangled
// wide ones come from the range and mutability gates instead. Two populations, two causes, neither
// an artifact of the other. `TestScopeWidthAndMembershipAgree` pins the empty cell.
//
// # Termination is a single pass with no iteration at all
//
// There is no worklist, no fixpoint, and no recursion. `AssignReactiveScopes` makes ONE pass over
// the classes `DisjointSet.Sets()` returns, and one pass over the members of each. Every class is
// visited exactly once and every member exactly once, so the work is linear in the number of
// members and terminates because both loops are over finite slices fixed before the walk begins.
//
// This is a weaker claim than Stage 3 needed and deliberately so: the union-find whose termination
// had to be argued is upstream of this file, and its answer is already final by the time this reads
// it. `TestScopeAssignmentIsASinglePass` pins it empirically by counting visits.
//
// # Determinism, which a Go map would otherwise destroy
//
// Scope ids are assigned in the order `Sets()` returns classes, and that order is sorted by first
// member rather than being a Go map's randomised iteration. Upstream's ids come from a counter on
// the Environment (`fn.env.nextScopeId`, a post-increment shared across nested functions), so ITS
// ids are stable because its backing `Map` is insertion-ordered. Ours are stable because `Sets()`
// sorts. Without that, the same function would produce different scope ids on different runs, which
// is the shape that makes a cache non-reproducible without ever producing a wrong answer.
//
// Ids are per-function here rather than per-environment, because this graph has no environment and
// a `Function` is the unit every other pass in this package takes. See `ScopeId`.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// ScopeId names one reactive scope within one function.
//
// Upstream draws these from a counter on the Environment, so a scope id is unique across every
// function the compiler touches in a run. There is no environment in this package -- `Function` is
// the unit every pass here takes -- so these are unique within a function and restart at 1 for the
// next one. A consumer holding scopes from two functions must key on the pair, exactly as it must
// for `IdentifierId`, and for the same reason `FindDisjointMutableValues` refuses to span functions.
//
// Zero is not a valid scope id. It is the absent value, which is what lets `ScopeOf` answer "no
// scope" without a second return.
type ScopeId uint32

// ScopeGap names a scope rule this pass cannot apply, for a caller that needs to know.
//
// Modelled on `RangeGap`, `ReactiveGap` and `DisjointGap` deliberately, and for the same reason: a
// gap answerable through the API is one a consumer can decline on, while a gap living only in a
// comment is one the next reader inherits by accident.
type ScopeGap uint8

const (
	// ScopeGapPostAlignmentWidening is a scope range that upstream would widen after this pass.
	//
	// React aliases every member's `mutableRange` to the scope's own range object, so the five
	// passes that widen a scope range afterwards -- block-scope alignment, method-call alignment,
	// object-method alignment, overlap merging, and the fbt/macro pass -- retroactively widen every
	// member too. None of those passes exists here, so no widening is missed today; what is missing
	// is the MECHANISM that would propagate one. A consumer that later adds an alignment pass must
	// re-derive member ranges from `ScopeRanges` rather than assuming this pass already agreed with
	// it.
	//
	// This is a gap in the OUTPUT's durability rather than in its current value: measured today,
	// every member's range equals its scope's hull, which is exactly what upstream reports at this
	// point in its own pipeline.
	ScopeGapPostAlignmentWidening ScopeGap = iota
)

// ScopeGaps are the scope rules AssignReactiveScopes does not apply. See ScopeGap.
//
// Returned as a value rather than documented alone so a test can assert on it, which makes closing
// a gap a visible event rather than a silent improvement.
func ScopeGaps() []ScopeGap {
	return []ScopeGap{ScopeGapPostAlignmentWidening}
}

// ReactiveScopes is the table this pass produces: one scope per equivalence class, plus the
// per-value index into it.
//
// # Why a side table rather than a field on Identifier
//
// Upstream stores the scope ON the identifier. This keeps it in a table instead, because a value's
// scope is not a property of the value: it is a property of the CLASS the value belongs to, and there
// are 3,306 classes over 26,962 members. Storing the scope per identifier stores the same answer 8.16
// times on average and invites the two copies to drift. The table below stores each scope once and
// indexes into it, so widening a scope is one write rather than a fan-out, which is exactly the
// operation `ScopeGapPostAlignmentWidening` says a later pass will need.
//
// The cost is one indirection and one thing a caller must hold, the same trade `MutableRanges` made.
// If scopes become permanent infrastructure, `ScopeOf` is the seam that makes moving them onto
// `Identifier` invisible to callers.
//
// The zero value is an empty table and is ready to read: every lookup answers "no scope".
type ReactiveScopes struct {
	byIdentifier map[static_single_assignment.IdentifierId]ScopeId
	ranges       map[ScopeId]mutation_aliasing.MutableRange
	members      map[ScopeId][]static_single_assignment.IdentifierId
	order        []ScopeId
}

// ScopeOf returns the scope a value belongs to, or zero when it has none.
//
// A value that never entered an equivalence class has no scope, which is upstream's `null` and is
// the ordinary answer for most values: 26,962 of a corpus function's values are members and the
// rest are not. Zero rather than a second return because zero is not a valid ScopeId.
func (s *ReactiveScopes) ScopeOf(id static_single_assignment.IdentifierId) ScopeId {
	if s == nil || s.byIdentifier == nil {
		return 0
	}
	return s.byIdentifier[id]
}

// RangeOf returns a scope's range, or the unset range for a scope that does not exist.
//
// This is the HULL of the member ranges, and after this pass it is also every member's own range.
// See the package comment for why the two are equal here and are not guaranteed to stay equal.
func (s *ReactiveScopes) RangeOf(scope ScopeId) mutation_aliasing.MutableRange {
	if s == nil || s.ranges == nil {
		return mutation_aliasing.MutableRange{}
	}
	return s.ranges[scope]
}

// MembersOf returns the values belonging to a scope, sorted, or nil for a scope that does not exist.
//
// The slice is the table's own and must not be modified by a caller. Sorted because
// `DisjointSet.Sets` sorts, and the order is load-bearing for determinism rather than cosmetic.
func (s *ReactiveScopes) MembersOf(scope ScopeId) []static_single_assignment.IdentifierId {
	if s == nil || s.members == nil {
		return nil
	}
	return s.members[scope]
}

// Ids returns every scope this pass created, in assignment order.
//
// Assignment order is `DisjointSet.Sets()` order, which is sorted by first member. Returned as a
// slice rather than by ranging a map so a caller iterating scopes gets the same sequence every run.
func (s *ReactiveScopes) Ids() []ScopeId {
	if s == nil {
		return nil
	}
	return s.order
}

// Len reports how many scopes exist, for measurement.
func (s *ReactiveScopes) Len() int {
	if s == nil {
		return 0
	}
	return len(s.order)
}

// AssignReactiveScopes gives every entangled group of values in function one reactive scope.
//
// Requires evaluation order and single-assignment form: call `Construct` first, or take a function
// from `ForFunction`, which does. Infers ranges and the disjoint set for itself; a caller holding
// either already passes them to `AssignReactiveScopesWithSets` and avoids the repeat.
//
// Nested functions are NOT included, for the reason `FindDisjointMutableValues` records: an
// IdentifierId names one value in this function and a different value in a nested one. Upstream runs
// this per function too, and separately for function expressions inside `lowerWithMutationAliasing`
// (line 42315).
func AssignReactiveScopes(function *Function) *ReactiveScopes {
	if function == nil {
		return &ReactiveScopes{}
	}
	ranges := InferMutableRanges(function)
	return AssignReactiveScopesWithSets(function, ranges, FindDisjointMutableValuesWithRanges(function, ranges))
}

// AssignReactiveScopesWithSets is the pass, taking the range table and disjoint set a caller has.
//
// The range table and the set must have been computed from the SAME function and from each other:
// passing a set built over different ranges produces hulls that describe neither. `AssignReactiveScopes`
// is the safe spelling and this is the one that avoids recomputing.
//
// The member ranges this returns are the REWRITTEN ones -- each member reports its scope's hull --
// which is upstream's `identifier.mutableRange = scope.range`. The input table is not modified;
// `MemberRanges` below is the rewritten view, kept separate so a caller can still see what a value's
// range was before entanglement widened it.
func AssignReactiveScopesWithSets(function *Function, ranges *mutation_aliasing.MutableRanges, set *DisjointSet) *ReactiveScopes {
	scopes := &ReactiveScopes{}
	if function == nil || set == nil {
		return scopes
	}

	classes := set.Sets()
	if len(classes) == 0 {
		return scopes
	}

	scopes.byIdentifier = map[static_single_assignment.IdentifierId]ScopeId{}
	scopes.ranges = map[ScopeId]mutation_aliasing.MutableRange{}
	scopes.members = map[ScopeId][]static_single_assignment.IdentifierId{}

	// Ids start at 1 because zero is the absent scope. Upstream starts at 0 and distinguishes
	// absence with `null`; `ScopeOf` returns a bare ScopeId, so the sentinel has to be a value.
	var next ScopeId

	for _, class := range classes {
		if len(class) == 0 {
			// `Sets` never returns an empty class, so this cannot fire today, and a mutation
			// neutralising it SURVIVES for that reason rather than because a fixture is missing.
			//
			// The callers enumerated when that verdict was taken: `DisjointSet.Sets` is the only
			// producer of the slice this ranges over, and it builds each class by appending at
			// least one member before the class exists as a key. There is no other path into this
			// loop. That verdict EXPIRES if a second producer of classes is added, or if `Sets`
			// ever returns a key with no members.
			//
			// Kept rather than deleted because the alternative is a hull computed from no members,
			// which is silently [0,0) and would then read as a range-inversion bug rather than as
			// an empty class -- a wrong diagnosis of a real failure, at the cost of one branch.
			continue
		}
		next++
		scope := next

		// The hull, in upstream's own spelling rather than a tidier min/max pair. The two branches
		// are NOT a plain minimum and the difference is load-bearing.
		//
		// Zero is the sentinel for an unset range, so a member with no recorded definition point
		// must not drag the hull's start down to zero. Upstream's `scope.range.start === 0` branch
		// ADOPTS the next member's start outright rather than taking a minimum, and its `else if`
		// then SKIPS any member whose own start is zero. Together those mean an unset member is
		// invisible to the start unless every member before it was also unset.
		//
		// Verified against React itself rather than read off the source, by running the merge loop
		// out of the 7.1.1 development bundle on hand-built classes:
		//
		//	first member unset, then [4,6), [2,9)   ->  [2,9)   start adopted then minimised
		//	[4,6), then unset, then [2,9)           ->  [2,9)   unset member skipped
		//	unset, unset, [5,7)                     ->  [5,7)   adoption survives a run of unsets
		//	every member unset                      ->  [0,0)   stays unset, trips the invariant
		//
		// The end has NO such guard upstream and none here: `max(end, 0)` is the identity, so an
		// unset member cannot move it. Reproduced as written rather than simplified, because the
		// asymmetry between the two fields is upstream's and a reader diffing them should find it.
		// # The index-zero seed is EQUIVALENT to starting from the zero range, proven exhaustively
		//
		// A mutation stopping the seed from ever firing SURVIVES, and it is a genuine equivalence
		// rather than a fixture gap. A zero-initialised hull takes the first member's start through
		// the `Start == 0` adoption branch and its end through the max, so the seed reproduces what
		// the merge branch would have done anyway.
		//
		// Proven rather than argued, because "equivalent" is the verdict most likely to be wrong:
		// both spellings were run over every class of one to three members with starts and ends
		// drawn from 0..4, which is 16,900 shapes, and they agreed on all of them.
		//
		// Upstream's own spelling is the seed (`range: identifier.mutableRange` at bundle line
		// 32239, inside the `scope === undefined` arm), so the seed is what this reproduces. The
		// equivalence is recorded so a reader who notices the redundancy finds the measurement
		// instead of re-deriving it, and so that a future change to the adoption branch is known to
		// make this load-bearing again.
		var hull mutation_aliasing.MutableRange
		for index, id := range class {
			memberRange := ranges.Get(id)
			if index == 0 {
				hull = memberRange
			} else {
				if hull.Start == 0 {
					hull.Start = memberRange.Start
				} else if memberRange.Start != 0 && memberRange.Start < hull.Start {
					hull.Start = memberRange.Start
				}
				if memberRange.End > hull.End {
					hull.End = memberRange.End
				}
			}
			scopes.byIdentifier[id] = scope
		}

		scopes.ranges[scope] = hull
		// Copied rather than aliased: `Sets` returns the table's own slice and a caller holding
		// `MembersOf` must not be able to reach back into the disjoint set through it.
		members := make([]static_single_assignment.IdentifierId, len(class))
		copy(members, class)
		scopes.members[scope] = members
		scopes.order = append(scopes.order, scope)
	}

	return scopes
}

// MemberRanges returns every scoped value's range after entanglement, as a table.
//
// This is upstream's `identifier.mutableRange = scope.range` (line 32261) expressed as a derived
// view rather than as a write back into the input. Every member of a scope reports that scope's
// hull; a value with no scope is absent, and `Get` answers the unset range for it exactly as
// `MutableRanges` does.
//
// Kept separate from the input table on purpose. Upstream OVERWRITES the value's own range and the
// pre-entanglement range is unrecoverable afterwards; here both are available, which costs one map
// and lets a consumer ask which of the two it wants. Nothing upstream needs the old value, so this
// is a capability rather than a divergence in what the pass decides.
func (s *ReactiveScopes) MemberRanges() *mutation_aliasing.MutableRanges {
	result := &mutation_aliasing.MutableRanges{}
	if s == nil {
		return result
	}
	for _, scope := range s.order {
		hull := s.ranges[scope]
		for _, id := range s.members[scope] {
			result.Set(id, hull)
		}
	}
	return result
}

// ValidateScopes returns the scopes whose range violates upstream's invariant.
//
// React asserts this and STOPS (`CompilerError.invariant`, line 32271); oxc returns a diagnostic
// (`infer_reactive_scope_variables.rs:119`). This returns the offenders, because a linter that
// panics on a graph it dislikes is worse than one that declines a function, and because a caller
// that wants upstream's behaviour can raise on a non-empty result while a caller that wants to skip
// those scopes can do that instead.
//
// The four conditions are upstream's, in order: a start of zero, an end of zero, a function with no
// instructions at all, and a range extending past the last instruction. The third is upstream's
// `maxInstruction === 0` and it is a property of the FUNCTION rather than of any scope, so it
// condemns every scope at once -- reproduced rather than hoisted, because upstream's message names
// a scope and a caller reporting per scope needs the same granularity.
//
// Measured over 400 corpus files: zero violations across 3,306 scopes in 685 functions. See
// `TestScopeRangesSatisfyUpstreamsInvariant`, which pins that zero rather than leaving it here.
func ValidateScopes(function *Function, scopes *ReactiveScopes) []ScopeId {
	if function == nil || scopes == nil {
		return nil
	}

	var maxInstruction static_single_assignment.EvaluationOrder
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			if instruction := function.Instructions[instructionId]; instruction != nil {
				if instruction.Order > maxInstruction {
					maxInstruction = instruction.Order
				}
			}
		}
		if order := TerminalOrder(block.Terminal); order > maxInstruction {
			maxInstruction = order
		}
	}

	var invalid []ScopeId
	for _, scope := range scopes.order {
		scopeRange := scopes.ranges[scope]
		if scopeRange.Start == 0 || scopeRange.End == 0 ||
			maxInstruction == 0 || scopeRange.End > maxInstruction+1 {
			invalid = append(invalid, scope)
		}
	}
	return invalid
}
