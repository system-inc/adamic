# Slice 3 graph-adoption boundary census

Task ef727kr. Base: `d4693fc61d2d01fb4e1e993667ae5028da109cc3`. Head: scratch merge of `b7e1b7f075071b62e512a2cd4e154fe1c3aadb53` onto that base, tree `f9be03c06047c9a26288b08e5662e48cd2b83883`. The merge has no conflicts and touches no adaptation. Both compilers use one tree produced by the base apply.sh from pinned TypeScript 6.0.3. This report covers **conditional unit-local boundary stops**, not whole-program graph membership or acceptance.

## 1. Newly stopped units and regressions

**0 units newly stop at a boundary view refusal**, representing 0 raw findings and 0 distinct reads. **0 had no other finding at the base**, representing 0 raw findings and 0 distinct reads. Those are the unit-local regressions.

[NEW-STOPS.md](NEW-STOPS.md) quotes every new unit, native location, exact reason and source expression or predicate site, and marks whether the base had another finding. new-stopped-units.json retains ordinary and boundary baseline findings and all counterfactual reasons. No incomplete or discarded attempt contributes a count.

## 2. Causes

Each head view refusal is rerun on fresh snapshots with the deeper fields/index/Map/Set walk disabled, unconstrained-parameter refusal disabled, and both disabled. No head view refusal occurred, so no tsc counterfactual rerun was needed. The counterfactual harness tests the real boundary, including another view that may become its first stop. Unit observations and deduplicated raw/distinct findings are separate.

| Cause of new unit view stop | Unit observations | Raw findings | Distinct reads |
| --- | ---: | ---: | ---: |
| a deeper host walk | 0 | 0 | 0 |
| b unconstrained parameter | 0 | 0 | 0 |
| a and b both necessary | 0 | 0 | 0 |
| a and b independently sufficient | 0 | 0 | 0 |
| neither widening removes the unit view stop | 0 | 0 | 0 |

The joint categories are explicit rather than silently charged to one widening. A stop that survives both disabled is reported separately. The [method](METHOD.md) defines these counterfactuals and the normalization.

## 3. Cleared stops, moved stops and panics

At the unit level, 385 base view-stopped units have no head view refusal; 0 remain view-stopped with a changed first location or reason. A cleared view refusal may expose a different boundary category. Across all modeled boundary first stops, 385 clear, 0 move, and 0 are new. Every row is retained in cleared-units.json, moved-units.json, and all-boundary-*-units.json.

The pair tool reports raw/distinct cleared view findings **207/202**, moved view findings **0/0**, new view findings **0/0**, and units whose view first stop changes **385/385**. New findings at an already blocked unit are not counted as a newly stopped unit in section 1. [view-delta.md](view-delta.md) and view-delta.json.gz preserve the site-level pair.

New ordinary panics: **0**. Probe errors in included selections: **0**. ordinary-delta.json retains ordinary first-stop changes and panic details. Every serialized ordinary header/file/unit/finding byte matches across every paired selection, excluding relation-report telemetry (hashes in ordinary-byte-verification.json). All combined ordinary file/unit/finding records equal across the paired runs: **true**.

## 4. Totals

| Measure | Base | Merged head |
| --- | ---: | ---: |
| Observed unit results | 4061 | 4061 |
| Recorded view sources | 735 | 735 |
| Units with a boundary view stop | 385 | 0 |
| Raw view-stop findings at reported positions | 207 | 0 |
| Distinct view-stop reads at reported positions | 202 | 0 |
| Units with another modeled boundary stop | 6 | 6 |
| Units with any modeled boundary stop | 391 | 6 |
| Ordinary finding observations | 5711 | 5711 |
| Ordinary panics | 0 | 0 |
| Native viewHosted calls | 0 | 6874539 |

The base has 23 view-stopped unit observations with an explicitly marked unit fallback because its legacy diagnostic site is absent from the unit snapshot. Base raw/distinct totals count reported native or fallback positions; they are not exact expression identities for those observations. The head has no view-stop fallback. position-verification.json retains the precision counts. The new-stop counts in sections 1 and 2 are zero and need no fallback attribution.

Ordinary unit inventory by status: base `{"attempted": 4061, "deferred_generic_instantiation": 504}`; head `{"attempted": 4061, "deferred_generic_instantiation": 504}`. Every selected file and disjoint unit range completed on both sides. The entry graph has 81 files; source-manifest.json hashes all 747 adapted source files.

## Reachability proof and limits

The stock meter never calls findCycles and therefore skips graphAdoptionBoundary. The scratch-only unit-end observer calls the real boundary on a fresh finder and the unit's snapshot, without selecting graph members. Native entry counters record **6874539 calls to the changed viewHosted** in the included tsc runs. Before measuring tsc, the native host-walk test passed and the identity control made 14 native walk calls. Forced-refusal controls produced a stop on both sides; real/mutant fixture ordinary output matches runs with observation disabled. final-control-results.json and raw/host-walk-tests.* retain the proof.

The pinned base has the blanket checked-view adapter boundary. The merged head replaces it with host-sensitive checks and introduces viewHosted. This pair is broader than a shallow/deep-only comparison, which is why already blocked units and moved first reasons are kept separate.

The first attempted full-boundary seam encountered a base diagnostic-only Where(nil) panic on a dynamic-shape stop with no graphDynamicSite in the unit snapshot. That attempt was stopped and excluded. A stack-traced isolated reproducer identifies the exact call. The final scratch formatter handles only missing diagnostic locations, preserves the real refusal and reason, and marks a unit fallback. The repaired unit reports its dynamic-shape refusal without a probe panic. Native head view-stop locations and source text are unchanged. See [METHOD.md](METHOD.md) for the exact scope and interpretation.

The census uses first-error top-level independent units. Callee bodies reached by them are included; standalone generics without a call witness are deferred. Views before an ordinary stop remain latent observations in already blocked units. This does not model whole-program membership, earlier whole-program cycle proofs, or unselected graph-pair membership. It does not establish native tsc acceptance.

[RUNS.md](RUNS.md) lists each child's own exit code, elapsed time and peak RSS. Every completed process is below 30 minutes; the longest recorded child took 651.504 seconds. Census children have a 1490-second cap. Raw streams over 1 MB are gzipped. Production Go evidence uses .go.txt names. The measurement seams remain outside the commit; provenance.json records their hashes. The commit changes only the census directory; it contains no production or adaptation edits.
