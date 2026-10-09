// Flattening the scopes a hook call sits in: a hook cannot be called conditionally, so no scope around
// one can be memoized.
//
// This is React's `flattenScopesWithHooksOrUseHIR`, as bundled in eslint-plugin-react-hooks 7.1.1,
// run over this graph. Upstream runs it immediately after `flattenReactiveLoopsHIR`, and so does
// `AnalyzePreservedManualMemoization`.
//
// # What it does
//
// Every scope still open when the walk reaches a hook call is flattened. Memoizing one would skip the
// hook on a render where the scope's dependencies did not change, which breaks the rules of hooks. A
// scope whose whole body is the hook call becomes a plain label, since there is nothing left to
// memoize. Any other becomes a pruned scope, which keeps its body and the scopes nested in it.
//
// # Why its absence reported memoization React Compiler keeps
//
// A pruned scope does not stand as a parent to the scopes inside it, so
// `MergeReactiveScopesThatInvalidateTogether` leaves a nested scope alone. Without this pass the hook
// call's scope stayed live, a nested scope with the same dependencies was folded into it and its id
// was lost, and preserve-manual-memoization read the nested memo block as unmemoized.
//
// The shape that measured it is a `useCallback` passed inline to `React.useSyncExternalStore`.
// Upstream's InferReactiveScopeVariables puts a method call's property in the call's own scope, so
// that scope opens at `useSyncExternalStore`, before the arguments, and the callback's scope nests
// inside it. TanStack Query's useMutation.ts and useMutationState.ts reported three findings there
// that React Compiler compiles clean (#zx5xvtg). The namespace spelling is the commonest way in and
// not the only one: useMutationState's shape reported with a bare `useSyncExternalStore` import too.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// FlattenScopesWithHooksOrUse flattens every reactive scope a hook call or `use` sits in.
//
// Returns the scope ids it pruned, for `BuildReactiveFunctionWithFlattenedScopes` to mark, the same
// channel `FlattenReactiveLoops` uses because this IR has no `PrunedScope` terminal. A scope that
// becomes a label is rewritten in place to the `Label` terminal, which this IR does have, so it is
// not in the returned set.
func FlattenScopesWithHooksOrUse(function *Function) map[ScopeId]bool {
	if function == nil {
		return nil
	}

	type activeScope struct {
		block         static_single_assignment.BlockId
		fallthroughId static_single_assignment.BlockId
	}
	var activeScopes []activeScope
	var prune []static_single_assignment.BlockId

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		// A scope whose fallthrough is this block has closed, so a hook call from here on is outside
		// it. Upstream's `retainWhere`, before the block's instructions are read.
		open := activeScopes[:0]
		for _, scope := range activeScopes {
			if scope.fallthroughId != block.Id {
				open = append(open, scope)
			}
		}
		activeScopes = open

		for _, instructionId := range block.Instructions {
			instruction := function.Instruction(instructionId)
			if instruction == nil || !callsHookOrUse(function, instruction) {
				continue
			}
			for _, scope := range activeScopes {
				prune = append(prune, scope.block)
			}
			activeScopes = activeScopes[:0]
		}

		if terminal, isScope := block.Terminal.(*Scope); isScope {
			activeScopes = append(activeScopes, activeScope{block: block.Id, fallthroughId: terminal.Fallthrough})
		}
	}

	flattened := map[ScopeId]bool{}
	for _, blockId := range prune {
		block, found := function.Block(blockId)
		if !found {
			continue
		}
		terminal, isScope := block.Terminal.(*Scope)
		if !isScope {
			continue
		}
		body, found := function.Block(terminal.Block)
		if found && len(body.Instructions) == 1 {
			if jump, isGoto := body.Terminal.(*Goto); isGoto && jump.Block == terminal.Fallthrough {
				block.Terminal = &Label{Block: terminal.Block, Fallthrough: terminal.Fallthrough, Order: terminal.Order}
				continue
			}
		}
		flattened[terminal.Scope] = true
	}

	if len(flattened) == 0 {
		return nil
	}
	return flattened
}

// callsHookOrUse reports an instruction that calls a hook or React's `use`, upstream's
// `getHookKind(fn.env, callee.identifier) != null || isUseOperator(callee.identifier)`.
//
// A hook is found by its name, through `IsHookCallee`, whose comment records how that differs from
// upstream's shape registry. `use` is found through the callee's module origin instead, since its
// name does not follow the hook convention and only the import says it is React's.
func callsHookOrUse(function *Function, instruction *Instruction) bool {
	switch call := instruction.Value.(type) {
	case *CallExpression:
		return IsHookCallee(function, call.Callee.Identifier) || isReactUse(call.CalleeOrigin)
	case *MethodCall:
		return IsHookCallee(function, call.Property.Identifier) || isReactUse(call.CalleeOrigin)
	}
	return false
}

func isReactUse(origin ModuleExportOrigin) bool {
	return origin.Module == "react" && origin.Export == "use"
}
