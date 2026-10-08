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

## Representation and lowering design

This section is a proposal for @system_adamic wherever it changes the language's accepted or refused programs. It is not authorization to implement those changes. Existing representation and iterable support may be completed without weakening any proof. Follow `docs/0.1.md`, `docs/escape-hatches.md`, and the no-collector ownership model in `docs/memory.md`.

| Shape | Representation and lowering | Admission boundary |
|---|---|---|
| string and string-literal unions | Existing string-key table, byte-content hashing and equality; retain keys | Already supported. Distinct allocated strings with equal content must match |
| number and numeric enum/literal unions | Existing double slots; SameValueZero hash and equality; canonical +0 at insertion | Already supported. All NaN payloads equal, infinities distinct, -0 shares +0's hash |
| boolean | Existing boolean-key flag, never read the slot as a double | Already supported |
| number or undefined | Existing packed MaybeNumber slots, reserved undefined tag distinct from every present NaN; optional-number hash/equality | Already supported. Insertion and iteration must preserve the packed tag |
| plain objects, tuples, classes and object-only unions | Existing identity-key table storing the original retained pointer; no structural hash, copy, coercion or field comparison | Already supported when represented as ir.Object and construction, mutable relations and cycle proofs hold |
| arrays, maps, sets and closures | Existing identity-key table; retain the original array/map/closure pointer | Already supported for homogeneous representations; closure identity is the closure value, not its function index |
| reference union mixing strings, arrays, objects, maps or closures | Proposed self-describing Union key, reference-kind discrimination; strings by content, other references by identity | NotYet until equality, hashing, boxing, iteration and ownership are validated together; never use box address as identity for a boxed primitive |
| scalar/reference unions, unknown and object | Proposed tagged Union key using the same runtime-kind representation as ordinary unions; numeric hash normalizes NaNs and zero; null and undefined use distinct tags | Proposal only. Invariant key slots and unknown-boundary checks remain required |
| boolean or undefined; nullish unions | Proposed explicit tags for absent values; no conflation of null with undefined or false | Proposal only; current single-null-pointer reference encoding cannot distinguish both nullish members |
| branded primitives (__String, Path and ResolvedConfigFilePath) | Proposed erasure of a proven primitive-only phantom brand to its primitive representation, with unchanged primitive equality | Proposal only. Construction assertions and brand member reads need a ruling; mere TypeScript assignability is not proof that a phantom property exists |
| iterable readonly pairs | Prove a fixed two-element tuple yield type; use the existing iterable plan to collect tuples, then existing MapNew/Pairs lowering | Compiler lesson under the existing iterator contract. Require exact slot representations or explicit checked conversions; preserve all existing iterator-origin checks |
| MapLike<T> and optional MapLike<T> | The record's shape plus dictionary from `docs/index-signatures.md`; one shared identity, string key order and checked field conversions | Separate approved index-signature design, with remaining implementation dependencies. Never translate a record view to a copied Map |

For proposed Union keys, add a dedicated key mode rather than treating ir.Union as an ordinary identity reference. Reuse the union's existing runtime tags and scalar boxes. Hash dispatch and equality dispatch must share one classification: string content; numeric SameValueZero; boolean value; null and undefined separately; other references by original identity. Canonicalize the numeric payload when retaining a new key. Equality must not invoke valueOf, toString, getters or user callbacks. Box allocation identity is never the identity of a numeric or boolean key.

The table owns one strong key edge and one strong value edge per live Map entry; Sets own only their elements. Overwrite releases the incoming duplicate key and the old value while preserving the first key's position. Delete/clear/free release each live edge once. Copying keys, entries, iterator results and constructor pairs retains the references each result owns. Borrowed loop payloads remain alive across delete, clear, callback mutation and exception paths. A key or element reaching its holder is still subject to the existing fresh-write and cycle analysis. No collector, weak-key fallback or leak-tolerating acceptance is proposed.

Live iteration continues through stable tombstones. Overwrite does not append; delete then reinsert does. A deletion before the next step is skipped; addition before exhaustion is visited; clear followed by add during iteration is visited. Exhaustion remains sticky. Retain the collection while its iterator exists; balance the active-iterator count on exhaustion and destruction. Never compact while any iterator needs its position. Snapshot-producing operations and live protocol iteration must stay distinct.

### Programs requiring a ruling, not implementation in this unit

These are reduced obligations, not trusted constructors:

```a
const mixed = new Map<string | number, string>();
mixed.set(1, 'number');
mixed.set('1', 'string');
console.log(mixed.size, mixed.get(1), mixed.get('1'));

const nullish = new Set<null | undefined>([null, undefined]);
console.log(nullish.size, nullish.has(null), nullish.has(undefined));

type Path = string & { readonly __pathBrand: unknown };
function store(path: Path): boolean {
    const paths = new Set<Path>();
    paths.add(path);
    return paths.has(path);
}
```

The first requires content/value equality across tags. The second requires two distinct nullish tags rather than one NULL. The third asks only for a representation, but does not establish how a Path can be constructed soundly or what reading __pathBrand means. Do not admit `'a' as Path` merely because its machine payload could be a string. The proposed primitive-brand rule must identify erased phantom members, reject observable unproven member contracts, and leave unchecked assertions refused.

The following must stay refused regardless of a key representation lesson:

```a
interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: () => string; }
const dogs = new Map<string, Dog>();
const animals: Map<string, Animal> = dogs;
animals.set('x', { name: 'cat' });
console.log(dogs.get('x')?.bark());

function widen(values: Set<Dog>): Set<Animal> { return values; }

const unrelated = 'wrong' as unknown as number;
const unsafe = new Map<number, number>();
unsafe.set(unrelated, 1);
```

Mutable key and value widening, unsafe callable variance, unrelated assertions, any, unproven runtime contracts and possible strong cycles remain refused with their existing reasons. Monomorphization failures, unresolved type parameters, iterator receiver-origin erasure, wrong tuple arity, optional/rest tuples and unsupported representations remain NotYet; never let a guessed slot layout reach clang.

### Silent-miscompile audit

1. Equal keys must hash alike: every NaN payload, both zero signs, distinct string allocations, absent packed numbers and reference identity. Different tags must not collapse 1, '1', true, null and undefined. Hash collisions must still compare the whole key.
2. A boxed numeric key cannot compare its box address; a reference key cannot compare its fields. Views and aliases must preserve the same original object, array, Map and closure identity. Separate closures from one function body must remain distinct.
3. Union fit/narrow conversions in literals, set/add/get/has/delete, constructor tuples, callback parameters, spread, keys/entries/values and iterator results must agree. Representation support in keyable alone is insufficient.
4. Present undefined values must not become absence. get's nullable result and has must distinguish them; iteration must retain them. Packed undefined must remain distinct from present NaN. null must not silently become undefined.
5. Evaluate receiver, key and value once, left to right. Callback and iterator protocol calls may mutate collections or throw. Constructor iteration must honor the existing done/value order and close semantics; tuple extraction must not re-evaluate the source.
6. Mutation must preserve insertion order, tombstones, sticky exhaustion and active iterator accounting, including nested iterators, clear, reinsert, early break and exceptional exits.
7. Ownership must survive overwrite with aliased key/value, runtime-built strings, deletion during callbacks, borrowed iterator payloads, copied entry tuples, growth and rehash. Sanitizers and leak checks need allocated inputs, not only immortal literals.
8. Cycle and invariant-slot proofs must see key edges as well as values. New lowering must record writes. Hidden iterator-to-collection edges cannot escape the cycle finder.
9. Maps and Sets share a native table but have different source contracts: add returns the Set, forEach receives value twice, and entries yields [value,value]. JavaScript lowering must use the actual collection kind.
10. MapLike integer-key enumeration, inherited names, __proto__, own presence, deletion, static/dynamic aliases, freeze and checked named-slot conversions must not be replaced by Map behavior. Keep this dependency explicit at the #whkxbc7 map-pair stop.

Acceptance requires original-source Node observations, both generated backends, native ASan/UBSan and leaks, recorded counts, and semantic mutants that emit valid code. A refusal mutant must be caught by a refused program becoming accepted, not by a C compile error. Historical hidden-byte credit is not a retirement claim: replay the exact original reason before and after the change, report any newly exposed blocker, and distinguish a diagnostic retired from a complete tsc root compiling.

## Reduced acceptance fixtures and baseline outcomes

The following source reductions retain the collection operation and key shape; Node edge cases expand the inputs rather than claim tsc actually inserts NaN or -0 at these sites. They are executable acceptance criteria, not a survey of runtime hot paths.

| Fixture | tsc reduction anchor | Base outcome |
|---|---|---|
| `internal/oracle/testdata/scout_map_objects.a` | core.ts:605 mapEntries, monomorphic compiler-node-shaped keys | Pass: distinct equal-field objects, alias lookup, overwrite position, delete/reinsert and Set identity |
| `internal/oracle/testdata/scout_map_numbers.a` | core.ts Map construction/update and Set deduplication shapes | Pass: NaN, -0 normalization, insertion order, delete/reinsert; edge inputs added explicitly |
| `internal/oracle/testdata/scout_map_references.a` | core.ts mapEntries instantiated for reference keys | Pass: arrays, Maps and distinct closures from one function body keep identity |
| `stage3/map-keys/branded_map.a` | Path-key Map uses in resolutionCache.ts and core.ts:605 | NotYet: `a value of type Path`, at the parameter before collection operations |
| `stage3/map-keys/branded_set.a` | builder.ts Set<Path> | NotYet: `a value of type Path`, at the parameter |
| `stage3/map-keys/mixed_map.a` | core.ts mapEntries instantiated with mixed K2 | NotYet: `a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions` |
| `stage3/map-keys/nullish_set.a` | Set deduplication, expanded to null/undefined | NotYet: `a Set of null \| undefined (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)` |
| `stage3/map-keys/maplike.a` | corePublic.ts:13, core.ts:1287 | Refused: `an index signature`; this implementation disagrees with the documented NotYet ruling, and is unchanged here |
| `stage3/map-keys/iterator_pairs.a` | commandLineParser.ts:141 readonly mapped pairs | NotYet: `new Map from something that isn't [key, value] pairs`; explicit iterator replaces the separate generator dependency |

`TestScoutMapKeyOutcomes` pins the six gap reductions, including diagnostic class. The source loader first proves they typecheck under Adamic's fixed options. The iterator reduction declares next, return and throw: omitting return or throw would trigger the existing optional-field compatibility refusal before reaching the pair-constructor gap. The original `core.ts:332` helper is a generator; `a generator function` and `yield (generators)` remain separate shared refusal dependencies, not key-table lessons. No generator frame or cancellation decision is implemented.

`source-node.json` records original-source Node exit/stdout/stderr for all six gap reductions. Brand parameter-only probes have no call and empty output: that preserves the unproven construction boundary rather than smuggling a cast into an acceptance test. The mixed and nullish programs print `2:number:string` and `2:true:true`; the record's missing read prints `-1`. The pair source prints four next calls, then two entries in first-insertion order, with the repeated key overwritten by its last value.

Scoped verification commands, after sourcing `/workspace/adamic-tools/env.sh`:

```sh
go test ./internal/lower -run '^TestScoutMapKeyOutcomes$' -count=1 -v > /tmp/scout-map-keys-outcomes.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_map_.*[.]a$' -count=1 -timeout 30m -v > /tmp/scout-map-keys-oracle.log 2>&1
go test ./internal/oracle -run '^TestScoutMapLookupMutants$' -count=1 -timeout 30m -v > /tmp/scout-map-keys-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout-map-keys-counts.log 2>&1
```

The gap run passed all six cases. The three new oracle programs passed source Node, generated JavaScript, sanitized native, release native and successful-program leak checks, with zero oracle cache hits. Three lost-lookup mutants replace emitted map_get calls with a valid helper returning absence. Each compiled under the normal warning policy, ran with exit 0 and empty stderr under ASan/UBSan and leak detection, and was killed only by source-Node stdout disagreement. A separate compiler mutant added ir.Union to keyable: the mixed_map outcome test failed with `got <nil>`, before code emission. The compiler source was restored. These mutants prove lookup comparison and that admission boundary; no claim is made that they independently prove every individual hash, ownership or order invariant.

The first counts refresh failed only because host fixtures required the missing pinned `@types/node` package. `npm ci --prefix stage3/api` installed the locked dependencies; the successful retry is recorded with the fixture evidence. No whole package test or full gate was run.
