# Task #dt55mtb — ownership and baseline

Fetched main at `df9d8ee60` and `apple/write-hides-field-4` at `71f6d8b1`. The fix's parent is `84657fa5`. The tracked Stage 3 trees at main and the fix are identical. One baseline `stage3/apply.sh --build-product` completed on the pinned TypeScript 6.0.3 source (`050880ce59e30b356b686bd3144efe24f875ebc8`). Stock TypeScript 6.0.3's AST identified the owners at the reported coordinates in that output.

| Location | Owner | Expression |
| --- | --- | --- |
| checker.ts:16636:13 | `createCanonicalSignature` | `map(signature.typeParameters, …)` |
| checker.ts:16975:70 | `getTypeArguments` | `concatenate(type.target.outerTypeParameters, getEffectiveTypeArguments(…))` |
| checker.ts:17031:47 | `getTypeFromClassOrInterfaceReference` | `concatenate(type.outerTypeParameters, fillMissingTypeArguments(…))` |
| checker.ts:17098:38 | `getTypeFromTypeAliasReference` | `aliasTypeArguments = getTypeArgumentsForAliasSymbol(newAliasSymbol)` |
| checker.ts:19948:37 | `getInferTypeParameters` | `result = append(result, getDeclaredTypeOfSymbol(symbol))` |
| checker.ts:19979:17 | `getTypeFromConditionalTypeNode` | `aliasTypeArguments` in the `ConditionalRoot` initializer |
| factory/nodeFactory.ts:7490:71 and :93 | `mergeEmitNode` | Both arguments to `mergeTokenSourceMapRanges` |

## Ownership stop

`99-correctness-axolotl-r/owners.json` pins the whole `getTypeArguments` function before and after its edit. `99-readonly-inferred-bindings/adapt.cjs` edits the `typeArguments` binding in that same owner and deliberately sorts after that pin. `95-checker-axolotl-b/sites.json` also hashes the whole of `getTypeArguments` and `getTypeFromClassOrInterfaceReference` as dependency proofs. `71-writable-views/tuples.json` edits the signature of `getTypeArguments`. The two concatenate calls accept real compiler type arguments (both second-argument producers return `Type[]`), so typing those values as `TypeParameter` would be incorrect. The supplied `append<TypeParameter>` edit applies to the single append at 19948. Its existing symbol flag guard is `SymbolFlags.TypeParameter`.

The factory helper is local with exactly one caller in the adapted source. Both arguments come from `EmitNode.tokenSourceMapRanges`, declared as `(SourceMapRange | undefined)[]` when present. The helper copies entries into the destination and returns the destination array. Adaptation 30 records two declined indexed-read sites in this helper.

## Two-site build

Landing commit `ce20cd2be` on `df9d8ee60` adds only the guarded append and the local factory-helper signature. Adaptation 30's `sites.json` records declined indexed accesses. Its planner checks expression occurrence and assertion state and has no whole-function pin for this helper. Both owners retained their existing line endings, including the mixed endings in the factory file. A direct replay edited two files on the first invocation and zero files on the second. Running it against the checker-restoring mutant refused the partially adapted tree and retained both original file hashes.

| Proof | Result |
| --- | --- |
| Stock 6.0.3 ES2024/ESNext JavaScript emit, comments retained | Identical for both files. `checker.ts` 2,893,847 bytes, SHA-256 `a14ae29295602e9c35e174bae3fffba2affdca2dfc3655a8f127cd2a13d1fc50`. `nodeFactory.ts` 272,883 bytes, SHA-256 `b072bccc113127b7409bc83c7db6a122696619c799a1af5dc175881e8724d669`. |
| Stock 6.0.3 `tsc -b src --force --pretty false` | All upstream source projects exited 0 before and after, both diagnostic logs empty. |
| Stock 6.0.3 compiler-project JavaScript emit with declaration and source-map output disabled | Identical. `checker.js` 3,048,140 bytes, SHA-256 `bc69c18843ee545c444fb74f70f8dfda3a4c6247ec5c06ef74e82c67a5bcc9e0`. `factory/nodeFactory.js` 281,751 bytes, SHA-256 `c6b1a7dc8c6676eaced02837d4e2a9b1499e62478b58549a476c1ab862de13cf`. |
| Public `built/local/typescript.d.ts` from upstream `hereby local` | Both 593,619 bytes and byte-identical, SHA-256 `65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd`. |
| Lane manifest | `generate_shards.py --check` passes. The runtime inputs and sanctioned API are identical. |

## Read-only findings for the remaining checker sites

Comparing AST-extracted owner text with the pinned upstream source, with line
endings normalized, confirmed that only `getTypeArguments` among the remaining
checker owners has an existing textual change. The exact dependency hash for
`getTypeFromClassOrInterfaceReference` still makes it an owned proof input.

### 16636:13 — createCanonicalSignature

`map(signature.typeParameters, tp => tp.target && !getConstraintOfTypeParameter(tp.target) ? tp.target : tp)` is passed to `getSignatureInstantiation`'s `typeArguments: Type[] | undefined`. `Signature.typeParameters` is a readonly array of TypeParameter. For a defined input, core's `map` creates a fresh array and pushes each callback result. The callback yields either the original element or its TypeParameter target. The array cannot alias the input storage.

The parameter's sole use is `fillMissingTypeArguments`. That function reads length, copies a supplied array via `slice()` before filling indices and setting length, and also returns a copy on the invalid-arity path when arguments are supplied. `getSignatureInstantiationWithoutFillingInTypeArguments` reads its returned copy for the cache key and passes it to the signature type mapper as readonly targets. The mapper retains or indexes read-only targets. Neither path writes to the mapped input array.

An honest sink is `typeArguments: readonly Type[] | undefined`. Alternatively the allocating call can declare `map<TypeParameter, Type>`, widening each element into a fresh `Type[]`. The ownership search found no existing whole-owner pin or direct edit for `createCanonicalSignature` or `getSignatureInstantiation`.

### 16975:70 — getTypeArguments

`concatenate(type.target.outerTypeParameters, getEffectiveTypeArguments(node, type.target.localTypeParameters!))` feeds `const typeArguments: readonly Type[] | undefined`, then optionally `TypeReference.resolvedTypeArguments?: readonly Type[] | undefined`. The first operand is `TypeParameter[] | undefined`; the second admits ordinary `Type` values. The mutable mixed-array overload is available and exposes the first operand as writable even though this binding and cache consume a readonly view.

The concatenate implementation can return either input. At this TypeReference-node arm, the first operand would be returned only if the second were empty or undefined. The closed-tree construction path creates a deferred TypeReferenceNode reference in `getTypeFromClassOrInterfaceReference` inside the `typeParameters` truthy guard. The class/interface producer creates a defined local list only after appending a real parameter. The other direct deferred-reference producer supplies array or tuple syntax; the instantiation call recreates a deferred reference with its existing target and node. For the TypeReference-node path, TypeScript arity failure previously returned `errorType`, and JavaScript filling permits arity recovery. Thus filling produces a nonempty second operand on those construction paths. The separate circular-resolution concatenations in this owner have their own fallback behavior.

On successful resolution the array enters `resolvedTypeArguments` directly or after `instantiateTypes`. That helper reads the input, returns it when unchanged, and allocates a new list before pushing mapped replacements. The cache has whole-property assignments and can be shared by `cloneTypeReference`. Its consumers read it or use the readonly return from `getTypeArguments` for indexing, iteration, comparisons, mapping, constructing references, and slices. Tracing maps IDs. I found no push, splice, index assignment, or mutable handoff of the cached array.

Keep the existing readonly binding and cache and select a readonly concatenate overload with a readonly first-operand view, for example by a typed local `readonly Type[] | undefined`. Instantiating the concatenate at `TypeParameter` would misdescribe the second operand. Amendments need to account for `99-correctness-axolotl-r`'s exact owner pin, `99-readonly-inferred-bindings`' later binding annotation, `95-checker-axolotl-b`'s earlier dependency hash, and `71-writable-views`' signature edit.

### 17031:47 — getTypeFromClassOrInterfaceReference

`concatenate(type.outerTypeParameters, fillMissingTypeArguments(typeArgumentsFromTypeReferenceNode(node), typeParameters, minTypeArgumentCount, isJs))` initializes the unannotated `const typeArguments`. The selected mutable overload infers `Type[] | undefined`; the next sink is `createTypeReference`'s declared `readonly Type[] | undefined` parameter.

The helper could return the outer parameter array if its second input were empty or undefined. This site is inside the nonempty local-parameter path described above. Invalid TypeScript arities return early and JavaScript supplies the permissive fill mode. For those paths, the fill result has the nonzero local parameter count. Concatenate therefore returns the filled array or a fresh combined array. Filling writes only to a new array or a slice. `createTypeReference` reads the array for IDs and propagation flags and stores it in the same readonly `resolvedTypeArguments` cache traced above. There is no element writer in that chain.

A readonly first-operand view selects the readonly overload, and an explicit `readonly Type[] | undefined` binding also documents the result. Annotating the result alone leaves overload selection at the call unchanged. `95-checker-axolotl-b` hashes the full owner as a dependency, so an amendment would also have to account for that pin.

### 17098:38 — getTypeFromTypeAliasReference

`aliasTypeArguments = getTypeArgumentsForAliasSymbol(newAliasSymbol)` writes to the local `let aliasTypeArguments: Type[] | undefined`. The helper delegates to `getLocalTypeParametersOfClassOrInterfaceOrTypeAlias`, which builds its own accumulator from undefined using `appendTypeParameters`. The result is the actual TypeParameter array created by that lookup. This path has no concatenate. The alternate branch can assign a `Type[]` built from syntax, an empty array, or undefined.

The local is passed to the readonly alias-argument parameter of `getTypeAliasInstantiation`. The latter reads it for `getAliasId` and passes it through `instantiateTypeWithAlias`. Its alias propagation paths store or forward readonly arrays on `Type.aliasTypeArguments` and readonly helper parameters. Among the consumers, node serialization maps it to type nodes, comparison reads index/length or uses `fillMissingTypeArguments` which mutates its copy, tracing maps IDs, and go-to-definition reads the first element. I found no element writes or mutable handoff of this array.

`let aliasTypeArguments: readonly Type[] | undefined` describes both producers and the read-only consumers. The ownership search found no existing whole-owner pin or direct edit in `getTypeFromTypeAliasReference`.

### 19979:17 — getTypeFromConditionalTypeNode

The shorthand `aliasTypeArguments` writes the inferred `TypeParameter[] | undefined` from `getTypeArgumentsForAliasSymbol(aliasSymbol)` into `ConditionalRoot.aliasTypeArguments?: Type[] | undefined`. It is the same accumulator produced by the lookup. This path has no concatenate. The local's other use tests whether alias arguments are present when choosing outer type parameters.

The sole read of `root.aliasTypeArguments` is `instantiateTypes(root.aliasTypeArguments, mapper)` in `getConditionalType` when no overriding alias symbol was supplied. It shares unchanged input or writes to a newly allocated list as traced above. The result is stored in the readonly `Type.aliasTypeArguments` field and follows the alias readers described above. I found no index write, push, splice, or mutable handoff. Writes to `root.instantiations` concern the separate cache map.

The field can honestly be `aliasTypeArguments?: readonly Type[] | undefined`. `ConditionalRoot` is exported, so this requires a public API comparison and a ruling on changing the array view. The ownership search found no existing whole-owner pin or direct edit in `getTypeFromConditionalTypeNode`. The apply snapshots show that adaptation 20 (`20-optional-declarations`) adds the explicit `| undefined` to the field. Adaptation 47 records that declaration in its evidence. A field amendment must preserve the optional contract from adaptation 20.

## First refusal stream

The original fix head `71f6d8b1` scanned the baseline at `src/tsc/tsc.ts` with the instrumented CommonJSBundle probe. It reported **3,524 refusal-scan events**, **2,005 distinct refusal triples**, and 1 first pipeline observation, for **3,525 total observations**. Its boundary was `production abort at lowering.declareModule`.

Raw stream `scan-before.json.gz`. Uncompressed SHA-256 `1846ae866b564166d36ad19b4002a8408142515ba4871188d102b570c69cedbb`.

## Fix-only paired scan and checker mutant

The original fix binary, used unchanged for both source variants, produced 3,525 total observations before, 3,521 after, and 3,522 for the restoring mutant. Each total includes one unchanged `lowering.declareModule` observation. The distinct counts including that observation were 2,006, 2,002, and 2,003. Raw streams are `scan-fix-only-after.json.gz` and `scan-fix-only-mutant.json.gz`.

Compared by multiset of (kind, rule, relative site, words), the after tree removes the three requested array refusals. It also removes the existing factory optional-field refusal at `7490:45` (TextRange lacks SourceMapRange's optional `source`) and the checker optional-field refusal at `19948:30` (Type lacks TypeParameter's optional `constraint`). It adds `castRefusal` at `checker.ts:19948:60` with these words:

> an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type

Restoring just the checker append returns both prior refusals at `19948:30` and `19948:37`, and removes that cast refusal. The factory delta is retained. Within the array-widening rule, exactly the requested checker refusal returns. Across all rules this fails the requested no-new-refusal and exact single-refusal mutant conditions. The guarded helper `getDeclaredTypeOfTypeParameter(symbol)` already returns TypeParameter. Calling it directly would change the emitted JavaScript callee. That alternative needs a ruling if byte identity remains required.

## Exact merged-head scan and whole applies

The measured merge is `9c4cd905651e8175d8c8433baae69bd4e769af3b`, with parents `71f6d8b1` and `ce20cd2be`. Its Stage 3 diff against the fix has the landing patch ID. The two full applies on the landing commit had the same tracked Stage 3 inputs as this merge. The manually staged scan tree’s two edited files matched the independent full apply byte for byte. The probe built from the exact merge ran all three source variants. Each probe process exited 0 and reached the same first pipeline boundary, `lowering.declareModule`. For each variant the normalized event multiset matches the fix-only probe, including the checked-cast blocker.

| Tree | Refusal-scan events | Total observations | Distinct observations | Raw stream SHA-256 (uncompressed) |
| --- | ---: | ---: | ---: | --- |
| before | 3,524 | 3,525 | 2,006 | `1846ae866b564166d36ad19b4002a8408142515ba4871188d102b570c69cedbb` |
| after | 3,520 | 3,521 | 2,002 | `e3c73ef002b771fbac0ecb419bef81e58e5af38fcdae7dd7aacf0ab57d48201b` |
| mutant | 3,521 | 3,522 | 2,003 | `5a54c0f535e49f56e470465bd472f6aec8e040b38cc93e9f5aced6d971704157` |

The full streams are `scan-merged-before.json.gz`, `scan-merged-after.json.gz`, and `scan-merged-mutant.json.gz`. The comparison retains duplicate events and normalizes only the source root in observation locations.

Two independent `stage3/apply.sh --build-product` runs with fresh output directories and separate caches both exited 0. All 81,506 working-tree file paths matched by content, mode, or link target as applicable (inventory SHA-256 `53edb9b7ea479936013d935d3e8801c06e9b1a685d6d4846fad9073f46067a91`). The Git refs matched separately. Git administrative metadata such as the index records build-specific timestamps and was excluded from the file comparison.

## Adapted-tree product

`go test -v -timeout 30m ./stage3/lane/adaptedtree -run '^TestProduct_Stage3AdaptedTree$'` passed on the landing commit with the locked package installs, pinned upstream checkout, and Go 1.27.1. It independently built the upstream compiler and test harness from the adapted tree. The whole lane uses that product from the local cache.

## Whole Stage 3 lane

All **430** listed tests passed, including all **398** manifest shards and the byte-comparisons shard. The initial `go test -v -parallel 2 -timeout 60m ./stage3/lane` recorded 224 passing tests before the execution environment restarted. I enumerated the package with `go test ./stage3/lane -list '^Test'` and resumed its 206 remaining tests with an exact name filter and `-parallel 3` against the same commit and cached adapted-tree product. The resumed invocation exited 0. The two pass sets are disjoint and cover every listed test. The lane runtime hash and sanctioned public API both stayed current. See `lane-first.log.gz` and `lane-resumed.log.gz`.

The landing branch contains the two requested edits and passes the stock compiler, byte, reproducibility, product and lane checks. The new cast refusal at the checker site blocks the refusal proof. The other five checker coordinates remain read-only findings for a ruling.

## Split Change A — token source-map ranges

Both split commits are based on the fetched main tip `f30d8927b`. Its Stage 3 files and cohere pin are identical to the previous base. The factory-only candidate is `af557aaa4`. The helper still has one caller, and both arguments come from the `EmitNode.tokenSourceMapRanges` slot. The same type now flows through the parameters, indexed assignment, and return. A type-only specifier joins the existing named compiler namespace import. Adaptation 30's two declined indexed accesses remain unchanged. The direct replay reports one file on first run and zero on replay. Restoring only the signature leaves a partial input which the adapter rejects before writing.

The stock 6.0.3 compiler-project JavaScript emit for `factory/nodeFactory.js` was byte-identical (SHA-256 `c6b1a7dc8c6676eaced02837d4e2a9b1499e62478b58549a476c1ab862de13cf`). `tsc -b src --force --pretty false` exited 0 with no diagnostics. The public `typescript.d.ts` from `hereby local` was byte-identical (SHA-256 `65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd`). These comparisons use the earlier baseline outputs from the identical before source.

## Split Change B — infer type parameter append

In the pinned checker, `getDeclaredTypeOfSymbol` returns `tryGetDeclaredTypeOfSymbol(symbol) || errorType`. The dispatcher first tests `Class | Interface` and calls `getDeclaredTypeOfClassOrInterface`. Then it tests `TypeAlias` and calls `getDeclaredTypeOfTypeAlias`. These are the only earlier flag checks. `TypeParameter` is next and calls `getDeclaredTypeOfTypeParameter`, which returns the cached object or creates and caches it. The later checks, in order, are `Enum`, `EnumMember`, and `Alias`, followed by an undefined return. Later flags cannot preempt the TypeParameter branch. The binder uses `TypeParameterExcludes = Type & ~TypeParameter`, and ordinary merging checks the exclusion mask. The checker also has an assignment merge exception. The dispatch order therefore identifies the flag overlap that needs to be measured instead of treating a bit test as an exact flag value.

The Node witness extracts the two actual `getInferTypeParameters` implementations, the pinned dispatcher and specialized helper, and core's `append`. It provides a fresh symbol-links cache and an instrumented type allocator. Both versions return undefined with absent locals or only nonparameter locals. For TypeParameter and TypeParameter|Transient, both versions return the same two object identities in order. The next call reuses those objects with no further allocator calls. See `infer-parameter-witness.cjs` and its output `infer-parameter-witness.json`.

### Change B whole-oracle count

The checker-only candidate is `3b721657e`. In a private copy of the full baseline, I captured `symbol.flags` before the calls, evaluated the original dispatcher and direct helper in that order, recorded `original === replacement`, and appended the original value. The recorder logs each call as `[fullFlags, sameObject, control]`. The preload installs the same recorder in Node and VM contexts.

The unfiltered oracle used all runners and eight workers. It recorded **9,810 calls**, all with full flags **262144** (`SymbolFlags.TypeParameter`), and **zero identity differences**. No observed symbol carried an earlier dispatch flag. One planted equal pair and one planted unequal pair used the same recorder and were each captured once. The controls are outside the 9,810 calls. The Node witness also tested cold allocation independently on each implementation, so the count's original-first order does not stand in for a cold-start witness.

The oracle completed with 106,366 passing, the one sanctioned public API comparison failure, and zero pending. The lane's count and failed-title parsing matched the platform expectations. Its API checker independently verified the recorded diff against the actual reference and local snapshots and the exact sanctioned changes. The split count changes no public API or lane file. See `infer-parameter-counts.json.gz` for the per-process raw records and `infer-parameter-oracle-verification.json` for the verified suite and API summary.

## Readonly type argument views (#dt55mtb)

### Public declaration diff

Upstream TypeScript 6.0.3 `hereby local` produced the following complete diff.

```diff
--- a/built/local/typescript.d.ts
+++ b/built/local/typescript.d.ts
@@ -6825,7 +6825,7 @@
         outerTypeParameters?: TypeParameter[] | undefined;
         instantiations?: Map<string, Type> | undefined;
         aliasSymbol?: Symbol | undefined;
-        aliasTypeArguments?: Type[] | undefined;
+        aliasTypeArguments?: readonly Type[] | undefined;
     }
     interface ConditionalType extends InstantiableType {
         root: ConditionalRoot;
```
