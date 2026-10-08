# Step 19: Maps and Sets with any key type

## Census inventory

Base: `8cb5e7c189fc431ab3cab7e0be2c0530e722a598`, delivery branch `codex/scout-map-keys`. The requested area tip overrides the usual main start. This is the map-keys-compiler inventory, including the MapLike placeholder and #whkxbc7 map-pair stop.

Observation sources: the committed entry-root full census under `stage3/meter/runs/20261008T035244Z.latent-full/tsc/full.jsonl.gz`; hidden ranking from `codex/stage3-hidden-ranking` at `6c4fc1af`; its pinned refusal table. The full census is measured on a checker-rejected program, not evidence that tsc compiles. The base includes the replay implementation; it does not contain a fresh full census produced at 8cb5e7c1. These inherited measurements are historical, not a same-base rerun.

Counts deliberately keep three units apart. Entry roots count distinct `(unit, where, kind, reason)` observations in the entry-root census. Ranking roots are its distinct boundary spans, not TypeScript project entry points. Hidden bytes are the frozen outermost-cause attribution, not measured bytes successfully compiled after a fix. A dash means the exact reason is absent from that ranking, not zero. Counts must not be added across the two censuses.

All collection-named Refused and NotYet reasons are retained, even value-only and generic blockers that this step cannot retire. `an index signature` is included for MapLike. A MapLike is a JavaScript record, not a Map: integer-key order, prototype behavior and aliasing remain the index-signature contract. Branded-string and general `T` reasons outside collection sites are shared representation dependencies, not claimed collection-only roots.

Witnesses are actual diagnostic file:line locations, supplemented from the entry-root census. When a reason has fewer than three distinct sites, the available sites are shown and the shortage is explicit. Fabricating extra witnesses would falsify a singleton measurement.

| Kind and exact reason | Entry roots | Ranking roots | Hidden bytes | Three witnesses when available |
|---|---:|---:|---:|---|
| NotYet: a Set of __String (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 14 | 19 | 2739 | binder.ts:576<br>binder.ts:3615<br>checker.ts:5232 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 137 | 195 | 2391 | binder.ts:2413<br>binder.ts:3346<br>binder.ts:3631 |
| NotYet: a Set of Path (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 6 | 12 | 931 | builder.ts:647<br>builder.ts:667<br>builder.ts:696 |
| NotYet: a value of type MapLike<T> | 3 | 3 | 725 | core.ts:1287 <br>only 1 distinct sites available |
| NotYet: a value of type { forEach: (callbackfn: (value: T, key: K, map: Map<K, T>) => void, thisArg?: any) => void; clear: () => void; } | 4 | 9 | 598 | utilities.ts:8172 <br>only 1 distinct sites available |
| NotYet: a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 1 | 1 | 309 | tsbuildPublic.ts:666 <br>only 1 distinct sites available |
| NotYet: a value of type MapLike<T> \| undefined | 0 | 1 | 234 | core.ts:1370 <br>only 1 distinct sites available |
| NotYet: new Map from something that isn't [key, value] pairs | 1 | 1 | 188 | commandLineParser.ts:141 <br>only 1 distinct sites available |
| NotYet: a value of type Set<K> | 1 | 1 | 176 | utilities.ts:8316 <br>only 1 distinct sites available |
| NotYet: a Map of ResolvedConfigFilePath | 1 | 1 | 112 | tsbuildPublic.ts:663 <br>only 1 distinct sites available |
| NotYet: a Map of CompilerOptionsValue | 1 | 1 | 61 | commandLineParser.ts:2785 <br>only 1 distinct sites available |
| NotYet: a Map of T | 1 | 1 | 39 | core.ts:1908 <br>only 1 distinct sites available |
| NotYet: a Map of HostFileInfo | 2 | 2 | 0 | watchPublic.ts:751<br>watchPublic.ts:823 <br>only 2 distinct sites available |
| NotYet: a Map of RedirectsCacheKey | 1 | 1 | 0 | moduleNameResolver.ts:1034 <br>only 1 distinct sites available |
| NotYet: a Map of VisitResult<ExportAssignment \| LateVisibilityPaintedStatement \| undefined> | 2 | 2 | 0 | transformers/declarations.ts:987<br>transformers/declarations.ts:1347 <br>only 2 distinct sites available |
| NotYet: a Map of false \| MutableFileSystemEntries | 3 | 2 | 0 | watchUtilities.ts:233<br>watchUtilities.ts:377<br>watchUtilities.ts:123 |
| NotYet: a Map of string \| number | 1 | - | - | commandLineParser.ts:3869 <br>only 1 distinct sites available |
| NotYet: a Map whose key and value types aren't known | 1 | 1 | 0 | resolutionCache.ts:1478 <br>only 1 distinct sites available |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 4 | - | - | commandLineParser.ts:2599<br>commandLineParser.ts:2743<br>executeCommandLine.ts:364 |
| NotYet: a field of type "boolean" \| "list" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 1 | - | - | executeCommandLine.ts:375 <br>only 1 distinct sites available |
| NotYet: a field of type "boolean" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 1 | - | - | commandLineParser.ts:1896 <br>only 1 distinct sites available |
| NotYet: a function returning Map<K, V1 \| V2> | 2 | - | - | core.ts:1409 <br>only 1 distinct sites available |
| NotYet: a value of type Map<K, T> | 3 | - | - | utilities.ts:8251<br>utilities.ts:8211 <br>only 2 distinct sites available |
| NotYet: a value of type Map<Path, ModeAwareCache<T>> \| undefined | 2 | 2 | 0 | program.ts:2007 <br>only 1 distinct sites available |
| NotYet: a value of type Map<string, SingleFileWatcher<T>> | 0 | 2 | 0 | sys.ts:500 <br>only 1 distinct sites available |
| NotYet: a value of type Map<string, WildcardDirectoryWatcher<T>> | 1 | 1 | 0 | watchUtilities.ts:515 <br>only 1 distinct sites available |
| NotYet: a value of type Map<string, [K, V[]]> | 1 | 1 | 0 | checker.ts:44581 <br>only 1 distinct sites available |
| Refused: a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 | - | - | builder.ts:1686<br>builder.ts:1694<br>builder.ts:1724 |
| Refused: a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> \| undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 | - | - | builder.ts:1886 <br>only 1 distinct sites available |
| Refused: a value of type Map<Path, DirectoryWatchesOfFailedLookup> seen as Map<string, DirectoryWatchesOfFailedLookup>, which can write string where Path is read | 1 | - | - | resolutionCache.ts:650 <br>only 1 distinct sites available |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 | - | - | tsbuildPublic.ts:2204 <br>only 1 distinct sites available |
| Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read | 1 | - | - | resolutionCache.ts:1471 <br>only 1 distinct sites available |
| Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read | 1 | - | - | resolutionCache.ts:1472 <br>only 1 distinct sites available |
| Refused: a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 | - | - | sys.ts:788 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 | - | - | tsbuildPublic.ts:680 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 | - | - | tsbuildPublic.ts:674 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 | - | - | tsbuildPublic.ts:681 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 | - | - | tsbuildPublic.ts:678 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 | - | - | tsbuildPublic.ts:682 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | - | - | tsbuildPublic.ts:676 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 | - | - | tsbuildPublic.ts:675 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 | - | - | tsbuildPublic.ts:677 <br>only 1 distinct sites available |
| Refused: a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 | - | - | tsbuildPublic.ts:679 <br>only 1 distinct sites available |
| Refused: a value of type Map<__String, TransientSymbol> seen as SymbolTable \| undefined, which can write Symbol where TransientSymbol is read | 0 | - | - |  <br>only 0 distinct sites available |
| Refused: a value of type Map<string, CachedResolvedModuleWithFailedLookupLocations> seen as Map<string, ResolutionWithFailedLookupLocations> \| Set<ResolutionWithFailedLookupLocations> \| undefined, which can write ResolutionWithFailedLookupLocations where CachedResolvedModuleWithFailedLookupLocations is read | 1 | - | - | resolutionCache.ts:1574 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 | - | - | commandLineParser.ts:838 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 | - | - | commandLineParser.ts:741 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 | - | - | utilities.ts:1385 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 | - | - | commandLineParser.ts:1672 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 | - | - | commandLineParser.ts:1098 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 | - | - | commandLineParser.ts:1427 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 | - | - | commandLineParser.ts:317 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 | - | - | commandLineParser.ts:305 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 | - | - | commandLineParser.ts:291 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause \| VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause \| VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | 1 | - | - | checker.ts:44643 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 | - | - | utilities.ts:1415<br>utilities.ts:1418<br>utilities.ts:1464 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 | - | - | commandLineParser.ts:709 <br>only 1 distinct sites available |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 | - | - | utilities.ts:1713<br>utilities.ts:1912<br>utilities.ts:1926 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | - | - | builderState.ts:312 <br>only 1 distinct sites available |
| Refused: a value of type Set<Path> \| undefined seen as Set<string> \| undefined, which can write string where Path is read | 1 | - | - | resolutionCache.ts:1573 <br>only 1 distinct sites available |
| Refused: a value of type Set<__String> seen as Set<__String \| undefined>, which can write __String \| undefined where __String is read | 1 | - | - | checker.ts:14225 <br>only 1 distinct sites available |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 3 | - | - | utilities.ts:9402<br>commandLineParser.ts:2933 <br>only 2 distinct sites available |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 | - | - | utilities.ts:9401 <br>only 1 distinct sites available |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 | - | - | commandLineParser.ts:3854 <br>only 1 distinct sites available |
| Refused: an index signature | 11 | - | - | corePublic.ts:14<br>tracing.ts:55<br>types.ts:7582 |

## Refusal-table ownership and disposition

The supplied refusal table classifies writable collection widening as adaptation. Preserve that refusal: a larger key domain permits writes that violate aliases of the smaller domain. Representation support cannot authorize variance. Exact rows supplied by that table follow.

| Exact reason | Owner | Refusal sites | Witnesses supplied by table |
|---|---|---:|---|
| a value of type Map<__String, TransientSymbol> seen as SymbolTable \| undefined, which can write Symbol where TransientSymbol is read | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | 1 | checker.ts:13997<br>none (one file:line site) |
| a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause \| VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause \| VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | 1 | checker.ts:44643<br>none (one file:line site) |
| a value of type Set<__String> seen as Set<__String \| undefined>, which can write __String \| undefined where __String is read | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | 1 | checker.ts:14225<br>none (one file:line site) |

## Priorities and scope

The frozen key-representation reason has 195 ranking roots and 2,391 hidden bytes; branded Sets add __String (19, 2,739), Path (12, 931), and ResolvedConfigFilePath (1, 309). These are not necessarily disjoint fixes: each actual key type and its construction proof must be audited. Object, array, Map and closure identity already have representation paths on the base. Mixed representations use ir.Union, which keyable currently excludes.

The map-pair stop is `commandLineParser.ts:141:65`, `new Map(mapIterator(jsxOptionMap.entries(), ...))`: one ranking root, 188 hidden bytes. Generic MapLike<T> has three ranking roots and 725 bytes; its optional form has one and 234. Generic structural collection signatures, Map value representations and unsafe widening remain dependencies rather than promises to retire them with key equality.
