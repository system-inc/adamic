# Container fit census (zrnkpdz)

Base d4c6b07a versus A plus statecopy 62ddd47969e5168e5a0e74a23498b85d144b2361. Logging measurements use 29de73556749a0374da691bb9599e0e024d1214a, the identical statecopy patch on f913c480. Report only; no optimizer feature is landed.

## Method and coverage

One adapted tree was produced by base stage3/apply.sh from TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. All source inventory entries are checked by source-manifest.json before each piece and again here; no adaptation was rerun. The actual inventory count is 746.

The fixed-array rule copies a fixed array by value only when its element copier returns the unchanged value. Struct/map/slice/IR elements retain the existing refusal. containerFits stays an ordinary owned snapshot map, not a programFact. The named struct-array mutant turns the guard red; the restored full latent Go test subtree passes; all three meter overlays compile. Evidence and the one-commit implementation are linked through evidence/latent-build and the statecopy branch.

The frozen plan has 26 pieces, 79 files and 4557 units. Each unit occurs exactly once. utilities.ts and utilitiesPublic.ts use disjoint contiguous unit chunks; all other files belong to one disjoint file group each. Every child has a 1,500 second ceiling with at most 10 seconds to terminate. Every piece was committed and ordinarily pushed on exit. Each piece loads the whole checked 79-root program. First-error mode, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_TIMES were enabled. No backend was run.

The reused selector admission audit verifies mode-off identity for base, head and log, fixture split first stops, and dropped/duplicate/misassigned-work controls. Generic instantiated witnesses remain scoped to the callers in each piece, exactly as observed; they are not normalized to claim a whole-run census. Repeated file scans are deduplicated only on identical finding records; lowering owners and all unit identities are checked. No new whole latent run was started. Main stage3/census/pair pairs the joined same-context streams.

Every completed utilitiesPublic chunk has byte-identical head and log record streams; individual raw SHA-256 values are in evidence/utilitiesPublic-byte-identity.json. The original whole-file p00 hit its 25 minute ceiling after partial timers, before any file finding record was emitted. That sizing attempt is saved and excluded from all paired findings and time sums. Every public unit is covered by the replacement chunk checks. The supplemental whole-file head attempt r00 also hit its hard 29 minute ceiling (1740.113 seconds); its matching logging run was not started. Whole-file byte comparison is incomplete. This intentional budget stop is archived and excluded from compiler panics, paired findings and timing sums. See evidence/whole-public-status.json. This check excludes the separate timing streams.

## First stops

| utilitiesPublic Map/Set first-stop site | Classification | Head first stops in its base attempting units |
| --- | --- | --- |
| src/compiler/utilitiesPublic.ts:620:26 | moved | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:620:9: checked predicate membership over an open tag domain |
| src/compiler/utilitiesPublic.ts:630:29 | moved | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:630:9: checked predicate membership over an open tag domain |
| src/compiler/utilitiesPublic.ts:638:29 | moved | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:638:12: checked predicate membership over an open tag domain |
| src/compiler/utilitiesPublic.ts:927:33 | moved | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:927:20: checked predicate membership over an open tag domain |
| src/compiler/utilitiesPublic.ts:1031:30 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1031:30: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1038:26 | moved | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1038:13: checked predicate membership over an open tag domain |
| src/compiler/utilitiesPublic.ts:1078:31 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1078:31: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1107:37 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1107:37: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1112:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1112:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1117:34 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1117:34: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1122:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1122:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1127:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1127:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1132:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1132:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1137:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1137:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1142:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1142:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1147:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1147:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1152:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1152:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1157:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1157:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1162:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1162:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1166:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1166:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1171:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1171:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1176:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1176:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1186:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1186:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1191:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1191:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1196:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1196:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1200:35 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1200:35: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1206:40 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1206:40: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1225:84 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1225:84: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:1319:26 | moved | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1319:9: checked predicate membership over an open tag domain |
| src/compiler/utilitiesPublic.ts:1353:28 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:1353:28: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:2634:76 | unchanged | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:2634:76: a Map or Set interface implemented by a value that isn't a Map or Set |
| src/compiler/utilitiesPublic.ts:2639:38 | moved | 1 unit(s): NotYet src/compiler/utilitiesPublic.ts:2639:18: checked predicate membership over an open tag domain |

Measured 32 matching utilitiesPublic sites (brief reference: 20) and 1 binder roots (brief reference: 50). The reference 20 find sites and 50 ordinary binder roots were not reproduced. This table enumerates all measured public Map/Set first-stop sites, without selecting an arbitrary 20. Ordinary roots are distinct from latent attempting units: the saved 79-root ordinary runs have 73 core.ts:736:36 cast refusals, four readiness/generic stops and two accepted roots, with no binder.ts:427:31 first stops. The cause of the reference-count discrepancy is unresolved; these measurements do not establish outcomes for an unidentified reference cohort. Full unit mappings are in first-stops.json.

| Base binder.ts:427:31 attempting unit | Classification | Head first stop and reason |
| --- | --- | --- |
| src/compiler/binder.ts:422:1 | moved | NotYet src/compiler/binder.ts:2258:13: writing a checked object field without its source-slot type certificate |

## Regressions and already-blocked units

Main pair tool: 0 distinct-read regressions; 0 raw regressions. There are 36 changed first stops in already-blocked units, kept separately in blocked-first-stops.json. Complete pairing: latent-delta.json and latent-pair-report.md.

Panics: 0 base; 0 head.

No regressions under the pair tool definition. New reasons exposed in already-blocked units are not counted as finding-free-unit regressions.

## Cap rows

No cap-reason finding row was observed in either completed stream.
All 0 cap rows are preserved in cap-rows.json; logged cap-stop walks: 0.

## Pair counts

Percentiles use nearest rank across all 4339 logged walks, including zero counts. Totals count work across independent walks; they are not distinct program-wide pair counts.

| Count | Total | p50 | p90 | p99 | Max per walk |
| --- | ---: | ---: | ---: | ---: | ---: |
| judged | 37766 | 0 | 4 | 361 | 1084 |
| past | 2919 | 0 | 0 | 27 | 79 |

Largest kept per unit: 2168. Unattributed setup walks: 0. Every unit's largest kept is preserved in pair-counts.json.

| Top judged walk unit | Piece:line | Judged | Past | Kept |
| --- | --- | ---: | ---: | ---: |
| src/compiler/utilities.ts:11025:1 | p13:220 | 1084 | 0 | 2168 |
| src/compiler/utilitiesPublic.ts:1550:1 | q03:27 | 1084 | 0 | 1084 |
| src/compiler/utilitiesPublic.ts:1536:1 | q03:21 | 1084 | 0 | 1084 |
| src/compiler/utilities.ts:11025:1 | p13:219 | 722 | 0 | 1084 |
| src/compiler/utilities.ts:4506:1 | p07:8 | 722 | 0 | 1084 |

## Time: one sample on this box

Cloud host 6feab438e03f, CPU quota 4 cores, 16 GiB cgroup memory, Go 1.27.1 linux/amd64, Node v24.19.0. All measured children ran sequentially. Head LATENT_TIMES includes the requested log file append overhead. Timers include unit registration, synchronous dependency lowering and recovery; checker loading, scans and other shared setup are excluded. Repeated setup caused by partitioning is not a whole-run time claim.

| Common-unit LATENT_TIMES sum | Base | Logging head |
| --- | ---: | ---: |
| wall_seconds | 6832.387048 | 7152.774806 |
| cpu_seconds | 7398.835660 | 7740.175999 |

Common timed units: 4056; base-only: 0; head-only: 0. Complete per-unit sums and differences are in latent-times.json.

20 largest wall increases:

| Unit | Base wall s | Head wall s | Delta wall s | Delta CPU s |
| --- | ---: | ---: | ---: | ---: |
| src/compiler/transformers/declarations.ts:259:1 | 0.278186 | 27.136242 | +26.858055 | +29.097173 |
| src/compiler/factory/utilities.ts:786:1 | 0.087519 | 25.981182 | +25.893662 | +27.907281 |
| src/compiler/utilities.ts:3928:1 | 0.079986 | 24.339970 | +24.259983 | +26.453269 |
| src/compiler/utilities.ts:6920:1 | 0.078024 | 24.000146 | +23.922122 | +26.215259 |
| src/compiler/factory/utilities.ts:1008:1 | 0.087486 | 23.761302 | +23.673816 | +25.723716 |
| src/compiler/utilities.ts:3945:1 | 0.171987 | 23.807823 | +23.635836 | +25.305561 |
| src/compiler/utilities.ts:4764:1 | 0.136980 | 23.174059 | +23.037079 | +25.162074 |
| src/compiler/utilities.ts:4698:1 | 0.084179 | 21.563168 | +21.478989 | +23.398195 |
| src/compiler/binder.ts:422:1 | 0.176370 | 13.167579 | +12.991209 | +14.159238 |
| src/compiler/transformers/utilities.ts:866:1 | 0.190637 | 12.788563 | +12.597926 | +13.186706 |
| src/compiler/utilities.ts:2742:1 | 0.156278 | 11.191140 | +11.034862 | +11.597821 |
| src/compiler/transformers/destructuring.ts:218:1 | 23.760036 | 31.267690 | +7.507654 | +8.271883 |
| src/compiler/utilitiesPublic.ts:1190:1 | 22.183249 | 29.030487 | +6.847238 | +7.149036 |
| src/compiler/transformers/utilities.ts:471:1 | 24.702768 | 30.887856 | +6.185088 | +6.182049 |
| src/compiler/transformers/esnext.ts:774:1 | 23.967233 | 29.981645 | +6.014412 | +6.898090 |
| src/compiler/transformers/utilities.ts:876:1 | 23.378445 | 28.388556 | +5.010111 | +5.637866 |
| src/compiler/emitter.ts:1147:1 | 27.537195 | 32.491420 | +4.954225 | +6.397568 |
| src/compiler/utilities.ts:7560:1 | 21.987065 | 26.906392 | +4.919327 | +5.090960 |
| src/compiler/transformers/destructuring.ts:197:1 | 23.155563 | 28.059003 | +4.903440 | +4.699400 |
| src/compiler/program.ts:959:1 | 25.150832 | 30.007099 | +4.856267 | +5.111426 |

20 largest wall decreases:

| Unit | Base wall s | Head wall s | Delta wall s | Delta CPU s |
| --- | ---: | ---: | ---: | ---: |
| src/compiler/utilities.ts:3139:1 | 27.562840 | 22.684583 | -4.878257 | -5.708084 |
| src/compiler/utilities.ts:10624:1 | 28.952696 | 24.164410 | -4.788286 | -5.200910 |
| src/compiler/transformers/namedEvaluation.ts:139:1 | 16.668344 | 11.986129 | -4.682215 | -5.868270 |
| src/compiler/utilities.ts:4506:1 | 34.711805 | 30.082136 | -4.629669 | -4.786210 |
| src/compiler/utilitiesPublic.ts:1399:1 | 25.914418 | 21.461922 | -4.452495 | -4.651993 |
| src/compiler/utilities.ts:4932:1 | 26.884166 | 22.481102 | -4.403064 | -4.560364 |
| src/compiler/utilities.ts:2129:1 | 25.600488 | 21.401129 | -4.199359 | -3.973526 |
| src/compiler/utilities.ts:3422:1 | 28.069287 | 23.922195 | -4.147093 | -5.063743 |
| src/compiler/utilities.ts:7578:1 | 15.008472 | 11.083711 | -3.924761 | -4.335319 |
| src/compiler/utilities.ts:4472:1 | 14.685754 | 11.256469 | -3.429285 | -3.576803 |
| src/compiler/utilities.ts:11054:1 | 34.971345 | 31.562411 | -3.408934 | -3.724068 |
| src/compiler/utilities.ts:7472:1 | 25.890944 | 22.510572 | -3.380372 | -3.022026 |
| src/compiler/utilities.ts:1980:1 | 25.829611 | 22.454196 | -3.375415 | -3.270413 |
| src/compiler/utilitiesPublic.ts:870:1 | 25.469670 | 22.217801 | -3.251869 | -3.422536 |
| src/compiler/utilities.ts:8287:1 | 27.191038 | 24.078306 | -3.112732 | -2.993566 |
| src/compiler/transformers/classFields.ts:3343:1 | 27.879590 | 24.788051 | -3.091538 | -2.920698 |
| src/compiler/utilities.ts:7594:1 | 26.413006 | 23.322098 | -3.090909 | -3.176579 |
| src/compiler/utilities.ts:7468:1 | 24.469981 | 21.422960 | -3.047022 | -3.615153 |
| src/compiler/utilities.ts:3950:1 | 25.049521 | 22.008343 | -3.041178 | -3.277273 |
| src/compiler/utilities.ts:5687:1 | 24.734076 | 21.737485 | -2.996591 | -2.971746 |

The prior complete ordinary 79-root censuses use base d4c6b07a and unmodified A 1197ee88. The statecopy commit affects measurement generation only, so those observations were retained. All ordinary first stops matched, with zero ordinary regressions, fatal panics or cap rows. Ordinary process walls: 780.291357 seconds base and 726.362797 seconds head. Per-root sums: 777.552194 and 724.051171 seconds. These are also one sample on this box and are not latent unit times. See ordinary-summary.json.

## Evidence and limits

Raw piece records, LATENT_TIMES, append logs, process logs, binary hashes and per-piece publication checkpoints are under evidence/. Merged streams normalize only the common adapted prefix; original archives are lossless. Source hashes, the planned unit partition, exact timer coverage, all raw hashes and the utilitiesPublic byte check are independently asserted by this script before reporting. Header checker_rejected is false; the fixed measurement label is retained verbatim. These observations are first stops in isolated latent units, not successful compilation or an exhaustive blocker list.
