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
| iterable readonly pairs | Prove a fixed two-element tuple yield type; use the existing iterable plan and insert each pair before asking for the next one | Compiler lesson under the existing iterator contract. Require exact slot representations or explicit checked conversions; preserve all existing iterator-origin checks |
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


## Implemented iterator-pair lesson

Concrete custom iterables yielding exactly two required tuple elements now lower into Map construction. The source expression is evaluated once. A generated function allocates the Map, starts the proven iterator, inserts the current key and value immediately, and reuses the existing close/exception machinery. Key and value slot representations must match exactly; conversion remains NotYet. Writes are registered with the existing cycle proof. No runtime, backend or acceptance-policy change was made.

Streaming matters: a readonly tuple can alias a mutable tuple reused by next(). Buffering tuples would observe later mutations instead of the original pairs. `reused_pair_mutation.a` records that source-Node case, but tuple-element assignment remains NotYet, so this alias hazard has no native acceptance claim yet. `iterator_pairs.a` now also checks one factory evaluation, duplicate overwrite order, allocated strings, and a throwing next(). Node does not call return after next itself throws; both backends agree and native cleanup is leak-free. `final-source-node.json` records the final seven reductions; `source-node.json` preserves the earlier baseline.

The original stop did not retire. Replaying commandLineParser.ts:141:65 still produces exactly `new Map from something that isn't [key, value] pairs`. All 81 adapted source hashes match the inherited census manifest. The replay is explicitly measured on a checker-rejected entry-root program. Evidence is in `stage3/map-keys/evidence/pair-replay.json` and the compressed full log. Retired original tsc roots: **0**. Closed reduced pair-constructor gaps: **1**. The original mapIterator/generator dependency remains, and no historical hidden bytes are credited as recovered. A fresh complete census on this compiler was not run; the inventory above is historical evidence, not a new base-wide measurement.

Final scoped commands (all test output redirected to logs):

```sh
go test ./internal/lower -run '^TestScoutMap(KeyOutcomes|IteratorPairsLower|IteratorPairRepresentationGap)$' -count=1 -v > /tmp/scout-map-keys-final-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^stage3$/^map-keys$/^iterator_pairs[.]a$' -count=1 -timeout 30m -v > /tmp/scout-map-keys-pairs-final-oracle.log 2>&1
go test ./internal/oracle -run '^TestScoutMapPairInsertionMutant$' -count=1 -v > /tmp/scout-map-keys-pair-mutant.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(user_iterators[.]a|library_map_set_construct[.]a)$' -count=1 -timeout 30m -v > /tmp/scout-map-keys-iterator-regression.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout-map-keys-final-counts.log 2>&1
gofmt -l internal/lower/collections.go internal/lower/map_iterator_pairs.go internal/lower/map_keys_scout_test.go internal/oracle/map_keys_scout_test.go > /tmp/scout-map-keys-format.log 2>&1
go vet ./internal/lower ./internal/oracle > /tmp/scout-map-keys-vet.log 2>&1
bash stage3/apply.sh /tmp/scout-map-keys-adapted > /tmp/scout-map-keys-adapt.log 2>&1
go run ./stage3/census/latent/replay -project /tmp/scout-map-keys-adapted/src/tsc/tsc.ts -where /tmp/scout-map-keys-adapted/src/compiler/commandLineParser.ts:141:65 -kind NotYet -reason "new Map from something that isn't [key, value] pairs" > /tmp/scout-map-keys-pair-replay.log 2>&1
```

All focused tests passed; oracle runs had zero cache hits. Formatting and vet produced no findings. The pair fixture passes original Node, generated JavaScript, sanitized native, release native and leaks. Its counts are 61 allocations, 61 frees, 48 retains, 78 releases, peak 19, zero regions. Counts refresh also canonicalized the existing logical_and_reference_maybe row's position and removed the stale taste/17_binder_flow row absent from this base's fixture registry; those are registry refresh effects, not compiler retirements.

The insertion mutant drops the second store while releasing its owned strings. It compiles and exits cleanly with no sanitizer/leak finding; Node stdout kills it. A compiler guard mutant changes the representation mismatch condition from OR to AND; `TestScoutMapIteratorPairRepresentationGap` fails because the widened value program becomes accepted. Restored compiler source passes. Together with the three lookup mutants, the union-admission mutant and the inventory-count mutant recorded above, these are the executed mutants; delayed tuple aliasing, all hash internals and every ownership invariant are not independently mutation-tested.

Setup used `GOPROXY=https://proxy.golang.org|direct`, `bash cloud/setup.sh`, then the printed `/workspace/adamic-tools/env.sh`. Observed cumulative timing lines: node 0.064s, Go 0.079s, clang 0.483s, markdown 1.941s, ready 2.135s, submodule 24.295s, build 198.426s, deferred 198.544s, warm 198.546s, done 198.580s. `nproc` printed 5. Setup succeeded. The pinned Node types installation described above resolved the only counts prerequisite failure.

Mixed union keys, nullish tags, primitive brands and MapLike remain proposals or existing gaps. MapLike's index-signature diagnostic classification discrepancy remains explicit. No generator implementation, unchecked cast, widening, GC, full package test or full gate is included. This branch completes a conservative compiler lesson toward roadmap step 19; it does not complete step 19.


## Step 19 ruling and resumed delivery

@system_adamic's #pvvzhmy ruling supersedes the proposal-only status for these six decisions: MultiMap expandos adapt to composition with explicit receiver helpers under #pcwk8fh; custom-equality createSet retains its own protocol; Program-region collections hold keys and values strongly with plain region pointers; Path and __String brands erase, the branded void arm is never, and unproven input is checked as string; same-map/same-key has/get without an intervening write proves presence, retaining stored undefined when V permits it; indexed loops prove reads from the range condition and stable length, otherwise reads are checked loudly. No weak fallback or substitution of built-in Set semantics is authorized.

Merged the authorized area-next line at dcdbb909 into the delivery branch, producing 04b105cf. The merge preserves the four previous units. The first ruled piece addresses string brands, which account for most concrete keys behind the 195-root general key reason and the branded Set reasons. No unrelated area branch was merged.

### Piece 1: Path and __String collection keys

Path and __String with their void markers are represented as strings. The __String branded void arm is uninhabited and omitted from representation selection. This implementation recognizes the two ruled marker names and void payloads; an unknown marker payload, other structural field contract, or branded literal refinement does not acquire permission. A string-to-brand cast erases only when the brand admits every string. An unproven input receives a single-evaluation typeof-string check and a panic before the value enters a string slot. Other unsafe cast rules stay in force. This is a local implementation of the ruled representation, without importing an unlanded area-views implementation.

`scout_map_brands.a` is reduced from types.ts:23,6201, binder.ts:576 and builder.ts:1429. It checks allocated equal-content keys, overwrite order, Map/Set iteration, an empty key and a successful unknown boundary. Both backends agree with source Node under sanitizers, release and leaks. The failing boundary fixture prints `boundary:1` then both generated backends panic with `brand boundary failed: expected string for __String`, exit 70. Source Node instead accepts the assertion and prints `number`; the checked-fixture contract records this intentional difference. Lowering controls cover number, undefined, boolean and null boundaries and a refused branded literal refinement. The earlier unknown-marker baseline reductions stay NotYet because their unknown payload is not the ruled void brand.

The semantic mutant removes the string check from the real generated helper. Both mutant backends exit 0, and native has no sanitizer or leak finding. The exact expected panic catches the omission. No compile-warning kill is counted.

A guarded batch replay rechecks 159 exact historical collection signatures before and after, using the same 81 source hashes. On the merged baseline 137 reproduce. After this piece, 103 of those diagnostic signatures no longer reproduce: 83 general Map-key sites, 14 Set<__String> sites and 6 Set<Path> sites. Eighteen distinct selected units have no remaining lowering blockers in the replay. The remaining signatures and all newly exposed blockers are recorded, rather than credited as compilation success. The measurement is on a checker-rejected entry-root program. Whole tsc entry roots compiling: zero. No hidden-byte recovery is measured. See `stage3/map-keys/evidence/ruling-brands/retirement.json`, before/after compressed records and the reproducible `stage3/map-keys/replay.py`.

Exact scoped validation, with the toolchain environment sourced and all output redirected:

```sh
python3 stage3/map-keys/replay.py /tmp/scout-map-keys-adapted /tmp/scout19-brands-before.jsonl > /tmp/scout19-brands-before.log 2>&1
go test ./internal/lower -run '^TestScoutMap(BrandBoundary|BrandRefinementRefused|KeyOutcomes)$' -count=1 -v > /tmp/scout19-brands-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_map_brand.*[.]a$' -count=1 -timeout 30m -v > /tmp/scout19-brands-oracle.log 2>&1
go test ./internal/oracle -run '^TestScoutMapBrandBoundaryMutant$' -count=1 -timeout 30m -v > /tmp/scout19-brands-mutant.log 2>&1
python3 stage3/map-keys/replay.py /tmp/scout-map-keys-adapted /tmp/scout19-brands-after.jsonl > /tmp/scout19-brands-after.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout19-brands-counts.log 2>&1
go test ./internal/lower -run '^(TestScoutMap.*|TestCast.*)$' -count=1 -v > /tmp/scout19-brands-regression-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(scout_map_.*[.]a|casts[.]a|cast_fails[.]a|unknown[.]a|library_map_set_keys[.]a)$' -count=1 -timeout 30m -v > /tmp/scout19-brands-regression-oracle.log 2>&1
go vet ./internal/lower ./internal/oracle > /tmp/scout19-brands-vet.log 2>&1
```

All passed. Successful brand fixture counts: 34 allocations, 34 frees, 48 retains, 71 releases, peak 11, zero regions. The failing fixture stops at panic and is not a successful-program leak claim. Counts refresh changes only the two added rows. Setup succeeded with node 0.021s, Go 0.022s, submodules 0.065s, markdown 0.074s, clang 0.178s, build 37.967s, deferred 38.089s, warm 38.090s, done 38.119s, cumulative as printed; nproc 5. Full logs are under `stage3/map-keys/evidence/ruling-brands`.

### Piece 2: collection presence and checked reads

The ruling now has a loader obligation and a lowering proof. The loader discharges only TS2322/TS2345 relations whose sole missing fact is an ordinary library Map/ReadonlyMap lookup's undefined arm. The present payload must remain assignable to the receiving type. Direct calls and const-local snapshots are recognized; wrong payloads, nullable payloads, custom methods named get and mutable local snapshots retain checker errors. An unsupported presence representation fails loudly in lowering.

Lowering proves the same receiver and key bindings (or the same primitive literal) under a positive has branch or after a missing-key guard that immediately returns or throws. Intervening assignments, updates, calls, constructors, property/index reads and other effectful operations invalidate the proof conservatively, including writes through aliases. A value type admitting undefined never loses its stored undefined. An unproven required read panics rather than supplying a scalar default. Existing checked TypeScript assertions benefit from the proof; Adamic's assertion refusal remains unchanged. The initial regression caught an unnecessarily broadened assertion exception, which was removed before delivery.

The two new fixtures reduce core.ts createSet's has/get pattern with monomorphic payloads. The successful fixture agrees with source Node for zero, NaN, absent keys, allocated string values, stored undefined and alias deletion with an explicit fallback. The invalidated required read prints reading and then both generated backends panic with collection lookup failed: map.get('key') is undefined, exit 70. Source Node instead prints undefined; this intentional checked contract is asserted separately. Successful native execution passes sanitizers and leak accounting: 19 allocations, 19 frees, 12 retains, 29 releases, peak 9, zero regions. The panic fixture is not a successful-program leak claim.

The production mutant makes collectionPrefixUnchanged always true. The flow tests fail for alias writes, unknown calls, key updates and an early-return guard followed by an alias write. The runtime contract fails because the mutant exits 0 and prints reading followed by 0 instead of panicking. The native mutant runs without a sanitizer or leak finding. Both mutant commands exit 1; the source is restored afterward.

Scoped commands, with environment sourced and output redirected to the named logs:

```sh
go test ./internal/load -run '^TestCollectionRead' -count=1 -v > /tmp/scout19-presence-load.log 2>&1
go test ./internal/lower -run '^(TestScoutMap.*|TestNonNull.*|TestWhatZeroOneRefusesIsRefusedWithAFix|TestLibraryMapSet.*)$' -count=1 -v > /tmp/scout19-presence-regression-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(scout_map_.*[.]a|non_null.*[.]a|library_map_set_keys[.]a)$' -count=1 -timeout 30m -v > /tmp/scout19-presence-regression-oracle.log 2>&1
go test ./internal/oracle -run '^TestScoutMapPresenceInvalidationContract$' -count=1 > /tmp/scout19-presence-contract.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout19-presence-counts.log 2>&1
go vet ./internal/load ./internal/lower ./internal/oracle > /tmp/scout19-presence-vet.log 2>&1
# With collectionPrefixUnchanged mutated, each must fail:
go test ./internal/lower -run '^TestScoutMapPresenceRules$' -count=1 > /tmp/scout19-presence-mutant-flow.log 2>&1
go test ./internal/oracle -run '^TestScoutMapPresenceInvalidationContract$' -count=1 > /tmp/scout19-presence-mutant-contract.log 2>&1
```

All unmutated checks pass. The uncached oracle run records 22 native misses and 16 Node misses, no hits. Evidence logs are in stage3/map-keys/evidence/ruling-presence. Additional frozen census roots credited: zero. This piece implements a ruled flow fact; it does not remove createSet's generic-key, union-value or generator blockers, and no entry-root or hidden-byte credit is inferred from a reduced fixture. Indexed arrays and collection protocol implementations remain separate pieces.

### Piece 3: indexed arrayToMap reads

An ordinary array index can carry the same checked absence obligation as Map.get. The receiving payload must remain assignable after excluding undefined; nullable payloads remain unsupported. The range proof follows a direct index or a const-local snapshot to a fresh let index initialized at zero, incremented with i++, and compared as i < array.length on the same array binding. It rejects effects between the comparison and the read, writes to the index anywhere in the body, captures of the index, hoisted var indices, different arrays, negative starts and elements that can themselves be undefined. No assertion is needed in the ruled source reduction.

Length changes after a read do not invalidate that completed read: the next iteration compares the current length again. The source-reduced arrayToMap fixture explicitly exercises a callback that shrinks the array after each read. It also covers empty input, zero, duplicate-key overwrite order and skipped undefined keys. Both backends agree with source Node, under sanitizers and successful-program leaks: 35 allocations, 35 frees, 25 retains, 57 releases, peak 11, zero regions. The invalidated fixture pops the only element before reading it. Source Node prints undefined; both generated backends print reading and panic with collection lookup failed: value is undefined, exit 70. Its terminal panic is not a successful-program leak claim.

The mutant omits the comparison-to-read effect check in collectionRangeProven. The alias-mutation flow control fails, and the independent runtime contract fails because the generated program prints reading then 0 and exits 0 instead of panicking. The native mutant is sanitizer/leak clean; the failure is semantic, not a build warning. The production source is restored. Frozen census roots credited: zero, since the full generic overload's key/value representation blockers remain. This proves the concrete loop rule without claiming whole tsc compilation or hidden-byte recovery.

Exact commands, after sourcing the toolchain and redirecting output:

```sh
go test ./internal/load -run '^TestCollectionRead' -count=1 -v > /tmp/scout19-range-load.log 2>&1
go test ./internal/lower -run '^TestScoutArrayRangeRules$' -count=1 -v > /tmp/scout19-range-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_array_to_map.*[.]a$' -count=1 -timeout 30m -v > /tmp/scout19-range-oracle.log 2>&1
go test ./internal/oracle -run '^TestScoutArrayRangeInvalidationContract$' -count=1 -v > /tmp/scout19-range-contract.log 2>&1
go test ./internal/lower -run '^(TestScoutArrayRangeRules|TestScoutMap.*|TestNonNull.*|TestWhatZeroOneRefusesIsRefusedWithAFix)$' -count=1 -v > /tmp/scout19-range-regression-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(scout_(map|array)_.*[.]a|array_reads.*[.]a|library_array.*[.]a)$' -count=1 -timeout 30m -v > /tmp/scout19-range-regression-oracle.log 2>&1
go vet ./internal/load ./internal/lower ./internal/oracle > /tmp/scout19-range-vet.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout19-range-counts.log 2>&1
# With the comparison-to-read effect check omitted, each must fail:
go test ./internal/lower -run '^TestScoutArrayRangeRules$' -count=1 > /tmp/scout19-range-mutant-flow.log 2>&1
go test ./internal/oracle -run '^TestScoutArrayRangeInvalidationContract$' -count=1 > /tmp/scout19-range-mutant-contract.log 2>&1
```

Logs are preserved as .log.txt evidence under stage3/map-keys/evidence/ruling-range; the preceding presence logs are also preserved with that extension.

All unmutated range checks passed. The broad affected-fixture run records 57 native misses and 40 Node misses, zero hits. No whole package test or full gate was run.

### Piece 4: MultiMap composition and concrete generic reads

Adaptation 76, tracked under the ruled #pcwk8fh adaptation list, replaces tsc's expando factory with a class owning a Map and explicit receiver helpers. It retains the MultiMap interface and factory signature. The delegated Map API preserves callback receiver identity, bound forEach callbacks and live iteration. The applied tree's extracted declarations pass strict TypeScript checking and emit the same declarations as the original. Composition changes native-brand reflection such as instanceof Map and own-field enumeration; that is a representation consequence of the sanctioned adaptation, not a claim that the class is an intrinsic Map. The adapter checks the exact reviewed source hash and the exact applied replacement. Its measured patch row is one file, 27 added lines and 11 removed lines. A repeat makes no writes; a changed-source mutant exits 1.

The native acceptance fixture uses the reduced generic composition library in stage3/map-keys/multimap.a. Its add/remove helpers preserve the returned bucket alias, first-occurrence strict equality, unordered removal, deletion of empty buckets, allocated string keys, object key identity, insertion order and stored undefined. It is held to Node in both backends with sanitizers and successful-program leak counts: 33 allocations, 33 frees, 59 retains, 90 releases, peak 16, zero regions. Independent Node verification extracts tsc's original expando implementation and compares both this reduction and the actual adapted class, including its complete delegated protocol. The source and native empty-bucket mutants are caught by changed size output; the native mutant exits 0 with no sanitizer or leak finding.

Concrete generic receiving types are now examined before imposing a checked lookup obligation. A receiving type instantiated with undefined keeps that value. Presence/range proofs also inspect instantiated element types. Removing the generic receiving-type guard makes both generated backends panic on source.get('key') while source Node returns stored undefined, and the oracle fails with exit 1.

The complete adapted class is still native NotYet: a method with a computed name at its Symbol.toStringTag getter. Its probe uncovered an accessor-name prepass panic. The prepass now skips unsupported computed names and lets the existing precise gap report; reverting that guard reproduces the panic and fails the test. The original expando shape remains loudly rejected. Neither the complete class's native execution nor whole tsc is claimed. Frozen census roots retired here: zero; no hidden-byte credit.

Exact scoped commands, with environment sourced and output redirected:

```sh
python3 stage3/apply.py /tmp/scout19-composed-tsc --write-table > /tmp/scout19-multimap-apply.log 2>&1
node stage3/adapt/76-multimap-composition/adapt.cjs /tmp/scout19-composed-tsc > /tmp/scout19-multimap-idempotence.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/map-keys/verify_multimap.cjs /tmp/scout-map-keys-adapted /tmp/scout19-composed-tsc /tmp/scout19-multimap-node.json > /tmp/scout19-multimap-node.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_multimap_composition[.]a$' -count=1 -timeout 30m -v > /tmp/scout19-multimap-oracle.log 2>&1
go test ./internal/oracle -run '^TestScoutMultiMapEmptyBucketMutant$' -count=1 -v > /tmp/scout19-multimap-mutant.log 2>&1
go test ./internal/lower -run '^(TestScout.*|TestClass.*Accessor.*|TestLiteral.*Accessor.*|TestAccessor.*)$' -count=1 -v > /tmp/scout19-multimap-lower.log 2>&1
go test ./internal/lower -run '^TestScoutMultiMap' -count=1 -v > /tmp/scout19-multimap-boundaries.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_(map|array|multimap)_.*[.]a$' -count=1 -timeout 30m -v > /tmp/scout19-multimap-regression.log 2>&1
go vet ./internal/load ./internal/lower ./internal/oracle > /tmp/scout19-multimap-vet.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout19-multimap-counts.log 2>&1
```

All unmutated checks pass, with no full package test or full gate. The uncached regression records 27 native misses, 20 Node misses and no hits. Compiler and adapter mutant logs are preserved under ruling-multimap alongside Node output and source hashes. The new stage 3 adaptation has Node/protocol/type evidence; no complete upstream test suite was run.

### Piece 5: collection-owned keys and values

The new lifetime fixture makes compiler-node-shaped keys and payloads inside a producer, returns only the Map or Set and then reads the held objects. Both backends agree with source Node. Sanitizers and leak accounting pass: 14 allocations, 14 frees, 12 retains, 21 releases, peak 9, zero regions. No weak edges are introduced. The two native ownership mutants drop the Map's owned key edge or owned value edge after insertion. Each builds and is caught by ASAN heap-use-after-free when the producer's locals have been released. These are lifetime failures, not compiler-warning kills.

Observed on this base: collection construction uses the counted heap; native/region.go plans statement regions for fresh objects, not a Program lifetime for Map/Set storage. These fixtures report zero regions. The ruling's plain pointers inside step 06's Program region cannot be claimed here. Its implementation is a dependency on the Program-region owner. Inferring region-safe cycles from counted-heap fixtures would be unsound, so the existing cycle refusals are preserved. Frozen census roots credited: zero.

Exact scoped commands, after sourcing the environment and redirecting output:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_map_strong_edges[.]a$' -count=1 -timeout 30m -v > /tmp/scout19-strong-oracle.log 2>&1
go test ./internal/oracle -run '^TestScoutCollectionStrongEdgeMutants$' -count=1 -v > /tmp/scout19-strong-mutants.log 2>&1
go vet ./internal/lower ./internal/oracle > /tmp/scout19-strong-vet.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout19-strong-counts.log 2>&1
```

All pass. The lifetime oracle records three native misses and two Node misses, zero hits. Evidence is preserved under ruling-strong. This piece holds the existing strong ownership contract; it does not implement Program-region allocation.

### Piece 6: original custom-equality Set acceptance boundary

custom_set.a retains the original createSet declaration/body from core.ts:1622-1743. The Node verifier compares that declaration against the pinned adapted source before executing it. Only its supporting helpers and driver are reduced. A local Set protocol describes the operations this collection actually supplies; the newer checker library's built-in Set additionally demands union, intersection, difference and other operations absent from tsc's original object. This fixture-local protocol keeps the original body and its own collection identity; it is not a global adaptation or an intrinsic Set replacement.

Source Node observes custom equality despite hash collisions, first representative retention, bucket iteration, callback receiver identity, equality-based deletion, paired entries, the original Map string tag and clear. Replacing the factory call with built-in Set changes size from 2 to 3, makes membership on a fresh equal-span object false, retains the duplicate representative and changes the string tag. The comparator catches this mutant. This is source behavior evidence; no generated-backend execution or sanitizer/leak success is claimed for the refused collection.

Native observation: the fixture loads, then lowering refuses its generator function with the existing suspended-frame ownership and cancellation diagnostic. The explicit outcome test and source Node contract pass. The generic hash/union bucket representation and the full protocol have not been implemented. The step 06 Program-region lifetime is also absent from this base. Inference: the original forEach method captures the returned collection through a closure; accepting this structure on the counted heap would require an ownership solution, rather than weakening its edges or replacing its protocol. The ruling specifies the desired semantics but does not supply those missing implementations. This part is blocked until Program-region and suspended-iterator ownership support are available; it remains a loud refusal.

Exact scoped commands, after sourcing the environment and redirecting output:

```sh
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/map-keys/verify_custom_set.cjs /tmp/scout-map-keys-adapted /tmp/scout19-custom-set-node.json > /tmp/scout19-custom-set-node.log 2>&1
go test ./internal/lower -run '^TestScoutOriginalCustomSetOutcome$' -count=1 -v > /tmp/scout19-custom-set-outcome.log 2>&1
go test ./internal/lower -run '^(TestScout.*|TestGeneratorsAreRefusedEvenWithoutYield)$' -count=1 -v > /tmp/scout19-custom-set-lower.log 2>&1
go test ./internal/oracle -run '^(TestScoutOwnCustomSetSourceContract|TestScoutCollectionStrongEdgeMutants)$' -count=1 -v > /tmp/scout19-custom-set-contract.log 2>&1
go vet ./internal/load ./internal/lower ./internal/oracle > /tmp/scout19-custom-set-vet.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/scout19-custom-set-counts.log 2>&1
```

All pass. Counts are refreshed; the refused custom collection has no native count row. It is held by explicit Refused tests, since the standard non-lowering oracle registry accepts NotYet only. Frozen census roots retired by this boundary fixture: zero. Evidence is preserved under ruling-custom-set.

### Remaining step 19 work

The string brand representation, concrete presence proof, checked required reads, lexical array range proof and composed add/remove helpers are implemented. The stage 3 MultiMap factory adaptation is written, guarded and verified against Node, with unchanged extracted declarations; its complete native class still stops at the computed string-tag accessor. Program-region collection allocation remains a step 06 implementation dependency. Original createSet needs its own union bucket representation, suspended iterators and region-safe ownership; it stays refused and is never lowered as built-in Set. Mixed/nullish union keys and MapLike remain the separately recorded representation/index-signature work. No complete tsc entry root or hidden-byte recovery is claimed.

### Final delivery verification

Merged area-next dcdbb909 in 04b105cf. Ruled implementation and evidence commits: 4274af93 (string brands), 9ed45ccb (presence), 79090307 (range), dcfead5e (MultiMap adaptation), 11ae89f4 (strong-edge fixture), ca3d58d5 (original custom Set boundary). Each finished piece was pushed to codex/scout-map-keys. No PR or merge to an owned main/area branch was made.

The final compiler replay on the unchanged frozen corpus confirms 159 probes, 137 originally reproduced signatures, 34 still reproduced and 103 retired diagnostic signatures: 83 general Map-key sites, 14 __String Set sites and six Path Set sites. Eighteen distinct selected units have no remaining lowering blockers. The entry-root program remains checker-rejected; complete tsc entry roots compiling is zero, and no hidden-byte recovery is measured. The new composition adaptation is verified separately, not mixed into this frozen-source comparison. Complete final findings and the exact probe list are in ruling-final.

Final scoped commands, with environment sourced and output redirected:

```sh
python3 stage3/map-keys/replay.py /tmp/scout-map-keys-adapted /tmp/scout19-rules-final.jsonl > /tmp/scout19-rules-final-replay.log 2>&1
go test ./internal/oracle -run '^TestScout' -count=1 -timeout 30m -v > /tmp/scout19-final-mutants.log 2>&1
go test ./internal/lower -run '^TestScout' -count=1 -v > /tmp/scout19-final-lower.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m > /tmp/scout19-final-counts.log 2>&1
go vet ./internal/load ./internal/lower ./internal/oracle > /tmp/scout19-final-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_(map|array|multimap)_.*[.]a$' -count=1 -timeout 30m -v > /tmp/scout19-final-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^stage3$/^map-keys$/^iterator_pairs[.]a$' -count=1 -timeout 30m -v > /tmp/scout19-final-pairs.log 2>&1
```

All pass. The last uncached native fixtures record 30 native misses and 22 Node misses; the pair fixture records three native misses and two Node misses; both have zero hits. Recorded counts match without an update. Vet has no findings. No whole package test or full gate was run.

Every executed mutant and its catcher:

| Mutant | What caught it |
|---|---|
| Inventory count changed | Exact inventory consistency assertion |
| Lost lookup in object-key fixture | Source Node stdout, clean native sanitizer/leaks |
| Lost lookup in numeric-key fixture | Source Node stdout, clean native sanitizer/leaks |
| Lost lookup in reference-key fixture | Source Node stdout, clean native sanitizer/leaks |
| Union key admission enabled | Mixed-key outcome became accepted and failed the gap test |
| Pair representation guard weakened | Widened pair program became accepted and failed the gap test |
| Second pair insertion omitted | Source Node stdout, clean native sanitizer/leaks |
| Brand boundary check omitted | Required panic contract; both mutant backends exit 0 |
| Presence effect guard bypassed | Alias/call/key-update flow controls and invalidated-read panic contract |
| Array range effect guard bypassed | Alias-pop flow control and invalidated-read panic contract |
| Concrete generic undefined allowance omitted | Node agrees on stored undefined; both mutant backends panic instead |
| Empty MultiMap bucket retained in source | Upstream Node size comparison |
| Empty MultiMap bucket retained in generated native code | Source Node size comparison, clean native sanitizer/leaks |
| Applied adapter source changed | Exact replacement drift guard exits 1 |
| Computed accessor prepass guard omitted | Native factory gap test fails with the old compiler panic |
| Strong key edge dropped | ASAN heap-use-after-free after producer return |
| Strong value edge dropped | ASAN heap-use-after-free after producer return |
| Custom Set replaced with built-in Set | Original Node equality, membership, size, representative and string-tag observations |

No backend compile-warning kill is counted. Compiler panic, source-drift and refusal tests are distinct from successful-runtime semantic mutants. Remaining native implementation dependencies are listed above; the branch advances step 19 but does not complete it.


### Union keys

Union keys now use a dedicated tagged-key Map mode in both Map and Set lowering. Numeric boxes hash and compare by SameValueZero, strings by content, booleans by their constant values, and references by identity. Kind discrimination keeps 1 separate from '1'. Ordinary union strict equality retains its existing NaN behavior. The Map owns the keys strongly; replacement consumes the incoming key, deletion and clear release the stored key, and copied entries retain their keys. Inserting zero creates a new positive-zero box rather than changing a shared input box.

scout_union_keys.a reduces core.ts mapEntries's generic key parameter and the Set deduplication path. It exercises primitive unions, allocated string equality, separate equal-looking objects, arrays, closures, copying a Map, overwrite order, delete and reinsert, clear, NaNs and signed zero. A negative-zero parameter remains negative outside the collection. mixed_map.a now lowers and is registered as an acceptance fixture rather than a gap.

The source Node oracle, generated JavaScript, sanitized native, release native and successful-program leak checks pass for both fixtures. TestScoutUnionKeyMutants runs three clean native semantic mutants: numeric 1 converted to its string spelling, NaN insertion omitted with owned arguments released, and stored zero changed back to negative zero. Each disagrees with Node stdout, without sanitizer, leak or compile-warning failure. Counts are refreshed. Exact commands and output are preserved in stage3/map-keys/evidence/union-keys; no whole package test or full gate was run.

The frozen 159-probe replay decreases reproduced diagnostics from 34 to 33. One additional signature retires at tracing.ts:224:38, the Map<object, number> key representation. The original merged-baseline total is now 104 retired signatures. Complete tsc entry roots compiling remains zero; hidden-byte recovery is not measured. The other dependencies exposed by that selected unit remain recorded in after.jsonl.gz. Nullish keys are the next piece. createSet, full MultiMap and MapLike remain assigned to steps 20, 06 and 22.

Commands: go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^scout_union_keys.a$' -v -count=1 -timeout 30m; the same test with /stage3/map-keys/^mixed_map.a$; go test ./internal/oracle -run '^TestScoutUnionKeyMutants$' -v -count=1; go test ./internal/lower -run '^TestScoutMapKeyOutcomes$' -count=1; go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts; go vet ./internal/lower ./internal/native ./internal/oracle; python3 stage3/map-keys/replay.py /tmp/scout-map-keys-adapted /tmp/step19-union-replay.jsonl. Oracle fixture runs use ADAMIC_GATE_UNCACHED=1; all outputs go to logs.
