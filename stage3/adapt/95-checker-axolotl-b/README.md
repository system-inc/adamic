# 95b: eight required reads, nine checker diagnostics

Built on 95a `ef4d26a4a38b7890317e1ba71f8ad928fd8588f6` (which includes
`d5c15cbf84b8c1732f01579d6bcd8ea8d3c974c1`). Branch:
`typescript/axolotl-checker-b`. TypeScript source and API are stock 6.0.3,
upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8`.

lane: run as part of a Correctness wave

Eight assertions remove the nine requested diagnostics. checker.ts own-file
17 to 8; whole-program tsc-entry diagnostics 131 to 122; no finding added
anywhere. No read was omitted for an unproved caller. This head does not edit
12034 or any other 95a survey row. No public API line changes.

## Each required-value obligation

Every insertion is `!`, a loud required-value check in Adamic, erased on Node.
All operands stay at their original evaluation points. No initializer assertion,
cast, default, optional chain, declaration change or source printer is introduced.
Existing casts in the inventoried expressions remain unchanged.

| Diagnostic site | Read | Invariant from compiler code |
| --- | --- | --- |
| 17946:29 | `elementTypes[unionIndex]!` | `findIndex` returns a nonnegative position only after its predicate reads that populated element's flags and identifies a union/never variadic constituent. The successful `checkCrossProductUnion` branch only reduces sizes; it neither writes the list nor invokes its error/tracing branch. The selected populated element remains there. |
| 19044:25 | `types[unionIndex]!` | The nonnegative `findIndex` result identifies the interpolation Type whose flags the predicate just read. Successful cross-product validation does not mutate the list. Required presence is at this selected read, not an assertion over all template types. |
| 20595:32 | `targets[i]!()` | The Deferred mapper's three producers create one callback per source, in the same order. The bounded sources loop selects the corresponding fresh callback. The Array mapper's separate, already asserted `targets[i]` is occurrence 1 and stays untouched; this edit is occurrence 2. Full producer audit below. |
| 22498:29, 22498:32 | `...stack[0]!` | The private incompatible stack contains diagnostic tuples pushed by the relation-error reporter. The `stack.length === 1` branch selects its sole populated tuple. Resetting the captured stack variable does not change the saved local array. One read accounts for both spread/iterator diagnostics. |
| 43517:20 | `getTypeArguments(type as GenericType)[0]!` | The branch tests reference-target identity with the validated arity-1 global Promise. Compiler-built immediate/deferred/cyclic references have a populated argument per declared parameter. Missing or malformed globals cannot introduce an emptyGenericType reference through the audited constructors. Full arity and constructor audit below. |
| 46114:128 | `yieldType!` at `resolveIterationType` | The first fast branch recognizes only one of eight validated arity-3 globals across its two fixed resolvers. Its first destructured argument is populated. The fallback occurrence of yieldType is unchanged. |
| 46114:208 | `returnType!` at `resolveIterationType` | The same validated arity-3 target identity guarantees the second argument. No assertion is added to nextType or to the later builtin returnType, which already has its own defined producer. |
| 46128:128 | `yieldType!` at `resolveIterationType` | The second fast branch recognizes only the arity-1 builtin targets in the two fixed resolver lists, so its first argument is populated. The fallback yieldType read stays unchanged. |

## All Deferred mapper producers

A stock-checker symbol inventory finds exactly four references to the private
`makeDeferredTypeMapper`: its definition and three direct calls, with no alias
or indirect escape. The adapter verifies that inventory, the exact call text,
and the full function text of the producer and each caller before writing.

| Caller | Sources | Targets and lifetime obligation |
| --- | --- | --- |
| `getInferredTypeParameterConstraint` | `typeParameters` of the enclosing type reference | `typeParameters.map((_, index) => () => getEffectiveTypeArgumentAtIndex(..., index))` supplies one fresh closure for each same-position source. The mapper is immediately consumed by constraint instantiation; no list write occurs between construction and use. |
| `makeFixingMapperForContext` | `map(context.inferences, i => i.typeParameter)` | A second map over the same unchanged inference list produces fresh closures. Callback bodies may later fix an inference but do not mutate either of these newly allocated mapper arrays. |
| `makeNonFixingMapperForContext` | `map(context.inferences, i => i.typeParameter)` | A second map over the same unchanged list supplies the matching fresh callback array. Neither map callback changes the inference list during construction. |

The Deferred object is constructed only by this private factory. Its targets
are these fresh map results, not a caller-owned callback array subsequently
shortened through an outside alias. getMappedType is the corresponding indexed
consumer. An array-of-functions declaration alone would not establish this.

## Both fast-path callers and every global getter

`getIterationTypesOfIterableFast` has exactly three symbol references: its
definition and two direct calls in `getIterationTypesOfIterableWorker`. There
is no callback alias. The calls use only `asyncIterationTypesResolver` and
`syncIterationTypesResolver`; both complete object declarations are pinned.
Their resolve callbacks cannot select arbitrary global getters.

| Getter group | Names | Validated arity |
| --- | --- | ---: |
| Promise getter | Promise | 1 |
| Sync first fast branch | Iterable, IteratorObject, IterableIterator, Generator | 3 each |
| Async first fast branch | AsyncIterable, AsyncIteratorObject, AsyncIterableIterator, AsyncGenerator | 3 each |
| Sync builtin list | ArrayIterator, MapIterator, SetIterator, StringIterator | 1 each |
| Async builtin list | ReadableStreamAsyncIterator | 1 |

Each of the nine individual Promise/first-branch getters passes the literal
arity to `getGlobalType`. Each builtin list calls `getGlobalBuiltinTypes` with
arity 1, and that helper applies `getGlobalType` to every entry. All those
functions are pinned. `getTypeOfGlobalSymbol` rejects missing symbols,
non-object types and parameter-count mismatches. Failed lookups return
undefined or the anonymous `emptyGenericType`, not a usable validated target.

The identity tests require ObjectFlags.Reference and the exact target, not
structural assignability or a similar spelling. However, the fallback's own
non-reference flags alone would not prove that no reference could target it:
the constructor callers must also be audited, as below.

## Reference construction, missing globals and argument filling

The symbol inventory also records all 15 createTypeReference calls, all three
createDeferredTypeReference calls, all eight tryCreateTypeReference calls,
and the two fast-path calls. Definition counts are 16, 4, 9 and 3 respectively;
no additional alias is accepted. Exact calls and owners are in sites.json.

| createTypeReference caller | Why it cannot supply a short argument list to the validated fast targets |
| --- | --- |
| `createTypeChecker`'s `getAnyAsyncIterableType` arrow | Rejects emptyGenericType, then supplies three anyType arguments. |
| `getTypeOfPrototypeProperty` | Maps every declared type parameter to a populated anyType. |
| `getTypeWithThisArgument` | Checks argument/parameter lengths before appending the populated this argument; it cannot remove an initial required argument. |
| `tryCreateTypeReference` | All eight callers are the decorator context/target/result wrappers. They pass literal arrays of one or two defined types. `some(typeArguments)` is therefore true, and a missing-global emptyGenericType returns unknownType before reference construction. |
| `getTypeFromClassOrInterfaceReference` | Invalid TypeScript argument counts return errorType. JavaScript recovery calls fillMissingTypeArguments with its JS mode, which fills the declared parameter list. Deferred references use the same effective-argument filler. |
| `createTypeFromGenericGlobalType` | Rejects emptyGenericType before construction. Its fixed global wrappers supply arity-1 descriptor/array arguments or arity-3 iterable/generator arguments; the readonly-array initialization supplies one argument. |
| `createNormalizedTypeReference` | Array/tuple syntax rejects emptyGenericType before this helper; tuple normalization constructs populated aligned elements. Instantiation's remaining caller maps an existing reference's complete arguments. |
| `createNormalizedTupleType` (two calls) | Tuple targets and their aligned element lists are compiler-built tuple state, not a validated nominal global target with missing parameters. |
| `getNormalizedType` | Reuses the original reference's resolved arguments. |
| `createMarkerType` | Instantiates every declared type parameter into a populated marker argument. |
| `getWidenedTypeWithContext` | sameMap widens the existing complete argument list without dropping positions. |
| `instantiateAliasOrInterfaceWithDefaults` | Checks the supplied count against the declared parameters and fills the remaining declared parameters before creating the interface reference. |
| `createPromiseType` | Rejects emptyGenericType, then constructs exactly `[promisedType]`, with an unknownType fallback for awaited resolution. |
| `createPromiseLikeType` | Likewise rejects the fallback and supplies exactly one populated argument. |

The eight tryCreateTypeReference callers are createClassDecoratorContextType,
createClassMethodDecoratorContextType, createClassGetterDecoratorContextType,
createClassSetterDecoratorContextType, createClassAccessorDecoratorContextType,
createClassFieldDecoratorContextType, createClassAccessorDecoratorTargetType
and createClassAccessorDecoratorResultType. None passes an empty array.

The three deferred constructors are called from
getTypeFromClassOrInterfaceReference, getTypeFromArrayOrTupleTypeNode, and
getObjectTypeInstantiation. The last preserves the original target and syntax
while composing its mapper. `getTypeArguments` resolves those syntax arguments
with getEffectiveTypeArguments/fillMissingTypeArguments. Circular resolution
supplies an errorType for each local parameter instead of a missing element.
Generic targets' self-instantiations carry their own populated parameters;
cloning and mapping preserve those argument positions. These obligations apply
to compiler-constructed types; they do not promise validity for fabricated
untyped JavaScript objects pretending to be compiler Types.

No caller breaking these obligations was found; none of the eight reads is
omitted. The adapter pins the relevant arity, recovery and constructor helper
functions in addition to the six edited owners. Its call inventories reject
new direct callers and aliases for the audited construction paths.

## Mechanical scope and composition

The adapter uses 47-host-errors' compiler(), stock TypeScript 6.0.3 parsing,
bound symbols, structural type fingerprints, exact parsed expression occurrence
counts, and before/after text hashes for each of the six edited functions.
The structural fingerprint sorts union members; display ordering of a tuple's
SourceFile/readonly-string union varied with checker query order during the
first attempt, which correctly stopped before writing. A type display string
is retained as documentation but is not the acceptance predicate.

There is **no checker.ts whole-file hash pin**. Each owner's recognized before
or after hash is independent of unrelated functions and shifts elsewhere.
Dependency function text, resolver/fallback declarations and caller inventories
protect the stated obligations. All edits are planned, syntax-checked, compared
through stock transpileModule in both comments modes and rechecked against the
unchanged source before the single write. CRLF and evaluation order are retained.
A mixed or otherwise unrecognized state inside an owner is refused.

verify.cjs is independent of the adapter's planner. It inventories non-null AST
nodes, requires exactly eight additions, checks each owner's complete before
and after contract, reconstructs the entire before file by restoring those six
owners, and requires all other source files identical. It independently compares
stock JavaScript bytes in both comments modes and checks CRLF counts.

## Meter counts and required mutant

Use the real census executable built from the 95a compiler, rooted only at
src/tsc/tsc.ts, with unchanged production internal/load options. Like 95a's
compare-census.py, this head's comparator verifies the root and aggregate
observations agree and counts one observation, rather than doubling them.
Primary-file locations determine checker.ts attribution; elaboration lines
are not counted as diagnostics.

| Tree | Whole program | checker.ts own file |
| --- | ---: | ---: |
| 95a control | 131 | 17 |
| 95a + 95b | 122 | 8 |
| Undo stack assertion | 124 | 10 |

proofs/census.json contains all nine exact removed diagnostics, zero added
findings, and both exact findings restored by the one-edit mutant. The isolated
mutant changes only `...stack[0]!` back to `...stack[0]`. The independent verifier
exits 1 for `all eight required-read edits must be present`, `7 !== 8`; the
adapter's check exits 1 for `required assertion missing: reportIncompatibleStack:1`.

A separate in-memory runtime mutant adds an extra reportError argument; stock
JavaScript comparison rejects it with comments retained and removed.
prove-mutants.cjs replays both controls; proofs/mutants.json records the results.
Neither mutant enters the oracle tree.

Measurements use clean apply trees before upstream npm installs. The first
mutant copy accidentally included the concurrently installed oracle packages;
source-map-support then resolved, removing one unrelated TS2307. The exact-set
comparison rejected that environment difference. Moving the copied node_modules
outside module resolution and rerunning restored the intended clean-tree
comparison: precisely two original tuple-spread diagnostics return.

## Commands and environment

All command and test output goes to logs. Toolchain setup uses
/workspace/adamic-tools/env.sh. nproc is 5, cgroup CPU quota 4; Go 1.27.1,
Node v24.19.0 and clang 20.1.8. Setup completed in 17.055 seconds; full timing
lines are retained in proofs/setup.log.gz. No compiler or Go source changes.

The retained control is the fresh 95a end-to-end apply output from the preceding
head. Reproduction with this directory installed: run ordinary apply to a new
path, copy that output as the control and restore only the six owner texts from
apply's refs/stage3/95-checker-axolotl-b-before Git tree (or restore the entire
checker.ts from that tree when no later heads are present). Do not rerun 95a's
whole-file pin on a tree already modified by a later head.

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH="$PWD/stage3/api/node_modules"
node stage3/adapt/95-checker-axolotl-b/adapt.cjs AFTER > /tmp/axolotl-b-adapt.log 2>&1
/tmp/axolotl-census BEFORE/src/tsc/tsc.ts before.jsonl > /tmp/axolotl-b-before-census.log 2>&1
/tmp/axolotl-census AFTER/src/tsc/tsc.ts after.jsonl > /tmp/axolotl-b-after-census.log 2>&1
node stage3/adapt/95-checker-axolotl-b/verify.cjs BEFORE AFTER > /tmp/axolotl-b-verify.json 2>&1
node stage3/adapt/95-checker-axolotl-b/adapt.cjs AFTER > /tmp/axolotl-b-idempotence.json 2>&1
# Copy AFTER to MUTANT, undo only ...stack[0]!, and keep its npm resolution equal to BEFORE.
/tmp/axolotl-census MUTANT/src/tsc/tsc.ts mutant.jsonl > /tmp/axolotl-b-mutant-census.log 2>&1
python3 stage3/adapt/95-checker-axolotl-b/compare-census.py before.jsonl after.jsonl mutant.jsonl > /tmp/axolotl-b-census.json 2>&1
node stage3/adapt/95-checker-axolotl-b/prove-mutants.cjs BEFORE AFTER MUTANT > /tmp/axolotl-b-mutants.json 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout --kill-after=20s 880 bash stage3/oracle/run.sh AFTER NEW_ORACLE --limit-seconds 720 > /tmp/axolotl-b-oracle.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/axolotl-upstream bash stage3/apply.sh NEW_APPLY > /tmp/axolotl-b-apply-final.log 2>&1
```

No stage3/lane, checked-in patch table, fixture, meter, other adaptation or Go
file is edited. Ordinary apply writes its generated patch table only inside its
output. Native tsc compilation, remaining checker diagnostics, full Go gate,
browser and ESLint integrations are outside this head.

## Completed proof results

Fresh full apply exits 0 in **208 seconds**, running 95a before 95b. Its
incremental row is **1 file, 7 added and 7 removed lines**: eight assertions
occupy seven lines because yieldType and returnType share line 46114. The fresh
output's complete **711-source** map equals the oracle tree's map. Its separate
real census reproduces exactly **122 whole-program / 8 checker own-file**
diagnostics, including the unchanged full diagnostic texts. Its second adapter
run makes **0 files / 0 assertions**. See proofs/apply.json,
proofs/final-source.json and proofs/final-idempotence.json.

Stock transpileModule output is byte-identical before/after in both modes:

| Comments | Bytes | SHA-256, before and after |
| --- | ---: | --- |
| Retained | 2,892,419 | `786ee138ff00765ab7c7ea7651de48df32c5981ec2c0810c68e302570d566f6e` |
| Removed | 2,354,716 | `3664a2f5fb2aba0e839a3153e6ff4d093b530db890b5b2433be6cd38e6e91aac` |

CRLF counts are **54,434 before and after**. Only checker.ts changes; all other
710 source files retain their bytes. Full-file hashes in proof artifacts are
observations, never adapter acceptance pins. There is no JavaScript-changing
head requiring separate before/after behavioral witnesses.

The standalone oracle used all default runners, no filter, eight workers,
`--light=false`, and lint disabled. It completed within 15 minutes:

| Phase | Exit | Seconds |
| --- | ---: | ---: |
| Install | 0 | 2.833 |
| Build | 0 | 25.580 |
| Tests | 1 | 358.988 |
| Complete oracle | 1 | 387.489 |

**106,366 passing, 1 failing, 0 pending**, matching main. Raw oracle status is
**fail**, solely for the standing
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
The sole baseline difference is api/typescript.d.ts; **0 new baseline
differences**. Both its built declaration and accepted reference match main's
recorded hashes in adaptation 40's brands-undefined-3/api-runtime.json:
`b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`
and `f87d6c30da287293aae741711fa94594b75d49e0977cea6bfd5982f90e2bafcf`.
The local API baseline equals the built declaration. Complete baseline.diff
is 51,123 bytes, SHA-256
`6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`.
No public API line is edited, no lane acceptance is claimed, and no timeout or
exception occurred. Raw report, compressed phase logs and complete diff are
retained under proofs/.

## Order independence and guard controls

The adapter applies to 95a alone, then reruns with zero edits. Remote branch
checks at the beginning and end found only the published 95a branch; neither
`typescript/axolotl-checker-c` nor `typescript/axolotl-checker-d` existed. No
combination with those unpublished heads is claimed. proofs/composition.json
records the exact refs and final UTC observation.

As an additional owner-pin control, prove-isolation.cjs appends an unrelated
comment outside every edited/dependency owner in an isolated 95a copy. This
changes the checker.ts full-file hash, but 95b still makes all **8 edits**, and
the next run makes **0**. This proves there is no hidden full-file acceptance
pin; it does not substitute for testing actual sibling branches when published.

In the same scratch copy, a separate producer mutant replaces the Deferred
factory's targets with `targets.slice(1)`. The caller/arity dependency guard
rejects makeDeferredTypeMapper before any write. The scratch file is restored
in finally. proofs/isolation.json records both controls. The required real-census
one-edit tuple mutant remains independently caught by exact counts and by the
source verifier; see proofs/mutants.json and proofs/census.json.

The initial auxiliary mutant runner expected an uninventoried-owner rejection
for the reverted tuple function. A fully reverted owner is instead a recognized
before state, so --check correctly rejects its missing assertion. The runner
was corrected to require the actual `required assertion missing:
reportIncompatibleStack:1` catcher and rerun successfully. No adaptation or
checker acceptance condition was weakened.
