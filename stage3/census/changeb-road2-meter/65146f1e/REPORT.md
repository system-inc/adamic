# Final road 2 (`65146f1e`) against main

**Zero main-to-final lowering regressions. Both earlier regressions are gone.** The two first-error censuses completed all 79 files concurrently on the same box and the unchanged adapted tree. This is **measured on a checker-rejected program**: a no-stop attempt is not a claim of ordinary accepted-program compilation or native output.

| Metric | A: main | B: final road 2 |
| --- | ---: | ---: |
| Distinct NotYet | 652 | 648 |
| Distinct Refused | 2,115 | 2,115 |
| Errors / panics | 0 / 0 | 0 / 0 |
| Attempted units without lowering finding | 2,930 | 2,930 |
| Timed attempted units | 4,017 | 4,017 |
| Unit lowering wall seconds (sum) | 512.228254 | 509.360610 |
| Unit lowering CPU seconds (sum) | 543.521684 | 549.259858 |
| Whole-run wall seconds | 928.102 | 922.344 |
| Peak RSS (KiB) | 6149568 | 5916992 |

No A-successful unit is absent or stopped under B; the [regression list is empty](lowering-regressions.json). Specifically, `computeLineOfPosition` (`scanner.ts:520:1`) and `indexOfNode` (`utilities.ts:1351:1`) are attempted without lowering-phase findings on **both** sides ([explicit check](prior-regressions-check.json)). There are six changed first stops, all in already-stopped units, and no stopped-to-success changes. The [complete first-stop table](unit-outcome-changes.md) lists each. Four move to the already-counted getter stop at `sourcemap.ts:474:9`; two move to the new predicate reasons below.

The [changed-reason table](changed-reasons.md) lists all eight changed rows, A/B counts and up to five example locations per side. Every changed row mentions a **predicate**. The new **LateBoundDeclaration** row says “a checked predicate target LateBoundDeclaration that narrows ClassElement & { name: ComputedPropertyName; }, which isn't its supertype” (0 → 1, `checker.ts:7626:37`). The new **LeftHandSideExpression** row says “a checked predicate target LeftHandSideExpression over a generic argument TIn & ({} | null), which narrows to an intersection no view covers” (0 → 1, `core.ts:1784:32`), explicitly mentioning a **view**. Six “checked predicate overload target …” rows go 1 → 0: **ClassElement, JSDocPropertyLikeTag, LateBoundDeclaration, LeftHandSideExpression, LiteralExpression, TypeElement**. The [complete view/predicate table](view-predicate-reasons.md) includes unchanged rows. No observed reason literally says “branded supertype”; no read/write-through-view row changes.

Observed total unit wall time is **-2.867644 seconds (-0.560%, 0.99440×)**. CPU is **+5.738174 seconds (+1.056%)**. Whole-run wall time is -5.758 seconds. The [20 largest B−A differences](timing-top20.md) show absolute seconds and ratios; [all units](timing-all.md), [JSON](unit-times.json), and both raw timing streams retain full coverage. Unit timing includes registration, lowering and recovery, excluding loading/checking and refusal scanning.

Noise floor: 1215 identical-outcome import/interface/type-alias units whose lowering does not run a runtime branded-predicate call. Median signed B−A is -0.101 ms; median absolute difference **2.179 ms**, p90 absolute **14.480 ms**. Median B/A is 0.9979; p10–p90 **0.8947–1.1166** ([baseline](noise-baseline.json)). The identical main binary/tree took 536.176 unit wall seconds in the first pair versus 512.228 now (-4.47%). This pair does not establish a causal speedup; GC, cache and scheduling variation remain.

**Generic `core.some` call-site inventory: 1 source site out of 305 resolved calls.** Stock TypeScript **6.0.3** resolves all calls against the three `core.ts` overload/implementation declarations. The only generic first argument is `array`, type **`readonly T[]`**, at **`src/compiler/core.ts:1203:10`** (inside `binarySearchKey`); it selects the one-argument type-predicate overload. There are zero generic-first-argument calls selecting the two-argument boolean overload. This is one source site, not one concrete instantiation or one affected census unit; it is the site behind the previous two concrete regressions. A generic source-site inventory is not itself a claim that the final head stops there.

The [site table](some-generic-sites.md), [full generic evidence](some-generic-sites.json), [all 305 calls](some-all-sites.json) and [API script](some-generic-sites.cjs) record file:line:column, argument text/type, resolved overload and generic components. “Generic” follows the compiler’s `mentionsTypeParameter`: instantiable flags through union/intersection members, alias arguments and reference arguments; the sole match contains an actual `T` type parameter. The script uses the adapted compiler tsconfig, stock parser/checker and the pinned API cache’s Node type declarations; it writes no source or emitted output. It scans all 79 source files regardless of checker eligibility or first-error stops. Stock diagnostics: zero option/syntax errors; one unrelated TS2307 at `sys.ts:1598:69` for missing `source-map-support`, retained in JSON. No `some` call has an unresolved signature. Inventory work ran before and after, never during, the pair.

Both runs started at 2026-11-01 14:52:39 UTC, within 0.3 ms; B ended at 15:08:02 UTC and A at 15:08:07 UTC. Neither hit the 5400-second bound; both exited 0. Box `6923d492b4a0`: 5 CPUs (`nproc`), cgroup quota 400000/100000 (4 CPU equivalents), 18,882,662,400 bytes RAM, no swap ([box](box.json), [A resources](A-resources.json), [B resources](B-resources.json)).

Pins and reproduction:

- A: `141ad1d244d6a2d9c5403858086d505dfdef5fd5`; existing main census binary reused.
- B: **`65146f1eff049a8ce9a35fe67c842f79bbe1407c`**, resolved after fetching `optimizer/branded-supertype-road2` with `git rev-parse 65146f1e^{commit}`.
- TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`; the tree built once from A for the previous pair was reused. All 79 hashes remain unchanged. Same checker header and all 4,543 unit eligibility records (4,017 attempted, 498 deferred, 28 skipped) on both sides. Main’s raw ledger is byte-identical to the previous pair.
- A binary SHA-256: `10371fb48fc2df1b6f414584ed0ec2c0f2320be156b2c11282c97b85561a4382`.
- B binary SHA-256: `3d9f5b617d133e6aec47d41d3ddefac3416731f5e72dd0a8249ef618c100c91e`.
- Meter/recorder code is identical at A and B; no backport. No compiler files or anything under `internal/` was changed. [Exact commands](commands.md), [provenance](provenance.json), [source hashes](input-sha256.json), [runner](run-censuses.py), [analysis](analyze.py) and [logs](logs/).
- [A ledger](A.jsonl.gz) and [B ledger](B.jsonl.gz), [A times](A-times.jsonl.gz) and [B times](B-times.jsonl.gz) round-trip to the raw streams. The requested [census delta](census-delta.json) and existing meter’s [lowering_summary coverage/count validator](validation.json) agree: **2,767 → 2,763** distinct NotYet/Refused sites using `(kind, where, reason, text)`. Three SkippedDependency sites on each side are excluded. `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1` throughout.
