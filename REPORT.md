# Function-constructor allocator census

Base `23cc416559634af639e51371b09a07f6fdb4eb14` and head `e8a7b0c36287126cd934b82051a5c7f1018b9b41`.
Both compiler revisions measured the same base-adapted TypeScript v6.0.3 tree from upstream `050880ce59e30b356b686bd3144efe24f875ebc8`.
Head apply restored the same keyed product from cache. The two apply outputs matched outside Git metadata. The SHA-256 of the sorted adapted-file checksum list is `ba5a38cbd633806781763b8fea82d38dad45648170e8d4fc2c54a7f28a4f4688`.

## 1. Node and Token getter stops

The pinned adaptation places the Node getter at `src/compiler/utilities.ts:8555` and the Token getter at `src/compiler/utilities.ts:8556`. The brief refers to the same getters at 8553 and 8555.

**Node’s stop stays, and Token does not become the lowering stop.** Both findings belong to the `objectAllocator` variable-statement unit at `src/compiler/utilities.ts:8554:1`.

| Getter | Position | Base | Head |
| --- | --- | --- | --- |
| `getNodeConstructor` | `src/compiler/utilities.ts:8555:82` | Refused in the refusal scan and at the first lowering stop | Same |
| `getTokenConstructor` | `src/compiler/utilities.ts:8556:32` | Refused in the refusal scan | Same |

At each position, the full diagnostic wording is:

> Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)

## 2. Other finding changes

None. No new, gone, changed, or reached findings.
All 3,254 finding events agree in stream order, including unit, position, kind, phase, reason, and message.
The repository pairing tool counted 2,060 unique Refused findings and 430 unique NotYet findings on each side, with zero unit first-stop changes.

## 3. Previously lowered units and new panics

0 previously lowered units stop at head, and there are 0 new panics.
Body lowering completed for 3,274 units at base and 3,274 at head. The scoped tail recorded 2 stops at base and 2 at head.

## 4. First stops across 79 roots

0 of 79 roots changed their production first stop.

The meter also returned the same next stop after its runtime-readiness adjustment for `src/compiler/debug.ts`, `src/compiler/diagnosticInformationMap.generated.ts`, `src/compiler/types.ts`.

## Measurement

The entry overlay, production census, and scan-all were built separately from each pinned compiler revision. The entry pass used the meter entry overlay rooted at `src/tsc/tsc.ts` with first-error and `LATENT_TAIL=1`. The entry and per-root production census both used `CENSUS_MODULE_FORMAT=CommonJSBundle`.
Finding signatures use kind, phase, reason, and message with source positions stripped. Comparisons retain unit and site so identical wording at a later stop remains visible. A reached finding is a later stop in a unit that already stopped at base.
The scoped entry ledger records the head’s whole-program `checkFunctionConstructorRoads` step as unrun in scoped mode. The per-root pass uses production lowering.
The raw per-root streams also include the three non-source rows and the meter’s whole-program row.
Each pass was pushed before the next started. The [gzipped streams](evidence/) retain the original positions and diagnostic text.
