# Paired depth-gate census — #rt5evf5

## Result

All six censuses completed normally within the 30-minute limit. Of the **40 historical sites**, head **lowers 4**, **stops at the identity bound at 34**, and **retains the real Map check at 2**. There are **0 cap sites** and **0 later-stop cases**.

The largest logged **pastDepth is 79**. There are **41 DEPTHGATE lines**, **0 above 1000**, and **28 DEPTHWALK lines**, all retained below and literally in the evidence. The supplied probe does not print exact comparison-node positions; those lines are attributed to their **containing timed units**, not guessed expression sites.

The expectation of no changes outside the 40 is **false**: `NotYet: reading last` at `utilitiesPublic.ts:2639:14` is **gone**. Seven rollback-boundary events are also gone (five distinct boundary sites). No non-target JSON finding is new, changed, or reached under the requested pairing key.

The largest wall-time increase is `utilities.ts`: **595.429 → 949.973 seconds (+59.5%)**, with peak RSS **1,791,304 → 1,726,568 KiB**. This is an observation of the enabled probe, including its instrumentation, rather than a claim about an uninstrumented implementation.

## Inputs, population, and method

* **Base / head 1:** `9d64615714e5beb154effd374bef22bcce5cfd8a`.
* **Head / head 2:** `54210cd4c9b850b7c2df13b30f4406c78451ea9b` (`optimizer/depth-gate-walk-probe`).
* Each entry tool was generated and built in its own isolated checkout, using that checkout's `make_overlay.py`, `entry_overlay.py`, and lowering source. Their distinct binary hashes are in every [run status](evidence/run-summary.json). The probe checkout is not merged into the report branch.
* Submodules were initialized at their gitlinks in both checkouts: cohere `5360753e19e4417675a90e92a1823bea66b13918`, checker TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Local clones supplied the same pins.
* **One shared adapted tree**, `/workspace/depthgate-adapted`, was produced successfully by **base's** `stage3/apply.sh`. Its upstream input was the cached `microsoft/TypeScript` `v6.0.3` checkout verified at `050880ce59e30b356b686bd3144efe24f875ebc8`, as `twice-daily.sh` requests. The pinned apply packages were installed first.
* Every run used `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_TAIL=1`, one **exact absolute** `LATENT_ONLY_FILE`, and the entry tool rooted at shared `src/tsc/tsc.ts`. Only head additionally set `ADAMIC_PROBE_DEPTHGATE=1`.
* Each header reports **full**, **CommonJSBundle**, **0 checker diagnostics**, and the same **81 source files**. Every stream has its normal `relation_report` completion trailer; no-output guards succeeded. Paired selected-file hashes match. [Source hashes](evidence/adapted-sources.sha256), [pins](evidence/build-pins.txt), [commands](evidence/commands.txt), and [runner](evidence/run-censuses.py) record the setup.

**Why 40, not 43:** the scout's 35 first-error sites plus 8 full-mode additions form 43 across its broader scope. Three of those sites are outside the requested files: `parser.ts:472:12`, `parser.ts:496:17`, and `factory/utilities.ts:1614:9`. The selected files contain **32 first-error + 8 additional = 40** historical sites: **29 utilitiesPublic, 7 utilities, 4 transformer utilities**. Exact source anchors in their enclosing functions map all 40 uniquely onto the new shared tree, with unchanged line/column coordinates; see [targets.json](evidence/targets.json).

The baseline population is still 40, but **38 use the renamed depth-gate words and 2 use the real Map-check words**. Those two are `utilities.ts:11649:41` and `:11715:72`; they already take the real Map path at base. This split is not a head-induced change.

Coordinates below are relative to `/workspace/depthgate-adapted/src/compiler/`.

## The 40 sites

Codes name the observed stop path:

* **G (base):** fixed-depth gate, `a generic comparison deeper than 4 positions that can reach a Map`.
* **I (head):** recursion-identity bound, with the same full reason: `a generic comparison deeper than 4 positions that can reach a Map`.
* **M:** the real `container_fit` representation check into `ir.Map`, `a Map or Set interface implemented by a value that isn't a Map or Set`. This labels the guard path; it does not establish that a runtime value is a FakeMap.
* **lowers:** the mapped comparison/statement no longer stops in the full census. This does not mean that the whole program is admitted or that usable IR was emitted.

| ID | Site | Measured owning unit(s) | Base | Head | Detail |
|---|---|---|---|---|---|
| S01 | `utilitiesPublic.ts:625:22` | `utilitiesPublic.ts:623:1` isEmptyBindingPattern | G | lowers | Owning units have no remaining findings. |
| S02 | `utilitiesPublic.ts:1035:30` | `utilitiesPublic.ts:1034:1` getModifiers | G | I | Gate words; recursion-identity bound. |
| S03 | `utilitiesPublic.ts:1082:31` | `utilitiesPublic.ts:1080:1` getJSDocTypeParameterTagsWorker | G | I | Gate words; recursion-identity bound. |
| S04 | `utilitiesPublic.ts:1111:37` | `utilitiesPublic.ts:1110:1` hasJSDocParameterTags | G | I | Gate words; recursion-identity bound. |
| S05 | `utilitiesPublic.ts:1116:35` | `utilitiesPublic.ts:1115:1` getJSDocAugmentsTag | G | I | Gate words; recursion-identity bound. |
| S06 | `utilitiesPublic.ts:1121:34` | `utilitiesPublic.ts:1120:1` getJSDocImplementsTags | G | I | Gate words; recursion-identity bound. |
| S07 | `utilitiesPublic.ts:1126:35` | `utilitiesPublic.ts:1125:1` getJSDocClassTag | G | I | Gate words; recursion-identity bound. |
| S08 | `utilitiesPublic.ts:1131:35` | `utilitiesPublic.ts:1130:1` getJSDocPublicTag | G | I | Gate words; recursion-identity bound. |
| S09 | `utilitiesPublic.ts:1136:35` | `utilitiesPublic.ts:1135:1` getJSDocPublicTagNoCache | G | I | Gate words; recursion-identity bound. |
| S10 | `utilitiesPublic.ts:1141:35` | `utilitiesPublic.ts:1140:1` getJSDocPrivateTag | G | I | Gate words; recursion-identity bound. |
| S11 | `utilitiesPublic.ts:1146:35` | `utilitiesPublic.ts:1145:1` getJSDocPrivateTagNoCache | G | I | Gate words; recursion-identity bound. |
| S12 | `utilitiesPublic.ts:1151:35` | `utilitiesPublic.ts:1150:1` getJSDocProtectedTag | G | I | Gate words; recursion-identity bound. |
| S13 | `utilitiesPublic.ts:1156:35` | `utilitiesPublic.ts:1155:1` getJSDocProtectedTagNoCache | G | I | Gate words; recursion-identity bound. |
| S14 | `utilitiesPublic.ts:1161:35` | `utilitiesPublic.ts:1160:1` getJSDocReadonlyTag | G | I | Gate words; recursion-identity bound. |
| S15 | `utilitiesPublic.ts:1166:35` | `utilitiesPublic.ts:1165:1` getJSDocReadonlyTagNoCache | G | I | Gate words; recursion-identity bound. |
| S16 | `utilitiesPublic.ts:1170:35` | `utilitiesPublic.ts:1169:1` getJSDocOverrideTagNoCache | G | I | Gate words; recursion-identity bound. |
| S17 | `utilitiesPublic.ts:1175:35` | `utilitiesPublic.ts:1174:1` getJSDocDeprecatedTag | G | I | Gate words; recursion-identity bound. |
| S18 | `utilitiesPublic.ts:1180:35` | `utilitiesPublic.ts:1179:1` getJSDocDeprecatedTagNoCache | G | I | Gate words; recursion-identity bound. |
| S19 | `utilitiesPublic.ts:1190:35` | `utilitiesPublic.ts:1189:1` getJSDocThisTag | G | I | Gate words; recursion-identity bound. |
| S20 | `utilitiesPublic.ts:1195:35` | `utilitiesPublic.ts:1194:1` getJSDocReturnTag | G | I | Gate words; recursion-identity bound. |
| S21 | `utilitiesPublic.ts:1200:35` | `utilitiesPublic.ts:1199:1` getJSDocTemplateTag | G | I | Gate words; recursion-identity bound. |
| S22 | `utilitiesPublic.ts:1204:35` | `utilitiesPublic.ts:1203:1` getJSDocSatisfiesTag | G | I | Gate words; recursion-identity bound. |
| S23 | `utilitiesPublic.ts:1210:40` | `utilitiesPublic.ts:1208:1` getJSDocTypeTag | G | I | Gate words; recursion-identity bound. |
| S24 | `utilitiesPublic.ts:1229:84` | `utilitiesPublic.ts:1228:1` getJSDocType | G | I | Gate words; recursion-identity bound. |
| S25 | `utilitiesPublic.ts:1339:38` | `utilitiesPublic.ts:1322:1` getEffectiveTypeParameterDeclarations | G | I | Gate words; recursion-identity bound. |
| S26 | `utilitiesPublic.ts:1342:20` | `utilitiesPublic.ts:1322:1` getEffectiveTypeParameterDeclarations | G | I | Gate words; recursion-identity bound. |
| S27 | `utilitiesPublic.ts:1357:28` | `utilitiesPublic.ts:1355:1` getEffectiveConstraintOfTypeParameter | G | I | Gate words; recursion-identity bound. |
| S28 | `utilitiesPublic.ts:2559:23` | `utilitiesPublic.ts:2556:1` hasJSDocNodes | G | I | Gate words; recursion-identity bound. |
| S29 | `utilitiesPublic.ts:2638:76` | `utilitiesPublic.ts:2637:1` hasRestParameter | G | lowers | Owning units have no remaining findings. |
| S30 | `utilities.ts:3095:20` | `utilities.ts:12296:1` getPropertyArrayElementValue<br>`utilities.ts:12304:1` getOptionsSyntaxByValue<br>`utilities.ts:3113:1` getTsConfigPropArrayElementValue | G | lowers | Owning units have no remaining findings. |
| S31 | `utilities.ts:4429:82` | `utilities.ts:4426:1` getNamespaceDeclarationNode | G | I | Gate words; recursion-identity bound. |
| S32 | `utilities.ts:4780:34` | `utilities.ts:4779:1` getJSDocOverloadTags | G | I | Gate words; recursion-identity bound. |
| S33 | `utilities.ts:10816:71` | `utilities.ts:10811:1` getContainingNodeArray | G | I | Gate words; recursion-identity bound. |
| S34 | `utilities.ts:11649:41` | `utilities.ts:11493:1` createNameResolver<br>`utilities.ts:11513:5` resolveNameHelper | M | M | Real Map-check path; unchanged from base. |
| S35 | `utilities.ts:11715:72` | `utilities.ts:11493:1` createNameResolver<br>`utilities.ts:11513:5` resolveNameHelper | M | M | Real Map-check path; unchanged from base. |
| S36 | `utilities.ts:11861:55` | `utilities.ts:11493:1` createNameResolver<br>`utilities.ts:11513:5` resolveNameHelper | G | I | Gate words; recursion-identity bound. |
| S37 | `transformers/utilities.ts:126:25` | `transformers/utilities.ts:124:1` containsDefaultReference | G | I | Gate words; recursion-identity bound. |
| S38 | `transformers/utilities.ts:127:17` | `transformers/utilities.ts:124:1` containsDefaultReference | G | lowers | Full-mode statement lowers; earlier `transformers/utilities.ts:126:25` still stops in this unit. |
| S39 | `transformers/utilities.ts:161:25` | `transformers/utilities.ts:153:1` getImportNeedsImportStarHelper | G | I | Gate words; recursion-identity bound. |
| S40 | `transformers/utilities.ts:175:118` | `transformers/utilities.ts:173:1` getImportNeedsImportDefaultHelper | G | I | Gate words; recursion-identity bound. |

The four sites that lower are `utilitiesPublic.ts:625:22`, `utilitiesPublic.ts:2638:76`, `utilities.ts:3095:20`, and `transformers/utilities.ts:127:17`. The first three have no remaining findings in their measured owning units. The fourth is a full-mode statement success inside a unit still blocked by the earlier target at `:126:25`.

Repeated owning contexts are retained. In particular, `utilities.ts:3095:20` is measured under three callers; the two real Map sites and `utilities.ts:11861:55` are each measured under both `createNameResolver` and `resolveNameHelper`. [All base/head target findings, messages, owning-unit records, and outcomes](evidence/site-outcomes.json) preserve those multiplicities.

### Cap sites, listed separately

**None.** Neither tool emits `a fit walk past 4096 distinct pairs` at any target or non-target finding in these runs. There are also no later-stop cases requiring a later reason to be reported.

## DEPTHGATE sizes

Largest N: **79**, at **head utilities stderr line 285**, in timed unit `utilities.ts:4426:1` (`getNamespaceDeclarationNode`), whose target context is S31, `utilities.ts:4429:82`. Its literal line is:

```text
DEPTHGATE pastDepth=79 stop="a generic comparison deeper than 4 positions that can reach a Map"
```

The complete frequency distribution counts **emitted walk lines**, not distinct source sites:

| pastDepth N | utilitiesPublic | utilities | transformer utilities | Total |
|---:|---:|---:|---:|---:|
| 3 | 0 | 3 | 0 | 3 |
| 5 | 1 | 0 | 0 | 1 |
| 8 | 1 | 0 | 0 | 1 |
| 22 | 0 | 0 | 1 | 1 |
| 43 | 1 | 0 | 0 | 1 |
| 48 | 1 | 0 | 0 | 1 |
| 52 | 0 | 1 | 0 | 1 |
| 57 | 1 | 0 | 0 | 1 |
| 58 | 2 | 0 | 0 | 2 |
| 59 | 2 | 0 | 0 | 2 |
| 62 | 0 | 1 | 0 | 1 |
| 74 | 0 | 0 | 1 | 1 |
| 76 | 20 | 2 | 0 | 22 |
| 77 | 0 | 0 | 2 | 2 |
| 79 | 0 | 1 | 0 | 1 |
| **Total lines** | **29** | **8** | **4** | **41** |

Of the 41 lines, **35 return the identity-bound reason** and **6 have `stop=""`**. Repeated instantiations account for the difference between walk-line and site counts. The two pre-existing real Map-check sites are classified from their census findings, rather than folded into this histogram. [Exact distribution TSV](evidence/pastDepth-distribution.tsv), [per-run probe summaries](evidence/probe-summary.json), and [every probe record](evidence/probe-lines.jsonl) are included.

**Every DEPTHGATE line above N=1000: none.** The [above-1000 appendix](evidence/above-1000.md) records that empty set.

## Every DEPTHWALK finding

There are **28**, including **22 in 11 units containing no historical target site**. They are not omitted because they are outside the 40. All 28 have `pastDepth=0` **at the instant printed**. Their `judged` values are 500 (17 lines), 1000 (8), and 1500 (3). `judged` includes shallow pairs; it is not the counter used by the 4096-pair cap. These observations alone do not identify new walk behavior at head, since base's probe logging is disabled.

The table lists every line. **P** = head utilitiesPublic, **U** = head utilities, **T** = head transformer utilities. The unit position is recovered from the next canonical timing record in the same stderr pipe; it is not an exact comparison position printed by the probe. Candidate historical sites in a unit are context only. All 28 literal lines, without further truncation, are in the [DEPTHWALK appendix](evidence/DEPTHWALK.md); all raw stderr bytes are retained per run.

| Run | stderr line | Containing unit site / name | judged | depth | pastDepth | Historical target context |
|---|---:|---|---:|---:|---:|---|
| P | 155 | `utilitiesPublic.ts:1540:1` isTypeOnlyImportDeclaration | 500 | 3 | 0 | — |
| P | 156 | `utilitiesPublic.ts:1540:1` isTypeOnlyImportDeclaration | 1000 | 4 | 0 | — |
| P | 158 | `utilitiesPublic.ts:1554:1` isTypeOnlyExportDeclaration | 500 | 3 | 0 | — |
| P | 159 | `utilitiesPublic.ts:1554:1` isTypeOnlyExportDeclaration | 1000 | 4 | 0 | — |
| U | 29 | `utilities.ts:961:1` aggregateChildData | 500 | 3 | 0 | — |
| U | 30 | `utilities.ts:961:1` aggregateChildData | 1000 | 4 | 0 | — |
| U | 31 | `utilities.ts:961:1` aggregateChildData | 1500 | 3 | 0 | — |
| U | 217 | `utilities.ts:3703:1` isInExpressionContext | 500 | 3 | 0 | — |
| U | 218 | `utilities.ts:3703:1` isInExpressionContext | 500 | 3 | 0 | — |
| U | 294 | `utilities.ts:4508:1` getSingleInitializerOfVariableStatementOrPropertyDeclaration | 500 | 3 | 0 | — |
| U | 318 | `utilities.ts:4853:1` getAssignmentTarget | 500 | 3 | 0 | — |
| U | 380 | `utilities.ts:5521:1` isNamedEvaluationSource | 500 | 3 | 0 | — |
| U | 805 | `utilities.ts:10811:1` getContainingNodeArray | 500 | 3 | 0 | S33 |
| U | 806 | `utilities.ts:10811:1` getContainingNodeArray | 500 | 3 | 0 | S33 |
| U | 818 | `utilities.ts:11027:1` isTypeDeclaration | 500 | 3 | 0 | — |
| U | 819 | `utilities.ts:11027:1` isTypeDeclaration | 500 | 3 | 0 | — |
| U | 820 | `utilities.ts:11027:1` isTypeDeclaration | 1000 | 4 | 0 | — |
| U | 855 | `utilities.ts:11493:1` createNameResolver | 500 | 3 | 0 | S34, S35, S36 |
| U | 856 | `utilities.ts:11493:1` createNameResolver | 1000 | 4 | 0 | S34, S35, S36 |
| U | 859 | `utilities.ts:11513:5` resolveNameHelper | 500 | 3 | 0 | S34, S35, S36 |
| U | 860 | `utilities.ts:11513:5` resolveNameHelper | 1000 | 4 | 0 | S34, S35, S36 |
| U | 863 | `utilities.ts:11908:5` useOuterVariableScopeInParameter | 500 | 3 | 0 | — |
| U | 864 | `utilities.ts:11908:5` useOuterVariableScopeInParameter | 1000 | 4 | 0 | — |
| U | 865 | `utilities.ts:11908:5` useOuterVariableScopeInParameter | 1500 | 3 | 0 | — |
| U | 868 | `utilities.ts:11939:9` requiresScopeChangeWorker | 500 | 3 | 0 | — |
| U | 869 | `utilities.ts:11939:9` requiresScopeChangeWorker | 1000 | 4 | 0 | — |
| U | 870 | `utilities.ts:11939:9` requiresScopeChangeWorker | 1500 | 3 | 0 | — |
| T | 16 | `transformers/utilities.ts:179:1` collectExternalModuleInfo | 500 | 3 | 0 | — |

This unit-level precision is the limit of the supplied probe output. No exact comparison-node position is guessed where the probe supplies none. All completed probe lines have a containing-unit attribution; none is left in an unfinished interval.

## Findings changed outside the 40

Pairing uses **kind, phase, reason, and message**, stripping `:line:column` substrings from reason/message. Identical normalized-key occurrences are paired at identical unit/site first, then within unit, then by normalized key. This prevents repeated generic-gate messages from assigning a removed boundary to an arbitrary occurrence. Event multiplicity is retained; a second view deduplicates `(kind, phase, where, reason, text)` across units.

**New: 0. Changed: 0. Reached: 0. Gone: 8 events / 6 distinct findings.** Specifically:

| Grade | Kind / phase | Site outside the 40 | Events | Full reason |
|---|---|---|---:|---|
| gone | Boundary / lowering | `utilitiesPublic.ts:625:9` | 1 | failed statement; state rolled back; continued at next statement or declaration |
| gone | Boundary / lowering | `utilitiesPublic.ts:2638:5` | 1 | failed statement; state rolled back; continued at next statement or declaration |
| gone | Boundary / lowering | `utilitiesPublic.ts:2639:5` | 1 | failed statement; state rolled back; continued at next statement or declaration |
| gone | NotYet / lowering | `utilitiesPublic.ts:2639:14` | 1 | reading last |
| gone | Boundary / lowering | `utilities.ts:3095:5` | 3 | failed statement; state rolled back; continued at next statement or declaration |
| gone | Boundary / lowering | `transformers/utilities.ts:127:5` | 1 | failed statement; state rolled back; continued at next statement or declaration |

The one non-Boundary change is **`reading last` at `utilitiesPublic.ts:2639:14`**, gone after `const last` at the preceding target `:2638:76` can lower. The boundary at `utilities.ts:3095:5` occurs under three owning units, so it contributes three gone events but one distinct finding. Boundary rows remain listed even when they accompany one of the 40 target comparisons at another column.

[Every changed event with its full reason, complete message, and owning unit](evidence/outside-target-changes.md) and the [structured pairing result](evidence/finding-comparisons.json) are included. The 28 bare DEPTHWALK probe observations are catalogued separately above: they do not carry the kind/phase/reason fields of a latent JSON finding and are not silently counted as unchanged JSON diagnoses.

Across all findings, including the targets, there are **14 gone events / 10 distinct findings**, with no new, changed, or reached JSON findings. The complete raw streams retain findings of every kind and phase, including scoped-tail findings and boundaries.

## Wall time and peak memory

Each run was a separate child process, executed sequentially. Wall time uses `time.monotonic`; peak RSS uses Linux `wait4(...).ru_maxrss`, in **KiB**. Times include graph loading, lowering, scoped tail, and enabled probe logging; builds/apply are excluded. These are single observations in the recorded order, not statistically repeated benchmarks.

| File | Base wall s | Head wall s | Head/base wall | Base peak RSS KiB | Head peak RSS KiB |
|---|---:|---:|---:|---:|---:|
| `utilitiesPublic.ts` | 200.275 | 208.964 | 1.043× | 1,124,284 | 1,102,760 |
| `utilities.ts` | 595.429 | 949.973 | 1.595× | 1,791,304 | 1,726,568 |
| `transformers/utilities.ts` | 83.902 | 80.464 | 0.959× | 1,044,884 | 1,048,020 |

Every run exited **0**, with `timed_out=false`. The runner was configured to send SIGTERM at **1800 seconds**, then SIGKILL after a 5-second grace period if necessary; neither action was needed. No partial output was substituted for a completed result.

[Per-run resource statuses](evidence/run-summary.json) also contain exact commands, flags, timestamps, binary/source hashes, CPU times, load averages, CPU quota, and memory limit. Raw run files are under `evidence/runs/`:

* [base utilitiesPublic stderr](evidence/runs/base-utilitiesPublic/stderr.log.gz) / [head utilitiesPublic stderr](evidence/runs/head-utilitiesPublic/stderr.log.gz)
* [base utilities stderr](evidence/runs/base-utilities/stderr.log.gz) / [head utilities stderr](evidence/runs/head-utilities/stderr.log.gz)
* [base transformer utilities stderr](evidence/runs/base-transformers-utilities/stderr.log.gz) / [head transformer utilities stderr](evidence/runs/head-transformers-utilities/stderr.log.gz)

Each directory contains its whole `latent.jsonl.gz`, whole stderr and stdout, and start/finish metadata. Stderr includes the canonical unit timing JSON records because `LATENT_TIMES=/dev/stderr` writes into the child's same pipe; the parent preserves that pipe byte-for-byte in its own per-run file. Separate extracted unit-time streams are also included.

## Scope and verification

Only this report and its evidence change on `typescript/depthgate-census`, whose source parent is **base**, not the probe head. No adaptation, ledger, or repository `internal/` file was changed. Canonical census overlays and generated tools are outside the source trees. No synthetic shape probe or implementation change was run or proposed for landing.

Validation checks the 40 exact anchors, all six completion trailers and zero-diagnostic headers, matching paired source hashes, the **4 + 34 + 2 = 40** outcome partition, both event and distinct-finding diffs, the **41-line** DEPTHGATE histogram, all **28** literal DEPTHWALK entries, gzip integrity, and unchanged source hashes. [SHA256SUMS](evidence/SHA256SUMS) covers the evidence.
