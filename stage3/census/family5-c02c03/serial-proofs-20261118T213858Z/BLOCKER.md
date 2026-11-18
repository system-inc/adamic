# C02 introduces an earlier checked-view refusal

At `utilities.ts:5339:11`, in `getFunctionFlags` (unit `5317:1`), the unchanged assertion is `(node as FunctionLikeDeclaration).body`. This unit has **no G02 finding** in the baseline and its lowering was already visited.

Before: `NotYet`, lowering, `checked view field modifiers of type NodeArray<Modifier> | NodeArray<ModifierLike> | undefined`.

After: `Refused`, lowering, `a checked view with callable field getPositionOfLineAndCharacter`. Full message: `Adamic 0.1 refuses a checked view with callable field getPositionOfLineAndCharacter; prove the callable body rather than asserting its signature`.

C02 adds this callable slot on Node; FunctionLikeDeclaration inherits it. `internal/lower/interface_cast.go:395–410` deliberately diagnoses callable contracts in its first property pass, before representation limitations in the second. Thus this is an **introduced earlier refusal**, not a later stop reached by clearing G02. The same diagnostic is mirrored by a Boundary at `5339:5`. Pairing grades 7 gone, 2 changed, 0 new, 0 reached: one changed finding plus its Boundary mirror. The comparison’s real exit is 1.

The product test passed (0). Both utilities censuses passed (0) with complete checker-clean relation_report trailers: seven G02 sites before, zero after, zero remaining G02 mentions. The required zero-growth condition nevertheless fails.

Once the completed utilities pair established the blocker, the unrelated in-progress baseline parser census was deliberately terminated: real child exit -15 (SIGTERM), 437.408 seconds. Its partial stream and whole stderr are retained. This cancellation is not an unexplained census or box failure. Matching parser, both program and checker pairs, and the mutant were not run. No C02 landing was pushed. C03 was not created or run. The requested evidence publication and README amendment are complete.
