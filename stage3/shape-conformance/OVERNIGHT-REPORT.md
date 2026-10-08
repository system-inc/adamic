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

## Represented callback references

The first flow group closes references through parentheses, assertion operands and both conditional arms. It also admits function values supplied to clean statically resolved generic function declarations, whose runtime bodies and argument joins are already modeled. This only removes an artificial escape annotation; it does not infer an identity from an asserted signature. A conditional's condition is not traversed as an alias. Diagnosed calls, unknown sinks, rest mappings and spread arguments retain their Unknown frontiers. The assumption is the same as direct generic allocation flow: type arguments do not choose a different runtime body.

Ten independent controls pass: four Free, two conforms-if with ready:number versus boolean, two host Unknown and two spread Unknown. All 35 previous latent controls and ten dynamic-key controls still pass. Four built mutants are semantically caught: drop-conditional-reference loses conditionalOne; drop-parenthesized-reference and drop-asserted-reference lose assertedOne; drop-generic-reference loses genericOne. Each mutant compiles and runs the measurement tool before the independent outcome assertion fails.

Two production witnesses combine conditional, asserted and generic callback calls. The good program agrees with Node in release, ASan/UBSan and JavaScript and has no leaks. The bad program observes Node true then 0, while both checked backends stop at the second ready read with exit 70, expected boolean and found number. Removing the field checks makes valid native execution finish with true then false and JavaScript true then 0; both mutants are caught. The uncached new oracle and count-row tests pass in 12.936s. No emitter change is needed for these already supported runtime forms.

The count-row handoff is:

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions | Graph regions | Graph merges |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| proven-callback-join.a | 4 | 4 | 12 | 14 | 3 | 0 | 0 | 0 |
| nonconforming-callback-join.a | 5 | 1 | 10 | 10 | 5 | 0 | 0 | 0 |

Full census before and after measurements are in progress. No corpus bucket improvement is claimed until both outputs finish and pass the independent audit. This checkpoint contains the passing rule, controls, witnesses, mutants and raw logs, so that completed work is available while the measurements run.

## Generic callable implementation identities

The next callback group follows generic function declaration aliases and generic arrow identities to their clean represented runtime bodies. Type arguments do not select another implementation; every actual invocation still joins through the existing callback graph. Diagnosed implementations keep their named diagnostic provenance. All callable rest implementations, generic or ordinary, now retain `callback rest arguments require an array allocation`, because a rest array cannot be replaced by the first supplied object. Spread expressions retain their existing unmodeled-input boundary.

Six new controls pass: declaration and arrow alias results are Free; a mixed good/number ready join is conforms-if; rest and spread are unsupported Unknown; a later host invocation makes the earlier clean invocation host Unknown. The previous 35 latent, ten dynamic and ten callback-join controls pass unchanged. Three valid built mutants are caught: drop-generic-declaration-identity loses declarationRead; drop-generic-arrow-identity loses arrowRead; ignore-callback-rest-array incorrectly makes restRead Free by treating its rest array as the first object. The control independently pins both the outcome and named rest cause.

This is an analysis-only identity improvement. The merged production compiler still refuses the actual generic-function-value alias witness with `stage 0 can't lower a generic function as a value yet`, at proven-generic-callable.a:4:15. It emits no C. That runtime pair is explicitly not certified; its source and exact diagnostic are retained. No production admission or representation rule was relaxed. A separately built third fixed-source census measures this group against the callback-join output; final bucket numbers await audited completion.
