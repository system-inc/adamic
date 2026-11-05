# 99r: blocked direct SortedArray mints and concatenate caller read

Base: origin/typescript/core-typeonly-adapt at
b729c7e07010fe6db34c6afc1d11c8ca63a86ae3.
This branch is evidence of a stopped proposal. It is not an accepted adaptation.
No adapted source, compiler, lowering code, brand declaration, concatenate
implementation, or checker caller is edited. Only this new directory is committed.
The adapter deliberately refuses before any write so a full apply cannot silently
claim that the blocked source change was accepted.

## Stock checker stop

The requested stop condition is reached at the first mint, core.ts:736.
Stock TypeScript 6.0.3 rejects `emptyArray as SortedReadonlyArray<T>` with TS2352.
The checker reports:

```text
Conversion of type 'never[]' to type 'SortedReadonlyArray<T>' may be a mistake because neither type sufficiently overlaps with the other. If this was intentional, convert the expression to 'unknown' first.
  Property '" __sortedArrayBrand"' is missing in type 'never[]' but required in type 'SortedReadonlyArray<T>'.
```

The exact diagnostic, all options, source hashes, brand declarations, and
complete cast inventory are retained in
probe-result.json. The trial exists only in the compiler host's memory. It changes
only the first mint, then stops. No unknown intermediate cast or declaration
change is substituted. The full stock run does not reproduce Adamic's embedded
Node/host overlays: its 59 ambient-context diagnostics are retained, and the
first mint adds exactly one diagnostic, TS2352. A separate stock witness extracts
the exact two brand declarations and emptyArray declaration from the applied tree;
it has zero diagnostics before and only TS2352 after the same expression change.
The 79-root Adamic baseline is independently checker-clean.

Reproduce against a fully applied base tree:

```sh
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-r/probe.cjs /workspace/r-before
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-r/read-callers.cjs /workspace/r-before
```

| Mint in core.ts | Requested direct form | Decision |
| --- | --- | --- |
| 736, deduplicateSorted | `emptyArray as SortedReadonlyArray<T>` | Stock TS2352; stop; unchanged |
| 759, deduplicateSorted | `deduplicated as SortedReadonlyArray<T>` | Not built after the first rejection; unchanged |
| 764, createSortedArray | `[] as SortedArray<T>` | Not built after the first rejection; unchanged |
| 1039, toSorted | `(array.length === 0 ? emptyArray : array.slice().sort(comparer)) as SortedReadonlyArray<T>` | Not built after the first rejection; unchanged |

Both adapted interfaces in corePublic.ts:17 and :21 add exactly one member to
ReadonlyArray<T> and Array<T>, respectively: the required property
`" __sortedArrayBrand": undefined`. The key includes a leading space. Neither
interface adds any other member. This erased brand ruling does not change stock
TypeScript's overlap check. The missing required property is the checker's stated
objection, even though the proposed assertion would erase from JavaScript.

The full adapted src cast inventory is in probe-result.json. In addition to the
four mints above:

| File:line | Cast source | Changed |
| --- | --- | --- |
| src/compiler/program.ts:3285 | `emptyArray as any as SortedReadonlyArray<Diagnostic>` | No, outside core.ts |
| src/compiler/utilities.ts:6061 | `[] as Diagnostic[] as SortedArray<Diagnostic>` | No, outside core.ts |
| src/compiler/utilities.ts:6062 | `[] as string[] as SortedArray<string>` | No, outside core.ts |
| src/compiler/utilities.ts:6099 | `[] as Diagnostic[] as SortedArray<DiagnosticWithLocation>` | No, outside core.ts |
| src/compiler/utilities.ts:6100 | `diagnostics as SortedArray<DiagnosticWithLocation>` | No, outside core.ts |
| src/compiler/utilities.ts:6108 | `nonFileDiagnostics.slice() as SortedArray<Diagnostic>` | No, outside core.ts |
| src/compiler/watchUtilities.ts:156 | `entries.files.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>` | No, outside core.ts |
| src/compiler/watchUtilities.ts:157 | `entries.directories.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>` | No, outside core.ts |
| src/server/project.ts:1778 | `emptyArray as SortedReadonlyArray<string>` | No, outside core.ts |
| src/server/utilitiesPublic.ts:140 | `[] as any as SortedArray<T>` | No, outside core.ts |
| src/services/services.ts:2957 | `[] as number[] as SortedArray<number>` | No, outside core.ts |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:386 | `cloned.entries.map(clone) as SortedArray<FSEntry>` | No, outside core.ts |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:976 | `[] as FSEntry[] as SortedArray<FSEntry>` | No, outside core.ts |

## Concatenate's six checker.ts diagnostics: read only

The complete five owning functions, relevant helper implementations, concatenate
overloads and implementation are captured with exact text and SHA-256 in
caller-owners.json. These are proposed caller fixes only. None is built or edited.
Lines below are diagnostic locations on the adapted stack, not necessarily the
line containing the originating concatenate call. In particular, 14120 and 14121
share the call at 14118, and 16980 is the final return after cache filling.

A defined second array alone does not prove a defined result: an empty second
array returns the first argument, which may be undefined. All six uses have an
array as the second argument. Both arguments cannot literally be absent at any
of them, but emptiness still matters. No fix belongs inside concatenate.

| Diagnostic and whole owner | Purpose and arguments | Absence and emptiness evidence | Smallest honest caller fix |
| --- | --- | --- | --- |
| 14120, TS18048, resolveTypeReferenceMembers | Instantiates declared members of a reference using source type parameters followed by its this parameter. `concatenate(source.typeParameters, [source.thisType!])` supplies the local `typeParameters`; the failing use compares its length with getTypeArguments(type).length. | source.typeParameters is optional, and adaptation 49 intentionally removed its old assertion. The second argument is a one-element literal, so is present and nonempty, even if its element were undefined. The result is always an array, but the truthful overload does not encode nonempty tuples. resolveDeclaredMembers and the Reference branch of resolveStructuredTypeMembers do not make source.typeParameters non-optional. | At this caller, pass `source.typeParameters ?? emptyArray` as the first argument. Both arrays are now statically present, selecting the already honest non-optional overload. The existing thisType assertion need not be added or changed. |
| 14121, TS2345, resolveTypeReferenceMembers | Same local result passed to resolveObjectTypeMembers as its non-optional readonly type-parameter array. The other arguments are type, source and paddedTypeArguments. The padding call uses getTypeArguments(type), a non-optional readonly Type[], and `[type]`. | Same optional source and nonempty singleton evidence as 14120. Both padding arguments are also defined. No separate unsafe padding-result proof is needed. | The same first-argument default at 14118 fixes both uses. No second edit, cast or new assertion. |
| 16980, TS2322, getTypeArguments | Lazily resolves and caches a reference's arguments, handles recursive resolution with errorType, then returns a readonly Type[]. Its failed-pop cache path concatenates optional outerTypeParameters with `localTypeParameters?.map(() => errorType) \|\| emptyArray`. Its TypeReference-node path concatenates optional outerTypeParameters with getEffectiveTypeArguments(...). | Both outer and local parameter properties explicitly include undefined. The failed-pop second argument is always an array but may be empty. getEffectiveTypeArguments returns Type[] and may also be empty. The initial failed-push early return already defaults the entire concatenation with `\|\| emptyArray`; the final cache paths lack that guarantee. The surrounding resolvedTypeArguments guard precedes these assignments and cannot narrow their new optional results. | At this function's final return, use `return type.resolvedTypeArguments ?? emptyArray;`. This checks the value at its own required-array use, handles both cache paths, and matches the existing failed-push empty-array policy. |
| 41574, TS2412, instantiateTypeWithSingleGenericCallSignature | During contextual generic function inference, adopts inferred parameters when the returned signature is non-generic and inference sets do not overlap. Concatenates optional context.inferredTypeParameters with uniqueTypeParameters. | The signature && signature.typeParameters guard supplies an array to getUniqueTypeParameters. That helper declares readonly TypeParameter[] and always returns its initialized result array. It can be empty; no length guard here proves otherwise. some(inferences, hasInferenceCandidates) concerns the context's inference list, not a proof that uniqueTypeParameters has elements. The optional context property excludes explicit undefined under exact optional properties. | Store the concatenate result in a local and assign context.inferredTypeParameters only under `if (combined !== undefined)`. If absent, leave the property as it was. This avoids introducing an empty array whose truthiness might affect later inference. |
| 41628, TS2345, getUniqueTypeParameters | Copies or renames signature type parameters to avoid collisions; sends `concatenate(context.inferredTypeParameters, result)` to getUniqueTypeParameterName. | result is a non-optional TypeParameter[] initialized to []. The call is inside `hasTypeParameterByName(context.inferredTypeParameters, name) \|\| hasTypeParameterByName(result, name)`. A true branch proves at least one list contains an entry, so the result is in fact present. The checker cannot derive that relational fact through the helper and disjunction. The sole caller supplies the guarded signature.typeParameters array. | Check at this argument's use: `getUniqueTypeParameterName(concatenate(context.inferredTypeParameters, result) ?? emptyArray, name)`. The fallback is unreachable under the branch's semantic invariant, and supplies the required readonly array without an assertion. |
| 49743, TS2322, getDiagnosticsWorker | Checks a file and includes newly deferred global diagnostics in its returned Diagnostic[]. Concatenates deferredGlobalDiagnostics with semanticDiagnostics after comparing the old and new global array references. | semanticDiagnostics comes from diagnostics.getDiagnostics(fileName), declared DiagnosticWithLocation[]; createDiagnosticCollection implements its missing-file case with []. Both global lists come from getGlobalDiagnostics, declared Diagnostic[], so neither is absent. relativeComplement conservatively returns `T[] \| undefined`, but its body with two defined array inputs returns an array on every path. The reference-inequality guard does not narrow that helper result for the checker. | Make the local declaration non-optional with `const deferredGlobalDiagnostics = relativeComplement(previousGlobalDiagnostics, currentGlobalDiagnostics, compareDiagnostics) ?? [];`. The existing two-array overload then returns Diagnostic[]. This checks the helper result at this caller and preserves its observed array behavior. |

The proofs above distinguish declarations the checker can use from facts visible
only by reading helper bodies. They do not claim that changing concatenate alone
would pass checking, or that an array's existence proves it is nonempty.

## Proof status

This proposal stops at stock TypeScript admission, before any mint is written.
Consequently there is no accepted 99r after-tree, owner rewrite contract, zero-edit
rewrite replay, emitted-JavaScript or public-declaration comparison, or fresh
79-root Adamic before/after census. Those successful-change proofs must not be
claimed for a checker-rejected proposal. The baseline full apply and stock checker
probe are separate from Adamic types and the compiler census.

| Check | Result |
| --- | --- |
| Prescribed setup with requested GOPROXY | Exit 0, 731.121 seconds |
| Base full apply with product packaging | Exit 0 |
| Baseline Adamic types over all 79 recursive roots | Exit 0, zero diagnostics; types-before-result.json |
| Exact-tree stock first-mint delta | One added TS2352; unrelated ambient diagnostics retained |
| Isolated stock witness from exact declarations | Zero diagnostics before; one TS2352 after |
| generate_shards --check with configured toolchain | Exit 0 |
| Real directory sort | 99r directly follows 99q |
| Blocked adapter refusal | Exit 1 before writes; all 746 source files have the same aggregate hash before and after |
| Accepted 99r replay, owner fixtures, JS/d.ts comparisons, lane Go tests, fresh census | Not run: stopped at the required checker gate |

The initial root preflight enumerated only 40 top-level files and refused its
79-root count assertion. It was corrected to recurse before the final checks.
The initial stock-context preflight expected zero diagnostics without installing
Adamic's host overlay, and refused that assertion. The final comparison retains
those ambient diagnostics, checks the exact one-diagnostic delta, and separately
requires the clean extracted witness. The first shard invocation lacked gofmt on
PATH; sourcing the setup environment resolved it. No checker diagnostics are
suppressed or relabeled as a clean full-stock build.

99q's retained CENSUS.md records the prior stack's 73 core.ts:736:36 refusals
(`a cast the runtime can't check`), two accepted roots, and four other NotYet
roots. That historical table is not relabeled as a fresh 99r measurement. No
single-cast lowering next-stop claim is possible because the source trial fails
stock checking first. The requested lowering work remains someone else's.

sort.json records the real Python directory sort, placing this proposal directly
after 99-correctness-axolotl-q. Reproduction requires a base-only checkout: this
blocked branch's adapt.cjs intentionally refuses rather than admitting a no-op.

## Historical 79-root next stops, not rerun

Source: ../99-correctness-axolotl-q/CENSUS.md and census-results.json, on its
recorded d40c2470e compiler. These rows describe the unchanged double-cast stack.
They are provided for context only, not as a single-cast 99r proof.

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | core.ts:736:36 | a cast the runtime can't check |
| 1 | NotYet | commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | debug.ts:243:59 | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 2 | accepted | corePublic.ts and hostErrors.ts | No stop |
