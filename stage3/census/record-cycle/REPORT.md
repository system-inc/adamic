Units newly stopped by lowering: **UNKNOWN**. First native seed in each: **UNKNOWN**. Current-head lowering coverage: **0 units with aggregate-ready graph-slot seeds**. No all-tsc zero is claimed.

# b529db90 static record-cycle read

**Static read, not a lowering measurement.** Current head is `b529db907bca6564edd2244ac8e5d102b85cf7e8` on `memory/record-allocation-stop`; `git merge-base --is-ancestor 55c375d9d245b580c010c3e0b5659f2849f24fb6 b529db90` succeeded. The stage3 source recipe is unchanged from that base and the superseded head. All static reads use the same prepared, hashed tree from TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`, with pinned TypeScript API 6.0.3. No new adaptation was made.

## Coverage and classification

The API pass inspected all **79 compiler root files**, **812,100 non-type-declaration walk nodes**, and **9,639 source units**. These are source inventory counts, not units lowered far enough to record seeds. It collected **5,413 source writes**, **168 candidate record literals** of **96 distinct checker types**, and **13,535 source-visible type-slot candidates**. It computed one static named-type edge graph over **46,046 types**; it did not assemble unit IR or compute production whole-program facts.

**There is no exhaustive native seed list.** `slotsOf` depends on concrete native type discovery and `fresh.ProveWrites` over coherent whole-program IR. Generated writes, generic instantiations, unknown writes and discarded statements cannot be reconstructed by this source inventory. An explicit UNKNOWN bucket covers those missing native seeds. Every enumerated type slot has UNKNOWN native seed eligibility. The static fallback below gives conditional judgments and evidence, not a claim that all 13,535 candidates are actual graph seeds.

| Inventory | stops | clean | unknown |
| --- | ---: | ---: | ---: |
| Source-write record judgments, **if the write is an eligible seed** | 2610 | 1734 | 1069 |
| Source-visible type-slot judgments, **if the slot is an eligible seed** | 1219 | 0 | 12316 |

`stops` is a positive named-type witness for the head's record predicate, conditional on the native seed existing, that literal actually lowering as record storage, and native/stock checker agreement. `clean` is a source/type-level negative with no unresolved edge on the exclusion route; source scalars account for most clean writes. `unknown` preserves missing structural-view roads, captured cells, unresolved generic types, native storage or write-index information. None is labeled an actual lowering outcome.

Of the conditional source-write positives, **21** are the existing `holdsRecord` arm, not newly introduced stops. **2,589** have a positive static witness for the deeper-record arm, across **996 source units**. Those are **not a count of newly stopped units**. Their first source candidate per unit, slot type, record type, record literal, reach back to holder and quoted line are listed in [STATIC-FIRST-SITES.tsv.gz](STATIC-FIRST-SITES.tsv.gz). Full named paths are in the hashed raw `writes.json`; table strings are capped for readability and marked when truncated.

Every enumerated source write and its classification/reason is in [STATIC-WRITES.tsv.gz](STATIC-WRITES.tsv.gz). Every enumerated type-slot candidate, declaration/use line, type, judgment, unknown seed status and corresponding source-write IDs is in [STATIC-SLOTS.tsv.gz](STATIC-SLOTS.tsv.gz). Candidate literal sites/types/elements and quoted lines are in [STATIC-RECORDS.tsv.gz](STATIC-RECORDS.tsv.gz). These gzip TSVs are compact summaries; full raw API output remains in scratch. [STATIC-SUMMARY.json](STATIC-SUMMARY.json) records counts, options and limitations. This conservative inventory does **not** certify the user's requested exhaustive set of actual `slotsOf` seeds.

## Exact rule read

- [`graph_records.go:17`](https://github.com/system-inc/adamic/blob/b529db907bca6564edd2244ac8e5d102b85cf7e8/internal/lower/graph_records.go#L17): `noteRecordSlot` stops when `holdsRecord(value) || writtenReachesRecord(value, holder, write)`. It saves the first stop; it does not judge every literal as 1c7f96e6 did.
- [`graph_records.go:110`](https://github.com/system-inc/adamic/blob/b529db907bca6564edd2244ac8e5d102b85cf7e8/internal/lower/graph_records.go#L110): `reachesRecord` iterates literal record types and requires both `reaches(value, record)` and `reaches(recordElement(record), holder)`. Missing record/element type fails closed.
- [`graph_records.go:150`](https://github.com/system-inc/adamic/blob/b529db907bca6564edd2244ac8e5d102b85cf7e8/internal/lower/graph_records.go#L150): plain function-field assignments use `writtenFunctionCells`. It checks every matching unproven field write, unknown writes, private fields, readable sites and named-function/literal form; it follows those actual captured cells. All other forms fall back to the slot type. The static pass marks these precise-function cases unknown; it never substitutes every compatible function signature as the written capture set.
- [`cycles.go:374`](https://github.com/system-inc/adamic/blob/b529db907bca6564edd2244ac8e5d102b85cf7e8/internal/lower/cycles.go#L374): `slotsOf` considers mutable arrays/tuples, map keys and values, set elements and non-accessor object fields (including readonly constructor-written fields). Readonly collections and weak links do not seed. A holder that is itself a record retains the older record-cycle refusal before graph seeding and is not counted as a new `noteRecordSlot` seed by the type-slot inventory.
- [`cycles.go:791`](https://github.com/system-inc/adamic/blob/b529db907bca6564edd2244ac8e5d102b85cf7e8/internal/lower/cycles.go#L791): `cycleUnproven` requires the native unproven-write index; unknown writes can match any holder and source field names alone are insufficient. This is why the source candidates are not promoted to known seeds.

The static implementation intentionally derives only named field, union and collection edges and records missing shape/capture edges as uncertainty. A positive named path remains useful static evidence; an omitted route cannot establish a negative. The stock checker emitted **one diagnostic**, TS2488 at `src/compiler/transformers/classFields.ts:1503` (`never` is not iterable); no source edit was made. API options align with `internal/load/load.go` for strictness/module/lib settings and explicitly load the pinned Node declarations. This is stock TypeScript, not a claim of native checker identity. Storage-sensitive Partial/Pick aliases, enumeration-backed fixed-object exceptions, and native representations need native confirmation and are part of the stated conditional scope.

## Why the seam stopped

Independent latent snapshots reuse local/function IDs and full mode discards failed-statement IR (`stage3/census/latent/full.go.txt`, `latentStatement`; `units.go.txt`, per-unit snapshot). `findCycles` needs one coherent `fresh.ProveWrites` result, shapes, local types, closure/accessor environments, captures and write sites. `writtenFunctionCells` makes a naive union particularly unsound: a cell/function index from one snapshot can name a different value in another. Appending IR without rebasing every identity is not a linker, and rebasing cannot recover rolled-back statements or missing instantiations. No aggregate graph, empty capture map or production-root gate is used for the current head.

This matches the concrete sibling evidence in [bind-extra-parameters, b092a62a](https://github.com/system-inc/adamic/blob/b092a62a/stage3/census/bind-extra-parameters/BLOCKED-LATENT-SEAM.md), which was read. Per your instruction, the seam was stopped rather than forced. No current-head latent aggregate control has passed, and no lowering coverage is inferred from a production-tail fixture.

## Current-head tiny production controls

These are **separate production-tail controls**, not validation of the stopped latent aggregate seam. Only the two tiny fixtures run here; no production tsc root gates the static classification. Pending controls are not zero results.

- `b529-positive`: NotYet: a record written into a graph slot, where the write at /workspace/scratch/record-cycle/fixtures/cycle.a:5:1 may close a cycle through it, in a graph program: a record can't join a graph region; NotYet: a record written into a graph slot, where the write at /workspace/scratch/record-cycle/fixtures/cycle.a:5:1 may close a cycle through it, in a graph program: a record can't join a graph region
- `b529-negative`: accepted; accepted
- `base-positive`: accepted; accepted
- `base-negative`: accepted; accepted

[CURRENT-CONTROLS.json](CURRENT-CONTROLS.json) preserves each repeated tuple. Inputs remain [cycle.a.txt](cycle.a.txt) and [acyclic.a.txt](acyclic.a.txt): the positive has `Node → Box.r: Record<string, Node> → Node`; the negative has a real Node-array graph but string-valued records. The existing compiler driver inspects each input twice. At b529db90, both positive observations stop at `cycle.a:5:1` on `a.box = { r: { k: a } };`; both negatives accept. Main accepts both fixtures twice. The identical static scanner also returns one positive source-write judgment for that assignment and two clean negative source-write judgments, with zero diagnostics: [STATIC-CONTROLS.json](STATIC-CONTROLS.json). These controls check the static read on these small shapes and the separate production rule; they do not repair missing aggregate input.

## Superseded head and ordinary table

All completed **1c7f96e6** results are retained in [REPORT-1c7f96e6-SUPERSEDED.md](REPORT-1c7f96e6-SUPERSEDED.md), with raw hashes, `.go.txt` seam patch and controls. Its ordinary table completed **23/79 roots** before retarget; root 24's incomplete raw output is not a completed result. `corePublic` was accepted twice. `hostErrors` was not reached by that cancelled run. No old control or root result is relabeled b529db90. The separately running main table is not reproduced here.

## Receipts and reproduction

Raw files remain under `/workspace/scratch/record-cycle`. [RAW-MANIFEST.json](RAW-MANIFEST.json) gives absolute path, byte size and SHA-256, including full static `writes.json`, `slots.json`, record inventory, diagnostics, pinned head source and controls. [RECEIPTS.json](RECEIPTS.json) gives bounded commands, exit codes, seconds and peak RSS. [static-read.cjs](static-read.cjs) is the read-only scanner. `API_TYPESCRIPT` names the current worktree's pinned API module; run `node --max-old-space-size=8192 static-read.cjs ADAPTED OUTPUT`. Its `stops` labels must be read with the conditional scope above.

The fresh current worktree is detached at b529db90, recursively initialized with depth-1 submodules, and received `npm ci` in `stage3/api`. `GOTMPDIR`, `TMPDIR`, Go cache and raw output are in scratch. One heavy job runs at a time, with 3500-second wrapper bounds. The final static pass exited 0 in 122.57 seconds; wrapper-reported peak RSS 3,895,612 KiB. RSS receipts use getrusage(RUSAGE_CHILDREN) and may include inherited/cumulative child accounting; these are operational receipts, not controlled performance comparisons. Older record-holder mutation rows are marked clean for this new check because their older refusal preempts `noteRecordSlot`; unresolved dynamic-element kinds remain unknown. The initial static raw output is retained separately as `static-b529db90-initial/`. No compiler or adaptation change is published; code archives are `.go.txt`. Active work on the superseded ordinary table was terminated on retarget. No current-head all-unit lowering result, exhaustive native seed count, or successful tsc native build is claimed.
