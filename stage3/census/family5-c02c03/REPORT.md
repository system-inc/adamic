# Family 5 C02/C03

[Current serial proof report](serial-proofs-20261118T213858Z/REPORT.md)

# Family 5 C02/C03: accepted lane and serial proofs

Stopped: C02 introduces a changed lowering finding outside G02 at utilities.ts:5339:11 in getFunctionFlags (unit 5317:1): checked view field modifiers NotYet becomes Refused, a checked view with callable field getPositionOfLineAndCharacter. Its mirrored Boundary also changes. No landing is green.

Main `fbe5b72ba115a8f3060d4b5229fd578574f4505f`; C02 `12b9e04184663f899acb76505043ad292966f236`; C03 `not created`. TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.

C02 lane is green by the owner’s acceptance: both alternating candidate runs passed 106,366 cases with only the one sanctioned failure, real lane exit 0. The earlier reds were nondeterminism. The 19 JavaScript bundles were byte-identical. See [the full alternating report](alternating-runs/REPORT.md), hunks, manifests and all run receipts. Its archive was not committed: SHA-256 `773ed33da75936a4d4c89b40cff991fd75cdbfc4c37b6663a53f7f04a2605b06` (2,084,049 bytes).

C02 was amended from measured `28a2e22015d0cf51667fcefa7acb246cab553817` only to add the requested README lane receipt. `readme-only-amendment.patch.gz` proves the adapter, pins, TypeScript source and JavaScript are unchanged. Previously green checks are carried with their original real exit receipts; fresh remaining proofs ran serially, never beside the lane or another census.

## Proof receipts

| Proof | Real child exit | Wall seconds |
|---|---:|---:|
| c02/checkpoint-comparison | 1 | 0.265 |
| c02/product | 0 | 767.633 |
| c02-after/utilities | 0 | 629.527 |
| c02-before/parser | -15 | 437.408 |
| c02-before/utilities | 0 | 652.015 |
| carried-c02/adamic-types | 0 | 12.383 |
| carried-c02/adapter-contract-final | 0 | 8.612 |
| carried-c02/public-build | 0 | 95.528 |
| carried-c02/public-byte-comparison | 0 | 0.057 |
| carried-c02/stock | 0 | 117.973 |
| carried-c02/stock-strict | 0 | 56.083 |
| carried-c02/whole-apply | 0 | 862.943 |

## Blocking finding

# C02 introduces an earlier checked-view refusal

At `utilities.ts:5339:11`, in `getFunctionFlags` (unit `5317:1`), the unchanged assertion is `(node as FunctionLikeDeclaration).body`. This unit has **no G02 finding** in the baseline and its lowering was already visited.

Before: `NotYet`, lowering, `checked view field modifiers of type NodeArray<Modifier> | NodeArray<ModifierLike> | undefined`.

After: `Refused`, lowering, `a checked view with callable field getPositionOfLineAndCharacter`. Full message: `Adamic 0.1 refuses a checked view with callable field getPositionOfLineAndCharacter; prove the callable body rather than asserting its signature`.

C02 adds this callable slot on Node; FunctionLikeDeclaration inherits it. `internal/lower/interface_cast.go:395–410` deliberately diagnoses callable contracts in its first property pass, before representation limitations in the second. Thus this is an **introduced earlier refusal**, not a later stop reached by clearing G02. The same diagnostic is mirrored by a Boundary at `5339:5`. Pairing grades 7 gone, 2 changed, 0 new, 0 reached: one changed finding plus its Boundary mirror. The comparison’s real exit is 1.

The product test passed (0). Both utilities censuses passed (0) with complete checker-clean relation_report trailers: seven G02 sites before, zero after, zero remaining G02 mentions. The required zero-growth condition nevertheless fails.

Once the completed utilities pair established the blocker, the unrelated in-progress baseline parser census was deliberately terminated: real child exit -15 (SIGTERM), 437.408 seconds. Its partial stream and whole stderr are retained. This cancellation is not an unexplained census or box failure. Matching parser, both program and checker pairs, and the mutant were not run. No C02 landing was pushed. C03 was not created or run. The requested evidence publication and README amendment are complete.


Stock checks require 79 roots and zero diagnostics in both configurations; adamic types exits 0. The public API cmp is byte-for-byte. TestProduct uses -timeout 45m. Full entry censuses use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, LATENT_TIMES=/dev/stderr, exact absolute LATENT_ONLY_FILE and tsc.ts as root. Complete stdout/stderr and JSONL streams are losslessly retained. No probe or internal/ edit.

## Census comparisons and every reached stop


`c02/checkpoint-comparison/comparison.json`:
```json
{
  "utilities": {
    "before_events": 1309,
    "after_events": 1302,
    "target_before_distinct_sites": 7,
    "target_after_distinct_sites": 0,
    "after_target_mentions_all_kinds": 0,
    "grades": {
      "changed": 2,
      "gone": 7
    }
  }
}
```

- **changed**: `/workspace/family5-c02-tree/src/compiler/utilities.ts:5339:5`; unit `/workspace/family5-c02-tree/src/compiler/utilities.ts:5317:1`; kind `Boundary`; phase `lowering`. Full reason: failed statement; state rolled back; continued at next statement or declaration

/workspace/family5-c02-tree/src/compiler/utilities.ts:5339:11: Adamic 0.1 refuses a checked view with callable field getPositionOfLineAndCharacter; prove the callable body rather than asserting its signature

- **changed**: `/workspace/family5-c02-tree/src/compiler/utilities.ts:5339:11`; unit `/workspace/family5-c02-tree/src/compiler/utilities.ts:5317:1`; kind `Refused`; phase `lowering`. Full reason: a checked view with callable field getPositionOfLineAndCharacter

/workspace/family5-c02-tree/src/compiler/utilities.ts:5339:11: Adamic 0.1 refuses a checked view with callable field getPositionOfLineAndCharacter; prove the callable body rather than asserting its signature

## Publication

```json
{}
```

Each landing is one commit containing only its adaptation directory and one appended ledger line. C03 must be directly parented by C02. A missing receipt is not a pass. The old five-result gap was six skipped case bodies minus one failed hook counted as a failure; the alternating report accounts for it.
