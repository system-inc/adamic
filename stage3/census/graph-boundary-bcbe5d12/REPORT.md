# Graph adoption boundary census on bcbe5d12

Task jawqwv7. Report only. Counts cover lowered unit results only, including retained successful statements after full-mode rollback. They do not measure accepted whole programs.

Compiler/apply: `bcbe5d1244ee8a218d860415dea19cca114bae46`. Reference report tree: `30057eca`. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `755daee0c6ba`; cgroup CPU quota `400000 100000`; memory limit `17179869184` bytes; 5 visible CPUs. Environment: `ccarenv_b64_Y2NhcmVudl85M2MyY2I4NzA3ZTA4MTkxODMwYTI0NDZlY2ZhNjAyZg`.

## Four-way split, lowered IR only

| Kind | Observations | Distinct reported locations |
| --- | ---: | ---: |
| views | 0 | 0 |
| record_literals | 0 | 0 |
| dynamic_reads | 0 | 0 |
| checked_object_calls | 0 | 0 |

Long pole by observed count: **views, 0**. Coverage caveat: lowered IR only; failed statements, checker-skipped bodies, deferred generics and uncompleted units are absent. Inference: this ranking may differ after those gaps lower.

## Coverage and interpretation

Observed 0 unit results: {}. First-error pass observed 0 results; full pieces observed 0. The summed table uses one result per unit, preferring full mode over first-error mode. Repeated sites in different units are repeated observations, not independent source sites.

Stock unit inventory has 0 units in 0 emitted files. Units without an observer result, by stock status: `{}`. Missing stock file records after an interrupted pass are not evidence of zero units.

In first-error mode a unit with a lowering finding contributes zero sites: its body is discarded. In full mode the stock hook rolls failed statements back and continues. A `stopped` full-mode observation can therefore contain retained successful statements; its failed statements contribute no sites. Panicked unit results contribute zero sites. `lowered` means no new lowering finding in that unit, not whole-program acceptance. File-level refusal-scan findings do not suppress independent-unit examination.

Every listed expression and metadata entry lacks a source-position field at this base. All positions are explicitly labeled unit fallbacks in the raw stream. Distinct reported locations deduplicate `(kind, unit file:line:column)`, and are a lower bound on distinct actual source sites. Multiple expressions or metadata entries at one fallback position cannot be separated into actual source sites from this IR. The observer does not use the boundary's last-seen graphViewSite or graphDynamicSite.

GraphTypes non-empty in 0 observed results. The latent hook never runs `findCycles`; GraphTypes is not computed per unit. Empty flags here do not establish acyclicity. No graph membership is inferred.

## Per file, lowered IR only

Attribution is to the unit's file, not an independent root. Dependency function bodies or metadata can be included in that unit's result.

| Unit file | Views | Record literals | Dynamic reads | Checked object calls | Any observed GraphTypes |
| --- | ---: | ---: | ---: | ---: | --- |

## Top 10 files by record_literals, lowered IR only

| Unit file | Count |
| --- | ---: |

Examples below are unit fallback positions and adapted unit source text, not exact expression locations.

## Top 10 files by views, lowered IR only

| Unit file | Count |
| --- | ---: |

Examples below are unit fallback positions and adapted unit source text, not exact expression locations.

## Pass exits and limits

| Pass | Child exit | Seconds | Stopping reason |
| --- | ---: | ---: | --- |
| apply | 0 | 917.51 | completed |
| census-build | 0 | 473.132 | completed |
| census-final-build | 0 | 1.006 | completed |
| entry-build-fixed | 0 | 33.1 | completed |
| entry-build | 1 | 10.017 | completed |
| entry-fixture | 0 | 1.007 | completed |
| instrument-entry-fixed | 0 | 1.001 | completed |
| instrument-entry | 0 | 1.002 | completed |
| npm-ci | 0 | not timed | completed |
| overlay-entry-fixed | 0 | 2.011 | completed |
| overlay-entry | 0 | 3.006 | completed |
| overlay | 0 | 6.029 | completed |
| selftest | 0 | 1.028 | completed |
| setup | 0 | 758.331 | completed |
| source | 0 | 12.004 | completed |
| verify-apply | 0 | 1.003 | completed |

## Method and reproduction

Read `graph_types.go:460` and its callers at `graph_types.go:355` and `program_region.go:140`. Count ViewOrigins entries plus CheckedFields keys without deduplicating their overlap; walk Main and every function body with the boundary's own walk for record ObjectLiteral, both dynamic-read nodes, and checked ObjectCall. These are hypothetical boundary candidates before the rest of ordinary lowering, not demonstrated graph-program refusals.

`instrument.py` extends the existing guarded latent overlay in scratch only. It retains successful top-level statement returns in the measurement result's Main, which the stock hook discards. Ordinary Load and Lower remain disabled and every real pass asserts those guards. `overlay/` contains all generated Go inputs as `.go.txt`; no production lower source is changed. The deterministic self-test exercises Main, a function body, both dynamic reads, both view sources, GraphTypes presence, and failed-unit exclusion.

Raw boundary streams, original latent streams, timing streams, logs and each child's own exit status are compressed separately beside this report in `raw/`. `source-manifest.json`, `provenance.json` and `SHA256SUMS` retain source and binary identities. Piece selections are retained in raw JSON. Each process group is killed and reaped by the supervisor on completion or stopping.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1
npm ci --prefix stage3/api > npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
bash stage3/apply.sh /workspace/graph-boundary-adapted > apply.log 2>&1
python3 stage3/census/latent/make_overlay.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay
python3 stage3/census/graph-boundary-bcbe5d12/instrument.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay
go build -overlay /workspace/scratch/jawqwv7/overlay/overlay.json -o /workspace/scratch/jawqwv7/census ./stage3/census/latent/tool
# run-pieces.py records the exact entry invocation and file selections
```
