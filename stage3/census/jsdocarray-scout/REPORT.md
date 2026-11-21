# JSDocArray census and scout · #3qh3m28

Adamic main `42edc2ed5dbe0c31adcf5c43db507283b5ae3199`. TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`.

## 1. First-error census

The entry root was `src/tsc/tsc.ts` and the directory root was `src/compiler` in the same adapted tree. Both runs used `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_FULL=0`, `LATENT_TAIL=1`, and `LATENT_ASSERT_NO_OUTPUT=1`. A unit’s first stop is its first recorded prepass, registration, lowering, deferred or scoped-tail finding. Refusal and deferred scans and the relation trailer contribute to the all-phase findings. The search matches the full reason as a substring. Sites and unit locations below use the adapted source coordinates.

- **Entry** — pending.
- **Directory** — pending.

## 2. TypeScript references and ownership

Neither `JSDocArray` nor `jsDocCache` appears in the upstream committed `tests/baselines/reference/api/typescript.d.ts`. The public `getJSDocTags(node: Node): readonly JSDocTag[]` signature does appear there at line 8750.

[references.tsv](references.tsv) lists the LanguageService results over the v6.0.3 `src/` tree, including the harness and test runner. The search starts at `JSDocContainer.jsDoc`, `JSDocArray.jsDocCache`, the array type, and the service `NodeObject.jsDoc` declaration. It uses checker symbol references and inspects the syntax of reported writes. In particular, `const { jsDoc } = node as JSDocContainer` in `utilitiesPublic.ts:2555` reads the node property when binding a fresh local. All TypeScript locations in parts 2 and 3 are upstream coordinates.

The copies and property-name-independent paths are below. The generic copies use dynamic keys, so an individual `jsDoc` symbol search cannot resolve their writes.

| Location | Effect |
| --- | --- |
| `src/compiler/factory/nodeFactory.ts:6338`, `cloneIdentifier` | Assigns the original identifier’s `jsDoc` to the clone. |
| `src/compiler/factory/nodeFactory.ts:6399–6404`, `cloneNode` | Copies original own enumerable properties that the new node lacks. This can copy `jsDoc` by identity. |
| `src/compiler/factory/nodeFactory.ts:6136–6145`, `cloneSourceFileWorker` | Copies own enumerable properties when the destination lacks the property, with a special case for `emitNode`. |
| `src/compiler/factory/nodeFactory.ts:6095–6096`, `createRedirectedSourceFile` | Uses the redirect target as the prototype. An inherited property would be an alias even without a write to the destination. |
| `src/compiler/transformers/declarations.ts:828–832`, `preserveJsDoc` | Copies `original.jsDoc` to `updated` when both pass `hasJSDocNodes`. |
| `src/services/textChanges.ts:1196–1204`, `updateJSDocHost` | Assigns the arrow function’s array to the property declaration or enclosing host. |
| `src/compiler/factory/nodeFactory.ts:7420–7428`, `setOriginalNode` | Links `original` and merges emit metadata. The JSDoc array copies are performed by the callers above. |

The cache can change a clone’s result today. I ran this probe with the v6.0.3 JavaScript compiler package. A parsed function with `/** @deprecated hello */` produces one `JSDocDeprecatedTag`. Calling `getJSDocTags` fills the cache. The factory clone has no parent and shares the function’s JSDoc array. Calling `getJSDocTags` on the clone returns the same deprecated tag. Calling `getJSDocCommentsAndTags(clone, true)` and applying the worker’s flattening returns an empty list. This is a concrete difference a fresh node-local cache would expose, regardless of its frequency in the lane workload.

Whole-lane oracle counts are pending.

## 3. Scout of the two roads

### Adaptation

The array-keyed side table is the semantics-preserving choice for cached tag results. A node-local cache would give a fresh clone a separate miss, as the clone probe demonstrates. Copying a node cache on clone would still separate later writes and incremental invalidation for nodes that share the array. Keep the current `node.jsDoc ??= []` so a cache fill on an undocumented node still materializes a fresh empty array. Several callers observe the array’s presence independently of its length.

| Edit site | Proposed source change | JavaScript effect |
| --- | --- | --- |
| `src/compiler/types.ts:959,963–965` | Replace the internal `jsDoc` field type with `JSDoc[]` and remove the now-unused `JSDocArray` interface. | Type-only. |
| `src/compiler/utilities.ts:386,4684` | Remove the type import and type `filterOwnedJSDocTags`’s comments as an ordinary array. | Type-only. |
| `src/compiler/utilitiesPublic.ts:1257–1268` | Keep a private array-keyed cache. In `getJSDocTagsWorker`, get the cached tags if the field holds an array. On a cacheable miss, compute tags as today, initialize the array if needed, and set the table entry under that exact array. Add an internal invalidation helper for the parser. | Adds module-level storage and a helper. Read and write operations become table lookups and updates. The array no longer exposes the own `jsDocCache` key in JavaScript. |
| `src/compiler/parser.ts:3171–3173` and its compiler utility import | Call the cache invalidation helper with the node’s array, if present. Delete the table entry. | Clearing through any alias clears the shared entry. JavaScript replaces the direct property write with a helper call. |

These proposed edits touch no line in the public declaration baseline. A weak-key table has the lifetime of the existing attached cache. A module-wide strong `Map` would retain arrays and tags unless their lifetimes were managed explicitly. Adamic support for an array-keyed weak table needs verification before treating this road as implementable in the lane. The side table changes JavaScript reflection on the array, even when all tag results agree.

### Fixed array representation

The two NodeArray rulings at `docs/rulings.md:78–79` require reserved metadata with separate presence and initialization, invisible to array length and JSON serialization. `nodeArrayLayoutOf` walks declaration ancestry to the library array, resolves numeric indices and the element representation, then gathers declared extras outside the base array’s members. It hands those contracts to `NodeArrayLayout` in `internal/lower/node_array_layout.go`. That schema admits number, boolean and string metadata and requires the four parser NodeArray fields `pos`, `end`, `hasTrailingComma`, and `transformFlags`. JSDocArray needs its own reviewed schema with one optional reference slot for `readonly JSDocTag[]`. The current planner can discover the declaration and member, but neither the metadata type nor the field set satisfies that policy.

Construction needs to allocate the reserved slot before either fresh array escapes. The two allocation sites are the parser’s `mapDefined` result at `parser.ts:1853–1854` and the empty fallback in `utilitiesPublic.ts:1266`. A layout-aware builder could fill the array directly. Another construction route could convert the still-private `mapDefined` result into a metadata-bearing array before publishing it. An ordinary structural assignment of the map result or the empty literal cannot reserve native storage. Reusing a shared empty array would accidentally unify caches for unrelated nodes. The existing literal-cast path in `constructNodeArray` recognizes a literal, reserves the metadata absent, and proves compatible element storage. Its proof does not cover a `mapDefined` call.

Reads at `utilitiesPublic.ts:1259` and `parser.ts:3171` need optional access to the reference slot. Writes at `utilitiesPublic.ts:1267` and `parser.ts:3173` need reference ownership and the existing presence semantics, including the parser’s explicit write of `undefined`. The lowerer and backends need to keep the array representation through ordinary array views, indexing, iteration, length, optional accesses and aliasing. Native metadata stores and reads must account for the tag array’s ownership and release. JavaScript must preserve the real array and the existing observable own-property behavior. The worker’s `??=` targets the node’s array field, while its next line targets array metadata.

Direct reads and writes of the array-valued field are enumerated in [references.tsv](references.tsv). The table below groups array exposures that exercise representation or aliasing. The factory’s `undefined` initializers are also in the reference list. They initialize the node field and allocate no array.

| Source | Array exposure |
| --- | --- |
| `compiler/parser.ts:1852–1854,3171–3173,10114,10224,10267` | Presence assertion, construction, metadata read and clear, AST walks. |
| `compiler/utilitiesPublic.ts:1259–1267,2555–2556` | Cache read, empty fallback, metadata write, array length through the destructured local. |
| `compiler/utilities.ts:1251,4662,4668,4684–4696,4819,10736,12210` | Index zero, forwarding to `filterOwnedJSDocTags`, its `last` and `flatMap` reads, selection of the last comment, iteration and child search. |
| `compiler/binder.ts:2809,2814` and `compiler/checker.ts:16381,43288,49070` | Comment traversal, presence tests and `forEach`. |
| `compiler/factory/nodeFactory.ts:6095–6096,6136–6145,6338,6399–6404` and `compiler/transformers/declarations.ts:830` | Alias and copy sites listed in part 2. |
| `services/textChanges.ts:707–744,1203` | Presence, iteration, comment and tag extraction, and host aliasing. |
| `services/codefixes/convertTypedefToType.ts:249`, `services/jsDoc.ts:496–497`, `services/navigationBar.ts:568` | `flatMap`, selection of the last comment through a local variable, traversal. |
| `services/services.ts:383,521,640,3508`, `services/smartSelection.ts:117–129`, `services/utilities.ts:2157` | Concrete service field type, traversal and a return through an ordinary array type, length and first-comment reads, presence check. |

The layout probe is pending.
