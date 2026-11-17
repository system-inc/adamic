# Core overloads — #ev14ch9

**Disposition: road (iii), stop and report, for all four sets.** No adaptation
was built, no landing branch exists, and no ledger entry was added. The reasons
below are constraints of the current proof and public declarations, not an
assertion that no future lowering rule could prove these functions.

Base: `cf468df0747075e50d326a97ff03e8530d7a1a88`. Upstream TypeScript v6.0.3:
`050880ce59e30b356b686bd3144efe24f875ebc8`. A fresh detached worktree, recursive
shallow submodules, stage3/api npm ci, and a separate shallow upstream clone
were used. Upstream npm ci also completed. Scratch paths and pins are in
[provenance.json](provenance.json). No Go program was run; the supplied scratch
environment points GOTMPDIR, TMPDIR and GOCACHE outside the checkout.

The initial report was pushed first, at `4f597dacf`, before constructing or
checking any proposed adaptation. Preparing the **unchanged base** adapted tree
was already in progress to obtain the source to read; it was not a proposed
adaptation build. All later experiments were in-memory checker inputs. They
wrote no TypeScript source and produced no landing adapter.

| Set (upstream line) | Supplied latent sites | Resolved calls | Caller files | Road | Landing branch / SHA |
| --- | ---: | ---: | ---: | --- | --- |
| arrayFrom (1337) | 1 | 119 | 43 | (iii), mapped-container correlation needs proof | none / none |
| arrayToMap (1403) | 1 | 10 | 8 | (iii), default/projected-value correlation needs proof | none / none |
| reduceLeft (1228) | 1 | 27 | 5 | (iii), seeded loop/callback proof needs a ruling | none / none |
| sameMap (342) | 1 | 36 | 7 | (iii), readonly propagation reaches public mutable next | none / none |

Counts are identical upstream and on the fully adapted base. These count static
call sites, not runtime invocations. There is no whole-lane execution count:
no JavaScript-changing candidate was selected or built.

## Read and resolve

Read whole before choosing roads:
`internal/lower/census_small.go`, `overload_predicate_result.go`,
`census_overload_proof.go`, and `overload_result_check.go`. Also inspected
`predicateEffects` and the callback-use exception. Their source SHA-256 values
and the complete adapter ownership search are in
[ownership-and-rules.json](ownership-and-rules.json).

[upstream-callers.json](upstream-callers.json) and
[adapted-callers.json](adapted-callers.json) retain every complete overload and
implementation, every resolved call, all argument texts and types, instantiated
signature, result type, enclosing expression and immediate-use classification.
The checker is stock TypeScript 6.0.3. Every counted call uses
`getResolvedSignature`, and its declaration must be the named function in
`src/compiler/core.ts`. Every source call is queried without a name filter.
Upstream: 709 source roots, 710 in-checkout source files, 100,503 call/new/tagged
expressions. Adapted: 713 roots, 714 files, 102,785 expressions.

Outside src, the AST index visits **38,760 tracked code files**, including tests,
20,549,966 upstream nodes (20,549,968 adapted) and 353,301 call/new/tagged expressions. Only `lib/_tsc.js` and
`lib/typescript.js` contain references to these owner identifiers; their calls
resolve to bundled declarations, never the source core owners. Candidate alias
tracking covers imports, variable aliases and destructuring. No text search is
used to count callers. The two bundles are checked in a separate program with
source core available, scanning 127,803 call expressions. See
[other-index.json](other-index.json), [adapted-other-index.json](adapted-other-index.json)
and both caller logs. Noncandidate files contain no owner identifier from which
such an alias could originate.

[adapted-consumers.json](adapted-consumers.json) follows result bindings by
checker symbol, through parentheses, assertions, conditional branches and direct
assignments, retaining downstream uses and contexts. The upstream direct-binding
inventory is retained separately as [upstream-consumers.json](upstream-consumers.json).
The earlier direct-binding trace counted seven identity readers; the completed
conditional-result trace counts eight. This corrects both that intermediate
count and the brief's six. None of those counts is a runtime counter.

Caller classes below are disjoint immediate-result contexts. Full details and
argument classes remain in the JSON, including all harness/test-runner callers.

| Immediate use | arrayFrom | arrayToMap | reduceLeft | sameMap |
| --- | ---: | ---: | ---: | ---: |
| variable initializer | 30 | 6 | 7 | 10 |
| function return | 18 | 1 | 7 | 6 |
| passed argument | 19 | 0 | 3 | 15 |
| member receiver | 17 | 0 | 0 | 0 |
| for-of input | 8 | 0 | 0 | 0 |
| assignment | 6 | 1 | 1 | 1 |
| conditional arm | 7 | 0 | 3 | 4 |
| property initializer | 7 | 0 | 0 | 0 |
| other binary expression | 4 | 2 | 3 | 0 |
| expression-bodied arrow | 3 | 0 | 3 | 0 |
| total | 119 | 10 | 27 | 36 |

## arrayFrom

1. **Read.** 97 calls omit map and 22 supply it. Product source has 89 calls,
   harness 18 and testRunner 12. Supplied mappers include tuple-to-ModulePath,
   ImportSpecifier-to-BindingElement, Path-to-file-ID, source-map conversion,
   build-info serialization, completion/signature-help records and grouped
   arrays. Unmapped results are returned, iterated, sorted/joined, stored or
   passed as element collections. All 43 caller files are enumerated in JSON.
2. **Failure.** `censusOverload` result relation compares implementation
   `(T | U)[]` to overload 1's `U[]` under rigid binders. Presence/phantoms cannot
   remove T from a mutable array. No predicate is involved. The return proof
   rejects the loop and local `result`; call-site narrowing excludes generic
   implementations and arrays.
3. **Road.** Tested the type-only candidate that drops both overloads and defaults
   U to never. It emits byte-identical JavaScript with comments kept and removed,
   but adds **11 diagnostics** over the zero-diagnostic 79-root base. Full texts
   are in [candidates.json](candidates.json). For example, JSX's mapped result
   reaches the public `NodeFactory.createObjectBindingPattern(elements: readonly
   BindingElement[])` (upstream API snapshot line 7609); the broader result also
   admits ImportSpecifier. Widening that public declaration is a stop. Constraining
   T extends U rejects this real transformation and the tuple/path projections.
   A body split still builds local arrays in loops. Returning another builder
   call is effectful and refused by `predicateEffects`. No road (i) or (ii) was
   established. Candidate next road: a ruling/proof for the contents of the
   existing loop-built array under the supplied map parameter; no new shape,
   cast, unchecked annotation or public change was implemented.

Ownership search finds no whole-owner pin; a successful future source edit would
need its own `99-correctness-overload-array-from` directory.

## arrayToMap

1. **Read.** Five calls omit makeValue: two command-line-option maps, two source-
   file maps and a string map keyed by Path. Five supply it: two file-info-to-
   boolean projections, one NamedSourceText-to-SourceFileWithText conversion,
   one build-info-to-BuilderFileEmit conversion, and one diagnostic tuple-to-
   diagnostic-array projection. Six calls are product source, four testRunner;
   none is harness. They return/store the map, inspect membership, read values,
   iterate entries or assign it to existing builder/project map storage.
2. **Failure.** `censusOverload` result relation cannot serve overload 2's
   `Map<K,V2>` from implementation `Map<K,V1|V2>`. The default identity explains
   the V1 arm but does not justify it for a supplied mapper. Mutable value views
   cannot be narrowed by a presence check. The default alone blocks the return
   proof; its loop and local return independently block it. This generic,
   untagged container is not a call-site narrowing result.
3. **Road.** Dropping all four overloads and defaulting V2 to never is JS-identical
   but adds **2 compiler diagnostics**, both in builder map assignments (full
   messages in candidates.json). The mapped input and output types differ, so
   V1 extends V2 is false on actual calls. Replacing the default with a branch
   still needs a proof of the local Map's contents, and returning `new Map` or
   another builder call is effectful under the present return proof. Splitting
   helpers would be a new shape. **Stop**, pending a ruling/proof of the existing
   supplied-mapper/default-identity correlation.

No whole-owner pin exists. A successful future edit would need its own
`99-correctness-overload-array-to-map` directory.

## reduceLeft

1. **Read.** Six calls omit initial (generic minimum, comma expressions,
   signature combination, and three Type selections); 21 supply initial, with
   two of those also supplying start. Seeded accumulator classes are: 12 numeric/
   flag accumulators, one boolean, one optional IndexInfo array, one optional
   Type array, two Type values, two Expression arrays and two Expression values.
   Element types are independently SourceFile, Type, Declaration, IndexInfo,
   Symbol, diagnostic-chain or syntax-node types. All calls are product source.
2. **Failure.** Strict `censusOverload` callback parameter relation fails. The
   initialized overload's U is independent of T; the implementation's callback
   consumes T|undefined after 49-honest-optionals and returns T. Even the second
   overload's callback need not accept undefined. `overloadCallbackUse` explicitly
   refuses any arguments read; this body reads arguments.length, then passes a
   mutable local accumulator around a while loop. It is not a bare forwarded
   overload parameter. Result annotation alone cannot fix a parameter refusal.
3. **Road.** Dropping both overloads adds **105 compiler diagnostics** while
   retaining JS bytes. SourceFile counters immediately infer SourceFile instead
   of number; initialized heterogeneous folds lose their required contract.
   Retaining only the initialized overload does not make U identical to T.
   Making the accumulator required or asserting it would undo the earlier
   undefined correction. A seeded-iteration callback proof must relate the
   initial value and each previous callback result to U, while preserving the
   actual arguments-length distinction, empty-array returns and optional start/
   count semantics. That needs a ruling/new proof; splitting helpers is a new
   shape. **Stop** without changing runtime behavior or claiming a body proof.

49-honest-optionals edits this body, but does not pin its whole owner. A future
new directory must pin the owner after that adapter and preserve its correction.

## sameMap

1. **Read.** All 36 calls supply array and callback. Sixteen resolve to mutable
   result overloads, twenty to readonly result overloads; two have a top-level
   undefined result. All calls are product source. Eight callers observe identity:

   | Adapted call | Comparison |
   | --- | --- |
   | checker.ts:14065 | types !== original intersection types |
   | checker.ts:15829 | reducedTypes === unionType.types |
   | checker.ts:22347 | normalizedTypes !== type.types |
   | checker.ts:22368 | elements !== normalizedElements |
   | checker.ts:23106 | constraints !== source intersection types |
   | checker.ts:25492 | primaryTypes === types (through a conditional initializer) |
   | services/codefixes/helpers.ts:411 | typeParameters !== newTypeParameters |
   | services/codefixes/helpers.ts:440 | parameters !== newParameters |

   Remaining uses pass transformed arrays into type/factory operations, return
   them, assign them, or use them in conditionals. Copying is not a candidate.
2. **Failure.** Axolotl-k already establishes T extends U; the remaining
   `censusOverload` refusal is readonly U[]|undefined against mutable U[].
   Presence alone cannot remove readonly. The body proof rejects reached casts/
   non-null assertions, loops and the local result; call-site narrowing rejects
   generic implementations and arrays.
3. **Road.** Dropping mutable overloads while retaining readonly overloads is
   JS-identical, but adds **4 compiler diagnostics**. Three can be pursued into
   internal checker parameter annotations. The fourth is builder's mutable
   DiagnosticMessageChain[] return. A second in-memory experiment propagates
   readonly through three checker helpers and the builder array helper. It
   reaches **three TS2322 errors at builder.ts:544, :550 and :559**, each assigning
   readonly arrays to public `DiagnosticMessageChain.next`. There is also one
   further internal readonly-parameter diagnostic. Full edits and messages:
   [samemap-boundary.json](samemap-boundary.json). The public snapshot explicitly
   has `next?: DiagnosticMessageChain[]` at upstream line 6921. Widening it is
   forbidden road (iii); a copy violates identity; a cast is not a proof.
   **Stop at that public boundary.** A future identity-preserving mutability
   correlation proof would need a ruling rather than an unchecked annotation.

Any sameMap edit is an amendment to `99-correctness-axolotl-k`: its complete
owner after-pin and SHA, README and contract would need updating. No separate
sameMap adapter was created and no existing pin was changed.

## Measurements and proof lines

The scratch type-only experiments used stock tsc 6.0.3's checker over exactly
79 compiler roots, with locked upstream Node types. Baseline diagnostics: **0**.
Each candidate is checked independently against the same base. Each reports
JavaScript equality with comments retained and removed. They are **rejected
candidates**, not passing adaptations or proof that their census sites disappear.
The scripts exit zero when measurement completes; diagnostic counts, not that
exit, determine rejection. [candidates.json](candidates.json) lists every new
candidate diagnostic in full and records exact edits; the baseline list is in
[candidates.json.baseline.json](candidates.json.baseline.json).

| Required proof line | arrayFrom | arrayToMap | reduceLeft | sameMap |
| --- | --- | --- | --- | --- |
| scan-all, site gone / 0 new / 0 changed | not run | not run | not run | not run |
| LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, owner + every caller file, CommonJSBundle | not run | not run | not run | not run |
| restoring mutant brings site back | not run | not run | not run | not run |
| landing candidate stock tsc, 0 new over 79 roots | not run | not run | not run | not run |
| adamic types, 0 | not run | not run | not run | not run |
| changed adaptation stage3/apply.sh | not run | not run | not run | not run |
| hereby dts; both public declarations byte-identical | not run | not run | not run | not run |
| TestProduct_Stage3AdaptedTree | not run | not run | not run | not run |
| whole unfiltered stage3/lane | not run | not run | not run | not run |
| road (ii) whole-oracle difference counter, Node witnesses, restoring mutant | not run | not run | not run | not run |
| landing ledger line | not run | not run | not run | not run |

The **unchanged base** `stage3/apply.sh --build-product` passed, exit 0; see
[apply.log](apply.log). This is tree preparation, not the product test that
builds it with stock tsc. No latent comparison ran, so there is no claim of
0 new/changed findings and no invented list of census movers. No public dts
build or byte-comparison ran. No source adapter, production lowerer, test
expectation or public declaration was changed by this task.

## Reproduce the evidence

The .cjs files here are the measurement scripts relocated from scratch; only
the parser/type-package paths were made configurable. With the base worktree,
its installed stage3/api and the pinned upstream prepared:

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/pinned/TypeScript
node --max-old-space-size=8000 stage3/census/overload-core/index.cjs "$ADAMIC_TYPESCRIPT_SOURCE" /scratch/other-index.json
CALLERS_INDEX=/scratch/other-index.json node --expose-gc --max-old-space-size=8000 stage3/census/overload-core/callers.cjs "$ADAMIC_TYPESCRIPT_SOURCE" /scratch/upstream-callers.json
# Prepare the unchanged base tree using stage3/apply.sh --build-product /scratch/adapted.
node --max-old-space-size=8000 stage3/census/overload-core/index.cjs /scratch/adapted /scratch/adapted-other-index.json
CALLERS_INDEX=/scratch/adapted-other-index.json node --expose-gc --max-old-space-size=8000 stage3/census/overload-core/callers.cjs /scratch/adapted /scratch/adapted-callers.json
node --max-old-space-size=8000 stage3/census/overload-core/consumers.cjs /scratch/adapted /scratch/adapted-consumers.json
node --max-old-space-size=8000 stage3/census/overload-core/candidates.cjs /scratch/adapted /scratch/candidates.json
node --max-old-space-size=8000 stage3/census/overload-core/samemap-boundary.cjs /scratch/adapted /scratch/samemap-boundary.json
```

The initial broad checker attempts encountered stack limits in unrelated bundled
regression inputs; the final traversal uses iterative AST walks and only asks
for bundled signatures after owner-symbol validation. A redundant batch scan
was stopped after the complete AST index established the only two candidates;
its partial log is retained. One diagnostic run was stopped before completion
because base adaptation was still running; its incomplete data was discarded,
and the complete stable-base rerun above is the reported measurement. No such
preparation attempt is represented as a passing proof line.
