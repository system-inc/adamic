# Shared flow design

Design only. Inspected Adamic main e8ba3d5 and cohere ca939db5. This document does not switch any pass. Algorithms stay in the cohere submodule; Adamic supplies adapters and its own IR semantics, never copied implementations.

## Modules and ownership

Adamic will require github.com/system-inc/cohere/static_single_assignment and github.com/system-inc/cohere/mutation_aliasing at v0.0.0 and replace each with its directory under ./cohere. Both replaces belong in Adamic's go.mod: dependency replaces are not inherited. The shared SSA module imports no checker. Mutation aliasing depends on that SSA module.

Keep flow.go's Function, BasicBlock, Instruction, Identifier, Place and terminals, build.go's IR-to-graph construction, liveness.go's backward analysis, infer.go's IR effect inference and escaped-value fixed point. Neither shared module builds Adamic's graph or decides which IR operations mutate. Captured variables and globals remain outside this graph. Each nested function remains a separate graph.

Retire graph.go, ssa.go, ssa_eliminate.go, ssa_verify.go and ranges.go, deleting each by its literal filename after its callers have moved. Remove the shared vocabulary and constructors from effects.go; retain the Adamic instruction-to-effects table or move that table next to infer.go. Do not retain algorithm copies under different filenames. Small forwarding functions are acceptable to preserve existing callers and unchanged range and trace tests.

The manifest cohere-mirrors.json is absent on this inspected Adamic main. If the DRY guard lands before implementation, remove only the retired entries then. Otherwise record the exact retired files for integration; do not invent a manifest format.

## SSA graph adapter

Implement static_single_assignment.Graph[*Function, *BasicBlock, Place] on an Adamic adapter. Alias BlockId, IdentifierId, DeclarationId and EvaluationOrder to the shared types so block slices and result tables need no translated copies. InstructionId stays Adamic's index into its instruction table. Use a compile-time interface assertion.

| Interface methods | Adamic answer |
| --- | --- |
| Entry, Blocks, SetBlocks | Function.Entry and Function.Blocks; replacement changes the existing function |
| Block, Retain | blocksById lookup; delete entries rejected by the keep predicate |
| BlockBound | Maximum of nextBlock and one past every id in blocksById, including manually built graphs |
| Id, Predecessors, SetPredecessors | BasicBlock.Id and its predecessor slice |
| Phis, SetPhis | BasicBlock.Phis, using the module's Phi[Place] |
| Placeholder | Unused because Adamic reports no structural Fallthrough; fail loudly if called |
| EachEdge | Preserve terminal target order: Goto target Real; If Consequent then Alternate Real; MayThrow Next Real then Handler Exceptional; every Choose target Real; Return, Throw and Unreachable have none |
| EndsInReturn | True only for Return |
| InstructionCount | Length of the block's instruction-id slice |
| EachInstructionPlace | Locate the instruction by its block index, visit Uses before Defines by pointer, translate PlaceRole to Use/Define |
| IsContextStore, ContextStoreDefines, Contextual | False: captured bindings are excluded from Adamic's graph |
| SetInstructionOrder | Write the instruction's Order |
| EachTerminalPlace | No visits: terminal operands are read by the preceding instruction |
| SetTerminalOrder | Write Order on all seven terminal variants; unknown variants fail loudly |
| Params | The original mutable Params slice, so construction renames it in place |
| Returns | Nil: return expressions are instruction uses, with no synthetic return slot |
| Declaration, Mint, Named | Existing identifier declaration, NewIdentifier preserving name and declaration, and name != empty string |
| PlaceString | Existing rendering for the requested identifier |
| IdentifierOf, WithIdentifier | Place.Identifier; return a place with that id |

No structural fallthrough is to be invented for If or Choose. Reverse postorder, removal of unreachable blocks, predecessor deduplication and evaluation numbering remain the module's work. Finalize remains a thin sequence of ReversePostorder, MarkPredecessors and MarkEvaluationOrder in that order. Nil-function guards stay at Adamic's entry points where the current API promises them.

Construct calls the module, which reestablishes graph invariants, clears old phis, renames definitions and eliminates redundant phis. Forward EliminateRedundantPhis, VerifySSA and CollectSSAStats too. Alias violation types, kinds and statistics instead of translating their meaning. Keep the independent reaching-definitions test as an independent implementation; do not replace it with the shared verifier.

## Phi representation

Alias Phi to static_single_assignment.Phi[Place]. Its Operands is a sorted PhiOperands[Place] slice, with one entry per predecessor. Use Get/At for lookup and Set/Delete for changes. A range loop reads operand.Place and operand.Predecessor, not a map key/value pair. PhiOperandsInOrder forwards to the module. No second sorting or map conversion belongs in Adamic.

Update infer.go's escaped-value propagation and flow_test.go's phi lookup mechanically. During the staged SSA switch, the still-local ranges.go's phi readers must use the same sorted operands. That temporary reader adaptation does not justify retaining ranges.go after the range switch. Preserve predecessor ordering, phi ids, source declarations and printed graph order; compare those observations before and after, rather than accepting merely valid SSA as equivalent behavior.

## Mutation aliasing adapter and vocabulary

Extend the SSA adapter with the effect table to implement mutation_aliasing.Graph[*Function, *BasicBlock, Place]. Alias AliasingEffectKind, EffectValueKind and AliasingEffect to the module's kinds and AliasingEffect[Place], with constant aliases where existing callers need their names. Replace create, flow and mutate with module CreateEffect, FlowEffect and MutationEffect calls or thin forwarding functions. Their From/Into/HasFrom meaning and effect list order must not change.

| Additional interface method | Adamic answer |
| --- | --- |
| InstructionOrder | Order and true for an existing non-nil instruction; false for a missing instruction |
| TerminalOrder | Existing TerminalOrder result |
| Effects | effects.Get(block.Instructions[index]), preserving effect order; nil table yields no effects |
| ParametersFrozen | False; Adamic is not a React component or hook |
| Context | Nil; captured cells are excluded |
| ReturnValue | Zero Place, false; Returns is nil |
| StoredContextValue | Zero identifier, false; no context stores |
| Closure | Nil captures, false; Adamic's captured cells are not React's frozen captures |

The extra methods are not all no-ops: instruction and terminal orders and Effects must answer from Adamic's graph. The remaining seams return the values above. ContextKinds remains nil.

InferMutableRanges still invokes Adamic's InferAliasingEffects. InferMutableRangesWithEffects passes that table through the adapter to mutation_aliasing.InferMutableRanges with Options{ParametersDefinedOnEntry: true}. This is the explicit seam for Adamic's 4748a636 behavior. Cohere's TestParametersDefinedOnEntryBothWays pins [2,4) without the option and [1,4) with it; a parameter not widened still gets its entry range. Verify that empty/unnumbered graphs and nil input retain Adamic's existing API behavior rather than assuming the module defaults match.

Alias MutableRange, MutableRanges and RangeGap and forward RangeOf, IsMutableAt, ValidateMutableRanges and RangeGaps. The alias graph, deferred mutations, creation sequence indices, immutable propagation, phi refinement, definition walk and half-open range membership come solely from the module. Use its exported Set only in focused table tests, not as a new analysis. The loop-carried inversion remains an unset range with RangeGapLoopCarriedInversion; never clamp it into apparent knowledge.

Adamic's effect inference remains conservative for escaped values, global/captured reads, unknown callees and callback operations. Primitive strings/numbers/booleans stay primitive. Cohere's signature-table inference, React parameter freezing, context-store widening, return-slot aliasing and closure freezing do not move into Adamic. Mutable ranges currently have no production native consumer on this checkout; reuse.go consumes Build, LiveOut and CanThrow directly. Do not change native optimization decisions as part of the migration.

## Callback alias mutation prerequisite

Runtime reports a failing TestEveryMutationIsInItsRange: a callee writes through a callback's alias of a value whose range otherwise ends before the call. The correction belongs to codex/flow-ranges-callback-alias, not this migration. That branch was not published at the time of this inspection, so its exact failing program and fix have not been inspected or verified here. The earlier blanket statement that inference is conservative for callback operations is an intent, not evidence that this hole is covered.

At cohere 7945d102, mutation_aliasing handles the supplied alias-and-mutation effects, but does not discover this callback behavior itself. For a mutable value v assigned or aliased into a, the Effects adapter must preserve the earlier Assign or Alias edge from v into a. A call that may write through a must have a mutation effect at the call's evaluation order c. The module records that mutation, then follows earlier backward alias edges and widens v's End to at least c + 1. It does not require v's old range to include the call before widening it. Transitive mutation also follows captured-value edges backward; MaybeAlias makes that traversal conditional. Edges created at or after the mutation's effect index are excluded. Effects within one instruction must preserve their ordering too.

The existing shared-module TestAMutationReachesWhatItIsAnAliasOf demonstrates a source defined at order 1, an alias at 2 and a mutation through the alias at 3: the source ends with [1,4). TestAnAliasMadeAfterAMutationDoesNotWiden demonstrates that an alias created after the mutation does not widen its source. Both tests passed against 7945d102 with `go test -count=1 ./... -run 'TestAMutationReachesWhatItIsAnAliasOf|TestAnAliasMadeAfterAMutationDoesNotWiden'` in cohere/mutation_aliasing; output is in /tmp/shared-flow-callback-module.log. These are hand-supplied effects, not a proof of Adamic's callback inference.

Adamic's infer.go remains responsible for emitting the call-site mutation on every reachable mutable alias, including values reachable through escaped or captured storage. Its current call/mutateEscaped path enumerates current declaration values; MakeClosure supplies an unknown shape without exposing captured cells in this graph. The shared module receives only those inferred effects through Graph.Effects. Graph.Closure returning nil and false supplies no missing capture relation, and AliasingEffectApply is a range no-op, not an interprocedural callee analysis. If inference omits the mutation or the reachable alias, switching to mutation_aliasing does not fix it. ParametersDefinedOnEntry changes range starts only and cannot repair this missing call-site end.

Before the range switch, obtain the other worker's exact program and correction, identify whether it changes inference or range propagation, and record a green baseline containing that fix. Preserve inference changes in infer.go. If the correction changes the retired alias/range algorithm, check the same case against the shared module; if it fails there, stop and have cohere implement the missing behavior rather than copy the local fix. Require the exact callback witness to pass TestEveryMutationIsInItsRange and the uncached Node oracle before and after the adapter switch, with trace expectations and counts unchanged. A targeted mutant must remove the call-site mutation or its alias relation and make that witness fail; restore it before proceeding. Do not skip this failure, relax range validation or treat the known red baseline as expected refactor behavior.

## Call sites to preserve

| Caller | Migration seam |
| --- | --- |
| build.go: Build | Finalize forwarder |
| flow_test.go: TestEveryFunctionIsInSingleAssignment | Construct, VerifySSA, CollectSSAStats; mechanical sorted-operand lookup in checkReaching |
| trace_test.go: traced | Construct; leave trace generation, runtime, walk and expected paths unchanged |
| infer.go: inference.run | Read sorted phi operands through operand.Place; keep its own effect inference |
| ranges.go before its deletion | Replace phi map readers during SSA staging, then remove with the shared range switch |
| ranges_test.go: TestEveryMutationIsInItsRange | InferMutableRanges and ValidateMutableRanges forwarders; leave test and thresholds unchanged |
| native/reuse.go and native/region.go | Existing Build/LiveOut/CanThrow integration; inspect results without changing emission |
| lower's fresh-write analysis through internal/fresh | Existing graph construction and instruction identity; ensure Build and finalization preserve its proof |

Repeat rg over exported flow APIs, Phi.Operands and retired helper names immediately before deletion. This inventory describes the inspected commits, not future callers.

## Known hole and behavior limits

#2yz9ra9 is excluded from this refactor. An exceptional edge currently looks up the predecessor's final definition, so a failed throwing store can give its handler the new value. Reporting Exceptional does not fix it: valueFromPredecessor is keyed by predecessor id and destination, with no edge kind passed to the lookup. MayThrow with the same Next and Handler has one predecessor/phi operand after deduplication. Preserve this known behavior and add a skipped test naming #2yz9ra9 if none exists. Its eventual edge-sensitive fix belongs in the module, with a separate Node witness.

No changes to trace tests, ranges_test.go, native semantics or counts.md are accepted as an incidental consequence. A changed trace, mutation interval or counts row stops the migration for diagnosis. SSA validity by itself does not prove the same reaching value; retain the independent reaching and Node checks. Existing trace exclusions (nonterminating output, size_class_churn and bitwise_sweep) remain exclusions, explicitly reported, while the native oracle still covers terminating oracle fixtures.

## Switch order and checkpoints

1. First resolve the callback alias mutation prerequisite above through the other worker's fix and verify it through the shared module. Record baseline on the compatible checker/submodule combination: whole internal/flow package, internal/fresh, internal/lower and internal/native; uncached oracle for all fixtures using these shared compiler paths; formatting, vet and unchanged counts. Preserve logs and per-function SSA statistics and range observations. The ca939db5 TypeScript pin differs from Adamic's current pin, so checker compatibility must be green before interpreting a flow failure. If integration has not supplied that compatibility, stop rather than hide failures by repinning the test constant alone.
2. Add both local module replaces and the SSA adapter, aliases and sorted phi representation. Route graph maintenance and all SSA calls to the module, then delete graph.go, ssa.go, ssa_eliminate.go and ssa_verify.go one named file at a time. Keep Adamic's effect inference and local ranges temporarily, adapting only their phi reads. Checkpoint: whole flow tests (SSA, independent reaching, trace, liveness and mutation ranges), fresh/lower/native packages, uncached Node oracle and exact unchanged counts. Mutant: misreport an If target or rewrite a use to a sibling definition; require the corresponding independent path/reaching check to fail, then restore.
3. Switch the shared effects vocabulary and constructors while retaining infer.go's algorithm and the effect table. Checkpoint: compare ordered effects on the same SSA functions, rerun whole flow, fresh/lower/native and uncached oracle; counts remain identical. Mutant: omit an inferred mutation on a mutable input; require the Node mutation-range test to fail, then restore.
4. Extend the adapter for mutation aliasing, select ParametersDefinedOnEntry, forward range APIs, then delete ranges.go and effects.go's shared vocabulary. Checkpoint: whole flow package with ranges_test.go and trace_test.go unchanged, exact ranges and unset-gap counts, fresh/lower/native, full uncached oracle, gofmt and vet. Mutant: turn off ParametersDefinedOnEntry and require the entry-mutation probe to distinguish [1,4) from [2,4); restore the option. No local alias graph or range worklist survives.
5. Recheck all callers and remove retired manifest entries if the guard has landed. Run the final gate uncached with output to a log, inspect any counts regeneration diff (it must be empty), commit and push only codex/shared-ssa. This implementation begins only after the design has been pushed and reported.

Each checkpoint logs actual commands, exit codes, named mutants and what caught them. Tests write to files and are never piped. New executable fixtures use .a. No behavior correction, including #2yz9ra9, is bundled into a deletion-only refactor.
