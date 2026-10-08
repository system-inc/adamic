# File inventory

Pinned cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`. Go has no intra-package import statements. “Package references” below are conservative declaration-name references after comments and literals are stripped; method-name collisions can add edges. This is a review aid, not a promised acyclic file graph. Test associations are symbol references; same-stem tests additionally test the named pass. All test files are inventoried too.

## align_case_test.go — 246 lines

Purpose: alignCase describes a minimal graph exercising one alignment mechanism.
Defines: alignCase, buildAlignCase, buildAlignCaseWithTerminal, buildInnermostGotoCase, buildNestedGotoCase.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, reactive_build.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, testing.

## align_method_calls.go — 178 lines

Purpose: Making a method call and its property agree about scopes.
Defines: AlignMethodCallScopes, rebuildWithAlignedMethodCalls.
Package references: align_scopes.go, dependencies.go, disjoint.go, effects.go, high_level_intermediate_representation.go, instruction.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, scope_terminals.go, scopes.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, sort.
Test references: align_method_calls_test.go.

## align_method_calls_test.go — 173 lines

Purpose: TestAlignMethodCallScopesRemovesTheInBetweenState asserts the pass's whole contract.
Defines: TestAlignMethodCallScopesReachesTheCorpus, TestAlignMethodCallScopesRemovesTheInBetweenState, methodCallScopeStates.
Package references: align_method_calls.go, disjoint.go, drop_manual_memoization.go, high_level_intermediate_representation.go, instruction.go, lower.go, ranges.go, reactive.go, scopes.go, ssa.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## align_scopes.go — 639 lines

Purpose: Aligning reactive scopes to block scopes: widening a scope out to the construct it straddles.
Defines: AlignGap, AlignGaps, AlignReactiveScopesToBlockScopes, AlignThenMergeReactiveScopes, AlignedScopes, Ids, Len, MemberRanges, MembersOf, RangeOf, Widened, activeScopeOf, alignSweepState, assignValueBlockNodes, fallthroughRange, recordPlace, startingIdOf, valueBlockNode, visitBlock.
Package references: dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, ranges.go, reactive.go, reactive_build.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, dependency_oracle_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, lower_for_of_kinds_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_scopes_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, prune_unused_scopes_test.go, ranges_test.go, reactive_build_test.go, reactive_transform_test.go, reactive_visitor_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go.

## align_scopes_test.go — 935 lines

Purpose: tagged is a nesting item plus which list it came from, so a violation can be attributed to a kind
Defines: TestAlignClosesTheFullBlockNestingAssertion, TestAlignDeclinesAPlaceReadOutsideItsScope, TestAlignExcludesBranchTerminals, TestAlignFallthroughsAreUniqueOnceBranchesAreExcluded, TestAlignGapsAreDeclared, TestAlignGotoSkipsTheInnermostOpenConstruct, TestAlignGotoWidensAcrossSeveralOpenConstructs, TestAlignHandlesAnEmptyScopeTable, TestAlignIsASingleSweep, TestAlignIsDeterministic, TestAlignMinimisesWithoutAGuard, TestAlignMustRunBeforeTheMerge, TestAlignPreservesScopeWidthAndMembership, TestAlignPushFilterDeclinesAScopeThatEndedBeforeTheTerminal, TestAlignReDerivesMemberRanges, TestAlignRecordsAScopeOnceCoversTheSeenGate, TestAlignVoidsTheMergesComparatorVerdicts, TestReactReversePostorderClosesFallthroughSelfNesting, alignedRangeMap, mergedRangeMap, scopeRangeMap, tagged, violationsByKind, widenedTableFrom.
Package references: align_scopes.go, dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, reactive_build.go, scope_terminals.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, sort, testing.

## align_survivor_helper_test.go — 81 lines

Purpose: mergeWithReversedEndOrder is the merge sweep with `sortByEndDescending` INVERTED.
Defines: mergeWithReversedEndOrder.
Package references: align_scopes.go, dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, primitive_property_constraints.go, ranges.go, reactive_build.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment, sort.

## always_invalidating.go — 92 lines

Purpose: Which values invalidate on every render, asked of the resident checker rather than inferred.
Defines: IsAlwaysInvalidatingType.
Package references: high_level_intermediate_representation.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.
Test references: always_invalidating_test.go, merge_invalidating_test.go.

## always_invalidating_test.go — 235 lines

Purpose: TestIsAlwaysInvalidatingTypeSeparatesTheFourShapes is the whole claim of this predicate.
Defines: TestIsAlwaysInvalidatingTypeDeclinesWithoutAChecker, TestIsAlwaysInvalidatingTypeHandlesNodelessIdentifiers, TestIsAlwaysInvalidatingTypeSeparatesTheFourShapes, forEachCorpusFunctionWithChecker, invalidatingAnswers.
Package references: always_invalidating.go, graph.go, high_level_intermediate_representation.go, lower.go, primitive_property_constraints.go, ranges.go, ssa.go.
Imports: path/filepath, testing, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## cache.go — 257 lines

Purpose: Per-file memoization of lowering, so the rules that share a function share the work.
Defines: AsCompilationUnit, ForFunction, ForFunctionWithoutManualMemoization, cacheKeyFor, mentionsManualMemoization.
Package references: drop_manual_memoization.go, high_level_intermediate_representation.go, inline_iife.go, lower.go, ranges.go, spelling.go, ssa.go.
Imports: strconv, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule.
Test references: cache_test.go.

## cache_test.go — 371 lines

Purpose: The component below is the shape the cache exists for: a `useMemo`, a state setter, and a JSX
Defines: TestForFunctionConstructsExactlyOnce, TestForFunctionLowersOncePerFunction, TestForFunctionRecordsOneFillPerFunction, TestForFunctionWithoutCacheStillLowers, TestForFunctionWithoutManualMemoizationSeesEscapedSpellings, TestMayHoldComponentOrHookAnswersFromTheText, cacheProbeSource, totalPhis.
Package references: cache.go, drop_manual_memoization.go, graph.go, high_level_intermediate_representation.go, lower.go, ranges.go, reactive_build.go, spelling.go, ssa.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## capture_test.go — 496 lines

Purpose: Tests for closure capture: the values a nested function closes over from an enclosing one.
Defines: TestCaptureAcceptanceCase, TestCaptureDoesNotClaimGlobals, TestCaptureIndicesPairAcrossTheBoundary, TestCaptureKeepsSSAValid, TestCaptureShadowingIsNotCaptured, TestCaptureSymbolIdentityCrossesFunctions, TestCaptureThroughTwoBoundaries, TestCaptureVerifierStillDetectsAViolation, TestCaptureWriteIsNotAGlobalWrite, captureNames, firstFunctionExpression, instructionNames, lowerTypedFunctions, moduleLocal.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, preserve_manual_memoization.go, print.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive_build.go, ssa.go, ssa_verify.go.
Imports: strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## clone.go — 118 lines

Purpose: Copying a finished graph, so a pass that rewrites one can start from the shared lowering.
Defines: CloneFunction.
Package references: graph.go, high_level_intermediate_representation.go, inline_remap.go, ranges.go, reactive_build.go, terminal.go.
Imports: github.com/system-inc/cohere/static_single_assignment, maps, slices.
Test references: clone_test.go.

## clone_test.go — 108 lines

Purpose: TestCloneFunctionSharesNothingAPassWrites runs the whole preserve-manual-memoization pipeline over a
Defines: TestCloneFunctionSharesNothingAPassWrites, forgetIdentifierSlabs.
Package references: clone.go, high_level_intermediate_representation.go, lower.go, preserve_manual_memoization.go, ranges.go, ssa.go.
Imports: reflect, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## computed_member_name_test.go — 109 lines

Purpose: A class member with a computed name lowers rather than crashing the file.
Defines: TestAComputedMemberNameLowersToNoName, TestLoweringAClassMemberWithAComputedName.
Package references: dependencies.go, lower.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/core, github.com/microsoft/TypeScript/tsc/shim/parser.

## context_identifiers.go — 168 lines

Purpose: Which bindings a function shares with the closures inside it, decided before anything is lowered.
Defines: assignmentTarget, contextIdentifiers, findContextIdentifiers, identifierUsage.
Package references: dependencies.go, lower.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker.

## controldominators_test.go — 110 lines

Purpose: ControlDominators finds the branch a block is control-dependent on.
Defines: TestControlDominatorsConsultsTheTest, TestControlDominatorsFindsTheDecidingBranch, TestControlDominatorsIsEmptyWithoutBranches, TestControlDominatorsIsNilSafe.
Package references: graph.go, high_level_intermediate_representation.go, postdominator.go, reactive_build.go.
Imports: github.com/system-inc/cohere/static_single_assignment, sort, testing.

## dead_code_elimination.go — 302 lines

Purpose: Removing instructions nothing reads, which is React's `deadCodeElimination`.
Defines: DeadCodeEliminationResult, EliminateDeadCode, findReferencedIdentifiers, functionHasBackEdge, liveIdentifiers, pruneableValue, reference, usedById, usedByIdOrName.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, reactive_build.go, ssa.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: dependency_oracle_test.go, effects_react_state_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, inline_iife_test.go, inline_remap_test.go, merge_consecutive_blocks_test.go, scope_oracle_test.go.

## dependencies.go — 1906 lines

Purpose: Scope dependencies: what each reactive scope reads from outside itself.
Defines: CollectScopeDependencies, CollectScopeDependenciesWithHoistable, Conflicts, DeclarationsOf, DependenciesOf, DependencyGap, DependencyGapOptionalChains, DependencyGapTypeExclusions, DependencyGaps, DependencyPathEntry, Ids, Len, NormalizeInferredDependency, OriginOf, PruneDeclarationsLastUsedBefore, PruneNonReactiveDependenciesOf, ReactiveScopeDependency, ReassignmentsOf, ScopeDependencies, addDependency, alreadyPresent, appendWithoutOptionalDuplicates, blocksAlwaysReached, checkValidDependency, collectMinimalInSubtree, collectTemporaries, collectTemporariesInto, containsIdentifier, copyScopeStack, currentScope, declaration, declare, dependencyCollector, dependencyNode, dependencyTree, deriveMinimalDependencies, enterScope, equalPaths, exitScope, findTemporariesUsedOutsideDeclaringScope, getProperty, handleInstruction, hoistableNode, isDeferredDependency, isDependencyAccess, isOptionalAccess, makeOrMergeProperty, mergeAccess, nestedAccesses, nestedAccessesByBlock, nestedOptionalInstructionsToDefer, newDependencyNode, newDependencyTree, objectMethodValues, optionalAccess, optionalHoistable, optionalProcessedInstructions, optionalProcessedTests, prefixesDisagree, propertyAccessType, recordPrefixOptionality, reduce, resolve, resolveWithPath, scopeBlockInfo, scopeBlockTraversal, scopeIsActive, temporaries, unconditionalAccess, unconditionalDependency, visitDependency, visitNestedFunction, visitReassignment, walk.
Package references: align_scopes.go, disjoint.go, drop_manual_memoization.go, effects.go, graph.go, high_level_intermediate_representation.go, hoistable.go, instruction.go, invoked_functions.go, lower.go, manual_memo_comparison.go, memoization_graph.go, memoization_level.go, merge_invalidating.go, merge_scopes.go, optional_chains.go, postdominator.go, preserve_manual_memoization.go, primitive_property_constraints.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive_build.go, scope_terminals.go, scopes.go, ssa.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, strconv, strings.
Test references: align_scopes_test.go, align_survivor_helper_test.go, capture_test.go, computed_member_name_test.go, dependencies_test.go, dependency_oracle_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_react_state_test.go, effects_test.go, high_level_intermediate_representation_test.go, hoistable_optional_test.go, hoistable_test.go, inline_remap_test.go, lower_test.go, manual_memo_comparison_test.go, memo_block_scope_oracle_test.go, memoization_graph_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, merge_scopes_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, prune_unused_scopes_test.go, ranges_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go, ssa_test.go.

## dependencies_nested_invocation_test.go — 31 lines

Purpose: Tests of the declarations listed below.
Defines: TestNestedDependencyHoistingRequiresAssumedInvocation, callback.
Package references: none.
Imports: strings, testing.

## dependencies_ref_current_test.go — 103 lines

Purpose: TestRefCurrentReadsTruncateToTheirRoot pins that `.current` is not a dependency.
Defines: TestRefCurrentReadsTruncateToTheirRoot.
Package references: none.
Imports: testing.

## dependencies_test.go — 810 lines

Purpose: dependenciesFor lowers one source and runs the whole scope pipeline over it.
Defines: TestDeclarationOriginDiffersFromHoldingScope, TestDependenciesExcludeValuesTheScopeItselfProduces, TestDependenciesFindTheRootOfAnAccessPath, TestDependenciesHandleANilFunction, TestDependenciesRequireScopeTerminals, TestDependencyCollectionIsASinglePass, TestDependencyCollectionIsDeterministic, TestDependencyDistributionIsReal, TestDependencyGapsAreDeclared, TestDependencyPathsCompareOptionality, TestDependencyTreeReducesToTheShallowestPath, TestDependencyTreeTruncatesWithoutAHoistableSet, TestMergeAccessIsAUnionOnDependencyAndAnIntersectionOnOptionality, TestOptionalChainProcessedStoreDoesNotLeakPrefix, TestOptionalChainProcessedStoresAreDeferred, TestPruneDeclarationsLastUsedBefore, TestScopesCloseAtTheirFallthrough, containsName, dependenciesFor, dependencyRootNames.
Package references: align_scopes.go, dependencies.go, disjoint.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, optional_chains.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive_build.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: os, path/filepath, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## dependency_oracle_test.go — 519 lines

Purpose: The compiled output in a golden names upstream's inferred dependencies, verbatim.
Defines: TestInferredDependenciesAgainstGoldenCacheSlots, goldenDependencies, identifierName, inferredDependencyStrings, isCodegenTemporary.
Package references: align_scopes.go, dead_code_elimination.go, dependencies.go, disjoint.go, drop_manual_memoization.go, high_level_intermediate_representation.go, inline_iife.go, lower.go, merge_consecutive_blocks.go, merge_scopes.go, outline_functions.go, ranges.go, reactive.go, scope_terminals.go, scopes.go, ssa.go.
Imports: os, regexp, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/rules/react/conformance, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## disjoint.go — 662 lines

Purpose: Disjoint mutable values: which values are so entangled by mutation that they must be treated as
Defines: DisjointGap, DisjointGapPrimitiveCallResult, DisjointGaps, DisjointSet, Find, FindDisjointMutableValues, FindDisjointMutableValuesWithRanges, Has, RepresentativeOf, Sets, Size, Union, appendMutableOperand, appendMutableOperands, appendStore, callResultIsPrimitive, declarationOf, mayAllocate, patternContainsSpread, taggedTemplateCalleeSyntaxName.
Package references: dependencies.go, effects.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, pattern.go, primitive_property_constraints.go, ranges.go, reactive_build.go, ssa.go, visitor.go.
Imports: sort, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: align_method_calls_test.go, dependencies_test.go, dependency_oracle_test.go, disjoint_test.go, effects_react_state_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_scopes_corpus_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, prune_unused_scopes_test.go, ranges_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go.

## disjoint_test.go — 1238 lines

Purpose: disjointFor lowers one source and returns the outermost function with its classes.
Defines: TestDisjointAllocationGetsItsOwnClass, TestDisjointCallAllocationUsesTheEffectSignature, TestDisjointClassSizesAreNotAllSingletons, TestDisjointDeclarationKeepsTheFirstValue, TestDisjointDestructureAllocatesOnlyWithSpread, TestDisjointFindTerminatesOnALongChain, TestDisjointGapsAreDeclared, TestDisjointHandlesANilFunction, TestDisjointIsDeterministic, TestDisjointIsIdempotent, TestDisjointLeavesIndependentValuesApart, TestDisjointMayAllocateMatchesReact, TestDisjointMethodCallAddsThePropertyUnconditionally, TestDisjointOperandGateNeedsBothHalves, TestDisjointPhiUnionIncludesTheDeclaration, TestDisjointPhiUsesReactsTwoTermTest, TestDisjointPrimitiveDoesNotAllocate, TestDisjointSetAbsentValueIsNotClassZero, TestDisjointSetDoesNotMutateTheCallersSlice, TestDisjointSetEmptyUnionIsANoOp, TestDisjointSetMatchesReact, TestDisjointUnifiesALoopCarriedPhi, TestDisjointUnifiesAMutatedObjectWithItsAlias, classOfName, disjointFor.
Package references: dependencies.go, disjoint.go, effects.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, merge_scopes.go, pattern.go, primitive_property_constraints.go, ranges.go, reactive.go, reactive_build.go, ssa.go.
Imports: os, path/filepath, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## drop_manual_memoization.go — 711 lines

Purpose: Dropping manual memoization: recognising `useMemo`/`useCallback` and recording what the developer
Defines: DropManualMemoization, ManualMemoKind, ManualMemoKindNone, ManualMemoKindUseMemo, ManualMemoization, String, calleeAnchor, collectManualMemoPhiDependencies, collectManualMemoTemporaries, extractManualMemoArguments, extractedMemoArguments, findManualMemoOptionalPlaces, insertQueuedMarkers, isNamedBinding, manualMemoSidemap, newMarkerInstruction, propagateManualMemoDependency, queuedMarkerInsert, recogniseManualMemoCall.
Package references: dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, manual_memo_comparison.go, memoization_level.go, optional_chains.go, preserve_manual_memoization.go, ranges.go, reactive_build.go, ssa.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_method_calls_test.go, cache_test.go, dependency_oracle_test.go, drop_manual_memoization_deps_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, identifier_slab_test.go, inline_iife_test.go, lower_corpus_test.go, manual_memo_comparison_test.go, memo_block_scope_oracle_test.go, memoization_level_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, scope_oracle_test.go.

## drop_manual_memoization_deps_test.go — 235 lines

Purpose: TestUnextractableDependencyAbandonsTheWholeList pins that a list we cannot read is not an empty
Defines: TestDropManualMemoizationDoesNotProjectOptionalThroughTernaryPhi, TestDropManualMemoizationRecoversOptionalGlobalDependency, TestDropManualMemoizationRecoversOptionalPlacesFromControlFlow, TestUnextractableDependencyAbandonsTheWholeList, memoDepsFor.
Package references: dependencies.go, drop_manual_memoization.go, graph.go, instruction.go, lower.go, ranges.go, reactive.go, reactive_build.go, ssa.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## effects.go — 1528 lines

Purpose: Aliasing effects: what a call does to its arguments, and what an instruction does to its operands.
Defines: AliasingEffect, AliasingEffects, EffectGap, EffectGapInterproceduralParameters, EffectGapSignatureTable, EffectGaps, Get, InferAliasingEffects, InferAliasingEffectsForNested, Len, ProjectEffects, String, aliasingOperand, aliasingReceiver, aliasingReturns, aliasingSignatureEffect, argumentMutationsFromCallbacks, calleeName, calleeSyntaxName, eachDestructureBinding, effectSignature, effectsForCall, effectsForInstruction, effectsForUnknownCall, effectsFromAliasingSignature, inferAliasingEffects, lookupSignature, mutatesOwnCapture, mutatesOwnParameter, nestedFunctionHeldBy, preserveExistingMemoizationEnabled, receiverSyntaxName, visitSignatureOperand.
Package references: align_scopes.go, dependencies.go, drop_manual_memoization.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, manual_memo_comparison.go, memoization_graph.go, memoization_level.go, merge_invalidating.go, merge_scopes.go, outline_functions.go, pattern.go, preserve_manual_memoization.go, primitive_property_constraints.go, prune_non_escaping_scopes.go, ranges.go, reactive_build.go, scopes.go.
Imports: strings, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, disjoint_test.go, effects_computed_load_test.go, effects_custom_hooks_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, hoistable_test.go, identifier_slab_test.go, lower_corpus_test.go, lower_for_of_kinds_test.go, manual_memo_comparison_test.go, memoization_level_test.go, merge_invalidating_test.go, merge_scopes_test.go, primitive_property_constraints_test.go, ranges_immutable_edges_test.go, ranges_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go.

## effects_computed_load_test.go — 76 lines

Purpose: Tests of the declarations listed below.
Defines: TestComputedLoadCopiesReceiverKind, TestManualMemoizationPreservesComputedFrozenRead, callback, item, value.
Package references: effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, ranges.go.
Imports: strings, testing, github.com/system-inc/cohere/mutation_aliasing.

## effects_custom_hooks.go — 111 lines

Purpose: calleeProducers is the table isModuleHookCall resolves a callee through: each value's producing
Defines: calleeProducers, customHookSignature, isModuleHookCall, newCalleeProducers, producer.
Package references: dependencies.go, effects.go, high_level_intermediate_representation.go, instruction.go, prune_non_escaping_scopes.go, reactive.go.
Imports: strings, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: disjoint_test.go, effects_custom_hooks_test.go, effects_test.go, prune_always_invalidating_test.go.

## effects_custom_hooks_test.go — 158 lines

Purpose: TestCalleeProducersAreBuiltOncePerFunction guards the quadratic this table replaced: every call
Defines: TestCalleeProducersAreBuiltOncePerFunction, TestCustomHookEffectsFollowModuleBindings, TestCustomHookPropertyAliasPreservesManualMemoization, callback, data, value.
Package references: effects.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, primitive_property_constraints.go, ranges.go, ssa.go.
Imports: strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing.

## effects_react_effect_test.go — 199 lines

Purpose: Tests of the declarations listed below.
Defines: React, TestEffectHookDoesNotPromoteStableSetter, TestEffectHookSignaturesFollowImportOrigin, TestImperativeHandleDoesNotPromoteStableSetter, callback, effectHookDeclarations, items, notifications, propItems, reference.
Package references: effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, preserve_manual_memoization.go, ranges.go, ssa.go.
Imports: fmt, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing.

## effects_react_state_test.go — 227 lines

Purpose: Tests of the declarations listed below.
Defines: React, TestStateEffectsFollowImportOrigin, TestStateMethodCallDoesNotMakeSetterReactive, TestStateSetterDependencyIsPrunedAfterScopeSeparation, canGo, setGlobal, toggle.
Package references: align_scopes.go, dead_code_elimination.go, dependencies.go, disjoint.go, drop_manual_memoization.go, effects.go, flatten_reactive_loops.go, graph.go, high_level_intermediate_representation.go, inline_iife.go, instruction.go, lower.go, merge_consecutive_blocks.go, merge_invalidating.go, merge_scopes.go, outline_functions.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive.go, reactive_build.go, scope_terminals.go, scopes.go, ssa.go.
Imports: strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing.

## effects_test.go — 1056 lines

Purpose: effectsFor lowers one function and returns its effect table plus the function.
Defines: TestAnUnknownMethodCallDoesNotMutateTheMethodItself, TestAwaitConditionallyMutatesTheAwaitedValue, TestCaptureReceiverIsAliasedExactlyOnce, TestDestructureEffectsNameEveryBinding, TestEffectGapsAreNamedInTheApi, TestEffectsAreDeterministic, TestEveryInstructionShapeProducesEffects, TestFreezeFollowsOnlySharedAssignIdentity, TestFreezeTraversesPhiValuesAndFunctionCaptures, TestInterproceduralParametersAreNotInferred, TestJsxFreezesWhatItInterpolates, TestKnownMutatingSignatureMutatesItsReceiver, TestKnownReadOnlySignatureDoesNotMutate, TestManualMemoMarkersFreezeTheirOperands, TestMemoFreezeRefinesAnUnknownMethodMutation, TestProjectEffectsNeedsRangesToDistinguishCaptureFromRead, TestPropertyStoreCapturesTheStoredValue, TestPropertyStoreMutatesItsObject, TestSortIsSilentBecauseUpstreamHasNoSignatureForIt, TestStartMemoizeFreezeRequiresAnInitializedAbstractValue, TestUnknownCallConservativelyMutatesItsArguments, effectsFor, effectsOn, hasEffect, typeName.
Package references: align_scopes.go, dependencies.go, drop_manual_memoization.go, effects.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, manual_memo_comparison.go, memoization_graph.go, memoization_level.go, merge_invalidating.go, merge_scopes.go, preserve_manual_memoization.go, primitive_property_constraints.go, ranges.go, reactive_build.go, scopes.go, ssa.go, terminal.go, visitor.go.
Imports: sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## flatten_reactive_loops.go — 110 lines

Purpose: Flattening reactive loops: a scope inside a loop is pruned, deliberately, before anything reads it.
Defines: FlattenReactiveLoops.
Package references: graph.go, high_level_intermediate_representation.go, reactive_build.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: effects_react_state_test.go, flatten_scopes_with_hooks_test.go.

## flatten_scopes_with_hooks.go — 124 lines

Purpose: Flattening the scopes a hook call sits in: a hook cannot be called conditionally, so no scope around
Defines: FlattenScopesWithHooksOrUse, callsHookOrUse, isReactUse.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, reactive.go, reactive_build.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: flatten_scopes_with_hooks_test.go.

## flatten_scopes_with_hooks_test.go — 121 lines

Purpose: hookFlattening is what FlattenScopesWithHooksOrUse did to one function.
Defines: TestFlattenScopesWithHooksOrUseLabelsALoneHookCall, TestFlattenScopesWithHooksOrUsePrunesAScopeHoldingMore, flattenHooksIn, hookFlattening.
Package references: align_scopes.go, dead_code_elimination.go, disjoint.go, drop_manual_memoization.go, flatten_reactive_loops.go, flatten_scopes_with_hooks.go, graph.go, inline_iife.go, lower.go, merge_consecutive_blocks.go, outline_functions.go, ranges.go, reactive.go, reactive_build.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## graph.go — 285 lines

Purpose: Graph maintenance: the passes lowering runs to bring a freshly built function into the state
Defines: Block, BlockBound, Blocks, ContextStoreDefines, Contextual, Declaration, EachEdge, EachInstructionPlace, EachTerminalPlace, EndsInReturn, Entry, Finalize, Id, IdentifierOf, InstructionCount, IsContextStore, MarkEvaluationOrder, MarkPredecessors, Mint, Named, Params, Phis, PlaceString, Placeholder, Predecessors, Retain, Returns, ReversePostorder, SetBlocks, SetInstructionOrder, SetPhis, SetPredecessors, SetTerminalOrder, WithIdentifier, setTerminalOrder, ssaGraph, ssaRole.
Package references: dependencies.go, high_level_intermediate_representation.go, instruction.go, primitive_property_constraints.go, reactive_build.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_case_test.go, align_scopes_test.go, align_survivor_helper_test.go, always_invalidating_test.go, cache_test.go, capture_test.go, controldominators_test.go, dependencies_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_computed_load_test.go, effects_custom_hooks_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, graph_test.go, high_level_intermediate_representation_test.go, identifier_slab_test.go, inline_branch_fallthrough_test.go, inline_iife_test.go, inline_remap_test.go, lower_corpus_test.go, lower_for_of_kinds_test.go, lower_test.go, memo_block_scope_oracle_test.go, merge_consecutive_blocks_test.go, merge_invalidating_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, postdominator_frontier_test.go, postdominator_test.go, primitive_property_constraints_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, ranges_immutable_edges_test.go, ranges_maybe_frozen_test.go, ranges_test.go, reactive_build_test.go, reactive_capture_identity_test.go, reactive_stable_hooks_test.go, reactive_test.go, reactive_visitor_test.go, scope_terminals_test.go, ssa_corpus_test.go, ssa_test.go.

## graph_test.go — 52 lines

Purpose: TestReversePostorderKeepsLoopBodyBeforeContinuation pins the ordering that mutable ranges read.
Defines: TestReversePostorderKeepsLoopBodyBeforeContinuation.
Package references: graph.go, instruction.go, print.go, reactive_build.go.
Imports: github.com/system-inc/cohere/static_single_assignment, testing.

## high_level_intermediate_representation.go — 644 lines

Purpose: Package hir is a high-level intermediate representation of one JavaScript or TypeScript
Defines: AddInstruction, BasicBlock, Block, BlockKind, BlockKindBlock, BlockKindSequence, BlockKindValue, Effect, EffectCapture, EffectConditionallyMutate, EffectFreeze, EffectStore, EffectUnknown, Function, FunctionId, FunctionKind, FunctionKindComponent, FunctionKindOther, HasBlock, Identifier, IdentifierOf, Instruction, InstructionId, InstructionKind, InstructionKindCatch, InstructionKindConst, InstructionKindFunction, InstructionKindHoistedLet, InstructionKindLet, InvalidBlock, IsMutable, NewBlock, NewFunction, NewIdentifier, Phi, PhiOperand, PhiOperands, Place, PlaceString, String.
Package references: dependencies.go, drop_manual_memoization.go, effects.go, graph.go, instruction.go, manual_memo_comparison.go, memoization_level.go, preserve_manual_memoization.go, ranges.go, reactive_build.go, terminal.go.
Imports: fmt, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/core, github.com/system-inc/cohere/static_single_assignment.
Test references: align_case_test.go, align_method_calls_test.go, align_scopes_test.go, align_survivor_helper_test.go, always_invalidating_test.go, cache_test.go, capture_test.go, clone_test.go, controldominators_test.go, dependencies_test.go, dependency_oracle_test.go, disjoint_test.go, effects_computed_load_test.go, effects_custom_hooks_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, high_level_intermediate_representation_test.go, hoistable_optional_test.go, hoistable_test.go, identifier_slab_test.go, inline_branch_fallthrough_test.go, inline_iife_test.go, inline_remap_places_test.go, inline_remap_test.go, invoked_functions_test.go, lower_corpus_test.go, lower_for_of_kinds_test.go, lower_test.go, manual_memo_comparison_test.go, memo_block_scope_oracle_test.go, memoization_level_test.go, merge_consecutive_blocks_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, outline_functions_test.go, postdominator_frontier_test.go, postdominator_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_score_test.go, preserve_manual_memoization_test.go, primitive_property_constraints_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, prune_unused_scopes_test.go, ranges_context_closures_test.go, ranges_maybe_frozen_test.go, ranges_test.go, reactive_build_test.go, reactive_capture_identity_test.go, reactive_stable_hooks_test.go, reactive_test.go, reactive_transform_test.go, reactive_visitor_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go, ssa_corpus_test.go, ssa_test.go.

## high_level_intermediate_representation_test.go — 418 lines

Purpose: TestEachPlaceCoversEveryInstructionValue is the exhaustiveness check Go's type switch cannot do.
Defines: TestClassifyFunction, TestEachBlockReferencePointerCoversEveryTerminal, TestEachBlockReferencePointerSeesEveryReadOnlyReference, TestEachPlaceCoversEveryInstructionValue, TestEachSuccessorCoversEveryTerminal, TestEffectLatticeOrdering, TestInstructionSetSizeIsWhatWeSaidItIs, TestPlaceStringIsStableAcrossIdenticalValues, TestPrintTerminalCoversEveryTerminal, TestPrintValueCoversEveryInstructionValue, TestSetTerminalOrderCoversEveryTerminal, TestTerminalOrderCoversEveryTerminal, contains, everyTerminalSample, forEachPackageFile, markerImplementers, typeSwitchCases.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, lower.go, primitive_property_constraints.go, reactive_build.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: github.com/system-inc/cohere/static_single_assignment, go/ast, go/parser, go/token, os, sort, testing.

## hoistable.go — 1215 lines

Purpose: Hoistable property loads: which accesses may be read before their scope runs.
Defines: CollectHoistablePropertyLoads, Converged, HoistableAnalysis, Iterations, analyseHoistableLoads, blockById, collectNonNullsInBlocks, dependencyArrayInstructions, dominatedByOtherReactiveScope, hoistableAt, hoistableIterationCap, hoistableTreeFor, identifierNode, invokedNonNullPaths, isImmutableAtInstruction, isKnownImmutableParameter, isLikelyComponentFunction, maybeNonNullInInstruction, newPathRegistry, pathIndex, pathNode, pathRegistry, propagateBackward, propagateForward, propagateNonNull, propagationDirection, propertyNode, recursivelyPropagateNonNull, reduceOptionalChains, sameNodeSet, sortedNodeIndices, traversalActive, traversalDone, traversalState.
Package references: align_scopes.go, dependencies.go, effects.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, invoked_functions.go, merge_scopes.go, preserve_manual_memoization.go, primitive_property_constraints.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive.go, reactive_build.go, scope_terminals.go, scopes.go, ssa_verify.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, sort.
Test references: hoistable_optional_test.go, hoistable_test.go.

## hoistable_branch_scope_test.go — 70 lines

Purpose: Tests of the declarations listed below.
Defines: TestBranchLocalHoistableFactsRespectDominatingReactiveScope.
Package references: none.
Imports: fmt, testing.

## hoistable_manual_memo_test.go — 40 lines

Purpose: Tests of the declarations listed below.
Defines: TestManualMemoDependenciesProveOnlyNonOptionalPrefixes, result.
Package references: none.
Imports: strings, testing.

## hoistable_optional_test.go — 111 lines

Purpose: TestOptionalLoadsProveNothingAboutTheirObject pins that `a?.b` does not make `a` hoistable.
Defines: TestOptionalHoistableIsNotRecorded, TestOptionalLoadsProveNothingAboutTheirObject.
Package references: dependencies.go, high_level_intermediate_representation.go, hoistable.go, instruction.go, terminal.go.
Imports: testing.

## hoistable_test.go — 794 lines

Purpose: hoistableFor runs the whole pipeline and returns the analysis.
Defines: TestATopLevelScopeInsideABranchUsesItsFacts, TestForwardPropagationIntersectsRatherThanUnions, TestHoistableAnalysisDeepensDependencyPaths, TestHoistableConflictsAreReportedNotSwallowed, TestHoistableConvergenceIsFarInsideTheBound, TestHoistableCorpusDistribution, TestHoistableHandlesNilInputs, TestHoistablePathsAreAlwaysPrefixesOfRealAccesses, TestHoistableReportsRatherThanHidingNonConvergence, TestOptionalChainReductionTerminates, hoistableFor.
Package references: align_scopes.go, dead_code_elimination.go, dependencies.go, disjoint.go, drop_manual_memoization.go, effects.go, high_level_intermediate_representation.go, hoistable.go, inline_iife.go, instruction.go, lower.go, memoization_graph.go, merge_consecutive_blocks.go, merge_invalidating.go, merge_scopes.go, outline_functions.go, primitive_property_constraints.go, ranges.go, reactive.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: path/filepath, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing.

## identifier_slab_test.go — 77 lines

Purpose: TestIdentifiersShareNoStorageAcrossFunctions holds what makes carving identifiers from chunks safe.
Defines: TestIdentifiersShareNoStorageAcrossFunctions.
Package references: drop_manual_memoization.go, effects.go, graph.go, high_level_intermediate_representation.go, lower.go, manual_memo_comparison.go, memoization_level.go, preserve_manual_memoization.go, ranges.go, ssa.go.
Imports: strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## inline_branch_fallthrough_test.go — 80 lines

Purpose: Tests of the declarations listed below.
Defines: TestConditionalManualMemosKeepSeparateDependencies, TestCopyNestedBodyRemapsBranchFallthrough, callback, firstSet, mounted, secondSet.
Package references: graph.go, high_level_intermediate_representation.go, inline_remap.go, reactive_build.go, terminal.go, visitor.go.
Imports: strings, testing.

## inline_iife.go — 484 lines

Purpose: Splicing an immediately-invoked function expression into its caller.
Defines: InlineImmediatelyInvokedFunctionExpressions, InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks, declareInlineResult, dropSplicedFunctionExpressions, forgetFunctionOperands, inlineInvokedFunctions, isStatementBlockKind, memoizedResults, nestedFunctionOf, returnedPlace, rewriteCopiedReturns, singleReturnExit, spliceInlinedCall.
Package references: graph.go, high_level_intermediate_representation.go, inline_remap.go, instruction.go, reactive_build.go, terminal.go, visitor.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/static_single_assignment.
Test references: dependency_oracle_test.go, effects_react_state_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, inline_iife_test.go, scope_oracle_test.go.

## inline_iife_test.go — 739 lines

Purpose: Tests for the CFG splice.
Defines: TestInlineDeclinesADroppedMemoCallback, TestInlineDeclinesTheCasesItCannotExpress, TestInlineDefinesTheCallResultOnEveryPathToTheContinuation, TestInlineFindsAMemoCallbackAcrossBlocks, TestInlineIsIdempotent, TestInlineLeavesEveryBlockReferenceResolvable, TestInlineRemovesTheCallAndEveryReturnOfTheInlinedBody, TestInlineStoresTheValueEachReturnProduces, TestMemoInclusiveInliningCallSitesAreReviewed, inlinedFixture, multipleReturnIifeSource, singleReturnIifeSource.
Package references: dead_code_elimination.go, drop_manual_memoization.go, graph.go, high_level_intermediate_representation.go, inline_iife.go, instruction.go, lower.go, preserve_manual_memoization.go, ranges.go, reactive_build.go, ssa.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: os, regexp, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## inline_remap.go — 373 lines

Purpose: Copying a nested function's body into its parent's id space.
Defines: CopyNestedBodyInto, InlineRemap, copyInstructionValue, copyTerminal, deepCopyAny, deepCopyValue, remapNestedFunctionIds.
Package references: align_scopes.go, dead_code_elimination.go, dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, ranges.go, reactive_build.go, scopes.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: github.com/system-inc/cohere/static_single_assignment, reflect.
Test references: inline_branch_fallthrough_test.go, inline_remap_places_test.go, inline_remap_test.go.

## inline_remap_places_test.go — 107 lines

Purpose: TestCopyNestedBodyLeavesTheNestedFunctionsPlacesUntouched is the sibling of
Defines: TestCopyNestedBodyLeavesTheNestedFunctionsPlacesUntouched.
Package references: high_level_intermediate_representation.go, inline_remap.go, visitor.go, visitor_mutate.go.
Imports: github.com/system-inc/cohere/static_single_assignment, testing.

## inline_remap_test.go — 501 lines

Purpose: branchingParentSource is the fixture both copy tests lower.
Defines: TestCopyNestedBodyDeclinesOnAContextMismatch, TestCopyNestedBodyKeepsCapturesPointingAtTheParent, TestCopyNestedBodyLeavesNoUnresolvedReference, TestCopyNestedBodyLeavesTheNestedFunctionUntouched, TestCopyNestedBodyMapsEveryPlaceOnce, TestCopyNestedBodyPreservesDeclarationEquivalenceClasses, branchingParentSource, loweredParentAndNested, overlappingIdsParentSource, sortedBlockIds.
Package references: dead_code_elimination.go, dependencies.go, graph.go, high_level_intermediate_representation.go, inline_remap.go, instruction.go, lower.go, ranges.go, reactive_build.go, ssa.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## instruction.go — 511 lines

Purpose: The instruction set.
Defines: Argument, ArrayElement, ArrayExpression, Await, BinaryExpression, CallExpression, ComputedDelete, ComputedLoad, ComputedStore, Debugger, DeclareContext, DeclareLocal, Destructure, FinishMemoize, FunctionExpression, GetIterator, GlobalBindingKind, GlobalBindingKindGlobal, GlobalBindingKindImportNamespace, GlobalBindingKindModuleLocal, InstructionValue, IteratorNext, JsxAttribute, JsxExpression, JsxFragment, JsxTag, JsxText, LoadContext, LoadGlobal, LoadLocal, ManualMemoDependency, ManualMemoRoot, MetaProperty, MethodCall, ModuleExportOrigin, NewExpression, NextPropertyOf, ObjectExpression, ObjectMethod, ObjectProperty, PostfixUpdate, PrefixUpdate, Primitive, PropertyDelete, PropertyLoad, PropertyStore, RegExpLiteral, StartMemoize, StoreContext, StoreGlobal, StoreLocal, TaggedTemplateExpression, TemplateLiteral, TypeCastExpression, UnaryExpression, UnsupportedNode, instructionValue.
Package references: dependencies.go, high_level_intermediate_representation.go, pattern.go, terminal.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/ast.
Test references: align_case_test.go, align_method_calls_test.go, capture_test.go, dependencies_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_computed_load_test.go, effects_custom_hooks_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, graph_test.go, hoistable_optional_test.go, hoistable_test.go, inline_iife_test.go, inline_remap_test.go, lower_corpus_test.go, lower_for_of_kinds_test.go, lower_global_test.go, lower_test.go, manual_memo_comparison_test.go, memo_block_scope_oracle_test.go, memoization_inputs_test.go, merge_nested_scopes_test.go, merge_scopes_test.go, postdominator_test.go, preserve_manual_memoization_test.go, primitive_property_constraints_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, ranges_context_closures_test.go, ranges_immutable_edges_test.go, ranges_test.go, reactive_build_test.go, reactive_capture_identity_test.go, reactive_namespace_refs_test.go, ssa_test.go.

## invoked_functions.go — 234 lines

Purpose: Which nested functions are assumed to be invoked before their scope ends.
Defines: AssumedInvokedFunctions, CollectAssumedInvokedFunctions, collectInvocations, collectInvokedCandidates, invokedCandidate, invokedWithinNested.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, ranges.go, reactive.go, reactive_build.go, terminal.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: invoked_functions_test.go.

## invoked_functions_test.go — 150 lines

Purpose: invokedFor lowers a source and returns which nested functions are assumed invoked.
Defines: TestAssumedInvokedFunctionsRecognisesEachCallShape, invokedFor.
Package references: high_level_intermediate_representation.go, invoked_functions.go, lower.go, ranges.go, ssa.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## lower.go — 1381 lines

Purpose: Lowering: TypeScript AST to HIR.
Defines: Lower, bind, builder, captureOf, capturedSymbolsInOrder, classifyFunction, declarationOf, emit, emitTo, ensureBlock, enter, functionBody, functionIsGenerator, functionName, functionParameters, gotoBlock, hasModifier, isLoopStatement, jumpTarget, lookupBreak, lookupContinue, lowerBreakStatement, lowerContinueStatement, lowerDoStatement, lowerForBinding, lowerForInStatement, lowerForOfStatement, lowerForStatement, lowerFunctionDeclaration, lowerIfStatement, lowerLabeledStatement, lowerNested, lowerNestedFunction, lowerParams, lowerReturnStatement, lowerStatement, lowerStatements, lowerSwitchStatement, lowerTryStatement, lowerVariableDeclaration, lowerVariableDeclarationList, lowerWhileStatement, newTemporary, newTemporaryUnder, popJump, pushJump, rangeOf, reserve, symbolOf, terminateAndEnter, terminateWith.
Package references: context_identifiers.go, dependencies.go, disjoint.go, drop_manual_memoization.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower_expression.go, manual_memo_comparison.go, memoization_level.go, pattern.go, preserve_manual_memoization.go, primitive_property_constraints.go, ranges.go, reactive_build.go, terminal.go, visitor.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/microsoft/TypeScript/tsc/shim/core, github.com/system-inc/cohere/internal/lint/ecmascript/property, github.com/system-inc/cohere/static_single_assignment.
Test references: align_case_test.go, align_method_calls_test.go, align_survivor_helper_test.go, always_invalidating_test.go, cache_test.go, capture_test.go, clone_test.go, computed_member_name_test.go, dependencies_test.go, dependency_oracle_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_custom_hooks_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, high_level_intermediate_representation_test.go, hoistable_test.go, identifier_slab_test.go, inline_iife_test.go, inline_remap_test.go, invoked_functions_test.go, lower_corpus_test.go, lower_global_test.go, lower_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, postdominator_frontier_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_score_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, ranges_immutable_edges_test.go, ranges_test.go, reactive_namespace_refs_test.go, reactive_stable_hooks_test.go, reactive_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go, ssa_corpus_test.go, ssa_test.go.

## lower_corpus_test.go — 372 lines

Purpose: corpusRoot is a real TypeScript codebase to lower against.
Defines: TestLowerRealCodebase, TestLowerRealCodebaseIsDeterministic, corpusRoot, fastTierVariable, forEachFunctionLike, itoa, pinnedCorpusCommit, pinnedCorpusFiles, pinnedCorpusRead, readPinnedCorpus, skipCorpusWalkInFastTier, skipWithoutCorpus.
Package references: drop_manual_memoization.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, manual_memo_comparison.go, memoization_level.go, preserve_manual_memoization.go, primitive_property_constraints.go, print.go, ranges.go.
Imports: bufio, bytes, fmt, io, os, os/exec, path/filepath, sort, strconv, strings, sync, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/core, github.com/microsoft/TypeScript/tsc/shim/parser, github.com/microsoft/TypeScript/tsc/shim/tspath, github.com/system-inc/cohere/internal/corpus.

## lower_expression.go — 1150 lines

Purpose: Expression lowering.
Defines: calleeModuleOrigin, compoundOperatorText, isCompoundAssignment, jsxOpeningParts, lowerArguments, lowerArrayAssignmentPattern, lowerArrayBindingPattern, lowerArrayLiteral, lowerAssignmentTarget, lowerBinaryExpression, lowerCallExpression, lowerCompoundAssignment, lowerConditionalExpression, lowerDeleteExpression, lowerElementAccess, lowerExpressionToPlace, lowerFunctionExpression, lowerIdentifier, lowerJsxAttributes, lowerJsxChildren, lowerJsxElement, lowerJsxFragment, lowerJsxTag, lowerLogicalExpression, lowerNewExpression, lowerObjectAssignmentPattern, lowerObjectBindingPattern, lowerObjectLiteral, lowerPattern, lowerPostfixUnary, lowerPrefixUnary, lowerPropertyAccess, lowerTaggedTemplate, lowerTemplateExpression, objectPropertyKey, operatorText, splitRegExp.
Package references: drop_manual_memoization.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, lower_global.go, lower_optional.go, manual_memo_comparison.go, memoization_level.go, pattern.go, preserve_manual_memoization.go, primitive_property_constraints.go, ranges.go, terminal.go, visitor.go.
Imports: strings, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/static_single_assignment.

## lower_for_of_kinds_test.go — 155 lines

Purpose: Tests of the declarations listed below.
Defines: TestForOfHeaderBlockKinds, TestForOfHeaderScopeAlignment, TestForOfScopedHeadersPreserveInstructions.
Package references: align_scopes.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, merge_scopes.go, ranges.go, reactive_build.go, scope_terminals.go, scopes.go, ssa.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, testing.

## lower_global.go — 25 lines

Purpose: globalBindingKinds maps how rule.ImportBindingOf found a name bound to the lowering's own kinds.
Defines: moduleGlobalLoad.
Package references: instruction.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule.

## lower_global_test.go — 66 lines

Purpose: Tests of the declarations listed below.
Defines: TestGlobalLoadsRetainBindingProvenance.
Package references: instruction.go, lower.go, ranges.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## lower_optional.go — 164 lines

Purpose: Optional chains: `a?.b`, lowered to the block shape upstream produces.
Defines: isOptionalChainLink, lowerOptionalChain, optionalChainParts.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, lower_expression.go, terminal.go, visitor.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/ast.

## lower_test.go — 1054 lines

Purpose: lowerSource parses one function out of source and lowers it.
Defines: TestEachSuccessorAndFallthroughReachesTheJoin, TestFallthroughIsNotAnEdge, TestLowerAsyncAndGeneratorModifiers, TestLowerDestructuring, TestLowerForOfProducesIteratorProtocol, TestLowerGlobalStoreIsOneInstruction, TestLowerJsx, TestLowerJsxComponentTagIsAValueReference, TestLowerLabeledBlockBreak, TestLowerLabeledBreakAndContinue, TestLowerLogicalIsControlFlow, TestLowerLoop, TestLowerMethodCallKeepsReceiver, TestLowerNestedFunction, TestLowerOptionalChaining, TestLowerReturnsIsOneIdentifier, TestLowerShorthandPropertyReadsLocalBinding, TestLowerSwitch, TestLowerSwitchWithoutDefaultCanSkipEveryCase, TestLowerTernary, TestLowerTryFinally, TestLowerTryFinallyWithoutCatch, TestLowerUnreachableCodeIsDropped, TestLowerUnsupportedSyntaxDoesNotFail, TestLowerWhileAndDoWhile, checkInvariants, lowerSource, requireInstruction.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, optional_chains.go, pattern.go, primitive_property_constraints.go, print.go, ranges.go, reactive_build.go, terminal.go, visitor.go.
Imports: strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/core, github.com/microsoft/TypeScript/tsc/shim/parser, github.com/system-inc/cohere/static_single_assignment.

## manual_memo_comparison.go — 152 lines

Purpose: Comparing the dependencies a developer wrote against the ones the compiler inferred.
Defines: CompareDependencyOk, CompareDependencyResult, CompareDependencyRootDifference, CompareDependencySubpath, CompareManualMemoDependencies, MergeCompareDependencyResults, String, manualMemoRootsEqual, pathReadsRefCurrent.
Package references: dependencies.go, drop_manual_memoization.go, effects.go, high_level_intermediate_representation.go, instruction.go, memoization_level.go, preserve_manual_memoization.go, ranges.go, terminal.go.
Imports: none.
Test references: effects_test.go, identifier_slab_test.go, lower_corpus_test.go, manual_memo_comparison_test.go, memoization_level_test.go, preserve_manual_memoization_test.go.

## manual_memo_comparison_test.go — 227 lines

Purpose: TestCompareManualMemoDependenciesAsymmetry pins the direction, which is the whole rule.
Defines: TestCompareManualMemoDependenciesAsymmetry, TestCompareManualMemoDependenciesComparesOptionality, TestCompareManualMemoDependenciesGlobalRoots, TestCompareManualMemoDependenciesWithdrawsTheRuleForRefs, TestMergeCompareDependencyResultsTakesTheMostSpecific.
Package references: dependencies.go, drop_manual_memoization.go, effects.go, high_level_intermediate_representation.go, instruction.go, manual_memo_comparison.go, memoization_level.go, preserve_manual_memoization.go, primitive_property_constraints.go, ranges.go, terminal.go.
Imports: github.com/system-inc/cohere/static_single_assignment, testing.

## memo_block_scope_oracle_test.go — 482 lines

Purpose: The relationship between a memo block's markers and the scopes that carry its dependencies,
Defines: TestMemoBlockScopeRelationshipAcrossCorpus, ensure, memoBlockObservation, memoBlockObservations, memoBlockObserver, observations, observeMemoBlocks, scopeCaptures, visitInstruction, walk, walkTerminal.
Package references: align_scopes.go, dependencies.go, disjoint.go, drop_manual_memoization.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, merge_invalidating.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/rules/react/conformance, github.com/system-inc/cohere/internal/lint/testing.

## memoization_graph.go — 242 lines

Purpose: The identifier graph `pruneNonEscapingScopes` walks, and the propagation over it.
Defines: AssociateScope, ComputeMemoized, Declare, EscapingCount, Len, MarkEscaping, MemoizationGraph, NewMemoizationGraph, Record, memoizationNode, sortedDeclarations, sortedScopes.
Package references: align_scopes.go, dependencies.go, effects.go, memoization_level.go, merge_invalidating.go, merge_scopes.go, primitive_property_constraints.go, scopes.go.
Imports: github.com/system-inc/cohere/static_single_assignment, sort.
Test references: align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, effects_test.go, hoistable_test.go, memoization_graph_test.go, merge_invalidating_test.go, merge_scopes_test.go, prune_non_escaping_scopes_test.go, ranges_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go.

## memoization_graph_test.go — 260 lines

Purpose: TestComputeMemoizedWalksFromEscapingRoots is the core propagation, and the case a boolean model
Defines: TestComputeMemoizedForcesScopeDependencies, TestComputeMemoizedHandlesUnknownDeclarations, TestComputeMemoizedIsRepeatable, TestComputeMemoizedRespectsTheLevels, TestComputeMemoizedTerminatesOnCycles, TestComputeMemoizedWalksFromEscapingRoots, TestRecordJoinsRatherThanAssigns.
Package references: dependencies.go, memoization_graph.go, memoization_level.go.
Imports: github.com/system-inc/cohere/static_single_assignment, testing.

## memoization_inputs.go — 169 lines

Purpose: Classifying every instruction value by how strongly it needs memoizing.
Defines: MemoizationInputsGap, MemoizationInputsGapCallSignatures, MemoizationInputsGapOperandMutability, MemoizationInputsGaps, MemoizationLevelOf, MemoizationLevelOfReactiveValue.
Package references: instruction.go, memoization_level.go, reactive_function.go.
Imports: none.
Test references: memoization_inputs_test.go.

## memoization_inputs_test.go — 179 lines

Purpose: TestMemoizationLevelOfCoversEveryInstructionValue is the guard that makes the default safe.
Defines: TestMemoizationInputsGapsAreDeclared, TestMemoizationLevelOfCoversEveryInstructionValue, TestMemoizationLevelOfMatchesUpstreamGroups, TestMemoizationLevelOfReactiveValueHandlesComposites.
Package references: instruction.go, memoization_inputs.go, memoization_level.go, reactive_function.go.
Imports: testing.

## memoization_level.go — 74 lines

Purpose: How strongly a value needs memoizing, which is the lattice `pruneNonEscapingScopes` walks.
Defines: JoinMemoizationLevels, MemoizationLevel, MemoizationMemoized, MemoizationNever, MemoizationUnmemoized, String.
Package references: drop_manual_memoization.go, effects.go, high_level_intermediate_representation.go, manual_memo_comparison.go, preserve_manual_memoization.go, ranges.go.
Imports: none.
Test references: effects_test.go, identifier_slab_test.go, lower_corpus_test.go, manual_memo_comparison_test.go, memoization_graph_test.go, memoization_inputs_test.go, memoization_level_test.go.

## memoization_level_test.go — 83 lines

Purpose: TestJoinMemoizationLevelsMatchesUpstreamChain asserts the join against upstream's own spelling.
Defines: TestJoinMemoizationLevelsMatchesUpstreamChain, TestMemoizationLevelOrderIsStrength.
Package references: drop_manual_memoization.go, effects.go, high_level_intermediate_representation.go, manual_memo_comparison.go, memoization_level.go, preserve_manual_memoization.go, ranges.go.
Imports: testing.

## merge_consecutive_blocks.go — 195 lines

Purpose: Merging blocks that always execute one after the other.
Defines: MergeConsecutiveBlocks, singlePhiOperand.
Package references: dead_code_elimination.go, dependencies.go, graph.go, high_level_intermediate_representation.go, instruction.go, prune_non_escaping_scopes.go, reactive_build.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: github.com/system-inc/cohere/static_single_assignment, slices.
Test references: dependency_oracle_test.go, effects_react_state_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, merge_consecutive_blocks_test.go, scope_oracle_test.go.

## merge_consecutive_blocks_test.go — 308 lines

Purpose: TestMergeCollapsesTheRegionTheSpliceCreated is the pass's reason for existing.
Defines: TestMergeCollapsesTheRegionTheSpliceCreated, TestMergeLeavesAJoinBlockAlone, TestMergeLeavesTheGraphResolvable, TestMergeLeavesTheLabeledSpliceAlone.
Package references: dead_code_elimination.go, graph.go, high_level_intermediate_representation.go, merge_consecutive_blocks.go, reactive_build.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: github.com/system-inc/cohere/static_single_assignment, testing.

## merge_invalidating.go — 596 lines

Purpose: Merging adjacent scopes that invalidate together: the tree-level merge.
Defines: AreEqualDependencies, AreLValuesLastUsedByScope, CanMergeScopes, FindLastUsage, LastUsage, LastUsedAt, Len, MergeReactiveScopesThatInvalidateTogether, MergeScopesResult, ScopeIsEligibleForMerging, declaredByOrAliasedFrom, mergeAllowedInstruction, mergeBlock, mergeCandidate, mergeTerminalBlocks, rewrite, scopeMerger.
Package references: align_scopes.go, always_invalidating.go, dependencies.go, disjoint.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, memoization_graph.go, merge_scopes.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.
Test references: align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, effects_react_state_test.go, effects_test.go, hoistable_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, merge_scopes_test.go, preserve_manual_memoization_pruned_test.go, prune_non_escaping_scopes_test.go, prune_unused_scopes_test.go, ranges_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go.

## merge_invalidating_test.go — 761 lines

Purpose: TestFindLastUsageRecordsTheHighestOrder is the property the merge condition reads.
Defines: TestAreEqualDependenciesIsSetEquality, TestAreLValuesLastUsedByScopeDeclinesOnMissingInformation, TestCanMergeScopesDeclinesReassignments, TestCanMergeScopesGuardsAreIndependentlyLoadBearing, TestCanMergeScopesRequiresRootedInvalidatingFlow, TestFindLastUsageCorpus, TestFindLastUsageRecordsTheHighestOrder, TestLastUsedAtReportsMissesRatherThanZero, TestMergeReactiveScopesFoldsInterleavedStatements, TestMergeReactiveScopesPreservesEveryStatement, TestMergeReactiveScopesRecordsAbsorbedIds, TestScopeIsEligibleForMergingTreatsNoDependenciesAsEligible, withInvalidatingScopes.
Package references: align_scopes.go, always_invalidating.go, dependencies.go, disjoint.go, effects.go, graph.go, high_level_intermediate_representation.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, primitive_property_constraints.go, ranges.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, ssa.go, terminal.go, visitor.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## merge_nested_scopes_test.go — 101 lines

Purpose: Tests of the declarations listed below.
Defines: TestManualMemoizationChecksNestedScopeAfterDependencyPruning, TestNestedScopeMergeMatchesEnclosingDependencies, callback, enabled, theme.
Package references: dependencies.go, high_level_intermediate_representation.go, instruction.go, merge_invalidating.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scopes.go, terminal.go.
Imports: github.com/system-inc/cohere/static_single_assignment, testing.

## merge_scopes.go — 765 lines

Purpose: Merging overlapping reactive scopes: making scope ranges nest, so a scope can become a block.
Defines: GroupOf, Ids, Len, MembersOf, MergeGap, MergeGapBlockScopeAlignment, MergeGapPrimitiveOperandSkip, MergeGaps, MergeOverlappingReactiveScopes, MergedScopes, RangeOf, Unions, activeScopeOf, apply, collectScopeInfo, descendingByPosition, find, indexOfScope, mergeSweepState, scopeSet, scopesAtPosition, sortByEndDescending, sortByStartDescending, union, visitInstructionId, visitPlace.
Package references: align_scopes.go, dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, lower.go, memoization_graph.go, merge_invalidating.go, primitive_property_constraints.go, ranges.go, reactive_build.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, sort.
Test references: align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, dependency_oracle_test.go, disjoint_test.go, effects_react_state_test.go, effects_test.go, hoistable_test.go, lower_for_of_kinds_test.go, merge_invalidating_test.go, merge_scopes_test.go, preserve_manual_memoization_test.go, prune_non_reactive_dependencies_test.go, ranges_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go.

## merge_scopes_corpus_test.go — 161 lines

Purpose: forEachCorpusFunction lowers every outermost function-like in the corpus and hands it to visit,
Defines: blockNestingViolations, forEachCorpusFunction, nestingItem, programBlockSubtrees, scopeNestingItems.
Package references: disjoint.go, graph.go, high_level_intermediate_representation.go, lower.go, primitive_property_constraints.go, ranges.go, reactive_build.go, scopes.go, ssa.go, terminal.go, visitor.go.
Imports: path/filepath, sort, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## merge_scopes_test.go — 1001 lines

Purpose: ---------------------------------------------------------------------------
Defines: TestMergeAssumesMonotoneEvaluationOrder, TestMergeConservesMembership, TestMergeDrivesNonNestedOverlapToZero, TestMergeFunctionOperandSkipUpperBound, TestMergeGapsAreDeclared, TestMergeHandlesAnEmptyScopeTable, TestMergeIsASingleSweep, TestMergeIsDeterministic, TestMergeIsNotAPairwiseOverlapRule, TestMergeLeavesBlockScopeAlignmentUnclosed, TestMergeMatchesReact, TestMergePreservesScopeWidthAndMembership, TestMergeRegistersScopesFromTerminals, TestMergeSkipsDegenerateScopes, TestMergeSortComparatorReachability, buildMergeCase, classify, countNonNestedPairs, mergeCase, mergeCorpusStats, unionsSkippingFunctionOperands.
Package references: align_scopes.go, dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, ranges.go, reactive_build.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, sort, testing.

## optional_chains.go — 373 lines

Purpose: Recovering optional chains from the control flow they lower to.
Defines: CollectOptionalChainSidemap, OptionalChainSidemap, matchOptionalTestBlock, optionalChainJoinBlocks, optionalTestMatch, optionalTraversal, recordOptionalChainJoinPhis, traverseFunction, traverseOptionalBlock.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, instruction.go, reactive_build.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: dependencies_test.go, lower_test.go.

## outline_functions.go — 128 lines

Purpose: Outlining function expressions that capture nothing.
Defines: OutlineFunctions, outlinedFunctionByName, outlinedFunctionName.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, ranges.go, reactive_build.go.
Imports: github.com/system-inc/cohere/static_single_assignment, strconv, strings.
Test references: dependency_oracle_test.go, effects_react_state_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, outline_functions_test.go, scope_oracle_test.go.

## outline_functions_test.go — 51 lines

Purpose: The name an outlined function takes must resolve back to that function.
Defines: TestOutlinedFunctionByNameDeclinesOrdinaryGlobals, TestOutlinedFunctionNameRoundTrip.
Package references: high_level_intermediate_representation.go, outline_functions.go.
Imports: testing.

## pattern.go — 60 lines

Purpose: Pattern is a destructuring target.
Defines: ArrayPattern, ArrayPatternElement, ObjectPattern, ObjectPatternProperty, Pattern, PlacePattern, pattern.
Package references: high_level_intermediate_representation.go.
Imports: none.
Test references: disjoint_test.go, lower_test.go, ssa_test.go.

## postdominator.go — 400 lines

Purpose: Post-dominance, and the set of blocks that must execute on the way out of a function.
Defines: ControlDominators, UnconditionalBlocks, computePostDominance, postDominanceTree, postDominatorExit, postDominatorFrontier, postDominatorsOf, reversePostorderFrom, terminalTestSatisfies.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, hoistable.go, preserve_manual_memoization.go, primitive_property_constraints.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, reactive_build.go, terminal.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: controldominators_test.go, postdominator_frontier_test.go, postdominator_test.go.

## postdominator_frontier_test.go — 177 lines

Purpose: TestPostDominatorFrontiersMatchTheChainWalk holds postDominatorFrontiers to the walk it replaced.
Defines: TestPostDominatorFrontiersMatchTheChainWalk, chainWalkFrontiers, functionLikeCount, pinnedCorpusFilesIfPresent.
Package references: graph.go, high_level_intermediate_representation.go, lower.go, postdominator.go, reactive.go, reactive_build.go.
Imports: fmt, reflect, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/core, github.com/microsoft/TypeScript/tsc/shim/parser, github.com/microsoft/TypeScript/tsc/shim/tspath, github.com/system-inc/cohere/static_single_assignment.

## postdominator_test.go — 369 lines

Purpose: The post-dominator tree is checked against the DEFINITION rather than against the algorithm.
Defines: TestPostDominatorDefinitionCheckCanFail, TestPostDominatorsMatchTheDefinition, TestUnconditionalBlocksAlwaysContainsTheEntry, TestUnconditionalBlocksHandlesDegenerateInput, TestUnconditionalBlocksIsASubsetOfPostDominators, TestUnconditionalBlocksOnMeasuredShapes, blockHoldingLastCall, postDominates, reachesAReturn, reachesAReturnWithout.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, postdominator.go, print.go, reactive_build.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment, testing.

## preserve_manual_memoization.go — 601 lines

Purpose: Checking that a developer's own memoization survived compilation.
Defines: AnalyzePreservedManualMemoization, ForEachFunctionLike, PreserveManualMemoizationDependencyMutable, PreserveManualMemoizationFinding, PreserveManualMemoizationKind, PreserveManualMemoizationValueUnmemoized, String, ValidatePreservedManualMemoization, ValidatePreservedManualMemoizationWithDependencies, ValidatePreservedManualMemoizationWithPruned, check, compareInferredDependencies, manualMemoValidator, visitInstruction, walk, walkTerminal.
Package references: align_scopes.go, dead_code_elimination.go, dependencies.go, disjoint.go, drop_manual_memoization.go, effects.go, flatten_reactive_loops.go, flatten_scopes_with_hooks.go, graph.go, high_level_intermediate_representation.go, hoistable.go, inline_iife.go, instruction.go, lower.go, manual_memo_comparison.go, memoization_level.go, merge_consecutive_blocks.go, merge_invalidating.go, outline_functions.go, primitive_property_constraints.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.
Test references: capture_test.go, clone_test.go, computed_member_name_test.go, dependencies_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, identifier_slab_test.go, inline_iife_test.go, lower_corpus_test.go, manual_memo_comparison_test.go, memo_block_scope_oracle_test.go, memoization_level_test.go, preserve_manual_memoization_score_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, ranges_test.go, reactive_capture_identity_test.go, reactive_namespace_refs_test.go.

## preserve_manual_memoization_pruned_test.go — 144 lines

Purpose: TestPrunedScopesAreRecordedRatherThanCompared pins upstream's split into two visitor methods.
Defines: TestPrunedScopesAreRecordedRatherThanCompared, prunedScopeShape.
Package references: align_scopes.go, dependencies.go, disjoint.go, drop_manual_memoization.go, high_level_intermediate_representation.go, lower.go, merge_invalidating.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## preserve_manual_memoization_score_test.go — 606 lines

Purpose: TestPreserveManualMemoizationAgainstGoldens scores the rule against upstream's own answers.
Defines: TestPreserveManualMemoizationAgainstGoldens, TestPreserveManualMemoizationFalsePositiveRate, findingsForSource, pipelineFindings, upstreamReportsInLogs.
Package references: high_level_intermediate_representation.go, lower.go, preserve_manual_memoization.go, ranges.go, ssa.go.
Imports: os, path/filepath, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/rules/react/conformance, github.com/system-inc/cohere/internal/lint/testing.

## preserve_manual_memoization_test.go — 353 lines

Purpose: TestValidatePreservedManualMemoizationFiresAndStaysSilent is the baseline.
Defines: TestConditionalOptionalArgumentPreservesManualMemoization, TestInferredDependencyComparisonIsBuiltAndGatedOnTruncation, TestValidatePreservedManualMemoizationAcceptsAMergedScope, TestValidatePreservedManualMemoizationFiresAndStaysSilent, TestValidatePreservedManualMemoizationIgnoresUnscopedValues, TestValidatePreservedManualMemoizationPairsMarkersById, TestValidatePreservedManualMemoizationSkipsPrunedAndUnopened, inferredDependencyShapes, memoStatement.
Package references: align_scopes.go, dependencies.go, disjoint.go, drop_manual_memoization.go, high_level_intermediate_representation.go, instruction.go, lower.go, manual_memo_comparison.go, merge_scopes.go, preserve_manual_memoization.go, ranges.go, reactive.go, reactive_function.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## primitive_property_constraints.go — 173 lines

Purpose: HIR declarations and operations.
Defines: alias, bind, inferPrimitivePropertyReads, primitiveConstraintKey, primitiveConstraintKind, primitiveConstraintPrimitive, primitiveConstraintState, primitiveConstraintUnknown, root, visit.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, ranges.go, reactive_build.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_survivor_helper_test.go, always_invalidating_test.go, disjoint_test.go, effects_custom_hooks_test.go, effects_test.go, high_level_intermediate_representation_test.go, hoistable_test.go, lower_corpus_test.go, lower_test.go, manual_memo_comparison_test.go, merge_invalidating_test.go, merge_scopes_corpus_test.go, primitive_property_constraints_test.go, prune_non_escaping_scopes_test.go, ssa_corpus_test.go.

## primitive_property_constraints_test.go — 109 lines

Purpose: Tests of the declarations listed below.
Defines: TestManualMemoizationPreservesPrimitivePropertyCapture, TestPrimitivePropertyConstraintsFollowOperators, TestPrimitivePropertyConstraintsRespectIdentity, assertPrimitivePropertyRead, callback, coefficient, theme.
Package references: effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, primitive_property_constraints.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, strings, testing.

## print.go — 392 lines

Purpose: Printing a function back out as text.
Defines: Print, gotoVariantMark, optionalMark, printArgs, printFunction, printManualMemoDependency, printPattern, printTerminal, printValue.
Package references: drop_manual_memoization.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, manual_memo_comparison.go, memoization_level.go, pattern.go, preserve_manual_memoization.go, ranges.go, reactive_build.go, ssa.go, terminal.go, visitor.go.
Imports: fmt, sort, strings.
Test references: capture_test.go, graph_test.go, lower_corpus_test.go, lower_test.go, postdominator_test.go, scope_terminals_test.go, ssa_test.go.

## prune_always_invalidating.go — 192 lines

Purpose: Pruning scopes that will always invalidate, so they stop paying for comparisons they cannot win.
Defines: PruneAlwaysInvalidatingScopes, alwaysInvalidatingPruner, visitInstruction, visitScope, walk, walkTerminal.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, instruction.go, preserve_manual_memoization.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, reactive_build.go, reactive_function.go, reactive_visitor.go, terminal.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: capture_test.go, computed_member_name_test.go, dependencies_test.go, effects_react_state_test.go, memo_block_scope_oracle_test.go, preserve_manual_memoization_pruned_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, ranges_test.go, scope_oracle_test.go.

## prune_always_invalidating_test.go — 204 lines

Purpose: TestPruneAlwaysInvalidatingPrunesSomethingAndKeepsSomething is the baseline.
Defines: TestPruneAlwaysInvalidatingPropagatesOutward, TestPruneAlwaysInvalidatingPrunesSomethingAndKeepsSomething, TestPruneAlwaysInvalidatingRespectsWithinScope, TestPruneAlwaysInvalidatingSeedsFromAllocationsOnly.
Package references: dependencies.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, prune_always_invalidating.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scopes.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.

## prune_non_escaping_scopes.go — 620 lines

Purpose: Pruning scopes whose values never escape, which is the reason the memoization lattice exists.
Defines: PruneNonEscapingScopes, PruneNonEscapingScopesResult, PruneNonEscapingScopesWithScopes, associate, associateChain, associateScope, eachMemoizationOperand, isRestPlaceOfPattern, memoizationCollector, memoizationLevelOfDefine, pruneNonEscapingScopesWith, prunePassOverMemoMarkers, prunePassOverScopes, resolve, visitInstruction, visitTerminal, walk.
Package references: align_scopes.go, dependencies.go, disjoint.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, memoization_graph.go, memoization_inputs.go, memoization_level.go, merge_invalidating.go, merge_scopes.go, pattern.go, preserve_manual_memoization.go, primitive_property_constraints.go, prune_always_invalidating.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, reactive.go, reactive_build.go, reactive_function.go, reactive_transform.go, reactive_visitor.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.
Test references: capture_test.go, computed_member_name_test.go, dependencies_test.go, effects_react_state_test.go, memo_block_scope_oracle_test.go, preserve_manual_memoization_pruned_test.go, prune_non_escaping_scopes_test.go, ranges_test.go, scope_oracle_test.go.

## prune_non_escaping_scopes_test.go — 557 lines

Purpose: TestPruneNonEscapingScopesPrunesAndKeeps is the baseline, and it guards two opposite failures.
Defines: TestMemoMarkersArePrunedOnlyWhenTheirScopeWas, TestPruneNonEscapingScopesIsDeterministic, TestPruneNonEscapingScopesKeepsAReturnedValue, TestPruneNonEscapingScopesPrunesAValueNothingHolds, TestPruneNonEscapingScopesPrunesAndKeeps, TestPruneNonEscapingScopesReportsWhatItReplaced, TestPruneNonEscapingScopesResolvesLoadLocalIndirection, TestPruneNonEscapingScopesTreatsHookArgumentsAsEscaping, TestPruneNonEscapingScopesVisitsASequenceNestedInALogical, hookArgumentRoots, markerCounts, memoMarkerCountsForSource, runTypedSource.
Package references: align_scopes.go, dependencies.go, disjoint.go, drop_manual_memoization.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, memoization_graph.go, merge_invalidating.go, preserve_manual_memoization.go, primitive_property_constraints.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive.go, reactive_build.go, reactive_function.go, reactive_transform.go, reactive_visitor.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/rules/react/conformance, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## prune_non_reactive_dependencies.go — 155 lines

Purpose: Dropping dependencies that cannot change, so a scope stops comparing values that never differ.
Defines: PruneNonReactiveDependencies, nonReactivePruner, visitScope, walk, walkTerminal.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_unused_scopes.go, reactive_build.go, reactive_function.go, reactive_visitor.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: capture_test.go, computed_member_name_test.go, dependencies_test.go, effects_react_state_test.go, memo_block_scope_oracle_test.go, preserve_manual_memoization_pruned_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, ranges_test.go, scope_oracle_test.go.

## prune_non_reactive_dependencies_test.go — 197 lines

Purpose: TestPruneNonReactiveDependenciesKeepsAndPrunes is the baseline, and it matters more here than
Defines: TestPruneNonReactiveDependenciesDoesNotPropagateFromAnInertScope, TestPruneNonReactiveDependenciesKeepsAndPrunes, TestPruneNonReactiveDependenciesPropagatesToScopeOutputs, TestPruneNonReactiveDependenciesVisitsInnermostFirst.
Package references: align_scopes.go, dependencies.go, graph.go, high_level_intermediate_representation.go, instruction.go, merge_scopes.go, prune_non_reactive_dependencies.go, reactive.go, reactive_function.go, scopes.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.

## prune_unused_scopes.go — 152 lines

Purpose: Turning scopes with no outputs into ordinary blocks.
Defines: PruneUnusedScopes, hasOwnDeclaration, scopePruner, visitScope, walk, walkTerminal.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, preserve_manual_memoization.go, primitive_property_constraints.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, reactive_build.go, reactive_function.go, reactive_visitor.go, terminal.go.
Imports: none.
Test references: capture_test.go, computed_member_name_test.go, dependencies_test.go, effects_react_state_test.go, memo_block_scope_oracle_test.go, preserve_manual_memoization_pruned_test.go, prune_non_escaping_scopes_test.go, prune_unused_scopes_test.go, ranges_test.go, scope_oracle_test.go.

## prune_unused_scopes_test.go — 185 lines

Purpose: TestPruneUnusedScopesPrunesSomethingAndKeepsSomething is the baseline before any specific claim.
Defines: TestPruneUnusedScopesKeepsScopesHoldingAReturn, TestPruneUnusedScopesKeepsScopesWithOwnDeclarations, TestPruneUnusedScopesKeepsScopesWithReassignments, TestPruneUnusedScopesPrunesSomethingAndKeepsSomething, prunableTree.
Package references: align_scopes.go, dependencies.go, disjoint.go, high_level_intermediate_representation.go, merge_invalidating.go, prune_unused_scopes.go, ranges.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.

## ranges.go — 243 lines

Purpose: Mutable ranges: over which span of a function's evaluation a value is still being written.
Defines: Closure, Context, Effects, InferMutableRanges, InferMutableRangesWithEffects, InstructionOrder, MutationSite, MutationSiteCall, MutationSiteComputedStore, MutationSiteKind, MutationSitePropertyStore, MutationSites, ParametersFrozen, RangesForNested, ReturnValue, StoredContextValue, String, TerminalOrder, newRangeGraph, rangeGraph.
Package references: drop_manual_memoization.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, manual_memo_comparison.go, memoization_level.go, preserve_manual_memoization.go, ranges_readonly_closures.go, reactive_build.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: align_method_calls_test.go, align_survivor_helper_test.go, always_invalidating_test.go, cache_test.go, capture_test.go, clone_test.go, dependencies_test.go, dependency_oracle_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_computed_load_test.go, effects_custom_hooks_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, identifier_slab_test.go, inline_iife_test.go, inline_remap_test.go, invoked_functions_test.go, lower_corpus_test.go, lower_for_of_kinds_test.go, lower_global_test.go, lower_test.go, manual_memo_comparison_test.go, memo_block_scope_oracle_test.go, memoization_level_test.go, merge_invalidating_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_score_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, prune_unused_scopes_test.go, ranges_context_closures_test.go, ranges_test.go, reactive_capture_identity_test.go, reactive_namespace_refs_test.go, reactive_stable_hooks_test.go, reactive_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go, ssa_corpus_test.go, ssa_test.go.

## ranges_context_closures_test.go — 77 lines

Purpose: Tests of the declarations listed below.
Defines: TestManualMemoizationPreservesMixedCaptureConditionalCall, TestMixedCaptureClosureRefinementRetainsEffectGuards, callback, item, value.
Package references: high_level_intermediate_representation.go, instruction.go, ranges.go.
Imports: strings, testing, github.com/system-inc/cohere/mutation_aliasing.

## ranges_immutable_edges_test.go — 107 lines

Purpose: Tests of the declarations listed below.
Defines: TestImmutableSourcesDoNotWidenThroughMutableFallbacks, TestManualMemoizationPreservesFrozenPropertyFallback, items, projects, request, result.
Package references: effects.go, graph.go, instruction.go, lower.go.
Imports: fmt, github.com/system-inc/cohere/static_single_assignment, testing.

## ranges_maybe_frozen_test.go — 122 lines

Purpose: Tests of the declarations listed below.
Defines: TestClosureRefExclusionFollowsDerivedValues, TestManualMemoizationPreservesMixedValueClosure, TestReadOnlyClosureEffectsRejectUnprovenBodies, copied, current, nested, reference, selected, unrelated.
Package references: graph.go, high_level_intermediate_representation.go, ranges_readonly_closures.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment, strings, testing.

## ranges_readonly_closures.go — 142 lines

Purpose: HIR declarations and operations.
Defines: hasReadOnlyClosureEffects, hasReadOnlyClosureEffectsForCaptures, refDerivedValues.
Package references: effects.go, effects_custom_hooks.go, graph.go, high_level_intermediate_representation.go, instruction.go, ranges.go, reactive.go, reactive_build.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: ranges_maybe_frozen_test.go.

## ranges_test.go — 1376 lines

Purpose: rangesFor lowers one source, constructs single-assignment form, and returns the outermost
Defines: TestAliasingRefinementMatchesAbstractKinds, TestBlockFirstOrderFallsBackToTheTerminal, TestCreateFromMutationPropagatesTransitively, TestMutatingMethodReceiverSurvivesArgumentControlFlow, TestMutationSitesNameTheMutatedValueNotTheStoredOne, TestMutationSitesNamesWhatStageOneWouldWiden, TestPhiNonMutabilityControlsUnknownCallScopes, TestRangeGapsAreNamed, TestRangeIsSetTreatsEitherFieldAsSet, TestRangesAreHalfOpen, TestRangesAreIdempotent, TestRangesDoNotWidenThroughAnEdgeThatDidNotExistYet, TestRangesDoNotWidenWithoutAMutation, TestRangesFollowCapturesOnlyForATransitiveMutation, TestRangesLeaveParametersUnset, TestRangesOpenAtTheDefiningInstruction, TestRangesProduceNoInvalidIntervalsOnRealCode, TestRangesRejectAnInvalidInterval, TestRangesRequireEvaluationOrder, TestRangesSpanTheGapBetweenDisjointWrites, TestRangesStayValidAcrossALoopBackEdge, TestRangesThirdLoopOpensAWidenedOperand, TestRangesWidenAStoreIntoACapturedBinding, TestRangesWidenToCoverAMutatingCall, rangeOfName, rangesFor, unknownCallAndArgumentPhi.
Package references: align_scopes.go, dependencies.go, disjoint.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive_build.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## reactive.go — 1067 lines

Purpose: Reactivity inference: which values can change between renders.
Defines: InferReactive, IsHookCallee, ReactiveGap, ReactiveGapAliasing, ReactiveGapMutation, ReactiveGaps, applyToPlaces, eachInstructionLValue, identifierNode, isHookCallee, isHookName, isReactive, isStable, isStableType, mark, maximumReactiveRounds, postDominatorFrontiers, propagateToNested, reactivity, recordHookResult, recordStablePositions, refreshControlled, run, setPlace, stableTypeName, terminalTestIsReactive, useRefResultValues, visitBlock.
Package references: align_scopes.go, effects.go, graph.go, high_level_intermediate_representation.go, hoistable.go, instruction.go, pattern.go, postdominator.go, primitive_property_constraints.go, ranges.go, reactive_build.go, reactive_visitor.go, ssa.go, terminal.go, visitor.go, visitor_mutate.go.
Imports: slices, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/static_single_assignment.
Test references: align_method_calls_test.go, dependency_oracle_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_react_state_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, memo_block_scope_oracle_test.go, postdominator_frontier_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, reactive_capture_identity_test.go, reactive_namespace_refs_test.go, reactive_stable_hooks_test.go, reactive_test.go, scope_oracle_test.go.

## reactive_build.go — 1264 lines

Purpose: The conversion itself: walking the graph once and emitting a tree.
Defines: BuildReactiveFunction, BuildReactiveFunctionWithFlattenedScopes, ReactiveBuildResult, block, breakTarget, continueTarget, controlFlowIf, controlFlowKind, controlFlowLoop, controlFlowSwitch, controlFlowTarget, eachNestedBlock, eachTerminalValue, emitGoto, emitStructuredValueTerminal, emitValueTerminal, extractValueBlockResult, isScheduled, labelFor, measureBlock, measureValue, newReactiveContext, reactiveContext, reactiveInstructions, reactiveValueBlockResult, reactiveValueTerminalResult, schedule, scheduleFallthrough, scheduleLoop, scheduleLoopTargets, traverse, unschedule, unscheduleAll, valueJoinPlace, valueOf, visitBlock, visitFallthrough, visitTerminal, visitTestValueBlock, visitValueBlock, visitValueBlockTerminal.
Package references: align_scopes.go, effects.go, graph.go, high_level_intermediate_representation.go, instruction.go, primitive_property_constraints.go, prune_non_escaping_scopes.go, reactive.go, reactive_function.go, reactive_visitor.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_case_test.go, align_scopes_test.go, align_survivor_helper_test.go, cache_test.go, capture_test.go, controldominators_test.go, dependencies_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, graph_test.go, high_level_intermediate_representation_test.go, inline_branch_fallthrough_test.go, inline_iife_test.go, inline_remap_test.go, lower_for_of_kinds_test.go, lower_test.go, memo_block_scope_oracle_test.go, merge_consecutive_blocks_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, postdominator_frontier_test.go, postdominator_test.go, preserve_manual_memoization_pruned_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_unused_scopes_test.go, ranges_test.go, reactive_build_test.go, reactive_stable_hooks_test.go, reactive_test.go, reactive_transform_test.go, reactive_visitor_test.go, scope_oracle_test.go, scope_terminals_test.go, ssa_test.go.

## reactive_build_test.go — 522 lines

Purpose: TestBuildReactiveFunctionShapes asserts the tree each control-flow construct converts to.
Defines: TestBuildReactiveFunctionCorpusConservation, TestBuildReactiveFunctionLogicalKeepsRightPrefix, TestBuildReactiveFunctionOptionalValue, TestBuildReactiveFunctionShapes, TestBuildReactiveFunctionTernaryValue, TestReactiveFunctionGapsAreDeclared, hasScopeTerminalInLoopValueBlock.
Package references: align_scopes.go, graph.go, high_level_intermediate_representation.go, instruction.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, testing.

## reactive_capture_identity_test.go — 75 lines

Purpose: Tests of the declarations listed below.
Defines: TestNestedReactivityUsesCaptureIdentity, stableValue.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, preserve_manual_memoization.go, ranges.go, reactive.go, ssa.go.
Imports: testing.

## reactive_function.go — 436 lines

Purpose: ReactiveFunction: the control-flow graph turned back into a tree.
Defines: ReactiveBlock, ReactiveBreak, ReactiveContinue, ReactiveDoWhile, ReactiveFor, ReactiveForIn, ReactiveForOf, ReactiveFunction, ReactiveFunctionGap, ReactiveFunctionGapPrunedScopes, ReactiveFunctionGapUnprunedLabels, ReactiveFunctionGapValueExpressions, ReactiveFunctionGaps, ReactiveIf, ReactiveInstruction, ReactiveInstructionStatement, ReactiveInstructionValue, ReactiveLabel, ReactiveLabelTerminal, ReactiveLogicalValue, ReactiveOptionalValue, ReactiveReturn, ReactiveScopeBlock, ReactiveSequenceValue, ReactiveStatement, ReactiveSwitch, ReactiveSwitchCase, ReactiveTargetImplicit, ReactiveTargetLabeled, ReactiveTerminal, ReactiveTerminalStatement, ReactiveTerminalTargetKind, ReactiveTernaryValue, ReactiveThrow, ReactiveTry, ReactiveValue, ReactiveWhile, reactiveStatement, reactiveTerminal, reactiveValue.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, scopes.go, terminal.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: memo_block_scope_oracle_test.go, memoization_inputs_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, prune_unused_scopes_test.go, reactive_build_test.go, reactive_transform_test.go, reactive_visitor_test.go, scope_oracle_test.go.

## reactive_namespace_refs_test.go — 192 lines

Purpose: Tests of the declarations listed below.
Defines: TestManualMemoizationPreservesNamespaceRefCallbacks, TestRefStabilityRecognizesNamespaceCalls, TestStableTupleHookPositionsRecognizeNamespaceCalls, callback, reference, result, second.
Package references: instruction.go, lower.go, preserve_manual_memoization.go, ranges.go, reactive.go, ssa.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## reactive_stable_hooks_test.go — 121 lines

Purpose: TestStableHookPositionsAreExemptedFromReactivity pins the positional route to React's stable set.
Defines: TestStableHookPositionsAreExemptedFromReactivity, reactivityByName.
Package references: graph.go, high_level_intermediate_representation.go, lower.go, ranges.go, reactive.go, reactive_build.go, ssa.go, visitor.go.
Imports: testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing.

## reactive_test.go — 807 lines

Purpose: reactiveDeclarations is the part of `@types/react` these tests read.
Defines: TestReactiveConvergesInFewRounds, TestReactiveDistinguishesHookNameFromOrdinaryCall, TestReactiveExemptsAliasedSetter, TestReactiveExemptsStableHookValues, TestReactiveGapsAreNamed, TestReactiveIsAnUnderApproximation, TestReactiveIsIdempotent, TestReactiveJoinsAtPhi, TestReactiveJoinsReactiveOperandUnderConstantBranch, TestReactiveLeavesConstantBranchAlone, TestReactiveLeavesConstantsAlone, TestReactiveMarksHookResults, TestReactiveMarksNamespacedHookResults, TestReactiveMarksParameters, TestReactiveMarksValueAssignedUnderReactiveBranch, TestReactiveMarksValueUnderReactiveSwitchCase, TestReactiveMarksValueUnderReactiveSwitchDiscriminant, TestReactivePhiResultIsMarkedByOperandJoin, TestReactivePhiResultIsMarkedByReactiveControl, TestReactivePropagatesThroughAliases, TestReactiveRequiresCapitalAfterUse, TestReactiveWithoutCheckerStillMarksParameters, assertNotReactive, assertReactive, collectReactiveNames, hasName, phiResultIsReactive, reactiveDeclarations, reactiveNames.
Package references: graph.go, high_level_intermediate_representation.go, lower.go, ranges.go, reactive.go, reactive_build.go, ssa.go, terminal.go, visitor.go.
Imports: sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## reactive_transform.go — 217 lines

Purpose: Rewriting a ReactiveFunction: the transform half of the walk.
Defines: KeepStatement, ReactiveTransformKeep, ReactiveTransformRemove, ReactiveTransformReplaceMany, ReactiveTransformed, ReactiveTransformedKind, ReactiveTransformer, RemoveStatement, ReplaceStatement, ReplaceStatements, TransformReactiveFunction, reactiveTransformWalker, transformBlock, transformStatement, transformTerminalBlocks.
Package references: graph.go, high_level_intermediate_representation.go, reactive_build.go, reactive_function.go, terminal.go.
Imports: none.
Test references: prune_non_escaping_scopes_test.go, reactive_transform_test.go.

## reactive_transform_test.go — 276 lines

Purpose: TestReactiveTransformKeepDoesNotReallocate is the property the lazy rebuild exists for.
Defines: TestReactiveTransformCorpus, TestReactiveTransformKeepDoesNotReallocate, TestReactiveTransformRemovesAndReplaces, TestReactiveTransformRewritesNestedBlocks, countInstructionStatements, countStatements.
Package references: align_scopes.go, high_level_intermediate_representation.go, reactive_build.go, reactive_function.go, reactive_transform.go, scope_terminals.go, scopes.go, terminal.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, testing.

## reactive_visitor.go — 282 lines

Purpose: Walking a ReactiveFunction: the traversal every remaining pass runs on.
Defines: ReactiveVisitor, VisitReactiveFunction, reactiveWalker, traverseBlock, traverseInstruction, traverseTerminal, traverseValue, visitBlock, visitInstruction, visitPlace, visitScope, visitTerminal, visitValue.
Package references: align_scopes.go, graph.go, high_level_intermediate_representation.go, merge_scopes.go, preserve_manual_memoization.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, reactive.go, reactive_build.go, reactive_function.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_survivor_helper_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, merge_scopes_test.go, preserve_manual_memoization_pruned_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_unused_scopes_test.go, reactive_build_test.go, reactive_visitor_test.go, scope_oracle_test.go.

## reactive_visitor_test.go — 340 lines

Purpose: TestReactiveVisitorReachesEveryInstruction is the conservation property, applied to the walk.
Defines: TestReactiveVisitorCorpus, TestReactiveVisitorPrunesWhenTraverseIsNotCalled, TestReactiveVisitorReachesEveryInstruction, TestReactiveVisitorSeesSequenceInstructions, countInstructionNodesInTerminal, countInstructionNodesInValue, countReactiveInstructionNodes.
Package references: align_scopes.go, graph.go, high_level_intermediate_representation.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, terminal.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, testing.

## scope_oracle_test.go — 406 lines

Purpose: Scoring our reactive scope structure against upstream's own compiled output.
Defines: TestScopeStructureAgainstUpstreamGuards, countScopeStages, scopeStageCounts.
Package references: align_scopes.go, dead_code_elimination.go, dependencies.go, disjoint.go, drop_manual_memoization.go, effects.go, high_level_intermediate_representation.go, inline_iife.go, lower.go, memoization_graph.go, merge_consecutive_blocks.go, merge_invalidating.go, merge_scopes.go, outline_functions.go, prune_always_invalidating.go, prune_non_escaping_scopes.go, prune_non_reactive_dependencies.go, prune_unused_scopes.go, ranges.go, reactive.go, reactive_build.go, reactive_function.go, reactive_visitor.go, scope_terminals.go, scopes.go, ssa.go, terminal.go.
Imports: fmt, os, regexp, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/checker, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/rules/react/conformance, github.com/system-inc/cohere/internal/lint/testing.

## scope_terminals.go — 666 lines

Purpose: Reactive scope terminals: turning the scope side table into control-flow structure.
Defines: BuildReactiveScopeTerminals, GroupOf, MergedScopeIdentity, RangeOf, ScopeIdentity, ScopeTerminals, ScopeTerminalsGap, ScopeTerminalsGapPrunedScope, ScopeTerminalsGaps, ScopeTerminalsPrecondition, applyScopeRewrites, fixScopeRanges, phisFor, queueScopeRewrites, reindexBlocks, scopeItem, scopeItemsInNestingOrder, scopeRewrite.
Package references: align_scopes.go, dependencies.go, effects.go, graph.go, high_level_intermediate_representation.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, primitive_property_constraints.go, ranges.go, reactive_build.go, scopes.go, ssa.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment, sort.
Test references: align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, dependency_oracle_test.go, effects_react_state_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, lower_for_of_kinds_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_scopes_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, prune_unused_scopes_test.go, ranges_test.go, reactive_build_test.go, reactive_transform_test.go, reactive_visitor_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go.

## scope_terminals_test.go — 724 lines

Purpose: terminalsFor lowers one source, runs the full scope pipeline, and builds the terminals.
Defines: GroupOf, RangeOf, TestAScopeEndingWhereAnotherBeginsClosesFirst, TestPrunedScopeIsNotDeclared, TestScopeEndJumpsWithBreak, TestScopeRewritesAreQueuedInTraversalOrderNotSortedByPosition, TestScopeTerminalIsReachableThroughFallthrough, TestScopeTerminalPrints, TestScopeTerminalsAreASinglePass, TestScopeTerminalsAreBuilt, TestScopeTerminalsDeclineOnUnnestedScopes, TestScopeTerminalsGapsAreDeclared, TestScopeTerminalsKeyOnTheMergedGroup, TestScopeTerminalsPreserveSingleAssignmentForm, TestScopeTerminalsRecoverEveryScopeFromTheGraph, TestScopeTerminalsRenumberTheGraph, overlappingIdentity, scopeTerminalsIn, terminalsFor.
Package references: align_scopes.go, dependencies.go, disjoint.go, effects.go, graph.go, high_level_intermediate_representation.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, print.go, ranges.go, reactive_build.go, scope_terminals.go, scopes.go, ssa.go, ssa_verify.go, terminal.go, visitor.go.
Imports: os, path/filepath, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## scopes.go — 440 lines

Purpose: Reactive scopes: which values are memoized together, and over what span.
Defines: AssignReactiveScopes, AssignReactiveScopesWithSets, Ids, Len, MemberRanges, MembersOf, RangeOf, ReactiveScopes, ScopeGap, ScopeGapPostAlignmentWidening, ScopeGaps, ScopeId, ScopeOf, ValidateScopes.
Package references: align_scopes.go, dependencies.go, disjoint.go, effects.go, graph.go, high_level_intermediate_representation.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, ranges.go, reactive_build.go, scope_terminals.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.
Test references: align_case_test.go, align_method_calls_test.go, align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, dependency_oracle_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, lower_for_of_kinds_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, prune_unused_scopes_test.go, ranges_test.go, reactive_build_test.go, reactive_transform_test.go, reactive_visitor_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go, ssa_test.go.

## scopes_test.go — 920 lines

Purpose: scopesFor lowers one source and returns the outermost function with its scopes.
Defines: TestScopeAbsentValueHasNoScope, TestScopeAssignmentIsASinglePass, TestScopeAssignmentIsDeterministic, TestScopeAssignmentIsIdempotent, TestScopeGapsAreDeclared, TestScopeHandlesANilFunction, TestScopeHullIsNotAPlainMinimum, TestScopeHullMatchesReact, TestScopeIsOnePerEquivalenceClass, TestScopeMatchesTheDisjointSetPartition, TestScopeMemberRangesAreRewrittenToTheHull, TestScopeMembersOfIsStableAcrossCalls, TestScopeRangesSatisfyUpstreamsInvariant, TestScopeWidthAndMembershipAgree, TestValidateScopesCatchesAStartOfZeroWithARealEnd, TestValidateScopesCatchesAnEmptyHull, TestValidateScopesCatchesAnEndOfZero, corpusScopeStats, scopeOfName, scopesFor.
Package references: align_scopes.go, dependencies.go, disjoint.go, effects.go, high_level_intermediate_representation.go, lower.go, memoization_graph.go, merge_invalidating.go, merge_scopes.go, ranges.go, scope_terminals.go, scopes.go, ssa.go.
Imports: os, path/filepath, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## spelling.go — 154 lines

Purpose: Whether source text can spell what a React Compiler pass recognises, answered before any lowering.
Defines: MayHoldComponentOrHook, MayNameManualMemoization, SpellsManualMemoization, cooksToName, mayHoldComponentOrHook, spellsHookName.
Package references: primitive_property_constraints.go, ranges.go.
Imports: strings, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/microsoft/TypeScript/tsc/shim/core, github.com/system-inc/cohere/internal/lint/ecmascript/react, github.com/system-inc/cohere/internal/lint/rule.
Test references: cache_test.go.

## ssa.go — 114 lines

Purpose: Single static assignment form: rename every definition so each value is written exactly once,
Defines: Construct, PhiOperandsInOrder.
Package references: graph.go, high_level_intermediate_representation.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_method_calls_test.go, always_invalidating_test.go, cache_test.go, capture_test.go, clone_test.go, dependencies_test.go, dependency_oracle_test.go, disjoint_test.go, drop_manual_memoization_deps_test.go, effects_custom_hooks_test.go, effects_react_effect_test.go, effects_react_state_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, hoistable_test.go, identifier_slab_test.go, inline_iife_test.go, inline_remap_test.go, invoked_functions_test.go, lower_for_of_kinds_test.go, memo_block_scope_oracle_test.go, merge_invalidating_test.go, merge_scopes_corpus_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_score_test.go, preserve_manual_memoization_test.go, prune_non_escaping_scopes_test.go, ranges_test.go, reactive_capture_identity_test.go, reactive_namespace_refs_test.go, reactive_stable_hooks_test.go, reactive_test.go, scope_oracle_test.go, scope_terminals_test.go, scopes_test.go, ssa_corpus_test.go, ssa_test.go.

## ssa_corpus_test.go — 151 lines

Purpose: Single static assignment construction over a real codebase.
Defines: TestSSAOverRealCodebase.
Package references: graph.go, high_level_intermediate_representation.go, lower.go, primitive_property_constraints.go, ranges.go, ssa.go, ssa_verify.go, terminal.go.
Imports: os, path/filepath, sort, strings, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/static_single_assignment.

## ssa_eliminate.go — 37 lines

Purpose: Redundant phi elimination.
Defines: EliminateRedundantPhis.
Package references: graph.go, high_level_intermediate_representation.go.
Imports: github.com/system-inc/cohere/static_single_assignment.

## ssa_test.go — 512 lines

Purpose: Tests for single static assignment construction.
Defines: TestConstructionIsIdempotent, TestConstructionIsNotReRunAfterTheInline, TestMutatingVisitorCoversEveryValue, TestPhiOperandsAreDeterministic, TestSSAAcrossEveryControlFlowConstruct, TestSSAMatchesReactSpecFixtures, TestSSAVerifierDetectsEachViolationClass, cohereEverywhere, lowerTypedForSSA.
Package references: dependencies.go, graph.go, high_level_intermediate_representation.go, instruction.go, lower.go, pattern.go, print.go, ranges.go, reactive_build.go, scopes.go, ssa.go, ssa_verify.go, visitor.go, visitor_mutate.go.
Imports: os, path/filepath, regexp, testing, github.com/microsoft/TypeScript/tsc/shim/ast, github.com/system-inc/cohere/internal/lint/rule, github.com/system-inc/cohere/internal/lint/testing, github.com/system-inc/cohere/mutation_aliasing, github.com/system-inc/cohere/static_single_assignment.

## ssa_verify.go — 67 lines

Purpose: Verification of single static assignment form: the property, checked rather than asserted.
Defines: CollectSSAStats, VerifySSA, computeDominance.
Package references: graph.go, high_level_intermediate_representation.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: capture_test.go, scope_terminals_test.go, ssa_corpus_test.go, ssa_test.go.

## terminal.go — 353 lines

Purpose: The terminal set.
Defines: Branch, DoWhile, For, ForIn, ForOf, Goto, GotoVariant, GotoVariantBreak, GotoVariantContinue, If, Label, Logical, MaybeThrow, Optional, Return, Scope, Sequence, Switch, SwitchCase, Terminal, Ternary, Throw, Try, Unreachable, Unsupported, While, terminal.
Package references: graph.go, high_level_intermediate_representation.go, scopes.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_case_test.go, align_scopes_test.go, align_survivor_helper_test.go, dependencies_test.go, drop_manual_memoization_deps_test.go, effects_test.go, flatten_scopes_with_hooks_test.go, high_level_intermediate_representation_test.go, hoistable_optional_test.go, hoistable_test.go, inline_branch_fallthrough_test.go, inline_iife_test.go, inline_remap_test.go, lower_for_of_kinds_test.go, lower_test.go, manual_memo_comparison_test.go, memo_block_scope_oracle_test.go, merge_consecutive_blocks_test.go, merge_invalidating_test.go, merge_nested_scopes_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, postdominator_test.go, preserve_manual_memoization_pruned_test.go, preserve_manual_memoization_test.go, prune_always_invalidating_test.go, prune_non_escaping_scopes_test.go, prune_non_reactive_dependencies_test.go, prune_unused_scopes_test.go, ranges_test.go, reactive_build_test.go, reactive_test.go, reactive_transform_test.go, reactive_visitor_test.go, scope_oracle_test.go, scope_terminals_test.go, ssa_corpus_test.go.

## visitor.go — 450 lines

Purpose: Visitor plumbing: the generic walks over an instruction's places, a terminal's successors, and a
Defines: EachInstructionPlace, EachPlace, EachSuccessor, EachSuccessorAndFallthrough, EachTerminalPlace, Fallthrough, PlaceRole, PlaceRoleDefine, PlaceRoleUse, TerminalOrder, eachArgumentPlace, eachPatternPlace, hasDefaultCase.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, pattern.go, primitive_property_constraints.go, ranges.go, reactive_build.go, terminal.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: align_case_test.go, align_scopes_test.go, align_survivor_helper_test.go, effects_test.go, high_level_intermediate_representation_test.go, inline_branch_fallthrough_test.go, inline_iife_test.go, inline_remap_places_test.go, inline_remap_test.go, lower_for_of_kinds_test.go, lower_test.go, merge_consecutive_blocks_test.go, merge_invalidating_test.go, merge_scopes_corpus_test.go, merge_scopes_test.go, postdominator_test.go, ranges_maybe_frozen_test.go, reactive_build_test.go, reactive_stable_hooks_test.go, reactive_test.go, scope_terminals_test.go, ssa_test.go.

## visitor_mutate.go — 328 lines

Purpose: Mutating place visitors: the write-side mirror of visitor.go.
Defines: EachBlockReferencePointer, EachInstructionPlacePointer, EachPlacePointer, EachTerminalPlacePointer, eachArgumentPlacePointer, eachPatternPlacePointer.
Package references: graph.go, high_level_intermediate_representation.go, instruction.go, pattern.go, primitive_property_constraints.go, reactive_build.go, terminal.go, visitor.go.
Imports: github.com/system-inc/cohere/static_single_assignment.
Test references: high_level_intermediate_representation_test.go, inline_iife_test.go, inline_remap_places_test.go, inline_remap_test.go, merge_consecutive_blocks_test.go, ssa_test.go.
