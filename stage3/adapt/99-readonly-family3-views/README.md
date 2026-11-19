# Readonly family 3 views, adaptation 99-readonly-family3-views

Adamic, task #k8ed1f2, family 3's roads 1 and 2, by Adamic's delegation:
readonly where tsc never writes, a shape earlier adaptations established. This
is type-only: the emitted JavaScript does not change. Every declaration it
touches is local to a function, file-local, or `@internal`, so the published
declaration files do not change, this directory carries no
`api-sanctions.json`, and `stage3/lane/sanctioned-api.json` is not
regenerated.

## Why

Family 3 is the refusal "a value of type A seen as B, which can write X where
Y is read". tsc hands a value to a slot whose type would let the holder write
something the value's own type never holds: a `Map<Path, ...>` seen as a
`Map<string, ...>` could take a plain string key. Adamic refuses each such view at its site.

Road 1 clears a site by making the view honest about what tsc does with it.
Where tsc only ever reads through the slot, the slot becomes a read-only view:
`readonly` on a property, `Readonly<T>` on a parameter, `ReadonlyMap` or
`ReadonlySet` for a collection, `readonly T[]` for an array, a readonly
constraint on a type parameter. A read-only view can't write, so the refusal
has nothing to stand on. One slot, cause 27's `cache`, takes its true key type
`Path` in place of a readonly view, because its body deletes from the map.

Road 2 clears a site by giving tsc the type it would otherwise infer wrongly.
At nine calls of `mutateMapSkippingNewValues` tsc infers the map's value type
as `unknown` from `noopOnDelete`, so each map is seen as a `Map<..., unknown>`
that could take any value. Each call gets its map's own key and value type as
explicit type arguments.

Together the two roads clear 69 of the 108 road 1 and 2 sites in family 3.
The causes, their sites and the reasoning behind each are in #k8ed1f2's cause
list (`k8ed1f2-causes.md`) and edit list (`edits-roads12.md`).

## The edit

`adapt.cjs` edits 34 slots in 12 files, all in `src/compiler`. Those 34 slots
are 34 of the edit list's 54 lines, one slot each; the other 20, causes 12, 13,
25 and 32 and cause 16's `directoryStructureHost` binding, are under Left out.
It parses each file with stock TypeScript 6.0.3 and finds each slot by name, never by
line, offset or whole-line text:

- a function's return type, named parameter or type parameter's constraint,
  an overload by its index among the declarations of that name, which must be
  one overload group of the expected size;
- an interface's or type alias's property, whose type changes or which gains
  `readonly` (its type must then be the one the ruling covers);
- an interface's method's return type;
- a `const` binding by its name in the function declaration nearest it, its
  initializer starting as upstream's does;
- a call by its callee and its first argument's text, in the function
  declaration nearest it.

The edit replaces the type node's span, inserts `readonly ` before a
property's name, inserts `: readonly Node[]` after `allDeclarations`' name, or
inserts `<...>` after the callee. Before any write it refuses unless every
slot is found exactly once, holds upstream's text or exactly its adapted
text, and its declaration is local to a function, file-local (a top-level
statement of a module, not exported), or carries `@internal` on itself or an
enclosing declaration. A slot already adapted keeps its bytes, so a second run
changes nothing. It prints one line per slot: its name, `edited` or `already
adapted`, where it is and why it is not public.

Sites are the meter's positions in the adapted tree, relative to
`src/compiler/`, as the cause list gives them. A cause's sites are listed once,
on its first slot; the slots under it clear them together.

| Cause | Slot | Upstream | Adapted | Sites cleared |
|---:|---|---|---|---|
| 4 | builder.ts `ReusableBuilderProgramState.semanticDiagnosticsPerFile` | `Map<Path, readonly ReusableDiagnostic[] \| readonly Diagnostic[]>` | `ReadonlyMap<...>` | 13: builder.ts:1709:90, 1717:28, 1747:62, 1771:62, 1936:62, 2035:62, 2113:62, 2153:62, 2214:62, 2299:41, 2300:37, 2340:41, 2341:37 |
| 4 | builder.ts `ReusableBuilderProgramState.emitDiagnosticsPerFile` | `Map<...> \| undefined` | `ReadonlyMap<...> \| undefined` | with the row above |
| 6 | builder.ts `assertSourceFileOkWithoutNextAffectedCall`, `state` | `BuilderProgramState` | `Readonly<BuilderProgramState>` | builder.ts:2037:55, 2115:55, 2215:51 |
| 6 | builder.ts `removeSemanticDiagnosticsOf`, `state` | `BuilderProgramState` | `Readonly<BuilderProgramState>` | builder.ts:820:41, 834:33, 871:33 |
| 6 | builder.ts `isChangedSignature`, `state` | `BuilderProgramState` | `Readonly<BuilderProgramState>` | builder.ts:965:29, 987:40 |
| 7 | types.ts `SyntacticTypeNodeBuilderContext.tracker` | `Required<Pick<SymbolTracker, "reportInferenceFallback">>` | the same, `readonly` | 11: checker.ts:10290:75, 6491:418, 6492:439, 6493:373, 7212:92, 8361:170, 9114:89, 9121:92, 9164:133, 9347:77, 9922:74 |
| 8 | tsbuildPublic.ts `createStateBuildOrder`, `mutateMapSkippingNewValues(state.configFileCache, ...)` | no type arguments | `<ResolvedConfigFilePath, ConfigFileCacheEntry>` | tsbuildPublic.ts:676:32 |
| 8 | the same, `state.projectStatus` | none | `<ResolvedConfigFilePath, UpToDateStatus>` | tsbuildPublic.ts:677:32 |
| 8 | the same, `state.builderPrograms` | none | `<ResolvedConfigFilePath, T>` | tsbuildPublic.ts:678:32 |
| 8 | the same, `state.diagnostics` | none | `<ResolvedConfigFilePath, readonly Diagnostic[]>` | tsbuildPublic.ts:679:32 |
| 8 | the same, `state.projectPendingBuild` | none | `<ResolvedConfigFilePath, ProgramUpdateLevel>` | tsbuildPublic.ts:680:32 |
| 8 | the same, `state.projectErrorsReported` | none | `<ResolvedConfigFilePath, true>` | tsbuildPublic.ts:681:32 |
| 8 | the same, `state.buildInfoCache` | none | `<ResolvedConfigFilePath, BuildInfoCacheEntry>` | tsbuildPublic.ts:682:32 |
| 8 | the same, `state.outputTimeStamps` | none | `<ResolvedConfigFilePath, Map<Path, Date>>` | tsbuildPublic.ts:683:32 |
| 8 | the same, `state.lastCachedPackageJsonLookups` | none | `<ResolvedConfigFilePath, Set<string> \| undefined>` | tsbuildPublic.ts:684:32 |
| 14 | types.ts `TypeChecker.getDiagnostics`, return (`@internal` member) | `Diagnostic[]` | `readonly Diagnostic[]` | 4: checker.ts:49753:24, 49759:63, 49765:62, 49768:20 |
| 14 | checker.ts `getDiagnostics`, return | `Diagnostic[]` | `readonly Diagnostic[]` | with the rows above |
| 14 | checker.ts `getDiagnosticsWorker`, return | `Diagnostic[]` | `readonly Diagnostic[]` | with the rows above |
| 15 | builder.ts `IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo`'s `signature` | `string \| false \| undefined` | the same, `readonly` | 3: builder.ts:1337:16, 1344:13, 1346:17 |
| 16 | watchUtilities.ts `createCachedDirectoryStructureHost`, `host` | `DirectoryStructureHost` | `Readonly<DirectoryStructureHost>` | watchPublic.ts:463:120 |
| 16 | program.ts `parseConfigHostFromCompilerHostLike`, `host` | `(CompilerHost \| ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }` | `Readonly<...>` of the same | watchPublic.ts:465:69 |
| 17 | checker.ts `discriminateTypeByDiscriminableItems`, `discriminators` | `(readonly [() => Type, __String])[]` | `readonly (readonly [() => Type, __String])[]` | 3: checker.ts:32666:21, 32690:21, 53778:101 |
| 21 | utilities.ts `createDiagnosticCollection`'s `getDiagnostics`, the implementation (third of three), return | `Diagnostic[]` | `readonly Diagnostic[]` | 2: utilities.ts:6125:20, two reasons at one position |
| 23 | core.ts `not`, `T`'s constraint | `unknown[]` | `readonly unknown[]` | 2: checker.ts:45156:68, 49042:113 |
| 26 | watchUtilities.ts `closeFileWatcherOf`, `T`'s constraint | `{ watcher: FileWatcher; }` | `{ readonly watcher: FileWatcher; }` | 2: watchUtilities.ts:527:32, 534:47 |
| 27 | resolutionCache.ts `removeResolutionsOfFileFromCache`, `cache` | `Map<string, ModeAwareCache<T>>` | `Map<Path, ModeAwareCache<T>>` | 2: resolutionCache.ts:1473:42, 1474:42 |
| 28 | checker.ts `getElaborationElementForJsxChild`, `nameType` | `LiteralType` | `Readonly<LiteralType>` | 2: checker.ts:21707:66, 21794:74 |
| 37 | checker.ts `combineIterationTypes`, `array` | `(IterationTypes \| undefined)[]` | `readonly (IterationTypes \| undefined)[]` | 1: checker.ts:45998:74 |
| 38 | watch.ts `createTabularErrorsDisplay`, `filesInError` | `(ReportFileInError \| undefined)[]` | `readonly (ReportFileInError \| undefined)[]` | 1: watch.ts:299:87 |
| 42 | emitter.ts `commentWillEmitNewLine`, `node` | `CommentRange` | `Readonly<CommentRange>` | 1: emitter.ts:3234:53 |
| 44 | moduleNameResolver.ts `ModuleResolutionState.host` | `ModuleResolutionHost` | the same, `readonly` | 1: moduleNameResolver.ts:2332:126 |
| 45 | program.ts `getReferencedFileLocation`, `ref` | `ReferencedFile` | `Readonly<ReferencedFile>` | 1: programDiagnostics.ts:191:71 |
| 46 | resolutionCache.ts `invalidateResolutions`, `resolutions` | `Set<ResolutionWithFailedLookupLocations> \| Map<string, ResolutionWithFailedLookupLocations> \| undefined` | `ReadonlySet<...> \| ReadonlyMap<string, ...> \| undefined` | 1: resolutionCache.ts:1576:39 |
| 47 | checker.ts `getObjectTypeInstantiation`'s `const allDeclarations` | none, inferred | `readonly Node[]` | 1: checker.ts:20774:140 |

Each row is one slot, and `adapt.cjs` lists the same 34. Where a cell says `...`, the elided text is the upstream cell's, unchanged.

Cause 6 clears 8 of its 12 sites; the other four are under Left out. Cause 21
clears the two reasons at utilities.ts:6125:20, and its third site is left
out. Cause 8's road 1 form, a set-free `map` parameter on the helper, is not
used: the nine type arguments clear the same nine sites without touching
`mutateMapSkippingNewValues`.

The directory sorts after 99-method-value-views and before
99-readonly-inferred-bindings, so it runs after every adaptation whose output
text it reads. Several earlier ones pin text that holds a slot here, and each
runs before this edit, so each pin reads its declarations as it expects:
99-correctness-binding-views pins resolutionCache.ts and its
`createResolutionCache`, which holds `removeResolutionsOfFileFromCache` and
`invalidateResolutions`; 99-correctness-receiver-printer pins `createPrinter`,
which holds `commentWillEmitNewLine`. Every slot's text in apply's tree of
Nov 17 is upstream's, and of the adaptations added since, none edits or pins
a declaration this one touches. The three that sort after it,
99-readonly-inferred-bindings, 99-readonly-script-target-entries and
99-wasm-crypto-import, neither read nor pin any of them: the first two edit
other declarations in builder.ts, checker.ts, program.ts and utilities.ts by
name, and the third edits sys.ts.

## Left out

39 of family 3's 108 road 1 and 2 sites, by cause:

- Cause 6, 4 sites (builder.ts:674:25, 892:21, 899:21 and 2075:21), the
  `state` parameters of `addToAffectedFilesPendingEmit` and
  `clearAffectedFilesPendingEmit`. Both bodies assign fields of `state`
  (`state.affectedFilesPendingEmit = ...`, `??=`,
  `state.programEmitPending = ...`), so `Readonly<>` would not type-check.
  `Omit<BuilderProgramState, "program">` would be true there; it was left for a
  multi-line edit of its own.
- Cause 12, 5 sites (executeCommandLine.ts:824:9, 859:66, 860:44, 890:62,
  891:39), the `options` parameters of `updateReportDiagnostic`,
  `shouldBePretty`, `createReportErrorSummary` and `createWatchStatusReporter`.
  `Readonly<CompilerOptions | BuildOptions>` is a mapped type, which Adamic
  stores as a record, so the 5 sites still stop, now as recordStorageView, and 7
  sites that did not stop do (executeCommandLine.ts:660, 672, 712, 935, 960,
  1028, 1050). Census with and without: J-fjw6z02, J-dh38kwx, J-7t4er8q,
  J-w3f32n6.
- Cause 13, 4 sites (utilities.ts:10005:9, 10019:64, 10020:63, 9999:16), the
  extension tables. With readonly inner arrays, Adamic's overload check makes a
  new stop at utilities.ts:10013: overload 1 of
  getSupportedExtensionsWithJsonIfResolveJsonModule, result
  `readonly (readonly Extension[])[]`, cannot be served by the implementation's
  result (lowering.declareModule). The constants feed those overloads, so the
  cause can't be cut in part: the four constants, getSupportedExtensions'
  overloads, getSupportedExtensionsWithJsonIfResolveJsonModule's parameters and
  returns, commandLineParser.ts's two `extensions` parameters and
  services/stringCompletions.ts's return all stay upstream's. The same census.
- Cause 16, 2 sites (watchPublic.ts:464:60 and 464:92), the
  `cachedDirectoryStructureHost || host` and the `host` on the initializer line
  of `createWatchProgram`'s local
  `directoryStructureHost: Readonly<DirectoryStructureHost>`. That binding
  cleared them but made 2 new stops, watchPublic.ts:465:75 and 493:90: a
  `Readonly<DirectoryStructureHost>` seen as `DirectoryStructureHost | undefined`,
  whose readonly fields become writable. The two host parameters stay. The same
  census.
- Cause 25, 2 sites (moduleSpecifiers.ts:614:129, 1290:21),
  `tryGetModuleNameFromPaths`' `paths`. `Readonly<MapLike<readonly string[]>>`
  is the same record storage, so both sites still stop and the slot clears
  nothing. The same census.
- Cause 9, 7 sites. `readonly Node[]` on `addDeprecatedSuggestion`'s
  `declarations` needs its worker to take `Node | readonly Node[]`, and the
  worker's `Array.isArray(declarations)` does not narrow `readonly Node[]` out
  of the false branch, so `getJSDocDeprecatedTag(declarations)` fails. That
  takes a body change (core's `isArray`), not an annotation.
- Cause 11, 6 sites, `ReusableRepopulateInfoChain.next` and
  `SerializedDiagnosticMessageChain`'s `next`. The edit list was written by
  line, and both declarations are the same line,
  `next?: ReusableDiagnosticMessageChain[];`. By name both are reachable, and
  `readonly` on them is true by the cause list's reading, but they were not in
  the 54 edits J-9z42vnh certified, so they wait for a run of their own.
- Cause 18, 3 sites, the local `target` in `getObjectTypeInstantiation`. It is
  handed to `instantiateAnonymousType`'s writable `AnonymousType` parameter, so
  a readonly view on the local only moves the site to that call, and
  `AnonymousType & { readonly target?: ... }` would not make the slot readonly,
  since an intersection's property is readonly only when every constituent's
  is. It needs a ruling.
- Cause 21, 1 site (utilities.ts:6075:27), `lookup`'s local `diagnostics`.
  By line it reads the same as `add`'s, which `insertSorted` writes. By name
  `lookup`'s is reachable on its own; it was not in the 54 J-9z42vnh
  certified.
- Cause 32, 1 site (resolutionCache.ts:652:9),
  `ResolutionCache.directoryWatchesOfFailedLookups`. `ReadonlyMap` there breaks
  the adapted tree's test build (J-vw1mey7, hereby build-tests):
  src/harness/incrementalUtils.ts(269,43) passes
  `actual.directoryWatchesOfFailedLookups` to
  `verifyDirectoryWatchesOfFailedLookups`, whose parameter is
  `Map<string, ts.DirectoryWatchesOfFailedLookup>`, so it is TS2345 there. The
  harness is outside the program stock tsc certified, and adapting it for one
  site is not worth it.
- Cause 24, 2 sites. `getEffectiveTypeRoots` is public, ruled out of this head.
- Cause 41, 1 site. Its line exists only in the adapted tree, not upstream.
- Cause 43, 1 site, `parseConfig`'s and `getExtendedConfig`'s
  `resolutionStack: string[]`, two declarations on the same line text, so
  neither was unique by line. By name both are reachable; they were not in
  the 54 J-9z42vnh certified.

## Writes elsewhere

Every write found to a slot this adaptation touches, and why none goes through
the view it makes readonly. Lines are upstream 6.0.3's.

- `semanticDiagnosticsPerFile` and `emitDiagnosticsPerFile` (cause 4). Map
  writes at builder.ts:405, 420, 496, 782, 894, 1068, 1858, 2225 and field
  assignments at 330, 497, 1886, 2168. Each goes through
  `BuilderProgramState`, which keeps its own `Map` override (213, 215), or
  through a local typed `BuilderProgramState["semanticDiagnosticsPerFile"]`.
  The Reusable view is only read (403, 418) or built fresh (2276-2277,
  2314-2315).
- `tracker` (cause 7). Assigned at checker.ts:6677, 8082, 8138 and 9392, each
  time through `NodeBuilderContext`, which redeclares
  `tracker: SymbolTrackerImpl` as mutable (54292). expressionToTypeNode.ts, the
  one module holding the `SyntacticTypeNodeBuilderContext` view, only calls
  `context.tracker.reportInferenceFallback`.
- `signature` (cause 15). Written at builder.ts:1939, 1943 and builderState.ts:395,
  458, through `BuilderState.FileInfo`. One alias path: builder.ts:2233 casts a
  build-info file info `as BuilderState.FileInfo`, and that object goes into
  `fileInfos` (2293), where those writes can reach it. tsc accepts the cast.
- `watcher` (cause 26). Assigned at watchPublic.ts:592 and resolutionCache.ts:1282
  through their concrete types, never through `closeFileWatcherOf`'s
  constraint.
- `ModuleResolutionState.host` (cause 44): no `state.host =` anywhere; states
  are built as literals.
- The cause 6 parameters, and the cause 16, 28, 42 and 45 parameters:
  no writes. The cause 6 bodies call map methods such as `delete` through
  `Readonly<>`, which is shallow and allows them.
- The cause 14, 17, 21, 37, 38, 46 and 47 arrays and collections: no
  writes. Cause 46 writes `resolution.isInvalidated` on elements, not on the
  collection.
- Cause 27's `cache` keeps a writable `Map`: its body calls `delete`. Both
  callers pass a `Map<Path, ...>` (resolutionCache.ts:604, 611).

## Public consequence

None. Each touched declaration is local to a function (the checker functions,
`createDiagnosticCollection`'s `getDiagnostics`, the resolution cache's two
functions, `commentWillEmitNewLine`, and `allDeclarations`), file-local (the builder.ts
functions and watch.ts's `createTabularErrorsDisplay`), or `@internal` on itself or its declaration
(`ReusableBuilderProgramState`, `SyntacticTypeNodeBuilderContext`,
`IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo`,
`ModuleResolutionState`, `TypeChecker.getDiagnostics`,
`createCachedDirectoryStructureHost`, `parseConfigHostFromCompilerHostLike`,
`not`, `closeFileWatcherOf` and `getReferencedFileLocation`). The script
refuses any slot that is none of these. Types and type arguments erase, so no
line of the emitted JavaScript and no line of typescript.d.ts moves. This
directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Proof

Sealed, at commit c6505c16 over main 375c551c, stock TypeScript 6.0.3.

Apply and stock tsc (J-pcnyjpe): apply builds the adapted tree with and without
this directory. A second run of `adapt.cjs` prints all 34 slots as already
adapted and changes no file. Exactly 12 files of `src/compiler` differ.
Stock tsc over tsc's whole program reports 0 diagnostics on both trees.
`stage3/lane` passes, and so does the adapted tree's own test build.

The census (J-6xnka7s and J-d0rz7gd without, J-4ngpspd and J-9vtn5dd with):
scan-all's instrumented probe over the closures of the 12 touched files, in two
halves per side, each half on one 64-core box at the same speed. A single run's
reach isn't fully repeatable, so each side is the union of its halves.

| Measurement | Without | With |
|---|---:|---:|
| Stock tsc diagnostics | 0 | 0 |
| Distinct stops over the closures | 2,540 | 2,483 |
| Sites that stop no more | | 56 |
| Sites that stop for a different reason | | 30 |
| Sites that stop and didn't before | | 0 |

The 30 that stop for a different reason are cause 4's nine in builder.ts
(`BuilderProgramState` seen as `ReusableBuilderProgramState`, now on its
`ReadonlyMap` fields), cause 7's eighteen in checker.ts
(`SyntacticTypeNodeBuilderContext` seen back as `NodeBuilderContext`, whose
readonly `tracker` it would make writable), and three host sites of cause 16's
parameters (program.ts:1584 and 5117, tsbuildPublic.ts:501). Each stopped
before. The scan stops at its usual `lowering.declareModule` boundary, so 56 is
a floor for the 69 listed sites; the daily meter's census is the full count.

## Mutants

Sealed, each taking one cause back out of `adapt.cjs` and scanning its file's
closure against the adaptation as it is (J-2gs5dhv, J-d1pazq4).

| Mutant | Brings back |
|---|---:|
| Road 2: the nine type arguments taken out of `createStateBuildOrder`'s prunes | 9, tsbuildPublic.ts:676 to 684, and nothing else |
| Cause 15: `readonly signature` taken out of `IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo` | 3, builder.ts:1337, 1344 and 1346, and nothing else |

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
