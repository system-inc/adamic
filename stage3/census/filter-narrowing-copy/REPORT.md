# Narrowing filter copy stop report — #ev14ch9

Base: `49f1d643064a9231f7f343d8bbd02ba718866e6a`.
TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3).
Freeze evidence: `df886e406b521bacf030d7e8ca76d6e828fed249`.

## Gate result

The two out-of-scope readonly narrowing overloads both refuse in
`lowering.declareModule` for the same generic result mismatch:

- Original overload 3, core.ts:265:1: `readonly U[]` cannot be served by
  `readonly T[] | undefined`.
- Original overload 7, core.ts:273:1: `readonly U[] | undefined` cannot be
  served by `readonly T[] | undefined`.

Each control removes preceding signatures only in a scratch copy, retaining
blank lines so original source coordinates survive. The resulting ordinal in
the diagnostic is 1; the original ordinals are 3 and 7. The implementation
is the complete unchanged adapted filter body, including baseline indexed-read
non-null assertions. No landing filter edit is proposed. The original-owner
probe stops at core.ts:261:1 on overload 1, with or without callers: declaration
judgment does not depend on whether the signature has callers.

An initial isolated probe used upstream's unadapted indexed reads and failed
Load with three TS2345 diagnostics. Those are not lowering evidence. The final
controls use actual adapted indexed reads and load successfully. Earlier controls
removing only mutable narrowing signatures first encounter mutable boolean
overload 2; they do not establish anything about later readonly signatures.
The separately isolated original-3 and original-7 controls do.

The scratch helper probe is checker-clean and its guarded full latent run has
zero findings, with concrete instantiation witnesses for both plain helpers,
including the undefined-array call. This reduced probe alone does not establish
full core.ts coverage, general freshness certification, backend acceptance, or
any of the unrun lane/declaration/mutant checks.

The authorized change stops here under condition 6. It is not widened to readonly
narrowing or boolean overloads. No adapter, ledger acceptance line or landing
branch is created.

## Remaining proof obligations (not claimed)

A complete accepted copy requires the premise that no reader compares a
narrowing filter's result with its argument by identity, directly or through a
slot. The freeze run preserves identity and cannot witness that premise. Its
242,183 additional allocations (232,451 at checker.ts:16764) remain the measured
price from df886e40, not a new copying-lane measurement. Should #4yrezf5 is the
specified later erasure task. There is no successful-copy claim for the ledger.

The four sites were read, not run: checker.ts:9417 and :9419 create temporary
arrays whose filtered identifiers flow through flatMap/map into export specifiers;
transformers/ts.ts:1088 reads a skipWhile result and adds elements into a separate
modifiersArray; services/callHierarchy.ts:482 groups a temporary reference-result
array and maps entries to spans. The inspected paths do not compare argument/result
array identity or write through either handle after the filter call. This hand
review does not replace a copying lane or prove arbitrary transitive aliases.

Not run after the stop: copying lane (106,366/1), alias-restoring mutant, Node
runtime witnesses from a built adapted compiler, stock 79-root diagnostic delta,
`adamic types`, public declaration builds/comparison, adapter contract/replay,
restoring adapter mutant, and TestProduct_Stage3AdaptedTree. The reduced probe's
measurement success does not stand in for those checks.

## Completed measurements

The unchanged `stage3/apply.sh --build-product` exits 0. The complete log is
[evidence/apply.log.gz](evidence/apply.log.gz). All logs and trial sources are
evidence only; the trial is not an accepted adaptation.

Resolved-signature census on 713 src roots:

| Original overload | Before calls | Trial calls |
| --- | ---: | ---: |
| 1 | 5 | 0 |
| 2 | 60 | 60 |
| 3 (readonly narrowing) | 14 | 14 |
| 4 | 41 | 41 |
| 5 | 11 | 0 |
| 6 | 13 | 13 |
| 7 (optional readonly narrowing) | 9 | 9 |
| 8 | 4 | 4 |
| Total | 157 | 141 |

Thus **two readonly narrowing declarations, with 23 calls**, retain the same
result-mismatch stop. No proposal to cover their calls is made. Core's existing
261:1 stop also remains: the original complete filter owner without any callers
still fails its declaration census at overload 1. The trial retains every filter
declaration and implementation byte; [owner identity](evidence/filter-owner-identity.json)
records its hash. Removing callers cannot remove a declaration-time judgment.

The full single-file core census uses `LATENT_FULL=1`,
`LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_ONLY_FILE` set to the exact core.ts path,
and `CENSUS_MODULE_FORMAT=CommonJSBundle`, with the complete `src/compiler`
directory loaded. Before and trial both have **14 findings**: **0 new, 0 gone,
0 changed** after normalizing the tree prefix only. Every finding, in full,
appears in [comparison.json](evidence/comparison.json), with raw streams in the
before/after-core-latent.jsonl.gz files. Both children exit 0. This mode defers
filter's generic units and does not establish absence of its declaration stop.

The two newly added helper units are also deferred and **not instantiated** in
the core-only trial. Their lack of findings is a coverage gap, not proof of their
bodies. The separately called reduced helper fixture has concrete instantiation
witnesses and zero findings, but does not replace whole-core instantiation or
name a general result/freshness certificate. Certificate A (#5fws6q8) is not
claimed either proved or disproved by that gap. No cast is introduced.

The 141 remaining call records have exactly the same file, overload and source
call text; added helper imports shift caller line coordinates. The requested
whole-transpiled-JavaScript comparison has not been performed, so no emitted-JS
identity claim is made. Only the trial preparation script changes the frozen
16 expressions and imports and appends the two helpers. Trial helpers use
`filterNarrowed` and `filterNarrowedOptional`: the `filter` family name describes
the operation, `Narrowed` identifies its predicate result, and `Optional`
identifies the undefined-array contract. Both declarations are plain, not
overloaded, and marked `@internal`.

## Reproduction and provenance

Setup used the requested detached worktree at the base, recursive shallow
submodules, `npm ci` in stage3/api, and the v6.0.3 upstream clone. Go 1.27.1,
Node v24.19.0. TMPDIR, GOTMPDIR and GOCACHE point under task scratch.
The measurement generators are unchanged main's
`stage3/census/latent/make_overlay.py` and
`stage3/census/scan-all/make_overlay.py`; their build logs are retained.

`evidence/count.cjs TREE OUTPUT` counts calls by stock tsc resolved signatures.
`evidence/prepare.py` records the scratch-only trial transformation. Its paths
point to this run's scratch and can be relocated for a rerun. The four service
calls contain repeated call text, so preparation requires the frozen text's
exact multiplicity rather than incorrectly requiring uniqueness per expression.
An initial preparation attempt failed on this multiplicity check; the final
script and comparison use the corrected preparation.

For an isolated control, run the scan binary with one corresponding
`evidence/isolated/overload-N/core.ts` file. Probe exit 0 means measurement
completed, not compilation acceptance: its JSON reports `Completed: false`
and the declaration stop. The original owner control has no callers and
reproduces 261:1. Removing signatures in isolated controls is measurement only;
no deletion is proposed for filter.

Ownership search evidence includes the 16 AST owner chains and the full search
of all owners.json, targets.json and adapt.cjs files. The transformer owner is
pinned whole by receiver-transformers; checker and service import owners are
pinned by core-multimap. No pins or adaptation directories were changed.

After the readonly stop, full caller-file censuses, scan-all of the entire
adapted import closures, JavaScript comparison and all remaining acceptance
checks were not run. This is the required stop report, not a landed copy.

## Measurement 1: all original filter overloads (2026-11-17)

Report only on base `49f1d643064a9231f7f343d8bbd02ba718866e6a`, using the unchanged completed adapted baseline from the first investigation. Each single-overload control keeps **only that signature and the complete unchanged adapted implementation**. Other lines are blanked, so original coordinates survive. The implementation is at line 277. There are no calls in the isolated files. Original ordinals below are distinct from the diagnostic ordinal, which is 1 in each single-overload control.

Each control runs both the requested full single-file latent census (`LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=<exact control core.ts> CENSUS_MODULE_FORMAT=CommonJSBundle`) and the unchanged scan-all measurement probe. **The declareModule answer comes from scan-all**, since all ten latent runs defer generic declarations and therefore do not perform that declaration judgment. Every child exits 0; this is measurement completion, not compiler acceptance.

| Original overload | Original line | Exact signature | Resolved calls (713 roots) | declareModule refuses | Exact reason words |
| --- | ---: | --- | ---: | --- | --- |
| 1 | 261 | `export function filter<T, U extends T>(array: T[], f: (x: T) => x is U): U[];` | 5 | yes | `overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined` |
| 2 | 263 | `export function filter<T>(array: T[], f: (x: T) => boolean): T[];` | 60 | yes | `overload 1 of filter result T[] cannot be served by implementation result readonly T[] \| undefined` |
| 3 | 265 | `export function filter<T, U extends T>(array: readonly T[], f: (x: T) => x is U): readonly U[];` | 14 | yes | `overload 1 of filter result readonly U[] cannot be served by implementation result readonly T[] \| undefined` |
| 4 | 267 | `export function filter<T>(array: readonly T[], f: (x: T) => boolean): readonly T[];` | 41 | no | `none observed; scan completed` |
| 5 | 269 | `export function filter<T, U extends T>(array: T[] \| undefined, f: (x: T) => x is U): U[] \| undefined;` | 11 | yes | `overload 1 of filter result U[] \| undefined cannot be served by implementation result readonly T[] \| undefined` |
| 6 | 271 | `export function filter<T>(array: T[] \| undefined, f: (x: T) => boolean): T[] \| undefined;` | 13 | yes | `overload 1 of filter result T[] \| undefined cannot be served by implementation result readonly T[] \| undefined` |
| 7 | 273 | `export function filter<T, U extends T>(array: readonly T[] \| undefined, f: (x: T) => x is U): readonly U[] \| undefined;` | 9 | yes | `overload 1 of filter result readonly U[] \| undefined cannot be served by implementation result readonly T[] \| undefined` |
| 8 | 275 | `export function filter<T>(array: readonly T[] \| undefined, f: (x: T) => boolean): readonly T[] \| undefined;` | 4 | no | `none observed; scan completed` |

| Group control | Original overloads retained | declareModule result |
| --- | --- | --- |
| boolean | 2, 4, 6, 8 | `263:1`: `overload 1 of filter result T[] cannot be served by implementation result readonly T[] \| undefined` |
| readonly | 3, 4, 7, 8 | `265:1`: `overload 1 of filter result readonly U[] cannot be served by implementation result readonly T[] \| undefined` |

**Answer:** every mutable-result overload (1, 2, 5, 6) and both readonly narrowing overloads (3, 7) refuse in these controls. The refusal is not limited to narrowing callbacks: boolean overloads 2 and 6 also refuse. Only readonly boolean overloads 4 and 8 escape this declaration refusal in isolation. The group controls report their first declaration stop only; their later overloads are judged individually by the eight single-overload controls. Overload 4 does exclude undefined from the implementation result, but that presence-only narrowing is admitted by the existing `censusNullableOverloadResult` rule; the broader statement that every formally narrower result refuses would therefore be false.

The full latent census records **0 findings for each of the ten controls**, with all retained signatures and the implementation marked `deferred_generic_instantiation`, without instantiation witnesses. This is a coverage gap, not evidence that the six refused declarations lower. Scan-all's `Completed: true` for controls 4 and 8 is measurement traversal only, not backend acceptance.

Reproduce with `python3 overload-measurement/run.py BASELINE_TREE SCRATCH SCAN_BINARY LATENT_BINARY` from this report's directory, using the unchanged main scan-all and latent overlays built earlier. [Raw controls and observations](overload-measurement/evidence/) include exact sources, stdout, stderr, exit codes, full latent streams, and combined results. The resolved counts are the existing completed 713-root census in `evidence/before-count.json.gz`.

No landing adaptation, ledger line or internal source change. This measurement is pushed before measurement 2 begins.

## Measurement 2: scoped readonly narrowing freeze count (2026-11-17)

Reports only on base `49f1d643064a9231f7f343d8bbd02ba718866e6a`. Measurement 1 was pushed at `2f473173f` before this measurement began. The 713-root stock checker selects exactly 23 original core filter calls: overload 3 has 14 and overload 7 has 9. The other 134 call expressions and filter itself remain unchanged. All 157 calls retain their resolved overload, argument type and result type after wrapping.

The entry/completion wrapper, V8 pause-on-all-exceptions observer, WeakMap array provenance, caught/escaping exception handling, VM bridge and control are from `df886e40:stage3/census/filter-freeze`. The observer, control, resolver and aggregator are byte-identical. Scope changes affect the selected declaration coordinates (265 and 273), expected set size (23), remaining-call count (134), and report-tool paths only; tool-provenance.json records both hashes. Each argument is evaluated once; a result equal to it is frozen only if it is an array. No copy, proxy or substitute result is used.

The same control imports the actual built adapted compiler filter. It records push, splice, sort, index assignment and length assignment through **both argument and result handles**: ten TypeErrors, all caught and recorded by the observer, no observer error or unclassified candidate. The control also proves the dropped-element result and input stay mutable, undefined counts without freezing, and a throwing predicate increments entry without completion. See [control validation](../filter-readonly-freeze/evidence/control-validation.json).

Self compilation uses the instrumented `built/local/tsc.js -b src/compiler src/tsc --force --pretty false`, after `npm ci` and the complete `npm run build` exited 0. Canonical self output is self-final-counts, collected after bootstrap completed. An earlier self run overlapping the final build tasks is retained separately and excluded. Bootstrap entries are separately counted and excluded.

The whole unfiltered lane uses the original run.py executed in memory with the original after-apply instrumentation hook and `FILTER_FREEZE_WORKERS=4`, matching df886e40's canonical run on this four-CPU environment. All runners, test selection and 40,000 ms timeout remain unchanged. The lane's eight-worker guard is untouched; its actual verdict is retained, never relabeled as a pass.

Table cells: **calls / returns identical to argument / array identities**. Calls count entries, not only successful returns. Array identities count returning invocations, not distinct arrays. Writes count post-call writes identified by observer provenance; every site's zero is retained.

| Original site | Overload | Self | Whole lane | Writes self / lane |
| --- | ---: | ---: | ---: | ---: |
| `src/compiler/utilitiesPublic.ts:1030:16` | 7 | 0 / 0 / 0 | 4,080 / 3,281 / 3,281 | 0 / 0 |
| `src/compiler/utilitiesPublic.ts:1036:16` | 7 | 0 / 0 / 0 | 6 / 6 / 6 | 0 / 0 |
| `src/compiler/utilities.ts:4691:20` | 7 | 0 / 0 / 0 | 1,725 / 229 / 160 | 0 / 0 |
| `src/compiler/utilities.ts:8019:12` | 3 | 0 / 0 / 0 | 2,893 / 2,225 / 2,225 | 0 / 0 |
| `src/compiler/commandLineParser.ts:3188:54` | 7 | 2 / 2 / 0 | 9,902 / 9,900 / 1,197 | 0 / 0 |
| `src/compiler/checker.ts:3138:58` | 3 | 0 / 0 / 0 | 80 / 0 / 0 | 0 / 0 |
| `src/compiler/checker.ts:11961:38` | 3 | 0 / 0 / 0 | 148 / 0 / 0 | 0 / 0 |
| `src/compiler/checker.ts:32639:25` | 3 | 596 / 24 / 24 | 2,516 / 371 / 371 | 0 / 0 |
| `src/compiler/checker.ts:46866:33` | 3 | 14,330 / 14,312 / 14,312 | 816,900 / 814,917 / 814,917 | 0 / 0 |
| `src/compiler/transformers/utilities.ts:603:12` | 3 | 0 / 0 / 0 | 11,433 / 1,771 / 1,771 | 0 / 0 |
| `src/compiler/transformers/ts.ts:1053:13` | 3 | 0 / 0 / 0 | 1,892 / 1,259 / 1,259 | 0 / 0 |
| `src/compiler/transformers/ts.ts:1086:38` | 7 | 0 / 0 / 0 | 88 / 85 / 85 | 0 / 0 |
| `src/compiler/transformers/ts.ts:1100:42` | 7 | 0 / 0 / 0 | 318 / 314 / 266 | 0 / 0 |
| `src/compiler/transformers/ts.ts:1102:42` | 7 | 0 / 0 / 0 | 318 / 48 / 0 | 0 / 0 |
| `src/compiler/transformers/ts.ts:1383:13` | 3 | 0 / 0 / 0 | 3,417 / 1,982 / 1,982 | 0 / 0 |
| `src/compiler/transformers/classFields.ts:852:25` | 7 | 0 / 0 / 0 | 422 / 205 / 63 | 0 / 0 |
| `src/compiler/transformers/classFields.ts:1707:16` | 3 | 0 / 0 / 0 | 21,883 / 104 / 104 | 0 / 0 |
| `src/services/utilities.ts:2616:38` | 3 | 0 / 0 / 0 | 667 / 82 / 82 | 0 / 0 |
| `src/services/codefixes/convertFunctionToEs6Class.ts:314:39` | 7 | 0 / 0 / 0 | 69 / 69 / 5 | 0 / 3 |
| `src/services/codefixes/convertToEsModule.ts:137:29` | 3 | 0 / 0 / 0 | 34 / 1 / 1 | 0 / 0 |
| `src/services/goToDefinition.ts:180:30` | 3 | 0 / 0 / 0 | 3 / 0 / 0 | 0 / 0 |
| `src/testRunner/unittests/services/organizeImports.ts:1133:25` | 3 | 0 / 0 / 0 | 30 / 30 / 30 | 0 / 0 |
| `src/testRunner/unittests/services/organizeImports.ts:1140:25` | 3 | 0 / 0 / 0 | 16 / 16 / 16 | 0 / 0 |

**self:** totals `{"calls": 14928, "completed": 14928, "frozen_arrays": 14336, "identity": 14338, "undefined_identity": 2}`; freeze writes 0, unrelated candidates 0, unclassified candidates 0, observer errors 0.

Sites with **no calls** in self: `src/compiler/utilitiesPublic.ts:1030:16`, `src/compiler/utilitiesPublic.ts:1036:16`, `src/compiler/utilities.ts:4691:20`, `src/compiler/utilities.ts:8019:12`, `src/compiler/checker.ts:3138:58`, `src/compiler/checker.ts:11961:38`, `src/compiler/transformers/utilities.ts:603:12`, `src/compiler/transformers/ts.ts:1053:13`, `src/compiler/transformers/ts.ts:1086:38`, `src/compiler/transformers/ts.ts:1100:42`, `src/compiler/transformers/ts.ts:1102:42`, `src/compiler/transformers/ts.ts:1383:13`, `src/compiler/transformers/classFields.ts:852:25`, `src/compiler/transformers/classFields.ts:1707:16`, `src/services/utilities.ts:2616:38`, `src/services/codefixes/convertFunctionToEs6Class.ts:314:39`, `src/services/codefixes/convertToEsModule.ts:137:29`, `src/services/goToDefinition.ts:180:30`, `src/testRunner/unittests/services/organizeImports.ts:1133:25`, `src/testRunner/unittests/services/organizeImports.ts:1140:25`.

Called sites with **no exercised array identity** in self: `src/compiler/commandLineParser.ts:3188:54`.

**lane:** totals `{"calls": 878840, "completed": 878840, "frozen_arrays": 827821, "identity": 836895, "undefined_identity": 9074}`; freeze writes 3, unrelated candidates 6, unclassified candidates 3, observer errors 0.

Sites with **no calls** in lane: none.

Called sites with **no exercised array identity** in lane: `src/compiler/checker.ts:3138:58`, `src/compiler/checker.ts:11961:38`, `src/compiler/transformers/ts.ts:1102:42`, `src/services/goToDefinition.ts:180:30`.

**bootstrap:** totals `{"calls": 0, "completed": 0, "frozen_arrays": 0, "identity": 0, "undefined_identity": 0}`; freeze writes 0, unrelated candidates 0, unclassified candidates 0, observer errors 0.

Actual lane verdict:

```text
FAIL stage3 landing lane: test counts: expected {'passing': 106366, 'failing': 1, 'pending': 0}, observed {'passing': 106363, 'failing': 4, 'pending': 0} [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=3, other=0]
```

The complete lane records **3 provenance-backed post-call write attempts**. They are array metadata writes to `__pos` by `setPos` in `src/services/textChanges.ts:191:16`, via `onBeforeEmitNodeArray` at :1434:13. The filtered modifier array is reused as the new class declaration modifiers and reaches the printer. These are not element/index/length writes. Blocked writes can abort the remainder of a failing test body, so these counts describe execution under freezing, not an unfrozen workload.

| Write site | Writer | Test case |
| --- | --- | --- |
| `src/services/codefixes/convertFunctionToEs6Class.ts:314:39` | `setPos` | fourslash tests tests/cases/fourslash/convertFunctionToEs6Class_asyncMethods.ts fourslash test convertFunctionToEs6Class_asyncMethods.ts runs correctly |
| `src/services/codefixes/convertFunctionToEs6Class.ts:314:39` | `setPos` | fourslash tests tests/cases/fourslash/convertFunctionToEs6Class_exportModifier2.ts fourslash test convertFunctionToEs6Class_exportModifier2.ts runs correctly |
| `src/services/codefixes/convertFunctionToEs6Class.ts:314:39` | `setPos` | fourslash tests tests/cases/fourslash/convertFunctionToEs6Class_exportModifier1.ts fourslash test convertFunctionToEs6Class_exportModifier1.ts runs correctly |

Exception classification after source review: `{"freeze-write": 3, "propagated-freeze-error": 3, "unrelated-explicit-decorator-error": 6}`; unresolved candidates: 0. The original byte-identical aggregator leaves provenance-free rethrows marked unclassified; classify.py records them separately only when process, test title and complete stack match a provenance-backed write and the top frame is the source-proven `runCode` catch/rethrow. Raw observations and original classifications are retained. See [classification](../filter-readonly-freeze/evidence/exception-classification.json) and [source proof](../filter-readonly-freeze/evidence/early-write-observation.json).

Whole-lane oracle counts: `{"failing": 4, "passing": 106363, "pending": 0}`; phase exits: `{"build": 0, "install": 0, "tests": 1}`. Workers: 4; runners: all; tests filter: None. The full [guard report](../filter-readonly-freeze/evidence/lane-report.json) also retains `workers: expected 8, observed 4`.

**Identity remains a premise.** Direct comparisons with the argument: 0. The reused static lexical/local-reference inventory does not see arbitrary later comparisons through a stored slot and is not a whole-program identity proof. Freezing preserves identity, so no observation here proves that changing identity by copying is safe. Likewise zero observed writes is limited to the workloads and paths exercised.

[Measurement tools and raw evidence](../filter-readonly-freeze/) contain the site inventory, type-preservation check, static comparison count, per-process counters, exception stacks, source hashes, logs and exact exit codes. Raw bootstrap, preliminary self, canonical self, control and lane observations are kept separate. No landing adaptation, ledger line or internal/ change.
