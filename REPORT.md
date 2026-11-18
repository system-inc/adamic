# #y852c3c paired TypeScript census

Both CommonJSBundle first-error pairs, all eight full single-file pairs, and both scan-all probes completed. The initial directory processes hit the worker’s memory limit before emitting findings. The directory totals use the completed memory-limited retries.

## 1. Record values reaching Unknown

**Yes. The head’s record node was Unknown in all 20 observed unit snapshots that held a checked slot read.** The graph tracked zero record-value writes in each. `recordValuesUnknown` named no write and returned “local or result has no tracked producer; no reaching allocation established.” The parser’s `AdamicJsonRecoveryValue` and `AdamicJsonRecoveryObject` supplied no observed cause.

The original `createProgram` probe stopped at “ir: virtual call has no target set.” Running the compiler’s `finishClassCalls` preparation on a disposable copy supplied the targets and let its graph query finish. It returned the same Unknown and zero writes. The prepared probe preserves all 340 findings from the focused program stream. The two overload units at `program.ts:1501:1` and `1516:1` also returned the same Unknown and zero writes after class-target preparation. Their prepared pass reproduced every one of the 932 selected findings from the unfiltered program stream. The whole-entry allocation graph remains unobserved. [Overload graph rows](evidence/head/class-target-overloads.log) · [Overload equivalence](evidence/overload-probe-equivalence.json) · [Prepared graph](evidence/head/class-target-probe.log) · [Program equivalence](evidence/class-target-probe-equivalence.json)


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
| `containerSeemsToBeEmptyDomElement` at `checker.ts:35213:5` | 2 | 0 |
| `verifyCompilerOptions` at `program.ts:4077:5` | 2 | 0 |
| `createProgram` at `program.ts:1501:1`, after `finishClassCalls` | 3 | 0 |
| `createProgram` at `program.ts:1516:1`, after `finishClassCalls` | 3 | 0 |
| `createProgram` at `program.ts:1517:1`, after `finishClassCalls` | 3 | 0 |

The observer queries the head allocation graph on a disposable copy of each unit snapshot. Its file streams reproduce all 504 parser, 640 resolver, 1,337 utilities, 107 program diagnostics, 198 resolution cache, and 77 watch findings at the original sites. See [the graph rows](evidence/record-values.json), [parser control](evidence/parser-probe-equivalence.json), [resolver control](evidence/resolver-probe-equivalence.json), and [utilities control](evidence/utilities-probe-equivalence.json), [program diagnostics control](evidence/programDiagnostics-probe-equivalence.json), [resolution cache control](evidence/resolutionCache-probe-equivalence.json), and [watch control](evidence/watch-probe-equivalence.json).

## 2. Slot-read sites and the Library comparison

The eight full pairs cleared 33 occurrences of the slot-read stop at 21 distinct source positions. Thirty occurrences lowered without reaching a new stop. Three occurrences of the same guard in `createProgram` moved to a later stop. All twelve historical Library refusals remain stopped.

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

### Program and checker slot reads

The unfiltered `program.ts` pair has 248 units per side, with 1,486 base findings and 1,467 head findings. All three `createProgram` declaration units lower the outer guard and reach the later call stop. [Program pair](evidence/program-pair.json)

| Site | Owning units | Base to head |
| --- | --- | --- |
| `program.ts:1817:18` | `createProgram` at `1501:1`, `1516:1`, and `1517:1` | Moved to a later stop at `1822:17` after the `options.lib` guard lowers. The rollback records “stage 0 can't lower a function returning U \| undefined yet” at `core.ts:33:17`. |
| `program.ts:4204:13` | The three `createProgram` units and `verifyCompilerOptions` at `4077:5` | Stopped to lowered in all four units. |
| `program.ts:4367:13` | The three `createProgram` units and `verifyCompilerOptions` at `4077:5` | Stopped to lowered in all four units. |
| `checker.ts:35214:17` | `containerSeemsToBeEmptyDomElement` | Stopped to lowered. The observer also records the next read at `35214:41`. |


The unfiltered checker pair measured 2,495 units and recorded 8,373 base findings against 8,371 head findings. The slot stop and its statement rollback are the entire checker delta. The selected checker observer reproduced all 1,198 findings in its coverage. [Checker pair](evidence/checker-pair.json) · [Checker probe equivalence](evidence/checker-probe-equivalence.json)

## 3. Other movements

- **Gone, NotYet during lowering.** “reading value” at `commandLineParser.ts:2950:65` clears when the declaration of `value` lowers.
- **Gone, rollback boundary during lowering.** “failed statement; state rolled back; continued at next statement or declaration” clears at `checker.ts:35214:9`, `commandLineParser.ts:2871:5`, `2949:9`, and `2950:9`, `utilities.ts:9028:5`, `programDiagnostics.ts:410:21` in two units, `resolutionCache.ts:1706:9` in two units, and `watch.ts:543:45`. It clears at `moduleNameResolver.ts:480:5`, `674:17` in two units, `699:13` in two units, `780:5`, `1583:5`, `1595:5`, `1634:9`, `2190:5`, `2196:5`, `3333:5`, and `3334:5`. It also clears at `program.ts:1817:13` in three units and at `program.ts:4204:9` and `4367:9` in four units each. The paired JSON preserves the owning units at every repeated site.
- **Reached, rollback boundary during lowering.** The same rollback reason is first reached at `program.ts:1822:17` in all three `createProgram` declaration units. Its message is “stage 0 can't lower a function returning U | undefined yet” at `core.ts:33:17`. This is the only rise in finding occurrences across the full file pairs. Across those pairs there are 68 gone findings, three reached, zero changed, and zero new.
- **Growth in conservative checked-view demand.** The checked slot reads add origins. In 17 snapshots, those reads are the entire set of origins. Each of the three `createProgram` units has 38 total origins, including three from slot reads. All 20 snapshot origin sets produce `viewed_unknown=true` with zero known allocation sites. The array check uses that flag to demand every eligible array read in the snapshot. The scan-all production boundary leaves whole-program demand unmeasured. [Graph observations](evidence/record-values.json)

The unfiltered `tsc.ts` first-error pair measured 81 source files and 4,568 units per side, with zero checker diagnostics. The base recorded 4,362 findings and the head recorded 4,355. The seven first-error slot stops listed above are the entire delta. There were zero new, changed, or reached findings. [Entry pair](evidence/entry-pair.json) · [Base stream](evidence/base/entry.jsonl.gz) · [Head stream](evidence/head/entry.jsonl.gz)

The unfiltered compiler-directory pair measured 79 source files and 4,561 units per side, with zero checker diagnostics. The base recorded 4,362 findings and the head recorded 4,355. Its entire delta is the same seven gone slot stops as the entry pair. [Directory pair](evidence/directory-pair.json) · [Base stream](evidence/base/directory.jsonl.gz) · [Head stream](evidence/head/directory.jsonl.gz)

### Whole-entry scan-all boundary

Both serial retries completed. All 4,716 events paired exactly, including 4,715 refusal-scan events and the same first production stop. Both stop during `lowering.declareModule` at `core.ts:261:1` with “overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined.” This boundary leaves the whole-entry allocation graph unobserved. [Scan pair](evidence/scan-pair.json) · [Base raw scan](evidence/base/scan.json.gz) · [Head raw scan](evidence/head/scan.json.gz)


## Inputs and evidence

Both sides used the same adapted TypeScript v6.0.3 source tree at upstream `050880ce59e30b356b686bd3144efe24f875ebc8`. The base worktree ran `git submodule update --init --recursive` and then `stage3/apply.sh` with the tag cloned at depth one, following the meter’s upstream bootstrap. The source digests and adaptation table are in [evidence](evidence/). The earlier scratch bootstrap hit a read-only cache directory after application. The measured preparation used a writable cache and produced byte-identical `src` files.

The two compiler worktrees were pinned to `19a80811d918e392290519cbcda54b8c8c07500d` and `ae0f4d8a8467749df78602d285e73c9f6e079c3c`, with their own latent, entry and scan-all overlays. Both pins use the same checker submodules. The first-error compiler-directory and entry runs used CommonJSBundle and `LATENT_ASSERT_NO_OUTPUT=1`. The full passes used the same format and guard, `LATENT_FULL=1`, and selected one file at a time in the loaded compiler project.

Pairing uses kind, phase, reason, and message with source positions stripped from the text. The table locations identify occurrences within their units. A gone finding is absent on the head. A changed finding has a head counterpart at the same site. A reached finding is a later stop first seen in a unit that already had a base stop. A new finding is a head-only stop without that progression. The paired JSON keeps the raw messages and coordinates. These are latent lowering observations. Both production scan-all probes stopped before whole-program lowering and exposed zero usable IR.

An edit to the scratch launcher while older shells were waiting on the tools caused shell errors after both full program streams, both entry streams, and the base directory stream closed. Each measurement binary wrote the expected file records and a final relation trailer. The original streams and shell diagnostics are in the evidence. The full checker tools separately returned exit code 0.
