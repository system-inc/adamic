# Five any members: stopped at the public config type

Main: `de5c28cc11d8a12a76cefb13e2bb1c78a2bb4585`. Upstream TypeScript:
`050880ce59e30b356b686bd3144efe24f875ebc8`, version 6.0.3. Branch:
`typescript/adapt40-any5`. No retype is implemented in this change.

The real config result type already exists in the adapted tree:
`AdamicJsonRecoveryObject`, from adaptation 43. A scratch change to only
`parseConfigFileTextToJson.config` builds successfully, including the tests,
but emits `config?: AdamicJsonRecoveryObject` in public `typescript.d.ts`
without defining that internal type. Stock 6.0.3 reports TS2304 at public
line 9249. The unchanged control public declaration has zero diagnostics.
See [public-check.json](evidence/any5/public-check.json) and
[the exact public diff](evidence/any5/probe-api.diff).

Both the object interface and its recursive value alias are marked
`@internal`. Publishing their definitions would add public declarations beyond
the required diff of retyped lines. Substituting unknown despite the existing
real result type would not implement the instruction to use the real type
where the code shows one. This work stops at that decision, before adding a
rule class or sanction. No reading code, runtime expression, alias visibility,
later adaptation pin, lane manifest or sanction is changed.

## Measurement

The five requested adaptation-40 files and adaptation 91's README were read
whole before editing. A fresh control `stage3/apply.sh` passes end to end.
The stock API program loads `src/compiler/tsconfig.json`; LanguageService
`findReferences` identifies member references, contextual object types identify
initializers, and checker symbol identity identifies uses of the `raw` and
`extendsRaw` aliases. [members.json](evidence/any5/members.json) retains every
compiler reference with its statement, location and type. Destructuring is a
read of the member even when LanguageService labels its binding a write.

Each member was separately changed to unknown in a virtual CompilerHost,
without writing a source file. The repeated measurement after dependency
installation starts from zero stock compiler diagnostics. The exact added
errors are in [unknown-candidates.json](evidence/any5/unknown-candidates.json).
These compiler-only checks do not establish compatibility with readers in
other source projects. The complete builds below expose that distinction.
All locations below refer to the fully adapted control tree.

## ParsedCommandLine.raw, types.ts:7716

| Item | Measured contract |
| --- | --- |
| Declaration | Optional `raw?: any`. |
| Values put there | The shorthand at commandLineParser.ts:3104 forwards `parsedConfig.raw` through the local binding at :3080. Its producers are the three ParsedTsconfig paths below. Omission remains possible through the public optional property. |
| Every compiler read | commandLineParser.ts:3355 passes `config.raw` to `hasProperty(object, "references")`; watchPublic.ts:979 passes it to `canJsonReportNoInputFiles(any)`; tsbuildPublic.ts:1245 passes it to the same helper. |
| Type decision | Retain any. The inputs include the unvalidated public JSON/config path, rather than only the typed recovery-object path. Unknown fails at commandLineParser.ts:3355 with TS2345. That read requires a check before passing the value to `hasProperty`. An optional object type would still permit absence and does not prove that read safe. |

## readConfigFile.config, commandLineParser.ts:2267

| Item | Measured contract |
| --- | --- |
| Declaration | Optional `config?: any` in the return type literal. |
| Every producing path | :2269 returns `{ config: {}, error: textOrDiagnostic }` on read failure, or forwards `parseConfigFileTextToJson` on success. The real constructed values are recovery objects. |
| Every compiler read | No direct member read. The successful return at :2269 transfers the other function's result into this return type. |
| Real candidate | `AdamicJsonRecoveryObject`; `{}` satisfies its index signature and the successful conversion produces it. |
| Type decision | Retain any under the no-reader-edit rule. The real candidate passes the compiler-only check but the complete build fails in jsTyping. Unknown cannot satisfy those existing readers either. No reader is edited. |

The complete build with both config members given the real type fails at these
reads outside src/compiler, retained in [both-configs-build.log](evidence/any5/both-configs-build.log):

| Read | Why it requires editing with the real config type |
| --- | --- |
| src/jsTyping/jsTyping.ts:81, `Object.entries<string>(result.config)` | The optional config is not a proven string-valued map. |
| :87-88, `result.config?.simpleMap` and `Object.entries<string>(result.config.simpleMap)` | Presence does not establish a string-valued object. The argument at :88 fails. |
| :213-215, `manifest = ...config`, then four dependency fields passed through `flatMap(..., getOwnKeys)` | Manifest may be absent; each dependency field is a recovery value, not a proved map. The field reads, callback contract and downstream string-array consumption fail. |
| :245, `combinePaths(..., typingName, ...)` | The dependency-key pipeline no longer establishes string elements. |
| :271, `const manifest: PackageJson = result.config` | The optional recovery object is not proved to be PackageJson. |

The harness additionally forwards `readResult.config` to the existing any-input
parser in src/harness/compilerImpl.ts:40. The real-type build does not diagnose
that forwarding. The compiler inventory remains scoped to src/compiler as
requested; the additional failed build is evidence that no-reader-edit safety
cannot be inferred from that inventory alone.

## parseConfigFileTextToJson.config, commandLineParser.ts:2277

| Item | Measured contract |
| --- | --- |
| Declaration | Optional `config?: any` in the return type literal. |
| Every producing path | :2280 calls `convertConfigFileToObject`. Its adapted return type is `AdamicJsonRecoveryObject`. Object roots are converted; missing roots return `{}`; invalid roots recover the first object in an array or return `{}`. Object properties hold recursive recovery values or undefined. |
| Every compiler read/transfer | utilities.ts:7797 assigns `looseResult.config` to the existing any local `result`; commandLineParser.ts:2269 forwards the whole return value through readConfigFile. |
| Type decision | The real type fits the producing expression and unchanged source readers. The isolated real-type probe passes the complete build. Stop: the public declaration references the internal type without its definition. The repository declaration is left any, rather than sanctioning that broken API. |

The other source callers forward config to any-input parsers or formatters in
src/testRunner/unittests/config/tsconfigParsing.ts:22 and config/helpers.ts:16,
and read `config.compilerOptions` after an existing absence check in
src/harness/fourslashImpl.ts:348 and :350. These readers type-check unchanged
with the real type. The isolated probe restores readConfigFile.config to any,
so its unrelated jsTyping errors cannot mask this member's result.

## ParsedTsconfig.raw, commandLineParser.ts:3382

| Item | Measured contract |
| --- | --- |
| Declaration | Required `raw: any`. |
| Every producing path | :3425 stores `json || convertToObject(sourceFile!, errors)` on circularity; :3519 stores the public any input `json` from parseOwnConfigOfJson; :3600 stores the typed recovery object from parseOwnConfigOfJsonSourceFile. `convertToObject` still has a public any return, unlike the object-root converter. |
| Subsequent field writes | :3449-3451 write inherited include/exclude/files string arrays. :3453 conditionally writes compileOnSave. These do not validate all original raw fields or eliminate the arbitrary-input path. |
| Direct member reads | :3080 destructures raw; :3449-3451 use ownConfig.raw as the mutation base; :3453 reads and mutates compileOnSave; :3466 binds extendsRaw; :3469 reads ownConfig.raw[propertyName]. |
| Every raw-alias read | :3104 forwards the raw binding; :3111 reads compileOnSave; :3119 calls hasProperty(raw, "extends"); :3197 calls canJsonReportNoInputFiles(raw); :3234 calls hasProperty and reads raw[prop]; :3235 reads raw[prop] for isArray; :3236 reads and asserts raw[prop] as T[]. |
| Every extendsRaw-alias read | :3470 tests extendsRaw[propertyName]; :3471 passes it to map with a string callback; :3483 tests compileOnSave; :3484 assigns compileOnSave to the optional boolean result field. |
| Type decision | Retain any. A recovery-object-only contract does not cover the arbitrary public input and unrestricted convertToObject path. Unknown adds 16 diagnostics and requires reader edits. |

The reads requiring edits with unknown are :3111, :3119, :3234 (both the
hasProperty argument and indexing), :3235, :3236, the mutation bases at
:3449-3451, both compileOnSave bases at :3453, indexing at :3469-3471,
and compileOnSave at :3483-3484. Mere forwarding to an any slot/helper at
:3104 and :3197 does not require an edit. Each diagnostic is retained rather
than treating the alias binding as the only read.

## SolutionBuilderState.timerToBuildInvalidatedProject, tsbuildPublic.ts:424

| Item | Measured contract |
| --- | --- |
| Declaration | Required `timerToBuildInvalidatedProject: any`. |
| Every write | :545 initializes undefined; :2068 assigns `hostWithWatch.setTimeout(...)`; :2080 resets undefined. The host timer API returns any and is host-polymorphic, so a NodeJS.Timeout-only contract is not proved. |
| Every read | :2065 tests truthiness; :2066 forwards the value to the host's any-input clearTimeout; :2094 tests truthiness. |
| Type decision | Unknown fits all writes and unchanged reads. Its isolated virtual stock check has zero diagnostics. Not implemented because the public config decision triggered the requested stop before implementation. |

## Proof status and reproduction

The requested setup command was launched with GOPROXY and its exact timeout,
logging to /workspace/scratch/setup.log. It was still fetching the cohere
TypeScript submodule when the public API blocker was confirmed. Its task-owned
process group was terminated then, exit 143. Setup is not claimed complete.
Pinned stage3/api and stage3/typescript dependencies were installed separately.
Control apply and both successful builds ran before this stop; no oracle ran
concurrently with setup or any other workload because no oracle was started.

| Required proof | Observed status |
| --- | --- |
| Main control full apply | Exit 0, all adaptations including 90, 91 and 92 executed. |
| Control build including tests | Exit 0. |
| Both-config real-type probe build | Exit 1, unchanged jsTyping readers fail. |
| Parse-config-only real-type probe build | Exit 0, including tests. |
| Stock validation of emitted public API | Control zero diagnostics; probe TS2304, missing AdamicJsonRecoveryObject. |
| All emitted JavaScript | All 10 .js/.mjs/.cjs artifacts under built are byte-identical between control and the parse-config-only probe. This is a scratch-probe comparison, not proof of an implemented rule. |
| Public API diff | The probe changes exactly line 9249, config any to AdamicJsonRecoveryObject. That line is invalid without the omitted definition. No sanction is written. |
| Retyped full apply and 90/91/92 own reruns | Not run; no rule is implemented and no pins move. |
| Complete paired sharded oracle | Not run. No pass/fail identity equivalence is claimed. |
| API undo-retype mutant | Not run; no valid public retype or sanction has been implemented. |
| Generated shards check | `python3 -B stage3/lane/generate_shards.py --check` exits 0; shards.json is unchanged. |

[provenance.json](evidence/any5/provenance.json) records commands, exits and pins.
The retained scripts use stock TypeScript 6.0.3 on NODE_PATH:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/40-explicit-any/evidence/any5/measure.cjs CONTROL > members.json 2> measure.log
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/40-explicit-any/evidence/any5/candidates.cjs CONTROL > unknown-candidates.json 2> candidates.log
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/40-explicit-any/evidence/any5/public-check.cjs CONTROL PARSE_ONLY_PROBE > public-check.json 2> public-check.log
```

Make the scratch probe by copying the built control, replacing only
parseConfigFileTextToJson's `config?: any` with
`config?: AdamicJsonRecoveryObject`, then running `npm run build` there.
The two-config failed probe additionally changes readConfigFile's config
annotation. No reading expression changes in either probe.
