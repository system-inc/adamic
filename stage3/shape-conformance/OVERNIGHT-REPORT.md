Built: views-only integration refresh and complete recheck of the reported graph/runtime and lower harness blockers.
Commits: views ffe428ab26eefb73154adbf570ad99e9c1b4f872 merged as 1e4964e6; evidence checkpoint follows.
Commands and outputs: uncached owned shape oracle PASS in 17.859s; focused shape/allocation/closure/native checks PASS; seven inherited lower tests and five inherited native graph harnesses remain red.
Mutants: existing unsafe field erasure mutants remain caught by pinned successful wrong output versus exit 70 in both backends.
Not covered: full gate; checker-ledger repairs; production array-result temporary remains Unknown; refreshed census is running against the exact adaptation.

The corrected instruction excludes codex/non-null-checked-area. Its in-progress merge was aborted, the unpublished ancestry-only merge discarded, and two untracked merge-resolution files removed. Nothing from that branch was pushed. The owned branch contains only the requested views integration dependency. No main or area reference was changed.

The views merge required no manual code conflict resolutions. Git combined the allocation graph changes automatically; the shape, allocation and closure tests verify the resulting graph. Every refusal is retained. The eight graph fixtures all agree with source Node in native release, sanitized native and JavaScript with leak checking. All 43 original graph count rows pass unchanged. Dynamic good and bad selected array shapes preserve the existing outcome: the read temporary lacks an initializer producer, so checks remain. The negative read still fails loudly and its erased-check mutant finishes with wrong output.

Every previously named lower blocker was rerun: TestOverloadedShorthandFunctionValueStaysNotYet, TestCensusPredicateMarkerKeepsProofBoundaries, TestCensusOverloadRelation/result_covariance, TestNestedFunctionGapsAreLoud/optional/default/rest, TestPhantomArrayRequiredCastsAreErased, TestPhantomArrayCastsAreErased and TestPhantomArrayProofs/cycle. All remain red with the same expectations versus integrated behavior. No expectations were weakened.

Every previously named native harness blocker was rerun: GraphRegionsMillion, GraphLazyRegions, GraphContainerBoundary and GraphRegionsRuntime retain old object-byte expectations; GraphRegionsRuntime also pins an obsolete leak size (observed 467 bytes); GraphClosureEnvironment still supplies a two-argument C callback to a three-argument runtime signature. None is certified passing. These are independent of the eight production fixture and 43 count-row successes.

Commands (all output captured in the linked logs):

```
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/ir ./internal/native ./internal/javascript -run 'TestShape|TestAllocationFlow|TestClosureTargets|TestOptionalWidening|TestOverloadedShorthandFunctionValueStaysNotYet|TestCensusPredicateMarkerKeepsProofBoundaries|TestCensusOverloadRelation|TestNestedFunctionGapsAreLoud|TestPhantomArray' -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestShapeGraphReportedFrontiers|TestShapeGraphCountSnapshot|TestShapeDynamicArrayExecution|TestCheckedViewShapeErasure|TestShapeErasureCountRows|TestShapeCallbackCountRows|TestShapeGenericCountRows)$' -count=1 -v -timeout 10m
go test ./internal/native -run 'TestGraphRegionsMillion|TestGraphLazyRegions|TestGraphContainerBoundary|TestGraphRegionsRuntime|TestGraphClosureEnvironment' -count=1 -v -timeout 10m
```

Logs: overnight/views-packages.log, overnight/views-oracle.log, overnight/views-native-blockers.log. The workspace restart removed prior scratch censuses and adaptation worktrees; their committed evidence remains. The fixed 234ab1aa adaptation is being regenerated from its own detached worktree rather than using newer stage3 sources.
