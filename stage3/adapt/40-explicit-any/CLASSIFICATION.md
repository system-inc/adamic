# Classification of explicit any tokens

Current adaptation status: 69 of the original 210 tokens are done; 141 remain.
Site 138 is done under Filesystem entry classification. The 207-token census
below is the historical classification snapshot; brands, enum display and
diagnostic arguments were completed earlier as recorded in PROGRESS.md.

This is a disjoint classification of the 207-token adapted census, not the original 210-token census. Every token has an ID and exact file:line:column in evidence/classification.json. Multiple tokens on one line remain separate sites. The counts below are semantic patterns; erased syntax is an independent axis, not a catch-all class. All suggested types remain proposals until checked against owners and consumers. A type-only candidate is not a completed proof.

| Pattern | Tokens | Proposed rule |
| --- | ---: | --- |
| JSON and config value pipeline | 38 | Unknown plus checked narrowing at genuinely untyped JSON/API boundaries; propagate validated recursive JSON/config value types through the internal pipeline. |
| Nominal and phantom markers | 36 | Real type: undefined for absent phantom marker values, preserving marker keys and optionality; audit all construction, reads and assignability before applying. |
| Enum and namespace display reflection | 23 | Real type: numeric-enum reflection map with string-or-number values, and the actual typeof namespace/enum at each owner. |
| Timer handles and variadic timer forwarding | 16 | Generic parameters for the host timer handle and argument tuple, carried together through registration, storage, callback and cancellation declarations. |
| Incomplete runtime constructors and storage initialization | 11 | Decline until constructor return and staged initialization contracts are modeled at their owners. |
| Generic enumerable copy and metadata copy | 10 | Decline until enumerable copy results and writable key relationships have an owner contract. |
| Standard runtime feature probes | 8 | Real type: existing Array, String and Error APIs, with optional host availability declared at their owners where older runtimes omit the API. |
| AST subtype properties and mutable manufactured nodes | 6 | Real type: the precise node union or manufactured node shape at the owning declaration. |
| Diagnostic substitution arguments | 6 | Real type: the diagnostic substitution domain established from callers, string \| number \| boolean \| readonly string[] \| SourceFile \| undefined, with scanner/parser arg0 restricted to optional string \| number. |
| Display projections on known internal objects | 6 | Real type: declare the actual optional display fields or recursive option display value at the object owner, preserving existing control flow. |
| Generic builder program factory escapes | 5 | Decline until the generic program type is tied to the selected factory and persisted builder result at the owner. |
| Broad generic API shape references | 4 | Real type: use the declared generic bound, BuilderProgram for ProgramHost and Node for NodeArray. |
| Sorted array brand assertions | 4 | Decline until sortedness proof and its branded result are expressible at the sorting/empty-array owner. |
| Heterogeneous node callback dispatch | 3 | Generic parameter associating the node kind and concrete node type with the callback input/output at registration and dispatch. |
| Generic defaults and constraints | 3 | Generic parameter: retain the argument tuple and checked input/output relationship explicitly, removing permissive defaults after auditing explicit generic callers. |
| Optional process and stream extensions | 3 | Unknown plus checked narrowing only for genuinely untyped host extensions at process.browser, process.recordreplay and stdout._handle. |
| Resolution instrumentation payload | 3 | Generic parameter for before/after instrumentation data, with void for an ignored after-hook result. |
| Heterogeneous watcher callbacks | 3 | Generic parameter carrying the watcher callback argument tuple through the cache and logging wrapper. |
| Recursive compiler data comparison | 2 | Real type: recursive compiler data domain including object/array, scalar, null/undefined and ignored function members. |
| Mutable Debug assertion table | 2 | Decline: the restore assignment stores a cache entry object where the declared slot is callable; a truthful callable replacement exposes a runtime contract mismatch. |
| Pragma argument name domain | 2 | Real type: string as the PragmaArgumentSpecification name parameter bound. |
| Optional builder program feature access | 2 | Real type: an optional getSemanticDiagnosticsOfNextAffectedFile capability on the program owner, preserving the existing presence test. |
| Generic array predicate input | 1 | Generic parameter for the internal value input while retaining the readonly unknown[] checked predicate. |
| Default string comparator escape | 1 | Decline until the string default comparer is restricted to a string element branch or a generic caller supplies its comparator. |
| System to config-host shape | 1 | Real type: the actual ParseConfigFileHost capability subset at the host declaration, proving required optional System methods before construction. |
| Distributive conditional type | 1 | Generic parameter: use the identity condition U extends U to preserve distribution without a top-type escape. |
| Filesystem entry classification (done) | 1 | Proven type: fs.Stats \| fs.Dirent \| undefined, with existing !stat continue retaining narrowing. |
| Delegated generator next-input type | 1 | Generic parameter carrying the delegated iterator next-input type through flatMapIterator. |
| Key-only mapped type | 1 | Real type: never as the unobserved mapped value, retaining the existing key remapping. |
| Module resolution enum key | 1 | Real type: the actual ModuleResolutionKind string-key domain at the declaration producing moduleKindName. |
| Variable-like declaration family | 1 | Real type: the variable-like declaration union including JSDoc property-like tags required by its callers. |
| Resolution result update | 1 | Real type: declare the mutable resolvedModule result field at its actual owner. |
| Unused scanner reduction state | 1 | Real type: undefined, the state passed by both getLeadingCommentRanges and getTrailingCommentRanges. |

Total: 207 tokens, 28 files, 33 disjoint classes. No token remains in an unclassified other bucket.

## Decision

Stop: largest class is JSON/config, 38 sites, and its whole-class rule requires runtime narrowing. No further adaptation or consumer edit.

The largest class is 38 JSON/config sites, not the 36 brands. For example, raw config extension merging writes ownConfig.raw.include/exclude/files and reads compileOnSave directly; public config conversion accepts arbitrary values before validation. Unknown would expose property-access errors. Adding object/field validation changes emitted JavaScript; a broad assertion would evade the requested proof. I therefore did not apply another class as a substitute for the largest class. The prior three-owner adapter and its proof are unchanged.

The guessed perfLogger/inspector class has zero sites in this compiler census. sys has timers, standard feature probes, three process/stream extensions and one typed filesystem union, which are separated above. Debug restoration is a mutation mismatch, not display reflection. Overloads and declaration shapes cross several classes: JSON, timers, diagnostics, generic API shapes and brands. Grouping all type-only syntax together would obscure the actual contracts.

Unknown is proposed only for the explicit boundary sites listed in classification.json. At 9/10/19/29/30 config APIs accept external values; 98/180 parse package/JSON text; 119 checks a sourcemap input; 61/136/137 inspect untyped runtime extensions. Internal conversion outputs and intermediate config values should carry validated real types. Existing checked leaf predicates at 20/22/118 can be adapted with their input relationship after boundary work; they do not justify blanket unknown across the pipeline.

## Locations and observations

### JSON and config value pipeline (38)

Config parsing and option validation, package.json, sourcemaps and JSON.parse. Boundary sites are 9, 10, 19, 29, 30, 98, 119 and 180: config/read APIs, public config converters, parsed package data, sourcemap guard and JSON.parse result. Other sites carry these values, construct JSON-shaped values, or validate them; they are not independent untyped host inputs. Recovery can produce undefined. parseOwnConfigOfJson and raw config merging read/write fields without proving the input object and each field shape. Existing truthiness and hasProperty checks do not establish those contracts. Whole-class unknown substitution would require runtime checks or expose consumer errors; declaring every value JSON would also falsely restrict public API inputs before validation.

Rule: Unknown plus checked narrowing at genuinely untyped JSON/API boundaries; propagate validated recursive JSON/config value types through the internal pipeline.

Locations:
- 9: `src/compiler/commandLineParser.ts:2267:110`
- 10: `src/compiler/commandLineParser.ts:2277:91`
- 11: `src/compiler/commandLineParser.ts:2430:16`
- 12: `src/compiler/commandLineParser.ts:2441:4`
- 13: `src/compiler/commandLineParser.ts:2467:84`
- 14: `src/compiler/commandLineParser.ts:2484:4`
- 15: `src/compiler/commandLineParser.ts:2494:8`
- 16: `src/compiler/commandLineParser.ts:2495:23`
- 17: `src/compiler/commandLineParser.ts:2538:110`
- 18: `src/compiler/commandLineParser.ts:2606:79`
- 19: `src/compiler/commandLineParser.ts:3013:50`
- 20: `src/compiler/commandLineParser.ts:3038:31`
- 21: `src/compiler/commandLineParser.ts:3060:11`
- 22: `src/compiler/commandLineParser.ts:3300:45`
- 23: `src/compiler/commandLineParser.ts:3354:48`
- 24: `src/compiler/commandLineParser.ts:3377:10`
- 25: `src/compiler/commandLineParser.ts:3406:11`
- 26: `src/compiler/commandLineParser.ts:3497:11`
- 27: `src/compiler/commandLineParser.ts:3599:16`
- 28: `src/compiler/commandLineParser.ts:3725:57`
- 29: `src/compiler/commandLineParser.ts:3733:61`
- 30: `src/compiler/commandLineParser.ts:3739:61`
- 31: `src/compiler/commandLineParser.ts:3752:60`
- 32: `src/compiler/commandLineParser.ts:3765:60`
- 33: `src/compiler/commandLineParser.ts:3771:57`
- 34: `src/compiler/commandLineParser.ts:3775:94`
- 35: `src/compiler/commandLineParser.ts:3776:94`
- 36: `src/compiler/commandLineParser.ts:3777:94`
- 37: `src/compiler/commandLineParser.ts:3803:12`
- 38: `src/compiler/commandLineParser.ts:3835:90`
- 39: `src/compiler/commandLineParser.ts:3880:22`
- 40: `src/compiler/commandLineParser.ts:3886:4`
- 41: `src/compiler/commandLineParser.ts:4276:48`
- 98: `src/compiler/moduleSpecifiers.ts:1256:54`
- 118: `src/compiler/sourcemap.ts:404:28`
- 119: `src/compiler/sourcemap.ts:408:28`
- 173: `src/compiler/types.ts:7705:11`
- 180: `src/compiler/utilities.ts:7809:45`

### Nominal and phantom markers (36)

Marker properties on branded primitives, arrays, AST families and SymbolLinks, including two cache-key markers without Brand suffix. This is distinct from JSON and overload signatures. No replacement is proved in this classification: changing marker value types can affect assignability and public declarations even though annotations erase.

Rule: Real type: undefined for absent phantom marker values, preserving marker keys and optionality; audit all construction, reads and assignability before applying.

Locations:
- 1: `src/compiler/builder.ts:1074:88`
- 2: `src/compiler/builder.ts:1076:100`
- 3: `src/compiler/checker.ts:1452:32`
- 62: `src/compiler/corePublic.ts:18:28`
- 63: `src/compiler/corePublic.ts:22:28`
- 96: `src/compiler/moduleNameResolver.ts:975:66`
- 97: `src/compiler/moduleNameResolver.ts:1113:65`
- 103: `src/compiler/path.ts:457:67`
- 142: `src/compiler/transformers/esnext.ts:774:103`
- 147: `src/compiler/types.ts:23:44`
- 148: `src/compiler/types.ts:958:27`
- 149: `src/compiler/types.ts:968:28`
- 150: `src/compiler/types.ts:974:26`
- 151: `src/compiler/types.ts:1757:24`
- 152: `src/compiler/types.ts:1944:25`
- 153: `src/compiler/types.ts:1981:26`
- 154: `src/compiler/types.ts:2064:36`
- 155: `src/compiler/types.ts:2185:21`
- 156: `src/compiler/types.ts:2397:23`
- 157: `src/compiler/types.ts:2412:28`
- 158: `src/compiler/types.ts:2418:29`
- 159: `src/compiler/types.ts:2449:35`
- 160: `src/compiler/types.ts:2453:29`
- 161: `src/compiler/types.ts:2457:30`
- 162: `src/compiler/types.ts:2785:30`
- 163: `src/compiler/types.ts:2982:26`
- 164: `src/compiler/types.ts:2997:55`
- 165: `src/compiler/types.ts:3010:26`
- 166: `src/compiler/types.ts:3034:26`
- 167: `src/compiler/types.ts:3168:26`
- 168: `src/compiler/types.ts:3322:22`
- 169: `src/compiler/types.ts:3568:25`
- 170: `src/compiler/types.ts:3573:24`
- 171: `src/compiler/types.ts:3903:22`
- 172: `src/compiler/types.ts:6059:24`
- 179: `src/compiler/utilities.ts:4289:40`

### Enum and namespace display reflection (23)

formatEnum and its cache reflect numeric enum objects; getEnumMembers filters typeof value === number. Namespace casts refer to known enum exports. FileIncludeKind indexes a real reverse enum map. These are not arbitrary untyped host objects.

Rule: Real type: numeric-enum reflection map with string-or-number values, and the actual typeof namespace/enum at each owner.

Locations:
- 69: `src/compiler/debug.ts:389:55`
- 70: `src/compiler/debug.ts:420:37`
- 71: `src/compiler/debug.ts:422:41`
- 72: `src/compiler/debug.ts:445:40`
- 73: `src/compiler/debug.ts:449:40`
- 74: `src/compiler/debug.ts:453:40`
- 75: `src/compiler/debug.ts:457:41`
- 76: `src/compiler/debug.ts:461:41`
- 77: `src/compiler/debug.ts:465:41`
- 78: `src/compiler/debug.ts:469:41`
- 79: `src/compiler/debug.ts:473:41`
- 80: `src/compiler/debug.ts:477:41`
- 81: `src/compiler/debug.ts:481:41`
- 82: `src/compiler/debug.ts:485:41`
- 83: `src/compiler/debug.ts:489:41`
- 84: `src/compiler/debug.ts:493:41`
- 85: `src/compiler/debug.ts:497:42`
- 86: `src/compiler/debug.ts:501:40`
- 87: `src/compiler/debug.ts:505:40`
- 88: `src/compiler/debug.ts:509:41`
- 89: `src/compiler/debug.ts:540:62`
- 91: `src/compiler/emitter.ts:5790:147`
- 106: `src/compiler/program.ts:3526:50`

### Timer handles and variadic timer forwarding (16)

Includes five callback/rest argument tokens and eleven opaque timer handle tokens. System and WatchHost timer contracts are host-polymorphic; forwarding an opaque handle does not justify unknown plus invented runtime narrowing. A blanket Node timer type would exclude other supported hosts.

Rule: Generic parameters for the host timer handle and argument tuple, carried together through registration, storage, callback and cancellation declarations.

Locations:
- 123: `src/compiler/sys.ts:50:48`
- 124: `src/compiler/sys.ts:50:82`
- 125: `src/compiler/sys.ts:51:39`
- 128: `src/compiler/sys.ts:463:24`
- 131: `src/compiler/sys.ts:617:36`
- 132: `src/compiler/sys.ts:1440:37`
- 133: `src/compiler/sys.ts:1440:74`
- 134: `src/compiler/sys.ts:1440:82`
- 135: `src/compiler/sys.ts:1441:30`
- 143: `src/compiler/tsbuildPublic.ts:424:37`
- 200: `src/compiler/watchPublic.ts:169:37`
- 201: `src/compiler/watchPublic.ts:169:74`
- 202: `src/compiler/watchPublic.ts:169:82`
- 203: `src/compiler/watchPublic.ts:171:30`
- 204: `src/compiler/watchPublic.ts:443:31`
- 205: `src/compiler/watchPublic.ts:444:51`

### Incomplete runtime constructors and storage initialization (11)

NodeLinks construction, Symbol.links initialization and nine objectAllocator constructor casts claim richer interfaces than the bare constructor initially supplies. A cast to the final interface would hide the same missing proof.

Rule: Decline until constructor return and staged initialization contracts are modeled at their owners.

Locations:
- 4: `src/compiler/checker.ts:2940:76`
- 183: `src/compiler/utilities.ts:8486:14`
- 184: `src/compiler/utilities.ts:8553:39`
- 185: `src/compiler/utilities.ts:8554:41`
- 186: `src/compiler/utilities.ts:8555:51`
- 187: `src/compiler/utilities.ts:8556:52`
- 188: `src/compiler/utilities.ts:8557:45`
- 189: `src/compiler/utilities.ts:8558:43`
- 190: `src/compiler/utilities.ts:8559:39`
- 191: `src/compiler/utilities.ts:8560:49`
- 192: `src/compiler/utilities.ts:8561:61`

### Generic enumerable copy and metadata copy (10)

clone, extend, copyProperties and SourceFile metadata copying enumerate string keys and populate initially empty objects. T or T1 & T2 does not prove that copying enumerable own strings preserves prototypes, symbol keys or all members of arbitrary input types.

Rule: Decline until enumerable copy results and writable key relationships have an owner contract.

Locations:
- 49: `src/compiler/core.ts:1479:19`
- 50: `src/compiler/core.ts:1482:37`
- 51: `src/compiler/core.ts:1496:35`
- 52: `src/compiler/core.ts:1499:24`
- 53: `src/compiler/core.ts:1499:46`
- 54: `src/compiler/core.ts:1505:24`
- 55: `src/compiler/core.ts:1505:45`
- 56: `src/compiler/core.ts:1516:23`
- 92: `src/compiler/factory/nodeFactory.ts:6145:22`
- 93: `src/compiler/factory/nodeFactory.ts:6145:43`

### Standard runtime feature probes (8)

Array.at, Error.captureStackTrace, String.fromCodePoint and Error.stackTraceLimit. Some are standard built-ins, others Node/V8 additions; availability is already tested. They are not parsed JSON.

Rule: Real type: existing Array, String and Error APIs, with optional host availability declared at their owners where older runtimes omit the API.

Locations:
- 47: `src/compiler/core.ts:1066:119`
- 48: `src/compiler/core.ts:1067:36`
- 66: `src/compiler/debug.ts:200:23`
- 67: `src/compiler/debug.ts:201:23`
- 116: `src/compiler/scanner.ts:4063:77`
- 117: `src/compiler/scanner.ts:4063:122`
- 126: `src/compiler/sys.ts:75:19`
- 127: `src/compiler/sys.ts:76:19`

### AST subtype properties and mutable manufactured nodes (6)

Fresh parentName.isTypeOf mutation; parent access on JSDoc parsing elements; optional declaration.type access in symbolWalker. Existing union narrowing and actual factory return shapes must be reconciled at owners.

Rule: Real type: the precise node union or manufactured node shape at the owning declaration.

Locations:
- 5: `src/compiler/checker.ts:6834:40`
- 101: `src/compiler/parser.ts:10207:25`
- 102: `src/compiler/parser.ts:10208:40`
- 120: `src/compiler/symbolWalker.ts:207:27`
- 121: `src/compiler/symbolWalker.ts:207:46`
- 122: `src/compiler/symbolWalker.ts:208:41`

### Diagnostic substitution arguments (6)

Diagnostic formatter, resolution trace and scanner/parser diagnostic arguments. They are formatted into messages, not stored arbitrary callback tuples. The proposed domain must still be checked against services and public callers.

Rule: Real type: the diagnostic substitution domain established from callers, string | number | boolean | readonly string[] | SourceFile | undefined; scanner/parser arg0 is optional string | number.

Locations:
- 8: `src/compiler/commandLineParser.ts:2206:72`
- 94: `src/compiler/moduleNameResolver.ts:118:88`
- 100: `src/compiler/parser.ts:2176:75`
- 112: `src/compiler/scanner.ts:39:81`
- 114: `src/compiler/scanner.ts:1161:87`
- 115: `src/compiler/scanner.ts:1162:94`

### Display projections on known internal objects (6)

__debugFlags, function.name, nested compiler option values, Type.objectFlags/intrinsicName and SourceFile.path. The transformer cast follows a SourceFile kind test. Optional Type fields differ by subtype; internal values are not host input warranting unknown.

Rule: Real type: declare the actual optional display fields or recursive option display value at the object owner, preserving existing control flow.

Locations:
- 6: `src/compiler/checker.ts:10953:77`
- 68: `src/compiler/debug.ts:373:29`
- 95: `src/compiler/moduleNameResolver.ts:954:69`
- 139: `src/compiler/tracing.ts:231:42`
- 140: `src/compiler/tracing.ts:317:41`
- 141: `src/compiler/transformer.ts:338:116`

### Generic builder program factory escapes (5)

Default emit-and-semantic factory and readBuilderProgram results are cast to arbitrary caller T. T can denote a different builder family; a replacement assertion T would perpetuate the mismatch.

Rule: Decline until the generic program type is tied to the selected factory and persisted builder result at the owner.

Locations:
- 146: `src/compiler/tsbuildPublic.ts:1355:64`
- 197: `src/compiler/watch.ts:862:91`
- 198: `src/compiler/watchPublic.ts:150:88`
- 199: `src/compiler/watchPublic.ts:151:61`
- 206: `src/compiler/watchPublic.ts:555:75`

### Broad generic API shape references (4)

ProgramHost readFile/host projections and emit NodeArray callbacks. Type-only generic instantiations are not evidence of untyped input; the projected member may not depend on the generic argument.

Rule: Real type: use the declared generic bound, BuilderProgram for ProgramHost and Node for NodeArray.

Locations:
- 104: `src/compiler/program.ts:393:27`
- 174: `src/compiler/types.ts:9857:64`
- 175: `src/compiler/types.ts:9858:63`
- 196: `src/compiler/watch.ts:754:69`

### Sorted array brand assertions (4)

Empty arrays are vacuously sorted; deduplication preserves its sorted input. as any as currently bridges the nominal SortedArray/SortedReadonlyArray marker. Merely swapping the bridge assertion for the target type would not constitute the owner proof requested.

Rule: Decline until sortedness proof and its branded result are expressible at the sorting/empty-array owner.

Locations:
- 43: `src/compiler/core.ts:736:50`
- 44: `src/compiler/core.ts:759:28`
- 45: `src/compiler/core.ts:764:18`
- 105: `src/compiler/program.ts:3284:130`

### Heterogeneous node callback dispatch (3)

ParenthesizerRule, forEachChildTable and visitEachChildTable erase heterogeneous function types. A broad Node callback loses contravariant input relationships; the stored callback and its invocation must share a type parameter.

Rule: Generic parameter associating the node kind and concrete node type with the callback input/output at registration and dispatch.

Locations:
- 90: `src/compiler/emitter.ts:1274:53`
- 99: `src/compiler/parser.ts:1253:78`
- 194: `src/compiler/visitorPublic.ts:603:82`

### Generic defaults and constraints (3)

maybeBind already carries A; cast and tryCast default TIn to any. Replacing the default with an unrelated universal type can break predicate variance or weaken the TOut extends TIn proof.

Rule: Generic parameter: retain the argument tuple and checked input/output relationship explicitly, removing permissive defaults after auditing explicit generic callers.

Locations:
- 57: `src/compiler/core.ts:1522:40`
- 59: `src/compiler/core.ts:1778:49`
- 60: `src/compiler/core.ts:1783:46`

### Optional process and stream extensions (3)

These three sites inspect external runtime extensions: browser marker, replay marker and private stream handle with optional setBlocking. Current truthiness or optional chaining does not prove an arbitrary handle method callable. Changing to checked unknown cannot be presumed byte-identical.

Rule: Unknown plus checked narrowing only for genuinely untyped host extensions at process.browser, process.recordreplay and stdout._handle.

Locations:
- 61: `src/compiler/core.ts:2593:25`
- 136: `src/compiler/sys.ts:1594:203`
- 137: `src/compiler/sys.ts:1609:51`

### Resolution instrumentation payload (3)

ResolutionCache host hooks return a before payload and pass it unchanged into the matching after hook; after return value is ignored. This is an opaque correlated payload, not inspector or perfLogger code.

Rule: Generic parameter for before/after instrumentation data, with void for an ignored after-hook result.

Locations:
- 108: `src/compiler/resolutionCache.ts:209:8`
- 109: `src/compiler/resolutionCache.ts:215:15`
- 110: `src/compiler/resolutionCache.ts:216:8`

### Heterogeneous watcher callbacks (3)

File and directory callback families have different arities; sys uses a never middle parameter as a variance workaround, while watchUtilities logs and forwards spread arguments. Cannot prove callback compatibility by changing just the spread array element type.

Rule: Generic parameter carrying the watcher callback argument tuple through the cache and logging wrapper.

Locations:
- 129: `src/compiler/sys.ts:516:30`
- 130: `src/compiler/sys.ts:516:58`
- 207: `src/compiler/watchUtilities.ts:814:23`

### Recursive compiler data comparison (2)

compareDataObjects recursively compares compiler option data and skips functions. It is not restricted to JSON, and unknown is not justified by an untyped external boundary here. Object indexing and recursive null behavior need an owner contract.

Rule: Real type: recursive compiler data domain including object/array, scalar, null/undefined and ignored function members.

Locations:
- 181: `src/compiler/utilities.ts:8147:41`
- 182: `src/compiler/utilities.ts:8147:51`

### Mutable Debug assertion table (2)

setAssertionLevel assigns cachedFunc, not cachedFunc.assertion, to Debug[key]. shouldAssertFunction installs noop. These are writes, not reflection for display. Fixing the restoration statement changes JavaScript, outside this unit.

Rule: Decline: the restore assignment stores a cache entry object where the declared slot is callable; a truthful callable replacement exposes a runtime contract mismatch.

Locations:
- 64: `src/compiler/debug.ts:170:31`
- 65: `src/compiler/debug.ts:190:23`

### Pragma argument name domain (2)

Type-level pragma shapes already constrain the name parameter to strings. These are two generic instantiations, not arbitrary runtime payloads.

Rule: Real type: string as the PragmaArgumentSpecification name parameter bound.

Locations:
- 177: `src/compiler/types.ts:10323:91`
- 178: `src/compiler/types.ts:10332:155`

### Optional builder program feature access (2)

A builder family is probed for a method and conditionally invoked. Calling the method on a union requires declaring the optional capability; claiming all programs have the full semantic builder interface is too strong.

Rule: Real type: an optional getSemanticDiagnosticsOfNextAffectedFile capability on the program owner, preserving the existing presence test.

Locations:
- 144: `src/compiler/tsbuildPublic.ts:992:34`
- 145: `src/compiler/tsbuildPublic.ts:993:33`

### Generic array predicate input (1)

isArray delegates to the runtime Array.isArray check. The incoming value is not inherently untyped JSON; a generic input preserves concrete caller types.

Rule: Generic parameter for the internal value input while retaining the readonly unknown[] checked predicate.

Locations:
- 58: `src/compiler/core.ts:1750:32`

### Default string comparator escape (1)

deduplicateSorted falls back to compareStringsCaseSensitive while T is unconstrained. Casting the string comparator to Comparer<T> has a missing domain proof.

Rule: Decline until the string default comparer is restricted to a string element branch or a generic caller supplies its comparator.

Locations:
- 46: `src/compiler/core.ts:810:120`

### System to config-host shape (1)

watch adapts a typed System into a config-reading host. It is not untyped host input, and the any cast may conceal optional versus required capabilities.

Rule: Real type: the actual ParseConfigFileHost capability subset at the host declaration, proving required optional System methods before construction.

Locations:
- 195: `src/compiler/watch.ts:221:49`

### Distributive conditional type (1)

UnionToIntersection uses a naked type parameter conditional solely to distribute the union. Must verify never, unions and generic instantiations before any change.

Rule: Generic parameter: use the identity condition U extends U to preserve distribution without a top-type escape.

Locations:
- 176: `src/compiler/types.ts:10320:49`

### Filesystem entry classification (done: 1, remaining: 0)

Assignments come from statSync or readdirSync dirents. This host API has real declarations and common isFile/isDirectory methods; unknown plus new checks is unnecessary.

Proven type: `import("fs").Stats | import("fs").Dirent | undefined`.
The local statSync declaration returns Stats | undefined; readdirSync supplies
Dirent. The existing !stat continue retains narrowing. Only the local annotation
changes; no emitted declaration or API sanction changes. Reproduction and
full-build wrong-type mutant are in filesystem-proof.cjs.

Locations:
- 138: `src/compiler/sys.ts:1853:31`

### Delegated generator next-input type (1)

yield* can forward caller next values to the delegated iterator, even though this generator does not name a received value. Declaring it void or undefined without checking delegated iterator contracts would overstate the proof.

Rule: Generic parameter carrying the delegated iterator next-input type through flatMapIterator.

Locations:
- 42: `src/compiler/core.ts:437:136`

### Key-only mapped type (1)

CompilerOptionKeys applies keyof to the mapped type; only its keys are consumed. No runtime value or external input is involved.

Rule: Real type: never as the unobserved mapped value, retaining the existing key remapping.

Locations:
- 193: `src/compiler/utilities.ts:9029:97`

### Module resolution enum key (1)

Runtime string indexes an enum and falls back to Node16; derive the valid key union rather than asserting every arbitrary string a valid enum key.

Rule: Real type: the actual ModuleResolutionKind string-key domain at the declaration producing moduleKindName.

Locations:
- 107: `src/compiler/program.ts:4379:81`

### Variable-like declaration family (1)

widenTypeForVariableLikeDeclaration is an internal checker helper. Derive its declaration union from callers and property reads, rather than treating AST objects as unknown.

Rule: Real type: the variable-like declaration union including JSDoc property-like tags required by its callers.

Locations:
- 7: `src/compiler/checker.ts:12480:87`

### Resolution result update (1)

The primary resolution result is updated after lookup; the any cast is on the assignment target. Replacing it with a value cast would not fix its ownership or readonly contract.

Rule: Real type: declare the mutable resolvedModule result field at its actual owner.

Locations:
- 111: `src/compiler/resolutionCache.ts:563:46`

### Unused scanner reduction state (1)

appendCommentRange does not read _state and its two owners pass undefined explicitly.

Rule: Real type: undefined, the state passed by both getLeadingCommentRanges and getTrailingCommentRanges.

Locations:
- 113: `src/compiler/scanner.ts:952:111`
