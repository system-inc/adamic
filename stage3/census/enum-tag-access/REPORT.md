# Paired open enum tag access census

Base: `b03fe07e48cce2b4a11f183bb3b4a8ba06bb60b9`. Head: `5b18178288676b943ccbde09fe9370cb37bfb7b4`.

Head branch: `optimizer/enum-tag-access-local-2`, task `qeanycw`. Measured on 2026-11-04 on one adapted TypeScript v6.0.3 tree. The head is exactly one commit on the base. The control is this base, with no older main measurement.

## Results

Open tag sites: **1006 base, 987 head**. The head removes 74 base findings and adds 55 open-tag findings at other source identities: 1006 - 74 + 55 = 987.

| Outcome of a base open tag site | Sites |
| --- | ---: |
| no finding | 74 |
| unchanged | 932 |
| moved to another reason | 0 |

Counts deduplicate the meter key `(kind, where, reason, text)`. Matching between sides uses source path, SHA-256 of the identical file bytes, and UTF-8 byte offset converted from its UTF-16 diagnostic position. It never matches by line alone. `DELTA.json` includes every changed site and every base open tag site.

All 1,006 base open-tag findings came from the refusal scan. Timed lowering attempts are a separate measurement. A no-finding outcome means that source identity has no finding in the head stream. First-error observation does not prove successful whole-program compilation, and a unit can advance to a stop at another source identity. The unit table below records those changes separately.

| Kind | Base | Head |
| --- | ---: | ---: |
| NotYet | 559 | 559 |
| Refused | 4036 | 4017 |
| panic | 0 | 0 |
| error | 0 | 0 |
| SkippedDependency | 0 | 0 |

Each side has 4,056 attempted units and 498 deferred generic units. Generic instantiation witness changes: 0. Deferred generic units are coverage gaps, not successful lowerings.

## Regressions

54 newly stopped source identities had no base finding; all stop on the open-tag rule. 0 units changed from no lowering stop to a stop, or from a non-panic stop to a panic.

Newly stopped source identities (all Refused on the open-tag rule):

- `src/compiler/checker.ts:15122:132`
- `src/compiler/checker.ts:15122:181`
- `src/compiler/checker.ts:21228:13`
- `src/compiler/checker.ts:31591:63`
- `src/compiler/checker.ts:31591:110`
- `src/compiler/checker.ts:31605:56`
- `src/compiler/checker.ts:31605:96`
- `src/compiler/checker.ts:33557:43`
- `src/compiler/checker.ts:35692:41`
- `src/compiler/checker.ts:39208:13`
- `src/compiler/checker.ts:39709:17`
- `src/compiler/checker.ts:41379:65`
- `src/compiler/checker.ts:41382:43`
- `src/compiler/checker.ts:42447:19`
- `src/compiler/checker.ts:42483:38`
- `src/compiler/checker.ts:42608:19`
- `src/compiler/checker.ts:47909:48`
- `src/compiler/checker.ts:47911:23`
- `src/compiler/checker.ts:53408:94`
- `src/compiler/checker.ts:53412:47`
- `src/compiler/checker.ts:53413:43`
- `src/compiler/checker.ts:53419:43`
- `src/compiler/expressionToTypeNode.ts:1046:61`
- `src/compiler/factory/nodeConverters.ts:126:86`
- `src/compiler/factory/utilities.ts:467:33`
- `src/compiler/factory/utilities.ts:1054:48`
- `src/compiler/moduleSpecifiers.ts:904:33`
- `src/compiler/moduleSpecifiers.ts:911:55`
- `src/compiler/transformers/classFields.ts:1818:76`
- `src/compiler/transformers/classFields.ts:2544:48`
- `src/compiler/transformers/classFields.ts:2544:95`
- `src/compiler/transformers/declarations.ts:624:106`
- `src/compiler/transformers/declarations.ts:625:43`
- `src/compiler/transformers/declarations.ts:1112:25`
- `src/compiler/transformers/declarations.ts:1128:25`
- `src/compiler/transformers/declarations.ts:1141:25`
- `src/compiler/transformers/declarations.ts:1153:25`
- `src/compiler/transformers/declarations.ts:1165:25`
- `src/compiler/transformers/declarations.ts:1176:25`
- `src/compiler/transformers/esDecorators.ts:549:94`
- `src/compiler/transformers/esDecorators.ts:549:114`
- `src/compiler/transformers/esDecorators.ts:1295:66`
- `src/compiler/transformers/esDecorators.ts:1296:57`
- `src/compiler/transformers/esDecorators.ts:1298:44`
- `src/compiler/transformers/jsx.ts:262:99`
- `src/compiler/transformers/jsx.ts:319:99`
- `src/compiler/utilities.ts:2140:37`
- `src/compiler/utilities.ts:2142:36`
- `src/compiler/utilities.ts:2142:72`
- `src/compiler/utilities.ts:2142:120`
- `src/compiler/utilities.ts:3608:45`
- `src/compiler/utilities.ts:11916:47`
- `src/compiler/utilities.ts:11917:47`
- `src/compiler/utilities.ts:12046:92`

These are source-level scan regressions, not whole-unit lowering regressions. The refusal scanner continues its child walk after each finding, so absence of a base finding at these points is not explained merely by an earlier stop. A different node of the same property access can already have been refused at another column; this ledger counts diagnostic source identities, not distinct runtime evaluations.

One further new open-tag finding is at `src/compiler/moduleSpecifiers.ts:906:35`, which already had a refused unchecked cast on the base and still has it on the head. It is an additional reason at an already blocked point, not a newly stopped site. No reason became a panic.

## Units whose first stop changed

0 timed units changed their first lowering stop.

First observations including refusal scans (separate from first lowering stops): 7 changed owners.

| Scan/unit owner | Base first observation | Head first observation |
| --- | --- | --- |
| `src/compiler/factory/utilities.ts:465:1` createExpressionForObjectLiteralElementLike | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/factory/utilities.ts:466:55` | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/factory/utilities.ts:467:33` |
| `src/compiler/factory/utilities.ts:1015:1` tryGetPropertyNameOfBindingOrAssignmentElement | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/factory/utilities.ts:1023:53` | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/factory/utilities.ts:1054:48` |
| `src/compiler/transformers/declarations.ts:259:1` transformDeclarations | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/transformers/declarations.ts:358:180` | Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where SourceFile is read at `src/compiler/transformers/declarations.ts:380:21` |
| `src/compiler/utilities.ts:2134:1` isModuleAugmentationExternal | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/utilities.ts:2140:42` | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/utilities.ts:2140:37` |
| `src/compiler/utilities.ts:3604:1` isEmptyStringLiteral | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/utilities.ts:3606:22` | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/utilities.ts:3608:45` |
| `src/compiler/utilities.ts:3953:1` hasExpandoValueProperty | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/utilities.ts:3957:11` | NotYet: writing a checked object field without its source-slot type certificate at `src/compiler/binder.ts:1109:17` |
| `src/compiler/utilities.ts:12033:1` isPrimitiveLiteralValue | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/utilities.ts:12046:97` | Refused: an object refinement using an open numeric enum as a literal tag at `src/compiler/utilities.ts:12046:92` |

## Unit lowering time

These are single sequential runs on the same box. CPU includes synchronous dependency work; totals sum the meter unit records and exclude loading and shared file setup. Timing differences are observations, not a benchmark confidence interval.

| Sum | Base seconds | Head seconds | Delta seconds |
| --- | ---: | ---: | ---: |
| wall_seconds | 572.126311 | 539.507429 | -32.618882 |
| cpu_seconds | 660.789528 | 628.583962 | -32.205566 |

Ten largest absolute wall-time movers:

| Unit | Base wall s | Head wall s | Delta wall s | Delta CPU s |
| --- | ---: | ---: | ---: | ---: |
| `src/compiler/utilities.ts:12033:1` | 3.873158 | 1.206374 | -2.666784 | -2.677553 |
| `src/compiler/utilities.ts:12306:1` | 1.831479 | 0.308276 | -1.523204 | -1.544319 |
| `src/compiler/utilities.ts:12059:1` | 1.801427 | 0.478520 | -1.322907 | -1.330905 |
| `src/compiler/utilities.ts:12221:1` | 1.618193 | 0.307685 | -1.310507 | -1.320372 |
| `src/compiler/utilities.ts:12067:1` | 1.696317 | 0.415580 | -1.280737 | -1.293243 |
| `src/compiler/utilities.ts:11028:1` | 2.473301 | 1.333214 | -1.140087 | -1.161085 |
| `src/compiler/utilities.ts:12439:1` | 1.371477 | 0.309335 | -1.062142 | -1.075611 |
| `src/compiler/utilities.ts:12204:1` | 1.223802 | 0.173196 | -1.050605 | -1.061839 |
| `src/compiler/utilities.ts:11494:1` | 1.419670 | 0.399265 | -1.020406 | -1.033535 |
| `src/compiler/utilities.ts:11051:1` | 1.389663 | 0.472571 | -0.917093 | -0.928049 |

## Peak RSS

| Run | Peak KiB | Peak GiB | Exit |
| --- | ---: | ---: | ---: |
| base | 6705744 | 6.395096 | 0 |
| head | 6853484 | 6.535992 | 0 |

Peak RSS is Linux `getrusage(RUSAGE_CHILDREN).ru_maxrss` for one census subprocess in a fresh Python wrapper. The image has no `/usr/bin/time`; its failed launch did not start a census. The cgroup memory limit was 17179869184 bytes.

## Method and provenance

Read the complete `stage3/meter/twice-daily.sh` and `stage3/meter/README.md`. Setup used the requested GOPROXY and timeout. Adaptation used the base snapshot's apply script and package lock, with upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Stage3 inputs and dependency gitlinks are identical across the pair. Both compiler snapshots use cohere `06af2e399c6c1fb93fd570b774027942747400b4` and its TypeScript pin `d92d9bfee114c80be2c375d72edae966176e3a4f`.

For each pinned compiler, generated its scratch latent overlay with `stage3/census/latent/make_overlay.py`, ran gofmt, and built `./stage3/census/latent/tool` using `go build -buildvcs=false -overlay=...`. The ordinary sources were unchanged. Base ran to completion before the head build/run. Both used:

```sh
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=... latent-census /workspace/scratch/adapted/src/compiler latent.jsonl
```

The observed header reports `checker_rejected: false`. The overlay retains its conventional measurement label, "measured on a checker-rejected program". No backend or usable IR output was enabled. Unit identities and eligibility were checked across streams. Raw observations and timing streams are retained compressed beside this report. All source hashes, per-unit times, first-stop transitions and per-site deltas are in `DELTA.json`.

## Accessor name rule

4 distinct head open tag sites stopped on a positive accessor-name query (2 were also base open tag sites).

| Name | Unique sites |
| --- | ---: |
| `pos` | 4 |

All four decisions are code-between checks, rather than access-path checks. They cover two flow references at `src/compiler/utilities.ts:11916:47` and `src/compiler/utilities.ts:11917:47`, each diagnosed both at the access and at its property name. Two of the four sites were already base open-tag sites; two are newly stopped sites. The program declares getter `pos` at `src/compiler/sourcemap.ts:477`. No other name rejects a measured open-tag site.

A positive-control fixture copied from the head accessor test records getter name `n` at the access-path gate. Its scan findings match the original head binary, and it verifies distinct property-name diagnostic and parent-reference positions. See `evidence/probe-audit.log`.

Attribution used an additional head scratch overlay that records positive `accessorName` queries while `enumTagRefinementProven` is active. This observes actual short-circuit decisions in access-local path checking and code-between checking. The supplementary scan-only replay retained the complete program and original scans for files with head open-tag findings. It omitted unit lowering and ran sequentially after both measured runs; its selected scan findings were canonically byte-identical to the corresponding original head findings. Names can overlap at one site. Its time is excluded from the paired timings. The trace and exact instrumentation script are retained beside this report.
