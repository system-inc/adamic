# #y852c3c paired TypeScript census

The full `commandLineParser.ts` pair is verified. The compiler-directory and `tsc.ts` first-error pairs, the scan-all runs, and the remaining file inventory are in progress. This checkpoint records the completed parser evidence and leaves the broader questions open.

## 1. Record values reaching Unknown

**Yes, in the measured head snapshot for `generateTSConfig` (`commandLineParser.ts:2849:1`).** The graph tracked zero record writes, so `recordValuesUnknown` named no write. Its exact result was “local or result has no tracked producer; no reaching allocation established.” The parser's `AdamicJsonRecoveryValue` and `AdamicJsonRecoveryObject` cannot be attributed as the cause of this observation. The full parser observer reproduced all 504 head findings at the same locations. See [the write probe log](evidence/head/record-probe-1.log) and [its equivalence check](evidence/parser-probe-equivalence.json).

## 2. Slot-read sites and the Library comparison
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

The full parser pair loses the NotYet “reading value” at `commandLineParser.ts:2950:65` after the declaration of `value` lowers. It also loses the three “failed statement; state rolled back; continued at next statement or declaration” boundary findings at `2871:5`, `2949:9` and `2950:9`. The head adds four checked slot-view observations in the `generateTSConfig` snapshot, including the second occurrence of `2949:23`. Its reach is Unknown with zero known viewed allocation sites, so the head's partial graph enters the conservative all-arrays fallback. Whole-entry allocation reach and the other file outcomes are pending.

## Inputs and evidence

Both sides used the same adapted TypeScript v6.0.3 source tree at upstream `050880ce59e30b356b686bd3144efe24f875ebc8`. The base worktree ran `git submodule update --init --recursive` and then `stage3/apply.sh` with the tag cloned at depth one, following the meter’s upstream bootstrap. The source digests and adaptation table are in [evidence](evidence/). The earlier scratch bootstrap hit a read-only cache directory after application. The measured preparation used a writable cache and produced byte-identical `src` files.

The two compiler worktrees were pinned to `19a80811d918e392290519cbcda54b8c8c07500d` and `ae0f4d8a8467749df78602d285e73c9f6e079c3c`, with their own latent, entry and scan-all overlays. Both pins use the same checker submodules. The first-error compiler-directory and entry runs used CommonJSBundle and `LATENT_ASSERT_NO_OUTPUT=1`. The full passes used the same format and guard, `LATENT_FULL=1`, and selected one file at a time in the loaded compiler project.

Pairing uses kind, phase, reason, and message with source positions stripped from the text. The table locations identify occurrences within their units. A gone finding is absent on the head. A changed finding has a head counterpart at the same site. A reached finding is a later stop first seen in a unit that already had a base stop. A new finding is a head-only stop without that progression. The paired JSON keeps the raw messages and coordinates.
