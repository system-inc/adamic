# Scout ty5myjt — resolution recommendation; diagnostics census stopped

| Helper | Recommended shape | Stock tsc 6.0.3, 79 roots | Full single-file census | Emitted JS |
| --- | --- | --- | --- | --- |
| getDiagnosticsHelper | Named readonly `file` view, **provisional only** | 0 diagnostics, 0 new | Baseline timed out; neither target disappearance verified | byte-identical |
| watchFailedLookupLocationsOfExternalModuleResolutions | Local named readonly `resolvedFileName` constraint view | 0 diagnostics, 0 new | 4 target refusals gone across this helper and the next two; 0 new, 0 changed | byte-identical |
| stopWatchFailedLookupLocationOfResolution | Same local readonly constraint view | 0 diagnostics, 0 new | Same complete resolutionCache.ts census | byte-identical |
| removeResolutionsOfFileFromCache | Same local readonly constraint view | 0 diagnostics, 0 new | Same complete resolutionCache.ts census | byte-identical |

**Stopped part:** the full baseline `program.ts` census produced only its header
in **602 seconds** (13:21 CPU time, 1,739,648 KiB RSS). It was terminated at the
announced 600-second limit. There is no file record or relation trailer, so no
baseline count, candidate full census, or finding-disappearance claim is made for
that file. `program-timeout.json`, `program-baseline-INCOMPLETE.jsonl`, and
`runs.log.txt` preserve this failure to obtain a complete measurement. No diagnostics
helper shape is cleared for adaptation. This is a partial scout with a concrete
stopping reason, not completion of the requested program.ts experiment matrix.

No adaptation or compiler change is delivered. All candidate edits were confined
to separate scratch adapted trees. This branch contains reporting and measurement
artifacts only; no Go files were added.

## Recommended resolution shape

```typescript
interface ResolutionReadView extends ResolutionWithResolvedFileName {
    readonly resolvedFileName: string | undefined;
}
```

Use this local view instead of `ResolutionWithResolvedFileName` in R's constraints
and its default in `GetResolutionWithResolvedFileName`. The scratch patch aligns
the `ResolutionCache` method, `ResolveNamesWithLocalCacheInput`,
`resolveNamesWithLocalCache`, and the three helpers. It preserves their generic
cache-wrapper and result types. The original interface remains unchanged.
See [resolution-readonly.patch](resolution-readonly.patch).

The complete guarded CommonJSBundle `resolutionCache.ts` census is
**221 → 217 raw findings; 133 → 129 unique findings; 0 new, 4 gone, 0 changed**.
All four gone findings are refusal_scan/Refused at baseline **774:17, 1053:13,
1056:17, 1471:9**, with reason “ResolvedModuleFull seen as
ResolutionWithResolvedFileName, which can write string | undefined where string
is read”. Every other finding, including lowering-phase findings, is retained.

Both baseline and readonly streams finish with `relation_report`, status 0, and
pass the no-output guards. They are archived as
`baseline-resolutionCache-full.jsonl.gz` and
`resolution-readonly-resolutionCache-full.jsonl.gz`.

## Bodies, reads and writes

[BODIES.md](BODIES.md) contains each helper's **full adapted signature and body**,
along with the forwarding `resolveNamesWithLocalCache` body and diagnostic sorter.

- `getDiagnosticsHelper`: no direct `file` access or write. Its two branches pass
  the returned `readonly T[]` to `sortAndDeduplicateDiagnostics`; the all-files
  branch first uses `flatMap` and cancellation checks. Those two sorter calls are
  the places exposing the diagnostics to the sorter's mutable Diagnostic
  constraint. The current sorter calls `sortAndDeduplicate` with
  `compareDiagnostics` and `diagnosticsEqualityComparer`; their implementations
  compare diagnostic fields rather than assigning `file`. A readonly helper view
  must still be checked against these downstream contracts: stock assignability
  alone does not prove Adamic accepts that forwarding.
- `watchFailedLookupLocationsOfExternalModuleResolutions`: its callback result R
  is only read, in the `resolved && resolved.resolvedFileName` guard and
  `resolutionHost.toPath(resolved.resolvedFileName)`. No R object is passed to a
  writer. It does write `(resolution.files ??= new Set())` and mutate sets/maps;
  those operations concern wrapper T and watcher state, not R.resolvedFileName.
- `stopWatchFailedLookupLocationOfResolution`: the same two resolved-file reads,
  no resolved-field write or R-object escape to a writer. It deletes from
  `resolution.files`, assigns `resolution.files = undefined`, deletes map/set
  entries, and decrements watcher counters. These are different fields/objects.
- `removeResolutionsOfFileFromCache`: no direct resolved-file access. It forwards
  wrapper T and the callback to `stopWatchFailedLookupLocationOfResolution`, then
  deletes the cache entry. Its callee's R result is read as described above.
- `resolveNamesWithLocalCache`: reads callback result `resolvedFileName` and
  `packageId` for trace messages and equality; it does not assign either widened
  resolved field. It forwards the callback to watch/stop. Its other cache/wrapper
  mutations do not write R.resolvedFileName.

Thus the widened fields are read-only needs of these implementations. A truthful
readonly view is the appropriate constraint direction. In resolution helpers,
**R** is the resolved-result binder; T is the separate cache-wrapper binder.

## Every caller

[CALLERS.md](CALLERS.md) enumerates every syntactic caller in all `src/`, including
harness, and its concrete or forwarded T/R. [call-inventory.json](call-inventory.json)
retains full call text and stock resolved signatures; [harness-call.json](harness-call.json)
independently resolves the harness call outside the 79 compiler roots.

Diagnostics has three callers: syntactic and declaration use
`DiagnosticWithLocation`; semantic uses `Diagnostic`. Resolution has module,
type-reference, generic forwarding, and a harness caller. The harness's
ExpectedResolution is the intersection of the two cached wrappers; its R is
`{ resolvedFileName: string | undefined }`. There is no single exact caller type
for either family.

## Candidate matrix

Each candidate starts from the same finished adapted tree and changes only the
file named in its patch. Stock results use exactly the 79 compiler roots plus
pinned Node ambient declarations. Baseline has **0** diagnostics. All original
stock libraries are unchanged; options, root names and every diagnostic are in
`*-stock.json`. The initial ambient-free control had 64 diagnostics; it is not the
baseline used in this table.

| Scratch candidate | Stock diagnostics / new | Full raw; unique | Full new / gone / changed |
| --- | --- | --- | --- |
| diag-intersection | 0 / 0 | not run: baseline timeout | unmeasured |
| diag-readonly | 0 / 0 | not run: baseline timeout | unmeasured |
| diag-exact | 1 / 1 | not run: baseline timeout | unmeasured |
| diag-overloads | 0 / 0 | not run: baseline timeout | unmeasured |
| resolution-readonly | 0 / 0 | 221→217; 133→129 | 0 / 4 / 0 |
| resolution-exact | 2 / 2 | 221→184; 133→121 | 2 / 14 / 0 |
| resolution-result-view | 0 / 0 | 221→217; 133→129 | 0 / 4 / 0 |

(a) **Literal intersection control:**
`T extends Readonly<Pick<Diagnostic, "file">> & Diagnostic`. Stock tsc still
permits `value.file = undefined`: intersecting with the mutable Diagnostic
restores a writable member. It does not meet the requirement that the body
cannot write the field. [readonly-control.json](readonly-control.json) records
this alongside the single expected TS2540 for an Omit-and-replace readonly view.
The named candidate instead uses a local interface extending Diagnostic and
redeclaring `readonly file: SourceFile | undefined`.

(b) **Exact-type negative controls:** diagnostics replaces T throughout the
helper with DiagnosticWithLocation, adding TS2322 at the semantic caller.
Resolution replaces result-polymorphic R with ResolvedModuleFull, adding TS2322
and TS2345 at type-reference callers. This control retains wrapper T; a fully
non-generic split over every wrapper family was **not built or measured**. The
mixed caller set prevents a single exact specialization from serving all callers.
The exact resolution census's extra disappearances include diagnosed-body skips;
its smaller total is not a successful repair because stock tsc rejects it.

(c) **Other source-suggested shapes:** diagnostics adds two non-generic overloads
for DiagnosticWithLocation and Diagnostic, with a Diagnostic implementation and
unchanged executable body. The result-view resolution candidate removes R from
the helper/input signatures and gives callbacks a fixed local readonly result
view, retaining wrapper T. That view extends ResolutionWithResolvedFileName to
preserve packageId reads. Its complete census clears exactly the same four
findings, with no new/changed findings. The readonly-constraint shape is preferred
because it preserves existing result polymorphism and requires fewer contract
changes.

**Every full-file new, gone or changed finding** is listed in
[FINDINGS.md](FINDINGS.md) and the three `*-delta.json` files. All three resolution
streams are archived, including the stock-rejected control. No scan-only run is
substituted for full mode.

## JavaScript

All seven candidates were emitted for **both** program.ts and resolutionCache.ts:
**14/14 byte-identical comparisons**, stock 6.0.3 `transpileModule`, ES2024,
CommonJS, comments retained. [javascript.json](javascript.json) gives byte counts
and SHA256s; each candidate's `.js.diff` is empty. Each `.patch` records the exact
scratch type changes. This establishes emission identity, not acceptance of the
stock-rejected controls or unmeasured diagnostics census.

## Who pins the owners

The direct searches are in [owner-grep.txt](owner-grep.txt) and
[resolver-grep.txt](resolver-grep.txt). No direct helper-name match appears in an
active adapt.cjs. Indirect ownership matters:

- `95-checker-axolotl-h/adapt.cjs:76–83` hashes the **entire createProgram owner**,
  which contains getDiagnosticsHelper. Its `sites.json` holds before/after owner
  hashes. A signature-only edit inside this helper changes that pin.
- `99-correctness-axolotl-x/ownership.json` explicitly records createProgram's
  ownership by 95-checker-axolotl-h; its inventory/replay evidence records the
  associated later pin work. A future adaptation must account for that chain.
- `98-axolotl-d/adapt.cjs:20–21` pins the shapes of
  ResolutionWithResolvedFileName.packageId and
  ResolutionWithFailedLookupLocations.files, not the full resolution-helper
  bodies. It cites the stop helper's files=undefined write as its invariant.
- No full createResolutionCache/helper owner hash was found by these searches.
  The resolution candidates edit its nested signatures, the ResolutionCache
  method signature, callback alias, and ResolveNamesWithLocalCacheInput; their
  direct-name searches found no active whole-owner pin. No existing pins were
  changed or bypassed in this scout.

## Setup and reproduction

Fresh `git worktree add --detach` of
`7534c1d7575db02bb7b6512119e0685987b7e333` at
`/workspace/scratch/scout-ty5myjt/adamic`; only then was this new report branch
created. Shared checkout source remains unchanged. Recursive shallow submodule
initialization succeeded: cohere `5360753e19e4417675a90e92a1823bea66b13918`, nested
TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. **No credential error.**

Microsoft TypeScript was cloned with `--depth 1 --branch v6.0.3` and confirmed at
`050880ce59e30b356b686bd3144efe24f875ebc8`. `npm ci` in scratch stage3/api succeeded.
`ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/scout-ty5myjt/TypeScript` was supplied
to `stage3/apply.sh`; the pipeline completed successfully (`apply.log.gz`).

The existing measurement overlay was built with make_overlay.py, not a production
compiler edit. Fresh-box Go setup initially reported
`go: creating work dir: stat /tmp/adamic-gate: no such file or directory`, then
`cgo: open /tmp/adamic-gate/cgo-gcc-input-1712609632: no such file or directory`.
The successful retry set GOTMPDIR and TMPDIR to writable scratch and
CGO_ENABLED=0. `build.log.txt` is empty, exit status 0.

The full invocation for each completed resolution tree was:

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE="$TREE/src/compiler/resolutionCache.ts" \
/workspace/scratch/scout-ty5myjt/latent "$TREE/src/compiler" "$OUTPUT"
```

Default module format is verified as CommonJSBundle in each header. Measured full
run wall times: baseline 88.04s, readonly constraint 84.29s, rejected exact control
68.34s, fixed readonly result 86.41s. The program baseline used the same flags with
program.ts selected, but did not complete within its bound.

`measure.cjs`, `stock-node.cjs`, `harness.cjs`, `variants.py`, `run.py`, and `delta.py`
preserve the exact scratch paths/workflow used. run.py stopped after its first
full census failed; the completed resolution series was separately invoked with
the command above into parallel-results. Delta identity is (phase, kind,
source-line-mapped where, normalized reason); changed means identical identity
but different normalized text. Roots in all diagnostic fields are normalized,
including references to dependency files. Repeated findings under different units
are counted in raw totals and deduplicated for unique totals.

The streams retain the tool's literal “measured on a checker-rejected program”
label, but baseline/readonly/fixed-view headers explicitly say
`checker_rejected: false`; only the rejected exact control has checker diagnostics.
No full adaptation oracle or native-output test was run. Full program candidate
censuses, their complete deltas, and fully non-generic cache-family splits remain
unmeasured because of the stated stopping point.
