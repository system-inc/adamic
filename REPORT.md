# #y852c3c paired TypeScript census

This checkpoint covers the verified full pairs for `commandLineParser.ts` and `moduleNameResolver.ts`. The compiler-directory and `tsc.ts` first-error pairs and the full `utilities.ts` pair are still running. The worker killed both initial scan-all processes after exceeding its memory allowance. A serial retry is underway.

## 1. Record values reaching Unknown

**Yes. The head's record node was Unknown in all 9 observed full-mode units that retained a checked slot read.** The graph tracked zero record writes in each. `recordValuesUnknown` named no write and returned “local or result has no tracked producer; no reaching allocation established.” The parser’s `AdamicJsonRecoveryValue` and `AdamicJsonRecoveryObject` supplied no observed cause. A source-filtered first-error probe also found the same graph result in `usesWildcardTypes` (`utilities.ts:9027:1`), with one checked slot read at `9028:17` and zero tracked writes. Whole-entry attribution is open until the broader passes finish.

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

The observer queries the head allocation graph on a disposable copy of each unit snapshot. Its file streams reproduce all 504 parser and 640 resolver findings at the original sites. See [the graph rows](evidence/record-values.json), [parser control](evidence/parser-probe-equivalence.json), and [resolver control](evidence/resolver-probe-equivalence.json).

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

### Utilities first-error finding

The focused first-error pair saw `usesWildcardTypes` stop at `utilities.ts:9028:17` on the base. On the head, the read of `options.types` lowered and the probe recorded no later stop in this unit. The full file pair is running. [Focused pair](evidence/candidate-pair.json)

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

## 3. Other movements

- **Gone, NotYet during lowering.** “reading value” at `commandLineParser.ts:2950:65` clears when the declaration of `value` lowers.
- **Gone, rollback boundary during lowering.** “failed statement; state rolled back; continued at next statement or declaration” clears at `commandLineParser.ts:2871:5`, `2949:9` and `2950:9`. It also clears at `moduleNameResolver.ts:480:5`, `674:17` in two units, `699:13` in two units, `780:5`, `1583:5`, `1595:5`, `1634:9`, `2190:5`, `2196:5`, `3333:5` and `3334:5`. The paired JSON identifies both owning units for each duplicate site.
- **Growth in the checked-view reach fallback.** Each observed head snapshot adds checked slot-view origins and has `viewed_unknown=true` with zero known viewed allocation sites. The array check treats that flag as demand for all eligible array reads in the snapshot. This is wider conservative demand than tracking the known allocation set. The first-error and scan-all runs have yet to establish the scope at entry. The two completed full pairs record zero new or changed findings. [Parser pair](evidence/parser-pair.json) · [Resolver pair](evidence/resolver-pair.json) · [Graph observations](evidence/record-values.json)

## Inputs and evidence

Both sides used the same adapted TypeScript v6.0.3 source tree at upstream `050880ce59e30b356b686bd3144efe24f875ebc8`. The base worktree ran `git submodule update --init --recursive` and then `stage3/apply.sh` with the tag cloned at depth one, following the meter’s upstream bootstrap. The source digests and adaptation table are in [evidence](evidence/). The earlier scratch bootstrap hit a read-only cache directory after application. The measured preparation used a writable cache and produced byte-identical `src` files.

The two compiler worktrees were pinned to `19a80811d918e392290519cbcda54b8c8c07500d` and `ae0f4d8a8467749df78602d285e73c9f6e079c3c`, with their own latent, entry and scan-all overlays. Both pins use the same checker submodules. The first-error compiler-directory and entry runs used CommonJSBundle and `LATENT_ASSERT_NO_OUTPUT=1`. The full passes used the same format and guard, `LATENT_FULL=1`, and selected one file at a time in the loaded compiler project.

Pairing uses kind, phase, reason, and message with source positions stripped from the text. The table locations identify occurrences within their units. A gone finding is absent on the head. A changed finding has a head counterpart at the same site. A reached finding is a later stop first seen in a unit that already had a base stop. A new finding is a head-only stop without that progression. The paired JSON keeps the raw messages and coordinates.
