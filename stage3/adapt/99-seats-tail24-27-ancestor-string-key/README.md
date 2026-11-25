This JavaScript-changing adaptation retypes `forEachAncestorDirectoryStoppingAtGlobalCache` to `string` with no `as P`, and builds `getIsExcluded`'s `Path` key with `toPath` in `src/services/exportInfoMap.ts`, Adamic's road (a) for the tail pass's family 27 (#bbssdcy, Nov 24 2026 21:47:33Z); Compiler accepts it under the widened split if `toPath` returns its input on every directory `getDirectoryPath(path)` can hand that callback, and Adamic does if it can change any.

The precedent is `src/compiler/moduleSpecifiers.ts:738`. The structural pins cover the wrapper declaration including its JSDoc, the namespace import declaration, the symlink key statement, and all 15 symbol-resolved call expressions before and after. A wrapper reference outside its declaration, import or export specifiers, and those calls refuses with its source position. The adapter checks every pin before it writes either file. The `adapter-table-mentions` gate does not apply.

## Why `toPath` returns its input here

The language service captures the host base and case policy at `src/services/services.ts:1664` and `:1681`, passes them to the compiler host at `:1757-1762`, and creates the program at `:1847`. The compiler host's path becomes `SourceFile.path` at `src/compiler/program.ts:3561` and `:3682`. The in-tree tsserver provider inherits its host project's current directory and project service at `src/server/project.ts:2777-2786`; both projects use that service's canonicalizer at `:587`. The provider hook is internal at `src/services/types.ts:431`.

Disk paths are normalized before case canonicalization by `src/compiler/path.ts:764`. URLs take the normalized-absolute branch. In the dynamic-name fixture, the nonempty absolute program base roots the stored path. With an empty program base a relative filename can remain relative. The identity claim applies to rooted, normalized stored paths under the same stable case policy at program creation and lookup. The walker starts with the parent and repeatedly removes a directory component without climbing above the root (`src/compiler/path.ts:312,1097`). The case-sensitive producer preserves case. The folding producer uses the same filename folding function on both passes (`src/compiler/core.ts:1897,2399`). The output test checks each yielded directory on both kinds of host, including different bases for the second pass. The wrapper calls the callback at the cache stop, then stops (`src/compiler/moduleNameResolver.ts:3061-3063`).

An outside provider can use its own case policy and base. The exclusion walk passes both programs the same lookup host (`src/services/exportInfoMap.ts:467,472`). A folding lookup host can change a sensitive provider's uppercase path. An empty-base provider can store a relative path and the lookup host can root it. Both measured changes go to Adamic. The repository's built-in provider uses the shared pair.

## Base callers

The TypeScript checker resolves the 15 call expressions by wrapper symbol, following import aliases. Of those callbacks, only `getIsExcluded` originally received `Path`. The table gives the original directory and callback types.

| Call | Enclosing function | Resolved directory | Callback |
| --- | --- | --- | --- |
| `src/compiler/moduleNameResolver.ts:796` | `resolvePackageNameToPackageJson` | `string` | `ancestorDirectory: string` |
| `src/compiler/moduleNameResolver.ts:2403` | `getPackageScopeForPath` | `string` | `dir: string` |
| `src/compiler/moduleNameResolver.ts:3032` | `loadModuleFromNearestNodeModulesDirectoryWorker.lookup` | `string` | `ancestorDirectory: string` |
| `src/compiler/moduleNameResolver.ts:3308` | `classicNameResolver.tryResolve` | `string` | `directory: string` |
| `src/compiler/moduleSpecifiers.ts:702` | `getNearestAncestorDirectoryWithPackageJson` | `string` | `directory: string` |
| `src/compiler/moduleSpecifiers.ts:738` | `forEachFileNameOfModule` | `string` | `realPathDirectory: string` |
| `src/services/utilities.ts:3539` | `findPackageJsons` | `string` | `ancestor: string` |
| `src/services/utilities.ts:3555` | `findPackageJson` | `string` | `ancestor: string` |
| `src/services/utilities.ts:3575` | `getPackageJsonsVisibleToFile` | `string` | `ancestor: string` |
| `src/services/exportInfoMap.ts:445` | `isImportablePath` | `string` | `ancestor: string` |
| `src/services/exportInfoMap.ts:515` | `getIsExcluded` | `Path` | `dirPath: Path` |
| `src/services/stringCompletions.ts:1094` | `getCompletionEntriesForNonRelativeModules` | `string` | `ancestor: string` |
| `src/server/editorServices.ts:5617` | `getPackageJsonsVisibleToFile` | `string` | `directory: string` |
| `src/server/editorServices.ts:5627` | `getNearestAncestorDirectoryWithPackageJson` | `string` | `directory: string` |
| `src/server/packageJsonCache.ts:43` | `createPackageJsonCache` | `string` | `ancestor: string` |


## Rulings at the pin

- **An `as` upcast and `satisfies` are accepted only under Adamic's own sound relation (immutable covariance, mutable invariance, nominal class identity, safe function variance, optional-field compatibility), then erased; TypeScript assignability alone is not enough; `satisfies` keeps the operand's own type.** Accept. Oct 6 2026, @system_adamic; source: #bw3xg7c (rr5sf2fx, item 2) / docs/escape-hatches.md:9,16,31,122,176-178,1281-1303. Spec: What's in, Types (`satisfies`); The soundness rules. Test: internal/lower/proven_relations_test.go:45 TestProvenRelationsErase; internal/lower/proven_relations_test.go:9 TestProvenRelationsRefuse.
- **Phantom string brands are approved, brand-adding casts included and erased, because a void member has no run-time content; removing `undefined` stays a run-time check; `unknown` values and unproven literal refinements stay refused; brand mints get their own line in the checked report.** Accept, Check for removing undefined. Oct 13 2026, @system_adamic; source: #fzx7q7y (btmknhnh), #bw3xg7c (5axpcgxn). Spec: What's in, Types (phantom brand), written Nov 1. Test: internal/lower/view_unions_mixed_test.go:168 TestMixedUnionContractPhantomVoidIsUndefined.
- **An overload result that differs from the implementation's only by a phantom brand is accepted with no check; brand members typed `any` are not admitted and are adapted to `undefined`.** Accept. Oct 7 2026, @system_adamic; source: #bw3xg7c (dtpc5t2z). Spec: not in the spec. The adaptation was `void` until Oct 29 (the entry below).
- **`Path` and `__String` are strings at run time with compile-time-only brands; `__String`'s void arm is `never`; a value arriving as `__String` from an unproven source is checked as a string at that boundary.** Check at the boundary. Oct 8 2026, @system_adamic; source: #pvvzhmy (03sbzdb5, item 4), #bw3xg7c (sffnja0f). Spec: not in the spec. Test: NONE (not found).
- **Meter adapters make only meaning-preserving in-memory edits (optional `| undefined`, optional method signatures as callable properties outside live class methods, `return void 0` only where the declared return includes `undefined`); disk sources stay unchanged.** Meaning. Undated, docs; source: docs/tsc-strictness-adaptation.md:11-48, docs/tsc-strictness-next-layer.md:8-47. Spec: not in the spec. Test: cmd/adamic-meter/optional_test.go:14 TestAdaptOptionalPropertiesPreservesRuntimeAndDisk.

> **Family 27: road (a), retype the wrapper to `string` and build the `Path` key with `toPath` in `getIsExcluded`.** It's JavaScript-changing in services, so it falls under tonight's widened split: Compiler accepts it if the `test.cjs` and her reader show `toPath` returns its input on every directory `getDirectoryPath(path)` can hand it, on case-sensitive and case-folding hosts both. Stock tsc must be at 0 on all three graphs. If `toPath` can change any input, it comes to me.

The import already supplies `hostGetCanonicalFileName`, and the new `toPath` import is shadowed only in unrelated scopes. The declarations and ancestry helper bodies are loaded from their own snapshot refs in the Node test. The directly used path and case helpers are read from the same snapshots, and their supporting primitives come from the stock 6.0.3 package.

The `adapter-output-test` gate uses both snapshots. It checks the allowed source spans, replays the adapter from the before snapshot, and confirms a second application writes nothing and preserves bytes and modification times. It exercises the original and candidate exclusion functions and checks four drifted pins and an escaping wrapper alias. A second TypeScript program passes through the snapshot's own services enumeration in each provider case.

| Provider case | Measured behavior |
| --- | --- |
| Shared base and case policy, both host settings | Each yielded directory is unchanged. Lookup sequences and module results agree. |
| Same policy and an unrelated provider base, both host settings | The fixture uses an absolute provider base. The stored paths and walked directories are rooted and canonical under the shared case policy. Keys and results agree. |
| Folding provider and a sensitive host | The lookup host preserves the provider's already-folded directory. Lookup sequences and module results agree. |
| Sensitive provider and a folding host | `/provider/node_modules/UPPER` becomes `/provider/node_modules/upper` and `/provider/node_modules/Äİ` becomes `/provider/node_modules/äİ`. The before walk matches the recorded symlink key and excludes the provider file. The after walk misses that key and includes the file. |
| Empty-base provider with the lookup host's case policy | A relative stored path is rooted against the lookup base. The original includes the file and the candidate excludes it after an absolute-key symlink hit. |

With an empty provider base, the stored path is `rel/node_modules/UPPER/index.ts` on the sensitive host and `rel/node_modules/upper/index.ts` on the folding host. The candidate roots the relative ancestor against `/Work`, producing `/Work/rel/node_modules/UPPER` and `/work/rel/node_modules/upper`, respectively. The original enumeration includes the provider. The candidate hits the absolute symlink key and excludes it. In the main-program twin with an empty program and lookup base, keys and results agree.

## What the key edit reads, and when

The adapted key calls `host.getCurrentDirectory()` and, through `hostGetCanonicalFileName`, reads `useCaseSensitiveFileNames` once per lookup. The language service captures its base and case policy once at creation (`src/services/services.ts:1664,1681-1682`). The live-policy fixture keeps the sensitive path `/Work/node_modules/UPPER/index.ts` and switches the same host to folding before the walk. The old key hits and excludes the file. The new key `/work/node_modules/upper/` misses, the remaining ancestors miss, and the file is included. The host stability contract remains Adamic's decision.

Adamic is the accepter under the quoted split. The measured outside-host differences are, on this road, evidence for her review. Both case policies preserved the ordinary fixtures, including POSIX disk roots, drive roots, UNC paths, file URLs and dynamic filenames resolved by the program. The Unicode case keeps `İ`, `ı` and `ß` under TypeScript's filename folding rule, while it folds `Ä`.

All three stock TypeScript 6.0.3 graphs emit with zero diagnostics. The tsc graph has no services files and emits identical JavaScript. The typescript and tsserver CommonJS graphs change only line 347 of `services/exportInfoMap.js`, adding the `toPath` call and the two host reads to the existing key argument. The namespace require is unchanged. The compiler wrapper's JavaScript and every stripInternal declaration are byte-identical.
