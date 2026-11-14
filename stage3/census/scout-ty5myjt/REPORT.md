# Scout ty5myjt — resolution full census; program.ts scan only

| Helper | Recommended shape | Stock tsc 6.0.3, 79 roots | Census and mode | Emitted JS |
| --- | --- | --- | --- | --- |
| getDiagnosticsHelper | Non-generic overloads: best measured scan-only candidate | 0 diagnostics, 0 new | **scan only: 81→79; 0 new, 2 gone, 0 changed; both targets clear** | byte-identical |
| watchFailedLookupLocationsOfExternalModuleResolutions | Local named readonly `resolvedFileName` constraint view | 0 diagnostics, 0 new | full: 4 target refusals gone across this helper and the next two; 0 new, 0 changed | byte-identical |
| stopWatchFailedLookupLocationOfResolution | Same local readonly constraint view | 0 diagnostics, 0 new | full: same complete resolutionCache.ts census | byte-identical |
| removeResolutionsOfFileFromCache | Same local readonly constraint view | 0 diagnostics, 0 new | full: same complete resolutionCache.ts census | byte-identical |

**Program follow-up is scan only.** All five diagnostics candidates now have
complete refusal-scan results. The named readonly file view removes the original
file-reason identities, but leaves numeric-field refusals at both target sites
and adds two helper→sorter refusals. Applying the same view to the sorter removes
those handoff refusals, but exposes two new comparator-handoff refusals in the
sorter's own file. The overload candidate alone clears both target sites with
zero new program refusal-scan findings and zero stock diagnostics. See
[PROGRAM-SCAN.md](PROGRAM-SCAN.md) for **every new, gone and changed finding**,
including the supplementary utilitiesPublic.ts scans.

**Historical full-mode stop retained:** program.ts full census was stopped at
602 seconds without a complete file record. Its incomplete stream and timeout
remain archived. This follow-up uses the requested scan-only wrapper instead;
it does not claim full body-lowering coverage or erase that failed full run.

No adaptation or compiler change is delivered. All candidate edits were confined
to separate scratch adapted trees. This branch contains reporting and measurement
artifacts only; no compilable Go files were added. Measurement Go templates end
in `.go.txt`.

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
files named in its patch. Stock results use exactly the 79 compiler roots plus
pinned Node ambient declarations. Baseline has **0** diagnostics. All original
stock libraries are unchanged; options, root names and every diagnostic are in
`*-stock.json`. The initial ambient-free control had 64 diagnostics; it is not the
baseline used in this table.

| Scratch candidate | Stock diagnostics / new | Mode; raw; unique | New / gone / changed |
| --- | --- | --- | --- |
| diag-intersection | 0 / 0 | scan only; 81→81; 81→81 | 2 / 2 / 0 |
| diag-readonly | 0 / 0 | scan only; 81→83; 81→83 | 4 / 2 / 0 |
| diag-readonly-sorter | 0 / 0 | scan only; 81→81; 81→81 | 2 / 2 / 0; plus 2 new in sorter file |
| diag-exact | 1 / 1 | scan only; 81→81; 81→81 | 2 / 2 / 0 |
| diag-overloads | 0 / 0 | scan only; 81→79; 81→79 | 0 / 2 / 0 |
| resolution-readonly | 0 / 0 | full; 221→217; 133→129 | 0 / 4 / 0 |
| resolution-exact | 2 / 2 | full; 221→184; 133→121 | 2 / 14 / 0 |
| resolution-result-view | 0 / 0 | full; 221→217; 133→129 | 0 / 4 / 0 |

(a) **Literal intersection control:**
`T extends Readonly<Pick<Diagnostic, "file">> & Diagnostic`. Stock tsc still
permits `value.file = undefined`: intersecting with the mutable Diagnostic
restores a writable member. It does not meet the requirement that the body
cannot write the field. Its scan only run retains both target sites with the same
file-width refusal and a newly spelled intersection constraint. [readonly-control.json](readonly-control.json) records
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

**Every full-mode resolution new, gone or changed finding** is listed in
[FINDINGS.md](FINDINGS.md) and the three `*-delta.json` files. All three resolution
streams are archived, including the stock-rejected control. Those resolution results remain full mode. Program results are separately and
explicitly labeled scan only in PROGRAM-SCAN.md and the tables above.

## JavaScript

All seven candidates were emitted for **both** program.ts and resolutionCache.ts:
**14/14 byte-identical comparisons**, stock 6.0.3 `transpileModule`, ES2024,
CommonJS, comments retained. [javascript.json](javascript.json) gives byte counts
and SHA256s; each candidate's `.js.diff` is empty. Each `.patch` records the exact
scratch type changes. This establishes emission identity, not acceptance of the
stock-rejected controls or full body-lowering acceptance of diagnostics shapes.
The additional helper-plus-sorter candidate is also stock-clean across 79 roots
and has byte-identical program.ts and utilitiesPublic.ts emissions, recorded in
diag-readonly-sorter-stock.json and diag-readonly-sorter-javascript.json.

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
No full adaptation oracle or native-output test was run. Full program body-lowering censuses and fully non-generic cache-family splits
remain unmeasured. The requested follow-up refusal-scan comparisons are complete;
they are documented below and in PROGRAM-SCAN.md.


## Program continuation — scan only

The wrapper is copied verbatim from the pinned
`typescript/promise-discard-sites` commit in `scan-wrapper-pin.txt`, path
`stage3/census/promise-discard-sites/scan-only.go.txt`. Only that measurement
wrapper was taken; **the compiler stays at base
7534c1d7575db02bb7b6512119e0685987b7e333**. It invokes the existing latentScan
visitor and emits zero lowering units. `scan-only.go.txt` and
`scan-driver.go.txt` preserve the wrapper and driver. The driver keeps both
no-output guards and calls LatentScanOnly in place of Latent; its header explicitly
says `latent_mode: "scan only"`. No production file was changed.

Build: start with the existing scratch make_overlay.py overlay, add the wrapper
as a virtual internal/lower/latent_scan_only.go, and map tool/main.go to the
scan-only driver. Build with `go build -buildvcs=false -overlay=<scan overlay.json>
-o <scratch>/scan-only ./stage3/census/latent/tool`, with the same scratch TMPDIR,
GOTMPDIR and CGO_ENABLED=0 as the full binary. `scan-build.log.txt` is empty,
status 0. `build_scan.py` reproduces those overlay mappings from the saved sources.

`run_scan.py` selects one absolute LATENT_ONLY_FILE, retains LATENT_FULL=1 and
LATENT_ASSERT_NO_OUTPUT=1, explicitly selects CommonJSBundle, and uses a 1500-second
process-group kill bound. Baseline ran first, named readonly view next; the
helper-plus-sorter candidate was then measured after its handoff refusals were
observed. All original diagnostics candidates were measured. `scan-runs.json`
records all eight complete runs (six program runs including baseline, two sorter
file runs), timings, status 0, no timeouts, and zero units. Every raw stream,
per-run metadata and log is archived under scan-*. Baseline program took 20.47s;
all program candidates finished without using the bound.

The two target baseline findings are confirmed present at 2796:16 and 2829:16.
All counts/deltas cover the entire selected file's emitted refusal scan, not just
those targets. `scan_delta.py` rejects incomplete streams, nonzero units, or a
non-refusal-scan phase, and applies the same normalized identity as the resolution
comparison. Reason changes count as gone plus new, not changed. The original two
identities disappear in every variant, but site clearance differs; the table in
PROGRAM-SCAN.md makes that distinction explicit.

The numeric replacement is explained by DiagnosticWithLocation also narrowing
**start and length**, each `number | undefined` in Diagnostic and `number` in
DiagnosticWithLocation. A readonly file alone does not remove those write
capabilities. The emitted reason does not identify which numeric member was
encountered first. Adding readonly to the sorter constraint carries file through
the helper handoff, but its comparers still accept mutable Diagnostic. They expose
new refusals at baseline utilitiesPublic.ts:304:47 and :304:67 (candidate :306:47
and :306:67). These supplementary **scan only** results are 34→36 unique (55→57 raw), 2 new, 0 gone,
0 changed, with every finding listed in PROGRAM-SCAN.md.

## Does the sorter or its call chain write file?

**No.** `sorter-call-chain.json` preserves the exact adapted declarations and
bodies used for this audit:

- sortAndDeduplicateDiagnostics forwards the readonly input array, compareDiagnostics,
  and diagnosticsEqualityComparer to sortAndDeduplicate.
- sortAndDeduplicate calls toSorted and deduplicateSorted. toSorted uses
  array.slice().sort(comparer): it reorders a **copy**. deduplicateSorted appends
  element references to a new buffer and updates local last/next values. Neither
  assigns a diagnostic property.
- compareDiagnostics and diagnosticsEqualityComparer read diagnostic file/path,
  start, length, code, message text, canonicalHead and relatedInformation. Their
  reachable diagnostic helpers are getDiagnosticFilePath, getDiagnosticCode,
  getDiagnosticMessage, compareDiagnosticsSkipRelatedInformation,
  compareRelatedInformation and compareMessageText; message-chain comparers recurse
  through message data, and related-information comparisons recurse through the
  same diagnostic comparator. None assigns file or a diagnostic location field.
- getDiagnosticFilePath is exactly a file-presence test followed by file.path;
  primitive string/number comparisons cannot write the diagnostic object.

Thus the current implementations only read these location fields. Their declared
mutable Diagnostic contracts are the capability obstacle. The measured file-only
view is insufficient both for start/length and for the sorter→comparer handoff.
This scout does not propagate new readonly comparator/helper contracts further or
claim those unmeasured shapes would pass. No adapter/owner pin is changed: the
existing createProgram enclosing hash still matters, and the direct sorter-name
search found no active whole-sorter owner pin.

`verify_scan.py` checks all archived scan-only counts, trailers, zero units, target
presence, wrapper provenance, stock cleanliness and JavaScript identity;
`scan-verification.json` records its passing result. Normal refusal-scan pruning
still applies: these are all emitted findings, not independent checks of every
possible rule after an earlier refusal.
