# TypeScript view-pairs census

Task vqzty8t. Measurement only. Counts describe observed latent unit lowerings, not accepted whole programs.

Box `755daee0c6ba`, CPU quota `400000 100000`, memory limit `17179869184` bytes. Source/compiler base `734aef5f3a28fdc6b21d0decb05d6338e78a027b`. Probe `86157bea8c6eafd207a3b518e1ec817fed52c072`. TypeScript 6.0.3 pin `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Base and probe

The base is 734aef5f, the main parent before Memory's first throwaway probe commit 8bedbc18. The named tip 86157bea fixes that probe's program-name conversion. Its immediate parent already contains the probe, so it is not used as a production base. bcbe5d12 has different adaptation inputs; this run builds the keyed adapted product at 734aef5f rather than reusing bcbe's adapted bytes. Prepared Go/cohere and pinned parser dependencies were reused.

Memory's original probe is preserved as memory-probe-original.go.txt and copied into scratch. The scratch context version calls logViewPairs() at every unit end beside the boundary observer, using only &cycleFinder{l:l, where:map[*checker.Type]*ast.Node{}}. findCycles never runs. Original pair traversal, per-unit checker-pointer seen set, first encountered path and depth bound 6 are preserved. The original program field names the entry; file and unit fields supply actual owner attribution. No probe is added to production lower source.

Annotations add full nested paths, nil-source target types and reference fields, and deriveRelated results for intersection records. Object classification accepts object/union shapes and intersections, with intersection targets counted as objects even when the original probe labels them target-not-object. Source intersections also qualify when the target is object/union shaped. The original skipped flags remain in raw rows. Intersections do not add a second recursive traversal, so their fields skipped by the original probe remain unwalked.

## Counts

**Unrelated object pair observations: 0; distinct textual (source, target, nested path) pairs: 0. Nil-source views with a known runtime reference field: 0.** The object-capable reference companion is 0. Nil-source views with no proven reference field but unknown fields or missing target: 0.

Selected-unit pair observations: **0**. Raw lines across all passes, including superseded first-error and retry observations: **0**.

| Split | Selected unit observations | All raw lines |
| --- | ---: | ---: |
| object_related_top | 0 | 0 |
| object_related_nested | 0 | 0 |
| object_unrelated_top | 0 | 0 |
| object_unrelated_nested | 0 | 0 |
| scalar_top | 0 | 0 |
| scalar_nested | 0 | 0 |
| nil_source | 0 | 0 |
| nil_target_only | 0 | 0 |

Intersection targets are a subset of the object counts, not an added bucket: 0, split as `{'object_related_top': 0, 'object_related_nested': 0, 'object_unrelated_top': 0, 'object_unrelated_nested': 0, 'scalar_top': 0, 'scalar_nested': 0, 'nil_source': 0, 'nil_target_only': 0}`. Scalar/non-object rows use the probe's flag test; top-level non-object rows are shown separately from nested scalar fields.

Reference field means a cycleFinder.fields field whose lowering representation is known and ir.Type.IsReference() is true. This runtime definition includes strings and Weak. The companion object-capable count excludes strings and Weak. Unknown representations remain explicitly unknown, not silently declared scalar. The nil-source count is a count of views; unlike non-nil pairs, the original probe does not deduplicate nil sources.

## Coverage

Observed **0 unit results**: `{}`. Unobserved stock units by status: `{}`. Full-mode files attempted: **0/81**; emitted complete full-mode file inventories: 0. Stock inventory has 0 units. Probe-incomplete selected units: 0; pair lines without a completed probe unit row: 0.

One result per (unit file, unit position) is selected: full mode supersedes first-error mode, and later full pieces supersede earlier ones. Repeated dependency pairs in different units remain repeated observations. TypeToString text can be truncated or reused for distinct checker identities; the distinct-pair list groups displayed source/target strings plus the full first-encountered nested path, not globally shared checker identities.

First-error stopped units expose the attempted lowering state before the stop, even when their IR body would be discarded. Full mode rolls failed statements back and retains successful statements. The probe sees the retained state at the end of that unit. Neither mode counts rolled-back state, skipped/deferred bodies, units not reached before a time limit, paths hidden behind a statement's first failure, pair paths suppressed by the probe's per-unit seen set, or recursion past depth 6. A completed file traversal still has those gaps. Missing buffered latent inventories are not evidence that a file had no units.

The latent hook does not compute GraphTypes or establish graph membership. GraphTypes was non-empty in 0 observed results. No acyclicity is inferred. Probe errors: 0.

## Every distinct unrelated object pair

All counts and owner positions are also in [unrelated-object-pairs.json](unrelated-object-pairs.json). Paths are the nested field path at its first encounter in each unit. Files below are owner-unit files, not independent roots.

| Source type | Target type | Nested path | Count | Files and units |
| --- | --- | --- | ---: | --- |
| none observed | | | 0 | |

## Every nil-source view with a runtime reference field

Each observation is retained, including repeats in one unit, in [nil-source-reference-views.json](nil-source-reference-views.json). Unknown-only cases are in [nil-source-reference-unknown.json](nil-source-reference-unknown.json).

| Unit | Target type | Runtime reference fields | Object-capable fields |
| --- | --- | --- | --- |
| none observed | | | |

## Per-file coverage

| Owner file | Lowered | Stopped | Selected pairs | Full mode |
| --- | ---: | ---: | ---: | --- |

## Pass exits

Each child is capped at 1490 seconds with five seconds for forced cleanup, below 25 minutes. Priority and files larger than 300000 bytes run alone. At most two smaller files run concurrently. Completed pieces are committed and pushed as they finish. Each process group is reaped and cleaned by supervise.py.

| Pass | Child exit | Seconds | Reason |
| --- | ---: | ---: | --- |

## Evidence and replay

Raw streams, pass environments, timings, logs and actual child exits are in raw/. selected-pairs.jsonl.gz and selected-units.jsonl.gz preserve the selected population. overlay/ archives scratch Go inputs as .go.txt only. source-manifest.json hashes the adapted TypeScript bytes. provenance.json records pins, box and binary identity. SHA256SUMS covers evidence. report.py can recount finished archived/scratch streams without lowering. run-census.py resumes finished pieces without rerunning them.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
bash stage3/apply.sh /workspace/view-pairs-adapted
python3 stage3/census/latent/make_overlay.py /workspace/view-pairs /workspace/scratch/vqzty8t/overlay
python3 stage3/census/view-pairs-734aef5f/instrument.py /workspace/view-pairs /workspace/scratch/vqzty8t/overlay
python3 stage3/census/view-pairs-734aef5f/prepare-probe.py
go build -overlay /workspace/scratch/vqzty8t/overlay/overlay.json -o /workspace/scratch/vqzty8t/census ./stage3/census/latent/tool
python3 stage3/census/view-pairs-734aef5f/run-census.py
```
