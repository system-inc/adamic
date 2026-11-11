# Paired census: assertEachNode loop adaptation

Task 8n7eet3, site 2. Report only. No compiler or adaptation change is landed.

The full debug census demonstrates removal and reintroduction of the overload-1 refusal. In debug-only full mode, head has no finding in assertEachNode or its overloads. Restoring the old body restores the refusal. Every other normalized debug finding is unchanged. The entry pair has zero regressions, two cleared sites and one new finding in both raw and distinct views. Five already-blocked units change first stop. The entry census exposes an overload-4 refusal in assertEachNode; its four affected callers remain blocked.

## Trees and inputs

- Optimizer remote tip checked before work: `34437289896454f4076396d7bf913e70f6b9a826`, unchanged from the requested tip.
- Adaptation: `31e90645713d6ebb144bb393e58f198b35843b5a`.
- Scratch head: automatic, conflict-free merge; tree `98e8dce112394cf04623b04ed01c203b7e1f4061`.
- Parent compiler: `640d3ff43e92170b6777c8ad46657fb90846ff74` with its own apply.sh.
- TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Cohere pin, identical for both compilers: `5360753e19e4417675a90e92a1823bea66b13918`.

Setup used source /workspace/adamic-tools/env.sh, bash cloud/setup.sh, and npm ci --prefix stage3/api. Setup exited 0 in 650.117 seconds. Head and parent apply exited 0 in 709.830 and 714.348 seconds. Only src/compiler/debug.ts differs between the generated source trees; source-manifest.json records every src file hash.

The restoring mutant used the exact head census binary on an isolated byte-verified copy of the head adapted tree. The checked-in restoring-mutant.patch was applied without manual edits. Reverse application restored the original bytes; reapplication produced a debug.ts byte-identical to the parent tree. All other source files were verified identical before the patch.

The latent overlay is extended by stage3/meter/entry_overlay.py, so all runs use the resolved src/tsc/tsc.ts entry program. Debug runs set LATENT_FULL=1, LATENT_ONLY_FILE to their exact adapted debug.ts, and LATENT_ASSERT_NO_OUTPUT=1. Entry runs use first-error mode and the same no-output guard. Every completed debug header is checker-clean with zero diagnostics. These are measurement-only censuses, not backend executable builds.

## Full debug census

| Tree | Units | Recorded observations | Raw unique findings | assertEachNode refusal observations | Wall seconds | Peak RSS KiB | Exit |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| head | 81 | 10 | 10 | 0 | 31.072 | 921716 | 0 |
| mutant | 81 | 14 | 13 | 2 | 31.685 | 924124 | 0 |
| parent | 81 | 14 | 13 | 2 | 32.473 | 910552 | 0 |

Raw unique findings deduplicate (kind, location, reason, complete text) across attempts. Boundary records are retained in these full-mode totals. Each tree also records two unchanged findings in imported core.ts, included above. Debug.ts itself has 8, 11 and 11 raw unique findings for head, mutant and parent respectively.

### Every assertEachNode finding

Head: none, across the overloads at lines 256 through 260 and implementation at line 261.

#### mutant

- **Refused** at `src/compiler/debug.ts:257:86`. Attempt contexts: `src/compiler/debug.ts:261:5 (lowering)`, `src/compiler/debug.ts:262:5 (lowering)`.

> overload 1 of assertEachNode parameter test cannot be served by implementation parameter test

```text
src/compiler/debug.ts:257:86: Adamic 0.1 refuses overload 1 of assertEachNode parameter test cannot be served by implementation parameter test; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

- **Boundary** at `src/compiler/debug.ts:261:5`. Attempt contexts: `src/compiler/debug.ts:261:5 (lowering)`.

> function signature or prologue failed; continued at next declaration

```text
src/compiler/debug.ts:257:86: Adamic 0.1 refuses overload 1 of assertEachNode parameter test cannot be served by implementation parameter test; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

- **Boundary** at `src/compiler/debug.ts:262:5`. Attempt contexts: `src/compiler/debug.ts:262:5 (lowering)`.

> function signature or prologue failed; continued at next declaration

```text
src/compiler/debug.ts:257:86: Adamic 0.1 refuses overload 1 of assertEachNode parameter test cannot be served by implementation parameter test; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

#### parent

- **Refused** at `src/compiler/debug.ts:257:86`. Attempt contexts: `src/compiler/debug.ts:261:5 (lowering)`, `src/compiler/debug.ts:262:5 (lowering)`.

> overload 1 of assertEachNode parameter test cannot be served by implementation parameter test

```text
src/compiler/debug.ts:257:86: Adamic 0.1 refuses overload 1 of assertEachNode parameter test cannot be served by implementation parameter test; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

- **Boundary** at `src/compiler/debug.ts:261:5`. Attempt contexts: `src/compiler/debug.ts:261:5 (lowering)`.

> function signature or prologue failed; continued at next declaration

```text
src/compiler/debug.ts:257:86: Adamic 0.1 refuses overload 1 of assertEachNode parameter test cannot be served by implementation parameter test; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

- **Boundary** at `src/compiler/debug.ts:262:5`. Attempt contexts: `src/compiler/debug.ts:262:5 (lowering)`.

> function signature or prologue failed; continued at next declaration

```text
src/compiler/debug.ts:257:86: Adamic 0.1 refuses overload 1 of assertEachNode parameter test cannot be served by implementation parameter test; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

### Other debug.ts differences

None. Parent and mutant normalize to exactly the same finding set. Against either one, head removes only the assertEachNode refusal and its two signature/prologue Boundary records. It adds no finding. The line map uses exactly equal complete source lines; raw locations and diagnostic text remain in the original streams. debug-delta.json retains the complete normalized comparison.

## Tsc entry first-error comparison

Both entry censuses exited 0, are checker-clean, and contain all 81 source records and 4,562 top-level units. entry-reconciliation.json records the eligibility/status census.

The requested head includes the adaptation branch's main ancestry. Global entry deltas measure that whole scratch merge. The restoring-mutant debug run isolates the old-body substitution.

The pair tool is unmodified. Because the requested inputs differ in debug.ts, entry-parent-mapped.jsonl maps the parent root and only locations on exactly equal complete debug source lines into head coordinates. It rejects an edited-line location or any other src file difference. pair-location-map.json retains every mapping; source-manifest.json retains both original hashes. The original streams are retained separately. The pair tool then validates matching headers, unit identities and eligibility and normalizes property/element access chains using pinned stock TypeScript.

The entry census exposes this exact new finding at head debug.ts:259:98:

> overload 4 of assertEachNode parameter test cannot be served by implementation parameter test

It replaces overload 1 as the first stop in four already-blocked destructuring helpers. The fifth changed unit, createNameResolver, reaches an already-known unchecked-cast refusal after its destructuring stop clears. All five remain blocked. Thus the debug-only full-mode result does not establish finding-free entry-context lowering of assertEachNode.

The entry tables use canonical head coordinates. The original parent overload-1 location is debug.ts:257:86, mapped to head debug.ts:256:86 only for pairing.

| Measure | Raw | Distinct |
| --- | ---: | ---: |
| Cleared sites, no remaining finding | 2 | 2 |
| Changed reason at an already finding-bearing site | 0 | 0 |
| New findings, including exposed reasons | 1 | 1 |
| New sites | 1 | 1 |
| Units whose first stop changed | 5 | 5 |
| Regressions | 0 | 0 |
| All disappeared finding keys, cleared plus changed | 2 | 2 |

| Kind | Parent raw | Head raw | Parent distinct | Head distinct |
| --- | ---: | ---: | ---: | ---: |
| NotYet | 533 | 532 | 526 | 525 |
| Refused | 4455 | 4455 | 4260 | 4260 |

### Every moved unit, with old and new stops

A changed first stop in an already-blocked unit is listed here separately from regressions. A unit that previously had no lowering stop and now stops is a regression.

| Views | Unit | Parent stop | Head stop |
| --- | --- | --- | --- |
| raw_findings, distinct_reads | src/compiler/transformers/destructuring.ts:600:1 makeArrayBindingPattern | Refused at src/compiler/debug.ts:256:86: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | Refused at src/compiler/debug.ts:259:98: overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |
| raw_findings, distinct_reads | src/compiler/transformers/destructuring.ts:605:1 makeArrayAssignmentPattern | Refused at src/compiler/debug.ts:256:86: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | Refused at src/compiler/debug.ts:259:98: overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |
| raw_findings, distinct_reads | src/compiler/transformers/destructuring.ts:610:1 makeObjectBindingPattern | Refused at src/compiler/debug.ts:256:86: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | Refused at src/compiler/debug.ts:259:98: overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |
| raw_findings, distinct_reads | src/compiler/transformers/destructuring.ts:615:1 makeObjectAssignmentPattern | Refused at src/compiler/debug.ts:256:86: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | Refused at src/compiler/debug.ts:259:98: overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |
| raw_findings, distinct_reads | src/compiler/utilities.ts:11492:1 createNameResolver | NotYet at src/compiler/utilities.ts:11500:5: a destructured name that isn't plain | Refused at src/compiler/utilities.ts:11528:53: a cast the runtime can't check |

### Every regression

None. No previously lowered unit now stops, no previously finding-free refusal-scan site now finds, and no new panic is reported.

### Entry run resources

| Run | Wall seconds | Peak RSS KiB | Exit |
| --- | ---: | ---: | ---: |
| parent | 1736.557 | 4056876 | 0 |
| head | 1700.713 | 3941480 | 0 |

pair-report.md contains all per-reason counts and the raw/distinct summaries. pair-delta.json retains every cleared finding, reason transition, new finding, first-stop change and regression. Large artifacts have a .gz suffix.


## Timing and limits

Peak memory is Linux child ru_maxrss, in KiB, measured in a fresh Python timing wrapper for each command. Wall time is monotonic elapsed time around the census child. Head debug ran after setup. Parent debug and mutant debug overlapped the long head entry census; the two entry censuses overlap. These are individual observed times, not a controlled throughput benchmark. The portable timing wrapper replaced an unavailable GNU time before any apply started. A scratch Go workspace replacement error was corrected before successful builds; failed-attempt logs are retained.

Raw streams, timing receipts, exact source snapshots, binary hashes, merged-tree provenance, source hashes, and generated overlays are retained beside this report. Generated Go evidence is named .go.txt. Files over 1 MB are gzip-compressed. Checkpoints were pushed after each apply and completed run. The final report commit has 34437289 as its sole parent and touches only this census directory.
