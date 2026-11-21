# JSDocArray census and scout · #3qh3m28

Adamic main `42edc2ed5dbe0c31adcf5c43db507283b5ae3199`. TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`.

I initialized the pinned recursive submodules, cloned the upstream tag with depth one, and prepared the tree with `stage3/apply.sh`. The census binaries came from `stage3/census/latent/make_overlay.py` and `stage3/meter/entry_overlay.py`, each built with `go build -overlay … ./stage3/census/latent/tool` on this main revision. The adapted Git tree after the final existing adaptation is `ccd4de0cc389430ec81f1fc19cc800f54009fcf5`. The entry checker measured 81 source files and reported zero diagnostics.

## 1. First-error census

The entry root was `src/tsc/tsc.ts` and the directory root was `src/compiler` in the same adapted tree. Both runs used `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_FULL=0`, `LATENT_TAIL=1`, and `LATENT_ASSERT_NO_OUTPUT=1`. A unit’s first stop is its first recorded prepass, registration, lowering, deferred or scoped-tail finding. Refusal and deferred scans and the relation trailer contribute to the all-phase findings. The search matches the full reason as a substring. Sites and unit locations below use the adapted source coordinates.

- **Entry** — exit 0, 81 file records, 0 checker diagnostics, relation trailer present. [entry.jsonl.gz](entry.jsonl.gz)
- **Directory** — pending.
- **Entry matches** — 18 first-stop units at 17 distinct sites. 18 findings across all phases at 17 distinct sites.

Distinct matching sites across the available streams — 17.

### `src/compiler/checker.ts:2972:55`

First-stop units

- Entry — `createTypeChecker` at `src/compiler/checker.ts:1490:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/checker.ts:1490:1`. a value of type ParameterPropertyDeclaration, an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/factory/utilities.ts:363:107`

First-stop units

- Entry — `createExpressionForAccessorDeclaration` at `src/compiler/factory/utilities.ts:363:1` (lowering)
- Entry — `createExpressionForObjectLiteralElementLike` at `src/compiler/factory/utilities.ts:466:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/factory/utilities.ts:363:1`. a value of type AccessorDeclaration & { readonly name: BigIntLiteral | ComputedPropertyName | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | StringLiteral; }, a union whose member GetAccessorDeclaration & { readonly name: BigIntLiteral | ComputedPropertyName | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | StringLiteral; } has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object
- Entry — lowering, NotYet, unit `src/compiler/factory/utilities.ts:466:1`. a value of type AccessorDeclaration & { readonly name: BigIntLiteral | ComputedPropertyName | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | StringLiteral; }, a union whose member GetAccessorDeclaration & { readonly name: BigIntLiteral | ComputedPropertyName | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | StringLiteral; } has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/moduleSpecifiers.ts:880:11`

First-stop units

- Entry — `tryGetModuleNameFromAmbientModule` at `src/compiler/moduleSpecifiers.ts:879:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/moduleSpecifiers.ts:879:1`. a value of type (ModuleDeclaration & { name: StringLiteral; }) | undefined, a union whose member ModuleDeclaration & { name: StringLiteral; } has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/transformers/classFields.ts:671:55`

First-stop units

- Entry — `transformClassFields` at `src/compiler/transformers/classFields.ts:356:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/transformers/classFields.ts:356:1`. a value of type PrivateIdentifierInExpression, an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/transformers/generators.ts:1386:43`

First-stop units

- Entry — `transformGenerators` at `src/compiler/transformers/generators.ts:337:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/transformers/generators.ts:337:1`. a value of type InitializedVariableDeclaration, an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/transformers/jsx.ts:109:13`

First-stop units

- Entry — `transformJsx` at `src/compiler/transformers/jsx.ts:93:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/transformers/jsx.ts:93:1`. a value of type (VariableDeclaration & { name: Identifier; }) | undefined, a union whose member VariableDeclaration & { name: Identifier; } has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/transformers/module/module.ts:1926:43`

First-stop units

- Entry — `transformModule` at `src/compiler/transformers/module/module.ts:176:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/transformers/module/module.ts:176:1`. a value of type InitializedVariableDeclaration, an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/transformers/namedEvaluation.ts:412:89`

First-stop units

- Entry — `transformNamedEvaluationOfAssignmentExpression` at `src/compiler/transformers/namedEvaluation.ts:412:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/transformers/namedEvaluation.ts:412:1`. a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) | (... & ... 1 more ... & BinaryExpression), a union whose member AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression has no representation: an intersection whose field left has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/transformers/ts.ts:1429:55`

First-stop units

- Entry — `transformTypeScript` at `src/compiler/transformers/ts.ts:236:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/transformers/ts.ts:236:1`. a value of type ParameterPropertyDeclaration, an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:3570:11`

First-stop units

- Entry — `classOrConstructorParameterIsDecorated` at `src/compiler/utilities.ts:3568:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:3568:1`. a value of type (ConstructorDeclaration & { body: Block; }) | undefined, a union whose member ConstructorDeclaration & { body: Block; } has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:3886:16`

First-stop units

- Entry — `getModuleSpecifierOfBareOrAccessedRequire` at `src/compiler/utilities.ts:3884:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:3884:1`. a value of type RequireOrImportCall, an intersection whose field expression has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:4348:54`

First-stop units

- Entry — `tryGetModuleSpecifierFromDeclaration` at `src/compiler/utilities.ts:4348:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:4348:1`. a value of type JSDocImportTag | CanHaveModuleSpecifier, a union whose member BindingElementOfBareOrAccessedRequire has no representation: an intersection whose field parent has no representation: a union whose member ArrayBindingPattern & { parent: VariableDeclarationInitializedTo<AccessExpression | RequireOrImportCall>; } has no representation: an intersection whose field parent has no representation: a union whose member VariableDeclaration & VariableDeclarationInitializedTo<AccessExpression | RequireOrImportCall> has no representation: an intersection whose field initializer has no representation: a union whose member Expression & ElementAccessExpression has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:4683:47`

First-stop units

- Entry — `filterOwnedJSDocTags` at `src/compiler/utilities.ts:4683:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:4683:1`. a value of type JSDocArray, an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:4704:39`

First-stop units

- Entry — `ownsJSDocTag` at `src/compiler/utilities.ts:4700:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:4700:1`. a value of type HasJSDoc, a union whose member EndOfFileToken has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:4795:11`

First-stop units

- Entry — `getEffectiveJSDocHost` at `src/compiler/utilities.ts:4794:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:4794:1`. a value of type HasJSDoc | undefined, a union whose member EndOfFileToken has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:4829:43`

First-stop units

- Entry — `getTypeParameterFromJsDoc` at `src/compiler/utilities.ts:4829:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:4829:1`. a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; }, an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

### `src/compiler/utilities.ts:5200:20`

First-stop units

- Entry — `getEffectiveBaseTypeNode` at `src/compiler/utilities.ts:5194:1` (lowering)

All matching findings

- Entry — lowering, NotYet, unit `src/compiler/utilities.ts:5194:1`. a value of type ExpressionWithTypeArguments & { readonly expression: Identifier | PropertyAccessEntityNameExpression; }, an intersection whose field expression has no representation: a union whose member LeftHandSideExpression & Identifier has no representation: an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object

## 2. TypeScript references and ownership

Neither `JSDocArray` nor `jsDocCache` appears in the upstream committed `tests/baselines/reference/api/typescript.d.ts`. The public `getJSDocTags(node: Node): readonly JSDocTag[]` signature does appear there at line 8750.

The reference list records the LanguageService results over the v6.0.3 `src/` tree, including the harness and test runner. The search starts at `JSDocContainer.jsDoc`, `JSDocArray.jsDocCache`, the array type, and the service `NodeObject.jsDoc` declaration. Across 709 source roots it found one cache declaration, two cache reads, and two cache writes. The node field has 71 direct write sites. The factory accounts for 66 initializations to `undefined`. The other five are three copies, the parser attachment, and the worker’s empty-array fallback. The search uses checker symbol references and inspects the syntax of reported writes. In particular, `const { jsDoc } = node as JSDocContainer` in `utilitiesPublic.ts:2555` reads the node property when binding a fresh local. The harness produced no direct property matches. The classifier’s `docCommentAndDiagnostics.jsDoc` is a separate result property holding one JSDoc node. All TypeScript locations in parts 2 and 3 are upstream coordinates. [references.tsv](references.tsv)

The copies and property-name-independent paths are below. The generic copies use dynamic keys, so an individual `jsDoc` symbol search cannot resolve their writes. Ordinary `SourceFile` nodes are outside `HasJSDoc`. I included their generic copy and redirect paths in the probe to cover structural aliases.

| Location | Effect |
| --- | --- |
| `src/compiler/factory/nodeFactory.ts:6338`, `cloneIdentifier` | Assigns the original identifier’s `jsDoc` to the clone. |
| `src/compiler/factory/nodeFactory.ts:6399–6404`, `cloneNode` | Copies original own enumerable properties that the new node lacks. This can copy `jsDoc` by identity. |
| `src/compiler/factory/nodeFactory.ts:6136–6145`, `cloneSourceFileWorker` | Copies own enumerable properties when the destination lacks the property, with a special case for `emitNode`. |
| `src/compiler/factory/nodeFactory.ts:6095–6096`, `createRedirectedSourceFile` | Uses the redirect target as the prototype. An inherited property would be an alias even without a write to the destination. |
| `src/compiler/transformers/declarations.ts:828–832`, `preserveJsDoc` | Copies `original.jsDoc` to `updated` when both pass `hasJSDocNodes`. |
| `src/services/textChanges.ts:1196–1204`, `updateJSDocHost` | Assigns the arrow function’s array to the property declaration or enclosing host. |
| `src/compiler/factory/nodeFactory.ts:7420–7428`, `setOriginalNode` | Links `original` and merges emit metadata. The JSDoc array copies are performed by the callers above. |

The cache can change a clone’s result today. I ran this probe with the v6.0.3 JavaScript compiler package. I cloned a parsed function with `/** @deprecated hello */` before filling the cache. The factory clone has no parent and shares the function’s JSDoc array. Calling `getJSDocTags` on the original then fills the cache with one `JSDocDeprecatedTag`. Calling `getJSDocTags` on the clone returns the same deprecated tag. Calling `getJSDocCommentsAndTags(clone, true)` and applying the worker’s flattening returns an empty list. This is a concrete difference a fresh node-local cache would expose, regardless of its frequency in the lane workload.

### Whole-lane oracle

I ran the lane’s default oracle command on a scratch copy of the census’s adapted tree, with all runners, eight workers and no test filter. The probe observes ordinary allocations and copies, prototype redirects, cache fills and invalidations, and reads. Identity tracking uses weak maps within each isolate. A shared array counts once when it first acquires a second distinct host. For a foreign cache hit, the probe computes the reader’s uncached tags and compares ordered tag identities while returning the original cached value.

Oracle phases — install exit 0, build exit 0, tests exit 1. Reporter totals — 106,366 passing, 1 failing, 0 pending. All eight workers recorded exit snapshots. The sole failure was the lane’s expected public API baseline check, with `api/typescript.d.ts` as the only baseline difference.

| Measurement | Count |
| --- | ---: |
| Distinct arrays observed | 771,808 |
| Arrays shared by at least two nodes | 8 |
| Cache fills | 130,748 |
| Incremental invalidations | 128 |
| Invalidations through a host other than the writer | 0 |
| Reads that reused cached tags | 1,032,700 |
| Reads that reused tags written through another host | 0 |
| Foreign hits with a different uncached tag sequence | 0 |
| Reads of cached tags bypassed by `noCache` | 22,504 |
| Bypassed reads through another host | 0 |
| Cache hits without a matching write observation | 0 |
| Foreign hits where recomputation threw | 0 |

All eight sharing observations came from the generic `cloneNode` copy. It yielded eight distinct shared arrays. Parser attachment registered 643,638 holder observations and cache fills registered 130,748. The lane saw zero cross-host cache consumption. The clone-before-fill probe above supplies the behavioral counterexample.

## 3. Scout of the two roads

### Adaptation

The array-keyed side table is the semantics-preserving choice for cached tag results. A node-local cache would give a fresh clone a separate miss, as the clone probe demonstrates. Copying a node cache at clone time would still separate later writes and incremental invalidation. Replacing an existing node’s `jsDoc` would also require invalidating any cache kept on that node. Keep the current `node.jsDoc ??= []` so a cache fill on an undocumented node still materializes a fresh empty array. Several callers observe the array’s presence independently of its length.

| Edit site | Proposed source change | JavaScript effect |
| --- | --- | --- |
| `src/compiler/types.ts:959,963–965` | Replace the internal `jsDoc` field type with `JSDoc[]` and remove the now-unused `JSDocArray` interface. | Type-only. |
| `src/compiler/utilities.ts:386,4684` | Remove the type import and type `filterOwnedJSDocTags`’s comments as an ordinary array. | Type-only. |
| New private compiler module, for example `src/compiler/jsDocCache.ts` | Hold the array-keyed table and its read, write and clear operations. Import these directly from the two callers and leave the namespace barrel alone. | Adds private module-level storage and functions. |
| `src/compiler/utilitiesPublic.ts:1257–1268` and a direct import | Look up the array’s cached tags. On a cacheable miss, compute tags, initialize the array if needed, and set the table entry under that exact array. | Read and write operations become table lookups and updates. The array no longer exposes the own `jsDocCache` key in JavaScript. |
| `src/compiler/parser.ts:3171–3173` and a direct import | Call the cache invalidation helper with the node’s array, if present. Delete the table entry. | Clearing through any alias clears the shared entry. JavaScript replaces the direct property write with a helper call. |

These proposed edits touch no line in the public declaration baseline. A weak-key table has the lifetime of the existing attached cache. A module-wide strong `Map` would retain arrays and tags unless their lifetimes were managed explicitly. Adamic’s `Map` lowering permits array keys by identity. I found no source-level `WeakMap` lowering, so the weak table is a design candidate whose lowering remains unverified. The side table changes JavaScript reflection on the array, even when all tag results agree.

### Fixed array representation

The two NodeArray rulings at `docs/rulings.md:78–79` require reserved metadata with separate presence and initialization, invisible to array length and JSON serialization. `nodeArrayLayoutOf` walks declaration ancestry to the library array, resolves numeric indices and the element representation, then gathers declared extras outside the base array’s members. It hands those contracts to `NodeArrayLayout` in `internal/lower/node_array_layout.go`. That schema admits number, boolean and string metadata and requires the four parser NodeArray fields `pos`, `end`, `hasTrailingComma`, and `transformFlags`. JSDocArray needs its own reviewed schema with one optional reference slot for `readonly JSDocTag[]`. The current planner discovers the declaration and member. Both the metadata type and the field set fail that policy.

Construction needs to allocate the reserved slot before either fresh array escapes. The two allocation sites are the parser’s `mapDefined` result at `parser.ts:1853–1854` and the empty fallback in `utilitiesPublic.ts:1266`. A layout-aware builder could fill the array directly. Another construction route could convert the still-private `mapDefined` result into a metadata-bearing array before publishing it. An ordinary structural assignment of the map result or the empty literal cannot reserve native storage. Reusing a shared empty array would accidentally unify caches for unrelated nodes. The existing literal-cast path in `constructNodeArray` recognizes a literal, reserves the metadata absent, and proves compatible element storage. Its proof does not cover a `mapDefined` call.

Reads at `utilitiesPublic.ts:1259` and `parser.ts:3171` need optional access to the reference slot. Writes at `utilitiesPublic.ts:1267` and `parser.ts:3173` need reference ownership and the existing presence semantics, including the parser’s explicit write of `undefined`. The lowerer and backends need to keep the array representation through ordinary array views, indexing, iteration, length, optional accesses and aliasing. Native metadata stores and reads must account for the tag array’s ownership and release. The allocation also needs to participate in cycle planning. Tags can lead back to AST hosts through parent links, and the current graph guard refuses NodeArray allocations that could lie on a cycle because the native constructor builds them outside graph regions. JavaScript must preserve the real array and the existing observable own-property behavior. The worker’s `??=` targets the node’s array field, while its next line targets array metadata.

The table groups array exposures that exercise representation or aliasing. The reference list includes each direct read and write, including factory `undefined` initializers that allocate no array. [references.tsv](references.tsv)

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

The ordinary array helper bodies that consume these views are `compiler/core.ts:33,399,418,1098,1116,1121,1131`. The parser currently builds the result through the ordinary allocation and append in `core.ts:494–504`. These paths use array elements and length. The cache property is accessed at the four sites named above.

#### Layout probe

I ran `go test -overlay … ./internal/lower -run '^TestScoutJSDocArrayLayout$' -v` on pinned Adamic main. The overlay supplied a scratch test with the upstream array declaration and small stand-ins for `JSDoc` and `JSDocTag`. It also loaded the adapted `src/tsc/tsc.ts` with the normal checker and repeated the calls on its real `src/compiler/types.ts` declaration. Both gave the same result.

| Probe result | Observed |
| --- | --- |
| Library ancestry and indexed element | `JSDoc[]` and `JSDoc`, with object storage for the element. |
| Extra field contract | `jsDocCache`, optional, array-representable `readonly JSDocTag[] \| undefined`. |
| `nodeArrayLayoutOf` | `ok=false`, no fields. The representation reason is the JSDocArray array-member stop. |
| Schema planner with inferred cache field | `stage 0 can't lower NodeArray metadata field jsDocCache of type readonly JSDocTag[] \| undefined yet`. |
| Schema planner with scalar control replacing just the cache type | `NodeArray requires pos with its declared required representation`. |

The test passed on the focused fixture and on the checker-clean adapted entry.
