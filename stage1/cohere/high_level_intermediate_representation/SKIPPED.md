# Skipped original Go tests

Pinned cohere: `7945d102a6c18dd36adf9114a758ce646e8b2359`. All **45** observed skips require the private Structure checkout (`COHERE_CORPUS_STRUCTURE`), which is unavailable in this environment. These are environment skips, not unsupported language or intentionally unreachable passes. **None is outside the eight rules’ reach.** The public construction denominator remains 1,465; absent private functions cannot be counted or claimed matched. Flow exclusions remain separate and unchanged.

The unit is the first port checkpoint that must cover the skipped test’s assertion. A later orchestration unit does not move an earlier pass’s validation obligation. Unit 2 applies to all constructed-graph rules; units 5–9 are required by `AnalyzePreservedManualMemoization` for preserve-manual-memoization. Small public cases still execute, but do not substitute for the missing corpus checks.

## Unit 2: Construction, SSA and postdominators (5)

**Needed in unit 2** for static-components and other constructed-graph rules; not safe to retire as out of reach.

| Skipped test | Go definition |
| --- | --- |
| `TestLowerRealCodebase` | `lower_corpus_test.go:192` |
| `TestLowerRealCodebaseIsDeterministic` | `lower_corpus_test.go:301` |
| `TestPostDominatorFrontiersMatchTheChainWalk/pinned_corpus` | `postdominator_frontier_test.go:23` |
| `TestConstructionIsIdempotent` | `ssa_test.go:401` |
| `TestSSAOverRealCodebase` | `ssa_corpus_test.go:32` |

## Unit 5: Disjoint sets (1)

**Needed in unit 5** for preserve-manual-memoization; not safe to retire as out of reach.

| Skipped test | Go definition |
| --- | --- |
| `TestDisjointClassSizesAreNotAllSingletons` | `disjoint_test.go:1145` |

## Unit 6: Scopes, alignment, merging and terminals (22)

**Needed in unit 6** for preserve-manual-memoization; not safe to retire as out of reach.

| Skipped test | Go definition |
| --- | --- |
| `TestMergeRegistersScopesFromTerminals` | `merge_scopes_test.go:927` |
| `TestMergeSortComparatorReachability` | `merge_scopes_test.go:879` |
| `TestMergeIsDeterministic` | `merge_scopes_test.go:838` |
| `TestMergeAssumesMonotoneEvaluationOrder` | `merge_scopes_test.go:791` |
| `TestMergeFunctionOperandSkipUpperBound` | `merge_scopes_test.go:684` |
| `TestMergeLeavesBlockScopeAlignmentUnclosed` | `merge_scopes_test.go:649` |
| `TestMergePreservesScopeWidthAndMembership` | `merge_scopes_test.go:560` |
| `TestMergeDrivesNonNestedOverlapToZero` | `merge_scopes_test.go:508` |
| `TestScopeMatchesTheDisjointSetPartition` | `scopes_test.go:838` |
| `TestScopeRangesSatisfyUpstreamsInvariant` | `scopes_test.go:645` |
| `TestScopeWidthAndMembershipAgree` | `scopes_test.go:580` |
| `TestScopeTerminalsRecoverEveryScopeFromTheGraph` | `scope_terminals_test.go:603` |
| `TestAlignIsDeterministic` | `align_scopes_test.go:892` |
| `TestAlignIsASingleSweep` | `align_scopes_test.go:849` |
| `TestReactReversePostorderClosesFallthroughSelfNesting` | `align_scopes_test.go:805` |
| `TestAlignVoidsTheMergesComparatorVerdicts` | `align_scopes_test.go:723` |
| `TestAlignPreservesScopeWidthAndMembership` | `align_scopes_test.go:669` |
| `TestAlignReDerivesMemberRanges` | `align_scopes_test.go:621` |
| `TestAlignFallthroughsAreUniqueOnceBranchesAreExcluded` | `align_scopes_test.go:570` |
| `TestAlignMustRunBeforeTheMerge` | `align_scopes_test.go:233` |
| `TestAlignClosesTheFullBlockNestingAssertion` | `align_scopes_test.go:117` |
| `TestAlignMethodCallScopesReachesTheCorpus` | `align_method_calls_test.go:53` |

## Unit 7: Dependencies, hoistability and invalidating types (4)

**Needed in unit 7** for preserve-manual-memoization; not safe to retire as out of reach.

| Skipped test | Go definition |
| --- | --- |
| `TestHoistableCorpusDistribution` | `hoistable_test.go:396` |
| `TestDeclarationOriginDiffersFromHoldingScope` | `dependencies_test.go:662` |
| `TestDependencyDistributionIsReal` | `dependencies_test.go:561` |
| `TestIsAlwaysInvalidatingTypeHandlesNodelessIdentifiers` | `always_invalidating_test.go:123` |

## Unit 8: Reactive tree, visitor and transform (3)

**Needed in unit 8** for preserve-manual-memoization; not safe to retire as out of reach.

| Skipped test | Go definition |
| --- | --- |
| `TestBuildReactiveFunctionCorpusConservation` | `reactive_build_test.go:351` |
| `TestReactiveVisitorCorpus` | `reactive_visitor_test.go:196` |
| `TestReactiveTransformCorpus` | `reactive_transform_test.go:190` |

## Unit 9: Pruning and invalidation merging (10)

**Needed in unit 9** for preserve-manual-memoization; not safe to retire as out of reach.

| Skipped test | Go definition |
| --- | --- |
| `TestMergeReactiveScopesRecordsAbsorbedIds` | `merge_invalidating_test.go:650` |
| `TestMergeReactiveScopesPreservesEveryStatement` | `merge_invalidating_test.go:567` |
| `TestFindLastUsageCorpus` | `merge_invalidating_test.go:209` |
| `TestPruneNonEscapingScopesPrunesAndKeeps` | `prune_non_escaping_scopes_test.go:21` |
| `TestPruneAlwaysInvalidatingPrunesSomethingAndKeepsSomething` | `prune_always_invalidating_test.go:14` |
| `TestPruneUnusedScopesPrunesSomethingAndKeepsSomething` | `prune_unused_scopes_test.go:15` |
| `TestPruneNonReactiveDependenciesKeepsAndPrunes` | `prune_non_reactive_dependencies_test.go:17` |
| `TestPruneNonEscapingScopesReportsWhatItReplaced` | `prune_non_escaping_scopes_test.go:515` |
| `TestPruneNonEscapingScopesIsDeterministic` | `prune_non_escaping_scopes_test.go:238` |
| `TestPruneNonEscapingScopesResolvesLoadLocalIndirection` | `prune_non_escaping_scopes_test.go:300` |

## Completion boundary

Provide the authorized private checkout and rerun the original observed suite at the named unit. Five unit-2 tests remain unavailable before construction can be described as tested over the private corpus. The postdominator frontier test’s handwritten cases pass; only its `pinned_corpus` subtest skips. No source or fixture filtering has been added to hide any of these gaps.
