# Paired census: assertEachNode loop adaptation

Task 8n7eet3, site 2. Report only. No compiler or adaptation change is landed.

The full debug census demonstrates removal and reintroduction of the overload-1 refusal. Head has no finding in assertEachNode or its overloads. Restoring the old body restores the refusal. Every other normalized debug finding is unchanged. The tsc entry comparison is still running.

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

Pending: both censuses are active. No regression conclusion is claimed yet.

## Timing and limits

Peak memory is Linux child ru_maxrss, in KiB, measured in a fresh Python timing wrapper for each command. Wall time is monotonic elapsed time around the census child. Head debug ran after setup. Parent debug and mutant debug overlapped the long head entry census; the two entry censuses overlap. These are individual observed times, not a controlled throughput benchmark. The portable timing wrapper replaced an unavailable GNU time before any apply started. A scratch Go workspace replacement error was corrected before successful builds; failed-attempt logs are retained.

Raw streams, timing receipts, exact source snapshots, binary hashes, merged-tree provenance, source hashes, and generated overlays are retained beside this report. Generated Go evidence is named .go.txt. Files over 1 MB are gzip-compressed. Checkpoints were pushed after each apply and completed run. The final report commit will have 34437289 as its sole parent and touch only this census directory.
