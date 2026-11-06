# 99r: SortedArray mints and truthful concatenate callers

Base: origin/typescript/core-typeonly-adapt at
b729c7e07010fe6db34c6afc1d11c8ca63a86ae3 (99q, on 99m and 99l).
Task #b5w3ycg, Adamic's ruling. Only this new adaptation directory is changed.
The blocked 9f1cd6e3 README, probe-result.json and caller-owners.json were read
first. Its deliberate refusal adapter was not copied. A fresh TypeScript v6.0.3
checkout at 050880ce59e30b356b686bd3144efe24f875ebc8 supplied both full applies.

## Exact edits

The four core.ts mints now spell `x as unknown as Brand<T>`: both returns in
deduplicateSorted, createSortedArray's empty literal, and toSorted's conditional.
The old any and readonly-array intermediate hops are removed only at these four
sites. The required `" __sortedArrayBrand": undefined` declarations are unchanged.
Stock TypeScript 6.0.3 parses the complete planned source. The unknown spelling
avoids the blocked direct-cast TS2352 without changing the brand declaration.

Only concatenate's two false mixed overload results change, to `T[] | undefined`
and `readonly T[] | undefined`. Its other overloads and complete implementation
are unchanged. concatenate-witness.cjs freshly confirms that an undefined first
array and empty second array return undefined, and checks both corrected inferred
results with stock TypeScript.

| Checker owner | Exact local change |
| --- | --- |
| resolveTypeReferenceMembers | First argument is `source.typeParameters ?? emptyArray` |
| getTypeArguments | Final return is `Debug.checkDefined(type.resolvedTypeArguments)` |
| instantiateTypeWithSingleGenericCallSignature | Assign `Debug.checkDefined(concatenate(context.inferredTypeParameters, uniqueTypeParameters))` |
| getUniqueTypeParameters | Argument is `Debug.checkDefined(concatenate(context.inferredTypeParameters, result))` |
| getDiagnosticsWorker | No edit; relativeComplement's truthful defined-array overload supplies Diagnostic[] |

There is no conditional assignment, no new cast, and no empty-array fallback at
sites 4, 5 or 6. The getUniqueTypeParameters `||` guard proves at least one input
contains a matching type parameter, so at least one array is nonempty and the
concatenation is defined. COUNT.md records the required whole-oracle counts before
any of those three edits was built. All three conditional counts were zero;
every counter ran, the real absence control fired 111,773 times, and each planted
undefined was recorded exactly once through the same recording path.

`javascriptChanging` is true. Edit 3 and the built checks in edits 4, 5 and 6
change JavaScript. Edits 1, 2 and 7 are type-only. Both complete core.ts transpiles,
with comments retained and removed, are byte-identical. Checks can throw if a
required invariant is violated; the count is evidence about this complete oracle,
not a proof that every possible client execution satisfies those invariants.

## Debug.checkDefined is unconditional

At the fully adapted head, debug.ts:227-237 contains:

```ts
export function assertIsDefined<T>(value: T, message?: string, stackCrawlMark?: AnyFunction): asserts value is NonNullable<T> {
    // eslint-disable-next-line no-restricted-syntax
    if (value === undefined || value === null) {
        fail(message, stackCrawlMark || assertIsDefined);
    }
}

export function checkDefined<T>(value: T | null | undefined, message?: string, stackCrawlMark?: AnyFunction): T { // eslint-disable-line no-restricted-syntax
    assertIsDefined(value, message, stackCrawlMark || checkDefined);
    return value;
}
```

The decisive lines are 229-230 and 235-236. Neither function consults the assertion
level or shouldAssertFunction. fail throws at line 182. The adapted
setAssertionLevel at 157-159 only stores the level. The extracted exact functions
were executed under Node at None (0), Normal (1), Aggressive (2) and VeryAggressive
(3): undefined throws `Debug Failure.` at every level, and a defined value returns
unchanged. See debug-proof.cjs and debug-proof.json. The six specialized node
assertions changed by 99-cli-debug-assertion-level are not used by these edits.

## resolveTypeReferenceMembers witness

With its second argument `[source.thisType!]`, concatenate returns that argument
unchanged when the first argument is absent or empty. Substituting emptyArray for
an absent first argument therefore preserves the result. resolve-witness.cjs
extracts the actual before and after owners and the unchanged concatenate body.
Node witnesses for absent, present-empty and present-nonempty typeParameters all
produce identical parameters and type arguments before and after. In the first
two cases the output parameters are `["this"]`; in the nonempty case they are
`["T", "this"]`. See resolve-witness.json.

## relativeComplement

The added first overload is exactly:

```ts
relativeComplement<T>(arrayA: T[], arrayB: T[], comparer: Comparer<T>): T[];
```

The original optional signature is retained as a callable overload, followed by
the unchanged implementation. With both arrays defined, the initial return
returns the defined arrayB (including either empty-input case). The remaining
path initializes `const result: T[] = []`, only pushes or continues in its loops,
and returns result. Thus every normal return is an array; comparer exceptions
do not add a normal undefined return.

The complete src/ search finds exactly one call, in getDiagnosticsWorker, plus
its import and the helper declaration. There are no other callers to edit. All
79 compiler roots and the full upstream build type-check with that caller
unchanged. The helper appears in neither freshly built public declaration file
before or after; both files are byte-identical. @internal documentation is kept
on both overloads, with the original comment on the implementation, preserving
JavaScript even when comments are retained.

An overload result narrower than its implementation is checked at the call unless
a per-overload body proof erases that check. The full proof contract is
[docs/overload-proof.md on compiler/overload-proof](https://github.com/system-inc/adamic/blob/7acd44960c497aac4b96e1a04e53f7b4e31b7c19/docs/overload-proof.md);
that complete proof is not on the measured main. Main's limited proof does not
prove this body, which contains non-null assertions, a local result and loops.
The result therefore needs the call-site presence check, with the overload result
Diagnostic[] rather than the optional implementation result. No erased body proof
or successful native compilation of the whole checker is claimed.

The read-only call-contract census checks the unchanged complete 79-root program
with Adamic's real checker, finds this owner and exact call text, verifies overload
1 is selected, and records the proof-disabling assertions. Its exact line is:

```json
{"bodyProof":"main censusReturnProof returns false at its assertion scan","call":"relativeComplement(previousGlobalDiagnostics, currentGlobalDiagnostics, compareDiagnostics)","callee":"relativeComplement","castsInBody":0,"checkerDiagnostics":0,"implementationAnnotation":"T[] | undefined","implementationHasUndefined":true,"measurement":"read-only checker call-contract census, not Lower output","nativeAdmission":"not claimed; ordinary census stops earlier","nativeBoundaryRule":"call-site undefined check if lowering reaches censusOverloadResult; no body-proof erasure","nonNullAssertionsInBody":7,"overload":1,"owner":"getDiagnosticsWorker","result":"Diagnostic[]","resultHasUndefined":false,"roots":79,"where":"src/compiler/checker.ts:49742:51"}
```

This is a call-contract census, not Lower output. It establishes the actual
signature and the assertion guard that makes main's limited proof return false.
At internal/lower/census_small.go:308-317 on measured main, that failed proof plus
an optional implementation result and defined overload result produces the
call-site undefined check. The ordinary censuses stop earlier at core.ts:736,
so they do not establish emitted native output for this call.

The initial supplemental full-checker meter probe was terminated without a
completed file record. The existing focused statement-piece mode then panicked
with `latent state copy: unexported IR field methodTables`, before completing
its assigned owner. Both failed-probe artifacts are retained; neither is counted
as a completed file or substituted for an ordinary census root. No compiler or
measurement copier was repaired. call-census.go.txt is a read-only checker probe,
explicitly labeled so it cannot be mistaken for lowering admission. It is kept
as .go.txt, like the census-export.go.txt files of adaptations 30 to 33, so it
is evidence and never a package the gate builds: copy it to a .go file in a
scratch module to rerun it.


## Mints outside core.ts stay unchanged

| File and baseline line | Mint source |
| --- | --- |
| compiler/program.ts:3285 | `emptyArray as any as SortedReadonlyArray<Diagnostic>` |
| compiler/utilities.ts:6061 | `[] as Diagnostic[] as SortedArray<Diagnostic>` |
| compiler/utilities.ts:6062 | `[] as string[] as SortedArray<string>` |
| compiler/utilities.ts:6099 | `[] as Diagnostic[] as SortedArray<DiagnosticWithLocation>` |
| compiler/utilities.ts:6100 | `diagnostics as SortedArray<DiagnosticWithLocation>` |
| compiler/utilities.ts:6108 | `nonFileDiagnostics.slice() as SortedArray<Diagnostic>` |
| compiler/watchUtilities.ts:156 | `entries.files.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>` |
| compiler/watchUtilities.ts:157 | `entries.directories.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>` |
| server/project.ts:1778 | `emptyArray as SortedReadonlyArray<string>` |
| server/utilitiesPublic.ts:140 | `[] as any as SortedArray<T>` |
| services/services.ts:2957 | `[] as number[] as SortedArray<number>` |
| testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:386 | `cloned.entries.map(clone) as SortedArray<FSEntry>` |
| testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:976 | `[] as FSEntry[] as SortedArray<FSEntry>` |

The exact 746-file source comparison permits changes only inside the nine pinned
owners in core.ts and checker.ts. All these other files remain byte-identical.

The measured uncounted reference is the required base b729c7e0. Its only adaptation
input differences from measured main 82b0c076 are 99l, 99m and 99q. Reversing their
six pinned core owners produces main's source preimage with byte-identical
JavaScript in both comment modes; all common apply, parser, source and oracle
inputs are identical. main-runtime-proof.json records this artifact equivalence.
The three oracle identity streams therefore compare the adapted runtime to the
same runtime artifact as main. This is not a claim of a fourth fresh main oracle
invocation. All three actual runs and their complete identities are retained.

Both fresh ordinary compiler censuses still refuse all 73 shared roots at
core.ts:736:36, including the clean b834a6a4 merge. CENSUS.md reports that measured
limitation and the pinned branch's failing supplemental controls. This adapter
is source- and oracle-validated; no native unblock of those roots is claimed.

## Owner contract and verification

owners.json pins exact LF-normalized before and after text and SHA-256 for five
core owners and four checker owners. The complete overload sets are included.
Checker owners must remain directly inside createTypeChecker; core owners must
remain top-level. No whole-file hash is an adapter precondition. Stock TypeScript
6.0.3 parses both complete files before planning and the planned results.

Changed, missing, duplicate, moved, nested, partially adapted and mixed-state
owners are refused before any write. Every original file is reread before writes.
The contract passes 85 refusal fixtures, preserves a foreign selectIndex edit,
including movement into arrows or nested blocks, and preserves LF and CRLF. Replay changes zero files and edits, makes zero
writes, and preserves bytes and mtimes. Python's real directory sort places 99r
immediately after 99q; sort.json records the full ordered list.

| Check | Result |
| --- | --- |
| Fresh baseline full apply | Exit 0 |
| New-stack full apply | Exit 0 |
| Adapter second run | Zero files, edits and writes |
| Stock TypeScript 6.0.3 complete-source parsing | Pass |
| Adamic types, all 79 recursive roots, before and after | Both exit 0, zero diagnostics |
| Whole core.ts JavaScript, comments kept and removed | Byte-identical |
| typescript.d.ts and tsserverlibrary.d.ts | Both byte-identical |
| Uncounted, counted and adapted full oracles | 106,366 passing, one sanctioned failure, zero pending each |
| Exact title/outcome multisets | All three equal, including multiplicities |
| Unchanged lane sanction checker | Pass for all three oracles |
| Node witnesses for edit 3 | All three before/after outputs equal |
| Brand mutant | Exit 1, changed-owner refusal, zero writes |
| relativeComplement overload mutant | Clean copied baseline, then exit 1 and TS2322 at checker.ts:49743:17 |
| generate_shards --check | Exit 0 with configured tool PATH |
| Lane Go fixtures | Exit 0, 139.543 seconds |
| Ordinary per-file census, both compiler stacks | See CENSUS.md and raw JSONL streams |

The initial shard invocation lacked gofmt on PATH and failed before checking the
generated file. Sourcing /workspace/adamic-tools/env.sh resolved it; the rerun
passes without any generator or shard edit. No test expectation, timeout, runner,
worker count, compiler refusal, or public baseline was weakened.

## Reproduction

Set NODE_PATH to stage3/api/node_modules and source the configured tool environment.
Outputs belong outside this directory because every file here participates in
apply's product key. COUNT.md and its scripts retain the complete count method.

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99r/typescript-source stage3/apply.sh BEFORE_TREE
node stage3/adapt/99-correctness-axolotl-r/contract.cjs BEFORE_TREE /absolute/path/to/built/adamic
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99r/typescript-source stage3/apply.sh AFTER_TREE
node stage3/adapt/99-correctness-axolotl-r/adapt.cjs AFTER_TREE
node stage3/adapt/99-correctness-axolotl-r/debug-proof.cjs AFTER_TREE
node stage3/adapt/99-correctness-axolotl-r/resolve-witness.cjs BEFORE_TREE AFTER_TREE
node stage3/adapt/99-correctness-axolotl-r/compare.cjs BEFORE_TREE AFTER_TREE
python3 stage3/lane/generate_shards.py --check
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99r/typescript-source go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
```

BEFORE_TREE is the full apply of base b729c7e0 without this directory, not an apply
of the new stack. The first command therefore runs from the base checkout. The
Go command is 99q's fixture command; the full oracle separately runs every
upstream runner with all eight workers and no filter. Compiler provenance and
both exact census stack labels are recorded in CENSUS.md.

## 2026-11-06 amendment: readonly toSorted arm

Adamic authorized this amendment to task b5w3ycg at 20:05Z. The empty arm is
`emptyArray as readonly T[]`; the existing unknown/SortedReadonlyArray mint and
every other owner are unchanged. MEASURE.md retains the pre-edit evidence and
the earlier stop; this authorization resolves its cast-free condition.

docs/rulings.md:14 requires Adamic's sound relation, not TypeScript assignability.
Run `node contract.cjs BEFORE_99R_TREE /absolute/path/to/built/adamic` with stock
TypeScript 6.0.3 on NODE_PATH. The compiler argument is mandatory, directly or
through ADAMIC_CONTRACT_COMPILER; there is no TypeScript-only fallback.
relation-contract.cjs extracts the actual adapted arm, its target, the unchanged
emptyArray declaration and toSorted's unconstrained T. It emits two small probes
with number and string instantiations: the actual `as` expression and the same
source `satisfies` the extracted target. It runs `adamic types` and `adamic c` on
both. The latter invokes ordinary Load and LowerWithOptions, including castProof
and provenRelation. The satisfies witness requires an erased sound relation,
so acceptance as a checked downcast cannot satisfy the contract. No compiler
source or refusal rule is patched. Set ADAMIC_RELATION_EVIDENCE to retain probes
and complete outputs at a chosen scratch directory.

castProof's upcast path is internal/lower/cast_proof.go:219-229; it calls
provenRelation in proven_relations.go:11-30, checking writable slots, optional
fields and nominal identity after the assignability candidate. cast then erases
the accepted view. The owner contract also rejects the former mutable arm with
an otherwise adapted mint as a partial owner.
