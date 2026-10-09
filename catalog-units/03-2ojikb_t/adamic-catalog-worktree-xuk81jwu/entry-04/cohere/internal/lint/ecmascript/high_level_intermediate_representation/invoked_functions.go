// Which nested functions are assumed to be invoked before their scope ends.
//
// React's `getAssumedInvokedFunctions` (`CollectHoistablePropertyLoads.ts`). The hoistable analysis
// descends into a callback only when it can assume the callback actually runs, because a property
// read inside a function that may never be called is not safe to hoist out of it.
//
// # What this buys, in one fixture pair
//
// `useCallback-infer-more-specific.ts` and `useMemo-infer-more-specific.ts` differ by one word:
//
//	useCallback(() => [x.y.z], [x])
//	useMemo(() => [x.y.z], [x])
//
// Upstream infers `x.y.z` for both. Measured before this pass existed, this tree inferred `x.y.z`
// for the first and bare `x` for the second, and the two lower to identical instruction streams
// except that `useMemo` ends in a `CallExpression` where `useCallback` ends in a `LoadLocal`. The
// dependency streams are identical too, and so is the nested hoistable seed. What differs is only
// that one callback is invoked -- which is exactly what this pass detects, and exactly what lets the
// hoistable analysis descend and make `x.y.z` hoistable in the enclosing scope.
//
// # Conservative in one direction only
//
// A false negative costs precision: the analysis does not descend, a path truncates to its root, and
// the scope invalidates more often than it needs to. A false positive is the dangerous direction --
// treating a read inside an uncalled function as hoistable moves a load to somewhere it may throw.
// So every rule here requires a syntactic reason to believe the call happens, and upstream's own
// comment marks the two places it accepts an assumption rather than a proof: arguments to hooks, and
// values handed to JSX.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// AssumedInvokedFunctions returns the nested functions this one is assumed to invoke.
//
// Keyed by `FunctionId` rather than by the lowered function pointer upstream uses, because a nested
// function is addressed by index here.
type AssumedInvokedFunctions map[FunctionId]bool

// invokedCandidate is a value known to hold a nested function, plus what that function may invoke.
//
// `mayInvoke` is upstream's transitive edge: if lambda A is invoked and A calls B, then B runs too.
// Recorded during the walk and resolved at the end, because A may be seen before B is known.
type invokedCandidate struct {
	function  FunctionId
	mayInvoke map[FunctionId]bool
}

// CollectAssumedInvokedFunctions reports which nested functions are assumed to run.
//
// Two passes, upstream's. The first maps every value that holds a nested function, following the
// stores and loads that move one around. The second reads the call sites.
func CollectAssumedInvokedFunctions(function *Function) AssumedInvokedFunctions {
	invoked := AssumedInvokedFunctions{}
	if function == nil {
		return invoked
	}
	candidates := map[static_single_assignment.IdentifierId]*invokedCandidate{}
	collectInvokedCandidates(function, candidates)
	collectInvocations(function, candidates, invoked)

	// Transitive closure over `mayInvoke`. A function reached only through another invoked function
	// still runs, which is upstream's `conditional-call-chain` case: "if lambda A calls lambda B, we
	// assume lambda B is safe to invoke if lambda A is -- even if lambda B is conditionally called".
	//
	// Iterated to a fixpoint rather than recursed, because the edges are discovered in instruction
	// order and a later function may name an earlier one. Bounded by the number of nested functions,
	// since each pass either adds one or stops.
	for {
		grew := false
		for _, candidate := range candidates {
			if !invoked[candidate.function] {
				continue
			}
			for reached := range candidate.mayInvoke {
				if !invoked[reached] {
					invoked[reached] = true
					grew = true
				}
			}
		}
		if !grew {
			return invoked
		}
	}
}

// collectInvokedCandidates maps each value that holds a nested function to which one it holds.
//
// Upstream's step 1, and its comment gives the reason it only matches `FunctionExpression`:
// "conservatively only match function expressions which can have guaranteed ssa. ObjectMethods and
// ObjectProperties do not". A value whose definition is not unique cannot be traced to one function.
func collectInvokedCandidates(function *Function, candidates map[static_single_assignment.IdentifierId]*invokedCandidate) {
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			switch value := instruction.Value.(type) {
			case *FunctionExpression:
				candidates[instruction.LValue.Identifier] = &invokedCandidate{
					function:  value.Function,
					mayInvoke: map[FunctionId]bool{},
				}
			case *StoreLocal:
				if held, ok := candidates[value.Value.Identifier]; ok {
					candidates[value.LValue.Identifier] = held
				}
			case *LoadLocal:
				if held, ok := candidates[value.Place.Identifier]; ok {
					candidates[instruction.LValue.Identifier] = held
				}
			}
		}
	}
}

// collectInvocations reads the call sites, upstream's step 2.
func collectInvocations(function *Function, candidates map[static_single_assignment.IdentifierId]*invokedCandidate,
	invoked AssumedInvokedFunctions) {
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			switch value := instruction.Value.(type) {
			case *CallExpression:
				if held, ok := candidates[value.Callee.Identifier]; ok {
					// A direct call, which is the only case here that is a proof rather than an
					// assumption.
					invoked[held.function] = true
					break
				}
				if !IsHookCallee(function, value.Callee.Identifier) {
					break
				}
				// Upstream: "assume arguments to all hooks are safe to invoke". This is what makes
				// `useMemo(() => ..., deps)` descend at all, since the callback is an argument
				// rather than a callee.
				for _, argument := range value.Args {
					if held, ok := candidates[argument.Place.Identifier]; ok {
						invoked[held.function] = true
					}
				}

			case *JsxExpression:
				// Upstream: "assume JSX attributes and children are safe to invoke". A spread is
				// skipped because the value it holds is not one identifiable place.
				for _, attribute := range value.Props {
					if attribute.Spread {
						continue
					}
					if held, ok := candidates[attribute.Value.Identifier]; ok {
						invoked[held.function] = true
					}
				}
				for _, child := range value.Children {
					if held, ok := candidates[child.Identifier]; ok {
						invoked[held.function] = true
					}
				}

			case *FunctionExpression:
				// Recorded as an edge rather than resolved now: this function is invoked only if
				// the one holding it is, which the closure above decides.
				nested := function.Functions[value.Function]
				if nested == nil {
					break
				}
				holder, ok := candidates[instruction.LValue.Identifier]
				if !ok {
					break
				}
				for id := range invokedWithinNested(function, nested) {
					holder.mayInvoke[id] = true
				}
			}
		}

		// Upstream: "assume directly returned functions are safe to call". A function handed back to
		// React is one React runs.
		if terminal, ok := block.Terminal.(*Return); ok {
			if held, present := candidates[terminal.Value.Identifier]; present {
				invoked[held.function] = true
			}
		}
	}
}

// invokedWithinNested reports which of THIS function's nested functions an inner one invokes.
//
// The inner function numbers its own children from zero, so its answer cannot be read directly in
// the parent's space. What survives the boundary is a capture: a function the inner one calls, that
// the parent also holds, arrives through `Context`. Anything else is inner-local and unreachable
// from here by construction.
func invokedWithinNested(parent *Function, nested *Function) map[FunctionId]bool {
	reached := map[FunctionId]bool{}
	if nested == nil || len(nested.Context) == 0 {
		return reached
	}
	innerInvoked := CollectAssumedInvokedFunctions(nested)
	if len(innerInvoked) == 0 {
		return reached
	}
	// Which of the inner function's captures name a function the parent holds.
	parentFunctions := map[static_single_assignment.IdentifierId]FunctionId{}
	for _, block := range parent.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := parent.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if expression, ok := instruction.Value.(*FunctionExpression); ok {
				parentFunctions[instruction.LValue.Identifier] = expression.Function
			}
		}
	}
	for _, context := range nested.Context {
		if id, ok := parentFunctions[context.Identifier]; ok {
			reached[id] = true
		}
	}
	return reached
}
