# sameMap scout: the constrained signature is not a drop-in replacement

Measured 2026-11-05. Recommendation: **do not adapt sameMap this way without separately migrating builder.ts:564**. Of 32 calls in the tsc build closure, 31 satisfy `T` assignable to `U`; one does not. Both proposed parameter orders are legal TypeScript, but both add two checker errors at that one call. The cast-free slice variant preserves the original emitted JavaScript exactly and encounters the same caller failure.

This commit contains only this report. All candidate source edits were confined to scratch adapted trees. No adaptation, compiler implementation, or public declaration is changed in the commit. No public declaration bundle build or native execution is claimed.

## Pins and method

| Input | Pin |
| --- | --- |
| origin/main at clone | `5c9fc8052ba0eb07c1c2a56e58b4e18cbd582602` |
| Requested S2a | `d231af0d34a4a145b14139dbc76d74f87e478b98` |
| cohere, identical in both | `06af2e399c6c1fb93fd570b774027942747400b4` |
| Nested TypeScript checker | `d92d9bfee114c80be2c375d72edae966176e3a4f` |
| Adapted upstream TypeScript 6.0.3 | `050880ce59e30b356b686bd3144efe24f875ebc8` |
| Stock checker and JS emitter | TypeScript 6.0.3 |
| Node | v24.19.0 |
| Go | go1.27.1 linux/amd64 |
| Adapted product input key | `d3d7119e05963133d152a90e3d34fcea1975bc1bdcf4a28bc35eb68cb382fddf` |

S2a is exactly two commits ahead of the measured main pin. The remote branch has advanced beyond the requested SHA; measurements use the requested `d231af0d`, not its later tip. Its stage3 adaptation inputs are unchanged from main, so both compiler builds measure identical adapted inputs.

Cloned system-inc/adamic from origin. Ran the requested `export GOPROXY='https://proxy.golang.org|direct'` and `timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1`. After explicit initialization of the pinned nested submodule, the successful setup completed in 213 seconds. Built the adapted baseline with `stage3/apply.sh`, supplying the pinned upstream checkout and lockfile-installed API packages. Read all four sameMap overloads and its complete implementation at adapted core.ts:342-368.

The stock checker program contains all 79 compiler roots plus the tsc entry and declarations. Its implementation closure is 81 files: the 79 compiler files plus `src/tsc/tsc.ts` and `src/tsc/_namespaces/ts.ts`. Every call was resolved with `getResolvedSignature`. `T` was read from the instantiated array parameter's number index type, `U` from the instantiated callback signature's return type, and the relation tested with `checker.isTypeAssignableTo(T, U)`. Explicit arguments were read from the call AST. A source search found no additional aliased or value-only sameMap uses in this closure. Four services calls outside this closure are not part of the tsc build and are excluded.

Two stock-checker configurations were measured:

* Adamic's `internal/load/load.go` options: strict, exact optional properties, unchecked indexed access, ES2024, ESNext/Bundler, forced modules, no emit. Loaded the repository's prelude with the loader's Node console declaration replacement and its pinned `@types/node` 25.3.3 plus undici dependency. Stock standard library declarations were used. Baseline: **0 diagnostics**.
* The compiler project's parsed `src/compiler/tsconfig.json`, with no emit, the same prelude, and pinned Node type roots. Baseline: **2 diagnostics**, both existing TS2322 generator/SetIterator incompatibilities at core.ts:1641:37 and core.ts:1746:21. Diagnostic differences were matched by file, code, and message, allowing later core.ts line numbers to shift with the added loop.

Adamic checks and censuses use the actual loader and its actual library overlays, independently of these stock checks.

## Call-site counts

| Class | Calls |
| --- | ---: |
| Total | 32 |
| Equal displayed T and U | 28 |
| Different T and U, T assignable to U | 3 |
| T not assignable to U | 1 |
| Inferred type arguments | 31 |
| Written explicit arguments | 1 |
| Stored result, including conditional assignments | 13 |
| Immediate returned result | 6 |
| Passed as an argument or nested expression | 13 |
| Direct identity comparison with input, overlaps stored-result class | 6 |
| Of those, `!==` | 4 |
| Of those, `===` | 2 |

Selected overloads: 14 mutable required, 15 readonly required, 2 mutable optional, 1 readonly optional. Thus **16 calls select readonly overloads**. The sole explicit call is `sameMap<Signature>`; both proposed orders preserve that call because its two types are equal. No call writes two explicit arguments.

## Every call in the build closure

Sites are baseline adapted-tree positions. Paths below are relative to `src/compiler/`. `infer` means no written type arguments. `O` abbreviates the checker's expanded `ObjectLiteralElementLike` union: `MethodDeclaration | GetAccessorDeclaration | SetAccessorDeclaration | PropertyAssignment | ShorthandPropertyAssignment | SpreadAssignment`.

| Site | T | U | T assignable to U | Arguments | Result use |
| --- | --- | --- | --- | --- | --- |
| `builder.ts:526:12` | `Diagnostic` | `Diagnostic` | yes | infer | return |
| `builder.ts:564:12` | `T` | `DiagnosticMessageChain` | **no** | infer | return; breaks A, B, C |
| `checker.ts:13195:16` | `Signature` | `Signature` | yes | `Signature` | return |
| `checker.ts:13313:30` | `TypeParameter` | `Type` | yes | infer | store |
| `checker.ts:14062:27` | `Type` | `Type` | yes | infer | store; `types !== (type as IntersectionType).types` |
| `checker.ts:14785:40` | `Type` | `Type` | yes | infer | argument/expression |
| `checker.ts:15826:30` | `Type` | `Type` | yes | infer | store; `reducedTypes === unionType.types` |
| `checker.ts:16593:16` | `TypeParameter` | `Type` | yes | infer | return |
| `checker.ts:18002:58` | `Type` | `Type` | yes | infer | argument/expression |
| `checker.ts:20246:85` | `IndexInfo` | `IndexInfo` | yes | infer | argument/expression |
| `checker.ts:22344:37` | `Type` | `Type` | yes | infer | store; `normalizedTypes !== type.types` |
| `checker.ts:22365:36` | `Type` | `Type` | yes | infer | store; `elements !== normalizedElements` |
| `checker.ts:23103:37` | `Type` | `Type` | yes | infer | store; `constraints !== (source as IntersectionType).types` |
| `checker.ts:25489:49` | `Type` | `Type` | yes | infer | store; `primaryTypes === types` |
| `checker.ts:26012:90` | `IndexInfo` | `IndexInfo` | yes | infer | argument/expression |
| `checker.ts:26035:38` | `Type` | `Type` | yes | infer | store |
| `checker.ts:26042:46` | `Type` | `Type` | yes | infer | argument/expression |
| `checker.ts:26045:59` | `Type` | `Type` | yes | infer | argument/expression |
| `checker.ts:26493:17` | `ElementFlags` | `ElementFlags` | yes | infer | store |
| `checker.ts:27597:54` | `Type` | `Type` | yes | infer | store |
| `checker.ts:27598:33` | `Type` | `Type` | yes | infer | store |
| `checker.ts:29463:62` | `Type` | `Type` | yes | infer | argument/expression |
| `checker.ts:33401:30` | `Type` | `Type` | yes | infer | argument/expression |
| `checker.ts:38246:48` | `Type` | `Type` | yes | infer | argument/expression |
| `checker.ts:38254:20` | `Signature` | `Signature` | yes | infer | return |
| `checker.ts:7432:37` | `Type` | `Type` | yes | infer | store |
| `factory/parenthesizerRules.ts:423:24` | `Expression` | `Expression` | yes | infer | store |
| `factory/parenthesizerRules.ts:516:40` | `TypeNode` | `TypeNode` | yes | infer | argument/expression |
| `factory/parenthesizerRules.ts:534:40` | `TypeNode` | `TypeNode` | yes | infer | argument/expression |
| `factory/parenthesizerRules.ts:614:40` | `TypeNode \| NamedTupleMember` | `TypeNode` | yes | infer | argument/expression |
| `factory/parenthesizerRules.ts:669:44` | `TypeNode` | `TypeNode` | yes | infer | argument/expression |
| `transformers/jsx.ts:451:20` | `O` | `O` | yes | infer | return |

The three compatible unequal pairs are parenthesizerRules.ts:614 (`TypeNode | NamedTupleMember` to `TypeNode`) and checker.ts:13313 and :16593 (`TypeParameter` to `Type`). All retain their original inferred pairs under A, B, and C. The failed call has no valid candidate instantiation; checker error-recovery types are not a successful inferred solution.

The six direct comparisons are listed in the table. Returned results can also carry identity to their callers. In particular, the builder helper's returned array is compared with `chain.next` in `convertOrRepopulateDiagnosticMessageChain` at builder.ts:555. Its caller at :530 compares the returned chain with `diag.messageText`. Replacing sameMap with unconditional map would require a separate identity review.

## The one incompatible caller

At builder.ts:558-564:

```ts
function convertOrRepopulateDiagnosticMessageChainArray<
    T extends DiagnosticMessageChain | ReusableDiagnosticMessageChain
>(
    array: T[] | undefined,
    sourceFile: SourceFile | undefined,
    newProgram: Program,
    repopulateInfo: (chain: T) => RepopulateDiagnosticChainInfo | undefined,
): DiagnosticMessageChain[] | undefined {
    return sameMap(array, chain =>
        convertOrRepopulateDiagnosticMessageChain(chain, sourceFile, newProgram, repopulateInfo));
}
```

Original inferred sameMap types are the enclosing `T` and `DiagnosticMessageChain`. The enclosing constraint permits `ReusableRepopulateInfoChain`, which lacks `messageText`, `category`, and `code`. Serialized chains can also contain recursively reusable `next` entries. Therefore the input element type is not a subtype of the output element type. The callback converts the input; equality of each actual returned element would justify the runtime identity path, but a bound on the entire input type cannot express that conditional fact.

This is the only existing caller relying on U not being a supertype of T. There is no additional concrete disjoint-type caller in the measured closure.

## Scratch candidates and checker results

Every candidate changes all four overloads and the implementation signature. Required/optional and mutable/readonly distinctions remain intact. The existing `item as unknown !== mapped` comparison is retained; it is an upcast and is not one of the two removed double casts.

| Candidate | Type parameters | Prefix construction | Stock with Adamic options | Stock project config | adamic types, 79 roots |
| --- | --- | --- | --- | --- | --- |
| Baseline | `<T, U = T>` | double-cast slice | 0 diagnostics | 2 existing diagnostics | exit 0 |
| A | `<T extends U, U = T>` | push loop | 2 new diagnostics | 4 total, 2 new | exit 1, 2 diagnostics |
| B | `<U, T extends U = U>` | push loop | 2 new diagnostics | 4 total, 2 new | exit 1, 2 diagnostics |
| C | `<T extends U, U = T>` | directly typed slice | 2 new diagnostics | 4 total, 2 new | exit 1, 2 diagnostics |

A is legal; it does not have a circular type-parameter error. Swapping order does not solve the actual caller incompatibility. None of A, B, or C keeps whole-program checking green.

A and B replace the prefix with:

```ts
const result: U[] = [];
for (let j = 0; j < i; j++) {
    result.push(array[j]!);
}
```

The new `!` is necessary under `noUncheckedIndexedAccess`; without it, the fresh indexed read is `T | undefined` and cannot be pushed into arbitrary `U[]`. This follows the existing adapted function's assertions on `array[i]!`. It adds no cast through unknown. All candidates then push `mapped`, retain the original remaining-element loop, and end with `return array;`.

C instead uses:

```ts
const result: U[] = array.slice(0, i);
```

Because `T extends U`, the slice result `T[]` is already assignable to `U[]`; no push-loop rewrite is needed for that assignment.

### Every new stock diagnostic

These two locations and codes occur for **each of A, B, and C**, in both stock configurations and in the actual Adamic loader. The stock Adamic-option messages are:

**src/compiler/builder.ts:564:20, TS2769**

```text
No overload matches this call.
  The last overload gave the following error.
    Argument of type 'T[] | undefined' is not assignable to parameter of type 'readonly DiagnosticMessageChain[] | undefined'.
      Type 'T[]' is not assignable to type 'readonly DiagnosticMessageChain[]'.
        Type 'T' is not assignable to type 'DiagnosticMessageChain'.
          Type 'DiagnosticMessageChain | ReusableDiagnosticMessageChain' is not assignable to type 'DiagnosticMessageChain'.
            Type 'SerializedDiagnosticMessageChain' is not assignable to type 'DiagnosticMessageChain' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
              Types of property 'next' are incompatible.
                Type 'ReusableDiagnosticMessageChain[] | undefined' is not assignable to type 'DiagnosticMessageChain[] | undefined'.
                  Type 'ReusableDiagnosticMessageChain[]' is not assignable to type 'DiagnosticMessageChain[]'.
                    Type 'ReusableDiagnosticMessageChain' is not assignable to type 'DiagnosticMessageChain'.
                      Type 'ReusableRepopulateInfoChain' is missing the following properties from type 'DiagnosticMessageChain': messageText, category, code
```

**src/compiler/builder.ts:564:78, TS2345**

```text
Argument of type 'DiagnosticMessageChain' is not assignable to parameter of type 'T'.
  'DiagnosticMessageChain' is assignable to the constraint of type 'T', but 'T' could be instantiated with a different subtype of constraint 'DiagnosticMessageChain | ReusableDiagnosticMessageChain'.
```

There are no other new errors, including no errors at the explicit `sameMap<Signature>` call. Under the project config, the two original core diagnostics move down three lines for A/B and remain at their baseline positions for C. The actual Adamic checker abbreviates the first error's explanatory chain by naming `ReusableRepopulateInfoChain` directly; its locations and codes are the same.

## Emitted JavaScript and caller-visible effects

Transpiled the complete isolated sameMap overload group and implementation with stock TypeScript 6.0.3, CommonJS, ES2024, comments removed. Compared emitted bytes and executed the emitted functions on Node.

* A and B emit identical JavaScript to each other. They return the exact input object on the no-change path, including empty arrays, and return undefined for undefined input. For ordinary dense arrays they copy the same current prefix elements in order, append the mapped first difference, then map the remainder in order. The prefix is copied after the first differing callback, as in the original. It is not mapped again.
* The dense/optional probe covered five inputs (`undefined`, `[]`, `[1]`, `[1,2,3]`, `[undefined,2,undefined]`) and four callback change positions each. All 60 candidate/baseline comparisons passed for identity, element values, keys, and callback arguments/order. The original strict equality comparison, including its NaN and signed-zero behavior, is unchanged.
* **Sparse arrays differ for A/B.** With a hole at index 0, values at 1 and 2, and the first change at 1, original/C result keys are `["1","2"]`; A/B keys are `["0","1","2"]`. Slice preserves the prefix hole, whereas the push loop reads and stores undefined. Holes in the remaining mapped suffix are already read and pushed by the original. If no element differs, all versions return the original sparse array.
* **Array subclasses differ for A/B.** A changed `Sub extends Array` input produces a `Sub` under baseline/C, but a plain `Array` under A/B. The push rewrite also removes slice's species/constructor behavior, any overridden slice call, and its HasProperty operations. Those can be observed by subclasses, accessors, proxies, or a modified array environment. Dense ordinary arrays alone do not establish equivalence for these inputs.
* **C emits byte-identical JavaScript to baseline.** Both casts erase, and neither the prefix operation nor the final identity return changes. Consequently its sparse-array, subclass, proxy, callback mutation, and species behavior is preserved by the same emitted function, not merely by the finite probes.
* The push-loop `!` erases in stock JavaScript. Adamic can turn non-null assertions into checks; full native equivalence of that added assertion is not established because the candidate program is checker-rejected. C introduces no such extra assertion.
* Keep readonly input and result overloads. Sixteen actual calls select readonly overloads; dropping readonly inputs rejects their arrays. Returning a mutable type for a readonly identity input would expose mutation of the original. Conversely, collapsing everything to readonly results would remove the mutable result contract used by the other sixteen calls. Keep the optional overloads as well: three calls select them.
* Changing parameter order alters the meaning of future or outside two-argument explicit instantiations. It causes no extra break among the measured calls, which contain only the one equal-type explicit instantiation. A preserves the original argument order and is preferable to B if the incompatible caller is separately resolved.

## Census first-stop list

Built the ordinary driver exactly as the meter's `build_census` function does:

```sh
source /workspace/adamic-tools/env.sh
go build -o /workspace/scratch/samemap/main-census ./stage3/census/tool
# Repeat in a detached d231af0d checkout for s2a-census.
```

Both checkouts use the same recorded cohere and nested checker pins. The detached S2a checkout shares the initialized, identical dependency checkout. No lowerer overlay or refusal suppression was used. The ordinary driver performs `load.Load` first and only invokes `lower.Lower` on checker-clean programs.

Used the authorized three-root fallback because the initial baseline roots take about 8 seconds per inspection and the driver also repeats its input as an aggregate row. A full 79-root census across eight input/compiler combinations would be much longer. Each file was passed directly to the unmodified meter census driver:

```sh
main-census /workspace/scratch/samemap/A/src/compiler/binder.ts A-main-binder.ts.jsonl
# Likewise checker.ts and utilities.ts; baseline, B, C; main-census and s2a-census.
```

Each invocation writes one per-file row and a duplicate one-root aggregate row. The following lists the per-file rows only. `checker` means lowering did not run.

| Input | Compiler | binder.ts | checker.ts | utilities.ts |
| --- | --- | --- | --- | --- |
| Baseline | main | NotYet core.ts:332:1 | NotYet core.ts:332:1 | NotYet core.ts:332:1 |
| Baseline | main plus S2a | Refused core.ts:356:37 | Refused core.ts:356:37 | Refused core.ts:356:37 |
| A | main | checker builder.ts:564:20 | checker builder.ts:564:20 | checker builder.ts:564:20 |
| A | main plus S2a | checker builder.ts:564:20 | checker builder.ts:564:20 | checker builder.ts:564:20 |
| B | main | checker builder.ts:564:20 | checker builder.ts:564:20 | checker builder.ts:564:20 |
| B | main plus S2a | checker builder.ts:564:20 | checker builder.ts:564:20 | checker builder.ts:564:20 |
| C | main | checker builder.ts:564:20 | checker builder.ts:564:20 | checker builder.ts:564:20 |
| C | main plus S2a | checker builder.ts:564:20 | checker builder.ts:564:20 | checker builder.ts:564:20 |

Top reasons, over these 24 per-file observations:

| Reason | Roots |
| --- | ---: |
| Checker TS2769 at builder.ts:564:20, followed by TS2345 at :564:78 | 18 |
| Refused: `a cast the runtime can't check` at core.ts:356:37 | 3 |
| NotYet: `a generator function (its counted frame is being built, #qk8rztp)` at core.ts:332:1 | 3 |

For each candidate separately, each compiler has three checker-stopped roots with those same two diagnostics. All 24 census command processes exit 0 because the driver successfully records refusals and checker failures; that is not an acceptance result. All four `adamic types` commands supply the complete sorted 79-root list: baseline exits 0; A/B/C exit 1.

Thus the constrained candidates do not demonstrate successful movement of the reported 73 roots to deeper lowering sites. On the three requested sample roots they replace a lowering blocker with an earlier checker failure. The other 76 roots were not censused individually, and the original 73/79 distribution is context supplied by the task, not a newly measured count here.

## Recommendation

Reject A and B as drop-in adaptations: one real caller fails, and the push rewrite has observable JavaScript differences. C is the better behavioral formulation if that caller is separately migrated: keep `<T extends U, U = T>`, keep all four overload shapes, assign the slice directly, and return the original array directly. Its JavaScript is exactly preserved, but it still fails the current caller set and cannot be landed as the requested identity-preserving cast removal today.

A migration for builder.ts must preserve its recursive conversion semantics and the `next === chain.next` identity test. An unchecked cast, a weakened overload promise, or an unconditional map should not be presented as proof that the input type extends the output type. This scout stops at that concrete incompatibility and does not implement a migration.

## Retained local evidence

Scratch artifacts are under `/workspace/scratch/samemap/`: `baseline-scan.json`, `A-scan.json`, `B-scan.json`, `C-scan.json`; the four `*-project.json` files; `runtime.json`; the four `*-types.log` files; `measurement-status.json`; and the 24 `*-{main,s2a}-*.ts.jsonl` census files with corresponding logs. Setup is `/workspace/scratch/setup.log`. Build/apply logs are also in that scratch directory. These local paths are not durable repository artifacts; all call sites, new errors, and first stops are reproduced above. Only REPORT.md is committed.

Binary SHA-256 values:

* `main-census`: `89c69f99a43b5d115bfe8947b53765777692752273ddbda88992fb4efe13e0fc`
* `s2a-census`: `b287c69758f61965d24612a3fff5fa9dd1b28e20b2a6d04628e6aecd84f6a771`
* `adamic`: `cacb4c4f01ccee90c3aee70e276cf7f9d37afb3e092822b006edd2339ed11f5c`
