# #y852c3c paired TypeScript census

Six full file pairs are verified. The focused full pair covers the remaining two files with observed base slot stops, and their unfiltered single-file pairs are underway. The original compiler-directory processes were killed at the worker’s memory limit before they emitted findings. The serial retry is underway. Both unfiltered `tsc.ts` first-error processes are still running. The focused first-error and scan-all pairs are complete.

## 1. Record values reaching Unknown

**Yes. The head's record node was Unknown in all 15 observed full-mode units that retained a checked slot read.** The graph tracked zero record writes in each. `recordValuesUnknown` named no write and returned “local or result has no tracked producer; no reaching allocation established.” The parser’s `AdamicJsonRecoveryValue` and `AdamicJsonRecoveryObject` supplied no observed cause. The focused full observer also found zero tracked writes and the same Unknown in `containerSeemsToBeEmptyDomElement` at `checker.ts:35213:5` and `verifyCompilerOptions` at `program.ts:4077:5`. The `createProgram` snapshot at `program.ts:1517:1` retained three slot-read observations. Its graph query stopped at “ir: virtual call has no target set” before it returned an Unknown decision or named a write. A supplemental query with class targets prepared is underway. The whole-entry allocation graph remains unobserved.

| Unit | Checked slot-read observations | Tracked writes |
| --- | ---: | ---: |
| `generateTSConfig` at `commandLineParser.ts:2849:1` | 4 | 0 |
| `getEffectiveTypeRoots` at `moduleNameResolver.ts:479:1` | 2 | 0 |
| `resolveTypeReferenceDirective` at `moduleNameResolver.ts:538:1` | 2 | 0 |
| `primaryLookup` at `moduleNameResolver.ts:662:5` | 1 | 0 |
| `secondaryLookup` at `moduleNameResolver.ts:695:5` | 1 | 0 |
| `getConditions` at `moduleNameResolver.ts:755:1` | 1 | 0 |
| `tryLoadModuleUsingRootDirs` at `moduleNameResolver.ts:1582:1` | 3 | 0 |
| `tryFile` at `moduleNameResolver.ts:2189:1` | 2 | 0 |
| `resolveFromTypeRoot` at `moduleNameResolver.ts:3332:1` | 2 | 0 |
| `usesWildcardTypes` at `utilities.ts:9027:1` | 1 | 0 |
| `createProgramDiagnostics` at `programDiagnostics.ts:89:1` | 1 | 0 |
| `fileIncludeReasonToRelatedInformation` at `programDiagnostics.ts:328:5` | 1 | 0 |
| `createResolutionCache` at `resolutionCache.ts:581:1` | 1 | 0 |
| `canWatchTypeRootPath` at `resolutionCache.ts:1704:5` | 1 | 0 |
| `fileIncludeReasonToDiagnostics` at `watch.ts:445:1` | 1 | 0 |

The observer queries the head allocation graph on a disposable copy of each unit snapshot. Its file streams reproduce all 504 parser, 640 resolver, 1,337 utilities, 107 program diagnostics, 198 resolution cache, and 77 watch findings at the original sites. See [the graph rows](evidence/record-values.json), [parser control](evidence/parser-probe-equivalence.json), [resolver control](evidence/resolver-probe-equivalence.json), and [utilities control](evidence/utilities-probe-equivalence.json), [program diagnostics control](evidence/programDiagnostics-probe-equivalence.json), [resolution cache control](evidence/resolutionCache-probe-equivalence.json), and [watch control](evidence/watch-probe-equivalence.json).

## 2. Slot-read sites and the Library comparison

### Resolver slot reads

The full `moduleNameResolver.ts` pair recorded 207 units on each side, 666 base findings, and 640 head findings. All thirteen occurrences of “a dictionary member narrowed to an array without an element-checked view” cleared at the eleven source sites below. Each read has a positive head observer receipt.

| Site in `moduleNameResolver.ts` | Owning unit | Base to head |
| --- | --- | --- |
| 480:9 | `getEffectiveTypeRoots` | Stopped to lowered. |
| 674:21 | `resolveTypeReferenceDirective`, `primaryLookup` | Stopped to lowered. |
| 699:18 | `resolveTypeReferenceDirective`, `secondaryLookup` | Stopped to lowered. |
| 780:36 | `getConditions` | Stopped to lowered. |
| 1583:10 | `tryLoadModuleUsingRootDirs` | Stopped to lowered. |
| 1595:27 | `tryLoadModuleUsingRootDirs` | Stopped to lowered. |
| 1634:31 | `tryLoadModuleUsingRootDirs` | Stopped to lowered. |
| 2190:10 | `tryFile` | Stopped to lowered. |
| 2196:20 | `tryFile` | Stopped to lowered. |
| 3333:10 | `resolveFromTypeRoot` | Stopped to lowered. |
| 3334:28 | `resolveFromTypeRoot` | Stopped to lowered. |

The guard at `480:9` also lets the following return read at `481:16` lower. The full run records no newly reached stop in this file. [Resolver pair](evidence/resolver-pair.json)

### Utilities slot read

The full pair measured 933 units on each side and 1,339 base findings against 1,337 head findings.

| Site | Owning unit | Expression | Base to head |
| --- | --- | --- | --- |
| `utilities.ts:9028:17` | `usesWildcardTypes` | `options.types` | Stopped to lowered. The head records no later stop in this unit. |

[Utilities pair](evidence/utilities-pair.json)

### Parser slot reads

The paired full census recorded 195 units on each side and 510 base findings against 504 head findings. The two findings with the reason “a dictionary member narrowed to an array without an element-checked view” both belonged to `generateTSConfig`.

| Site | Expression | Base to head |
| --- | --- | --- |
| `commandLineParser.ts:2871:9` | `options.lib` in the condition | Stopped to lowered. The head also lowers the following argument read at `2872:27`. |
| `commandLineParser.ts:2949:23` | `options[setting]` in `emitOption` | Stopped to lowered. The lost local `value` is available again at `2950:65`. The refusal “a cast the runtime can't check” at `2950:124` was present on both sides of the full run. |

The statements are measured within a still-blocked unit. See [the paired parser findings](evidence/parser-pair.json).

### Library’s twelve option-value refusals in the parser

All twelve remain stopped at the same refusal. The historical locations and full messages come from [the scan at `03c140ec`](https://github.com/system-inc/adamic/blob/c39d7b07852973aba01c15551d5164509f48d770/stage3/census/scan-all-sys-03c140ec/CommonJSBundle.raw.json) and are verified in [both full parser streams](evidence/library-pair.json).

| Site in `commandLineParser.ts` | Refusal in both compilers | Disposition |
| --- | --- | --- |
| 1901:20 | a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | Remains stopped at the same refusal |
| 2823:42 | a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |
| 2831:42 | a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |
| 2950:125 | a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | Remains stopped at the same refusal |
| 2950:134 | a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | Remains stopped at the same refusal |
| 2994:28 | a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |
| 3012:24 | a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |
| 3095:9 | a value of type MapLike<string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined> seen as OptionsBase \| undefined, which can write string \| number where string is read | Remains stopped at the same refusal |
| 3298:60 | a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |
| 3313:57 | a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |
| 3807:67 | a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |
| 3876:37 | a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | Remains stopped at the same refusal |

The focused first-error pair selected source units containing candidate reads. It recorded seven instances of the slot-read stop in the base and none in the head. The six resolver instances were at `480:9`, `674:21`, `780:36`, `1583:10`, `2190:10`, and `3333:10`. The other was `utilities.ts:9028:17`. All seven were gone and none of those units acquired a later stop in the focused pass. The two parser sites were visible in full mode. [Focused first-error pair](evidence/candidate-pair.json)

### Program diagnostics slot read

| Site | Owning unit | Base to head |
| --- | --- | --- |
| `programDiagnostics.ts:410:122` | `createProgramDiagnostics` and `fileIncludeReasonToRelatedInformation` | Stopped to lowered in both units. The head observer records both reads of `options.lib`, and the full pair records no newly reached stop. |

[Program diagnostics pair](evidence/programDiagnostics-pair.json)

### Resolution cache and watch slot reads

| Site | Owning unit | Base to head |
| --- | --- | --- |
| `resolutionCache.ts:1706:13` | `createResolutionCache` and `canWatchTypeRootPath` | Stopped to lowered in both units. The option read is `resolutionHost.getCompilationSettings().typeRoots`. |
| `watch.ts:543:151` | `fileIncludeReasonToDiagnostics` | Stopped to lowered. The head observer records the read of `options.lib`. |

Neither pair records a later stop newly reached in these units. See [resolution cache](evidence/resolutionCache-pair.json) and [watch](evidence/watch-pair.json).

### Checker and program slot reads from the focused full pair

Four source locations account for six findings in the focused full pair. The unfiltered single-file runs are still underway. [Focused full pair](evidence/discovery-pair.json)

| Site | Owning unit | Base to head |
| --- | --- | --- |
| `checker.ts:35214:17` | `checker.ts:35213:5` | Stopped to lowered. The observer also records the next read at `35214:41`. |
| `program.ts:1817:18` | `createProgram` | Moved to a later stop at `1822:17` after the `options.lib` guard lowers. The new rollback records “stage 0 can't lower a function returning U \| undefined yet” at `core.ts:33:17`. |
| `program.ts:4204:13` | `createProgram` and `verifyCompilerOptions` | Stopped to lowered in both owners. |
| `program.ts:4367:13` | `createProgram` and `verifyCompilerOptions` | Stopped to lowered in both owners. |


## 3. Other movements

- **Gone, NotYet during lowering.** “reading value” at `commandLineParser.ts:2950:65` clears when the declaration of `value` lowers.
- **Gone, rollback boundary during lowering.** “failed statement; state rolled back; continued at next statement or declaration” clears at `commandLineParser.ts:2871:5`, `2949:9`, and `2950:9`, at `utilities.ts:9028:5`, at `programDiagnostics.ts:410:21` in two units, at `resolutionCache.ts:1706:9` in two units, and at `watch.ts:543:45`. It also clears at `moduleNameResolver.ts:480:5`, `674:17` in two units, `699:13` in two units, `780:5`, `1583:5`, `1595:5`, `1634:9`, `2190:5`, `2196:5`, `3333:5`, and `3334:5`. The paired JSON names both owners for each duplicated site.
- **Growth in conservative checked-view demand.** The new origins make `viewedFrom` return `Unknown` with zero known allocation sites in all fifteen measured unit snapshots. The array check uses that flag to demand every eligible array read in those snapshots. The known-site set remains empty. The six unfiltered full file pairs and focused first-error pair have zero new, changed, or reached findings. The focused full pair reaches the later rollback at `program.ts:1822:17`, after the earlier guard lowers. It loses the older rollbacks at `checker.ts:35214:9`, `program.ts:1817:13`, and `program.ts:4204:9` and `4367:9` in two owners each. The unfiltered census is still pending. [Graph observations](evidence/record-values.json)

### Whole-entry scan-all boundary

Both serial retries completed. All 4,716 events paired exactly, including 4,715 refusal-scan events and the same first production stop. Both stop during `lowering.declareModule` at `core.ts:261:1` with “overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined.” This boundary leaves the whole-entry allocation graph unobserved. [Scan pair](evidence/scan-pair.json) · [Base raw scan](evidence/base/scan.json.gz) · [Head raw scan](evidence/head/scan.json.gz)


## Inputs and evidence

Both sides used the same adapted TypeScript v6.0.3 source tree at upstream `050880ce59e30b356b686bd3144efe24f875ebc8`. The base worktree ran `git submodule update --init --recursive` and then `stage3/apply.sh` with the tag cloned at depth one, following the meter’s upstream bootstrap. The source digests and adaptation table are in [evidence](evidence/). The earlier scratch bootstrap hit a read-only cache directory after application. The measured preparation used a writable cache and produced byte-identical `src` files.

The two compiler worktrees were pinned to `19a80811d918e392290519cbcda54b8c8c07500d` and `ae0f4d8a8467749df78602d285e73c9f6e079c3c`, with their own latent, entry and scan-all overlays. Both pins use the same checker submodules. The first-error compiler-directory and entry runs used CommonJSBundle and `LATENT_ASSERT_NO_OUTPUT=1`. The full passes used the same format and guard, `LATENT_FULL=1`, and selected one file at a time in the loaded compiler project.

Pairing uses kind, phase, reason, and message with source positions stripped from the text. The table locations identify occurrences within their units. A gone finding is absent on the head. A changed finding has a head counterpart at the same site. A reached finding is a later stop first seen in a unit that already had a base stop. A new finding is a head-only stop without that progression. The paired JSON keeps the raw messages and coordinates.
