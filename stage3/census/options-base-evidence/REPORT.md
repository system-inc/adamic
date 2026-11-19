# OptionsBase scout for #6c4swvt piece 3

- Adamic main was `32d927c43c08d2df0cd7f891ffa82942c8fa85ca`.
- Library's record-to-fixed head was `bc3285c066a8914ffa34ddd047cafd0793d0c4bb`. The ancestry check exited 1, with merge base `7ed9da77f6125271b18de835f5391abf056dc274`.
- Both worktrees used cohere `5360753e19e4417675a90e92a1823bea66b13918` and its TypeScript checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. Both recursive submodule updates exited 0.
- The meter's shallow `v6.0.3` TypeScript clone and pin check both exited 0. The pin was `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Readonly index probe

Stock `tsc --ignoreConfig --strict --noEmit readonly-index-probe.ts` exited 0. A prior invocation without `--ignoreConfig` exited 1 with TS5112 because its working directory contained a tsconfig. Stock `checker.getIndexInfosOfType(checker.getTypeAtLocation(parameter))` gives one string index with `isReadonly: true` for `readOption` and one writable index for `probe`'s source. Adamic's full latent probe exited 0 with zero checker diagnostics and reports these stops.

- At the call, `a Narrow seen as Readonly<Wide> (fixed objects and records have different storage; copy explicitly)`.
- At the index access, `a value of type Readonly<Wide>, a type with an index signature that isn't a string record`.

The census exit records stream completion. The two findings are lowering stops. Adamic's `recordElement` excludes a readonly string index from ordinary record storage on both pins.

## Receiver trace

| Adapted site | Expression and receiver | Writes and road |
| --- | --- | --- |
| 1994:100 | `watchOptions \|\| (watchOptions = {})` goes to `parseOptionValue.options: OptionsBase`. | The initializer can replace the caller's local binding with a fresh object. `parseOptionValue` writes `options[opt.name]` in eight branches. Two handle config-only null and false. Six handle normal number, boolean, string, list, custom type and null. The validators and custom/list parsers compute values or add diagnostics. Road (b) needs a writable contract for the actual input and the option key/value pairs it handles. A readonly parameter would forbid these stores. |
| 2712:79 | The inner `impliedCompilerOptions` goes to `serializeCompilerOptions.options: CompilerOptions`. The outer `assign` source resolves as `object \| undefined`. | The local accumulator receives each computed value before serialization. Serialization reads it through `serializeOptionBaseObject`, which writes a fresh map. `optionMapToObject` creates a new object via `Object.fromEntries`. `assign` checks own keys and writes `t[p] = arg[p]` into `config.compilerOptions`. Its rest argument is read. Road (a) describes the serializer input. Road (b) would have to describe the raw accumulator, whose producer uses a key-specific computed-option table exposed through a string index. Annotating the serialized result as `CompilerOptions` would overstate its named fields. Upstream `--ignoreConfig --showConfig --target es5 src/compiler/core.ts` exited 0 and returned the implied `"module": "commonjs"`. The compiler option is numeric `ModuleKind` before serialization. |
| 2792:38 | `serializeWatchOptions.options` goes to `serializeOptionBaseObject.options: OptionsBase`. | The receiver is enumerated and read. The ownership test only reads. Path and custom-value conversions write a fresh result map and create new arrays where needed. Road (a) fits the access pattern. |
| 3275:54 | `watchOptions` goes to `handleOptionConfigDirTemplateSubstitution.options: OptionsBase \| undefined`. | The helper reads the input. Its nested `setOptionValue` writes `(result ??= assign({}, options))[option.name]` into a fresh shallow copy. The list and map helpers write to `list.slice()` and `assign({}, mapLike)` respectively. The no-change path returns the original object. Road (a) fits the input reads and must preserve the original-object return path. |

The receiving declarations are internal. The v6.0.3 public `typescript.d.ts` exposes `CompilerOptions` and `WatchOptions`, while the helper declarations and `OptionsBase` stay internal.

The stock checker API found no declared interface or class in the compiler program extending `OptionsBase`, `CompilerOptions`, or `WatchOptions`. A readonly parameter view leaves the original declarations intact. Its equivalent guard uses `checker.getTypeAtLocation(parameter)`, each non-undefined union constituent, and `checker.getIndexInfosOfType(type)` to require `isReadonly` on the effective string index. For a declaration-level change, follow `checker.getBaseTypes` transitively and reject any derived type whose effective string index is writable. A readonly receiver does not certify an existing function that exports the same object under a writable type.

The stage3 owner search found `99-correctness-axolotl-b` on the parse expression inside `parseOptionValue`, `81-readonly-option-values` on casts inside the serialization and substitution workers, and the whole `serializeOptionBaseObject` owner pins in `99-correctness-axolotl-j`. `99-correctness-core-assign` delegates the whole `hasProperty` owner to adaptation 40 and preserves `assign`. `99-correctness-axolotl-c` has member pins on the internal `TSConfig` interface.

Both whole applies exited 0. The first stock checks exited 1 with 65 diagnostics per tree because the isolated trees did not resolve Node's declarations. Both stock reruns used lockfile-pinned `@types/node` through an explicit `typeRoots`. Each checked 79 roots with zero diagnostics and exited 0. The resolved argument query exited 0. The first single-file census attempts started before whole apply finished and exited 0 against partially adapted trees. The retained runs use the completed trees and all compiler roots with the walk filtered to `commandLineParser.ts`.

## Pinned full census

The two full runs used their own lowerers built from the meter overlays. Both used `CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1`, set `LATENT_ONLY_FILE` to the adapted parser's absolute path, and pointed the census binary at the adapted `src/compiler` directory. Main and library each exited 0 and ended with a relation trailer. Both headers report `checker_rejected: false` and zero diagnostics. The driver also prints a generic measurement label that says “checker-rejected program.”

The normalized site inventories compare equal (exit 0). The full parser inventories contain 195 units each, with 441 findings on main and 438 on library.

- At 1994:100, the scan reports `a value of type WatchOptions seen as OptionsBase, which can write TsConfigSourceFile | CompilerOptionsValue where CompilerOptionsValue is read`. It reports a second variant with `WatchOptions | undefined` as the source. Lowering also finds `an assignment result with different storage` at 1994:117.
- At 2712:79, the scan reports `NotYet` with `a Record<string, CompilerOptionsValue> seen as CompilerOptions (fixed objects and records have different storage; copy explicitly)`. At 2712:12 it separately finds `reading config`.
- At 2792:38, the scan reports `a value of type WatchOptions seen as OptionsBase, which can write TsConfigSourceFile | CompilerOptionsValue where CompilerOptionsValue is read`.
- At 3275:54, the scan reports `a value of type WatchOptions | undefined seen as OptionsBase | undefined, which can write TsConfigSourceFile | CompilerOptionsValue where CompilerOptionsValue is read`. The return assertion at 3275:12 separately reports `a cast the runtime can't check`.

Both streams were published as they finished. The main commit was `1ac616880` and the library commit was `d235fcd85`. Both pushes exited 0.

## Scratch readonly probe and stop

The scratch patch changed the two read-only workers' input types to `Readonly<OptionsBase>`, retaining `| undefined` on the substitution worker. Stock TypeScript checked 79 roots with zero diagnostics (exit 0). The parser JavaScript compared byte-identical with comments kept and removed (exit 0). The full parser census finished with its trailer (exit 0).

The probe replaced the refusals at 2792:38 and 3275:54 with storage stops at the same calls. The 1994 and 2712 findings persisted. The additional deltas are recorded with exact text, phase, location, unit and multiplicity in `readonly-parser-delta.json`.

- The storage stop in the compiler serializer at 2788:38 changed its target to the readonly view. The other template caller at 3095:9 acquired a storage stop.
- The serializer acquired a detached `hasOwnProperty` stop inside the call to its ownership helper.
- The substitution worker acquired a dynamic indexed-read stop at 3286:13, earlier than the previous switch stop at 3289:17. That displaced the switch stop and its rollback marker.
- Substitution acquired a writable-result storage stop at 3313:21. Two attempts to enumerate the readonly source inside `assign` stopped at core.ts:1350:25, one from the worker replay and one from its nested setter replay.
- The four added boundary records track statement rollbacks for the new lowering stops. The normalized delta totals 13 added and 5 removed findings, including those records.

The first scratch-copy command exited 1 because the upstream checkout has no repository-root tsconfig. The rerun preserved the upstream source layout and exited 0. The first scratch stock check exited 1 while Node declarations were unresolved. The explicit lockfile-pinned `typeRoots` rerun exited 0.

The scout stopped here. The two readonly views fail lowering and the writable parser needs its own contract. That leaves no established four-site adaptation from the proposed roads. The first-error entry pair, public declaration byte comparison, restoring mutant, product test, and lane were gated on a viable adaptation.
