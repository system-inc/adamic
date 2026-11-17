# Core overloads — #ev14ch9

Base: cf468df0747075e50d326a97ff03e8530d7a1a88. Upstream TypeScript v6.0.3:
050880ce59e30b356b686bd3144efe24f875ebc8. This report is checkpointed before
building any proposed adaptation. There are four sets and one supplied latent
site per set. No landing change has been built or committed.

| Set | Upstream src resolved calls | Initial road | Build count |
| --- | ---: | --- | ---: |
| arrayFrom | 119 (97 without map, 22 with map) | (iii) stop; investigate truthful union versus mapped callers | 0 |
| arrayToMap | 10 (5 identity, 5 mapped) | (iii) stop; investigate invariant Map union versus mapped callers | 0 |
| reduceLeft | 27 (6 without initial, 21 with initial) | (iii) stop; callback accumulator proof requires loop/arguments reasoning | 0 |
| sameMap | 36 | (iii) stop; readonly implementation cannot promise a mutable alias | 0 |

These are source-checker counts, including src/harness and src/testRunner, not
text matches. The whole-checkout traversal and fully adapted base traversal
are still running; final counts and consumer classifications will follow in a
checkpoint on this branch. The checker is stock TypeScript 6.0.3 and each counted
call's getResolvedSignature declaration is the named function in compiler/core.ts.

## Evidence read before road choice

Read whole: internal/lower/census_small.go, overload_predicate_result.go,
census_overload_proof.go, overload_result_check.go. Also read predicateEffects
in predicates.go and overloadCallbackUse in overload_callback_use.go.

The return proof rejects defaults, assigned parameters, reached casts, loops,
returned local identifiers, and calls/new/property reads in return expressions.
The call-site narrowing road excludes generic implementations and arrays; a Map
has no supported literal discriminant or class-instance check here. No callback
in these four sets is a result-narrowing predicate. Presence alone cannot erase
an element union, Map value union or readonly/mutable difference.

Ownership search over every owners.json, targets.json and adapt.cjs finds a
whole-owner pin only for sameMap, in 99-correctness-axolotl-k. Any sameMap edit
must amend that adapter. 49-honest-optionals edits reduceLeft's callback memo
and local result to include undefined and removes initial!. It does not pin the
whole owner. 76-truthful-casts only mentions sameMap as pending in its log.

## Preliminary blockers and candidate roads

arrayFrom: the implementation returns a local (T | U)[], while overload 1
promises U[]. Mapped callers include module-path tuples mapped to ModulePath,
import specifiers mapped to binding elements, and path strings mapped to file
IDs. T extends U is not a truthful constraint on these calls. A wider overload
would expose T to consumers expecting only U. A branch-per-map rewrite still
returns local arrays and iterates; the existing proof cannot prove it. Calling
another builder in the return is effectful and also fails the proof. A new
shape or a rule for loop-built containers needs a ruling.

arrayToMap: overload 2 promises Map<K,V2> while the implementation returns
Map<K,V1|V2>. The identity default is essential on the five two-argument calls.
Mapped calls project file-info objects to booleans and diagnostics tuples to
diagnostic arrays. V1 extends V2 does not hold. A union Map is invariant and
cannot be used as the projected Map. Keeping one implementation for both paths
needs correlation the declaration comparison does not express. A default also
blocks the return proof, which cannot read the map-building loop or local
return even if the default were restructured. Splitting the helper is a new
shape, so it is a stop candidate rather than an adapter.

reduceLeft: the adapted implementation's callback takes memo T|undefined,
but overload 1's accumulator is U independently of array element T; overload 2
also admits callbacks taking only T. Strict callback contravariance refuses
this. The callback-use exception explicitly refuses arguments reads, and the
body uses arguments.length to select initialization, then repeatedly passes a
mutable local accumulator. Dropping the uninitialized overload does not solve
the heterogeneous initialized calls. Reintroducing a required accumulator or
cast would undo 49-honest-optionals' correction. A proof for seeded iteration,
or splitting initialization paths into a new shape, requires a ruling.

sameMap: axolotl-k already fixes T extends U and preserves unchanged-array
identity. Its implementation result remains readonly U[]|undefined while the
first overload promises U[]. Presence only cannot change readonly to mutable.
Returning a copy would change identity and is prohibited. Dropping mutable
overloads or widening their results needs caller assessment; the return proof
cannot handle the existing loops and local return, nor its reached comparison
cast. Conditional/generic correlation between input mutability and return
mutability would require new body evidence rather than an unchecked annotation.

## Proof status

For each set: scan-all, full single-file latent census (owner and callers),
restoring mutant, stock 79-root tsc delta, adamic types, changed apply, public dts
comparison, TestProduct_Stage3AdaptedTree, whole lane, and ledger: **not run**.
There is no proposed adapter to validate. Preparing the unchanged base adapted
tree for reading is separate and does not establish any of these proof lines.
No landing branch or SHA exists. No production lowerer or public declaration
has been edited. Evidence and subsequent updates belong only on this branch.
