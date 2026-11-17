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
| 1 | 261 | `export function filter<T, U extends T>(array: T[], f: (x: T) => x is U): U[];` | 5 | yes | `overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined` |
| 2 | 263 | `export function filter<T>(array: T[], f: (x: T) => boolean): T[];` | 60 | yes | `overload 1 of filter result T[] cannot be served by implementation result readonly T[] | undefined` |
| 3 | 265 | `export function filter<T, U extends T>(array: readonly T[], f: (x: T) => x is U): readonly U[];` | 14 | yes | `overload 1 of filter result readonly U[] cannot be served by implementation result readonly T[] | undefined` |
| 4 | 267 | `export function filter<T>(array: readonly T[], f: (x: T) => boolean): readonly T[];` | 41 | no | `none observed; scan completed` |
| 5 | 269 | `export function filter<T, U extends T>(array: T[] | undefined, f: (x: T) => x is U): U[] | undefined;` | 11 | yes | `overload 1 of filter result U[] | undefined cannot be served by implementation result readonly T[] | undefined` |
| 6 | 271 | `export function filter<T>(array: T[] | undefined, f: (x: T) => boolean): T[] | undefined;` | 13 | yes | `overload 1 of filter result T[] | undefined cannot be served by implementation result readonly T[] | undefined` |
| 7 | 273 | `export function filter<T, U extends T>(array: readonly T[] | undefined, f: (x: T) => x is U): readonly U[] | undefined;` | 9 | yes | `overload 1 of filter result readonly U[] | undefined cannot be served by implementation result readonly T[] | undefined` |
| 8 | 275 | `export function filter<T>(array: readonly T[] | undefined, f: (x: T) => boolean): readonly T[] | undefined;` | 4 | no | `none observed; scan completed` |

| Group control | Original overloads retained | declareModule result |
| --- | --- | --- |
| boolean | 2, 4, 6, 8 | `263:1`: `overload 1 of filter result T[] cannot be served by implementation result readonly T[] | undefined` |
| readonly | 3, 4, 7, 8 | `265:1`: `overload 1 of filter result readonly U[] cannot be served by implementation result readonly T[] | undefined` |

**Answer:** every mutable-result overload (1, 2, 5, 6) and both readonly narrowing overloads (3, 7) refuse in these controls. The refusal is not limited to narrowing callbacks: boolean overloads 2 and 6 also refuse. Only readonly boolean overloads 4 and 8 escape this declaration refusal in isolation. The group controls report their first declaration stop only; their later overloads are judged individually by the eight single-overload controls.

The full latent census records **0 findings for each of the ten controls**, with all retained signatures and the implementation marked `deferred_generic_instantiation`, without instantiation witnesses. This is a coverage gap, not evidence that the six refused declarations lower. Scan-all's `Completed: true` for controls 4 and 8 is measurement traversal only, not backend acceptance.

Reproduce with `python3 overload-measurement/run.py BASELINE_TREE SCRATCH SCAN_BINARY LATENT_BINARY` from this report's directory, using the unchanged main scan-all and latent overlays built earlier. [Raw controls and observations](overload-measurement/evidence/) include exact sources, stdout, stderr, exit codes, full latent streams, and combined results. The resolved counts are the existing completed 713-root census in `evidence/before-count.json.gz`.

No landing adaptation, ledger line or internal source change. This measurement is pushed before measurement 2 begins.
