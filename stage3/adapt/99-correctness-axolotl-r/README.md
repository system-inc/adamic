# 99r: SortedArray mints through unknown

This revision is one commit on 9f1cd6e37ad89f2db0e10fba4bdb51a026363193,
which stopped the original single-cast proposal on stock TS2352. The underlying
stack is origin/typescript/core-typeonly-adapt at
b729c7e07010fe6db34c6afc1d11c8ca63a86ae3.

Adamic's subsequent #b5w3ycg ruling specifies `x as unknown as Brand`.
This adapter changes only the four core.ts mints to that spelling. It does not
change the checker callers, brand declarations, concatenate, or any compiler
lowering code. The source change is type-only. The old direct-cast rejection and
fail-closed adapter evidence are archived under initial-direct-cast/.

The source, type, replay, JavaScript, declaration, shard and census proofs are
complete. The lane Go proof remains failed: both API-generator wrappers exceed
their unchanged 55-second limits even in isolation. The underlying 22 Python
cases pass in 111.715 seconds. PROOF.md retains all attempts and this proof gap.

## Mints and own operand types

Stock TypeScript 6.0.3 checks each innermost operand in the actual applied program.
The audit strips only parentheses, requires an outer `as SortedArray<T>` or
`as SortedReadonlyArray<T>` whose operand is exactly an `as unknown` expression,
and requires the innermost operand's own type to be an array, never any or unknown.
For toSorted it also requires that operand to be the complete conditional.

| core.ts site and owner | Before | After | Innermost operand's own type |
| --- | --- | --- | --- |
| 736, deduplicateSorted | `emptyArray as any as SortedReadonlyArray<T>` | `emptyArray as unknown as SortedReadonlyArray<T>` | `never[]` |
| 759, deduplicateSorted | `deduplicated as any as SortedReadonlyArray<T>` | `deduplicated as unknown as SortedReadonlyArray<T>` | `T[]` |
| 764, createSortedArray | `[] as any as SortedArray<T>` | `[] as unknown as SortedArray<T>` | `never[]` |
| 1039, toSorted | `(array.length === 0 ? emptyArray : array.slice().sort(comparer)) as readonly T[] as SortedReadonlyArray<T>` | `(array.length === 0 ? emptyArray : array.slice().sort(comparer)) as unknown as SortedReadonlyArray<T>` | `T[]` for the whole conditional |

The adapted corePublic.ts:17 and :21 targets each add exactly one required member,
`" __sortedArrayBrand": undefined`, to ReadonlyArray<T> and Array<T>, respectively.
The key includes its leading space. Neither adds any other member.

audit-result.json retains all own types, flags, exact expressions and stock options.
The full-stock before/after diagnostic lists are identical. They include 59
unrelated ambient errors because this stock host does not reproduce Adamic's
embedded Node declarations. No diagnostic is suppressed. A clean witness formed
from the exact applied brand and emptyArray declarations, and all four exact
mint expressions, separately has zero stock diagnostics. Both complete 79-root
Adamic types checks use the real embedded host and are zero.

## Whole-src cast inventory

The complete initial whole-src AST inventory contains 17 brand casts: these four
core mints and the following 13 casts. Only the four core mints change. The source
comparison checks all 746 files and refuses any change outside the pinned owners.

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

## Owner contract

owners.json pins the complete exact normalized text and SHA-256 of the three
before/after owners, not a whole-file hash. Missing, duplicate, nested, moved,
changed or mixed owners refuse before any write. Both ways of half-adapting the
two-mint deduplicateSorted owner are explicitly refused. All files are reread
before writing. Foreign selectIndex edits and LF/CRLF are preserved.

The contract has 24 refused fixtures, one accepted foreign-edit fixture, both
line endings, and replay with zero writes plus identical bytes and mtime.
Whole core.ts JavaScript is required to match with comments kept and removed.
The directory's real Python sort remains directly after 99-correctness-axolotl-q.

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


## Proofs and next stops

See PROOF.md for commands, results and provenance, CENSUS.md for the ordinary
79-root next-stop table, and site-results.json for each mint's exact reason.
The ordinary census uses the unchanged production compiler. A separate existing
measurement-only overlay scans the complete core.ts to see beyond the first stop;
it has output disabled and is never used for the ordinary next-stop census.
This change does not claim that those four mints lower successfully before
Diagnostics' corresponding compiler change lands.

## Exact current lowering reasons

| Site | Kind | Exact reason |
| --- | --- | --- |
| core.ts:736:36 | Refused | a cast the runtime can't check |
| core.ts:759:12 | Refused | a cast the runtime can't check |
| core.ts:764:12 | Refused | a cast the runtime can't check |
| core.ts:1039:12 | Refused | a cast the runtime can't check |

The unchanged castProof in internal/lower/cast_proof.go explicitly refuses an
outer cast whose parenthesis-stripped operand is an as expression typed unknown.
The array operand proof does not bypass that current guard. Diagnostics' matching
brand-construction recognition is not included in this source-only commit.

The complete core scan also retains the pre-existing core.ts:1039:34 refusal:
`a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold`.
It concerns the emptyArray branch's mutable array relation; the conditional's own type is T[]. It is present before
and after; this adaptation does not claim to fix it. Both full-core scans have
82 distinct (kind, location, exact reason) findings, with none added or removed.
The measurement stream's historical label says "measured on a checker-rejected
program"; its actual metadata has checker_rejected=false and no diagnostics in
both runs. It remains a no-output measurement, not a successful-lowering claim.
