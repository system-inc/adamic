# Overload body proof scout — #5fws6q8

Report only. No `internal/`, adaptation, public declaration, or ledger change.
The earlier group report is pushed at `472fc1926`; it selected no adaptation
builds. This scout measures the unchanged fully adapted tree of Adamic main
`cf468df0747075e50d326a97ff03e8530d7a1a88`, using upstream TypeScript
`050880ce59e30b356b686bd3144efe24f875ebc8` and stock checker 6.0.3.

**Recommendation:** two small, separate fresh-container certificates, plus a
bounded fold certificate. Do not merely allow returned locals, loops, defaults,
or `arguments.length`. The fold certificate must solve callback/result
provenance **and** enforce the admitted runtime arity. A written maximum
parameter count alone is unsound, as the executable alias witness demonstrates.

## Exact failures on cf468df0

All paths/lines in this section are `internal/lower/` at that base, not lines of
an imagined implementation. These are source traces; no instrumented lowering
or replacement compiler was built.

| Set | Decisive current proof failure | Final refusal |
| --- | --- | --- |
| arrayFrom, overload 1 | `census_overload_proof.go:213`: `censusReturnProof` falls through its statement switch on the `for-of`. | `census_small.go:256`, result U[] versus (T\|U)[] |
| arrayToMap, overload 2 | `census_overload_proof.go:67–68`: `censusReturnProof` refuses `makeValue`'s parameter initializer before walking the body. | `census_small.go:256`, Map<K,V2> versus Map<K,V1\|V2> |
| reduceLeft, overload 1 | `overload_callback_use.go:60–61`: `overloadCallbackUse` first fails callback-return relatedness, U to T under the implementation's rigid T mapping. It has **not reached** the arguments test. | `census_small.go:211–218`, parameter f |
| reduceLeft, overload 2 | `overload_callback_use.go:64–65`: after matching callback results T to T, the blanket `arguments` read guard refuses. | `census_small.go:211–218`, parameter f |

The other requested steps, in actual order:

* `censusOverloadParameter`, `census_small.go:261–281`, extracts these ordinary
  non-rest parameters successfully. At :270–274 it reads their types and applies
  the rigid mapper. It does **not** prove correlations or reject a loop. A missing
  parameter returns nil at :267–268; `censusOverload` supplies undefined at
  :202–204. Defaults are incoming optional values: `censusCallableParameterType`,
  :405–409, adds undefined even though the default's body-local type excludes it.
  Do not drop that incoming undefined when adding a result proof.
* `arrayFrom` and `arrayToMap` pass parameter comparison at
  `census_small.go:211`; the callback-use exception is not needed there. Their
  implementation results fail ordinary relatedness at :223.
* Both then try `predicateOverloadResultProven` via `census_small.go:234`.
  `overload_predicate_result.go:40–42` finds no `x is U` callback, so
  :95–97 returns false with an empty predicate-callback list. This is the first
  negative predicate attempt, not a failed checked-return witness. Returning an
  array of transformed values is not the existing single-value predicate proof.
* `arrayFrom`'s body proof next stops on the loop at :213. Even if a loop were
  blindly skipped, returning the local `result` would fail at
  `census_overload_proof.go:129–131`.
* `arrayToMap` stops on its default at :68 first. Removing that guard alone
  would expose its unsupported for-loop at :213 and local return at :131.
  Its fourth overload additionally has fewer generic binders than the
  implementation, which the standalone body proof refuses at :50–52. A new
  certificate must use the validated witness mapping established by
  `census_small.go:151–175`, not blindly positional substitution.
* `reduceLeft` is refused during parameter comparison, so neither the predicate
  nor return proof runs for the failing declaration. If only the callback gate
  were relaxed, the current return proof would first reject the **outer**
  `array && array.length > 0` test at `census_overload_proof.go:174–175`:
  `censusTrustedTest` has no greater-than case (:275–293). It would not yet reach
  the nested `arguments.length <= 2` test. Later blockers include checked index
  assertions, the while-loop, and the returned mutable local. Relaxing :64 alone
  fixes none of the independent accumulator-return or body-proof obligations.
  After repairing f alone, overload 1 would also fail on parameter initial:
  the admitted U does not relate to implementation T|undefined at
  `census_small.go:211`; its non-callable callback exception fails at
  `overload_callback_use.go:43–45`. The fold certificate must prove this seed's
  actual uses as well as f's uses, not just suppress the first error.

Generic call-site narrowing remains unavailable at
`overload_result_check.go:46–48`; arrays are also uncheckable there at :102–103.
No proposal below removes either restriction or mutable container invariance.

## Measurement before choosing the bounded rules

[scout.cjs](scout.cjs) scans **247 adapted compiler/services files**, **15,180
function-like bodies**, and **195 function overload sets**. It uses ASTs and
checker symbols, not textual function-name matches, to identify locals, aliases,
closure references, defaults, declared accessors and resolved target calls.
`arguments` means the builtin symbol with no declarations, matching
`arguments_length.go:12–18`; an AST property or destructuring key named
`arguments` is not a read of that object.

The broad container search starts at fresh empty `[]` or an empty intrinsic
`new Map()`, requires a direct local return or a return through local initializer
aliases, and requires an owner-level loop. It includes non-overloaded producers
so that the hazards around a tempting generalization are visible. It does not
claim to enumerate every latent refusal or every arbitrary heap alias.
Unknown stores/calls involving the local are classified as escapes and refused;
this is deliberately not interprocedural alias acceptance.

| Domain | Broad owners (locals) | Overloaded owners | Eligible for bounded rule | Other broad owners excluded | Other overloaded owners excluded |
| --- | ---: | ---: | ---: | ---: | ---: |
| fresh loop-built returned array | 86 (87) | 5 | 1: arrayFrom | 85 | 4 |
| fresh loop-built returned Map | 4 (4) | 1 | 1: arrayToMap | 3 | 0 |
| builtin arguments reader | 1 | 1 | 1 conditional fold: reduceLeft | 0 | 0 |

**These are static eligibility/proposed-coverage counts, not executions of a new
Go acceptance rule.** The two exact builder grammars match one owner each;
per-overload semantic obligations for those owners are derived below. The fold
has one candidate and no neighbouring arguments reader, but acceptance remains
conditional on the arity/escape gate. No changed census was run or claimed.

The four overloaded array neighbours are `takeWhile`, `relativeComplement`,
`arrayToNumericMap`, and checker
`createTypeChecker/createNodeBuilder/symbolTableToDeclarationStatements/serializePropertySymbolsForClassOrInterface`.
They need predicate, indexed/sparse-store, other-branch or other-expression
proofs outside the bounded arrayFrom grammar. Some already have related results;
being a neighbour is not a claim of a current refusal.

The other Map producers are `createSymbolTable`, `serializeOptionBaseObject`,
and services `getDirectImportsMap`. The first two read properties as set values;
the third captures the map and reads it with get. These are excluded by the
bounded grammar, not all claimed semantically incorrect.

### Hazard roads and counts

Counts are **distinct owners**, may overlap, and are not additive. Full rows,
all reference locations/contexts, allocation records, source hashes and target
call audits are in [inventory.json](inventory.json); compact counts are in
[summary.json](summary.json).

| Road found in broad candidate domain | Array owners | Map owners | Fold candidate |
| --- | ---: | ---: | --- |
| direct write/mutator other than push or set | 9 | 0 | 3 memo assignment sites, all required seed/step sites; no extra assignment |
| alias/store/pass of local | 13 (14 locals) | 0 | 0 memo aliases |
| nested closure captures local | 8 | 1 | 0 memo captures |
| explicitly declared getter read in a pushed/set argument | 0 | 0 | not a container rule |
| property/element reads in pushed/set arguments; may dispatch a getter | 33 | 2 | 2 existing checked array reads |
| default parameter could stand in for an omitted/undefined overload argument | 1 | 1 | 0 default parameters |
| builtin arguments read elsewhere than the allowed literal length test | 0 | 0 | 0; exactly one read, the <= 2 test |

The array default is `arrayToNumericMap`, which also writes by index. The Map
default is `arrayToMap` itself: it is an exposure requiring a path proof, **not**
permission to pretend identity is a supplied mapper. The mapped overloads exclude
undefined at that parameter; omitted-mapper overloads must prove identity instead.

Zero declared getters does not establish zero possible getters through structural
property types. The 33/2 rows deliberately count those broader possible roads.
Do not trust arbitrary property-flow narrowing on the strength of the zero;
`array-getter.ts` and `map-getter.ts` execute the counterexample.

Nine array owners also use the local outside the restricted push/return read
whitelist. A non-mutating Map.get is a read, not counted as a non-set write.
No array/Map source candidate reads builtin arguments elsewhere. Across the
entire measured compiler/services domain the only builtin reader is reduceLeft;
there are zero other readers to grandfather.

The target reference audit resolves calls through the checker and records **70
arrayFrom calls** (55 unmapped, 15 mapped), **4 arrayToMap calls** (2 identity,
2 mapped), and **27 reduceLeft calls** (6 two-argument, 19 three-argument,
2 four-argument). There are **0 callable-value uses** and **0 spread calls** for
these three within this domain, excluding imports, declarations and type uses.
These are compiler/services counts; the preceding group report's 119/10/27
whole-src counts include server/harness/testRunner. The local zero-escape count
is not a certificate about unknown external consumers of an exported function.

## Proposed rule A: a fresh push-only array certificate

Scope the first increment to a body consisting of allocation, one for-of loop,
and return, exactly the observed arrayFrom shape. Match by resolved bindings,
not the function name. Reuse the existing rigid overload mapping and iterable
parameter/for-of element provenance; do not infer a generic binder from one
lucky call.

1. Allocate a fresh empty array local. Track its allocation, not its annotation.
   Every reference before return is the receiver of the intrinsic push; the only
   other reference is the final return. Refuse rebinding, indexed/length writes,
   mutators, aliases, stores/passes, bound/extracted methods and captures, even
   captures that appear not to run. No wider mutable view may escape.
2. Whitelist the one loop over the admitted iterator parameter. Its const element
   binding carries the overload's element type. The pushed expression may be
   only that element or a direct call of the unchanged callback parameter with
   that element, selected by the bare callback parameter's presence test.
   No spread, getter/property-based guard, arbitrary local flow alias, cast,
   unknown call or new control construct becomes proof evidence.
3. Under overload 1, map is an admitted function, so only map(value) is reachable;
   its declared overload callback result is U. Under overload 2, the missing map
   is undefined, so only the input T value is reachable. The empty allocation
   vacuously satisfies either element promise. Each iteration preserves the
   invariant; the final return transfers the fresh allocation with that invariant.
4. Require intrinsic receiver/method stability, retaining existing checks against
   method/prototype writes. A syntactic `.push` is insufficient. Preserve all
   incoming callback and element contracts and invariant mutable-view checks.

This adds a bounded loop/local-return certificate to `censusReturnProof`'s result
road. It does not generally bless a local identifier or ask
`predicateOverloadResultProven` to treat a mapper as a predicate. Predicted
coverage: **1 overload owner**, with **4 overloaded** and **85 broad** neighbours
outside the rule. The zero-getter observation does not justify widening its
expression whitelist.

## Proposed rule B: a fresh set-only Map certificate with default provenance

Use the observed empty intrinsic Map allocation, canonical bounded index loop,
const element/key bindings, undefined-key guard, set and return. The measured
shape matches only arrayToMap. Do not generalize to arbitrary for-loop code or
accept arbitrary local casts just because the return is a Map.

1. Enforce the same non-escape/local-reference conditions as rule A, with intrinsic
   set as the sole mutation. Every set must prove both its key and value against
   the **overload's** Map<K,V> promise; merely seeing calls named set is not enough.
2. Recognize the canonical index loop and existing checked array element read.
   The element's provenance is the admitted input array. `makeKey(value)` has
   K|undefined and the immediately enclosing const-key test excludes undefined.
   The value is the direct result of the unchanged makeValue parameter. No getter
   narrowing or unrelated assignment may supply those facts.
3. Split incoming/default state correctly. For mapped overloads 2 and 4, makeValue
   is required **and excludes undefined**, so its default cannot execute. Its
   callback result is V2/U, not the implementation annotation's V1|V2 union.
   For omitted-mapper overloads, prove the actual default binding: resolve identity
   to its unmodified one-parameter `return x` body, and prove its binding stable.
   That path produces the admitted input value. A spelling `identity`, a function
   type annotation, or a supplied-but-undefined argument is not that proof.
4. Use the signature witness mapping already validated by censusOverload, including
   overloads with fewer binders. Do not copy censusReturnProof's positional
   binder-count refusal into the new certificate or erase the extra binder.

The general default ban stays in place. This certificate discharges this one
initializer using incoming non-undefined evidence or the independently proved
identity branch. Predicted coverage: **1 overload owner**, **0 other overloaded**
Map owners, and **3 broad** Map neighbours excluded. One default exposure is
accounted for, not silently ignored. Preserve intrinsic set/prototype stability
and the current Map mutable-value relation everywhere else.

## Proposed rule C: arity-partitioned, seeded fold certificate

There is one measured owner, reduceLeft. `arguments.length` facts alone cannot
solve its first refusal: the first overload also needs a new callback-result
provenance proof before `overload_callback_use.go:60`.

Use one coupled summary whose obligations are checked against each overload:

* Admit only the exact numeric `arguments.length <= 2` partition on the builtin,
  read-only arguments count of this ordinary function. No indexed arguments,
  alias, write, arrow capture, rest, default initializer or other arguments read
  is allowed. Keep `argumentsObjectRefusal`'s existing :46–69 checks.
* For the initialized overload, minimum arity 3 proves the false branch. Abstract
  accumulator type starts as the caller's U. The only recurrence is
  `result = f(result, checked array element, numeric index)`. Prove the callback's
  admitted arguments U,T,number and its declared result U; induction preserves U.
  Empty/undefined arrays return initial U without invoking f. No assumption U=T
  is needed or permitted.
* For the uninitialized overload, a **validated exact two-argument entry** proves
  the true branch; missing initial is undefined. Empty arrays return undefined.
  The nonempty path seeds from the existing checked first element T, advances
  the index and folds T,T,number -> T, giving T|undefined overall. Retain the
  existing checked index reads; this is not permission to erase assertions or
  accept unchecked sparse/out-of-range reads.
* Recognize the existing numeric size/index/bounds temporaries and redundant
  size>0 guard as part of this bounded fold summary. Do not teach all of
  censusTrustedTest that arbitrary local/property guards supply facts. Reject
  any other accumulator assignment, alias, closure capture, callback escape or
  parameter rebind. Prove every return; do not just notice one correctly typed
  callback call somewhere in a loop.
* **Arity gate:** written maximum 2 is not sufficient. Existing callable
  assignment can present a two-argument overload through a type with an extra
  argument; JavaScript still passes that argument. Require a whole-program
  call/value-reach certificate that all entries used for this exact-arity claim
  are checked two-argument calls, or an equivalent checked call contract that
  rejects the offending alias. Unknown/exported/opaque reach fails closed.
  `.call`, `.apply`, bound calls, spreads, callback values and callable storage
  need the same guarantee. Do not infer maximum from parameter-list length.

Measured source facts: **three memo assignments**, exactly the two seeds and
one recurrence; **one callback call**, **zero aliases/captures/defaults**, one
arguments-length comparison and no other arguments reads. All **27** direct
calls have known, non-spread arity; **6** use the two-argument overload. This
makes **one conditionally eligible owner, zero arguments-reader near misses**
in this source domain. The measured zero value uses is useful evidence, but
cannot by itself seal exported/unknown whole-program reach.

The certificate must discharge f's use, the seed parameter's use and returns
together, before the generic parameter loop at `census_small.go:186–211`.
Initial U is used only as the accumulator seed or directly returned; certify
those operations under U instead of pretending U relates to implementation T.
Changing only the callback-return test at `overload_callback_use.go:60` would
merely expose the independent initial-parameter mismatch. This is not a general
weakening of strict callback contravariance or parameter relatedness. The implementation's annotated T return is not evidence
that a heterogeneous U accumulator is T. Existing lowering representation/fitting
obligations still apply; a scout is not a claim that result fitting or codegen has
already been implemented. **Without the arity gate, accept zero owners under an
unconditional arity-only rule.**

## Unsound roads, each with a tiny executable TypeScript program

Every linked file is standalone and contains its overload promise, wider
implementation, typed call and JSON observation. [witnesses.cjs](witnesses.cjs)
checked each with stock tsc 6.0.3, strict mode, then executed its transpiled JS
in a fresh Node VM context. **15/15: zero diagnostics; 15/15: a wrong runtime
value type observed.** Full outputs and diagnostics: [witnesses.json](witnesses.json).

These demonstrate what an overly loose proof would bless. They are not claims
that current Adamic accepts the programs: existing arguments/prototype/variance
checks may reject them independently and must be retained.

| Rule/unsound relaxation | Tiny program | Promised → observed |
| --- | --- | --- |
| A: ignore indexed writes after push | [array-index-write.ts](counterexamples/array-index-write.ts) | string element → 0 |
| A: track only the local's spelling, allow aliases | [array-alias.ts](counterexamples/array-alias.ts) | string element → alias pushes 0 |
| A: ignore closure captures | [array-capture.ts](counterexamples/array-capture.ts) | string element → closure pushes 0 |
| A: trust narrowed property read across getter calls | [array-getter.ts](counterexamples/array-getter.ts) | string element → second getter read is 0 |
| A: ignore an optional mapper's default | [array-default.ts](counterexamples/array-default.ts) | string element → identity yields 0 |
| A: assume a method name guarantees intrinsic push | [array-prototype.ts](counterexamples/array-prototype.ts) | string element → replaced push writes 0 |
| B: every write is set, so ignore value types | [map-wrong-set.ts](counterexamples/map-wrong-set.ts) | string value → later set writes 0 |
| B: allow a second view of the Map | [map-alias.ts](counterexamples/map-alias.ts) | string value → alias overwrites with 0 |
| B: ignore closure captures | [map-capture.ts](counterexamples/map-capture.ts) | string value → closure overwrites with 0 |
| B: trust property narrowing across getter calls | [map-getter.ts](counterexamples/map-getter.ts) | string value → second getter read is 0 |
| B: present argument position means default cannot run | [map-default-undefined.ts](counterexamples/map-default-undefined.ts) | string value → supplied undefined selects identity, yields 0 |
| C: written maximum arity is runtime maximum | [arity-extra-alias.ts](counterexamples/arity-extra-alias.ts) | number result → extra seed is "bad" |
| C: allow other arguments reads/writes | [arity-arguments-write.ts](counterexamples/arity-arguments-write.ts) | number result → arguments[2] becomes "bad" |
| C: >=3 arguments proves seed wasn't defaulted | [arity-default.ts](counterexamples/arity-default.ts) | string result → explicit undefined defaults to 0 |
| C: arity alone proves recurrence arguments | [arity-wrong-recurrence.ts](counterexamples/arity-wrong-recurrence.ts) | string accumulator → input element passed as memo, yields 0 |

The getter examples have **no alias of the output and no closure capture of it**;
a getter's own state changes between the test and push/set. This is why an
otherwise clean producer still cannot trust arbitrary narrowed property reads.
The default examples show that arity and incoming undefined are separate facts.
The extra-argument example has zero stock diagnostics: it is not a call-site
error that the proposed proof may assume cannot arise.

## Evidence, limits and reproduction

[provenance.json](provenance.json) pins the base, parser, source scope and lowerer
hashes. [inventory.json](inventory.json) contains SHA-256 for all 247 measured
source files, all owner records, every broad container and hazard reference,
target resolved calls and the fold facts. [summary.json](summary.json) and
[inventory.log](inventory.log) contain the compact counts. Static shape counts
were measured before selecting the proposed rule boundaries; tightening the
profiles to the exact allocation/loop/return grammar retained one match each.

```sh
# From this evidence checkout after npm ci in stage3/api:
node --max-old-space-size=8000 stage3/census/overload-body-proof-scout/scout.cjs /path/to/base-adapted /scratch/inventory.json
node stage3/census/overload-body-proof-scout/witnesses.cjs /scratch/witnesses.json
```

Both final commands exited 0. The fixture runner exits nonzero if a fixture has
any stock diagnostic, throws unexpectedly, or does not demonstrate a broken
promise; wrappers return that child's exit code. No Go file or compiler rule
was created. No production census, lane, product build or new acceptance count
was run. No general semantic equivalence claim follows from a shape match.

Before implementing, the compiler task should turn the three bounded summaries
and all 15 negative witnesses into positive/negative proof tests, settle the
arity-entry guard, and measure the actual census delta without relaxing existing
parameter, predicate, invariant-view or runtime assertion checks. This scout
provides the measured scope and required guards for that decision.
