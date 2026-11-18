# Depth-gate submission census — #rt5evf5

## Result

All **12 runs completed normally** within 30 minutes. Both repeats give the same site and owning-unit findings. Of the **40 historical sites**, head **lowers 34**, **stops later at 4**, and **retains the real Map check at 2**. There are **0 identity-bound sites** and **0 cap sites**, including outside the target population.

Of the probe head `54210cd4`’s **34 identity-bound sites**, **30 now lower** and **4 reach later refusals**. All 34 lose their old gate refusal, but the four later cases are not counted as successful lowering.

**Growth outside the 40 exists:** `utilitiesPublic.ts:2559:13` newly reaches `a value of type JSDocArray | undefined`; the existing filter-overload refusal at `core.ts:261:1` and checked-field-write refusal at `binder.ts:1031:17` gain one owning context each. Two boundary messages change and two additional rollback boundaries are reached. These changes reproduce in both repeats. The other two measured files have only removals outside the 40.

For **utilities.ts**, base wall times are **620.994 / 578.356 seconds**; head is **826.809 / 798.073 seconds**. The two-run mean changes **+35.5%**. The timing table retains every wall/RSS observation.

## Inputs and reproducible method

Base is `c56b022995e1e960bad71626fd2074bc46aee485`; head is `26d208af47f81143ff1db2280a5b43292deb96ba`. Each canonical entry tool was generated and built from its own SHA in an isolated worktree, using that SHA’s `make_overlay.py` and `entry_overlay.py`. Head uses main’s `typeNesting` (limit 5) for the growing-generic path bound; the beyond-depth-4 walk and 4096-pair cap are measured as supplied. No lowering source was patched. The report branch starts at base and does not merge the candidate.

Both checkouts initialized their pinned submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, checker TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. One shared adapted tree, `/workspace/depthgate2-adapted`, was built successfully with **base’s `stage3/apply.sh`**, using TypeScript **v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`**, as the meter fetches it. Local verified clones supplied those commits.

Every run uses `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_TAIL=1`, and an exact absolute `LATENT_ONLY_FILE`, through the entry tool rooted at shared `src/tsc/tsc.ts`. **Neither side sets `ADAMIC_PROBE_DEPTHGATE`**; inherited `ADAMIC_PROBE_*` flags are removed. The canonical `LATENT_TIMES=/dev/stderr` records per-unit timings; its pipe is captured whole. There are no DEPTHGATE or DEPTHWALK lines. Every stream has a full-mode/CommonJSBundle header, zero checker diagnostics, the same 81 source files, and its normal `relation_report` trailer. The no-output guard succeeds in all 12 runs.

All runs are fresh child processes on the same machine, **sequential**, with no concurrent census or tool build. Per file the order is **base repeat 1, head repeat 1, head repeat 2, base repeat 2 (ABBA)**; file order is utilitiesPublic, transformer utilities, utilities. The runner measures end-to-end child wall time with a monotonic clock and child peak RSS with Linux `wait4.ru_maxrss` (KiB). Host, CPU/memory limits, timestamps, binary and selected-source hashes are in each status. The runner sends SIGTERM at 1,800 seconds and SIGKILL five seconds later if needed; **no timeout or kill was required**. Small report/progress reads and analysis occurred during the sequence; the machine was not benchmark-isolated.

Evidence: [runner](evidence/run-censuses.py), [commands](evidence/commands.txt), [pins](evidence/build-pins.txt), [all run statuses and counts](evidence/run-summary.json), [source manifest](evidence/adapted-sources.sha256), [repeat stability](evidence/repeat-stability.json), and [raw run directories](evidence/runs/). Each run directory retains the entire stream, stderr and stdout, compressed losslessly.

### Adapted tree compared with the earlier run

**Not byte-identical.** Comparing all directory names, symlink targets and regular-file bytes (excluding `.git`, timestamps and permission bits) against `/workspace/depthgate-adapted` finds exactly **three changed paths among 81,506 regular files**: `patch-set.md`, `src/compiler/parser.ts`, and `src/compiler/types.ts`. Base’s additional adaptations annotate `lastQueriedPosition: number` and change Node/NodeArray heritage from ReadonlyTextRange to TextRange; both are type-only source edits. The three selected files are byte-identical to their earlier versions. These are pre-existing base adaptations, not changes made by this report. The prior-probe outcome comparison is therefore across these stated input differences as well as the candidate change. [Exact old/new hashes and tree manifests](evidence/tree-comparison.json).

## The 40 at head

The population is unchanged: **29 utilitiesPublic + 7 utilities + 4 transformer utilities = 40**. The broader scout’s 35 first-error sites plus eight extras made 43; `parser.ts:472:12`, `parser.ts:496:17` and `factory/utilities.ts:1614:9` are outside these three files. All 40 uniquely map by source anchors in their enclosing functions, with unchanged coordinates. Base has **38 fixed-depth-gate sites + 2 real Map-check sites**, confirmed in both repeats. [Exact targets](evidence/targets.json).

Coordinates below are relative to `src/compiler/`. “Lowers” means the mapped comparison/statement no longer stops in this full census; it does not assert successful emission of the whole program. A new later refusal in the same owning attempt is conservatively classified **later**, even when the original comparison itself clears. “Real Map check” labels `container_fit`’s object-into-`ir.Map` guard path, not a new claim about runtime construction.

Full reasons used below:

* **G:** `a generic comparison deeper than 4 positions that can reach a Map` (fixed-depth at base; identity-bound at old probe head).
* **M:** `a Map or Set interface implemented by a value that isn't a Map or Set`.
* **L1:** `overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined` (Refused, lowering; `core.ts:261:1`).
* **L2:** `writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds` (NotYet, lowering; `binder.ts:1031:17`).
* **L3:** `a value of type JSDocArray | undefined` (NotYet, lowering; `utilitiesPublic.ts:2559:13`).

| ID | Site | Measured owning unit(s) | Function | Base | 54210cd4 | 26d208af (both repeats) | Reason / scope |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S01 | `utilitiesPublic.ts:625:22` | `utilitiesPublic.ts:623:1` | isEmptyBindingPattern | G | lowers | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S02 | `utilitiesPublic.ts:1035:30` | `utilitiesPublic.ts:1034:1` | getModifiers | G | identity bound | later | L1 at `core.ts:261:1` |
| S03 | `utilitiesPublic.ts:1082:31` | `utilitiesPublic.ts:1080:1` | getJSDocTypeParameterTagsWorker | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S04 | `utilitiesPublic.ts:1111:37` | `utilitiesPublic.ts:1110:1` | hasJSDocParameterTags | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S05 | `utilitiesPublic.ts:1116:35` | `utilitiesPublic.ts:1115:1` | getJSDocAugmentsTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S06 | `utilitiesPublic.ts:1121:34` | `utilitiesPublic.ts:1120:1` | getJSDocImplementsTags | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S07 | `utilitiesPublic.ts:1126:35` | `utilitiesPublic.ts:1125:1` | getJSDocClassTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S08 | `utilitiesPublic.ts:1131:35` | `utilitiesPublic.ts:1130:1` | getJSDocPublicTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S09 | `utilitiesPublic.ts:1136:35` | `utilitiesPublic.ts:1135:1` | getJSDocPublicTagNoCache | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S10 | `utilitiesPublic.ts:1141:35` | `utilitiesPublic.ts:1140:1` | getJSDocPrivateTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S11 | `utilitiesPublic.ts:1146:35` | `utilitiesPublic.ts:1145:1` | getJSDocPrivateTagNoCache | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S12 | `utilitiesPublic.ts:1151:35` | `utilitiesPublic.ts:1150:1` | getJSDocProtectedTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S13 | `utilitiesPublic.ts:1156:35` | `utilitiesPublic.ts:1155:1` | getJSDocProtectedTagNoCache | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S14 | `utilitiesPublic.ts:1161:35` | `utilitiesPublic.ts:1160:1` | getJSDocReadonlyTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S15 | `utilitiesPublic.ts:1166:35` | `utilitiesPublic.ts:1165:1` | getJSDocReadonlyTagNoCache | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S16 | `utilitiesPublic.ts:1170:35` | `utilitiesPublic.ts:1169:1` | getJSDocOverrideTagNoCache | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S17 | `utilitiesPublic.ts:1175:35` | `utilitiesPublic.ts:1174:1` | getJSDocDeprecatedTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S18 | `utilitiesPublic.ts:1180:35` | `utilitiesPublic.ts:1179:1` | getJSDocDeprecatedTagNoCache | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S19 | `utilitiesPublic.ts:1190:35` | `utilitiesPublic.ts:1189:1` | getJSDocThisTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S20 | `utilitiesPublic.ts:1195:35` | `utilitiesPublic.ts:1194:1` | getJSDocReturnTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S21 | `utilitiesPublic.ts:1200:35` | `utilitiesPublic.ts:1199:1` | getJSDocTemplateTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S22 | `utilitiesPublic.ts:1204:35` | `utilitiesPublic.ts:1203:1` | getJSDocSatisfiesTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S23 | `utilitiesPublic.ts:1210:40` | `utilitiesPublic.ts:1208:1` | getJSDocTypeTag | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S24 | `utilitiesPublic.ts:1229:84` | `utilitiesPublic.ts:1228:1` | getJSDocType | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S25 | `utilitiesPublic.ts:1339:38` | `utilitiesPublic.ts:1322:1` | getEffectiveTypeParameterDeclarations | G | identity bound | later | L2 at `binder.ts:1031:17` |
| S26 | `utilitiesPublic.ts:1342:20` | `utilitiesPublic.ts:1322:1` | getEffectiveTypeParameterDeclarations | G | identity bound | later | L2 at `binder.ts:1031:17` |
| S27 | `utilitiesPublic.ts:1357:28` | `utilitiesPublic.ts:1355:1` | getEffectiveConstraintOfTypeParameter | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S28 | `utilitiesPublic.ts:2559:23` | `utilitiesPublic.ts:2556:1` | hasJSDocNodes | G | identity bound | later | L3 at `utilitiesPublic.ts:2559:13` |
| S29 | `utilitiesPublic.ts:2638:76` | `utilitiesPublic.ts:2637:1` | hasRestParameter | G | lowers | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S30 | `utilities.ts:3095:20` | `utilities.ts:12296:1`<br>`utilities.ts:12304:1`<br>`utilities.ts:3113:1` | forEachPropertyAssignment | G | lowers | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S31 | `utilities.ts:4429:82` | `utilities.ts:4426:1` | getNamespaceDeclarationNode | G | identity bound | lowers | No stop at mapped site. Existing owning-unit refusals remain; full records in site-outcomes.json. |
| S32 | `utilities.ts:4780:34` | `utilities.ts:4779:1` | getJSDocOverloadTags | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S33 | `utilities.ts:10816:71` | `utilities.ts:10811:1` | getContainingNodeArray | G | identity bound | lowers | No stop at mapped site. Existing owning-unit refusals remain; full records in site-outcomes.json. |
| S34 | `utilities.ts:11649:41` | `utilities.ts:11493:1`<br>`utilities.ts:11513:5` | resolveNameHelper | M | real Map check | real Map check | M (unchanged) |
| S35 | `utilities.ts:11715:72` | `utilities.ts:11493:1`<br>`utilities.ts:11513:5` | resolveNameHelper | M | real Map check | real Map check | M (unchanged) |
| S36 | `utilities.ts:11861:55` | `utilities.ts:11493:1`<br>`utilities.ts:11513:5` | resolveNameHelper | G | identity bound | lowers | No stop at mapped site. Existing owning-unit refusals remain; full records in site-outcomes.json. |
| S37 | `transformers/utilities.ts:126:25` | `transformers/utilities.ts:124:1` | containsDefaultReference | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S38 | `transformers/utilities.ts:127:17` | `transformers/utilities.ts:124:1` | containsDefaultReference | G | lowers | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S39 | `transformers/utilities.ts:161:25` | `transformers/utilities.ts:153:1` | getImportNeedsImportStarHelper | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |
| S40 | `transformers/utilities.ts:175:118` | `transformers/utilities.ts:173:1` | getImportNeedsImportDefaultHelper | G | identity bound | lowers | No stop at mapped site. No remaining owning-unit findings. |

**Which former 34 lift:** S03, S04, S05, S06, S07, S08, S09, S10, S11, S12, S13, S14, S15, S16, S17, S18, S19, S20, S21, S22, S23, S24, S27, S31, S32, S33, S36, S37, S39, S40. **S02, S25, S26 and S28 reach later stops** (L1, L2, L2 and L3 respectively). S25/S26 share one owning attempt and one later checked-write finding; they are two target sites, not two distinct later diagnoses. The four sites that already lowered at 54210cd4 (S01, S29, S30, S38) still lower. S34/S35 retain M under both `createNameResolver` and `resolveNameHelper`. S36 now lowers at its mapped site while those pre-existing Map refusals remain in its owning units.

### Cap sites, listed separately

**None.** No target or non-target finding has `a fit walk past 4096 distinct pairs` in any of the 12 runs. Head also has no finding using G; there is no observed remaining identity-bound site. This census does not measure how close a successful walk gets to the cap.

## Findings outside the 40

Pairing is the same as `typescript/depthgate-census` `7276d4ef`: key **(kind, phase, reason, message)** after stripping `:line:column` positions in reason/message. Equal keys pair first at the same site and owning unit, then within unit, then across remaining occurrences. Unmatched same-site/phase/unit records are **changed**; remaining base records are **gone**; remaining head records are **reached** when their point lies in a base rollback-boundary span in the same unit, otherwise **new**. Both raw event multiplicity and distinct `(kind, phase, site, reason, message)` findings are retained. No target-site events enter this outside-target comparison, but their rollback boundaries do.

| Measured file | Counting view | New | Gone | Changed | Reached | Repeat check |
| --- | --- | --- | --- | --- | --- | --- |
| utilitiesPublic | owning-context events | 2 | 35 | 2 | 3 | same in repeats 1 and 2 |
| utilitiesPublic | distinct source findings | 0 | 35 | 2 | 3 | same in repeats 1 and 2 |
| utilities | owning-context events | 0 | 8 | 0 | 0 | same in repeats 1 and 2 |
| utilities | distinct source findings | 0 | 5 | 0 | 0 | same in repeats 1 and 2 |
| transformers-utilities | owning-context events | 0 | 4 | 0 | 0 | same in repeats 1 and 2 |
| transformers-utilities | distinct source findings | 0 | 4 | 0 | 0 | same in repeats 1 and 2 |

Counts are **per paired comparison**, not the sum of duplicate confirmations. All positive outside-target changes are in utilitiesPublic. The two **new raw events** are **additional owning contexts for existing distinct diagnoses**, so the distinct-finding table has zero new diagnoses. Growth is still present and is named below.

| Grade | Kind | Phase | Site | Owner | Full reason / change |
| --- | --- | --- | --- | --- | --- |
| changed | Boundary | lowering | `utilitiesPublic.ts:1339:5` | `utilitiesPublic.ts:1322:1` | failed statement; state rolled back; continued at next statement or declaration; message changes from G to L2 |
| changed | Boundary | lowering | `utilitiesPublic.ts:2559:5` | `utilitiesPublic.ts:2556:1` | failed statement; state rolled back; continued at next statement or declaration; message changes from G to L3 |
| reached | Boundary | lowering | `utilitiesPublic.ts:1348:9` | `utilitiesPublic.ts:1322:1` | failed statement; state rolled back; continued at next statement or declaration; message newly reached rollback boundary |
| reached | Boundary | lowering | `utilitiesPublic.ts:1036:9` | `utilitiesPublic.ts:1034:1` | failed statement; state rolled back; continued at next statement or declaration; message newly reached rollback boundary |
| reached | NotYet | lowering | `utilitiesPublic.ts:2559:13` | `utilitiesPublic.ts:2556:1` | a value of type JSDocArray \| undefined |
| new | NotYet | lowering | `binder.ts:1031:17` | `utilitiesPublic.ts:1322:1` | writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds (existing source diagnosis; additional owner) |
| new | Refused | lowering | `core.ts:261:1` | `utilitiesPublic.ts:1034:1` | overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined (existing source diagnosis; additional owner) |

Non-boundary removals outside the 40 are `reading last` at `utilitiesPublic.ts:2639:14` and `reading tag` at `utilitiesPublic.ts:1211:9`, `:1230:10` and `:1234:12`. The other gone records are rollback boundaries. Utilities has eight gone boundary events (five distinct sites); transformer utilities has four gone boundaries. The [complete outside-target appendix](evidence/outside-target-changes.md) lists **every grade with full before/after messages and owners**, including removals; the [machine-readable pairing](evidence/finding-comparisons.json) includes all events, matched counts, both repeats, and reached-span proofs.

## Wall time and peak memory

| File | Side | Repeat | Wall seconds | Peak RSS KiB | Peak RSS MiB | Started UTC | Exit |
| --- | --- | --- | --- | --- | --- | --- | --- |
| utilitiesPublic | base | 1 | 214.800 | 1,143,692 | 1116.9 | 2026-11-18T12:05:28Z | 0 |
| utilitiesPublic | base | 2 | 207.600 | 1,130,508 | 1104.0 | 2026-11-18T12:16:18Z | 0 |
| utilitiesPublic | head | 1 | 222.477 | 1,112,704 | 1086.6 | 2026-11-18T12:09:03Z | 0 |
| utilitiesPublic | head | 2 | 212.481 | 1,121,140 | 1094.9 | 2026-11-18T12:12:46Z | 0 |
| utilities | base | 1 | 620.994 | 1,724,956 | 1684.5 | 2026-11-18T12:25:33Z | 0 |
| utilities | base | 2 | 578.356 | 1,732,240 | 1691.6 | 2026-11-18T13:03:00Z | 0 |
| utilities | head | 1 | 826.809 | 1,735,264 | 1694.6 | 2026-11-18T12:35:54Z | 0 |
| utilities | head | 2 | 798.073 | 1,772,232 | 1730.7 | 2026-11-18T12:49:42Z | 0 |
| transformers-utilities | base | 1 | 90.484 | 1,026,616 | 1002.6 | 2026-11-18T12:19:46Z | 0 |
| transformers-utilities | base | 2 | 82.497 | 1,015,940 | 992.1 | 2026-11-18T12:24:11Z | 0 |
| transformers-utilities | head | 1 | 85.784 | 1,077,836 | 1052.6 | 2026-11-18T12:21:17Z | 0 |
| transformers-utilities | head | 2 | 87.663 | 1,013,080 | 989.3 | 2026-11-18T12:22:43Z | 0 |

| File | Base mean seconds | Head mean seconds | Mean wall change | Base RSS range MiB | Head RSS range MiB |
| --- | --- | --- | --- | --- | --- |
| utilitiesPublic | 211.200 | 217.479 | +3.0% | 1104.0–1116.9 | 1086.6–1094.9 |
| utilities | 599.675 | 812.441 | +35.5% | 1684.5–1691.6 | 1694.6–1730.7 |
| transformers-utilities | 86.490 | 86.723 | +0.3% | 992.1–1002.6 | 989.3–1052.6 |

There are only two observations per side. Public and transformer timings overlap across sides. Both utilities head observations exceed both base observations, so the utilities increase exceeds the observed within-side noise. The earlier **595.429 → 949.973 seconds** measured different base/head tools with probe logging enabled at head and the earlier adapted tree. It supplies context, not a controlled logging-only experiment. This rerun compares the submission head against its own base, using the same current tree and no probe logging.

## Audit scope

Only this report and its evidence are committed. No adaptation, ledger, `internal/`, checker or probe implementation change is made or landed. [Validation](evidence/validation.txt) records completion, source/binary pins, repeat agreement, archive integrity and protected-path checks; [SHA256SUMS](evidence/SHA256SUMS) covers every other evidence file. The full messages and unit/pass records are retained in the lossless streams and [site outcomes](evidence/site-outcomes.json).
